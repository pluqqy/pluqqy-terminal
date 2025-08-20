package files

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// UpdateTagRegistryOnArchive removes tags from registry if they're no longer used by active items
func UpdateTagRegistryOnArchive(archivedTags []string) error {
	if len(archivedTags) == 0 {
		return nil
	}

	// Load the tag registry
	registryPath := filepath.Join(PluqqyDir, "tags.yaml")
	registry, err := loadTagRegistry(registryPath)
	if err != nil {
		// If registry doesn't exist, nothing to remove
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load tag registry: %w", err)
	}

	// Get all tags currently in use by active items
	activeTags, err := getActiveItemTags()
	if err != nil {
		return fmt.Errorf("failed to get active tags: %w", err)
	}

	// Remove tags that are no longer in use
	registry.Tags = filterTagsInUse(registry.Tags, activeTags)

	// Save the updated registry
	return saveTagRegistry(registryPath, registry)
}

// UpdateTagRegistryOnUnarchive adds tags back to registry from unarchived items
func UpdateTagRegistryOnUnarchive(unarchivedTags []string) error {
	if len(unarchivedTags) == 0 {
		return nil
	}

	// Load the tag registry
	registryPath := filepath.Join(PluqqyDir, "tags.yaml")
	registry, err := loadTagRegistry(registryPath)
	if err != nil {
		// If registry doesn't exist, create a new one
		if os.IsNotExist(err) {
			registry = &models.TagRegistry{
				Tags: []models.Tag{},
			}
		} else {
			return fmt.Errorf("failed to load tag registry: %w", err)
		}
	}

	// Add tags that don't already exist
	for _, tagName := range unarchivedTags {
		normalizedName := models.NormalizeTagName(tagName)
		
		// Check if tag already exists
		exists := false
		for _, existingTag := range registry.Tags {
			if models.NormalizeTagName(existingTag.Name) == normalizedName {
				exists = true
				break
			}
		}

		// If tag doesn't exist, add it with auto-assigned color
		if !exists {
			newTag := models.Tag{
				Name:  normalizedName,
				Color: models.GetTagColor(normalizedName, ""),
			}
			registry.Tags = append(registry.Tags, newTag)
		}
	}

	// Save the updated registry
	return saveTagRegistry(registryPath, registry)
}

// getActiveItemTags returns all unique tags from active (non-archived) items
func getActiveItemTags() (map[string]bool, error) {
	activeTags := make(map[string]bool)

	// Get tags from active components
	for _, compType := range []string{models.ComponentTypePrompt, models.ComponentTypeContext, models.ComponentTypeRules} {
		components, err := ListComponents(compType)
		if err != nil {
			continue // Skip on error
		}

		for _, compFile := range components {
			compPath := filepath.Join(ComponentsDir, compType, compFile)
			comp, err := ReadComponent(compPath)
			if err != nil {
				continue // Skip components that can't be read
			}

			for _, tag := range comp.Tags {
				normalized := models.NormalizeTagName(tag)
				activeTags[normalized] = true
			}
		}
	}

	// Get tags from active pipelines
	pipelines, err := ListPipelines()
	if err == nil {
		for _, pipelineFile := range pipelines {
			pipeline, err := ReadPipeline(pipelineFile)
			if err != nil {
				continue // Skip pipelines that can't be read
			}

			for _, tag := range pipeline.Tags {
				normalized := models.NormalizeTagName(tag)
				activeTags[normalized] = true
			}
		}
	}

	return activeTags, nil
}

// filterTagsInUse returns only tags that are in the active tags set
func filterTagsInUse(allTags []models.Tag, activeTags map[string]bool) []models.Tag {
	var filteredTags []models.Tag
	
	for _, tag := range allTags {
		normalizedName := models.NormalizeTagName(tag.Name)
		if activeTags[normalizedName] {
			filteredTags = append(filteredTags, tag)
		}
	}
	
	return filteredTags
}

// loadTagRegistry loads the tag registry from disk
func loadTagRegistry(path string) (*models.TagRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var registry models.TagRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse tag registry: %w", err)
	}

	return &registry, nil
}

// saveTagRegistry saves the tag registry to disk
func saveTagRegistry(path string, registry *models.TagRegistry) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal tag registry: %w", err)
	}

	// Write atomically
	return writeFileAtomic(path, data, 0644)
}