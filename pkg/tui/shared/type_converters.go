package shared

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ConvertTUIComponentItemToShared converts a TUI componentItem to shared ComponentItem
func ConvertTUIComponentItemToShared(name, path, compType string, lastModified time.Time, usageCount, tokenCount int, tags []string, isArchived bool) ComponentItem {
	return ComponentItem{
		Name:         name,
		Path:         path,
		CompType:     compType,
		LastModified: lastModified,
		UsageCount:   usageCount,
		TokenCount:   tokenCount,
		Tags:         tags,
		IsArchived:   isArchived,
	}
}

// ConvertTUIPipelineItemToShared converts a TUI pipelineItem to shared PipelineItem
func ConvertTUIPipelineItemToShared(name, path string, tags []string, tokenCount int, isArchived bool) PipelineItem {
	// Get modified time from file system
	var modTime time.Time
	if isArchived {
		archivedPath := filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, path)
		if stat, err := os.Stat(archivedPath); err == nil {
			modTime = stat.ModTime()
		}
	} else {
		pipelinePath := filepath.Join(files.PluqqyDir, files.PipelinesDir, path)
		if stat, err := os.Stat(pipelinePath); err == nil {
			modTime = stat.ModTime()
		}
	}
	
	return PipelineItem{
		Name:       name,
		Path:       path,
		Tags:       tags,
		TokenCount: tokenCount,
		IsArchived: isArchived,
		Modified:   modTime,
	}
}

// ConvertSharedComponentItemsToTUI converts shared ComponentItems back to TUI format for compatibility
// This function returns the fields needed to reconstruct TUI componentItem structs
func ConvertSharedComponentItemsToTUI(items []ComponentItem) (names, paths, compTypes []string, lastModifieds []time.Time, usageCounts, tokenCounts []int, tags [][]string, isArchiveds []bool) {
	count := len(items)
	names = make([]string, count)
	paths = make([]string, count)
	compTypes = make([]string, count)
	lastModifieds = make([]time.Time, count)
	usageCounts = make([]int, count)
	tokenCounts = make([]int, count)
	tags = make([][]string, count)
	isArchiveds = make([]bool, count)
	
	for i, item := range items {
		names[i] = item.Name
		paths[i] = item.Path
		compTypes[i] = item.CompType
		lastModifieds[i] = item.LastModified
		usageCounts[i] = item.UsageCount
		tokenCounts[i] = item.TokenCount
		tags[i] = item.Tags
		isArchiveds[i] = item.IsArchived
	}
	
	return
}

// ConvertSharedPipelineItemsToTUI converts shared PipelineItems back to TUI format for compatibility
// This function returns the fields needed to reconstruct TUI pipelineItem structs
func ConvertSharedPipelineItemsToTUI(items []PipelineItem) (names, paths []string, tags [][]string, tokenCounts []int, isArchiveds []bool) {
	count := len(items)
	names = make([]string, count)
	paths = make([]string, count)
	tags = make([][]string, count)
	tokenCounts = make([]int, count)
	isArchiveds = make([]bool, count)
	
	for i, item := range items {
		names[i] = item.Name
		paths[i] = item.Path
		tags[i] = item.Tags
		tokenCounts[i] = item.TokenCount
		isArchiveds[i] = item.IsArchived
	}
	
	return
}

// FilterSearchResultsByType filters search results to separate components by type
func FilterSearchResultsByType(componentResults []SearchResult[*ComponentItemWrapper]) (prompts, contexts, rules []ComponentItem) {
	for _, result := range componentResults {
		item := result.Item
		sharedItem := ComponentItem{
			Name:         item.GetName(),
			Path:         item.GetPath(),
			CompType:     item.GetSubType(),
			LastModified: item.GetModified(),
			UsageCount:   item.GetUsageCount(),
			TokenCount:   item.GetTokenCount(),
			Tags:         item.GetTags(),
			IsArchived:   item.IsArchived(),
		}
		
		switch item.GetSubType() {
		case models.ComponentTypePrompt:
			prompts = append(prompts, sharedItem)
		case models.ComponentTypeContext:
			contexts = append(contexts, sharedItem)
		case models.ComponentTypeRules:
			rules = append(rules, sharedItem)
		}
	}
	
	return
}

// ConvertPipelineResults converts pipeline search results to shared PipelineItem format
func ConvertPipelineResults(pipelineResults []SearchResult[*PipelineItemWrapper]) []PipelineItem {
	var pipelines []PipelineItem
	
	for _, result := range pipelineResults {
		item := result.Item
		sharedItem := PipelineItem{
			Name:       item.GetName(),
			Path:       item.GetPath(),
			Tags:       item.GetTags(),
			TokenCount: item.GetTokenCount(),
			IsArchived: item.IsArchived(),
			Modified:   item.GetModified(),
		}
		pipelines = append(pipelines, sharedItem)
	}
	
	return pipelines
}

// CreateComponentItemsFromTUITypes creates ComponentItems from the individual TUI fields
func CreateComponentItemsFromTUITypes(names, paths, compTypes []string, lastModifieds []time.Time, usageCounts, tokenCounts []int, tags [][]string, isArchiveds []bool) []ComponentItem {
	if len(names) == 0 {
		return []ComponentItem{}
	}
	
	items := make([]ComponentItem, len(names))
	for i := range names {
		items[i] = ComponentItem{
			Name:         names[i],
			Path:         paths[i],
			CompType:     compTypes[i],
			LastModified: lastModifieds[i],
			UsageCount:   usageCounts[i],
			TokenCount:   tokenCounts[i],
			Tags:         tags[i],
			IsArchived:   isArchiveds[i],
		}
	}
	
	return items
}

// CreatePipelineItemsFromTUITypes creates PipelineItems from the individual TUI fields
func CreatePipelineItemsFromTUITypes(names, paths []string, tags [][]string, tokenCounts []int, isArchiveds []bool, modifieds []time.Time) []PipelineItem {
	if len(names) == 0 {
		return []PipelineItem{}
	}
	
	items := make([]PipelineItem, len(names))
	for i := range names {
		modified := time.Time{}
		if i < len(modifieds) {
			modified = modifieds[i]
		}
		
		items[i] = PipelineItem{
			Name:       names[i],
			Path:       paths[i],
			Tags:       tags[i],
			TokenCount: tokenCounts[i],
			IsArchived: isArchiveds[i],
			Modified:   modified,
		}
	}
	
	return items
}

// SeparateComponentsByType separates a list of ComponentItems by their type
func SeparateComponentsByType(components []ComponentItem) (prompts, contexts, rules []ComponentItem) {
	for _, comp := range components {
		switch comp.CompType {
		case models.ComponentTypePrompt:
			prompts = append(prompts, comp)
		case models.ComponentTypeContext:
			contexts = append(contexts, comp)
		case models.ComponentTypeRules:
			rules = append(rules, comp)
		}
	}
	return
}

// CombineComponentsByType combines separated component types into a single slice
func CombineComponentsByType(prompts, contexts, rules []ComponentItem) []ComponentItem {
	var combined []ComponentItem
	combined = append(combined, prompts...)
	combined = append(combined, contexts...)
	combined = append(combined, rules...)
	return combined
}

// ExtractTUIComponentFields extracts fields from ComponentItems for use in TUI structs
func ExtractTUIComponentFields(items []ComponentItem) (names, paths, compTypes []string, lastModifieds []time.Time, usageCounts, tokenCounts []int, tags [][]string, isArchiveds []bool) {
	return ConvertSharedComponentItemsToTUI(items)
}