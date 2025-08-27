package tui

import (
	"sort"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

// SearchManager handles search initialization and operations
type SearchManager struct {
	// Legacy engine for backward compatibility
	engine *search.Engine
	
	// New unified search helper
	helper *shared.SearchHelper
	
	query  string
}

// NewSearchManager creates a new search manager
func NewSearchManager() *SearchManager {
	return &SearchManager{
		helper: shared.NewSearchHelper(),
	}
}

// InitializeEngine creates and initializes the search engine
func (s *SearchManager) InitializeEngine() error {
	engine := search.NewEngine()
	if err := engine.BuildIndex(); err != nil {
		return err
	}
	s.engine = engine
	return nil
}

// GetEngine returns the search engine
func (s *SearchManager) GetEngine() *search.Engine {
	return s.engine
}

// SetQuery updates the search query
func (s *SearchManager) SetQuery(query string) {
	s.query = query
}

// GetQuery returns the current search query
func (s *SearchManager) GetQuery() string {
	return s.query
}

// Search performs a search with the current query
func (s *SearchManager) Search() ([]search.SearchResult, error) {
	if s.query == "" || s.engine == nil {
		return nil, nil
	}
	return s.engine.Search(s.query)
}

// UnifiedSearchComponents performs a unified search on components
func (s *SearchManager) UnifiedSearchComponents(prompts, contexts, rules []shared.ComponentItem) ([]shared.ComponentItem, []shared.ComponentItem, []shared.ComponentItem, error) {
	return s.helper.UnifiedFilterComponents(s.query, prompts, contexts, rules)
}

// UnifiedSearchPipelines performs a unified search on pipelines  
func (s *SearchManager) UnifiedSearchPipelines(pipelines []shared.PipelineItem) ([]shared.PipelineItem, error) {
	return s.helper.UnifiedFilterPipelines(s.query, pipelines)
}

// UnifiedSearchAll performs a unified search across all items
func (s *SearchManager) UnifiedSearchAll(prompts, contexts, rules []shared.ComponentItem, pipelines []shared.PipelineItem) ([]shared.ComponentItem, []shared.ComponentItem, []shared.ComponentItem, []shared.PipelineItem, error) {
	return s.helper.UnifiedFilterAll(s.query, prompts, contexts, rules, pipelines)
}

// GetSearchHelper returns the unified search helper for advanced operations
func (s *SearchManager) GetSearchHelper() *shared.SearchHelper {
	return s.helper
}

// FilterSearchResultsUnified uses the new unified search system for better performance
func FilterSearchResultsUnified(query string, pipelines []pipelineItem, components []componentItem) ([]pipelineItem, []componentItem) {
	// Convert TUI types to shared types
	sharedPipelines := convertTUIPipelinesToShared(pipelines)
	sharedComponents := convertTUIComponentsToShared(components)
	
	// Use unified search helper
	helper := shared.NewSearchHelper()
	helper.SetSearchOptions(shared.ShouldIncludeArchived(query), 1000, "relevance")
	
	// Separate components by type for the search
	var prompts, contexts, rules []shared.ComponentItem
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
		// Fallback to original function
		return FilterSearchResults(nil, pipelines, components)
	}
	
	// Convert back to TUI types
	resultComponents := shared.CombineComponentsByType(filteredPrompts, filteredContexts, filteredRules)
	tuiComponents := convertSharedComponentsToTUIList(resultComponents)
	tuiPipelines := convertSharedPipelinesToTUI(filteredPipelines)
	
	return tuiPipelines, tuiComponents
}

// FilterSearchResults filters pipelines and components based on search results (legacy compatibility)
func FilterSearchResults(results []search.SearchResult, pipelines []pipelineItem, components []componentItem) ([]pipelineItem, []componentItem) {
	if results == nil || len(results) == 0 {
		// No search results, return empty lists
		return []pipelineItem{}, []componentItem{}
	}

	// Build filtered lists from search results
	filteredPipelines := []pipelineItem{}
	filteredComponents := []componentItem{}

	// Use maps to track what's already added
	addedPipelines := make(map[string]bool)
	addedComponents := make(map[string]bool)

	for _, result := range results {
		if result.Item.Type == search.ItemTypePipeline {
			// Match by path for accuracy (handles archived items correctly)
			if !addedPipelines[result.Item.Path] {
				for _, p := range pipelines {
					if p.path == result.Item.Path {
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
					if c.path == result.Item.Path {
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
		orderI, okI := typeOrder[filteredComponents[i].compType]
		if !okI {
			orderI = 4
		}
		orderJ, okJ := typeOrder[filteredComponents[j].compType]
		if !okJ {
			orderJ = 4
		}

		// Sort by type order first
		if orderI != orderJ {
			return orderI < orderJ
		}

		// Within same type, sort by name
		return filteredComponents[i].name < filteredComponents[j].name
	})

	return filteredPipelines, filteredComponents
}

// Helper functions for type conversion

// convertTUIPipelinesToShared converts TUI pipelineItem slice to shared PipelineItem slice
func convertTUIPipelinesToShared(pipelines []pipelineItem) []shared.PipelineItem {
	shared_pipelines := make([]shared.PipelineItem, len(pipelines))
	for i, p := range pipelines {
		shared_pipelines[i] = shared.ConvertTUIPipelineItemToShared(p.name, p.path, p.tags, p.tokenCount, p.isArchived)
	}
	return shared_pipelines
}

// convertTUIComponentsToShared converts TUI componentItem slice to shared ComponentItem slice  
func convertTUIComponentsToShared(components []componentItem) []shared.ComponentItem {
	shared_components := make([]shared.ComponentItem, len(components))
	for i, c := range components {
		shared_components[i] = shared.ConvertTUIComponentItemToShared(c.name, c.path, c.compType, c.lastModified, c.usageCount, c.tokenCount, c.tags, c.isArchived)
	}
	return shared_components
}

// convertSharedPipelinesToTUI converts shared PipelineItem slice back to TUI pipelineItem slice
func convertSharedPipelinesToTUI(pipelines []shared.PipelineItem) []pipelineItem {
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

// convertSharedComponentsToTUIList converts shared ComponentItem slice back to TUI componentItem slice (list version)
func convertSharedComponentsToTUIList(components []shared.ComponentItem) []componentItem {
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
