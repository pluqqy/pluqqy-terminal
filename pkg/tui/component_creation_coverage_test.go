package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestListView_ComponentCreatorIntegration tests the integration between
// the list view and the shared ComponentCreator
func TestListView_ComponentCreatorIntegration(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *MainListModel
		tests func(t *testing.T, m *MainListModel)
	}{
		{
			name: "list view component creator initial state",
			setup: func() *MainListModel {
				return NewMainListModel()
			},
			tests: func(t *testing.T, m *MainListModel) {
				// Test that the list view has a component creator
				if m.operations.ComponentCreator == nil {
					t.Error("List view should have a component creator")
				}

				// Test initial state through list view
				if m.operations.ComponentCreator.IsActive() {
					t.Error("Component creator should not be active initially in list view")
				}

				if m.operations.ComponentCreator.GetCurrentStep() != 0 {
					t.Errorf("Expected initial step 0 in list view, got %d", m.operations.ComponentCreator.GetCurrentStep())
				}
			},
		},
		{
			name: "list view component creator after activation",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.operations.ComponentCreator.Start()
				m.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Context
				for _, char := range "Test Component" {
					m.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
				}
				m.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
				return m
			},
			tests: func(t *testing.T, m *MainListModel) {
				// Test state after setup through list view
				if m.operations.ComponentCreator.GetCurrentStep() != 2 {
					t.Errorf("Expected step 2 in list view, got %d", m.operations.ComponentCreator.GetCurrentStep())
				}

				if m.operations.ComponentCreator.GetComponentType() != models.ComponentTypeContext {
					t.Errorf("Expected context type in list view, got %s", m.operations.ComponentCreator.GetComponentType())
				}

				if m.operations.ComponentCreator.GetComponentName() != "Test Component" {
					t.Errorf("Expected 'Test Component' in list view, got %s", m.operations.ComponentCreator.GetComponentName())
				}

				// Test that enhanced editor is active in list view
				if !m.editors.Enhanced.IsActive() {
					t.Error("List view enhanced editor should be active after component creation setup")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			tt.tests(t, m)
		})
	}
}

// TestBuilderView_ComponentCreatorIntegration tests the integration between
// the builder view and the shared ComponentCreator
func TestBuilderView_ComponentCreatorIntegration(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *PipelineBuilderModel
		input          tea.KeyMsg
		expectedType   string
		expectedStep   int
		expectedCursor int
		handled        bool
	}{
		{
			name: "builder view handles type selection",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				return m
			},
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypeContext,
			expectedStep:   1, // Move to name input
			expectedCursor: 0,
			handled:        true,
		},
		{
			name: "builder view handles navigation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				return m
			},
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 1,
			handled:        true,
		},
		{
			name: "builder view handles cancellation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				return m
			},
			input:          tea.KeyMsg{Type: tea.KeyEsc},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0,
			handled:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Handle type selection through builder view
			handled := m.editors.ComponentCreator.HandleTypeSelection(tt.input)

			// Check if handled correctly
			if handled != tt.handled {
				t.Errorf("Expected handled=%v, got %v in builder view", tt.handled, handled)
			}

			// Check resulting state in builder view
			if m.editors.ComponentCreator.GetComponentType() != tt.expectedType {
				t.Errorf("Expected type %s, got %s in builder view", tt.expectedType, m.editors.ComponentCreator.GetComponentType())
			}

			if m.editors.ComponentCreator.GetCurrentStep() != tt.expectedStep {
				t.Errorf("Expected step %d, got %d in builder view", tt.expectedStep, m.editors.ComponentCreator.GetCurrentStep())
			}

			if m.editors.ComponentCreator.GetTypeCursor() != tt.expectedCursor {
				t.Errorf("Expected cursor %d, got %d in builder view", tt.expectedCursor, m.editors.ComponentCreator.GetTypeCursor())
			}
		})
	}
}

// TestViewIntegration_CompleteComponentCreationFlow tests the complete creation flow
// through both list and builder views to ensure consistent behavior
func TestViewIntegration_CompleteComponentCreationFlow(t *testing.T) {
	tests := []struct {
		name string
		setupView func() interface{}
		getCreator func(view interface{}) interface {
			Start()
			IsActive() bool
			HandleTypeSelection(tea.KeyMsg) bool
			HandleNameInput(tea.KeyMsg) bool
			GetCurrentStep() int
			GetComponentName() string
			GetComponentType() string
			IsEnhancedEditorActive() bool
		}
		getEnhancedEditor func(view interface{}) interface{}
	}{
		{
			name: "complete flow through list view",
			setupView: func() interface{} {
				return NewMainListModel()
			},
			getCreator: func(view interface{}) interface {
				Start()
				IsActive() bool
				HandleTypeSelection(tea.KeyMsg) bool
				HandleNameInput(tea.KeyMsg) bool
				GetCurrentStep() int
				GetComponentName() string
				GetComponentType() string
				IsEnhancedEditorActive() bool
			} {
				m := view.(*MainListModel)
				return m.operations.ComponentCreator
			},
			getEnhancedEditor: func(view interface{}) interface{} {
				m := view.(*MainListModel)
				return m.editors.Enhanced
			},
		},
		{
			name: "complete flow through builder view",
			setupView: func() interface{} {
				return NewPipelineBuilderModel()
			},
			getCreator: func(view interface{}) interface {
				Start()
				IsActive() bool
				HandleTypeSelection(tea.KeyMsg) bool
				HandleNameInput(tea.KeyMsg) bool
				GetCurrentStep() int
				GetComponentName() string
				GetComponentType() string
				IsEnhancedEditorActive() bool
			} {
				m := view.(*PipelineBuilderModel)
				return m.editors.ComponentCreator
			},
			getEnhancedEditor: func(view interface{}) interface{} {
				m := view.(*PipelineBuilderModel)
				return m.editors.Enhanced
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.setupView()
			creator := tt.getCreator(view)

			// Step 1: Start creation
			creator.Start()
			if !creator.IsActive() {
				t.Error("Creator should be active after start")
			}

			// Step 2: Select type (Context)
			creator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
			if creator.GetCurrentStep() != 1 {
				t.Error("Should move to name input step")
			}

			// Step 3: Enter name
			for _, char := range "Test" {
				creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
			}

			if creator.GetComponentName() != "Test" {
				t.Errorf("Expected name 'Test', got '%s'", creator.GetComponentName())
			}

			// Move to content step
			creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
			if creator.GetCurrentStep() != 2 {
				t.Error("Should move to content input step")
			}

			// Step 4: Verify enhanced editor is active
			if !creator.IsEnhancedEditorActive() {
				t.Error("Enhanced editor should be active")
			}

			// Verify the component state is ready for save
			if creator.GetComponentName() != "Test" {
				t.Error("Component name should be set")
			}
			if creator.GetComponentType() != models.ComponentTypeContext {
				t.Error("Component type should be context")
			}
		})
	}
}

// TestBuilderView_HandleComponentCreation tests the handleComponentCreation function in builder
// which coordinates between the view and the shared ComponentCreator
func TestBuilderView_HandleComponentCreation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PipelineBuilderModel
		input    tea.KeyMsg
		validate func(t *testing.T, m *PipelineBuilderModel)
	}{
		{
			name: "builder view handles type selection through handleComponentCreation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyEnter},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.ComponentCreator.GetCurrentStep() != 1 {
					t.Error("Builder view should advance to name input step")
				}
			},
		},
		{
			name: "builder view handles name input through handleComponentCreation",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				m.editors.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.ComponentCreator.GetComponentName() != "A" {
					t.Error("Builder view should add character to name")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Handle component creation input through builder view's method
			updatedModel, _ := m.handleComponentCreation(tt.input)
			updatedM := updatedModel.(*PipelineBuilderModel)

			// Validate builder view state
			tt.validate(t, updatedM)
		})
	}
}

// TestListView_HandleComponentCreation tests component creation in main list
// through the view's Update method and integration with shared ComponentCreator
func TestListView_HandleComponentCreation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MainListModel
		input    tea.KeyMsg
		validate func(t *testing.T, m *MainListModel)
	}{
		{
			name: "list view starts component creation with 'n'",
			setup: func() *MainListModel {
				return NewMainListModel()
			},
			input: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			validate: func(t *testing.T, m *MainListModel) {
				if !m.operations.ComponentCreator.IsActive() {
					t.Error("List view component creator should be active after 'n'")
				}
			},
		},
		{
			name: "list view handles type selection navigation",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.operations.ComponentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, m *MainListModel) {
				// Simulate the navigation - this would be handled by handleComponentCreation
				// which delegates to the shared ComponentCreator
				m.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyDown})
				if m.operations.ComponentCreator.GetTypeCursor() != 1 {
					t.Error("List view cursor should move down through shared creator")
				}
			},
		},
		{
			name: "list view handles creation cancellation",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.operations.ComponentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyEsc},
			validate: func(t *testing.T, m *MainListModel) {
				// Simulate escape handling through list view
				handled := m.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEsc})
				if !handled {
					t.Error("Escape should be handled by list view component creator")
				}
				if m.operations.ComponentCreator.IsActive() {
					t.Error("List view component creator should not be active after cancel")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// For the 'n' key test, process through Update which triggers component creation start
			if tt.input.String() == "n" {
				updatedModel, _ := m.Update(tt.input)
				updatedM := updatedModel.(*MainListModel)
				tt.validate(t, updatedM)
			} else {
				// For other tests, validate the setup directly
				tt.validate(t, m)
			}
		})
	}
}
