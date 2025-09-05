package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

// NewRestoreCommand creates the restore command
func NewRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <item>",
		Short: "Restore an archived component or pipeline",
		Long: `Restore an archived component or pipeline back to active use.

The item will be moved from the archive directory back to its
original location.

Examples:
  # Restore a component
  pluqqy restore api-docs
  
  # Restore a pipeline
  pluqqy restore my-pipeline
  
  # Restore without confirmation
  pluqqy restore old-component -y`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runRestore,
	}

	return cmd
}

func runRestore(cmd *cobra.Command, args []string) error {
	itemRef := args[0]

	// Search in archive directories
	var archivedPath string
	var itemType string

	// Check component archives
	componentTypes := []string{"contexts", "prompts", "rules"}
	for _, ct := range componentTypes {
		path := filepath.Join(files.PluqqyDir, "archive", "components", ct, itemRef+".md")
		if _, err := os.Stat(path); err == nil {
			archivedPath = path
			itemType = "component"
			break
		}
	}

	// Check pipeline archive if not found as component
	if archivedPath == "" {
		path := filepath.Join(files.PluqqyDir, "archive", "pipelines", itemRef+".yaml")
		if _, err := os.Stat(path); err == nil {
			archivedPath = path
			itemType = "pipeline"
		}
	}

	if archivedPath == "" {
		// Provide more helpful error message
		return fmt.Errorf("archived item '%s' not found\n\nUse 'pluqqy list --archived' to see available archived items", itemRef)
	}

	// Determine target path
	var targetPath string
	if itemType == "component" {
		// Extract component type from path
		parts := strings.Split(archivedPath, string(os.PathSeparator))
		for i, part := range parts {
			if part == "components" && i+1 < len(parts) {
				componentType := parts[i+1]
				targetPath = filepath.Join(files.PluqqyDir, "components", componentType, filepath.Base(archivedPath))
				break
			}
		}
	} else {
		targetPath = filepath.Join(files.PluqqyDir, "pipelines", filepath.Base(archivedPath))
	}

	// Check if target already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("%s already exists in active directory: %s", itemType, itemRef)
	}

	// Confirm restore
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	if !skipConfirm {
		confirmed, err := cli.Confirm(fmt.Sprintf("Restore %s '%s'?", itemType, itemRef), false)
		if err != nil {
			return err
		}
		if !confirmed {
			cli.PrintInfo("Restore cancelled")
			return nil
		}
	}

	// Perform restore
	var err error
	if itemType == "component" {
		// UnarchiveComponent expects a relative path like "components/contexts/item.md"
		// Extract the relative path from the archived path
		parts := strings.Split(archivedPath, string(os.PathSeparator))
		var relativePath string
		for i, part := range parts {
			if part == "components" && i+2 < len(parts) {
				// Join components/type/filename
				relativePath = filepath.Join("components", parts[i+1], parts[i+2])
				break
			}
		}
		if relativePath == "" {
			return fmt.Errorf("could not determine component type from path: %s", archivedPath)
		}
		err = files.UnarchiveComponent(relativePath)
	} else {
		// UnarchivePipeline expects just the filename
		err = files.UnarchivePipeline(filepath.Base(archivedPath))
	}

	if err != nil {
		return fmt.Errorf("failed to restore %s: %w", itemType, err)
	}

	cli.PrintSuccess("Restored %s: %s", itemType, itemRef)
	return nil
}