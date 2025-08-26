package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

var (
	clipboardFormat string
)

// NewClipboardCommand creates the clipboard command
func NewClipboardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clipboard <pipeline|component>",
		Short: "Copy pipeline or component content to clipboard",
		Long: `Copy a pipeline's or component's generated content to the system clipboard.

This command works with both:
- Pipeline files (e.g., cli-development)
- Individual components (e.g., contexts/api-docs, prompts/user-story)

The content is generated in the same format as when setting
it as active, ready to be pasted into your AI tool.

Examples:
  # Copy pipeline content to clipboard
  pluqqy clipboard add-cli-command
  
  # Copy component content to clipboard
  pluqqy clipboard contexts/api-docs
  pluqqy clipboard prompts/user-story
  
  # Handle ambiguous names by specifying type
  pluqqy clipboard prompts/123`,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"clip", "copy"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runClipboard,
	}

	cmd.Flags().StringVar(&clipboardFormat, "format", "markdown", "Output format (markdown)")

	return cmd
}

func runClipboard(cmd *cobra.Command, args []string) error {
	itemRef := args[0]
	
	// Load settings for output configuration
	settings, err := files.ReadSettings()
	if err != nil {
		// Use default settings if can't read
		settings = models.DefaultSettings()
	}

	var content string
	var itemType string
	var itemName string

	// Check if it's a component reference (contains /)
	if strings.Contains(itemRef, "/") {
		// It's a component reference like "contexts/api-docs"
		componentPath, err := findComponentFileForClipboard(itemRef)
		if err != nil {
			return fmt.Errorf("failed to find component: %w", err)
		}

		// Load the component
		component, err := files.LoadComponent(componentPath)
		if err != nil {
			return fmt.Errorf("failed to load component: %w", err)
		}

		// Compose the component
		content, err = composer.ComposeComponentWithSettings(component, settings)
		if err != nil {
			return fmt.Errorf("failed to compose component: %w", err)
		}

		itemType = "Component"
		itemName = component.Name
	} else {
		// Try as pipeline first
		pipelineName := strings.TrimSuffix(itemRef, ".yaml")
		pipeline, pipelineErr := files.ReadPipeline(pipelineName + ".yaml")
		
		if pipelineErr == nil {
			// It's a pipeline
			content, err = composer.ComposePipelineWithSettings(pipeline, settings)
			if err != nil {
				return fmt.Errorf("failed to compose pipeline: %w", err)
			}
			itemType = "Pipeline"
			itemName = pipeline.Name
		} else {
			// Try as component without type prefix
			componentPath, componentErr := findComponentFileForClipboard(itemRef)
			if componentErr == nil {
				component, err := files.LoadComponent(componentPath)
				if err != nil {
					return fmt.Errorf("failed to load component: %w", err)
				}

				content, err = composer.ComposeComponentWithSettings(component, settings)
				if err != nil {
					return fmt.Errorf("failed to compose component: %w", err)
				}
				itemType = "Component"
				itemName = component.Name
			} else {
				// Not found as either pipeline or component
				if os.IsNotExist(pipelineErr) || strings.Contains(pipelineErr.Error(), "no such file or directory") {
					var errMsg strings.Builder
					errMsg.WriteString(fmt.Sprintf("Item '%s' not found as pipeline or component\n\n", itemRef))
					errMsg.WriteString("Usage:\n")
					errMsg.WriteString("  For pipelines: pluqqy clipboard <pipeline-filename>\n")
					errMsg.WriteString("  For components: pluqqy clipboard <type>/<component-name>\n\n")
					errMsg.WriteString("Examples:\n")
					errMsg.WriteString("  pluqqy clipboard cli-development\n")
					errMsg.WriteString("  pluqqy clipboard contexts/api-docs\n")
					errMsg.WriteString("  pluqqy clipboard prompts/user-story\n\n")
					errMsg.WriteString("Run 'pluqqy list' to see available items\n")
					
					return fmt.Errorf(errMsg.String())
				}
				return fmt.Errorf("failed to load item: %w", pipelineErr)
			}
		}
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	cli.PrintSuccess("✓ %s '%s' copied to clipboard", itemType, itemName)
	
	// Show token count
	tokenCount := composer.EstimateTokens(content)
	cli.PrintInfo("Estimated tokens: %d", tokenCount)
	
	// Show a preview of what was copied
	lines := strings.Split(content, "\n")
	preview := lines[0]
	if len(lines) > 1 {
		preview += " ..."
	}
	if len(preview) > 80 {
		preview = preview[:77] + "..."
	}
	cli.PrintInfo("Preview: %s", preview)

	return nil
}

// findComponentFileForClipboard finds a component file by reference
func findComponentFileForClipboard(ref string) (string, error) {
	// If ref contains a slash, treat it as a path hint
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		componentType := parts[0]
		componentName := parts[1]
		
		// Normalize component type
		switch strings.ToLower(componentType) {
		case "context", "contexts":
			componentType = "contexts"
		case "prompt", "prompts":
			componentType = "prompts"
		case "rule", "rules":
			componentType = "rules"
		default:
			return "", fmt.Errorf("invalid component type: %s", componentType)
		}
		
		// Add .md extension if not present
		if !strings.HasSuffix(componentName, ".md") {
			componentName += ".md"
		}
		
		path := filepath.Join(files.PluqqyDir, "components", componentType, componentName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("component not found: %s/%s", componentType, strings.TrimSuffix(componentName, ".md"))
	}

	// Search for the component in all directories
	componentTypes := []string{"contexts", "prompts", "rules"}
	var foundPaths []string

	for _, ct := range componentTypes {
		// Try with .md extension
		path := filepath.Join(files.PluqqyDir, "components", ct, ref+".md")
		if _, err := os.Stat(path); err == nil {
			foundPaths = append(foundPaths, path)
		}
		
		// Try without extension (in case user included it)
		if strings.HasSuffix(ref, ".md") {
			path = filepath.Join(files.PluqqyDir, "components", ct, ref)
			if _, err := os.Stat(path); err == nil {
				foundPaths = append(foundPaths, path)
			}
		}
	}

	if len(foundPaths) == 0 {
		return "", fmt.Errorf("component not found: %s", ref)
	}

	if len(foundPaths) > 1 {
		return "", fmt.Errorf("multiple components found with name '%s'. Please specify the type (e.g., contexts/%s)", ref, ref)
	}

	return foundPaths[0], nil
}

func listAvailablePipelines() ([]string, error) {
	pipelineDir := filepath.Join(files.PluqqyDir, "pipelines")
	entries, err := os.ReadDir(pipelineDir)
	if err != nil {
		return nil, err
	}

	var pipelines []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			// Remove .yaml extension for display
			name := strings.TrimSuffix(entry.Name(), ".yaml")
			pipelines = append(pipelines, name)
		}
	}

	return pipelines, nil
}