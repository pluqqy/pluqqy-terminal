package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNewTagEditor(t *testing.T) {
	te := NewTagEditor()
	
	assert.NotNil(t, te)
	assert.False(t, te.Active)
	assert.Empty(t, te.CurrentTags)
	assert.Empty(t, te.AvailableTags)
	assert.NotNil(t, te.TagDeleteConfirm)
	assert.NotNil(t, te.ExitConfirm)
	assert.NotNil(t, te.TagReloader)
	assert.NotNil(t, te.TagDeletionState)
	assert.Equal(t, TagEditorModeNormal, te.Mode)
}

func TestTagEditor_Start(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		currentTags []string
		itemType    string
		itemName    string
		wantActive  bool
	}{
		{
			name:        "Start with empty tags",
			path:        "/test/path",
			currentTags: []string{},
			itemType:    "component",
			itemName:    "test-component",
			wantActive:  true,
		},
		{
			name:        "Start with existing tags",
			path:        "/test/path",
			currentTags: []string{"tag1", "tag2"},
			itemType:    "pipeline",
			itemName:    "test-pipeline",
			wantActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTagEditor()
			te.Start(tt.path, tt.currentTags, tt.itemType, tt.itemName)
			
			assert.Equal(t, tt.wantActive, te.Active)
			assert.Equal(t, tt.path, te.Path)
			assert.Equal(t, tt.itemType, te.ItemType)
			assert.Equal(t, tt.itemName, te.ItemName)
			assert.Equal(t, tt.currentTags, te.CurrentTags)
			assert.Equal(t, tt.currentTags, te.OriginalTags)
			assert.Equal(t, TagEditorModeNormal, te.Mode)
		})
	}
}

func TestTagEditor_Reset(t *testing.T) {
	te := NewTagEditor()
	te.Start("/test/path", []string{"tag1", "tag2"}, "component", "test")
	te.TagInput = "newtag"
	te.TagCursor = 1
	te.TagCloudActive = true
	te.Mode = TagEditorModeDeleting
	
	te.Reset()
	
	assert.False(t, te.Active)
	assert.Empty(t, te.Path)
	assert.Empty(t, te.ItemType)
	assert.Empty(t, te.ItemName)
	assert.Empty(t, te.CurrentTags)
	assert.Empty(t, te.OriginalTags)
	assert.Empty(t, te.TagInput)
	assert.Equal(t, 0, te.TagCursor)
	assert.False(t, te.TagCloudActive)
	assert.Equal(t, TagEditorModeNormal, te.Mode)
}

func TestTagEditor_HasChanges(t *testing.T) {
	tests := []struct {
		name         string
		originalTags []string
		currentTags  []string
		wantChanges  bool
	}{
		{
			name:         "No changes - empty",
			originalTags: []string{},
			currentTags:  []string{},
			wantChanges:  false,
		},
		{
			name:         "No changes - same tags",
			originalTags: []string{"tag1", "tag2"},
			currentTags:  []string{"tag1", "tag2"},
			wantChanges:  false,
		},
		{
			name:         "Has changes - added tag",
			originalTags: []string{"tag1"},
			currentTags:  []string{"tag1", "tag2"},
			wantChanges:  true,
		},
		{
			name:         "Has changes - removed tag",
			originalTags: []string{"tag1", "tag2"},
			currentTags:  []string{"tag1"},
			wantChanges:  true,
		},
		{
			name:         "No changes - different order",
			originalTags: []string{"tag1", "tag2"},
			currentTags:  []string{"tag2", "tag1"},
			wantChanges:  false, // Order doesn't matter in HasChanges
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTagEditor()
			te.OriginalTags = tt.originalTags
			te.CurrentTags = tt.currentTags
			
			assert.Equal(t, tt.wantChanges, te.HasChanges())
		})
	}
}

func TestTagEditor_HasTag(t *testing.T) {
	te := NewTagEditor()
	te.CurrentTags = []string{"tag1", "Tag2", "TAG3"}

	tests := []struct {
		name    string
		tag     string
		wantHas bool
	}{
		{"Exact match lowercase", "tag1", true},
		{"Case insensitive match", "TAG1", true},
		{"Mixed case match", "tAg2", true},
		{"Not found", "tag4", false},
		{"Empty tag", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantHas, te.HasTag(tt.tag))
		})
	}
}


func TestTagEditor_RemoveTagAtCursor(t *testing.T) {
	te := NewTagEditor()
	te.CurrentTags = []string{"tag1", "tag2", "tag3"}

	// Remove middle tag
	te.TagCursor = 1
	te.RemoveTagAtCursor()
	assert.Equal(t, []string{"tag1", "tag3"}, te.CurrentTags)
	assert.Equal(t, 1, te.TagCursor)

	// Remove last tag
	te.TagCursor = 1
	te.RemoveTagAtCursor()
	assert.Equal(t, []string{"tag1"}, te.CurrentTags)
	assert.Equal(t, 0, te.TagCursor)

	// Try to remove with invalid cursor
	te.TagCursor = 5
	te.RemoveTagAtCursor()
	assert.Equal(t, []string{"tag1"}, te.CurrentTags)
}

func TestTagEditor_GetSuggestions(t *testing.T) {
	te := NewTagEditor()
	te.AvailableTags = []string{"alpha", "beta", "gamma", "alphabet", "better"}
	te.CurrentTags = []string{"beta"}

	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantFirst  string
	}{
		{
			name:      "Match prefix 'al'",
			input:     "al",
			wantCount: 2,
			wantFirst: "alpha",
		},
		{
			name:      "Match prefix 'bet'",
			input:     "bet",
			wantCount: 2, // "better" (prefix) and "alphabet" (contains)
			wantFirst: "better", // prefix matches come first
		},
		{
			name:      "No matches",
			input:     "xyz",
			wantCount: 0,
			wantFirst: "",
		},
		{
			name:      "Empty input",
			input:     "",
			wantCount: 0,
			wantFirst: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te.TagInput = tt.input
			suggestions := te.GetSuggestions()
			
			assert.Equal(t, tt.wantCount, len(suggestions))
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, suggestions[0])
			}
		})
	}
}

func TestTagEditor_GetAvailableTagsForCloud(t *testing.T) {
	te := NewTagEditor()
	te.AvailableTags = []string{"tag1", "tag2", "tag3", "tag4"}
	te.CurrentTags = []string{"tag2", "tag4"}

	available := te.GetAvailableTagsForCloud()
	
	assert.Equal(t, 2, len(available))
	assert.Contains(t, available, "tag1")
	assert.Contains(t, available, "tag3")
	assert.NotContains(t, available, "tag2")
	assert.NotContains(t, available, "tag4")
}

func TestTagEditor_HandleInput(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		setup       func(*TagEditor)
		wantHandled bool
		check       func(*testing.T, *TagEditor)
	}{
		{
			name: "Tab switches panes",
			key:  "tab",
			setup: func(te *TagEditor) {
				te.TagCloudActive = false
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.True(t, te.TagCloudActive)
			},
		},
		{
			name: "Enter adds tag from input",
			key:  "enter",
			setup: func(te *TagEditor) {
				te.TagInput = "newtag"
				te.CurrentTags = []string{"existing"}
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.Contains(t, te.CurrentTags, "newtag")
				assert.Empty(t, te.TagInput)
			},
		},
		{
			name: "Enter adds tag from cloud",
			key:  "enter",
			setup: func(te *TagEditor) {
				te.TagCloudActive = true
				te.AvailableTags = []string{"cloudtag"}
				te.TagCloudCursor = 0
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.Contains(t, te.CurrentTags, "cloudtag")
			},
		},
		{
			name: "Escape with no changes",
			key:  "esc",
			setup: func(te *TagEditor) {
				te.OriginalTags = []string{"tag1"}
				te.CurrentTags = []string{"tag1"}
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.False(t, te.Active)
			},
		},
		{
			name: "Escape with changes shows confirmation",
			key:  "esc",
			setup: func(te *TagEditor) {
				te.OriginalTags = []string{"tag1"}
				te.CurrentTags = []string{"tag1", "tag2"}
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.True(t, te.ExitConfirm.Active())
			},
		},
		{
			name: "Ctrl+D removes tag from current",
			key:  "ctrl+d",
			setup: func(te *TagEditor) {
				te.CurrentTags = []string{"tag1", "tag2"}
				te.TagCursor = 0
				te.TagCloudActive = false
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.Equal(t, 1, len(te.CurrentTags))
				assert.Equal(t, "tag2", te.CurrentTags[0])
			},
		},
		{
			name: "Arrow navigation",
			key:  "right",
			setup: func(te *TagEditor) {
				te.CurrentTags = []string{"tag1", "tag2"}
				te.TagCursor = 0
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.Equal(t, 1, te.TagCursor)
			},
		},
		{
			name: "Type character adds to input",
			key:  "a",
			setup: func(te *TagEditor) {
				te.TagInput = "t"
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.Equal(t, "ta", te.TagInput)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTagEditor()
			te.Active = true
			if tt.setup != nil {
				tt.setup(te)
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if len(tt.key) > 1 {
				// Special keys
				switch tt.key {
				case "tab":
					msg = tea.KeyMsg{Type: tea.KeyTab}
				case "enter":
					msg = tea.KeyMsg{Type: tea.KeyEnter}
				case "esc":
					msg = tea.KeyMsg{Type: tea.KeyEsc}
				case "ctrl+d":
					msg = tea.KeyMsg{Type: tea.KeyCtrlD}
				case "ctrl+s":
					msg = tea.KeyMsg{Type: tea.KeyCtrlS}
				case "right":
					msg = tea.KeyMsg{Type: tea.KeyRight}
				}
			}

			handled, _ := te.HandleInput(msg)
			assert.Equal(t, tt.wantHandled, handled)
			
			if tt.check != nil {
				tt.check(t, te)
			}
		})
	}
}

func TestTagEditor_SetSize(t *testing.T) {
	te := NewTagEditor()
	te.SetSize(100, 50)
	
	assert.Equal(t, 100, te.Width)
	assert.Equal(t, 50, te.Height)
}

func TestTagEditor_DeleteTagFromRegistry(t *testing.T) {
	te := NewTagEditor()
	te.DeletingTag = "test-tag"
	
	cmd := te.DeleteTagFromRegistry()
	assert.NotNil(t, cmd)
	assert.True(t, te.TagDeletionState.Active)
	assert.Equal(t, TagEditorModeDeleting, te.Mode)
}

func TestTagEditor_HandleMessage(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		setup       func(*TagEditor)
		wantHandled bool
		check       func(*testing.T, *TagEditor)
	}{
		{
			name: "Handle deletion complete",
			msg: tagDeletionCompleteMsg{
				Result: TagDeletionResult{
					TagName:      "deleted-tag",
					FilesUpdated: 5,
				},
			},
			setup: func(te *TagEditor) {
				te.Active = true
				te.TagDeletionState.Active = true
				te.Mode = TagEditorModeDeleting
				te.DeletingTag = "deleted-tag"
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.False(t, te.TagDeletionState.Active)
				assert.Equal(t, TagEditorModeNormal, te.Mode)
				assert.Empty(t, te.DeletingTag)
			},
		},
		{
			name: "Handle reload complete",
			msg:  tagReloadCompleteMsg{},
			setup: func(te *TagEditor) {
				te.Active = true
				te.Mode = TagEditorModeReloading
				te.TagReloader.Active = true
			},
			wantHandled: true,
			check: func(t *testing.T, te *TagEditor) {
				assert.False(t, te.TagReloader.Active)
				assert.Equal(t, TagEditorModeNormal, te.Mode)
			},
		},
		{
			name: "Ignore when inactive",
			msg:  tagDeletionCompleteMsg{},
			setup: func(te *TagEditor) {
				te.Active = false
			},
			wantHandled: false,
			check: func(t *testing.T, te *TagEditor) {
				assert.False(t, te.Active)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTagEditor()
			if tt.setup != nil {
				tt.setup(te)
			}

			handled, _ := te.HandleMessage(tt.msg)
			assert.Equal(t, tt.wantHandled, handled)
			
			if tt.check != nil {
				tt.check(t, te)
			}
		})
	}
}

func TestTagEditorRenderer_Render(t *testing.T) {
	te := NewTagEditor()
	te.Active = true
	te.CurrentTags = []string{"tag1", "tag2"}
	te.AvailableTags = []string{"tag3", "tag4"}
	
	renderer := NewTagEditorRenderer(te, 100, 30)
	output := renderer.Render()
	
	// Check that output contains expected elements
	assert.Contains(t, output, "EDIT TAGS")
	assert.Contains(t, output, "AVAILABLE TAGS")
	assert.Contains(t, output, "Current tags:")
	
	// Check help text is present
	assert.Contains(t, output, "tab")
	assert.Contains(t, output, "enter")
	assert.Contains(t, output, "esc")
}

func TestTagEditor_AddTagFromCloud(t *testing.T) {
	te := NewTagEditor()
	te.AvailableTags = []string{"cloudtag1", "cloudtag2"}
	te.CurrentTags = []string{"existing"}
	te.TagCloudCursor = 0
	
	te.AddTagFromCloud()
	
	assert.Contains(t, te.CurrentTags, "cloudtag1")
	assert.Equal(t, 2, len(te.CurrentTags))
	assert.Equal(t, 1, te.TagCursor) // Cursor moves to new tag
}


func TestFormatDeletionResult(t *testing.T) {
	tests := []struct {
		name     string
		result   TagDeletionResult
		expected string
	}{
		{
			name: "Successful deletion",
			result: TagDeletionResult{
				TagName:      "test-tag",
				FilesUpdated: 5,
				FilesScanned: 10,
				Errors:       []string{},
			},
			expected: "✓ Tag 'test-tag' deleted from registry and 5 files",
		},
		{
			name: "No files updated",
			result: TagDeletionResult{
				TagName:      "unused-tag",
				FilesUpdated: 0,
				FilesScanned: 10,
				Errors:       []string{},
			},
			expected: "✓ Tag 'unused-tag' removed from registry (not used in any files)",
		},
		{
			name: "With errors",
			result: TagDeletionResult{
				TagName:      "problem-tag",
				FilesUpdated: 2,
				FilesScanned: 10,
				Errors:       []string{"Error 1", "Error 2"},
			},
			expected: "× Tag 'problem-tag' partially deleted: 2 files updated, 2 errors:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDeletionResult(tt.result)
			assert.Contains(t, result, tt.expected[:20]) // Check prefix
		})
	}
}

// Benchmark tests
func BenchmarkTagEditor_GetSuggestions(b *testing.B) {
	te := NewTagEditor()
	te.AvailableTags = make([]string, 1000)
	for i := 0; i < 1000; i++ {
		te.AvailableTags[i] = "tag" + string(rune(i))
	}
	te.TagInput = "tag"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		te.GetSuggestions()
	}
}

func BenchmarkTagEditor_GetAvailableTagsForCloud(b *testing.B) {
	te := NewTagEditor()
	te.AvailableTags = make([]string, 100)
	for i := 0; i < 100; i++ {
		te.AvailableTags[i] = "available" + string(rune(i))
	}
	te.CurrentTags = make([]string, 50)
	for i := 0; i < 50; i++ {
		te.CurrentTags[i] = "current" + string(rune(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		te.GetAvailableTagsForCloud()
	}
}

// Helper function to compare tag slices ignoring order
func equalTagSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	aMap := make(map[string]bool)
	for _, tag := range a {
		normalized := models.NormalizeTagName(tag)
		aMap[normalized] = true
	}
	
	for _, tag := range b {
		normalized := models.NormalizeTagName(tag)
		if !aMap[normalized] {
			return false
		}
	}
	
	return true
}

func TestEqualTagSets(t *testing.T) {
	tests := []struct {
		name  string
		a     []string
		b     []string
		equal bool
	}{
		{"Empty sets", []string{}, []string{}, true},
		{"Same tags", []string{"a", "b"}, []string{"a", "b"}, true},
		{"Different order", []string{"a", "b"}, []string{"b", "a"}, true},
		{"Case insensitive", []string{"Tag1", "TAG2"}, []string{"tag1", "tag2"}, true},
		{"Different tags", []string{"a", "b"}, []string{"a", "c"}, false},
		{"Different lengths", []string{"a"}, []string{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, equalTagSets(tt.a, tt.b))
		})
	}
}