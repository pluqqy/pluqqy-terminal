package search

import (
	"fmt"
	"time"
)

// SearchAPI provides a high-level interface for searching
type SearchAPI struct {
	engine *Engine
}

// NewSearchAPI creates a new search API instance
func NewSearchAPI() (*SearchAPI, error) {
	api := &SearchAPI{
		engine: NewEngine(),
	}
	
	// Build initial index
	if err := api.engine.BuildIndex(); err != nil {
		return nil, fmt.Errorf("failed to build search index: %w", err)
	}
	
	return api, nil
}

// RefreshIndex rebuilds the search index
func (api *SearchAPI) RefreshIndex() error {
	return api.engine.BuildIndex()
}

// Search performs a search with the given query
func (api *SearchAPI) Search(query string) ([]SearchResult, error) {
	return api.engine.Search(query)
}

// SearchOptions represents options for searching
type SearchOptions struct {
	Query      string
	Tags       []string
	Types      []string
	MaxResults int
	SortBy     string // "relevance", "name", "modified"
}

// SearchWithOptions performs a search with specific options
func (api *SearchAPI) SearchWithOptions(opts SearchOptions) ([]SearchResult, error) {
	// Build query from options
	query := opts.Query
	
	// Add tag filters
	for _, tag := range opts.Tags {
		if query != "" {
			query += " AND "
		}
		query += fmt.Sprintf("tag:%s", tag)
	}
	
	// Add type filters
	for i, typ := range opts.Types {
		if i == 0 && query != "" {
			query += " AND ("
		} else if i == 0 {
			query += "("
		} else {
			query += " OR "
		}
		query += fmt.Sprintf("type:%s", typ)
		if i == len(opts.Types)-1 {
			query += ")"
		}
	}
	
	// Perform search
	results, err := api.engine.Search(query)
	if err != nil {
		return nil, err
	}
	
	// Apply sorting
	switch opts.SortBy {
	case "name":
		sortResultsByName(results)
	case "modified":
		sortResultsByModified(results)
	// "relevance" is default, already sorted by score
	}
	
	// Limit results
	if opts.MaxResults > 0 && len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}
	
	return results, nil
}

// QuickFilter provides simple tag-based filtering
func (api *SearchAPI) QuickFilter(tags []string) ([]SearchResult, error) {
	if len(tags) == 0 {
		// Return all items
		return api.Search("")
	}
	
	query := ""
	for i, tag := range tags {
		if i > 0 {
			query += " AND "
		}
		query += fmt.Sprintf("tag:%s", tag)
	}
	
	return api.engine.Search(query)
}

// SearchComponents searches only components
func (api *SearchAPI) SearchComponents(query string, componentType string) ([]SearchResult, error) {
	fullQuery := query
	if componentType != "" {
		if fullQuery != "" {
			fullQuery += " AND "
		}
		fullQuery += fmt.Sprintf("type:%s", componentType)
	} else {
		if fullQuery != "" {
			fullQuery += " AND "
		}
		fullQuery += "type:component"
	}
	
	return api.engine.Search(fullQuery)
}

// SearchPipelines searches only pipelines
func (api *SearchAPI) SearchPipelines(query string) ([]SearchResult, error) {
	fullQuery := query
	if fullQuery != "" {
		fullQuery += " AND "
	}
	fullQuery += "type:pipeline"
	
	return api.engine.Search(fullQuery)
}

// GetRecentItems returns items modified within the specified duration
func (api *SearchAPI) GetRecentItems(since time.Duration) ([]SearchResult, error) {
	// Convert duration to search query format
	days := int(since.Hours() / 24)
	query := fmt.Sprintf("modified:>%dd", days)
	
	return api.engine.Search(query)
}

// GetItemsByTags returns items that have all specified tags
func (api *SearchAPI) GetItemsByTags(tags []string) ([]SearchResult, error) {
	if len(tags) == 0 {
		return []SearchResult{}, nil
	}
	
	query := ""
	for i, tag := range tags {
		if i > 0 {
			query += " AND "
		}
		query += fmt.Sprintf("tag:%s", tag)
	}
	
	return api.engine.Search(query)
}

// GetItemsByAnyTags returns items that have any of the specified tags
func (api *SearchAPI) GetItemsByAnyTags(tags []string) ([]SearchResult, error) {
	if len(tags) == 0 {
		return []SearchResult{}, nil
	}
	
	query := ""
	for i, tag := range tags {
		if i > 0 {
			query += " OR "
		}
		query += fmt.Sprintf("tag:%s", tag)
	}
	
	return api.engine.Search(query)
}

// SearchByContent performs a full-text content search
func (api *SearchAPI) SearchByContent(searchTerm string) ([]SearchResult, error) {
	return api.engine.Search(fmt.Sprintf("content:%q", searchTerm))
}

// Helper functions for sorting

func sortResultsByName(results []SearchResult) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Item.Name < results[i].Item.Name {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func sortResultsByModified(results []SearchResult) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Item.Modified.After(results[i].Item.Modified) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}