package unified

import (
	"strings"
	"testing"
	"time"
)

func TestSearchEngine_Search(t *testing.T) {
	// Create test items
	testItems := createTestItems()
	
	tests := []struct {
		name          string
		query         string
		options       SearchOptions
		expectedCount int
		expectedNames []string
		checkRelevance func(t *testing.T, results []SearchResult[*TestSearchableItem])
	}{
		// Basic searches
		{
			name:          "empty query returns all non-archived items",
			query:         "",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 3, // Only non-archived items
			expectedNames: []string{"Error Handler", "Authentication Module", "Data Pipeline"},
		},
		{
			name:          "empty query with archived",
			query:         "",
			options:       SearchOptions{IncludeArchived: true},
			expectedCount: 4, // All items including archived
			expectedNames: []string{"Error Handler", "Authentication Module", "Data Pipeline", "Archived Component"},
		},
		
		// Tag searches
		{
			name:          "tag search - exact match",
			query:         "tag:security",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
			checkRelevance: func(t *testing.T, results []SearchResult[*TestSearchableItem]) {
				if len(results) > 0 && !results[0].Relevance.TagMatch {
					t.Error("Expected TagMatch to be true")
				}
			},
		},
		{
			name:          "tag search - multiple results",
			query:         "tag:error",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 2,
			expectedNames: []string{"Error Handler", "Data Pipeline"},
		},
		
		// Type searches
		{
			name:          "type search - singular",
			query:         "type:component",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 2,
			expectedNames: []string{"Error Handler", "Authentication Module"},
		},
		{
			name:          "type search - plural",
			query:         "type:pipelines",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Data Pipeline"},
		},
		{
			name:          "type search - with subtype plural",
			query:         "type:prompts",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Error Handler"},
		},
		
		// Content searches
		{
			name:          "content search - single word",
			query:         "content:error",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 2,
			expectedNames: []string{"Error Handler", "Data Pipeline"},
			checkRelevance: func(t *testing.T, results []SearchResult[*TestSearchableItem]) {
				for _, r := range results {
					if !r.Relevance.ContentMatch {
						t.Errorf("Expected ContentMatch to be true for %s", r.Item.GetName())
					}
				}
			},
		},
		{
			name:          "content search - with quotes",
			query:         `content:"JWT tokens"`,
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
		},
		{
			name:          "content search - case insensitive",
			query:         `content:AUTHENTICATION`,
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
		},
		
		// Name searches
		{
			name:          "name search - partial match",
			query:         "name:Handler",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Error Handler"},
			checkRelevance: func(t *testing.T, results []SearchResult[*TestSearchableItem]) {
				if len(results) > 0 && !results[0].Relevance.NameMatch {
					t.Error("Expected NameMatch to be true")
				}
			},
		},
		{
			name:          "name search - case insensitive",
			query:         "name:pipeline",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedNames: []string{"Data Pipeline"},
		},
		
		// Status searches
		{
			name:          "status search - archived",
			query:         "status:archived",
			options:       SearchOptions{IncludeArchived: true},
			expectedCount: 1,
			expectedNames: []string{"Archived Component"},
		},
		{
			name:          "status search - active",
			query:         "status:active",
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 3,
			expectedNames: []string{"Error Handler", "Authentication Module", "Data Pipeline"},
		},
		
		// Simple text search (non-structured)
		{
			name:          "simple text search",
			query:         "error",
			options:       SearchOptions{IncludeArchived: false, Mode: SearchModeAll},
			expectedCount: 2,
			expectedNames: []string{"Error Handler", "Data Pipeline"},
		},
		
		// Search modes
		{
			name:          "search mode - name only",
			query:         "error",
			options:       SearchOptions{IncludeArchived: false, Mode: SearchModeName},
			expectedCount: 1,
			expectedNames: []string{"Error Handler"},
		},
		{
			name:          "search mode - content only",
			query:         "authentication",
			options:       SearchOptions{IncludeArchived: false, Mode: SearchModeContent},
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
		},
		{
			name:          "search mode - tags only",
			query:         "security",
			options:       SearchOptions{IncludeArchived: false, Mode: SearchModeTags},
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create search engine
			engine := NewSearchEngine[*TestSearchableItem]()
			engine.SetItems(testItems)
			engine.SetOptions(tt.options)
			
			// Perform search
			results, err := engine.Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			
			// Check result count
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
					t.Errorf("Expected item %s not found in results", expectedName)
				}
			}
			
			// Run additional relevance checks if provided
			if tt.checkRelevance != nil {
				tt.checkRelevance(t, results)
			}
		})
	}
}

func TestSearchEngine_Sorting(t *testing.T) {
	// Create items with different scores
	items := []*TestSearchableItem{
		{name: "Low Score", content: "test", modified: time.Now().Add(-48 * time.Hour)},
		{name: "High Score Test", content: "test test test", modified: time.Now()},
		{name: "Medium Score", content: "test test", modified: time.Now().Add(-24 * time.Hour)},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	engine.SetOptions(SearchOptions{SortBy: "relevance"})
	
	results, err := engine.Search("test")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Check results are sorted by score (descending)
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	
	// High Score Test should be first (most matches + recent)
	if results[0].Item.GetName() != "High Score Test" {
		t.Errorf("Expected 'High Score Test' first, got %s", results[0].Item.GetName())
	}
	
	// Verify scores are descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not sorted by score: %.2f > %.2f", results[i].Score, results[i-1].Score)
		}
	}
}

func TestSearchEngine_Deduplication(t *testing.T) {
	// Create items with content that would match multiple times
	items := []*TestSearchableItem{
		{
			name:    "Test Component",
			content: "error handling error recovery error logging",
			tags:    []string{"error", "handling"},
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	
	// Search for "error" which appears multiple times
	results, err := engine.Search("content:error")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Should only return the item once, not multiple times
	if len(results) != 1 {
		t.Errorf("Expected 1 result (deduplicated), got %d", len(results))
	}
}

func TestSearchEngine_ExtractExcerpts(t *testing.T) {
	content := "This is a test document with multiple occurrences of the word test in various test positions."
	
	excerpts := extractExcerpts(content, "test", 2, 30)
	
	if len(excerpts) == 0 {
		t.Error("Expected excerpts to be extracted")
	}
	
	// Check that excerpts contain the search term
	for _, excerpt := range excerpts {
		if !strings.Contains(strings.ToLower(excerpt), "test") {
			t.Errorf("Excerpt doesn't contain search term: %s", excerpt)
		}
	}
	
	// Check excerpt length constraint
	for _, excerpt := range excerpts {
		if len(excerpt) > 100 { // Some buffer for "..." additions
			t.Errorf("Excerpt too long: %d characters", len(excerpt))
		}
	}
}

// Test helper functions

func createTestItems() []*TestSearchableItem {
	return []*TestSearchableItem{
		{
			name:     "Error Handler",
			path:     "/test/error.md",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"error", "handling"},
			content:  "This component handles errors gracefully with proper logging.",
			modified: time.Now(),
			archived: false,
		},
		{
			name:     "Authentication Module",
			path:     "/test/auth.md",
			itemType: "component",
			subType:  "contexts",
			tags:     []string{"security", "auth"},
			content:  "Provides secure authentication using JWT tokens and OAuth2.",
			modified: time.Now().Add(-24 * time.Hour),
			archived: false,
		},
		{
			name:     "Data Pipeline",
			path:     "/test/pipeline.yaml",
			itemType: "pipeline",
			subType:  "",
			tags:     []string{"etl", "data", "error"},
			content:  "Processes data through transformation stages with error handling.",
			modified: time.Now().Add(-48 * time.Hour),
			archived: false,
		},
		{
			name:     "Archived Component",
			path:     "/test/old.md",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"deprecated"},
			content:  "Old component that has been archived.",
			modified: time.Now().Add(-72 * time.Hour),
			archived: true,
		},
	}
}

// TestSearchableItem implements Searchable interface for testing
type TestSearchableItem struct {
	name       string
	path       string
	itemType   string
	subType    string
	tags       []string
	content    string
	modified   time.Time
	archived   bool
	tokenCount int
	usageCount int
}

func (t *TestSearchableItem) GetName() string        { return t.name }
func (t *TestSearchableItem) GetPath() string        { return t.path }
func (t *TestSearchableItem) GetType() string        { return t.itemType }
func (t *TestSearchableItem) GetSubType() string     { return t.subType }
func (t *TestSearchableItem) GetTags() []string      { return t.tags }
func (t *TestSearchableItem) GetContent() string     { return t.content }
func (t *TestSearchableItem) GetModified() time.Time { return t.modified }
func (t *TestSearchableItem) IsArchived() bool       { return t.archived }
func (t *TestSearchableItem) GetTokenCount() int     { return t.tokenCount }
func (t *TestSearchableItem) GetUsageCount() int     { return t.usageCount }