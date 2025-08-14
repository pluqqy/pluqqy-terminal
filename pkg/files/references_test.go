package files

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"gopkg.in/yaml.v3"
)

func TestRemoveComponentReferences(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pluqqy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)
	
	// Create necessary directories
	dirs := []string{
		filepath.Join(PluqqyDir, ComponentsDir),
		filepath.Join(PluqqyDir, ComponentsDir, "contexts"),
		filepath.Join(PluqqyDir, PipelinesDir),
		filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create test components
	comp1Path := filepath.Join(ComponentsDir, "contexts", "auth.md")
	comp1Content := `---
name: Auth Context
---
# Auth Context
Authentication context for the system.`
	
	comp2Path := filepath.Join(ComponentsDir, "contexts", "database.md")
	comp2Content := `---
name: Database Context
---
# Database Context
Database context for the system.`
	
	// Write components
	if err := os.WriteFile(filepath.Join(PluqqyDir, comp1Path), []byte(comp1Content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(PluqqyDir, comp2Path), []byte(comp2Content), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create active pipeline that references both components
	activePipeline := &models.Pipeline{
		Name: "Test Pipeline",
		Components: []models.ComponentRef{
			{
				Path:  "../" + comp1Path,
				Type:  "context",
				Order: 1,
			},
			{
				Path:  "../" + comp2Path,
				Type:  "context",
				Order: 2,
			},
		},
	}
	
	// Write active pipeline
	activePipelineData, _ := yaml.Marshal(activePipeline)
	activePipelinePath := filepath.Join(PluqqyDir, PipelinesDir, "test-pipeline.yaml")
	if err := os.WriteFile(activePipelinePath, activePipelineData, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create archived pipeline that references both components
	archivedPipeline := &models.Pipeline{
		Name: "Archived Pipeline",
		Components: []models.ComponentRef{
			{
				Path:  "../" + comp1Path,
				Type:  "context",
				Order: 1,
			},
			{
				Path:  "../" + comp2Path,
				Type:  "context",
				Order: 2,
			},
		},
	}
	
	// Write archived pipeline
	archivedPipelineData, _ := yaml.Marshal(archivedPipeline)
	archivedPipelinePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, "archived-pipeline.yaml")
	if err := os.WriteFile(archivedPipelinePath, archivedPipelineData, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test removing references to comp1
	t.Run("Remove component references", func(t *testing.T) {
		// Remove references to comp1
		if err := RemoveComponentReferences(comp1Path); err != nil {
			t.Errorf("RemoveComponentReferences failed: %v", err)
		}
		
		// Check active pipeline
		updatedActivePipeline, err := ReadPipeline("test-pipeline.yaml")
		if err != nil {
			t.Errorf("Failed to read updated active pipeline: %v", err)
		}
		
		// Should only have comp2 now
		if len(updatedActivePipeline.Components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(updatedActivePipeline.Components))
		}
		
		if updatedActivePipeline.Components[0].Path != "../"+comp2Path {
			t.Errorf("Expected component path %s, got %s", "../"+comp2Path, updatedActivePipeline.Components[0].Path)
		}
		
		// Check archived pipeline
		updatedArchivedPipeline, err := ReadArchivedPipeline("archived-pipeline.yaml")
		if err != nil {
			t.Errorf("Failed to read updated archived pipeline: %v", err)
		}
		
		// Should only have comp2 now
		if len(updatedArchivedPipeline.Components) != 1 {
			t.Errorf("Expected 1 component in archived pipeline, got %d", len(updatedArchivedPipeline.Components))
		}
		
		if updatedArchivedPipeline.Components[0].Path != "../"+comp2Path {
			t.Errorf("Expected component path %s in archived pipeline, got %s", "../"+comp2Path, updatedArchivedPipeline.Components[0].Path)
		}
	})
	
	// Test removing all components
	t.Run("Remove all component references", func(t *testing.T) {
		// Remove references to comp2
		if err := RemoveComponentReferences(comp2Path); err != nil {
			t.Errorf("RemoveComponentReferences failed: %v", err)
		}
		
		// Check active pipeline
		updatedActivePipeline, err := ReadPipeline("test-pipeline.yaml")
		if err != nil {
			t.Errorf("Failed to read updated active pipeline: %v", err)
		}
		
		// Should have no components now
		if len(updatedActivePipeline.Components) != 0 {
			t.Errorf("Expected 0 components, got %d", len(updatedActivePipeline.Components))
		}
		
		// Check archived pipeline
		updatedArchivedPipeline, err := ReadArchivedPipeline("archived-pipeline.yaml")
		if err != nil {
			t.Errorf("Failed to read updated archived pipeline: %v", err)
		}
		
		// Should have no components now
		if len(updatedArchivedPipeline.Components) != 0 {
			t.Errorf("Expected 0 components in archived pipeline, got %d", len(updatedArchivedPipeline.Components))
		}
	})
}

func TestDeleteComponentWithReferences(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "pluqqy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)
	
	// Create necessary directories
	dirs := []string{
		filepath.Join(PluqqyDir, ComponentsDir),
		filepath.Join(PluqqyDir, ComponentsDir, "contexts"),
		filepath.Join(PluqqyDir, PipelinesDir),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create test component
	compPath := filepath.Join(ComponentsDir, "contexts", "test.md")
	compContent := `---
name: Test Context
---
# Test Context
Test context content.`
	
	// Write component
	if err := os.WriteFile(filepath.Join(PluqqyDir, compPath), []byte(compContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create pipeline that references the component
	pipeline := &models.Pipeline{
		Name: "Test Pipeline",
		Components: []models.ComponentRef{
			{
				Path:  "../" + compPath,
				Type:  "context",
				Order: 1,
			},
		},
	}
	
	// Write pipeline
	pipelineData, _ := yaml.Marshal(pipeline)
	pipelinePath := filepath.Join(PluqqyDir, PipelinesDir, "test-pipeline.yaml")
	if err := os.WriteFile(pipelinePath, pipelineData, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test deleting the component
	t.Run("Delete component with references", func(t *testing.T) {
		// Delete the component
		if err := DeleteComponent(compPath); err != nil {
			t.Errorf("DeleteComponent failed: %v", err)
		}
		
		// Check that component file is deleted
		if _, err := os.Stat(filepath.Join(PluqqyDir, compPath)); !os.IsNotExist(err) {
			t.Error("Component file should be deleted")
		}
		
		// Check that pipeline is updated
		updatedPipeline, err := ReadPipeline("test-pipeline.yaml")
		if err != nil {
			t.Errorf("Failed to read updated pipeline: %v", err)
		}
		
		// Should have no components now
		if len(updatedPipeline.Components) != 0 {
			t.Errorf("Expected 0 components after deletion, got %d", len(updatedPipeline.Components))
		}
	})
}