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

	// Use the ViewTitle component for consistent styling
	viewTitle := NewViewTitle(title)
	viewTitle.SetWidth(0) // We'll handle width calculation separately for header

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
		
		// Get the rendered title from ViewTitle component
		titleRendered := viewTitle.View()
		
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
		
		// Use content width and place title on left, logo on right
		headerContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			titleRendered,
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