package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestInitProjectStructure(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	err := InitProjectStructure()
	if err != nil {
		t.Fatalf("InitProjectStructure failed: %v", err)
	}

	expectedDirs := []string{
		PluqqyDir,
		filepath.Join(PluqqyDir, PipelinesDir),
		filepath.Join(PluqqyDir, ComponentsDir),
		filepath.Join(PluqqyDir, ComponentsDir, PromptsDir),
		filepath.Join(PluqqyDir, ComponentsDir, ContextsDir),
		filepath.Join(PluqqyDir, ComponentsDir, RulesDir),
		filepath.Join(PluqqyDir, ArchiveDir),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s does not exist", dir)
		}
	}
}

func TestReadWriteComponent(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	err := InitProjectStructure()
	if err != nil {
		t.Fatalf("InitProjectStructure failed: %v", err)
	}

	componentPath := filepath.Join(ComponentsDir, PromptsDir, "test.md")
	content := "This is a test prompt"

	err = WriteComponent(componentPath, content)
	if err != nil {
		t.Fatalf("WriteComponent failed: %v", err)
	}

	component, err := ReadComponent(componentPath)
	if err != nil {
		t.Fatalf("ReadComponent failed: %v", err)
	}

	if component.Content != content {
		t.Errorf("Expected content %q, got %q", content, component.Content)
	}

	if component.Type != models.ComponentTypePrompt {
		t.Errorf("Expected type %q, got %q", models.ComponentTypePrompt, component.Type)
	}
}

func TestReadWritePipeline(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	err := InitProjectStructure()
	if err != nil {
		t.Fatalf("InitProjectStructure failed: %v", err)
	}

	pipeline := &models.Pipeline{
		Name: "test-pipeline",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypeContext, Path: "../components/contexts/test.md", Order: 1},
			{Type: models.ComponentTypePrompt, Path: "../components/prompts/test.md", Order: 2},
			{Type: models.ComponentTypeRules, Path: "../components/rules/test.md", Order: 3},
		},
	}

	err = WritePipeline(pipeline)
	if err != nil {
		t.Fatalf("WritePipeline failed: %v", err)
	}

	readPipeline, err := ReadPipeline("test-pipeline.yaml")
	if err != nil {
		t.Fatalf("ReadPipeline failed: %v", err)
	}

	if readPipeline.Name != pipeline.Name {
		t.Errorf("Expected name %q, got %q", pipeline.Name, readPipeline.Name)
	}

	if len(readPipeline.Components) != len(pipeline.Components) {
		t.Errorf("Expected %d components, got %d", len(pipeline.Components), len(readPipeline.Components))
	}
}

func TestListPipelines(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	err := InitProjectStructure()
	if err != nil {
		t.Fatalf("InitProjectStructure failed: %v", err)
	}

	pipeline1 := &models.Pipeline{
		Name: "pipeline1",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypePrompt, Path: "../components/prompts/test.md", Order: 1},
		},
	}
	pipeline2 := &models.Pipeline{
		Name: "pipeline2",
		Components: []models.ComponentRef{
			{Type: models.ComponentTypeContext, Path: "../components/contexts/test.md", Order: 1},
		},
	}
	
	WritePipeline(pipeline1)
	WritePipeline(pipeline2)

	pipelines, err := ListPipelines()
	if err != nil {
		t.Fatalf("ListPipelines failed: %v", err)
	}

	if len(pipelines) != 2 {
		t.Errorf("Expected 2 pipelines, got %d", len(pipelines))
	}
}

func TestListComponents(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	err := InitProjectStructure()
	if err != nil {
		t.Fatalf("InitProjectStructure failed: %v", err)
	}

	WriteComponent(filepath.Join(ComponentsDir, PromptsDir, "prompt1.md"), "prompt 1")
	WriteComponent(filepath.Join(ComponentsDir, PromptsDir, "prompt2.md"), "prompt 2")
	WriteComponent(filepath.Join(ComponentsDir, ContextsDir, "context1.md"), "context 1")

	prompts, err := ListComponents("prompts")
	if err != nil {
		t.Fatalf("ListComponents failed: %v", err)
	}

	if len(prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(prompts))
	}

	contexts, err := ListComponents("contexts")
	if err != nil {
		t.Fatalf("ListComponents failed: %v", err)
	}

	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}
}

func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	_, err := ReadComponent("nonexistent.md")
	if err == nil {
		t.Error("Expected error when reading nonexistent component")
	}

	_, err = ReadPipeline("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error when reading nonexistent pipeline")
	}

	_, err = ListComponents("invalid-type")
	if err == nil {
		t.Error("Expected error for invalid component type")
	}
}