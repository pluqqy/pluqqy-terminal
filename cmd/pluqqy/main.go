package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/tui"
)

// Version is set during build with -ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "pluqqy",
	Short: "Terminal-based tool for managing LLM prompt pipelines",
	Long:  `Pluqqy is a terminal-based tool for managing LLM prompt pipelines. It stores everything as plain text files (Markdown and YAML) and provides a TUI for easy interaction.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if .pluqqy directory exists
		if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: No .pluqqy directory found in the current directory.\n")
			fmt.Fprintf(os.Stderr, "Please run 'pluqqy init' first to initialize a new project.\n")
			os.Exit(1)
		}

		// Launch TUI
		app := tui.NewApp()
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to start the terminal user interface: %v\n", err)
			fmt.Fprintf(os.Stderr, "This could be due to terminal compatibility issues. Try running in a different terminal.\n")
			os.Exit(1)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Pluqqy project",
	Long:  `Creates the .pluqqy folder structure in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to determine current directory: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Initializing Pluqqy project in %s...\n", cwd)
		
		if err := files.InitProjectStructure(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize project structure: %v\n", err)
			fmt.Fprintf(os.Stderr, "Make sure you have write permissions in the current directory.\n")
			os.Exit(1)
		}
		
		fmt.Println("✓ Created .pluqqy folder structure")
		fmt.Println("✓ You can now create pipelines and components!")
		fmt.Println("\nRun 'pluqqy' to start the interactive TUI.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pluqqy",
	Long:  `Display the current version of the Pluqqy CLI tool`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pluqqy version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Command execution failed: %v\n", err)
		os.Exit(1)
	}
}