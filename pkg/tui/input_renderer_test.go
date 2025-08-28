package tui

import (
	"strings"
	"testing"
)

func TestInputRenderer_RenderInputField(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		cursorPos     int
		placeholder   string
		showCursor    bool
		cursorVisible bool
		checkContent  func(string) bool
	}{
		{
			name:          "empty input with placeholder and visible cursor",
			text:          "",
			cursorPos:     0,
			placeholder:   "Enter text...",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain placeholder text
				return strings.Contains(output, "Enter text...")
			},
		},
		{
			name:          "empty input with placeholder and hidden cursor",
			text:          "",
			cursorPos:     0,
			placeholder:   "Enter text...",
			showCursor:    true,
			cursorVisible: false,
			checkContent: func(output string) bool {
				// Should still contain placeholder
				return strings.Contains(output, "Enter text...")
			},
		},
		{
			name:          "text with cursor at beginning",
			text:          "Hello",
			cursorPos:     0,
			placeholder:   "",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain the text
				return strings.Contains(output, "Hello") || strings.Contains(output, "ello")
			},
		},
		{
			name:          "text with cursor in middle",
			text:          "Hello",
			cursorPos:     2,
			placeholder:   "",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain parts of the text
				return strings.Contains(output, "He") || strings.Contains(output, "llo")
			},
		},
		{
			name:          "text with cursor at end",
			text:          "Hello",
			cursorPos:     5,
			placeholder:   "",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain the text
				return strings.Contains(output, "Hello")
			},
		},
		{
			name:          "no cursor shown",
			text:          "Hello",
			cursorPos:     2,
			placeholder:   "",
			showCursor:    false,
			cursorVisible: false,
			checkContent: func(output string) bool {
				// Should contain the text
				return strings.Contains(output, "Hello") || strings.Contains(output, "llo")
			},
		},
		{
			name:          "invalid cursor position corrected",
			text:          "Hi",
			cursorPos:     10, // Beyond text length
			placeholder:   "",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain the text
				return strings.Contains(output, "Hi")
			},
		},
		{
			name:          "negative cursor position corrected",
			text:          "Hi",
			cursorPos:     -1,
			placeholder:   "",
			showCursor:    true,
			cursorVisible: true,
			checkContent: func(output string) bool {
				// Should contain the text
				return strings.Contains(output, "Hi") || strings.Contains(output, "i")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := NewInputRenderer(40)
			output := ir.RenderInputField(
				tt.text,
				tt.cursorPos,
				tt.placeholder,
				tt.showCursor,
				tt.cursorVisible,
			)

			// Check that output is not empty
			if output == "" {
				t.Error("Output should not be empty")
			}

			// Check specific content requirements
			if !tt.checkContent(output) {
				t.Errorf("Output content check failed. Got: %q", output)
			}
		})
	}
}

func TestInputRenderer_RenderInputFieldWithLabel(t *testing.T) {
	ir := NewInputRenderer(40)
	output := ir.RenderInputFieldWithLabel(
		"Username:",
		"john_doe",
		8,
		"Enter username...",
		true,
		true,
	)

	// Should contain the label
	if !strings.Contains(output, "Username:") {
		t.Error("Output should contain the label")
	}

	// Should contain the text
	if !strings.Contains(output, "john_doe") {
		t.Error("Output should contain the input text")
	}

	// Should have a newline between label and input
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Error("Output should have at least 2 lines (label and input)")
	}
}

func TestInputRenderer_Width(t *testing.T) {
	// Test that the renderer respects width setting
	ir := NewInputRenderer(20)
	output := ir.RenderInputField(
		"Short",
		5,
		"",
		false,
		false,
	)

	// The output should be styled (contains ANSI codes), but we can't easily test the exact width
	// Just ensure it's not empty
	if output == "" {
		t.Error("Output should not be empty")
	}
}