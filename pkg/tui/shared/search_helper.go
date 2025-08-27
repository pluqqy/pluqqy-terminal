package shared

import (
	"sort"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// SearchHelper provides compatibility functions for the existing TUI search interface
type SearchHelper struct {
	unifiedManager *UnifiedSearchManager
}

// NewSearchHelper creates a new search helper
func NewSearchHelper() *SearchHelper {
	return &SearchHelper{
		unifiedManager: NewUnifiedSearchManager(),
	}
}

// FilterSearchResults filters pipelines and components based on search results
// This maintains the same signature as the existing function for backward compatibility
func (sh *SearchHelper) FilterSearchResults(results []search.SearchResult, pipelines []PipelineItem, components []ComponentItem) ([]PipelineItem, []ComponentItem) {
	if results == nil || len(results) == 0 {
		// No search results, return empty lists
		return []PipelineItem{}, []ComponentItem{}
	}

	// Build filtered lists from search results
	filteredPipelines := []PipelineItem{}
	filteredComponents := []ComponentItem{}

	// Use maps to track what's already added
	addedPipelines := make(map[string]bool)
	addedComponents := make(map[string]bool)

	for _, result := range results {
		if result.Item.Type == search.ItemTypePipeline {
			// Match by path for accuracy (handles archived items correctly)
			if !addedPipelines[result.Item.Path] {
				for _, p := range pipelines {
					if p.Path == result.Item.Path {
						filteredPipelines = append(filteredPipelines, p)
						addedPipelines[result.Item.Path] = true
						break
					}
				}
			}
		} else {
			// Match components by path
			if !addedComponents[result.Item.Path] {
				for _, c := range components {
					if c.Path == result.Item.Path {
						filteredComponents = append(filteredComponents, c)
						addedComponents[result.Item.Path] = true
						break
					}
				}
			}
		}
	}

	// Sort filtered components by type to ensure proper grouping
	sort.Slice(filteredComponents, func(i, j int) bool {
		// Define type order
		typeOrder := map[string]int{
			models.ComponentTypeContext: 1,
			models.ComponentTypePrompt:  2,
			models.ComponentTypeRules:   3,
		}

		// Get order values, defaulting to 4 for unknown types
		orderI, okI := typeOrder[filteredComponents[i].CompType]
		if !okI {
			orderI = 4
		}
		orderJ, okJ := typeOrder[filteredComponents[j].CompType]
		if !okJ {
			orderJ = 4
		}

		// Sort by type order first
		if orderI != orderJ {
			return orderI < orderJ
		}

		// Within same type, sort by name
		return filteredComponents[i].Name < filteredComponents[j].Name
	})

	return filteredPipelines, filteredComponents
}

// UnifiedFilterComponents uses the new unified search to filter components
func (sh *SearchHelper) UnifiedFilterComponents(query string, prompts, contexts, rules []ComponentItem) ([]ComponentItem, []ComponentItem, []ComponentItem, error) {
	// Load components into the unified manager
	sh.unifiedManager.LoadComponentItems(prompts, contexts, rules)
	
	// Configure search options
	sh.unifiedManager.SetIncludeArchived(ShouldIncludeArchived(query))
	
	// Perform the search
	if query == "" {
		// No search, return all items
		return prompts, contexts, rules, nil
	}
	
	results, err := sh.unifiedManager.SearchComponents(query, nil)
	if err != nil {
		// On error, return all items
		return prompts, contexts, rules, err
	}
	
	// Convert results back to component lists
	filteredPrompts, filteredContexts, filteredRules := FilterSearchResultsByType(results)
	return filteredPrompts, filteredContexts, filteredRules, nil
}

// UnifiedFilterPipelines uses the new unified search to filter pipelines
func (sh *SearchHelper) UnifiedFilterPipelines(query string, pipelines []PipelineItem) ([]PipelineItem, error) {
	// Load pipelines into the unified manager
	sh.unifiedManager.LoadPipelineItems(pipelines)
	
	// Configure search options
	sh.unifiedManager.SetIncludeArchived(ShouldIncludeArchived(query))
	
	// Perform the search
	if query == "" {
		// No search, return all items
		return pipelines, nil
	}
	
	results, err := sh.unifiedManager.SearchPipelines(query)
	if err != nil {
		// On error, return all items
		return pipelines, err
	}
	
	// Convert results back to pipeline list
	return ConvertPipelineResults(results), nil
}

// UnifiedFilterAll performs a unified search across both components and pipelines
func (sh *SearchHelper) UnifiedFilterAll(query string, prompts, contexts, rules []ComponentItem, pipelines []PipelineItem) ([]ComponentItem, []ComponentItem, []ComponentItem, []PipelineItem, error) {
	// Load data into the unified manager
	sh.unifiedManager.LoadComponentItems(prompts, contexts, rules)
	sh.unifiedManager.LoadPipelineItems(pipelines)
	
	// Configure search options
	sh.unifiedManager.SetIncludeArchived(ShouldIncludeArchived(query))
	
	if query == "" {
		// No search, return all items
		return prompts, contexts, rules, pipelines, nil
	}
	
	// Perform unified search
	componentResults, pipelineResults, err := sh.unifiedManager.SearchAll(query)
	if err != nil {
		// On error, return all items
		return prompts, contexts, rules, pipelines, err
	}
	
	// Convert results back to original format
	filteredPrompts, filteredContexts, filteredRules := FilterSearchResultsByType(componentResults)
	filteredPipelines := ConvertPipelineResults(pipelineResults)
	
	return filteredPrompts, filteredContexts, filteredRules, filteredPipelines, nil
}

// GetUnifiedManager returns the underlying unified manager for advanced usage
func (sh *SearchHelper) GetUnifiedManager() *UnifiedSearchManager {
	return sh.unifiedManager
}

// SetSearchOptions configures search behavior
func (sh *SearchHelper) SetSearchOptions(includeArchived bool, maxResults int, sortBy string) {
	sh.unifiedManager.SetIncludeArchived(includeArchived)
	sh.unifiedManager.SetMaxResults(maxResults)
	
	// Update sort options on both engines
	componentOptions := sh.unifiedManager.componentEngine.options
	componentOptions.SortBy = sortBy
	sh.unifiedManager.componentEngine.SetOptions(componentOptions)
	
	pipelineOptions := sh.unifiedManager.pipelineEngine.options
	pipelineOptions.SortBy = sortBy
	sh.unifiedManager.pipelineEngine.SetOptions(pipelineOptions)
}

// IsStructuredQuery checks if the query uses structured search syntax
func (sh *SearchHelper) IsStructuredQuery(query string) bool {
	return sh.unifiedManager.IsStructuredQuery(query)
}

// ParseSearchQuery parses a search query to understand its structure
func (sh *SearchHelper) ParseSearchQuery(query string) (*search.Query, error) {
	return sh.unifiedManager.ParseSearchQuery(query)
}

// GetSearchSuggestions provides search suggestions based on current data
func (sh *SearchHelper) GetSearchSuggestions(partial string) []string {
	return sh.unifiedManager.GetSearchSuggestions(partial)
}