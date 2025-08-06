package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTagReloader(t *testing.T) {
	tr := NewTagReloader()
	
	if tr == nil {
		t.Fatal("NewTagReloader returned nil")
	}
	
	if tr.Active {
		t.Error("New tag reloader should not be active")
	}
	
	if tr.IsReloading {
		t.Error("New tag reloader should not be reloading")
	}
	
	if tr.TagsFound == nil {
		t.Error("TagsFound map should be initialized")
	}
	
	if len(tr.TagsFound) != 0 {
		t.Error("TagsFound should be empty initially")
	}
}

func TestTagReloader_Start(t *testing.T) {
	tr := NewTagReloader()
	
	cmd := tr.Start()
	
	if !tr.Active {
		t.Error("Tag reloader should be active after Start")
	}
	
	if !tr.IsReloading {
		t.Error("Tag reloader should be reloading after Start")
	}
	
	if tr.ReloadResult != nil {
		t.Error("ReloadResult should be nil when starting")
	}
	
	if tr.LastError != nil {
		t.Error("LastError should be nil when starting")
	}
	
	if cmd == nil {
		t.Error("Start should return a command")
	}
}

func TestTagReloader_HandleMessage(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *TagReloader
		msg         tea.Msg
		wantHandled bool
		wantCmd     bool
		checkState  func(*testing.T, *TagReloader)
	}{
		{
			name: "successful reload message",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.IsReloading = true
				return tr
			},
			msg: TagReloadMsg{
				Result: &TagReloadResult{
					ComponentsScanned: 10,
					PipelinesScanned:  5,
					TotalTags:        15,
					NewTags:          []string{"new-tag-1", "new-tag-2"},
				},
				Error: nil,
			},
			wantHandled: true,
			wantCmd:     true,
			checkState: func(t *testing.T, tr *TagReloader) {
				if tr.IsReloading {
					t.Error("Should not be reloading after successful message")
				}
				if tr.ReloadResult == nil {
					t.Error("ReloadResult should be set")
				}
				if tr.LastError != nil {
					t.Error("LastError should be nil for successful reload")
				}
				if tr.ComponentsProcessed != 10 {
					t.Errorf("ComponentsProcessed = %d, want 10", tr.ComponentsProcessed)
				}
				if tr.PipelinesProcessed != 5 {
					t.Errorf("PipelinesProcessed = %d, want 5", tr.PipelinesProcessed)
				}
			},
		},
		{
			name: "failed reload message",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.IsReloading = true
				return tr
			},
			msg: TagReloadMsg{
				Result: &TagReloadResult{
					ComponentsScanned: 3,
					PipelinesScanned:  1,
				},
				Error: errors.New("failed to load registry"),
			},
			wantHandled: true,
			wantCmd:     false,
			checkState: func(t *testing.T, tr *TagReloader) {
				if tr.IsReloading {
					t.Error("Should not be reloading after error message")
				}
				if tr.LastError == nil {
					t.Error("LastError should be set")
				}
				if !strings.Contains(tr.LastError.Error(), "failed to load registry") {
					t.Errorf("Unexpected error: %v", tr.LastError)
				}
			},
		},
		{
			name: "unrelated message",
			setup: func() *TagReloader {
				return NewTagReloader()
			},
			msg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: false,
			wantCmd:     false,
			checkState: func(t *testing.T, tr *TagReloader) {
				// State should be unchanged
				if tr.Active {
					t.Error("Should not be active")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.setup()
			handled, cmd := tr.HandleMessage(tt.msg)
			
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			
			if (cmd != nil) != tt.wantCmd {
				t.Errorf("cmd returned = %v, want cmd = %v", cmd != nil, tt.wantCmd)
			}
			
			if tt.checkState != nil {
				tt.checkState(t, tr)
			}
		})
	}
}

func TestTagReloader_HandleComplete(t *testing.T) {
	tr := NewTagReloader()
	tr.Active = true
	tr.IsReloading = true
	
	tr.HandleComplete()
	
	if tr.Active {
		t.Error("Should not be active after HandleComplete")
	}
	
	if tr.IsReloading {
		t.Error("Should not be reloading after HandleComplete")
	}
}

func TestTagReloader_Reset(t *testing.T) {
	tr := NewTagReloader()
	tr.Active = true
	tr.IsReloading = true
	tr.ComponentsProcessed = 10
	tr.PipelinesProcessed = 5
	tr.TagsFound["test-tag"] = 3
	tr.ReloadResult = &TagReloadResult{TotalTags: 15}
	tr.LastError = errors.New("test error")
	
	tr.Reset()
	
	if tr.Active {
		t.Error("Should not be active after Reset")
	}
	
	if tr.IsReloading {
		t.Error("Should not be reloading after Reset")
	}
	
	if tr.ComponentsProcessed != 0 {
		t.Error("ComponentsProcessed should be 0 after Reset")
	}
	
	if tr.PipelinesProcessed != 0 {
		t.Error("PipelinesProcessed should be 0 after Reset")
	}
	
	if len(tr.TagsFound) != 0 {
		t.Error("TagsFound should be empty after Reset")
	}
	
	if tr.ReloadResult != nil {
		t.Error("ReloadResult should be nil after Reset")
	}
	
	if tr.LastError != nil {
		t.Error("LastError should be nil after Reset")
	}
}

func TestTagReloader_GetStatus(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *TagReloader
		wantEmpty bool
		wantContains string
	}{
		{
			name: "inactive reloader",
			setup: func() *TagReloader {
				return NewTagReloader()
			},
			wantEmpty: true,
		},
		{
			name: "reloading state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.IsReloading = true
				return tr
			},
			wantContains: "Reloading tags",
		},
		{
			name: "error state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.LastError = errors.New("test error")
				return tr
			},
			wantContains: "Tag reload failed",
		},
		{
			name: "success with new tags",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.ReloadResult = &TagReloadResult{
					NewTags:   []string{"tag1", "tag2"},
					TotalTags: 10,
				}
				return tr
			},
			wantContains: "2 new tags found",
		},
		{
			name: "success without new tags",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.ReloadResult = &TagReloadResult{
					TotalTags: 8,
				}
				return tr
			},
			wantContains: "8 total tags",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.setup()
			status := tr.GetStatus()
			
			if tt.wantEmpty {
				if status != "" {
					t.Errorf("Expected empty status, got: %s", status)
				}
			} else if tt.wantContains != "" {
				if !strings.Contains(status, tt.wantContains) {
					t.Errorf("Status %q does not contain %q", status, tt.wantContains)
				}
			}
		})
	}
}

func TestTagReloader_IsActive(t *testing.T) {
	tr := NewTagReloader()
	
	if tr.IsActive() {
		t.Error("New reloader should not be active")
	}
	
	tr.Active = true
	
	if !tr.IsActive() {
		t.Error("Reloader should be active when Active is true")
	}
}

func TestTagReloadRenderer_NewTagReloadRenderer(t *testing.T) {
	renderer := NewTagReloadRenderer(100, 50)
	
	if renderer == nil {
		t.Fatal("NewTagReloadRenderer returned nil")
	}
	
	if renderer.Width != 100 {
		t.Errorf("Width = %d, want 100", renderer.Width)
	}
	
	if renderer.Height != 50 {
		t.Errorf("Height = %d, want 50", renderer.Height)
	}
}

func TestTagReloadRenderer_SetSize(t *testing.T) {
	renderer := NewTagReloadRenderer(100, 50)
	
	renderer.SetSize(200, 75)
	
	if renderer.Width != 200 {
		t.Errorf("Width = %d, want 200", renderer.Width)
	}
	
	if renderer.Height != 75 {
		t.Errorf("Height = %d, want 75", renderer.Height)
	}
}

func TestTagReloadRenderer_RenderStatus(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *TagReloader
		wantEmpty    bool
		wantContains []string
	}{
		{
			name: "inactive reloader",
			setup: func() *TagReloader {
				return NewTagReloader()
			},
			wantEmpty: true,
		},
		{
			name: "reloading state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.IsReloading = true
				return tr
			},
			wantContains: []string{"Reloading Tags", "Scanning"},
		},
		{
			name: "error state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.LastError = errors.New("registry error")
				return tr
			},
			wantContains: []string{"Tag Reload Failed", "registry error"},
		},
		{
			name: "success state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.ReloadResult = &TagReloadResult{
					ComponentsScanned: 15,
					PipelinesScanned:  8,
					TotalTags:        23,
					NewTags:          []string{"api", "database"},
				}
				return tr
			},
			wantContains: []string{
				"Tag Reload Complete",
				"Components scanned: 15",
				"Pipelines scanned: 8",
				"Total tags found: 23",
				"New tags added: 2",
				"api",
				"database",
			},
		},
		{
			name: "success with failed files",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.ReloadResult = &TagReloadResult{
					ComponentsScanned: 10,
					PipelinesScanned:  5,
					TotalTags:        15,
					FailedFiles:      []string{"bad1.yaml", "bad2.yaml"},
				}
				return tr
			},
			wantContains: []string{
				"Failed files: 2",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTagReloadRenderer(100, 50)
			tr := tt.setup()
			output := renderer.RenderStatus(tr)
			
			if tt.wantEmpty {
				if output != "" {
					t.Errorf("Expected empty output, got: %s", output)
				}
			} else {
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("Output does not contain %q\nOutput: %s", want, output)
					}
				}
			}
		})
	}
}

func TestTagReloadRenderer_RenderInlineStatus(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *TagReloader
		wantEmpty    bool
		wantContains string
	}{
		{
			name: "inactive reloader",
			setup: func() *TagReloader {
				return NewTagReloader()
			},
			wantEmpty: true,
		},
		{
			name: "reloading state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.IsReloading = true
				return tr
			},
			wantContains: "Reloading tags",
		},
		{
			name: "error state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.LastError = errors.New("load failed")
				return tr
			},
			wantContains: "Tag reload failed",
		},
		{
			name: "success state",
			setup: func() *TagReloader {
				tr := NewTagReloader()
				tr.Active = true
				tr.ReloadResult = &TagReloadResult{
					NewTags:   []string{"new1"},
					TotalTags: 5,
				}
				return tr
			},
			wantContains: "1 new tags found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTagReloadRenderer(100, 50)
			tr := tt.setup()
			output := renderer.RenderInlineStatus(tr)
			
			if tt.wantEmpty {
				if output != "" {
					t.Errorf("Expected empty output, got: %s", output)
				}
			} else if tt.wantContains != "" {
				// Strip ANSI codes for testing
				stripped := stripANSI(output)
				if !strings.Contains(stripped, tt.wantContains) {
					t.Errorf("Output %q does not contain %q", stripped, tt.wantContains)
				}
			}
		})
	}
}

// stripANSI removes ANSI escape codes from a string for testing
func stripANSI(s string) string {
	// Simple implementation for testing
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			result.WriteRune(r)
		}
	}
	return result.String()
}