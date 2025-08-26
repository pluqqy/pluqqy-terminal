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
	outputFilename string
)

// NewSetCommand creates the set command
func NewSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <pipeline|component>",
		Short: "Set active pipeline or component (generates output file)",
		Long: `Set the specified pipeline or component as active and generate the output file.

This command can work with either:
- Pipeline files (e.g., cli-development or cli-development.yaml)
- Individual components (e.g., contexts/api-docs, prompts/user-story)

The output is written to PLUQQY.md by default, or to a custom filename if specified.

Examples:
  # Set a pipeline
  pluqqy set cli-development
  
  # Set a specific component
  pluqqy set contexts/api-docs
  pluqqy set prompts/user-story
  
  # Set with custom output filename
  pluqqy set cli-development --output-file MY_PROMPT.md
  pluqqy set contexts/api-docs --output-file CONTEXT.md
  
  # Set a pipeline with quiet output
  pluqqy set cli-development -q`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			
			// Don't validate here since it could be a pipeline or component
			return nil
		},
		RunE: runSet,
	}

	cmd.Flags().StringVar(&outputFilename, "output-file", "", "Custom output filename (default: PLUQQY.md)")

	return cmd
}

func runSet(cmd *cobra.Command, args []string) error {
	itemRef := args[0]
	
	// Load settings for output configuration
	settings, err := files.ReadSettings()
	if err != nil {
		// Use default settings if can't read
		settings = models.DefaultSettings()
	}

	// Override output filename if specified
	if outputFilename != "" {
		settings.Output.DefaultFilename = outputFilename
	}
	
	var composed string
	var itemType string
	var itemName string

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

		// Compose the component
		composed, err = composer.ComposeComponentWithSettings(component, settings)
		if err != nil {
			return fmt.Errorf("failed to compose component: %w", err)
		}

		itemType = "Component"
		itemName = component.Name
	} else {
		// Try as pipeline first
		pipelineName := strings.TrimSuffix(itemRef, ".yaml")
		pipelinePath := filepath.Join(files.PluqqyDir, "pipelines", pipelineName+".yaml")
		pipeline, pipelineErr := files.LoadPipeline(pipelinePath)
		
		if pipelineErr == nil {
			// It's a pipeline
			composed, err = composer.ComposePipelineWithSettings(pipeline, settings)
			if err != nil {
				return fmt.Errorf("failed to compose pipeline: %w", err)
			}
			itemType = "Pipeline"
			itemName = pipeline.Name
		} else {
			// Try as component without type prefix
			componentPath, componentErr := findComponentFile(itemRef)
			if componentErr == nil {
				component, err := files.LoadComponent(componentPath)
				if err != nil {
					return fmt.Errorf("failed to load component: %w", err)
				}

				composed, err = composer.ComposeComponentWithSettings(component, settings)
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
					errMsg.WriteString("  For pipelines: pluqqy set <pipeline-filename>\n")
					errMsg.WriteString("  For components: pluqqy set <type>/<component-name>\n\n")
					errMsg.WriteString("Examples:\n")
					errMsg.WriteString("  pluqqy set cli-development\n")
					errMsg.WriteString("  pluqqy set contexts/api-docs\n")
					errMsg.WriteString("  pluqqy set prompts/user-story\n\n")
					errMsg.WriteString("Run 'pluqqy list' to see available items\n")
					
					return fmt.Errorf("%s", errMsg.String())
				}
				return fmt.Errorf("failed to load item: %w", pipelineErr)
			}
		}
	}

	// Determine output path
	outputPath := filepath.Join(settings.Output.ExportPath, settings.Output.DefaultFilename)
	
	// Create parent directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the output file
	if err := os.WriteFile(outputPath, []byte(composed), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	cli.PrintSuccess("%s '%s' set as active", itemType, itemName)
	cli.PrintInfo("Output written to: %s", outputPath)
	
	// Show token count if not quiet
	tokenCount := composer.EstimateTokens(composed)
	if !cmd.Flags().Lookup("quiet").Changed {
		cli.PrintInfo("Estimated tokens: %d", tokenCount)
	}

	return nil
}

// findComponentFile is defined in edit.go