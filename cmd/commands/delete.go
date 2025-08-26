package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

var (
	deleteForce bool
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <item>",
		Short: "Delete a component or pipeline",
		Long: `Permanently delete a component or pipeline.

This action cannot be undone. Consider archiving instead if you
might need the item later.

Examples:
  # Delete a component (with confirmation)
  pluqqy delete api-docs
  
  # Delete a pipeline
  pluqqy delete my-pipeline
  
  # Force delete without confirmation
  pluqqy delete old-component --force`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runDelete,
	}

	cmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Force deletion without confirmation")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Create command context and resolver
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	resolver := cli.NewItemResolver(ctx.ProjectPath)

	// Try to resolve the item
	itemTypeResolved, itemPath, resolveErr := resolver.ResolveItem(itemRef)
	if resolveErr != nil {
		return resolveErr
	}

	var itemType string
	var isArchived bool

	switch itemTypeResolved {
	case "pipeline":
		itemType = "pipeline"
		isArchived = false
	case "component":
		itemType = "component"
		isArchived = false
	case "archived":
		isArchived = true
		// Determine if it's a component or pipeline based on path
		if strings.Contains(itemPath, "/components/") {
			itemType = "component"
		} else if strings.Contains(itemPath, "/pipelines/") {
			itemType = "pipeline"
		} else {
			return fmt.Errorf("cannot determine type of archived item: %s", itemPath)
		}
	default:
		return fmt.Errorf("unknown item type: %s", itemTypeResolved)
	}

	// Show warning for non-archived items
	if !isArchived && !deleteForce {
		cli.PrintWarning("Item '%s' is not archived. Consider archiving instead of deleting.", itemRef)
		cli.PrintInfo("Use 'pluqqy archive %s' to archive, or --force to delete anyway.", itemRef)
	}

	// Confirm deletion (only one confirmation)
	if !deleteForce {
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			prompt := fmt.Sprintf("Permanently delete %s '%s'? This cannot be undone.", itemType, itemRef)
			confirmed, err := cli.Confirm(prompt, false)
			if err != nil {
				return err
			}
			if !confirmed {
				cli.PrintInfo("Deletion cancelled")
				return nil
			}
		}
	}

	// Perform deletion
	var deleteErr error
	if itemType == "component" {
		// DeleteComponent and DeleteArchivedComponent expect relative paths
		// Use resolver to convert to relative path
		relativePath := resolver.ConvertToRelativePath(itemPath)
		
		if isArchived {
			// For archived components, remove the "archive/" prefix to get the standard component path
			relativePath = strings.TrimPrefix(relativePath, "archive"+string(os.PathSeparator))
			deleteErr = files.DeleteArchivedComponent(relativePath)
		} else {
			deleteErr = files.DeleteComponent(relativePath)
		}
	} else {
		// For pipelines, use standard os.Remove
		deleteErr = os.Remove(itemPath)
	}

	if deleteErr != nil {
		return fmt.Errorf("failed to delete %s: %w", itemType, deleteErr)
	}

	cli.PrintSuccess("Deleted %s: %s", itemType, itemRef)
	if isArchived {
		cli.PrintInfo("Deleted from archive")
	}
	
	return nil
}