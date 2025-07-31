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

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Pink color to match
		Bold(true)

	// Header padding style (matching pane padding)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Width(width)

	// Render the complete logo with version
	logoRendered := logoStyle.Render(logo)

	var headerContent string
	
	// If there's a title, render it on the left
	if title != "" {
		// Split the logo into lines to align title with version row
		logoLines := strings.Split(logo, "\n")
		
		// Create padding for title to align with 4th line (version line)
		titlePadding := strings.Repeat("\n", 3) // 3 newlines to reach 4th line
		titleWithPadding := titlePadding + title
		titleRendered := titleStyle.Render(titleWithPadding)
		
		// Calculate available width for content (accounting for padding)
		contentWidth := width - 2 // -2 for left and right padding
		
		// Use content width and place title on left, logo on right
		headerContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			titleRendered,
			lipgloss.NewStyle().Width(contentWidth - lipgloss.Width(title) - lipgloss.Width(logoLines[0])).Render(""),
			logoRendered,
		)
	} else {
		// No title, just right-align the logo
		rightAlign := lipgloss.NewStyle().
			Width(width - 2). // -2 for padding
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