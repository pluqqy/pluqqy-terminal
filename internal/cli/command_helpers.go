package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// CommandContext manages project validation and common command context
type CommandContext struct {
	ProjectPath string
	Settings    *models.Settings
	validated   bool
}

// NewCommandContext creates a new command context
func NewCommandContext() (*CommandContext, error) {
	projectPath := files.PluqqyDir
	return &CommandContext{
		ProjectPath: projectPath,
	}, nil
}

// ValidateProject ensures the project is initialized
func (c *CommandContext) ValidateProject() error {
	if c.validated {
		return nil
	}
	
	if _, err := os.Stat(c.ProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
	}
	
	c.validated = true
	return nil
}

// LoadSettingsWithDefault loads settings or returns default if error
func (c *CommandContext) LoadSettingsWithDefault() *models.Settings {
	if c.Settings != nil {
		return c.Settings
	}
	
	settings, err := files.ReadSettings()
	if err != nil {
		// Use default settings if can't read
		settings = models.DefaultSettings()
	}
	
	c.Settings = settings
	return settings
}

// EditorLauncher handles all editor-related operations
type EditorLauncher struct {
	DefaultEditor string
}

// NewEditorLauncher creates a new editor launcher
func NewEditorLauncher() *EditorLauncher {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return &EditorLauncher{
		DefaultEditor: editor,
	}
}

// OpenFile opens a file in the configured editor
func (e *EditorLauncher) OpenFile(filepath string) error {
	parts := strings.Fields(e.DefaultEditor)
	
	var editorCmd *exec.Cmd
	if len(parts) > 1 {
		editorCmd = exec.Command(parts[0], append(parts[1:], filepath)...)
	} else {
		editorCmd = exec.Command(e.DefaultEditor, filepath)
	}
	
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	return nil
}

// OpenTempFile creates a temp file with content and opens it
func (e *EditorLauncher) OpenTempFile(name, content string) (string, error) {
	tmpFile, err := os.CreateTemp("", name)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()
	
	if content != "" {
		if _, err := tmpFile.WriteString(content); err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to write to temp file: %w", err)
		}
	}
	
	if err := e.OpenFile(tmpFile.Name()); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	
	return tmpFile.Name(), nil
}

// ItemResolver manages finding and resolving items (components/pipelines)
type ItemResolver struct {
	ProjectPath string
	finder      *ComponentFinder
}

// NewItemResolver creates a new item resolver
func NewItemResolver(projectPath string) *ItemResolver {
	return &ItemResolver{
		ProjectPath: projectPath,
		finder:      NewComponentFinder(projectPath),
	}
}

// FindComponent finds a component by reference
func (r *ItemResolver) FindComponent(ref string) (string, error) {
	return r.finder.FindByReference(ref)
}

// FindPipeline finds a pipeline by name
func (r *ItemResolver) FindPipeline(name string) (string, error) {
	pipelinePath := filepath.Join(r.ProjectPath, "pipelines", name+".yaml")
	if _, err := os.Stat(pipelinePath); err == nil {
		return pipelinePath, nil
	}
	
	// Check without extension
	pipelinePath = filepath.Join(r.ProjectPath, "pipelines", name)
	if _, err := os.Stat(pipelinePath); err == nil {
		return pipelinePath, nil
	}
	
	return "", fmt.Errorf("pipeline '%s' not found", name)
}

// ResolveItem tries to resolve an item as either a pipeline or component
func (r *ItemResolver) ResolveItem(ref string) (itemType string, path string, err error) {
	// Try as pipeline first
	pipelinePath, err := r.FindPipeline(ref)
	if err == nil {
		return "pipeline", pipelinePath, nil
	}
	
	// Try as component
	componentPath, err := r.FindComponent(ref)
	if err == nil {
		return "component", componentPath, nil
	}
	
	// Check in archive
	archivedItems, err := r.finder.SearchInArchive(ref)
	if err == nil && len(archivedItems) > 0 {
		if len(archivedItems) == 1 {
			return "archived", archivedItems[0], nil
		}
		return "", "", fmt.Errorf("multiple archived items found matching '%s'", ref)
	}
	
	return "", "", fmt.Errorf("no pipeline or component found matching '%s'", ref)
}

// ConvertToRelativePath converts an absolute path to relative from project root
func (r *ItemResolver) ConvertToRelativePath(absolutePath string) string {
	return strings.TrimPrefix(absolutePath, r.ProjectPath+string(os.PathSeparator))
}

// ComponentFinder specialized component search logic
type ComponentFinder struct {
	ProjectPath string
}

// NewComponentFinder creates a new component finder
func NewComponentFinder(projectPath string) *ComponentFinder {
	return &ComponentFinder{
		ProjectPath: projectPath,
	}
}

// FindByReference finds a component by reference (name or path)
func (f *ComponentFinder) FindByReference(ref string) (string, error) {
	// If ref contains a slash, treat it as a path hint
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		componentType := parts[0]
		componentName := parts[1]
		
		// Normalize component type
		switch strings.ToLower(componentType) {
		case "context", "contexts":
			componentType = "contexts"
		case "prompt", "prompts":
			componentType = "prompts"
		case "rule", "rules":
			componentType = "rules"
		default:
			// Try as-is for other types
		}
		
		// Add .md extension if not present
		if !strings.HasSuffix(componentName, ".md") {
			componentName += ".md"
		}
		
		path := filepath.Join(f.ProjectPath, "components", componentType, componentName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("component not found: %s", ref)
	}
	
	// Search all component types
	matches, err := f.SearchAllTypes(ref)
	if err != nil {
		return "", err
	}
	
	if len(matches) == 0 {
		return "", fmt.Errorf("component '%s' not found", ref)
	}
	
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple components found with name '%s'. Please specify the type (e.g., contexts/%s)", ref, ref)
	}
	
	return matches[0], nil
}

// SearchAllTypes searches for a component across all types
func (f *ComponentFinder) SearchAllTypes(name string) ([]string, error) {
	var matches []string
	componentTypes := []string{"contexts", "prompts", "rules"}
	
	for _, ct := range componentTypes {
		// Try with .md extension
		path := filepath.Join(f.ProjectPath, "components", ct, name+".md")
		if _, err := os.Stat(path); err == nil {
			matches = append(matches, path)
		}
		
		// Try without extension (in case user included it)
		if strings.HasSuffix(name, ".md") {
			path = filepath.Join(f.ProjectPath, "components", ct, name)
			if _, err := os.Stat(path); err == nil {
				// Avoid duplicates
				alreadyAdded := false
				for _, m := range matches {
					if m == path {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					matches = append(matches, path)
				}
			}
		}
	}
	
	return matches, nil
}

// SearchInArchive searches for items in the archive directory
func (f *ComponentFinder) SearchInArchive(name string) ([]string, error) {
	var matches []string
	componentTypes := []string{"contexts", "prompts", "rules"}
	
	// Check archived components
	for _, ct := range componentTypes {
		archivePath := filepath.Join(f.ProjectPath, "archive", "components", ct, name+".md")
		if _, err := os.Stat(archivePath); err == nil {
			matches = append(matches, archivePath)
		}
		
		// Also check without extension if user included it
		if strings.HasSuffix(name, ".md") {
			archivePath = filepath.Join(f.ProjectPath, "archive", "components", ct, name)
			if _, err := os.Stat(archivePath); err == nil {
				matches = append(matches, archivePath)
			}
		}
	}
	
	// Check archived pipelines
	pipelineArchivePath := filepath.Join(f.ProjectPath, "archive", "pipelines", name+".yaml")
	if _, err := os.Stat(pipelineArchivePath); err == nil {
		matches = append(matches, pipelineArchivePath)
	}
	
	return matches, nil
}