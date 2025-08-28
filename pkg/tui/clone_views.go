package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

// CloneRenderer handles the visual rendering of the clone dialog
type CloneRenderer struct {
	Width  int
	Height int
}

// NewCloneRenderer creates a new clone renderer instance
func NewCloneRenderer() *CloneRenderer {
	return &CloneRenderer{}
}

// SetSize updates the renderer dimensions
func (cr *CloneRenderer) SetSize(width, height int) {
	cr.Width = width
	cr.Height = height
}

// Render creates the visual representation of the clone dialog
func (cr *CloneRenderer) Render(state *CloneState) string {
	// Lock for reading state
	state.mu.RLock()
	active := state.Active
	if !active {
		state.mu.RUnlock()
		return ""
	}

	// Copy needed values while holding lock
	itemType := state.ItemType
	isArchived := state.IsArchived
	originalName := state.OriginalName
	newName := state.NewName
	cursorPos := state.CursorPos
	cloneToArchive := state.CloneToArchive
	validationError := state.ValidationError
	isValid := state.NewName != "" && state.ValidationError == ""
	slugified := ""
	if state.NewName != "" {
		slugified = files.Slugify(state.NewName)
	}
	state.mu.RUnlock()

	var b strings.Builder

	// Title based on item type
	title := fmt.Sprintf("CLONE %s", strings.ToUpper(itemType))
	if isArchived {
		title = fmt.Sprintf("CLONE ARCHIVED %s", strings.ToUpper(itemType))
	}

	// Title with style
	b.WriteString(TypeHeaderStyle.Render(title))
	b.WriteString("\n\n")

	// Original name display
	b.WriteString(HeaderStyle.Render("Original:"))
	b.WriteString(" ")
	b.WriteString(originalName)
	b.WriteString("\n\n")

	// Input field
	b.WriteString(HeaderStyle.Render("New name:"))
	b.WriteString(" ")

	if newName == "" {
		// Show placeholder with cursor
		placeholder := DescriptionStyle.Render("Enter name for clone...")
		b.WriteString(placeholder)
		b.WriteString(CursorStyle.Render("█"))
	} else {
		// Show input with cursor at correct position
		runes := []rune(newName)
		if cursorPos >= 0 && cursorPos <= len(runes) {
			// Show text before cursor
			if cursorPos > 0 {
				b.WriteString(string(runes[:cursorPos]))
			}
			// Show cursor
			b.WriteString(CursorStyle.Render("█"))
			// Show text after cursor
			if cursorPos < len(runes) {
				b.WriteString(string(runes[cursorPos:]))
			}
		} else {
			// Fallback to showing cursor at end
			b.WriteString(newName)
			b.WriteString(CursorStyle.Render("█"))
		}
	}
	b.WriteString("\n")

	// Show what filename this will become
	if newName != "" {
		b.WriteString("\n")
		ext := ".md"
		if itemType == "pipeline" {
			ext = ".yaml"
		}
		preview := fmt.Sprintf("Will save as: %s%s", slugified, ext)
		b.WriteString(DescriptionStyle.Render(preview))
		b.WriteString("\n")
	}

	// Clone destination
	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render("Destination:"))
	b.WriteString(" ")

	if cloneToArchive {
		b.WriteString("Archive")
		b.WriteString(DescriptionStyle.Render(" (Tab to change)"))
	} else {
		b.WriteString("Active")
		b.WriteString(DescriptionStyle.Render(" (Tab to change)"))
	}
	b.WriteString("\n")

	// Validation error
	if validationError != "" {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("⚠ " + validationError))
		b.WriteString("\n")
	}

	// Help text with colored key backgrounds
	b.WriteString("\n\n") // Extra line for consistent spacing

	// Create styles for the keys
	enterKeyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSuccess)).
		Foreground(lipgloss.Color(ColorWhite)).
		Padding(0, 1).
		Bold(true)

	tabKeyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorWarning)).
		Foreground(lipgloss.Color(ColorDark)).
		Padding(0, 1).
		Bold(true)

	escKeyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorDanger)).
		Foreground(lipgloss.Color(ColorWhite)).
		Padding(0, 1).
		Bold(true)

	// Build help text with colored keys
	var helpParts []string

	if isValid {
		helpParts = append(helpParts, enterKeyStyle.Render("Enter")+" Clone")
	}

	helpParts = append(helpParts, tabKeyStyle.Render("Tab")+" Toggle destination")
	helpParts = append(helpParts, escKeyStyle.Render("Esc")+" Cancel")

	b.WriteString(strings.Join(helpParts, "  "))

	// Calculate dialog dimensions
	dialogWidth := cr.Width / 2
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	// Apply border and center the dialog
	dialogStyle := ActiveBorderStyle.
		Width(dialogWidth).
		Padding(1, 2)

	// Center the dialog on screen
	centeredStyle := lipgloss.NewStyle().
		Width(cr.Width).
		Height(cr.Height).
		Align(lipgloss.Center, lipgloss.Center)

	return centeredStyle.Render(dialogStyle.Render(b.String()))
}

// RenderOverlay creates an overlay view for the clone dialog
func (cr *CloneRenderer) RenderOverlay(baseView string, state *CloneState) string {
	if !state.Active {
		return baseView
	}

	// Create the clone dialog
	cloneView := cr.Render(state)

	// Overlay the dialog on the base view
	return overlayViews(baseView, cloneView)
}
