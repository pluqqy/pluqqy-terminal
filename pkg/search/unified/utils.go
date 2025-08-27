package unified

import (
	"strings"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ShouldIncludeArchived checks if the search query includes archived items
func ShouldIncludeArchived(searchQuery string) bool {
	if searchQuery == "" {
		return false
	}

	// Check for simple status:archived pattern
	lowerQuery := strings.ToLower(searchQuery)
	if strings.Contains(lowerQuery, "status:archived") {
		return true
	}

	// Parse the search query to check for status:archived using unified parser
	parsedQuery := ParseQuery(searchQuery)
	for _, filter := range parsedQuery.Filters {
		if filter.Type == "status" && strings.ToLower(filter.Value) == "archived" {
			return true
		}
	}

	return false
}

// FilterSearchResultsByType separates component results by type
func FilterSearchResultsByType(results []SearchResult[*ComponentItemWrapper]) ([]ComponentItem, []ComponentItem, []ComponentItem) {
	var prompts, contexts, rules []ComponentItem

	for _, result := range results {
		item := ComponentItem{
			Name:         result.Item.name,
			Path:         result.Item.path,
			CompType:     result.Item.compType,
			LastModified: result.Item.lastModified,
			UsageCount:   result.Item.usageCount,
			TokenCount:   result.Item.tokenCount,
			Tags:         result.Item.tags,
			IsArchived:   result.Item.isArchived,
		}

		switch result.Item.compType {
		case models.ComponentTypePrompt:
			prompts = append(prompts, item)
		case models.ComponentTypeContext:
			contexts = append(contexts, item)
		case models.ComponentTypeRules:
			rules = append(rules, item)
		}
	}

	return prompts, contexts, rules
}

// ConvertPipelineResults converts pipeline search results back to PipelineItem
func ConvertPipelineResults(results []SearchResult[*PipelineItemWrapper]) []PipelineItem {
	var pipelines []PipelineItem

	for _, result := range results {
		pipelines = append(pipelines, PipelineItem{
			Name:       result.Item.name,
			Path:       result.Item.path,
			Tags:       result.Item.tags,
			TokenCount: result.Item.tokenCount,
			IsArchived: result.Item.isArchived,
			Modified:   result.Item.modified,
		})
	}

	return pipelines
}

// CombineComponentsByType combines component slices into a single slice
func CombineComponentsByType(prompts, contexts, rules []ComponentItem) []ComponentItem {
	var components []ComponentItem
	components = append(components, prompts...)
	components = append(components, contexts...)
	components = append(components, rules...)
	return components
}