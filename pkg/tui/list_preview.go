package tui

import (
	"fmt"
	"strings"
	
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
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
func (r *PreviewRenderer) RenderPipelinePreview(pipelinePath string) string {
	if !r.ShowPreview {
		return ""
	}
	
	// Load pipeline
	pipeline, err := files.ReadPipeline(pipelinePath)
	if err != nil {
		return fmt.Sprintf("Error loading pipeline: %v", err)
	}
	
	// Generate preview
	output, err := composer.ComposePipeline(pipeline)
	if err != nil {
		return fmt.Sprintf("Error generating preview: %v", err)
	}
	
	return output
}

// RenderComponentPreview generates preview content for a component
func (r *PreviewRenderer) RenderComponentPreview(comp componentItem) string {
	if !r.ShowPreview {
		return ""
	}
	
	// Read component content
	content, err := files.ReadComponent(comp.path)
	if err != nil {
		return fmt.Sprintf("Error loading component: %v", err)
	}
	
	// Format component preview with metadata
	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("# %s\n\n", comp.name))
	preview.WriteString(fmt.Sprintf("**Type:** %s\n", strings.Title(comp.compType)))
	preview.WriteString(fmt.Sprintf("**Path:** %s\n", comp.path))
	preview.WriteString(fmt.Sprintf("**Usage Count:** %d\n", comp.usageCount))
	preview.WriteString(fmt.Sprintf("**Token Count:** ~%d\n", comp.tokenCount))
	if !comp.lastModified.IsZero() {
		preview.WriteString(fmt.Sprintf("**Last Modified:** %s\n", comp.lastModified.Format("2006-01-02 15:04:05")))
	}
	preview.WriteString("\n---\n\n")
	preview.WriteString(content.Content)
	
	return preview.String()
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