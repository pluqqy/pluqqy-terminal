package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

// ComponentEditor manages the state for component editing
type ComponentEditor struct {
	// Editing state
	Active            bool
	ComponentPath     string
	ComponentName     string
	Content           string
	OriginalContent   string
	EditViewport      viewport.Model
	
	// Exit confirmation
	ExitConfirm       *ConfirmationModel
	ExitConfirmActive bool
}

// NewComponentEditor creates a new component editor
func NewComponentEditor() *ComponentEditor {
	return &ComponentEditor{
		EditViewport: viewport.New(80, 20), // Default size, will be adjusted
		ExitConfirm:  NewConfirmation(),
	}
}

// IsActive returns whether component editing is active
func (e *ComponentEditor) IsActive() bool {
	return e.Active
}

// StartEditing initializes the editor with a component
func (e *ComponentEditor) StartEditing(path, name, content string) {
	e.Active = true
	e.ComponentPath = path
	e.ComponentName = name
	e.Content = content
	e.OriginalContent = content
	e.ExitConfirmActive = false
}

// HandleInput processes keyboard input during component editing
func (e *ComponentEditor) HandleInput(msg tea.KeyMsg, width, height int) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Handle exit confirmation if active
	if e.ExitConfirmActive {
		cmd = e.ExitConfirm.Update(msg)
		if cmd != nil {
			return nil, cmd
		}
	}
	
	// Handle viewport scrolling
	switch msg.String() {
	case "up", "k", "pgup":
		e.EditViewport, cmd = e.EditViewport.Update(msg)
		return nil, cmd
	case "down", "j", "pgdown":
		e.EditViewport, cmd = e.EditViewport.Update(msg)
		return nil, cmd
	case "ctrl+s":
		// Save component and exit
		return nil, e.saveComponent()
	case "ctrl+x":
		// Save current content and open in external editor
		// First save any unsaved changes
		if e.Content != e.OriginalContent {
			err := files.WriteComponent(e.ComponentPath, e.Content)
			if err != nil {
				return nil, func() tea.Msg {
					return StatusMsg(fmt.Sprintf("× Failed to save before external edit: %v", err))
				}
			}
		}
		// Open in external editor
		return nil, e.openInEditor()
	case "esc":
		// Check if content has changed
		if e.Content != e.OriginalContent {
			// Show confirmation dialog
			e.ExitConfirmActive = true
			e.ExitConfirm.ShowDialog(
				"⚠️  Unsaved Changes",
				"You have unsaved changes to this component.",
				"Exit without saving?",
				true, // destructive
				width - 4,
				10,
				func() tea.Cmd {
					// Exit without saving
					e.Reset()
					return nil
				},
				func() tea.Cmd {
					// Cancel exit
					e.ExitConfirmActive = false
					return nil
				},
			)
			return nil, nil
		}
		// No changes, exit immediately
		e.Reset()
		return nil, nil
	case "enter":
		e.Content += "\n"
	case "backspace":
		if len(e.Content) > 0 {
			e.Content = e.Content[:len(e.Content)-1]
		}
	case "tab":
		e.Content += "    "
	case " ":
		e.Content += " "
	default:
		if msg.Type == tea.KeyRunes {
			e.Content += string(msg.Runes)
		}
	}
	
	return nil, nil
}

// Reset clears the editing state
func (e *ComponentEditor) Reset() {
	e.Active = false
	e.ComponentPath = ""
	e.ComponentName = ""
	e.Content = ""
	e.OriginalContent = ""
	e.ExitConfirmActive = false
}

// saveComponent saves the edited component
func (e *ComponentEditor) saveComponent() tea.Cmd {
	return func() tea.Msg {
		// Write component
		err := files.WriteComponent(e.ComponentPath, e.Content)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save: %v", err))
		}
		
		// Clear editing state
		savedName := e.ComponentName
		e.Reset()
		
		// Return a reload message to refresh data
		return ReloadMsg{
			Message: fmt.Sprintf("✓ Saved: %s", savedName),
		}
	}
}

// openInEditor opens the component in an external editor
func (e *ComponentEditor) openInEditor() tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set. Please set it to your preferred editor.")
		}

		// Construct full path
		fullPath := filepath.Join(files.PluqqyDir, e.ComponentPath)
		
		// Create command with proper argument parsing for editors with flags
		cmd := createEditorCommand(editor, fullPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}

		// Return a reload message to refresh data
		return ReloadMsg{
			Message: fmt.Sprintf("Edited: %s", filepath.Base(e.ComponentPath)),
		}
	}
}

// GetViewport returns the current viewport for rendering
func (e *ComponentEditor) GetViewport() viewport.Model {
	return e.EditViewport
}

// UpdateViewport updates the viewport with new content
func (e *ComponentEditor) UpdateViewport(width, height int) {
	// Update viewport dimensions if needed
	viewportWidth := width - 8  // Account for borders and padding
	viewportHeight := height - 12 // Account for header, help, etc.
	
	if e.EditViewport.Width != viewportWidth || e.EditViewport.Height != viewportHeight {
		e.EditViewport.Width = viewportWidth
		e.EditViewport.Height = viewportHeight
	}
}

// HasUnsavedChanges returns true if there are unsaved changes
func (e *ComponentEditor) HasUnsavedChanges() bool {
	return e.Content != e.OriginalContent
}