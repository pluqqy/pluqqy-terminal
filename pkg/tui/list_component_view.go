package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	TableRenderer       *ComponentTableRenderer
}

// NewComponentViewRenderer creates a new component view renderer
func NewComponentViewRenderer(width, height int) *ComponentViewRenderer {
	return &ComponentViewRenderer{
		Width:         width,
		Height:        height,
		TableRenderer: NewComponentTableRenderer(width, height-6, true), // true for showUsageColumn
	}
}

// Render generates the complete component pane view
func (r *ComponentViewRenderer) Render() string {
	var content strings.Builder
	
	// Calculate column width
	columnWidth := (r.Width - 6) / 2 // Account for gap, padding, and ensure border visibility
	
	// Update table renderer state
	r.TableRenderer.SetSize(columnWidth, r.Height)
	r.TableRenderer.SetComponents(r.FilteredComponents)
	r.TableRenderer.SetCursor(r.ComponentCursor)
	r.TableRenderer.SetActive(r.ActivePane == componentsPane)
	
	// Handle empty state with search context
	if len(r.FilteredComponents) == 0 && len(r.AllComponents) > 0 && r.SearchQuery != "" {
		// Override empty message for search results
		r.TableRenderer.Components = nil // Force empty state
	}
	
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
	
	// Render table header
	content.WriteString(headerPadding.Render(r.TableRenderer.RenderHeader()))
	content.WriteString("\n\n")
	
	// Add padding to table content
	viewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	content.WriteString(viewportPadding.Render(r.TableRenderer.RenderTable()))
	
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

