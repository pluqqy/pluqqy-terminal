package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

// NewArchiveCommand creates the archive command
func NewArchiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <item>",
		Short: "Archive a component or pipeline",
		Long: `Archive a component or pipeline to move it out of active use.

Archived items are moved to the archive directory and won't appear
in normal listings unless specifically requested.

Examples:
  # Archive a component
  pluqqy archive api-docs
  
  # Archive a pipeline
  pluqqy archive my-pipeline
  
  # Archive without confirmation
  pluqqy archive old-component -y`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			return ctx.ValidateProject()
		},
		RunE: runArchive,
	}

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Create command context and resolver
	ctx, err := cli.NewCommandContext()
	if err != nil {
		return err
	}
	resolver := cli.NewItemResolver(ctx.ProjectPath)

	// Try to resolve the item
	itemTypeResolved, itemPath, err := resolver.ResolveItem(itemRef)
	if err != nil {
		return err
	}

	var itemType string
	switch itemTypeResolved {
	case "pipeline":
		itemType = "pipeline"
	case "component":
		itemType = "component"
	case "archived":
		return fmt.Errorf("item is already archived: %s", itemRef)
	default:
		return fmt.Errorf("unknown item type: %s", itemTypeResolved)
	}

	// Confirm archive
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	if !skipConfirm {
		confirmed, err := cli.Confirm(fmt.Sprintf("Archive %s '%s'?", itemType, itemRef), false)
		if err != nil {
			return err
		}
		if !confirmed {
			cli.PrintInfo("Archive cancelled")
			return nil
		}
	}

	// Perform archive
	var archiveErr error
	if itemType == "component" {
		// ArchiveComponent expects a relative path like "components/contexts/item.md"
		relativePath := resolver.ConvertToRelativePath(itemPath)
		archiveErr = files.ArchiveComponent(relativePath)
	} else {
		// ArchivePipeline expects just the filename
		archiveErr = files.ArchivePipeline(filepath.Base(itemPath))
	}

	if archiveErr != nil {
		return fmt.Errorf("failed to archive %s: %w", itemType, archiveErr)
	}

	cli.PrintSuccess("Archived %s: %s", itemType, itemRef)
	return nil
}