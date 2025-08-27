package shared

import (
	"testing"
	"time"
	
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestSearchEngine_TypeFiltering(t *testing.T) {
	// Create test data
	testPipelines := []PipelineItem{
		{
			Name:     "Pipeline 1",
			Path:     "/test/pipeline1.yaml",
			Tags:     []string{"test"},
			Modified: time.Now(),
		},
		{
			Name:     "Pipeline 2",
			Path:     "/test/pipeline2.yaml",
			Tags:     []string{"demo"},
			Modified: time.Now(),
		},
	}
	
	testPrompts := []ComponentItem{
		{
			Name:         "Prompt 1",
			Path:         "/test/prompt1.md",
			CompType:     models.ComponentTypePrompt,
			Tags:         []string{"test"},
			LastModified: time.Now(),
		},
		{
			Name:         "Prompt 2",
			Path:         "/test/prompt2.md",
			CompType:     models.ComponentTypePrompt,
			Tags:         []string{"demo"},
			LastModified: time.Now(),
		},
	}
	
	testContexts := []ComponentItem{
		{
			Name:         "Context 1",
			Path:         "/test/context1.md",
			CompType:     models.ComponentTypeContext,
			Tags:         []string{"demo"},
			LastModified: time.Now(),
		},
	}
	
	testRules := []ComponentItem{
		{
			Name:         "Rules 1",
			Path:         "/test/rules1.md",
			CompType:     models.ComponentTypeRules,
			Tags:         []string{"test"},
			LastModified: time.Now(),
		},
	}
	
	tests := []struct {
		name              string
		query             string
		expectedPipelines int
		expectedPrompts   int
		expectedContexts  int
		expectedRules     int
	}{
		// Pipeline type queries
		{
			name:              "type:pipeline singular",
			query:             "type:pipeline",
			expectedPipelines: 2,
			expectedPrompts:   0,
			expectedContexts:  0,
			expectedRules:     0,
		},
		{
			name:              "type:pipelines plural",
			query:             "type:pipelines",
			expectedPipelines: 2,
			expectedPrompts:   0,
			expectedContexts:  0,
			expectedRules:     0,
		},
		// Component type queries - plural forms (as stored)
		{
			name:              "type:prompts plural",
			query:             "type:prompts",
			expectedPipelines: 0,
			expectedPrompts:   2,
			expectedContexts:  0,
			expectedRules:     0,
		},
		{
			name:              "type:contexts plural",
			query:             "type:contexts",
			expectedPipelines: 0,
			expectedPrompts:   0,
			expectedContexts:  1,
			expectedRules:     0,
		},
		{
			name:              "type:rules plural",
			query:             "type:rules",
			expectedPipelines: 0,
			expectedPrompts:   0,
			expectedContexts:  0,
			expectedRules:     1,
		},
		// Component type queries - singular forms
		// Note: Component types are stored as plural in models.ComponentType constants,
		// so singular searches will match if we trim 's' from the stored plural form
		{
			name:              "type:prompt singular",
			query:             "type:prompt",
			expectedPipelines: 0,
			expectedPrompts:   2,
			expectedContexts:  0,
			expectedRules:     0,
		},
		{
			name:              "type:context singular",
			query:             "type:context",
			expectedPipelines: 0,
			expectedPrompts:   0,
			expectedContexts:  1,
			expectedRules:     0,
		},
		{
			name:              "type:rule singular", 
			query:             "type:rule",
			expectedPipelines: 0,
			expectedPrompts:   0,
			expectedContexts:  0,
			expectedRules:     1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create search helper
			searchHelper := NewSearchHelper()
			searchHelper.SetSearchOptions(false, 1000, "relevance")
			
			// Perform search
			filteredPrompts, filteredContexts, filteredRules, filteredPipelines, err := searchHelper.UnifiedFilterAll(
				tt.query, testPrompts, testContexts, testRules, testPipelines,
			)
			
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			
			// Check results
			if len(filteredPipelines) != tt.expectedPipelines {
				t.Errorf("Expected %d pipelines, got %d", tt.expectedPipelines, len(filteredPipelines))
			}
			
			if len(filteredPrompts) != tt.expectedPrompts {
				t.Errorf("Expected %d prompts, got %d", tt.expectedPrompts, len(filteredPrompts))
			}
			
			if len(filteredContexts) != tt.expectedContexts {
				t.Errorf("Expected %d contexts, got %d", tt.expectedContexts, len(filteredContexts))
			}
			
			if len(filteredRules) != tt.expectedRules {
				t.Errorf("Expected %d rules, got %d", tt.expectedRules, len(filteredRules))
			}
		})
	}
}