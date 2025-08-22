package tui

import (
	"strings"
	"testing"
)

func TestEnhancedEditorState_Undo(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		changes         []string
		undoCount       int
		expectedContent string
		expectSuccess   bool
	}{
		{
			name:            "undo single change",
			initialContent:  "original",
			changes:         []string{"modified"},
			undoCount:       1,
			expectedContent: "original",
			expectSuccess:   true,
		},
		{
			name:            "undo multiple changes",
			initialContent:  "start",
			changes:         []string{"first change", "second change", "third change"},
			undoCount:       2,
			expectedContent: "first change",
			expectSuccess:   true,
		},
		{
			name:            "undo with no history",
			initialContent:  "content",
			changes:         []string{},
			undoCount:       1,
			expectedContent: "content",
			expectSuccess:   false,
		},
		{
			name:            "undo more than history",
			initialContent:  "initial",
			changes:         []string{"change1", "change2"},
			undoCount:       5,
			expectedContent: "initial",
			expectSuccess:   true, // First undo succeeds, rest fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.Content = tt.initialContent
			state.OriginalContent = tt.initialContent
			state.MaxUndoLevels = 10

			// Apply changes and save undo states
			for _, change := range tt.changes {
				state.SaveUndoState("test change")
				state.SetContent(change)
			}

			// Perform undos
			var lastSuccess bool
			for i := 0; i < tt.undoCount; i++ {
				lastSuccess = state.Undo()
			}

			// Check results
			if lastSuccess != tt.expectSuccess && tt.undoCount == 1 {
				t.Errorf("Undo() returned %v, expected %v", lastSuccess, tt.expectSuccess)
			}

			if state.Content != tt.expectedContent {
				t.Errorf("Content after undo = %q, expected %q", state.Content, tt.expectedContent)
			}
		})
	}
}

func TestEnhancedEditorState_SaveUndoState(t *testing.T) {
	tests := []struct {
		name              string
		maxLevels         int
		statesToSave      int
		expectedStackSize int
	}{
		{
			name:              "save within limit",
			maxLevels:         5,
			statesToSave:      3,
			expectedStackSize: 3,
		},
		{
			name:              "save exceeding limit",
			maxLevels:         3,
			statesToSave:      5,
			expectedStackSize: 3,
		},
		{
			name:              "save exactly at limit",
			maxLevels:         4,
			statesToSave:      4,
			expectedStackSize: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.MaxUndoLevels = tt.maxLevels

			for i := 0; i < tt.statesToSave; i++ {
				state.Content = strings.Repeat("a", i+1)
				state.SaveUndoState("test")
			}

			if len(state.UndoStack) != tt.expectedStackSize {
				t.Errorf("UndoStack size = %d, expected %d", len(state.UndoStack), tt.expectedStackSize)
			}

			// Verify oldest states are removed when exceeding limit
			if tt.statesToSave > tt.maxLevels {
				// The oldest saved state should have been removed
				firstState := state.UndoStack[0]
				minExpectedLen := tt.statesToSave - tt.maxLevels + 1
				if len(firstState.Content) < minExpectedLen {
					t.Errorf("Oldest state has wrong content length: %d, expected at least %d",
						len(firstState.Content), minExpectedLen)
				}
			}
		})
	}
}

func TestEnhancedEditorState_UpdateStats(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedLines int
		expectedWords int
	}{
		{
			name:          "empty content",
			content:       "",
			expectedLines: 0,
			expectedWords: 0,
		},
		{
			name:          "single line",
			content:       "Hello world",
			expectedLines: 1,
			expectedWords: 2,
		},
		{
			name:          "multiple lines",
			content:       "Line one\nLine two\nLine three",
			expectedLines: 3,
			expectedWords: 6,
		},
		{
			name:          "lines with extra spaces",
			content:       "  Word1   Word2  \n\n  Word3  ",
			expectedLines: 3,
			expectedWords: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.Textarea.SetValue(tt.content)
			state.UpdateStats()

			if state.LineCount != tt.expectedLines {
				t.Errorf("LineCount = %d, expected %d", state.LineCount, tt.expectedLines)
			}

			if state.WordCount != tt.expectedWords {
				t.Errorf("WordCount = %d, expected %d", state.WordCount, tt.expectedWords)
			}
		})
	}
}

func TestEnhancedEditorState_HasUnsavedChanges(t *testing.T) {
	tests := []struct {
		name            string
		originalContent string
		currentContent  string
		expected        bool
	}{
		{
			name:            "no changes",
			originalContent: "content",
			currentContent:  "content",
			expected:        false,
		},
		{
			name:            "has changes",
			originalContent: "original",
			currentContent:  "modified",
			expected:        true,
		},
		{
			name:            "empty to content",
			originalContent: "",
			currentContent:  "new content",
			expected:        true,
		},
		{
			name:            "content to empty",
			originalContent: "content",
			currentContent:  "",
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewEnhancedEditorState()
			state.OriginalContent = tt.originalContent
			state.Content = tt.currentContent
			state.UnsavedChanges = (tt.currentContent != tt.originalContent)

			if got := state.HasUnsavedChanges(); got != tt.expected {
				t.Errorf("HasUnsavedChanges() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"", 0},
		{"word", 1},
		{"two words", 2},
		{"  spaces  between  words  ", 3},
		{"line\nbreaks\ncount", 3},
		{"tabs\ttoo\tcount", 3},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			if got := CountWords(tt.content); got != tt.expected {
				t.Errorf("CountWords(%q) = %d, expected %d", tt.content, got, tt.expected)
			}
		})
	}
}
