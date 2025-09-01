package tui

import (
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search/unified"
)

// SearchManager handles search initialization and operations
type SearchManager struct {
	// Unified search helper
	helper *unified.SearchHelper
	
	query  string
}

// NewSearchManager creates a new search manager
func NewSearchManager() *SearchManager {
	return &SearchManager{
		helper: unified.NewSearchHelper(),
	}
}

// InitializeEngine creates and initializes the search engine (legacy compatibility)
func (s *SearchManager) InitializeEngine() error {
	// No longer needed with unified search - always return success
	return nil
}

// SetQuery updates the search query
func (s *SearchManager) SetQuery(query string) {
	s.query = query
}

// GetQuery returns the current search query
func (s *SearchManager) GetQuery() string {
	return s.query
}

// Search performs a search with the current query (legacy compatibility)
// Returns empty results as this method is no longer used
func (s *SearchManager) Search() ([]interface{}, error) {
	return nil, nil
}

// UnifiedSearchComponents performs a unified search on components
func (s *SearchManager) UnifiedSearchComponents(prompts, contexts, rules []unified.ComponentItem) ([]unified.ComponentItem, []unified.ComponentItem, []unified.ComponentItem, error) {
	return s.helper.UnifiedFilterComponents(s.query, prompts, contexts, rules)
}

// UnifiedSearchPipelines performs a unified search on pipelines  
func (s *SearchManager) UnifiedSearchPipelines(pipelines []unified.PipelineItem) ([]unified.PipelineItem, error) {
	return s.helper.UnifiedFilterPipelines(s.query, pipelines)
}

// UnifiedSearchAll performs a unified search across all items
func (s *SearchManager) UnifiedSearchAll(prompts, contexts, rules []unified.ComponentItem, pipelines []unified.PipelineItem) ([]unified.ComponentItem, []unified.ComponentItem, []unified.ComponentItem, []unified.PipelineItem, error) {
	return s.helper.UnifiedFilterAll(s.query, prompts, contexts, rules, pipelines)
}

// GetSearchHelper returns the unified search helper for advanced operations
func (s *SearchManager) GetSearchHelper() *unified.SearchHelper {
	return s.helper
}

// FilterSearchResultsUnified uses the new unified search system for better performance
func FilterSearchResultsUnified(query string, pipelines []pipelineItem, components []componentItem) ([]pipelineItem, []componentItem) {
	// Convert TUI types to shared types
	sharedPipelines := convertTUIPipelinesToShared(pipelines)
	sharedComponents := convertTUIComponentsToShared(components)
	
	// Use unified search helper
	helper := unified.NewSearchHelper()
	helper.SetSearchOptions(unified.ShouldIncludeArchived(query), 1000, "relevance")
	
	// Separate components by type for the search
	var prompts, contexts, rules []unified.ComponentItem
	for _, comp := range sharedComponents {
		switch comp.CompType {
		case models.ComponentTypePrompt:
			prompts = append(prompts, comp)
		case models.ComponentTypeContext:
			contexts = append(contexts, comp)
		case models.ComponentTypeRules:
			rules = append(rules, comp)
		}
	}
	
	// Perform unified search
	filteredPrompts, filteredContexts, filteredRules, filteredPipelines, err := helper.UnifiedFilterAll(query, prompts, contexts, rules, sharedPipelines)
	if err != nil {
		// Fallback to return all items on error
		return pipelines, components
	}
	
	// Convert back to TUI types
	resultComponents := unified.CombineComponentsByType(filteredPrompts, filteredContexts, filteredRules)
	tuiComponents := convertSharedComponentsToTUIList(resultComponents)
	tuiPipelines := convertSharedPipelinesToTUI(filteredPipelines)
	
	return tuiPipelines, tuiComponents
}


// Helper functions for type conversion

// convertTUIPipelinesToShared converts TUI pipelineItem slice to unified PipelineItem slice
func convertTUIPipelinesToShared(pipelines []pipelineItem) []unified.PipelineItem {
	shared_pipelines := make([]unified.PipelineItem, len(pipelines))
	for i, p := range pipelines {
		shared_pipelines[i] = unified.ConvertTUIPipelineItemToShared(p.name, p.path, p.tags, p.tokenCount, p.isArchived)
	}
	return shared_pipelines
}

// convertTUIComponentsToShared converts TUI componentItem slice to unified ComponentItem slice  
func convertTUIComponentsToShared(components []componentItem) []unified.ComponentItem {
	shared_components := make([]unified.ComponentItem, len(components))
	for i, c := range components {
		shared_components[i] = unified.ConvertTUIComponentItemToShared(c.name, c.path, c.compType, c.lastModified, c.usageCount, c.tokenCount, c.tags, c.isArchived)
	}
	return shared_components
}

// convertSharedPipelinesToTUI converts unified PipelineItem slice back to TUI pipelineItem slice
func convertSharedPipelinesToTUI(pipelines []unified.PipelineItem) []pipelineItem {
	tui_pipelines := make([]pipelineItem, len(pipelines))
	for i, p := range pipelines {
		tui_pipelines[i] = pipelineItem{
			name:       p.Name,
			path:       p.Path,
			tags:       p.Tags,
			tokenCount: p.TokenCount,
			isArchived: p.IsArchived,
		}
	}
	return tui_pipelines
}

// convertSharedComponentsToTUIList converts unified ComponentItem slice back to TUI componentItem slice (list version)
func convertSharedComponentsToTUIList(components []unified.ComponentItem) []componentItem {
	tui_components := make([]componentItem, len(components))
	for i, c := range components {
		tui_components[i] = componentItem{
			name:         c.Name,
			path:         c.Path,
			compType:     c.CompType,
			lastModified: c.LastModified,
			usageCount:   c.UsageCount,
			tokenCount:   c.TokenCount,
			tags:         c.Tags,
			isArchived:   c.IsArchived,
		}
	}
	return tui_components
}
