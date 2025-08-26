package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
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
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runEdit,
	}

	return cmd
}

func runEdit(cmd *cobra.Command, args []string) error {
	componentRef := args[0]

	// Find the component file
	componentPath, err := findComponentFile(componentRef)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return fmt.Errorf("component not found: %s", componentRef)
	}

	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Parse editor command to handle arguments like "--wait" or "-w"
	parts := strings.Fields(editor)
	var editorCmd *exec.Cmd
	if len(parts) > 1 {
		// Editor has arguments (e.g., "code --wait")
		editorCmd = exec.Command(parts[0], append(parts[1:], componentPath)...)
	} else {
		// Simple editor command (e.g., "vim")
		editorCmd = exec.Command(editor, componentPath)
	}
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	cli.PrintInfo("Opening %s in editor...", componentPath)
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	cli.PrintSuccess("Component edited successfully")
	return nil
}

// findComponentFile locates a component file by name or path
func findComponentFile(ref string) (string, error) {
	// If ref contains a slash, treat it as a path hint
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		componentType := parts[0]
		componentName := parts[1]
		
		// Normalize component type
		switch strings.ToLower(componentType) {
		case "context", "contexts":
			componentType = "contexts"
		case "prompt", "prompts":
			componentType = "prompts"
		case "rule", "rules":
			componentType = "rules"
		default:
			return "", fmt.Errorf("invalid component type: %s", componentType)
		}
		
		// Add .md extension if not present
		if !strings.HasSuffix(componentName, ".md") {
			componentName += ".md"
		}
		
		path := filepath.Join(files.PluqqyDir, "components", componentType, componentName)
		return path, nil
	}

	// Search for the component in all directories
	componentTypes := []string{"contexts", "prompts", "rules"}
	var foundPaths []string

	for _, ct := range componentTypes {
		// Try with .md extension
		path := filepath.Join(files.PluqqyDir, "components", ct, ref+".md")
		if _, err := os.Stat(path); err == nil {
			foundPaths = append(foundPaths, path)
		}
		
		// Try without extension (in case user included it)
		if strings.HasSuffix(ref, ".md") {
			path = filepath.Join(files.PluqqyDir, "components", ct, ref)
			if _, err := os.Stat(path); err == nil {
				foundPaths = append(foundPaths, path)
			}
		}
	}

	// Check archived components if not found
	if len(foundPaths) == 0 {
		for _, ct := range componentTypes {
			path := filepath.Join(files.PluqqyDir, "archive", "components", ct, ref+".md")
			if _, err := os.Stat(path); err == nil {
				foundPaths = append(foundPaths, path)
			}
		}
	}

	if len(foundPaths) == 0 {
		return "", fmt.Errorf("component not found: %s", ref)
	}

	if len(foundPaths) > 1 {
		return "", fmt.Errorf("multiple components found with name '%s'. Please specify the type (e.g., contexts/%s)", ref, ref)
	}

	return foundPaths[0], nil
}