package files

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	SettingsFile      = "settings.yaml"
	
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
	// Load settings to get output path
	settings := models.DefaultSettings()
	
	dirs := []string{
		PluqqyDir,
		filepath.Join(PluqqyDir, PipelinesDir),
		filepath.Join(PluqqyDir, ComponentsDir),
		filepath.Join(PluqqyDir, ComponentsDir, PromptsDir),
		filepath.Join(PluqqyDir, ComponentsDir, ContextsDir),
		filepath.Join(PluqqyDir, ComponentsDir, RulesDir),
		filepath.Join(PluqqyDir, ArchiveDir),
		filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir),
		filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir),
		filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, PromptsDir),
		filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, ContextsDir),
		filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, RulesDir),
		filepath.Join(PluqqyDir, strings.TrimSuffix(settings.Output.OutputPath, "/")), // tmp directory
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Create or update .gitignore to ignore tmp directory
	gitignorePath := filepath.Join(PluqqyDir, ".gitignore")
	gitignoreContent := "/tmp/\n"
	
	// Check if .gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		// Read existing content
		existing, err := os.ReadFile(gitignorePath)
		if err == nil {
			// Check if tmp is already ignored
			if !strings.Contains(string(existing), "/tmp/") {
				// Append to existing content
				gitignoreContent = string(existing)
				if !strings.HasSuffix(gitignoreContent, "\n") {
					gitignoreContent += "\n"
				}
				gitignoreContent += "/tmp/\n"
			} else {
				// Already contains /tmp/, don't modify
				gitignoreContent = ""
			}
		}
	}
	
	// Write .gitignore if needed
	if gitignoreContent != "" {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	return nil
}

// componentFrontmatter represents the YAML frontmatter in component files
type componentFrontmatter struct {
	Tags []string `yaml:"tags,omitempty"`
}

// extractFrontmatter extracts YAML frontmatter from markdown content
func extractFrontmatter(content []byte) (*componentFrontmatter, []byte, error) {
	if !bytes.HasPrefix(content, []byte("---")) {
		// No frontmatter
		return &componentFrontmatter{}, content, nil
	}
	
	// Find the end of frontmatter
	parts := bytes.SplitN(content, []byte("\n---\n"), 3)
	if len(parts) < 2 {
		// Malformed frontmatter
		return &componentFrontmatter{}, content, nil
	}
	
	// Parse the frontmatter (excluding the first ---)
	frontmatterBytes := bytes.TrimPrefix(parts[0], []byte("---\n"))
	
	var frontmatter componentFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
		// If parsing fails, just return empty frontmatter
		return &componentFrontmatter{}, content, nil
	}
	
	// Return the content without frontmatter
	remainingContent := content
	if len(parts) >= 2 {
		remainingContent = parts[1]
		if len(parts) == 3 {
			remainingContent = bytes.Join([][]byte{[]byte("---\n"), parts[2]}, nil)
		}
	}
	
	return &frontmatter, remainingContent, nil
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
	
	// Extract frontmatter to get tags
	frontmatter, _, _ := extractFrontmatter(content)
	
	return &models.Component{
		Path:     path,
		Type:     componentType,
		Content:  string(content), // Keep original content with frontmatter
		Modified: info.ModTime(),
		Tags:     frontmatter.Tags,
	}, nil
}

// formatComponentContent adds or updates frontmatter with tags
func formatComponentContent(content string, tags []string) string {
	contentBytes := []byte(content)
	frontmatter, contentWithoutFrontmatter, _ := extractFrontmatter(contentBytes)
	
	// Update tags if provided
	if tags != nil {
		frontmatter.Tags = tags
	}
	
	// If no tags and no existing frontmatter, return content as-is
	if len(frontmatter.Tags) == 0 && !bytes.HasPrefix(contentBytes, []byte("---")) {
		return content
	}
	
	// Build new content with frontmatter
	var buf bytes.Buffer
	
	// Write frontmatter if there are tags
	if len(frontmatter.Tags) > 0 {
		buf.WriteString("---\n")
		frontmatterBytes, _ := yaml.Marshal(frontmatter)
		buf.Write(frontmatterBytes)
		buf.WriteString("---\n")
	}
	
	// Write the content
	buf.Write(contentWithoutFrontmatter)
	
	return buf.String()
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

// WriteComponentWithTags writes a component file with tags
func WriteComponentWithTags(path string, content string, tags []string) error {
	formattedContent := formatComponentContent(content, tags)
	return WriteComponent(path, formattedContent)
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
	
	// Normalize component types for backward compatibility (singular -> plural)
	for i := range pipeline.Components {
		switch pipeline.Components[i].Type {
		case "context":
			pipeline.Components[i].Type = models.ComponentTypeContext
		case "prompt":
			pipeline.Components[i].Type = models.ComponentTypePrompt
		case "rule":
			pipeline.Components[i].Type = models.ComponentTypeRules
		// Already plural forms are correct - no change needed
		case "contexts", "prompts", "rules":
			// These are already in the correct format
		}
	}
	
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
	case models.ComponentTypePrompt:
		subDir = PromptsDir
	case models.ComponentTypeContext:
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

// ListArchivedPipelines returns a list of archived pipeline files
func ListArchivedPipelines() ([]string, error) {
	archivePipelinesPath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir)
	
	entries, err := os.ReadDir(archivePipelinesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read archived pipelines directory '%s': %w", archivePipelinesPath, err)
	}
	
	var pipelines []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			pipelines = append(pipelines, entry.Name())
		}
	}
	
	return pipelines, nil
}

// ListArchivedComponents returns a list of archived component files for a given type
func ListArchivedComponents(componentType string) ([]string, error) {
	var subDir string
	switch componentType {
	case models.ComponentTypePrompt:
		subDir = PromptsDir
	case models.ComponentTypeContext:
		subDir = ContextsDir
	case models.ComponentTypeRules:
		subDir = RulesDir
	default:
		return nil, fmt.Errorf("invalid component type '%s': must be one of: prompt, context, rules", componentType)
	}

	archiveComponentsPath := filepath.Join(PluqqyDir, ArchiveDir, ComponentsDir, subDir)
	
	entries, err := os.ReadDir(archiveComponentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read archived components directory '%s': %w", archiveComponentsPath, err)
	}

	var components []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			components = append(components, entry.Name())
		}
	}

	return components, nil
}

// ReadArchivedPipeline reads an archived pipeline file
func ReadArchivedPipeline(path string) (*models.Pipeline, error) {
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid pipeline path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, path)
	
	// Validate file size before reading
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archived pipeline not found at path '%s'", path)
		}
		return nil, fmt.Errorf("failed to stat archived pipeline file '%s': %w", path, err)
	}
	if fileInfo.Size() > MaxFileSize {
		return nil, fmt.Errorf("archived pipeline file '%s' is too large (%d bytes)", path, fileInfo.Size())
	}
	
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archived pipeline file '%s': %w", path, err)
	}
	
	var pipeline models.Pipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse archived pipeline file '%s': %w", path, err)
	}
	
	// Set the path
	pipeline.Path = path
	
	return &pipeline, nil
}

// ReadArchivedComponent reads an archived component file
func ReadArchivedComponent(path string) (*models.Component, error) {
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid component path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, ArchiveDir, path)
	
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archived component not found at path '%s'", path)
		}
		return nil, fmt.Errorf("failed to stat archived component file '%s': %w", path, err)
	}
	if fileInfo.Size() > MaxFileSize {
		return nil, fmt.Errorf("archived component file '%s' is too large (%d bytes)", path, fileInfo.Size())
	}
	
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archived component file '%s': %w", path, err)
	}
	
	// Extract frontmatter to get tags
	frontmatter, _, _ := extractFrontmatter(data)
	
	comp := &models.Component{
		Content:     string(data),
		Path:        path,
		Type:        getComponentType(path),
		Modified:    fileInfo.ModTime(),
		Tags:        frontmatter.Tags,
	}
	
	return comp, nil
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

// DeleteComponent removes a component file
func DeleteComponent(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid component path: %w", err)
	}
	
	absPath := filepath.Join(PluqqyDir, path)
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("component not found at path '%s'", path)
	}
	
	// Remove the file
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete component '%s': %w", path, err)
	}
	
	return nil
}

// ArchivePipeline moves a pipeline file to the archive directory
func ArchivePipeline(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid pipeline path: %w", err)
	}
	
	sourcePath := filepath.Join(PluqqyDir, PipelinesDir, path)
	archivePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, path)
	
	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("pipeline not found at path '%s'", path)
	}
	
	// Create archive directory if it doesn't exist
	archiveDir := filepath.Dir(archivePath)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	
	// Move the file
	if err := os.Rename(sourcePath, archivePath); err != nil {
		return fmt.Errorf("failed to archive pipeline '%s': %w", path, err)
	}
	
	return nil
}

// ArchiveComponent moves a component file to the archive directory
func ArchiveComponent(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid component path: %w", err)
	}
	
	sourcePath := filepath.Join(PluqqyDir, path)
	archivePath := filepath.Join(PluqqyDir, ArchiveDir, path)
	
	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("component not found at path '%s'", path)
	}
	
	// Create archive directory if it doesn't exist
	archiveDir := filepath.Dir(archivePath)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	
	// Move the file
	if err := os.Rename(sourcePath, archivePath); err != nil {
		return fmt.Errorf("failed to archive component '%s': %w", path, err)
	}
	
	return nil
}

// ReadSettings reads the settings file
func ReadSettings() (*models.Settings, error) {
	settingsPath := filepath.Join(PluqqyDir, SettingsFile)
	
	// Check if settings file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// Return default settings if file doesn't exist
		return models.DefaultSettings(), nil
	}
	
	// Read the file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}
	
	// Parse YAML
	var settings models.Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}
	
	// Merge with defaults to ensure all fields are populated
	defaults := models.DefaultSettings()
	mergeSettings(&settings, defaults)
	
	return &settings, nil
}

// WriteSettings writes the settings file
func WriteSettings(settings *models.Settings) error {
	settingsPath := filepath.Join(PluqqyDir, SettingsFile)
	
	// Marshal to YAML
	data, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	
	// Write atomically
	if err := writeFileAtomic(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	
	return nil
}

// mergeSettings fills in any missing values from defaults
func mergeSettings(settings *models.Settings, defaults *models.Settings) {
	// Merge output settings
	if settings.Output.DefaultFilename == "" {
		settings.Output.DefaultFilename = defaults.Output.DefaultFilename
	}
	if settings.Output.ExportPath == "" {
		settings.Output.ExportPath = defaults.Output.ExportPath
	}
	
	// Merge sections configuration
	if len(settings.Output.Formatting.Sections) == 0 {
		settings.Output.Formatting.Sections = defaults.Output.Formatting.Sections
	}
}

// CountComponentUsage returns a map of component paths to their usage count across all pipelines
func CountComponentUsage() (map[string]int, error) {
	usageCount := make(map[string]int)
	
	// List all pipelines
	pipelines, err := ListPipelines()
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	
	// For each pipeline, count component usage
	for _, pipelinePath := range pipelines {
		pipeline, err := ReadPipeline(pipelinePath)
		if err != nil {
			// Skip pipelines that can't be read
			continue
		}
		
		// Count each component reference
		for _, comp := range pipeline.Components {
			// Normalize the path to match how components are stored
			normalizedPath := filepath.Clean(comp.Path)
			usageCount[normalizedPath]++
		}
	}
	
	return usageCount, nil
}

// GetComponentStats returns detailed stats for a component including last modified time
func GetComponentStats(componentPath string) (time.Time, error) {
	absPath := filepath.Join(PluqqyDir, componentPath)
	
	info, err := os.Stat(absPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get component stats: %w", err)
	}
	
	return info.ModTime(), nil
}

// UpdateComponentTags updates only the tags of a component without modifying its content
func UpdateComponentTags(path string, tags []string) error {
	// Read the component
	component, err := ReadComponent(path)
	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}
	
	// Update the content with new tags
	updatedContent := formatComponentContent(component.Content, tags)
	
	// Write back
	return WriteComponent(path, updatedContent)
}

// AddComponentTag adds a single tag to a component
func AddComponentTag(path string, tag string) error {
	component, err := ReadComponent(path)
	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}
	
	// Check if tag already exists
	for _, t := range component.Tags {
		if t == tag {
			return nil // Tag already exists
		}
	}
	
	// Add the tag
	newTags := append(component.Tags, tag)
	return UpdateComponentTags(path, newTags)
}

// RemoveComponentTag removes a single tag from a component
func RemoveComponentTag(path string, tag string) error {
	component, err := ReadComponent(path)
	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}
	
	// Filter out the tag
	newTags := make([]string, 0, len(component.Tags))
	for _, t := range component.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	
	return UpdateComponentTags(path, newTags)
}