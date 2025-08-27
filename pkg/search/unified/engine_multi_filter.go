package unified

import (
	"strings"
)

// searchWithMultipleFilters handles queries with multiple filters using AND logic
func (se *SearchEngine[T]) searchWithMultipleFilters(parsedQuery *ParsedQuery) []SearchResult[T] {
	var results []SearchResult[T]
	
	// Process each item and check if it matches ALL filters
	for _, item := range se.items {
		// Skip archived items if not included (unless status:archived is specified)
		if !se.shouldIncludeItem(item, parsedQuery) {
			continue
		}
		
		// Check if item matches ALL filters
		matches, score, relevance := se.checkAllFilters(item, parsedQuery)
		if !matches {
			continue
		}
		
		// If there's free text, also check that
		if parsedQuery.FreeText != "" {
			textScore, textRelevance := se.calculateSimpleScore(item, strings.Fields(parsedQuery.FreeText))
			if textScore == 0 {
				continue // Doesn't match free text
			}
			// Combine scores and relevance
			score += textScore
			relevance = se.combineRelevance(relevance, textRelevance)
		}
		
		results = append(results, SearchResult[T]{
			Item:      item,
			Score:     score,
			Relevance: relevance,
		})
	}
	
	// Sort results by score
	se.sortResults(results)
	
	// Apply max results limit
	if se.options.MaxResults > 0 && len(results) > se.options.MaxResults {
		results = results[:se.options.MaxResults]
	}
	
	return results
}

// shouldIncludeItem checks if an item should be included based on archived status
func (se *SearchEngine[T]) shouldIncludeItem(item T, parsedQuery *ParsedQuery) bool {
	// Check if there's a status filter
	if statusFilter, hasStatus := parsedQuery.GetFilter("status"); hasStatus {
		isArchived := item.IsArchived()
		if statusFilter.Value == "archived" {
			return isArchived
		} else if statusFilter.Value == "active" {
			return !isArchived
		}
	}
	
	// Default behavior - exclude archived unless explicitly included
	if !se.options.IncludeArchived && item.IsArchived() {
		return false
	}
	
	return true
}

// checkAllFilters checks if an item matches ALL filters in the query
func (se *SearchEngine[T]) checkAllFilters(item T, parsedQuery *ParsedQuery) (bool, float64, SearchRelevance) {
	totalScore := 0.0
	relevance := SearchRelevance{
		Highlights: make(map[string][]string),
	}
	
	for _, filter := range parsedQuery.Filters {
		matches, filterScore, filterRelevance := se.checkSingleFilter(item, filter)
		if !matches {
			return false, 0, SearchRelevance{} // Item doesn't match this filter
		}
		
		totalScore += filterScore
		relevance = se.combineRelevance(relevance, filterRelevance)
	}
	
	return true, totalScore, relevance
}

// checkSingleFilter checks if an item matches a single filter
func (se *SearchEngine[T]) checkSingleFilter(item T, filter QueryFilter) (bool, float64, SearchRelevance) {
	relevance := SearchRelevance{
		Highlights: make(map[string][]string),
	}
	
	switch filter.Type {
	case "tag":
		// Check if item has the requested tag
		for _, tag := range item.GetTags() {
			if strings.EqualFold(tag, filter.Value) {
				relevance.TagMatch = true
				relevance.ExactMatch = true
				relevance.Highlights["tags"] = []string{filter.Value}
				return true, 10.0, relevance
			}
		}
		return false, 0, relevance
		
	case "type":
		// Check type and subtype
		itemType := strings.ToLower(item.GetType())
		itemSubType := strings.ToLower(item.GetSubType())
		typeQuery := strings.ToLower(filter.Value)
		
		// Check exact match
		if itemType == typeQuery || itemSubType == typeQuery {
			relevance.ExactMatch = true
			relevance.Highlights["type"] = []string{filter.Value}
			return true, 10.0, relevance
		}
		
		// Check plural/singular variations
		if itemType+"s" == typeQuery || itemSubType+"s" == typeQuery {
			relevance.Highlights["type"] = []string{filter.Value}
			return true, 10.0, relevance
		}
		if strings.TrimSuffix(itemType, "s") == typeQuery || strings.TrimSuffix(itemSubType, "s") == typeQuery {
			relevance.Highlights["type"] = []string{filter.Value}
			return true, 10.0, relevance
		}
		
		return false, 0, relevance
		
	case "status":
		// Check archived status
		isArchived := item.IsArchived()
		statusQuery := strings.ToLower(filter.Value)
		
		if (statusQuery == "archived" && isArchived) || (statusQuery == "active" && !isArchived) {
			relevance.ExactMatch = true
			relevance.Highlights["status"] = []string{filter.Value}
			return true, 10.0, relevance
		}
		return false, 0, relevance
		
	case "name":
		// Check if name contains the search term
		name := strings.ToLower(item.GetName())
		nameQuery := strings.ToLower(filter.Value)
		
		if strings.Contains(name, nameQuery) {
			score := 5.0
			exactMatch := false
			if name == nameQuery {
				score = 10.0
				exactMatch = true
			}
			
			relevance.NameMatch = true
			relevance.ExactMatch = exactMatch
			relevance.Highlights["name"] = []string{item.GetName()}
			return true, score, relevance
		}
		return false, 0, relevance
		
	case "content":
		// Check if content contains the search term
		content := strings.ToLower(item.GetContent())
		contentQuery := strings.ToLower(filter.Value)
		
		if strings.Contains(content, contentQuery) {
			// Count occurrences
			count := strings.Count(content, contentQuery)
			score := float64(count) * 2.0
			
			relevance.ContentMatch = true
			if excerpts := extractExcerpts(item.GetContent(), filter.Value, 2, 50); len(excerpts) > 0 {
				relevance.Highlights["content"] = excerpts
			}
			return true, score, relevance
		}
		return false, 0, relevance
		
	default:
		// Unknown filter type - ignore for now
		return true, 0, relevance
	}
}

// combineRelevance combines two SearchRelevance objects
func (se *SearchEngine[T]) combineRelevance(r1, r2 SearchRelevance) SearchRelevance {
	combined := SearchRelevance{
		NameMatch:    r1.NameMatch || r2.NameMatch,
		ContentMatch: r1.ContentMatch || r2.ContentMatch,
		TagMatch:     r1.TagMatch || r2.TagMatch,
		ExactMatch:   r1.ExactMatch || r2.ExactMatch,
		Highlights:   make(map[string][]string),
	}
	
	// Merge highlights
	for key, values := range r1.Highlights {
		combined.Highlights[key] = append(combined.Highlights[key], values...)
	}
	for key, values := range r2.Highlights {
		if _, exists := combined.Highlights[key]; !exists {
			combined.Highlights[key] = values
		} else {
			// Avoid duplicates
			for _, v := range values {
				found := false
				for _, existing := range combined.Highlights[key] {
					if existing == v {
						found = true
						break
					}
				}
				if !found {
					combined.Highlights[key] = append(combined.Highlights[key], v)
				}
			}
		}
	}
	
	return combined
}