package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

const (
	PluqqyDir         = ".pluqqy"
	PipelinesDir      = "pipelines"
	ComponentsDir     = "components"
	PromptsDir        = "prompts"
	ContextsDir       = "contexts"
	RulesDir          = "rules"
	ArchiveDir        = "archive"
	DefaultOutputFile = "PLUQQY.md"
	
	// MaxFileSize is the maximum size for component and pipeline files (10MB)
	MaxFileSize = 10 * 1024 * 1024
)

// validatePath ensures the path doesn't contain directory traversal attempts
func validatePath(path string) error {
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("invalid path: contains directory traversal")
	}
	return nil
}

// validateFileSize checks if the file size is within acceptable limits
func validateFileSize(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	if info.Size() > MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes", info.Size(), MaxFileSize)
	}
	
	return nil
}

func InitProjectStructure() error {
	dirs := []string{
		PluqqyDir,
		filepath.Join(PluqqyDir, PipelinesDir),
		filepath.Join(PluqqyDir, ComponentsDir),
		filepath.Join(PluqqyDir, ComponentsDir, PromptsDir),
		filepath.Join(PluqqyDir, ComponentsDir, ContextsDir),
		filepath.Join(PluqqyDir, ComponentsDir, RulesDir),
		filepath.Join(PluqqyDir, ArchiveDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func ReadComponent(path string) (*models.Component, error) {
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid component path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, path)
	
	// Validate file size before reading
	if err := validateFileSize(absPath); err != nil {
		return nil, fmt.Errorf("component file validation failed: %w", err)
	}
	
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("component not found at path '%s': file does not exist", path)
		}
		return nil, fmt.Errorf("failed to read component file '%s': %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for component '%s': %w", path, err)
	}

	componentType := getComponentType(path)
	
	return &models.Component{
		Path:     path,
		Type:     componentType,
		Content:  string(content),
		Modified: info.ModTime(),
	}, nil
}

func WriteComponent(path string, content string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid component path: %w", err)
	}
	
	// Validate content size
	if len(content) > MaxFileSize {
		return fmt.Errorf("content size %d bytes exceeds maximum allowed size of %d bytes", len(content), MaxFileSize)
	}
	
	absPath := filepath.Join(PluqqyDir, path)
	
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create component directory '%s': %w", dir, err)
	}

	if err := writeFileAtomic(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write component file '%s': %w", path, err)
	}

	return nil
}

func ReadPipeline(path string) (*models.Pipeline, error) {
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid pipeline path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, PipelinesDir, path)
	
	// Validate file size before reading
	if err := validateFileSize(absPath); err != nil {
		return nil, fmt.Errorf("pipeline file validation failed: %w", err)
	}
	
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pipeline not found at path '%s': file does not exist", path)
		}
		return nil, fmt.Errorf("failed to read pipeline file '%s': %w", path, err)
	}

	var pipeline models.Pipeline
	if err := yaml.Unmarshal(content, &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in pipeline '%s': %w", path, err)
	}

	pipeline.Path = path
	
	return &pipeline, nil
}

func WritePipeline(pipeline *models.Pipeline) error {
	// Validate pipeline before writing
	if err := pipeline.Validate(); err != nil {
		return fmt.Errorf("invalid pipeline: %w", err)
	}
	
	if pipeline.Path == "" {
		pipeline.Path = fmt.Sprintf("%s.yaml", pipeline.Name)
	}

	if err := validatePath(pipeline.Path); err != nil {
		return fmt.Errorf("invalid pipeline path: %w", err)
	}

	absPath := filepath.Join(PluqqyDir, PipelinesDir, pipeline.Path)
	
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create pipeline directory '%s': %w", dir, err)
	}

	content, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to serialize pipeline '%s' to YAML: %w", pipeline.Name, err)
	}

	if err := writeFileAtomic(absPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write pipeline file '%s': %w", pipeline.Path, err)
	}

	return nil
}

func ListPipelines() ([]string, error) {
	pipelinesPath := filepath.Join(PluqqyDir, PipelinesDir)
	
	entries, err := os.ReadDir(pipelinesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read pipelines directory '%s': %w", pipelinesPath, err)
	}

	var pipelines []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			pipelines = append(pipelines, entry.Name())
		}
	}

	return pipelines, nil
}

func ListComponents(componentType string) ([]string, error) {
	var subDir string
	switch componentType {
	case models.ComponentTypePrompt, "prompts":
		subDir = PromptsDir
	case models.ComponentTypeContext, "contexts":
		subDir = ContextsDir
	case models.ComponentTypeRules:
		subDir = RulesDir
	default:
		return nil, fmt.Errorf("invalid component type '%s': must be one of: prompt, context, rules", componentType)
	}

	componentsPath := filepath.Join(PluqqyDir, ComponentsDir, subDir)
	
	entries, err := os.ReadDir(componentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read components directory '%s': %w", componentsPath, err)
	}

	var components []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			components = append(components, entry.Name())
		}
	}

	return components, nil
}

func getComponentType(path string) string {
	if strings.Contains(path, PromptsDir) {
		return models.ComponentTypePrompt
	} else if strings.Contains(path, ContextsDir) {
		return models.ComponentTypeContext
	} else if strings.Contains(path, RulesDir) {
		return models.ComponentTypeRules
	}
	return "unknown"
}

// WriteFile writes content to a file (for PLUQQY.md output)
func WriteFile(path string, content string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid output file path: %w", err)
	}
	
	// Validate content size
	if len(content) > MaxFileSize {
		return fmt.Errorf("content size %d bytes exceeds maximum allowed size of %d bytes", len(content), MaxFileSize)
	}
	
	if err := writeFileAtomic(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output file '%s': %w", path, err)
	}
	return nil
}

// writeFileAtomic writes data to a file atomically by writing to a temp file first
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	// Create temp file in the same directory as target
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	
	// Ensure temp file is cleaned up
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()
	
	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	
	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	
	// Close the temp file before renaming
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	
	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file to target: %w", err)
	}
	
	return nil
}

// DeletePipeline removes a pipeline file
func DeletePipeline(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid pipeline path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, PipelinesDir, path)
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("pipeline not found at path '%s'", path)
	}
	
	// Remove the file
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete pipeline '%s': %w", path, err)
	}
	
	return nil
}