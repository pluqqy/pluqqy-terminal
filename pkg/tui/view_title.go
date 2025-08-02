package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// ViewTitle creates a standardized title component with consistent styling
// Used for all view titles (Component Edit, Pipeline Builder, etc.)
type ViewTitle struct {
	text  string
	width int
}

// NewViewTitle creates a new view title with the given text
func NewViewTitle(text string) *ViewTitle {
	return &ViewTitle{
		text: text,
	}
}

// SetWidth sets the width for the title component
func (v *ViewTitle) SetWidth(width int) {
	v.width = width
}

// View renders the title with consistent styling
func (v *ViewTitle) View() string {
	if v.text == "" {
		return ""
	}

	// Title style with white text on black background
	// Note: Lipgloss doesn't support rounded corners on backgrounds directly,
	// only on borders. To achieve a similar effect, we could use a border
	// with matching background, but it would add extra height/width.
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")). // White text
		Background(lipgloss.Color("0")).   // Black background for better contrast
		Bold(true).
		Padding(0, 1) // Add padding inside the background

	// Add vertical padding for consistent height
	titleWithPadding := "\n" + v.text + "\n"
	return titleStyle.Render(titleWithPadding)
}

// ViewWithAlignment renders the title with alignment and padding
func (v *ViewTitle) ViewWithAlignment(width int) string {
	if v.text == "" {
		return ""
	}

	titleRendered := v.View()

	// Left align with padding
	alignStyle := lipgloss.NewStyle().
		Width(width).
		PaddingLeft(2).
		PaddingRight(2)

	return alignStyle.Render(titleRendered)
}

// ViewTitleHeight returns the consistent height of view titles
func ViewTitleHeight() int {
	return 3 // 1 line for text + 2 lines for vertical padding
}