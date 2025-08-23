package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestPipelineBuilderModel_RenameInitialization(t *testing.T) {
	m := NewPipelineBuilderModel()

	if m.editors.Rename.State == nil {
		t.Error("renameState should be initialized")
	}

	if m.editors.Rename.Renderer == nil {
		t.Error("renameRenderer should be initialized")
	}

	if m.editors.Rename.Operator == nil {
		t.Error("renameOperator should be initialized")
	}

	// Verify rename is not active initially
	if m.editors.Rename.State.IsActive() {
		t.Error("rename should not be active initially")
	}
}

func TestPipelineBuilderModel_RenameKeyHandler(t *testing.T) {
	tests := []struct {
		name               string
		setup              func() *PipelineBuilderModel
		activeColumn       column
		expectRenameActive bool
	}{
		{
			name: "R key in left column starts component rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Exit name editing mode
				m.editors.EditingName = false
				m.editors.NameInput = "Test Pipeline"
				m.data.Pipeline.Name = "Test Pipeline"
				// Add test components
				m.data.Prompts = []componentItem{
					{name: "Test Prompt", path: "components/prompts/test.md"},
				}
				m.data.FilteredPrompts = m.data.Prompts
				m.ui.LeftCursor = 0
				return m
			},
			activeColumn:       leftColumn,
			expectRenameActive: true,
		},
		{
			name: "R key in right column starts pipeline rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Exit name editing mode
				m.editors.EditingName = false
				m.editors.NameInput = "Test Pipeline"
				m.data.Pipeline = &models.Pipeline{
					Name: "Test Pipeline",
					Path: "test-pipeline.yaml",
				}
				return m
			},
			activeColumn:       rightColumn,
			expectRenameActive: true,
		},
		{
			name: "R key with no pipeline in right column does nothing",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = nil
				return m
			},
			activeColumn:       rightColumn,
			expectRenameActive: false,
		},
		{
			name: "R key with empty components list does nothing",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Prompts = []componentItem{}
				m.data.Contexts = []componentItem{}
				m.data.Rules = []componentItem{}
				m.data.FilteredPrompts = m.data.Prompts
				m.data.FilteredContexts = m.data.Contexts
				m.data.FilteredRules = m.data.Rules
				return m
			},
			activeColumn:       leftColumn,
			expectRenameActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			m.ui.ActiveColumn = tt.activeColumn

			// Send 'R' key
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
			newModel, _ := m.Update(msg)
			updatedModel := newModel.(*PipelineBuilderModel)

			if updatedModel.editors.Rename.State.IsActive() != tt.expectRenameActive {
				t.Errorf("rename active = %v, want %v",
					updatedModel.editors.Rename.State.IsActive(), tt.expectRenameActive)
			}
		})
	}
}

func TestPipelineBuilderModel_RenameStateHandling(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *PipelineBuilderModel
		input       tea.KeyMsg
		wantHandled bool
	}{
		{
			name: "escape cancels rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.editors.Rename.State.Start("Test", "component", "test.md", false)
				return m
			},
			input:       tea.KeyMsg{Type: tea.KeyEsc},
			wantHandled: true,
		},
		{
			name: "character input handled during rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.editors.Rename.State.Start("Test", "component", "test.md", false)
				return m
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: true,
		},
		{
			name: "normal keys ignored when rename inactive",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Don't activate rename
				return m
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Update with input
			_, cmd := m.Update(tt.input)

			// Check if the input was handled (cmd returned or state changed)
			if tt.wantHandled {
				// When rename handles input, it should either return a command
				// or update the rename state
				if !m.editors.Rename.State.IsActive() && cmd == nil {
					// If rename became inactive without a command, that's ok (escape)
				} else if m.editors.Rename.State.IsActive() && m.editors.Rename.State.NewName == "" && tt.input.Type == tea.KeyRunes {
					t.Error("Character input not processed during rename")
				}
			}
		})
	}
}

func TestPipelineBuilderModel_RenameMessages(t *testing.T) {
	tests := []struct {
		name       string
		msg        tea.Msg
		setup      func() *PipelineBuilderModel
		checkState func(*testing.T, *PipelineBuilderModel)
	}{
		{
			name: "RenameSuccessMsg resets state",
			msg: RenameSuccessMsg{
				ItemType: "component",
				OldName:  "old.md",
				NewName:  "New Component",
			},
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.editors.Rename.State.Start("Old", "component", "old.md", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.Rename.State.IsActive() {
					t.Error("rename should be inactive after success")
				}
			},
		},
		{
			name: "RenameSuccessMsg for pipeline updates pipeline",
			msg: RenameSuccessMsg{
				ItemType: "pipeline",
				OldName:  "old-pipeline.yaml",
				NewName:  "New Pipeline",
			},
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "Old Pipeline",
					Path: "old-pipeline.yaml",
				}
				m.editors.Rename.State.Start("Old Pipeline", "pipeline", "old-pipeline.yaml", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.data.Pipeline.Name != "New Pipeline" {
					t.Errorf("pipeline name = %q, want %q", m.data.Pipeline.Name, "New Pipeline")
				}
				if m.editors.NameInput != "New Pipeline" {
					t.Errorf("nameInput = %q, want %q", m.editors.NameInput, "New Pipeline")
				}
			},
		},
		{
			name: "RenameErrorMsg sets validation error",
			msg: RenameErrorMsg{
				Error: makeTestError("rename failed"),
			},
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.editors.Rename.State.Start("Test", "component", "test.md", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editors.Rename.State.ValidationError == "" {
					t.Error("validation error should be set")
				}
				if !contains(m.editors.Rename.State.ValidationError, "rename failed") {
					t.Errorf("validation error = %q, want to contain %q",
						m.editors.Rename.State.ValidationError, "rename failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Process the message
			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(*PipelineBuilderModel)

			// Check the resulting state
			tt.checkState(t, updatedModel)
		})
	}
}

func TestPipelineBuilderModel_RenameView(t *testing.T) {
	tests := []struct {
		name             string
		setup            func() *PipelineBuilderModel
		wantRenameDialog bool
		wantContains     []string
	}{
		{
			name: "shows rename dialog when active",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Exit name editing mode so the main view is shown
				m.editors.EditingName = false
				m.editors.Rename.State.Start("Test Component", "component", "test.md", false)
				m.editors.Rename.Renderer.SetSize(100, 30)
				return m
			},
			wantRenameDialog: true,
			wantContains: []string{
				"RENAME COMPONENT",
				"Test Component",
			},
		},
		{
			name: "shows normal view when inactive",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Don't activate rename
				return m
			},
			wantRenameDialog: false,
			wantContains: []string{
				"CREATE NEW PIPELINE", // Shows new pipeline creation view
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			view := m.View()

			for _, want := range tt.wantContains {
				if !contains(view, want) {
					t.Errorf("view missing %q\nGot:\n%s", want, view)
				}
			}
		})
	}
}

func TestPipelineBuilderModel_RenameWindowResize(t *testing.T) {
	m := makeTestBuilderModel()

	// Send window resize message
	newWidth := 120
	newHeight := 40
	msg := tea.WindowSizeMsg{Width: newWidth, Height: newHeight}

	newModel, _ := m.Update(msg)
	updatedModel := newModel.(*PipelineBuilderModel)

	// Check that rename renderer was updated
	if updatedModel.editors.Rename.Renderer.Width != newWidth {
		t.Errorf("renameRenderer.Width = %d, want %d",
			updatedModel.editors.Rename.Renderer.Width, newWidth)
	}

	if updatedModel.editors.Rename.Renderer.Height != newHeight {
		t.Errorf("renameRenderer.Height = %d, want %d",
			updatedModel.editors.Rename.Renderer.Height, newHeight)
	}
}

func TestPipelineBuilderModel_SetSize(t *testing.T) {
	m := makeTestBuilderModel()

	newWidth := 150
	newHeight := 50

	m.SetSize(newWidth, newHeight)

	// Check that rename renderer was updated
	if m.editors.Rename.Renderer.Width != newWidth {
		t.Errorf("renameRenderer.Width = %d, want %d",
			m.editors.Rename.Renderer.Width, newWidth)
	}

	if m.editors.Rename.Renderer.Height != newHeight {
		t.Errorf("renameRenderer.Height = %d, want %d",
			m.editors.Rename.Renderer.Height, newHeight)
	}
}

// Helper to create a test error
func makeTestError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
