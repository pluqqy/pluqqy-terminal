package unified

import (
	"testing"
	"time"
)

func TestSearchEngine_MultipleFilters(t *testing.T) {
	// Create test items with various combinations of attributes
	items := []*TestSearchableItem{
		{
			name:     "TUI Error Handler",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"tui", "error", "handling"},
			content:  "This component handles errors in the TUI with proper coding standards",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "API Error Handler",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"api", "error", "rest"},
			content:  "Handles API errors with proper logging",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "TUI Authentication",
			itemType: "component",
			subType:  "contexts",
			tags:     []string{"tui", "auth", "security"},
			content:  "Authentication context for TUI components",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Coding Standards",
			itemType: "component",
			subType:  "rules",
			tags:     []string{"standards", "quality", "coding"},
			content:  "Rules for maintaining code quality and coding standards",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Old TUI Component",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"tui", "deprecated"},
			content:  "Old TUI component that is archived",
			archived: true,
			modified: time.Now().Add(-30 * 24 * time.Hour),
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	engine.SetOptions(SearchOptions{IncludeArchived: false})
	
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "single tag filter",
			query:         "tag:tui",
			expectedCount: 2,
			expectedNames: []string{"TUI Error Handler", "TUI Authentication"},
		},
		{
			name:          "multiple filters AND - tag and content",
			query:         "tag:tui content:coding",
			expectedCount: 1,
			expectedNames: []string{"TUI Error Handler"},
		},
		{
			name:          "multiple filters AND - tag and type",
			query:         "tag:tui type:prompts",
			expectedCount: 1,
			expectedNames: []string{"TUI Error Handler"},
		},
		{
			name:          "multiple filters with no results",
			query:         "tag:api content:coding",
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "three filters",
			query:         "tag:tui type:component content:error",
			expectedCount: 1,
			expectedNames: []string{"TUI Error Handler"},
		},
		{
			name:          "filter with free text",
			query:         "tag:tui Handler",
			expectedCount: 1,
			expectedNames: []string{"TUI Error Handler"},
		},
		{
			name:          "name filter with tag",
			query:         "name:Handler tag:error",
			expectedCount: 2,
			expectedNames: []string{"TUI Error Handler", "API Error Handler"},
		},
		{
			name:          "status archived with tag",
			query:         "status:archived tag:tui",
			expectedCount: 1,
			expectedNames: []string{"Old TUI Component"},
		},
		{
			name:          "content filter with special term",
			query:         "content:standards tag:coding",
			expectedCount: 1,
			expectedNames: []string{"Coding Standards"},
		},
		{
			name:          "quoted content with tag",
			query:         `content:"code quality" tag:standards`,
			expectedCount: 1,
			expectedNames: []string{"Coding Standards"},
		},
		{
			name:          "multiple tags (any match)",
			query:         "tag:error tag:handling",
			expectedCount: 1,
			expectedNames: []string{"TUI Error Handler"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special handling for archived test
			if tt.query == "status:archived tag:tui" {
				engine.SetOptions(SearchOptions{IncludeArchived: true})
			} else {
				engine.SetOptions(SearchOptions{IncludeArchived: false})
			}
			
			results, err := engine.Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
				for _, r := range results {
					t.Logf("  - %s (score: %.2f)", r.Item.GetName(), r.Score)
				}
			}
			
			// Check expected items are in results
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, result := range results {
					if result.Item.GetName() == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected item '%s' not found in results", expectedName)
				}
			}
		})
	}
}

func TestSearchEngine_ComplexMultiFilter(t *testing.T) {
	items := []*TestSearchableItem{
		{
			name:     "Production API Handler",
			itemType: "component",
			subType:  "handlers",
			tags:     []string{"production", "api", "stable"},
			content:  "Handles production API requests with error recovery",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Development API Handler",
			itemType: "component", 
			subType:  "handlers",
			tags:     []string{"development", "api", "testing"},
			content:  "Development version with debug logging",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Production Database Handler",
			itemType: "component",
			subType:  "handlers",
			tags:     []string{"production", "database", "stable"},
			content:  "Database connection handler for production",
			archived: false,
			modified: time.Now(),
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	
	// Test: Find production handlers that deal with API
	results, err := engine.Search("tag:production type:handlers content:API")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	
	if len(results) > 0 && results[0].Item.GetName() != "Production API Handler" {
		t.Errorf("Expected 'Production API Handler', got '%s'", results[0].Item.GetName())
	}
	
	// Verify relevance information
	if len(results) > 0 {
		relevance := results[0].Relevance
		if !relevance.TagMatch {
			t.Error("Expected TagMatch to be true")
		}
		if !relevance.ContentMatch {
			t.Error("Expected ContentMatch to be true")
		}
		
		// Check highlights
		if len(relevance.Highlights["tags"]) == 0 {
			t.Error("Expected tag highlights")
		}
		if len(relevance.Highlights["content"]) == 0 {
			t.Error("Expected content highlights")
		}
	}
}