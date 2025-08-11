package tui

import (
	"strings"
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDetectAtTrigger(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		setup    func(state *EnhancedEditorState)
		expected bool
	}{
		{
			name:    "detects @ at end of content",
			content: "Some text @",
			setup: func(state *EnhancedEditorState) {
				state.Textarea.SetValue("Some text @")
			},
			expected: true,
		},
		{
			name:    "no @ trigger",
			content: "Some text without trigger",
			setup: func(state *EnhancedEditorState) {
				state.Textarea.SetValue("Some text without trigger")
			},
			expected: false,
		},
		{
			name:    "@ in middle of text",
			content: "Some @ text",
			setup: func(state *EnhancedEditorState) {
				state.Textarea.SetValue("Some @ text")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "", []string{})
			tt.setup(state)
			
			result := DetectAtTrigger(state)
			if result != tt.expected {
				t.Errorf("Expected DetectAtTrigger to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestProcessFileSelection(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "formats simple path",
			path:     "/path/to/file.md",
			expected: "@/path/to/file.md",
		},
		{
			name:     "formats relative path",
			path:     "components/prompts/test.md",
			expected: "@components/prompts/test.md",
		},
		{
			name:     "handles empty path",
			path:     "",
			expected: "@.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessFileSelection(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateComponentContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid content",
			content: "# Component\n\nSome valid content",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "   \n\t  ",
			wantErr: true,
			errMsg:  "component cannot be empty",
		},
		{
			name:    "valid YAML frontmatter",
			content: "---\ntitle: Test\n---\nContent",
			wantErr: false,
		},
		{
			name:    "invalid YAML frontmatter",
			content: "---\nIncomplete frontmatter",
			wantErr: true,
			errMsg:  "invalid frontmatter structure",
		},
		{
			name:    "valid file reference",
			content: "Check @components/test.md for details",
			wantErr: false,
		},
		{
			name:    "invalid file reference with path traversal",
			content: "Check @../../etc/passwd for details",
			wantErr: true,
			errMsg:  "path traversal not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateComponentContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateComponentContent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestExtractFileReferences(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single reference",
			content:  "See @components/test.md for more",
			expected: []string{"@components/test.md"},
		},
		{
			name:     "multiple references",
			content:  "Check @file1.md and @file2.md",
			expected: []string{"@file1.md", "@file2.md"},
		},
		{
			name:     "reference at line start",
			content:  "@start.md is the beginning",
			expected: []string{"@start.md"},
		},
		{
			name:     "reference at line end",
			content:  "The end is @end.md",
			expected: []string{"@end.md"},
		},
		{
			name:     "no references",
			content:  "No file references here",
			expected: []string{},
		},
		{
			name:     "multiline with references",
			content:  "Line 1 @ref1.md\nLine 2 @ref2.md\nLine 3",
			expected: []string{"@ref1.md", "@ref2.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFileReferences(tt.content)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d references, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, ref := range result {
				if ref != tt.expected[i] {
					t.Errorf("Expected reference %d to be %s, got %s", i, tt.expected[i], ref)
				}
			}
		})
	}
}

func TestValidateFileReference(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid reference",
			ref:     "@components/test.md",
			wantErr: false,
		},
		{
			name:    "missing @ prefix",
			ref:     "components/test.md",
			wantErr: true,
			errMsg:  "must start with @",
		},
		{
			name:    "empty path after @",
			ref:     "@",
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name:    "path traversal attempt",
			ref:     "@../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal not allowed",
		},
		{
			name:    "valid nested path",
			ref:     "@components/prompts/nested/file.md",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileReference(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileReference() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestParseFileReference(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "valid reference",
			ref:      "@components/test.md",
			wantPath: "components/test.md",
			wantErr:  false,
		},
		{
			name:     "invalid format",
			ref:      "components/test.md",
			wantPath: "",
			wantErr:  true,
		},
		{
			name:     "empty path",
			ref:      "@",
			wantPath: "",
			wantErr:  true,
		},
		{
			name:     "nested path",
			ref:      "@a/b/c/d.txt",
			wantPath: "a/b/c/d.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ParseFileReference(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFileReference() error = %v, wantErr %v", err, tt.wantErr)
			}
			if path != tt.wantPath {
				t.Errorf("Expected path %s, got %s", tt.wantPath, path)
			}
		})
	}
}

func TestInsertTextAtCursor(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		text     string
		pos      int
		expected string
	}{
		{
			name:     "insert at beginning",
			content:  "existing text",
			text:     "new ",
			pos:      0,
			expected: "new existing text",
		},
		{
			name:     "insert at end",
			content:  "existing text",
			text:     " new",
			pos:      13,
			expected: "existing text new",
		},
		{
			name:     "insert in middle",
			content:  "existing text",
			text:     "NEW",
			pos:      9,
			expected: "existing NEWtext",
		},
		{
			name:     "insert with negative position",
			content:  "text",
			text:     "start",
			pos:      -5,
			expected: "starttext",
		},
		{
			name:     "insert with position beyond content",
			content:  "text",
			text:     "end",
			pos:      100,
			expected: "textend",
		},
		{
			name:     "insert into empty content",
			content:  "",
			text:     "text",
			pos:      0,
			expected: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InsertTextAtCursor(tt.content, tt.text, tt.pos)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHandleEnhancedEditorInput(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(state *EnhancedEditorState)
		keyMsg      tea.KeyMsg
		width       int
		expectHandled bool
		validateFunc func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd)
	}{
		{
			name: "inactive state returns false",
			setup: func(state *EnhancedEditorState) {
				// Keep state inactive
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			width:  80,
			expectHandled: false,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if handled {
					t.Error("Expected inactive state to not handle input")
				}
				if cmd != nil {
					t.Error("Expected no command from inactive state")
				}
			},
		},
		{
			name: "active state in normal mode handles input",
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			width:  80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !handled {
					t.Error("Expected active state to handle input")
				}
			},
		},
		{
			name: "file picker mode handles input",
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
				state.StartFilePicker()
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyUp},
			width:  80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !handled {
					t.Error("Expected file picker mode to handle input")
				}
			},
		},
		{
			name: "exit confirmation active handles input",
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
				state.SetContent("modified content")
				onConfirm := func() tea.Cmd { return nil }
				onCancel := func() tea.Cmd { return nil }
				state.ShowExitConfirmation(80, onConfirm, onCancel)
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
			width:  80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !handled {
					t.Error("Expected exit confirmation to handle input")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			tt.setup(state)
			
			handled, cmd := HandleEnhancedEditorInput(state, tt.keyMsg, tt.width)
			
			if handled != tt.expectHandled {
				t.Errorf("Expected handled to be %v, got %v", tt.expectHandled, handled)
			}
			
			tt.validateFunc(t, state, handled, cmd)
		})
	}
}

func TestHandleNormalEditorInputKeyHandling(t *testing.T) {
	tests := []struct {
		name        string
		keyMsg      tea.KeyMsg
		setup       func(state *EnhancedEditorState)
		width       int
		expectHandled bool
		validateFunc func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd)
	}{
		{
			name:   "ctrl+s triggers save",
			keyMsg: tea.KeyMsg{Type: tea.KeyCtrlS},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
			},
			width:         80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if cmd == nil {
					t.Error("Expected save command to be returned")
				}
			},
		},
		{
			name:   "esc with no changes exits",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
			},
			width:         80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if state.IsActive() {
					t.Error("Expected state to be reset after esc with no changes")
				}
			},
		},
		{
			name:   "esc with unsaved changes shows confirmation",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
				state.SetContent("modified content")
			},
			width:         80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !state.IsExitConfirmActive() {
					t.Error("Expected exit confirmation to be shown for unsaved changes")
				}
			},
		},
		{
			name:   "@ triggers file picker",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@")},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
			},
			width:         80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !state.IsFilePicking() {
					t.Error("Expected @ to trigger file picker mode")
				}
			},
		},
		{
			name:   "escaped @ adds literal @",
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@")},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content\\", []string{})
			},
			width:         80,
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if state.IsFilePicking() {
					t.Error("Expected escaped @ to not trigger file picker")
				}
				if !strings.Contains(state.Content, "content@") {
					t.Error("Expected escaped @ to add literal @ to content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			tt.setup(state)
			
			handled, cmd := handleNormalEditorInput(state, tt.keyMsg, tt.width)
			
			if handled != tt.expectHandled {
				t.Errorf("Expected handled to be %v, got %v", tt.expectHandled, handled)
			}
			
			tt.validateFunc(t, state, handled, cmd)
		})
	}
}

func TestHandleFilePickerInputKeys(t *testing.T) {
	tests := []struct {
		name        string
		keyMsg      tea.KeyMsg
		setup       func(state *EnhancedEditorState)
		expectHandled bool
		validateFunc func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd)
	}{
		{
			name:   "esc cancels file picking",
			keyMsg: tea.KeyMsg{Type: tea.KeyEsc},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content@", []string{})
				state.StartFilePicker()
			},
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if state.IsFilePicking() {
					t.Error("Expected esc to stop file picking")
				}
				if strings.Contains(state.Content, "@") {
					t.Error("Expected @ to be removed after canceling file picker")
				}
			},
		},
		{
			name:   "other keys update file picker",
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			setup: func(state *EnhancedEditorState) {
				state.StartEditing("test.md", "test", "prompt", "content", []string{})
				state.StartFilePicker()
			},
			expectHandled: true,
			validateFunc: func(t *testing.T, state *EnhancedEditorState, handled bool, cmd tea.Cmd) {
				if !state.IsFilePicking() {
					t.Error("Expected to still be in file picking mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			tt.setup(state)
			
			handled, cmd := handleFilePickerInput(state, tt.keyMsg)
			
			if handled != tt.expectHandled {
				t.Errorf("Expected handled to be %v, got %v", tt.expectHandled, handled)
			}
			
			tt.validateFunc(t, state, handled, cmd)
		})
	}
}

func TestInitializeEnhancedFilePicker(t *testing.T) {
	tests := []struct {
		name        string
		validateFunc func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd)
	}{
		{
			name: "initializes file picker with current directory",
			validateFunc: func(t *testing.T, state *EnhancedEditorState, cmd tea.Cmd) {
				if cmd == nil {
					t.Error("Expected initialization command to be returned")
				}
				// FilePicker should be configured
				if state.FilePicker.ShowHidden != false {
					t.Error("Expected ShowHidden to be false")
				}
				if state.FilePicker.DirAllowed != true {
					t.Error("Expected DirAllowed to be true")
				}
				if state.FilePicker.FileAllowed != true {
					t.Error("Expected FileAllowed to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", "content", []string{})
			state.StartFilePicker()
			
			cmd := InitializeEnhancedFilePicker(state)
			tt.validateFunc(t, state, cmd)
		})
	}
}

func TestInsertFileReference(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		reference   string
		expected    string
	}{
		{
			name:      "inserts reference removing triggering @",
			content:   "content@",
			reference: "@file.md",
			expected:  "content@file.md",
		},
		{
			name:      "inserts reference with no @ at end",
			content:   "content",
			reference: "@file.md",
			expected:  "content@file.md",
		},
		{
			name:      "inserts reference in empty content",
			content:   "@",
			reference: "@file.md",
			expected:  "@file.md",
		},
		{
			name:      "handles complex reference path",
			content:   "text@",
			reference: "@components/prompts/complex-file.md",
			expected:  "text@components/prompts/complex-file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.StartEditing("test.md", "test", "prompt", tt.content, []string{})
			
			cmd := InsertFileReference(state, tt.reference)
			
			if state.Content != tt.expected {
				t.Errorf("Expected content %q, got %q", tt.expected, state.Content)
			}
			
			if cmd != nil {
				t.Error("Expected no command from InsertFileReference")
			}
		})
	}
}