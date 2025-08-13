package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func makeTestRenameOperator() *RenameOperator {
	return NewRenameOperator()
}

func TestNewRenameOperator(t *testing.T) {
	ro := NewRenameOperator()
	
	if ro == nil {
		t.Fatal("NewRenameOperator() returned nil")
	}
	
	if ro.confirmDialog == nil {
		t.Error("confirmDialog should be initialized")
	}
}

func TestRenameOperator_ValidateDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		newName  string
		itemType string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid component name",
			newName:  "My Component",
			itemType: "component",
			wantErr:  false,
		},
		{
			name:     "valid pipeline name",
			newName:  "My Pipeline",
			itemType: "pipeline",
			wantErr:  false,
		},
		{
			name:     "empty name",
			newName:  "",
			itemType: "component",
			wantErr:  true,
			errMsg:   "name cannot be empty",
		},
		{
			name:     "whitespace only name",
			newName:  "   ",
			itemType: "component",
			wantErr:  true,
			errMsg:   "name cannot be empty",
		},
		{
			name:     "name too long",
			newName:  "This is a very long name that exceeds one hundred characters which is the maximum allowed length for a component or pipeline name in the system",
			itemType: "component",
			wantErr:  true,
			errMsg:   "name is too long",
		},
		{
			name:     "name at max length",
			newName:  "This name is exactly one hundred characters long including spaces and punctuation marks right here!!",
			itemType: "component",
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			err := ro.ValidateDisplayName(tt.newName, tt.itemType)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDisplayName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestRenameOperator_PrepareRenameComponent(t *testing.T) {
	tests := []struct {
		name            string
		item            componentItem
		wantDisplayName string
		wantPath        string
		wantArchived    bool
	}{
		{
			name: "active component",
			item: componentItem{
				name:       "Test Component",
				path:       "components/contexts/test.md",
				isArchived: false,
			},
			wantDisplayName: "Test Component",
			wantPath:        "components/contexts/test.md",
			wantArchived:    false,
		},
		{
			name: "archived component",
			item: componentItem{
				name:       "Old Component",
				path:       "archive/components/prompts/old.md",
				isArchived: true,
			},
			wantDisplayName: "Old Component",
			wantPath:        "archive/components/prompts/old.md",
			wantArchived:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			displayName, path, isArchived := ro.PrepareRenameComponent(tt.item)
			
			if displayName != tt.wantDisplayName {
				t.Errorf("displayName = %q, want %q", displayName, tt.wantDisplayName)
			}
			
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
			
			if isArchived != tt.wantArchived {
				t.Errorf("isArchived = %v, want %v", isArchived, tt.wantArchived)
			}
		})
	}
}

func TestRenameOperator_PrepareRenamePipeline(t *testing.T) {
	tests := []struct {
		name            string
		item            pipelineItem
		wantDisplayName string
		wantPath        string
		wantArchived    bool
	}{
		{
			name: "active pipeline",
			item: pipelineItem{
				name:       "Test Pipeline",
				path:       "test-pipeline.yaml",
				isArchived: false,
			},
			wantDisplayName: "Test Pipeline",
			wantPath:        "test-pipeline.yaml",
			wantArchived:    false,
		},
		{
			name: "archived pipeline",
			item: pipelineItem{
				name:       "Old Pipeline",
				path:       "archive/pipelines/old-pipeline.yaml",
				isArchived: true,
			},
			wantDisplayName: "Old Pipeline",
			wantPath:        "archive/pipelines/old-pipeline.yaml",
			wantArchived:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			displayName, path, isArchived := ro.PrepareRenamePipeline(tt.item)
			
			if displayName != tt.wantDisplayName {
				t.Errorf("displayName = %q, want %q", displayName, tt.wantDisplayName)
			}
			
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
			
			if isArchived != tt.wantArchived {
				t.Errorf("isArchived = %v, want %v", isArchived, tt.wantArchived)
			}
		})
	}
}

func TestRenameOperator_HandleInput(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *RenameState
		input       tea.KeyMsg
		wantHandled bool
	}{
		{
			name: "delegates to state",
			setup: func() *RenameState {
				rs := NewRenameState()
				rs.Start("Test", "component", "test.md", false)
				return rs
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: true,
		},
		{
			name: "inactive state returns false",
			setup: func() *RenameState {
				return NewRenameState() // Not started
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantHandled: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			state := tt.setup()
			handled, err := ro.HandleInput(tt.input, state)
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

func TestRenameOperator_GetSuccessMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     RenameSuccessMsg
		want    string
	}{
		{
			name: "component rename",
			msg: RenameSuccessMsg{
				ItemType: "component",
				OldName:  "old-component.md",
				NewName:  "New Component",
				IsArchived: false,
			},
			want: "✓ Renamed component to: New Component",
		},
		{
			name: "pipeline rename",
			msg: RenameSuccessMsg{
				ItemType: "pipeline",
				OldName:  "old-pipeline.yaml",
				NewName:  "New Pipeline",
				IsArchived: false,
			},
			want: "✓ Renamed pipeline to: New Pipeline",
		},
		{
			name: "archived component rename",
			msg: RenameSuccessMsg{
				ItemType: "component",
				OldName:  "old-component.md",
				NewName:  "Renamed Archived",
				IsArchived: true,
			},
			want: "✓ Renamed archived component to: Renamed Archived",
		},
		{
			name: "archived pipeline rename",
			msg: RenameSuccessMsg{
				ItemType: "pipeline",
				OldName:  "old-pipeline.yaml",
				NewName:  "Renamed Archived Pipeline",
				IsArchived: true,
			},
			want: "✓ Renamed archived pipeline to: Renamed Archived Pipeline",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			got := ro.GetSuccessMessage(tt.msg)
			
			if got != tt.want {
				t.Errorf("GetSuccessMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenameOperator_GetErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    string
	}{
		{
			name: "simple error",
			err:  fmt.Errorf("file not found"),
			want: "✗ Rename failed: file not found",
		},
		{
			name: "complex error",
			err:  fmt.Errorf("component with name 'test' already exists"),
			want: "✗ Rename failed: component with name 'test' already exists",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			got := ro.GetErrorMessage(tt.err)
			
			if got != tt.want {
				t.Errorf("GetErrorMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenameOperator_ExecuteRename(t *testing.T) {
	tests := []struct {
		name           string
		oldPath        string
		newDisplayName string
		itemType       string
		isArchived     bool
		checkMsgType   func(tea.Msg) bool
	}{
		{
			name:           "component rename command",
			oldPath:        "components/contexts/test.md",
			newDisplayName: "New Component",
			itemType:       "component",
			isArchived:     false,
			checkMsgType: func(msg tea.Msg) bool {
				// The command will execute and return either success or error
				// Since we're testing without actual files, it will likely error
				_, isError := msg.(RenameErrorMsg)
				_, isSuccess := msg.(RenameSuccessMsg)
				return isError || isSuccess
			},
		},
		{
			name:           "pipeline rename command",
			oldPath:        "test-pipeline.yaml",
			newDisplayName: "New Pipeline",
			itemType:       "pipeline",
			isArchived:     false,
			checkMsgType: func(msg tea.Msg) bool {
				_, isError := msg.(RenameErrorMsg)
				_, isSuccess := msg.(RenameSuccessMsg)
				return isError || isSuccess
			},
		},
		{
			name:           "archived component rename command",
			oldPath:        "archive/components/contexts/old.md",
			newDisplayName: "Renamed Archived",
			itemType:       "component",
			isArchived:     true,
			checkMsgType: func(msg tea.Msg) bool {
				_, isError := msg.(RenameErrorMsg)
				_, isSuccess := msg.(RenameSuccessMsg)
				return isError || isSuccess
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := makeTestRenameOperator()
			cmd := ro.ExecuteRename(tt.oldPath, tt.newDisplayName, tt.itemType, tt.isArchived)
			
			if cmd == nil {
				t.Fatal("ExecuteRename() returned nil command")
			}
			
			// Execute the command
			msg := cmd()
			
			if !tt.checkMsgType(msg) {
				t.Errorf("unexpected message type: %T", msg)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}