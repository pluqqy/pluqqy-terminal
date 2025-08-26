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
			ctx, err := cli.NewCommandContext()
			if err != nil {
				return err
			}
			if err := ctx.ValidateProject(); err != nil {
				return err
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

	// Create template content
	template := createComponentTemplate(componentName, componentType, createTags)
	
	// Open in editor with template
	launcher := cli.NewEditorLauncher()
	cli.PrintInfo("Opening editor to create component...")
	tmpFilePath, err := launcher.OpenTempFile(componentName+"_*.md", template)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFilePath)

	// Read the edited content
	content, err := os.ReadFile(tmpFilePath)
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