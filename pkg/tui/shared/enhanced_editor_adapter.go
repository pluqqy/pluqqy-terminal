package shared

import tea "github.com/charmbracelet/bubbletea"

// EnhancedEditorAdapter wraps the existing EnhancedEditorState to implement EnhancedEditorInterface
// This allows the shared ComponentCreator to work with the existing enhanced editor
type EnhancedEditorAdapter struct {
	editor EnhancedEditor
}

// EnhancedEditor defines the interface for the TUI's enhanced editor state
// This matches the methods available on EnhancedEditorState in the TUI package
type EnhancedEditor interface {
	IsActive() bool
	GetContent() string
	StartEditing(path, name, componentType, content string, tags []string)
	SetContent(content string)
	SetTextareaDimensions(width, height int)
	IsFilePicking() bool
	UpdateFilePicker(msg tea.Msg) tea.Cmd
}

// NewEnhancedEditorAdapter creates a new adapter for the enhanced editor
func NewEnhancedEditorAdapter(editor EnhancedEditor) *EnhancedEditorAdapter {
	return &EnhancedEditorAdapter{
		editor: editor,
	}
}

// IsActive returns whether the enhanced editor is active
func (a *EnhancedEditorAdapter) IsActive() bool {
	if a.editor == nil {
		return false
	}
	return a.editor.IsActive()
}

// GetContent returns the current content of the editor
func (a *EnhancedEditorAdapter) GetContent() string {
	if a.editor == nil {
		return ""
	}
	return a.editor.GetContent()
}

// StartEditing starts editing with the given parameters
func (a *EnhancedEditorAdapter) StartEditing(path, name, componentType, content string, tags []string) {
	if a.editor != nil {
		a.editor.StartEditing(path, name, componentType, content, tags)
	}
}

// SetValue sets the content of the editor
func (a *EnhancedEditorAdapter) SetValue(content string) {
	if a.editor != nil {
		a.editor.SetContent(content)
	}
}

// Focus focuses the editor - this is handled by StartEditing
func (a *EnhancedEditorAdapter) Focus() {
	// Focus is handled by StartEditing method in the enhanced editor
	// No separate focus method needed
}

// SetSize sets the size of the editor
func (a *EnhancedEditorAdapter) SetSize(width, height int) {
	if a.editor != nil {
		a.editor.SetTextareaDimensions(width, height)
	}
}

// IsFilePicking returns whether the file picker is active
func (a *EnhancedEditorAdapter) IsFilePicking() bool {
	if a.editor == nil {
		return false
	}
	return a.editor.IsFilePicking()
}

// UpdateFilePicker updates the file picker with a message
func (a *EnhancedEditorAdapter) UpdateFilePicker(msg interface{}) interface{} {
	if a.editor == nil {
		return nil
	}
	// Convert interface{} back to tea.Msg for the underlying editor
	if teaMsg, ok := msg.(tea.Msg); ok {
		return a.editor.UpdateFilePicker(teaMsg)
	}
	return nil
}

// GetUnderlyingEditor returns the underlying enhanced editor for direct access
// This is needed for render methods that expect the concrete type
func (a *EnhancedEditorAdapter) GetUnderlyingEditor() EnhancedEditor {
	return a.editor
}