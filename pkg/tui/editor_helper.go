package tui

import (
	"os/exec"
	"strings"
)

// parseEditor splits the EDITOR environment variable into command and arguments
// This allows support for editors with flags like "code --wait" or "zed -w"
func parseEditor(editor string) (string, []string) {
	// Handle quoted strings properly
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := parts[0]
	args := parts[1:]

	return cmd, args
}

// createEditorCommand creates an exec.Command with proper argument handling
func createEditorCommand(editor string, filepath string) *exec.Cmd {
	cmd, args := parseEditor(editor)

	// Append the filepath to the arguments
	allArgs := append(args, filepath)

	return exec.Command(cmd, allArgs...)
}
