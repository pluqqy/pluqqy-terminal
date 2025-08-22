package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	ColorActive   = "170" // Purple/magenta for active elements
	ColorInactive = "240" // Gray for inactive elements
	ColorSelected = "236" // Dark gray for background selection
	ColorNormal   = "245" // Light gray for normal text
	ColorDim      = "241" // Dimmer gray
	ColorVeryDim  = "242" // Even dimmer gray
	ColorWarning  = "214" // Orange/yellow for warnings
	ColorDanger   = "196" // Red for dangerous actions
	ColorSuccess  = "28"  // Green for success
	ColorWhite    = "255" // White
	ColorDark     = "235" // Dark for contrast
	ColorBorder   = "243" // Border gray
	ColorPrimary  = "33"  // Blue for primary actions
	ColorError    = "196" // Red for errors (same as danger)
)

// Common styles
var (
	// Border styles
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorActive))

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorInactive))

	// Selection styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorActive)).
			Background(lipgloss.Color(ColorSelected)).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNormal))

	// Header styles
	TypeHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorWarning))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorDim))

	// Padding styles
	HeaderPaddingStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				PaddingRight(1)

	ContentPaddingStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				PaddingRight(1)

	// Message styles
	EmptyActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWarning)).
				Bold(true)

	EmptyInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorVeryDim))

	// Confirmation styles
	ConfirmDangerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDanger)).
				Bold(true).
				Padding(1)

	ConfirmWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWarning)).
				Bold(true).
				Padding(1)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorActive)).
			Padding(0, 1)

	// Description styles
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDim))

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError))

	// Cursor style
	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorActive)).
			Bold(true)

	// Placeholder style
	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDim)).
				Italic(true)
)

// Token badge styles based on count
func GetTokenBadgeStyle(tokenCount int) lipgloss.Style {
	status := GetTokenStatus(tokenCount)
	switch status {
	case "good":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSuccess)).
			Foreground(lipgloss.Color(ColorWhite)).
			Padding(0, 1).
			Bold(true)
	case "warning":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(ColorWarning)).
			Foreground(lipgloss.Color(ColorDark)).
			Padding(0, 1).
			Bold(true)
	case "danger":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(ColorDanger)).
			Foreground(lipgloss.Color(ColorWhite)).
			Padding(0, 1).
			Bold(true)
	default:
		return lipgloss.NewStyle().
			Padding(0, 1)
	}
}

// Get token status based on count
func GetTokenStatus(tokenCount int) string {
	if tokenCount < 10000 {
		return "good"
	} else if tokenCount < 50000 {
		return "warning"
	}
	return "danger"
}

// Dynamic styles that depend on state
func GetActiveHeaderStyle(isActive bool) lipgloss.Style {
	color := ColorInactive
	if isActive {
		color = ColorActive
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color))
}

func GetActiveColonStyle(isActive bool) lipgloss.Style {
	color := ColorInactive
	if isActive {
		color = ColorActive
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color))
}

// Help border style (always inactive looking)
var HelpBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color(ColorInactive))

// Tag chip style
func GetTagChipStyle(color string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color(ColorWhite)).
		Padding(0, 1)
}

// Mermaid diagram colors (Tokyo Night theme)
var (
	MermaidPipelineColor       = "#bb9af7"
	MermaidPipelineStroke      = "#9a7ecc"
	MermaidContextColor        = "#7aa2f7"
	MermaidContextStroke       = "#5a82d7"
	MermaidPromptColor         = "#9ece6a"
	MermaidPromptStroke        = "#7eae4a"
	MermaidRulesColor          = "#f7768e"
	MermaidRulesStroke         = "#d7566e"
	MermaidBackgroundColor     = "#1a1b26"
	MermaidStrokeWidth         = "3px"
	MermaidPipelineStrokeWidth = "4px"
)

// Mermaid constants
const (
	EstimatedTokensPerComponent = 350
	TokensPerCharacter          = 4
	MaxFunctionLines            = 50
	MermaidTmpSubdir            = "diagrams"
)
