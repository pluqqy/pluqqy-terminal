package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// TagEditorRenderer handles the rendering of the tag editor
type TagEditorRenderer struct {
	Editor *TagEditor
	Width  int
	Height int
}

// NewTagEditorRenderer creates a new tag editor renderer
func NewTagEditorRenderer(editor *TagEditor, width, height int) *TagEditorRenderer {
	return &TagEditorRenderer{
		Editor: editor,
		Width:  width,
		Height: height,
	}
}

// Render returns the complete tag editor view
func (ter *TagEditorRenderer) Render() string {
	// Handle confirmation dialogs
	if ter.Editor.TagDeleteConfirm.Active() {
		// Add padding to match other views (same as Enhanced Editor)
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		return contentStyle.Render(ter.Editor.TagDeleteConfirm.View())
	}
	
	if ter.Editor.ExitConfirm.Active() {
		// Add padding to match other views (same as Enhanced Editor)
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		return contentStyle.Render(ter.Editor.ExitConfirm.View())
	}
	
	// Calculate dimensions for side-by-side layout
	paneWidth := (ter.Width - 6) / 2
	paneHeight := ter.Height - 10 // Leave room for help pane
	
	// Render main pane
	mainPane := ter.renderMainPane(paneWidth, paneHeight)
	
	// Render tag cloud pane
	cloudPane := ter.renderCloudPane(paneWidth, paneHeight)
	
	// Join panes side by side
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainPane,
		" ",
		cloudPane,
	)
	
	// Render help section
	helpView := ter.renderHelp()
	
	// Combine all views
	var s strings.Builder
	contentStyle := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	
	s.WriteString(contentStyle.Render(mainView))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpView))
	
	finalView := s.String()
	
	// Overlay spinner if deletion is in progress
	if ter.Editor.TagDeletionState != nil && ter.Editor.TagDeletionState.Active {
		overlay := ter.Editor.TagDeletionState.View()
		if overlay != "" {
			return overlayViews(finalView, overlay)
		}
	}
	
	// Overlay tag reload status if active
	if ter.Editor.TagReloader != nil && ter.Editor.TagReloader.IsActive() {
		reloadRenderer := NewTagReloadRenderer(ter.Width, ter.Height)
		overlay := reloadRenderer.RenderStatus(ter.Editor.TagReloader)
		if overlay != "" {
			return overlayViews(finalView, overlay)
		}
	}
	
	return finalView
}

// renderMainPane renders the left pane with current tags and input
func (ter *TagEditorRenderer) renderMainPane(width, height int) string {
	var content strings.Builder
	headerPadding := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	
	// Header
	heading := fmt.Sprintf("EDIT TAGS: %s", strings.ToUpper(ter.Editor.ItemName))
	remainingWidth := width - len(heading) - 7
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	// Dynamic styles based on which pane is active
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if !ter.Editor.TagCloudActive {
				return ColorActive // Purple when active
			}
			return ColorWarning // Orange when inactive
		}()))
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if !ter.Editor.TagCloudActive {
				return ColorActive // Purple when active
			}
			return ColorDim // Gray when inactive
		}()))
	
	content.WriteString(headerPadding.Render(headerStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")
	
	// Current tags section
	content.WriteString(headerPadding.Render("Current tags:"))
	content.WriteString("\n\n")
	
	if len(ter.Editor.CurrentTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorVeryDim))
		content.WriteString(headerPadding.Render(dimStyle.Render("  (no tags)")))
	} else {
		// Group tags in rows for better display
		var tagRows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		maxRowWidth := width - 6
		
		// Render tags with cursor
		for i, tag := range ter.Editor.CurrentTags {
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}
			
			tagStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(color)).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
			
			var tagDisplay string
			if i == ter.Editor.TagCursor && ter.Editor.TagInput == "" && !ter.Editor.TagCloudActive {
				// Show cursor
				cursorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorActive)).
					Bold(true)
				tagDisplay = cursorStyle.Render("▶ ") + tagStyle.Render(tag) + cursorStyle.Render(" ◀")
			} else {
				tagDisplay = "  " + tagStyle.Render(tag) + "  "
			}
			
			tagWidth := lipgloss.Width(tagDisplay) + 2
			
			// Check if we need to start a new row
			if rowTags > 0 && currentRowWidth+tagWidth+2 > maxRowWidth {
				tagRows.WriteString("\n\n")
				rowTags = 0
				currentRowWidth = 0
			}
			
			// Add the tag
			if rowTags > 0 {
				tagRows.WriteString("  ")
				currentRowWidth += 2
			}
			tagRows.WriteString(tagDisplay)
			currentRowWidth += tagWidth
			rowTags++
		}
		
		content.WriteString(headerPadding.Render(tagRows.String()))
	}
	content.WriteString("\n\n")
	
	// Input section
	content.WriteString(headerPadding.Render("Add tag:"))
	content.WriteString("\n")
	
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if !ter.Editor.TagCloudActive {
				return ColorActive
			}
			return ColorDim
		}())).
		Padding(0, 1).
		Width(width - 4)
	
	inputContent := ter.Editor.TagInput
	if inputContent == "" && !ter.Editor.TagCloudActive {
		inputContent = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorVeryDim)).Render("type to add tag...")
	}
	content.WriteString(headerPadding.Render(inputStyle.Render(inputContent)))
	
	// Suggestions
	if ter.Editor.ShowSuggestions && ter.Editor.TagInput != "" {
		content.WriteString("\n")
		suggestions := ter.Editor.GetSuggestions()
		if len(suggestions) > 0 {
			content.WriteString(headerPadding.Render("Suggestions:"))
			content.WriteString("\n")
			
			maxSuggestions := len(suggestions)
			if maxSuggestions > 6 {
				maxSuggestions = 6
			}
			
			for i := 0; i < maxSuggestions; i++ {
				suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorVeryDim))
				if i == ter.Editor.SuggestionCursor {
					suggestionStyle = lipgloss.NewStyle().
						Background(lipgloss.Color("236")).
						Foreground(lipgloss.Color(ColorActive))
				}
				content.WriteString(headerPadding.Render("  " + suggestionStyle.Render(suggestions[i])))
				content.WriteString("\n")
			}
		}
	}
	
	// Border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if !ter.Editor.TagCloudActive {
				return ColorActive
			}
			return ColorDim
		}())).
		Width(width).
		Height(height)
	
	return borderStyle.Render(content.String())
}

// renderCloudPane renders the right pane with available tags
func (ter *TagEditorRenderer) renderCloudPane(width, height int) string {
	var content strings.Builder
	headerPadding := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	
	// Header
	heading := "AVAILABLE TAGS"
	remainingWidth := width - len(heading) - 7
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if ter.Editor.TagCloudActive {
				return ColorActive
			}
			return ColorWarning
		}()))
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if ter.Editor.TagCloudActive {
				return ColorActive
			}
			return ColorDim
		}()))
	
	content.WriteString(headerPadding.Render(headerStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")
	
	// Get available tags that haven't been added yet
	availableForCloud := ter.Editor.GetAvailableTagsForCloud()
	
	if len(availableForCloud) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorVeryDim))
		content.WriteString(headerPadding.Render(dimStyle.Render("  (no available tags)")))
	} else {
		// Group tags in rows for better display
		var tagRows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		maxRowWidth := width - 6
		
		for i, tag := range availableForCloud {
			// Get tag color
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}
			
			tagStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(color)).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
			
			// Calculate tag display width
			var tagDisplay string
			if ter.Editor.TagCloudActive && i == ter.Editor.TagCloudCursor {
				indicatorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorActive)).
					Bold(true)
				tagDisplay = indicatorStyle.Render("▶ ") + tagStyle.Render(tag) + indicatorStyle.Render(" ◀")
			} else {
				tagDisplay = "  " + tagStyle.Render(tag) + "  "
			}
			
			tagWidth := lipgloss.Width(tagDisplay) + 2
			
			// Check if we need to start a new row
			if rowTags > 0 && currentRowWidth+tagWidth+2 > maxRowWidth {
				tagRows.WriteString("\n\n")
				rowTags = 0
				currentRowWidth = 0
			}
			
			// Add the tag
			if rowTags > 0 {
				tagRows.WriteString("  ")
				currentRowWidth += 2
			}
			tagRows.WriteString(tagDisplay)
			currentRowWidth += tagWidth
			rowTags++
		}
		
		content.WriteString(headerPadding.Render(tagRows.String()))
	}
	
	// Border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if ter.Editor.TagCloudActive {
				return ColorActive
			}
			return ColorDim
		}())).
		Width(width).
		Height(height)
	
	return borderStyle.Render(content.String())
}

// renderHelp renders the help section
func (ter *TagEditorRenderer) renderHelp() string {
	var help []string
	
	if ter.Editor.TagCloudActive {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←→ navigate",
			"^d delete tag",
			"^s save",
			"esc cancel",
		}
	} else {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←↑↓→ navigate",
			"^d remove tag",
			"^t reload tags",
			"^s save",
			"esc cancel",
		}
	}
	
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorDim)).
		Width(ter.Width - 4).
		Padding(0, 1)
	
	helpContent := formatHelpText(help)
	
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(ter.Width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	
	return helpBorderStyle.Render(alignedHelp)
}