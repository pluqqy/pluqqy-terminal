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

// ListResult represents the output structure for list command
type ListResult struct {
	Type  string      `json:"type" yaml:"type"`
	Items []ListItem  `json:"items" yaml:"items"`
	Count int         `json:"count" yaml:"count"`
}

// ListItem represents a single item in the list
type ListItem struct {
	Name        string   `json:"name" yaml:"name"`
	Filename    string   `json:"filename,omitempty" yaml:"filename,omitempty"`
	Type        string   `json:"type" yaml:"type"`
	Tags        []string `json:"tags" yaml:"tags"`
	Path        string   `json:"path,omitempty" yaml:"path,omitempty"`
	Components  int      `json:"components,omitempty" yaml:"components,omitempty"`
	IsArchived  bool     `json:"is_archived,omitempty" yaml:"is_archived,omitempty"`
}

var (
	listShowArchived bool
	listShowPaths    bool
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [type]",
		Short: "List pipelines and components",
		Long: `List all pipelines and components in the current project.

Types:
  pipelines   - List only pipelines
  components  - List all components  
  contexts    - List only context components
  prompts     - List only prompt components
  rules       - List only rule components
  all         - List everything (default)

Examples:
  # List all items
  pluqqy list
  
  # List only pipelines
  pluqqy list pipelines
  
  # List only rules
  pluqqy list rules
  
  # List components with JSON output
  pluqqy list components -o json
  
  # Show only archived items
  pluqqy list --archived
  
  # Show file paths
  pluqqy list --paths`,
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"pipelines", "components", "contexts", "prompts", "rules", "all"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runList,
	}

	cmd.Flags().BoolVarP(&listShowArchived, "archived", "a", false, "Show only archived items")
	cmd.Flags().BoolVar(&listShowPaths, "paths", false, "Show file paths")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")

	var result ListResult
	result.Type = listType

	// List pipelines if requested
	if listType == "all" || listType == "pipelines" {
		pipelines, err := listPipelines()
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}
		
		if listType == "pipelines" {
			result.Items = pipelines
		} else {
			result.Items = append(result.Items, pipelines...)
		}
	}

	// List components if requested
	if listType == "all" || listType == "components" || listType == "contexts" || listType == "prompts" || listType == "rules" {
		var components []ListItem
		var err error
		
		// If a specific component type was requested, filter to just that type
		if listType == "contexts" || listType == "prompts" || listType == "rules" {
			components, err = listSpecificComponentType(listType)
		} else {
			components, err = listComponents()
		}
		
		if err != nil {
			return fmt.Errorf("failed to list components: %w", err)
		}
		
		if listType != "all" {
			result.Items = components
		} else {
			result.Items = append(result.Items, components...)
		}
	}

	result.Count = len(result.Items)

	// Output results
	switch outputFormat {
	case "json", "yaml":
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, result)
	default:
		return outputListText(cmd, result)
	}
}

func listPipelines() ([]ListItem, error) {
	var items []ListItem

	if listShowArchived {
		// Show ONLY archived pipelines when --archived flag is used
		archiveDir := filepath.Join(files.PluqqyDir, "archive", "pipelines")
		archivedPipelines, err := listPipelinesFromDir(archiveDir, true)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		items = append(items, archivedPipelines...)
	} else {
		// Show only regular (non-archived) pipelines by default
		pipelineDir := filepath.Join(files.PluqqyDir, "pipelines")
		regularPipelines, err := listPipelinesFromDir(pipelineDir, false)
		if err != nil {
			return nil, err
		}
		items = append(items, regularPipelines...)
	}

	return items, nil
}

func listPipelinesFromDir(dir string, isArchived bool) ([]ListItem, error) {
	var items []ListItem

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		pipelinePath := filepath.Join(dir, entry.Name())
		
		var pipeline *models.Pipeline
		var err error
		
		if isArchived {
			// For archived pipelines, use ReadArchivedPipeline with just the filename
			pipeline, err = files.ReadArchivedPipeline(entry.Name())
		} else {
			// For regular pipelines, use LoadPipeline
			pipeline, err = files.LoadPipeline(pipelinePath)
		}
		
		if err != nil {
			cli.PrintWarning("Failed to load pipeline %s: %v", name, err)
			continue
		}

		item := ListItem{
			Name:       pipeline.Name,
			Filename:   name,  // The filename without .yaml extension
			Type:       "pipeline",
			Tags:       pipeline.Tags,
			Components: len(pipeline.Components),
			IsArchived: isArchived,
		}

		if listShowPaths {
			item.Path = pipelinePath
		}

		items = append(items, item)
	}

	return items, nil
}

func listSpecificComponentType(componentType string) ([]ListItem, error) {
	var items []ListItem
	
	// componentType comes in as singular (e.g., "contexts", "prompts", "rules")
	// Get singular form for display
	componentTypeSingular := strings.TrimSuffix(componentType, "s")
	
	if listShowArchived {
		// Show ONLY archived components when --archived flag is used
		archiveDir := filepath.Join(files.PluqqyDir, "archive", "components", componentType)
		archivedComponents, err := listComponentsFromDir(archiveDir, componentTypeSingular, true)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		items = append(items, archivedComponents...)
	} else {
		// Show only regular (non-archived) components by default
		componentDir := filepath.Join(files.PluqqyDir, "components", componentType)
		regularComponents, err := listComponentsFromDir(componentDir, componentTypeSingular, false)
		if err != nil {
			return nil, err
		}
		items = append(items, regularComponents...)
	}
	
	return items, nil
}

func listComponents() ([]ListItem, error) {
	var items []ListItem

	// List all component types
	componentTypes := []string{
		"contexts",
		"prompts",
		"rules",
	}

	for _, ct := range componentTypes {
		// Get singular form for display
		componentTypeSingular := strings.TrimSuffix(ct, "s")
		
		if listShowArchived {
			// Show ONLY archived components when --archived flag is used
			archiveDir := filepath.Join(files.PluqqyDir, "archive", "components", ct)
			archivedComponents, err := listComponentsFromDir(archiveDir, componentTypeSingular, true)
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			items = append(items, archivedComponents...)
		} else {
			// Show only regular (non-archived) components by default
			componentDir := filepath.Join(files.PluqqyDir, "components", ct)
			regularComponents, err := listComponentsFromDir(componentDir, componentTypeSingular, false)
			if err != nil {
				return nil, err
			}
			items = append(items, regularComponents...)
		}
	}

	return items, nil
}

func listComponentsFromDir(dir string, componentType string, isArchived bool) ([]ListItem, error) {
	var items []ListItem

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		componentPath := filepath.Join(dir, entry.Name())
		
		component, err := files.LoadComponent(componentPath)
		if err != nil {
			cli.PrintWarning("Failed to load component %s: %v", name, err)
			continue
		}

		item := ListItem{
			Name:       component.Name,
			Filename:   name,  // The filename without .md extension
			Type:       componentType,
			Tags:       component.Tags,
			IsArchived: isArchived,
		}

		if listShowPaths {
			item.Path = componentPath
		}

		items = append(items, item)
	}

	return items, nil
}

func outputListText(cmd *cobra.Command, result ListResult) error {
	if result.Count == 0 {
		cli.PrintInfo("No items found")
		return nil
	}

	// Group items by type for better display
	pipelines := []ListItem{}
	components := map[string][]ListItem{
		"context": {},
		"prompt":  {},
		"rule":    {},
	}

	for _, item := range result.Items {
		if item.Type == "pipeline" {
			pipelines = append(pipelines, item)
		} else {
			components[item.Type] = append(components[item.Type], item)
		}
	}

	// Display pipelines
	if len(pipelines) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\nPIPELINES")
		fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))
		
		table := cli.NewTableFormatter(cmd.OutOrStdout())
		if listShowPaths {
			table.Header("Name", "Filename", "Components", "Tags", "Path")
		} else {
			table.Header("Name", "Filename", "Components", "Tags")
		}
		
		for _, p := range pipelines {
			tags := strings.Join(p.Tags, ", ")
			if tags == "" {
				tags = "-"
			}
			
			if listShowPaths {
				table.Row(p.Name, p.Filename, fmt.Sprintf("%d", p.Components), tags, p.Path)
			} else {
				table.Row(p.Name, p.Filename, fmt.Sprintf("%d", p.Components), tags)
			}
		}
		table.Flush()
	}

	// Display components
	for _, ct := range []string{"context", "prompt", "rule"} {
		if len(components[ct]) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%sS\n", strings.ToUpper(ct))
			fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))
			
			table := cli.NewTableFormatter(cmd.OutOrStdout())
			if listShowPaths {
				table.Header("Name", "Filename", "Tags", "Path")
			} else {
				table.Header("Name", "Filename", "Tags")
			}
			
			for _, c := range components[ct] {
				tags := strings.Join(c.Tags, ", ")
				if tags == "" {
					tags = "-"
				}
				
				if listShowPaths {
					table.Row(c.Name, c.Filename, tags, c.Path)
				} else {
					table.Row(c.Name, c.Filename, tags)
				}
			}
			table.Flush()
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d items\n", result.Count)
	
	return nil
}