package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenameRenderer handles the visual rendering of the rename dialog
type RenameRenderer struct {
	Width  int
	Height int
}

// NewRenameRenderer creates a new rename renderer
func NewRenameRenderer() *RenameRenderer {
	return &RenameRenderer{}
}

// SetSize updates the renderer dimensions
func (rr *RenameRenderer) SetSize(width, height int) {
	rr.Width = width
	rr.Height = height
}

// Render creates the visual representation of the rename dialog
func (rr *RenameRenderer) Render(state *RenameState) string {
	if !state.Active {
		return ""
	}

	// Calculate dialog dimensions early so we can use them
	dialogWidth := rr.Width / 2
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	var b strings.Builder

	// Title based on item type
	title := fmt.Sprintf("RENAME %s", strings.ToUpper(state.ItemType))
	if state.IsArchived {
		title = fmt.Sprintf("RENAME ARCHIVED %s", strings.ToUpper(state.ItemType))
	}

	// Title with style
	b.WriteString(TypeHeaderStyle.Render(title))
	b.WriteString("\n\n")

	// Current name display
	b.WriteString(HeaderStyle.Render("Current:"))
	b.WriteString(" ")
	b.WriteString(state.OriginalName)
	b.WriteString("\n\n")

	// Input field
	b.WriteString(HeaderStyle.Render("New name:"))
	b.WriteString("\n")
	
	// Calculate input field width (dialog width minus some padding)
	inputFieldWidth := dialogWidth - 8
	if inputFieldWidth < 40 {
		inputFieldWidth = 40
	}
	
	// Create highlighted input style (similar to selected items)
	inputFieldStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSelected)).
		Foreground(lipgloss.Color(ColorNormal)).
		Width(inputFieldWidth).
		Padding(0, 1)
	
	// Create cursor style with inverted colors for visibility
	cursorOnHighlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorActive)).
		Foreground(lipgloss.Color(ColorWhite)).
		Bold(true)
	
	if state.NewName == "" {
		// Show placeholder with cursor at beginning
		var inputContent strings.Builder
		inputContent.WriteString(cursorOnHighlightStyle.Render(" "))
		placeholder := "Enter new display name..."
		inputContent.WriteString(DescriptionStyle.Render(placeholder))
		
		// Pad to fill the width
		currentLen := 1 + len(placeholder)
		if currentLen < inputFieldWidth-2 {
			inputContent.WriteString(strings.Repeat(" ", inputFieldWidth-2-currentLen))
		}
		
		b.WriteString(inputFieldStyle.Render(inputContent.String()))
	} else {
		// Show input with cursor at correct position
		runes := []rune(state.NewName)
		var inputContent strings.Builder
		
		// Text before cursor
		if state.CursorPos > 0 {
			inputContent.WriteString(string(runes[:state.CursorPos]))
		}
		
		// Cursor
		if state.CursorPos < len(runes) {
			// Cursor on a character - highlight the character
			inputContent.WriteString(cursorOnHighlightStyle.Render(string(runes[state.CursorPos])))
			// Text after cursor
			if state.CursorPos+1 < len(runes) {
				inputContent.WriteString(string(runes[state.CursorPos+1:]))
			}
		} else {
			// Cursor at end - show a space with cursor styling
			inputContent.WriteString(cursorOnHighlightStyle.Render(" "))
		}
		
		// Pad to fill the width if needed
		textLen := len(runes) + 1
		if textLen < inputFieldWidth-2 {
			inputContent.WriteString(strings.Repeat(" ", inputFieldWidth-2-textLen))
		}
		
		// Apply the background highlight to the entire input
		b.WriteString(inputFieldStyle.Render(inputContent.String()))
	}
	b.WriteString("\n")

	// Show what filename this will become
	if state.NewName != "" && state.NewName != state.OriginalName {
		b.WriteString("\n")
		slugified := state.GetSlugifiedName()
		ext := ".md"
		if state.ItemType == "pipeline" {
			ext = ".yaml"
		}
		preview := fmt.Sprintf("Will save as: %s%s", slugified, ext)
		b.WriteString(DescriptionStyle.Render(preview))
		b.WriteString("\n")
	}

	// Validation error
	if state.ValidationError != "" {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("⚠ " + state.ValidationError))
		b.WriteString("\n")
	}

	// Affected pipelines (only for components)
	if state.ItemType == "component" && state.HasAffectedPipelines() {
		b.WriteString("\n")
		b.WriteString(HeaderStyle.Render("This will update references in:"))
		b.WriteString("\n")

		if len(state.AffectedActive) > 0 {
			b.WriteString("\n")
			b.WriteString(DescriptionStyle.Render("Active pipelines:"))
			b.WriteString("\n")
			for _, p := range state.AffectedActive {
				b.WriteString(fmt.Sprintf("  • %s\n", p))
			}
		}

		if len(state.AffectedArchive) > 0 {
			b.WriteString("\n")
			b.WriteString(DescriptionStyle.Render("Archived pipelines:"))
			b.WriteString("\n")
			for _, p := range state.AffectedArchive {
				b.WriteString(fmt.Sprintf("  • %s\n", p))
			}
		}
	}

	// Help text with colored key backgrounds
	b.WriteString("\n")
	
	// Create styles for the keys
	enterKeyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSuccess)).
		Foreground(lipgloss.Color(ColorWhite)).
		Bold(true)
	
	escKeyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorDanger)).
		Foreground(lipgloss.Color(ColorWhite)).
		Bold(true)
	
	// Build help text with colored keys
	var helpParts []string
	
	if state.IsValid() {
		helpParts = append(helpParts, "["+enterKeyStyle.Render("Enter")+"] Save")
		helpParts = append(helpParts, "["+escKeyStyle.Render("Esc")+"] Cancel")
	} else if state.NewName == state.OriginalName {
		helpParts = append(helpParts, "["+escKeyStyle.Render("Esc")+"] Cancel")
		helpParts = append(helpParts, DescriptionStyle.Render("(no changes)"))
	} else {
		helpParts = append(helpParts, "["+escKeyStyle.Render("Esc")+"] Cancel")
	}
	
	b.WriteString(strings.Join(helpParts, "  "))

	// Apply border and center the dialog
	dialogStyle := ActiveBorderStyle.
		Width(dialogWidth).
		Padding(1, 2)

	// Center the dialog on screen
	centeredStyle := lipgloss.NewStyle().
		Width(rr.Width).
		Height(rr.Height).
		Align(lipgloss.Center, lipgloss.Center)

	return centeredStyle.Render(dialogStyle.Render(b.String()))
}

// RenderOverlay creates an overlay view for the rename dialog
func (rr *RenameRenderer) RenderOverlay(baseView string, state *RenameState) string {
	if !state.Active {
		return baseView
	}

	// Create the rename dialog
	renameView := rr.Render(state)

	// Overlay the dialog on the base view
	return overlayViews(baseView, renameView)
}