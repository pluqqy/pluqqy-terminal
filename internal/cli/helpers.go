package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for confirmation
func Confirm(prompt string, defaultYes bool) (bool, error) {
	if skipConfirm {
		return true, nil
	}

	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	fmt.Print(prompt + suffix)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "" {
		return defaultYes, nil
	}

	return response == "y" || response == "yes", nil
}

// PrintSuccess prints a success message unless quiet mode is enabled
func PrintSuccess(format string, args ...interface{}) {
	if !quiet {
		msg := fmt.Sprintf(format, args...)
		if !noColor {
			fmt.Printf("✓ %s\n", msg)
		} else {
			fmt.Printf("OK: %s\n", msg)
		}
	}
}

// PrintInfo prints an info message unless quiet mode is enabled
func PrintInfo(format string, args ...interface{}) {
	if !quiet {
		msg := fmt.Sprintf(format, args...)
		if !noColor {
			fmt.Printf("ℹ %s\n", msg)
		} else {
			fmt.Printf("INFO: %s\n", msg)
		}
	}
}

// PrintWarning prints a warning message to stderr
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", msg)
	}
}

// PrintError prints an error message to stderr
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !noColor {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
	}
}

// Global flags (will be set from cmd package)
var (
	quiet       bool
	noColor     bool
	skipConfirm bool
)

// SetGlobalFlags sets the global flag values from the cmd package
func SetGlobalFlags(q, nc, sc bool) {
	quiet = q
	noColor = nc
	skipConfirm = sc
}