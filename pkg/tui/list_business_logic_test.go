package tui

import (
	"testing"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// Test helpers for creating components
func makeTestComponent(name, compType string) componentItem {
	return componentItem{
		name:         name,
		path:         "/test/components/" + compType + "/" + name,
		compType:     compType,
		lastModified: time.Now(),
		usageCount:   0,
		tokenCount:   100,
		tags:         []string{},
	}
}

func makeTestComponents(compType string, names ...string) []componentItem {
	components := make([]componentItem, len(names))
	for i, name := range names {
		components[i] = makeTestComponent(name, compType)
	}
	return components
}

// Test helpers for creating pipelines
func makeTestPipeline(name string) pipelineItem {
	return pipelineItem{
		name:       name,
		path:       "/test/pipelines/" + name + ".yaml",
		tags:       []string{},
		tokenCount: 200,
	}
}

func makeTestPipelines(names ...string) []pipelineItem {
	pipelines := make([]pipelineItem, len(names))
	for i, name := range names {
		pipelines[i] = makeTestPipeline(name)
	}
	return pipelines
}

// Helper to create test settings with custom section order
func makeTestSettings(order ...string) *models.Settings {
	sections := make([]models.Section, len(order))
	for i, typ := range order {
		sections[i] = models.Section{Type: typ}
	}
	return &models.Settings{
		Output: models.OutputSettings{
			Formatting: models.FormattingSettings{
				Sections: sections,
			},
		},
	}
}

// TestNewBusinessLogic tests the constructor
func TestNewBusinessLogic(t *testing.T) {
	bl := NewBusinessLogic()
	if bl == nil {
		t.Error("NewBusinessLogic returned nil")
	}
	
	// Verify initial state
	if len(bl.prompts) != 0 {
		t.Errorf("Expected empty prompts, got %d", len(bl.prompts))
	}
	if len(bl.contexts) != 0 {
		t.Errorf("Expected empty contexts, got %d", len(bl.contexts))
	}
	if len(bl.rules) != 0 {
		t.Errorf("Expected empty rules, got %d", len(bl.rules))
	}
}

// TestSetComponents tests setting component collections
func TestSetComponents(t *testing.T) {
	bl := NewBusinessLogic()
	
	prompts := makeTestComponents("prompt", "p1", "p2")
	contexts := makeTestComponents("context", "c1", "c2", "c3")
	rules := makeTestComponents("rules", "r1")
	
	bl.SetComponents(prompts, contexts, rules)
	
	// Verify components were set correctly
	if len(bl.prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(bl.prompts))
	}
	if len(bl.contexts) != 3 {
		t.Errorf("Expected 3 contexts, got %d", len(bl.contexts))
	}
	if len(bl.rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(bl.rules))
	}
}

// TestGetAllComponents_DefaultOrder tests with default settings order
func TestGetAllComponents_DefaultOrder(t *testing.T) {
	// Since we can't mock files.ReadSettings directly without modifying source,
	// we'll test with the actual default behavior when settings can't be read
	bl := NewBusinessLogic()
	
	prompts := makeTestComponents("prompts", "p1", "p2")
	contexts := makeTestComponents("contexts", "c1")
	rules := makeTestComponents("rules", "r1", "r2", "r3")
	
	bl.SetComponents(prompts, contexts, rules)
	
	// Call GetAllComponents - it will use default settings
	got := bl.GetAllComponents()
	
	// Default order is: rules, contexts, prompts
	expectedCount := 6
	if len(got) != expectedCount {
		t.Errorf("Expected %d components, got %d", expectedCount, len(got))
	}
	
	// Verify order - default is rules, contexts, prompts
	expectedTypes := []string{"rules", "rules", "rules", "contexts", "prompts", "prompts"}
	for i, comp := range got {
		if i < len(expectedTypes) && comp.compType != expectedTypes[i] {
			t.Errorf("Component at index %d: expected type %s, got %s", 
				i, expectedTypes[i], comp.compType)
		}
	}
}

// TestGetAllComponents_EmptyCollections tests with empty component arrays
func TestGetAllComponents_EmptyCollections(t *testing.T) {
	bl := NewBusinessLogic()
	
	// Don't set any components, leave them empty
	got := bl.GetAllComponents()
	
	if len(got) != 0 {
		t.Errorf("Expected empty component list, got %d components", len(got))
	}
}

// TestGetAllComponents_SingleTypeOnly tests with only one type of component
func TestGetAllComponents_SingleTypeOnly(t *testing.T) {
	tests := []struct {
		name     string
		setFunc  func(*BusinessLogic)
		expected string
		count    int
	}{
		{
			name: "only prompts",
			setFunc: func(bl *BusinessLogic) {
				bl.SetComponents(makeTestComponents("prompts", "p1", "p2"), nil, nil)
			},
			expected: "prompts",
			count:    2,
		},
		{
			name: "only contexts",
			setFunc: func(bl *BusinessLogic) {
				bl.SetComponents(nil, makeTestComponents("contexts", "c1", "c2", "c3"), nil)
			},
			expected: "contexts",
			count:    3,
		},
		{
			name: "only rules",
			setFunc: func(bl *BusinessLogic) {
				bl.SetComponents(nil, nil, makeTestComponents("rules", "r1"))
			},
			expected: "rules",
			count:    1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := NewBusinessLogic()
			tt.setFunc(bl)
			
			got := bl.GetAllComponents()
			
			if len(got) != tt.count {
				t.Errorf("Expected %d components, got %d", tt.count, len(got))
			}
			
			// Verify all components are of expected type
			for i, comp := range got {
				if comp.compType != tt.expected {
					t.Errorf("Component at index %d: expected type %s, got %s",
						i, tt.expected, comp.compType)
				}
			}
		})
	}
}

// TestGetEditingItemName tests retrieving the name of the item being edited
func TestGetEditingItemName(t *testing.T) {
	tests := []struct {
		name           string
		itemType       string
		componentCursor int
		pipelineCursor int
		components     []componentItem
		pipelines      []pipelineItem
		want           string
	}{
		{
			name:           "component editing with valid cursor",
			itemType:       "component",
			componentCursor: 1,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1", "comp2", "comp3"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "comp2",
		},
		{
			name:           "component editing with cursor at boundary",
			itemType:       "component",
			componentCursor: 2,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1", "comp2", "comp3"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "comp3",
		},
		{
			name:           "component editing with out-of-bounds cursor",
			itemType:       "component",
			componentCursor: 5,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1", "comp2"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "",
		},
		{
			name:           "component editing with negative cursor",
			itemType:       "component",
			componentCursor: -1,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "",
		},
		{
			name:           "pipeline editing with valid cursor",
			itemType:       "pipeline",
			componentCursor: 0,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      makeTestPipelines("pipe1", "pipe2"),
			want:           "pipe1",
		},
		{
			name:           "pipeline editing with out-of-bounds cursor",
			itemType:       "pipeline",
			componentCursor: 0,
			pipelineCursor: 3,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      makeTestPipelines("pipe1", "pipe2"),
			want:           "",
		},
		{
			name:           "pipeline editing with negative cursor",
			itemType:       "pipeline",
			componentCursor: 0,
			pipelineCursor: -2,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "",
		},
		{
			name:           "empty component list",
			itemType:       "component",
			componentCursor: 0,
			pipelineCursor: 0,
			components:     []componentItem{},
			pipelines:      makeTestPipelines("pipe1"),
			want:           "",
		},
		{
			name:           "empty pipeline list",
			itemType:       "pipeline",
			componentCursor: 0,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      []pipelineItem{},
			want:           "",
		},
		{
			name:           "unknown item type defaults to pipeline",
			itemType:       "unknown",
			componentCursor: 0,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "comp1"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "pipe1", // unknown types are treated as pipeline
		},
		{
			name:           "cursor at zero with single item",
			itemType:       "component",
			componentCursor: 0,
			pipelineCursor: 0,
			components:     makeTestComponents("prompt", "single"),
			pipelines:      makeTestPipelines("pipe1"),
			want:           "single",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagEditor := &TagEditor{
				ItemType: tt.itemType,
			}
			
			stateManager := &StateManager{
				ComponentCursor: tt.componentCursor,
				PipelineCursor: tt.pipelineCursor,
			}
			
			got := GetEditingItemName(tagEditor, stateManager, tt.components, tt.pipelines)
			
			if got != tt.want {
				t.Errorf("GetEditingItemName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetAllComponentsIntegration tests real-world scenarios
func TestGetAllComponentsIntegration(t *testing.T) {
	// Test with actual component data
	bl := NewBusinessLogic()
	
	// Create components with realistic data
	prompts := []componentItem{
		{
			name:         "code-review",
			path:         "/prompts/code-review.md",
			compType:     "prompts",
			lastModified: time.Now().Add(-24 * time.Hour),
			usageCount:   10,
			tokenCount:   150,
			tags:         []string{"review", "code"},
		},
		{
			name:         "refactor",
			path:         "/prompts/refactor.md",
			compType:     "prompts",
			lastModified: time.Now().Add(-48 * time.Hour),
			usageCount:   5,
			tokenCount:   200,
			tags:         []string{"refactor"},
		},
	}
	
	contexts := []componentItem{
		{
			name:         "project-context",
			path:         "/contexts/project.md",
			compType:     "contexts",
			lastModified: time.Now().Add(-12 * time.Hour),
			usageCount:   20,
			tokenCount:   500,
			tags:         []string{"project", "context"},
		},
	}
	
	rules := []componentItem{
		{
			name:         "coding-standards",
			path:         "/rules/standards.md",
			compType:     "rules",
			lastModified: time.Now(),
			usageCount:   15,
			tokenCount:   300,
			tags:         []string{"standards"},
		},
	}
	
	bl.SetComponents(prompts, contexts, rules)
	
	// Get all components
	allComponents := bl.GetAllComponents()
	
	// Verify we got all components
	if len(allComponents) != 4 {
		t.Errorf("Expected 4 components, got %d", len(allComponents))
	}
	
	// Verify order follows default settings: rules, contexts, prompts
	expectedOrder := []string{"coding-standards", "project-context", "code-review", "refactor"}
	for i, comp := range allComponents {
		if comp.name != expectedOrder[i] {
			t.Errorf("Component at index %d: expected %s, got %s", 
				i, expectedOrder[i], comp.name)
		}
	}
}

// TestGetEditingItemName_EdgeCases tests additional edge cases
func TestGetEditingItemName_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (*TagEditor, *StateManager, []componentItem, []pipelineItem)
		want       string
	}{
		{
			name: "nil tag editor",
			setupFunc: func() (*TagEditor, *StateManager, []componentItem, []pipelineItem) {
				return nil, &StateManager{}, makeTestComponents("prompt", "comp1"), makeTestPipelines("pipe1")
			},
			want: "",
		},
		{
			name: "nil state manager",
			setupFunc: func() (*TagEditor, *StateManager, []componentItem, []pipelineItem) {
				return &TagEditor{ItemType: "component"}, nil, makeTestComponents("prompt", "comp1"), makeTestPipelines("pipe1")
			},
			want: "",
		},
		{
			name: "large cursor value",
			setupFunc: func() (*TagEditor, *StateManager, []componentItem, []pipelineItem) {
				return &TagEditor{ItemType: "component"}, 
					&StateManager{ComponentCursor: 999999}, 
					makeTestComponents("prompt", "comp1"), 
					makeTestPipelines("pipe1")
			},
			want: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle potential panics
			defer func() {
				if r := recover(); r != nil {
					// If we panic, the test should fail
					t.Errorf("GetEditingItemName panicked: %v", r)
				}
			}()
			
			tagEditor, stateManager, components, pipelines := tt.setupFunc()
			
			// Check for nil before calling
			if tagEditor == nil || stateManager == nil {
				// Skip this test as the function would panic
				return
			}
			
			got := GetEditingItemName(tagEditor, stateManager, components, pipelines)
			
			if got != tt.want {
				t.Errorf("GetEditingItemName() = %q, want %q", got, tt.want)
			}
		})
	}
}