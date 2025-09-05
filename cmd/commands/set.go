package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
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
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runSet,
	}

	cmd.Flags().StringVar(&outputFilename, "output-file", "", "Custom output filename (default: PLUQQY.md)")

	return cmd
}

func runSet(cmd *cobra.Command, args []string) error {
	itemRef := args[0]
	
	// Create command context and load settings
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	settings := ctx.LoadSettingsWithDefault()

	// Override output filename if specified
	if outputFilename != "" {
		settings.Output.DefaultFilename = outputFilename
	}
	
	var composed string
	var itemType string
	var itemName string

	// Create resolver to find items
	resolver := cli.NewItemResolver(ctx.ProjectPath)

	// Try to resolve the item
	itemTypeResolved, itemPath, err := resolver.ResolveItem(itemRef)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
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
		return err
	}

	switch itemTypeResolved {
	case "pipeline":
		pipeline, err := files.LoadPipeline(itemPath)
		if err != nil {
			return fmt.Errorf("failed to load pipeline: %w", err)
		}

		composed, err = composer.ComposePipelineWithSettings(pipeline, settings)
		if err != nil {
			return fmt.Errorf("failed to compose pipeline: %w", err)
		}
		itemType = "Pipeline"
		itemName = pipeline.Name

	case "component":
		component, err := files.LoadComponent(itemPath)
		if err != nil {
			return fmt.Errorf("failed to load component: %w", err)
		}

		composed, err = composer.ComposeComponentWithSettings(component, settings)
		if err != nil {
			return fmt.Errorf("failed to compose component: %w", err)
		}
		itemType = "Component"
		itemName = component.Name

	case "archived":
		return fmt.Errorf("cannot set archived item '%s'. Please restore it first", itemRef)

	default:
		return fmt.Errorf("unknown item type: %s", itemTypeResolved)
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