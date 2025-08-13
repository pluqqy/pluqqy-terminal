package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
	"github.com/muesli/reflow/wordwrap"
)

// MainViewRenderer handles the orchestration of all view components
type MainViewRenderer struct {
	Width           int
	Height          int
	ShowPreview     bool
	ActivePane      pane
	LastDataPane    pane
	SearchBar       *SearchBar
	PreviewViewport viewport.Model
	PreviewContent  string
}

// NewMainViewRenderer creates a new main view renderer
func NewMainViewRenderer(width, height int) *MainViewRenderer {
	return &MainViewRenderer{
		Width:  width,
		Height: height,
	}
}

// RenderErrorView renders an error state
func (r *MainViewRenderer) RenderErrorView(err error) string {
	return fmt.Sprintf("Error: Failed to load pipelines: %v\n\nPress 'q' to quit", err)
}

// RenderHelpPane renders the help text in a bordered pane
func (r *MainViewRenderer) RenderHelpPane(searchActive bool) string {
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.Width - 4).  // Account for left/right padding (2) and borders (2)
		Padding(0, 1)  // Internal padding for help text
	
	var helpContent string
	if searchActive {
		// Show search syntax help when search is active
		helpRows := [][]string{
			{"esc clear+exit search", "enter search"},
			{"tag:<name>", "type:<type>", "status:archived", "<keyword>", "combine with spaces"},
		}
		helpContent = formatHelpTextRows(helpRows, r.Width - 8) // -8 for borders and padding
	} else {
		// Show normal navigation help - grouped by function
		helpRows := [][]string{
			// Row 1: Navigation & viewing
			{"/ search", "tab switch pane", "↑/↓ nav", "enter view", "p preview"},
			// Row 2: CRUD operations & system
			{"n new", "e edit", "R rename", "^x external", "t tag", "M diagram", "^d delete", "a archive/unarchive", "S set", "s settings", "^c quit"},
		}
		helpContent = formatHelpTextRows(helpRows, r.Width - 8) // -8 for borders and padding
	}
	
	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	return contentStyle.Render(helpBorderStyle.Render(helpContent))
}

// RenderPreviewPane renders the preview pane with border and header
func (r *MainViewRenderer) RenderPreviewPane(pipelines []pipelineItem, components []componentItem, pipelineCursor, componentCursor int) string {
	if !r.ShowPreview || r.PreviewContent == "" {
		return ""
	}
	
	// Calculate token count
	tokenCount := utils.EstimateTokens(r.PreviewContent)
	
	// Create token badge with appropriate color
	tokenBadgeStyle := GetTokenBadgeStyle(tokenCount)
	tokenBadge := tokenBadgeStyle.Render(utils.FormatTokenCount(tokenCount))
	
	// Apply active/inactive style to preview border
	var previewBorderStyle lipgloss.Style
	if r.ActivePane == previewPane {
		previewBorderStyle = ActiveBorderStyle
	} else {
		previewBorderStyle = InactiveBorderStyle
	}
	previewBorderStyle = previewBorderStyle.Width(r.Width - 4) // Account for padding (2) and border (2)
	
	// Build preview content with header inside
	var previewContent strings.Builder
	
	// Create heading with colons and token info
	var previewHeading string
	
	// Determine what we're previewing based on lastDataPane
	if r.LastDataPane == pipelinesPane && len(pipelines) > 0 && pipelineCursor >= 0 && pipelineCursor < len(pipelines) {
		pipelineName := pipelines[pipelineCursor].name
		previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
	} else if r.LastDataPane == componentsPane && len(components) > 0 && componentCursor >= 0 && componentCursor < len(components) {
		comp := components[componentCursor]
		previewHeading = fmt.Sprintf("COMPONENT PREVIEW (%s)", comp.name)
	} else {
		previewHeading = "PREVIEW"
	}
	
	// Calculate the actual rendered width of token info
	tokenInfoWidth := lipgloss.Width(tokenBadge)
	
	// Calculate total available width inside the border
	totalWidth := r.Width - 8 // accounting for border padding and header padding
	
	// Calculate space for colons between heading and token info
	colonSpace := totalWidth - len(previewHeading) - tokenInfoWidth - 2 // -2 for spaces
	if colonSpace < 3 {
		colonSpace = 3
	}
	
	// Build the complete header line
	// Dynamic header and colon styles based on active pane
	previewHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == previewPane {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	previewColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == previewPane {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	previewHeaderPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	previewContent.WriteString(previewHeaderPadding.Render(previewHeaderStyle.Render(previewHeading) + " " + previewColonStyle.Render(strings.Repeat(":", colonSpace)) + " " + tokenBadge))
	previewContent.WriteString("\n\n")
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := preprocessContent(r.PreviewContent)
	// Wrap content to viewport width to prevent overflow
	wrappedContent := wordwrap.String(processedContent, r.PreviewViewport.Width)
	r.PreviewViewport.SetContent(wrappedContent)
	
	// Add padding to preview viewport content
	previewViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	previewContent.WriteString(previewViewportPadding.Render(r.PreviewViewport.View()))
	
	// Render the border around the entire preview with same padding as top columns
	var result strings.Builder
	result.WriteString("\n")
	previewPaddingStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	result.WriteString(previewPaddingStyle.Render(previewBorderStyle.Render(previewContent.String())))
	
	return result.String()
}

// RenderConfirmationDialogs renders delete/archive confirmation dialogs
func (r *MainViewRenderer) RenderConfirmationDialogs(pipelineOperator *PipelineOperator) string {
	var result strings.Builder
	
	// Show delete confirmation if active
	if pipelineOperator.IsDeleteConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		result.WriteString("\n")
		result.WriteString(confirmStyle.Render(pipelineOperator.ViewDeleteConfirm(r.Width - 4)))
	}
	
	// Show archive confirmation if active
	if pipelineOperator.IsArchiveConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange for archive
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		result.WriteString("\n")
		result.WriteString(confirmStyle.Render(pipelineOperator.ViewArchiveConfirm(r.Width - 4)))
	}
	
	return result.String()
}

// CalculateContentHeight calculates the height for the main content area
func (r *MainViewRenderer) CalculateContentHeight() int {
	searchBarHeight := 3 // Height for search bar
	contentHeight := r.Height - 13 - searchBarHeight // Reserve space for header, search bar, help pane, and spacing
	
	if r.ShowPreview {
		contentHeight = contentHeight / 2
	}
	
	// Ensure minimum height for content
	if contentHeight < 10 {
		contentHeight = 10
	}
	
	return contentHeight
}