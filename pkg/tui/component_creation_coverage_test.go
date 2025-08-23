package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// TestComponentCreator_GetterMethods tests all getter methods for coverage
func TestComponentCreator_GetterMethods(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *ComponentCreator
		tests func(t *testing.T, c *ComponentCreator)
	}{
		{
			name: "test all getters at initial state",
			setup: func() *ComponentCreator {
				return NewComponentCreator()
			},
			tests: func(t *testing.T, c *ComponentCreator) {
				// Test GetCurrentStep
				if step := c.GetCurrentStep(); step != 0 {
					t.Errorf("Expected initial step 0, got %d", step)
				}

				// Test GetTypeCursor
				if cursor := c.GetTypeCursor(); cursor != 0 {
					t.Errorf("Expected type cursor 0, got %d", cursor)
				}

				// Test GetComponentType
				if compType := c.GetComponentType(); compType != "" {
					t.Errorf("Expected empty component type, got %s", compType)
				}

				// Test GetComponentName
				if name := c.GetComponentName(); name != "" {
					t.Errorf("Expected empty component name, got %s", name)
				}

				// Test GetComponentContent
				if content := c.GetComponentContent(); content != "" {
					t.Errorf("Expected empty content, got %s", content)
				}

				// Test GetEnhancedEditor
				if editor := c.GetEnhancedEditor(); editor == nil {
					t.Error("Expected enhanced editor to be initialized")
				}
			},
		},
		{
			name: "test getters after setup",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.componentCreationType = models.ComponentTypeContext
				c.componentName = "Test Component"
				c.creationStep = 2
				c.typeCursor = 1
				// Initialize and activate enhanced editor with content
				c.initializeEnhancedEditor()
				c.enhancedEditor.Active = true
				c.enhancedEditor.Content = "Test content"
				c.enhancedEditor.Textarea.SetValue("Test content")
				return c
			},
			tests: func(t *testing.T, c *ComponentCreator) {
				// Test GetCurrentStep
				if step := c.GetCurrentStep(); step != 2 {
					t.Errorf("Expected step 2, got %d", step)
				}

				// Test GetTypeCursor
				if cursor := c.GetTypeCursor(); cursor != 1 {
					t.Errorf("Expected type cursor 1, got %d", cursor)
				}

				// Test GetComponentType
				if compType := c.GetComponentType(); compType != models.ComponentTypeContext {
					t.Errorf("Expected context type, got %s", compType)
				}

				// Test GetComponentName
				if name := c.GetComponentName(); name != "Test Component" {
					t.Errorf("Expected 'Test Component', got %s", name)
				}

				// Test GetComponentContent
				if content := c.GetComponentContent(); content != "Test content" {
					t.Errorf("Expected 'Test content', got %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			tt.tests(t, c)
		})
	}
}

// TestComponentCreator_HandleTypeSelection tests type selection handling
func TestComponentCreator_HandleTypeSelection(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *ComponentCreator
		input          tea.KeyMsg
		expectedType   string
		expectedStep   int
		expectedCursor int
		handled        bool
	}{
		{
			name: "select context type with enter",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 0 // Context
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypeContext,
			expectedStep:   1, // Move to name input
			expectedCursor: 0,
			handled:        true,
		},
		{
			name: "select prompt type with enter",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 1 // Prompt
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypePrompt,
			expectedStep:   1,
			expectedCursor: 1,
			handled:        true,
		},
		{
			name: "select rules type with enter",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 2 // Rules
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypeRules,
			expectedStep:   1,
			expectedCursor: 2,
			handled:        true,
		},
		{
			name: "navigate down with j",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 0
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 1,
			handled:        true,
		},
		{
			name: "navigate up with k",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 1
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0,
			handled:        true,
		},
		{
			name: "stops at last when going down",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 2 // Last position
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyDown},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 2, // Stays at last
			handled:        true,
		},
		{
			name: "stops at first when going up",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				c.typeCursor = 0 // First position
				return c
			},
			input:          tea.KeyMsg{Type: tea.KeyUp},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0, // Stays at first
			handled:        true,
		},
		{
			name: "cancel with escape",
			setup: func() *ComponentCreator {
				c := NewComponentCreator()
				c.Start()
				return c
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
			c := tt.setup()

			// Handle type selection
			handled := c.HandleTypeSelection(tt.input)

			// Check if handled correctly
			if handled != tt.handled {
				t.Errorf("Expected handled=%v, got %v", tt.handled, handled)
			}

			// Check resulting state
			if c.componentCreationType != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, c.componentCreationType)
			}

			if c.creationStep != tt.expectedStep {
				t.Errorf("Expected step %d, got %d", tt.expectedStep, c.creationStep)
			}

			if c.typeCursor != tt.expectedCursor {
				t.Errorf("Expected cursor %d, got %d", tt.expectedCursor, c.typeCursor)
			}
		})
	}
}

// TestComponentCreator_FullFlow tests the complete creation flow
func TestComponentCreator_FullFlow(t *testing.T) {
	t.Run("complete component creation flow", func(t *testing.T) {
		c := NewComponentCreator()

		// Step 1: Start creation
		c.Start()
		if !c.IsActive() {
			t.Error("Creator should be active after start")
		}

		// Step 2: Select type (Context)
		c.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
		if c.GetCurrentStep() != 1 {
			t.Error("Should move to name input step")
		}

		// Step 3: Enter name
		c.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
		c.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		c.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		c.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

		if c.GetComponentName() != "Test" {
			t.Errorf("Expected name 'Test', got '%s'", c.GetComponentName())
		}

		// Move to content step
		c.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
		if c.GetCurrentStep() != 2 {
			t.Error("Should move to content input step")
		}

		// Step 4: Set content in enhanced editor
		c.initializeEnhancedEditor()
		c.enhancedEditor.Content = "Test content for component"
		c.enhancedEditor.Textarea.SetValue("Test content for component")

		// Verify the component state is ready for save
		if c.componentName != "Test" {
			t.Error("Component name should be set")
		}
		if c.GetComponentContent() == "" {
			t.Error("Component content should be set")
		}
		if c.componentCreationType != models.ComponentTypeContext {
			t.Error("Component type should be context")
		}
	})
}

// TestHandleComponentCreation tests the handleComponentCreation function in builder
func TestHandleComponentCreation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PipelineBuilderModel
		input    tea.KeyMsg
		validate func(t *testing.T, m *PipelineBuilderModel)
	}{
		{
			name: "handle type selection in builder",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.componentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyEnter},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.componentCreator.GetCurrentStep() != 1 {
					t.Error("Should advance to name input step")
				}
			},
		},
		{
			name: "handle name input in builder",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.componentCreator.Start()
				m.componentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.componentCreator.GetComponentName() != "A" {
					t.Error("Should add character to name")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Handle component creation input
			updatedModel, _ := m.handleComponentCreation(tt.input)
			updatedM := updatedModel.(*PipelineBuilderModel)

			// Validate
			tt.validate(t, updatedM)
		})
	}
}

// TestMainList_HandleComponentCreation tests component creation in main list
func TestMainList_HandleComponentCreation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MainListModel
		input    tea.KeyMsg
		validate func(t *testing.T, m *MainListModel)
	}{
		{
			name: "start component creation with 'n'",
			setup: func() *MainListModel {
				return NewMainListModel()
			},
			input: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			validate: func(t *testing.T, m *MainListModel) {
				if !m.componentCreator.IsActive() {
					t.Error("Component creator should be active")
				}
			},
		},
		{
			name: "handle type selection in main list",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.componentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, m *MainListModel) {
				if m.componentCreator.GetTypeCursor() != 1 {
					t.Error("Cursor should move down")
				}
			},
		},
		{
			name: "cancel creation with escape",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.componentCreator.Start()
				return m
			},
			input: tea.KeyMsg{Type: tea.KeyEsc},
			validate: func(t *testing.T, m *MainListModel) {
				// After escape in type selection, should reset
				handled := m.componentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEsc})
				if !handled {
					t.Error("Escape should be handled")
				}
				if m.componentCreator.IsActive() {
					t.Error("Component creator should not be active after cancel")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Process input
			updatedModel, _ := m.Update(tt.input)
			updatedM := updatedModel.(*MainListModel)

			// Validate
			tt.validate(t, updatedM)
		})
	}
}
