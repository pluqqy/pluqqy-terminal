package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTagEditor_HandleInput_VimKeysAsText(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *TagEditor
		inputs    []tea.KeyMsg
		wantInput string
		desc      string
	}{
		{
			name: "can type j in tag name",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				return te
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("a")},
				{Type: tea.KeyRunes, Runes: []rune("v")},
				{Type: tea.KeyRunes, Runes: []rune("a")},
			},
			wantInput: "java",
			desc:      "j should be typed as regular text",
		},
		{
			name: "can type k in tag name",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				return te
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("o")},
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("i")},
				{Type: tea.KeyRunes, Runes: []rune("n")},
			},
			wantInput: "kotlin",
			desc:      "k and l should be typed as regular text",
		},
		{
			name: "can type l in tag name",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				return te
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("i")},
				{Type: tea.KeyRunes, Runes: []rune("s")},
				{Type: tea.KeyRunes, Runes: []rune("p")},
			},
			wantInput: "lisp",
			desc:      "l should be typed as regular text",
		},
		{
			name: "can type h in tag name",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				return te
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("h")},
				{Type: tea.KeyRunes, Runes: []rune("a")},
				{Type: tea.KeyRunes, Runes: []rune("s")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
			},
			wantInput: "haskell",
			desc:      "h, k, and l should all be typed as regular text",
		},
		{
			name: "can type mixed vim keys in tag name",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				return te
			},
			inputs: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyRunes, Runes: []rune("l")},
				{Type: tea.KeyRunes, Runes: []rune("h")},
			},
			wantInput: "jklh",
			desc:      "all vim navigation keys should be typed as text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := tt.setup()

			// Process all inputs
			for _, input := range tt.inputs {
				handled, _ := te.HandleInput(input)
				if !handled {
					t.Errorf("Expected input to be handled")
				}
			}

			// Check the result
			if te.TagInput != tt.wantInput {
				t.Errorf("TagInput = %q, want %q (%s)", te.TagInput, tt.wantInput, tt.desc)
			}
		})
	}
}

func TestTagEditor_HandleInput_ArrowNavigation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *TagEditor
		input       tea.KeyMsg
		wantHandled bool
		desc        string
	}{
		{
			name: "up arrow navigates suggestions",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				te.TagInput = "test"
				te.ShowSuggestions = true
				te.SuggestionCursor = 1
				return te
			},
			input:       tea.KeyMsg{Type: tea.KeyUp},
			wantHandled: true,
			desc:        "up arrow should navigate suggestions, not type text",
		},
		{
			name: "down arrow navigates suggestions",
			setup: func() *TagEditor {
				te := NewTagEditor()
				te.Start("/test/path", []string{}, "component")
				te.TagInput = "test"
				te.ShowSuggestions = true
				te.SuggestionCursor = 0
				te.AvailableTags = []string{"test1", "test2"}
				return te
			},
			input:       tea.KeyMsg{Type: tea.KeyDown},
			wantHandled: true,
			desc:        "down arrow should navigate suggestions, not type text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := tt.setup()
			handled, _ := te.HandleInput(tt.input)

			if handled != tt.wantHandled {
				t.Errorf("HandleInput() handled = %v, want %v (%s)", handled, tt.wantHandled, tt.desc)
			}
		})
	}
}