package search

import (
	"testing"
)

func TestSearchAPI(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	api, err := NewSearchAPI()
	if err != nil {
		t.Fatalf("Failed to create search API: %v", err)
	}
	
	t.Run("Search", func(t *testing.T) {
		results, err := api.Search("tag:api")
		if err != nil {
			t.Errorf("Search failed: %v", err)
		}
		
		if len(results) == 0 {
			t.Error("Expected results but got none")
		}
	})
	
	t.Run("QuickFilter", func(t *testing.T) {
		results, err := api.QuickFilter([]string{"api", "security"})
		if err != nil {
			t.Errorf("QuickFilter failed: %v", err)
		}
		
		// Should return items that have both tags
		for _, result := range results {
			hasAPI := false
			hasSecurity := false
			
			for _, tag := range result.Item.Tags {
				if tag == "api" {
					hasAPI = true
				}
				if tag == "security" {
					hasSecurity = true
				}
			}
			
			if !hasAPI || !hasSecurity {
				t.Errorf("Result %s doesn't have both required tags", result.Item.Name)
			}
		}
	})
	
	t.Run("SearchComponents", func(t *testing.T) {
		results, err := api.SearchComponents("", "prompts")
		if err != nil {
			t.Errorf("SearchComponents failed: %v", err)
		}
		
		for _, result := range results {
			if result.Item.Type != ItemTypeComponent {
				t.Errorf("Expected component type, got %s", result.Item.Type)
			}
		}
	})
	
	t.Run("SearchPipelines", func(t *testing.T) {
		results, err := api.SearchPipelines("")
		if err != nil {
			t.Errorf("SearchPipelines failed: %v", err)
		}
		
		for _, result := range results {
			if result.Item.Type != ItemTypePipeline {
				t.Errorf("Expected pipeline type, got %s", result.Item.Type)
			}
		}
	})
	
	t.Run("GetItemsByTags", func(t *testing.T) {
		results, err := api.GetItemsByTags([]string{"api", "production"})
		if err != nil {
			t.Errorf("GetItemsByTags failed: %v", err)
		}
		
		// Only api-pipeline should have both tags
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		
		if len(results) > 0 && results[0].Item.Name != "api-pipeline" {
			t.Errorf("Expected api-pipeline, got %s", results[0].Item.Name)
		}
	})
	
	t.Run("GetItemsByAnyTags", func(t *testing.T) {
		results, err := api.GetItemsByAnyTags([]string{"critical", "documentation"})
		if err != nil {
			t.Errorf("GetItemsByAnyTags failed: %v", err)
		}
		
		// Should return items with either tag
		if len(results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(results))
		}
	})
	
	t.Run("SearchByContent", func(t *testing.T) {
		results, err := api.SearchByContent("authentication")
		if err != nil {
			t.Errorf("SearchByContent failed: %v", err)
		}
		
		if len(results) == 0 {
			t.Error("Expected results for 'authentication' content search")
		}
	})
	
	t.Run("RefreshIndex", func(t *testing.T) {
		err := api.RefreshIndex()
		if err != nil {
			t.Errorf("RefreshIndex failed: %v", err)
		}
	})
}

func TestSearchAPISorting(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()
	
	api, _ := NewSearchAPI()
	
	tests := []struct {
		name   string
		sortBy string
		check  func(*testing.T, []SearchResult)
	}{
		{
			name:   "sort by name",
			sortBy: "name",
			check: func(t *testing.T, results []SearchResult) {
				for i := 1; i < len(results); i++ {
					if results[i].Item.Name < results[i-1].Item.Name {
						t.Errorf("Results not sorted by name: %s < %s",
							results[i].Item.Name, results[i-1].Item.Name)
					}
				}
			},
		},
		{
			name:   "sort by modified",
			sortBy: "modified",
			check: func(t *testing.T, results []SearchResult) {
				for i := 1; i < len(results); i++ {
					if results[i].Item.Modified.After(results[i-1].Item.Modified) {
						t.Error("Results not sorted by modified date (newest first)")
					}
				}
			},
		},
		{
			name:   "sort by relevance",
			sortBy: "relevance",
			check: func(t *testing.T, results []SearchResult) {
				for i := 1; i < len(results); i++ {
					if results[i].Score > results[i-1].Score {
						t.Errorf("Results not sorted by score: %f > %f",
							results[i].Score, results[i-1].Score)
					}
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SearchOptions{
				Query:  "",
				SortBy: tt.sortBy,
			}
			
			results, err := api.SearchWithOptions(opts)
			if err != nil {
				t.Errorf("Search failed: %v", err)
				return
			}
			
			if len(results) < 2 {
				t.Skip("Not enough results to test sorting")
			}
			
			tt.check(t, results)
		})
	}
}