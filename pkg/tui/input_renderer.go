package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// InputRenderer provides consistent rendering for input fields across the TUI
type InputRenderer struct {
	Width int
}

// NewInputRenderer creates a new input renderer
func NewInputRenderer(width int) *InputRenderer {
	return &InputRenderer{Width: width}
}

// RenderInputField renders a standardized input field with cursor
func (ir *InputRenderer) RenderInputField(
	text string,
	cursorPos int,
	placeholder string,
	showCursor bool,
	cursorVisible bool,
) string {
	// Create the background style for the input field
	inputFieldStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSelected)).
		Foreground(lipgloss.Color(ColorNormal)).
		Width(ir.Width).
		Padding(0, 1)

	// Cursor style when visible - inverted colors for visibility
	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorActive)).
		Foreground(lipgloss.Color(ColorWhite)).
		Bold(true)

	// Placeholder style
	placeholderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorVeryDim))

	var content strings.Builder

	if text == "" {
		// Empty input - show placeholder with cursor at beginning
		if showCursor && cursorVisible {
			content.WriteString(cursorStyle.Render(" "))
			if placeholder != "" {
				content.WriteString(placeholderStyle.Render(placeholder))
			}
		} else if showCursor {
			// Cursor is in blink-off state
			content.WriteString(" ")
			if placeholder != "" {
				content.WriteString(placeholderStyle.Render(placeholder))
			}
		} else {
			// No cursor (not focused)
			if placeholder != "" {
				content.WriteString(placeholderStyle.Render(placeholder))
			}
		}
	} else {
		// Input has text
		runes := []rune(text)

		// Ensure cursor position is valid
		if cursorPos < 0 {
			cursorPos = 0
		}
		if cursorPos > len(runes) {
			cursorPos = len(runes)
		}

		// Build the text with cursor
		for i := 0; i < len(runes); i++ {
			if showCursor && i == cursorPos && cursorVisible {
				// Highlight the character at cursor position
				content.WriteString(cursorStyle.Render(string(runes[i])))
			} else {
				content.WriteString(string(runes[i]))
			}
		}
		
		// Add cursor at end if needed
		if showCursor && cursorPos == len(runes) {
			if cursorVisible {
				content.WriteString(cursorStyle.Render(" "))
			}
			// Don't add space when cursor is not visible - this prevents the gap
		}
	}

	// Don't add extra padding - let the style handle width
	// The Width setting in the style will handle padding automatically
	
	return inputFieldStyle.Render(content.String())
}

// RenderInputFieldWithLabel renders an input field with a label above it
func (ir *InputRenderer) RenderInputFieldWithLabel(
	label string,
	text string,
	cursorPos int,
	placeholder string,
	showCursor bool,
	cursorVisible bool,
) string {
	var result strings.Builder

	// Render the label
	result.WriteString(HeaderStyle.Render(label))
	result.WriteString("\n")

	// Render the input field
	result.WriteString(ir.RenderInputField(text, cursorPos, placeholder, showCursor, cursorVisible))

	return result.String()
}