package unified

import (
	"strings"
)

// QueryFilter represents a single filter in a search query
type QueryFilter struct {
	Type  string // "tag", "type", "status", "name", "content"
	Value string
}

// ParsedQuery represents a parsed search query with multiple filters
type ParsedQuery struct {
	Filters []QueryFilter
	FreeText string // Any remaining text that isn't part of a filter
}

// ParseQuery parses a search query into structured filters
// Example: "tag:tui content:coding" returns two filters
func ParseQuery(query string) *ParsedQuery {
	result := &ParsedQuery{
		Filters: []QueryFilter{},
	}
	
	// Track what parts of the query have been processed
	processedParts := make(map[string]bool)
	
	// Define the filter prefixes we support
	filterPrefixes := []string{"tag:", "type:", "status:", "name:", "content:", "modified:"}
	
	// Split query into parts but preserve quoted strings
	parts := splitQueryPreservingQuotes(query)
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		found := false
		lowerPart := strings.ToLower(part)
		
		// Check if this part matches any filter prefix
		for _, prefix := range filterPrefixes {
			if strings.HasPrefix(lowerPart, prefix) {
				filterType := strings.TrimSuffix(prefix, ":")
				filterValue := part[len(prefix):]
				filterValue = strings.TrimSpace(filterValue)
				
				// Remove quotes if present
				filterValue = strings.Trim(filterValue, `"'`)
				
				result.Filters = append(result.Filters, QueryFilter{
					Type:  filterType,
					Value: filterValue,
				})
				
				processedParts[part] = true
				found = true
				break
			}
		}
		
		// If not a filter, add to free text
		if !found && !processedParts[part] {
			if result.FreeText != "" {
				result.FreeText += " "
			}
			result.FreeText += part
		}
	}
	
	return result
}

// splitQueryPreservingQuotes splits a query string into parts while preserving quoted strings
func splitQueryPreservingQuotes(query string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	
	for _, r := range query {
		switch {
		case !inQuotes && (r == '"' || r == '\''):
			// Start of quoted section
			inQuotes = true
			quoteChar = r
			current.WriteRune(r)
		case inQuotes && r == quoteChar:
			// End of quoted section
			inQuotes = false
			current.WriteRune(r)
		case !inQuotes && r == ' ':
			// Space outside quotes - split here
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	
	// Add any remaining text
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
}

// HasFilter checks if the parsed query has a specific filter type
func (pq *ParsedQuery) HasFilter(filterType string) bool {
	for _, filter := range pq.Filters {
		if filter.Type == filterType {
			return true
		}
	}
	return false
}

// GetFilter returns the first filter of the specified type
func (pq *ParsedQuery) GetFilter(filterType string) (QueryFilter, bool) {
	for _, filter := range pq.Filters {
		if filter.Type == filterType {
			return filter, true
		}
	}
	return QueryFilter{}, false
}

// GetFilters returns all filters of the specified type
func (pq *ParsedQuery) GetFilters(filterType string) []QueryFilter {
	var filters []QueryFilter
	for _, filter := range pq.Filters {
		if filter.Type == filterType {
			filters = append(filters, filter)
		}
	}
	return filters
}