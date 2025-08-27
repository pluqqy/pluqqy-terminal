package shared

import (
	"testing"
	"time"
)

func TestSearchEngine_ContentSearch(t *testing.T) {
	// Create test items with mock content
	testItems := []*TestSearchableItem{
		{
			name:     "Error Handler",
			path:     "/test/error.md",
			itemType: "component",
			subType:  "prompts",
			tags:     []string{"error-handling"},
			content:  "This component handles errors gracefully with proper logging and recovery mechanisms.",
			modified: time.Now(),
		},
		{
			name:     "Authentication Module",
			path:     "/test/auth.md",
			itemType: "component",
			subType:  "contexts",
			tags:     []string{"security"},
			content:  "Provides secure authentication using JWT tokens and OAuth2 protocols.",
			modified: time.Now(),
		},
		{
			name:     "Data Pipeline",
			path:     "/test/pipeline.yaml",
			itemType: "pipeline",
			subType:  "",
			tags:     []string{"etl"},
			content:  "Processes data through multiple transformation stages with error handling.",
			modified: time.Now(),
		},
		{
			name:     "API Documentation",
			path:     "/test/api.md",
			itemType: "component",
			subType:  "contexts",
			tags:     []string{"docs"},
			content:  "Complete REST API documentation with authentication endpoints and data models.",
			modified: time.Now(),
		},
	}
	
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "content search with quotes - exact phrase",
			query:         `content:"error handling"`,
			expectedCount: 1,
			expectedNames: []string{"Data Pipeline"}, // Only this one has exact phrase "error handling"
		},
		{
			name:          "content search for errors",
			query:         `content:errors`,
			expectedCount: 1,
			expectedNames: []string{"Error Handler"}, // Has "errors gracefully"
		},
		{
			name:          "content search without quotes",
			query:         `content:authentication`,
			expectedCount: 2,
			expectedNames: []string{"Authentication Module", "API Documentation"},
		},
		{
			name:          "content search for JWT",
			query:         `content:JWT`,
			expectedCount: 1,
			expectedNames: []string{"Authentication Module"},
		},
		{
			name:          "content search case insensitive",
			query:         `content:"REST API"`,
			expectedCount: 1,
			expectedNames: []string{"API Documentation"},
		},
		{
			name:          "name search",
			query:         `name:Handler`,
			expectedCount: 1,
			expectedNames: []string{"Error Handler"},
		},
		{
			name:          "name search with pipeline",
			query:         `name:Pipeline`,
			expectedCount: 1,
			expectedNames: []string{"Data Pipeline"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create search engine
			engine := NewSearchEngine[*TestSearchableItem]()
			engine.SetItems(testItems)
			engine.SetOptions(SearchOptions{
				IncludeArchived: false,
				MaxResults:      100,
				Mode:           SearchModeAll,
			})
			
			// Perform search
			results, err := engine.Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			
			// Check result count
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
				for _, r := range results {
					t.Logf("  - %s", r.Item.GetName())
				}
			}
			
			// Check expected items are in results
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, result := range results {
					if result.Item.GetName() == expectedName {
						found = true
						// Verify content match flag is set for content searches
						if tt.query[:8] == "content:" && !result.Relevance.ContentMatch {
							t.Errorf("ContentMatch flag not set for item %s", expectedName)
						}
						// Verify name match flag is set for name searches
						if tt.query[:5] == "name:" && !result.Relevance.NameMatch {
							t.Errorf("NameMatch flag not set for item %s", expectedName)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected item %s not found in results", expectedName)
				}
			}
		})
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

func (t *TestSearchableItem) GetName() string       { return t.name }
func (t *TestSearchableItem) GetPath() string       { return t.path }
func (t *TestSearchableItem) GetType() string       { return t.itemType }
func (t *TestSearchableItem) GetSubType() string    { return t.subType }
func (t *TestSearchableItem) GetTags() []string     { return t.tags }
func (t *TestSearchableItem) GetContent() string    { return t.content }
func (t *TestSearchableItem) GetModified() time.Time { return t.modified }
func (t *TestSearchableItem) IsArchived() bool      { return t.archived }
func (t *TestSearchableItem) GetTokenCount() int    { return t.tokenCount }
func (t *TestSearchableItem) GetUsageCount() int    { return t.usageCount }