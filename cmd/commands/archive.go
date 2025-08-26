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
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runArchive,
	}

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Try to find as component first
	componentPath, componentErr := findComponentFile(itemRef)
	
	// Try to find as pipeline
	pipelinePath := filepath.Join(files.PluqqyDir, "pipelines", itemRef+".yaml")
	_, pipelineErr := os.Stat(pipelinePath)

	var itemPath string
	var itemType string

	if componentErr == nil {
		itemPath = componentPath
		itemType = "component"
	} else if pipelineErr == nil {
		itemPath = pipelinePath
		itemType = "pipeline"
	} else {
		// Check if the component error is about multiple matches
		if componentErr != nil && strings.Contains(componentErr.Error(), "multiple components found") {
			return componentErr
		}
		return fmt.Errorf("item not found: %s", itemRef)
	}

	// Check if already archived
	if strings.Contains(itemPath, "/archive/") {
		return fmt.Errorf("item is already archived: %s", itemRef)
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
	var err error
	if itemType == "component" {
		// ArchiveComponent expects a relative path like "components/contexts/item.md"
		// Extract the relative path from the full path
		relativePath := strings.TrimPrefix(itemPath, files.PluqqyDir+string(os.PathSeparator))
		err = files.ArchiveComponent(relativePath)
	} else {
		// ArchivePipeline expects just the filename
		err = files.ArchivePipeline(filepath.Base(itemPath))
	}

	if err != nil {
		return fmt.Errorf("failed to archive %s: %w", itemType, err)
	}

	cli.PrintSuccess("Archived %s: %s", itemType, itemRef)
	return nil
}