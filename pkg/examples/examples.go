package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// ExampleSet represents a collection of related examples
type ExampleSet struct {
	Category    string
	Name        string
	Description string
	Components  []ExampleComponent
	Pipelines   []ExamplePipeline
}

// ExampleComponent represents an example component template
type ExampleComponent struct {
	Name        string
	Filename    string
	Type        string // contexts, prompts, or rules
	Content     string
	Tags        []string
}

// ExamplePipeline represents an example pipeline configuration
type ExamplePipeline struct {
	Name        string
	Filename    string
	Description string
	Components  []models.ComponentRef
	Tags        []string
}

// GetExamples returns example sets for the given category
func GetExamples(category string) []ExampleSet {
	switch category {
	case "general":
		sets := getGeneralExamples()
		for i := range sets {
			sets[i].Category = "general"
		}
		return sets
	case "web":
		sets := getWebExamples()
		for i := range sets {
			sets[i].Category = "web"
		}
		return sets
	case "ai":
		sets := getAIExamples()
		for i := range sets {
			sets[i].Category = "ai"
		}
		return sets
	case "claude":
		sets := getClaudeExamples()
		for i := range sets {
			sets[i].Category = "claude"
		}
		return sets
	case "all":
		var all []ExampleSet
		
		general := getGeneralExamples()
		for i := range general {
			general[i].Category = "general"
		}
		all = append(all, general...)
		
		web := getWebExamples()
		for i := range web {
			web[i].Category = "web"
		}
		all = append(all, web...)
		
		ai := getAIExamples()
		for i := range ai {
			ai[i].Category = "ai"
		}
		all = append(all, ai...)
		
		claude := getClaudeExamples()
		for i := range claude {
			claude[i].Category = "claude"
		}
		all = append(all, claude...)
		
		return all
	default:
		return []ExampleSet{}
	}
}

// InstallComponent installs a component example to the user's .pluqqy directory
func InstallComponent(comp ExampleComponent, force bool) (bool, error) {
	// Build the file path
	path := filepath.Join("components", comp.Type, comp.Filename)
	fullPath := filepath.Join(files.PluqqyDir, path)
	
	// Check if file exists
	if !force {
		if _, err := os.Stat(fullPath); err == nil {
			return false, fmt.Errorf("component already exists at %s", path)
		}
	}
	
	// Format content with frontmatter
	content := formatComponentContent(comp)
	
	// Write the component
	if err := files.WriteComponent(path, content); err != nil {
		return false, err
	}
	
	return true, nil
}

// InstallPipeline installs a pipeline example to the user's .pluqqy directory
func InstallPipeline(pipeline ExamplePipeline, force bool) (bool, error) {
	// Build the file path
	path := filepath.Join(files.PluqqyDir, files.PipelinesDir, pipeline.Filename)
	
	// Check if file exists
	if !force {
		if _, err := os.Stat(path); err == nil {
			return false, fmt.Errorf("pipeline already exists at %s", pipeline.Filename)
		}
	}
	
	// Create pipeline model
	p := &models.Pipeline{
		Name:       pipeline.Name,
		Path:       pipeline.Filename,
		Components: pipeline.Components,
		Tags:       pipeline.Tags,
	}
	
	// Write the pipeline
	if err := files.WritePipeline(p); err != nil {
		return false, err
	}
	
	return true, nil
}

// formatComponentContent formats a component with YAML frontmatter
func formatComponentContent(comp ExampleComponent) string {
	var content strings.Builder
	
	// Add frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("name: %s\n", comp.Name))
	if len(comp.Tags) > 0 {
		content.WriteString("tags:\n")
		for _, tag := range comp.Tags {
			content.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}
	content.WriteString("---\n")
	
	// Add content
	content.WriteString(comp.Content)
	
	return content.String()
}