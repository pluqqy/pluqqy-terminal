package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestListArchivedPipelines(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create archive directory structure
	archiveDir := filepath.Join(".pluqqy", "archive", "pipelines")
	err := os.MkdirAll(archiveDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create archive directory: %v", err)
	}

	tests := []struct {
		name          string
		setup         func()
		expectedCount int
		expectedNames []string
		expectError   bool
	}{
		{
			name: "no archived pipelines",
			setup: func() {
				// Archive directory exists but is empty
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "single archived pipeline",
			setup: func() {
				pipelinePath := filepath.Join(archiveDir, "test-pipeline.yaml")
				content := `---
name: test-pipeline
tags: [api, v1]
---

# Test Pipeline`
				err := os.WriteFile(pipelinePath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test pipeline: %v", err)
				}
			},
			expectedCount: 1,
			expectedNames: []string{"test-pipeline.yaml"},
		},
		{
			name: "multiple archived pipelines",
			setup: func() {
				pipelines := []string{"pipeline1.yaml", "pipeline2.yaml", "pipeline3.yaml"}
				for _, p := range pipelines {
					content := `---
name: ` + p + `
tags: []
---

# Pipeline`
					err := os.WriteFile(filepath.Join(archiveDir, p), []byte(content), 0644)
					if err != nil {
						t.Fatalf("Failed to write pipeline %s: %v", p, err)
					}
				}
			},
			expectedCount: 3,
			expectedNames: []string{"pipeline1.yaml", "pipeline2.yaml", "pipeline3.yaml"},
		},
		{
			name: "filters out non-yaml files",
			setup: func() {
				// Create both .yaml and non-.yaml files
				os.WriteFile(filepath.Join(archiveDir, "valid.yaml"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(archiveDir, "invalid.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(archiveDir, "README"), []byte("content"), 0644)
			},
			expectedCount: 1,
			expectedNames: []string{"valid.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean archive directory
			os.RemoveAll(archiveDir)
			os.MkdirAll(archiveDir, 0755)

			// Run setup
			tt.setup()

			// Test
			pipelines, err := ListArchivedPipelines()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(pipelines) != tt.expectedCount {
				t.Errorf("Expected %d pipelines, got %d", tt.expectedCount, len(pipelines))
			}

			// Check pipeline names match
			for i, expectedName := range tt.expectedNames {
				if i >= len(pipelines) {
					break
				}
				baseName := filepath.Base(pipelines[i])
				if baseName != expectedName {
					t.Errorf("Pipeline[%d]: expected %s, got %s", i, expectedName, baseName)
				}
			}
		})
	}
}

func TestListArchivedComponents(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	tests := []struct {
		name          string
		componentType string
		setup         func()
		expectedCount int
		expectedNames []string
		expectError   bool
	}{
		{
			name:          "no archived components",
			componentType: models.ComponentTypeContext,
			setup: func() {
				// Create empty archive directory
				archiveDir := filepath.Join(".pluqqy", "archive", "components", "contexts")
				os.MkdirAll(archiveDir, 0755)
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "archived contexts",
			componentType: models.ComponentTypeContext,
			setup: func() {
				archiveDir := filepath.Join(".pluqqy", "archive", "components", "contexts")
				os.MkdirAll(archiveDir, 0755)
				
				contexts := []string{"api-context.md", "auth-context.md"}
				for _, c := range contexts {
					content := `---
name: ` + c + `
tags: [context]
---

# Context`
					err := os.WriteFile(filepath.Join(archiveDir, c), []byte(content), 0644)
					if err != nil {
						t.Fatalf("Failed to write context %s: %v", c, err)
					}
				}
			},
			expectedCount: 2,
			expectedNames: []string{"api-context.md", "auth-context.md"},
		},
		{
			name:          "archived prompts",
			componentType: models.ComponentTypePrompt,
			setup: func() {
				archiveDir := filepath.Join(".pluqqy", "archive", "components", "prompts")
				os.MkdirAll(archiveDir, 0755)
				
				prompts := []string{"test-prompt.md", "review-prompt.md"}
				for _, p := range prompts {
					content := `---
name: ` + p + `
tags: [prompt]
---

# Prompt`
					err := os.WriteFile(filepath.Join(archiveDir, p), []byte(content), 0644)
					if err != nil {
						t.Fatalf("Failed to write prompt %s: %v", p, err)
					}
				}
			},
			expectedCount: 2,
			expectedNames: []string{"test-prompt.md", "review-prompt.md"},
		},
		{
			name:          "archived rules",
			componentType: models.ComponentTypeRules,
			setup: func() {
				archiveDir := filepath.Join(".pluqqy", "archive", "components", "rules")
				os.MkdirAll(archiveDir, 0755)
				
				rules := []string{"validation-rules.md"}
				for _, r := range rules {
					content := `---
name: ` + r + `
tags: [rules]
---

# Rules`
					err := os.WriteFile(filepath.Join(archiveDir, r), []byte(content), 0644)
					if err != nil {
						t.Fatalf("Failed to write rules %s: %v", r, err)
					}
				}
			},
			expectedCount: 1,
			expectedNames: []string{"validation-rules.md"},
		},
		{
			name:          "filters out non-markdown files",
			componentType: models.ComponentTypeContext,
			setup: func() {
				archiveDir := filepath.Join(".pluqqy", "archive", "components", "contexts")
				os.MkdirAll(archiveDir, 0755)
				
				// Create both .md and non-.md files
				os.WriteFile(filepath.Join(archiveDir, "valid.md"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(archiveDir, "invalid.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(archiveDir, ".DS_Store"), []byte("content"), 0644)
			},
			expectedCount: 1,
			expectedNames: []string{"valid.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean entire archive directory
			archiveBase := filepath.Join(tmpDir, ".pluqqy", "archive")
			os.RemoveAll(archiveBase)

			// Run setup
			tt.setup()

			// Test
			components, err := ListArchivedComponents(tt.componentType)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(components) != tt.expectedCount {
				t.Errorf("Expected %d components, got %d", tt.expectedCount, len(components))
			}

			// Check component names match
			for i, expectedName := range tt.expectedNames {
				if i >= len(components) {
					break
				}
				baseName := filepath.Base(components[i])
				if baseName != expectedName {
					t.Errorf("Component[%d]: expected %s, got %s", i, expectedName, baseName)
				}
			}
		})
	}
}

func TestReadArchivedPipeline(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	archiveDir := filepath.Join(".pluqqy", "archive", "pipelines")
	os.MkdirAll(archiveDir, 0755)

	tests := []struct {
		name        string
		setup       func() string // returns path to test
		expectError bool
		validate    func(*testing.T, *models.Pipeline)
	}{
		{
			name: "read valid archived pipeline",
			setup: func() string {
				pipelinePath := filepath.Join(archiveDir, "test-pipeline.yaml")
				content := `---
name: Test Pipeline
tags: [api, v2]
---

# Test Pipeline

This is a test pipeline.`
				err := os.WriteFile(pipelinePath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test pipeline: %v", err)
				}
				return filepath.Base(pipelinePath)
			},
			validate: func(t *testing.T, p *models.Pipeline) {
				if p.Name != "Test Pipeline" {
					t.Errorf("Expected name 'Test Pipeline', got %s", p.Name)
				}
				if len(p.Tags) != 2 {
					t.Errorf("Expected 2 tags, got %d", len(p.Tags))
				}
				// Pipeline doesn't have Content field, just verify it was read
			},
		},
		{
			name: "read non-existent pipeline",
			setup: func() string {
				return "non-existent.yaml"
			},
			expectError: true,
		},
		{
			name: "read pipeline with minimal yaml",
			setup: func() string {
				pipelinePath := filepath.Join(archiveDir, "no-frontmatter.yaml")
				content := `name: no-frontmatter
components: []`
				err := os.WriteFile(pipelinePath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test pipeline: %v", err)
				}
				return filepath.Base(pipelinePath)
			},
			validate: func(t *testing.T, p *models.Pipeline) {
				if p.Name != "no-frontmatter" {
					t.Errorf("Expected name derived from filename, got %s", p.Name)
				}
				if len(p.Tags) != 0 {
					t.Errorf("Expected no tags, got %d", len(p.Tags))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			pipeline, err := ReadArchivedPipeline(path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, pipeline)
			}
		})
	}
}

func TestReadArchivedComponent(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	archiveDir := filepath.Join(".pluqqy", "archive", "components", "contexts")
	os.MkdirAll(archiveDir, 0755)

	tests := []struct {
		name        string
		setup       func() string // returns path to test
		expectError bool
		validate    func(*testing.T, *models.Component)
	}{
		{
			name: "read valid archived component",
			setup: func() string {
				componentPath := filepath.Join(archiveDir, "test-context.md")
				content := `---
name: Test Context
tags: [api, context]
---

# Test Context

This is a test context component.`
				err := os.WriteFile(componentPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test component: %v", err)
				}
				return filepath.Base(componentPath)
			},
			validate: func(t *testing.T, c *models.Component) {
				// Component doesn't have Name field, check from path
				baseName := filepath.Base(c.Path)
				if baseName != "test-context.md" {
					t.Errorf("Expected filename 'test-context.md', got %s", baseName)
				}
				if c.Type != models.ComponentTypeContext {
					t.Errorf("Expected type 'context', got %s", c.Type)
				}
				if len(c.Tags) != 2 {
					t.Errorf("Expected 2 tags, got %d", len(c.Tags))
				}
				if c.Content == "" {
					t.Error("Expected non-empty content")
				}
			},
		},
		{
			name: "read non-existent component",
			setup: func() string {
				return "non-existent.yaml"
			},
			expectError: true,
		},
		{
			name: "read component with no frontmatter",
			setup: func() string {
				componentPath := filepath.Join(archiveDir, "no-frontmatter.md")
				content := `# Component without frontmatter

Just content here.`
				err := os.WriteFile(componentPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test component: %v", err)
				}
				return filepath.Base(componentPath)
			},
			validate: func(t *testing.T, c *models.Component) {
				baseName := filepath.Base(c.Path)
				if baseName != "no-frontmatter.md" {
					t.Errorf("Expected filename 'no-frontmatter.md', got %s", baseName)
				}
				if c.Type != models.ComponentTypeContext {
					t.Errorf("Expected type 'context' from path, got %s", c.Type)
				}
				if len(c.Tags) != 0 {
					t.Errorf("Expected no tags, got %d", len(c.Tags))
				}
			},
		},
		{
			name: "component type from path - prompts",
			setup: func() string {
				promptDir := filepath.Join(".pluqqy", "archive", "components", "prompts")
				os.MkdirAll(promptDir, 0755)
				componentPath := filepath.Join(promptDir, "test-prompt.md")
				content := `# Test Prompt`
				err := os.WriteFile(componentPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test component: %v", err)
				}
				return filepath.Base(componentPath)
			},
			validate: func(t *testing.T, c *models.Component) {
				if c.Type != models.ComponentTypePrompt {
					t.Errorf("Expected type 'prompt' from path, got %s", c.Type)
				}
			},
		},
		{
			name: "component type from path - rules",
			setup: func() string {
				rulesDir := filepath.Join(".pluqqy", "archive", "components", "rules")
				os.MkdirAll(rulesDir, 0755)
				componentPath := filepath.Join(rulesDir, "test-rules.md")
				content := `# Test Rules`
				err := os.WriteFile(componentPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test component: %v", err)
				}
				return filepath.Base(componentPath)
			},
			validate: func(t *testing.T, c *models.Component) {
				if c.Type != models.ComponentTypeRules {
					t.Errorf("Expected type 'rules' from path, got %s", c.Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			component, err := ReadArchivedComponent(path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, component)
			}
		})
	}
}
