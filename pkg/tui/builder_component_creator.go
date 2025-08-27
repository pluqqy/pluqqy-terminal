package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

// BuilderComponentCreator wraps the shared ComponentCreator and adds TUI-specific methods
type BuilderComponentCreator struct {
	*shared.ComponentCreator
	enhancedEditor *EnhancedEditorState // Direct reference to the enhanced editor
}

// NewBuilderComponentCreator creates a new component creator for the builder view
func NewBuilderComponentCreator(reloadCallback func(), enhancedEditor *EnhancedEditorState) *BuilderComponentCreator {
	creator := &BuilderComponentCreator{
		ComponentCreator: shared.NewComponentCreator(reloadCallback),
		enhancedEditor:   enhancedEditor,
	}

	// Set up the enhanced editor adapter
	if enhancedEditor != nil {
		adapter := shared.NewEnhancedEditorAdapter(enhancedEditor)
		creator.SetEnhancedEditor(adapter)
	}

	return creator
}

// HandleEnhancedEditorInput handles input when enhanced editor is active
// This method provides the TUI-specific functionality that was in the original ComponentCreator
func (c *BuilderComponentCreator) HandleEnhancedEditorInput(msg tea.KeyMsg, width int) (bool, tea.Cmd) {
	if c.enhancedEditor == nil || !c.enhancedEditor.IsActive() {
		return false, nil
	}

	// Let the enhanced editor handle ALL input, including Ctrl+S
	// This matches how editing existing components works
	handled, cmd := HandleEnhancedEditorInput(c.enhancedEditor, msg, width)

	// Check if a save occurred (Ctrl+S was pressed and handled)
	if msg.String() == "ctrl+s" && handled {
		// Mark that a save was successful for the views to reload components
		c.MarkSaveSuccessful()
	}

	// Check if editor was closed (ESC or similar)
	if !c.enhancedEditor.IsActive() {
		// Exit component creation entirely instead of going back to name input
		c.Reset()
		return true, cmd
	}

	return handled, cmd
}