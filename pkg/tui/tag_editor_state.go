package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// TagEditor is a reusable component for editing tags
type TagEditor struct {
	// Embedded state
	*TagEditorState
	
	// Sub-components
	TagDeleteConfirm *ConfirmationModel
	ExitConfirm      *ConfirmationModel
	TagReloader      *TagReloader
	TagDeletionState *TagDeletionState
}

// NewTagEditor creates a new tag editor instance
func NewTagEditor() *TagEditor {
	return &TagEditor{
		TagEditorState: &TagEditorState{
			TagEditorDataStore: &TagEditorDataStore{
				CurrentTags:   []string{},
				OriginalTags:  []string{},
				AvailableTags: []string{},
			},
			TagEditorStateManager: &TagEditorStateManager{},
			TagEditorInputComponents: &TagEditorInputComponents{},
			TagEditorUIComponents: &TagEditorUIComponents{},
		},
		TagDeleteConfirm: NewConfirmation(),
		ExitConfirm:      NewConfirmation(),
		TagReloader:      NewTagReloader(),
		TagDeletionState: NewTagDeletionState(),
	}
}

// Start initializes the tag editor for a specific item
func (te *TagEditor) Start(path string, currentTags []string, itemType, itemName string) {
	// Initialize state using composition helper methods
	te.TagEditorStateManager.Start(path, itemType, itemName)
	te.TagEditorDataStore.CopyTags(currentTags)
	te.TagEditorInputComponents.Reset()
	te.TagEditorUIComponents.Reset()
	
	// Load available tags from registry
	te.LoadAvailableTags()
}

// Reset clears the tag editor state
func (te *TagEditor) Reset() {
	// Reset using composition helper methods
	te.TagEditorStateManager.Reset()
	te.TagEditorDataStore.ResetTags()
	te.TagEditorInputComponents.Reset()
	te.TagEditorUIComponents.Reset()
	
	// Reset tag reloader if it exists
	if te.TagReloader != nil {
		te.TagReloader.Reset()
	}
	
	// Reset tag deletion state if it exists
	if te.TagDeletionState != nil {
		te.TagDeletionState.Active = false
	}
}

// LoadAvailableTags loads tags from the registry
func (te *TagEditor) LoadAvailableTags() {
	registry, err := tags.NewRegistry()
	if err != nil {
		te.AvailableTags = []string{}
		return
	}
	
	allTags := registry.ListTags()
	te.AvailableTags = make([]string, 0, len(allTags))
	for _, tag := range allTags {
		te.AvailableTags = append(te.AvailableTags, tag.Name)
	}
}

// HasChanges returns true if tags have been modified
func (te *TagEditor) HasChanges() bool {
	return te.TagEditorDataStore.HasChanges()
}

// HandleInput processes keyboard input for tag editing
func (te *TagEditor) HandleInput(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if !te.Active {
		return false, nil
	}
	
	// Handle confirmation dialogs first
	if te.TagDeleteConfirm.Active() {
		cmd := te.TagDeleteConfirm.Update(msg)
		return true, cmd
	}
	
	if te.ExitConfirm.Active() {
		cmd := te.ExitConfirm.Update(msg)
		return true, cmd
	}
	
	// Handle normal input based on current mode
	switch msg.String() {
	case "esc":
		if te.TagInput != "" {
			// Clear input
			te.TagInput = ""
			te.ShowSuggestions = false
			te.SuggestionCursor = 0
			te.HasNavigatedSuggestions = false
		} else if te.HasChanges() {
			// Show exit confirmation if there are unsaved changes
			te.ExitConfirm.Show(ConfirmationConfig{
				Title:       "⚠️  Unsaved Changes",
				Message:     "You have unsaved tag changes.",
				Warning:     "Exit without saving?",
				YesLabel:    "Exit",
				NoLabel:     "Cancel",
				Destructive: true,
				Type:        ConfirmTypeDialog,
				Width:       te.Width - 4,
				Height:      10,
			}, func() tea.Cmd {
				// onConfirm - user chose to exit without saving
				te.Reset()
				if te.Callbacks.OnExit != nil {
					te.Callbacks.OnExit(false)
				}
				return nil
			}, func() tea.Cmd {
				// onCancel - user chose to stay and keep editing
				return nil
			})
		} else {
			// No changes, exit directly
			te.Reset()
			if te.Callbacks.OnExit != nil {
				te.Callbacks.OnExit(false)
			}
		}
		return true, nil
		
	case "ctrl+s":
		// Save tags
		return true, te.Save()
		
	case "ctrl+t":
		// Reload tags from all components and pipelines
		if !te.TagReloader.IsActive() {
			te.Mode = TagEditorModeReloading
			return true, te.TagReloader.Start()
		}
		return true, nil
		
	case "ctrl+d":
		if te.TagCloudActive {
			// Delete tag from cloud
			return true, te.StartTagDeletion()
		} else if te.TagInput == "" && len(te.CurrentTags) > 0 {
			// Remove tag at cursor
			te.RemoveTagAtCursor()
		}
		return true, nil
		
	case "enter":
		if te.TagCloudActive {
			// Add tag from cloud
			te.AddTagFromCloud()
		} else if te.ShowSuggestions && te.TagInput != "" {
			// Add tag from suggestions or input
			suggestions := te.GetSuggestions()
			if te.HasNavigatedSuggestions && len(suggestions) > 0 && te.SuggestionCursor < len(suggestions) {
				te.AddSelectedSuggestion()
			} else {
				te.AddTagFromInput()
			}
		} else {
			// Add tag from input
			te.AddTagFromInput()
		}
		return true, nil
		
	case "tab":
		// Toggle between main pane and tag cloud
		te.TagCloudActive = !te.TagCloudActive
		if te.TagCloudActive {
			te.TagCloudCursor = 0
			te.ShowSuggestions = false
		} else {
			te.TagInput = ""
		}
		return true, nil
		
	case "up":
		if !te.TagCloudActive && te.ShowSuggestions && te.TagInput != "" {
			// Navigate up in suggestions
			if te.SuggestionCursor > 0 {
				te.SuggestionCursor--
			}
			te.HasNavigatedSuggestions = true
		}
		return true, nil
		
	case "down":
		if !te.TagCloudActive && te.ShowSuggestions && te.TagInput != "" {
			// Navigate down in suggestions
			suggestions := te.GetSuggestions()
			maxSuggestions := len(suggestions)
			if maxSuggestions > 6 {
				maxSuggestions = 6 // Limit to 6 suggestions
			}
			if te.SuggestionCursor < maxSuggestions-1 {
				te.SuggestionCursor++
			}
			te.HasNavigatedSuggestions = true
		}
		return true, nil
		
	case "left":
		if te.TagCloudActive {
			// Navigate in tag cloud
			if te.TagCloudCursor > 0 {
				te.TagCloudCursor--
			}
		} else {
			// Move cursor left in current tags
			if te.TagInput == "" && te.TagCursor > 0 {
				te.TagCursor--
			}
		}
		return true, nil
		
	case "right":
		if te.TagCloudActive {
			// Navigate in tag cloud
			availableForSelection := te.GetAvailableTagsForCloud()
			if te.TagCloudCursor < len(availableForSelection)-1 {
				te.TagCloudCursor++
			}
		} else {
			// Move cursor right in current tags
			if te.TagInput == "" && te.TagCursor < len(te.CurrentTags)-1 {
				te.TagCursor++
			}
		}
		return true, nil
		
	case "home":
		if te.TagCloudActive {
			te.TagCloudCursor = 0
		} else if te.TagInput == "" {
			te.TagCursor = 0
		}
		return true, nil
		
	case "end":
		if te.TagCloudActive {
			availableForSelection := te.GetAvailableTagsForCloud()
			if len(availableForSelection) > 0 {
				te.TagCloudCursor = len(availableForSelection) - 1
			}
		} else if te.TagInput == "" {
			if len(te.CurrentTags) > 0 {
				te.TagCursor = len(te.CurrentTags) - 1
			}
		}
		return true, nil
		
	case "backspace":
		if !te.TagCloudActive {
			if te.TagInput != "" {
				// Remove last character from input
				if len(te.TagInput) > 0 {
					te.TagInput = te.TagInput[:len(te.TagInput)-1]
					te.UpdateSuggestions()
				}
			} else if len(te.CurrentTags) > 0 {
				// Remove tag at cursor
				te.RemoveTagAtCursor()
			}
		}
		return true, nil
		
	default:
		// Handle text input when not in tag cloud
		if !te.TagCloudActive && msg.Type == tea.KeyRunes {
			te.TagInput += string(msg.Runes)
			te.UpdateSuggestions()
			return true, nil
		}
	}
	
	return false, nil
}

// UpdateSuggestions updates the suggestion list based on current input
func (te *TagEditor) UpdateSuggestions() {
	te.TagEditorInputComponents.UpdateSuggestions()
}

// GetSuggestions returns tag suggestions based on input
func (te *TagEditor) GetSuggestions() []string {
	if te.TagInput == "" {
		return []string{}
	}
	
	input := strings.ToLower(te.TagInput)
	var suggestions []string
	
	// First, exact prefix matches
	for _, tag := range te.AvailableTags {
		if strings.HasPrefix(strings.ToLower(tag), input) && !te.TagEditorDataStore.HasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	// Then, contains matches
	for _, tag := range te.AvailableTags {
		lowerTag := strings.ToLower(tag)
		if !strings.HasPrefix(lowerTag, input) && strings.Contains(lowerTag, input) && !te.TagEditorDataStore.HasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	return suggestions
}

// HasTag checks if a tag is already in the current tags
func (te *TagEditor) HasTag(tag string) bool {
	return te.TagEditorDataStore.HasTag(tag)
}

// AddTagFromInput adds a tag from the current input
func (te *TagEditor) AddTagFromInput() {
	if te.TagInput == "" {
		return
	}
	
	// Normalize the tag name
	tagName := models.NormalizeTagName(te.TagInput)
	
	// Check if tag already exists
	if te.HasTag(tagName) {
		te.TagEditorInputComponents.ClearInput()
		return
	}
	
	// Add to current tags
	te.CurrentTags = append(te.CurrentTags, tagName)
	te.TagCursor = len(te.CurrentTags) - 1
	
	// Clear input
	te.TagEditorInputComponents.ClearInput()
}

// AddSelectedSuggestion adds the currently selected suggestion
func (te *TagEditor) AddSelectedSuggestion() {
	suggestions := te.GetSuggestions()
	if te.SuggestionCursor < len(suggestions) {
		tagName := suggestions[te.SuggestionCursor]
		if !te.HasTag(tagName) {
			te.CurrentTags = append(te.CurrentTags, tagName)
			te.TagCursor = len(te.CurrentTags) - 1
		}
		// Clear input
		te.TagEditorInputComponents.ClearInput()
	}
}

// AddTagFromCloud adds the selected tag from the cloud
func (te *TagEditor) AddTagFromCloud() {
	availableForSelection := te.GetAvailableTagsForCloud()
	if te.TagCloudCursor < len(availableForSelection) {
		tagName := availableForSelection[te.TagCloudCursor]
		if !te.HasTag(tagName) {
			te.CurrentTags = append(te.CurrentTags, tagName)
			te.TagCursor = len(te.CurrentTags) - 1
		}
	}
}

// RemoveTagAtCursor removes the tag at the current cursor position
func (te *TagEditor) RemoveTagAtCursor() {
	if te.TagCursor < len(te.CurrentTags) {
		// Remove tag at cursor
		te.CurrentTags = append(te.CurrentTags[:te.TagCursor], te.CurrentTags[te.TagCursor+1:]...)
		// Adjust cursor
		if te.TagCursor >= len(te.CurrentTags) && te.TagCursor > 0 {
			te.TagCursor = len(te.CurrentTags) - 1
		}
	}
}

// GetAvailableTagsForCloud returns tags available for selection in the cloud
func (te *TagEditor) GetAvailableTagsForCloud() []string {
	return te.TagEditorDataStore.GetAvailableTagsForCloud()
}

// StartTagDeletion starts the process of deleting a tag from the registry
func (te *TagEditor) StartTagDeletion() tea.Cmd {
	availableForCloud := te.GetAvailableTagsForCloud()
	if te.TagCloudCursor >= len(availableForCloud) {
		return nil
	}
	
	tagToDelete := availableForCloud[te.TagCloudCursor]
	te.DeletingTag = tagToDelete
	
	// Get usage stats for the tag
	usage, err := tags.CountTagUsage(tagToDelete)
	if err == nil {
		te.DeletingTagUsage = usage
		
		// Show confirmation dialog
		usageCount := usage.ComponentCount + usage.PipelineCount
		message := fmt.Sprintf(
			"Delete tag '%s'?\n\nThis tag is used in %d item%s.\nIt will be removed from all files and the registry.",
			tagToDelete,
			usageCount,
			func() string {
				if usageCount == 1 {
					return ""
				}
				return "s"
			}(),
		)
		
		te.TagDeleteConfirm.Show(ConfirmationConfig{
			Title:       "⚠️  Delete Tag",
			Message:     message,
			YesLabel:    "Delete",
			NoLabel:     "Cancel",
			Destructive: true,
			Type:        ConfirmTypeDialog,
			Width:       te.Width - 4,
			Height:      10,
		}, func() tea.Cmd {
			return te.DeleteTagFromRegistry()
		}, func() tea.Cmd {
			te.DeletingTag = ""
			te.DeletingTagUsage = nil
			return nil
		})
	}
	
	return nil
}

// DeleteTagFromRegistry deletes a tag from the registry and all files
func (te *TagEditor) DeleteTagFromRegistry() tea.Cmd {
	// Start the deletion spinner
	te.TagDeletionState.Start()
	te.Mode = TagEditorModeDeleting
	
	// Create progress callback
	progressCallback := func(currentFile string, progress int, total int) {
		// Progress updates will be handled through messages
	}
	
	// Return the comprehensive deletion command
	return DeleteTagCompletely(te.DeletingTag, progressCallback)
}

// HandleMessage processes messages related to tag editing
func (te *TagEditor) HandleMessage(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	if !te.Active {
		return false, nil
	}
	
	// Handle specific message types first, before delegating to sub-components
	switch msg := msg.(type) {
	case tagDeletionCompleteMsg:
		te.TagDeletionState.Active = false
		te.Mode = TagEditorModeNormal
		
		// Clear deletion state
		te.DeletingTag = ""
		te.DeletingTagUsage = nil
		
		// Clear the available tags completely first
		te.AvailableTags = []string{}
		
		// Then reload from registry to get the updated state
		// The registry has been updated by DeleteTagCompletely
		te.LoadAvailableTags()
		
		// Adjust cursor if needed
		availableForCloud := te.GetAvailableTagsForCloud()
		if te.TagCloudCursor >= len(availableForCloud) && te.TagCloudCursor > 0 {
			te.TagCloudCursor = len(availableForCloud) - 1
		}
		
		// Return nil to trigger re-render
		return true, nil
		
	case tagDeletionProgressMsg:
		if te.TagDeletionState != nil {
			te.TagDeletionState.Update(msg)
		}
		return true, nil
		
	case spinner.TickMsg:
		if te.TagDeletionState != nil && te.TagDeletionState.Active {
			handled, cmd := te.TagDeletionState.Update(msg)
			if handled {
				return true, cmd
			}
		}
	}
	
	// Handle tag reload messages if reloader is active
	if te.TagReloader != nil {
		switch msg := msg.(type) {
		case TagReloadMsg:
			handled, cmd := te.TagReloader.HandleMessage(msg)
			if handled {
				te.Mode = TagEditorModeNormal
				// Reload available tags after successful reload
				if msg.Error == nil {
					te.LoadAvailableTags()
				}
				return true, cmd
			}
		case tagReloadCompleteMsg:
			te.TagReloader.HandleComplete()
			te.Mode = TagEditorModeNormal
			return true, nil
		}
	}
	
	return false, nil
}

// SetSize updates the dimensions of the tag editor
func (te *TagEditor) SetSize(width, height int) {
	te.TagEditorUIComponents.SetSize(width, height)
}