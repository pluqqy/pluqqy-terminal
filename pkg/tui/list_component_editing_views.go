package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// ComponentEditingViewRenderer handles rendering of the component editing view
type ComponentEditingViewRenderer struct {
	width  int
	height int
}

// NewComponentEditingViewRenderer creates a new renderer
func NewComponentEditingViewRenderer(width, height int) *ComponentEditingViewRenderer {
	return &ComponentEditingViewRenderer{
		width:  width,
		height: height,
	}
}

// RenderEditView renders the component editing interface
func (r *ComponentEditingViewRenderer) RenderEditView(componentName, content string, editViewport viewport.Model) string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	// Calculate dimensions  
	contentWidth := r.width - 4 // Match help pane width
	contentHeight := r.height - 7 // Reserve space for help pane (3) + spacing (3) + status bar (1)

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true) // Purple for active single pane

	heading := fmt.Sprintf("EDITING: %s", componentName)
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Update viewport dimensions if needed
	viewportWidth := contentWidth - 4 // 2 for border, 2 for headerPadding
	viewportHeight := contentHeight - 5 // Account for header and spacing
	if editViewport.Width != viewportWidth || editViewport.Height != viewportHeight {
		editViewport.Width = viewportWidth
		editViewport.Height = viewportHeight
	}
	
	// Editor content with cursor
	contentWithCursor := content + "│" // cursor
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := preprocessContent(contentWithCursor)
	
	// Wrap content to viewport width to prevent overflow
	wrappedContent := wordwrap.String(processedContent, viewportWidth)
	
	// Update viewport content
	editViewport.SetContent(wrappedContent)
	
	// Use viewport for scrollable content
	mainContent.WriteString(headerPadding.Render(editViewport.View()))

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"↑/↓ scroll",
		"^s save",
		"^x edit external",
		"esc cancel",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

	// Combine all elements
	var s strings.Builder

	// Add top margin to ensure content is not cut off
	s.WriteString("\n")

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(mainPane))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}