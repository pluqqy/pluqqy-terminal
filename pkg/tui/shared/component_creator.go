package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComponentCreator manages the state and logic for component creation
// following the composition pattern established in the codebase
type ComponentCreator struct {
	// State fields
	creatingComponent     bool
	componentCreationType string
	componentName         string
	creationStep          int // 0: type, 1: name, 2: content
	typeCursor            int
	lastSaveSuccessful    bool
	validationError       string

	// Enhanced editor integration
	enhancedEditor EnhancedEditorInterface

	// Callback for reloading components after creation
	reloadCallback func()
}

// EnhancedEditorInterface defines the interface for enhanced editor integration
// This allows the shared creator to work with different enhanced editor implementations
type EnhancedEditorInterface interface {
	IsActive() bool
	GetContent() string
	StartEditing(path, name, componentType, content string, tags []string)
	SetValue(content string) // For setting initial content
	Focus()                  // For focusing the editor
	SetSize(width, height int) // For setting editor size
	
	// Additional methods needed for file picker functionality
	IsFilePicking() bool
	UpdateFilePicker(msg interface{}) interface{}
}

// NewComponentCreator creates a new component creator instance
func NewComponentCreator(reloadCallback func()) *ComponentCreator {
	return &ComponentCreator{
		reloadCallback: reloadCallback,
	}
}

// SetEnhancedEditor sets the enhanced editor implementation
func (c *ComponentCreator) SetEnhancedEditor(editor EnhancedEditorInterface) {
	c.enhancedEditor = editor
}

// Reset resets the component creator state
func (c *ComponentCreator) Reset() {
	c.creatingComponent = false
	c.componentName = ""
	c.creationStep = 0
	c.typeCursor = 0
	c.componentCreationType = ""
	c.lastSaveSuccessful = false
	c.validationError = ""
}

// Start initiates component creation
func (c *ComponentCreator) Start() {
	c.creatingComponent = true
	c.creationStep = 0
	c.typeCursor = 0
	c.componentName = ""
	c.componentCreationType = ""
	c.validationError = ""
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
	if c.enhancedEditor != nil && c.enhancedEditor.IsActive() {
		return c.enhancedEditor.GetContent()
	}
	return ""
}

// GetValidationError returns the current validation error
func (c *ComponentCreator) GetValidationError() string {
	return c.validationError
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
		c.validationError = ""
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
		c.validationError = ""
		return true
	case "enter":
		if c.componentName != "" {
			// Validate component name
			if err := c.validateComponentName(c.componentName); err != nil {
				c.validationError = err.Error()
				return true
			}
			
			// Check for duplicates
			if err := c.checkForDuplicates(c.componentName, c.componentCreationType); err != nil {
				c.validationError = err.Error()
				return true
			}

			c.creationStep = 2
			c.validationError = ""
			// Initialize enhanced editor for content editing
			c.initializeEnhancedEditor()
		}
		return true
	case "backspace":
		if len(c.componentName) > 0 {
			c.componentName = c.componentName[:len(c.componentName)-1]
		}
		c.validationError = ""
		return true
	case " ":
		c.componentName += " "
		c.validationError = ""
		return true
	default:
		if msg.Type == tea.KeyRunes {
			c.componentName += string(msg.Runes)
			c.validationError = ""
			return true
		}
	}
	return false
}

// validateComponentName validates the component name using shared validation logic
func (c *ComponentCreator) validateComponentName(name string) error {
	// Use the shared CLI validator
	if err := cli.ValidateComponentName(name); err != nil {
		return err
	}

	// Additional TUI-specific validations
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("component name cannot be empty or only whitespace")
	}

	// Check for file extension conflicts
	if strings.HasSuffix(strings.ToLower(name), ".md") {
		return fmt.Errorf("component name should not include .md extension")
	}

	return nil
}

// checkForDuplicates checks if a component with the same name already exists
func (c *ComponentCreator) checkForDuplicates(name, componentType string) error {
	// Convert component type to directory name
	var subDir string
	switch componentType {
	case models.ComponentTypeContext:
		subDir = models.ComponentTypeContext
	case models.ComponentTypePrompt:
		subDir = models.ComponentTypePrompt
	case models.ComponentTypeRules:
		subDir = models.ComponentTypeRules
	default:
		return fmt.Errorf("invalid component type: %s", componentType)
	}

	// Generate the expected filename and path
	filename := SanitizeFileName(name) + ".md"
	relativePath := filepath.Join(files.ComponentsDir, subDir, filename)
	fullPath := filepath.Join(files.PluqqyDir, relativePath)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("a %s named '%s' already exists", componentType, name)
	}

	// Also check archived components
	archivedPath := filepath.Join(files.PluqqyDir, files.ArchiveDir, relativePath)
	if _, err := os.Stat(archivedPath); err == nil {
		return fmt.Errorf("an archived %s named '%s' already exists", componentType, name)
	}

	return nil
}

// SanitizeFileName converts a user-provided name into a safe filename
// This matches the implementation in builder.go for consistency
func SanitizeFileName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	filename := strings.ToLower(name)
	filename = strings.ReplaceAll(filename, " ", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	var cleanName strings.Builder
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanName.WriteRune(r)
		}
	}

	result := cleanName.String()

	// Ensure the filename is not empty
	if result == "" {
		result = "untitled"
	}

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	return result
}

// GetStatusMessage returns a status message after successful save
func (c *ComponentCreator) GetStatusMessage() string {
	filename := SanitizeFileName(c.componentName) + ".md"
	return fmt.Sprintf("✓ Created %s: %s", c.componentCreationType, filename)
}

// initializeEnhancedEditor sets up the enhanced editor for component creation
func (c *ComponentCreator) initializeEnhancedEditor() {
	if c.enhancedEditor == nil {
		return
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
	os.MkdirAll(dir, 0755)

	// Generate filename and path
	filename := SanitizeFileName(c.componentName) + ".md"
	relativePath := filepath.Join(files.ComponentsDir, subDir, filename)

	// Start editing with the enhanced editor (this also handles focus)
	c.enhancedEditor.StartEditing(relativePath, c.componentName, c.componentCreationType, "", []string{})
	
	// Set reasonable size for the editor
	c.enhancedEditor.SetSize(80, 20)
}

// WasSaveSuccessful checks if a save just occurred and resets the flag
func (c *ComponentCreator) WasSaveSuccessful() bool {
	if c.lastSaveSuccessful {
		c.lastSaveSuccessful = false
		// Trigger reload callback if available
		if c.reloadCallback != nil {
			c.reloadCallback()
		}
		return true
	}
	return false
}

// MarkSaveSuccessful marks that a save operation was successful
// This should be called by the enhanced editor wrapper when Ctrl+S succeeds
func (c *ComponentCreator) MarkSaveSuccessful() {
	c.lastSaveSuccessful = true
}

// HandleContentEditing handles the content editing phase
// This method provides a standard interface for content editing workflow
func (c *ComponentCreator) HandleContentEditing(msg tea.KeyMsg) (bool, tea.Cmd) {
	// This is typically delegated to the enhanced editor
	// The actual implementation would be provided by the view-specific wrapper
	return false, nil
}

// IsEnhancedEditorActive returns true if the enhanced editor is currently active
func (c *ComponentCreator) IsEnhancedEditorActive() bool {
	return c.enhancedEditor != nil && c.enhancedEditor.IsActive() && c.creationStep == 2
}

// GetEnhancedEditor returns the enhanced editor instance
func (c *ComponentCreator) GetEnhancedEditor() EnhancedEditorInterface {
	return c.enhancedEditor
}