package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// PipelineBuilderConfig holds configuration options for the Pipeline Builder
type PipelineBuilderConfig struct {
	// UseEnhancedEditor determines whether to use the enhanced editor
	// for component editing (default: true)
	UseEnhancedEditor bool
	
	// ShowPreviewByDefault controls whether the preview pane is shown
	// automatically when the builder opens (default: true)
	ShowPreviewByDefault bool
}

// DefaultPipelineBuilderConfig returns the default configuration
func DefaultPipelineBuilderConfig() *PipelineBuilderConfig {
	return &PipelineBuilderConfig{
		UseEnhancedEditor:    true,  // Enhanced editor enabled by default
		ShowPreviewByDefault: true,  // Preview shown by default
	}
}

// NewPipelineBuilderModelWithConfig creates a new Pipeline Builder with custom configuration
func NewPipelineBuilderModelWithConfig(config *PipelineBuilderConfig) *PipelineBuilderModel {
	if config == nil {
		config = DefaultPipelineBuilderConfig()
	}
	
	m := &PipelineBuilderModel{
		activeColumn:       leftColumn,
		showPreview:        config.ShowPreviewByDefault,
		editingName:        true,
		nameInput:          "",
		pipeline: &models.Pipeline{
			Name:       "",
			Components: []models.ComponentRef{},
		},
		originalComponents: []models.ComponentRef{},
		previewViewport:    viewport.New(80, 20),
		leftTableRenderer:  NewComponentTableRenderer(40, 20, true),
		rightViewport:      viewport.New(40, 20),
		searchBar:          NewSearchBar(),
		exitConfirm:        NewConfirmation(),
		tagDeleteConfirm:   NewConfirmation(),
		deleteConfirm:      NewConfirmation(),
		archiveConfirm:     NewConfirmation(),
		enhancedEditor:     NewEnhancedEditorState(),
		useEnhancedEditor:  config.UseEnhancedEditor,
		renameState:        NewRenameState(),
		renameRenderer:     NewRenameRenderer(),
		renameOperator:     NewRenameOperator(),
		cloneState:         NewCloneState(),
		cloneRenderer:      NewCloneRenderer(),
		cloneOperator:      NewCloneOperator(),
	}
	
	// Initialize mermaid state and operator
	mermaidState := NewMermaidState()
	m.mermaidState = mermaidState
	m.mermaidOperator = NewMermaidOperator(mermaidState)
	
	// Initialize search engine
	m.searchEngine = search.NewEngine()
	
	// Configure table renderer for pipeline builder
	m.leftTableRenderer.ShowAddedIndicator = true
	
	m.loadAvailableComponents()
	return m
}

// LoadPipelineWithConfig loads an existing pipeline with custom configuration
func LoadPipelineWithConfig(pipeline *models.Pipeline, config *PipelineBuilderConfig) *PipelineBuilderModel {
	if config == nil {
		config = DefaultPipelineBuilderConfig()
	}
	
	m := NewPipelineBuilderModelWithConfig(config)
	m.pipeline = pipeline
	m.editingName = false // Don't edit name for existing pipeline
	m.nameInput = pipeline.Name
	
	// Copy components to track original state
	m.originalComponents = make([]models.ComponentRef, len(pipeline.Components))
	copy(m.originalComponents, pipeline.Components)
	
	// Copy components for editing
	m.selectedComponents = make([]models.ComponentRef, len(pipeline.Components))
	copy(m.selectedComponents, pipeline.Components)
	
	// Set the right cursor position if there are components
	if len(m.selectedComponents) > 0 {
		m.rightCursor = 0
	}
	
	// Update preview
	m.updatePreview()
	
	return m
}