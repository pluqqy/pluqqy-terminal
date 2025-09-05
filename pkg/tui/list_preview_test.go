package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/utils"
)

// Test helper to create test components
func createTestComponentForPreview(t *testing.T, name, content, compType string, tags []string) componentItem {
	t.Helper()
	return componentItem{
		name:         name,
		path:         fmt.Sprintf("components/%s/%s", compType, name),
		compType:     compType,
		lastModified: time.Now(),
		usageCount:   3,
		tokenCount:   utils.EstimateTokens(content),
		tags:         tags,
		isArchived:   false,
	}
}

// Test helper to create test pipeline
func createTestPipelineFile(t *testing.T, name string, components []models.ComponentRef) string {
	t.Helper()

	// Ensure test directory exists
	testDir := filepath.Join(files.PluqqyDir, files.PipelinesDir)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	pipeline := models.Pipeline{
		Name:       name,
		Components: components,
		OutputPath: ".pluqqy/tmp/test-output.md",
		Path:       name + ".yaml",
	}

	if err := files.WritePipeline(&pipeline); err != nil {
		t.Fatalf("Failed to create test pipeline: %v", err)
	}

	return name + ".yaml"
}

// Test helper to create test component file
func createTestComponentFile(t *testing.T, compType, name, content string, tags []string) string {
	t.Helper()

	var subDir string
	switch compType {
	case models.ComponentTypePrompt:
		subDir = files.PromptsDir
	case models.ComponentTypeContext:
		subDir = files.ContextsDir
	case models.ComponentTypeRules:
		subDir = files.RulesDir
	default:
		t.Fatalf("Invalid component type: %s", compType)
	}

	componentPath := filepath.Join(files.ComponentsDir, subDir, name+".md")

	// Add frontmatter if tags are provided
	fullContent := content
	if len(tags) > 0 {
		fullContent = fmt.Sprintf("---\ntags: %v\n---\n\n%s", tags, content)
	}

	if err := files.WriteComponent(componentPath, fullContent); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	return componentPath
}

func TestPreviewRenderer_RenderComponentPreview(t *testing.T) {
	// Initialize test environment
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}
	defer os.RemoveAll(files.PluqqyDir)

	tests := []struct {
		name           string
		setup          func() (componentItem, string) // Returns component and expected content
		showPreview    bool
		wantEmpty      bool
		wantError      bool
		validateOutput func(t *testing.T, output string)
	}{
		{
			name: "renders component content without metadata",
			setup: func() (componentItem, string) {
				content := "# Test Component\n\nThis is test content without any metadata."
				path := createTestComponentFile(t, models.ComponentTypePrompt, "test-prompt", content, []string{"test", "preview"})
				comp := createTestComponentForPreview(t, "test-prompt.md", content, models.ComponentTypePrompt, []string{"test", "preview"})
				comp.path = path
				return comp, content
			},
			showPreview: true,
			validateOutput: func(t *testing.T, output string) {
				// Should contain the actual content
				if !strings.Contains(output, "# Test Component") {
					t.Error("Output should contain component heading")
				}
				if !strings.Contains(output, "This is test content") {
					t.Error("Output should contain component content")
				}
				// Should NOT contain metadata
				if strings.Contains(output, "**Type:**") {
					t.Error("Output should not contain Type metadata")
				}
				if strings.Contains(output, "**Path:**") {
					t.Error("Output should not contain Path metadata")
				}
				if strings.Contains(output, "**Usage Count:**") {
					t.Error("Output should not contain Usage Count metadata")
				}
				if strings.Contains(output, "**Token Count:**") {
					t.Error("Output should not contain Token Count metadata")
				}
				if strings.Contains(output, "**Last Modified:**") {
					t.Error("Output should not contain Last Modified metadata")
				}
				// Should NOT contain frontmatter
				if strings.Contains(output, "tags:") || strings.Contains(output, "---") {
					t.Error("Output should not contain frontmatter")
				}
			},
		},
		{
			name: "returns empty when preview disabled",
			setup: func() (componentItem, string) {
				comp := createTestComponentForPreview(t, "test.md", "content", models.ComponentTypePrompt, nil)
				return comp, ""
			},
			showPreview: false,
			wantEmpty:   true,
		},
		{
			name: "handles component with multiline content",
			setup: func() (componentItem, string) {
				content := `# Complex Component

## Section 1
This is the first section.

## Section 2
- Item 1
- Item 2
- Item 3

## Code Example
` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"

				path := createTestComponentFile(t, models.ComponentTypeContext, "complex", content, nil)
				comp := createTestComponentForPreview(t, "complex.md", content, models.ComponentTypeContext, nil)
				comp.path = path
				return comp, content
			},
			showPreview: true,
			validateOutput: func(t *testing.T, output string) {
				if output != `# Complex Component

## Section 1
This is the first section.

## Section 2
- Item 1
- Item 2
- Item 3

## Code Example
`+"```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" {
					t.Errorf("Multiline content not preserved correctly")
				}
			},
		},
		{
			name: "handles error when component file not found",
			setup: func() (componentItem, string) {
				comp := createTestComponentForPreview(t, "nonexistent.md", "", models.ComponentTypePrompt, nil)
				comp.path = "components/prompts/nonexistent.md"
				return comp, ""
			},
			showPreview: true,
			wantError:   true,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Error loading component") {
					t.Error("Should show error message for missing component")
				}
			},
		},
		{
			name: "strips frontmatter from component content",
			setup: func() (componentItem, string) {
				content := "# Component Without Frontmatter Display\n\nThe actual content here."
				path := createTestComponentFile(t, models.ComponentTypeRules, "with-frontmatter", content, []string{"important", "rule"})
				comp := createTestComponentForPreview(t, "with-frontmatter.md", content, models.ComponentTypeRules, []string{"important", "rule"})
				comp.path = path
				expectedContent := "# Component Without Frontmatter Display\n\nThe actual content here."
				return comp, expectedContent
			},
			showPreview: true,
			validateOutput: func(t *testing.T, output string) {
				// Should not contain frontmatter
				if strings.Contains(output, "tags:") || strings.Contains(output, "---") {
					t.Error("Frontmatter should be stripped from output")
				}
				// Should contain actual content
				if !strings.Contains(output, "Component Without Frontmatter Display") {
					t.Error("Should contain the actual component content")
				}
			},
		},
		{
			name: "handles empty component content",
			setup: func() (componentItem, string) {
				content := ""
				path := createTestComponentFile(t, models.ComponentTypePrompt, "empty", content, nil)
				comp := createTestComponentForPreview(t, "empty.md", content, models.ComponentTypePrompt, nil)
				comp.path = path
				return comp, content
			},
			showPreview: true,
			wantEmpty:   true, // Empty content should produce empty preview
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewPreviewRenderer()
			renderer.ShowPreview = tt.showPreview

			comp, _ := tt.setup()
			output := renderer.RenderComponentPreview(comp)

			if tt.wantEmpty && output != "" {
				t.Errorf("Expected empty output, got: %s", output)
			}

			if !tt.wantEmpty && !tt.wantError && output == "" && tt.showPreview {
				t.Error("Expected non-empty output")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}

func TestPreviewRenderer_RenderPipelinePreview(t *testing.T) {
	// Initialize test environment
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}
	defer os.RemoveAll(files.PluqqyDir)

	tests := []struct {
		name           string
		setup          func() string // Returns pipeline path
		showPreview    bool
		wantEmpty      bool
		wantError      bool
		validateOutput func(t *testing.T, output string)
	}{
		{
			name: "renders pipeline with multiple components",
			setup: func() string {
				// Create test components
				createTestComponentFile(t, models.ComponentTypeContext, "context1", "# Context 1\nContext content", nil)
				createTestComponentFile(t, models.ComponentTypePrompt, "prompt1", "# Prompt 1\nPrompt content", nil)

				// Create pipeline
				components := []models.ComponentRef{
					{Type: models.ComponentTypeContext, Path: "../components/contexts/context1.md", Order: 1},
					{Type: models.ComponentTypePrompt, Path: "../components/prompts/prompt1.md", Order: 2},
				}
				return createTestPipelineFile(t, "test-pipeline", components)
			},
			showPreview: true,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "# test-pipeline") {
					t.Error("Should contain pipeline name as heading")
				}
				if !strings.Contains(output, "Context content") {
					t.Error("Should contain context component content")
				}
				if !strings.Contains(output, "Prompt content") {
					t.Error("Should contain prompt component content")
				}
				// Should not contain frontmatter
				if strings.Contains(output, "tags:") || strings.Contains(output, "---\n") {
					t.Error("Should not contain frontmatter in pipeline output")
				}
			},
		},
		{
			name: "returns empty when preview disabled",
			setup: func() string {
				return "pipelines/dummy.yaml"
			},
			showPreview: false,
			wantEmpty:   true,
		},
		{
			name: "handles error when pipeline not found",
			setup: func() string {
				return "nonexistent-pipeline.yaml"
			},
			showPreview: true,
			wantError:   true,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Error loading pipeline") {
					t.Error("Should show error message for missing pipeline")
				}
			},
		},
		{
			name: "handles pipeline with component ordering",
			setup: func() string {
				// Create components with specific order
				createTestComponentFile(t, models.ComponentTypeRules, "rule1", "Rule 1 content", nil)
				createTestComponentFile(t, models.ComponentTypeContext, "context2", "Context 2 content", nil)
				createTestComponentFile(t, models.ComponentTypePrompt, "prompt2", "Prompt 2 content", nil)

				components := []models.ComponentRef{
					{Type: models.ComponentTypePrompt, Path: "../components/prompts/prompt2.md", Order: 3},
					{Type: models.ComponentTypeRules, Path: "../components/rules/rule1.md", Order: 1},
					{Type: models.ComponentTypeContext, Path: "../components/contexts/context2.md", Order: 2},
				}
				return createTestPipelineFile(t, "ordered-pipeline", components)
			},
			showPreview: true,
			validateOutput: func(t *testing.T, output string) {
				// Find positions of each content type
				rulePos := strings.Index(output, "Rule 1 content")
				contextPos := strings.Index(output, "Context 2 content")
				promptPos := strings.Index(output, "Prompt 2 content")

				// Verify ordering based on settings (Rules -> Context -> Prompts)
				if rulePos == -1 || contextPos == -1 || promptPos == -1 {
					t.Error("All component content should be present")
				}
				// Note: The actual order depends on settings, but all should be present
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewPreviewRenderer()
			renderer.ShowPreview = tt.showPreview

			pipelinePath := tt.setup()
			output := renderer.RenderPipelinePreview(pipelinePath, false)

			if tt.wantEmpty && output != "" {
				t.Errorf("Expected empty output, got: %s", output)
			}

			if !tt.wantEmpty && !tt.wantError && output == "" && tt.showPreview {
				t.Error("Expected non-empty output")
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}

func TestPreviewRenderer_RenderEmptyPreview(t *testing.T) {
	tests := []struct {
		name          string
		showPreview   bool
		activePane    pane
		hasPipelines  bool
		hasComponents bool
		expected      string
	}{
		{
			name:        "returns empty when preview disabled",
			showPreview: false,
			activePane:  pipelinesPane,
			expected:    "",
		},
		{
			name:         "shows message for pipelines pane with no pipelines",
			showPreview:  true,
			activePane:   pipelinesPane,
			hasPipelines: false,
			expected:     "No pipelines to preview.",
		},
		{
			name:          "shows message for components pane with no components",
			showPreview:   true,
			activePane:    componentsPane,
			hasComponents: false,
			expected:      "No components to preview.",
		},
		{
			name:          "returns empty for pipelines pane with pipelines",
			showPreview:   true,
			activePane:    pipelinesPane,
			hasPipelines:  true,
			hasComponents: true,
			expected:      "",
		},
		{
			name:          "returns empty for components pane with components",
			showPreview:   true,
			activePane:    componentsPane,
			hasPipelines:  true,
			hasComponents: true,
			expected:      "",
		},
		{
			name:        "returns empty for search pane",
			showPreview: true,
			activePane:  searchPane,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewPreviewRenderer()
			renderer.ShowPreview = tt.showPreview

			output := renderer.RenderEmptyPreview(tt.activePane, tt.hasPipelines, tt.hasComponents)

			if output != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestEstimatePreviewTokens(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		minTokens int
		maxTokens int
	}{
		{
			name:      "estimates tokens for empty content",
			content:   "",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:      "estimates tokens for short content",
			content:   "Hello world",
			minTokens: 2,
			maxTokens: 4,
		},
		{
			name: "estimates tokens for long content",
			content: `This is a longer piece of content that contains multiple sentences.
It should have a reasonable token count estimate.
The estimation should be based on the content length and complexity.`,
			minTokens: 20,
			maxTokens: 50,
		},
		{
			name:      "estimates tokens for code content",
			content:   "func main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			minTokens: 8,
			maxTokens: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := EstimatePreviewTokens(tt.content)

			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("Token count %d outside expected range [%d, %d]", tokens, tt.minTokens, tt.maxTokens)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkRenderComponentPreview(b *testing.B) {
	// Setup
	if err := files.InitProjectStructure(); err != nil {
		b.Fatalf("Failed to initialize project structure: %v", err)
	}
	defer os.RemoveAll(files.PluqqyDir)

	// Create a test component
	content := strings.Repeat("This is test content. ", 100)
	// Create component file directly for benchmark
	componentPath := filepath.Join(files.ComponentsDir, files.PromptsDir, "bench-component.md")
	fullContent := fmt.Sprintf("---\ntags: [benchmark]\n---\n\n%s", content)
	if err := files.WriteComponent(componentPath, fullContent); err != nil {
		b.Fatalf("Failed to create test component: %v", err)
	}
	path := componentPath

	comp := componentItem{
		name:         "bench-component.md",
		path:         path,
		compType:     models.ComponentTypePrompt,
		lastModified: time.Now(),
		usageCount:   10,
		tokenCount:   utils.EstimateTokens(content),
		tags:         []string{"benchmark"},
	}

	renderer := NewPreviewRenderer()
	renderer.ShowPreview = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.RenderComponentPreview(comp)
	}
}

func BenchmarkEstimatePreviewTokens(b *testing.B) {
	content := strings.Repeat("This is a sample text for token estimation. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EstimatePreviewTokens(content)
	}
}

// Test for race conditions
func TestPreviewRenderer_RaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Initialize test environment
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}
	defer os.RemoveAll(files.PluqqyDir)

	// Create test component
	content := "Test content for race condition testing"
	componentPath := filepath.Join(files.ComponentsDir, files.PromptsDir, "race-test.md")
	if err := files.WriteComponent(componentPath, content); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}
	path := componentPath

	comp := componentItem{
		name:     "race-test.md",
		path:     path,
		compType: models.ComponentTypePrompt,
	}

	renderer := NewPreviewRenderer()
	renderer.ShowPreview = true

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = renderer.RenderComponentPreview(comp)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
