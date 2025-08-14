package composer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func ComposePipeline(pipeline *models.Pipeline) (string, error) {
	if pipeline == nil {
		return "", fmt.Errorf("cannot compose pipeline: nil pipeline provided")
	}

	if len(pipeline.Components) == 0 {
		return "", fmt.Errorf("cannot compose pipeline '%s': no components defined", pipeline.Name)
	}

	// Load settings
	settings, err := files.ReadSettings()
	if err != nil {
		// Use defaults if settings can't be loaded
		settings = models.DefaultSettings()
	}

	// Sort components by order field
	sortedComponents := make([]models.ComponentRef, len(pipeline.Components))
	copy(sortedComponents, pipeline.Components)
	sort.Slice(sortedComponents, func(i, j int) bool {
		return sortedComponents[i].Order < sortedComponents[j].Order
	})

	// Group components by type while maintaining order
	var output strings.Builder
	output.WriteString(fmt.Sprintf("# %s\n\n", pipeline.Name))

	// Track which types we've seen and their components
	typeGroups := make(map[string][]componentWithContent)
	typeOrder := []string{}
	var missingComponents []string

	// Load all components and group by type
	for _, compRef := range sortedComponents {
		// Component paths in YAML are relative to the pipelines directory
		// We need to resolve them from the .pluqqy directory
		componentPath := filepath.Join(files.PipelinesDir, compRef.Path)
		componentPath = filepath.Clean(componentPath)
		
		component, err := files.ReadComponent(componentPath)
		if err != nil {
			// Track missing components instead of failing immediately
			missingComponents = append(missingComponents, compRef.Path)
			continue
		}

		if _, exists := typeGroups[compRef.Type]; !exists {
			typeOrder = append(typeOrder, compRef.Type)
		}

		typeGroups[compRef.Type] = append(typeGroups[compRef.Type], componentWithContent{
			ref:     compRef,
			content: component.Content,
		})
	}

	// If there are missing components, add a warning section
	if len(missingComponents) > 0 {
		output.WriteString("⚠️ **Warning: Missing Components**\n\n")
		output.WriteString("The following components could not be found:\n")
		for _, path := range missingComponents {
			output.WriteString(fmt.Sprintf("- %s\n", path))
		}
		output.WriteString("\nThese components may have been deleted or moved. Consider updating this pipeline.\n\n")
		output.WriteString("---\n\n")
	}

	// Write components grouped by type, ordered by settings.Sections
	// First write sections in the configured order
	for _, section := range settings.Output.Formatting.Sections {
		components, exists := typeGroups[section.Type]
		if !exists || len(components) == 0 {
			// Skip sections that have no components
			continue
		}
		
		// Write type header if enabled in settings
		if settings.Output.Formatting.ShowHeadings {
			output.WriteString(fmt.Sprintf("%s\n\n", section.Heading))
		}

		// Write components of this type
		for _, comp := range components {
			// Write content
			output.WriteString(strings.TrimSpace(comp.content))
			output.WriteString("\n\n")
		}

		output.WriteString("\n")
	}
	
	// Then write any remaining types not in Sections (for backwards compatibility)
	for _, componentType := range typeOrder {
		// Skip if already written
		written := false
		for _, section := range settings.Output.Formatting.Sections {
			if componentType == section.Type {
				written = true
				break
			}
		}
		if written {
			continue
		}
		
		components := typeGroups[componentType]
		// Write type header if enabled in settings
		if settings.Output.Formatting.ShowHeadings {
			// Use default heading for types not in sections config
			heading := fmt.Sprintf("## %s", capitalizeType(componentType))
			output.WriteString(fmt.Sprintf("%s\n\n", heading))
		}

		// Write components of this type
		for _, comp := range components {
			// Write content
			output.WriteString(strings.TrimSpace(comp.content))
			output.WriteString("\n\n")
		}

		output.WriteString("\n")
	}

	return output.String(), nil
}

type componentWithContent struct {
	ref     models.ComponentRef
	content string
}

func capitalizeType(componentType string) string {
	switch componentType {
	case models.ComponentTypePrompt:
		return "PROMPTS"
	case models.ComponentTypeContext:
		return "CONTEXT"
	case models.ComponentTypeRules:
		return "IMPORTANT RULES"
	default:
		// Uppercase the whole type
		return strings.ToUpper(componentType)
	}
}


// WritePLUQQYFile writes the composed pipeline to the output file
func WritePLUQQYFile(content string, outputPath string) error {
	// Load settings for default output path
	settings, err := files.ReadSettings()
	if err != nil {
		settings = models.DefaultSettings()
	}

	if outputPath == "" {
		outputPath = filepath.Join(settings.Output.ExportPath, settings.Output.DefaultFilename)
	}

	if err := files.WriteFile(outputPath, content); err != nil {
		return fmt.Errorf("failed to write composed pipeline to '%s': %w", outputPath, err)
	}

	return nil
}