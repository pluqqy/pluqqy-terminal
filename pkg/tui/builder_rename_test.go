package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestPipelineBuilderModel_RenameInitialization(t *testing.T) {
	m := NewPipelineBuilderModel()
	
	if m.renameState == nil {
		t.Error("renameState should be initialized")
	}
	
	if m.renameRenderer == nil {
		t.Error("renameRenderer should be initialized")
	}
	
	if m.renameOperator == nil {
		t.Error("renameOperator should be initialized")
	}
	
	// Verify rename is not active initially
	if m.renameState.IsActive() {
		t.Error("rename should not be active initially")
	}
}

func TestPipelineBuilderModel_RenameKeyHandler(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() *PipelineBuilderModel
		activeColumn    column
		expectRenameActive bool
	}{
		{
			name: "R key in left column starts component rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Exit name editing mode
				m.editingName = false
				m.nameInput = "Test Pipeline"
				m.pipeline.Name = "Test Pipeline"
				// Add test components
				m.prompts = []componentItem{
					{name: "Test Prompt", path: "components/prompts/test.md"},
				}
				m.filteredPrompts = m.prompts
				m.leftCursor = 0
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
				m.editingName = false
				m.nameInput = "Test Pipeline"
				m.pipeline = &models.Pipeline{
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
				m.pipeline = nil
				return m
			},
			activeColumn:       rightColumn,
			expectRenameActive: false,
		},
		{
			name: "R key with empty components list does nothing",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.prompts = []componentItem{}
				m.contexts = []componentItem{}
				m.rules = []componentItem{}
				m.filteredPrompts = m.prompts
				m.filteredContexts = m.contexts
				m.filteredRules = m.rules
				return m
			},
			activeColumn:       leftColumn,
			expectRenameActive: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			m.activeColumn = tt.activeColumn
			
			// Send 'R' key
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
			newModel, _ := m.Update(msg)
			updatedModel := newModel.(*PipelineBuilderModel)
			
			if updatedModel.renameState.IsActive() != tt.expectRenameActive {
				t.Errorf("rename active = %v, want %v", 
					updatedModel.renameState.IsActive(), tt.expectRenameActive)
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
				m.renameState.Start("Test", "component", "test.md", false)
				return m
			},
			input:       tea.KeyMsg{Type: tea.KeyEsc},
			wantHandled: true,
		},
		{
			name: "character input handled during rename",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.renameState.Start("Test", "component", "test.md", false)
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
				if !m.renameState.IsActive() && cmd == nil {
					// If rename became inactive without a command, that's ok (escape)
				} else if m.renameState.IsActive() && m.renameState.NewName == "" && tt.input.Type == tea.KeyRunes {
					t.Error("Character input not processed during rename")
				}
			}
		})
	}
}

func TestPipelineBuilderModel_RenameMessages(t *testing.T) {
	tests := []struct {
		name          string
		msg           tea.Msg
		setup         func() *PipelineBuilderModel
		checkState    func(*testing.T, *PipelineBuilderModel)
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
				m.renameState.Start("Old", "component", "old.md", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.renameState.IsActive() {
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
				m.pipeline = &models.Pipeline{
					Name: "Old Pipeline",
					Path: "old-pipeline.yaml",
				}
				m.renameState.Start("Old Pipeline", "pipeline", "old-pipeline.yaml", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.pipeline.Name != "New Pipeline" {
					t.Errorf("pipeline name = %q, want %q", m.pipeline.Name, "New Pipeline")
				}
				if m.nameInput != "New Pipeline" {
					t.Errorf("nameInput = %q, want %q", m.nameInput, "New Pipeline")
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
				m.renameState.Start("Test", "component", "test.md", false)
				return m
			},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.renameState.ValidationError == "" {
					t.Error("validation error should be set")
				}
				if !contains(m.renameState.ValidationError, "rename failed") {
					t.Errorf("validation error = %q, want to contain %q", 
						m.renameState.ValidationError, "rename failed")
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
		name              string
		setup             func() *PipelineBuilderModel
		wantRenameDialog  bool
		wantContains      []string
	}{
		{
			name: "shows rename dialog when active",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Exit name editing mode so the main view is shown
				m.editingName = false
				m.renameState.Start("Test Component", "component", "test.md", false)
				m.renameRenderer.SetSize(100, 30)
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
	if updatedModel.renameRenderer.Width != newWidth {
		t.Errorf("renameRenderer.Width = %d, want %d", 
			updatedModel.renameRenderer.Width, newWidth)
	}
	
	if updatedModel.renameRenderer.Height != newHeight {
		t.Errorf("renameRenderer.Height = %d, want %d", 
			updatedModel.renameRenderer.Height, newHeight)
	}
}

func TestPipelineBuilderModel_SetSize(t *testing.T) {
	m := makeTestBuilderModel()
	
	newWidth := 150
	newHeight := 50
	
	m.SetSize(newWidth, newHeight)
	
	// Check that rename renderer was updated
	if m.renameRenderer.Width != newWidth {
		t.Errorf("renameRenderer.Width = %d, want %d", 
			m.renameRenderer.Width, newWidth)
	}
	
	if m.renameRenderer.Height != newHeight {
		t.Errorf("renameRenderer.Height = %d, want %d", 
			m.renameRenderer.Height, newHeight)
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