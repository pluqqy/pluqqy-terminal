package tui

import (
	"fmt"
	"strings"

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

// UndoState represents a saved state for undo functionality
type UndoState struct {
	Content     string
	CursorPos   int
	Description string // What action created this state
}

// EnhancedEditorState manages ONLY the state of the enhanced editor - no business logic
// Uses composition pattern to organize related fields
type EnhancedEditorState struct {
	EditorCore
	EditorUIState
	EditorOperations
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
		EditorCore: EditorCore{
			Active:     false,
			Mode:       EditorModeNormal,
			Textarea:   ta,
			FilePicker: fp,
		},
		EditorUIState: EditorUIState{
			ExitConfirm: NewConfirmation(),
		},
		EditorOperations: EditorOperations{
			ClipboardStatus: &ClipboardStatus{},
			StatusManager:   NewStatusManager(),
			ActionFeedback:  NewEditorActionFeedback(),
			PasteHelper:     NewPasteHelper(),
			RecentFiles:     NewRecentFilesTracker(),
			UndoStack:       make([]UndoState, 0, 10),
			MaxUndoLevels:   10,
		},
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
	e.EditorCore.Content = content
	e.EditorCore.Textarea.SetValue(content)
	e.EditorOperations.UnsavedChanges = (content != e.EditorCore.OriginalContent)
}

// IsActive checks if the editor is active
func (e *EnhancedEditorState) IsActive() bool {
	return e.EditorCore.Active
}

// GetContent returns the current content from the textarea
func (e *EnhancedEditorState) GetContent() string {
	if e.EditorCore.Textarea.Value() != "" {
		return e.EditorCore.Textarea.Value()
	}
	return e.EditorCore.Content
}

// HasUnsavedChanges checks for unsaved changes
func (e *EnhancedEditorState) HasUnsavedChanges() bool {
	return e.EditorOperations.UnsavedChanges
}

// StartEditing initializes the editor with a component
func (e *EnhancedEditorState) StartEditing(path, name, compType, content string, tags []string) {
	e.EditorCore.Active = true
	e.EditorCore.Mode = EditorModeNormal
	e.EditorCore.ComponentPath = path
	e.EditorCore.ComponentName = name
	e.EditorCore.ComponentType = compType
	e.EditorCore.ComponentTags = tags // Store tags to preserve them on save
	e.EditorCore.Content = content
	e.EditorCore.OriginalContent = content
	
	// Set the content and focus the textarea
	e.EditorCore.Textarea.SetValue(content)
	e.EditorCore.Textarea.Focus()
	
	e.EditorOperations.UnsavedChanges = false
	e.EditorUIState.ExitConfirmActive = false
	e.EditorUIState.CursorPosition = 0
	e.EditorUIState.InsertionPoint = 0
}

// Reset clears the editor state
func (e *EnhancedEditorState) Reset() {
	e.EditorCore.Active = false
	e.EditorCore.Mode = EditorModeNormal
	e.EditorCore.ComponentPath = ""
	e.EditorCore.ComponentName = ""
	e.EditorCore.ComponentType = ""
	e.EditorCore.ComponentTags = nil
	e.EditorCore.Content = ""
	e.EditorCore.OriginalContent = ""
	e.EditorUIState.CursorPosition = 0
	e.EditorUIState.InsertionPoint = 0
	e.EditorOperations.UnsavedChanges = false
	e.EditorUIState.ExitConfirmActive = false
	e.EditorCore.Textarea.Reset()
	e.EditorCore.Textarea.Blur()
}

// UpdateTextarea updates the textarea model (for bubbletea updates)
func (e *EnhancedEditorState) UpdateTextarea(msg tea.Msg) tea.Cmd {
	// Check if this is a paste event (detect content size jump)
	oldContent := e.Textarea.Value()
	oldLen := len(oldContent)

	// Save undo state periodically (every significant change)
	shouldSaveUndo := false

	var cmd tea.Cmd
	// Explicitly update the textarea in the embedded EditorCore struct
	e.EditorCore.Textarea, cmd = e.EditorCore.Textarea.Update(msg)

	// Track content changes
	newContent := e.EditorCore.Textarea.Value()
	if newContent != e.EditorCore.Content {
		// Check if this looks like a paste (significant content increase or has TUI borders)
		newLen := len(newContent)
		isPaste := false
		pastedContent := ""

		// Detect paste by content size jump
		if newLen > oldLen+10 && oldLen >= 0 {
			isPaste = true
			pastedContent = newContent[oldLen:]
			shouldSaveUndo = true // Save undo before paste
		} else if strings.Contains(newContent, "│") && !strings.Contains(oldContent, "│") {
			// Also detect paste if new content suddenly has TUI borders
			isPaste = true
			pastedContent = newContent
			shouldSaveUndo = true
		}

		// Save undo state for significant changes (not every keystroke)
		// Check for: newline added, significant length change, or periodic saves
		if !isPaste && oldContent != "" {
			// Save on newline, deletion of multiple chars, or every 20 char changes
			oldLines := strings.Count(oldContent, "\n")
			newLines := strings.Count(newContent, "\n")
			lengthDiff := len(newContent) - len(oldContent)

			if oldLines != newLines || // New line added/removed
				lengthDiff < -5 || // Deleted 5+ chars
				lengthDiff > 20 || // Added 20+ chars
				(len(e.UndoStack) == 0 && lengthDiff != 0) { // First change
				shouldSaveUndo = true
			}
		}

		// Save undo state if needed
		if shouldSaveUndo && oldContent != "" {
			e.SaveUndoState("Edit")
		}

		if isPaste && pastedContent != "" {
			// Clean the pasted content
			cleanedPaste := e.PasteHelper.CleanPastedContent(pastedContent)

			// If content was cleaned, update it
			if cleanedPaste != pastedContent {
				var cleanedContent string
				if pastedContent == newContent {
					// Entire content was pasted
					cleanedContent = cleanedPaste
				} else {
					// Appended paste
					cleanedContent = oldContent + cleanedPaste
				}
				e.EditorCore.Textarea.SetValue(cleanedContent)
				e.EditorCore.Content = cleanedContent
				e.ActionFeedback.RecordAction(fmt.Sprintf("✓ Pasted %d lines (cleaned)", CountLines(cleanedPaste)))
			} else {
				e.EditorCore.Content = newContent
				e.ActionFeedback.RecordAction(fmt.Sprintf("✓ Pasted %d lines", CountLines(pastedContent)))
			}
		} else {
			e.EditorCore.Content = newContent
		}

		e.UnsavedChanges = (e.EditorCore.Content != e.EditorCore.OriginalContent)
		e.UpdateStats()
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
		width-4,
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

// UpdateClipboardStatus updates the clipboard status
func (e *EnhancedEditorState) UpdateClipboardStatus(content string) {
	if content == "" {
		e.ClipboardStatus.HasContent = false
		e.ClipboardStatus.LineCount = 0
		e.ClipboardStatus.WillClean = false
		return
	}

	e.ClipboardStatus.HasContent = true
	e.ClipboardStatus.LineCount = CountLines(content)
	e.ClipboardStatus.WillClean = e.PasteHelper.WillClean(content)
}

// SaveUndoState saves the current state to the undo stack
func (e *EnhancedEditorState) SaveUndoState(description string) {
	state := UndoState{
		Content:     e.Content,
		CursorPos:   e.CursorPosition,
		Description: description,
	}

	// Add to stack
	e.UndoStack = append(e.UndoStack, state)

	// Trim stack if it exceeds max
	if len(e.UndoStack) > e.MaxUndoLevels {
		e.UndoStack = e.UndoStack[len(e.UndoStack)-e.MaxUndoLevels:]
	}
}

// Undo restores the previous state
func (e *EnhancedEditorState) Undo() bool {
	if len(e.UndoStack) == 0 {
		return false
	}

	// Pop the last state
	lastIndex := len(e.UndoStack) - 1
	state := e.UndoStack[lastIndex]
	e.UndoStack = e.UndoStack[:lastIndex]

	// Restore state
	e.Content = state.Content
	e.Textarea.SetValue(state.Content)
	e.CursorPosition = state.CursorPos
	e.UnsavedChanges = (e.Content != e.OriginalContent)

	// Update stats
	e.UpdateStats()

	return true
}

// UpdateStats updates line count, word count, and cursor position
func (e *EnhancedEditorState) UpdateStats() {
	content := e.Textarea.Value()

	// Line count
	e.LineCount = CountLines(content)

	// Word count
	e.WordCount = CountWords(content)

	// Cursor position is tracked internally by textarea
	// We'll use the simple position for now
	e.CurrentLine = 0 // Will be updated by textarea events
	e.CurrentColumn = 0
}

// CountWords counts words in content
func CountWords(content string) int {
	if content == "" {
		return 0
	}
	words := strings.Fields(content)
	return len(words)
}
