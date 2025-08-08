package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
	"gopkg.in/yaml.v3"
)

// Test helper to create a temporary test environment
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pipeline-ops-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(files.PluqqyDir, "pipelines"), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, "components/prompts"), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, "components/contexts"), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, "components/rules"), 0755)

	// Cleanup function
	cleanup := func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// Test helper to create a test pipeline
func createTestPipeline(t *testing.T, name string, tags []string) string {
	pipeline := &models.Pipeline{
		Name: name,
		Path: name + ".yaml",
		Tags: tags,
	}
	
	pipelinePath := filepath.Join(files.PluqqyDir, "pipelines", pipeline.Path)
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		t.Fatalf("Failed to write pipeline: %v", err)
	}
	
	return pipeline.Path // Return just the filename, not the full path
}

// Test helper to create a test component
func createTestComponent(t *testing.T, name, compType string, componentTags []string) componentItem {
	componentPath := filepath.Join("components", compType, name+".md")
	fullPath := filepath.Join(files.PluqqyDir, componentPath)
	
	// Create component with YAML front matter
	content := "---\n"
	if len(componentTags) > 0 {
		content += "tags: ["
		for i, tag := range componentTags {
			if i > 0 {
				content += ", "
			}
			content += tag
		}
		content += "]\n"
	}
	content += "---\nTest content for " + name
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write component: %v", err)
	}
	
	return componentItem{
		name:     name,
		path:     componentPath,
		compType: compType,
		tags:     componentTags,
	}
}

// Test helper to create a test tag registry
func createTestRegistry(t *testing.T, testTags []string) {
	registry := &models.TagRegistry{
		Tags: make([]models.Tag, len(testTags)),
	}
	
	for i, tag := range testTags {
		registry.Tags[i] = models.Tag{
			Name:  tag,
			Color: "#3498db",
		}
	}
	
	registryPath := filepath.Join(files.PluqqyDir, tags.TagsRegistryFile)
	data, err := yaml.Marshal(registry)
	if err != nil {
		t.Fatalf("Failed to marshal registry: %v", err)
	}
	
	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		t.Fatalf("Failed to write registry: %v", err)
	}
}

func TestPipelineOperator_DeletePipeline(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (string, []string)
		expectedStatus string
		expectError    bool
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "delete pipeline with orphaned tags",
			setupFunc: func(t *testing.T) (string, []string) {
				// Create registry with tags
				createTestRegistry(t, []string{"tag1", "tag2", "orphaned-tag"})
				
				// Create pipelines
				createTestPipeline(t, "test-pipeline", []string{"tag1", "orphaned-tag"})
				createTestPipeline(t, "other-pipeline", []string{"tag1", "tag2"})
				
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
			setupFunc: func(t *testing.T) (string, []string) {
				createTestPipeline(t, "no-tags-pipeline", nil)
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
			setupFunc: func(t *testing.T) (string, []string) {
				createTestRegistry(t, []string{"shared-tag", "unique-tag"})
				
				// Create multiple pipelines sharing tags
				createTestPipeline(t, "pipeline1", []string{"shared-tag", "unique-tag"})
				createTestPipeline(t, "pipeline2", []string{"shared-tag"})
				
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
			_, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Setup test case
			pipelinePath, pipelineTags := tt.setupFunc(t)
			
			// Create pipeline operator
			po := NewPipelineOperator()
			
			// Track if reload was called
			reloadCalled := false
			reloadFunc := func() {
				reloadCalled = true
			}
			
			// Execute delete
			cmd := po.DeletePipeline(pipelinePath, pipelineTags, reloadFunc)
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
		setupFunc      func(t *testing.T) componentItem
		expectedStatus string
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "delete component with orphaned tags",
			setupFunc: func(t *testing.T) componentItem {
				createTestRegistry(t, []string{"comp-tag1", "comp-tag2", "orphaned-comp-tag"})
				
				// Create components
				comp1 := createTestComponent(t, "test-prompt", "prompts", []string{"comp-tag1", "orphaned-comp-tag"})
				createTestComponent(t, "other-prompt", "prompts", []string{"comp-tag1", "comp-tag2"})
				
				return comp1
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
			setupFunc: func(t *testing.T) componentItem {
				return createTestComponent(t, "no-tags-context", "contexts", nil)
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
			setupFunc: func(t *testing.T) componentItem {
				createTestRegistry(t, []string{"rule-shared", "rule-unique"})
				
				comp1 := createTestComponent(t, "rule1", "rules", []string{"rule-shared", "rule-unique"})
				createTestComponent(t, "rule2", "rules", []string{"rule-shared"})
				
				return comp1
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
			_, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Setup test case
			comp := tt.setupFunc(t)
			
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
		setupFunc      func(t *testing.T) string
		expectedStatus string
		verifyFunc     func(t *testing.T)
	}{
		{
			name: "archive pipeline successfully",
			setupFunc: func(t *testing.T) string {
				return createTestPipeline(t, "test-pipeline", []string{"tag1"})
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
			_, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Setup test case
			pipelinePath := tt.setupFunc(t)
			
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