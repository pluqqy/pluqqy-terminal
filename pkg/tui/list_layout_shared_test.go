package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

func TestSharedLayout_NewSharedLayout(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		showPreview bool
		wantColumn  int
		wantContent int
	}{
		{
			name:        "standard dimensions without preview",
			width:       100,
			height:      40,
			showPreview: false,
			wantColumn:  47, // (100 - 6) / 2
			wantContent: 24, // 40 - 16
		},
		{
			name:        "standard dimensions with preview",
			width:       100,
			height:      40,
			showPreview: true,
			wantColumn:  47, // (100 - 6) / 2
			wantContent: 12, // (40 - 16) / 2
		},
		{
			name:        "minimum window size",
			width:       40,
			height:      20,
			showPreview: false,
			wantColumn:  17, // (40 - 6) / 2
			wantContent: 10, // min height enforced
		},
		{
			name:        "minimum window size with preview",
			width:       40,
			height:      20,
			showPreview: true,
			wantColumn:  17, // (40 - 6) / 2
			wantContent: 10, // min height enforced
		},
		{
			name:        "large window",
			width:       200,
			height:      80,
			showPreview: false,
			wantColumn:  97, // (200 - 6) / 2
			wantContent: 64, // 80 - 16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := NewSharedLayout(tt.width, tt.height, tt.showPreview)
			
			if sl == nil {
				t.Fatal("NewSharedLayout returned nil")
			}
			if sl.Width != tt.width {
				t.Errorf("Width = %v, want %v", sl.Width, tt.width)
			}
			if sl.Height != tt.height {
				t.Errorf("Height = %v, want %v", sl.Height, tt.height)
			}
			if sl.ShowPreview != tt.showPreview {
				t.Errorf("ShowPreview = %v, want %v", sl.ShowPreview, tt.showPreview)
			}
			if got := sl.GetColumnWidth(); got != tt.wantColumn {
				t.Errorf("GetColumnWidth() = %v, want %v", got, tt.wantColumn)
			}
			if got := sl.GetContentHeight(); got != tt.wantContent {
				t.Errorf("GetContentHeight() = %v, want %v", got, tt.wantContent)
			}
		})
	}
}

func TestSharedLayout_SetSize(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	// Initial values
	if got := sl.GetColumnWidth(); got != 47 {
		t.Errorf("Initial GetColumnWidth() = %v, want 47", got)
	}
	if got := sl.GetContentHeight(); got != 24 {
		t.Errorf("Initial GetContentHeight() = %v, want 24", got)
	}
	
	// Update size
	sl.SetSize(120, 50)
	if sl.Width != 120 {
		t.Errorf("Width after SetSize = %v, want 120", sl.Width)
	}
	if sl.Height != 50 {
		t.Errorf("Height after SetSize = %v, want 50", sl.Height)
	}
	if got := sl.GetColumnWidth(); got != 57 { // (120 - 6) / 2
		t.Errorf("GetColumnWidth() after SetSize = %v, want 57", got)
	}
	if got := sl.GetContentHeight(); got != 34 { // 50 - 16
		t.Errorf("GetContentHeight() after SetSize = %v, want 34", got)
	}
}

func TestSharedLayout_SetShowPreview(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	// Initial without preview
	if got := sl.GetContentHeight(); got != 24 {
		t.Errorf("Initial GetContentHeight() = %v, want 24", got)
	}
	
	// Enable preview
	sl.SetShowPreview(true)
	if !sl.ShowPreview {
		t.Error("ShowPreview should be true after SetShowPreview(true)")
	}
	if got := sl.GetContentHeight(); got != 12 { // Should be halved
		t.Errorf("GetContentHeight() with preview = %v, want 12", got)
	}
	
	// Disable preview
	sl.SetShowPreview(false)
	if sl.ShowPreview {
		t.Error("ShowPreview should be false after SetShowPreview(false)")
	}
	if got := sl.GetContentHeight(); got != 24 { // Should be restored
		t.Errorf("GetContentHeight() without preview = %v, want 24", got)
	}
}

func TestSharedLayout_RenderHelpPane(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	tests := []struct {
		name     string
		helpRows [][]string
		validate func(t *testing.T, result string)
	}{
		{
			name: "single row help",
			helpRows: [][]string{
				{"tab switch", "esc exit", "enter select"},
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "tab switch") {
					t.Error("Result should contain 'tab switch'")
				}
				if !strings.Contains(result, "esc exit") {
					t.Error("Result should contain 'esc exit'")
				}
				if !strings.Contains(result, "enter select") {
					t.Error("Result should contain 'enter select'")
				}
				// Should have borders
				if !strings.Contains(result, "─") {
					t.Error("Result should contain border character '─'")
				}
				if !strings.Contains(result, "│") {
					t.Error("Result should contain border character '│'")
				}
			},
		},
		{
			name: "multi row help",
			helpRows: [][]string{
				{"/ search", "tab switch", "↑↓ nav"},
				{"n new", "e edit", "d delete"},
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "/ search") {
					t.Error("Result should contain '/ search'")
				}
				if !strings.Contains(result, "n new") {
					t.Error("Result should contain 'n new'")
				}
				if !strings.Contains(result, "e edit") {
					t.Error("Result should contain 'e edit'")
				}
				// Should have borders
				if !strings.Contains(result, "─") {
					t.Error("Result should contain border character '─'")
				}
			},
		},
		{
			name:     "empty help",
			helpRows: [][]string{},
			validate: func(t *testing.T, result string) {
				// Should still have borders
				if !strings.Contains(result, "─") {
					t.Error("Result should contain border character '─'")
				}
				if !strings.Contains(result, "│") {
					t.Error("Result should contain border character '│'")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sl.RenderHelpPane(tt.helpRows)
			tt.validate(t, result)
		})
	}
}

func TestSharedLayout_RenderHeader(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	tests := []struct {
		name      string
		heading   string
		active    bool
		badge     string
		width     int
		validate  func(t *testing.T, result string)
	}{
		{
			name:    "active header without badge",
			heading: "COMPONENTS",
			active:  true,
			badge:   "",
			width:   50,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "COMPONENTS") {
					t.Error("Result should contain 'COMPONENTS'")
				}
				if !strings.Contains(result, ":") {
					t.Error("Result should contain ':'")
				}
				// Should have colons filling the space
				colonCount := strings.Count(result, ":")
				if colonCount <= 10 {
					t.Errorf("Should have more than 10 colons, got %v", colonCount)
				}
			},
		},
		{
			name:    "inactive header with badge",
			heading: "PREVIEW",
			active:  false,
			badge:   "1.2k",
			width:   50,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "PREVIEW") {
					t.Error("Result should contain 'PREVIEW'")
				}
				if !strings.Contains(result, "1.2k") {
					t.Error("Result should contain '1.2k'")
				}
				if !strings.Contains(result, ":") {
					t.Error("Result should contain ':'")
				}
			},
		},
		{
			name:    "long heading",
			heading: "THIS IS A VERY LONG HEADING THAT TAKES UP SPACE",
			active:  true,
			badge:   "100",
			width:   60,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "THIS IS A VERY LONG HEADING") {
					t.Error("Result should contain the heading")
				}
				if !strings.Contains(result, "100") {
					t.Error("Result should contain '100'")
				}
				// Should still have at least minimum colons
				if !strings.Contains(result, ":::") {
					t.Error("Result should contain at least ':::'")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sl.RenderHeader(tt.heading, tt.active, tt.badge, tt.width)
			tt.validate(t, result)
		})
	}
}

func TestSharedLayout_RenderSearchBar(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	searchBar := NewSearchBar()
	
	result := sl.RenderSearchBar(searchBar)
	
	// Should set width and render
	if result == "" {
		t.Error("RenderSearchBar should return non-empty result")
	}
	if searchBar.width != 100 {
		t.Errorf("SearchBar width = %v, want 100", searchBar.width)
	}
}

func TestSharedLayout_RenderPreviewPane(t *testing.T) {
	sl := NewSharedLayout(100, 40, true)
	
	tests := []struct {
		name     string
		config   PreviewConfig
		validate func(t *testing.T, result string)
	}{
		{
			name: "active preview pane",
			config: PreviewConfig{
				Content:     "This is preview content\nLine 2\nLine 3",
				Heading:     "PIPELINE PREVIEW (test.yaml)",
				ActivePane:  previewPane,
				PreviewPane: previewPane,
				Viewport:    viewport.New(80, 10),
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "PIPELINE PREVIEW (test.yaml)") {
					t.Error("Result should contain preview heading")
				}
				if !strings.Contains(result, "This is preview content") {
					t.Error("Result should contain preview content")
				}
				// Should have borders
				if !strings.Contains(result, "─") {
					t.Error("Result should contain border character '─'")
				}
			},
		},
		{
			name: "inactive preview with column type",
			config: PreviewConfig{
				Content:     "Component content here",
				Heading:     "COMPONENT PREVIEW (test.md)",
				ActivePane:  leftColumn,
				PreviewPane: previewColumn,
				Viewport:    viewport.New(80, 10),
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "COMPONENT PREVIEW (test.md)") {
					t.Error("Result should contain component heading")
				}
				if !strings.Contains(result, "Component content here") {
					t.Error("Result should contain component content")
				}
			},
		},
		{
			name: "empty content",
			config: PreviewConfig{
				Content:     "",
				Heading:     "PREVIEW",
				ActivePane:  pipelinesPane,
				PreviewPane: previewPane,
				Viewport:    viewport.New(80, 10),
			},
			validate: func(t *testing.T, result string) {
				// Should return empty string for empty content
				if result != "" {
					t.Error("Result should be empty for empty content")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sl.RenderPreviewPane(tt.config)
			tt.validate(t, result)
		})
	}
}

func TestSharedLayout_RenderColumnHeader(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	tests := []struct {
		name     string
		config   ColumnHeaderConfig
		validate func(t *testing.T, result string)
	}{
		{
			name: "active column",
			config: ColumnHeaderConfig{
				Heading:     "AVAILABLE COMPONENTS",
				Active:      true,
				ColumnWidth: 47,
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "AVAILABLE COMPONENTS") {
					t.Error("Result should contain 'AVAILABLE COMPONENTS'")
				}
				if !strings.Contains(result, ":") {
					t.Error("Result should contain ':'")
				}
			},
		},
		{
			name: "inactive column",
			config: ColumnHeaderConfig{
				Heading:     "PIPELINE COMPONENTS",
				Active:      false,
				ColumnWidth: 47,
			},
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "PIPELINE COMPONENTS") {
					t.Error("Result should contain 'PIPELINE COMPONENTS'")
				}
				if !strings.Contains(result, ":") {
					t.Error("Result should contain ':'")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sl.RenderColumnHeader(tt.config)
			tt.validate(t, result)
		})
	}
}

func TestSharedLayout_BuildConfirmationDialog(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	t.Run("active confirmation", func(t *testing.T) {
		confirm := NewConfirmation()
		confirm.ShowInline("Delete this?", true, nil, nil)
		
		result := sl.BuildConfirmationDialog(confirm, "Delete this?", true)
		// Debug output to see what we're getting
		t.Logf("Confirmation dialog result: %q", result)
		
		if !strings.Contains(result, "Delete this?") {
			t.Error("Result should contain 'Delete this?'")
		}
		// The result should not be empty for an active confirmation
		if len(result) < 10 {
			t.Errorf("Result seems too short for an active confirmation: %q", result)
		}
	})
	
	t.Run("inactive confirmation", func(t *testing.T) {
		confirm := NewConfirmation()
		
		result := sl.BuildConfirmationDialog(confirm, "Delete this?", true)
		if result != "" {
			t.Error("Result should be empty for inactive confirmation")
		}
	})
}

func TestSharedLayout_EdgeCases(t *testing.T) {
	t.Run("very small dimensions", func(t *testing.T) {
		sl := NewSharedLayout(20, 10, true)
		// Should handle gracefully with minimum values
		if got := sl.GetColumnWidth(); got != 7 { // (20 - 6) / 2
			t.Errorf("GetColumnWidth() = %v, want 7", got)
		}
		if got := sl.GetContentHeight(); got != 10 { // minimum enforced
			t.Errorf("GetContentHeight() = %v, want 10", got)
		}
	})
	
	t.Run("odd width", func(t *testing.T) {
		sl := NewSharedLayout(101, 40, false)
		// Should handle odd numbers correctly
		if got := sl.GetColumnWidth(); got != 47 { // (101 - 6) / 2 = 47 (integer division)
			t.Errorf("GetColumnWidth() = %v, want 47", got)
		}
	})
	
	t.Run("rapid dimension changes", func(t *testing.T) {
		sl := NewSharedLayout(100, 40, false)
		
		// Simulate rapid resizing
		for i := 50; i <= 150; i += 10 {
			sl.SetSize(i, i/2)
			// Should always maintain valid dimensions
			if got := sl.GetColumnWidth(); got <= 0 {
				t.Errorf("GetColumnWidth() = %v, should be > 0", got)
			}
			if got := sl.GetContentHeight(); got < 10 {
				t.Errorf("GetContentHeight() = %v, should be >= 10", got)
			}
		}
	})
}

func BenchmarkSharedLayout_RenderHelpPane(b *testing.B) {
	sl := NewSharedLayout(100, 40, false)
	helpRows := [][]string{
		{"/ search", "tab switch", "↑↓ nav", "p preview", "s settings"},
		{"n new", "e edit", "d delete", "r rename", "c clone"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.RenderHelpPane(helpRows)
	}
}

func BenchmarkSharedLayout_RenderPreviewPane(b *testing.B) {
	sl := NewSharedLayout(100, 40, true)
	config := PreviewConfig{
		Content:     strings.Repeat("This is a line of preview content.\n", 100),
		Heading:     "PREVIEW (test.yaml)",
		ActivePane:  previewPane,
		PreviewPane: previewPane,
		Viewport:    viewport.New(80, 20),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.RenderPreviewPane(config)
	}
}

func TestSharedLayout_CachedValues(t *testing.T) {
	sl := NewSharedLayout(100, 40, false)
	
	// Get initial values
	width1 := sl.GetColumnWidth()
	height1 := sl.GetContentHeight()
	
	// Call multiple times without changes
	width2 := sl.GetColumnWidth()
	height2 := sl.GetContentHeight()
	width3 := sl.GetColumnWidth()
	height3 := sl.GetContentHeight()
	
	// Should return same cached values
	if width1 != width2 || width1 != width3 {
		t.Errorf("GetColumnWidth() should return cached value: %v, %v, %v", width1, width2, width3)
	}
	if height1 != height2 || height1 != height3 {
		t.Errorf("GetContentHeight() should return cached value: %v, %v, %v", height1, height2, height3)
	}
	
	// Change size
	sl.SetSize(120, 50)
	
	// Should return new values
	width4 := sl.GetColumnWidth()
	height4 := sl.GetContentHeight()
	
	if width1 == width4 {
		t.Error("GetColumnWidth() should return new value after SetSize")
	}
	if height1 == height4 {
		t.Error("GetContentHeight() should return new value after SetSize")
	}
}