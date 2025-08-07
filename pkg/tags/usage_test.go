package tags

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"gopkg.in/yaml.v3"
)

func TestCleanupOrphanedTags(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "tags-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory so .pluqqy is created there
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(files.PluqqyDir, "pipelines"), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, "components/prompts"), 0755)

	// Create a test registry with multiple tags
	registry := &models.TagRegistry{
		Tags: []models.Tag{
			{Name: "used-tag", Color: "#3498db"},
			{Name: "orphaned-tag", Color: "#e74c3c"},
			{Name: "another-orphaned", Color: "#2ecc71"},
			{Name: "pipeline-tag", Color: "#f39c12"},
		},
	}

	// Save the registry
	registryPath := filepath.Join(files.PluqqyDir, TagsRegistryFile)
	data, _ := yaml.Marshal(registry)
	os.WriteFile(registryPath, data, 0644)

	// Create a test pipeline that uses some tags
	pipeline := &models.Pipeline{
		Name: "test-pipeline",
		Path: "test.yaml",
		Tags: []string{"used-tag", "pipeline-tag"},
	}
	pipelineData, _ := yaml.Marshal(pipeline)
	os.WriteFile(filepath.Join(files.PluqqyDir, "pipelines", "test.yaml"), pipelineData, 0644)

	// Write component with front matter (components are stored as markdown with YAML front matter)
	componentPath := filepath.Join(files.PluqqyDir, "components/prompts/test.md")
	componentContent := "---\ntags: [used-tag]\n---\nTest content"
	os.WriteFile(componentPath, []byte(componentContent), 0644)

	tests := []struct {
		name         string
		tagsToCheck  []string
		wantRemoved  []string
		wantKept     []string
	}{
		{
			name:        "cleanup single orphaned tag",
			tagsToCheck: []string{"orphaned-tag"},
			wantRemoved: []string{"orphaned-tag"},
			wantKept:    []string{"used-tag", "pipeline-tag", "another-orphaned"},
		},
		{
			name:        "cleanup multiple orphaned tags",
			tagsToCheck: []string{"orphaned-tag", "another-orphaned"},
			wantRemoved: []string{"orphaned-tag", "another-orphaned"},
			wantKept:    []string{"used-tag", "pipeline-tag"},
		},
		{
			name:        "don't cleanup used tags",
			tagsToCheck: []string{"used-tag", "orphaned-tag"},
			wantRemoved: []string{"orphaned-tag"},
			wantKept:    []string{"used-tag", "pipeline-tag"},
		},
		{
			name:        "handle non-existent tags gracefully",
			tagsToCheck: []string{"non-existent", "orphaned-tag"},
			wantRemoved: []string{"orphaned-tag"},
			wantKept:    []string{"used-tag", "pipeline-tag"},
		},
		{
			name:        "empty tags list",
			tagsToCheck: []string{},
			wantRemoved: []string{},
			wantKept:    []string{"used-tag", "pipeline-tag", "orphaned-tag", "another-orphaned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset registry for each test
			data, _ := yaml.Marshal(registry)
			os.WriteFile(registryPath, data, 0644)

			// Run cleanup
			removed, err := CleanupOrphanedTags(tt.tagsToCheck)
			if err != nil {
				t.Errorf("CleanupOrphanedTags() error = %v", err)
				return
			}

			// Check removed tags match expected
			if len(removed) != len(tt.wantRemoved) {
				t.Errorf("Removed %d tags, want %d", len(removed), len(tt.wantRemoved))
			}

			for _, tag := range tt.wantRemoved {
				found := false
				for _, r := range removed {
					if r == tag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' to be removed", tag)
				}
			}

			// Verify registry state
			r, err := NewRegistry()
			if err != nil {
				t.Fatalf("Failed to load registry: %v", err)
			}

			registryTags := r.ListTags()
			
			// Check that kept tags are still in registry
			for _, keepTag := range tt.wantKept {
				found := false
				for _, tag := range registryTags {
					if tag.Name == keepTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' to be kept in registry", keepTag)
				}
			}

			// Check that removed tags are not in registry
			for _, removeTag := range tt.wantRemoved {
				for _, tag := range registryTags {
					if tag.Name == removeTag {
						t.Errorf("Tag '%s' should have been removed from registry", removeTag)
					}
				}
			}
		})
	}
}

func TestCleanupOrphanedTags_ConcurrentSafety(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "tags-concurrent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(files.PluqqyDir, "pipelines"), 0755)

	// Create a test registry with tags
	registry := &models.TagRegistry{
		Tags: []models.Tag{
			{Name: "tag1", Color: "#3498db"},
			{Name: "tag2", Color: "#e74c3c"},
			{Name: "tag3", Color: "#2ecc71"},
			{Name: "tag4", Color: "#f39c12"},
			{Name: "tag5", Color: "#9b59b6"},
		},
	}

	// Save the registry
	registryPath := filepath.Join(files.PluqqyDir, TagsRegistryFile)
	data, _ := yaml.Marshal(registry)
	os.WriteFile(registryPath, data, 0644)

	// Run multiple cleanup operations concurrently
	done := make(chan bool, 5)
	tagSets := [][]string{
		{"tag1"},
		{"tag2"},
		{"tag3"},
		{"tag4"},
		{"tag5"},
	}

	for _, tags := range tagSets {
		go func(tagsToCheck []string) {
			CleanupOrphanedTags(tagsToCheck)
			done <- true
		}(tags)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify registry is still valid and not corrupted
	r, err := NewRegistry()
	if err != nil {
		t.Fatalf("Registry corrupted after concurrent cleanup: %v", err)
	}

	// All tags should be removed since none are in use
	registryTags := r.ListTags()
	if len(registryTags) != 0 {
		t.Errorf("Expected all tags to be removed, but %d remain", len(registryTags))
	}
}

func TestCleanupOrphanedTags_RegistryLoadError(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Change to a directory where we can't create .pluqqy
	if err := os.Chdir("/"); err != nil {
		t.Skip("Can't change to root directory")
	}

	// Should handle registry load error gracefully
	removed, err := CleanupOrphanedTags([]string{"some-tag"})
	// This might or might not error depending on permissions
	// but should not panic
	_ = err
	_ = removed
}

func BenchmarkCleanupOrphanedTags(b *testing.B) {
	// Save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "tags-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create necessary directories
	os.MkdirAll(filepath.Join(files.PluqqyDir, "pipelines"), 0755)

	// Create a registry with many tags
	registry := &models.TagRegistry{
		Tags: make([]models.Tag, 100),
	}
	for i := 0; i < 100; i++ {
		registry.Tags[i] = models.Tag{
			Name:  fmt.Sprintf("tag-%d", i),
			Color: "#3498db",
		}
	}

	// Save the registry
	registryPath := filepath.Join(files.PluqqyDir, TagsRegistryFile)
	data, _ := yaml.Marshal(registry)
	os.WriteFile(registryPath, data, 0644)

	// Tags to check for cleanup
	tagsToCheck := []string{"tag-0", "tag-1", "tag-2", "tag-3", "tag-4"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanupOrphanedTags(tagsToCheck)
	}
}