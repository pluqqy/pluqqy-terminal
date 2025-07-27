package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/pluqqy/pkg/models"
	"gopkg.in/yaml.v3"
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
)

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
	absPath := filepath.Join(PluqqyDir, path)
	
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat component %s: %w", path, err)
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
	absPath := filepath.Join(PluqqyDir, path)
	
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for component: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write component %s: %w", path, err)
	}

	return nil
}

func ReadPipeline(path string) (*models.Pipeline, error) {
	absPath := filepath.Join(PluqqyDir, PipelinesDir, path)
	
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pipeline %s: %w", path, err)
	}

	var pipeline models.Pipeline
	if err := yaml.Unmarshal(content, &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline YAML %s: %w", path, err)
	}

	pipeline.Path = path
	
	return &pipeline, nil
}

func WritePipeline(pipeline *models.Pipeline) error {
	if pipeline.Path == "" {
		pipeline.Path = fmt.Sprintf("%s.yaml", pipeline.Name)
	}

	absPath := filepath.Join(PluqqyDir, PipelinesDir, pipeline.Path)
	
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for pipeline: %w", err)
	}

	content, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline to YAML: %w", err)
	}

	if err := os.WriteFile(absPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write pipeline %s: %w", pipeline.Path, err)
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
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
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
	case "prompt", "prompts":
		subDir = PromptsDir
	case "context", "contexts":
		subDir = ContextsDir
	case "rules":
		subDir = RulesDir
	default:
		return nil, fmt.Errorf("invalid component type: %s", componentType)
	}

	componentsPath := filepath.Join(PluqqyDir, ComponentsDir, subDir)
	
	entries, err := os.ReadDir(componentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list components: %w", err)
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
		return "prompt"
	} else if strings.Contains(path, ContextsDir) {
		return "context"
	} else if strings.Contains(path, RulesDir) {
		return "rules"
	}
	return "unknown"
}

// WriteFile writes content to a file (for PLUQQY.md output)
func WriteFile(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}