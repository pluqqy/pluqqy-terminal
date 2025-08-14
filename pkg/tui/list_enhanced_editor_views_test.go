package tui

import (
	"strings"
	"testing"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestNewEnhancedEditorRenderer(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		validate func(t *testing.T, renderer *EnhancedEditorRenderer)
	}{
		{
			name:   "creates renderer with normal dimensions",
			width:  80,
			height: 24,
			validate: func(t *testing.T, renderer *EnhancedEditorRenderer) {
				if renderer.Width != 80 {
					t.Errorf("Expected Width to be 80, got %d", renderer.Width)
				}
				if renderer.Height != 24 {
					t.Errorf("Expected Height to be 24, got %d", renderer.Height)
				}
			},
		},
		{
			name:   "creates renderer with large dimensions",
			width:  200,
			height: 60,
			validate: func(t *testing.T, renderer *EnhancedEditorRenderer) {
				if renderer.Width != 200 {
					t.Errorf("Expected Width to be 200, got %d", renderer.Width)
				}
				if renderer.Height != 60 {
					t.Errorf("Expected Height to be 60, got %d", renderer.Height)
				}
			},
		},
		{
			name:   "creates renderer with small dimensions",
			width:  40,
			height: 10,
			validate: func(t *testing.T, renderer *EnhancedEditorRenderer) {
				if renderer.Width != 40 {
					t.Errorf("Expected Width to be 40, got %d", renderer.Width)
				}
				if renderer.Height != 10 {
					t.Errorf("Expected Height to be 10, got %d", renderer.Height)
				}
			},
		},
		{
			name:   "handles zero dimensions",
			width:  0,
			height: 0,
			validate: func(t *testing.T, renderer *EnhancedEditorRenderer) {
				if renderer.Width != 0 {
					t.Errorf("Expected Width to be 0, got %d", renderer.Width)
				}
				if renderer.Height != 0 {
					t.Errorf("Expected Height to be 0, got %d", renderer.Height)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, tt.height)
			if renderer == nil {
				t.Fatal("Expected renderer to be created, got nil")
			}
			tt.validate(t, renderer)
		})
	}
}

func TestEnhancedEditorRenderer_Render_InactiveState(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected string
	}{
		{
			name:     "returns empty string when state inactive",
			width:    80,
			height:   24,
			expected: "",
		},
		{
			name:     "returns empty string with large dimensions",
			width:    200,
			height:   60,
			expected: "",
		},
		{
			name:     "returns empty string with small dimensions",
			width:    20,
			height:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, tt.height)
			state := NewEnhancedEditorState() // Inactive by default
			
			result := renderer.Render(state)
			if result != tt.expected {
				t.Errorf("Expected empty string for inactive state, got %q", result)
			}
		})
	}
}

func TestEnhancedEditorRenderer_Render_NormalMode(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		componentName  string
		componentType  string
		content        string
		tags           []string
		hasUnsaved     bool
		validateOutput func(t *testing.T, output string)
	}{
		{
			name:          "renders normal editing mode",
			width:         80,
			height:        24,
			componentName: "Test Component",
			componentType: "prompt",
			content:       "# Test Component\nSome content here",
			tags:          []string{"tag1", "tag2"},
			hasUnsaved:    false,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "EDITING: Test Component") {
					t.Error("Expected header to contain 'EDITING: Test Component'")
				}
				if !strings.Contains(output, "○ Saved") {
					t.Error("Expected status bar to show saved status")
				}
				if !strings.Contains(output, "^s save") {
					t.Error("Expected help text to contain save shortcut")
				}
				if !strings.Contains(output, "@ insert file ref") {
					t.Error("Expected help text to contain file reference shortcut")
				}
			},
		},
		{
			name:          "renders with unsaved changes",
			width:         80,
			height:        24,
			componentName: "Modified Config",
			componentType: "context",
			content:       "modified content",
			tags:          []string{},
			hasUnsaved:    true,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "EDITING: Modified Config") {
					t.Error("Expected header to contain 'EDITING: Modified Config'")
				}
				if !strings.Contains(output, "● Modified") {
					t.Error("Expected status bar to show modified status")
				}
			},
		},
		{
			name:          "renders with long component name",
			width:         80,
			height:        24,
			componentName: "Very Long Component Name That Might Overflow",
			componentType: "rule",
			content:       "",
			tags:          []string{"long", "name", "test"},
			hasUnsaved:    false,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Very Long Component Name That Might Overflow") {
					t.Error("Expected header to contain full component name")
				}
			},
		},
		{
			name:          "renders with small dimensions",
			width:         40,
			height:        10,
			componentName: "Small Component",
			componentType: "prompt",
			content:       "small",
			tags:          []string{},
			hasUnsaved:    false,
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "EDITING: Small Component") {
					t.Error("Expected header to contain component name even with small dimensions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, tt.height)
			state := NewEnhancedEditorState()
			
			// Set up the state
			state.StartEditing("path/"+tt.componentName, tt.componentName, tt.componentType, tt.content, tt.tags)
			if tt.hasUnsaved {
				state.SetContent(tt.content + " modified")
			}
			
			result := renderer.Render(state)
			
			// Basic sanity checks
			if len(result) == 0 {
				t.Error("Expected non-empty render output")
			}
			
			tt.validateOutput(t, result)
		})
	}
}

func TestEnhancedEditorRenderer_Render_FilePickerMode(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		componentName  string
		validateOutput func(t *testing.T, output string)
	}{
		{
			name:          "renders file picker mode",
			width:         80,
			height:        24,
			componentName: "Test Component",
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "SELECT FILE REFERENCE") {
					t.Error("Expected header to contain 'SELECT FILE REFERENCE'")
				}
				if !strings.Contains(output, "Current:") {
					t.Error("Expected to show current directory")
				}
				if !strings.Contains(output, "enter select") {
					t.Error("Expected help text to contain selection shortcut")
				}
				if !strings.Contains(output, "esc cancel") {
					t.Error("Expected help text to contain cancel shortcut")
				}
			},
		},
		{
			name:          "renders file picker with small dimensions",
			width:         40,
			height:        12,
			componentName: "small.md",
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "SELECT FILE REFERENCE") {
					t.Error("Expected header even with small dimensions")
				}
			},
		},
		{
			name:          "renders file picker with large dimensions",
			width:         120,
			height:        40,
			componentName: "large.yaml",
			validateOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "SELECT FILE REFERENCE") {
					t.Error("Expected header with large dimensions")
				}
				if !strings.Contains(output, "↑/↓ navigate") {
					t.Error("Expected navigation help text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, tt.height)
			state := NewEnhancedEditorState()
			
			// Set up the state for file picking
			state.StartEditing("path/"+tt.componentName, tt.componentName, "prompt", "content", []string{})
			state.StartFilePicker()
			
			result := renderer.Render(state)
			
			// Basic sanity checks
			if len(result) == 0 {
				t.Error("Expected non-empty render output for file picker mode")
			}
			
			tt.validateOutput(t, result)
		})
	}
}

func TestEnhancedEditorRenderer_renderHeader(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		componentName string
		expected      []string // Strings that should be present in the header
		notExpected   []string // Strings that should NOT be present
	}{
		{
			name:          "renders basic header",
			width:         80,
			componentName: "test.md",
			expected:      []string{"EDITING:", "test.md"},
			notExpected:   []string{},
		},
		{
			name:          "renders header with special characters",
			width:         80,
			componentName: "Test File With Special@Chars",
			expected:      []string{"EDITING:", "Test File With Special@Chars"},
			notExpected:   []string{},
		},
		{
			name:          "renders header with short name",
			width:         120,
			componentName: "A",
			expected:      []string{"EDITING:", "A", ":"},
			notExpected:   []string{},
		},
		{
			name:          "handles empty component name",
			width:         80,
			componentName: "",
			expected:      []string{"EDITING:"},
			notExpected:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, 24)
			state := NewEnhancedEditorState()
			state.StartEditing("path/"+tt.componentName, tt.componentName, "prompt", "content", []string{})
			
			result := renderer.renderHeader(state)
			
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected header to contain %q, but it didn't. Header: %q", expected, result)
				}
			}
			
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected header NOT to contain %q, but it did. Header: %q", notExpected, result)
				}
			}
		})
	}
}

func TestEnhancedEditorRenderer_renderStatusBar(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		hasUnsaved     bool
		expectedTexts  []string
	}{
		{
			name:          "renders saved status",
			width:         80,
			hasUnsaved:    false,
			expectedTexts: []string{"○ Saved"},
		},
		{
			name:          "renders modified status",
			width:         80,
			hasUnsaved:    true,
			expectedTexts: []string{"● Modified"},
		},
		{
			name:          "renders saved status with small width",
			width:         40,
			hasUnsaved:    false,
			expectedTexts: []string{"○ Saved"},
		},
		{
			name:          "renders modified status with large width",
			width:         120,
			hasUnsaved:    true,
			expectedTexts: []string{"● Modified"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, 24)
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "original content", []string{})
			
			if tt.hasUnsaved {
				state.SetContent("modified content")
			}
			
			result := renderer.renderStatusBar(state)
			
			for _, expectedText := range tt.expectedTexts {
				if !strings.Contains(result, expectedText) {
					t.Errorf("Expected status bar to contain %q, got: %q", expectedText, result)
				}
			}
		})
	}
}

func TestEnhancedEditorRenderer_renderHelpPane(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		mode           EditorMode
		expectedHelp   []string
		unexpectedHelp []string
	}{
		{
			name:         "renders normal mode help",
			width:        80,
			mode:         EditorModeNormal,
			expectedHelp: []string{"^s save", "^x external", "esc cancel", "@ insert file ref", "\\@ literal @"},
			unexpectedHelp: []string{"enter select", "↑/↓ navigate"},
		},
		{
			name:           "renders file picker mode help",
			width:          80,
			mode:           EditorModeFilePicking,
			expectedHelp:   []string{"↑/↓ navigate", "enter select", "esc cancel", "tab toggle hidden"},
			unexpectedHelp: []string{"^s save", "@ insert file ref"},
		},
		{
			name:         "renders help with small width",
			width:        40,
			mode:         EditorModeNormal,
			expectedHelp: []string{"^s save", "^x external", "esc", "cancel"},
			unexpectedHelp: []string{},
		},
		{
			name:         "renders help with large width",
			width:        120,
			mode:         EditorModeFilePicking,
			expectedHelp: []string{"↑/↓ navigate", "enter select", "esc cancel"},
			unexpectedHelp: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, 24)
			
			result := renderer.renderHelpPane(tt.mode)
			
			for _, expected := range tt.expectedHelp {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected help to contain %q, got: %q", expected, result)
				}
			}
			
			for _, unexpected := range tt.unexpectedHelp {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected help NOT to contain %q, got: %q", unexpected, result)
				}
			}
		})
	}
}

func TestEnhancedEditorRenderer_renderTextarea(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		content        string
		validateOutput func(t *testing.T, output string, state *EnhancedEditorState)
	}{
		{
			name:    "renders textarea with content",
			width:   80,
			height:  24,
			content: "# Test Content\nLine 2\nLine 3",
			validateOutput: func(t *testing.T, output string, state *EnhancedEditorState) {
				// The output should contain the rendered textarea
				if len(output) == 0 {
					t.Error("Expected non-empty textarea render")
				}
				// Verify dimensions were set on the textarea
				// Note: We can't directly check the textarea dimensions as they're private
				// but we can verify the method doesn't panic and produces output
			},
		},
		{
			name:    "renders textarea with empty content",
			width:   60,
			height:  20,
			content: "",
			validateOutput: func(t *testing.T, output string, state *EnhancedEditorState) {
				if len(output) == 0 {
					t.Error("Expected non-empty textarea render even with empty content")
				}
			},
		},
		{
			name:    "renders textarea with small dimensions",
			width:   40,
			height:  10,
			content: "Small content",
			validateOutput: func(t *testing.T, output string, state *EnhancedEditorState) {
				if len(output) == 0 {
					t.Error("Expected non-empty textarea render with small dimensions")
				}
			},
		},
		{
			name:    "renders textarea with large dimensions",
			width:   150,
			height:  50,
			content: "Large content area\nWith multiple lines\nTo test large dimensions",
			validateOutput: func(t *testing.T, output string, state *EnhancedEditorState) {
				if len(output) == 0 {
					t.Error("Expected non-empty textarea render with large dimensions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, tt.height)
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", tt.content, []string{})
			
			result := renderer.renderTextarea(state)
			
			tt.validateOutput(t, result, state)
		})
	}
}

func TestEnhancedEditorRenderer_renderFileIcon(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{
			name:     "markdown file icon",
			fileName: "readme.md",
			expected: "📝",
		},
		{
			name:     "yaml file icon",
			fileName: "config.yaml",
			expected: "⚙️",
		},
		{
			name:     "yml file icon",
			fileName: "docker-compose.yml",
			expected: "⚙️",
		},
		{
			name:     "json file icon",
			fileName: "package.json",
			expected: "📋",
		},
		{
			name:     "text file icon",
			fileName: "notes.txt",
			expected: "📄",
		},
		{
			name:     "hidden file icon",
			fileName: ".gitignore",
			expected: "👻",
		},
		{
			name:     "unknown file icon",
			fileName: "script.py",
			expected: "📄",
		},
		{
			name:     "file without extension",
			fileName: "LICENSE",
			expected: "📄",
		},
		{
			name:     "empty filename",
			fileName: "",
			expected: "📄",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(80, 24)
			
			result := renderer.renderFileIcon(tt.fileName)
			if result != tt.expected {
				t.Errorf("Expected icon %q for file %q, got %q", tt.expected, tt.fileName, result)
			}
		})
	}
}

func TestEnhancedEditorRenderer_renderBreadcrumbs(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		path     string
		expected []string
	}{
		{
			name:     "renders simple path",
			width:    80,
			path:     "components/prompts/test.md",
			expected: []string{"components", "›", "prompts", "›", "test.md"},
		},
		{
			name:     "renders root path",
			width:    80,
			path:     "file.md",
			expected: []string{"file.md"},
		},
		{
			name:     "renders nested path",
			width:    100,
			path:     "a/b/c/d/e/file.txt",
			expected: []string{"a", "›", "b", "›", "c", "›", "d", "›", "e", "›", "file.txt"},
		},
		{
			name:     "handles empty path",
			width:    80,
			path:     "",
			expected: []string{},
		},
		{
			name:     "handles path with no separators",
			width:    60,
			path:     "filename",
			expected: []string{"filename"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewEnhancedEditorRenderer(tt.width, 24)
			
			result := renderer.renderBreadcrumbs(tt.path)
			
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected breadcrumbs to contain %q for path %q, got: %q", expected, tt.path, result)
				}
			}
		})
	}
}

// Edge case tests
func TestEnhancedEditorRenderer_EdgeCases(t *testing.T) {
	t.Run("handles nil state gracefully", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(80, 24)
		
		// This should not panic - testing robustness
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Renderer should handle edge cases gracefully, but panicked: %v", r)
			}
		}()
		
		// Note: We can't actually pass nil to Render() as it expects *EnhancedEditorState
		// But we can test with an uninitialized state
		state := &EnhancedEditorState{}
		result := renderer.Render(state)
		
		if result != "" {
			t.Error("Expected empty result for uninitialized state")
		}
	})

	t.Run("handles very small dimensions", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(5, 3)
		state := NewEnhancedEditorState()
		state.StartEditing("test.md", "test", "prompt", "content", []string{})
		
		// Should not panic
		result := renderer.Render(state)
		if len(result) == 0 {
			t.Error("Expected some output even with very small dimensions")
		}
	})

	t.Run("handles very large content", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(80, 24)
		state := NewEnhancedEditorState()
		
		// Create very large content
		largeContent := strings.Repeat("This is a very long line that should be handled properly by the renderer.\n", 1000)
		state.StartEditing("large.md", "large", "prompt", largeContent, []string{})
		
		// Should not panic or freeze
		result := renderer.Render(state)
		if len(result) == 0 {
			t.Error("Expected output even with very large content")
		}
	})

	t.Run("mode transitions are handled correctly", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(80, 24)
		state := NewEnhancedEditorState()
		state.StartEditing("test.md", "test", "prompt", "content", []string{})
		
		// Normal mode
		normalResult := renderer.Render(state)
		if !strings.Contains(normalResult, "EDITING:") {
			t.Error("Expected normal mode to show editing header")
		}
		
		// Switch to file picker mode
		state.StartFilePicker()
		pickerResult := renderer.Render(state)
		if !strings.Contains(pickerResult, "SELECT FILE REFERENCE") {
			t.Error("Expected file picker mode to show file selection header")
		}
		
		// Switch back to normal mode
		state.StopFilePicker()
		backToNormalResult := renderer.Render(state)
		if !strings.Contains(backToNormalResult, "EDITING:") {
			t.Error("Expected normal mode to show editing header after returning from file picker")
		}
	})
}

// Performance and consistency tests
func TestEnhancedEditorRenderer_Consistency(t *testing.T) {
	t.Run("consistent output for same state", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(80, 24)
		state := NewEnhancedEditorState()
		state.StartEditing("test.md", "test", "prompt", "content", []string{})
		
		// Render multiple times and ensure consistency
		result1 := renderer.Render(state)
		result2 := renderer.Render(state)
		result3 := renderer.Render(state)
		
		if result1 != result2 {
			t.Error("Expected consistent rendering output")
		}
		if result2 != result3 {
			t.Error("Expected consistent rendering output")
		}
	})

	t.Run("different states produce different output", func(t *testing.T) {
		renderer := NewEnhancedEditorRenderer(80, 24)
		
		state1 := NewEnhancedEditorState()
		state1.StartEditing("test1.md", "test1", "prompt", "content1", []string{})
		
		state2 := NewEnhancedEditorState()
		state2.StartEditing("test2.md", "test2", "context", "content2", []string{})
		
		result1 := renderer.Render(state1)
		result2 := renderer.Render(state2)
		
		// Both should produce non-empty output since states are active
		if len(result1) == 0 {
			t.Error("Expected non-empty output for active state1")
		}
		if len(result2) == 0 {
			t.Error("Expected non-empty output for active state2")
		}
		
		if result1 == result2 {
			t.Error("Expected different output for different states")
		}
		
		// Verify they contain their respective component names
		if !strings.Contains(result1, "test1") {
			t.Errorf("Expected result1 to contain test1, got: %q", result1[:min(200, len(result1))])
		}
		if !strings.Contains(result2, "test2") {
			t.Errorf("Expected result2 to contain test2, got: %q", result2[:min(200, len(result2))])
		}
	})
}