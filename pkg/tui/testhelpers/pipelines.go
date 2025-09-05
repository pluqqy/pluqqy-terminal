package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"gopkg.in/yaml.v3"
)

// PipelineItem represents a pipeline in the TUI lists
// This mirrors the pipelineItem struct in the tui package
type PipelineItem struct {
	Name       string
	Path       string
	Tags       []string
	TokenCount int
	IsArchived bool
}

// PipelineBuilder provides a fluent interface for building test pipelines
type PipelineBuilder struct {
	pipeline *models.Pipeline
	item     PipelineItem
}

// NewPipelineBuilder creates a new pipeline builder with default values
func NewPipelineBuilder(name string) *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: &models.Pipeline{
			Name:       name,
			Path:       name + ".yaml",
			Tags:       []string{},
			Components: []models.ComponentRef{},
		},
		item: PipelineItem{
			Name:       name,
			Path:       "/test/pipelines/" + name + ".yaml",
			Tags:       []string{},
			TokenCount: 200,
			IsArchived: false,
		},
	}
}

// WithPath sets a custom path for the pipeline
func (b *PipelineBuilder) WithPath(path string) *PipelineBuilder {
	b.pipeline.Path = path
	b.item.Path = path
	return b
}

// WithComponents adds component references to the pipeline
func (b *PipelineBuilder) WithComponents(refs ...models.ComponentRef) *PipelineBuilder {
	b.pipeline.Components = append(b.pipeline.Components, refs...)
	// Update token count based on number of components
	b.item.TokenCount = 200 + len(b.pipeline.Components)*100
	return b
}

// WithTags sets tags for the pipeline
func (b *PipelineBuilder) WithTags(tags ...string) *PipelineBuilder {
	b.pipeline.Tags = tags
	b.item.Tags = tags
	return b
}

// WithTokenCount sets a specific token count
func (b *PipelineBuilder) WithTokenCount(count int) *PipelineBuilder {
	b.item.TokenCount = count
	return b
}

// Archived marks the pipeline as archived
func (b *PipelineBuilder) Archived() *PipelineBuilder {
	b.item.IsArchived = true
	return b
}

// Build returns the built pipeline model
func (b *PipelineBuilder) Build() *models.Pipeline {
	return b.pipeline
}

// BuildItem returns the built pipeline item for TUI lists
func (b *PipelineBuilder) BuildItem() PipelineItem {
	return b.item
}

// WriteToFile writes the pipeline to a file in the test environment
func (b *PipelineBuilder) WriteToFile(t *testing.T, tmpDir string) string {
	t.Helper()
	
	pipelinePath := filepath.Join(tmpDir, files.PluqqyDir, "pipelines", b.pipeline.Path)
	
	// Ensure directory exists
	dir := filepath.Dir(pipelinePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	
	// Marshal and write the pipeline
	data, err := yaml.Marshal(b.pipeline)
	if err != nil {
		t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		t.Fatalf("Failed to write pipeline file: %v", err)
	}
	
	return b.pipeline.Path
}

// Factory Functions for Common Cases

// MakeSimplePipeline creates a basic pipeline without components
func MakeSimplePipeline(name string) *models.Pipeline {
	return NewPipelineBuilder(name).Build()
}

// MakePipelineWithComponents creates a pipeline with the specified number of components
func MakePipelineWithComponents(name string, componentCount int) *models.Pipeline {
	builder := NewPipelineBuilder(name)
	
	for i := 0; i < componentCount; i++ {
		ref := models.ComponentRef{
			Type: models.ComponentTypePrompt,
			Path: "component-" + string(rune('a'+i)),
		}
		builder.WithComponents(ref)
	}
	
	return builder.Build()
}

// MakeTestPipeline creates a pipeline item for TUI testing
func MakeTestPipeline(name string) PipelineItem {
	return NewPipelineBuilder(name).BuildItem()
}

// MakeTestPipelines creates multiple pipeline items
func MakeTestPipelines(names ...string) []PipelineItem {
	pipelines := make([]PipelineItem, len(names))
	for i, name := range names {
		pipelines[i] = MakeTestPipeline(name)
	}
	return pipelines
}

// MakeTestPipelineWithTags creates a pipeline item with tags
func MakeTestPipelineWithTags(name string, tags []string) PipelineItem {
	return NewPipelineBuilder(name).
		WithTags(tags...).
		BuildItem()
}

// CreateTestPipeline creates and writes a pipeline to disk
func CreateTestPipeline(t *testing.T, tmpDir, name string, tags []string) string {
	t.Helper()
	
	pipeline := NewPipelineBuilder(name).
		WithTags(tags...).
		Build()
	
	pipelinePath := filepath.Join(tmpDir, files.PluqqyDir, "pipelines", pipeline.Path)
	
	// Ensure directory exists
	dir := filepath.Dir(pipelinePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		t.Fatalf("Failed to write pipeline: %v", err)
	}
	
	return pipeline.Path
}

// CreateTestPipelineFile creates a pipeline file with component references
func CreateTestPipelineFile(t *testing.T, tmpDir, name string, components []models.ComponentRef) string {
	t.Helper()
	
	pipeline := NewPipelineBuilder(name).
		WithComponents(components...).
		Build()
	
	pipelinePath := filepath.Join(tmpDir, files.PluqqyDir, "pipelines", pipeline.Path)
	
	// Ensure directory exists
	dir := filepath.Dir(pipelinePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		t.Fatalf("Failed to write pipeline: %v", err)
	}
	
	return pipelinePath
}

// GeneratePipelines creates multiple test pipelines with incrementing values
func GeneratePipelines(count int) []PipelineItem {
	pipelines := make([]PipelineItem, count)
	
	for i := 0; i < count; i++ {
		pipelines[i] = NewPipelineBuilder("pipeline-" + string(rune('a'+i))).
			WithTokenCount(200 * (i + 1)).
			BuildItem()
	}
	
	return pipelines
}