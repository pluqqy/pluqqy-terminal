package tui

import (
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

func renderHeader(width int, title string) string {
	// ASCII art logo from pluqqy.txt with integrated version
	logo := `▄▖▖ ▖▖▄▖▄▖▖▖
▙▌▌ ▌▌▌▌▌▌▌▌
▌ ▙▖▙▌█▌█▌▐
v0.1.0 ▘ ▘▘`

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Pink/magenta color
		Bold(true)

	// Header padding style (matching pane borders which add 1 char on each side)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2).
		Width(width)

	// Render the complete logo with version
	logoRendered := logoStyle.Render(logo)

	var headerContent string
	
	// If there's a title, render it on the left
	if title != "" {
		// Split the logo into lines to align title with version row
		logoLines := strings.Split(logoRendered, "\n")
		
		// Create title style directly for header (without vertical padding)
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // White text
			Background(lipgloss.Color("0")).   // Black background
			Bold(true).
			Padding(0, 1) // Horizontal padding only
		
		titleRendered := titleStyle.Render(title)
		
		// Calculate available width for content (accounting for padding)
		contentWidth := width - 4 // -4 for left and right padding (2 each side)
		
		// Get the actual rendered width of the title (including background padding)
		titleRenderedWidth := lipgloss.Width(titleRendered)
		logoWidth := lipgloss.Width(logoLines[0])
		
		// Calculate spacer width
		spacerWidth := contentWidth - titleRenderedWidth - logoWidth
		if spacerWidth < 0 {
			spacerWidth = 0
		}
		
		// Build header with title aligned to the version line (4th line)
		// Add 3 empty lines above the title to align with line 4
		emptyLine := lipgloss.NewStyle().Width(titleRenderedWidth).Render("")
		titleColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			emptyLine, // Line 1
			emptyLine, // Line 2
			emptyLine, // Line 3
			titleRendered, // Line 4 - aligned with version
		)
		
		// Use content width and place title on left, logo on right
		headerContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			titleColumn,
			lipgloss.NewStyle().Width(spacerWidth).Render(""),
			logoRendered,
		)
	} else {
		// No title, just right-align the logo
		rightAlign := lipgloss.NewStyle().
			Width(width - 4). // -4 for padding (2 each side)
			Align(lipgloss.Right)
		
		headerContent = rightAlign.Render(logoRendered)
	}
	
	return headerPadding.Render(headerContent)
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}