package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/testhelpers"
)

func TestPipelineOperator_DeletePipeline(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, env *testhelpers.TestEnvironment) (string, []string)
		expectedStatus string
		expectError    bool
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "delete pipeline with orphaned tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) (string, []string) {
				// Create registry with tags
				env.CreateTagRegistry([]string{"tag1", "tag2", "orphaned-tag"})

				// Create pipelines
				env.CreatePipelineFile("test-pipeline", nil, []string{"tag1", "orphaned-tag"})
				env.CreatePipelineFile("other-pipeline", nil, []string{"tag1", "tag2"})

				return "test-pipeline.yaml", []string{"tag1", "orphaned-tag"}
			},
			expectedStatus: "✓ Deleted pipeline: test-pipeline",
			verifyFunc: func(t *testing.T) {
				// Wait for async cleanup
				time.Sleep(100 * time.Millisecond)

				// Verify pipeline was deleted
				pipelinePath := filepath.Join(files.PluqqyDir, "pipelines/test-pipeline.yaml")
				if _, err := os.Stat(pipelinePath); !os.IsNotExist(err) {
					t.Error("Pipeline file should have been deleted")
				}

				// Verify orphaned tag was removed from registry
				registry, _ := tags.NewRegistry()
				for _, tag := range registry.ListTags() {
					if tag.Name == "orphaned-tag" {
						t.Error("Orphaned tag should have been removed from registry")
					}
				}

				// Verify used tags remain
				foundTag1 := false
				foundTag2 := false
				for _, tag := range registry.ListTags() {
					if tag.Name == "tag1" {
						foundTag1 = true
					}
					if tag.Name == "tag2" {
						foundTag2 = true
					}
				}
				if !foundTag1 || !foundTag2 {
					t.Error("Used tags should remain in registry")
				}
			},
		},
		{
			name: "delete pipeline with no tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) (string, []string) {
				env.CreatePipelineFile("no-tags-pipeline", nil, nil)
				return "no-tags-pipeline.yaml", nil
			},
			expectedStatus: "✓ Deleted pipeline: no-tags-pipeline",
			verifyFunc: func(t *testing.T) {
				pipelinePath := filepath.Join(files.PluqqyDir, "pipelines/no-tags-pipeline.yaml")
				if _, err := os.Stat(pipelinePath); !os.IsNotExist(err) {
					t.Error("Pipeline file should have been deleted")
				}
			},
		},
		{
			name: "delete pipeline with shared tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) (string, []string) {
				env.CreateTagRegistry([]string{"shared-tag", "unique-tag"})

				// Create multiple pipelines sharing tags
				env.CreatePipelineFile("pipeline1", nil, []string{"shared-tag", "unique-tag"})
				env.CreatePipelineFile("pipeline2", nil, []string{"shared-tag"})

				return "pipeline1.yaml", []string{"shared-tag", "unique-tag"}
			},
			expectedStatus: "✓ Deleted pipeline: pipeline1",
			verifyFunc: func(t *testing.T) {
				// Wait for async cleanup
				time.Sleep(100 * time.Millisecond)

				// Verify shared tag remains
				registry, _ := tags.NewRegistry()
				foundShared := false
				for _, tag := range registry.ListTags() {
					if tag.Name == "shared-tag" {
						foundShared = true
					}
				}
				if !foundShared {
					t.Error("Shared tag should remain in registry")
				}

				// Verify unique tag was removed
				for _, tag := range registry.ListTags() {
					if tag.Name == "unique-tag" {
						t.Error("Unique tag should have been removed")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testhelpers.NewTestEnvironment(t)
			defer env.Cleanup()
			env.ChangeToTempDir()
			env.InitProjectStructure()

			// Setup test case
			pipelinePath, pipelineTags := tt.setupFunc(t, env)

			// Create pipeline operator
			po := NewPipelineOperator()

			// Track if reload was called
			reloadCalled := false
			reloadFunc := func() {
				reloadCalled = true
			}

			// Execute delete (pass false for isArchived in tests)
			cmd := po.DeletePipeline(pipelinePath, pipelineTags, false, reloadFunc)
			msg := cmd()

			// Verify status message
			if statusMsg, ok := msg.(StatusMsg); ok {
				if string(statusMsg) != tt.expectedStatus {
					t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, string(statusMsg))
				}
			} else if !tt.expectError {
				t.Error("Expected StatusMsg, got different message type")
			}

			// Verify reload was called
			if !reloadCalled {
				t.Error("Reload function should have been called")
			}

			// Run additional verifications
			if tt.verifyFunc != nil {
				tt.verifyFunc(t)
			}
		})
	}
}

func TestPipelineOperator_DeleteComponent(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, env *testhelpers.TestEnvironment) componentItem
		expectedStatus string
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "delete component with orphaned tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) componentItem {
				env.CreateTagRegistry([]string{"comp-tag1", "comp-tag2", "orphaned-comp-tag"})

				// Create components
				env.CreateComponentFile("prompts", "test-prompt", "Test content", []string{"comp-tag1", "orphaned-comp-tag"})
				env.CreateComponentFile("prompts", "other-prompt", "Other content", []string{"comp-tag1", "comp-tag2"})

				return componentItem{
					name:     "test-prompt",
					path:     "components/prompts/test-prompt.md",
					compType: "prompts",
					tags:     []string{"comp-tag1", "orphaned-comp-tag"},
				}
			},
			expectedStatus: "✓ Deleted prompts: test-prompt",
			verifyFunc: func(t *testing.T) {
				// Wait for async cleanup
				time.Sleep(100 * time.Millisecond)

				// Verify component was deleted
				componentPath := filepath.Join(files.PluqqyDir, "components/prompts/test-prompt.md")
				if _, err := os.Stat(componentPath); !os.IsNotExist(err) {
					t.Error("Component file should have been deleted")
				}

				// Verify orphaned tag was removed
				registry, _ := tags.NewRegistry()
				for _, tag := range registry.ListTags() {
					if tag.Name == "orphaned-comp-tag" {
						t.Error("Orphaned component tag should have been removed")
					}
				}
			},
		},
		{
			name: "delete component with no tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) componentItem {
				env.CreateComponentFile("contexts", "no-tags-context", "Test content", nil)
				
				return componentItem{
					name:     "no-tags-context",
					path:     "components/contexts/no-tags-context.md",
					compType: "contexts",
					tags:     nil,
				}
			},
			expectedStatus: "✓ Deleted contexts: no-tags-context",
			verifyFunc: func(t *testing.T) {
				componentPath := filepath.Join(files.PluqqyDir, "components/contexts/no-tags-context.md")
				if _, err := os.Stat(componentPath); !os.IsNotExist(err) {
					t.Error("Component file should have been deleted")
				}
			},
		},
		{
			name: "delete rule component with shared tags",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) componentItem {
				env.CreateTagRegistry([]string{"rule-shared", "rule-unique"})

				env.CreateComponentFile("rules", "rule1", "Rule content", []string{"rule-shared", "rule-unique"})
				env.CreateComponentFile("rules", "rule2", "Rule content", []string{"rule-shared"})

				return componentItem{
					name:     "rule1",
					path:     "components/rules/rule1.md",
					compType: "rules",
					tags:     []string{"rule-shared", "rule-unique"},
				}
			},
			expectedStatus: "✓ Deleted rules: rule1",
			verifyFunc: func(t *testing.T) {
				// Wait for async cleanup
				time.Sleep(100 * time.Millisecond)

				// Verify shared tag remains
				registry, _ := tags.NewRegistry()
				foundShared := false
				for _, tag := range registry.ListTags() {
					if tag.Name == "rule-shared" {
						foundShared = true
					}
				}
				if !foundShared {
					t.Error("Shared tag should remain in registry")
				}

				// Verify unique tag was removed
				for _, tag := range registry.ListTags() {
					if tag.Name == "rule-unique" {
						t.Error("Unique tag should have been removed")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testhelpers.NewTestEnvironment(t)
			defer env.Cleanup()
			env.ChangeToTempDir()
			env.InitProjectStructure()

			// Setup test case
			comp := tt.setupFunc(t, env)

			// Create pipeline operator
			po := NewPipelineOperator()

			// Track if reload was called
			reloadCalled := false
			reloadFunc := func() {
				reloadCalled = true
			}

			// Execute delete
			cmd := po.DeleteComponent(comp, reloadFunc)
			msg := cmd()

			// Verify status message
			if statusMsg, ok := msg.(StatusMsg); ok {
				if string(statusMsg) != tt.expectedStatus {
					t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, string(statusMsg))
				}
			} else {
				t.Error("Expected StatusMsg")
			}

			// Verify reload was called
			if !reloadCalled {
				t.Error("Reload function should have been called")
			}

			// Run additional verifications
			if tt.verifyFunc != nil {
				tt.verifyFunc(t)
			}
		})
	}
}

func TestPipelineOperator_ArchivePipeline(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, env *testhelpers.TestEnvironment) string
		expectedStatus string
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "archive pipeline successfully",
			setupFunc: func(t *testing.T, env *testhelpers.TestEnvironment) string {
				return env.CreatePipelineFile("test-pipeline", nil, []string{"tag1"})
			},
			expectedStatus: "✓ Archived pipeline: test-pipeline",
			verifyFunc: func(t *testing.T) {
				// Verify pipeline was moved to archive
				originalPath := filepath.Join(files.PluqqyDir, "pipelines/test-pipeline.yaml")
				archivePath := filepath.Join(files.PluqqyDir, "archive/pipelines/test-pipeline.yaml")

				if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
					t.Error("Original pipeline file should have been moved")
				}

				if _, err := os.Stat(archivePath); os.IsNotExist(err) {
					t.Error("Archived pipeline file should exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testhelpers.NewTestEnvironment(t)
			defer env.Cleanup()
			env.ChangeToTempDir()
			env.InitProjectStructure()

			// Setup test case
			pipelinePath := tt.setupFunc(t, env)

			// Create pipeline operator
			po := NewPipelineOperator()

			// Track if reload was called
			reloadCalled := false
			reloadFunc := func() {
				reloadCalled = true
			}

			// Execute archive
			cmd := po.ArchivePipeline(pipelinePath, reloadFunc)
			msg := cmd()

			// Verify status message
			if statusMsg, ok := msg.(StatusMsg); ok {
				if string(statusMsg) != tt.expectedStatus {
					t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, string(statusMsg))
				}
			}

			// Verify reload was called
			if !reloadCalled {
				t.Error("Reload function should have been called")
			}

			// Run additional verifications
			if tt.verifyFunc != nil {
				tt.verifyFunc(t)
			}
		})
	}
}

func TestPipelineOperator_Confirmations(t *testing.T) {
	po := NewPipelineOperator()

	t.Run("delete confirmation workflow", func(t *testing.T) {
		// Initially not active
		if po.IsDeleteConfirmActive() {
			t.Error("Delete confirmation should not be active initially")
		}

		// Show confirmation
		confirmCalled := false
		cancelCalled := false

		po.ShowDeleteConfirmation(
			"Delete test?",
			func() tea.Cmd {
				confirmCalled = true
				return nil
			},
			func() tea.Cmd {
				cancelCalled = true
				return nil
			},
		)

		// Should now be active
		if !po.IsDeleteConfirmActive() {
			t.Error("Delete confirmation should be active after showing")
		}

		// Simulate pressing 'y'
		cmd := po.UpdateDeleteConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		if cmd != nil {
			// Execute the command if it exists
			cmd()
		}

		// Verify confirm was called
		if !confirmCalled {
			t.Error("Confirm callback should have been called")
		}
		if cancelCalled {
			t.Error("Cancel callback should not have been called")
		}
	})

	t.Run("archive confirmation workflow", func(t *testing.T) {
		// Initially not active
		if po.IsArchiveConfirmActive() {
			t.Error("Archive confirmation should not be active initially")
		}

		// Show confirmation
		confirmCalled := false
		cancelCalled := false

		po.ShowArchiveConfirmation(
			"Archive test?",
			func() tea.Cmd {
				confirmCalled = true
				return nil
			},
			func() tea.Cmd {
				cancelCalled = true
				return nil
			},
		)

		// Should now be active
		if !po.IsArchiveConfirmActive() {
			t.Error("Archive confirmation should be active after showing")
		}

		// Simulate pressing 'n'
		cmd := po.UpdateArchiveConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		if cmd != nil {
			// Execute the command if it exists
			cmd()
		}

		// Verify cancel was called
		if confirmCalled {
			t.Error("Confirm callback should not have been called")
		}
		if !cancelCalled {
			t.Error("Cancel callback should have been called")
		}
	})
}
