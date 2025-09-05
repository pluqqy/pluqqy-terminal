package tui

import (
	"strings"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// TagEditorDataStore manages all tag-related data including current state,
// original values for change detection, and available tags for suggestions
type TagEditorDataStore struct {
	// Tag collections
	CurrentTags   []string
	OriginalTags  []string
	AvailableTags []string
	
	// Deletion state data
	DeletingTag      string
	DeletingTagUsage *tags.UsageStats
}

// TagEditorStateManager manages core editor state and context information
// about where the editor was invoked from and what entity it's editing
type TagEditorStateManager struct {
	// Core state
	Active   bool
	Mode     TagEditorMode
	
	// Context information
	Path     string
	ItemType string // "component" or "pipeline"
	ItemName string
}

// TagEditorInputComponents handles text input, cursor positioning, 
// and the auto-completion suggestion system
type TagEditorInputComponents struct {
	// Text input state
	TagInput                string
	TagCursor               int
	
	// Suggestion system
	ShowSuggestions         bool
	SuggestionCursor        int
	HasNavigatedSuggestions bool
}

// TagEditorUIComponents manages UI state including tag cloud navigation,
// window dimensions, and callbacks for parent view communication
type TagEditorUIComponents struct {
	// Tag cloud navigation
	TagCloudActive bool
	TagCloudCursor int
	
	// Dimensions
	Width  int
	Height int
	
	// Callbacks
	Callbacks TagEditorCallbacks
}

// Helper methods for TagEditorDataStore
func (d *TagEditorDataStore) HasChanges() bool {
	if len(d.CurrentTags) != len(d.OriginalTags) {
		return true
	}
	
	// Create a map of original tags for quick lookup
	originalMap := make(map[string]bool)
	for _, tag := range d.OriginalTags {
		originalMap[tag] = true
	}
	
	// Check if all current tags exist in original
	for _, tag := range d.CurrentTags {
		if !originalMap[tag] {
			return true
		}
	}
	
	return false
}

func (d *TagEditorDataStore) HasTag(tag string) bool {
	for _, existingTag := range d.CurrentTags {
		if strings.EqualFold(existingTag, tag) {
			return true
		}
	}
	return false
}

func (d *TagEditorDataStore) GetAvailableTagsForCloud() []string {
	var available []string
	for _, tag := range d.AvailableTags {
		if !d.HasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

func (d *TagEditorDataStore) CopyTags(sourceTags []string) {
	d.CurrentTags = make([]string, len(sourceTags))
	copy(d.CurrentTags, sourceTags)
	d.OriginalTags = make([]string, len(sourceTags))
	copy(d.OriginalTags, sourceTags)
}

func (d *TagEditorDataStore) ResetTags() {
	d.CurrentTags = []string{}
	d.OriginalTags = []string{}
	d.DeletingTag = ""
	d.DeletingTagUsage = nil
}

func (d *TagEditorDataStore) UpdateOriginalTags() {
	// Update original tags to match current (used after save)
	d.OriginalTags = make([]string, len(d.CurrentTags))
	copy(d.OriginalTags, d.CurrentTags)
}

// Helper methods for TagEditorStateManager
func (s *TagEditorStateManager) Reset() {
	s.Active = false
	s.Mode = TagEditorModeNormal
	s.Path = ""
	s.ItemType = ""
	s.ItemName = ""
}

func (s *TagEditorStateManager) Start(path, itemType, itemName string) {
	s.Active = true
	s.Path = path
	s.ItemType = itemType
	s.ItemName = itemName
	s.Mode = TagEditorModeNormal
}

// Helper methods for TagEditorInputComponents  
func (i *TagEditorInputComponents) Reset() {
	i.TagInput = ""
	i.TagCursor = 0
	i.ShowSuggestions = false
	i.SuggestionCursor = 0
	i.HasNavigatedSuggestions = false
}

func (i *TagEditorInputComponents) UpdateSuggestions() {
	if i.TagInput == "" {
		i.ShowSuggestions = false
		i.SuggestionCursor = 0
		i.HasNavigatedSuggestions = false
	} else {
		i.ShowSuggestions = true
		i.SuggestionCursor = 0
		i.HasNavigatedSuggestions = false
	}
}

func (i *TagEditorInputComponents) ClearInput() {
	i.TagInput = ""
	i.ShowSuggestions = false
	i.SuggestionCursor = 0
	i.HasNavigatedSuggestions = false
}

// Helper methods for TagEditorUIComponents
func (u *TagEditorUIComponents) Reset() {
	u.TagCloudActive = false
	u.TagCloudCursor = 0
}

func (u *TagEditorUIComponents) SetSize(width, height int) {
	u.Width = width
	u.Height = height
}

func (u *TagEditorUIComponents) ToggleTagCloud() {
	u.TagCloudActive = !u.TagCloudActive
	if u.TagCloudActive {
		u.TagCloudCursor = 0
	}
}