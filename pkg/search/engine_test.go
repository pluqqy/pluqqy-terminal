package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pluqqy-search-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Save original PluqqyDir
	originalDir := files.PluqqyDir
	files.PluqqyDir = tmpDir
	
	// Initialize project structure
	if err := files.InitProjectStructure(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init project structure: %v", err)
	}
	
	// Create test data
	createTestData(t)
	
	// Return cleanup function
	return tmpDir, func() {
		files.PluqqyDir = originalDir
		os.RemoveAll(tmpDir)
	}
}

func createTestData(t *testing.T) {
	// Create test components
	testComponents := []struct {
		path    string
		content string
		tags    []string
	}{
		{
			path:    filepath.Join(files.ComponentsDir, files.PromptsDir, "api-prompt.md"),
			content: "# API Prompt\n\nThis is an API-related prompt for error handling.",
			tags:    []string{"api", "error-handling", "v2"},
		},
		{
			path:    filepath.Join(files.ComponentsDir, files.PromptsDir, "auth-prompt.md"),
			content: "# Authentication Prompt\n\nHandle user authentication and authorization.",
			tags:    []string{"auth", "security", "api"},
		},
		{
			path:    filepath.Join(files.ComponentsDir, files.ContextsDir, "api-context.md"),
			content: "# API Context\n\nBackground information about the API system.",
			tags:    []string{"api", "documentation"},
		},
		{
			path:    filepath.Join(files.ComponentsDir, files.RulesDir, "security-rules.md"),
			content: "# Security Rules\n\nImportant security constraints and guidelines.",
			tags:    []string{"security", "critical"},
		},
	}
	
	for _, tc := range testComponents {
		err := files.WriteComponentWithTags(tc.path, tc.content, tc.tags)
		if err != nil {
			t.Fatalf("Failed to create test component %s: %v", tc.path, err)
		}
	}
	
	// Create test pipelines
	testPipelines := []struct {
		pipeline *models.Pipeline
	}{
		{
			pipeline: &models.Pipeline{
				Name: "api-pipeline",
				Tags: []string{"api", "production"},
				Components: []models.ComponentRef{
					{Type: files.PromptsDir, Path: "api-prompt.md", Order: 1},
					{Type: files.ContextsDir, Path: "api-context.md", Order: 2},
				},
			},
		},
		{
			pipeline: &models.Pipeline{
				Name: "auth-pipeline",
				Tags: []string{"auth", "security", "production"},
				Components: []models.ComponentRef{
					{Type: files.PromptsDir, Path: "auth-prompt.md", Order: 1},
					{Type: files.RulesDir, Path: "security-rules.md", Order: 2},
				},
			},
		},
	}
	
	for _, tp := range testPipelines {
		err := files.WritePipeline(tp.pipeline)
		if err != nil {
			t.Fatalf("Failed to create test pipeline %s: %v", tp.pipeline.Name, err)
		}
	}
}

func TestEngineBuildIndex(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	engine := NewEngine()
	
	err := engine.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}
	
	// Check that items were indexed
	if len(engine.index.items) == 0 {
		t.Error("No items were indexed")
	}
	
	// Check tag index
	if len(engine.index.tagIndex) == 0 {
		t.Error("No tags were indexed")
	}
	
	// Check specific tag
	if indices, exists := engine.index.tagIndex["api"]; !exists {
		t.Error("Tag 'api' not found in index")
	} else if len(indices) < 2 {
		t.Errorf("Expected at least 2 items with 'api' tag, got %d", len(indices))
	}
}

func TestEngineSearch(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	engine := NewEngine()
	engine.BuildIndex()
	
	tests := []struct {
		name          string
		query         string
		expectError   bool
		minResults    int
		checkResults  func(*testing.T, []SearchResult)
	}{
		{
			name:       "search by tag",
			query:      "tag:api",
			minResults: 3, // 2 components + 1 pipeline
			checkResults: func(t *testing.T, results []SearchResult) {
				for _, r := range results {
					hasAPI := false
					for _, tag := range r.Item.Tags {
						if models.NormalizeTagName(tag) == "api" {
							hasAPI = true
							break
						}
					}
					if !hasAPI {
						t.Errorf("Result %s doesn't have 'api' tag", r.Item.Name)
					}
				}
			},
		},
		{
			name:       "search by type",
			query:      "type:pipeline",
			minResults: 2,
			checkResults: func(t *testing.T, results []SearchResult) {
				for _, r := range results {
					if r.Item.Type != ItemTypePipeline {
						t.Errorf("Expected pipeline type, got %s", r.Item.Type)
					}
				}
			},
		},
		{
			name:       "combined search",
			query:      "tag:api AND type:component",
			minResults: 2,
		},
		{
			name:       "OR search",
			query:      "tag:auth OR tag:security",
			minResults: 3,
		},
		{
			name:       "NOT search",
			query:      "tag:api NOT tag:security",
			minResults: 2,
		},
		{
			name:       "content search",
			query:      "content:error",
			minResults: 1,
			checkResults: func(t *testing.T, results []SearchResult) {
				if results[0].Item.Name != "api-prompt" {
					t.Errorf("Expected api-prompt in results, got %s", results[0].Item.Name)
				}
			},
		},
		{
			name:       "name search",
			query:      "name:auth",
			minResults: 2, // auth-prompt and auth-pipeline
		},
		{
			name:        "invalid query",
			query:       "invalid:field",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.Search(tt.query)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}
			
			if tt.checkResults != nil {
				tt.checkResults(t, results)
			}
		})
	}
}

func TestEngineScoring(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	engine := NewEngine()
	engine.BuildIndex()
	
	// Search for "api" - should rank exact name matches higher
	results, err := engine.Search("name:api")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) < 2 {
		t.Fatal("Expected at least 2 results")
	}
	
	// Check that items with "api" at the start of name rank higher
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not properly sorted by score: %f > %f", 
				results[i].Score, results[i-1].Score)
		}
	}
}

func TestEngineHighlights(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	engine := NewEngine()
	engine.BuildIndex()
	
	// Search for content
	results, err := engine.Search("content:error")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) == 0 {
		t.Fatal("No results found")
	}
	
	// Check highlights
	result := results[0]
	contentHighlights, exists := result.Highlights["content"]
	if !exists || len(contentHighlights) == 0 {
		t.Error("Expected content highlights")
	}
}

func TestExtractExcerpts(t *testing.T) {
	content := "This is a test document about error handling. Error messages should be clear. Avoid cryptic error codes."
	
	excerpts := extractExcerpts(content, "error", 3, 10)
	
	if len(excerpts) != 3 {
		t.Errorf("Expected 3 excerpts, got %d", len(excerpts))
	}
	
	// Check that excerpts contain the search term
	for _, excerpt := range excerpts {
		if !strings.Contains(strings.ToLower(excerpt), "error") {
			t.Errorf("Excerpt doesn't contain search term: %s", excerpt)
		}
	}
}

func TestTokenizeContent(t *testing.T) {
	content := "This is a test! With some punctuation, and MIXED case."
	
	tokens := tokenizeContent(content)
	
	expected := []string{"this", "test", "with", "some", "punctuation", "and", "mixed", "case"}
	
	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		return
	}
	
	for i, token := range tokens {
		if token != expected[i] {
			t.Errorf("Token[%d] = %q, want %q", i, token, expected[i])
		}
	}
}

func TestIntersectUnionSlices(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := []int{3, 4, 5, 6}
	
	// Test intersection
	intersection := intersectSlices(a, b)
	if len(intersection) != 2 {
		t.Errorf("Expected intersection of length 2, got %d", len(intersection))
	}
	
	// Test union
	union := unionSlices(a, b)
	if len(union) != 6 {
		t.Errorf("Expected union of length 6, got %d", len(union))
	}
}