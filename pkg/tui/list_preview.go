package tui

import (
	"fmt"

	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/utils"
)

// PreviewRenderer handles rendering preview content
type PreviewRenderer struct {
	ShowPreview bool
}

// NewPreviewRenderer creates a new preview renderer
func NewPreviewRenderer() *PreviewRenderer {
	return &PreviewRenderer{}
}

// RenderPipelinePreview generates preview content for a pipeline
func (r *PreviewRenderer) RenderPipelinePreview(pipelinePath string, isArchived bool) string {
	if !r.ShowPreview {
		return ""
	}

	// Load pipeline - use appropriate function based on archive status
	var pipeline *models.Pipeline
	var err error

	if isArchived {
		// For archived pipelines, path is just the filename
		pipeline, err = files.ReadArchivedPipeline(pipelinePath)
	} else {
		// For active pipelines, use regular ReadPipeline
		pipeline, err = files.ReadPipeline(pipelinePath)
	}

	if err != nil {
		return fmt.Sprintf("Error loading pipeline: %v", err)
	}

	// Generate preview
	output, err := composer.ComposePipeline(pipeline)
	if err != nil {
		// For archived pipelines, provide a more helpful message
		if isArchived {
			return fmt.Sprintf("Preview unavailable for archived pipeline.\n\nNote: Archived pipelines may reference components that have also been archived or deleted.\n\nTo fully restore this pipeline, you may need to unarchive its components first.")
		}
		return fmt.Sprintf("Error generating preview: %v", err)
	}

	return output
}

// RenderComponentPreview generates preview content for a component
func (r *PreviewRenderer) RenderComponentPreview(comp componentItem) string {
	if !r.ShowPreview {
		return ""
	}

	// Read component content - use appropriate function based on archive status
	var content *models.Component
	var err error

	if comp.isArchived {
		// For archived components, use ReadArchivedComponent
		content, err = files.ReadArchivedComponent(comp.path)
	} else {
		// For active components, use ReadComponent
		content, err = files.ReadComponent(comp.path)
	}

	if err != nil {
		return fmt.Sprintf("Error loading component: %v", err)
	}

	// Return just the component content without metadata
	return content.Content
}

// RenderEmptyPreview returns appropriate empty preview message
func (r *PreviewRenderer) RenderEmptyPreview(activePane pane, hasPipelines bool, hasComponents bool) string {
	if !r.ShowPreview {
		return ""
	}

	if activePane == pipelinesPane {
		if !hasPipelines {
			return "No pipelines to preview."
		}
	} else if activePane == componentsPane {
		if !hasComponents {
			return "No components to preview."
		}
	}

	return ""
}

// EstimatePreviewTokens estimates tokens for a preview content
func EstimatePreviewTokens(content string) int {
	return utils.EstimateTokens(content)
}
