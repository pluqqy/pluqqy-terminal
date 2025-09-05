package tags

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

const (
	TagsRegistryFile = "tags.yaml"
)

// Registry manages the tag registry for a project
type Registry struct {
	mu       sync.RWMutex
	registry *models.TagRegistry
	path     string
}

// NewRegistry creates a new tag registry manager
func NewRegistry() (*Registry, error) {
	registryPath := filepath.Join(files.PluqqyDir, TagsRegistryFile)
	
	r := &Registry{
		path: registryPath,
	}
	
	// Load existing registry or create new one
	if err := r.Load(); err != nil {
		// If file doesn't exist, create empty registry
		if os.IsNotExist(err) {
			r.registry = &models.TagRegistry{
				Tags: []models.Tag{},
			}
			return r, nil
		}
		return nil, fmt.Errorf("failed to load tag registry: %w", err)
	}
	
	return r, nil
}

// Load reads the tag registry from disk
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	data, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}
	
	var registry models.TagRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to parse tag registry: %w", err)
	}
	
	r.registry = &registry
	return nil
}

// Save writes the tag registry to disk
func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	data, err := yaml.Marshal(r.registry)
	if err != nil {
		return fmt.Errorf("failed to marshal tag registry: %w", err)
	}
	
	// Write atomically
	tmpFile := r.path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tag registry: %w", err)
	}
	
	if err := os.Rename(tmpFile, r.path); err != nil {
		os.Remove(tmpFile) // Clean up temp file
		return fmt.Errorf("failed to save tag registry: %w", err)
	}
	
	return nil
}

// GetTag retrieves tag metadata by name
func (r *Registry) GetTag(name string) (*models.Tag, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	normalizedName := models.NormalizeTagName(name)
	
	for _, tag := range r.registry.Tags {
		if models.NormalizeTagName(tag.Name) == normalizedName {
			return &tag, true
		}
	}
	
	return nil, false
}

// AddTag adds or updates a tag in the registry
func (r *Registry) AddTag(tag models.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Validate tag name
	if err := models.ValidateTagName(tag.Name); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}
	
	// Normalize tag name
	tag.Name = models.NormalizeTagName(tag.Name)
	
	// Check if tag already exists
	for i, existing := range r.registry.Tags {
		if models.NormalizeTagName(existing.Name) == tag.Name {
			// Update existing tag
			r.registry.Tags[i] = tag
			return nil
		}
	}
	
	// Add new tag
	r.registry.Tags = append(r.registry.Tags, tag)
	return nil
}

// RemoveTag removes a tag from the registry
func (r *Registry) RemoveTag(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	normalizedName := models.NormalizeTagName(name)
	
	newTags := make([]models.Tag, 0, len(r.registry.Tags))
	found := false
	
	for _, tag := range r.registry.Tags {
		if models.NormalizeTagName(tag.Name) != normalizedName {
			newTags = append(newTags, tag)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("tag '%s' not found in registry", name)
	}
	
	r.registry.Tags = newTags
	return nil
}

// ListTags returns all tags in the registry
func (r *Registry) ListTags() []models.Tag {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Return a copy to prevent external modification
	tags := make([]models.Tag, len(r.registry.Tags))
	copy(tags, r.registry.Tags)
	return tags
}

// GetOrCreateTag gets an existing tag or creates a new one with default color
func (r *Registry) GetOrCreateTag(name string) (*models.Tag, error) {
	// Check if tag exists
	if tag, exists := r.GetTag(name); exists {
		return tag, nil
	}
	
	// Create new tag with auto-assigned color
	normalizedName := models.NormalizeTagName(name)
	newTag := models.Tag{
		Name:  normalizedName,
		Color: models.GetTagColor(normalizedName, ""),
	}
	
	// Add to registry
	if err := r.AddTag(newTag); err != nil {
		return nil, err
	}
	
	// Save registry
	if err := r.Save(); err != nil {
		return nil, fmt.Errorf("failed to save tag registry: %w", err)
	}
	
	return &newTag, nil
}

// RenameTag renames a tag throughout the system
func (r *Registry) RenameTag(oldName, newName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Validate new name
	if err := models.ValidateTagName(newName); err != nil {
		return fmt.Errorf("invalid new tag name: %w", err)
	}
	
	oldNormalized := models.NormalizeTagName(oldName)
	newNormalized := models.NormalizeTagName(newName)
	
	// Check if old tag exists
	found := false
	for i, tag := range r.registry.Tags {
		if models.NormalizeTagName(tag.Name) == oldNormalized {
			r.registry.Tags[i].Name = newNormalized
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("tag '%s' not found", oldName)
	}
	
	// TODO: Update all components and pipelines that use this tag
	// This would require scanning all files and updating their tags
	
	return nil
}

// GetTagStats returns usage statistics for all tags
func (r *Registry) GetTagStats() (map[string]int, error) {
	stats := make(map[string]int)
	
	// Count component usage
	for _, compType := range []string{models.ComponentTypePrompt, models.ComponentTypeContext, models.ComponentTypeRules} {
		components, err := files.ListComponents(compType)
		if err != nil {
			return nil, fmt.Errorf("failed to list %s components: %w", compType, err)
		}
		
		for _, compFile := range components {
			compPath := filepath.Join(files.ComponentsDir, compType, compFile)
			comp, err := files.ReadComponent(compPath)
			if err != nil {
				continue // Skip components that can't be read
			}
			
			for _, tag := range comp.Tags {
				normalized := models.NormalizeTagName(tag)
				stats[normalized]++
			}
		}
	}
	
	// Count pipeline usage
	pipelines, err := files.ListPipelines()
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	
	for _, pipelineFile := range pipelines {
		pipeline, err := files.ReadPipeline(pipelineFile)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		for _, tag := range pipeline.Tags {
			normalized := models.NormalizeTagName(tag)
			stats[normalized]++
		}
	}
	
	return stats, nil
}