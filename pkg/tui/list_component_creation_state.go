package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
}

// NewComponentCreator creates a new component creator instance
func NewComponentCreator() *ComponentCreator {
	return &ComponentCreator{}
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
		}
		return true
	case "backspace":
		if len(c.componentName) > 0 {
			c.componentName = c.componentName[:len(c.componentName)-1]
		}
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
	// Ensure directory exists
	var dir string
	switch c.componentCreationType {
	case models.ComponentTypeContext:
		dir = ".pluqqy/components/contexts"
	case models.ComponentTypePrompt:
		dir = ".pluqqy/components/prompts"
	case models.ComponentTypeRules:
		dir = ".pluqqy/components/rules"
	default:
		return fmt.Errorf("unknown component type: %s", c.componentCreationType)
	}
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Generate filename
	filename := sanitizeFileName(c.componentName) + ".md"
	fullPath := filepath.Join(dir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("component already exists: %s", filename)
	}
	
	// Write the file
	content := c.componentContent
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// GetStatusMessage returns a status message after successful save
func (c *ComponentCreator) GetStatusMessage() string {
	filename := sanitizeFileName(c.componentName) + ".md"
	return fmt.Sprintf("✓ Created %s: %s", c.componentCreationType, filename)
}