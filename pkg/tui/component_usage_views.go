package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ComponentUsageRenderer handles the rendering of the component usage modal
type ComponentUsageRenderer struct {
	Width  int
	Height int
}

// NewComponentUsageRenderer creates a new component usage renderer
func NewComponentUsageRenderer(width, height int) *ComponentUsageRenderer {
	return &ComponentUsageRenderer{
		Width:  width,
		Height: height,
	}
}

// Render returns the complete component usage view as a modal overlay
func (cur *ComponentUsageRenderer) Render(state *ComponentUsageState) string {
	if !state.Active {
		return ""
	}

	// Calculate modal dimensions (80% of screen, max 100 chars wide, 30 lines tall)
	modalWidth := min(100, int(float64(cur.Width)*0.8))
	modalHeight := min(30, int(float64(cur.Height)*0.8))
	
	// Build modal content
	var content strings.Builder
	
	// Title
	title := fmt.Sprintf("Pipelines using: %s", state.SelectedComponent.name)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorActive)).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	
	// Component info
	componentInfo := fmt.Sprintf("Type: %s | Path: %s", 
		state.SelectedComponent.compType, 
		state.SelectedComponent.path)
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorDim)).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render(componentInfo))
	content.WriteString("\n\n")
	
	// Calculate visible lines (needed for both rendering and help text)
	visibleLines := state.getVisibleLines()
	
	// Pipeline list or empty message
	if len(state.PipelinesUsingComponent) == 0 {
		emptyMsg := "No pipelines are currently using this component"
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Italic(true).
			Width(modalWidth - 4).
			Align(lipgloss.Center)
		content.WriteString(emptyStyle.Render(emptyMsg))
	} else {
		// Header
		headerText := fmt.Sprintf("Found in %d pipeline(s):", len(state.PipelinesUsingComponent))
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorNormal))
		content.WriteString(headerStyle.Render(headerText))
		content.WriteString("\n\n")
		
		// Calculate visible range
		startIdx := state.ScrollOffset
		endIdx := min(len(state.PipelinesUsingComponent), startIdx+visibleLines)
		
		// Render visible pipelines
		for i := startIdx; i < endIdx; i++ {
			usage := state.PipelinesUsingComponent[i]
			isSelected := i == state.SelectedIndex
			
			// Pipeline name with selection indicator
			var line strings.Builder
			if isSelected {
				line.WriteString("→ ")
			} else {
				line.WriteString("  ")
			}
			
			// Pipeline name
			nameStyle := lipgloss.NewStyle().Bold(true)
			if isSelected {
				nameStyle = nameStyle.Foreground(lipgloss.Color(ColorActive))
			} else {
				nameStyle = nameStyle.Foreground(lipgloss.Color(ColorNormal))
			}
			line.WriteString(nameStyle.Render(usage.Name))
			
			// Position info
			positionText := fmt.Sprintf(" (position %d of %d)", 
				usage.ComponentOrder, usage.TotalComponents)
			positionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim))
			line.WriteString(positionStyle.Render(positionText))
			
			content.WriteString(line.String())
			content.WriteString("\n")
			
			// Path on next line (indented)
			pathText := fmt.Sprintf("    %s", usage.Path)
			pathStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorVeryDim)).
				Italic(true)
			content.WriteString(pathStyle.Render(pathText))
			content.WriteString("\n")
		}
		
		// Scroll indicator if needed
		if len(state.PipelinesUsingComponent) > visibleLines {
			content.WriteString("\n")
			scrollInfo := fmt.Sprintf("Showing %d-%d of %d", 
				startIdx+1, endIdx, len(state.PipelinesUsingComponent))
			scrollStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDim)).
				Width(modalWidth - 4).
				Align(lipgloss.Center)
			content.WriteString(scrollStyle.Render(scrollInfo))
		}
	}
	
	// Help text at bottom
	content.WriteString("\n\n")
	var helpText string
	if len(state.PipelinesUsingComponent) > visibleLines {
		// Only show navigation help if there are more items than can fit
		helpText = "↑/↓ Navigate • ESC/q/u Close"
	} else {
		helpText = "ESC/q/u Close"
	}
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorDim)).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(helpStyle.Render(helpText))
	
	// Create modal box with border
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorActive)).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight)
	
	modal := modalStyle.Render(content.String())
	
	// Center the modal on screen
	return centerInScreen(modal, cur.Width, cur.Height)
}

// centerInScreen centers content in the available screen space
func centerInScreen(content string, screenWidth, screenHeight int) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	contentWidth := 0
	for _, line := range lines {
		if lineWidth := lipgloss.Width(line); lineWidth > contentWidth {
			contentWidth = lineWidth
		}
	}
	
	// Calculate padding
	topPadding := max(0, (screenHeight-contentHeight)/2)
	leftPadding := max(0, (screenWidth-contentWidth)/2)
	
	// Build centered view
	var result strings.Builder
	
	// Add top padding
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}
	
	// Add left padding to each line
	leftPad := strings.Repeat(" ", leftPadding)
	for _, line := range lines {
		result.WriteString(leftPad)
		result.WriteString(line)
		result.WriteString("\n")
	}
	
	return result.String()
}