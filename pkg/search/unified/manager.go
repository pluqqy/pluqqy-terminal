package unified

import (
	"strings"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// UnifiedSearchManager manages search operations across different TUI views
type UnifiedSearchManager struct {
	// Component search engine
	componentEngine *SearchEngine[*ComponentItemWrapper]
	
	// Pipeline search engine  
	pipelineEngine *SearchEngine[*PipelineItemWrapper]
	
	// Combined search engine for mixed searches
	combinedItems []Searchable
	
	// Search configuration
	includeArchived bool
	maxResults      int
}

// NewUnifiedSearchManager creates a new unified search manager
func NewUnifiedSearchManager() *UnifiedSearchManager {
	return &UnifiedSearchManager{
		componentEngine: NewSearchEngine[*ComponentItemWrapper](),
		pipelineEngine:  NewSearchEngine[*PipelineItemWrapper](),
		maxResults:      100,
	}
}

// SetIncludeArchived sets whether to include archived items in searches
func (usm *UnifiedSearchManager) SetIncludeArchived(include bool) {
	usm.includeArchived = include
	
	// Update component engine options
	componentOptions := usm.componentEngine.options
	componentOptions.IncludeArchived = include
	usm.componentEngine.SetOptions(componentOptions)
	
	// Update pipeline engine options
	pipelineOptions := usm.pipelineEngine.options
	pipelineOptions.IncludeArchived = include
	usm.pipelineEngine.SetOptions(pipelineOptions)
}

// SetMaxResults sets the maximum number of results to return
func (usm *UnifiedSearchManager) SetMaxResults(max int) {
	usm.maxResults = max
	
	// Update component engine options
	componentOptions := usm.componentEngine.options
	componentOptions.MaxResults = max
	usm.componentEngine.SetOptions(componentOptions)
	
	// Update pipeline engine options
	pipelineOptions := usm.pipelineEngine.options
	pipelineOptions.MaxResults = max
	usm.pipelineEngine.SetOptions(pipelineOptions)
}

// LoadComponentItems loads and wraps component items for searching
func (usm *UnifiedSearchManager) LoadComponentItems(prompts, contexts, rules []ComponentItem) {
	var componentItems []*ComponentItemWrapper
	
	// Convert prompts
	for _, item := range prompts {
		wrapper := NewComponentItemWrapper(
			item.Name,
			item.Path,
			item.CompType,
			item.LastModified,
			item.UsageCount,
			item.TokenCount,
			item.Tags,
			item.IsArchived,
			usm.loadComponentContent(item.Path), // Load actual content for searching
		)
		componentItems = append(componentItems, wrapper)
	}
	
	// Convert contexts
	for _, item := range contexts {
		wrapper := NewComponentItemWrapper(
			item.Name,
			item.Path,
			item.CompType,
			item.LastModified,
			item.UsageCount,
			item.TokenCount,
			item.Tags,
			item.IsArchived,
			usm.loadComponentContent(item.Path),
		)
		componentItems = append(componentItems, wrapper)
	}
	
	// Convert rules
	for _, item := range rules {
		wrapper := NewComponentItemWrapper(
			item.Name,
			item.Path,
			item.CompType,
			item.LastModified,
			item.UsageCount,
			item.TokenCount,
			item.Tags,
			item.IsArchived,
			usm.loadComponentContent(item.Path),
		)
		componentItems = append(componentItems, wrapper)
	}
	
	// Set items in the component engine
	usm.componentEngine.SetItems(componentItems)
}

// LoadPipelineItems loads and wraps pipeline items for searching
func (usm *UnifiedSearchManager) LoadPipelineItems(pipelines []PipelineItem) {
	var pipelineItems []*PipelineItemWrapper
	
	for _, item := range pipelines {
		wrapper := NewPipelineItemWrapper(
			item.Name,
			item.Path,
			item.Tags,
			item.TokenCount,
			item.IsArchived,
			item.Modified,
			usm.loadPipelineContent(item.Path), // Load pipeline content for searching
		)
		pipelineItems = append(pipelineItems, wrapper)
	}
	
	// Set items in the pipeline engine
	usm.pipelineEngine.SetItems(pipelineItems)
}

// SearchComponents performs a search across components only
func (usm *UnifiedSearchManager) SearchComponents(query string, componentTypes []string) ([]SearchResult[*ComponentItemWrapper], error) {
	// If component types are specified, filter by them first
	if len(componentTypes) > 0 {
		// Get filtered results
		filteredResults := usm.componentEngine.GetFilteredItems(componentTypes, usm.includeArchived)
		if query == "" {
			return filteredResults, nil
		}
		
		// Create a temporary engine with filtered items
		tempEngine := NewSearchEngine[*ComponentItemWrapper]()
		var filteredItems []*ComponentItemWrapper
		for _, result := range filteredResults {
			filteredItems = append(filteredItems, result.Item)
		}
		tempEngine.SetItems(filteredItems)
		tempEngine.SetOptions(usm.componentEngine.options)
		
		return tempEngine.Search(query)
	}
	
	return usm.componentEngine.Search(query)
}

// SearchPipelines performs a search across pipelines only
func (usm *UnifiedSearchManager) SearchPipelines(query string) ([]SearchResult[*PipelineItemWrapper], error) {
	return usm.pipelineEngine.Search(query)
}

// SearchAll performs a unified search across both components and pipelines
func (usm *UnifiedSearchManager) SearchAll(query string) ([]SearchResult[*ComponentItemWrapper], []SearchResult[*PipelineItemWrapper], error) {
	// Search components and pipelines separately
	componentResults, err := usm.componentEngine.Search(query)
	if err != nil {
		return nil, nil, err
	}
	
	pipelineResults, err := usm.pipelineEngine.Search(query)
	if err != nil {
		return componentResults, nil, err
	}
	
	return componentResults, pipelineResults, nil
}

// FilterComponentsByQuery filters components using the existing search query parsing
func (usm *UnifiedSearchManager) FilterComponentsByQuery(query string, allPrompts, allContexts, allRules []ComponentItem) ([]ComponentItem, []ComponentItem, []ComponentItem, error) {
	// Load the components into the search engine
	usm.LoadComponentItems(allPrompts, allContexts, allRules)
	
	// Perform the search
	results, err := usm.componentEngine.Search(query)
	if err != nil {
		// Return all items on error
		return allPrompts, allContexts, allRules, err
	}
	
	// If no results, return empty slices
	if len(results) == 0 {
		return []ComponentItem{}, []ComponentItem{}, []ComponentItem{}, nil
	}
	
	// Convert results back to component items
	return usm.convertToComponentItems(results, allPrompts, allContexts, allRules)
}

// FilterPipelinesByQuery filters pipelines using the existing search query parsing
func (usm *UnifiedSearchManager) FilterPipelinesByQuery(query string, allPipelines []PipelineItem) ([]PipelineItem, error) {
	// Load the pipelines into the search engine
	usm.LoadPipelineItems(allPipelines)
	
	// Perform the search
	results, err := usm.pipelineEngine.Search(query)
	if err != nil {
		// Return all items on error
		return allPipelines, err
	}
	
	// If no results, return empty slice
	if len(results) == 0 {
		return []PipelineItem{}, nil
	}
	
	// Convert results back to pipeline items
	return usm.convertToPipelineItems(results, allPipelines), nil
}

// Helper methods

// loadComponentContent loads the actual content of a component for searching
func (usm *UnifiedSearchManager) loadComponentContent(path string) string {
	if path == "" {
		return ""
	}
	
	var component *models.Component
	var err error
	
	// Try to load as regular component first
	component, err = files.ReadComponent(path)
	if err != nil {
		// Try as archived component
		component, err = files.ReadArchivedComponent(path)
		if err != nil {
			return ""
		}
	}
	
	return component.Content
}

// loadPipelineContent loads pipeline information for searching
func (usm *UnifiedSearchManager) loadPipelineContent(path string) string {
	if path == "" {
		return ""
	}
	
	var pipeline *models.Pipeline
	var err error
	
	// Try to load as regular pipeline first
	pipeline, err = files.ReadPipeline(path)
	if err != nil {
		// Try as archived pipeline
		pipeline, err = files.ReadArchivedPipeline(path)
		if err != nil {
			return ""
		}
	}
	
	// Create searchable content from pipeline metadata
	content := pipeline.Name
	if len(pipeline.Tags) > 0 {
		content += " " + strings.Join(pipeline.Tags, " ")
	}
	
	// Add component paths for better searchability
	for _, comp := range pipeline.Components {
		if comp.Path != "" {
			content += " " + comp.Path
		}
	}
	
	return content
}

// convertToComponentItems converts search results back to component items
func (usm *UnifiedSearchManager) convertToComponentItems(results []SearchResult[*ComponentItemWrapper], allPrompts, allContexts, allRules []ComponentItem) ([]ComponentItem, []ComponentItem, []ComponentItem, error) {
	// Create path-to-item maps for quick lookup
	promptMap := make(map[string]ComponentItem)
	contextMap := make(map[string]ComponentItem)
	rulesMap := make(map[string]ComponentItem)
	
	for _, prompt := range allPrompts {
		promptMap[prompt.Path] = prompt
	}
	for _, context := range allContexts {
		contextMap[context.Path] = context
	}
	for _, rule := range allRules {
		rulesMap[rule.Path] = rule
	}
	
	// Convert results back to original types
	var filteredPrompts []ComponentItem
	var filteredContexts []ComponentItem
	var filteredRules []ComponentItem
	
	for _, result := range results {
		wrapper := result.Item
		path := wrapper.GetPath()
		compType := wrapper.GetSubType()
		
		switch compType {
		case models.ComponentTypePrompt:
			if item, exists := promptMap[path]; exists {
				filteredPrompts = append(filteredPrompts, item)
			}
		case models.ComponentTypeContext:
			if item, exists := contextMap[path]; exists {
				filteredContexts = append(filteredContexts, item)
			}
		case models.ComponentTypeRules:
			if item, exists := rulesMap[path]; exists {
				filteredRules = append(filteredRules, item)
			}
		}
	}
	
	return filteredPrompts, filteredContexts, filteredRules, nil
}

// convertToPipelineItems converts search results back to pipeline items
func (usm *UnifiedSearchManager) convertToPipelineItems(results []SearchResult[*PipelineItemWrapper], allPipelines []PipelineItem) []PipelineItem {
	// Create path-to-item map for quick lookup
	pipelineMap := make(map[string]PipelineItem)
	for _, pipeline := range allPipelines {
		pipelineMap[pipeline.Path] = pipeline
	}
	
	// Convert results back to original types
	var filteredPipelines []PipelineItem
	for _, result := range results {
		wrapper := result.Item
		path := wrapper.GetPath()
		
		if item, exists := pipelineMap[path]; exists {
			filteredPipelines = append(filteredPipelines, item)
		}
	}
	
	return filteredPipelines
}

// IsStructuredQuery checks if a query uses structured search syntax
func (usm *UnifiedSearchManager) IsStructuredQuery(query string) bool {
	return usm.componentEngine.isStructuredQuery(query)
}

// ParseSearchQuery parses a search query using the unified parser
func (usm *UnifiedSearchManager) ParseSearchQuery(query string) (*ParsedQuery, error) {
	return ParseQuery(query), nil
}

// GetSearchSuggestions returns search suggestions based on current items
func (usm *UnifiedSearchManager) GetSearchSuggestions(partial string) []string {
	var suggestions []string
	seen := make(map[string]bool)
	
	partial = strings.ToLower(partial)
	
	// Get tag suggestions from components
	for _, item := range usm.componentEngine.items {
		for _, tag := range item.GetTags() {
			normalizedTag := models.NormalizeTagName(tag)
			if strings.HasPrefix(strings.ToLower(normalizedTag), partial) && !seen[normalizedTag] {
				suggestions = append(suggestions, "tag:"+normalizedTag)
				seen[normalizedTag] = true
			}
		}
		
		// Get name suggestions
		name := strings.ToLower(item.GetName())
		if strings.HasPrefix(name, partial) && !seen[name] {
			suggestions = append(suggestions, item.GetName())
			seen[name] = true
		}
	}
	
	// Get tag suggestions from pipelines
	for _, item := range usm.pipelineEngine.items {
		for _, tag := range item.GetTags() {
			normalizedTag := models.NormalizeTagName(tag)
			if strings.HasPrefix(strings.ToLower(normalizedTag), partial) && !seen[normalizedTag] {
				suggestions = append(suggestions, "tag:"+normalizedTag)
				seen[normalizedTag] = true
			}
		}
		
		// Get name suggestions
		name := strings.ToLower(item.GetName())
		if strings.HasPrefix(name, partial) && !seen[name] {
			suggestions = append(suggestions, item.GetName())
			seen[name] = true
		}
	}
	
	// Add type suggestions
	typeFilters := []string{"type:prompt", "type:context", "type:rules", "type:pipeline"}
	for _, filter := range typeFilters {
		if strings.HasPrefix(filter, "type:"+partial) && !seen[filter] {
			suggestions = append(suggestions, filter)
			seen[filter] = true
		}
	}
	
	// Add status suggestions
	statusFilters := []string{"status:archived", "status:active"}
	for _, filter := range statusFilters {
		if strings.HasPrefix(filter, "status:"+partial) && !seen[filter] {
			suggestions = append(suggestions, filter)
			seen[filter] = true
		}
	}
	
	return suggestions
}