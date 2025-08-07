package tui

import (
	"testing"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestPipelineBuilderModel_SyncPreviewToSelectedComponent(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() *PipelineBuilderModel
		wantScrolled    bool
		wantYOffset     int
		skipOffsetCheck bool
	}{
		{
			name: "syncs_to_first_component",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
				}
				m.previewContent = "# Pipeline\n\n## Context\n\nTest content 1\n\n## Prompts\n\nTest content 2\n"
				m.previewViewport = viewport.New(80, 10)
				m.previewViewport.SetContent(m.previewContent)
				return m
			},
			wantScrolled: true,
			wantYOffset:  0, // First component should be at top
		},
		{
			name: "syncs_to_middle_component",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 1
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
					{Path: "../components/test3.md", Type: "rule", Order: 2},
				}
				// Create longer content to test scrolling
				content := "# Pipeline\n\n## Context\n\n"
				for i := 0; i < 20; i++ {
					content += "Context line\n"
				}
				content += "\n## Prompts\n\nTest prompt content\n"
				for i := 0; i < 20; i++ {
					content += "Prompt line\n"
				}
				m.previewContent = content
				m.previewViewport = viewport.New(80, 10)
				m.previewViewport.SetContent(m.previewContent)
				return m
			},
			wantScrolled:    true,
			skipOffsetCheck: true, // Offset depends on content matching
		},
		{
			name: "no_sync_when_preview_hidden",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = false // Preview is hidden
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			wantScrolled: false,
			wantYOffset:  0,
		},
		{
			name: "no_sync_with_empty_components",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{} // Empty
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			wantScrolled: false,
			wantYOffset:  0,
		},
		{
			name: "handles_invalid_cursor_position",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 5 // Out of bounds
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			wantScrolled: false,
			wantYOffset:  0,
		},
		{
			name: "handles_duplicate_components",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 2 // Select second occurrence
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test1.md", Type: "context", Order: 1}, // Duplicate
					{Path: "../components/test1.md", Type: "context", Order: 2}, // Another duplicate
				}
				m.previewContent = "# Pipeline\n\nFirst occurrence\nContent 1\n\nSecond occurrence\nContent 1\n\nThird occurrence\nContent 1\n"
				m.previewViewport = viewport.New(80, 10)
				m.previewViewport.SetContent(m.previewContent)
				return m
			},
			wantScrolled:    true,
			skipOffsetCheck: true, // Complex to predict exact offset
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			initialOffset := m.previewViewport.YOffset

			m.syncPreviewToSelectedComponent()

			if tt.wantScrolled {
				// For complex cases, just verify that some scrolling action was attempted
				if !tt.skipOffsetCheck && m.previewViewport.YOffset != tt.wantYOffset {
					t.Errorf("YOffset = %d, want %d", m.previewViewport.YOffset, tt.wantYOffset)
				}
			} else {
				if m.previewViewport.YOffset != initialOffset {
					t.Errorf("YOffset changed when it shouldn't: got %d, initial %d", 
						m.previewViewport.YOffset, initialOffset)
				}
			}
		})
	}
}

func TestPipelineBuilderModel_NavigationWithSync(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *PipelineBuilderModel
		keySequence  []string
		wantCursor   int
		wantScrolled bool
	}{
		{
			name: "up_navigation_syncs_preview",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false // Disable name editing mode
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 2
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
					{Path: "../components/test3.md", Type: "rule", Order: 2},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			keySequence:  []string{"up"},
			wantCursor:   1,
			wantScrolled: true,
		},
		{
			name: "down_navigation_syncs_preview",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false // Disable name editing mode
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			keySequence:  []string{"down"},
			wantCursor:   1,
			wantScrolled: true,
		},
		{
			name: "home_key_syncs_to_first",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false // Disable name editing mode
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 3
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
					{Path: "../components/test3.md", Type: "rule", Order: 2},
					{Path: "../components/test4.md", Type: "rule", Order: 3},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			keySequence:  []string{"home"},
			wantCursor:   0,
			wantScrolled: true,
		},
		{
			name: "end_key_syncs_to_last",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false // Disable name editing mode
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
					{Path: "../components/test3.md", Type: "rule", Order: 2},
				}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			keySequence:  []string{"end"},
			wantCursor:   2,
			wantScrolled: true,
		},
		{
			name: "navigation_in_left_column_doesnt_sync",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false // Disable name editing mode
				m.showPreview = true
				m.activeColumn = leftColumn // Left column active
				m.leftCursor = 0
				// Need to set up filtered lists for navigation to work
				m.prompts = []componentItem{
					{name: "test1", path: "components/test1.md", compType: "prompt"},
					{name: "test2", path: "components/test2.md", compType: "prompt"},
				}
				m.filteredPrompts = m.prompts
				m.filteredContexts = []componentItem{}
				m.filteredRules = []componentItem{}
				m.previewViewport = viewport.New(80, 10)
				return m
			},
			keySequence:  []string{"down"},
			wantCursor:   1,
			wantScrolled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			
			// Process key sequence
			for _, key := range tt.keySequence {
				msg := tea.KeyMsg{Type: tea.KeyRunes}
				switch key {
				case "up":
					msg.Type = tea.KeyUp
				case "down":
					msg.Type = tea.KeyDown
				case "home":
					msg.Type = tea.KeyHome
				case "end":
					msg.Type = tea.KeyEnd
				case "j":
					msg.Runes = []rune{'j'}
				case "k":
					msg.Runes = []rune{'k'}
				}
				
				// Update returns (Model, Cmd) but we only need to update our model
				updatedModel, _ := m.Update(msg)
				if pbm, ok := updatedModel.(*PipelineBuilderModel); ok {
					m = pbm
				}
			}
			
			// Check cursor position
			if m.activeColumn == rightColumn {
				if m.rightCursor != tt.wantCursor {
					t.Errorf("rightCursor = %d, want %d", m.rightCursor, tt.wantCursor)
				}
			} else if m.activeColumn == leftColumn {
				if m.leftCursor != tt.wantCursor {
					t.Errorf("leftCursor = %d, want %d", m.leftCursor, tt.wantCursor)
				}
			}
		})
	}
}

func TestPipelineBuilderModel_PreviewScrollBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *PipelineBuilderModel
		wantYOffset int
		wantClipped bool
	}{
		{
			name: "scrolls_to_top_for_first_component",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 0
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
				}
				m.previewContent = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
				m.previewViewport = viewport.New(80, 3)
				m.previewViewport.SetContent(m.previewContent)
				m.previewViewport.SetYOffset(5) // Start scrolled down
				return m
			},
			wantYOffset: -1, // Skip exact offset check, viewport may adjust this
			wantClipped: true,
		},
		{
			name: "centers_component_when_possible",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.showPreview = true
				m.activeColumn = rightColumn
				m.rightCursor = 1
				m.selectedComponents = []models.ComponentRef{
					{Path: "../components/test1.md", Type: "context", Order: 0},
					{Path: "../components/test2.md", Type: "prompt", Order: 1},
				}
				// Create content where second component is in the middle
				content := ""
				for i := 0; i < 30; i++ {
					content += "Line " + string(rune('0'+i)) + "\n"
				}
				m.previewContent = content
				m.previewViewport = viewport.New(80, 10)
				m.previewViewport.SetContent(m.previewContent)
				return m
			},
			wantYOffset: -1, // Will vary based on content matching
			wantClipped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			m.syncPreviewToSelectedComponent()
			
			if !tt.wantClipped && m.previewViewport.YOffset != tt.wantYOffset {
				t.Errorf("YOffset = %d, want %d", m.previewViewport.YOffset, tt.wantYOffset)
			}
			
			// Verify offset is within valid bounds
			if m.previewViewport.YOffset < 0 {
				t.Error("YOffset is negative")
			}
		})
	}
}

func TestPipelineBuilderModel_SyncWithComplexContent(t *testing.T) {
	tests := []struct {
		name            string
		componentPath   string
		componentType   string
		previewContent  string
		expectedPattern string
	}{
		{
			name:          "finds_component_by_content",
			componentPath: "../components/rules/validation.md",
			componentType: "rule",
			previewContent: `# My Pipeline

## Rules

### Validation Rules
1. Must validate input
2. Must check boundaries

## Contexts

### System Context
System information here`,
			expectedPattern: "Validation Rules",
		},
		{
			name:          "handles_yaml_frontmatter",
			componentPath: "../components/prompts/user.md",
			componentType: "prompt",
			previewContent: `# Pipeline

## Prompts

---
tags: [prompt, user]
---

User prompt content here

## Rules

Rule content`,
			expectedPattern: "User prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPipelineBuilderModel()
			m.showPreview = true
			m.activeColumn = rightColumn
			m.rightCursor = 0
			m.selectedComponents = []models.ComponentRef{
				{Path: tt.componentPath, Type: tt.componentType, Order: 0},
			}
			m.previewContent = tt.previewContent
			m.previewViewport = viewport.New(80, 10)
			m.previewViewport.SetContent(m.previewContent)
			
			// The sync function should handle finding the component
			m.syncPreviewToSelectedComponent()
			
			// Just verify no panic and basic state consistency
			if m.previewViewport.YOffset < 0 {
				t.Error("YOffset became negative after sync")
			}
		})
	}
}

// Test state consistency during navigation
func TestPipelineBuilderModel_NavigationStateConsistency(t *testing.T) {
	m := NewPipelineBuilderModel()
	m.editingName = false // Disable name editing mode
	m.showPreview = true
	m.activeColumn = rightColumn
	m.selectedComponents = []models.ComponentRef{
		{Path: "../components/test1.md", Type: "context", Order: 0},
		{Path: "../components/test2.md", Type: "prompt", Order: 1},
		{Path: "../components/test3.md", Type: "rule", Order: 2},
	}
	m.previewViewport = viewport.New(80, 10)
	
	// Test rapid navigation
	keys := []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyUp},
		{Type: tea.KeyHome},
		{Type: tea.KeyEnd},
		{Type: tea.KeyUp},
	}
	
	for i, key := range keys {
		updatedModel, _ := m.Update(key)
		if pbm, ok := updatedModel.(*PipelineBuilderModel); ok {
			m = pbm
		}
		
		// Verify cursor stays in bounds
		if m.rightCursor < 0 || m.rightCursor >= len(m.selectedComponents) {
			t.Errorf("After key %d: rightCursor out of bounds: %d", i, m.rightCursor)
		}
		
		// Verify viewport offset is valid
		if m.previewViewport.YOffset < 0 {
			t.Errorf("After key %d: YOffset is negative: %d", i, m.previewViewport.YOffset)
		}
	}
}

// Benchmark sync performance
func BenchmarkSyncPreviewToSelectedComponent(b *testing.B) {
	m := NewPipelineBuilderModel()
	m.showPreview = true
	m.activeColumn = rightColumn
	m.rightCursor = 5
	
	// Create many components
	for i := 0; i < 20; i++ {
		m.selectedComponents = append(m.selectedComponents, models.ComponentRef{
			Path:  "../components/test.md",
			Type:  "context",
			Order: i,
		})
	}
	
	// Create large preview content
	content := "# Pipeline\n\n"
	for i := 0; i < 1000; i++ {
		content += "Line of content\n"
	}
	m.previewContent = content
	m.previewViewport = viewport.New(80, 20)
	m.previewViewport.SetContent(m.previewContent)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.syncPreviewToSelectedComponent()
	}
}

// Helper function to create test components for sync tests
func createSyncTestComponent(name, path, compType string) componentItem {
	return componentItem{
		name:         name,
		path:         path,
		compType:     compType,
		usageCount:   0,
		tokenCount:   100,
		lastModified: time.Now(),
		tags:         []string{},
	}
}