package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"gopkg.in/yaml.v3"
)

func TestUnarchivePipeline(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		pipelinePath string
		wantErr     bool
		errContains string
		validate    func(t *testing.T)
	}{
		{
			name: "successful unarchive",
			setup: func(t *testing.T) string {
				// Create temp directory structure
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create archive structure with pipeline
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)
				os.MkdirAll(archiveDir, 0755)
				
				// Create active pipelines dir
				activeDir := filepath.Join(PluqqyDir, PipelinesDir)
				os.MkdirAll(activeDir, 0755)
				
				// Create archived pipeline
				pipeline := models.Pipeline{
					Name: "Test Pipeline",
					Tags: []string{"test"},
				}
				data, _ := yaml.Marshal(pipeline)
				os.WriteFile(filepath.Join(archiveDir, "test.yaml"), data, 0644)
				
				return "test.yaml"
			},
			pipelinePath: "test.yaml",
			wantErr:     false,
			validate: func(t *testing.T) {
				// Check file exists in active directory
				activePath := filepath.Join(PluqqyDir, PipelinesDir, "test.yaml")
				if _, err := os.Stat(activePath); os.IsNotExist(err) {
					t.Error("Pipeline not found in active directory after unarchive")
				}
				
				// Check file removed from archive
				archivePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, "test.yaml")
				if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
					t.Error("Pipeline still exists in archive after unarchive")
				}
			},
		},
		{
			name: "unarchive non-existent pipeline",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create directory structure but no file
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)
				os.MkdirAll(archiveDir, 0755)
				
				return "nonexistent.yaml"
			},
			pipelinePath: "nonexistent.yaml",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "unarchive with invalid path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				return "../../../etc/passwd"
			},
			pipelinePath: "../../../etc/passwd",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name: "unarchive when active file already exists",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create both archived and active versions
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)
				os.MkdirAll(archiveDir, 0755)
				activeDir := filepath.Join(PluqqyDir, PipelinesDir)
				os.MkdirAll(activeDir, 0755)
				
				// Create files in both locations
				pipeline := models.Pipeline{Name: "Test"}
				data, _ := yaml.Marshal(pipeline)
				os.WriteFile(filepath.Join(archiveDir, "test.yaml"), data, 0644)
				os.WriteFile(filepath.Join(activeDir, "test.yaml"), []byte("existing"), 0644)
				
				return "test.yaml"
			},
			pipelinePath: "test.yaml",
			wantErr:     true,
			errContains: "cannot unarchive: active pipeline already exists",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			if path == "" {
				path = tt.pipelinePath
			}
			
			err := UnarchivePipeline(path)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t)
				}
			}
		})
	}
}

func TestUnarchiveComponent(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		componentPath string
		wantErr      bool
		errContains  string
		validate     func(t *testing.T)
	}{
		{
			name: "successful unarchive context",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create archive structure
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, ContextsDir)
				os.MkdirAll(archiveDir, 0755)
				
				// Create active directory
				activeDir := filepath.Join(PluqqyDir, ComponentsDir, ContextsDir)
				os.MkdirAll(activeDir, 0755)
				
				// Create archived component
				content := `---
tags: [test]
---
# Test Context
Content here`
				archivePath := filepath.Join(archiveDir, "test.md")
				os.WriteFile(archivePath, []byte(content), 0644)
				
				return filepath.Join(ComponentsDir, ContextsDir, "test.md")
			},
			componentPath: "",
			wantErr:      false,
			validate: func(t *testing.T) {
				// Check file exists in active directory
				activePath := filepath.Join(PluqqyDir, ComponentsDir, ContextsDir, "test.md")
				if _, err := os.Stat(activePath); os.IsNotExist(err) {
					t.Error("Component not found in active directory after unarchive")
				}
				
				// Check file removed from archive
				archivePath := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, ContextsDir, "test.md")
				if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
					t.Error("Component still exists in archive after unarchive")
				}
			},
		},
		{
			name: "successful unarchive prompt",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create archive structure
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, PromptsDir)
				os.MkdirAll(archiveDir, 0755)
				
				// Create active directory
				activeDir := filepath.Join(PluqqyDir, ComponentsDir, PromptsDir)
				os.MkdirAll(activeDir, 0755)
				
				// Create archived component
				content := "# Test Prompt"
				archivePath := filepath.Join(archiveDir, "test.md")
				os.WriteFile(archivePath, []byte(content), 0644)
				
				return filepath.Join(ComponentsDir, PromptsDir, "test.md")
			},
			componentPath: "",
			wantErr:      false,
		},
		{
			name: "unarchive non-existent component",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				
				// Create directory structure but no file
				archiveDir := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, ContextsDir)
				os.MkdirAll(archiveDir, 0755)
				
				return filepath.Join(ComponentsDir, ContextsDir, "nonexistent.md")
			},
			componentPath: "",
			wantErr:      true,
			errContains:  "not found",
		},
		{
			name: "unarchive with invalid path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				return ""
			},
			componentPath: "../../../etc/passwd",
			wantErr:      true,
			errContains:  "invalid",
		},
		{
			name: "unarchive with path traversal in component",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				t.Cleanup(func() { os.Chdir(oldWd) })
				os.Chdir(tmpDir)
				return ""
			},
			componentPath: "components/../../../etc/passwd",
			wantErr:      true,
			errContains:  "invalid",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			if tt.componentPath != "" {
				path = tt.componentPath
			}
			
			err := UnarchiveComponent(path)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t)
				}
			}
		})
	}
}

// Test concurrent unarchive operations
func TestUnarchiveConcurrent(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)
	
	// Create multiple archived files
	archivePipelineDir := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)
	os.MkdirAll(archivePipelineDir, 0755)
	activePipelineDir := filepath.Join(PluqqyDir, PipelinesDir)
	os.MkdirAll(activePipelineDir, 0755)
	
	archiveComponentDir := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, ContextsDir)
	os.MkdirAll(archiveComponentDir, 0755)
	activeComponentDir := filepath.Join(PluqqyDir, ComponentsDir, ContextsDir)
	os.MkdirAll(activeComponentDir, 0755)
	
	// Create test files
	for i := 0; i < 5; i++ {
		pipeline := models.Pipeline{Name: "Test"}
		data, _ := yaml.Marshal(pipeline)
		os.WriteFile(filepath.Join(archivePipelineDir, filepath.Base(filepath.Clean(string(rune('a'+i)))+".yaml")), data, 0644)
		os.WriteFile(filepath.Join(archiveComponentDir, filepath.Base(filepath.Clean(string(rune('a'+i)))+".md")), []byte("test"), 0644)
	}
	
	// Run concurrent unarchive operations
	done := make(chan bool, 10)
	errors := make(chan error, 10)
	
	for i := 0; i < 5; i++ {
		go func(idx int) {
			err := UnarchivePipeline(filepath.Base(filepath.Clean(string(rune('a'+idx))) + ".yaml"))
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
		
		go func(idx int) {
			err := UnarchiveComponent(filepath.Join(ComponentsDir, ContextsDir, filepath.Base(filepath.Clean(string(rune('a'+idx)))+".md")))
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}
	
	// Wait for all operations
	for i := 0; i < 10; i++ {
		<-done
	}
	
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
	
	// Verify all files were moved
	for i := 0; i < 5; i++ {
		pipelinePath := filepath.Join(PluqqyDir, PipelinesDir, filepath.Base(filepath.Clean(string(rune('a'+i)))+".yaml"))
		if _, err := os.Stat(pipelinePath); os.IsNotExist(err) {
			t.Errorf("Pipeline %d not unarchived", i)
		}
		
		componentPath := filepath.Join(PluqqyDir, ComponentsDir, ContextsDir, filepath.Base(filepath.Clean(string(rune('a'+i)))+".md"))
		if _, err := os.Stat(componentPath); os.IsNotExist(err) {
			t.Errorf("Component %d not unarchived", i)
		}
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && s[0:len(substr)] == substr || (len(s) > len(substr) && containsString(s[1:], substr)))
}