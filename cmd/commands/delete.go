package commands

import (
	"fmt"
	"os"
	"path/filepath"
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
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runDelete,
	}

	cmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Force deletion without confirmation")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Try to find the item
	var itemPath string
	var itemType string
	var isArchived bool

	// Try to find as component (active or archived)
	componentPath, componentErr := findComponentFile(itemRef)
	if componentErr == nil {
		itemPath = componentPath
		itemType = "component"
		isArchived = strings.Contains(componentPath, "/archive/")
	} else {
		// Try to find as pipeline (active)
		pipelinePath := filepath.Join(files.PluqqyDir, "pipelines", itemRef+".yaml")
		if _, err := os.Stat(pipelinePath); err == nil {
			itemPath = pipelinePath
			itemType = "pipeline"
		} else {
			// Try to find as archived pipeline
			archivedPipelinePath := filepath.Join(files.PluqqyDir, "archive", "pipelines", itemRef+".yaml")
			if _, err := os.Stat(archivedPipelinePath); err == nil {
				itemPath = archivedPipelinePath
				itemType = "pipeline"
				isArchived = true
			} else {
				// Check if the component error is about multiple matches
				if componentErr != nil && strings.Contains(componentErr.Error(), "multiple components found") {
					return componentErr
				}
				return fmt.Errorf("item not found: %s", itemRef)
			}
		}
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
	var err error
	if itemType == "component" {
		// DeleteComponent and DeleteArchivedComponent expect relative paths
		// Extract the relative path from the full path
		relativePath := strings.TrimPrefix(itemPath, files.PluqqyDir+string(os.PathSeparator))
		
		if isArchived {
			// For archived components, remove the "archive/" prefix to get the standard component path
			relativePath = strings.TrimPrefix(relativePath, "archive"+string(os.PathSeparator))
			err = files.DeleteArchivedComponent(relativePath)
		} else {
			err = files.DeleteComponent(relativePath)
		}
	} else {
		// For pipelines, use standard os.Remove
		err = os.Remove(itemPath)
	}

	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", itemType, err)
	}

	cli.PrintSuccess("Deleted %s: %s", itemType, itemRef)
	if isArchived {
		cli.PrintInfo("Deleted from archive")
	}
	
	return nil
}