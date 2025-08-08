package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
)

// PipelineViewRenderer handles rendering of the pipeline pane
type PipelineViewRenderer struct {
	Width              int
	Height             int
	ActivePane         pane
	Pipelines          []pipelineItem
	FilteredPipelines  []pipelineItem
	PipelineCursor     int
	SearchQuery        string
	Viewport           viewport.Model
}

// NewPipelineViewRenderer creates a new pipeline view renderer
func NewPipelineViewRenderer(width, height int) *PipelineViewRenderer {
	return &PipelineViewRenderer{
		Width:  width,
		Height: height,
	}
}

// Render generates the complete pipeline pane view
func (r *PipelineViewRenderer) Render() string {
	var content strings.Builder
	
	// Calculate column width
	columnWidth := (r.Width - 6) / 2 // Account for gap, padding, and ensure border visibility
	
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Create heading with colons spanning the width
	heading := "PIPELINES"
	remainingWidth := columnWidth - len(heading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	// Dynamic header and colon styles based on active pane
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == pipelinesPane {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == pipelinesPane {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	
	content.WriteString(headerPadding.Render(headerStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")
	
	// Table header for pipelines with token count
	pipelineHeaderStyle := HeaderStyle
	
	// Table column widths for pipelines (adjusted for viewport width)
	pipelineViewportWidth := columnWidth - 4 // Same as viewport.Width
	pipelineNameWidth, pipelineTagsWidth, pipelineTokenWidth, _ := formatColumnWidths(pipelineViewportWidth, false)
	
	// Render table header with consistent spacing to match rows
	pipelineHeader := fmt.Sprintf("  %-*s %-*s %*s", 
		pipelineNameWidth, "Name",
		pipelineTagsWidth, "Tags",
		pipelineTokenWidth, "~Tokens")
	content.WriteString(headerPadding.Render(pipelineHeaderStyle.Render(pipelineHeader)))
	content.WriteString("\n\n")
	
	// Build scrollable content for pipelines viewport
	scrollContent := r.buildScrollableContent(pipelineNameWidth, pipelineTagsWidth, pipelineTokenWidth)
	
	// Update viewport with content
	r.Viewport.SetContent(scrollContent)
	
	// Update viewport to follow cursor
	r.updateViewportScroll()
	
	// Add padding to viewport content
	viewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	content.WriteString(viewportPadding.Render(r.Viewport.View()))
	
	// Apply border
	borderStyle := InactiveBorderStyle
	if r.ActivePane == pipelinesPane {
		borderStyle = ActiveBorderStyle
	}
	
	return borderStyle.
		Width(columnWidth).
		Height(r.Height).
		Render(content.String())
}

// buildScrollableContent creates the content for the pipelines viewport
func (r *PipelineViewRenderer) buildScrollableContent(nameWidth, tagsWidth, tokenWidth int) string {
	var content strings.Builder
	normalStyle := NormalStyle
	dimmedStyle := EmptyInactiveStyle
	
	if len(r.FilteredPipelines) == 0 {
		if r.ActivePane == pipelinesPane {
			// Active pane - show prominent message
			emptyStyle := EmptyActiveStyle
			
			// Check if we have pipelines but they're filtered out
			if len(r.Pipelines) > 0 && r.SearchQuery != "" {
				content.WriteString(emptyStyle.Render("No pipelines match your search."))
			} else {
				content.WriteString(emptyStyle.Render("No pipelines found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			
			// Check if we have pipelines but they're filtered out
			if len(r.Pipelines) > 0 && r.SearchQuery != "" {
				content.WriteString(dimmedStyle.Render("No pipelines match your search."))
			} else {
				content.WriteString(dimmedStyle.Render("No pipelines found."))
			}
		}
	} else {
		for i, pipeline := range r.FilteredPipelines {
			// Format the pipeline name with archived indicator
			nameStr := pipeline.name
			if pipeline.isArchived {
				nameStr = "📦 " + nameStr + " [ARCHIVED]"
			}
			nameStr = truncateName(nameStr, nameWidth)
			
			// Format tags
			tagsStr := renderTagChipsWithWidth(pipeline.tags, tagsWidth, 2) // Show max 2 tags inline
			
			// Format token count - right-aligned
			tokenStr := fmt.Sprintf("%d", pipeline.tokenCount)
			
			// Build the row components separately for proper styling
			namePart := fmt.Sprintf("%-*s", nameWidth, nameStr)
			
			// For tags, we need to pad based on rendered width
			tagsPadding := tagsWidth - lipgloss.Width(tagsStr)
			if tagsPadding < 0 {
				tagsPadding = 0
			}
			tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)
			
			tokenPart := fmt.Sprintf("%*s", tokenWidth, tokenStr)
			
			// Build row with styling
			var row string
			if r.ActivePane == pipelinesPane && i == r.PipelineCursor {
				// Apply selection styling only to name column
				if pipeline.isArchived {
					// Dimmed style for archived items
					row = "▸ " + dimmedStyle.Render(namePart) + " " + tagsPart + " " + dimmedStyle.Render(tokenPart)
				} else {
					row = "▸ " + SelectedStyle.Render(namePart) + " " + tagsPart + " " + normalStyle.Render(tokenPart)
				}
			} else {
				// Normal row styling
				if pipeline.isArchived {
					// Dimmed style for archived items
					row = "  " + dimmedStyle.Render(namePart) + " " + tagsPart + " " + dimmedStyle.Render(tokenPart)
				} else {
					row = "  " + normalStyle.Render(namePart) + " " + tagsPart + " " + normalStyle.Render(tokenPart)
				}
			}
			
			content.WriteString(row)
			
			if i < len(r.FilteredPipelines)-1 {
				content.WriteString("\n")
			}
		}
	}
	
	return content.String()
}

// updateViewportScroll updates the viewport to follow the cursor
func (r *PipelineViewRenderer) updateViewportScroll() {
	if r.ActivePane == pipelinesPane && len(r.Pipelines) > 0 {
		// For pipelines, each item is one line
		currentLine := r.PipelineCursor
		
		// Ensure the cursor line is visible
		if currentLine < r.Viewport.YOffset {
			r.Viewport.SetYOffset(currentLine)
		} else if currentLine >= r.Viewport.YOffset+r.Viewport.Height {
			r.Viewport.SetYOffset(currentLine - r.Viewport.Height + 1)
		}
	}
}