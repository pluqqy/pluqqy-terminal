package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

// SharedLayout provides common layout calculations and rendering functions
// used by both Main List view and Pipeline Builder view
type SharedLayout struct {
	Width       int
	Height      int
	ShowPreview bool

	// Cached computed values for performance
	contentHeight int
	columnWidth   int
}

// NewSharedLayout creates a new shared layout with given dimensions
func NewSharedLayout(width, height int, showPreview bool) *SharedLayout {
	sl := &SharedLayout{
		Width:       width,
		Height:      height,
		ShowPreview: showPreview,
	}
	sl.recalculateDimensions()
	return sl
}

// SetSize updates the layout dimensions and recalculates cached values
func (sl *SharedLayout) SetSize(width, height int) {
	sl.Width = width
	sl.Height = height
	sl.recalculateDimensions()
}

// SetShowPreview updates the preview visibility and recalculates dimensions
func (sl *SharedLayout) SetShowPreview(show bool) {
	sl.ShowPreview = show
	sl.recalculateDimensions()
}

// recalculateDimensions updates cached dimension values
func (sl *SharedLayout) recalculateDimensions() {
	sl.columnWidth = (sl.Width - 6) / 2 // Account for gap, padding, and borders

	// Calculate content height
	// Base reservation: 13 lines (header, help pane, spacing)
	// Plus 3 lines for search bar
	sl.contentHeight = sl.Height - 16

	if sl.ShowPreview {
		sl.contentHeight = sl.contentHeight / 2
	}

	// Ensure minimum height
	if sl.contentHeight < 10 {
		sl.contentHeight = 10
	}
}

// GetContentHeight returns the calculated content height for columns
func (sl *SharedLayout) GetContentHeight() int {
	return sl.contentHeight
}

// GetColumnWidth returns the calculated width for each column
func (sl *SharedLayout) GetColumnWidth() int {
	return sl.columnWidth
}

// RenderHelpPane renders the help text in a bordered pane
func (sl *SharedLayout) RenderHelpPane(helpRows [][]string) string {
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(sl.Width-4). // Account for left/right padding (2) and borders (2)
		Padding(0, 1)      // Internal padding for help text

	helpContent := formatHelpTextRows(helpRows, sl.Width-8) // -8 for borders and padding

	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	return contentStyle.Render(helpBorderStyle.Render(helpContent))
}

// RenderHeader renders a header with colons and optional badge
func (sl *SharedLayout) RenderHeader(heading string, active bool, badge string, availableWidth int) string {
	// Dynamic header and colon styles based on active state
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if active {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))

	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if active {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))

	// Calculate badge width if provided
	badgeWidth := 0
	if badge != "" {
		badgeWidth = lipgloss.Width(badge)
	}

	// Calculate space for colons
	colonSpace := availableWidth - len(heading) - badgeWidth - 2 // -2 for spaces
	if badge != "" {
		colonSpace -= 2 // Additional spaces around badge
	}
	if colonSpace < 3 {
		colonSpace = 3
	}

	// Build header
	var result strings.Builder
	result.WriteString(headerStyle.Render(heading))
	result.WriteString(" ")
	result.WriteString(colonStyle.Render(strings.Repeat(":", colonSpace)))
	if badge != "" {
		result.WriteString(" ")
		result.WriteString(badge)
	}

	return result.String()
}

// RenderSearchBar renders the search bar at the top of the view
func (sl *SharedLayout) RenderSearchBar(searchBar *SearchBar) string {
	searchBar.SetWidth(sl.Width)
	return searchBar.View()
}

// PreviewConfig holds configuration for rendering the preview pane
type PreviewConfig struct {
	Content     string
	Heading     string
	ActivePane  interface{} // Can be pane or column type
	PreviewPane interface{} // The pane being previewed (for determining active state)
	Viewport    viewport.Model
}

// RenderPreviewPane renders the preview pane with border, header, and content
func (sl *SharedLayout) RenderPreviewPane(config PreviewConfig) string {
	if !sl.ShowPreview || config.Content == "" {
		return ""
	}

	// Calculate token count and create badge
	tokenCount := utils.EstimateTokens(config.Content)
	tokenBadgeStyle := GetTokenBadgeStyle(tokenCount)
	tokenBadge := tokenBadgeStyle.Render(utils.FormatTokenCount(tokenCount))

	// Determine if preview is active
	isActive := false
	// Handle both pane type (from MainListView) and column type (from Builder)
	switch v := config.ActivePane.(type) {
	case pane:
		isActive = v == previewPane
	case column:
		isActive = v == previewColumn
	}

	// Apply active/inactive style to preview border
	var previewBorderStyle lipgloss.Style
	if isActive {
		previewBorderStyle = ActiveBorderStyle
	} else {
		previewBorderStyle = InactiveBorderStyle
	}
	previewBorderStyle = previewBorderStyle.Width(sl.Width - 4) // Account for padding and border

	// Build preview content with header inside
	var previewContent strings.Builder

	// Calculate available width for header
	totalWidth := sl.Width - 8 // accounting for border padding and header padding

	// Render header with token badge
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	header := sl.RenderHeader(config.Heading, isActive, tokenBadge, totalWidth)
	previewContent.WriteString(headerPadding.Render(header))
	previewContent.WriteString("\n\n")

	// Preprocess and wrap content
	processedContent := preprocessContent(config.Content)
	wrappedContent := wordwrap.String(processedContent, config.Viewport.Width)
	config.Viewport.SetContent(wrappedContent)

	// Add padding to preview viewport content
	previewViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	previewContent.WriteString(previewViewportPadding.Render(config.Viewport.View()))

	// Render the border around the entire preview with padding
	var result strings.Builder
	result.WriteString("\n")
	previewPaddingStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	result.WriteString(previewPaddingStyle.Render(previewBorderStyle.Render(previewContent.String())))

	return result.String()
}

// ColumnHeaderConfig holds configuration for rendering column headers
type ColumnHeaderConfig struct {
	Heading     string
	Active      bool
	ColumnWidth int
}

// RenderColumnHeader renders a column header with colons
func (sl *SharedLayout) RenderColumnHeader(config ColumnHeaderConfig) string {
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Calculate remaining width for colons
	remainingWidth := config.ColumnWidth - len(config.Heading) - 5 // -5 for space and padding
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	header := sl.RenderHeader(config.Heading, config.Active, "", config.ColumnWidth-5)
	return headerPadding.Render(header)
}

// BuildConfirmationDialog builds a confirmation dialog string
func (sl *SharedLayout) BuildConfirmationDialog(confirmModel interface{}, message string, isDestructive bool) string {
	var confirmStyle lipgloss.Style
	if isDestructive {
		confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
	} else {
		confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
	}

	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Use type assertion to handle different confirmation models
	switch cm := confirmModel.(type) {
	case *ConfirmationModel:
		if cm.Active() {
			return "\n" + contentStyle.Render(confirmStyle.Render(cm.ViewWithWidth(sl.Width-4)))
		}
	}

	return ""
}
