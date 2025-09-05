package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
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
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runExport,
	}

	cmd.Flags().StringVarP(&exportToFile, "file", "f", "", "Export to file instead of stdout")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Create command context and load settings
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	settings := ctx.LoadSettingsWithDefault()

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")

	var output string
	var itemType string
	var itemName string
	var exportData interface{} // For structured output formats

	// Create resolver to find items
	resolver := cli.NewItemResolver(ctx.ProjectPath)

	// Try to resolve the item
	itemTypeResolved, itemPath, err := resolver.ResolveItem(itemRef)
	if err != nil {
		return err
	}

	switch itemTypeResolved {
	case "pipeline":
		pipeline, err := files.LoadPipeline(itemPath)
		if err != nil {
			return fmt.Errorf("failed to load pipeline: %w", err)
		}

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

	case "component":
		component, err := files.LoadComponent(itemPath)
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

	case "archived":
		return fmt.Errorf("item '%s' is archived at %s", itemRef, itemPath)

	default:
		return fmt.Errorf("unknown item type: %s", itemTypeResolved)
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