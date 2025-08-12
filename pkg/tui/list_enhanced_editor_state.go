package tui

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// EditorMode represents the current mode of the enhanced editor
type EditorMode int

const (
	// EditorModeNormal is the default editing mode
	EditorModeNormal EditorMode = iota
	// EditorModeFilePicking is when the file picker is active
	EditorModeFilePicking
)

// EnhancedEditorState manages ONLY the state of the enhanced editor - no business logic
type EnhancedEditorState struct {
	// Basic state
	Active         bool
	Mode           EditorMode
	
	// Component information
	ComponentPath  string
	ComponentName  string
	ComponentType  string
	ComponentTags  []string // Store original tags to preserve them
	
	// Content management
	Content        string
	OriginalContent string
	CursorPosition int
	InsertionPoint int
	
	// UI components
	Textarea       textarea.Model
	FilePicker     filepicker.Model
	
	// Change tracking
	UnsavedChanges bool
	
	// Exit confirmation
	ExitConfirm    *ConfirmationModel
	ExitConfirmActive bool
}

// NewEnhancedEditorState creates a new enhanced editor state
func NewEnhancedEditorState() *EnhancedEditorState {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = "  " // Use spaces instead of vertical line
	ta.CharLimit = 0 // No limit
	ta.SetWidth(80)
	ta.SetHeight(20)
	
	fp := filepicker.New()
	fp.ShowHidden = false
	fp.DirAllowed = true
	fp.FileAllowed = true
	// Set to show all files by default
	fp.AllowedTypes = []string{}
	
	return &EnhancedEditorState{
		Active:      false,
		Mode:        EditorModeNormal,
		Textarea:    ta,
		FilePicker:  fp,
		ExitConfirm: NewConfirmation(),
	}
}

// SetMode updates the editor mode (state update only)
func (e *EnhancedEditorState) SetMode(mode EditorMode) {
	e.Mode = mode
}

// UpdateCursor tracks cursor position (state update only)
func (e *EnhancedEditorState) UpdateCursor(pos int) {
	e.CursorPosition = pos
}

// SetContent updates the content (state update only)
func (e *EnhancedEditorState) SetContent(content string) {
	e.Content = content
	e.Textarea.SetValue(content)
	e.UnsavedChanges = (content != e.OriginalContent)
}

// IsActive checks if the editor is active
func (e *EnhancedEditorState) IsActive() bool {
	return e.Active
}

// HasUnsavedChanges checks for unsaved changes
func (e *EnhancedEditorState) HasUnsavedChanges() bool {
	return e.UnsavedChanges
}

// StartEditing initializes the editor with a component
func (e *EnhancedEditorState) StartEditing(path, name, compType, content string, tags []string) {
	e.Active = true
	e.Mode = EditorModeNormal
	e.ComponentPath = path
	e.ComponentName = name
	e.ComponentType = compType
	e.ComponentTags = tags // Store tags to preserve them on save
	e.Content = content
	e.OriginalContent = content
	e.Textarea.SetValue(content)
	e.Textarea.Focus()
	e.UnsavedChanges = false
	e.ExitConfirmActive = false
	e.CursorPosition = 0
	e.InsertionPoint = 0
}

// Reset clears the editor state
func (e *EnhancedEditorState) Reset() {
	e.Active = false
	e.Mode = EditorModeNormal
	e.ComponentPath = ""
	e.ComponentName = ""
	e.ComponentType = ""
	e.ComponentTags = nil
	e.Content = ""
	e.OriginalContent = ""
	e.CursorPosition = 0
	e.InsertionPoint = 0
	e.UnsavedChanges = false
	e.ExitConfirmActive = false
	e.Textarea.Reset()
	e.Textarea.Blur()
}

// UpdateTextarea updates the textarea model (for bubbletea updates)
func (e *EnhancedEditorState) UpdateTextarea(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	e.Textarea, cmd = e.Textarea.Update(msg)
	
	// Track content changes
	newContent := e.Textarea.Value()
	if newContent != e.Content {
		e.Content = newContent
		e.UnsavedChanges = (newContent != e.OriginalContent)
	}
	
	return cmd
}

// UpdateFilePicker updates the file picker model (for bubbletea updates)
func (e *EnhancedEditorState) UpdateFilePicker(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	e.FilePicker, cmd = e.FilePicker.Update(msg)
	return cmd
}

// SetTextareaDimensions updates the textarea dimensions
func (e *EnhancedEditorState) SetTextareaDimensions(width, height int) {
	e.Textarea.SetWidth(width)
	e.Textarea.SetHeight(height)
}

// GetMode returns the current editor mode
func (e *EnhancedEditorState) GetMode() EditorMode {
	return e.Mode
}

// SetInsertionPoint marks where to insert file reference
func (e *EnhancedEditorState) SetInsertionPoint(pos int) {
	e.InsertionPoint = pos
}

// GetInsertionPoint returns the insertion point for file references
func (e *EnhancedEditorState) GetInsertionPoint() int {
	return e.InsertionPoint
}

// StartFilePicker activates the file picker mode
func (e *EnhancedEditorState) StartFilePicker() {
	e.Mode = EditorModeFilePicking
	e.InsertionPoint = e.CursorPosition
	// Don't recreate the file picker, just update its settings
	// The picker will be properly initialized by InitializeEnhancedFilePicker
}

// StopFilePicker deactivates the file picker mode
func (e *EnhancedEditorState) StopFilePicker() {
	e.Mode = EditorModeNormal
	e.Textarea.Focus()
}

// IsFilePicking checks if currently picking a file
func (e *EnhancedEditorState) IsFilePicking() bool {
	return e.Mode == EditorModeFilePicking
}

// ShowExitConfirmation shows the exit confirmation dialog
func (e *EnhancedEditorState) ShowExitConfirmation(width int, onConfirm, onCancel func() tea.Cmd) {
	e.ExitConfirmActive = true
	e.ExitConfirm.ShowDialog(
		"⚠️  Unsaved Changes",
		"You have unsaved changes to this component.",
		"Exit without saving?",
		true, // destructive
		width - 4,
		10,
		onConfirm,
		onCancel,
	)
}

// HideExitConfirmation hides the exit confirmation dialog
func (e *EnhancedEditorState) HideExitConfirmation() {
	e.ExitConfirmActive = false
}

// IsExitConfirmActive checks if exit confirmation is showing
func (e *EnhancedEditorState) IsExitConfirmActive() bool {
	return e.ExitConfirmActive
}