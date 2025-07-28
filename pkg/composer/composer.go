package composer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/user/pluqqy/pkg/files"
	"github.com/user/pluqqy/pkg/models"
)

func ComposePipeline(pipeline *models.Pipeline) (string, error) {
	if pipeline == nil {
		return "", fmt.Errorf("pipeline is nil")
	}

	if len(pipeline.Components) == 0 {
		return "", fmt.Errorf("pipeline has no components")
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
			return "", fmt.Errorf("failed to read component %s: %w", compRef.Path, err)
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
		
		// Write type header
		output.WriteString(fmt.Sprintf("## %s\n\n", capitalizeType(componentType)))

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
	case "prompt", "prompts":
		return "PROMPTS"
	case "context", "contexts":
		return "CONTEXT"
	case "rules":
		return "IMPORTANT RULES"
	default:
		// Uppercase the whole type
		return strings.ToUpper(componentType)
	}
}

// WritePLUQQYFile writes the composed pipeline to the output file
func WritePLUQQYFile(content string, outputPath string) error {
	if outputPath == "" {
		outputPath = files.DefaultOutputFile
	}

	if err := files.WriteFile(outputPath, content); err != nil {
		return fmt.Errorf("failed to write PLUQQY.md: %w", err)
	}

	return nil
}