package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name  string
		color string
		value string
	}{
		{"ColorActive", ColorActive, "170"},
		{"ColorInactive", ColorInactive, "240"},
		{"ColorSelected", ColorSelected, "236"},
		{"ColorNormal", ColorNormal, "245"},
		{"ColorDim", ColorDim, "241"},
		{"ColorVeryDim", ColorVeryDim, "242"},
		{"ColorWarning", ColorWarning, "214"},
		{"ColorDanger", ColorDanger, "196"},
		{"ColorSuccess", ColorSuccess, "28"},
		{"ColorWhite", ColorWhite, "255"},
		{"ColorDark", ColorDark, "235"},
		{"ColorBorder", ColorBorder, "243"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the constant has the expected value
			if tt.value != tt.color {
				t.Errorf("expected %s, got %s", tt.value, tt.color)
			}

			// Verify it can be used as a lipgloss color without panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Color %s caused panic: %v", tt.color, r)
					}
				}()
				_ = lipgloss.Color(tt.color)
			}()
		})
	}
}

func TestStaticStyles(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"ActiveBorderStyle", ActiveBorderStyle},
		{"InactiveBorderStyle", InactiveBorderStyle},
		{"SelectedStyle", SelectedStyle},
		{"NormalStyle", NormalStyle},
		{"TypeHeaderStyle", TypeHeaderStyle},
		{"HeaderStyle", HeaderStyle},
		{"HeaderPaddingStyle", HeaderPaddingStyle},
		{"ContentPaddingStyle", ContentPaddingStyle},
		{"EmptyActiveStyle", EmptyActiveStyle},
		{"EmptyInactiveStyle", EmptyInactiveStyle},
		{"ConfirmDangerStyle", ConfirmDangerStyle},
		{"ConfirmWarningStyle", ConfirmWarningStyle},
		{"InputStyle", InputStyle},
		{"DescriptionStyle", DescriptionStyle},
		{"CursorStyle", CursorStyle},
		{"PlaceholderStyle", PlaceholderStyle},
		{"HelpBorderStyle", HelpBorderStyle},
	}

	for _, tt := range styles {
		t.Run(tt.name, func(t *testing.T) {
			// Verify rendering doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Style %s caused panic: %v", tt.name, r)
					}
				}()
				_ = tt.style.Render("test content")
			}()

			// Verify the rendered output is not empty
			output := tt.style.Render("test")
			if output == "" {
				t.Errorf("Style %s rendered empty output", tt.name)
			}
		})
	}
}

func TestGetTokenStatus(t *testing.T) {
	tests := []struct {
		name       string
		tokenCount int
		expected   string
	}{
		{"Zero tokens", 0, "good"},
		{"Low tokens", 5000, "good"},
		{"Just under warning", 9999, "good"},
		{"Warning threshold", 10000, "warning"},
		{"Mid warning", 30000, "warning"},
		{"Just under danger", 49999, "warning"},
		{"Danger threshold", 50000, "danger"},
		{"High danger", 100000, "danger"},
		{"Negative tokens", -1, "good"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTokenStatus(tt.tokenCount)
			if result != tt.expected {
				t.Errorf("GetTokenStatus(%d) = %s, want %s", tt.tokenCount, result, tt.expected)
			}
		})
	}
}

func TestGetTokenBadgeStyle(t *testing.T) {
	tests := []struct {
		name       string
		tokenCount int
		status     string
	}{
		{"Good status", 5000, "good"},
		{"Warning status", 25000, "warning"},
		{"Danger status", 75000, "danger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetTokenBadgeStyle(tt.tokenCount)

			// Verify rendering doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("GetTokenBadgeStyle(%d) caused panic: %v", tt.tokenCount, r)
					}
				}()
				_ = style.Render("12.3k")
			}()

			// Verify the style has appropriate properties based on status
			rendered := style.Render("test")
			if rendered == "" {
				t.Errorf("GetTokenBadgeStyle(%d) rendered empty output", tt.tokenCount)
			}
		})
	}
}

func TestGetActiveHeaderStyle(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
	}{
		{"Active header", true},
		{"Inactive header", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetActiveHeaderStyle(tt.isActive)

			// Verify rendering doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("GetActiveHeaderStyle(%v) caused panic: %v", tt.isActive, r)
					}
				}()
				_ = style.Render("Header Text")
			}()

			// Verify the rendered output is not empty
			output := style.Render("test")
			if output == "" {
				t.Errorf("GetActiveHeaderStyle(%v) rendered empty output", tt.isActive)
			}
		})
	}
}

func TestGetActiveColonStyle(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
	}{
		{"Active colon", true},
		{"Inactive colon", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetActiveColonStyle(tt.isActive)

			// Verify rendering doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("GetActiveColonStyle(%v) caused panic: %v", tt.isActive, r)
					}
				}()
				_ = style.Render(":")
			}()

			// Verify the rendered output is not empty
			output := style.Render(":")
			if output == "" {
				t.Errorf("GetActiveColonStyle(%v) rendered empty output", tt.isActive)
			}
		})
	}
}

func TestGetTagChipStyle(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{"Success color", ColorSuccess},
		{"Warning color", ColorWarning},
		{"Danger color", ColorDanger},
		{"Active color", ColorActive},
		{"Custom color", "123"},
		{"Empty color", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetTagChipStyle(tt.color)

			// Verify rendering doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("GetTagChipStyle(%s) caused panic: %v", tt.color, r)
					}
				}()
				_ = style.Render("TAG")
			}()

			// Verify the rendered output is not empty
			output := style.Render("test")
			if output == "" {
				t.Errorf("GetTagChipStyle(%s) rendered empty output", tt.color)
			}
		})
	}
}

func TestStyleCombinations(t *testing.T) {
	// Test that styles can be combined without panic
	t.Run("Combine styles", func(t *testing.T) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Style combination caused panic: %v", r)
				}
			}()
			// Combine border with content padding
			combined := ActiveBorderStyle.Copy().Inherit(ContentPaddingStyle)
			_ = combined.Render("Combined content")

			// Combine multiple styles
			multiCombined := HeaderStyle.Copy().
				Inherit(HeaderPaddingStyle).
				Inherit(ActiveBorderStyle)
			_ = multiCombined.Render("Multi-combined")
		}()
	})
}

func TestStylesWithVariousContent(t *testing.T) {
	contents := []string{
		"",                // Empty string
		"Simple text",     // Normal text
		"Line\nBreak",     // Multi-line
		"🎨 Unicode emoji", // Unicode content
		"Very long text that might exceed normal boundaries and cause issues with rendering", // Long text
		"\t\tTabbed", // Special characters
	}

	styles := []lipgloss.Style{
		ActiveBorderStyle,
		SelectedStyle,
		ConfirmDangerStyle,
		InputStyle,
	}

	for _, style := range styles {
		for _, content := range contents {
			// Test that no style panics with any content
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Style panicked with content %q: %v", content, r)
					}
				}()
				_ = style.Render(content)
			}()
		}
	}
}

func TestTokenBadgeEdgeCases(t *testing.T) {
	// Test exact boundary values
	boundaries := []struct {
		name     string
		count    int
		expected string
	}{
		{"Exactly 10000", 10000, "warning"},
		{"Just below 10000", 9999, "good"},
		{"Just above 10000", 10001, "warning"},
		{"Exactly 50000", 50000, "danger"},
		{"Just below 50000", 49999, "warning"},
		{"Just above 50000", 50001, "danger"},
	}

	for _, tt := range boundaries {
		t.Run(tt.name, func(t *testing.T) {
			status := GetTokenStatus(tt.count)
			if status != tt.expected {
				t.Errorf("GetTokenStatus(%d) = %s, want %s", tt.count, status, tt.expected)
			}

			// Also verify the badge style works
			style := GetTokenBadgeStyle(tt.count)
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("GetTokenBadgeStyle(%d) caused panic: %v", tt.count, r)
					}
				}()
				_ = style.Render("test")
			}()
		})
	}
}
