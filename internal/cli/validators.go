package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateComponentType validates a component type string
func ValidateComponentType(t string) error {
	validTypes := []string{"context", "contexts", "prompt", "prompts", "rule", "rules"}
	normalized := strings.ToLower(t)
	
	for _, valid := range validTypes {
		if normalized == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid component type: %s (must be: context, prompt, or rule)", t)
}

// NormalizeComponentType converts type variants to standard form
func NormalizeComponentType(t string) string {
	normalized := strings.ToLower(t)
	switch normalized {
	case "context", "contexts":
		return "contexts"
	case "prompt", "prompts":
		return "prompts"
	case "rule", "rules":
		return "rules"
	default:
		return "contexts" // default fallback
	}
}

// ValidateFilePath validates that a file path exists and is a file
func ValidateFilePath(path string) error {
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("error accessing path: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, expected file: %s", path)
	}

	return nil
}

// ValidateDirectoryPath validates that a directory path exists
func ValidateDirectoryPath(path string) error {
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("error accessing directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// ValidateOutputFormat validates the output format flag
func ValidateOutputFormat(format string) error {
	validFormats := []string{"text", "json", "yaml"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format: %s (must be: text, json, or yaml)", format)
}

// ValidatePipelineName validates a pipeline name
func ValidatePipelineName(name string) error {
	if name == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", "..", "~", "$", "`"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("pipeline name contains invalid character: %s", char)
		}
	}
	
	return nil
}

// ValidateComponentName validates a component name
func ValidateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", "..", "~", "$", "`"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("component name contains invalid character: %s", char)
		}
	}
	
	return nil
}

// Contains checks if a string is in a slice
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}