package tags

import (
	"fmt"
	
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// UsageStats represents usage statistics for a tag
type UsageStats struct {
	ComponentCount int
	PipelineCount  int
	TotalCount     int
}

// CountTagUsage counts how many times a tag is used across all components and pipelines
func CountTagUsage(tagName string) (*UsageStats, error) {
	stats := &UsageStats{}
	normalized := models.NormalizeTagName(tagName)
	
	// Count in components
	for _, compType := range []string{"prompts", "contexts", "rules"} {
		components, err := files.ListComponents(compType)
		if err != nil {
			continue
		}
		
		for _, compFile := range components {
			path := "components/" + compType + "/" + compFile
			comp, err := files.ReadComponent(path)
			if err != nil {
				continue
			}
			
			for _, tag := range comp.Tags {
				if models.NormalizeTagName(tag) == normalized {
					stats.ComponentCount++
					stats.TotalCount++
					break
				}
			}
		}
	}
	
	// Count in pipelines
	pipelines, err := files.ListPipelines()
	if err == nil {
		for _, pipelineFile := range pipelines {
			pipeline, err := files.ReadPipeline(pipelineFile)
			if err != nil {
				continue
			}
			
			for _, tag := range pipeline.Tags {
				if models.NormalizeTagName(tag) == normalized {
					stats.PipelineCount++
					stats.TotalCount++
					break
				}
			}
		}
	}
	
	return stats, nil
}

// GetAllTagUsage returns usage statistics for all tags in the registry
func GetAllTagUsage() (map[string]*UsageStats, error) {
	usage := make(map[string]*UsageStats)
	
	// Get all tags from registry
	registry, err := NewRegistry()
	if err != nil {
		return usage, err
	}
	
	// Initialize usage stats for all registry tags
	for _, tag := range registry.ListTags() {
		usage[tag.Name] = &UsageStats{}
	}
	
	// Count component usage
	for _, compType := range []string{"prompts", "contexts", "rules"} {
		components, err := files.ListComponents(compType)
		if err != nil {
			continue
		}
		
		for _, compFile := range components {
			path := "components/" + compType + "/" + compFile
			comp, err := files.ReadComponent(path)
			if err != nil {
				continue
			}
			
			for _, tag := range comp.Tags {
				normalized := models.NormalizeTagName(tag)
				if stats, exists := usage[normalized]; exists {
					stats.ComponentCount++
					stats.TotalCount++
				} else {
					// Tag not in registry but still in use
					usage[normalized] = &UsageStats{
						ComponentCount: 1,
						TotalCount:     1,
					}
				}
			}
		}
	}
	
	// Count pipeline usage
	pipelines, err := files.ListPipelines()
	if err == nil {
		for _, pipelineFile := range pipelines {
			pipeline, err := files.ReadPipeline(pipelineFile)
			if err != nil {
				continue
			}
			
			for _, tag := range pipeline.Tags {
				normalized := models.NormalizeTagName(tag)
				if stats, exists := usage[normalized]; exists {
					stats.PipelineCount++
					stats.TotalCount++
				} else {
					// Tag not in registry but still in use
					usage[normalized] = &UsageStats{
						PipelineCount: 1,
						TotalCount:    1,
					}
				}
			}
		}
	}
	
	return usage, nil
}

// CleanupOrphanedTags removes tags from the registry that are no longer used
// Returns the list of tags that were removed
func CleanupOrphanedTags(tagsToCheck []string) ([]string, error) {
	registry, err := NewRegistry()
	if err != nil {
		return nil, err
	}
	
	var removedTags []string
	
	for _, tagName := range tagsToCheck {
		normalized := models.NormalizeTagName(tagName)
		
		// Check if tag is still in use
		stats, err := CountTagUsage(normalized)
		if err != nil {
			// If we can't check usage, skip this tag to be safe
			continue
		}
		
		// If tag is not used anywhere, remove it from registry
		if stats.TotalCount == 0 {
			if err := registry.RemoveTag(normalized); err == nil {
				removedTags = append(removedTags, normalized)
			}
		}
	}
	
	// Save the updated registry if any tags were removed
	if len(removedTags) > 0 {
		if err := registry.Save(); err != nil {
			return removedTags, fmt.Errorf("removed tags but failed to save registry: %w", err)
		}
	}
	
	return removedTags, nil
}