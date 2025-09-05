package unified

import (
	"testing"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

func TestSearchHelper_UnifiedFilterAll(t *testing.T) {
	helper := NewSearchHelper()
	
	// Create test data
	prompts := []ComponentItem{
		{
			Name:     "Error Handler Prompt",
			Tags:     []string{"error", "handling"},
			CompType: models.ComponentTypePrompt,
		},
		{
			Name:     "Input Validator Prompt",
			Tags:     []string{"validation", "input"},
			CompType: models.ComponentTypePrompt,
		},
	}
	
	contexts := []ComponentItem{
		{
			Name:     "API Documentation",
			Tags:     []string{"api", "docs"},
			CompType: models.ComponentTypeContext,
		},
		{
			Name:     "Error Codes Context",
			Tags:     []string{"error", "reference"},
			CompType: models.ComponentTypeContext,
		},
	}
	
	rules := []ComponentItem{
		{
			Name:     "Coding Standards",
			Tags:     []string{"standards", "quality"},
			CompType: models.ComponentTypeRules,
		},
	}
	
	pipelines := []PipelineItem{
		{
			Name: "Error Processing Pipeline",
			Tags: []string{"error", "pipeline"},
		},
		{
			Name: "Data Validation Pipeline",
			Tags: []string{"validation", "pipeline"},
		},
	}
	
	tests := []struct {
		name               string
		query              string
		includeArchived    bool
		expectedPrompts    int
		expectedContexts   int
		expectedRules      int
		expectedPipelines  int
	}{
		{
			name:               "empty query returns all",
			query:              "",
			includeArchived:    false,
			expectedPrompts:    2,
			expectedContexts:   2,
			expectedRules:      1,
			expectedPipelines:  2,
		},
		{
			name:               "tag filter - error",
			query:              "tag:error",
			includeArchived:    false,
			expectedPrompts:    1,
			expectedContexts:   1,
			expectedRules:      0,
			expectedPipelines:  1,
		},
		{
			name:               "type filter - prompts",
			query:              "type:prompts",
			includeArchived:    false,
			expectedPrompts:    2,
			expectedContexts:   0,
			expectedRules:      0,
			expectedPipelines:  0,
		},
		{
			name:               "type filter - pipeline",
			query:              "type:pipeline",
			includeArchived:    false,
			expectedPrompts:    0,
			expectedContexts:   0,
			expectedRules:      0,
			expectedPipelines:  2,
		},
		{
			name:               "name filter - validator",
			query:              "name:Validator",
			includeArchived:    false,
			expectedPrompts:    1,
			expectedContexts:   0,
			expectedRules:      0,
			expectedPipelines:  0,
		},
		{
			name:               "name filter - error",
			query:              "name:Error",
			includeArchived:    false,
			expectedPrompts:    1,
			expectedContexts:   1,
			expectedRules:      0,
			expectedPipelines:  1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set options
			helper.SetSearchOptions(tt.includeArchived, 100, "relevance")
			
			// Perform search
			filteredPrompts, filteredContexts, filteredRules, filteredPipelines, err := helper.UnifiedFilterAll(
				tt.query, prompts, contexts, rules, pipelines,
			)
			
			if err != nil {
				t.Fatalf("UnifiedFilterAll failed: %v", err)
			}
			
			// Check results
			if len(filteredPrompts) != tt.expectedPrompts {
				t.Errorf("Expected %d prompts, got %d", tt.expectedPrompts, len(filteredPrompts))
				for _, p := range filteredPrompts {
					t.Logf("  - %s", p.Name)
				}
			}
			
			if len(filteredContexts) != tt.expectedContexts {
				t.Errorf("Expected %d contexts, got %d", tt.expectedContexts, len(filteredContexts))
				for _, c := range filteredContexts {
					t.Logf("  - %s", c.Name)
				}
			}
			
			if len(filteredRules) != tt.expectedRules {
				t.Errorf("Expected %d rules, got %d", tt.expectedRules, len(filteredRules))
				for _, r := range filteredRules {
					t.Logf("  - %s", r.Name)
				}
			}
			
			if len(filteredPipelines) != tt.expectedPipelines {
				t.Errorf("Expected %d pipelines, got %d", tt.expectedPipelines, len(filteredPipelines))
				for _, p := range filteredPipelines {
					t.Logf("  - %s", p.Name)
				}
			}
		})
	}
}

func TestSearchHelper_SetSearchOptions(t *testing.T) {
	helper := NewSearchHelper()
	
	// Test setting options
	helper.SetSearchOptions(true, 50, "name")
	
	manager := helper.GetUnifiedManager()
	if manager == nil {
		t.Fatal("UnifiedManager is nil")
	}
	
	// Check that options were set on both engines
	if !manager.componentEngine.options.IncludeArchived {
		t.Error("IncludeArchived not set on component engine")
	}
	
	if manager.componentEngine.options.MaxResults != 50 {
		t.Error("MaxResults not set on component engine")
	}
	
	if manager.componentEngine.options.SortBy != "name" {
		t.Error("SortBy not set on component engine")
	}
}

func TestSearchHelper_StructuredQueries(t *testing.T) {
	helper := NewSearchHelper()
	
	tests := []struct {
		query       string
		isStructured bool
	}{
		{"tag:test", true},
		{"type:pipeline", true},
		{"content:search", true},
		{"name:component", true},
		{"status:archived", true},
		{"simple search", false},
		{"test", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := helper.IsStructuredQuery(tt.query)
			if result != tt.isStructured {
				t.Errorf("IsStructuredQuery(%q) = %v, want %v", tt.query, result, tt.isStructured)
			}
		})
	}
}