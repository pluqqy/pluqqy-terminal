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

var (
	createTags []string
)

// NewCreateCommand creates the create command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <type> <name>",
		Short: "Create a new component",
		Long: `Create a new component of the specified type.

Types:
  context  - Create a context component
  prompt   - Create a prompt component  
  rule     - Create a rule component

The command will open your default editor ($EDITOR) to write the content.

Examples:
  # Create a new context component
  pluqqy create context api-docs
  
  # Create a prompt with tags
  pluqqy create prompt user-story --tags feature,agile
  
  # Create with specific editor
  EDITOR=vim pluqqy create rule security-rules`,
		Args: cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			
			// Validate component type
			if err := cli.ValidateComponentType(args[0]); err != nil {
				return err
			}
			
			// Validate component name
			if err := cli.ValidateComponentName(args[1]); err != nil {
				return err
			}
			
			return nil
		},
		RunE: runCreate,
	}

	cmd.Flags().StringSliceVar(&createTags, "tags", []string{}, "Tags for the component (comma-separated)")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	componentType := cli.NormalizeComponentType(args[0])
	componentName := args[1]

	// Determine the relative path for the new component
	// WriteComponent expects a relative path from .pluqqy directory
	componentRelativePath := filepath.Join(
		"components", 
		componentType,
		componentName+".md",
	)
	
	// Full path for checking existence and display
	componentPath := filepath.Join(files.PluqqyDir, componentRelativePath)

	// Check if component already exists
	if _, err := os.Stat(componentPath); err == nil {
		return fmt.Errorf("component '%s' already exists", componentName)
	}

	// Create a temporary file for editing
	tmpFile, err := os.CreateTemp("", componentName+"_*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial template to temp file
	template := createComponentTemplate(componentName, componentType, createTags)
	if err := os.WriteFile(tmpFile.Name(), []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
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
		editorCmd = exec.Command(parts[0], append(parts[1:], tmpFile.Name())...)
	} else {
		// Simple editor command (e.g., "vim")
		editorCmd = exec.Command(editor, tmpFile.Name())
	}
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	cli.PrintInfo("Opening editor to create component...")
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited content: %w", err)
	}

	// Check if content is empty or unchanged
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("component content is empty, creation cancelled")
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(componentPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	// Write the component (use relative path)
	if err := files.WriteComponentWithNameAndTags(componentRelativePath, string(content), componentName, createTags); err != nil {
		return fmt.Errorf("failed to save component: %w", err)
	}

	cli.PrintSuccess("Created %s component: %s", componentType, componentName)
	cli.PrintInfo("Path: %s", componentPath)
	
	if len(createTags) > 0 {
		cli.PrintInfo("Tags: %s", strings.Join(createTags, ", "))
	}

	return nil
}

func createComponentTemplate(name string, componentType string, tags []string) string {
	var template strings.Builder

	// Add frontmatter with metadata
	template.WriteString("---\n")
	template.WriteString(fmt.Sprintf("name: %s\n", name))
	if len(tags) > 0 {
		template.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tags, ", ")))
	}
	template.WriteString("---\n\n")

	// Add helpful template based on type
	switch componentType {
	case "contexts":
		template.WriteString(fmt.Sprintf("# %s Context\n\n", name))
		template.WriteString("<!-- Provide context information that the AI should know -->\n\n")
		template.WriteString("## Overview\n\n")
		template.WriteString("<!-- Describe the context here -->\n\n")
		
	case "prompts":
		template.WriteString(fmt.Sprintf("# %s Prompt\n\n", name))
		template.WriteString("<!-- Define specific instructions or prompts for the AI -->\n\n")
		template.WriteString("## Instructions\n\n")
		template.WriteString("<!-- Write your prompt instructions here -->\n\n")
		
	case "rules":
		template.WriteString(fmt.Sprintf("# %s Rules\n\n", name))
		template.WriteString("<!-- Define rules and constraints for the AI to follow -->\n\n")
		template.WriteString("## Rules\n\n")
		template.WriteString("1. <!-- First rule -->\n")
		template.WriteString("2. <!-- Second rule -->\n")
		template.WriteString("3. <!-- Third rule -->\n\n")
	}

	return template.String()
}