package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComponentTableRenderer handles standardized rendering of component tables.
// It provides a consistent UI for displaying components across different views
// in the application, including the Main List View and Pipeline Builder.
//
// The renderer manages column widths dynamically, handles scrolling through
// viewports, and provides visual indicators for selected, archived, and
// already-added components.
type ComponentTableRenderer struct {
	Width              int
	Height             int
	Components         []componentItem
	Cursor             int
	IsActive           bool
	ShowUsageColumn    bool
	ShowArchived       bool
	ShowAddedIndicator bool
	AddedComponents    map[string]bool // For tracking which components are already added
	Viewport           viewport.Model
}

// NewComponentTableRenderer creates a new component table renderer with the specified dimensions.
//
// Parameters:
//   - width: The total width available for the table
//   - height: The total height available for the table
//   - showUsageColumn: Whether to display the usage count column (true for Main List View, false for other views)
//
// Returns a configured ComponentTableRenderer ready for use.
func NewComponentTableRenderer(width, height int, showUsageColumn bool) *ComponentTableRenderer {
	vp := viewport.New(width-4, height-6) // Account for padding and headers
	return &ComponentTableRenderer{
		Width:           width,
		Height:          height,
		ShowUsageColumn: showUsageColumn,
		Viewport:        vp,
		AddedComponents: make(map[string]bool),
	}
}

// SetSize updates the dimensions of the table renderer.
// This should be called when the terminal is resized or when the available
// space for the table changes. The viewport is automatically adjusted to
// account for padding and headers.
func (r *ComponentTableRenderer) SetSize(width, height int) {
	r.Width = width
	r.Height = height
	r.Viewport.Width = width - 4
	r.Viewport.Height = height - 6
}

// SetComponents updates the list of components to display.
// This triggers a re-render of the table content and updates the viewport.
// The components should be pre-sorted and filtered as needed before passing.
func (r *ComponentTableRenderer) SetComponents(components []componentItem) {
	r.Components = components
	r.updateContent()
}

// SetCursor updates the cursor position to highlight a specific component.
// The viewport will automatically scroll to ensure the cursor is visible.
// The cursor value should be within the bounds of the component list.
func (r *ComponentTableRenderer) SetCursor(cursor int) {
	r.Cursor = cursor
	r.updateContent()
	r.updateViewportScroll()
}

// SetActive updates whether this table is the actively focused element.
// When active, the selected component will be highlighted and the table
// will respond to cursor movements.
func (r *ComponentTableRenderer) SetActive(active bool) {
	r.IsActive = active
	r.updateContent()
}

// MarkAsAdded marks a component as already added to the current selection.
// This is primarily used by the Pipeline Builder to show which components
// are already in the pipeline. Added components display a checkmark (✓).
//
// The componentPath should match the path format used in componentItem.
func (r *ComponentTableRenderer) MarkAsAdded(componentPath string) {
	r.AddedComponents[componentPath] = true
}

// ClearAddedMarks removes all "added" indicators from components.
// This resets the visual state when starting a new selection or clearing
// the current pipeline.
func (r *ComponentTableRenderer) ClearAddedMarks() {
	r.AddedComponents = make(map[string]bool)
}

// RenderHeader returns the formatted table header with column names.
// The header includes Name, Tags, ~Tokens, and optionally Usage columns.
// Column widths are calculated dynamically based on the available width.
func (r *ComponentTableRenderer) RenderHeader() string {
	// Calculate column widths using the shared helper
	viewportWidth := r.Width - 4 // Same as Main List View
	nameWidth, tagsWidth, tokenWidth, usageWidth := formatColumnWidths(viewportWidth, r.ShowUsageColumn)

	headerStyle := HeaderStyle

	var header string
	if r.ShowUsageColumn {
		// Components view with usage column
		header = fmt.Sprintf("  %-*s %-*s  %*s %*s",
			nameWidth, "Name",
			tagsWidth, "Tags",
			tokenWidth, "~Tokens",
			usageWidth, "Usage")
	} else {
		// Pipeline view without usage column (if needed)
		header = fmt.Sprintf("  %-*s %-*s  %*s",
			nameWidth, "Name",
			tagsWidth, "Tags",
			tokenWidth, "~Tokens")
	}

	return headerStyle.Render(header)
}

// RenderTable returns the rendered table content within the viewport.
// This includes all visible components with proper scrolling support.
// The content is already formatted and ready for display.
func (r *ComponentTableRenderer) RenderTable() string {
	return r.Viewport.View()
}

// updateContent rebuilds the content for the viewport.
// This is called internally whenever the component list, cursor, or active
// state changes. It recalculates column widths and regenerates the table.
func (r *ComponentTableRenderer) updateContent() {
	// Calculate column widths
	viewportWidth := r.Width - 4
	nameWidth, tagsWidth, tokenWidth, usageWidth := formatColumnWidths(viewportWidth, r.ShowUsageColumn)

	content := r.buildTableContent(nameWidth, tagsWidth, tokenWidth, usageWidth)
	r.Viewport.SetContent(content)
}

// buildTableContent creates the scrollable table content with proper formatting.
// It handles type headers, selection indicators, archived status, and added marks.
// The content is formatted according to the provided column widths.
//
// Parameters:
//   - nameWidth: Width allocated for the component name column
//   - tagsWidth: Width allocated for the tags column
//   - tokenWidth: Width allocated for the token count column
//   - usageWidth: Width allocated for the usage count column (0 if not shown)
//
// Returns the complete formatted table content as a string.
func (r *ComponentTableRenderer) buildTableContent(nameWidth, tagsWidth, tokenWidth, usageWidth int) string {
	var content strings.Builder

	normalStyle := NormalStyle
	dimmedStyle := EmptyInactiveStyle
	typeHeaderStyle := TypeHeaderStyle
	selectedStyle := SelectedStyle

	if len(r.Components) == 0 {
		if r.IsActive {
			emptyStyle := EmptyActiveStyle
			content.WriteString(emptyStyle.Render("No components found.\n\nPress 'n' to create one"))
		} else {
			content.WriteString(dimmedStyle.Render("No components found."))
		}
		return content.String()
	}

	currentType := ""
	for i, comp := range r.Components {
		// Add type headers
		if comp.compType != currentType {
			if currentType != "" {
				content.WriteString("\n")
			}
			currentType = comp.compType

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

		// Check if component is added (for Pipeline Builder)
		isAdded := false
		if r.ShowAddedIndicator {
			componentPath := "../" + comp.path
			isAdded = r.AddedComponents[componentPath]
		}

		// Format name with indicators
		nameStr := comp.name
		if comp.isArchived {
			nameStr = "[A] " + nameStr
		}
		if isAdded {
			nameStr = nameStr + " ✓"
		}
		nameStr = truncateName(nameStr, nameWidth)

		// Format other columns
		tokenStr := fmt.Sprintf("%d", comp.tokenCount)
		tagsStr := renderTagChipsWithWidth(comp.tags, tagsWidth, 2)

		// Build row parts
		namePart := fmt.Sprintf("%-*s", nameWidth, nameStr)

		// Pad tags based on rendered width
		tagsPadding := tagsWidth - lipgloss.Width(tagsStr)
		if tagsPadding < 0 {
			tagsPadding = 0
		}
		tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)

		tokenPart := fmt.Sprintf("%*s", tokenWidth, tokenStr)

		// Build complete row
		var row string
		rowPrefix := "  "

		// Determine styling
		isSelected := r.IsActive && i == r.Cursor
		if isSelected {
			rowPrefix = "▸ "
		}

		if r.ShowUsageColumn {
			usageStr := fmt.Sprintf("%d", comp.usageCount)
			usagePart := fmt.Sprintf("%*s", usageWidth, usageStr)

			if isSelected {
				if comp.isArchived {
					row = rowPrefix + dimmedStyle.Render(namePart) + " " + tagsPart + "  " + dimmedStyle.Render(tokenPart+" "+usagePart)
				} else {
					row = rowPrefix + selectedStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart+" "+usagePart)
				}
			} else if isAdded {
				addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
				row = rowPrefix + addedStyle.Render(namePart) + " " + tagsPart + "  " + addedStyle.Render(tokenPart+" "+usagePart)
			} else if comp.isArchived {
				row = rowPrefix + dimmedStyle.Render(namePart) + " " + tagsPart + "  " + dimmedStyle.Render(tokenPart+" "+usagePart)
			} else {
				row = rowPrefix + normalStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart+" "+usagePart)
			}
		} else {
			// Without usage column
			if isSelected {
				if comp.isArchived {
					row = rowPrefix + dimmedStyle.Render(namePart) + " " + tagsPart + "  " + dimmedStyle.Render(tokenPart)
				} else {
					row = rowPrefix + selectedStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart)
				}
			} else if isAdded {
				addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
				row = rowPrefix + addedStyle.Render(namePart) + " " + tagsPart + "  " + addedStyle.Render(tokenPart)
			} else if comp.isArchived {
				row = rowPrefix + dimmedStyle.Render(namePart) + " " + tagsPart + "  " + dimmedStyle.Render(tokenPart)
			} else {
				row = rowPrefix + normalStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart)
			}
		}

		content.WriteString(row)

		if i < len(r.Components)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

// updateViewportScroll adjusts the viewport offset to ensure the cursor is visible.
// It calculates the actual line position of the cursor, accounting for type headers
// and spacing between sections, then scrolls the viewport as needed.
// This method is called automatically when the cursor position changes.
func (r *ComponentTableRenderer) updateViewportScroll() {
	if !r.IsActive || len(r.Components) == 0 {
		return
	}

	// Calculate the line position of the cursor
	currentLine := 0
	for i := 0; i < r.Cursor && i < len(r.Components); i++ {
		currentLine++ // Component line
		// Check if it's the first item of a new type to add header line
		if i == 0 || r.Components[i].compType != r.Components[i-1].compType {
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
