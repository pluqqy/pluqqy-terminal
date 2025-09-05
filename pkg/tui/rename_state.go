package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

// RenameState manages the state for renaming components and pipelines
type RenameState struct {
	Active          bool     // Whether rename mode is active
	ItemType        string   // "component" or "pipeline"
	OriginalName    string   // Current display name
	OriginalPath    string   // Current file path
	NewName         string   // User input for new display name
	CursorPos       int      // Current cursor position in the input
	ValidationError string   // Real-time validation error
	AffectedActive  []string // Active pipelines affected (display names)
	AffectedArchive []string // Archived pipelines affected (display names)
	IsArchived      bool     // Whether the item being renamed is archived
}

// NewRenameState creates a new rename state instance
func NewRenameState() *RenameState {
	return &RenameState{}
}

// HandleInput processes keyboard input for rename mode
func (rs *RenameState) HandleInput(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if !rs.Active {
		return false, nil
	}

	switch msg.String() {
	case "esc":
		rs.Reset()
		return true, nil

	case "enter":
		if rs.NewName != "" && rs.ValidationError == "" {
			// Trigger rename operation
			return true, rs.executeRename()
		}
		return true, nil

	case "backspace":
		if len(rs.NewName) > 0 && rs.CursorPos > 0 {
			// Handle UTF-8 properly
			runes := []rune(rs.NewName)
			// Remove character before cursor
			rs.NewName = string(runes[:rs.CursorPos-1]) + string(runes[rs.CursorPos:])
			rs.CursorPos--
			rs.validate()
		}
		return true, nil

	case "delete":
		runes := []rune(rs.NewName)
		if rs.CursorPos < len(runes) {
			// Remove character at cursor
			rs.NewName = string(runes[:rs.CursorPos]) + string(runes[rs.CursorPos+1:])
			rs.validate()
		}
		return true, nil

	case "left", "ctrl+b":
		if rs.CursorPos > 0 {
			rs.CursorPos--
		}
		return true, nil

	case "right", "ctrl+f":
		runes := []rune(rs.NewName)
		if rs.CursorPos < len(runes) {
			rs.CursorPos++
		}
		return true, nil

	case "home", "ctrl+a":
		rs.CursorPos = 0
		return true, nil

	case "end", "ctrl+e":
		runes := []rune(rs.NewName)
		rs.CursorPos = len(runes)
		return true, nil

	case "ctrl+u":
		// Clear from cursor to beginning
		if rs.CursorPos > 0 {
			runes := []rune(rs.NewName)
			rs.NewName = string(runes[rs.CursorPos:])
			rs.CursorPos = 0
			rs.validate()
		}
		return true, nil

	case "ctrl+k":
		// Clear from cursor to end
		runes := []rune(rs.NewName)
		if rs.CursorPos < len(runes) {
			rs.NewName = string(runes[:rs.CursorPos])
			rs.validate()
		}
		return true, nil

	case " ":
		// Insert space at cursor position
		runes := []rune(rs.NewName)
		rs.NewName = string(runes[:rs.CursorPos]) + " " + string(runes[rs.CursorPos:])
		rs.CursorPos++
		rs.validate()
		return true, nil

	case "tab":
		// Ignore tab in rename mode
		return true, nil

	default:
		if msg.Type == tea.KeyRunes {
			// Insert characters at cursor position
			runes := []rune(rs.NewName)
			newRunes := string(msg.Runes)
			rs.NewName = string(runes[:rs.CursorPos]) + newRunes + string(runes[rs.CursorPos:])
			rs.CursorPos += len([]rune(newRunes))
			rs.validate()
			return true, nil
		}
	}

	return false, nil
}

// Start initiates rename mode for an item
func (rs *RenameState) Start(displayName, itemType, path string, isArchived bool) {
	rs.Active = true
	rs.ItemType = itemType
	rs.OriginalName = displayName
	rs.OriginalPath = path
	rs.NewName = displayName                // Pre-fill with current name
	rs.CursorPos = len([]rune(displayName)) // Set cursor at end
	rs.ValidationError = ""
	rs.IsArchived = isArchived
	rs.AffectedActive = nil
	rs.AffectedArchive = nil

	// If renaming a component, find affected pipelines
	if itemType == "component" {
		rs.findAffectedPipelines()
	}
}

// validate checks if the new name is valid
func (rs *RenameState) validate() {
	// Clear previous error
	rs.ValidationError = ""

	// Check for empty name
	if strings.TrimSpace(rs.NewName) == "" {
		rs.ValidationError = "Name cannot be empty"
		return
	}

	// Check if name hasn't changed
	if rs.NewName == rs.OriginalName {
		return // No error, but also no change
	}

	// Validate the rename operation
	err := files.ValidateRename(rs.OriginalPath, rs.NewName, rs.ItemType)
	if err != nil {
		rs.ValidationError = err.Error()
	}
}

// Reset clears the rename state
func (rs *RenameState) Reset() {
	rs.Active = false
	rs.ItemType = ""
	rs.OriginalName = ""
	rs.OriginalPath = ""
	rs.NewName = ""
	rs.CursorPos = 0
	rs.ValidationError = ""
	rs.AffectedActive = nil
	rs.AffectedArchive = nil
	rs.IsArchived = false
}

// executeRename performs the actual rename operation
func (rs *RenameState) executeRename() tea.Cmd {
	return func() tea.Msg {
		var err error

		// Perform the rename based on item type and archive status
		if rs.IsArchived {
			if rs.ItemType == "component" {
				err = files.RenameComponentInArchive(rs.OriginalPath, rs.NewName)
			} else {
				err = files.RenamePipelineInArchive(rs.OriginalPath, rs.NewName)
			}
		} else {
			if rs.ItemType == "component" {
				err = files.RenameComponent(rs.OriginalPath, rs.NewName)
			} else {
				err = files.RenamePipeline(rs.OriginalPath, rs.NewName)
			}
		}

		if err != nil {
			return RenameErrorMsg{Error: err}
		}

		return RenameSuccessMsg{
			ItemType:   rs.ItemType,
			OldName:    rs.OriginalName,
			NewName:    rs.NewName,
			IsArchived: rs.IsArchived,
		}
	}
}

// findAffectedPipelines finds pipelines that reference the component
func (rs *RenameState) findAffectedPipelines() {
	active, archived, err := files.FindAffectedPipelines(rs.OriginalPath)
	if err == nil {
		rs.AffectedActive = active
		rs.AffectedArchive = archived
	}
}

// GetSlugifiedName returns what the filename will become
func (rs *RenameState) GetSlugifiedName() string {
	if rs.NewName == "" {
		return ""
	}
	return files.Slugify(rs.NewName)
}

// HasAffectedPipelines returns true if there are pipelines that will be updated
func (rs *RenameState) HasAffectedPipelines() bool {
	return len(rs.AffectedActive) > 0 || len(rs.AffectedArchive) > 0
}

// IsValid returns true if the current name is valid for renaming
func (rs *RenameState) IsValid() bool {
	return rs.NewName != "" &&
		rs.ValidationError == "" &&
		rs.NewName != rs.OriginalName
}

// IsActive returns whether rename mode is active
func (rs *RenameState) IsActive() bool {
	return rs.Active
}

// GetError returns the validation error if any
func (rs *RenameState) GetError() error {
	if rs.ValidationError != "" {
		return nil // We have a validation error but not a general error
	}
	return nil
}

// GetItemType returns the type of item being renamed
func (rs *RenameState) GetItemType() string {
	return rs.ItemType
}

// GetNewName returns the new name entered by the user
func (rs *RenameState) GetNewName() string {
	return rs.NewName
}

// StartRename initiates rename mode for an item
func (rs *RenameState) StartRename(path, displayName, itemType string) {
	rs.Start(displayName, itemType, path, false) // Default to not archived
}

// RenameSuccessMsg is sent when a rename operation succeeds
type RenameSuccessMsg struct {
	ItemType   string
	OldName    string
	NewName    string
	IsArchived bool
}

// RenameErrorMsg is sent when a rename operation fails
type RenameErrorMsg struct {
	Error error
}
