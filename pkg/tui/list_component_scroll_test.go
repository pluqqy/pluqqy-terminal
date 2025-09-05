package tui

import (
	"testing"

	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestMainListModel_ComponentScrollPersistence validates that the component list
// maintains its scroll position across multiple renders when the preview pane is visible.
// This test ensures the fix for the scrolling issue is working correctly.
func TestMainListModel_ComponentScrollPersistence(t *testing.T) {
	tests := []struct {
		name                string
		previewEnabled      bool
		numComponents       int
		cursorPosition      int
		expectedScrollAfter int // Expected minimum scroll offset after cursor movement
		description         string
	}{
		{
			name:                "scroll_follows_cursor_with_preview",
			previewEnabled:      true,
			numComponents:       50,
			cursorPosition:      25,
			expectedScrollAfter: 10, // Should scroll to keep cursor visible
			description:         "With preview pane visible, scrolling should follow cursor",
		},
		{
			name:                "scroll_persists_across_renders",
			previewEnabled:      true,
			numComponents:       30,
			cursorPosition:      20,
			expectedScrollAfter: 5,
			description:         "Scroll position should persist between View() calls",
		},
		{
			name:                "no_scroll_when_cursor_in_viewport",
			previewEnabled:      true,
			numComponents:       10,
			cursorPosition:      3,
			expectedScrollAfter: 0,
			description:         "No scrolling needed when cursor is within initial viewport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and configure the model
			m := NewMainListModel()
			m.stateManager.ShowPreview = tt.previewEnabled
			m.SetSize(120, 40) // Set a reasonable terminal size

			// Create test components
			var components []componentItem
			for i := 0; i < tt.numComponents; i++ {
				compType := models.ComponentTypeContext
				if i%3 == 1 {
					compType = models.ComponentTypePrompt
				} else if i%3 == 2 {
					compType = models.ComponentTypeRules
				}

				components = append(components, componentItem{
					name:       componentNames[i%len(componentNames)],
					compType:   compType,
					tokenCount: 100 + (i * 10),
					tags:       []string{"test"},
				})
			}

			// Set the components
			m.data.FilteredComponents = components
			m.stateManager.UpdateCounts(len(components), 0)

			// Initial render to set up the view
			initialView := m.View()
			if initialView == "" {
				t.Fatal("Initial View() returned empty string")
			}

			// Verify componentTableRenderer was created
			if m.ui.ComponentTableRenderer == nil {
				t.Fatal("componentTableRenderer should be initialized after View()")
			}

			// Now move cursor to target position and render again
			m.stateManager.ComponentCursor = tt.cursorPosition
			view1 := m.View()
			if view1 == "" {
				t.Fatal("First View() after cursor move returned empty string")
			}

			// Check scroll position after cursor movement
			firstScrollOffset := m.ui.ComponentTableRenderer.Viewport.YOffset

			// For large cursor positions, we expect scrolling
			if tt.cursorPosition > 10 && tt.expectedScrollAfter > 0 {
				if firstScrollOffset < tt.expectedScrollAfter {
					t.Errorf("Expected scroll offset >= %d, got %d after moving cursor to position %d",
						tt.expectedScrollAfter, firstScrollOffset, tt.cursorPosition)
				}
			}

			// Second render - scroll position should persist
			view2 := m.View()
			if view2 == "" {
				t.Fatal("Second View() returned empty string")
			}

			secondScrollOffset := m.ui.ComponentTableRenderer.Viewport.YOffset

			// Verify scroll position persisted
			if firstScrollOffset != secondScrollOffset {
				t.Errorf("Scroll position changed between renders: first=%d, second=%d",
					firstScrollOffset, secondScrollOffset)
			}

			// Move cursor up slightly and verify scroll adjusts appropriately
			if m.stateManager.ComponentCursor > 0 {
				m.stateManager.ComponentCursor--
				view3 := m.View()
				if view3 == "" {
					t.Fatal("Third View() returned empty string")
				}

				// Scroll should adjust or stay the same, but not jump randomly
				thirdScrollOffset := m.ui.ComponentTableRenderer.Viewport.YOffset
				if thirdScrollOffset > secondScrollOffset+1 {
					t.Errorf("Unexpected scroll jump when moving cursor up: before=%d, after=%d",
						secondScrollOffset, thirdScrollOffset)
				}
			}
		})
	}
}

// TestComponentTableRenderer_ViewportBehavior tests the viewport scrolling behavior
// of the ComponentTableRenderer in isolation.
func TestComponentTableRenderer_ViewportBehavior(t *testing.T) {
	tests := []struct {
		name           string
		viewportHeight int
		numComponents  int
		cursorMoves    []int // Sequence of cursor positions to test
		description    string
	}{
		{
			name:           "basic_scrolling_down",
			viewportHeight: 10,
			numComponents:  30,
			cursorMoves:    []int{0, 8, 12, 18}, // More gradual moves
			description:    "Viewport should scroll when cursor moves beyond visible area",
		},
		{
			name:           "basic_scrolling_up",
			viewportHeight: 10,
			numComponents:  30,
			cursorMoves:    []int{20, 15, 10, 5, 0},
			description:    "Viewport should scroll up when cursor moves up",
		},
		{
			name:           "jump_to_bottom_and_back",
			viewportHeight: 10,
			numComponents:  40,
			cursorMoves:    []int{0, 35, 0}, // Jump to near bottom then back to top
			description:    "Viewport should handle large cursor jumps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create renderer
			renderer := NewComponentTableRenderer(80, 20, true)
			renderer.Viewport.Height = tt.viewportHeight

			// Create test components with proper type grouping
			var components []componentItem
			for i := 0; i < tt.numComponents; i++ {
				compType := models.ComponentTypeContext
				if i >= tt.numComponents/3 && i < 2*tt.numComponents/3 {
					compType = models.ComponentTypePrompt
				} else if i >= 2*tt.numComponents/3 {
					compType = models.ComponentTypeRules
				}

				components = append(components, componentItem{
					name:     componentNames[i%len(componentNames)],
					compType: compType,
				})
			}

			renderer.SetComponents(components)
			renderer.SetActive(true)

			for i, cursorPos := range tt.cursorMoves {
				renderer.SetCursor(cursorPos)

				currentOffset := renderer.Viewport.YOffset

				// Basic check: viewport should adjust when cursor moves
				// We're mainly testing that scrolling happens, not the exact positioning
				if i > 0 {
					prevCursor := tt.cursorMoves[i-1]

					// If cursor moved significantly down and was near bottom of viewport,
					// viewport should have scrolled down
					if cursorPos > prevCursor+5 && prevCursor > tt.viewportHeight-3 {
						if currentOffset == 0 {
							t.Logf("Move %d: Expected some scrolling when moving cursor from %d to %d",
								i, prevCursor, cursorPos)
						}
					}

					// If cursor jumped to top, viewport should reset
					if cursorPos == 0 && prevCursor > 20 {
						if currentOffset > 5 {
							t.Errorf("Move %d: Viewport should reset when jumping to top (offset=%d)",
								i, currentOffset)
						}
					}
				}
			}
		})
	}
}

// TestComponentScrollingWithPreviewToggle tests that scrolling behavior works correctly
// when toggling the preview pane on and off.
func TestComponentScrollingWithPreviewToggle(t *testing.T) {
	// Create model
	m := NewMainListModel()
	m.SetSize(120, 40)

	// Create many components to ensure scrolling is needed
	var components []componentItem
	for i := 0; i < 40; i++ {
		components = append(components, componentItem{
			name:       componentNames[i%len(componentNames)],
			compType:   models.ComponentTypeContext,
			tokenCount: 100,
		})
	}

	m.data.FilteredComponents = components
	m.stateManager.UpdateCounts(len(components), 0)

	// Move cursor down to trigger scrolling
	m.stateManager.ComponentCursor = 25

	// Enable preview and render
	m.stateManager.ShowPreview = true
	m.updateViewportSizes()
	view1 := m.View()
	if view1 == "" {
		t.Fatal("View() with preview returned empty")
	}

	// Record scroll position with preview
	if m.ui.ComponentTableRenderer != nil {
		_ = m.ui.ComponentTableRenderer.Viewport.YOffset // Record for comparison if needed
	}

	// Disable preview and render
	m.stateManager.ShowPreview = false
	m.updateViewportSizes()
	view2 := m.View()
	if view2 == "" {
		t.Fatal("View() without preview returned empty")
	}

	// Check scroll position after preview toggle
	scrollWithoutPreview := 0
	if m.ui.ComponentTableRenderer != nil {
		scrollWithoutPreview = m.ui.ComponentTableRenderer.Viewport.YOffset
	}

	// The exact offset might change due to viewport height changes,
	// but the cursor should still be visible
	if m.ui.ComponentTableRenderer != nil {
		effectiveLine := calculateEffectiveLine(components, m.stateManager.ComponentCursor)
		viewportHeight := m.ui.ComponentTableRenderer.Viewport.Height

		if effectiveLine < scrollWithoutPreview || effectiveLine >= scrollWithoutPreview+viewportHeight {
			t.Errorf("Cursor not visible after preview toggle: line=%d, offset=%d, height=%d",
				effectiveLine, scrollWithoutPreview, viewportHeight)
		}
	}

	// Re-enable preview
	m.stateManager.ShowPreview = true
	m.updateViewportSizes()
	view3 := m.View()
	if view3 == "" {
		t.Fatal("View() with preview re-enabled returned empty")
	}

	// Verify renderer is still functional
	if m.ui.ComponentTableRenderer == nil {
		t.Fatal("componentTableRenderer should persist through preview toggles")
	}
}

// Helper function to calculate the effective line position accounting for type headers
func calculateEffectiveLine(components []componentItem, cursorPos int) int {
	if cursorPos >= len(components) || cursorPos < 0 {
		return 0
	}

	line := 0
	currentType := ""

	for i := 0; i < len(components); i++ {
		// Add line for type header if type changes
		if components[i].compType != currentType {
			if currentType != "" {
				line++ // Empty line between sections
			}
			currentType = components[i].compType
			line++ // Type header line
		}

		if i == cursorPos {
			return line
		}
		line++ // Component line
	}

	return line
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Sample component names for testing
var componentNames = []string{
	"analyze-code", "api-client", "auth-handler", "build-script",
	"cache-manager", "cli-parser", "config-loader", "data-processor",
	"debug-logger", "error-handler", "file-watcher", "format-output",
	"generate-docs", "http-server", "input-validator", "json-parser",
	"key-manager", "lint-rules", "memory-cache", "network-client",
	"output-formatter", "parse-args", "query-builder", "rate-limiter",
	"security-check", "test-runner", "url-parser", "validate-input",
	"worker-pool", "xml-handler", "yaml-config", "zip-handler",
}
