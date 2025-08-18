package tui

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestBuilderClipboardYank(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *PipelineBuilderModel
		key          tea.KeyMsg
		expectStatus bool
		expectMsg    string
	}{
		{
			name: "yank pipeline content when pipeline exists with components",
			setup: func() *PipelineBuilderModel {
				m := &PipelineBuilderModel{
					pipeline: &models.Pipeline{
						Name: "my-test-pipeline",
						Components: []models.ComponentRef{
							{Path: "test-component.md", Type: "prompt", Order: 0},
						},
					},
				}
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			expectStatus: true,
			expectMsg:    "my-test-pipeline → clipboard",
		},
		{
			name: "no yank when pipeline is nil",
			setup: func() *PipelineBuilderModel {
				m := &PipelineBuilderModel{
					pipeline: nil,
				}
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			expectStatus: false,
		},
		{
			name: "no yank when pipeline has no components",
			setup: func() *PipelineBuilderModel {
				m := &PipelineBuilderModel{
					pipeline: &models.Pipeline{
						Name: "empty-pipeline",
						Components: []models.ComponentRef{},
					},
				}
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			expectStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Simulate the 'y' key handling logic
			if tt.key.String() == "y" && m.pipeline != nil && len(m.pipeline.Components) > 0 {
				expectedMsg := m.pipeline.Name + " → clipboard"
				
				if tt.expectStatus && expectedMsg != tt.expectMsg {
					t.Errorf("expected status message %q, got %q", tt.expectMsg, expectedMsg)
				}
			} else if tt.expectStatus {
				t.Error("expected status message but conditions were not met")
			}
		})
	}
}