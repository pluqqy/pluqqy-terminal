package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationType defines the visual style of the confirmation
type ConfirmationType int

const (
	ConfirmTypeInline ConfirmationType = iota // Simple inline message
	ConfirmTypeDialog                          // Full dialog with border and centered layout
)

// ConfirmationConfig holds the configuration for a confirmation prompt
type ConfirmationConfig struct {
	Title       string           // Title for dialog type (optional)
	Message     string           // Main confirmation message
	Warning     string           // Optional warning text (shown in orange)
	Details     []string         // Optional detail lines
	Destructive bool             // If true, Yes is red, No is green
	Type        ConfirmationType // Visual style
	YesLabel    string           // Custom label for Yes (default: "Yes")
	NoLabel     string           // Custom label for No (default: "No")
	Width       int              // Width for dialog type
	Height      int              // Height for dialog type
}

// ConfirmationModel handles confirmation prompts
type ConfirmationModel struct {
	active   bool
	config   ConfirmationConfig
	onConfirm func() tea.Cmd
	onCancel  func() tea.Cmd
	viewWidth int // Width for centering inline messages
}

// NewConfirmation creates a new confirmation model
func NewConfirmation() *ConfirmationModel {
	return &ConfirmationModel{}
}

// Show activates the confirmation with the given configuration
func (m *ConfirmationModel) Show(config ConfirmationConfig, onConfirm, onCancel func() tea.Cmd) {
	m.active = true
	m.config = config
	m.onConfirm = onConfirm
	m.onCancel = onCancel
	
	// Set defaults
	if m.config.YesLabel == "" {
		m.config.YesLabel = "Yes"
	}
	if m.config.NoLabel == "" {
		m.config.NoLabel = "No"
	}
}

// Hide deactivates the confirmation
func (m *ConfirmationModel) Hide() {
	m.active = false
}

// Active returns whether the confirmation is currently shown
func (m *ConfirmationModel) Active() bool {
	return m.active
}

// Update handles key events for the confirmation
func (m *ConfirmationModel) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	
	switch msg.String() {
	case "y", "Y":
		m.active = false
		if m.onConfirm != nil {
			return m.onConfirm()
		}
		return nil
		
	case "n", "N", "esc":
		m.active = false
		if m.onCancel != nil {
			return m.onCancel()
		}
		return nil
	}
	
	return nil
}

// View renders the confirmation based on its type
func (m *ConfirmationModel) View() string {
	if !m.active {
		return ""
	}
	
	switch m.config.Type {
	case ConfirmTypeInline:
		return m.renderInline()
	case ConfirmTypeDialog:
		return m.renderDialog()
	default:
		return m.renderInline()
	}
}

// ViewWithWidth renders the confirmation with a specific width for centering
func (m *ConfirmationModel) ViewWithWidth(width int) string {
	m.viewWidth = width
	return m.View()
}

// renderInline renders a simple inline confirmation message
func (m *ConfirmationModel) renderInline() string {
	options := formatConfirmOptions(m.config.Destructive)
	message := fmt.Sprintf("%s %s", m.config.Message, options)
	
	// Center the message if width is provided
	if m.viewWidth > 0 {
		messageWidth := lipgloss.Width(message)
		if messageWidth < m.viewWidth {
			centeredStyle := lipgloss.NewStyle().
				Width(m.viewWidth).
				Align(lipgloss.Center)
			return centeredStyle.Render(message)
		}
	}
	
	return message
}

// renderDialog renders a full dialog with border
func (m *ConfirmationModel) renderDialog() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))
		
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange
		
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")) // Orange for warning
	
	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Normal text
	
	// Calculate dimensions
	width := m.config.Width
	if width == 0 {
		width = 60 // Default width
	}
	height := m.config.Height
	if height == 0 {
		height = 10 // Default height
	}
	
	contentWidth := width - 4 // Account for border and padding
	
	var mainContent strings.Builder
	
	// Title
	if m.config.Title != "" {
		centeredTitle := lipgloss.NewStyle().
			Width(contentWidth - 4).
			Align(lipgloss.Center).
			Render(headerStyle.Render(m.config.Title))
		mainContent.WriteString(centeredTitle)
		mainContent.WriteString("\n\n")
	}
	
	// Message
	if m.config.Message != "" {
		message := m.config.Message
		// Center the message if it's short
		if len(message) < contentWidth-4 {
			message = lipgloss.NewStyle().
				Width(contentWidth - 4).
				Align(lipgloss.Center).
				Render(message)
		}
		mainContent.WriteString(message)
		mainContent.WriteString("\n")
	}
	
	// Warning
	if m.config.Warning != "" {
		mainContent.WriteString("\n")
		warning := warningStyle.Render(m.config.Warning)
		// Center the warning if it's short
		if lipgloss.Width(warning) < contentWidth-4 {
			warning = lipgloss.NewStyle().
				Width(contentWidth - 4).
				Align(lipgloss.Center).
				Render(warning)
		}
		mainContent.WriteString(warning)
		mainContent.WriteString("\n")
	}
	
	// Details
	if len(m.config.Details) > 0 {
		mainContent.WriteString("\n")
		for _, detail := range m.config.Details {
			mainContent.WriteString(detailStyle.Render("  • " + detail))
			mainContent.WriteString("\n")
		}
	}
	
	// Add spacing before options
	mainContent.WriteString("\n")
	
	// Options
	yesNoLabels := fmt.Sprintf("(%s / %s)", 
		strings.ToLower(m.config.YesLabel), 
		strings.ToLower(m.config.NoLabel))
	options := formatConfirmOptions(m.config.Destructive) + "  " + yesNoLabels
	centeredOptions := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(options)
	mainContent.WriteString(centeredOptions)
	
	// Apply border to main content
	return borderStyle.
		Width(width).
		Height(height).
		Render(mainContent.String())
}

// Helper function to create a quick inline confirmation
func (m *ConfirmationModel) ShowInline(message string, destructive bool, onConfirm, onCancel func() tea.Cmd) {
	m.Show(ConfirmationConfig{
		Message:     message,
		Destructive: destructive,
		Type:        ConfirmTypeInline,
	}, onConfirm, onCancel)
}

// Helper function to create a quick dialog confirmation
func (m *ConfirmationModel) ShowDialog(title, message, warning string, destructive bool, width, height int, onConfirm, onCancel func() tea.Cmd) {
	m.Show(ConfirmationConfig{
		Title:       title,
		Message:     message,
		Warning:     warning,
		Destructive: destructive,
		Type:        ConfirmTypeDialog,
		Width:       width,
		Height:      height,
	}, onConfirm, onCancel)
}