package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/pluqqy/pkg/files"
)

var rootCmd = &cobra.Command{
	Use:   "pluqqy",
	Short: "Terminal-based tool for managing LLM prompt pipelines",
	Long:  `Pluqqy is a terminal-based tool for managing LLM prompt pipelines. It stores everything as plain text files (Markdown and YAML) and provides a TUI for easy interaction.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pluqqy TUI will be implemented soon!")
		fmt.Println("Use 'pluqqy init' to initialize a new project")
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