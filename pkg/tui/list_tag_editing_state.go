package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
)

// TagEditor manages the state and logic for tag editing
type TagEditor struct {
	// Core state
	Active       bool
	Path         string
	ItemType     string // "component" or "pipeline"
	CurrentTags  []string
	OriginalTags []string

	// Input state
	TagInput                string
	TagCursor               int
	ShowSuggestions         bool
	SuggestionCursor        int
	HasNavigatedSuggestions bool // Track if user actively navigated suggestions

	// Tag cloud state
	TagCloudActive bool
	TagCloudCursor int
	AvailableTags  []string

	// Tag deletion state
	DeletingTag      string
	DeletingTagUsage *tags.UsageStats

	// Confirmation models
	TagDeleteConfirm *ConfirmationModel
	ExitConfirm      *ConfirmationModel
	PendingExit      bool // Flag to indicate exit should happen after confirmation

	// Tag reloader
	TagReloader *TagReloader

	// Dimensions
	Width int
}

// NewTagEditor creates a new tag editor instance
func NewTagEditor() *TagEditor {
	return &TagEditor{
		TagDeleteConfirm: NewConfirmation(),
		ExitConfirm:      NewConfirmation(),
		TagReloader:      NewTagReloader(),
	}
}

// Start initializes the tag editor for a specific item
func (t *TagEditor) Start(path string, currentTags []string, itemType string) {
	t.Active = true
	t.Path = path
	t.ItemType = itemType
	t.CurrentTags = make([]string, len(currentTags))
	copy(t.CurrentTags, currentTags)
	t.OriginalTags = make([]string, len(currentTags))
	copy(t.OriginalTags, currentTags)
	t.TagInput = ""
	t.TagCursor = 0
	t.ShowSuggestions = false
	t.SuggestionCursor = 0
	t.HasNavigatedSuggestions = false
	t.TagCloudActive = false
	t.TagCloudCursor = 0

	// Load available tags from registry
	t.LoadAvailableTags()
}

// Reset clears the tag editor state
func (t *TagEditor) Reset() {
	t.Active = false
	t.Path = ""
	t.ItemType = ""
	t.CurrentTags = nil
	t.OriginalTags = nil
	t.TagInput = ""
	t.TagCursor = 0
	t.ShowSuggestions = false
	t.SuggestionCursor = 0
	t.HasNavigatedSuggestions = false
	t.TagCloudActive = false
	t.TagCloudCursor = 0
	t.DeletingTag = ""
	t.DeletingTagUsage = nil
	t.PendingExit = false
}

// HasUnsavedChanges checks if tags have been modified
func (t *TagEditor) HasUnsavedChanges() bool {
	// Check if the number of tags has changed
	if len(t.CurrentTags) != len(t.OriginalTags) {
		return true
	}

	// Create maps for efficient comparison
	originalMap := make(map[string]bool)
	for _, tag := range t.OriginalTags {
		originalMap[tag] = true
	}

	// Check if all current tags exist in original
	for _, tag := range t.CurrentTags {
		if !originalMap[tag] {
			return true
		}
	}

	return false
}

// SetSize updates the width for proper dialog display
func (t *TagEditor) SetSize(width int) {
	t.Width = width
}

// LoadAvailableTags loads tags from the registry
func (t *TagEditor) LoadAvailableTags() {
	registry, err := tags.NewRegistry()
	if err != nil {
		t.AvailableTags = []string{}
		return
	}

	allTags := registry.ListTags()
	t.AvailableTags = make([]string, 0, len(allTags))
	for _, tag := range allTags {
		t.AvailableTags = append(t.AvailableTags, tag.Name)
	}
}

// HandleInput processes keyboard input for tag editing
func (t *TagEditor) HandleInput(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if !t.Active {
		return false, nil
	}

	// Handle tag deletion confirmation
	if t.TagDeleteConfirm.Active() {
		return true, t.TagDeleteConfirm.Update(msg)
	}

	// Handle exit confirmation
	if t.ExitConfirm.Active() {
		cmd := t.ExitConfirm.Update(msg)
		// Check if confirmation was completed by checking if dialog is no longer active
		if !t.ExitConfirm.Active() && t.PendingExit {
			// Dialog closed, check which key was pressed to determine action
			if msg.String() == "y" || msg.String() == "Y" {
				// User confirmed, exit without saving
				t.Reset()
			}
			// Clear pending flag regardless of choice
			t.PendingExit = false
		}
		return true, cmd
	}

	switch msg.String() {
	case "esc":
		// Check for unsaved changes
		if t.HasUnsavedChanges() {
			// Set pending exit flag
			t.PendingExit = true
			// Show exit confirmation
			width := t.Width
			if width == 0 {
				width = 80 // Default width if not set
			}
			t.ExitConfirm.Show(ConfirmationConfig{
				Title:       "⚠️  Unsaved Changes",
				Message:     "You have unsaved changes to tags.",
				Warning:     "Exit without saving?",
				Destructive: true,
				Type:        ConfirmTypeDialog,
				Width:       width - 4,
				Height:      10,
			}, func() tea.Cmd {
				// On confirm: The actual reset will happen in the Update handler above
				return nil
			}, func() tea.Cmd {
				// On cancel: The pending flag will be cleared in the Update handler above
				return nil
			})
			return true, nil
		} else {
			// No changes, exit directly
			t.Reset()
			return true, nil
		}

	case "ctrl+s":
		// Save tags
		return true, t.SaveTags()

	case "ctrl+t":
		// Reload tags from all components and pipelines
		if !t.TagReloader.IsActive() {
			return true, t.TagReloader.Start()
		}
		return true, nil

	case "enter":
		if t.TagCloudActive {
			// Add tag from cloud
			t.AddTagFromCloud()
		} else if t.ShowSuggestions && t.TagInput != "" {
			// Only use suggestion if user actively navigated to select it
			suggestions := t.GetSuggestions()
			if t.HasNavigatedSuggestions && len(suggestions) > 0 && t.SuggestionCursor < len(suggestions) {
				t.AddSelectedSuggestion()
			} else {
				// User just typed and hit enter - add exactly what they typed
				t.AddTagFromInput()
			}
		} else {
			// Add tag from input
			t.AddTagFromInput()
		}
		return true, nil

	case "tab":
		// Toggle between main pane and tag cloud
		t.TagCloudActive = !t.TagCloudActive
		if t.TagCloudActive {
			t.TagCloudCursor = 0
			t.ShowSuggestions = false
		} else {
			t.TagInput = ""
		}
		return true, nil

	case "up":
		if !t.TagCloudActive && t.ShowSuggestions && t.TagInput != "" {
			// Navigate up in suggestions
			if t.SuggestionCursor > 0 {
				t.SuggestionCursor--
			}
			// Mark as navigated even if cursor doesn't move (for single suggestion case)
			t.HasNavigatedSuggestions = true
		}
		return true, nil

	case "down":
		if !t.TagCloudActive && t.ShowSuggestions && t.TagInput != "" {
			// Navigate down in suggestions
			suggestions := t.GetSuggestions()
			maxSuggestions := len(suggestions)
			if maxSuggestions > 6 {
				maxSuggestions = 6 // Limit to 6 suggestions as per the view
			}
			if t.SuggestionCursor < maxSuggestions-1 {
				t.SuggestionCursor++
			}
			// Mark as navigated even if cursor doesn't move (for single suggestion case)
			t.HasNavigatedSuggestions = true
		}
		return true, nil

	case "left":
		if t.TagCloudActive {
			// Navigate in tag cloud
			if t.TagCloudCursor > 0 {
				t.TagCloudCursor--
			}
		} else {
			// Move cursor left in current tags
			if t.TagInput == "" && t.TagCursor > 0 {
				t.TagCursor--
			}
		}
		return true, nil

	case "right":
		if t.TagCloudActive {
			// Navigate in tag cloud
			availableForSelection := t.GetAvailableTagsForCloud()
			if t.TagCloudCursor < len(availableForSelection)-1 {
				t.TagCloudCursor++
			}
		} else {
			// Move cursor right in current tags
			if t.TagInput == "" && t.TagCursor < len(t.CurrentTags)-1 {
				t.TagCursor++
			}
		}
		return true, nil

	case "home":
		if t.TagCloudActive {
			// Jump to first tag in cloud
			t.TagCloudCursor = 0
		} else {
			// Jump to first tag in current tags
			if t.TagInput == "" && len(t.CurrentTags) > 0 {
				t.TagCursor = 0
			}
		}
		return true, nil

	case "end":
		if t.TagCloudActive {
			// Jump to last tag in cloud
			availableForSelection := t.GetAvailableTagsForCloud()
			if len(availableForSelection) > 0 {
				t.TagCloudCursor = len(availableForSelection) - 1
			}
		} else {
			// Jump to last tag in current tags
			if t.TagInput == "" && len(t.CurrentTags) > 0 {
				t.TagCursor = len(t.CurrentTags) - 1
			}
		}
		return true, nil

	case "ctrl+d":
		if t.TagCloudActive {
			// Delete tag from registry
			cmd = t.HandleTagRegistryDeletion()
		} else {
			// Remove tag from current item
			t.RemoveCurrentTag()
		}
		return true, cmd

	case "backspace", "delete":
		if !t.TagCloudActive && t.TagInput != "" {
			// Delete from input
			if len(t.TagInput) > 0 {
				t.TagInput = t.TagInput[:len(t.TagInput)-1]
				t.ShowSuggestions = len(t.TagInput) > 0
				t.SuggestionCursor = 0 // Reset to first suggestion
				t.HasNavigatedSuggestions = false // Reset navigation flag when editing
			}
		}
		return true, nil

	default:
		// Add to input only when in main pane
		if !t.TagCloudActive && len(msg.String()) == 1 {
			t.TagInput += msg.String()
			t.ShowSuggestions = true
			t.SuggestionCursor = 0 // Reset to first suggestion
			t.HasNavigatedSuggestions = false // Reset navigation flag when typing
		}
		return true, nil
	}
}

// AddSelectedSuggestion adds the currently selected suggestion
func (t *TagEditor) AddSelectedSuggestion() {
	suggestions := t.GetSuggestions()
	if t.SuggestionCursor >= 0 && t.SuggestionCursor < len(suggestions) {
		// Limit to 6 suggestions as per the view
		maxIndex := len(suggestions)
		if maxIndex > 6 {
			maxIndex = 6
		}
		if t.SuggestionCursor < maxIndex {
			tag := suggestions[t.SuggestionCursor]
			if !t.HasTag(tag) {
				t.CurrentTags = append(t.CurrentTags, tag)
			}
			t.TagInput = ""
			t.ShowSuggestions = false
			t.SuggestionCursor = 0
			t.HasNavigatedSuggestions = false
		}
	}
}

// AddTagFromInput adds a tag from the input field
func (t *TagEditor) AddTagFromInput() {
	if t.TagInput != "" {
		// Normalize the tag
		normalized := models.NormalizeTagName(t.TagInput)
		if normalized != "" && !t.HasTag(normalized) {
			t.CurrentTags = append(t.CurrentTags, normalized)
		}
		t.TagInput = ""
		t.ShowSuggestions = false
		t.SuggestionCursor = 0
		t.HasNavigatedSuggestions = false
	}
}

// AddTagFromCloud adds a tag from the tag cloud
func (t *TagEditor) AddTagFromCloud() {
	availableForSelection := t.GetAvailableTagsForCloud()
	if t.TagCloudCursor >= 0 && t.TagCloudCursor < len(availableForSelection) {
		tag := availableForSelection[t.TagCloudCursor]
		if !t.HasTag(tag) {
			t.CurrentTags = append(t.CurrentTags, tag)
		}
	}
}

// RemoveCurrentTag removes the currently selected tag
func (t *TagEditor) RemoveCurrentTag() {
	if t.TagInput == "" && len(t.CurrentTags) > 0 {
		if t.TagCursor >= 0 && t.TagCursor < len(t.CurrentTags) {
			t.CurrentTags = append(t.CurrentTags[:t.TagCursor], t.CurrentTags[t.TagCursor+1:]...)
			if t.TagCursor >= len(t.CurrentTags) && t.TagCursor > 0 {
				t.TagCursor--
			}
		}
	}
}

// HandleTagRegistryDeletion handles the deletion of a tag from the registry
func (t *TagEditor) HandleTagRegistryDeletion() tea.Cmd {
	availableForSelection := t.GetAvailableTagsForCloud()
	if t.TagCloudCursor >= 0 && t.TagCloudCursor < len(availableForSelection) {
		tagToDelete := availableForSelection[t.TagCloudCursor]

		// Get usage stats
		usage, err := tags.CountTagUsage(tagToDelete)
		if err != nil {
			return func() tea.Msg {
				return StatusMsg(fmt.Sprintf("× Failed to check tag usage: %v", err))
			}
		}

		t.DeletingTag = tagToDelete
		t.DeletingTagUsage = usage

		// Show tag deletion confirmation with details
		var details []string
		if usage.PipelineCount > 0 {
			details = append(details, fmt.Sprintf("Used in %d pipeline(s)", usage.PipelineCount))
		}
		if usage.ComponentCount > 0 {
			details = append(details, fmt.Sprintf("Used in %d component(s)", usage.ComponentCount))
		}

		warning := ""
		if usage.PipelineCount > 0 || usage.ComponentCount > 0 {
			warning = "The tag will be removed from the registry but will remain on items that use it."
		}

		// Configure and show confirmation
		t.TagDeleteConfirm.Show(ConfirmationConfig{
			Title:       "⚠️  Delete Tag from Registry?",
			Message:     fmt.Sprintf("Delete tag '%s'?", tagToDelete),
			Warning:     warning,
			Details:     details,
			Destructive: true,
			Type:        ConfirmTypeDialog,
			Width:       80,
			Height:      12,
		}, func() tea.Cmd {
			return t.DeleteTagFromRegistry()
		}, func() tea.Cmd {
			t.DeletingTag = ""
			t.DeletingTagUsage = nil
			return nil
		})
	}

	return nil
}

// DeleteTagFromRegistry deletes a tag from the registry
func (t *TagEditor) DeleteTagFromRegistry() tea.Cmd {
	return func() tea.Msg {
		// Delete from registry
		registry, err := tags.NewRegistry()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to load tag registry: %v", err))
		}

		registry.RemoveTag(t.DeletingTag)

		if err := registry.Save(); err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save tag registry: %v", err))
		}

		// Update available tags
		t.LoadAvailableTags()

		// Adjust cursor if needed
		availableForCloud := t.GetAvailableTagsForCloud()
		if t.TagCloudCursor >= len(availableForCloud) && t.TagCloudCursor > 0 {
			t.TagCloudCursor = len(availableForCloud) - 1
		}

		// Clear deletion state
		deletedTag := t.DeletingTag
		t.DeletingTag = ""
		t.DeletingTagUsage = nil

		return StatusMsg(fmt.Sprintf("✓ Deleted tag '%s' from registry", deletedTag))
	}
}

// SaveTags saves the current tags to the file
func (t *TagEditor) SaveTags() tea.Cmd {
	return func() tea.Msg {
		var err error

		if t.ItemType == "component" {
			// Update component tags
			err = files.UpdateComponentTags(t.Path, t.CurrentTags)
		} else {
			// Update pipeline tags
			pipeline, err := files.ReadPipeline(t.Path)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to read pipeline: %v", err))
			}
			pipeline.Tags = t.CurrentTags
			err = files.WritePipeline(pipeline)
		}

		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save tags: %v", err))
		}

		// Update the registry with any new tags
		registry, _ := tags.NewRegistry()
		if registry != nil {
			for _, tag := range t.CurrentTags {
				registry.GetOrCreateTag(tag)
			}
			registry.Save()
		}

		// Reset the editor
		t.Reset()

		return ReloadMsg{Message: "✓ Tags saved"}
	}
}

// HasTag checks if a tag is already in the current tags
func (t *TagEditor) HasTag(tag string) bool {
	for _, existingTag := range t.CurrentTags {
		if strings.EqualFold(existingTag, tag) {
			return true
		}
	}
	return false
}

// GetSuggestions returns tag suggestions based on input
func (t *TagEditor) GetSuggestions() []string {
	if t.TagInput == "" {
		return []string{}
	}

	input := strings.ToLower(t.TagInput)
	var suggestions []string

	// First, exact prefix matches
	for _, tag := range t.AvailableTags {
		if strings.HasPrefix(strings.ToLower(tag), input) && !t.HasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}

	// Then, contains matches
	for _, tag := range t.AvailableTags {
		lowerTag := strings.ToLower(tag)
		if !strings.HasPrefix(lowerTag, input) && strings.Contains(lowerTag, input) && !t.HasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}

	return suggestions
}

// GetAvailableTagsForCloud returns tags available for selection in the cloud
func (t *TagEditor) GetAvailableTagsForCloud() []string {
	var available []string
	for _, tag := range t.AvailableTags {
		if !t.HasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

// HasChanges returns true if tags have been modified
func (t *TagEditor) HasChanges() bool {
	if len(t.CurrentTags) != len(t.OriginalTags) {
		return true
	}

	// Create a map of original tags for quick lookup
	originalMap := make(map[string]bool)
	for _, tag := range t.OriginalTags {
		originalMap[tag] = true
	}

	// Check if all current tags exist in original
	for _, tag := range t.CurrentTags {
		if !originalMap[tag] {
			return true
		}
	}

	return false
}

// HandleMessage processes messages related to tag reloading
func (t *TagEditor) HandleMessage(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	if !t.Active {
		return false, nil
	}

	// Handle tag reload messages if reloader is active
	if t.TagReloader != nil {
		switch msg := msg.(type) {
		case TagReloadMsg:
			handled, cmd := t.TagReloader.HandleMessage(msg)
			if handled {
				// Reload available tags after successful reload
				if msg.Error == nil {
					t.LoadAvailableTags()
				}
				return true, cmd
			}
		case tagReloadCompleteMsg:
			t.TagReloader.HandleComplete()
			return true, nil
		}
	}

	return false, nil
}
