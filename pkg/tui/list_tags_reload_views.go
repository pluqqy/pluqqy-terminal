package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TagReloadRenderer handles the rendering of tag reload status and results
type TagReloadRenderer struct {
	Width  int
	Height int
}

// NewTagReloadRenderer creates a new tag reload renderer
func NewTagReloadRenderer(width, height int) *TagReloadRenderer {
	return &TagReloadRenderer{
		Width:  width,
		Height: height,
	}
}

// RenderStatus renders the tag reload status overlay
func (r *TagReloadRenderer) RenderStatus(reloader *TagReloader) string {
	if !reloader.Active {
		return ""
	}

	var content strings.Builder

	// Create box style for the overlay
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	if reloader.IsReloading {
		// Show loading state
		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorPrimary)).
			Render("⟳ Reloading Tags"))
		content.WriteString("\n\n")
		content.WriteString("Scanning components and pipelines...")

	} else if reloader.LastError != nil {
		// Show error state
		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorError)).
			Render("✗ Tag Reload Failed"))
		content.WriteString("\n\n")
		content.WriteString(ErrorStyle.Render(reloader.LastError.Error()))

	} else if reloader.ReloadResult != nil {
		// Show success state
		result := reloader.ReloadResult

		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorSuccess)).
			Render("✓ Tag Reload Complete"))
		content.WriteString("\n\n")

		// Statistics
		stats := []string{
			fmt.Sprintf("Components scanned: %d", result.ComponentsScanned),
			fmt.Sprintf("Pipelines scanned: %d", result.PipelinesScanned),
			fmt.Sprintf("Total tags found: %d", result.TotalTags),
		}

		if len(result.NewTags) > 0 {
			stats = append(stats, fmt.Sprintf("New tags added: %d", len(result.NewTags)))
		}

		if len(result.FailedFiles) > 0 {
			stats = append(stats, fmt.Sprintf("Failed files: %d", len(result.FailedFiles)))
		}

		for _, stat := range stats {
			content.WriteString(DescriptionStyle.Render(stat))
			content.WriteString("\n")
		}

		// Show new tags if any
		if len(result.NewTags) > 0 && len(result.NewTags) <= 5 {
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().
				Bold(true).
				Render("New tags:"))
			content.WriteString("\n")
			for _, tag := range result.NewTags {
				content.WriteString(fmt.Sprintf("  • %s\n", tag))
			}
		}
	}

	// Apply box style
	box := boxStyle.Render(content.String())

	// Center the box on screen
	return r.centerOverlay(box)
}

// centerOverlay centers content as an overlay on the screen
func (r *TagReloadRenderer) centerOverlay(content string) string {
	lines := strings.Split(content, "\n")
	boxHeight := len(lines)
	boxWidth := 0
	for _, line := range lines {
		if width := lipgloss.Width(line); width > boxWidth {
			boxWidth = width
		}
	}

	// Calculate padding to center the box
	topPadding := (r.Height - boxHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	leftPadding := (r.Width - boxWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Build centered output
	var output strings.Builder

	// Add top padding
	for i := 0; i < topPadding; i++ {
		output.WriteString("\n")
	}

	// Add content with left padding
	for _, line := range lines {
		output.WriteString(strings.Repeat(" ", leftPadding))
		output.WriteString(line)
		output.WriteString("\n")
	}

	return output.String()
}

// RenderInlineStatus renders a compact status message for the bottom of the screen
func (r *TagReloadRenderer) RenderInlineStatus(reloader *TagReloader) string {
	status := reloader.GetStatus()
	if status == "" {
		return ""
	}

	// Style based on state
	var style lipgloss.Style
	if reloader.IsReloading {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true)
	} else if reloader.LastError != nil {
		style = ErrorStyle
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess))
	}

	return style.Render(status)
}

// SetSize updates the renderer dimensions
func (r *TagReloadRenderer) SetSize(width, height int) {
	r.Width = width
	r.Height = height
}
