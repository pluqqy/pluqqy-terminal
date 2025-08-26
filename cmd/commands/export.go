package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

var (
	exportToFile string
)

// NewExportCommand creates the export command
func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <pipeline|component>",
		Short: "Export composed pipeline or component to stdout or file",
		Long: `Export a pipeline or component by composing and outputting the result.

This command works with both:
- Pipeline files (e.g., my-assistant)
- Individual components (e.g., contexts/api-docs, prompts/user-story)

By default, the composed content is written to stdout. You can redirect
it to a file using shell redirection or the --file flag.

Examples:
  # Export pipeline to stdout
  pluqqy export my-assistant
  
  # Export component to stdout
  pluqqy export contexts/api-docs
  pluqqy export prompts/user-story
  
  # Export to file using redirection
  pluqqy export my-assistant > output.md
  
  # Export to file using flag
  pluqqy export my-assistant --file output.md
  pluqqy export contexts/api-docs --file context.md
  
  # Export as YAML format (pipelines only)
  pluqqy export my-assistant -o yaml`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			
			// Don't validate here since it could be a pipeline or component
			return nil
		},
		RunE: runExport,
	}

	cmd.Flags().StringVarP(&exportToFile, "file", "f", "", "Export to file instead of stdout")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Load settings for composition
	settings, err := files.ReadSettings()
	if err != nil {
		// Use default settings if can't read
		settings = models.DefaultSettings()
	}

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")

	var output string
	var itemType string
	var itemName string
	var exportData interface{} // For structured output formats

	// Check if it's a component reference (contains /)
	if strings.Contains(itemRef, "/") {
		// It's a component reference like "contexts/api-docs"
		componentPath, err := findComponentFile(itemRef)
		if err != nil {
			return fmt.Errorf("failed to find component: %w", err)
		}

		// Load the component
		component, err := files.LoadComponent(componentPath)
		if err != nil {
			return fmt.Errorf("failed to load component: %w", err)
		}

		itemType = "Component"
		itemName = component.Name
		exportData = component

		// Compose the component for text output
		if outputFormat != "json" && outputFormat != "yaml" {
			composed, err := composer.ComposeComponentWithSettings(component, settings)
			if err != nil {
				return fmt.Errorf("failed to compose component: %w", err)
			}
			output = composed
		}
	} else {
		// Try as pipeline first
		pipelineName := strings.TrimSuffix(itemRef, ".yaml")
		pipelinePath := filepath.Join(files.PluqqyDir, "pipelines", pipelineName+".yaml")
		pipeline, pipelineErr := files.LoadPipeline(pipelinePath)
		
		if pipelineErr == nil {
			// It's a pipeline
			itemType = "Pipeline"
			itemName = pipeline.Name
			exportData = pipeline

			// Compose the pipeline for text output
			if outputFormat != "json" && outputFormat != "yaml" {
				composed, err := composer.ComposePipelineWithSettings(pipeline, settings)
				if err != nil {
					return fmt.Errorf("failed to compose pipeline: %w", err)
				}
				output = composed
			}
		} else {
			// Try as component without type prefix
			componentPath, componentErr := findComponentFile(itemRef)
			if componentErr == nil {
				component, err := files.LoadComponent(componentPath)
				if err != nil {
					return fmt.Errorf("failed to load component: %w", err)
				}

				itemType = "Component"
				itemName = component.Name
				exportData = component

				// Compose the component for text output
				if outputFormat != "json" && outputFormat != "yaml" {
					composed, err := composer.ComposeComponentWithSettings(component, settings)
					if err != nil {
						return fmt.Errorf("failed to compose component: %w", err)
					}
					output = composed
				}
			} else {
				// Not found as either pipeline or component
				if os.IsNotExist(pipelineErr) || strings.Contains(pipelineErr.Error(), "no such file or directory") {
					return fmt.Errorf("item '%s' not found as pipeline or component", itemRef)
				}
				return fmt.Errorf("failed to load item: %w", pipelineErr)
			}
		}
	}

	// Handle structured output formats
	if outputFormat == "json" || outputFormat == "yaml" {
		if exportToFile != "" {
			// Create file for structured output
			file, err := os.Create(exportToFile)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()
			
			err = cli.OutputResults(file, outputFormat, exportData)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}
			
			cli.PrintSuccess("%s exported to: %s (%s format)", itemType, exportToFile, outputFormat)
		} else {
			// Write to stdout
			err = cli.OutputResults(cmd.OutOrStdout(), outputFormat, exportData)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}
		}
		return nil
	}

	// Write text/markdown output
	if exportToFile != "" {
		// Write to file
		if err := os.WriteFile(exportToFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		cli.PrintSuccess("%s '%s' exported to: %s", itemType, itemName, exportToFile)
		
		// Show token count
		tokenCount := composer.EstimateTokens(output)
		cli.PrintInfo("Estimated tokens: %d", tokenCount)
	} else {
		// Write to stdout
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	return nil
}