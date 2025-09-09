package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/pkg/examples"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

func NewExamplesCommand() *cobra.Command {
	var category string
	var listOnly bool
	var force bool

	cmd := &cobra.Command{
		Use:   "examples [category]",
		Short: "Add example components and pipelines to your project",
		Long: `Add example components and pipelines to your .pluqqy directory.

Examples demonstrate useful patterns and approaches that have worked well
for using Pluqqy with AI coding assistants. They include templates with
placeholders that you can customize for your specific needs.

Categories:
  general       - General-purpose development examples (default)
  web          - Web development (React, APIs, databases)
  ai           - AI assistant optimization patterns
  claude       - CLAUDE.md distiller for migrating existing files
  all          - Install all example categories

The examples will be added to your .pluqqy directory with an 'example-' prefix
to distinguish them from your own components.`,
		Example: `  # Add general examples
  pluqqy examples
  
  # Add web development examples
  pluqqy examples web
  
  # List available examples without installing
  pluqqy examples --list
  
  # Add all examples
  pluqqy examples all
  
  # Force overwrite existing examples
  pluqqy examples general --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}

			// Determine category
			if len(args) > 0 {
				category = args[0]
			} else if category == "" {
				// For listing, show all categories by default
				// For installing, use general as default
				if listOnly {
					category = "all"
				} else {
					category = "general"
				}
			}

			// Validate category
			validCategories := []string{"general", "web", "ai", "claude", "all"}
			isValid := false
			for _, valid := range validCategories {
				if category == valid {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("invalid category '%s'. Valid categories: %s", 
					category, strings.Join(validCategories, ", "))
			}

			// List mode
			if listOnly {
				return listExamples(cmd, category)
			}

			// Install examples
			return installExamples(cmd, category, force)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Category of examples to add")
	cmd.Flags().BoolVarP(&listOnly, "list", "l", false, "List available examples without installing")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing example files")

	return cmd
}

func listExamples(cmd *cobra.Command, category string) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	
	if !quiet {
		if category == "all" {
			fmt.Fprintf(cmd.OutOrStdout(), "Available examples (all categories):\n\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Available examples in category '%s':\n\n", category)
		}
	}

	// Get examples for the category
	exampleSets := examples.GetExamples(category)
	
	for _, set := range exampleSets {
		if !quiet {
			if category == "all" && set.Category != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "📦 [%s] %s\n", set.Category, set.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "📦 %s\n", set.Name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "   %s\n\n", set.Description)
		}
		
		if len(set.Components) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Components:\n")
			for _, comp := range set.Components {
				fmt.Fprintf(cmd.OutOrStdout(), "   • %s (%s)\n", comp.Name, comp.Type)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
		
		if len(set.Pipelines) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Pipelines:\n")
			for _, pipeline := range set.Pipelines {
				fmt.Fprintf(cmd.OutOrStdout(), "   • %s\n", pipeline.Name)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	if !quiet {
		if category == "all" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nTo install all examples, run: pluqqy examples all\n")
			fmt.Fprintf(cmd.OutOrStdout(), "To install a specific category, run: pluqqy examples <category>\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "\nTo install these examples, run: pluqqy examples %s\n", category)
		}
	}

	return nil
}

func installExamples(cmd *cobra.Command, category string, force bool) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	
	if !quiet {
		fmt.Fprintf(cmd.OutOrStdout(), "Installing %s examples...\n\n", category)
	}

	// Get examples for the category
	exampleSets := examples.GetExamples(category)
	
	totalComponents := 0
	totalPipelines := 0
	skipped := 0
	
	for _, set := range exampleSets {
		if !quiet {
			fmt.Fprintf(cmd.OutOrStdout(), "📦 Installing %s...\n", set.Name)
		}
		
		// Install components
		for _, comp := range set.Components {
			installed, err := examples.InstallComponent(comp, force)
			if err != nil {
				if !force && strings.Contains(err.Error(), "already exists") {
					skipped++
					if !quiet {
						fmt.Fprintf(cmd.OutOrStdout(), "   ⚠️  Skipped %s (already exists, use --force to overwrite)\n", comp.Name)
					}
					continue
				}
				return fmt.Errorf("failed to install component %s: %w", comp.Name, err)
			}
			
			if installed {
				totalComponents++
				if !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "   ✓ Installed %s/%s\n", comp.Type, comp.Filename)
				}
			}
		}
		
		// Install pipelines
		for _, pipeline := range set.Pipelines {
			installed, err := examples.InstallPipeline(pipeline, force)
			if err != nil {
				if !force && strings.Contains(err.Error(), "already exists") {
					skipped++
					if !quiet {
						fmt.Fprintf(cmd.OutOrStdout(), "   ⚠️  Skipped %s (already exists, use --force to overwrite)\n", pipeline.Name)
					}
					continue
				}
				return fmt.Errorf("failed to install pipeline %s: %w", pipeline.Name, err)
			}
			
			if installed {
				totalPipelines++
				if !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "   ✓ Installed pipeline %s\n", pipeline.Filename)
				}
			}
		}
		
		if !quiet {
			fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	// Summary
	if !quiet {
		fmt.Fprintf(cmd.OutOrStdout(), "✨ Installation complete!\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Installed:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • %d components\n", totalComponents)
		fmt.Fprintf(cmd.OutOrStdout(), "  • %d pipelines\n", totalPipelines)
		
		if skipped > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  • %d items skipped (already exist)\n", skipped)
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "\n💡 Tips:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Run 'pluqqy' to explore the examples in the TUI\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Look for items with 'example-' prefix\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Customize the templates by replacing {{PLACEHOLDERS}}\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Use 'pluqqy set example-<pipeline>' to activate a pipeline\n")
		
		// Special message for claude category
		if category == "claude" || category == "all" {
			fmt.Fprintf(cmd.OutOrStdout(), "\n🔄 CLAUDE.md Migration:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  • Use 'pluqqy set example-claude-distiller' to convert CLAUDE.md files\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  • The pipeline will help extract components from existing files\n")
		}
	}

	return nil
}