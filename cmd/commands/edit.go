package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
)

// NewEditCommand creates the edit command
func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <component>",
		Short: "Edit a component in your editor",
		Long: `Edit an existing component in your default editor ($EDITOR).

The component can be specified by name or path. The command will search
for the component in all component directories.

Examples:
  # Edit a component by name
  pluqqy edit api-docs
  
  # Edit a specific component type
  pluqqy edit prompts/user-story
  
  # Edit with specific editor
  EDITOR=vim pluqqy edit security-rules`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runEdit,
	}

	return cmd
}

func runEdit(cmd *cobra.Command, args []string) error {
	componentRef := args[0]

	// Create command context and finder
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	finder := cli.NewComponentFinder(ctx.ProjectPath)

	// Find the component file
	componentPath, err := finder.FindByReference(componentRef)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return fmt.Errorf("component not found: %s", componentRef)
	}

	// Open in editor
	launcher := cli.NewEditorLauncher()
	cli.PrintInfo("Opening %s in editor...", componentPath)
	if err := launcher.OpenFile(componentPath); err != nil {
		return err
	}

	cli.PrintSuccess("Component edited successfully")
	return nil
}

