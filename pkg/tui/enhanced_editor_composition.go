// Package tui provides terminal user interface components
package tui

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
)

// EditorCore contains core editor functionality and state.
// This includes the basic editor state, component information, content management,
// and the primary UI components (textarea and file picker)
type EditorCore struct {
	// Basic state
	Active bool
	Mode   EditorMode

	// Component information
	ComponentPath string
	ComponentName string
	ComponentType string
	ComponentTags []string // Store original tags to preserve them

	// Content management
	Content         string
	OriginalContent string

	// UI components
	Textarea   textarea.Model
	FilePicker filepicker.Model

	// Component creation callbacks
	SaveRequested   bool
	CancelRequested bool
	IsNewComponent  bool // True when creating a new component (not yet on disk)
}

// EditorUIState manages UI-specific state and components.
// This includes cursor position tracking, editor statistics,
// and confirmation dialogs
type EditorUIState struct {
	// Cursor and insertion management
	CursorPosition int
	InsertionPoint int

	// Editor stats
	LineCount     int
	WordCount     int
	CurrentLine   int
	CurrentColumn int

	// Exit confirmation
	ExitConfirm       *ConfirmationModel
	ExitConfirmActive bool
}

// EditorOperations manages editor operations and features.
// This includes change tracking, clipboard management, status feedback,
// and undo/redo functionality
type EditorOperations struct {
	// Change tracking
	UnsavedChanges bool

	// Clipboard and status management
	ClipboardStatus *ClipboardStatus
	StatusManager   *StatusManager
	ActionFeedback  *EditorActionFeedback
	PasteHelper     *PasteHelper
	RecentFiles     *RecentFilesTracker

	// Undo management
	UndoStack     []UndoState
	MaxUndoLevels int
}