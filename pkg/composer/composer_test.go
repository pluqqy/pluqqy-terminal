package composer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestComposePipeline(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}

	// Create test components
	promptContent := "Please help me debug this issue"
	contextContent := "The system is running on Linux"
	rulesContent := "Be concise and technical"

	files.WriteComponent(filepath.Join(files.ComponentsDir, files.PromptsDir, "debug.md"), promptContent)
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.ContextsDir, "system.md"), contextContent)
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.RulesDir, "technical.md"), rulesContent)

	// Create test pipeline
	pipeline := &models.Pipeline{
		Name: "test-debug",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypeContext, Path: "../components/contexts/system.md", Order: 1},
			{Type: models.ComponentTypePrompt, Path: "../components/prompts/debug.md", Order: 2},
			{Type: models.ComponentTypeRules, Path: "../components/rules/technical.md", Order: 3},
		},
	}

	// Test composition
	output, err := ComposePipeline(pipeline)
	if err != nil {
		t.Fatalf("ComposePipeline failed: %v", err)
	}

	// Verify output contains expected elements
	expectedElements := []string{
		"# test-debug",
		"## CONTEXT",
		"## PROMPTS",
		"## IMPORTANT RULES",
		contextContent,
		promptContent,
		rulesContent,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected element: %s", expected)
		}
	}
}

func TestComposePipelineWithMultipleComponents(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := files.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}

	// Create multiple components of same type
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.PromptsDir, "step1.md"), "First step")
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.PromptsDir, "step2.md"), "Second step")
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.ContextsDir, "context1.md"), "Context 1")
	files.WriteComponent(filepath.Join(files.ComponentsDir, files.ContextsDir, "context2.md"), "Context 2")

	// Create pipeline with multiple components per type
	pipeline := &models.Pipeline{
		Name: "multi-component",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypeContext, Path: "../components/contexts/context1.md", Order: 1},
			{Type: models.ComponentTypeContext, Path: "../components/contexts/context2.md", Order: 2},
			{Type: models.ComponentTypePrompt, Path: "../components/prompts/step1.md", Order: 3},
			{Type: models.ComponentTypePrompt, Path: "../components/prompts/step2.md", Order: 4},
		},
	}

	output, err := ComposePipeline(pipeline)
	if err != nil {
		t.Fatalf("ComposePipeline failed: %v", err)
	}

	// Verify NO separators between same-type components
	if strings.Contains(output, "---") {
		t.Error("Output should not contain separators between components")
	}

	// Verify order is maintained
	context1Pos := strings.Index(output, "Context 1")
	context2Pos := strings.Index(output, "Context 2")
	prompt1Pos := strings.Index(output, "First step")
	prompt2Pos := strings.Index(output, "Second step")

	if context1Pos > context2Pos || context2Pos > prompt1Pos || prompt1Pos > prompt2Pos {
		t.Error("Components not in correct order")
	}
}

func TestComposePipelineErrors(t *testing.T) {
	// Test nil pipeline
	_, err := ComposePipeline(nil)
	if err == nil {
		t.Error("Expected error for nil pipeline")
	}

	// Test empty components
	pipeline := &models.Pipeline{Name: "empty"}
	_, err = ComposePipeline(pipeline)
	if err == nil {
		t.Error("Expected error for pipeline with no components")
	}

	// Test missing component file - should now succeed with warning
	pipeline = &models.Pipeline{
		Name: "missing-component",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypePrompt, Path: "nonexistent.md", Order: 1},
		},
	}
	output, err := ComposePipeline(pipeline)
	if err != nil {
		t.Errorf("Should not error for missing component file, got: %v", err)
	}
	if !strings.Contains(output, "Warning: Missing Components") {
		t.Error("Expected warning about missing components in output")
	}
	if !strings.Contains(output, "nonexistent.md") {
		t.Error("Expected missing component path in warning")
	}
}

func TestWritePLUQQYFile(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	content := "# Test PLUQQY Content"
	
	// Test with default filename
	err := WritePLUQQYFile(content, "")
	if err != nil {
		t.Fatalf("WritePLUQQYFile failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(files.DefaultOutputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(data) != content {
		t.Error("Output file content doesn't match")
	}

	// Test with custom filename
	customPath := "custom-output.md"
	err = WritePLUQQYFile(content, customPath)
	if err != nil {
		t.Fatalf("WritePLUQQYFile with custom path failed: %v", err)
	}

	// Verify custom file was created
	_, err = os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read custom output file: %v", err)
	}
}