package tui

import (
	"testing"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ExamplePipelineBuilderConfig demonstrates how to use custom configuration
// for the Pipeline Builder to control editor preferences
func ExamplePipelineBuilderConfig() {
	// Example 1: Create a Pipeline Builder with default configuration
	// This enables the enhanced editor by default
	defaultBuilder := NewPipelineBuilderModel()
	_ = defaultBuilder // Use the builder...

	// Example 2: Create a Pipeline Builder without preview by default
	noPreviewConfig := &PipelineBuilderConfig{
		ShowPreviewByDefault: false, // Hide preview initially
	}
	noPreviewBuilder := NewPipelineBuilderModelWithConfig(noPreviewConfig)
	_ = noPreviewBuilder // Use the builder...

	// Example 3: Load an existing pipeline with custom configuration
	existingPipeline := &models.Pipeline{
		Name: "My Pipeline",
		Components: []models.ComponentRef{
			{Path: "../prompts/example.md", Order: 1},
		},
	}
	customConfig := &PipelineBuilderConfig{
		ShowPreviewByDefault: true,
	}
	loadedBuilder := LoadPipelineWithConfig(existingPipeline, customConfig)
	_ = loadedBuilder // Use the builder...
}

// TestPipelineBuilderConfig_Defaults verifies default configuration values
func TestPipelineBuilderConfig_Defaults(t *testing.T) {
	config := DefaultPipelineBuilderConfig()

	if !config.ShowPreviewByDefault {
		t.Error("Expected ShowPreviewByDefault to be true by default")
	}
}

// TestPipelineBuilderConfig_CustomSettings verifies custom configuration is applied
func TestPipelineBuilderConfig_CustomSettings(t *testing.T) {
	// Test with custom configuration
	customConfig := &PipelineBuilderConfig{
		ShowPreviewByDefault: false,
	}

	m := NewPipelineBuilderModelWithConfig(customConfig)

	if m.ui.ShowPreview {
		t.Error("Expected showPreview to be false with custom config")
	}
}

// TestPipelineBuilderConfig_NilConfig verifies nil config uses defaults
func TestPipelineBuilderConfig_NilConfig(t *testing.T) {
	m := NewPipelineBuilderModelWithConfig(nil)

	if !m.ui.ShowPreview {
		t.Error("Expected showPreview to be true with nil config")
	}
}

// TestLoadPipelineWithConfig verifies loading existing pipeline with config
func TestLoadPipelineWithConfig(t *testing.T) {
	pipeline := &models.Pipeline{
		Name: "Test Pipeline",
		Components: []models.ComponentRef{
			{Path: "../prompts/test1.md", Order: 1},
			{Path: "../contexts/test2.md", Order: 2},
		},
	}

	config := &PipelineBuilderConfig{
		ShowPreviewByDefault: true,
	}

	m := LoadPipelineWithConfig(pipeline, config)

	// Verify pipeline data was loaded
	if m.data.Pipeline.Name != "Test Pipeline" {
		t.Errorf("Expected pipeline name 'Test Pipeline', got %s", m.data.Pipeline.Name)
	}

	if len(m.data.SelectedComponents) != 2 {
		t.Errorf("Expected 2 selected components, got %d", len(m.data.SelectedComponents))
	}

	if len(m.data.OriginalComponents) != 2 {
		t.Errorf("Expected 2 original components, got %d", len(m.data.OriginalComponents))
	}

	// Verify we're not in name editing mode for existing pipeline
	if m.editors.EditingName {
		t.Error("Expected editingName to be false for existing pipeline")
	}
}
