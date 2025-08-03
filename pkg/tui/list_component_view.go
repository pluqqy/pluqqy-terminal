package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComponentViewRenderer handles rendering of the component pane
type ComponentViewRenderer struct {
	Width               int
	Height              int
	ActivePane          pane
	FilteredComponents  []componentItem
	AllComponents       []componentItem
	ComponentCursor     int
	SearchQuery         string
	Viewport            viewport.Model
}

// NewComponentViewRenderer creates a new component view renderer
func NewComponentViewRenderer(width, height int) *ComponentViewRenderer {
	return &ComponentViewRenderer{
		Width:  width,
		Height: height,
	}
}

// Render generates the complete component pane view
func (r *ComponentViewRenderer) Render() string {
	var content strings.Builder
	
	// Calculate column width
	columnWidth := (r.Width - 6) / 2 // Account for gap, padding, and ensure border visibility
	
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Create heading with colons spanning the width
	heading := "COMPONENTS"
	remainingWidth := columnWidth - len(heading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	// Dynamic header and colon styles based on active pane
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == componentsPane {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if r.ActivePane == componentsPane {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	
	content.WriteString(headerPadding.Render(headerStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")
	
	// Table header styles
	headerStyle = HeaderStyle
	
	// Table column widths (adjusted for viewport width)
	viewportWidth := columnWidth - 4 // Same as viewport.Width
	nameWidth, tagsWidth, tokenWidth, usageWidth := formatColumnWidths(viewportWidth, true)
	
	// Render table header with consistent spacing
	header := fmt.Sprintf("  %-*s %-*s  %*s %*s", 
		nameWidth, "Name",
		tagsWidth, "Tags",
		tokenWidth, "~Tokens",
		usageWidth, "Usage")
	content.WriteString(headerPadding.Render(headerStyle.Render(header)))
	content.WriteString("\n\n")
	
	// Build scrollable content for components viewport
	scrollContent := r.buildScrollableContent(nameWidth, tagsWidth, tokenWidth, usageWidth)
	
	// Update viewport with content
	r.Viewport.SetContent(scrollContent)
	
	// Update viewport to follow cursor
	r.updateViewportScroll()
	
	// Add padding to viewport content
	viewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	content.WriteString(viewportPadding.Render(r.Viewport.View()))
	
	// Apply border
	borderStyle := InactiveBorderStyle
	if r.ActivePane == componentsPane {
		borderStyle = ActiveBorderStyle
	}
	
	return borderStyle.
		Width(columnWidth).
		Height(r.Height).
		Render(content.String())
}

// buildScrollableContent creates the content for the components viewport
func (r *ComponentViewRenderer) buildScrollableContent(nameWidth, tagsWidth, tokenWidth, usageWidth int) string {
	var content strings.Builder
	normalStyle := NormalStyle
	typeHeaderStyle := TypeHeaderStyle
	
	// Use filtered components instead of all components
	if len(r.FilteredComponents) == 0 {
		if r.ActivePane == componentsPane {
			// Active pane - show prominent message
			emptyStyle := EmptyActiveStyle
			
			// Check if we have components but they're filtered out
			if len(r.AllComponents) > 0 && r.SearchQuery != "" {
				content.WriteString(emptyStyle.Render("No components match your search."))
			} else {
				content.WriteString(emptyStyle.Render("No components found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := EmptyInactiveStyle
			
			// Check if we have components but they're filtered out
			if len(r.AllComponents) > 0 && r.SearchQuery != "" {
				content.WriteString(dimmedStyle.Render("No components match your search."))
			} else {
				content.WriteString(dimmedStyle.Render("No components found."))
			}
		}
	} else {
		currentType := ""
		
		for i, comp := range r.FilteredComponents {
			if comp.compType != currentType {
				if currentType != "" {
					content.WriteString("\n")
				}
				currentType = comp.compType
				// Convert to uppercase plural form
				var typeHeader string
				switch currentType {
				case models.ComponentTypeContext:
					typeHeader = "CONTEXTS"
				case models.ComponentTypePrompt:
					typeHeader = "PROMPTS"
				case models.ComponentTypeRules:
					typeHeader = "RULES"
				default:
					typeHeader = strings.ToUpper(currentType)
				}
				content.WriteString(typeHeaderStyle.Render(fmt.Sprintf("▸ %s", typeHeader)) + "\n")
			}
			
			// Format the row data
			nameStr := truncateName(comp.name, nameWidth)
			
			// Format usage count
			usageStr := fmt.Sprintf("%d", comp.usageCount)
			
			// Format token count - right-aligned with consistent width
			tokenStr := fmt.Sprintf("%d", comp.tokenCount)
			
			// Format tags
			tagsStr := renderTagChipsWithWidth(comp.tags, tagsWidth, 2) // Show max 2 tags inline
			
			// Build the row components separately for proper styling
			namePart := fmt.Sprintf("%-*s", nameWidth, nameStr)
			
			// For tags, we need to pad based on rendered width
			tagsPadding := tagsWidth - lipgloss.Width(tagsStr)
			if tagsPadding < 0 {
				tagsPadding = 0
			}
			tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)
			
			tokenPart := fmt.Sprintf("%*s", tokenWidth, tokenStr)
			usagePart := fmt.Sprintf("%*s", usageWidth, usageStr)
			
			// Build row with styling
			var row string
			if r.ActivePane == componentsPane && i == r.ComponentCursor {
				// Apply selection styling only to name column
				row = "▸ " + SelectedStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart + " " + usagePart)
			} else {
				// Normal row styling
				row = "  " + normalStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart + " " + usagePart)
			}
			
			content.WriteString(row)
			
			if i < len(r.FilteredComponents)-1 {
				content.WriteString("\n")
			}
		}
	}
	
	return content.String()
}

// updateViewportScroll updates the viewport to follow the cursor
func (r *ComponentViewRenderer) updateViewportScroll() {
	if r.ActivePane == componentsPane && len(r.FilteredComponents) > 0 {
		// Calculate the line position of the cursor
		currentLine := 0
		for i := 0; i < r.ComponentCursor && i < len(r.FilteredComponents); i++ {
			currentLine++ // Component line
			// Check if it's the first item of a new type to add header line
			if i == 0 || r.FilteredComponents[i].compType != r.FilteredComponents[i-1].compType {
				currentLine++ // Type header line
				if i > 0 {
					currentLine++ // Empty line between sections
				}
			}
		}
		
		// Ensure the cursor line is visible
		if currentLine < r.Viewport.YOffset {
			r.Viewport.SetYOffset(currentLine)
		} else if currentLine >= r.Viewport.YOffset+r.Viewport.Height {
			r.Viewport.SetYOffset(currentLine - r.Viewport.Height + 1)
		}
	}
}