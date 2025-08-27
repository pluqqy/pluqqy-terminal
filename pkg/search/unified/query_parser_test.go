package unified

import (
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *ParsedQuery
	}{
		{
			name:  "empty query",
			query: "",
			expected: &ParsedQuery{
				Filters:  []QueryFilter{},
				FreeText: "",
			},
		},
		{
			name:  "single tag filter",
			query: "tag:tui",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "tui"},
				},
				FreeText: "",
			},
		},
		{
			name:  "multiple filters",
			query: "tag:tui content:coding",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "tui"},
					{Type: "content", Value: "coding"},
				},
				FreeText: "",
			},
		},
		{
			name:  "filters with free text",
			query: "tag:api search term type:component",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "api"},
					{Type: "type", Value: "component"},
				},
				FreeText: "search term",
			},
		},
		{
			name:  "quoted value in filter",
			query: `name:"Error Handler" tag:error`,
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "name", Value: "Error Handler"},
					{Type: "tag", Value: "error"},
				},
				FreeText: "",
			},
		},
		{
			name:  "mixed case filters",
			query: "TAG:api TYPE:Component",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "api"},
					{Type: "type", Value: "Component"},
				},
				FreeText: "",
			},
		},
		{
			name:  "all filter types",
			query: "tag:test type:prompt status:active name:Handler content:error",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "test"},
					{Type: "type", Value: "prompt"},
					{Type: "status", Value: "active"},
					{Type: "name", Value: "Handler"},
					{Type: "content", Value: "error"},
				},
				FreeText: "",
			},
		},
		{
			name:  "only free text",
			query: "search for this text",
			expected: &ParsedQuery{
				Filters:  []QueryFilter{},
				FreeText: "search for this text",
			},
		},
		{
			name:  "content with quotes",
			query: `content:"JWT tokens"`,
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "content", Value: "JWT tokens"},
				},
				FreeText: "",
			},
		},
		{
			name:  "multiple tags",
			query: "tag:api tag:rest tag:graphql",
			expected: &ParsedQuery{
				Filters: []QueryFilter{
					{Type: "tag", Value: "api"},
					{Type: "tag", Value: "rest"},
					{Type: "tag", Value: "graphql"},
				},
				FreeText: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseQuery(tt.query)
			
			// Check filters
			if len(result.Filters) != len(tt.expected.Filters) {
				t.Errorf("Expected %d filters, got %d", len(tt.expected.Filters), len(result.Filters))
				t.Logf("Filters: %+v", result.Filters)
			}
			
			for i, filter := range result.Filters {
				if i >= len(tt.expected.Filters) {
					break
				}
				if filter.Type != tt.expected.Filters[i].Type || filter.Value != tt.expected.Filters[i].Value {
					t.Errorf("Filter %d: expected %+v, got %+v", i, tt.expected.Filters[i], filter)
				}
			}
			
			// Check free text
			if result.FreeText != tt.expected.FreeText {
				t.Errorf("Expected free text '%s', got '%s'", tt.expected.FreeText, result.FreeText)
			}
		})
	}
}

func TestSplitQueryPreservingQuotes(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "simple split",
			query:    "tag:api type:component",
			expected: []string{"tag:api", "type:component"},
		},
		{
			name:     "quoted string",
			query:    `name:"Error Handler" tag:error`,
			expected: []string{`name:"Error Handler"`, "tag:error"},
		},
		{
			name:     "single quotes",
			query:    `content:'test content' type:prompt`,
			expected: []string{`content:'test content'`, "type:prompt"},
		},
		{
			name:     "multiple spaces",
			query:    "tag:api    type:component",
			expected: []string{"tag:api", "type:component"},
		},
		{
			name:     "quotes with spaces inside",
			query:    `name:"My Component Name" content:"with some content"`,
			expected: []string{`name:"My Component Name"`, `content:"with some content"`},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitQueryPreservingQuotes(tt.query)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParsedQuery_Methods(t *testing.T) {
	query := ParseQuery("tag:api type:component content:test")
	
	// Test HasFilter
	if !query.HasFilter("tag") {
		t.Error("Expected HasFilter('tag') to be true")
	}
	if query.HasFilter("status") {
		t.Error("Expected HasFilter('status') to be false")
	}
	
	// Test GetFilter
	filter, found := query.GetFilter("tag")
	if !found {
		t.Error("Expected to find 'tag' filter")
	}
	if filter.Value != "api" {
		t.Errorf("Expected tag value 'api', got '%s'", filter.Value)
	}
	
	_, found = query.GetFilter("status")
	if found {
		t.Error("Expected not to find 'status' filter")
	}
	
	// Test GetFilters with multiple same-type filters
	multiQuery := ParseQuery("tag:api tag:rest tag:graphql")
	tagFilters := multiQuery.GetFilters("tag")
	if len(tagFilters) != 3 {
		t.Errorf("Expected 3 tag filters, got %d", len(tagFilters))
	}
	
	expectedTags := []string{"api", "rest", "graphql"}
	for i, filter := range tagFilters {
		if filter.Value != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], filter.Value)
		}
	}
}