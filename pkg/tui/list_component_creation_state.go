package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComponentCreator manages the state and logic for component creation
type ComponentCreator struct {
	// State fields
	creatingComponent     bool
	componentCreationType string
	componentName         string
	componentContent      string
	creationStep          int // 0: type, 1: name, 2: content
	typeCursor           int
	
	// Enhanced editor integration
	enhancedEditor        *EnhancedEditorState
	useEnhancedEditor     bool
}

// NewComponentCreator creates a new component creator instance
func NewComponentCreator() *ComponentCreator {
	return &ComponentCreator{
		enhancedEditor:    NewEnhancedEditorState(),
		useEnhancedEditor: true, // Enable enhanced editor by default
	}
}

// Reset resets the component creator state
func (c *ComponentCreator) Reset() {
	c.creatingComponent = false
	c.componentName = ""
	c.componentContent = ""
	c.creationStep = 0
	c.typeCursor = 0
	c.componentCreationType = ""
}

// Start initiates component creation
func (c *ComponentCreator) Start() {
	c.creatingComponent = true
	c.creationStep = 0
	c.typeCursor = 0
	c.componentName = ""
	c.componentContent = ""
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

// GetComponentContent returns the component content
func (c *ComponentCreator) GetComponentContent() string {
	return c.componentContent
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
			if c.useEnhancedEditor {
				c.initializeEnhancedEditor()
			}
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

// HandleContentEdit handles keyboard input for content editing step
func (c *ComponentCreator) HandleContentEdit(msg tea.KeyMsg) (bool, error) {
	switch msg.String() {
	case "esc":
		c.creationStep = 1
		c.componentContent = ""
		return true, nil
	case "ctrl+s":
		err := c.SaveComponent()
		if err != nil {
			return true, err
		}
		c.Reset()
		return true, nil
	case "backspace":
		if len(c.componentContent) > 0 {
			c.componentContent = c.componentContent[:len(c.componentContent)-1]
		}
		return true, nil
	case "enter":
		c.componentContent += "\n"
		return true, nil
	case "tab":
		c.componentContent += "    "
		return true, nil
	case " ":
		c.componentContent += " "
		return true, nil
	default:
		if msg.Type == tea.KeyRunes {
			c.componentContent += string(msg.Runes)
			return true, nil
		}
	}
	return false, nil
}

// SaveComponent saves the component to disk
func (c *ComponentCreator) SaveComponent() error {
	// Determine the component subdirectory
	var subDir string
	switch c.componentCreationType {
	case models.ComponentTypeContext:
		subDir = models.ComponentTypeContext
	case models.ComponentTypePrompt:
		subDir = models.ComponentTypePrompt
	case models.ComponentTypeRules:
		subDir = models.ComponentTypeRules
	default:
		return fmt.Errorf("unknown component type: %s", c.componentCreationType)
	}
	
	// Ensure directory exists
	dir := filepath.Join(files.PluqqyDir, files.ComponentsDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Generate filename
	filename := sanitizeFileName(c.componentName) + ".md"
	relativePath := filepath.Join(files.ComponentsDir, subDir, filename)
	fullPath := filepath.Join(files.PluqqyDir, relativePath)
	
	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("component already exists: %s", filename)
	}
	
	// Prepare content
	content := c.componentContent
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	
	// Use the files package to write with name in frontmatter
	// The display name is the component name provided by the user
	err := files.WriteComponentWithNameAndTags(relativePath, content, c.componentName, nil)
	if err != nil {
		return fmt.Errorf("failed to write component: %w", err)
	}
	
	return nil
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
	
	// Generate filename and path
	filename := sanitizeFileName(c.componentName) + ".md"
	relativePath := filepath.Join(files.ComponentsDir, subDir, filename)
	
	// Configure enhanced editor for creation mode
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
	if !c.useEnhancedEditor || c.enhancedEditor == nil || !c.enhancedEditor.Active {
		return false, nil
	}
	
	// Let the enhanced editor handle the input
	handled, cmd := HandleEnhancedEditorInput(c.enhancedEditor, msg, width)
	
	// Check if save was requested (Ctrl+S pressed)
	if msg.String() == "ctrl+s" && c.enhancedEditor.Active {
		c.componentContent = c.enhancedEditor.GetContent()
		if err := c.SaveComponent(); err == nil {
			c.enhancedEditor.Active = false
			c.Reset()
			return true, cmd
		}
	}
	
	// Check if editor was closed (ESC or similar)
	if !c.enhancedEditor.IsActive() {
		c.creationStep = 1 // Go back to name input
		c.componentContent = ""
		return true, cmd
	}
	
	return handled, cmd
}

// IsEnhancedEditorActive returns true if the enhanced editor is currently active
func (c *ComponentCreator) IsEnhancedEditorActive() bool {
	return c.useEnhancedEditor && c.enhancedEditor != nil && c.enhancedEditor.Active && c.creationStep == 2
}

// GetEnhancedEditor returns the enhanced editor instance
func (c *ComponentCreator) GetEnhancedEditor() *EnhancedEditorState {
	return c.enhancedEditor
}