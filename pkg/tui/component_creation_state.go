package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// ComponentCreator manages the state and logic for component creation
type ComponentCreator struct {
	// State fields
	creatingComponent     bool
	componentCreationType string
	componentName         string
	creationStep          int // 0: type, 1: name, 2: content
	typeCursor            int
	lastSaveSuccessful    bool // Track if last save was successful

	// Enhanced editor integration
	enhancedEditor *EnhancedEditorState
}

// NewComponentCreator creates a new component creator instance
func NewComponentCreator() *ComponentCreator {
	return &ComponentCreator{
		enhancedEditor: NewEnhancedEditorState(),
	}
}

// Reset resets the component creator state
func (c *ComponentCreator) Reset() {
	c.creatingComponent = false
	c.componentName = ""
	c.creationStep = 0
	c.typeCursor = 0
	c.componentCreationType = ""
	c.lastSaveSuccessful = false
}

// Start initiates component creation
func (c *ComponentCreator) Start() {
	c.creatingComponent = true
	c.creationStep = 0
	c.typeCursor = 0
	c.componentName = ""
	c.componentCreationType = ""
}

// IsActive returns whether component creation is active
func (c *ComponentCreator) IsActive() bool {
	return c.creatingComponent
}

// GetCurrentStep returns the current creation step
func (c *ComponentCreator) GetCurrentStep() int {
	return c.creationStep
}

// GetTypeCursor returns the current type cursor position
func (c *ComponentCreator) GetTypeCursor() int {
	return c.typeCursor
}

// GetComponentType returns the selected component type
func (c *ComponentCreator) GetComponentType() string {
	return c.componentCreationType
}

// GetComponentName returns the component name
func (c *ComponentCreator) GetComponentName() string {
	return c.componentName
}

// GetComponentContent returns the component content from the enhanced editor
func (c *ComponentCreator) GetComponentContent() string {
	if c.enhancedEditor != nil && c.enhancedEditor.Active {
		return c.enhancedEditor.GetContent()
	}
	return ""
}

// HandleTypeSelection handles keyboard input for type selection step
func (c *ComponentCreator) HandleTypeSelection(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "esc":
		c.creatingComponent = false
		return true
	case "up", "k":
		if c.typeCursor > 0 {
			c.typeCursor--
		}
		return true
	case "down", "j":
		if c.typeCursor < 2 {
			c.typeCursor++
		}
		return true
	case "enter":
		types := []string{models.ComponentTypeContext, models.ComponentTypePrompt, models.ComponentTypeRules}
		c.componentCreationType = types[c.typeCursor]
		c.creationStep = 1
		return true
	}
	return false
}

// HandleNameInput handles keyboard input for name input step
func (c *ComponentCreator) HandleNameInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "esc":
		c.creationStep = 0
		c.componentName = ""
		return true
	case "enter":
		if c.componentName != "" {
			c.creationStep = 2
			// Initialize enhanced editor for content editing
			c.initializeEnhancedEditor()
		}
		return true
	case "backspace":
		if len(c.componentName) > 0 {
			c.componentName = c.componentName[:len(c.componentName)-1]
		}
		return true
	case " ":
		c.componentName += " "
		return true
	default:
		if msg.Type == tea.KeyRunes {
			c.componentName += string(msg.Runes)
			return true
		}
	}
	return false
}


// GetStatusMessage returns a status message after successful save
func (c *ComponentCreator) GetStatusMessage() string {
	filename := sanitizeFileName(c.componentName) + ".md"
	return fmt.Sprintf("✓ Created %s: %s", c.componentCreationType, filename)
}

// initializeEnhancedEditor sets up the enhanced editor for component creation
func (c *ComponentCreator) initializeEnhancedEditor() {
	if c.enhancedEditor == nil {
		c.enhancedEditor = NewEnhancedEditorState()
	}

	// Generate the component path that will be used when saving
	var subDir string
	switch c.componentCreationType {
	case models.ComponentTypeContext:
		subDir = models.ComponentTypeContext
	case models.ComponentTypePrompt:
		subDir = models.ComponentTypePrompt
	case models.ComponentTypeRules:
		subDir = models.ComponentTypeRules
	}

	// Ensure directory exists for new components
	dir := filepath.Join(files.PluqqyDir, files.ComponentsDir, subDir)
	os.MkdirAll(dir, 0755) // Create directory if it doesn't exist

	// Generate filename and path
	filename := sanitizeFileName(c.componentName) + ".md"
	relativePath := filepath.Join(files.ComponentsDir, subDir, filename)

	// Check if we're returning to the editor (it was already active before)
	// If so, just reactivate it without clearing content
	if c.enhancedEditor.ComponentPath == relativePath {
		// Returning to editor - just reactivate without losing content
		c.enhancedEditor.Active = true
		c.enhancedEditor.IsNewComponent = true  // Ensure this stays true for new components
		// Restore the textarea with the preserved content
		c.enhancedEditor.Textarea.SetValue(c.enhancedEditor.Content)
		c.enhancedEditor.Textarea.Focus()
		return
	}

	// First time entering editor for this component - initialize everything
	c.enhancedEditor.Active = true
	c.enhancedEditor.Mode = EditorModeNormal
	c.enhancedEditor.ComponentPath = relativePath // Set the path for external editor
	c.enhancedEditor.ComponentName = c.componentName
	c.enhancedEditor.ComponentType = c.componentCreationType
	c.enhancedEditor.Content = ""
	c.enhancedEditor.OriginalContent = ""
	c.enhancedEditor.UnsavedChanges = false

	// Mark this as a new component (not yet saved to disk)
	c.enhancedEditor.IsNewComponent = true

	// Clear and setup the textarea
	c.enhancedEditor.Textarea.SetValue("")
	c.enhancedEditor.Textarea.Focus()

	// Set reasonable size for the textarea
	c.enhancedEditor.Textarea.SetWidth(80)
	c.enhancedEditor.Textarea.SetHeight(20)
}

// HandleEnhancedEditorInput handles input when enhanced editor is active
func (c *ComponentCreator) HandleEnhancedEditorInput(msg tea.KeyMsg, width int) (bool, tea.Cmd) {
	if c.enhancedEditor == nil || !c.enhancedEditor.Active {
		return false, nil
	}

	// Let the enhanced editor handle ALL input, including Ctrl+S
	// This matches how editing existing components works
	handled, cmd := HandleEnhancedEditorInput(c.enhancedEditor, msg, width)

	// Check if a save occurred (Ctrl+S was pressed and handled)
	if msg.String() == "ctrl+s" && handled {
		// Mark that a save was successful for the views to reload components
		c.lastSaveSuccessful = true
	}

	// Check if editor was closed (ESC or similar)
	if !c.enhancedEditor.IsActive() {
		// Exit component creation entirely instead of going back to name input
		c.Reset()
		return true, cmd
	}

	return handled, cmd
}

// IsEnhancedEditorActive returns true if the enhanced editor is currently active
func (c *ComponentCreator) IsEnhancedEditorActive() bool {
	return c.enhancedEditor != nil && c.enhancedEditor.Active && c.creationStep == 2
}

// GetEnhancedEditor returns the enhanced editor instance
func (c *ComponentCreator) GetEnhancedEditor() *EnhancedEditorState {
	return c.enhancedEditor
}

// WasSaveSuccessful checks if a save just occurred and resets the flag
func (c *ComponentCreator) WasSaveSuccessful() bool {
	if c.lastSaveSuccessful {
		c.lastSaveSuccessful = false // Reset flag after checking
		return true
	}
	return false
}
