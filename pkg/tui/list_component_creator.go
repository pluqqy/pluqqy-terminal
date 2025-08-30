package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

// ListComponentCreator wraps the shared ComponentCreator and adds TUI-specific methods
type ListComponentCreator struct {
	*shared.ComponentCreator
	enhancedEditor *EnhancedEditorState // Direct reference to the enhanced editor
}

// NewListComponentCreator creates a new component creator for the list view
func NewListComponentCreator(reloadCallback func(), enhancedEditor *EnhancedEditorState) *ListComponentCreator {
	creator := &ListComponentCreator{
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

// HandleNameInput overrides the shared method to ensure IsNewComponent is set
func (c *ListComponentCreator) HandleNameInput(msg tea.KeyMsg) bool {
	// Call the parent implementation
	handled := c.ComponentCreator.HandleNameInput(msg)
	
	// If we're moving to the content step (step 2), ensure IsNewComponent is set
	if handled && c.GetCurrentStep() == 2 && c.enhancedEditor != nil {
		c.enhancedEditor.IsNewComponent = true
	}
	
	return handled
}

// HandleEnhancedEditorInput handles input when enhanced editor is active
// This method provides the TUI-specific functionality that was in the original ComponentCreator
func (c *ListComponentCreator) HandleEnhancedEditorInput(msg tea.KeyMsg, width int) (bool, tea.Cmd) {
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
		// For new components, go back to naming step
		if c.enhancedEditor.IsNewComponent {
			// Go back to naming step (step 1)
			c.SetCreationStep(1)
			// Keep the component name and type so user doesn't have to re-enter
			// The enhanced editor state is preserved so content isn't lost
			return true, cmd
		}
		// For existing components being edited (shouldn't happen in creation flow)
		c.Reset()
		return true, cmd
	}

	return handled, cmd
}