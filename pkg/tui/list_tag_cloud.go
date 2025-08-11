package tui

import (
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	tagspkg "github.com/pluqqy/pluqqy-cli/pkg/tags"
)

// TagCloudRenderer handles the rendering of the tag cloud pane
type TagCloudRenderer struct {
	Width         int
	Height        int
	IsActive      bool
	CursorIndex   int
	AvailableTags []string
	CurrentTags   []string
}

// NewTagCloudRenderer creates a new tag cloud renderer
func NewTagCloudRenderer(width, height int) *TagCloudRenderer {
	return &TagCloudRenderer{
		Width:  width,
		Height: height,
	}
}

// Render generates the tag cloud pane content
func (r *TagCloudRenderer) Render() string {
	var content strings.Builder
	
	// Header
	title := "AVAILABLE TAGS"
	remainingWidth := r.Width - len(title) - 7 // Adjust for smaller width
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	// Dynamic styles based on active state
	headerStyle := GetActiveHeaderStyle(r.IsActive)
	if !r.IsActive {
		// Use orange for inactive tag cloud header
		headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorWarning))
	}
	colonStyle := GetActiveColonStyle(r.IsActive)
	
	content.WriteString(HeaderPaddingStyle.Render(
		headerStyle.Render(title) + " " + 
		colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")
	
	// Get available tags (excluding current tags)
	availableForCloud := r.getAvailableForCloud()
	
	// Display tags
	if len(availableForCloud) == 0 {
		content.WriteString(HeaderPaddingStyle.Render(
			DescriptionStyle.Render("(no available tags)")))
	} else {
		// Render tag rows
		tagRows := r.renderTagRows(availableForCloud)
		content.WriteString(HeaderPaddingStyle.Render(tagRows))
	}
	
	// Apply border
	borderStyle := InactiveBorderStyle
	if r.IsActive {
		borderStyle = ActiveBorderStyle
	}
	
	return borderStyle.
		Width(r.Width).
		Height(r.Height).
		Render(content.String())
}

// getAvailableForCloud returns tags that aren't already selected
func (r *TagCloudRenderer) getAvailableForCloud() []string {
	var available []string
	for _, tag := range r.AvailableTags {
		if !r.hasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

// hasTag checks if a tag is already in current tags
func (r *TagCloudRenderer) hasTag(tag string) bool {
	for _, t := range r.CurrentTags {
		if t == tag {
			return true
		}
	}
	return false
}

// renderTagRows renders tags in rows that fit within the width
func (r *TagCloudRenderer) renderTagRows(tags []string) string {
	var rows strings.Builder
	rowTags := 0
	currentRowWidth := 0
	maxRowWidth := r.Width - 6 // Account for padding
	
	for i, tag := range tags {
		// Get tag color
		registry, _ := tagspkg.NewRegistry()
		color := models.GetTagColor(tag, "")
		if registry != nil {
			if t, exists := registry.GetTag(tag); exists && t.Color != "" {
				color = t.Color
			}
		}
		
		tagStyle := GetTagChipStyle(color)
		
		// Calculate tag display
		var tagDisplay string
		if r.IsActive && i == r.CursorIndex {
			indicatorStyle := CursorStyle
			tagDisplay = indicatorStyle.Render("▶ ") + tagStyle.Render(tag) + indicatorStyle.Render(" ◀")
		} else {
			tagDisplay = "  " + tagStyle.Render(tag) + "  "
		}
		
		tagWidth := lipgloss.Width(tagDisplay) + 2 // Add spacing
		
		// Check if we need a new row
		if rowTags > 0 && currentRowWidth + tagWidth > maxRowWidth {
			rows.WriteString("\n\n") // Double newline for vertical spacing
			rowTags = 0
			currentRowWidth = 0
		}
		
		rows.WriteString(tagDisplay)
		rows.WriteString("  ")
		currentRowWidth += tagWidth + 2
		rowTags++
	}
	
	return rows.String()
}

// GetHelpText returns help text for the tag cloud
func (r *TagCloudRenderer) GetHelpText() []string {
	return []string{
		"tab switch pane",
		"enter add tag",
		"←/→ navigate",
		"^d delete tag",
		"^s save",
		"esc cancel",
	}
}