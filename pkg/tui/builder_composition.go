package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// BuilderDataStore manages all component and pipeline data
type BuilderDataStore struct {
	// Core pipeline data
	Pipeline *models.Pipeline

	// Available components (left column)
	Prompts  []componentItem
	Contexts []componentItem
	Rules    []componentItem

	// Selected components (right column)
	SelectedComponents []models.ComponentRef

	// Filtered components (after search)
	FilteredPrompts  []componentItem
	FilteredContexts []componentItem
	FilteredRules    []componentItem

	// Change tracking
	OriginalComponents []models.ComponentRef // Original components for existing pipelines
}

// BuilderViewportManager manages all UI viewports and dimensions
type BuilderViewportManager struct {
	Preview       viewport.Model
	LeftTable     *ComponentTableRenderer // For available components table
	RightViewport viewport.Model          // For selected components
	Width         int
	Height        int
}

// BuilderEditorComponents groups all editor-related functionality
type BuilderEditorComponents struct {
	Enhanced         *EnhancedEditorState
	TagEditor        *TagEditor
	Rename           *BuilderRenameComponents
	Clone            *BuilderCloneComponents
	ComponentCreator *BuilderComponentCreator

	// Component editing state
	EditingComponent     bool
	EditingComponentPath string
	EditingComponentName string
	ComponentContent     string // Content being edited
	EditSaveMessage      string
	EditSaveTimer        *time.Timer
	OriginalContent      string // Track original content for unsaved changes

	// Name input state
	EditingName bool
	NameInput   string
}

// BuilderRenameComponents groups rename-related components
type BuilderRenameComponents struct {
	State    *RenameState
	Renderer *RenameRenderer
	Operator *RenameOperator
}

// BuilderCloneComponents groups clone-related components
type BuilderCloneComponents struct {
	State    *CloneState
	Renderer *CloneRenderer
	Operator *CloneOperator
}

// BuilderSearchComponents groups search-related functionality
type BuilderSearchComponents struct {
	Engine       *search.Engine
	Bar          *SearchBar
	Query        string
	FilterHelper *SearchFilterHelper
}

// BuilderUIComponents groups UI-specific components
type BuilderUIComponents struct {
	// Dialogs
	ExitConfirm          *ConfirmationModel
	ExitConfirmationType string // "pipeline" or "component" or "tags"
	DeleteConfirm        *ConfirmationModel
	ArchiveConfirm       *ConfirmationModel

	// UI state
	ActiveColumn   column
	LeftCursor     int
	RightCursor    int
	ShowPreview    bool
	PreviewContent string
	StatusMessage  string
	SharedLayout   *SharedLayout

	// Mermaid state
	MermaidState    *MermaidState
	MermaidOperator *MermaidOperator

	// Legacy tag UI state (TO BE REMOVED after full migration)
	EditingTags             bool
	EditingTagsPath         string
	CurrentTags             []string
	OriginalTags            []string
	TagInput                string
	TagCursor               int
	AvailableTags           []string
	ShowTagSuggestions      bool
	TagSuggestionCursor     int
	HasNavigatedSuggestions bool
	TagCloudActive          bool
	TagCloudCursor          int
	TagDeleteConfirm        *ConfirmationModel
	DeletingTag             string
	DeletingTagUsage        []string
}

// Helper methods for BuilderDataStore
func (d *BuilderDataStore) GetAllAvailableComponents() []componentItem {
	allComponents := make([]componentItem, 0, len(d.Prompts)+len(d.Contexts)+len(d.Rules))
	allComponents = append(allComponents, d.Prompts...)
	allComponents = append(allComponents, d.Contexts...)
	allComponents = append(allComponents, d.Rules...)
	return allComponents
}

func (d *BuilderDataStore) GetFilteredComponents() []componentItem {
	allFiltered := make([]componentItem, 0, len(d.FilteredPrompts)+len(d.FilteredContexts)+len(d.FilteredRules))
	allFiltered = append(allFiltered, d.FilteredPrompts...)
	allFiltered = append(allFiltered, d.FilteredContexts...)
	allFiltered = append(allFiltered, d.FilteredRules...)
	return allFiltered
}

func (d *BuilderDataStore) HasUnsavedPipelineChanges() bool {
	// For new pipelines, check if we have any components
	if d.Pipeline == nil && len(d.SelectedComponents) > 0 {
		return true
	}

	// For existing pipelines, check if components have changed
	if d.Pipeline != nil {
		if len(d.SelectedComponents) != len(d.OriginalComponents) {
			return true
		}
		for i, comp := range d.SelectedComponents {
			if comp.Path != d.OriginalComponents[i].Path {
				return true
			}
		}
	}

	return false
}


// Helper methods for BuilderViewportManager
func (v *BuilderViewportManager) UpdateSizes(width, height int, showPreview bool) {
	v.Width = width
	v.Height = height

	// Calculate dimensions
	searchBarHeight := 3
	nameInputHeight := 3
	footerHeight := 10
	contentHeight := height - footerHeight - searchBarHeight - nameInputHeight

	if showPreview {
		// Split horizontally: components above, preview below
		contentHeight = contentHeight / 2
		
		// Preview takes the bottom half
		previewHeight := height/2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		v.Preview.Width = width - 8
		v.Preview.Height = previewHeight
	}

	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Calculate column widths (50/50 split)
	columnWidth := (width - 6) / 2

	// Update viewports for components
	viewportHeight := contentHeight - 4
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	// Left table renderer width
	if v.LeftTable != nil {
		v.LeftTable.Width = columnWidth - 4
		v.LeftTable.Height = viewportHeight
	}

	// Right viewport for selected components
	v.RightViewport.Width = columnWidth - 4
	v.RightViewport.Height = viewportHeight
}

// Helper methods for BuilderEditorComponents
func (e *BuilderEditorComponents) IsAnyEditorActive() bool {
	return e.EditingComponent ||
		e.EditingName ||
		e.Enhanced.Active ||
		e.TagEditor.Active ||
		e.Rename.State.Active ||
		e.Clone.State.Active ||
		(e.ComponentCreator != nil && e.ComponentCreator.IsActive())
}

func (e *BuilderEditorComponents) DeactivateAll() {
	e.EditingComponent = false
	e.EditingName = false
	e.Enhanced.Active = false
	e.TagEditor.Active = false
	e.Rename.State.Active = false
	e.Clone.State.Active = false
	if e.ComponentCreator != nil {
		e.ComponentCreator.Reset()
	}
}

func (e *BuilderEditorComponents) HasUnsavedChanges() bool {
	// Check component content changes
	if e.EditingComponent && e.ComponentContent != e.OriginalContent {
		return true
	}

	// Check enhanced editor changes
	if e.Enhanced.Active && e.Enhanced.HasUnsavedChanges() {
		return true
	}

	// Check tag editor changes
	if e.TagEditor.Active && e.TagEditor.HasChanges() {
		return true
	}

	return false
}

// Helper methods for BuilderSearchComponents
func (s *BuilderSearchComponents) IsSearchActive() bool {
	return s.Bar != nil && s.Bar.Value() != ""
}

func (s *BuilderSearchComponents) ClearSearch() {
	s.Query = ""
	if s.Bar != nil {
		s.Bar.SetValue("")
	}
}

// Helper methods for BuilderUIComponents
func (u *BuilderUIComponents) HasTagChanges() bool {
	if !u.EditingTags {
		return false
	}

	if len(u.CurrentTags) != len(u.OriginalTags) {
		return true
	}

	// Create maps for efficient comparison
	currentMap := make(map[string]bool)
	for _, tag := range u.CurrentTags {
		currentMap[tag] = true
	}

	for _, tag := range u.OriginalTags {
		if !currentMap[tag] {
			return true
		}
	}

	return false
}

func (u *BuilderUIComponents) IsInEditMode() bool {
	return u.ExitConfirm != nil ||
		u.DeleteConfirm != nil ||
		u.ArchiveConfirm != nil ||
		u.EditingTags ||
		u.TagDeleteConfirm != nil ||
		(u.MermaidState != nil && u.MermaidState.Active)
}