package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedTags    []string
		expectedContent string
	}{
		{
			name: "with tags",
			content: `---
tags: [api, authentication, v2]
---

# Component Content
This is the actual content.`,
			expectedTags:    []string{"api", "authentication", "v2"},
			expectedContent: "\n# Component Content\nThis is the actual content.",
		},
		{
			name: "no frontmatter",
			content: `# Component Content
This is the actual content.`,
			expectedTags:    nil,
			expectedContent: "# Component Content\nThis is the actual content.",
		},
		{
			name: "empty tags",
			content: `---
tags: []
---

# Component Content`,
			expectedTags:    []string{},
			expectedContent: "\n# Component Content",
		},
		{
			name: "malformed frontmatter",
			content: `---
This is not valid YAML

# Component Content`,
			expectedTags:    nil,
			expectedContent: "---\nThis is not valid YAML\n\n# Component Content",
		},
		{
			name: "frontmatter with other fields",
			content: `---
title: My Component
tags: [api, v2]
author: John Doe
---

# Content`,
			expectedTags:    []string{"api", "v2"},
			expectedContent: "\n# Content",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, content, err := extractFrontmatter([]byte(tt.content))
			if err != nil {
				t.Errorf("extractFrontmatter() error = %v", err)
				return
			}
			
			// Check tags
			if len(frontmatter.Tags) != len(tt.expectedTags) {
				t.Errorf("extractFrontmatter() tags = %v, want %v", frontmatter.Tags, tt.expectedTags)
			} else {
				for i, tag := range frontmatter.Tags {
					if tag != tt.expectedTags[i] {
						t.Errorf("extractFrontmatter() tag[%d] = %q, want %q", i, tag, tt.expectedTags[i])
					}
				}
			}
			
			// Check content
			if string(content) != tt.expectedContent {
				t.Errorf("extractFrontmatter() content = %q, want %q", string(content), tt.expectedContent)
			}
		})
	}
}


func TestComponentTagOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pluqqy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Save original working directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)
	
	// Initialize project structure
	if err := InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}
	
	// Test component path
	componentPath := filepath.Join(ComponentsDir, PromptsDir, "test-component.md")
	content := "# Test Component\nThis is a test."
	
	t.Run("WriteComponentWithNameAndTags", func(t *testing.T) {
		tags := []string{"test", "api", "v2"}
		err := WriteComponentWithNameAndTags(componentPath, content, "Test Component", tags)
		if err != nil {
			t.Errorf("WriteComponentWithNameAndTags() error = %v", err)
			return
		}
		
		// Read and verify
		comp, err := ReadComponent(componentPath)
		if err != nil {
			t.Errorf("ReadComponent() error = %v", err)
			return
		}
		
		if len(comp.Tags) != len(tags) {
			t.Errorf("Component tags = %v, want %v", comp.Tags, tags)
		}
	})
	
	t.Run("AddComponentTag", func(t *testing.T) {
		err := AddComponentTag(componentPath, "new-tag")
		if err != nil {
			t.Errorf("AddComponentTag() error = %v", err)
			return
		}
		
		comp, err := ReadComponent(componentPath)
		if err != nil {
			t.Errorf("ReadComponent() error = %v", err)
			return
		}
		
		found := false
		for _, tag := range comp.Tags {
			if tag == "new-tag" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tag 'new-tag' not found in component tags: %v", comp.Tags)
		}
	})
	
	t.Run("RemoveComponentTag", func(t *testing.T) {
		err := RemoveComponentTag(componentPath, "api")
		if err != nil {
			t.Errorf("RemoveComponentTag() error = %v", err)
			return
		}
		
		comp, err := ReadComponent(componentPath)
		if err != nil {
			t.Errorf("ReadComponent() error = %v", err)
			return
		}
		
		for _, tag := range comp.Tags {
			if tag == "api" {
				t.Errorf("Tag 'api' still found in component tags: %v", comp.Tags)
			}
		}
	})
	
	t.Run("UpdateComponentTags", func(t *testing.T) {
		newTags := []string{"updated", "fresh", "clean"}
		err := UpdateComponentTags(componentPath, newTags)
		if err != nil {
			t.Errorf("UpdateComponentTags() error = %v", err)
			return
		}
		
		comp, err := ReadComponent(componentPath)
		if err != nil {
			t.Errorf("ReadComponent() error = %v", err)
			return
		}
		
		if len(comp.Tags) != len(newTags) {
			t.Errorf("Component tags = %v, want %v", comp.Tags, newTags)
		}
		
		for i, tag := range comp.Tags {
			if tag != newTags[i] {
				t.Errorf("Component tag[%d] = %q, want %q", i, tag, newTags[i])
			}
		}
	})
}