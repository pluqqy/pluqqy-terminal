package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// TODO: Remove this test after full migration to TagEditor
// This tests the deprecated handleTagEditing method which is no longer in use.
// Modern tag editing is handled by TagEditor component.
func TestPipelineBuilder_HandleTagEditing_VimKeysAsText(t *testing.T) {
	t.Skip("Skipping deprecated tag editing test - handleTagEditing method has been removed in favor of TagEditor")
	tests := []struct {
		name      string
		setup     func() *PipelineBuilderModel
		inputs    []tea.KeyMsg
		wantInput string
		desc      string
	}{
		{
			name: "can type j in tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("s")},
				{Type: tea.KeyRunes, Runes: []rune("o")},
				{Type: tea.KeyRunes, Runes: []rune("n")},
			},
			wantInput: "json",
			desc:      "j should be typed as regular text in tag input",
		},
		{
			name: "can type k in tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("u")},
				{Type: tea.KeyRunes, Runes: []rune("b")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
			},
			wantInput: "kube",
			desc:      "k should be typed as regular text in tag input",
		},
		{
			name: "can type l in tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("o")},
				{Type: tea.KeyRunes, Runes: []rune("g")},
			},
			wantInput: "log",
			desc:      "l should be typed as regular text in tag input",
		},
		{
			name: "can type h in tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("h")},
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("m")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
			},
			wantInput: "html",
			desc:      "h and l should be typed as regular text in tag input",
		},
		{
			name: "can type all vim keys in tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("h")},
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
			},
			wantInput: "hjkl",
			desc:      "all vim navigation keys should be typed as text",
		},
		{
			name: "vim keys work in complex tag name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = false
				return m
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("y")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
			},
			wantInput: "jekyll",
			desc:      "j, k, and double l should work in tag names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Process all inputs through handleTagEditing
			// NOTE: handleTagEditing has been removed - modern tag editing uses TagEditor
			// for _, input := range tt.inputs {
			// 	_, _ = m.handleTagEditing(input)
			// }

			// Check the result
			if m.ui.TagInput != tt.wantInput {
				t.Errorf("tagInput = %q, want %q (%s)", m.ui.TagInput, tt.wantInput, tt.desc)
			}
		})
	}
}

func TestPipelineBuilder_HandleTagEditing_ArrowNavigation(t *testing.T) {
	t.Skip("Skipping deprecated tag editing test - handleTagEditing method has been removed in favor of TagEditor")
	tests := []struct {
		name            string
		setup           func() *PipelineBuilderModel
		input           string
		wantCloudCursor int
		desc            string
	}{
		{
			name: "up arrow navigates tag cloud",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = true
				m.ui.TagCloudCursor = 2
				m.ui.AvailableTags = []string{"tag1", "tag2", "tag3"}
				return m
			},
			input:           "up",
			wantCloudCursor: 1,
			desc:            "up arrow should navigate tag cloud up",
		},
		{
			name: "down arrow navigates tag cloud",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = &models.Pipeline{
					Name: "test",
					Path: "/test/pipeline.yaml",
				}
				m.ui.EditingTags = true
				m.ui.TagCloudActive = true
				m.ui.TagCloudCursor = 0
				m.ui.AvailableTags = []string{"tag1", "tag2", "tag3"}
				return m
			},
			input:           "down",
			wantCloudCursor: 1,
			desc:            "down arrow should navigate tag cloud down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Create the appropriate key message
			// var msg tea.KeyMsg
			// switch tt.input {
			// case "up":
			// 	msg = tea.KeyMsg{Type: tea.KeyUp}
			// case "down":
			// 	msg = tea.KeyMsg{Type: tea.KeyDown}
			// }

			// Process the input
			// NOTE: handleTagEditing has been removed - modern tag editing uses TagEditor
			// _, _ = m.handleTagEditing(msg)

			// Check the result
			if m.ui.TagCloudCursor != tt.wantCloudCursor {
				t.Errorf("tagCloudCursor = %d, want %d (%s)", m.ui.TagCloudCursor, tt.wantCloudCursor, tt.desc)
			}
		})
	}
}
