package unified

import (
	"testing"
	"time"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

func TestShouldIncludeArchived(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "empty query",
			query:    "",
			expected: false,
		},
		{
			name:     "simple search",
			query:    "test search",
			expected: false,
		},
		{
			name:     "status archived lowercase",
			query:    "status:archived",
			expected: true,
		},
		{
			name:     "status archived uppercase",
			query:    "STATUS:ARCHIVED",
			expected: true,
		},
		{
			name:     "status archived mixed case",
			query:    "Status:Archived",
			expected: true,
		},
		{
			name:     "status active",
			query:    "status:active",
			expected: false,
		},
		{
			name:     "tag query",
			query:    "tag:test",
			expected: false,
		},
		{
			name:     "complex query with archived",
			query:    "tag:test status:archived type:pipeline",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldIncludeArchived(tt.query)
			if result != tt.expected {
				t.Errorf("ShouldIncludeArchived(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestFilterSearchResultsByType(t *testing.T) {
	// Create test results with different types
	results := []SearchResult[*ComponentItemWrapper]{
		{
			Item: &ComponentItemWrapper{
				name:     "Test Prompt",
				compType: models.ComponentTypePrompt,
				tags:     []string{"test"},
			},
			Score: 10.0,
		},
		{
			Item: &ComponentItemWrapper{
				name:     "API Context",
				compType: models.ComponentTypeContext,
				tags:     []string{"api"},
			},
			Score: 8.0,
		},
		{
			Item: &ComponentItemWrapper{
				name:     "Coding Rules",
				compType: models.ComponentTypeRules,
				tags:     []string{"standards"},
			},
			Score: 6.0,
		},
		{
			Item: &ComponentItemWrapper{
				name:     "Another Prompt",
				compType: models.ComponentTypePrompt,
				tags:     []string{"another"},
			},
			Score: 5.0,
		},
	}
	
	prompts, contexts, rules := FilterSearchResultsByType(results)
	
	// Check counts
	if len(prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(prompts))
	}
	
	if len(contexts) != 1 {
		t.Errorf("Expected 1 context, got %d", len(contexts))
	}
	
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	
	// Check specific items
	if prompts[0].Name != "Test Prompt" {
		t.Errorf("Expected first prompt to be 'Test Prompt', got %s", prompts[0].Name)
	}
	
	if contexts[0].Name != "API Context" {
		t.Errorf("Expected context to be 'API Context', got %s", contexts[0].Name)
	}
	
	if rules[0].Name != "Coding Rules" {
		t.Errorf("Expected rule to be 'Coding Rules', got %s", rules[0].Name)
	}
}

func TestConvertPipelineResults(t *testing.T) {
	// Create test pipeline results
	results := []SearchResult[*PipelineItemWrapper]{
		{
			Item: &PipelineItemWrapper{
				name:       "Test Pipeline",
				path:       "/test/pipeline.yaml",
				tags:       []string{"test", "pipeline"},
				tokenCount: 100,
				isArchived: false,
				modified:   time.Now(),
			},
			Score: 10.0,
		},
		{
			Item: &PipelineItemWrapper{
				name:       "Archived Pipeline",
				path:       "/archive/old.yaml",
				tags:       []string{"old"},
				tokenCount: 50,
				isArchived: true,
				modified:   time.Now().Add(-24 * time.Hour),
			},
			Score: 5.0,
		},
	}
	
	pipelines := ConvertPipelineResults(results)
	
	// Check count
	if len(pipelines) != 2 {
		t.Fatalf("Expected 2 pipelines, got %d", len(pipelines))
	}
	
	// Check first pipeline
	if pipelines[0].Name != "Test Pipeline" {
		t.Errorf("Expected first pipeline name to be 'Test Pipeline', got %s", pipelines[0].Name)
	}
	
	if pipelines[0].IsArchived {
		t.Error("First pipeline should not be archived")
	}
	
	// Check second pipeline
	if pipelines[1].Name != "Archived Pipeline" {
		t.Errorf("Expected second pipeline name to be 'Archived Pipeline', got %s", pipelines[1].Name)
	}
	
	if !pipelines[1].IsArchived {
		t.Error("Second pipeline should be archived")
	}
}

func TestCombineComponentsByType(t *testing.T) {
	prompts := []ComponentItem{
		{Name: "Prompt1"},
		{Name: "Prompt2"},
	}
	
	contexts := []ComponentItem{
		{Name: "Context1"},
	}
	
	rules := []ComponentItem{
		{Name: "Rule1"},
		{Name: "Rule2"},
		{Name: "Rule3"},
	}
	
	combined := CombineComponentsByType(prompts, contexts, rules)
	
	// Check total count
	expectedCount := len(prompts) + len(contexts) + len(rules)
	if len(combined) != expectedCount {
		t.Errorf("Expected %d combined items, got %d", expectedCount, len(combined))
	}
	
	// Check order (prompts, then contexts, then rules)
	expectedNames := []string{"Prompt1", "Prompt2", "Context1", "Rule1", "Rule2", "Rule3"}
	for i, expected := range expectedNames {
		if i >= len(combined) {
			t.Errorf("Missing item at index %d: expected %s", i, expected)
			continue
		}
		if combined[i].Name != expected {
			t.Errorf("Item at index %d: expected %s, got %s", i, expected, combined[i].Name)
		}
	}
}