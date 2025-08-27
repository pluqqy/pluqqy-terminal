package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/search/unified"
)

// ListDataStore manages all component and pipeline data
type ListDataStore struct {
	// Core data
	Pipelines []pipelineItem
	Prompts   []componentItem
	Contexts  []componentItem
	Rules     []componentItem

	// Filtered data (after search)
	FilteredPipelines  []pipelineItem
	FilteredComponents []componentItem
}

// ListViewportManager manages all UI viewports and dimensions
type ListViewportManager struct {
	Preview    viewport.Model
	Pipelines  viewport.Model
	Components viewport.Model
	Width      int
	Height     int
}

// ListEditorComponents groups all editor-related functionality
type ListEditorComponents struct {
	Enhanced      *EnhancedEditorState
	FileReference *FileReferenceState
	TagEditor     *TagEditor
	Rename        *ListRenameComponents
	Clone         *ListCloneComponents
}

// ListRenameComponents groups rename-related components
type ListRenameComponents struct {
	State    *RenameState
	Renderer *RenameRenderer
	Operator *RenameOperator
}

// ListCloneComponents groups clone-related components
type ListCloneComponents struct {
	State    *CloneState
	Renderer *CloneRenderer
	Operator *CloneOperator
}

// ListSearchComponents groups search-related functionality
type ListSearchComponents struct {
	// Unified search manager
	UnifiedManager *unified.UnifiedSearchManager
	
	// UI components
	Bar          *SearchBar
	Query        string
	FilterHelper *SearchFilterHelper
}

// ListOperationComponents groups business operation handlers
type ListOperationComponents struct {
	BusinessLogic    *BusinessLogic
	PipelineOperator *PipelineOperator
	ComponentCreator *ListComponentCreator
	MermaidOperator  *MermaidOperator
	TagReloader      *TagReloader
}

// ListUIComponents groups UI-specific components
type ListUIComponents struct {
	ExitConfirm            *ConfirmationModel
	ExitConfirmationType   string
	PreviewContent         string
	ComponentTableRenderer *ComponentTableRenderer
	TagReloadRenderer      *TagReloadRenderer
	MermaidState           *MermaidState
}

// Helper methods for ListDataStore
func (d *ListDataStore) GetAllComponents() []componentItem {
	allComponents := make([]componentItem, 0, len(d.Prompts)+len(d.Contexts)+len(d.Rules))
	allComponents = append(allComponents, d.Prompts...)
	allComponents = append(allComponents, d.Contexts...)
	allComponents = append(allComponents, d.Rules...)
	return allComponents
}

func (d *ListDataStore) GetFilteredComponentCounts() (prompts, contexts, rules int) {
	for _, comp := range d.FilteredComponents {
		switch comp.compType {
		case "prompt":
			prompts++
		case "context":
			contexts++
		case "rules":
			rules++
		}
	}
	return
}

// Helper methods for ListViewportManager
func (v *ListViewportManager) UpdateSizes(width, height int, showPreview bool) {
	v.Width = width
	v.Height = height

	// Calculate dimensions
	columnWidth := (width - 6) / 2
	searchBarHeight := 3
	contentHeight := height - 13 - searchBarHeight

	if showPreview {
		contentHeight = contentHeight / 2
	}

	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Update pipelines and components viewports
	viewportHeight := contentHeight - 4
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	v.Pipelines.Width = columnWidth - 4
	v.Pipelines.Height = viewportHeight
	v.Components.Width = columnWidth - 4
	v.Components.Height = viewportHeight

	// Update preview viewport
	if showPreview {
		previewHeight := height/2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		v.Preview.Width = width - 8
		v.Preview.Height = previewHeight
	}
}

// Helper methods for ListEditorComponents
func (e *ListEditorComponents) IsAnyEditorActive() bool {
	return e.Enhanced.Active ||
		e.FileReference.Active ||
		e.TagEditor.Active ||
		e.Rename.State.Active ||
		e.Clone.State.Active
}

func (e *ListEditorComponents) DeactivateAll() {
	e.Enhanced.Active = false
	e.FileReference.Active = false
	e.TagEditor.Active = false
	e.Rename.State.Active = false
	e.Clone.State.Active = false
}

// Helper methods for ListSearchComponents
func (s *ListSearchComponents) IsSearchActive() bool {
	return s.Bar != nil && s.Bar.Value() != ""
}

func (s *ListSearchComponents) ClearSearch() {
	s.Query = ""
	if s.Bar != nil {
		// Clear the search bar by setting a new empty value
		s.Bar.SetValue("")
	}
}

// InitializeUnifiedManager initializes the unified search manager if not already done
func (s *ListSearchComponents) InitializeUnifiedManager() {
	if s.UnifiedManager == nil {
		s.UnifiedManager = unified.NewUnifiedSearchManager()
	}
}

// ShouldIncludeArchived checks if archived items should be included based on search query
func (s *ListSearchComponents) ShouldIncludeArchived() bool {
	if s.UnifiedManager != nil {
		return s.UnifiedManager.IsStructuredQuery(s.Query) && 
			unified.ShouldIncludeArchived(s.Query)
	}
	// Fallback to existing logic
	return unified.ShouldIncludeArchived(s.Query)
}