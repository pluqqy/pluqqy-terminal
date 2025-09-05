package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
	"gopkg.in/yaml.v3"
)

// TestEnvironment provides a complete test environment for TUI tests
type TestEnvironment struct {
	t          *testing.T
	TempDir    string
	OriginalWd string
	cleanup    []func()
}

// NewTestEnvironment creates a new test environment with a temporary directory
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()
	
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	env := &TestEnvironment{
		t:          t,
		TempDir:    tmpDir,
		OriginalWd: originalWd,
		cleanup:    []func(){},
	}
	
	// Add cleanup to restore working directory and remove temp dir
	env.cleanup = append(env.cleanup, func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	})
	
	return env
}

// Cleanup performs all cleanup operations
func (e *TestEnvironment) Cleanup() {
	for i := len(e.cleanup) - 1; i >= 0; i-- {
		e.cleanup[i]()
	}
}

// ChangeToTempDir changes the working directory to the temp directory
func (e *TestEnvironment) ChangeToTempDir() {
	if err := os.Chdir(e.TempDir); err != nil {
		e.t.Fatalf("Failed to change to temp dir: %v", err)
	}
}

// InitProjectStructure creates the standard Pluqqy directory structure
func (e *TestEnvironment) InitProjectStructure() error {
	dirs := []string{
		filepath.Join(e.TempDir, files.PluqqyDir, "pipelines"),
		filepath.Join(e.TempDir, files.PluqqyDir, "pipelines", "archive"),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypePrompt),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypeContext),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypeRules),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypePrompt, "archive"),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypeContext, "archive"),
		filepath.Join(e.TempDir, files.PluqqyDir, "components", models.ComponentTypeRules, "archive"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	return nil
}

// CreateComponentFile creates a component file with optional YAML front matter
func (e *TestEnvironment) CreateComponentFile(compType, name, content string, componentTags []string) string {
	e.t.Helper()
	
	componentPath := filepath.Join("components", compType, name+".md")
	fullPath := filepath.Join(e.TempDir, files.PluqqyDir, componentPath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create component directory: %v", err)
	}
	
	// Build content with YAML front matter if tags are provided
	var fullContent string
	if len(componentTags) > 0 {
		fullContent = "---\ntags: ["
		for i, tag := range componentTags {
			if i > 0 {
				fullContent += ", "
			}
			fullContent += tag
		}
		fullContent += "]\n---\n"
	}
	fullContent += content
	
	if err := os.WriteFile(fullPath, []byte(fullContent), 0644); err != nil {
		e.t.Fatalf("Failed to write component file: %v", err)
	}
	
	return componentPath
}

// CreateArchivedComponentFile creates an archived component file
func (e *TestEnvironment) CreateArchivedComponentFile(compType, name, content string) string {
	e.t.Helper()
	
	componentPath := filepath.Join("components", compType, "archive", name+".md")
	fullPath := filepath.Join(e.TempDir, files.PluqqyDir, componentPath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create component directory: %v", err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write component file: %v", err)
	}
	
	return componentPath
}

// CreatePipelineFile creates a pipeline file with the given configuration
func (e *TestEnvironment) CreatePipelineFile(name string, components []models.ComponentRef, pipelineTags []string) string {
	e.t.Helper()
	
	pipeline := &models.Pipeline{
		Name:       name,
		Path:       name + ".yaml",
		Tags:       pipelineTags,
		Components: components,
	}
	
	pipelinePath := filepath.Join(e.TempDir, files.PluqqyDir, "pipelines", pipeline.Path)
	
	// Ensure directory exists
	dir := filepath.Dir(pipelinePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		e.t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		e.t.Fatalf("Failed to write pipeline file: %v", err)
	}
	
	return pipeline.Path
}

// CreateArchivedPipelineFile creates an archived pipeline file
func (e *TestEnvironment) CreateArchivedPipelineFile(name string) string {
	e.t.Helper()
	
	pipeline := &models.Pipeline{
		Name:       name,
		Path:       name + ".yaml",
	}
	
	pipelinePath := filepath.Join(e.TempDir, files.PluqqyDir, "pipelines", "archive", pipeline.Path)
	
	// Ensure directory exists
	dir := filepath.Dir(pipelinePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		e.t.Fatalf("Failed to marshal pipeline: %v", err)
	}
	
	if err := os.WriteFile(pipelinePath, data, 0644); err != nil {
		e.t.Fatalf("Failed to write pipeline file: %v", err)
	}
	
	return filepath.Join("archive", pipeline.Path)
}

// CreateTagRegistry creates a tag registry file with the given tags
func (e *TestEnvironment) CreateTagRegistry(testTags []string) {
	e.t.Helper()
	
	registry := &models.TagRegistry{
		Tags: make([]models.Tag, len(testTags)),
	}
	
	colors := []string{"#3498db", "#e74c3c", "#2ecc71", "#f39c12", "#9b59b6"}
	for i, tag := range testTags {
		registry.Tags[i] = models.Tag{
			Name:  tag,
			Color: colors[i%len(colors)],
		}
	}
	
	registryPath := filepath.Join(e.TempDir, files.PluqqyDir, tags.TagsRegistryFile)
	
	// Ensure directory exists
	dir := filepath.Dir(registryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create registry directory: %v", err)
	}
	
	data, err := yaml.Marshal(registry)
	if err != nil {
		e.t.Fatalf("Failed to marshal registry: %v", err)
	}
	
	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		e.t.Fatalf("Failed to write registry file: %v", err)
	}
}

// CreateSettings creates a settings file with custom configuration
func (e *TestEnvironment) CreateSettings(settings *models.Settings) {
	e.t.Helper()
	
	settingsPath := filepath.Join(e.TempDir, files.PluqqyDir, "settings.yaml")
	
	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.t.Fatalf("Failed to create settings directory: %v", err)
	}
	
	data, err := yaml.Marshal(settings)
	if err != nil {
		e.t.Fatalf("Failed to marshal settings: %v", err)
	}
	
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		e.t.Fatalf("Failed to write settings file: %v", err)
	}
}

// GetProjectDir returns the Pluqqy project directory path
func (e *TestEnvironment) GetProjectDir() string {
	return filepath.Join(e.TempDir, files.PluqqyDir)
}

// SetupTestEnvironment is a convenience function that creates and initializes a test environment
// This function is for backward compatibility with existing tests
func SetupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()
	
	env := NewTestEnvironment(t)
	env.ChangeToTempDir()
	
	if err := env.InitProjectStructure(); err != nil {
		t.Fatalf("Failed to initialize project structure: %v", err)
	}
	
	return env.TempDir, env.Cleanup
}

// SetupArchiveTestEnvironment creates a test environment with archive directories
func SetupArchiveTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()
	
	tmpDir, cleanup := SetupTestEnvironment(t)
	
	// Archive directories are already created by InitProjectStructure
	// This function exists for backward compatibility
	
	return tmpDir, cleanup
}