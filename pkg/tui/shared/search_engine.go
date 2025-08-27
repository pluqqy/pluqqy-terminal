package shared

import (
	"sort"
	"strings"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// SearchMode defines the type of search being performed
type SearchMode string

const (
	SearchModeAll     SearchMode = "all"     // Search name, content, and tags
	SearchModeName    SearchMode = "name"    // Search only name/title
	SearchModeContent SearchMode = "content" // Search only content
	SearchModeTags    SearchMode = "tags"    // Search only tags
)

// Searchable represents an item that can be searched
type Searchable interface {
	GetName() string
	GetPath() string
	GetType() string        // e.g., "pipeline", "prompt", "context", "rules"
	GetSubType() string     // For pipelines: "", for components: component type
	GetTags() []string
	GetContent() string     // Searchable content
	GetModified() time.Time
	IsArchived() bool
	GetTokenCount() int
	GetUsageCount() int // For components, returns usage count
}

// SearchResult represents a ranked search result
type SearchResult[T Searchable] struct {
	Item      T
	Score     float64
	Relevance SearchRelevance
}

// SearchRelevance provides detailed relevance information
type SearchRelevance struct {
	NameMatch    bool
	ContentMatch bool
	TagMatch     bool
	ExactMatch   bool // Exact name or tag match
	Highlights   map[string][]string // Field -> highlighted excerpts
}

// SearchOptions configures search behavior
type SearchOptions struct {
	Mode           SearchMode
	MaxResults     int
	IncludeArchived bool
	SortBy         string // "relevance", "name", "modified", "usage"
}

// SearchEngine provides a unified search interface for both components and pipelines
type SearchEngine[T Searchable] struct {
	// Internal search engine from pkg/search
	engine *search.Engine
	
	// Search query parsing
	parser *search.Parser
	
	// Current items being searched
	items []T
	
	// Search options
	options SearchOptions
}

// NewSearchEngine creates a new search engine for a specific type
func NewSearchEngine[T Searchable]() *SearchEngine[T] {
	return &SearchEngine[T]{
		engine: search.NewEngine(),
		parser: search.NewParser(),
		options: SearchOptions{
			Mode:           SearchModeAll,
			MaxResults:     100,
			IncludeArchived: false,
			SortBy:         "relevance",
		},
	}
}

// SetItems updates the items to search through
func (se *SearchEngine[T]) SetItems(items []T) {
	se.items = items
}

// SetOptions updates search configuration
func (se *SearchEngine[T]) SetOptions(options SearchOptions) {
	se.options = options
}

// Search performs a search with the given query
func (se *SearchEngine[T]) Search(query string) ([]SearchResult[T], error) {
	if query == "" {
		// No query, return all items (filtered by archived status)
		return se.getAllItems(), nil
	}
	
	// Try to use the search engine first for structured queries
	if se.isStructuredQuery(query) {
		return se.searchWithEngine(query)
	}
	
	// Fall back to simple text search
	return se.simpleTextSearch(query), nil
}

// SearchByMode performs a search in a specific mode
func (se *SearchEngine[T]) SearchByMode(query string, mode SearchMode) ([]SearchResult[T], error) {
	originalMode := se.options.Mode
	se.options.Mode = mode
	
	results, err := se.Search(query)
	
	se.options.Mode = originalMode
	return results, err
}

// GetFilteredItems returns items filtered by type and archived status
func (se *SearchEngine[T]) GetFilteredItems(itemTypes []string, includeArchived bool) []SearchResult[T] {
	var results []SearchResult[T]
	
	for _, item := range se.items {
		// Check archived status
		if !includeArchived && item.IsArchived() {
			continue
		}
		
		// Check type filter
		if len(itemTypes) > 0 {
			typeMatch := false
			itemType := strings.ToLower(item.GetType())
			itemSubType := strings.ToLower(item.GetSubType())
			
			for _, filterType := range itemTypes {
				filterType = strings.ToLower(filterType)
				if itemType == filterType || itemSubType == filterType {
					typeMatch = true
					break
				}
				// Handle plural forms - both ways
				if itemType+"s" == filterType || itemSubType+"s" == filterType ||
				   itemType == strings.TrimSuffix(filterType, "s") || itemSubType == strings.TrimSuffix(filterType, "s") {
					typeMatch = true
					break
				}
			}
			
			if !typeMatch {
				continue
			}
		}
		
		results = append(results, SearchResult[T]{
			Item:  item,
			Score: 1.0,
			Relevance: SearchRelevance{
				NameMatch:    false,
				ContentMatch: false,
				TagMatch:     false,
				ExactMatch:   false,
				Highlights:   make(map[string][]string),
			},
		})
	}
	
	return results
}

// isStructuredQuery checks if the query contains structured search syntax
func (se *SearchEngine[T]) isStructuredQuery(query string) bool {
	// Check for field-based searches (tag:, type:, status:, etc.)
	structuredPatterns := []string{"tag:", "type:", "status:", "name:", "content:", "modified:"}
	
	lowerQuery := strings.ToLower(query)
	for _, pattern := range structuredPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return true
		}
	}
	
	return false
}

// searchWithEngine uses the search engine for structured queries
func (se *SearchEngine[T]) searchWithEngine(query string) ([]SearchResult[T], error) {
	// Parse structured query to handle tag:, type:, status: etc.
	lowerQuery := strings.ToLower(query)
	var results []SearchResult[T]
	
	// Handle tag: queries
	if strings.HasPrefix(lowerQuery, "tag:") {
		tagQuery := strings.TrimPrefix(lowerQuery, "tag:")
		tagQuery = strings.TrimSpace(tagQuery)
		
		for _, item := range se.items {
			// Skip archived items if not included
			if !se.options.IncludeArchived && item.IsArchived() {
				continue
			}
			
			// Check if item has the requested tag
			tags := item.GetTags()
			hasTag := false
			for _, tag := range tags {
				if strings.EqualFold(tag, tagQuery) {
					hasTag = true
					break
				}
			}
			
			if hasTag {
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: 10.0, // High score for exact tag match
					Relevance: SearchRelevance{
						TagMatch:   true,
						ExactMatch: true,
						Highlights: map[string][]string{
							"tags": {tagQuery},
						},
					},
				})
			}
		}
		
		// Sort by score and name
		se.sortResults(results)
		return results, nil
	}
	
	// Handle status: queries
	if strings.HasPrefix(lowerQuery, "status:") {
		statusQuery := strings.TrimPrefix(lowerQuery, "status:")
		statusQuery = strings.TrimSpace(statusQuery)
		
		for _, item := range se.items {
			isArchived := item.IsArchived()
			
			// Match based on status
			if (statusQuery == "archived" && isArchived) ||
			   (statusQuery == "active" && !isArchived) {
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: 10.0,
					Relevance: SearchRelevance{
						ExactMatch: true,
						Highlights: make(map[string][]string),
					},
				})
			}
		}
		
		// Sort by score and name
		se.sortResults(results)
		return results, nil
	}
	
	// Handle type: queries
	if strings.HasPrefix(lowerQuery, "type:") {
		typeQuery := strings.TrimPrefix(lowerQuery, "type:")
		typeQuery = strings.TrimSpace(typeQuery)
		
		for _, item := range se.items {
			// Skip archived items if not included
			if !se.options.IncludeArchived && item.IsArchived() {
				continue
			}
			
			itemType := strings.ToLower(item.GetType())
			itemSubType := strings.ToLower(item.GetSubType())
			
			// Match exact type or handle plural/singular forms
			// Check exact match
			if itemType == typeQuery || itemSubType == typeQuery {
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: 10.0,
					Relevance: SearchRelevance{
						ExactMatch: true,
						Highlights: make(map[string][]string),
					},
				})
				continue
			}
			
			// Check plural/singular variations
			// If query is plural, check if item is singular
			if itemType+"s" == typeQuery || itemSubType+"s" == typeQuery {
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: 10.0,
					Relevance: SearchRelevance{
						ExactMatch: true,
						Highlights: make(map[string][]string),
					},
				})
				continue
			}
			
			// If query is singular, check if item is plural
			if strings.TrimSuffix(itemType, "s") == typeQuery || strings.TrimSuffix(itemSubType, "s") == typeQuery {
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: 10.0,
					Relevance: SearchRelevance{
						ExactMatch: true,
						Highlights: make(map[string][]string),
					},
				})
			}
		}
		
		// Sort by score and name
		se.sortResults(results)
		return results, nil
	}
	
	// Handle content: queries
	if strings.HasPrefix(lowerQuery, "content:") {
		contentQuery := strings.TrimPrefix(lowerQuery, "content:")
		contentQuery = strings.TrimSpace(contentQuery)
		// Remove quotes if present
		contentQuery = strings.Trim(contentQuery, `"'`)
		
		for _, item := range se.items {
			// Skip archived items if not included
			if !se.options.IncludeArchived && item.IsArchived() {
				continue
			}
			
			// Check if item content contains the search term
			content := strings.ToLower(item.GetContent())
			if strings.Contains(content, strings.ToLower(contentQuery)) {
				// Calculate score based on number of matches
				score := float64(strings.Count(content, strings.ToLower(contentQuery)))
				
				// Extract excerpts around matches
				excerpts := extractExcerpts(item.GetContent(), contentQuery, 3, 100)
				
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: score,
					Relevance: SearchRelevance{
						ContentMatch: true,
						ExactMatch:   false,
						Highlights: map[string][]string{
							"content": excerpts,
						},
					},
				})
			}
		}
		
		// Sort by score and name
		se.sortResults(results)
		return results, nil
	}
	
	// Handle name: queries
	if strings.HasPrefix(lowerQuery, "name:") {
		nameQuery := strings.TrimPrefix(lowerQuery, "name:")
		nameQuery = strings.TrimSpace(nameQuery)
		// Remove quotes if present
		nameQuery = strings.Trim(nameQuery, `"'`)
		
		for _, item := range se.items {
			// Skip archived items if not included
			if !se.options.IncludeArchived && item.IsArchived() {
				continue
			}
			
			// Check if item name contains the search term
			name := strings.ToLower(item.GetName())
			if strings.Contains(name, strings.ToLower(nameQuery)) {
				// Higher score for exact match
				score := 5.0
				exactMatch := false
				if name == strings.ToLower(nameQuery) {
					score = 10.0
					exactMatch = true
				}
				
				results = append(results, SearchResult[T]{
					Item:  item,
					Score: score,
					Relevance: SearchRelevance{
						NameMatch:  true,
						ExactMatch: exactMatch,
						Highlights: map[string][]string{
							"name": {item.GetName()},
						},
					},
				})
			}
		}
		
		// Sort by score and name
		se.sortResults(results)
		return results, nil
	}
	
	// For other structured queries, try to use the search engine if available
	if se.engine != nil {
		// Build the search index from current items
		if err := se.buildSearchIndex(); err != nil {
			// Fallback to simple search
			return se.simpleTextSearch(query), nil
		}
		
		// Perform the search
		engineResults, err := se.engine.Search(query)
		if err != nil {
			// Fallback to simple search
			return se.simpleTextSearch(query), nil
		}
		
		// Convert search results to our format
		return se.convertSearchResults(engineResults), nil
	}
	
	// Fallback to simple search for unrecognized structured queries
	return se.simpleTextSearch(query), nil
}

// buildSearchIndex builds the search index from current items
func (se *SearchEngine[T]) buildSearchIndex() error {
	// Clear the existing index
	se.engine = search.NewEngine()
	
	// This would require integration with the search engine's BuildIndex method
	// For now, we'll implement this as a simple placeholder
	// In a real implementation, we'd need to adapt the search.Engine to work with generic types
	return nil
}

// convertSearchResults converts search.SearchResult to our SearchResult format
func (se *SearchEngine[T]) convertSearchResults(results []search.SearchResult) []SearchResult[T] {
	var converted []SearchResult[T]
	
	// Create a path-to-item map for quick lookup
	pathMap := make(map[string]T)
	for _, item := range se.items {
		pathMap[item.GetPath()] = item
	}
	
	for _, result := range results {
		if item, exists := pathMap[result.Item.Path]; exists {
			converted = append(converted, SearchResult[T]{
				Item:  item,
				Score: result.Score,
				Relevance: SearchRelevance{
					NameMatch:    se.hasNameMatch(result.Highlights),
					ContentMatch: se.hasContentMatch(result.Highlights),
					TagMatch:     se.hasTagMatch(result.Highlights),
					ExactMatch:   result.Score > 2.0, // High score indicates exact match
					Highlights:   result.Highlights,
				},
			})
		}
	}
	
	return converted
}

// simpleTextSearch performs a simple text-based search
func (se *SearchEngine[T]) simpleTextSearch(query string) []SearchResult[T] {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return se.getAllItems()
	}
	
	var results []SearchResult[T]
	queryTerms := strings.Fields(query)
	
	for _, item := range se.items {
		// Skip archived items if not included
		if !se.options.IncludeArchived && item.IsArchived() {
			continue
		}
		
		score, relevance := se.calculateSimpleScore(item, queryTerms)
		if score > 0 {
			results = append(results, SearchResult[T]{
				Item:      item,
				Score:     score,
				Relevance: relevance,
			})
		}
	}
	
	// Sort by score
	se.sortResults(results)
	
	// Apply max results limit
	if se.options.MaxResults > 0 && len(results) > se.options.MaxResults {
		results = results[:se.options.MaxResults]
	}
	
	return results
}

// calculateSimpleScore calculates relevance score for simple text search
func (se *SearchEngine[T]) calculateSimpleScore(item T, queryTerms []string) (float64, SearchRelevance) {
	name := strings.ToLower(item.GetName())
	content := strings.ToLower(item.GetContent())
	tags := make([]string, len(item.GetTags()))
	for i, tag := range item.GetTags() {
		tags[i] = strings.ToLower(models.NormalizeTagName(tag))
	}
	
	score := 0.0
	relevance := SearchRelevance{
		Highlights: make(map[string][]string),
	}
	
	for _, term := range queryTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		
		// Check name matches
		if se.shouldSearchField("name") {
			if name == term {
				score += 10.0 // Exact name match
				relevance.ExactMatch = true
				relevance.NameMatch = true
				relevance.Highlights["name"] = []string{item.GetName()}
			} else if strings.HasPrefix(name, term) {
				score += 5.0 // Name prefix match
				relevance.NameMatch = true
				relevance.Highlights["name"] = []string{item.GetName()}
			} else if strings.Contains(name, term) {
				score += 3.0 // Name contains match
				relevance.NameMatch = true
				relevance.Highlights["name"] = []string{item.GetName()}
			}
		}
		
		// Check tag matches
		if se.shouldSearchField("tags") {
			for _, tag := range tags {
				if tag == term {
					score += 8.0 // Exact tag match
					relevance.ExactMatch = true
					relevance.TagMatch = true
				} else if strings.HasPrefix(tag, term) {
					score += 4.0 // Tag prefix match
					relevance.TagMatch = true
				} else if strings.Contains(tag, term) {
					score += 2.0 // Tag contains match
					relevance.TagMatch = true
				}
			}
		}
		
		// Check content matches
		if se.shouldSearchField("content") && strings.Contains(content, term) {
			score += 1.0 // Content match
			relevance.ContentMatch = true
			// Extract excerpt around the match
			if excerpts := extractExcerpts(item.GetContent(), term, 1, 50); len(excerpts) > 0 {
				relevance.Highlights["content"] = excerpts
			}
		}
	}
	
	// Boost for recent items
	age := time.Since(item.GetModified())
	if age < 24*time.Hour {
		score += 1.0
	} else if age < 7*24*time.Hour {
		score += 0.5
	}
	
	// Boost for frequently used components
	if usage := item.GetUsageCount(); usage > 0 {
		score += float64(usage) * 0.1
	}
	
	return score, relevance
}

// shouldSearchField checks if a field should be searched based on current mode
func (se *SearchEngine[T]) shouldSearchField(field string) bool {
	switch se.options.Mode {
	case SearchModeAll:
		return true
	case SearchModeName:
		return field == "name"
	case SearchModeContent:
		return field == "content"
	case SearchModeTags:
		return field == "tags"
	default:
		return true
	}
}

// getAllItems returns all items as search results (respecting archived filter)
func (se *SearchEngine[T]) getAllItems() []SearchResult[T] {
	var results []SearchResult[T]
	
	for _, item := range se.items {
		if !se.options.IncludeArchived && item.IsArchived() {
			continue
		}
		
		results = append(results, SearchResult[T]{
			Item:  item,
			Score: 1.0,
			Relevance: SearchRelevance{
				Highlights: make(map[string][]string),
			},
		})
	}
	
	se.sortResults(results)
	return results
}

// sortResults sorts search results based on current sort option
func (se *SearchEngine[T]) sortResults(results []SearchResult[T]) {
	switch se.options.SortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Item.GetName() < results[j].Item.GetName()
		})
	case "modified":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Item.GetModified().After(results[j].Item.GetModified())
		})
	case "usage":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Item.GetUsageCount() > results[j].Item.GetUsageCount()
		})
	case "relevance":
		fallthrough
	default:
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
}

// Helper functions for search result analysis

func (se *SearchEngine[T]) hasNameMatch(highlights map[string][]string) bool {
	_, exists := highlights["name"]
	return exists
}

func (se *SearchEngine[T]) hasContentMatch(highlights map[string][]string) bool {
	_, exists := highlights["content"]
	return exists
}

func (se *SearchEngine[T]) hasTagMatch(highlights map[string][]string) bool {
	_, exists := highlights["tags"]
	return exists
}

// extractExcerpts extracts highlighted excerpts from content
func extractExcerpts(content, searchTerm string, maxExcerpts, contextChars int) []string {
	var excerpts []string
	lowerContent := strings.ToLower(content)
	lowerTerm := strings.ToLower(searchTerm)
	
	index := 0
	for i := 0; i < maxExcerpts; i++ {
		pos := strings.Index(lowerContent[index:], lowerTerm)
		if pos == -1 {
			break
		}
		
		pos += index
		start := pos - contextChars
		if start < 0 {
			start = 0
		}
		
		end := pos + len(searchTerm) + contextChars
		if end > len(content) {
			end = len(content)
		}
		
		excerpt := content[start:end]
		if start > 0 {
			excerpt = "..." + excerpt
		}
		if end < len(content) {
			excerpt = excerpt + "..."
		}
		
		excerpts = append(excerpts, excerpt)
		index = pos + len(searchTerm)
	}
	
	return excerpts
}

// FilterByTypeAndArchiveStatus is a utility function that filters items by type and archive status
func FilterByTypeAndArchiveStatus[T Searchable](items []T, itemTypes []string, includeArchived bool) []T {
	var filtered []T
	
	for _, item := range items {
		// Check archived status
		if !includeArchived && item.IsArchived() {
			continue
		}
		
		// Check type filter
		if len(itemTypes) > 0 {
			typeMatch := false
			itemType := strings.ToLower(item.GetType())
			itemSubType := strings.ToLower(item.GetSubType())
			
			for _, filterType := range itemTypes {
				filterType = strings.ToLower(filterType)
				if itemType == filterType || itemSubType == filterType {
					typeMatch = true
					break
				}
				// Handle plural forms and partial matches
				if strings.Contains(filterType, itemType) || strings.Contains(filterType, itemSubType) {
					typeMatch = true
					break
				}
			}
			
			if !typeMatch {
				continue
			}
		}
		
		filtered = append(filtered, item)
	}
	
	return filtered
}