package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
)

// tagStyle represents a tag with its display style
type tagStyle struct {
	name  string
	style lipgloss.Style
}

// renderTags renders tags as colored chips
func renderTags(tagNames []string, maxWidth int) string {
	if len(tagNames) == 0 {
		return ""
	}

	// Get tag registry for colors
	registry, err := tags.NewRegistry()
	if err != nil {
		// If registry fails, render without colors
		return renderTagsPlain(tagNames, maxWidth)
	}

	var tagStyles []tagStyle
	for _, tagName := range tagNames {
		tag, exists := registry.GetTag(tagName)
		color := models.GetTagColor(tagName, "")
		if exists && tag.Color != "" {
			color = tag.Color
		}

		// Create tag style with background color
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("255")). // White text
			Padding(0, 1).
			MarginRight(1)

		tagStyles = append(tagStyles, tagStyle{
			name:  tagName,
			style: style,
		})
	}

	// Render tags
	var result strings.Builder
	currentWidth := 0

	for i, ts := range tagStyles {
		rendered := ts.style.Render(ts.name)
		renderedWidth := lipgloss.Width(rendered)

		// Check if adding this tag would exceed max width
		if currentWidth+renderedWidth > maxWidth && i > 0 {
			result.WriteString("...")
			break
		}

		result.WriteString(rendered)
		currentWidth += renderedWidth
	}

	return result.String()
}

// renderTagsPlain renders tags without colors as a fallback
func renderTagsPlain(tagNames []string, maxWidth int) string {
	if len(tagNames) == 0 {
		return ""
	}

	// Simple comma-separated list
	tagStr := strings.Join(tagNames, ", ")
	if len(tagStr) > maxWidth {
		tagStr = tagStr[:maxWidth-3] + "..."
	}

	return tagStr
}

// renderTagChips renders tags as small colored chips for inline display
func renderTagChips(tagNames []string, maxTags int) string {
	if len(tagNames) == 0 {
		return ""
	}

	// Limit number of tags shown
	tagsToShow := tagNames
	if len(tagNames) > maxTags {
		tagsToShow = tagNames[:maxTags]
	}

	// Get tag registry for colors
	registry, err := tags.NewRegistry()
	if err != nil {
		// If registry fails, show plain
		result := strings.Join(tagsToShow, " ")
		if len(tagNames) > maxTags {
			result += " ..."
		}
		return result
	}

	var chips []string
	for _, tagName := range tagsToShow {
		tag, exists := registry.GetTag(tagName)
		color := models.GetTagColor(tagName, "")
		if exists && tag.Color != "" {
			color = tag.Color
		}

		// Create compact chip style
		chipStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)

		chips = append(chips, chipStyle.Render(tagName))
	}

	result := strings.Join(chips, " ")
	if len(tagNames) > maxTags {
		moreStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		result += " " + moreStyle.Render(fmt.Sprintf("+%d", len(tagNames)-maxTags))
	}

	return result
}

// renderTagChipsWithWidth renders tags as colored chips with width constraint
func renderTagChipsWithWidth(tagNames []string, maxWidth int, maxTags int) string {
	if len(tagNames) == 0 {
		return ""
	}

	// Get tag registry for colors
	registry, err := tags.NewRegistry()
	if err != nil {
		// If registry fails, show plain
		result := strings.Join(tagNames, " ")
		if len(result) > maxWidth-3 {
			result = result[:maxWidth-3] + "..."
		}
		return result
	}

	// Reserve space for the "+N" indicator if we have more tags than maxTags
	reserveForMore := 0
	if len(tagNames) > maxTags {
		// Calculate space needed for "+N" indicator
		moreText := fmt.Sprintf("+%d", len(tagNames)-maxTags)
		moreStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		moreIndicator := moreStyle.Render(moreText)
		reserveForMore = lipgloss.Width(moreIndicator) + 1 // +1 for space before indicator
	}

	var chips []string
	currentWidth := 0
	tagsShown := 0
	availableWidth := maxWidth - reserveForMore

	for i, tagName := range tagNames {
		if tagsShown >= maxTags {
			break
		}

		tag, exists := registry.GetTag(tagName)
		color := models.GetTagColor(tagName, "")
		if exists && tag.Color != "" {
			color = tag.Color
		}

		// Create compact chip style
		chipStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)

		// Truncate individual tag if needed
		displayName := tagName
		maxTagWidth := availableWidth - currentWidth
		if i > 0 {
			maxTagWidth -= 1 // Space between chips
		}

		// Ensure minimum tag width
		if maxTagWidth < 6 { // Minimum to show "a..." with padding
			break
		}

		// Check if we need to truncate
		testChip := chipStyle.Render(displayName)
		testWidth := lipgloss.Width(testChip)

		if currentWidth+testWidth+(func() int {
			if i > 0 {
				return 1
			} else {
				return 0
			}
		}()) > availableWidth {
			// Need to truncate or skip
			if maxTagWidth > 8 { // Room for at least "ab..." with padding
				// Truncate the tag
				for len(displayName) > 1 {
					displayName = displayName[:len(displayName)-1]
					testChip = chipStyle.Render(displayName + "...")
					testWidth = lipgloss.Width(testChip)
					if currentWidth+testWidth+(func() int {
						if i > 0 {
							return 1
						} else {
							return 0
						}
					}()) <= availableWidth {
						displayName = displayName + "..."
						break
					}
				}
			} else {
				// No room, stop here
				break
			}
		}

		chip := chipStyle.Render(displayName)
		chipWidth := lipgloss.Width(chip)

		// Add the chip
		chips = append(chips, chip)
		currentWidth += chipWidth
		if i > 0 {
			currentWidth += 1 // Space between chips
		}
		tagsShown++
	}

	result := strings.Join(chips, " ")
	if tagsShown < len(tagNames) {
		moreStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		result += " " + moreStyle.Render(fmt.Sprintf("+%d", len(tagNames)-tagsShown))
	}

	return result
}
