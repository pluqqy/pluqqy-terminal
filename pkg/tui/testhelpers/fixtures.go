package testhelpers

import (
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// Common test data constants
var (
	// Sample component content
	SamplePromptContent  = "Generate a summary of the following text:"
	SampleContextContent = "You are a helpful assistant specialized in data analysis."
	SampleRulesContent   = "Always be polite and professional. Provide accurate information."
	
	// Sample component names
	DefaultPromptNames  = []string{"summarize", "analyze", "translate"}
	DefaultContextNames = []string{"assistant", "expert", "advisor"}
	DefaultRulesNames   = []string{"guidelines", "policies", "standards"}
	
	// Sample tags
	DefaultTags   = []string{"test", "sample"}
	ExtendedTags  = []string{"test", "sample", "qa", "production", "staging"}
	CategoryTags  = []string{"nlp", "ml", "data", "api", "ui"}
	PriorityTags  = []string{"high", "medium", "low"}
	
	// Sample colors for tags
	TagColors = []string{"#3498db", "#e74c3c", "#2ecc71", "#f39c12", "#9b59b6", "#1abc9c", "#34495e"}
	
	// Token counts
	SmallTokenCount  = 100
	MediumTokenCount = 500
	LargeTokenCount  = 2000
	
	// Usage counts
	LowUsageCount  = 1
	MedUsageCount  = 10
	HighUsageCount = 50
)

// TestData provides structured test data sets
type TestData struct {
	Components []ComponentItem
	Pipelines  []PipelineItem
	Tags       []string
}

// GetSmallTestData returns a small set of test data
func GetSmallTestData() TestData {
	return TestData{
		Components: []ComponentItem{
			MakePromptComponent("summarize", SmallTokenCount),
			MakeContextComponent("assistant", SmallTokenCount),
			MakeRulesComponent("guidelines", SmallTokenCount),
		},
		Pipelines: []PipelineItem{
			MakeTestPipeline("basic-pipeline"),
			MakeTestPipeline("advanced-pipeline"),
		},
		Tags: DefaultTags,
	}
}

// GetMediumTestData returns a medium set of test data
func GetMediumTestData() TestData {
	components := []ComponentItem{}
	
	// Add prompts
	for _, name := range DefaultPromptNames {
		components = append(components, 
			NewComponentBuilder(name).
				WithType(models.ComponentTypePrompt).
				WithTokens(MediumTokenCount).
				WithTags(DefaultTags...).
				WithUsageCount(MedUsageCount).
				Build())
	}
	
	// Add contexts
	for _, name := range DefaultContextNames {
		components = append(components, 
			NewComponentBuilder(name).
				WithType(models.ComponentTypeContext).
				WithTokens(MediumTokenCount).
				WithTags(CategoryTags[:2]...).
				WithUsageCount(MedUsageCount).
				Build())
	}
	
	// Add rules
	for _, name := range DefaultRulesNames {
		components = append(components, 
			NewComponentBuilder(name).
				WithType(models.ComponentTypeRules).
				WithTokens(SmallTokenCount).
				WithTags(PriorityTags[:1]...).
				WithUsageCount(LowUsageCount).
				Build())
	}
	
	pipelines := []PipelineItem{}
	pipelineNames := []string{"data-processing", "text-analysis", "api-gateway", "ml-pipeline", "ui-builder"}
	for _, name := range pipelineNames {
		pipelines = append(pipelines, 
			NewPipelineBuilder(name).
				WithTags(ExtendedTags[:3]...).
				WithTokenCount(MediumTokenCount * 2).
				BuildItem())
	}
	
	return TestData{
		Components: components,
		Pipelines:  pipelines,
		Tags:       ExtendedTags,
	}
}

// GetLargeTestData returns a large set of test data
func GetLargeTestData() TestData {
	components := GenerateComponents(20, models.ComponentTypePrompt)
	components = append(components, GenerateComponents(15, models.ComponentTypeContext)...)
	components = append(components, GenerateComponents(10, models.ComponentTypeRules)...)
	
	pipelines := GeneratePipelines(25)
	
	return TestData{
		Components: components,
		Pipelines:  pipelines,
		Tags:       append(ExtendedTags, CategoryTags...),
	}
}

// GetArchivedTestData returns test data with archived items
func GetArchivedTestData() TestData {
	components := []ComponentItem{
		MakeArchivedComponent("old-prompt", models.ComponentTypePrompt),
		MakeArchivedComponent("deprecated-context", models.ComponentTypeContext),
		MakeArchivedComponent("legacy-rules", models.ComponentTypeRules),
		MakePromptComponent("active-prompt", SmallTokenCount),
	}
	
	pipelines := []PipelineItem{
		NewPipelineBuilder("archived-pipeline").Archived().BuildItem(),
		NewPipelineBuilder("old-pipeline").Archived().BuildItem(),
		MakeTestPipeline("active-pipeline"),
	}
	
	return TestData{
		Components: components,
		Pipelines:  pipelines,
		Tags:       DefaultTags,
	}
}

// GetTaggedTestData returns test data with extensive tagging
func GetTaggedTestData() TestData {
	components := []ComponentItem{
		NewComponentBuilder("multi-tag-prompt").
			WithType(models.ComponentTypePrompt).
			WithTags(ExtendedTags...).
			Build(),
		NewComponentBuilder("category-context").
			WithType(models.ComponentTypeContext).
			WithTags(CategoryTags...).
			Build(),
		NewComponentBuilder("priority-rules").
			WithType(models.ComponentTypeRules).
			WithTags(PriorityTags...).
			Build(),
	}
	
	pipelines := []PipelineItem{
		NewPipelineBuilder("tagged-pipeline").
			WithTags(append(DefaultTags, CategoryTags...)...).
			BuildItem(),
	}
	
	return TestData{
		Components: components,
		Pipelines:  pipelines,
		Tags:       append(append(ExtendedTags, CategoryTags...), PriorityTags...),
	}
}

// CreateTestComponentRefs creates a slice of component references
func CreateTestComponentRefs(count int) []models.ComponentRef {
	refs := make([]models.ComponentRef, count)
	types := []string{models.ComponentTypePrompt, models.ComponentTypeContext, models.ComponentTypeRules}
	
	for i := 0; i < count; i++ {
		refs[i] = models.ComponentRef{
			Type: types[i%len(types)],
			Path: "component-" + string(rune('a'+i)),
		}
	}
	return refs
}

// CreateTestSections creates test sections for settings
func CreateTestSections(types ...string) []models.Section {
	sections := make([]models.Section, len(types))
	for i, typ := range types {
		sections[i] = models.Section{Type: typ}
	}
	return sections
}

// CreateTestTagRegistry creates a test tag registry
func CreateTestTagRegistry(tags []string) *models.TagRegistry {
	registry := &models.TagRegistry{
		Tags: make([]models.Tag, len(tags)),
	}
	
	for i, tag := range tags {
		registry.Tags[i] = models.Tag{
			Name:  tag,
			Color: TagColors[i%len(TagColors)],
		}
	}
	
	return registry
}

// SamplePipelineYAML returns sample YAML content for a pipeline
func SamplePipelineYAML() string {
	return `name: sample-pipeline
path: sample-pipeline.yaml
tags: [test, sample]
components:
  - type: prompts
    name: summarize
  - type: contexts
    name: assistant
  - type: rules
    name: guidelines
`
}

// SampleComponentMarkdown returns sample Markdown content for a component
func SampleComponentMarkdown(tags []string) string {
	content := ""
	if len(tags) > 0 {
		content = "---\ntags: ["
		for i, tag := range tags {
			if i > 0 {
				content += ", "
			}
			content += tag
		}
		content += "]\n---\n"
	}
	content += `# Component Documentation

This is a sample component for testing purposes.

## Features
- Feature 1
- Feature 2
- Feature 3

## Usage
Use this component when you need to test functionality.
`
	return content
}