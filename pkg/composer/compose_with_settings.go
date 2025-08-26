package composer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComposePipelineWithSettings composes a pipeline using provided settings
func ComposePipelineWithSettings(pipeline *models.Pipeline, settings *models.Settings) (string, error) {
	if pipeline == nil {
		return "", fmt.Errorf("cannot compose pipeline: nil pipeline provided")
	}

	if len(pipeline.Components) == 0 {
		return "", fmt.Errorf("cannot compose pipeline '%s': no components defined", pipeline.Name)
	}

	// Sort components by order field
	sortedComponents := make([]models.ComponentRef, len(pipeline.Components))
	copy(sortedComponents, pipeline.Components)
	sort.Slice(sortedComponents, func(i, j int) bool {
		return sortedComponents[i].Order < sortedComponents[j].Order
	})

	// Group components by type while maintaining order
	var output strings.Builder
	
	// Track which types we've seen and their components
	typeGroups := make(map[string][]componentWithContentSettings)
	var missingComponents []string

	// Load all components and group by type
	for _, compRef := range sortedComponents {
		// Component paths in YAML are relative to the pipelines directory
		// We need to resolve them from the .pluqqy directory
		componentPath := filepath.Join(files.PipelinesDir, compRef.Path)
		componentPath = filepath.Clean(componentPath)

		// Check if it's an archived component
		isArchived := strings.Contains(componentPath, "/archive/")
		
		component, err := files.ReadArchivedOrActiveComponent(componentPath, isArchived)
		if err != nil {
			missingComponents = append(missingComponents, compRef.Path)
			continue
		}

		typeGroups[component.Type] = append(typeGroups[component.Type], componentWithContentSettings{
			ref:     compRef,
			content: component.Content,
		})
	}

	// Build output according to settings
	for _, section := range settings.Output.Formatting.Sections {
		components, exists := typeGroups[strings.ToLower(section.Type)]
		if !exists || len(components) == 0 {
			continue
		}

		// Add section heading if enabled
		if settings.Output.Formatting.ShowHeadings {
			output.WriteString(fmt.Sprintf("%s\n\n", section.Heading))
		}

		// Add components
		for _, comp := range components {
			output.WriteString(comp.content)
			if !strings.HasSuffix(comp.content, "\n") {
				output.WriteString("\n")
			}
			output.WriteString("\n")
		}
	}

	// Add warning about missing components
	if len(missingComponents) > 0 {
		output.WriteString("\n---\n")
		output.WriteString("⚠️  Warning: The following components could not be loaded:\n")
		for _, path := range missingComponents {
			output.WriteString(fmt.Sprintf("   - %s\n", path))
		}
	}

	return output.String(), nil
}

// componentWithContentSettings is the version used with settings
type componentWithContentSettings struct {
	ref     models.ComponentRef
	content string
}