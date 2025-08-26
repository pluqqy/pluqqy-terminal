package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

var (
	showMetadata bool
)

// NewShowCommand creates the show command
func NewShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <pipeline|component>",
		Short: "Display pipeline or component content",
		Long: `Display the content of a pipeline or component.

For pipelines, shows the composed output (all components combined).
For components, shows the component content (Markdown).

The item can be specified by name or path. The command will search
for the item in the appropriate directories.

Examples:
  # Show a pipeline
  pluqqy show cli-development
  
  # Show a component by name
  pluqqy show api-docs
  
  # Show with metadata
  pluqqy show api-docs --metadata
  pluqqy show cli-development --metadata
  
  # Show a specific component type
  pluqqy show prompts/user-story
  
  # Output as JSON
  pluqqy show api-docs -o json
  pluqqy show cli-development -o json`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runShow,
	}

	cmd.Flags().BoolVarP(&showMetadata, "metadata", "m", false, "Show component metadata")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	itemRef := args[0]
	
	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")

	// Create command context and resolver
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	resolver := cli.NewItemResolver(ctx.ProjectPath)

	// Resolve the item
	itemType, itemPath, err := resolver.ResolveItem(itemRef)
	if err != nil {
		return err
	}

	switch itemType {
	case "pipeline":
		pipeline, err := files.LoadPipeline(itemPath)
		if err != nil {
			return fmt.Errorf("failed to load pipeline: %w", err)
		}
		return showPipeline(cmd, pipeline, itemPath, outputFormat)
		case "component":
		return showComponentByPath(cmd, itemPath, outputFormat)
	case "archived":
		return fmt.Errorf("item '%s' is archived at %s", itemRef, itemPath)
	default:
		return fmt.Errorf("unknown item type: %s", itemType)
	}
}

func showPipeline(cmd *cobra.Command, pipeline interface{}, pipelinePath string, outputFormat string) error {
	// Cast pipeline to correct type
	p, ok := pipeline.(*models.Pipeline)
	if !ok {
		return fmt.Errorf("invalid pipeline type")
	}

	switch outputFormat {
	case "json", "yaml":
		// For structured formats, show the pipeline definition
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, pipeline)
		
	default:
		// Text output - show composed pipeline content
		
		// Load settings for composition
		ctx, _ := cli.NewCommandContext()
		settings := ctx.LoadSettingsWithDefault()

		// Compose the pipeline
		composed, err := composer.ComposePipelineWithSettings(p, settings)
		if err != nil {
			return fmt.Errorf("failed to compose pipeline: %w", err)
		}

		// Show metadata if requested
		if showMetadata {
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", p.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Type: pipeline\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Components: %d\n", len(p.Components))
			if len(p.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Tags: %s\n", strings.Join(p.Tags, ", "))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", pipelinePath)
			
			// Show token count
			tokenCount := composer.EstimateTokens(composed)
			fmt.Fprintf(cmd.OutOrStdout(), "Estimated tokens: %d\n", tokenCount)
			
			fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))
		}
		
		// Show the composed content
		fmt.Fprint(cmd.OutOrStdout(), composed)
		
		return nil
	}
}

func showComponent(cmd *cobra.Command, componentRef string, outputFormat string) error {
	// Find the component file
	ctx, _ := cli.NewCommandContext()
	finder := cli.NewComponentFinder(ctx.ProjectPath)
	componentPath, err := finder.FindByReference(componentRef)
	if err != nil {
		return err
	}

	return showComponentByPath(cmd, componentPath, outputFormat)
}

func showComponentByPath(cmd *cobra.Command, componentPath string, outputFormat string) error {
	// Load the component
	component, err := files.LoadComponent(componentPath)
	if err != nil {
		return fmt.Errorf("failed to load component: %w", err)
	}

	// Output based on format
	switch outputFormat {
	case "json", "yaml":
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, component)
		
	default:
		// Text output
		if showMetadata {
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", component.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Type: %s\n", component.Type)
			if len(component.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Tags: %s\n", strings.Join(component.Tags, ", "))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", componentPath)
			fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))
		}
		
		fmt.Fprintln(cmd.OutOrStdout(), component.Content)
	}

	return nil
}