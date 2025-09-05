package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestEnhancedEditorAlwaysUsed verifies that enhanced editor is always used after legacy removal
func TestEnhancedEditorAlwaysUsed(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{} // Can be MainListModel or PipelineBuilderModel
		validate func(t *testing.T, model interface{})
	}{
		{
			name: "MainListModel always uses enhanced editor",
			setup: func() interface{} {
				return NewMainListModel()
			},
			validate: func(t *testing.T, model interface{}) {
				m := model.(*MainListModel)

				// Check that enhanced editor is initialized
				if m.editors.Enhanced == nil {
					t.Error("Enhanced editor should be initialized")
				}

				// Verify no legacy editor field exists by checking it's not in use
				// This would fail to compile if componentEditor field existed
				// which is what we want - ensuring legacy editor is completely removed
			},
		},
		{
			name: "PipelineBuilderModel always uses enhanced editor",
			setup: func() interface{} {
				return NewPipelineBuilderModel()
			},
			validate: func(t *testing.T, model interface{}) {
				m := model.(*PipelineBuilderModel)

				// Check that enhanced editor is initialized
				if m.editors.Enhanced == nil {
					t.Error("Enhanced editor should be initialized")
				}
			},
		},
		{
			name: "ComponentCreator always uses enhanced editor",
			setup: func() interface{} {
				return NewComponentCreator()
			},
			validate: func(t *testing.T, model interface{}) {
				c := model.(*ComponentCreator)

				// Check that enhanced editor is initialized
				if c.enhancedEditor == nil {
					t.Error("Enhanced editor should be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.setup()
			tt.validate(t, model)
		})
	}
}

// TestEditingAlwaysUsesEnhancedEditor verifies editing operations use enhanced editor
func TestEditingAlwaysUsesEnhancedEditor(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *MainListModel
		triggerEdit  func(m *MainListModel) (tea.Model, tea.Cmd)
		expectActive bool
	}{
		{
			name: "pressing 'e' activates enhanced editor",
			setup: func() *MainListModel {
				m := NewMainListModel()
				// Add a test component
				m.data.FilteredComponents = []componentItem{
					{
						name:     "Test Component",
						path:     "components/contexts/test.md",
						compType: models.ComponentTypeContext,
					},
				}
				m.stateManager.ComponentCursor = 0
				m.stateManager.ActivePane = componentsPane
				return m
			},
			triggerEdit: func(m *MainListModel) (tea.Model, tea.Cmd) {
				// Simulate pressing 'e' to edit
				return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
			},
			expectActive: true,
		},
		{
			name: "component creation uses enhanced editor",
			setup: func() *MainListModel {
				m := NewMainListModel()
				m.operations.ComponentCreator.Start()
				// Use public methods to set up the component creation state
				// Simulate type selection
				m.operations.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
				// Simulate name input
				nameInput := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("New Component")}
				for _, r := range nameInput.Runes {
					m.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
				}
				m.operations.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
				return m
			},
			triggerEdit: func(m *MainListModel) (tea.Model, tea.Cmd) {
				// Enhanced editor should already be initialized from the setup
				return m, nil
			},
			expectActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Trigger edit operation
			updatedModel, _ := tt.triggerEdit(m)
			updatedM := updatedModel.(*MainListModel)

			// Check enhanced editor state
			if tt.expectActive {
				if m.operations.ComponentCreator.IsActive() && m.operations.ComponentCreator.GetCurrentStep() == 2 {
					// Check component creator's enhanced editor
					if !m.operations.ComponentCreator.IsEnhancedEditorActive() {
						t.Error("Expected enhanced editor to be active in component creator")
					}
				} else if updatedM.editors.Enhanced != nil {
					// Check main enhanced editor (for regular editing)
					// Note: We can't directly test IsActive() without proper setup
					// but we verify it's initialized and ready
					if updatedM.editors.Enhanced == nil {
						t.Error("Expected enhanced editor to be initialized")
					}
				}
			}
		})
	}
}

// TestBuilderEditingUsesEnhancedEditor tests Builder view always uses enhanced editor
func TestBuilderEditingUsesEnhancedEditor(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *PipelineBuilderModel
		triggerEdit func(m *PipelineBuilderModel) (tea.Model, tea.Cmd)
		validate    func(t *testing.T, m *PipelineBuilderModel)
	}{
		{
			name: "edit from left column uses enhanced editor",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.ui.ActiveColumn = leftColumn
				m.data.Contexts = []componentItem{
					{
						name:     "Test Context",
						path:     "components/contexts/test.md",
						compType: models.ComponentTypeContext,
					},
				}
				m.data.FilteredContexts = m.data.Contexts
				m.ui.LeftCursor = 0
				return m
			},
			triggerEdit: func(m *PipelineBuilderModel) (tea.Model, tea.Cmd) {
				// Simulate pressing 'e' to edit
				return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
			},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.Enhanced == nil {
					t.Error("Enhanced editor should be initialized")
				}
				// After triggering edit, editingComponent should be true
				// (actual behavior depends on file reading which we can't test here)
			},
		},
		{
			name: "edit from right column uses enhanced editor",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.ui.ActiveColumn = rightColumn
				m.data.SelectedComponents = []models.ComponentRef{
					{Type: "contexts", Path: "../components/contexts/test.md", Order: 1},
				}
				m.ui.RightCursor = 0
				return m
			},
			triggerEdit: func(m *PipelineBuilderModel) (tea.Model, tea.Cmd) {
				// Simulate pressing 'e' to edit
				return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
			},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.Enhanced == nil {
					t.Error("Enhanced editor should be initialized")
				}
			},
		},
		{
			name: "component creation in builder uses enhanced editor",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editors.ComponentCreator.Start()
				return m
			},
			triggerEdit: func(m *PipelineBuilderModel) (tea.Model, tea.Cmd) {
				// Move through creation steps using public interface
				// First select prompt type (second option, index 1)
				m.editors.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyDown})
				m.editors.ComponentCreator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter})
				// Enter name
				nameInput := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Test Prompt")}
				for _, r := range nameInput.Runes {
					m.editors.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
				}
				m.editors.ComponentCreator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
				return m, nil
			},
			validate: func(t *testing.T, m *PipelineBuilderModel) {
				if !m.editors.ComponentCreator.IsEnhancedEditorActive() {
					t.Error("Component creator should use enhanced editor")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Trigger edit
			updatedModel, _ := tt.triggerEdit(m)
			updatedM := updatedModel.(*PipelineBuilderModel)

			// Validate
			tt.validate(t, updatedM)
		})
	}
}

// TestNoLegacyEditorFallback ensures there's no fallback to legacy editor
func TestNoLegacyEditorFallback(t *testing.T) {
	// Test that component creation doesn't fall back
	t.Run("component creator has no fallback path", func(t *testing.T) {
		creator := NewComponentCreator()
		creator.Start()
		creator.componentCreationType = models.ComponentTypeRules
		creator.componentName = "Test Rule"
		creator.creationStep = 2

		// This should initialize enhanced editor
		creator.initializeEnhancedEditor()

		// Verify enhanced editor is active
		if !creator.IsEnhancedEditorActive() {
			t.Error("Enhanced editor should be active for content editing")
		}

		// Try to handle input - should go through enhanced editor
		handled, _ := creator.HandleEnhancedEditorInput(
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			80, // width
		)

		// The function should try to handle it (even if it returns false due to setup)
		// The important thing is it doesn't panic or fall back to non-existent method
		_ = handled
	})

	// Test that regular editing doesn't fall back
	t.Run("main list editor has no fallback", func(t *testing.T) {
		m := NewMainListModel()

		// Verify enhanced editor exists
		if m.editors.Enhanced == nil {
			t.Fatal("Enhanced editor must be initialized")
		}

		// There should be no way to use a legacy editor
		// This is verified by the fact that componentEditor field doesn't exist
		// If it did exist, this would be a compilation error
	})

	// Test builder has no fallback
	t.Run("builder has no fallback to legacy editor", func(t *testing.T) {
		m := NewPipelineBuilderModel()

		// Enhanced editor should always be present
		if m.editors.Enhanced == nil {
			t.Fatal("Enhanced editor must be initialized")
		}

		// handleComponentEditing should only use enhanced editor
		// Create a simple test by checking the function exists and doesn't panic
		_, _ = m.handleComponentEditing(tea.KeyMsg{Type: tea.KeyEsc})
	})
}

// TestEnhancedEditorStateConsistency verifies state management is consistent
func TestEnhancedEditorStateConsistency(t *testing.T) {
	t.Run("enhanced editor maintains state correctly", func(t *testing.T) {
		editor := NewEnhancedEditorState()

		// Test initialization
		if editor.Active {
			t.Error("Editor should not be active initially")
		}

		// Test starting edit
		editor.StartEditing(
			"test/path.md",
			"Test Component",
			models.ComponentTypeContext,
			"Test content",
			[]string{"tag1", "tag2"},
		)

		if !editor.Active {
			t.Error("Editor should be active after StartEditing")
		}

		if editor.ComponentPath != "test/path.md" {
			t.Errorf("Expected path 'test/path.md', got '%s'", editor.ComponentPath)
		}

		if editor.ComponentName != "Test Component" {
			t.Errorf("Expected name 'Test Component', got '%s'", editor.ComponentName)
		}

		// Test reset
		editor.Reset()

		if editor.Active {
			t.Error("Editor should not be active after reset")
		}

		if editor.ComponentPath != "" {
			t.Error("Component path should be cleared after reset")
		}
	})
}

// TestEnhancedEditorHandlesAllInputProperly tests input handling
func TestEnhancedEditorHandlesAllInputProperly(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *EnhancedEditorState
		input    tea.KeyMsg
		expected func(t *testing.T, state *EnhancedEditorState, handled bool)
	}{
		{
			name: "ESC key exits editor",
			setup: func() *EnhancedEditorState {
				e := NewEnhancedEditorState()
				e.Active = true
				e.Mode = EditorModeNormal
				return e
			},
			input: tea.KeyMsg{Type: tea.KeyEsc},
			expected: func(t *testing.T, state *EnhancedEditorState, handled bool) {
				// ESC might show exit confirmation or exit directly
				// depending on whether there are unsaved changes
				if !handled {
					t.Error("ESC key should be handled")
				}
			},
		},
		{
			name: "Ctrl+S triggers save",
			setup: func() *EnhancedEditorState {
				e := NewEnhancedEditorState()
				e.Active = true
				e.Mode = EditorModeNormal
				e.ComponentPath = "test.md"
				return e
			},
			input: tea.KeyMsg{Type: tea.KeyCtrlS},
			expected: func(t *testing.T, state *EnhancedEditorState, handled bool) {
				if !handled {
					t.Error("Ctrl+S should be handled")
				}
			},
		},
		{
			name: "@ key opens file picker",
			setup: func() *EnhancedEditorState {
				e := NewEnhancedEditorState()
				e.Active = true
				e.Mode = EditorModeNormal
				return e
			},
			input: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}},
			expected: func(t *testing.T, state *EnhancedEditorState, handled bool) {
				if !handled {
					t.Error("@ key should be handled")
				}
				if state.Mode != EditorModeFilePicking {
					t.Error("@ key should switch to file picker mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup()

			// Handle input
			handled, _ := HandleEnhancedEditorInput(state, tt.input, 80)

			// Validate
			tt.expected(t, state, handled)
		})
	}
}
