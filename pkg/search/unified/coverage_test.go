package unified

import (
	"strings"
	"testing"
	"time"

	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestUncoveredFunctions tests functions that weren't covered in other tests
func TestComponentItemWrapper_GetContent(t *testing.T) {
	wrapper := NewComponentItemWrapper(
		"Test Component",
		"/path/to/component",
		models.ComponentTypePrompt,
		time.Now(),
		0,
		0,
		[]string{},
		false,
		"This is test content",
	)
	
	content := wrapper.GetContent()
	expected := "Test Component This is test content"
	if content != expected {
		t.Errorf("Expected '%s', got %s", expected, content)
	}
}

func TestComponentItemWrapper_GetModified(t *testing.T) {
	testTime := time.Now()
	wrapper := NewComponentItemWrapper(
		"Test Component",
		"/path/to/component",
		models.ComponentTypePrompt,
		testTime,
		0,
		0,
		[]string{},
		false,
		"",
	)
	
	modified := wrapper.GetModified()
	if !modified.Equal(testTime) {
		t.Errorf("Expected %v, got %v", testTime, modified)
	}
}

func TestComponentItemWrapper_GetTokenCount(t *testing.T) {
	wrapper := NewComponentItemWrapper(
		"Test Component",
		"/path/to/component",
		models.ComponentTypePrompt,
		time.Now(),
		0,
		100,
		[]string{},
		false,
		"",
	)
	
	count := wrapper.GetTokenCount()
	if count != 100 {
		t.Errorf("Expected 100, got %d", count)
	}
}

func TestComponentItemWrapper_GetUsageCount(t *testing.T) {
	wrapper := NewComponentItemWrapper(
		"Test Component",
		"/path/to/component",
		models.ComponentTypePrompt,
		time.Now(),
		5,
		0,
		[]string{},
		false,
		"",
	)
	
	count := wrapper.GetUsageCount()
	if count != 5 {
		t.Errorf("Expected 5, got %d", count)
	}
}

func TestPipelineItemWrapper_GetContent(t *testing.T) {
	wrapper := NewPipelineItemWrapper(
		"Test Pipeline",
		"/path/to/pipeline",
		[]string{},
		0,
		false,
		time.Now(),
		"Pipeline content here",
	)
	
	content := wrapper.GetContent()
	expected := "Test Pipeline Pipeline content here"
	if content != expected {
		t.Errorf("Expected '%s', got %s", expected, content)
	}
}

func TestPipelineItemWrapper_GetModified(t *testing.T) {
	testTime := time.Now()
	wrapper := NewPipelineItemWrapper(
		"Test Pipeline",
		"/path/to/pipeline",
		[]string{},
		0,
		false,
		testTime,
		"",
	)
	
	modified := wrapper.GetModified()
	if !modified.Equal(testTime) {
		t.Errorf("Expected %v, got %v", testTime, modified)
	}
}

func TestPipelineItemWrapper_GetTokenCount(t *testing.T) {
	wrapper := NewPipelineItemWrapper(
		"Test Pipeline",
		"/path/to/pipeline",
		[]string{},
		200,
		false,
		time.Now(),
		"",
	)
	
	count := wrapper.GetTokenCount()
	if count != 200 {
		t.Errorf("Expected 200, got %d", count)
	}
}

func TestPipelineItemWrapper_GetUsageCount(t *testing.T) {
	// PipelineItemWrapper doesn't track usage count, always returns 0
	wrapper := NewPipelineItemWrapper(
		"Test Pipeline",
		"/path/to/pipeline",
		[]string{},
		0,
		false,
		time.Now(),
		"",
	)
	
	count := wrapper.GetUsageCount()
	if count != 0 {
		t.Errorf("Expected 0 (pipelines don't track usage), got %d", count)
	}
}

func TestGetSearchSuggestions(t *testing.T) {
	manager := NewUnifiedSearchManager()
	
	// Add some test items
	components := []ComponentItem{
		{Name: "Test Component", Tags: []string{"test", "component"}},
		{Name: "API Handler", Tags: []string{"api", "test"}},
	}
	pipelines := []PipelineItem{
		{Name: "Test Pipeline", Tags: []string{"pipeline", "test"}},
	}
	
	manager.LoadComponentItems(components, nil, nil)
	manager.LoadPipelineItems(pipelines)
	
	suggestions := manager.GetSearchSuggestions("te")
	
	// Should get tag suggestions that start with "te"
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "test") {
			found = true
			break
		}
	}
	if !found && len(suggestions) > 0 {
		t.Error("Expected to find 'test' in suggestions")
	}
}

func TestSearchEngine_ContentSearch(t *testing.T) {
	items := []*TestSearchableItem{
		{
			name:    "Document One",
			content: "This document contains important information about testing.",
			tags:    []string{"docs"},
			modified: time.Now(),
		},
		{
			name:    "Code Sample",
			content: "func main() { fmt.Println(\"Hello, World!\") }",
			tags:    []string{"code"},
			modified: time.Now(),
		},
		{
			name:    "README",
			content: "This is a readme file with installation instructions and usage examples.",
			tags:    []string{"documentation"},
			modified: time.Now(),
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	engine.SetOptions(SearchOptions{Mode: SearchModeContent})
	
	// Test content-only search
	results, err := engine.Search("installation")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'installation', got %d", len(results))
	}
	
	if len(results) > 0 && results[0].Item.GetName() != "README" {
		t.Errorf("Expected 'README' to match, got %s", results[0].Item.GetName())
	}
}

func TestSearchEngine_UsageBoost(t *testing.T) {
	items := []*TestSearchableItem{
		{
			name:       "Popular Component",
			content:    "test",
			usageCount: 10,
			modified:   time.Now().Add(-48 * time.Hour),
		},
		{
			name:       "Unpopular Component",
			content:    "test",
			usageCount: 0,
			modified:   time.Now().Add(-48 * time.Hour),
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	
	results, err := engine.Search("test")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	
	// Popular component should rank higher due to usage boost
	if results[0].Item.GetName() != "Popular Component" {
		t.Errorf("Expected 'Popular Component' to rank first, got %s", results[0].Item.GetName())
	}
}

func TestSearchEngine_RecencyBoost(t *testing.T) {
	items := []*TestSearchableItem{
		{
			name:     "Old Item",
			content:  "test content",
			modified: time.Now().Add(-30 * 24 * time.Hour), // 30 days old
		},
		{
			name:     "Recent Item",
			content:  "test content",
			modified: time.Now().Add(-2 * time.Hour), // 2 hours old
		},
		{
			name:     "Week Old Item",
			content:  "test content",
			modified: time.Now().Add(-5 * 24 * time.Hour), // 5 days old
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	
	results, err := engine.Search("test")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	
	// Recent item should rank highest
	if results[0].Item.GetName() != "Recent Item" {
		t.Errorf("Expected 'Recent Item' to rank first, got %s", results[0].Item.GetName())
	}
	
	// Week old should rank second
	if results[1].Item.GetName() != "Week Old Item" {
		t.Errorf("Expected 'Week Old Item' to rank second, got %s", results[1].Item.GetName())
	}
}

func TestSearchEngine_ComplexQueries(t *testing.T) {
	items := []*TestSearchableItem{
		{
			name:     "API Handler",
			itemType: "component",
			subType:  "handlers",
			tags:     []string{"api", "rest"},
			content:  "Handles API requests",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Database Layer",
			itemType: "component", 
			subType:  "database",
			tags:     []string{"database", "sql"},
			content:  "Database access layer",
			archived: false,
			modified: time.Now(),
		},
		{
			name:     "Old API Handler",
			itemType: "component",
			subType:  "handlers",
			tags:     []string{"api", "deprecated"},
			content:  "Old API handler",
			archived: true,
			modified: time.Now().Add(-30 * 24 * time.Hour),
		},
	}
	
	engine := NewSearchEngine[*TestSearchableItem]()
	engine.SetItems(items)
	
	tests := []struct {
		name          string
		query         string
		options       SearchOptions
		expectedCount int
		expectedFirst string
	}{
		{
			name:          "search with quotes",
			query:         `content:"API requests"`,
			options:       SearchOptions{IncludeArchived: false},
			expectedCount: 1,
			expectedFirst: "API Handler",
		},
		{
			name:          "name search with spaces",
			query:         `name:"API Handler"`,
			options:       SearchOptions{IncludeArchived: true},
			expectedCount: 2,
			expectedFirst: "API Handler",
		},
		{
			name:          "subtype search",
			query:         `type:handlers`,
			options:       SearchOptions{IncludeArchived: true},
			expectedCount: 2,
			expectedFirst: "API Handler",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine.SetOptions(tt.options)
			results, err := engine.Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
			
			if len(results) > 0 && results[0].Item.GetName() != tt.expectedFirst {
				t.Errorf("Expected first result to be %s, got %s", tt.expectedFirst, results[0].Item.GetName())
			}
		})
	}
}

func TestShouldIncludeArchived_EdgeCases(t *testing.T) {
	// Test with status: in the middle of query
	query := "tag:test status:archived type:component"
	if !ShouldIncludeArchived(query) {
		t.Error("Should include archived when status:archived is in query")
	}
	
	// Test with status:active
	query2 := "status:active"
	if ShouldIncludeArchived(query2) {
		t.Error("Should not include archived when status:active")
	}
}