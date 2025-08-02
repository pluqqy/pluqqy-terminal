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
	
	var chips []string
	currentWidth := 0
	tagsShown := 0
	
	for i, tagName := range tagNames {
		if tagsShown >= maxTags {
			break
		}
		
		tag, exists := registry.GetTag(tagName)
		color := models.GetTagColor(tagName, "")
		if exists && tag.Color != "" {
			color = tag.Color
		}
		
		// Truncate individual tag if needed
		displayName := tagName
		// Account for padding (2 chars) in chip
		maxTagWidth := maxWidth - currentWidth - 2
		if i > 0 {
			maxTagWidth -= 1 // Space between chips
		}
		
		// If this is the last tag we can show, leave room for "..."
		if i == maxTags-1 && len(tagNames) > maxTags {
			maxTagWidth -= 4 // Space + "..."
		}
		
		if len(displayName) > maxTagWidth-2 && maxTagWidth > 5 {
			displayName = displayName[:maxTagWidth-5] + "..."
		}
		
		// Create compact chip style
		chipStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)
		
		chip := chipStyle.Render(displayName)
		chipWidth := lipgloss.Width(chip)
		
		// Check if adding this chip exceeds width
		newWidth := currentWidth + chipWidth
		if i > 0 {
			newWidth += 1 // Space between chips
		}
		
		if newWidth > maxWidth && i > 0 {
			// Can't fit this chip
			break
		}
		
		chips = append(chips, chip)
		currentWidth = newWidth
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