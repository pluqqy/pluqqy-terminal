package tui

import (
	"os"
	"strings"
	"testing"

	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestPipelineBuilderModel_DeleteComponentFromLeft tests the delete command creation
func TestPipelineBuilderModel_DeleteComponentFromLeft(t *testing.T) {
	// This test verifies that the deleteComponentFromLeft method creates a command
	// We can't test actual file deletion without setting up real files

	tests := []struct {
		name              string
		setup             func() *PipelineBuilderModel
		componentToDelete componentItem
		expectCommand     bool
	}{
		{
			name: "creates delete command for component",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.data.Contexts = []componentItem{{
					name:     "Test Component",
					path:     "components/contexts/test-component.md",
					compType: models.ComponentTypeContext,
					tags:     []string{"test"},
				}}
				return m
			},
			componentToDelete: componentItem{
				name:     "Test Component",
				path:     "components/contexts/test-component.md",
				compType: models.ComponentTypeContext,
				tags:     []string{"test"},
			},
			expectCommand: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Check that delete command is created
			cmd := m.deleteComponentFromLeft(tt.componentToDelete)

			if tt.expectCommand && cmd == nil {
				t.Fatal("Expected a command to be returned")
			}
			if !tt.expectCommand && cmd != nil {
				t.Fatal("Expected no command to be returned")
			}
		})
	}
}

// TestPipelineBuilderModel_DeleteComponentKeyHandler tests that deleteConfirm is initialized
func TestPipelineBuilderModel_DeleteComponentKeyHandler(t *testing.T) {
	// This test just verifies that the deleteConfirm field is properly initialized
	// Actual key handling would require more complex setup

	m := NewPipelineBuilderModel()

	if m.ui.DeleteConfirm == nil {
		t.Fatal("deleteConfirm should be initialized")
	}

	// Verify it's not active initially
	if m.ui.DeleteConfirm.Active() {
		t.Error("deleteConfirm should not be active initially")
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
				m.data.Pipeline = &models.Pipeline{
					Name: "New Test Pipeline",
					Components: []models.ComponentRef{
						{Type: "contexts", Path: "../components/contexts/test.md", Order: 1},
					},
				}
				m.editors.NameInput = "New Test Pipeline"
				m.data.SelectedComponents = m.data.Pipeline.Components
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
				m.data.Pipeline = &models.Pipeline{
					Name: "Existing Pipeline",
					Path: "pipelines/existing.yaml",
					Components: []models.ComponentRef{
						{Type: "contexts", Path: "../components/contexts/new.md", Order: 1},
					},
				}
				m.editors.NameInput = "Existing Pipeline"
				m.data.SelectedComponents = m.data.Pipeline.Components
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
			// Clean up any existing test pipelines first
			os.RemoveAll(".pluqqy/pipelines/New Test Pipeline.yaml")
			os.RemoveAll(".pluqqy/pipelines/Existing Pipeline.yaml")
			defer os.RemoveAll(".pluqqy/pipelines/New Test Pipeline.yaml")
			defer os.RemoveAll(".pluqqy/pipelines/Existing Pipeline.yaml")

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
		expectedCount  int
	}{
		{
			name: "add first component to empty pipeline",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.data.SelectedComponents = []models.ComponentRef{}
				return m
			},
			componentToAdd: componentItem{
				name:     "Test Context",
				path:     "components/contexts/test.md",
				compType: models.ComponentTypeContext,
			},
			expectedCount: 1,
		},
		{
			name: "add component to existing pipeline",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.data.SelectedComponents = []models.ComponentRef{
					{Type: "prompts", Path: "../components/prompts/existing.md", Order: 1},
				}
				return m
			},
			componentToAdd: componentItem{
				name:     "New Context",
				path:     "components/contexts/new.md",
				compType: models.ComponentTypeContext,
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Set up the state as if the component was selected
			m.data.Contexts = []componentItem{tt.componentToAdd}
			m.data.FilteredContexts = m.data.Contexts
			m.ui.LeftCursor = 0
			m.addSelectedComponent()

			// Validate count
			if len(m.data.SelectedComponents) != tt.expectedCount {
				t.Errorf("Expected %d components, got %d", tt.expectedCount, len(m.data.SelectedComponents))
			}

			// Check that component was added
			found := false
			for _, comp := range m.data.SelectedComponents {
				if comp.Path == "../"+tt.componentToAdd.path {
					found = true
					// Order is managed by reorganizeComponentsByType, so we don't test specific values
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
				m.data.SelectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/test.md", Order: 1},
				}
				m.ui.RightCursor = 0
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
				m.data.SelectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/c1.md", Order: 1},
					{Type: "prompts", Path: "../components/prompts/p1.md", Order: 2},
					{Type: "rules", Path: "../components/rules/r1.md", Order: 3},
				}
				m.ui.RightCursor = 1
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
			if len(m.data.SelectedComponents) != tt.expectedCount {
				t.Errorf("Expected %d components after removal, got %d", tt.expectedCount, len(m.data.SelectedComponents))
			}

			// Validate ordering
			tt.validateOrder(t, m.data.SelectedComponents)
		})
	}
}
