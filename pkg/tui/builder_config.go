package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
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
		ComponentCreator: NewComponentCreator(),
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
		Engine:       search.NewEngine(),
		Bar:          NewSearchBar(),
		FilterHelper: NewSearchFilterHelper(),
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
		TagDeleteConfirm:     NewConfirmation(), // Initialize for compatibility
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
