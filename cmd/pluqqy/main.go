package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user/pluqqy/pkg/files"
	"github.com/user/pluqqy/pkg/tui"
)

var rootCmd = &cobra.Command{
	Use:   "pluqqy",
	Short: "Terminal-based tool for managing LLM prompt pipelines",
	Long:  `Pluqqy is a terminal-based tool for managing LLM prompt pipelines. It stores everything as plain text files (Markdown and YAML) and provides a TUI for easy interaction.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if .pluqqy directory exists
		if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
			fmt.Println("No .pluqqy directory found. Run 'pluqqy init' first.")
			os.Exit(1)
		}

		// Launch TUI
		app := tui.NewApp()
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Pluqqy project",
	Long:  `Creates the .pluqqy folder structure in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Pluqqy project...")
		if err := files.InitProjectStructure(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created .pluqqy folder structure")
		fmt.Println("You can now create pipelines and components!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}