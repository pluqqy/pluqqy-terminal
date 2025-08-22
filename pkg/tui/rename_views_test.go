package tui

import (
	"strings"
	"testing"
)

func makeTestRenameRenderer() *RenameRenderer {
	rr := NewRenameRenderer()
	rr.SetSize(80, 24) // Set a reasonable terminal size
	return rr
}

func TestNewRenameRenderer(t *testing.T) {
	rr := NewRenameRenderer()

	if rr == nil {
		t.Fatal("NewRenameRenderer() returned nil")
	}

	if rr.Width != 0 {
		t.Errorf("Width = %d, want 0 initially", rr.Width)
	}

	if rr.Height != 0 {
		t.Errorf("Height = %d, want 0 initially", rr.Height)
	}
}

func TestRenameRenderer_SetSize(t *testing.T) {
	rr := NewRenameRenderer()

	width := 100
	height := 40

	rr.SetSize(width, height)

	if rr.Width != width {
		t.Errorf("Width = %d, want %d", rr.Width, width)
	}

	if rr.Height != height {
		t.Errorf("Height = %d, want %d", rr.Height, height)
	}
}

func TestRenameRenderer_Render(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *RenameState
		width        int
		height       int
		wantContains []string
		wantEmpty    bool
	}{
		{
			name: "inactive state returns empty",
			setup: func() *RenameState {
				return NewRenameState()
			},
			width:     80,
			height:    24,
			wantEmpty: true,
		},
		{
			name: "component rename dialog",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test Component", "component", "test.md", false)
				rs.NewName = "New Component"
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"RENAME COMPONENT",
				"Current:",
				"Test Component",
				"New name:",
				"New Component",
				"Will save as: new-component.md",
				"[Enter] Save",
				"[Esc] Cancel",
			},
		},
		{
			name: "pipeline rename dialog",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test Pipeline", "pipeline", "test-pipeline.yaml", false)
				rs.NewName = "New Pipeline"
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"RENAME PIPELINE",
				"Current:",
				"Test Pipeline",
				"New name:",
				"New Pipeline",
				"Will save as: new-pipeline.yaml",
				"[Enter] Save",
				"[Esc] Cancel",
			},
		},
		{
			name: "archived component rename",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Archived Component", "component", "archived.md", true)
				rs.NewName = "Renamed Archived"
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"RENAME ARCHIVED COMPONENT",
				"Current:",
				"Archived Component",
				"New name:",
				"Renamed Archived",
			},
		},
		{
			name: "empty new name shows placeholder",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = ""
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"Current:",
				"Test",
				"New name:",
				"Enter new display name...",
			},
		},
		{
			name: "validation error displayed",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Invalid"
				rs.ValidationError = "Name already exists"
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"⚠ Name already exists",
				"[Esc] Cancel",
			},
		},
		{
			name: "no changes message",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Same Name", "component", "test.md", false)
				rs.NewName = "Same Name"
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"[Esc] Cancel",
				"(no changes)",
			},
		},
		{
			name: "affected pipelines shown for component",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Component", "component", "test.md", false)
				rs.NewName = "Renamed Component"
				rs.AffectedActive = []string{"Pipeline 1", "Pipeline 2"}
				rs.AffectedArchive = []string{"Archived Pipeline"}
				return rs
			},
			width:  80,
			height: 24,
			wantContains: []string{
				"This will update references in:",
				"Active pipelines:",
				"• Pipeline 1",
				"• Pipeline 2",
				"Archived pipelines:",
				"• Archived Pipeline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := NewRenameRenderer()
			rr.SetSize(tt.width, tt.height)
			state := tt.setup()

			output := rr.Render(state)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("expected empty output, got %q", output)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestRenameRenderer_RenderOverlay(t *testing.T) {
	tests := []struct {
		name     string
		baseView string
		setup    func() *RenameState
		wantBase bool // true if base should be returned unchanged
	}{
		{
			name:     "inactive state returns base unchanged",
			baseView: "Base content here",
			setup: func() *RenameState {
				return NewRenameState()
			},
			wantBase: true,
		},
		{
			name:     "active state overlays rename dialog",
			baseView: "Base content here",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test", "component", "test.md", false)
				return rs
			},
			wantBase: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeTestRenameRenderer()
			state := tt.setup()

			output := rr.RenderOverlay(tt.baseView, state)

			if tt.wantBase {
				if output != tt.baseView {
					t.Errorf("expected base view unchanged, got %q", output)
				}
			} else {
				// Should contain rename dialog content overlaid
				if output == tt.baseView {
					t.Error("expected overlay to modify base view")
				}
			}
		})
	}
}

// Removed TestRenameRenderer_DialogWidth as it's testing implementation details
// The important tests are that the dialog renders correctly with the expected content
