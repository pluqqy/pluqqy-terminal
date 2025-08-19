package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
)

// TagEditingViewRenderer handles the rendering of tag editing views
type TagEditingViewRenderer struct {
	Width            int
	Height           int
	ItemName         string
	CurrentTags      []string
	TagInput         string
	TagCursor        int
	ShowSuggestions  bool
	SuggestionCursor int
	TagCloudActive   bool
	TagCloudCursor   int
	AvailableTags    []string

	// Tag deletion state
	TagDeleteConfirm *ConfirmationModel
	DeletingTag      string
	DeletingTagUsage *tags.UsageStats

	// Callback for getting tag suggestions
	GetSuggestionsFunc func(input string, availableTags []string, currentTags []string) []string

	// Callback for getting available tags for cloud
	GetAvailableTagsForCloudFunc func(availableTags []string, currentTags []string) []string
}

// NewTagEditingViewRenderer creates a new tag editing view renderer
func NewTagEditingViewRenderer(width, height int) *TagEditingViewRenderer {
	return &TagEditingViewRenderer{
		Width:  width,
		Height: height,
	}
}

// Render returns the complete tag editing view
func (r *TagEditingViewRenderer) Render() string {
	// Show deletion confirmation if active
	if r.TagDeleteConfirm != nil && r.TagDeleteConfirm.Active() {
		return r.TagDeleteConfirm.View()
	}

	// Calculate dimensions for side-by-side layout
	paneWidth := (r.Width - 6) / 2
	paneHeight := r.Height - 10

	// Render main editing pane
	mainPane := r.renderMainPane(paneWidth, paneHeight)

	// Render tag cloud pane
	tagCloudPane := r.renderTagCloudPane(paneWidth, paneHeight)

	// Render help section
	helpSection := r.renderHelpSection()

	// Combine panes side by side
	sideBySide := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainPane,
		" ",
		tagCloudPane,
	)

	// Combine all elements
	var s strings.Builder
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	s.WriteString(contentStyle.Render(sideBySide))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpSection))

	return s.String()
}

// renderMainPane renders the main tag editing pane
func (r *TagEditingViewRenderer) renderMainPane(paneWidth, paneHeight int) string {
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	var content strings.Builder

	// Render title
	heading := fmt.Sprintf("EDIT TAGS: %s", strings.ToUpper(r.ItemName))
	remainingWidth := paneWidth - len(heading) - 7
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Dynamic styles based on which pane is active
	mainHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if !r.TagCloudActive {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	mainColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if !r.TagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))

	content.WriteString(headerPadding.Render(mainHeaderStyle.Render(heading) + " " + mainColonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")

	// Render current tags
	content.WriteString(headerPadding.Render("Current tags:\n"))
	content.WriteString(headerPadding.Render(r.renderCurrentTags()))
	content.WriteString("\n\n")

	// Render input field
	content.WriteString(headerPadding.Render("Add tag:"))
	content.WriteString("\n")
	content.WriteString(headerPadding.Render(r.renderInputField()))
	content.WriteString("\n\n")

	// Render suggestions
	if r.ShowSuggestions && len(r.TagInput) > 0 {
		content.WriteString(headerPadding.Render("Suggestions:\n"))
		content.WriteString(r.renderSuggestions(headerPadding))
		content.WriteString("\n")
	}

	// Apply border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if !r.TagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))

	return borderStyle.
		Width(paneWidth).
		Height(paneHeight).
		Render(content.String())
}

// renderCurrentTags renders the current tags with selection
func (r *TagEditingViewRenderer) renderCurrentTags() string {
	if len(r.CurrentTags) == 0 {
		dimStyle := DescriptionStyle
		return dimStyle.Render("(no tags)")
	}

	var tagDisplay strings.Builder
	for i, tag := range r.CurrentTags {
		// Get color from registry
		registry, _ := tags.NewRegistry()
		color := models.GetTagColor(tag, "")
		if registry != nil {
			if t, exists := registry.GetTag(tag); exists && t.Color != "" {
				color = t.Color
			}
		}

		style := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)

		// Add selection indicators with consistent spacing
		if i == r.TagCursor && r.TagInput == "" {
			// Selected tag with triangle indicators
			indicatorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)
			tagDisplay.WriteString(indicatorStyle.Render("▶ "))
			tagDisplay.WriteString(style.Render(tag))
			tagDisplay.WriteString(indicatorStyle.Render(" ◀"))
		} else {
			// Add invisible spacers to maintain consistent width
			tagDisplay.WriteString("  ")
			tagDisplay.WriteString(style.Render(tag))
			tagDisplay.WriteString("  ")
		}

		// Add space between tags
		if i < len(r.CurrentTags)-1 {
			tagDisplay.WriteString("  ") // Double space for better separation
		}
	}

	return tagDisplay.String()
}

// renderInputField renders the tag input field
func (r *TagEditingViewRenderer) renderInputField() string {
	inputStyle := InputStyle.Width(40)
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	inputDisplay := r.TagInput
	if !r.TagCloudActive && r.TagInput != "" {
		// Add cursor to existing input when active
		inputDisplay = r.TagInput + cursorStyle.Render("│")
	}

	// Show placeholder if empty
	if r.TagInput == "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		if !r.TagCloudActive {
			inputDisplay = placeholderStyle.Render("Type to add a new tag...") + cursorStyle.Render("│")
		} else {
			inputDisplay = placeholderStyle.Render("Type to add a new tag...")
		}
	}

	// Highlight input border when active
	activeInputStyle := inputStyle
	if !r.TagCloudActive {
		activeInputStyle = inputStyle.BorderForeground(lipgloss.Color("170"))
	}

	return activeInputStyle.Render(inputDisplay)
}

// renderSuggestions renders the tag suggestions
func (r *TagEditingViewRenderer) renderSuggestions(padding lipgloss.Style) string {
	if r.GetSuggestionsFunc == nil {
		return ""
	}

	suggestions := r.GetSuggestionsFunc(r.TagInput, r.AvailableTags, r.CurrentTags)
	suggestionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	selectedSuggestionStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("170"))

	var content strings.Builder
	for i, suggestion := range suggestions {
		if i > 5 { // Limit to 6 suggestions
			break
		}
		var suggestionLine string
		if i == r.SuggestionCursor {
			// Selected suggestion with background
			suggestionLine = selectedSuggestionStyle.
				Padding(0, 1).
				Render(suggestion)
		} else {
			// Regular suggestion
			suggestionLine = suggestionStyle.Render("  " + suggestion)
		}
		content.WriteString(padding.Render(suggestionLine))
		content.WriteString("\n")
	}

	return content.String()
}

// renderTagCloudPane renders the tag cloud pane
func (r *TagEditingViewRenderer) renderTagCloudPane(paneWidth, paneHeight int) string {
	tagCloudRenderer := NewTagCloudRenderer(paneWidth, paneHeight)
	tagCloudRenderer.IsActive = r.TagCloudActive
	tagCloudRenderer.CursorIndex = r.TagCloudCursor

	// Get available tags for cloud (excluding current tags)
	if r.GetAvailableTagsForCloudFunc != nil {
		tagCloudRenderer.AvailableTags = r.GetAvailableTagsForCloudFunc(r.AvailableTags, r.CurrentTags)
	} else {
		tagCloudRenderer.AvailableTags = r.AvailableTags
	}

	tagCloudRenderer.CurrentTags = r.CurrentTags

	return tagCloudRenderer.Render()
}

// renderHelpSection renders the help section
func (r *TagEditingViewRenderer) renderHelpSection() string {
	var help []string
	if r.TagCloudActive {
		tagCloudRenderer := NewTagCloudRenderer(0, 0)
		help = tagCloudRenderer.GetHelpText()
	} else {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←/→ select tag",
			"↑↓ navigate suggestions",
			"^d delete tag",
			"^t reload tags",
			"^s save",
			"esc cancel",
		}
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.Width-4).
		Padding(0, 1)

	helpContent := formatHelpText(help)

	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.Width - 8).
		Align(lipgloss.Right).
		Render(helpContent)

	return helpBorderStyle.Render(alignedHelp)
}
