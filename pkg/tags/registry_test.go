package tags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

func TestRegistry(t *testing.T) {
	// Create a temporary directory and change to it
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// Initialize project structure
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}
	
	t.Run("NewRegistry", func(t *testing.T) {
		registry, err := NewRegistry()
		if err != nil {
			t.Errorf("NewRegistry() error = %v", err)
			return
		}
		
		if registry == nil {
			t.Error("NewRegistry() returned nil")
		}
		
		// Should start with empty tags
		tags := registry.ListTags()
		if len(tags) != 0 {
			t.Errorf("New registry has %d tags, want 0", len(tags))
		}
	})
	
	t.Run("AddTag", func(t *testing.T) {
		registry, _ := NewRegistry()
		
		tag := models.Tag{
			Name:        "API",
			Color:       "#3498db",
			Description: "API-related components",
		}
		
		err := registry.AddTag(tag)
		if err != nil {
			t.Errorf("AddTag() error = %v", err)
			return
		}
		
		// Tag should be normalized
		retrieved, exists := registry.GetTag("api")
		if !exists {
			t.Error("Tag not found after adding")
			return
		}
		
		if retrieved.Name != "api" {
			t.Errorf("Tag name = %q, want %q", retrieved.Name, "api")
		}
		
		if retrieved.Color != tag.Color {
			t.Errorf("Tag color = %q, want %q", retrieved.Color, tag.Color)
		}
	})
	
	t.Run("GetOrCreateTag", func(t *testing.T) {
		registry, _ := NewRegistry()
		
		// Create new tag
		tag1, err := registry.GetOrCreateTag("frontend")
		if err != nil {
			t.Errorf("GetOrCreateTag() error = %v", err)
			return
		}
		
		if tag1.Name != "frontend" {
			t.Errorf("Tag name = %q, want %q", tag1.Name, "frontend")
		}
		
		// Should have auto-assigned color
		if tag1.Color == "" {
			t.Error("Tag color not auto-assigned")
		}
		
		// Get existing tag
		tag2, err := registry.GetOrCreateTag("frontend")
		if err != nil {
			t.Errorf("GetOrCreateTag() error = %v", err)
			return
		}
		
		if tag2.Color != tag1.Color {
			t.Errorf("Tag color changed: %q != %q", tag2.Color, tag1.Color)
		}
	})
	
	t.Run("RemoveTag", func(t *testing.T) {
		registry, _ := NewRegistry()
		
		// Add a tag
		tag := models.Tag{Name: "temp-tag"}
		registry.AddTag(tag)
		
		// Remove it
		err := registry.RemoveTag("temp-tag")
		if err != nil {
			t.Errorf("RemoveTag() error = %v", err)
			return
		}
		
		// Should not exist
		_, exists := registry.GetTag("temp-tag")
		if exists {
			t.Error("Tag still exists after removal")
		}
		
		// Remove non-existent tag
		err = registry.RemoveTag("non-existent")
		if err == nil {
			t.Error("RemoveTag() should error for non-existent tag")
		}
	})
	
	t.Run("SaveAndLoad", func(t *testing.T) {
		registry1, _ := NewRegistry()
		
		// Add some tags
		tags := []models.Tag{
			{Name: "api", Color: "#3498db", Description: "API stuff"},
			{Name: "frontend", Color: "#e74c3c", Description: "Frontend stuff"},
			{Name: "backend", Color: "#2ecc71"},
		}
		
		for _, tag := range tags {
			registry1.AddTag(tag)
		}
		
		// Save
		err := registry1.Save()
		if err != nil {
			t.Errorf("Save() error = %v", err)
			return
		}
		
		// Create new registry and load
		registry2, err := NewRegistry()
		if err != nil {
			t.Errorf("NewRegistry() error = %v", err)
			return
		}
		
		// Verify tags were loaded
		loadedTags := registry2.ListTags()
		if len(loadedTags) != len(tags) {
			t.Errorf("Loaded %d tags, want %d", len(loadedTags), len(tags))
		}
		
		// Verify tag details
		for _, originalTag := range tags {
			loaded, exists := registry2.GetTag(originalTag.Name)
			if !exists {
				t.Errorf("Tag %q not found after load", originalTag.Name)
				continue
			}
			
			if loaded.Color != originalTag.Color {
				t.Errorf("Tag %q color = %q, want %q", originalTag.Name, loaded.Color, originalTag.Color)
			}
		}
	})
	
	t.Run("RenameTag", func(t *testing.T) {
		registry, _ := NewRegistry()
		
		// Add a tag
		tag := models.Tag{Name: "old-name", Color: "#123456"}
		registry.AddTag(tag)
		
		// Rename it
		err := registry.RenameTag("old-name", "new-name")
		if err != nil {
			t.Errorf("RenameTag() error = %v", err)
			return
		}
		
		// Old name should not exist
		_, exists := registry.GetTag("old-name")
		if exists {
			t.Error("Old tag name still exists")
		}
		
		// New name should exist with same color
		newTag, exists := registry.GetTag("new-name")
		if !exists {
			t.Error("New tag name not found")
		}
		
		if newTag.Color != tag.Color {
			t.Errorf("Tag color changed: %q != %q", newTag.Color, tag.Color)
		}
	})
	
	t.Run("ValidateTagName", func(t *testing.T) {
		registry, _ := NewRegistry()
		
		// Invalid tag names
		invalidTags := []models.Tag{
			{Name: ""},                    // Empty
			{Name: "tag@with#special"},    // Special chars
			{Name: strings.Repeat("a", 51)}, // Too long
		}
		
		for _, tag := range invalidTags {
			err := registry.AddTag(tag)
			if err == nil {
				t.Errorf("AddTag(%q) should have failed", tag.Name)
			}
		}
	})
}

func TestGetTagStats(t *testing.T) {
	// Create a temporary directory and change to it
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)
	
	// Initialize project structure
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}
	
	registry, _ := NewRegistry()
	
	// Create some components with tags
	component1 := filepath.Join(files.ComponentsDir, files.PromptsDir, "test1.md")
	files.WriteComponentWithNameAndTags(component1, "# Test 1", "Test 1", []string{"api", "v2"})
	
	component2 := filepath.Join(files.ComponentsDir, files.PromptsDir, "test2.md")
	files.WriteComponentWithNameAndTags(component2, "# Test 2", "Test 2", []string{"api", "frontend"})
	
	// Create a pipeline with tags
	pipeline := &models.Pipeline{
		Name: "test-pipeline",
		Tags: []string{"api", "production"},
		Components: []models.ComponentRef{
			{Type: files.PromptsDir, Path: component1, Order: 1},
		},
	}
	files.WritePipeline(pipeline)
	
	// Get stats
	stats, err := registry.GetTagStats()
	if err != nil {
		t.Errorf("GetTagStats() error = %v", err)
		return
	}
	
	// Verify counts
	expectedStats := map[string]int{
		"api":        3, // 2 components + 1 pipeline
		"v2":         1,
		"frontend":   1,
		"production": 1,
	}
	
	for tag, expectedCount := range expectedStats {
		if stats[tag] != expectedCount {
			t.Errorf("Tag %q count = %d, want %d", tag, stats[tag], expectedCount)
		}
	}
}

