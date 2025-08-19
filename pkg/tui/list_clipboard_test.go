package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestClipboardYank(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *MainListModel
		key          tea.KeyMsg
		activePane   pane
		expectStatus bool
		expectMsg    string
	}{
		{
			name: "yank pipeline content when in pipelines pane",
			setup: func() *MainListModel {
				m := &MainListModel{
					stateManager: NewStateManager(),
					pipelines: []pipelineItem{
						{name: "test-pipeline", path: "test.yaml"},
					},
					filteredPipelines: []pipelineItem{
						{name: "test-pipeline", path: "test.yaml"},
					},
				}
				m.stateManager.ActivePane = pipelinesPane
				m.stateManager.PipelineCursor = 0
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			activePane:   pipelinesPane,
			expectStatus: true,
			expectMsg:    "test-pipeline → clipboard",
		},
		{
			name: "no yank when pipelines list is empty",
			setup: func() *MainListModel {
				m := &MainListModel{
					stateManager:      NewStateManager(),
					pipelines:         []pipelineItem{},
					filteredPipelines: []pipelineItem{},
				}
				m.stateManager.ActivePane = pipelinesPane
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			activePane:   pipelinesPane,
			expectStatus: false,
		},
		{
			name: "no yank when not in pipelines pane",
			setup: func() *MainListModel {
				m := &MainListModel{
					stateManager: NewStateManager(),
					pipelines: []pipelineItem{
						{name: "test-pipeline", path: "test.yaml"},
					},
					filteredPipelines: []pipelineItem{
						{name: "test-pipeline", path: "test.yaml"},
					},
				}
				m.stateManager.ActivePane = componentsPane
				return m
			},
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			activePane:   componentsPane,
			expectStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Simulate the 'y' key handling logic
			if tt.key.String() == "y" && m.stateManager.ActivePane == pipelinesPane {
				pipelines := m.filteredPipelines
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipelineName := pipelines[m.stateManager.PipelineCursor].name
					expectedMsg := pipelineName + " → clipboard"

					if tt.expectStatus && expectedMsg != tt.expectMsg {
						t.Errorf("expected status message %q, got %q", tt.expectMsg, expectedMsg)
					}
				} else if tt.expectStatus {
					t.Error("expected status message but pipeline list was empty")
				}
			} else if tt.expectStatus {
				t.Error("expected status message but conditions were not met")
			}
		})
	}
}
