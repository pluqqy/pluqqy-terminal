package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test helpers
func makeTestRenameState() *RenameState {
	return NewRenameState()
}

func TestNewRenameState(t *testing.T) {
	rs := NewRenameState()
	
	if rs == nil {
		t.Fatal("NewRenameState() returned nil")
	}
	
	if rs.Active {
		t.Error("Active should be false initially")
	}
	
	if rs.ItemType != "" {
		t.Errorf("ItemType = %q, want empty", rs.ItemType)
	}
	
	if rs.OriginalName != "" {
		t.Errorf("OriginalName = %q, want empty", rs.OriginalName)
	}
	
	if rs.NewName != "" {
		t.Errorf("NewName = %q, want empty", rs.NewName)
	}
	
	if rs.ValidationError != "" {
		t.Errorf("ValidationError = %q, want empty", rs.ValidationError)
	}
}

func TestRenameState_Start(t *testing.T) {
	tests := []struct {
		name         string
		displayName  string
		itemType     string
		path         string
		isArchived   bool
		wantActive   bool
		wantNewName  string
		wantCursorPos int
	}{
		{
			name:         "start rename for component",
			displayName:  "Test Component",
			itemType:     "component",
			path:         "components/contexts/test.md",
			isArchived:   false,
			wantActive:   true,
			wantNewName:  "Test Component", // Pre-filled with current name
			wantCursorPos: 14, // Cursor at end of "Test Component"
		},
		{
			name:         "start rename for pipeline",
			displayName:  "Test Pipeline",
			itemType:     "pipeline",
			path:         "test-pipeline.yaml",
			isArchived:   false,
			wantActive:   true,
			wantNewName:  "Test Pipeline",
			wantCursorPos: 13, // Cursor at end of "Test Pipeline"
		},
		{
			name:         "start rename for archived component",
			displayName:  "Archived Component",
			itemType:     "component",
			path:         "archive/components/contexts/old.md",
			isArchived:   true,
			wantActive:   true,
			wantNewName:  "Archived Component",
			wantCursorPos: 18, // Cursor at end of "Archived Component"
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := makeTestRenameState()
			rs.Start(tt.displayName, tt.itemType, tt.path, tt.isArchived)
			
			if rs.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", rs.Active, tt.wantActive)
			}
			
			if rs.ItemType != tt.itemType {
				t.Errorf("ItemType = %q, want %q", rs.ItemType, tt.itemType)
			}
			
			if rs.OriginalName != tt.displayName {
				t.Errorf("OriginalName = %q, want %q", rs.OriginalName, tt.displayName)
			}
			
			if rs.OriginalPath != tt.path {
				t.Errorf("OriginalPath = %q, want %q", rs.OriginalPath, tt.path)
			}
			
			if rs.NewName != tt.wantNewName {
				t.Errorf("NewName = %q, want %q", rs.NewName, tt.wantNewName)
			}
			
			if rs.IsArchived != tt.isArchived {
				t.Errorf("IsArchived = %v, want %v", rs.IsArchived, tt.isArchived)
			}
			
			if rs.CursorPos != tt.wantCursorPos {
				t.Errorf("CursorPos = %d, want %d", rs.CursorPos, tt.wantCursorPos)
			}
		})
	}
}

func TestRenameState_StartRename(t *testing.T) {
	rs := makeTestRenameState()
	
	path := "components/contexts/test.md"
	displayName := "Test Component"
	itemType := "component"
	
	rs.StartRename(path, displayName, itemType)
	
	if !rs.Active {
		t.Error("Active should be true after StartRename")
	}
	
	if rs.ItemType != itemType {
		t.Errorf("ItemType = %q, want %q", rs.ItemType, itemType)
	}
	
	if rs.OriginalPath != path {
		t.Errorf("OriginalPath = %q, want %q", rs.OriginalPath, path)
	}
	
	if rs.OriginalName != displayName {
		t.Errorf("OriginalName = %q, want %q", rs.OriginalName, displayName)
	}
	
	// Should default to not archived
	if rs.IsArchived {
		t.Error("IsArchived should be false by default")
	}
}

func TestRenameState_HandleInput(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *RenameState
		input         tea.KeyMsg
		wantHandled   bool
		wantActive    bool
		wantNewName   string
		wantCursorPos int
		checkCommand  bool
	}{
		{
			name: "escape cancels rename",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Modified"
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyEsc},
			wantHandled: true,
			wantActive:  false,
			wantNewName: "", // Should be reset
			wantCursorPos: 0,
		},
		{
			name: "enter with valid name triggers rename",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "New Name"
				rs.CursorPos = 8  // Set cursor to end of "New Name"
				rs.ValidationError = "" // No error
				return rs
			},
			input:        tea.KeyMsg{Type: tea.KeyEnter},
			wantHandled:  true,
			wantActive:   true, // Still active until command completes
			wantNewName:  "New Name",
			wantCursorPos: 8,  // Cursor position stays at 8
			checkCommand: true,
		},
		{
			name: "enter with empty name does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = ""
				rs.CursorPos = 0
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyEnter},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "",
			wantCursorPos: 0,
		},
		{
			name: "enter with validation error does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Invalid"
				rs.ValidationError = "Name already exists"
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyEnter},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Invalid",
			wantCursorPos: 4,
		},
		{
			name: "backspace removes character before cursor",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test Name"
				rs.CursorPos = 5 // After "Test "
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyBackspace},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "TestName",
			wantCursorPos: 4,
		},
		{
			name: "backspace at beginning does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 0
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyBackspace},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 0,
		},
		{
			name: "space adds space at cursor position",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "TestName"
				rs.CursorPos = 4 // After "Test"
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeySpace},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test Name",
			wantCursorPos: 5,
		},
		{
			name: "tab is ignored",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyTab},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 4,
		},
		{
			name: "character input adds at cursor position",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Tet"
				rs.CursorPos = 2 // After "Te"
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 3,
		},
		{
			name: "left arrow moves cursor left",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 2
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyLeft},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 1,
		},
		{
			name: "left arrow at beginning does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 0
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyLeft},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 0,
		},
		{
			name: "right arrow moves cursor right",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 2
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyRight},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 3,
		},
		{
			name: "right arrow at end does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 4
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyRight},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 4,
		},
		{
			name: "home moves cursor to beginning",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 3
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyHome},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 0,
		},
		{
			name: "end moves cursor to end",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 1
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyEnd},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 4,
		},
		{
			name: "delete removes character at cursor",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 1 // On 'e'
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyDelete},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Tst",
			wantCursorPos: 1,
		},
		{
			name: "delete at end does nothing",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.Start("Test", "component", "test.md", false)
				rs.NewName = "Test"
				rs.CursorPos = 4
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyDelete},
			wantHandled: true,
			wantActive:  true,
			wantNewName: "Test",
			wantCursorPos: 4,
		},
		{
			name: "inactive state ignores input",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				// Don't start it
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: false,
			wantActive:  false,
			wantNewName: "",
			wantCursorPos: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := tt.setup()
			handled, cmd := rs.HandleInput(tt.input)
			
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			
			if rs.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", rs.Active, tt.wantActive)
			}
			
			if rs.NewName != tt.wantNewName {
				t.Errorf("NewName = %q, want %q", rs.NewName, tt.wantNewName)
			}
			
			if rs.CursorPos != tt.wantCursorPos {
				t.Errorf("CursorPos = %d, want %d", rs.CursorPos, tt.wantCursorPos)
			}
			
			if tt.checkCommand && cmd == nil {
				t.Error("Expected command but got nil")
			}
		})
	}
}

func TestRenameState_Reset(t *testing.T) {
	rs := makeTestRenameState()
	
	// Set up some state
	rs.Start("Test Component", "component", "test.md", true)
	rs.NewName = "Modified Name"
	rs.ValidationError = "Some error"
	rs.AffectedActive = []string{"pipeline1"}
	rs.AffectedArchive = []string{"pipeline2"}
	
	// Reset
	rs.Reset()
	
	// Verify everything is cleared
	if rs.Active {
		t.Error("Active should be false after reset")
	}
	
	if rs.ItemType != "" {
		t.Errorf("ItemType = %q, want empty", rs.ItemType)
	}
	
	if rs.OriginalName != "" {
		t.Errorf("OriginalName = %q, want empty", rs.OriginalName)
	}
	
	if rs.OriginalPath != "" {
		t.Errorf("OriginalPath = %q, want empty", rs.OriginalPath)
	}
	
	if rs.NewName != "" {
		t.Errorf("NewName = %q, want empty", rs.NewName)
	}
	
	if rs.ValidationError != "" {
		t.Errorf("ValidationError = %q, want empty", rs.ValidationError)
	}
	
	if rs.AffectedActive != nil {
		t.Error("AffectedActive should be nil after reset")
	}
	
	if rs.AffectedArchive != nil {
		t.Error("AffectedArchive should be nil after reset")
	}
	
	if rs.IsArchived {
		t.Error("IsArchived should be false after reset")
	}
}

func TestRenameState_IsActive(t *testing.T) {
	rs := makeTestRenameState()
	
	if rs.IsActive() {
		t.Error("IsActive() should return false initially")
	}
	
	rs.Start("Test", "component", "test.md", false)
	
	if !rs.IsActive() {
		t.Error("IsActive() should return true after Start")
	}
	
	rs.Reset()
	
	if rs.IsActive() {
		t.Error("IsActive() should return false after Reset")
	}
}

func TestRenameState_GetError(t *testing.T) {
	rs := makeTestRenameState()
	
	// Initially no error
	if rs.GetError() != nil {
		t.Error("GetError() should return nil initially")
	}
	
	// With validation error, still returns nil (as per implementation)
	rs.ValidationError = "Some validation error"
	if rs.GetError() != nil {
		t.Error("GetError() should return nil even with validation error")
	}
}

func TestRenameState_GetItemType(t *testing.T) {
	rs := makeTestRenameState()
	
	if rs.GetItemType() != "" {
		t.Errorf("GetItemType() = %q, want empty", rs.GetItemType())
	}
	
	rs.ItemType = "component"
	if rs.GetItemType() != "component" {
		t.Errorf("GetItemType() = %q, want %q", rs.GetItemType(), "component")
	}
	
	rs.ItemType = "pipeline"
	if rs.GetItemType() != "pipeline" {
		t.Errorf("GetItemType() = %q, want %q", rs.GetItemType(), "pipeline")
	}
}

func TestRenameState_GetNewName(t *testing.T) {
	rs := makeTestRenameState()
	
	if rs.GetNewName() != "" {
		t.Errorf("GetNewName() = %q, want empty", rs.GetNewName())
	}
	
	rs.NewName = "New Component Name"
	if rs.GetNewName() != "New Component Name" {
		t.Errorf("GetNewName() = %q, want %q", rs.GetNewName(), "New Component Name")
	}
}

func TestRenameState_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *RenameState
		wantValid bool
	}{
		{
			name: "empty name is invalid",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.NewName = ""
				rs.OriginalName = "Original"
				rs.ValidationError = ""
				return rs
			},
			wantValid: false,
		},
		{
			name: "validation error makes invalid",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.NewName = "New Name"
				rs.OriginalName = "Original"
				rs.ValidationError = "Error"
				return rs
			},
			wantValid: false,
		},
		{
			name: "same as original is invalid",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.NewName = "Same Name"
				rs.OriginalName = "Same Name"
				rs.ValidationError = ""
				return rs
			},
			wantValid: false,
		},
		{
			name: "different valid name is valid",
			setup: func() *RenameState {
				rs := makeTestRenameState()
				rs.NewName = "New Name"
				rs.OriginalName = "Original"
				rs.ValidationError = ""
				return rs
			},
			wantValid: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := tt.setup()
			if got := rs.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestRenameState_HasAffectedPipelines(t *testing.T) {
	tests := []struct {
		name           string
		affectedActive []string
		affectedArchive []string
		want           bool
	}{
		{
			name:           "no affected pipelines",
			affectedActive: nil,
			affectedArchive: nil,
			want:           false,
		},
		{
			name:           "empty slices",
			affectedActive: []string{},
			affectedArchive: []string{},
			want:           false,
		},
		{
			name:           "has active affected",
			affectedActive: []string{"pipeline1"},
			affectedArchive: nil,
			want:           true,
		},
		{
			name:           "has archived affected",
			affectedActive: nil,
			affectedArchive: []string{"archived1"},
			want:           true,
		},
		{
			name:           "has both affected",
			affectedActive: []string{"pipeline1"},
			affectedArchive: []string{"archived1"},
			want:           true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := makeTestRenameState()
			rs.AffectedActive = tt.affectedActive
			rs.AffectedArchive = tt.affectedArchive
			
			if got := rs.HasAffectedPipelines(); got != tt.want {
				t.Errorf("HasAffectedPipelines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenameState_GetSlugifiedName(t *testing.T) {
	tests := []struct {
		name     string
		newName  string
		wantSlug string
	}{
		{
			name:     "empty name",
			newName:  "",
			wantSlug: "",
		},
		{
			name:     "simple name",
			newName:  "Test Component",
			wantSlug: "test-component",
		},
		{
			name:     "name with special chars",
			newName:  "Test's Component!",
			wantSlug: "test-s-component",
		},
		{
			name:     "name with numbers",
			newName:  "Component #1",
			wantSlug: "component-1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := makeTestRenameState()
			rs.NewName = tt.newName
			
			if got := rs.GetSlugifiedName(); got != tt.wantSlug {
				t.Errorf("GetSlugifiedName() = %q, want %q", got, tt.wantSlug)
			}
		})
	}
}