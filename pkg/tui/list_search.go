package tui

import (
	"sort"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// SearchManager handles search initialization and operations
type SearchManager struct {
	engine *search.Engine
	query  string
}

// NewSearchManager creates a new search manager
func NewSearchManager() *SearchManager {
	return &SearchManager{}
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

// FilterSearchResults filters pipelines and components based on search results
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
