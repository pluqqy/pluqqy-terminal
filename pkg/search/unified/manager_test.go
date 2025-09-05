package unified

import (
	"testing"
	"time"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

func TestUnifiedSearchManager_LoadAndSearch(t *testing.T) {
	manager := NewUnifiedSearchManager()
	
	// Create test data
	testComponents := []ComponentItem{
		{
			Name:         "Test Prompt",
			Path:         "/test/prompt.md",
			CompType:     models.ComponentTypePrompt,
			Tags:         []string{"test", "prompt"},
			LastModified: time.Now(),
		},
		{
			Name:         "API Context",
			Path:         "/test/context.md",
			CompType:     models.ComponentTypeContext,
			Tags:         []string{"api", "context"},
			LastModified: time.Now(),
		},
		{
			Name:         "Validation Rules",
			Path:         "/test/rules.md",
			CompType:     models.ComponentTypeRules,
			Tags:         []string{"validation"},
			LastModified: time.Now(),
		},
	}
	
	testPipelines := []PipelineItem{
		{
			Name:     "Test Pipeline",
			Path:     "/test/pipeline.yaml",
			Tags:     []string{"test", "pipeline"},
			Modified: time.Now(),
		},
		{
			Name:     "API Pipeline",
			Path:     "/test/api-pipeline.yaml",
			Tags:     []string{"api", "pipeline"},
			Modified: time.Now(),
		},
	}
	
	tests := []struct {
		name              string
		query             string
		includeArchived   bool
		expectedCompCount int
		expectedPipeCount int
	}{
		{
			name:              "search all with empty query",
			query:             "",
			includeArchived:   false,
			expectedCompCount: 3,
			expectedPipeCount: 2,
		},
		{
			name:              "search by tag",
			query:             "tag:test",
			includeArchived:   false,
			expectedCompCount: 1,
			expectedPipeCount: 1,
		},
		{
			name:              "search by type components",
			query:             "type:prompts",
			includeArchived:   false,
			expectedCompCount: 1,
			expectedPipeCount: 0,
		},
		{
			name:              "search by type pipelines",
			query:             "type:pipeline",
			includeArchived:   false,
			expectedCompCount: 0,
			expectedPipeCount: 2,
		},
		{
			name:              "search by name",
			query:             "name:API",
			includeArchived:   false,
			expectedCompCount: 1,
			expectedPipeCount: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load items into manager
			manager.LoadComponentItems(
				filterByType(testComponents, models.ComponentTypePrompt),
				filterByType(testComponents, models.ComponentTypeContext),
				filterByType(testComponents, models.ComponentTypeRules),
			)
			manager.LoadPipelineItems(testPipelines)
			
			// Set options
			manager.SetIncludeArchived(tt.includeArchived)
			
			// Perform search
			compResults, pipeResults, err := manager.SearchAll(tt.query)
			if err != nil {
				t.Fatalf("SearchAll failed: %v", err)
			}
			
			// Check component results
			if len(compResults) != tt.expectedCompCount {
				t.Errorf("Expected %d component results, got %d", tt.expectedCompCount, len(compResults))
			}
			
			// Check pipeline results
			if len(pipeResults) != tt.expectedPipeCount {
				t.Errorf("Expected %d pipeline results, got %d", tt.expectedPipeCount, len(pipeResults))
			}
		})
	}
}

func TestUnifiedSearchManager_FilterByQuery(t *testing.T) {
	manager := NewUnifiedSearchManager()
	
	// Test data
	prompts := []ComponentItem{
		{Name: "Error Prompt", Tags: []string{"error"}, CompType: models.ComponentTypePrompt},
		{Name: "Success Prompt", Tags: []string{"success"}, CompType: models.ComponentTypePrompt},
	}
	
	contexts := []ComponentItem{
		{Name: "API Context", Tags: []string{"api"}, CompType: models.ComponentTypeContext},
		{Name: "Error Context", Tags: []string{"error"}, CompType: models.ComponentTypeContext},
	}
	
	rules := []ComponentItem{
		{Name: "Validation Rules", Tags: []string{"validation"}, CompType: models.ComponentTypeRules},
	}
	
	pipelines := []PipelineItem{
		{Name: "Error Pipeline", Tags: []string{"error", "pipeline"}},
		{Name: "Success Pipeline", Tags: []string{"success", "pipeline"}},
	}
	
	tests := []struct {
		name             string
		query            string
		expectedPrompts  int
		expectedContexts int
		expectedRules    int
		expectedPipes    int
	}{
		{
			name:             "filter by tag error",
			query:            "tag:error",
			expectedPrompts:  1,
			expectedContexts: 1,
			expectedRules:    0,
			expectedPipes:    1,
		},
		{
			name:             "filter by tag success",
			query:            "tag:success",
			expectedPrompts:  1,
			expectedContexts: 0,
			expectedRules:    0,
			expectedPipes:    1,
		},
		{
			name:             "filter by name contains Error",
			query:            "name:Error",
			expectedPrompts:  1,
			expectedContexts: 1,
			expectedRules:    0,
			expectedPipes:    1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Filter components
			filteredPrompts, filteredContexts, filteredRules, err := manager.FilterComponentsByQuery(
				tt.query, prompts, contexts, rules,
			)
			if err != nil {
				t.Fatalf("FilterComponentsByQuery failed: %v", err)
			}
			
			// Filter pipelines
			filteredPipelines, err := manager.FilterPipelinesByQuery(tt.query, pipelines)
			if err != nil {
				t.Fatalf("FilterPipelinesByQuery failed: %v", err)
			}
			
			// Check results
			if len(filteredPrompts) != tt.expectedPrompts {
				t.Errorf("Expected %d prompts, got %d", tt.expectedPrompts, len(filteredPrompts))
			}
			
			if len(filteredContexts) != tt.expectedContexts {
				t.Errorf("Expected %d contexts, got %d", tt.expectedContexts, len(filteredContexts))
			}
			
			if len(filteredRules) != tt.expectedRules {
				t.Errorf("Expected %d rules, got %d", tt.expectedRules, len(filteredRules))
			}
			
			if len(filteredPipelines) != tt.expectedPipes {
				t.Errorf("Expected %d pipelines, got %d", tt.expectedPipes, len(filteredPipelines))
			}
		})
	}
}

func TestUnifiedSearchManager_Options(t *testing.T) {
	manager := NewUnifiedSearchManager()
	
	// Test SetIncludeArchived
	manager.SetIncludeArchived(true)
	if !manager.componentEngine.options.IncludeArchived {
		t.Error("SetIncludeArchived(true) did not set component engine option")
	}
	if !manager.pipelineEngine.options.IncludeArchived {
		t.Error("SetIncludeArchived(true) did not set pipeline engine option")
	}
	
	// Test SetMaxResults
	manager.SetMaxResults(50)
	if manager.componentEngine.options.MaxResults != 50 {
		t.Error("SetMaxResults did not set component engine option")
	}
	if manager.pipelineEngine.options.MaxResults != 50 {
		t.Error("SetMaxResults did not set pipeline engine option")
	}
}

// Helper function to filter components by type
func filterByType(items []ComponentItem, compType string) []ComponentItem {
	var filtered []ComponentItem
	for _, item := range items {
		if item.CompType == compType {
			filtered = append(filtered, item)
		}
	}
	return filtered
}