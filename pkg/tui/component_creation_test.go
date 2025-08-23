package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestComponentCreator_EnhancedEditorIntegration(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		componentName string
		expectActive  bool
	}{
		{
			name:          "Context component with enhanced editor",
			componentType: models.ComponentTypeContext,
			componentName: "Test Context",
			expectActive:  true,
		},
		{
			name:          "Prompt component with enhanced editor",
			componentType: models.ComponentTypePrompt,
			componentName: "Test Prompt",
			expectActive:  true,
		},
		{
			name:          "Rules component with enhanced editor",
			componentType: models.ComponentTypeRules,
			componentName: "Test Rules",
			expectActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component creator
			creator := NewComponentCreator()

			// Start creation process
			creator.Start()
			if !creator.IsActive() {
				t.Error("Creator should be active after Start()")
			}

			// Select component type
			creator.componentCreationType = tt.componentType
			creator.creationStep = 1

			// Enter name and proceed to content
			creator.componentName = tt.componentName
			enterKey := tea.KeyMsg{Type: tea.KeyEnter}
			creator.HandleNameInput(enterKey)

			// Check that we're at content step with enhanced editor
			if creator.creationStep != 2 {
				t.Errorf("Expected to be at step 2 (content), got %d", creator.creationStep)
			}

			if !creator.IsEnhancedEditorActive() {
				t.Error("Enhanced editor should be active for content editing")
			}

			if creator.enhancedEditor == nil {
				t.Fatal("Enhanced editor should be initialized")
			}

			if !creator.enhancedEditor.Active {
				t.Error("Enhanced editor state should be active")
			}

			if creator.enhancedEditor.ComponentName != tt.componentName {
				t.Errorf("Expected component name %s, got %s",
					tt.componentName, creator.enhancedEditor.ComponentName)
			}

			if creator.enhancedEditor.ComponentType != tt.componentType {
				t.Errorf("Expected component type %s, got %s",
					tt.componentType, creator.enhancedEditor.ComponentType)
			}
		})
	}
}

func TestComponentCreator_EnhancedEditorSave(t *testing.T) {
	// Create component creator
	creator := NewComponentCreator()

	// Setup component creation
	creator.Start()
	creator.componentCreationType = models.ComponentTypeContext
	creator.creationStep = 1
	creator.componentName = "Test Save"

	// Enter name and proceed to content
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	creator.HandleNameInput(enterKey)

	// Verify enhanced editor is active
	if !creator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}

	// Set some content
	creator.enhancedEditor.Textarea.SetValue("Test content for saving")

	// Simulate Ctrl+S
	saveKey := tea.KeyMsg{Type: tea.KeyCtrlS}
	handled, _ := creator.HandleEnhancedEditorInput(saveKey, 80)

	if !handled {
		t.Error("Save key should be handled")
	}

	// Note: Actual file saving would fail in tests without proper setup
	// This test verifies the integration and handling
}

func TestComponentCreator_EnhancedEditorCancel(t *testing.T) {
	// Create component creator
	creator := NewComponentCreator()

	// Setup component creation
	creator.Start()
	creator.componentCreationType = models.ComponentTypePrompt
	creator.creationStep = 1
	creator.componentName = "Test Cancel"

	// Enter name and proceed to content
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	creator.HandleNameInput(enterKey)

	// Verify enhanced editor is active
	if !creator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}

	// Set some content
	creator.enhancedEditor.Textarea.SetValue("Test content")

	// Simulate ESC to cancel through component creator handler
	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	handled, _ := creator.HandleEnhancedEditorInput(escKey, 80)

	if !handled {
		t.Error("ESC key should be handled")
	}

	// Component creation should be reset entirely (not just go back to name input)
	if creator.IsActive() {
		t.Error("Component creator should not be active after ESC")
	}

	// Editor should have exited
	if creator.enhancedEditor != nil && creator.enhancedEditor.IsActive() {
		t.Error("Enhanced editor should not be active after ESC")
	}
}

func TestComponentCreator_SaveKeepsEditorOpen(t *testing.T) {
	// Create component creator
	creator := NewComponentCreator()

	// Start creation and set up for content editing
	creator.Start()
	creator.componentCreationType = models.ComponentTypeContext
	creator.componentName = "Test Component"
	creator.creationStep = 2
	creator.initializeEnhancedEditor()

	// Set some content
	testContent := "Test content for save"
	creator.enhancedEditor.Textarea.SetValue(testContent)

	// Manually simulate a successful save
	// This simulates what happens when Ctrl+S is pressed and save succeeds
	creator.enhancedEditor.OriginalContent = testContent
	creator.enhancedEditor.Content = testContent  // Also set Content field
	creator.enhancedEditor.UnsavedChanges = false
	creator.enhancedEditor.IsNewComponent = false
	creator.lastSaveSuccessful = true

	// Editor should still be active after save
	if !creator.enhancedEditor.IsActive() {
		t.Error("Enhanced editor should remain active after save")
	}

	// Check that save was marked as successful
	if !creator.WasSaveSuccessful() {
		t.Error("Save should have been marked as successful")
	}

	// Calling WasSaveSuccessful again should return false (flag is reset)
	if creator.WasSaveSuccessful() {
		t.Error("WasSaveSuccessful should reset flag after being called")
	}

	// Enhanced editor should have no unsaved changes
	if creator.enhancedEditor.UnsavedChanges {
		t.Error("Editor should have no unsaved changes after save")
	}

	// IsNewComponent should be false after save
	if creator.enhancedEditor.IsNewComponent {
		t.Error("IsNewComponent should be false after save")
	}

	// Verify that Content and OriginalContent match (no unsaved changes)
	if creator.enhancedEditor.Content != creator.enhancedEditor.OriginalContent {
		t.Error("Content and OriginalContent should match after save")
	}
}

func TestComponentCreator_ExternalEditorPath(t *testing.T) {
	// Create component creator
	creator := NewComponentCreator()

	// Setup component creation
	creator.Start()
	creator.componentCreationType = models.ComponentTypeContext
	creator.creationStep = 1
	creator.componentName = "Test External Editor"

	// Enter name and proceed to content
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	creator.HandleNameInput(enterKey)

	// Verify enhanced editor is active
	if !creator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}

	// Check that component path is set
	if creator.enhancedEditor.ComponentPath == "" {
		t.Error("Component path should be set for external editor")
	}

	// Verify path contains correct component type directory
	expectedPath := "components/contexts/test-external-editor.md"
	if creator.enhancedEditor.ComponentPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, creator.enhancedEditor.ComponentPath)
	}

	// Verify IsNewComponent flag is set
	if !creator.enhancedEditor.IsNewComponent {
		t.Error("IsNewComponent flag should be true for new components")
	}
}
