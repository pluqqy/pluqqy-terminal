package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"gopkg.in/yaml.v3"
)

// DeleteArchivedPipeline deletes an archived pipeline
func DeleteArchivedPipeline(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid pipeline path: %w", err)
	}
	
	// Use archive path: .pluqqy/archive/pipelines/xxx.yaml
	absPath := filepath.Join(PluqqyDir, "archive", PipelinesDir, path)
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("archived pipeline not found at path '%s'", path)
	}
	
	// Delete the file
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete archived pipeline '%s': %w", path, err)
	}
	
	return nil
}

// DeleteArchivedComponent deletes an archived component
func DeleteArchivedComponent(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid component path: %w", err)
	}
	
	// Use archive path: .pluqqy/archive/components/xxx/yyy.md
	absPath := filepath.Join(PluqqyDir, "archive", path)
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("archived component not found at path '%s'", path)
	}
	
	// Delete the file
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete archived component '%s': %w", path, err)
	}
	
	return nil
}

// GetFullPath returns the full absolute path for an item, considering if it's archived
func GetFullPath(relativePath string, isArchived bool, isComponent bool) string {
	if isArchived {
		if isComponent {
			return filepath.Join(PluqqyDir, "archive", relativePath)
		}
		return filepath.Join(PluqqyDir, "archive", PipelinesDir, relativePath)
	}
	
	if isComponent {
		return filepath.Join(PluqqyDir, relativePath)
	}
	return filepath.Join(PluqqyDir, PipelinesDir, relativePath)
}

// ReadArchivedOrActiveComponent reads a component from either archive or active location
func ReadArchivedOrActiveComponent(path string, isArchived bool) (*models.Component, error) {
	if isArchived {
		return ReadArchivedComponent(path)
	}
	return ReadComponent(path)
}

// ReadArchivedOrActivePipeline reads a pipeline from either archive or active location
func ReadArchivedOrActivePipeline(path string, isArchived bool) (*models.Pipeline, error) {
	if isArchived {
		return ReadArchivedPipeline(path)
	}
	return ReadPipeline(path)
}

// WriteComponentToArchive writes a component directly to the archive
func WriteComponentToArchive(path string, content string) error {
	// Ensure path doesn't include the archive directory
	cleanPath := path
	if strings.HasPrefix(path, "archive/") {
		cleanPath = strings.TrimPrefix(path, "archive/")
	}
	
	// Build the full archive path
	fullPath := filepath.Join(PluqqyDir, "archive", cleanPath)
	
	// Create the directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	
	// Write the file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write archived component: %w", err)
	}
	
	return nil
}

// WritePipelineToArchive writes a pipeline directly to the archive
func WritePipelineToArchive(pipeline *models.Pipeline) error {
	// Generate YAML content
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline: %w", err)
	}
	
	// Build the full archive path
	filename := Slugify(pipeline.Name) + ".yaml"
	fullPath := filepath.Join(PluqqyDir, "archive", PipelinesDir, filename)
	
	// Create the directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	
	// Write the file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write archived pipeline: %w", err)
	}
	
	// Update the pipeline path to reflect the new location
	pipeline.Path = filename
	
	return nil
}