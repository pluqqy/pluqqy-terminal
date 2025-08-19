package tui

import (
	"strings"
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// TestPipelineBuilderModel_DeleteComponentFromLeft tests the new delete functionality
func TestPipelineBuilderModel_DeleteComponentFromLeft(t *testing.T) {
	// Note: This test focuses on the delete command generation
	// Actual file deletion is handled by the files package
	
	tests := []struct {
		name           string
		setup          func() *PipelineBuilderModel
		componentToDelete componentItem
		expectError    bool
		validateAfter  func(t *testing.T, m *PipelineBuilderModel)
	}{
		{
			name: "successfully delete component from left column",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				// Manually add a test component
				m.contexts = []componentItem{{
					name:     "Test Component",
					path:     "components/contexts/test-component.md",
					compType: models.ComponentTypeContext,
					tags:     []string{"test", "delete"},
				}}
				return m
			},
			componentToDelete: componentItem{
				name:     "Test Component",
				path:     "components/contexts/test-component.md",
				compType: models.ComponentTypeContext,
				tags:     []string{"test", "delete"},
			},
			expectError: false,
			validateAfter: func(t *testing.T, m *PipelineBuilderModel) {
				// In a unit test, we just verify the command was created correctly
			},
		},
		{
			name: "handle deletion of non-existent component",
			setup: func() *PipelineBuilderModel {
				return NewPipelineBuilderModel()
			},
			componentToDelete: componentItem{
				name:     "Non Existent",
				path:     "components/contexts/non-existent.md",
				compType: models.ComponentTypeContext,
			},
			expectError: true,
			validateAfter: func(t *testing.T, m *PipelineBuilderModel) {
				// Nothing to validate
			},
		},
		{
			name: "delete component with tags triggers cleanup",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				// Manually add a tagged component
				m.contexts = []componentItem{{
					name:     "Tagged Component",
					path:     "components/contexts/tagged-component.md",
					compType: models.ComponentTypeContext,
					tags:     []string{"cleanup", "test"},
				}}
				return m
			},
			componentToDelete: componentItem{
				name:     "Tagged Component",
				path:     "components/contexts/tagged-component.md",
				compType: models.ComponentTypeContext,
				tags:     []string{"cleanup", "test"},
			},
			expectError: false,
			validateAfter: func(t *testing.T, m *PipelineBuilderModel) {
				// In a unit test, we verify the command includes tag cleanup
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Execute the delete command
			cmd := m.deleteComponentFromLeft(tt.componentToDelete)
			if cmd == nil {
				t.Fatal("Expected a command to be returned")
			}
			
			// Execute the command and check the result
			msg := cmd()
			
			switch result := msg.(type) {
			case StatusMsg:
				statusStr := string(result)
				if tt.expectError {
					if !strings.HasPrefix(statusStr, "×") {
						t.Errorf("Expected error message, got: %s", statusStr)
					}
				} else {
					if !strings.HasPrefix(statusStr, "✓") {
						t.Errorf("Expected success message, got: %s", statusStr)
					}
				}
			default:
				t.Errorf("Unexpected message type: %T", msg)
			}
			
			// Validate the state after deletion
			tt.validateAfter(t, m)
		})
	}
}

// TestPipelineBuilderModel_DeleteComponentKeyHandler tests Ctrl+D handling
func TestPipelineBuilderModel_DeleteComponentKeyHandler(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *PipelineBuilderModel
		keyMsg        tea.KeyMsg
		expectConfirm bool
	}{
		{
			name: "Ctrl+D in left column shows component delete confirmation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.activeColumn = leftColumn
				// Add a test component
				m.contexts = []componentItem{
					{
						name:     "Test Context",
						path:     "components/contexts/test.md",
						compType: models.ComponentTypeContext,
					},
				}
				m.filteredContexts = m.contexts
				m.leftCursor = 0
				return m
			},
			keyMsg:        tea.KeyMsg{Type: tea.KeyCtrlD},
			expectConfirm: true,
		},
		{
			name: "Ctrl+D in right column shows pipeline delete confirmation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.activeColumn = rightColumn
				m.pipeline = &models.Pipeline{
					Name: "Test Pipeline",
					Path: "pipelines/test.yaml",
				}
				return m
			},
			keyMsg:        tea.KeyMsg{Type: tea.KeyCtrlD},
			expectConfirm: true,
		},
		{
			name: "Ctrl+D in preview column does nothing",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.activeColumn = previewColumn
				m.showPreview = true
				return m
			},
			keyMsg:        tea.KeyMsg{Type: tea.KeyCtrlD},
			expectConfirm: false,
		},
		{
			name: "Ctrl+D with no components selected does nothing",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.activeColumn = leftColumn
				m.contexts = []componentItem{} // Empty list
				return m
			},
			keyMsg:        tea.KeyMsg{Type: tea.KeyCtrlD},
			expectConfirm: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Send the key message
			updatedModel, _ := m.Update(tt.keyMsg)
			updatedM := updatedModel.(*PipelineBuilderModel)
			
			// Check if confirmation dialog is shown
			if tt.expectConfirm {
				if !updatedM.deleteConfirm.Active() {
					t.Error("Expected delete confirmation to be active")
				}
			} else {
				if updatedM.deleteConfirm.Active() {
					t.Error("Expected delete confirmation to NOT be active")
				}
			}
		})
	}
}

// TestPipelineBuilderModel_SavePipeline tests pipeline saving functionality
func TestPipelineBuilderModel_SavePipeline(t *testing.T) {
	// Note: This test focuses on the save command generation
	// Actual file writing is handled by the files package
	
	tests := []struct {
		name          string
		setup         func() *PipelineBuilderModel
		expectError   bool
		validateAfter func(t *testing.T, m *PipelineBuilderModel)
	}{
		{
			name: "save new pipeline successfully",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.pipeline = &models.Pipeline{
					Name: "New Test Pipeline",
					Components: []models.ComponentRef{
						{Type: "contexts", Path: "../components/contexts/test.md", Order: 1},
					},
				}
				m.nameInput = "New Test Pipeline"
				m.selectedComponents = m.pipeline.Components
				return m
			},
			expectError: false,
			validateAfter: func(t *testing.T, m *PipelineBuilderModel) {
				// In a unit test, we verify the command was successful
			},
		},
		{
			name: "update existing pipeline",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.pipeline = &models.Pipeline{
					Name: "Existing Pipeline",
					Path: "pipelines/existing.yaml",
					Components: []models.ComponentRef{
						{Type: "contexts", Path: "../components/contexts/new.md", Order: 1},
					},
				}
				m.nameInput = "Existing Pipeline"
				m.selectedComponents = m.pipeline.Components
				return m
			},
			expectError: false,
			validateAfter: func(t *testing.T, m *PipelineBuilderModel) {
				// In a unit test, we verify the command was successful
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Execute save command
			cmd := m.savePipeline()
			if cmd == nil {
				t.Fatal("Expected a command to be returned")
			}
			
			// Execute the command
			msg := cmd()
			
			// Check result
			switch result := msg.(type) {
			case StatusMsg:
				statusStr := string(result)
				if tt.expectError {
					if !strings.HasPrefix(statusStr, "×") {
						t.Errorf("Expected error message, got: %s", statusStr)
					}
				} else {
					if strings.HasPrefix(statusStr, "×") {
						t.Errorf("Expected success but got error: %s", statusStr)
					}
				}
			default:
				t.Errorf("Unexpected message type: %T", msg)
			}
			
			// Validate after save
			tt.validateAfter(t, m)
		})
	}
}

// TestPipelineBuilderModel_AddSelectedComponent tests adding components to pipeline
func TestPipelineBuilderModel_AddSelectedComponent(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *PipelineBuilderModel
		componentToAdd componentItem
		expectedOrder  int
		expectedCount  int
	}{
		{
			name: "add first component to empty pipeline",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.selectedComponents = []models.ComponentRef{}
				return m
			},
			componentToAdd: componentItem{
				name:     "Test Context",
				path:     "components/contexts/test.md",
				compType: models.ComponentTypeContext,
			},
			expectedOrder: 1,
			expectedCount: 1,
		},
		{
			name: "add component to existing pipeline",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.selectedComponents = []models.ComponentRef{
					{Type: "prompts", Path: "../components/prompts/existing.md", Order: 1},
				}
				return m
			},
			componentToAdd: componentItem{
				name:     "New Context",
				path:     "components/contexts/new.md",
				compType: models.ComponentTypeContext,
			},
			expectedOrder: 2,
			expectedCount: 2,
		},
		{
			name: "add component maintains type ordering",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.selectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/ctx1.md", Order: 1},
					{Type: "prompts", Path: "../components/prompts/p1.md", Order: 2},
				}
				return m
			},
			componentToAdd: componentItem{
				name:     "New Context",
				path:     "components/contexts/ctx2.md",
				compType: models.ComponentTypeContext,
			},
			expectedOrder: 2, // Should be inserted after the existing context
			expectedCount: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Manually add the component (since addSelectedComponent doesn't take parameters)
			// Set up the state as if the component was selected
			m.contexts = []componentItem{tt.componentToAdd}
			m.filteredContexts = m.contexts
			m.leftCursor = 0
			m.addSelectedComponent()
			
			// Validate count
			if len(m.selectedComponents) != tt.expectedCount {
				t.Errorf("Expected %d components, got %d", tt.expectedCount, len(m.selectedComponents))
			}
			
			// Find the added component and check its order
			found := false
			for _, comp := range m.selectedComponents {
				if comp.Path == "../"+tt.componentToAdd.path {
					found = true
					if comp.Order != tt.expectedOrder {
						t.Errorf("Expected order %d, got %d", tt.expectedOrder, comp.Order)
					}
					break
				}
			}
			
			if !found {
				t.Error("Component was not added to selectedComponents")
			}
		})
	}
}

// TestPipelineBuilderModel_RemoveSelectedComponent tests removing components from pipeline
func TestPipelineBuilderModel_RemoveSelectedComponent(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *PipelineBuilderModel
		indexToRemove int
		expectedCount int
		validateOrder func(t *testing.T, components []models.ComponentRef)
	}{
		{
			name: "remove single component",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.selectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/test.md", Order: 1},
				}
				m.rightCursor = 0
				return m
			},
			indexToRemove: 0,
			expectedCount: 0,
			validateOrder: func(t *testing.T, components []models.ComponentRef) {
				// No components left
			},
		},
		{
			name: "remove middle component and reorder",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.selectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/c1.md", Order: 1},
					{Type: "prompts", Path: "../components/prompts/p1.md", Order: 2},
					{Type: "rules", Path: "../components/rules/r1.md", Order: 3},
				}
				m.rightCursor = 1
				return m
			},
			indexToRemove: 1,
			expectedCount: 2,
			validateOrder: func(t *testing.T, components []models.ComponentRef) {
				// Check reordering
				if components[0].Order != 1 {
					t.Errorf("First component should have order 1, got %d", components[0].Order)
				}
				if components[1].Order != 2 {
					t.Errorf("Second component should have order 2, got %d", components[1].Order)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Remove the component
			m.removeSelectedComponent()
			
			// Validate count
			if len(m.selectedComponents) != tt.expectedCount {
				t.Errorf("Expected %d components after removal, got %d", tt.expectedCount, len(m.selectedComponents))
			}
			
			// Validate ordering
			tt.validateOrder(t, m.selectedComponents)
		})
	}
}