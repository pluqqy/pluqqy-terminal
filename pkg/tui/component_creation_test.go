package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// TestListComponentCreator_EnhancedEditorIntegration tests the integration between
// the list view's component creator and the enhanced editor
func TestListComponentCreator_EnhancedEditorIntegration(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		componentName string
		expectActive  bool
	}{
		{
			name:          "Context component with enhanced editor in list view",
			componentType: models.ComponentTypeContext,
			componentName: "Test Context",
			expectActive:  true,
		},
		{
			name:          "Prompt component with enhanced editor in list view",
			componentType: models.ComponentTypePrompt,
			componentName: "Test Prompt",
			expectActive:  true,
		},
		{
			name:          "Rules component with enhanced editor in list view",
			componentType: models.ComponentTypeRules,
			componentName: "Test Rules",
			expectActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create list model with component creator
			listModel := NewMainListModel()
			
			// Start creation process
			listModel.operations.ComponentCreator.Start()
			if !listModel.operations.ComponentCreator.IsActive() {
				t.Error("Creator should be active after Start()")
			}

			// Select component type and proceed through steps
			// This tests the view-specific integration
			// First set the correct cursor position for the expected type
			expectedIndex := 0 // context by default
			switch tt.componentType {
			case models.ComponentTypePrompt:
				expectedIndex = 1
			case models.ComponentTypeRules:
				expectedIndex = 2
			}
			
			// Navigate to the correct type
			for i := 0; i < expectedIndex; i++ {
				listModel.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyDown})
			}
			
			// Select the type
			listModel.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
			
			// Add component name
			for _, char := range tt.componentName {
				listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
			}
			
			// Proceed to content step
			listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

			// Check that enhanced editor is active through the list view wrapper
			if !listModel.operations.ComponentCreator.IsEnhancedEditorActive() {
				t.Error("Enhanced editor should be active for content editing")
			}

			// Check that the enhanced editor state is properly configured
			if !listModel.editors.Enhanced.IsActive() {
				t.Error("List view enhanced editor should be active")
			}

			if listModel.editors.Enhanced.ComponentName != tt.componentName {
				t.Errorf("Expected component name %s, got %s",
					tt.componentName, listModel.editors.Enhanced.ComponentName)
			}

			if listModel.editors.Enhanced.ComponentType != tt.componentType {
				t.Errorf("Expected component type %s, got %s",
					tt.componentType, listModel.editors.Enhanced.ComponentType)
			}
		})
	}
}

// TestListComponentCreator_HandleEnhancedEditorInput tests the list view's handling
// of enhanced editor input during component creation
func TestListComponentCreator_HandleEnhancedEditorInput(t *testing.T) {
	// Create list model
	listModel := NewMainListModel()

	// Setup component creation to content step
	listModel.operations.ComponentCreator.Start()
	listModel.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Context
	for _, char := range "Test Save" {
		listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify enhanced editor is active
	if !listModel.operations.ComponentCreator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}

	// Set some content
	listModel.editors.Enhanced.Textarea.SetValue("Test content for saving")

	// Simulate Ctrl+S through the list view wrapper
	saveKey := tea.KeyMsg{Type: tea.KeyCtrlS}
	handled, _ := listModel.operations.ComponentCreator.HandleEnhancedEditorInput(saveKey, 80)

	if !handled {
		t.Error("Save key should be handled by list view wrapper")
	}

	// Verify that the save was marked successful
	if !listModel.operations.ComponentCreator.WasSaveSuccessful() {
		t.Error("Save should be marked as successful through list view wrapper")
	}

	// Note: Actual file saving would fail in tests without proper setup
	// This test verifies the view-specific integration and handling
}

// TestListComponentCreator_EnhancedEditorCancel tests cancelling component creation
// from the enhanced editor step through the list view
func TestListComponentCreator_EnhancedEditorCancel(t *testing.T) {
	// Create list model
	listModel := NewMainListModel()

	// Setup component creation to content step
	listModel.operations.ComponentCreator.Start()
	listModel.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Context
	for _, char := range "Test Cancel" {
		listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify enhanced editor is active
	if !listModel.operations.ComponentCreator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}

	// Set some content
	listModel.editors.Enhanced.Textarea.SetValue("Test content")

	// Verify the editor is active before cancelling
	if !listModel.editors.Enhanced.IsActive() {
		t.Fatal("List enhanced editor should be active before cancel")
	}

	// Simulate ESC to cancel through the list view wrapper
	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	handled, _ := listModel.operations.ComponentCreator.HandleEnhancedEditorInput(escKey, 80)

	if !handled {
		t.Error("ESC key should be handled by list view wrapper")
	}

	// Component creation should be reset entirely through the list view logic
	if listModel.operations.ComponentCreator.IsActive() {
		t.Error("Component creator should not be active after ESC in list view")
	}

	// Editor should have exited in the list view
	if listModel.editors.Enhanced.IsActive() {
		t.Error("List enhanced editor should not be active after ESC")
	}
}

// TestListComponentCreator_SaveKeepsEditorOpen tests that saving a new component
// keeps the editor open in the list view (allowing continued editing)
func TestListComponentCreator_SaveKeepsEditorOpen(t *testing.T) {
	// Create list model
	listModel := NewMainListModel()

	// Setup component creation to content step
	listModel.operations.ComponentCreator.Start()
	listModel.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Context
	for _, char := range "Test Component" {
		listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	listModel.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	// Set some content
	testContent := "Test content for save"
	listModel.editors.Enhanced.Textarea.SetValue(testContent)

	// Manually simulate a successful save through the list view state
	// This simulates what happens when Ctrl+S is pressed and save succeeds
	listModel.editors.Enhanced.OriginalContent = testContent
	listModel.editors.Enhanced.Content = testContent
	listModel.editors.Enhanced.UnsavedChanges = false
	listModel.editors.Enhanced.IsNewComponent = false
	listModel.operations.ComponentCreator.MarkSaveSuccessful()

	// Editor should still be active after save in the list view
	if !listModel.editors.Enhanced.IsActive() {
		t.Error("List enhanced editor should remain active after save")
	}

	// Check that save was marked as successful through the list view wrapper
	if !listModel.operations.ComponentCreator.WasSaveSuccessful() {
		t.Error("Save should have been marked as successful in list view")
	}

	// Calling WasSaveSuccessful again should return false (flag is reset)
	if listModel.operations.ComponentCreator.WasSaveSuccessful() {
		t.Error("WasSaveSuccessful should reset flag after being called in list view")
	}

	// Enhanced editor in list view should have no unsaved changes
	if listModel.editors.Enhanced.UnsavedChanges {
		t.Error("List editor should have no unsaved changes after save")
	}

	// IsNewComponent should be false after save in list view
	if listModel.editors.Enhanced.IsNewComponent {
		t.Error("IsNewComponent should be false after save in list view")
	}

	// Verify that Content and OriginalContent match (no unsaved changes)
	if listModel.editors.Enhanced.Content != listModel.editors.Enhanced.OriginalContent {
		t.Error("Content and OriginalContent should match after save in list view")
	}
}

// TestBuilderComponentCreator_ExternalEditorPath tests that component paths
// are correctly set for external editor integration in the builder view
func TestBuilderComponentCreator_ExternalEditorPath(t *testing.T) {
	// Create builder model
	builderModel := NewPipelineBuilderModel()

	// Setup component creation to content step
	builderModel.editors.ComponentCreator.Start()
	builderModel.editors.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Context
	for _, char := range "Test External Editor" {
		builderModel.editors.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	builderModel.editors.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify enhanced editor is active through builder view
	if !builderModel.editors.ComponentCreator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active in builder view")
	}

	// Check that component path is set in builder view's enhanced editor
	if builderModel.editors.Enhanced.ComponentPath == "" {
		t.Error("Component path should be set for external editor in builder view")
	}

	// Verify path contains correct component type directory
	expectedPath := "components/contexts/test-external-editor.md"
	if builderModel.editors.Enhanced.ComponentPath != expectedPath {
		t.Errorf("Expected path %s, got %s in builder view", expectedPath, builderModel.editors.Enhanced.ComponentPath)
	}

	// Note: IsNewComponent flag is set separately during the save workflow
	// This test focuses on the integration between the shared ComponentCreator 
	// and the builder view's enhanced editor
}
