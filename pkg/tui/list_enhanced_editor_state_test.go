package tui

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)

func TestEnhancedEditorState_Initialization(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, state *EnhancedEditorState)
	}{
		{
			name: "creates new state with defaults",
			validate: func(t *testing.T, state *EnhancedEditorState) {
				if state.Active {
					t.Error("Expected Active to be false")
				}
				if state.Mode != EditorModeNormal {
					t.Error("Expected Mode to be EditorModeNormal")
				}
				if state.Content != "" {
					t.Error("Expected Content to be empty")
				}
				if state.UnsavedChanges {
					t.Error("Expected UnsavedChanges to be false")
				}
				if state.IsActive() {
					t.Error("Expected IsActive to be false for new state")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			tt.validate(t, state)
		})
	}
}

func TestEnhancedEditorState_StartEditing(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		compName    string
		compType    string
		content     string
		tags        []string
		validate    func(t *testing.T, state *EnhancedEditorState)
	}{
		{
			name:     "initializes editor with component data",
			path:     "components/prompts/test.md",
			compName: "test.md",
			compType: "prompt",
			content:  "# Test Component\nContent here",
			tags:     []string{"test", "example"},
			validate: func(t *testing.T, state *EnhancedEditorState) {
				if !state.Active {
					t.Error("Expected Active to be true")
				}
				if state.ComponentPath != "components/prompts/test.md" {
					t.Errorf("Expected path to be %s, got %s", "components/prompts/test.md", state.ComponentPath)
				}
				if state.ComponentName != "test.md" {
					t.Errorf("Expected name to be %s, got %s", "test.md", state.ComponentName)
				}
				if state.Content != "# Test Component\nContent here" {
					t.Errorf("Expected content to match, got %s", state.Content)
				}
				if state.UnsavedChanges {
					t.Error("Expected UnsavedChanges to be false initially")
				}
				if !state.IsActive() {
					t.Error("Expected IsActive to be true after StartEditing")
				}
			},
		},
		{
			name:     "handles empty content",
			path:     "components/contexts/empty.yaml",
			compName: "empty.yaml",
			compType: "context",
			content:  "",
			tags:     []string{},
			validate: func(t *testing.T, state *EnhancedEditorState) {
				if !state.Active {
					t.Error("Expected Active to be true")
				}
				if state.Content != "" {
					t.Error("Expected Content to be empty")
				}
				if state.OriginalContent != "" {
					t.Error("Expected OriginalContent to be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing(tt.path, tt.compName, tt.compType, tt.content, tt.tags)
			tt.validate(t, state)
		})
	}
}

func TestEnhancedEditorState_ContentChanges(t *testing.T) {
	tests := []struct {
		name            string
		originalContent string
		newContent      string
		expectUnsaved   bool
	}{
		{
			name:            "detects content changes",
			originalContent: "original",
			newContent:      "modified",
			expectUnsaved:   true,
		},
		{
			name:            "no changes when content same",
			originalContent: "unchanged",
			newContent:      "unchanged",
			expectUnsaved:   false,
		},
		{
			name:            "detects changes with empty original",
			originalContent: "",
			newContent:      "new content",
			expectUnsaved:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", tt.originalContent, []string{})
			state.SetContent(tt.newContent)
			
			if state.HasUnsavedChanges() != tt.expectUnsaved {
				t.Errorf("Expected unsaved changes to be %v, got %v", tt.expectUnsaved, state.HasUnsavedChanges())
			}
		})
	}
}

func TestEnhancedEditorState_ModeTransitions(t *testing.T) {
	tests := []struct {
		name        string
		transitions []EditorMode
		finalMode   EditorMode
	}{
		{
			name:        "transitions to file picking mode",
			transitions: []EditorMode{EditorModeFilePicking},
			finalMode:   EditorModeFilePicking,
		},
		{
			name:        "transitions back to normal mode",
			transitions: []EditorMode{EditorModeFilePicking, EditorModeNormal},
			finalMode:   EditorModeNormal,
		},
		{
			name:        "multiple mode changes",
			transitions: []EditorMode{EditorModeFilePicking, EditorModeNormal, EditorModeFilePicking, EditorModeNormal},
			finalMode:   EditorModeNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "content", []string{})
			
			for _, mode := range tt.transitions {
				state.SetMode(mode)
			}
			
			if state.GetMode() != tt.finalMode {
				t.Errorf("Expected final mode to be %v, got %v", tt.finalMode, state.GetMode())
			}
		})
	}
}

func TestEnhancedEditorState_FilePicker(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(state *EnhancedEditorState)
		validate   func(t *testing.T, state *EnhancedEditorState)
	}{
		{
			name: "starts file picker",
			setup: func(state *EnhancedEditorState) {
				state.StartFilePicker()
			},
			validate: func(t *testing.T, state *EnhancedEditorState) {
				if !state.IsFilePicking() {
					t.Error("Expected IsFilePicking to be true")
				}
				if state.Mode != EditorModeFilePicking {
					t.Error("Expected Mode to be EditorModeFilePicking")
				}
			},
		},
		{
			name: "stops file picker",
			setup: func(state *EnhancedEditorState) {
				state.StartFilePicker()
				state.StopFilePicker()
			},
			validate: func(t *testing.T, state *EnhancedEditorState) {
				if state.IsFilePicking() {
					t.Error("Expected IsFilePicking to be false")
				}
				if state.Mode != EditorModeNormal {
					t.Error("Expected Mode to be EditorModeNormal")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "content", []string{})
			tt.setup(state)
			tt.validate(t, state)
		})
	}
}

func TestEnhancedEditorState_Reset(t *testing.T) {
	state := NewEnhancedEditorState()
	
	// Set up state with data
	state.StartEditing("test.md", "test", "prompt", "content", []string{})
	state.SetContent("modified content")
	state.StartFilePicker()
	state.UpdateCursor(10)
	state.SetInsertionPoint(5)
	
	// Reset the state
	state.Reset()
	
	// Verify all fields are reset
	if state.Active {
		t.Error("Expected Active to be false after reset")
	}
	if state.Mode != EditorModeNormal {
		t.Error("Expected Mode to be EditorModeNormal after reset")
	}
	if state.ComponentPath != "" {
		t.Error("Expected ComponentPath to be empty after reset")
	}
	if state.ComponentName != "" {
		t.Error("Expected ComponentName to be empty after reset")
	}
	if state.Content != "" {
		t.Error("Expected Content to be empty after reset")
	}
	if state.OriginalContent != "" {
		t.Error("Expected OriginalContent to be empty after reset")
	}
	if state.CursorPosition != 0 {
		t.Error("Expected CursorPosition to be 0 after reset")
	}
	if state.InsertionPoint != 0 {
		t.Error("Expected InsertionPoint to be 0 after reset")
	}
	if state.UnsavedChanges {
		t.Error("Expected UnsavedChanges to be false after reset")
	}
}

func TestEnhancedEditorState_TextareaDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "sets normal dimensions",
			width:  80,
			height: 24,
		},
		{
			name:   "sets large dimensions",
			width:  200,
			height: 50,
		},
		{
			name:   "sets small dimensions",
			width:  40,
			height: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.SetTextareaDimensions(tt.width, tt.height)
			
			// Note: We can't directly test the textarea dimensions
			// as they're internal to the textarea model
			// This test ensures the method doesn't panic
		})
	}
}

func TestEnhancedEditorState_InsertionPoint(t *testing.T) {
	tests := []struct {
		name      string
		position  int
		expected  int
	}{
		{
			name:     "sets insertion point",
			position: 10,
			expected: 10,
		},
		{
			name:     "sets insertion point to zero",
			position: 0,
			expected: 0,
		},
		{
			name:     "sets large insertion point",
			position: 1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.SetInsertionPoint(tt.position)
			
			if state.GetInsertionPoint() != tt.expected {
				t.Errorf("Expected insertion point to be %d, got %d", tt.expected, state.GetInsertionPoint())
			}
		})
	}
}

func TestEnhancedEditorState_UpdateTextarea(t *testing.T) {
	tests := []struct {
		name        string
		initialContent string
		keyMsg      tea.KeyMsg
		expectedContentChange bool
		validateFunc func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd)
	}{
		{
			name:           "updates content with key runes",
			initialContent: "initial",
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("new")},
			expectedContentChange: true, // UpdateTextarea does track content changes
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// The textarea should have been updated - command may or may not be returned
			},
		},
		{
			name:           "handles backspace",
			initialContent: "delete me",
			keyMsg:         tea.KeyMsg{Type: tea.KeyBackspace},
			expectedContentChange: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// Command may or may not be returned, test ensures no panic
			},
		},
		{
			name:           "handles enter key",
			initialContent: "line1",
			keyMsg:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedContentChange: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// Command may or may not be returned, test ensures no panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", tt.initialContent, []string{})
			
			originalContent := state.Content
			cmd := state.UpdateTextarea(tt.keyMsg)
			
			tt.validateFunc(t, state, cmd)
			
			// Check if content changed as expected
			contentChanged := state.Content != originalContent
			if contentChanged != tt.expectedContentChange {
				t.Errorf("Expected content change: %v, got: %v", tt.expectedContentChange, contentChanged)
			}
		})
	}
}

func TestEnhancedEditorState_UpdateFilePicker(t *testing.T) {
	tests := []struct {
		name        string
		keyMsg      tea.KeyMsg
		validateFunc func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd)
	}{
		{
			name:   "handles up arrow in file picker",
			keyMsg: tea.KeyMsg{Type: tea.KeyUp},
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// Command may or may not be returned, test just ensures no panic
				// The method executed successfully if we get here
			},
		},
		{
			name:   "handles down arrow in file picker",
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// Command may or may not be returned, test just ensures no panic
			},
		},
		{
			name:   "handles enter key in file picker",
			keyMsg: tea.KeyMsg{Type: tea.KeyEnter},
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				// Command may or may not be returned, test just ensures no panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "content", []string{})
			state.StartFilePicker()
			
			cmd := state.UpdateFilePicker(tt.keyMsg)
			tt.validateFunc(t, state, cmd)
		})
	}
}

func TestEnhancedEditorState_ExitConfirmation(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		setup       func(state *EnhancedEditorState)
		validateFunc func(t *testing.T, state *EnhancedEditorState)
	}{
		{
			name:  "shows exit confirmation dialog",
			width: 80,
			setup: func(state *EnhancedEditorState) {
				onConfirm := func() tea.Cmd { return nil }
				onCancel := func() tea.Cmd { return nil }
				state.ShowExitConfirmation(80, onConfirm, onCancel)
			},
			validateFunc: func(t *testing.T, state *EnhancedEditorState) {
				if !state.IsExitConfirmActive() {
					t.Error("Expected exit confirmation to be active")
				}
			},
		},
		{
			name:  "hides exit confirmation dialog",
			width: 80,
			setup: func(state *EnhancedEditorState) {
				onConfirm := func() tea.Cmd { return nil }
				onCancel := func() tea.Cmd { return nil }
				state.ShowExitConfirmation(80, onConfirm, onCancel)
				state.HideExitConfirmation()
			},
			validateFunc: func(t *testing.T, state *EnhancedEditorState) {
				if state.IsExitConfirmActive() {
					t.Error("Expected exit confirmation to be inactive after hiding")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "content", []string{})
			
			// Initially should not be active
			if state.IsExitConfirmActive() {
				t.Error("Expected exit confirmation to be inactive initially")
			}
			
			tt.setup(state)
			tt.validateFunc(t, state)
		})
	}
}