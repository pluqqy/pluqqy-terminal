package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
)

// TagEditorConfig contains configuration for the tag editor
type TagEditorConfig struct {
	Width  int
	Height int
}

// TagEditorMode represents the current mode of the tag editor
type TagEditorMode int

const (
	TagEditorModeNormal TagEditorMode = iota
	TagEditorModeCloud
	TagEditorModeDeleting
	TagEditorModeReloading
)

// TagEditResult represents the result of a tag editing session
type TagEditResult struct {
	Saved    bool
	Tags     []string
	Canceled bool
}

// TagEditorCallbacks contains callback functions for tag editor operations
type TagEditorCallbacks struct {
	// OnSave is called when tags are saved
	OnSave func(path string, tags []string) error
	
	// OnExit is called when the editor is closed
	OnExit func(saved bool)
	
	// OnReload is called when tags need to be reloaded
	OnReload func()
}

// TagEditorState represents the complete state of the tag editor
type TagEditorState struct {
	// Core state
	Active       bool
	Mode         TagEditorMode
	Path         string
	ItemType     string // "component" or "pipeline"
	ItemName     string
	
	// Tags
	CurrentTags   []string
	OriginalTags  []string
	AvailableTags []string
	
	// Input handling
	TagInput                string
	TagCursor               int
	ShowSuggestions         bool
	SuggestionCursor        int
	HasNavigatedSuggestions bool
	
	// Tag cloud navigation
	TagCloudActive bool
	TagCloudCursor int
	
	// Deletion state
	DeletingTag      string
	DeletingTagUsage *tags.UsageStats
	
	// Dimensions
	Width  int
	Height int
	
	// Callbacks
	Callbacks TagEditorCallbacks
}

// TagDeletionResult holds the results of a comprehensive tag deletion
type TagDeletionResult struct {
	TagName      string
	FilesUpdated int
	FilesScanned int
	Errors       []string
}

// tagDeletionProgressMsg is sent during long-running deletion operations
type tagDeletionProgressMsg struct {
	CurrentFile string
	Progress    int
	Total       int
}

// tagDeletionCompleteMsg is sent when deletion is complete
type tagDeletionCompleteMsg struct {
	Result TagDeletionResult
}

// TagDeletionState manages the spinner and progress for tag deletion
type TagDeletionState struct {
	Active       bool
	Spinner      spinner.Model
	Progress     string
	CurrentFile  string
	FilesScanned int
	FilesUpdated int
}

// NewTagDeletionState creates a new tag deletion state with spinner
func NewTagDeletionState() *TagDeletionState {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	
	return &TagDeletionState{
		Spinner: s,
	}
}

// Start begins the deletion process with spinner
func (tds *TagDeletionState) Start() tea.Cmd {
	tds.Active = true
	tds.FilesScanned = 0
	tds.FilesUpdated = 0
	tds.Progress = "Starting tag deletion..."
	return tds.Spinner.Tick
}

// Update handles spinner updates
func (tds *TagDeletionState) Update(msg tea.Msg) (bool, tea.Cmd) {
	if !tds.Active {
		return false, nil
	}
	
	switch msg := msg.(type) {
	case tagDeletionProgressMsg:
		tds.CurrentFile = msg.CurrentFile
		tds.FilesScanned = msg.Progress
		tds.Progress = fmt.Sprintf("Scanning files... (%d/%d)", msg.Progress, msg.Total)
		return true, nil
		
	case tagDeletionCompleteMsg:
		tds.Active = false
		return true, nil
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		tds.Spinner, cmd = tds.Spinner.Update(msg)
		return true, cmd
	}
	
	return false, nil
}

// View renders the spinner and progress
func (tds *TagDeletionState) View() string {
	if !tds.Active {
		return ""
	}
	
	// Add padding to avoid overlaying on borders
	spinnerLine := fmt.Sprintf("  %s %s", tds.Spinner.View(), tds.Progress)
	// Return with a newline at the beginning to position below the border
	return "\n" + spinnerLine
}