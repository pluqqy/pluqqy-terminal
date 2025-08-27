package unified

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

// FilterSearchResults is deprecated - legacy search results are no longer supported
// Use UnifiedFilterAll or other unified search methods instead
func (sh *SearchHelper) FilterSearchResults(results interface{}, pipelines []PipelineItem, components []ComponentItem) ([]PipelineItem, []ComponentItem) {
	// Return empty results as this method is deprecated
	return []PipelineItem{}, []ComponentItem{}
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
func (sh *SearchHelper) ParseSearchQuery(query string) (*ParsedQuery, error) {
	return sh.unifiedManager.ParseSearchQuery(query)
}

// GetSearchSuggestions provides search suggestions based on current data
func (sh *SearchHelper) GetSearchSuggestions(partial string) []string {
	return sh.unifiedManager.GetSearchSuggestions(partial)
}