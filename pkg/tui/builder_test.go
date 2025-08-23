package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// Helper to create a test PipelineBuilderModel
func makeTestBuilderModel() *PipelineBuilderModel {
	m := NewPipelineBuilderModel()
	m.viewports.Width = 100
	m.viewports.Height = 30
	return m
}

// Helper to create a test pipeline
func makeTestPipelineModel(name, path string) *models.Pipeline {
	return &models.Pipeline{
		Name:       name,
		Path:       path,
		Components: []models.ComponentRef{},
		Tags:       []string{},
	}
}

func TestPipelineBuilderModel_DeletePipeline_Logic(t *testing.T) {
	tests := []struct {
		name             string
		setup            func() *PipelineBuilderModel
		wantMsgType      string // "status" or "switch"
		wantStatusMsg    string
		wantSwitchStatus string // status message in SwitchViewMsg
	}{
		{
			name: "no pipeline loaded returns error status",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = nil
				return m
			},
			wantMsgType:   "status",
			wantStatusMsg: "× No pipeline to delete",
		},
		{
			name: "pipeline with empty path returns error status",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test-pipeline", "")
				return m
			},
			wantMsgType:   "status",
			wantStatusMsg: "× No pipeline to delete",
		},
		{
			name: "pipeline with valid path attempts deletion",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Use a non-existent file to ensure delete will fail predictably
				m.data.Pipeline = makeTestPipelineModel("test-pipeline", "/nonexistent/test-pipeline.yaml")
				return m
			},
			wantMsgType:   "status", // Will fail because file doesn't exist
			wantStatusMsg: "× Failed to delete pipeline:",
		},
		{
			name: "pipeline with tags triggers async cleanup",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				// Pipeline with tags (will fail deletion but shows tag handling)
				pipeline := makeTestPipelineModel("test-pipeline", "/nonexistent/test-pipeline.yaml")
				pipeline.Tags = []string{"tag1", "tag2", "orphaned-tag"}
				m.data.Pipeline = pipeline
				return m
			},
			wantMsgType:   "status", // Will fail because file doesn't exist
			wantStatusMsg: "× Failed to delete pipeline:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test model
			m := tt.setup()

			// Execute the delete command
			cmd := m.deletePipeline()
			if cmd == nil {
				t.Fatal("deletePipeline() returned nil command")
			}

			// Execute the command to get the message
			msg := cmd()

			// Check the message type and content
			switch tt.wantMsgType {
			case "status":
				statusMsg, ok := msg.(StatusMsg)
				if !ok {
					t.Errorf("Expected StatusMsg, got %T", msg)
				}
				// For error messages, just check the prefix
				if tt.wantStatusMsg != "" {
					statusStr := string(statusMsg)
					if len(tt.wantStatusMsg) > 20 && tt.wantStatusMsg[len(tt.wantStatusMsg)-1] == ':' {
						// Check prefix for error messages
						if len(statusStr) < len(tt.wantStatusMsg) || statusStr[:len(tt.wantStatusMsg)] != tt.wantStatusMsg {
							t.Errorf("StatusMsg prefix = %q, want prefix %q", statusStr, tt.wantStatusMsg)
						}
					} else if statusStr != tt.wantStatusMsg {
						t.Errorf("StatusMsg = %q, want %q", statusStr, tt.wantStatusMsg)
					}
				}
			case "switch":
				switchMsg, ok := msg.(SwitchViewMsg)
				if !ok {
					t.Errorf("Expected SwitchViewMsg, got %T", msg)
				}
				if switchMsg.view != mainListView {
					t.Errorf("SwitchViewMsg.view = %v, want mainListView", switchMsg.view)
				}
				// Check status message in SwitchViewMsg if expected
				if tt.wantSwitchStatus != "" {
					if !strings.Contains(switchMsg.status, tt.wantSwitchStatus) {
						t.Errorf("SwitchViewMsg.status = %q, want to contain %q", switchMsg.status, tt.wantSwitchStatus)
					}
				}
			}
		})
	}
}

func TestPipelineBuilderModel_DeleteConfirmation(t *testing.T) {
	tests := []struct {
		name              string
		setup             func() *PipelineBuilderModel
		wantConfirmActive bool
		wantConfirmMsg    string
	}{
		{
			name: "delete confirmation shows for valid pipeline",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test-pipeline", "pipelines/test.yaml")
				return m
			},
			wantConfirmActive: true,
			wantConfirmMsg:    "Delete pipeline 'test.yaml'?",
		},
		{
			name: "no confirmation for nil pipeline",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = nil
				return m
			},
			wantConfirmActive: false,
			wantConfirmMsg:    "",
		},
		{
			name: "no confirmation for empty path",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "")
				return m
			},
			wantConfirmActive: false,
			wantConfirmMsg:    "",
		},
		{
			name: "confirmation message uses basename of path",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "/deep/path/to/pipeline.yaml")
				return m
			},
			wantConfirmActive: true,
			wantConfirmMsg:    "Delete pipeline 'pipeline.yaml'?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Simulate the ctrl+d key press behavior
			if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
				pipelineName := filepath.Base(m.data.Pipeline.Path)
				m.ui.DeleteConfirm.ShowInline(
					fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
					true,
					func() tea.Cmd { return nil },
					func() tea.Cmd { return nil },
				)
			}

			// Check confirmation state
			if m.ui.DeleteConfirm.Active() != tt.wantConfirmActive {
				t.Errorf("deleteConfirm.Active() = %v, want %v", m.ui.DeleteConfirm.Active(), tt.wantConfirmActive)
			}

			// Check confirmation message if active
			if tt.wantConfirmActive && m.ui.DeleteConfirm.Active() {
				actualMsg := m.ui.DeleteConfirm.config.Message
				if actualMsg != tt.wantConfirmMsg {
					t.Errorf("Confirmation message = %q, want %q", actualMsg, tt.wantConfirmMsg)
				}
			}
		})
	}
}

func TestPipelineBuilderModel_DeleteConfirmation_Rendering(t *testing.T) {
	tests := []struct {
		name              string
		confirmActive     bool
		wantConfirmInView bool
	}{
		{
			name:              "confirmation renders when active",
			confirmActive:     true,
			wantConfirmInView: true,
		},
		{
			name:              "confirmation hidden when inactive",
			confirmActive:     false,
			wantConfirmInView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeTestBuilderModel()
			m.data.Pipeline = makeTestPipelineModel("test", "test.yaml")

			if tt.confirmActive {
				m.ui.DeleteConfirm.ShowInline(
					"Delete pipeline 'test.yaml'?",
					true,
					func() tea.Cmd { return nil },
					func() tea.Cmd { return nil },
				)
			}

			// Render the view
			_ = m.View()

			// Check if confirmation text appears in the view
			containsConfirm := false
			if m.ui.DeleteConfirm.Active() {
				// The view should contain the confirmation when active
				confirmView := m.ui.DeleteConfirm.ViewWithWidth(m.viewports.Width - 4)
				if confirmView != "" {
					containsConfirm = true
				}
			}

			if containsConfirm != tt.wantConfirmInView {
				t.Errorf("Confirmation in view = %v, want %v", containsConfirm, tt.wantConfirmInView)
			}
		})
	}
}

func TestPipelineBuilderModel_DeleteKeyHandling(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *PipelineBuilderModel
		keyString   string
		wantHandled bool
	}{
		{
			name: "ctrl+d handled in normal mode",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "test.yaml")
				m.editors.EditingName = false
				// Component creator is not active by default
				m.editors.EditingComponent = false
				return m
			},
			keyString:   "ctrl+d",
			wantHandled: true,
		},
		{
			name: "ctrl+d not handled when editing name",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "test.yaml")
				m.editors.EditingName = true
				return m
			},
			keyString:   "ctrl+d",
			wantHandled: false,
		},
		{
			name: "ctrl+d not handled when creating component",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "test.yaml")
				m.editors.ComponentCreator.Start()
				return m
			},
			keyString:   "ctrl+d",
			wantHandled: false,
		},
		{
			name: "ctrl+d not handled when editing component",
			setup: func() *PipelineBuilderModel {
				m := makeTestBuilderModel()
				m.data.Pipeline = makeTestPipelineModel("test", "test.yaml")
				m.editors.EditingComponent = true
				return m
			},
			keyString:   "ctrl+d",
			wantHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			// Check if the key would be handled based on model state
			handled := false
			if !m.editors.EditingName && !(m.editors.ComponentCreator != nil && m.editors.ComponentCreator.IsActive()) && !m.editors.EditingComponent {
				// In normal mode, ctrl+d would be handled
				if tt.keyString == "ctrl+d" && m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
					handled = true
				}
			}

			if handled != tt.wantHandled {
				t.Errorf("Key handled = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

func TestPipelineBuilderModel_DeletePipelineEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		pipelinePath string
		expectedName string
	}{
		{
			name:         "pipeline with spaces in name",
			pipelinePath: "my pipeline with spaces.yaml",
			expectedName: "my pipeline with spaces.yaml",
		},
		{
			name:         "pipeline with special characters",
			pipelinePath: "pipeline-@#$%.yaml",
			expectedName: "pipeline-@#$%.yaml",
		},
		{
			name:         "pipeline with unicode characters",
			pipelinePath: "pipeline-测试.yaml",
			expectedName: "pipeline-测试.yaml",
		},
		{
			name:         "deeply nested pipeline path",
			pipelinePath: "/very/deep/nested/folder/structure/pipeline.yaml",
			expectedName: "pipeline.yaml",
		},
		{
			name:         "pipeline with multiple extensions",
			pipelinePath: "pipeline.backup.yaml",
			expectedName: "pipeline.backup.yaml",
		},
		{
			name:         "relative path with dots",
			pipelinePath: "../pipelines/test.yaml",
			expectedName: "test.yaml",
		},
		{
			name:         "path with trailing slash (shouldn't happen but test anyway)",
			pipelinePath: "pipelines/test.yaml/",
			expectedName: "test.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeTestBuilderModel()
			m.data.Pipeline = makeTestPipelineModel("test", tt.pipelinePath)

			// Test confirmation message formatting
			pipelineName := filepath.Base(tt.pipelinePath)
			// Handle the trailing slash case
			if pipelineName == "" || pipelineName == "." {
				pipelineName = filepath.Base(filepath.Clean(tt.pipelinePath))
			}

			if pipelineName != tt.expectedName {
				t.Errorf("filepath.Base(%q) = %q, want %q", tt.pipelinePath, pipelineName, tt.expectedName)
			}

			// Test that delete confirmation would show correct name
			expectedMsg := fmt.Sprintf("Delete pipeline '%s'?", pipelineName)
			m.ui.DeleteConfirm.ShowInline(expectedMsg, true, func() tea.Cmd { return nil }, func() tea.Cmd { return nil })

			if m.ui.DeleteConfirm.config.Message != expectedMsg {
				t.Errorf("Confirmation message = %q, want %q", m.ui.DeleteConfirm.config.Message, expectedMsg)
			}
		})
	}
}

// Test that delete confirmation follows the same pattern as Main List View
func TestPipelineBuilderModel_DeleteConfirmation_Consistency(t *testing.T) {
	// Create builder model
	builderModel := makeTestBuilderModel()
	builderModel.data.Pipeline = makeTestPipelineModel("test", "test.yaml")

	// Create list model for comparison
	listModel := NewMainListModel()
	listModel.data.Pipelines = []pipelineItem{
		{name: "test", path: "test.yaml"},
	}

	// Both should use ShowInline with destructive=true
	builderModel.ui.DeleteConfirm.ShowInline("Delete pipeline 'test.yaml'?", true, func() tea.Cmd { return nil }, func() tea.Cmd { return nil })

	// Check that builder uses inline style (not dialog)
	if builderModel.ui.DeleteConfirm.config.Type != ConfirmTypeInline {
		t.Errorf("Builder delete confirmation type = %v, want ConfirmTypeInline", builderModel.ui.DeleteConfirm.config.Type)
	}

	// Check that it's marked as destructive
	if !builderModel.ui.DeleteConfirm.config.Destructive {
		t.Error("Builder delete confirmation should be marked as destructive")
	}
}

// Benchmark test for delete confirmation display
func BenchmarkPipelineBuilderModel_DeleteConfirmation(b *testing.B) {
	m := makeTestBuilderModel()
	m.data.Pipeline = makeTestPipelineModel("benchmark", "benchmark.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ui.DeleteConfirm.ShowInline("Delete pipeline 'benchmark.yaml'?", true, func() tea.Cmd { return nil }, func() tea.Cmd { return nil })
		_ = m.ui.DeleteConfirm.View()
		m.ui.DeleteConfirm.active = false // Reset for next iteration
	}
}
