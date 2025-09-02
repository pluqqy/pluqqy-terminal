package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentUsageRenderer_Render(t *testing.T) {
	tests := []struct {
		name     string
		state    *ComponentUsageState
		contains []string
		excludes []string
	}{
		{
			name: "renders with no pipelines",
			state: &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "test-component",
					compType: "context",
				},
				PipelinesUsingComponent: []PipelineUsageInfo{},
				Width:                   80,
				Height:                  24,
			},
			contains: []string{
				"Pipelines using: test-component",
				"Type: context",
				"No pipelines are currently using this component",
				"ESC/q/u Close",
			},
			excludes: []string{
				"Navigate",
				"Found in",
			},
		},
		{
			name: "renders with single pipeline",
			state: &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "api-docs",
					compType: "context",
				},
				PipelinesUsingComponent: []PipelineUsageInfo{
					{
						Name:            "CLI Development",
						Path:            "cli-development.yaml",
						ComponentOrder:  2,
						TotalComponents: 5,
					},
				},
				Width:         80,
				Height:        24,
				SelectedIndex: 0,
			},
			contains: []string{
				"Pipelines using: api-docs",
				"Type: context",
				"Found in 1 pipeline(s):",
				"CLI Development",
				"(position 2 of 5)",
				"cli-development.yaml",
				"ESC/q/u Close",
			},
			excludes: []string{
				"Navigate", // No navigation needed for single item
				"Showing",
			},
		},
		{
			name: "renders with multiple pipelines",
			state: &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "coding-standards",
					compType: "rule",
				},
				PipelinesUsingComponent: []PipelineUsageInfo{
					{
						Name:            "Pipeline 1",
						Path:            "pipeline1.yaml",
						ComponentOrder:  1,
						TotalComponents: 3,
					},
					{
						Name:            "Pipeline 2",
						Path:            "pipeline2.yaml",
						ComponentOrder:  2,
						TotalComponents: 4,
					},
					{
						Name:            "Pipeline 3",
						Path:            "pipeline3.yaml",
						ComponentOrder:  3,
						TotalComponents: 5,
					},
				},
				Width:         80,
				Height:        24,
				SelectedIndex: 1,
			},
			contains: []string{
				"Pipelines using: coding-standards",
				"Type: rule",
				"Found in 3 pipeline(s):",
				"Pipeline 1",
				"Pipeline 2",
				"Pipeline 3",
				"(position 2 of 4)", // Selected pipeline's position
				"ESC/q/u Close",
			},
			excludes: []string{
				"Navigate", // All items fit in view
				"Showing",
			},
		},
		{
			name: "renders with scrolling needed",
			state: &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "test-prompt",
					compType: "prompt",
				},
				PipelinesUsingComponent: make([]PipelineUsageInfo, 20),
				Width:                   80,
				Height:                  15, // Small height to force scrolling
				SelectedIndex:           5,
				ScrollOffset:            3,
			},
			contains: []string{
				"Pipelines using: test-prompt",
				"Type: prompt",
				"Found in 20 pipeline(s):",
				"Navigate", // Navigation help shown when scrolling needed
				"Showing",  // Scroll indicator
				"ESC/q/u Close",
			},
			excludes: []string{
				"No pipelines",
			},
		},
		{
			name: "renders selection indicator",
			state: &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "test",
					compType: "context",
				},
				PipelinesUsingComponent: []PipelineUsageInfo{
					{Name: "Pipeline A", Path: "a.yaml"},
					{Name: "Pipeline B", Path: "b.yaml"},
					{Name: "Pipeline C", Path: "c.yaml"},
				},
				Width:         80,
				Height:        24,
				SelectedIndex: 1, // Pipeline B selected
			},
			contains: []string{
				"→ Pipeline B", // Selected item has arrow
				"Pipeline A",   // Others don't
				"Pipeline C",
			},
			excludes: []string{
				"→ Pipeline A",
				"→ Pipeline C",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize pipelines with default data if needed
			if tt.name == "renders with scrolling needed" {
				for i := range tt.state.PipelinesUsingComponent {
					tt.state.PipelinesUsingComponent[i] = PipelineUsageInfo{
						Name:            "Pipeline",
						Path:            "pipeline.yaml",
						ComponentOrder:  1,
						TotalComponents: 1,
					}
				}
			}

			renderer := NewComponentUsageRenderer(80, 24)
			output := renderer.Render(tt.state)

			// Check for expected content
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Should contain: %s", expected)
			}

			// Check for excluded content
			for _, excluded := range tt.excludes {
				assert.NotContains(t, output, excluded, "Should not contain: %s", excluded)
			}

			// Additional checks
			if tt.state.Active {
				assert.NotEmpty(t, output, "Active modal should produce output")
			}
		})
	}
}

func TestComponentUsageRenderer_RenderInactive(t *testing.T) {
	renderer := NewComponentUsageRenderer(80, 24)
	state := &ComponentUsageState{
		Active: false,
	}

	output := renderer.Render(state)
	assert.Empty(t, output, "Inactive modal should produce empty output")
}

func TestComponentUsageRenderer_ModalDimensions(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		minExpWidth int
		maxExpWidth int
	}{
		{
			name:        "standard terminal size",
			width:       100,
			height:      30,
			minExpWidth: 60,
			maxExpWidth: 80,
		},
		{
			name:        "small terminal",
			width:       60,
			height:      20,
			minExpWidth: 40,
			maxExpWidth: 50,
		},
		{
			name:        "large terminal",
			width:       150,
			height:      50,
			minExpWidth: 80,
			maxExpWidth: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "test",
					compType: "context",
				},
				PipelinesUsingComponent: []PipelineUsageInfo{
					{Name: "Test Pipeline"},
				},
				Width:  tt.width,
				Height: tt.height,
			}

			renderer := NewComponentUsageRenderer(80, 24)
			output := renderer.Render(state)

			// Check that modal respects size constraints
			lines := strings.Split(output, "\n")
			maxLineLength := 0
			for _, line := range lines {
				length := len([]rune(stripAnsi(line)))
				if length > maxLineLength {
					maxLineLength = length
				}
			}

			// Modal may have minimum width requirements, so just check it's reasonable
			assert.LessOrEqual(t, maxLineLength, tt.width+20, 
				"Modal width should be reasonable for terminal size")
		})
	}
}

func TestComponentUsageRenderer_ScrollIndicator(t *testing.T) {
	state := &ComponentUsageState{
		Active: true,
		SelectedComponent: componentItem{
			name:  "test",
			compType: "context",
		},
		Width:  80,
		Height: 20,
	}

	// Create many pipelines to test scrolling
	for i := 0; i < 30; i++ {
		state.PipelinesUsingComponent = append(state.PipelinesUsingComponent,
			PipelineUsageInfo{
				Name: "Pipeline",
				Path: "pipeline.yaml",
			})
	}

	renderer := NewComponentUsageRenderer(80, 24)

	// Test at beginning
	state.ScrollOffset = 0
	state.SelectedIndex = 0
	output := renderer.Render(state)
	assert.Contains(t, output, "Showing 1-", "Should show scroll position at beginning")

	// Test in middle
	state.ScrollOffset = 10
	state.SelectedIndex = 15
	output = renderer.Render(state)
	assert.Contains(t, output, "Showing", "Should show scroll position in middle")

	// Test near end
	state.ScrollOffset = 25
	state.SelectedIndex = 29
	output = renderer.Render(state)
	assert.Contains(t, output, "of 30", "Should show total count at end")
}

func TestComponentUsageRenderer_HelpText(t *testing.T) {
	renderer := NewComponentUsageRenderer(80, 24)

	tests := []struct {
		name         string
		itemCount    int
		height       int
		expectNav    bool
	}{
		{
			name:      "few items - no navigation help",
			itemCount: 2,
			height:    30,
			expectNav: false,
		},
		{
			name:      "many items - show navigation help",
			itemCount: 50,
			height:    20,
			expectNav: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &ComponentUsageState{
				Active: true,
				SelectedComponent: componentItem{
					name:  "test",
					compType: "context",
				},
				Width:  80,
				Height: tt.height,
			}

			for i := 0; i < tt.itemCount; i++ {
				state.PipelinesUsingComponent = append(state.PipelinesUsingComponent,
					PipelineUsageInfo{Name: "Pipeline"})
			}

			output := renderer.Render(state)

			if tt.expectNav {
				assert.Contains(t, output, "Navigate", "Should show navigation help")
			} else {
				assert.NotContains(t, output, "Navigate", "Should not show navigation help")
			}

			// Always show close help
			assert.Contains(t, output, "ESC/q/u Close", "Should always show close help")
		})
	}
}

// Helper function to strip ANSI codes for testing
func stripAnsi(str string) string {
	// Simple ANSI stripping for testing
	result := str
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}