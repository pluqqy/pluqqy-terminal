package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// UsageResult represents the output structure for usage command
type UsageResult struct {
	Component string           `json:"component" yaml:"component"`
	Path      string           `json:"path" yaml:"path"`
	Type      string           `json:"type" yaml:"type"`
	Usage     []PipelineUsage  `json:"usage" yaml:"usage"`
	Count     int              `json:"count" yaml:"count"`
}

// PipelineUsage represents a pipeline using the component
type PipelineUsage struct {
	Name            string `json:"name" yaml:"name"`
	Path            string `json:"path" yaml:"path"`
	Position        int    `json:"position" yaml:"position"`
	TotalComponents int    `json:"total_components" yaml:"total_components"`
	IsArchived      bool   `json:"is_archived,omitempty" yaml:"is_archived,omitempty"`
}

var (
	usageShowAll bool
)

// NewUsageCommand creates the usage command
func NewUsageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage <component-name>",
		Short: "Show which pipelines use a component",
		Long: `Show all pipelines that use a specific component.

You can specify the component in multiple ways:
- Just the name: 'greeting'
- With type prefix: 'prompts/greeting' or 'prompt/greeting'
- With .md extension: 'greeting.md'

Examples:
  # Show usage of a prompt (by name only)
  pluqqy usage greeting
  
  # Show usage with type prefix (useful for disambiguation)
  pluqqy usage prompts/greeting
  pluqqy usage contexts/api-docs
  pluqqy usage rules/coding-standards
  
  # Include archived pipelines
  pluqqy usage greeting --all
  
  # Output as JSON
  pluqqy usage contexts/api-docs -o json`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runUsage,
	}

	cmd.Flags().BoolVarP(&usageShowAll, "all", "a", false, "Include archived pipelines")

	return cmd
}

func runUsage(cmd *cobra.Command, args []string) error {
	componentName := args[0]
	
	// Find the component by name in all component directories
	componentPath, compType, err := findComponentByName(componentName)
	if err != nil {
		return err
	}
	
	// Find all pipelines using this component
	usage, err := findComponentUsage(componentPath, usageShowAll)
	if err != nil {
		return fmt.Errorf("failed to find component usage: %w", err)
	}
	
	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")
	
	result := UsageResult{
		Component: componentName,
		Path:      componentPath,
		Type:      compType,
		Usage:     usage,
		Count:     len(usage),
	}
	
	// Format and print output
	switch outputFormat {
	case "json", "yaml":
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, result)
	default:
		// Custom table format for usage
		return printUsageTable(cmd, result)
	}
}

func findComponentByName(componentName string) (string, string, error) {
	var searchTypes []string
	var searchName string
	
	// Check if the user specified a type prefix (e.g., "prompts/greeting" or "contexts/api-docs")
	if strings.Contains(componentName, "/") {
		parts := strings.SplitN(componentName, "/", 2)
		typePrefix := parts[0]
		searchName = parts[1]
		
		// Map singular or plural forms to the correct directory name
		switch typePrefix {
		case "prompt", "prompts":
			searchTypes = []string{"prompts"}
		case "context", "contexts":
			searchTypes = []string{"contexts"}
		case "rule", "rules":
			searchTypes = []string{"rules"}
		default:
			return "", "", fmt.Errorf("unknown component type: %s", typePrefix)
		}
	} else {
		// No type specified, search all types
		searchTypes = []string{"prompts", "contexts", "rules"}
		searchName = componentName
	}
	
	// Add .md extension if not present
	if !strings.HasSuffix(searchName, ".md") {
		searchName = searchName + ".md"
	}
	
	// Search in specified component types
	for _, compType := range searchTypes {
		// Check regular components
		componentDir := filepath.Join(files.PluqqyDir, "components", compType)
		fullPath := filepath.Join(componentDir, searchName)
		
		if _, err := os.Stat(fullPath); err == nil {
			// Found it!
			relativePath := filepath.Join("components", compType, searchName)
			typeSingular := strings.TrimSuffix(compType, "s")
			return relativePath, typeSingular, nil
		}
		
		// Check archived components if --all flag is set
		if usageShowAll {
			archivedDir := filepath.Join(files.PluqqyDir, "archive", "components", compType)
			archivedPath := filepath.Join(archivedDir, searchName)
			
			if _, err := os.Stat(archivedPath); err == nil {
				// Found in archive
				relativePath := filepath.Join("archive", "components", compType, searchName)
				typeSingular := strings.TrimSuffix(compType, "s")
				return relativePath, typeSingular, nil
			}
		}
	}
	
	// Build error message based on what was searched
	if len(searchTypes) == 1 {
		return "", "", fmt.Errorf("component not found: %s in %s", searchName, searchTypes[0])
	}
	return "", "", fmt.Errorf("component not found: %s", componentName)
}

func getComponentType(path string) string {
	if strings.Contains(path, "prompts") {
		return "prompt"
	} else if strings.Contains(path, "contexts") {
		return "context"
	} else if strings.Contains(path, "rules") {
		return "rule"
	}
	return "unknown"
}

func findComponentUsage(componentPath string, includeArchived bool) ([]PipelineUsage, error) {
	var usage []PipelineUsage
	
	// Get all pipelines
	pipelines, err := files.ListPipelines()
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	
	// Check archived pipelines if requested
	if includeArchived {
		archivedPipelines, err := files.ListArchivedPipelines()
		if err == nil {
			pipelines = append(pipelines, archivedPipelines...)
		}
	}
	
	// Check each pipeline
	for _, pipelinePath := range pipelines {
		// Determine if this is an archived pipeline
		// Now that ListArchivedPipelines returns paths with archive/ prefix, this will work correctly
		isArchived := strings.HasPrefix(pipelinePath, files.ArchiveDir+"/")
		
		// Read the pipeline (both functions now handle the new path format)
		var pipeline *models.Pipeline
		var err error
		
		if isArchived {
			pipeline, err = files.ReadArchivedPipeline(pipelinePath)
		} else {
			pipeline, err = files.ReadPipeline(pipelinePath)
		}
		
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline uses the component
		for _, comp := range pipeline.Components {
			// Normalize both paths - remove any ../ prefix and clean
			normalizedCompPath := filepath.Clean(comp.Path)
			if strings.HasPrefix(normalizedCompPath, "../") {
				normalizedCompPath = strings.TrimPrefix(normalizedCompPath, "../")
			}
			if normalizedCompPath == componentPath {
				usage = append(usage, PipelineUsage{
					Name:            pipeline.Name,
					Path:            pipelinePath,
					Position:        comp.Order,
					TotalComponents: len(pipeline.Components),
					IsArchived:      isArchived,
				})
				break
			}
		}
	}
	
	return usage, nil
}

func printUsageTable(cmd *cobra.Command, result UsageResult) error {
	out := cmd.OutOrStdout()
	
	// Print header
	fmt.Fprintf(out, "Component: %s\n", result.Component)
	fmt.Fprintf(out, "Type: %s\n", result.Type)
	fmt.Fprintf(out, "Path: %s\n", result.Path)
	fmt.Fprintln(out)
	
	if result.Count == 0 {
		fmt.Fprintln(out, "No pipelines are using this component.")
		return nil
	}
	
	fmt.Fprintf(out, "Used in %d pipeline(s):\n\n", result.Count)
	
	// Print usage table
	for _, usage := range result.Usage {
		status := ""
		if usage.IsArchived {
			status = " (archived)"
		}
		
		fmt.Fprintf(out, "  • %s%s\n", usage.Name, status)
		fmt.Fprintf(out, "    Position: %d of %d\n", usage.Position, usage.TotalComponents)
		fmt.Fprintf(out, "    Path: %s\n", usage.Path)
		fmt.Fprintln(out)
	}
	
	return nil
}