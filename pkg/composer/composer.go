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

	// Load all components and group by type
	for _, compRef := range sortedComponents {
		// Component paths in YAML are relative to the pipelines directory
		// We need to resolve them from the .pluqqy directory
		componentPath := filepath.Join(files.PipelinesDir, compRef.Path)
		componentPath = filepath.Clean(componentPath)
		
		component, err := files.ReadComponent(componentPath)
		if err != nil {
			return "", fmt.Errorf("failed to read component '%s' for pipeline '%s': %w", compRef.Path, pipeline.Name, err)
		}

		if _, exists := typeGroups[compRef.Type]; !exists {
			typeOrder = append(typeOrder, compRef.Type)
		}

		typeGroups[compRef.Type] = append(typeGroups[compRef.Type], componentWithContent{
			ref:     compRef,
			content: component.Content,
		})
	}

	// Write components grouped by type
	for _, componentType := range typeOrder {
		components := typeGroups[componentType]
		
		// Write type header if enabled in settings
		if settings.Output.Formatting.ShowHeadings {
			heading := getCustomHeading(componentType, settings)
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
	case models.ComponentTypePrompt, "prompts":
		return "PROMPTS"
	case models.ComponentTypeContext, "contexts":
		return "CONTEXT"
	case models.ComponentTypeRules:
		return "IMPORTANT RULES"
	default:
		// Uppercase the whole type
		return strings.ToUpper(componentType)
	}
}

func getCustomHeading(componentType string, settings *models.Settings) string {
	switch componentType {
	case models.ComponentTypeContext:
		return settings.Output.Formatting.Headings.Context
	case models.ComponentTypePrompt:
		return settings.Output.Formatting.Headings.Prompts
	case models.ComponentTypeRules:
		return settings.Output.Formatting.Headings.Rules
	default:
		// Fallback to default capitalization
		return fmt.Sprintf("## %s", capitalizeType(componentType))
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