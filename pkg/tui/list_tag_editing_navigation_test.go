package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTagEditor_NavigationSuggestions(t *testing.T) {
	tests := []struct {
		name                    string
		input                   string
		keySequence            []tea.KeyMsg
		expectedHasNavigated   bool
		expectedSuggestionCursor int
	}{
		{
			name:  "typing without navigation shows suggestions but not navigated",
			input: "te",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
			},
			expectedHasNavigated:   false,
			expectedSuggestionCursor: 0,
		},
		{
			name:  "pressing down arrow sets navigation flag",
			input: "te",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyDown},
			},
			expectedHasNavigated:   true,
			expectedSuggestionCursor: 1,
		},
		{
			name:  "pressing up arrow sets navigation flag even at position 0",
			input: "te",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyUp},
			},
			expectedHasNavigated:   true,
			expectedSuggestionCursor: 0,
		},
		{
			name:  "typing after navigation resets flag",
			input: "tes",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyDown},
				{Type: tea.KeyRunes, Runes: []rune("s")},
			},
			expectedHasNavigated:   false,
			expectedSuggestionCursor: 0,
		},
		{
			name:  "backspace after navigation resets flag",
			input: "t",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("t")},
				{Type: tea.KeyRunes, Runes: []rune("e")},
				{Type: tea.KeyDown},
				{Type: tea.KeyBackspace},
			},
			expectedHasNavigated:   false,
			expectedSuggestionCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewTagEditor()
			editor.Start("test.md", []string{}, "component")
			editor.AvailableTags = []string{"test", "testing", "temp", "template"}
			
			// Process key sequence
			for _, key := range tt.keySequence {
				editor.HandleInput(key)
			}
			
			// Check results
			if editor.TagInput != tt.input {
				t.Errorf("expected input %q, got %q", tt.input, editor.TagInput)
			}
			
			if editor.HasNavigatedSuggestions != tt.expectedHasNavigated {
				t.Errorf("expected HasNavigatedSuggestions=%v, got %v", 
					tt.expectedHasNavigated, editor.HasNavigatedSuggestions)
			}
			
			if editor.SuggestionCursor != tt.expectedSuggestionCursor {
				t.Errorf("expected SuggestionCursor=%d, got %d", 
					tt.expectedSuggestionCursor, editor.SuggestionCursor)
			}
		})
	}
}

func TestTagEditor_SingleSuggestionNavigation(t *testing.T) {
	// Test the specific case of a single suggestion
	editor := NewTagEditor()
	editor.Start("test.md", []string{}, "component")
	editor.AvailableTags = []string{"oooo"}
	
	// Type "oo" - should show "oooo" as suggestion
	editor.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	editor.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	
	// Should have suggestion but not navigated
	if editor.HasNavigatedSuggestions {
		t.Error("expected HasNavigatedSuggestions=false after typing")
	}
	
	suggestions := editor.GetSuggestions()
	if len(suggestions) != 1 || suggestions[0] != "oooo" {
		t.Errorf("expected single suggestion 'oooo', got %v", suggestions)
	}
	
	// Press down arrow - even though cursor can't move, should set navigation flag
	editor.HandleInput(tea.KeyMsg{Type: tea.KeyDown})
	
	if !editor.HasNavigatedSuggestions {
		t.Error("expected HasNavigatedSuggestions=true after pressing down on single suggestion")
	}
	
	if editor.SuggestionCursor != 0 {
		t.Errorf("expected SuggestionCursor=0 for single suggestion, got %d", editor.SuggestionCursor)
	}
}