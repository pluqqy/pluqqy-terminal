package tui

import (
	"regexp"
	"strings"
)

// SearchFilterHelper provides functions to manipulate search query filters
type SearchFilterHelper struct{}

// NewSearchFilterHelper creates a new search filter helper
func NewSearchFilterHelper() *SearchFilterHelper {
	return &SearchFilterHelper{}
}

// ToggleArchivedFilter adds or removes the status:archived filter
func (sfh *SearchFilterHelper) ToggleArchivedFilter(query string) string {
	if strings.Contains(query, "status:archived") {
		// Remove the filter
		return sfh.removeFilter(query, "status:archived")
	}
	// Add the filter
	return sfh.appendFilter(query, "status:archived")
}

// CycleTypeFilter cycles through type filters: all -> pipelines -> prompts -> contexts -> rules -> all
func (sfh *SearchFilterHelper) CycleTypeFilter(query string) string {
	// Extract current type filter if any
	currentType := sfh.extractTypeFilter(query)
	
	// Define the cycle order (empty string means "all"/no filter)
	typeOrder := []string{"", "pipelines", "prompts", "contexts", "rules"}
	
	// Find current index and get next
	currentIndex := 0
	for i, t := range typeOrder {
		if t == currentType {
			currentIndex = i
			break
		}
	}
	
	nextIndex := (currentIndex + 1) % len(typeOrder)
	nextType := typeOrder[nextIndex]
	
	// Remove existing type filter
	query = sfh.removeTypeFilter(query)
	
	// Add new type filter if not cycling back to "all"
	if nextType != "" {
		query = sfh.appendFilter(query, "type:"+nextType)
	}
	
	return strings.TrimSpace(query)
}

// CycleTypeFilterForComponents cycles through type filters for component views (skips pipelines)
// Cycles: all -> prompts -> contexts -> rules -> all
func (sfh *SearchFilterHelper) CycleTypeFilterForComponents(query string) string {
	// Extract current type filter if any
	currentType := sfh.extractTypeFilter(query)
	
	// Define the cycle order without pipelines (empty string means "all"/no filter)
	typeOrder := []string{"", "prompts", "contexts", "rules"}
	
	// Find current index and get next
	currentIndex := 0
	for i, t := range typeOrder {
		if t == currentType {
			currentIndex = i
			break
		}
	}
	
	// If current type is "pipelines" (shouldn't happen in builder but handle gracefully),
	// treat it as if we're at "all" and go to prompts
	if currentType == "pipelines" {
		currentIndex = 0
	}
	
	nextIndex := (currentIndex + 1) % len(typeOrder)
	nextType := typeOrder[nextIndex]
	
	// Remove existing type filter
	query = sfh.removeTypeFilter(query)
	
	// Add new type filter if not cycling back to "all"
	if nextType != "" {
		query = sfh.appendFilter(query, "type:"+nextType)
	}
	
	return strings.TrimSpace(query)
}

// extractTypeFilter finds and returns the current type filter value (or empty string if none)
func (sfh *SearchFilterHelper) extractTypeFilter(query string) string {
	re := regexp.MustCompile(`type:(\w+)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// removeTypeFilter removes any type:xxx filter from the query
func (sfh *SearchFilterHelper) removeTypeFilter(query string) string {
	re := regexp.MustCompile(`\s*type:\w+\s*`)
	result := re.ReplaceAllString(query, " ")
	return sfh.cleanupSpaces(result)
}

// removeFilter removes a specific filter from the query
func (sfh *SearchFilterHelper) removeFilter(query, filter string) string {
	result := strings.ReplaceAll(query, filter, "")
	return sfh.cleanupSpaces(result)
}

// appendFilter adds a filter to the query with proper spacing
func (sfh *SearchFilterHelper) appendFilter(query, filter string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return filter
	}
	return query + " " + filter
}

// cleanupSpaces removes extra spaces and trims the result
func (sfh *SearchFilterHelper) cleanupSpaces(s string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// GetCurrentFilters returns a list of active filters for display purposes
func (sfh *SearchFilterHelper) GetCurrentFilters(query string) []string {
	var filters []string
	
	if strings.Contains(query, "status:archived") {
		filters = append(filters, "Archived")
	}
	
	typeFilter := sfh.extractTypeFilter(query)
	if typeFilter != "" {
		filters = append(filters, "Type: "+typeFilter)
	}
	
	return filters
}