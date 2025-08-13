package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// SanitizeFileName is an alias for Slugify - converts a display name to a safe filename
func SanitizeFileName(displayName string) string {
	return Slugify(displayName)
}

// Slugify converts a display name to a valid filename
// Examples:
//   "Auth Context" → "auth-context"
//   "User's Profile!" → "users-profile"
//   "My Component #1" → "my-component-1"
func Slugify(displayName string) string {
	// Convert to lowercase
	slug := strings.ToLower(displayName)
	
	// Replace any non-alphanumeric characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Replace multiple consecutive hyphens with a single hyphen
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	
	// Ensure we have a valid filename
	if slug == "" {
		slug = "unnamed"
	}
	
	return slug
}

// ExtractDisplayName extracts a display name from a filename
// Examples:
//   "auth-context.md" → "Auth Context"
//   "users-profile.yaml" → "Users Profile"
func ExtractDisplayName(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Split by hyphens
	parts := strings.Split(name, "-")
	
	// Capitalize each part
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	
	// Join with spaces
	return strings.Join(parts, " ")
}

// updateMarkdownDisplayName updates the first # header in markdown content
func updateMarkdownDisplayName(content, newDisplayName string) string {
	lines := strings.Split(content, "\n")
	headerUpdated := false
	
	for i, line := range lines {
		// Find the first # header (not ##, ###, etc.)
		if strings.HasPrefix(strings.TrimSpace(line), "# ") && !headerUpdated {
			lines[i] = "# " + newDisplayName
			headerUpdated = true
			break
		}
	}
	
	// If no header was found, add one after frontmatter (if exists)
	if !headerUpdated {
		// Check if content has frontmatter
		if strings.HasPrefix(content, "---") {
			// Find end of frontmatter
			inFrontmatter := false
			for i, line := range lines {
				if i == 0 && line == "---" {
					inFrontmatter = true
					continue
				}
				if inFrontmatter && line == "---" {
					// Insert header after frontmatter
					newLines := append(lines[:i+1], append([]string{"", "# " + newDisplayName, ""}, lines[i+1:]...)...)
					return strings.Join(newLines, "\n")
				}
			}
		}
		// No frontmatter, add header at the beginning
		return "# " + newDisplayName + "\n\n" + content
	}
	
	return strings.Join(lines, "\n")
}

// ExtractMarkdownDisplayName extracts the display name from markdown content
func ExtractMarkdownDisplayName(content string) string {
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		// Find the first # header (not ##, ###, etc.)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			// Return the header text without the "# " prefix, trimmed
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	
	// No header found, return empty
	return ""
}

// RenameComponent renames a component file and updates all references
func RenameComponent(oldPath, newDisplayName string) error {
	// Validate inputs
	if err := validatePath(oldPath); err != nil {
		return fmt.Errorf("invalid old path: %w", err)
	}
	
	if newDisplayName == "" {
		return fmt.Errorf("new display name cannot be empty")
	}
	
	// Generate new filename
	newSlug := Slugify(newDisplayName)
	dir := filepath.Dir(oldPath)
	ext := filepath.Ext(oldPath)
	newPath := filepath.Join(dir, newSlug+ext)
	
	// Build absolute paths for file operations
	absOldPath := filepath.Join(PluqqyDir, oldPath)
	absNewPath := filepath.Join(PluqqyDir, newPath)
	
	// Check if target already exists
	if oldPath != newPath {
		if _, err := os.Stat(absNewPath); err == nil {
			return fmt.Errorf("component with name '%s' already exists", newSlug)
		}
	}
	
	// Read the component using the relative path
	component, err := ReadComponent(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}
	
	// Create a backup for rollback using absolute path
	backupContent, err := os.ReadFile(absOldPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Write to new path with updated name in frontmatter
	if err := WriteComponentWithNameAndTags(newPath, component.Content, newDisplayName, component.Tags); err != nil {
		return fmt.Errorf("failed to write renamed component: %w", err)
	}
	
	// If paths are different, delete the old file (use absolute path)
	if oldPath != newPath {
		if err := os.Remove(absOldPath); err != nil {
			// Rollback: delete new file and restore old
			os.Remove(absNewPath)
			return fmt.Errorf("failed to remove old file: %w", err)
		}
	}
	
	// Update references in pipelines (uses relative paths)
	if err := UpdateComponentReferences(oldPath, newPath, newDisplayName); err != nil {
		// Rollback: restore old file and remove new (use absolute paths)
		os.WriteFile(absOldPath, backupContent, 0644)
		if oldPath != newPath {
			os.Remove(absNewPath)
		}
		return fmt.Errorf("failed to update references: %w", err)
	}
	
	return nil
}

// RenamePipeline renames a pipeline file
func RenamePipeline(oldPath, newDisplayName string) error {
	// Validate inputs
	if err := validatePath(oldPath); err != nil {
		return fmt.Errorf("invalid old path: %w", err)
	}
	
	if newDisplayName == "" {
		return fmt.Errorf("new display name cannot be empty")
	}
	
	// oldPath is just the filename, we need to handle it properly
	oldFilename := oldPath
	newSlug := Slugify(newDisplayName)
	ext := filepath.Ext(oldFilename)
	newFilename := newSlug + ext
	
	// Build absolute paths for file operations (add PipelinesDir)
	absOldPath := filepath.Join(PluqqyDir, PipelinesDir, oldFilename)
	absNewPath := filepath.Join(PluqqyDir, PipelinesDir, newFilename)
	
	// Check if target already exists
	if oldFilename != newFilename {
		if _, err := os.Stat(absNewPath); err == nil {
			return fmt.Errorf("pipeline with name '%s' already exists", newSlug)
		}
	}
	
	// Read the pipeline (ReadPipeline expects just filename)
	pipeline, err := ReadPipeline(oldFilename)
	if err != nil {
		return fmt.Errorf("failed to read pipeline: %w", err)
	}
	
	// Update the display name in the pipeline
	pipeline.Name = newDisplayName
	
	// Create backup for rollback using absolute path
	backupContent, err := os.ReadFile(absOldPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Set the new path for writing
	originalPath := pipeline.Path
	pipeline.Path = newFilename
	
	// Write to new location
	if err := WritePipeline(pipeline); err != nil {
		// Restore original path before returning error
		pipeline.Path = originalPath
		return fmt.Errorf("failed to write renamed pipeline: %w", err)
	}
	
	// If filenames are different, delete the old file (use absolute path)
	if oldFilename != newFilename {
		if err := os.Remove(absOldPath); err != nil {
			// Rollback: restore old file and remove new (use absolute paths)
			os.WriteFile(absOldPath, backupContent, 0644)
			os.Remove(absNewPath)
			return fmt.Errorf("failed to remove old file: %w", err)
		}
	}
	
	return nil
}

// ValidateRename checks if a rename operation would be valid
func ValidateRename(oldPath, newDisplayName, itemType string) error {
	// Validate inputs
	if err := validatePath(oldPath); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	
	if newDisplayName == "" {
		return fmt.Errorf("name cannot be empty")
	}
	
	// Check for invalid characters that can't be slugified nicely
	if regexp.MustCompile(`^[^a-zA-Z0-9\s\-_#]+$`).MatchString(newDisplayName) {
		return fmt.Errorf("name contains only special characters")
	}
	
	// Generate the slug to check for conflicts
	newSlug := Slugify(newDisplayName)
	dir := filepath.Dir(oldPath)
	ext := filepath.Ext(oldPath)
	newPath := filepath.Join(dir, newSlug+ext)
	
	// Check if it would conflict with existing file
	if oldPath != newPath {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("%s with name '%s' already exists", itemType, newSlug)
		}
	}
	
	return nil
}

// RenameComponentInArchive renames an archived component
func RenameComponentInArchive(oldPath, newDisplayName string) error {
	// The oldPath should be relative to .pluqqy directory
	// If it contains archive/, use it as is; otherwise add archive/ prefix
	relativePath := oldPath
	if !strings.Contains(oldPath, ArchiveDir) {
		// Add archive prefix if not present
		relativePath = filepath.Join(ArchiveDir, oldPath)
	}
	
	// Call regular rename with the archive path
	return RenameComponent(relativePath, newDisplayName)
}

// RenamePipelineInArchive renames an archived pipeline
func RenamePipelineInArchive(oldPath, newDisplayName string) error {
	// oldPath is just the filename for pipelines
	// Need to handle archived pipelines differently
	
	// Validate inputs
	if err := validatePath(oldPath); err != nil {
		return fmt.Errorf("invalid old path: %w", err)
	}
	
	if newDisplayName == "" {
		return fmt.Errorf("new display name cannot be empty")
	}
	
	// oldPath is just the filename
	oldFilename := oldPath
	newSlug := Slugify(newDisplayName)
	ext := filepath.Ext(oldFilename)
	newFilename := newSlug + ext
	
	// Build absolute paths for archived pipelines
	absOldPath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, oldFilename)
	absNewPath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, newFilename)
	
	// Check if target already exists
	if oldFilename != newFilename {
		if _, err := os.Stat(absNewPath); err == nil {
			return fmt.Errorf("archived pipeline with name '%s' already exists", newSlug)
		}
	}
	
	// Read the archived pipeline
	pipeline, err := ReadArchivedPipeline(oldFilename)
	if err != nil {
		return fmt.Errorf("failed to read archived pipeline: %w", err)
	}
	
	// Update the display name
	pipeline.Name = newDisplayName
	
	// Create backup
	backupContent, err := os.ReadFile(absOldPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// We need to write the pipeline to the new location
	// WritePipeline expects the path to be set relative to PipelinesDir
	// For archived pipelines, we need to write directly
	pipelineData, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline: %w", err)
	}
	
	// Write to new location
	if err := os.WriteFile(absNewPath, pipelineData, 0644); err != nil {
		return fmt.Errorf("failed to write renamed archived pipeline: %w", err)
	}
	
	// If filenames are different, delete the old file
	if oldFilename != newFilename {
		if err := os.Remove(absOldPath); err != nil {
			// Rollback
			os.WriteFile(absOldPath, backupContent, 0644)
			os.Remove(absNewPath)
			return fmt.Errorf("failed to remove old file: %w", err)
		}
	}
	
	return nil
}