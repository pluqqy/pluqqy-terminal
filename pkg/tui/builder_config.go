package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search/unified"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

// PipelineBuilderConfig holds configuration options for the Pipeline Builder
type PipelineBuilderConfig struct {
	// ShowPreviewByDefault controls whether the preview pane is shown
	// automatically when the builder opens (default: true)
	ShowPreviewByDefault bool
}

// DefaultPipelineBuilderConfig returns the default configuration
func DefaultPipelineBuilderConfig() *PipelineBuilderConfig {
	return &PipelineBuilderConfig{
		ShowPreviewByDefault: true, // Preview shown by default
	}
}

// NewPipelineBuilderModelWithConfig creates a new Pipeline Builder with custom configuration
func NewPipelineBuilderModelWithConfig(config *PipelineBuilderConfig) *PipelineBuilderModel {
	if config == nil {
		config = DefaultPipelineBuilderConfig()
	}

	// Initialize data store
	dataStore := &BuilderDataStore{
		Pipeline: &models.Pipeline{
			Name:       "",
			Components: []models.ComponentRef{},
		},
		SelectedComponents: []models.ComponentRef{},
		OriginalComponents: []models.ComponentRef{},
	}

	// Initialize viewport manager
	viewportManager := &BuilderViewportManager{
		Preview:       viewport.New(80, 20),
		LeftTable:     NewComponentTableRenderer(40, 20, true),
		RightViewport: viewport.New(40, 20),
		Width:         80,
		Height:        24,
	}

	// Configure table renderer for pipeline builder
	viewportManager.LeftTable.ShowAddedIndicator = true

	// Initialize editor components
	editorComponents := &BuilderEditorComponents{
		Enhanced:         NewEnhancedEditorState(),
		TagEditor:        NewTagEditor(),
		ComponentUsage:   NewComponentUsageState(),
		ComponentCreator: nil, // Will be set after model is created
		EditingName:      true,
		NameInput:        "",
		Rename: &BuilderRenameComponents{
			State:    NewRenameState(),
			Renderer: NewRenameRenderer(),
			Operator: NewRenameOperator(),
		},
		Clone: &BuilderCloneComponents{
			State:    NewCloneState(),
			Renderer: NewCloneRenderer(),
			Operator: NewCloneOperator(),
		},
	}

	// Initialize search components
	searchComponents := &BuilderSearchComponents{
		UnifiedManager: unified.NewUnifiedSearchManager(),
		Bar:            NewSearchBar(),
		FilterHelper:   NewSearchFilterHelper(),
	}

	// Initialize mermaid state and operator
	mermaidState := NewMermaidState()
	mermaidOperator := NewMermaidOperator(mermaidState)

	// Initialize UI components
	uiComponents := &BuilderUIComponents{
		ActiveColumn:         leftColumn,
		ShowPreview:          config.ShowPreviewByDefault,
		ExitConfirm:          NewConfirmation(),
		DeleteConfirm:        NewConfirmation(),
		ArchiveConfirm:       NewConfirmation(),
		MermaidState:         mermaidState,
		MermaidOperator:      mermaidOperator,
		SharedLayout:         NewSharedLayout(80, 24, config.ShowPreviewByDefault),
	}

	m := &PipelineBuilderModel{
		data:      dataStore,
		viewports: viewportManager,
		editors:   editorComponents,
		search:    searchComponents,
		ui:        uiComponents,
	}

	m.loadAvailableComponents()

	// Create and setup the component creator after the model is initialized
	m.editors.ComponentCreator = createBuilderComponentCreator(m)
	m.setupBuilderComponentCreatorEditor()

	return m
}

// LoadPipelineWithConfig loads an existing pipeline with custom configuration
func LoadPipelineWithConfig(pipeline *models.Pipeline, config *PipelineBuilderConfig) *PipelineBuilderModel {
	if config == nil {
		config = DefaultPipelineBuilderConfig()
	}

	m := NewPipelineBuilderModelWithConfig(config)
	m.data.Pipeline = pipeline
	m.editors.EditingName = false // Don't edit name for existing pipeline
	m.editors.NameInput = pipeline.Name

	// Copy components to track original state
	m.data.OriginalComponents = make([]models.ComponentRef, len(pipeline.Components))
	copy(m.data.OriginalComponents, pipeline.Components)

	// Copy components for editing
	m.data.SelectedComponents = make([]models.ComponentRef, len(pipeline.Components))
	copy(m.data.SelectedComponents, pipeline.Components)

	// Set the right cursor position if there are components
	if len(m.data.SelectedComponents) > 0 {
		m.ui.RightCursor = 0
	}

	// Update preview
	m.updatePreview()

	return m
}

// createBuilderComponentCreator creates a BuilderComponentCreator for the builder view
func createBuilderComponentCreator(model *PipelineBuilderModel) *BuilderComponentCreator {
	// Create the reload callback that's specific to the builder view
	reloadCallback := func() {
		if model != nil {
			model.loadAvailableComponents()
			// Also reload available tags for tag editor
			if model.editors.TagEditor != nil {
				model.editors.TagEditor.LoadAvailableTags()
			}
		}
	}

	// Create the builder-specific component creator
	// Note: Enhanced editor will be nil initially, will be set up later
	creator := NewBuilderComponentCreator(reloadCallback, nil)
	
	return creator
}

// setupBuilderComponentCreatorEditor sets up the enhanced editor for component creation
// This must be called after the PipelineBuilderModel is fully initialized
func (m *PipelineBuilderModel) setupBuilderComponentCreatorEditor() {
	if m.editors.ComponentCreator != nil && m.editors.Enhanced != nil {
		// Update the enhanced editor reference in the builder component creator
		m.editors.ComponentCreator.enhancedEditor = m.editors.Enhanced
		
		// Set up the enhanced editor adapter
		adapter := shared.NewEnhancedEditorAdapter(m.editors.Enhanced)
		m.editors.ComponentCreator.SetEnhancedEditor(adapter)
	}
}
