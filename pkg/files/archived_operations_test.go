package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

func TestWriteComponentToArchive(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)

	// Initialize project structure
	if err := InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}

	// Test writing a component to archive
	componentPath := "components/prompts/test-component.md"
	content := "# Test Component\n\nThis is a test"

	err := WriteComponentToArchive(componentPath, content)
	if err != nil {
		t.Fatalf("Failed to write component to archive: %v", err)
	}

	// Verify the file was created in the correct location
	expectedPath := filepath.Join(PluqqyDir, "archive", componentPath)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Component was not created at expected path: %s", expectedPath)
	}

	// Read the content back
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read archived component: %v", err)
	}

	if string(data) != content {
		t.Errorf("Content mismatch. Got: %s, Want: %s", string(data), content)
	}
}

func TestWritePipelineToArchive(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)

	// Initialize project structure
	if err := InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}

	// Create a test pipeline
	pipeline := &models.Pipeline{
		Name: "test-pipeline",
		Path: "test-pipeline.yaml",
		Components: []models.ComponentRef{
			{Path: "../components/prompts/test.md", Type: "prompt"},
		},
	}

	err := WritePipelineToArchive(pipeline)
	if err != nil {
		t.Fatalf("Failed to write pipeline to archive: %v", err)
	}

	// Verify the file was created in the correct location
	expectedPath := filepath.Join(PluqqyDir, "archive", PipelinesDir, "test-pipeline.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Pipeline was not created at expected path: %s", expectedPath)
	}

	// Read the pipeline back
	archivedPipeline, err := ReadArchivedPipeline("test-pipeline.yaml")
	if err != nil {
		t.Fatalf("Failed to read archived pipeline: %v", err)
	}

	if archivedPipeline.Name != pipeline.Name {
		t.Errorf("Pipeline name mismatch. Got: %s, Want: %s", archivedPipeline.Name, pipeline.Name)
	}
}

func TestDeleteArchivedComponent(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)

	// Initialize project structure
	if err := InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}

	// Create an archived component
	componentPath := "components/prompts/delete-test.md"
	content := "# Delete Test"
	err := WriteComponentToArchive(componentPath, content)
	if err != nil {
		t.Fatalf("Failed to write component to archive: %v", err)
	}

	// Delete the archived component
	err = DeleteArchivedComponent(componentPath)
	if err != nil {
		t.Fatalf("Failed to delete archived component: %v", err)
	}

	// Verify the file was deleted
	expectedPath := filepath.Join(PluqqyDir, "archive", componentPath)
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Errorf("Component should have been deleted: %s", expectedPath)
	}
}

func TestDeleteArchivedPipeline(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)

	// Initialize project structure
	if err := InitProjectStructure(); err != nil {
		t.Fatalf("Failed to init project structure: %v", err)
	}

	// Create an archived pipeline
	pipeline := &models.Pipeline{
		Name: "delete-test",
		Path: "delete-test.yaml",
	}
	err := WritePipelineToArchive(pipeline)
	if err != nil {
		t.Fatalf("Failed to write pipeline to archive: %v", err)
	}

	// Delete the archived pipeline
	err = DeleteArchivedPipeline("delete-test.yaml")
	if err != nil {
		t.Fatalf("Failed to delete archived pipeline: %v", err)
	}

	// Verify the file was deleted
	expectedPath := filepath.Join(PluqqyDir, "archive", PipelinesDir, "delete-test.yaml")
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Errorf("Pipeline should have been deleted: %s", expectedPath)
	}
}