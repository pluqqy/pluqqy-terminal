package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

type SettingsEditorModel struct {
	SettingsDataStore
	SettingsUIComponents
	SettingsViewportManager
	SettingsFormInputs
	SettingsSectionManager
}

const (
	fieldDefaultFilename = iota
	fieldExportPath
	fieldOutputPath
	fieldShowHeadings
	fieldSections // This is where sections list starts
)

func NewSettingsEditorModel() *SettingsEditorModel {
	m := &SettingsEditorModel{
		SettingsDataStore: SettingsDataStore{},
		SettingsUIComponents: SettingsUIComponents{
			viewport:    viewport.New(80, 20), // Default size
			exitConfirm: NewConfirmation(),
		},
		SettingsViewportManager: SettingsViewportManager{},
		SettingsFormInputs: SettingsFormInputs{
			defaultFilenameInput: textinput.New(),
			exportPathInput:      textinput.New(),
			outputPathInput:      textinput.New(),
			sectionTypeInput:     textinput.New(),
			sectionHeadingInput:  textinput.New(),
			showHeadings:         true,
			focusIndex:           0,
		},
		SettingsSectionManager: SettingsSectionManager{},
	}

	// Configure text inputs
	m.defaultFilenameInput.Placeholder = "PLUQQY.md"
	m.defaultFilenameInput.CharLimit = 255
	m.defaultFilenameInput.Width = 40

	m.exportPathInput.Placeholder = "./"
	m.exportPathInput.CharLimit = 255
	m.exportPathInput.Width = 40

	m.outputPathInput.Placeholder = "tmp/"
	m.outputPathInput.CharLimit = 255
	m.outputPathInput.Width = 40

	m.sectionTypeInput.Placeholder = "contexts/prompts/rules"
	m.sectionTypeInput.CharLimit = 50
	m.sectionTypeInput.Width = 30

	m.sectionHeadingInput.Placeholder = "## HEADING"
	m.sectionHeadingInput.CharLimit = 100
	m.sectionHeadingInput.Width = 40

	// Set initial focus
	m.updateFocus()

	return m
}

func (m *SettingsEditorModel) Init() tea.Cmd {
	// Load current settings
	return m.loadSettings()
}

func (m *SettingsEditorModel) loadSettings() tea.Cmd {
	return func() tea.Msg {
		settings, err := files.ReadSettings()
		if err != nil {
			// Use defaults if can't read
			settings = models.DefaultSettings()
		}

		// Make a deep copy for comparison
		originalSettings := &models.Settings{
			Output: models.OutputSettings{
				DefaultFilename: settings.Output.DefaultFilename,
				ExportPath:      settings.Output.ExportPath,
				OutputPath:      settings.Output.OutputPath,
				Formatting: models.FormattingSettings{
					ShowHeadings: settings.Output.Formatting.ShowHeadings,
					Sections:     make([]models.Section, len(settings.Output.Formatting.Sections)),
				},
			},
		}

		// Deep copy sections
		for i, section := range settings.Output.Formatting.Sections {
			originalSettings.Output.Formatting.Sections[i] = models.Section{
				Type:    section.Type,
				Heading: section.Heading,
			}
		}

		return settingsLoadedMsg{settings: settings, originalSettings: originalSettings}
	}
}

type settingsLoadedMsg struct {
	settings         *models.Settings
	originalSettings *models.Settings
}

func (m *SettingsEditorModel) updateFocus() {
	// Calculate total fields
	if m.settings != nil {
		m.totalFields = 4 + len(m.settings.Output.Formatting.Sections) // 4 basic fields + sections
	}

	// Disable all inputs first
	m.defaultFilenameInput.Blur()
	m.exportPathInput.Blur()
	m.outputPathInput.Blur()
	m.sectionTypeInput.Blur()
	m.sectionHeadingInput.Blur()

	// Focus the current field
	switch m.focusIndex {
	case fieldDefaultFilename:
		m.defaultFilenameInput.Focus()
	case fieldExportPath:
		m.exportPathInput.Focus()
	case fieldOutputPath:
		m.outputPathInput.Focus()
	}
}

func (m *SettingsEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSize()

	case settingsLoadedMsg:
		m.settings = msg.settings
		m.originalSettings = msg.originalSettings

		// Set input values
		m.defaultFilenameInput.SetValue(m.settings.Output.DefaultFilename)
		m.exportPathInput.SetValue(m.settings.Output.ExportPath)
		m.outputPathInput.SetValue(m.settings.Output.OutputPath)
		m.showHeadings = m.settings.Output.Formatting.ShowHeadings

		m.updateFocus()
		m.updateViewportContent()

	case tea.KeyMsg:
		if m.editingSection {
			// Handle section editing
			switch msg.String() {
			case "esc":
				m.editingSection = false
				m.sectionTypeInput.SetValue("")
				m.sectionHeadingInput.SetValue("")
				m.updateViewportContent()
				return m, nil

			case "enter":
				// Save section changes
				if m.sectionCursor >= 0 && m.sectionCursor < len(m.settings.Output.Formatting.Sections) {
					newType := m.sectionTypeInput.Value()
					newHeading := m.sectionHeadingInput.Value()

					if newType != "" {
						m.settings.Output.Formatting.Sections[m.sectionCursor].Type = newType
					}
					if newHeading != "" {
						m.settings.Output.Formatting.Sections[m.sectionCursor].Heading = newHeading
					}

					m.hasChanges = true
				}

				m.editingSection = false
				m.sectionTypeInput.SetValue("")
				m.sectionHeadingInput.SetValue("")
				m.updateViewportContent()
				return m, nil

			case "tab":
				// Switch between type and heading inputs
				if m.sectionTypeInput.Focused() {
					m.sectionTypeInput.Blur()
					m.sectionHeadingInput.Focus()
				} else {
					m.sectionHeadingInput.Blur()
					m.sectionTypeInput.Focus()
				}
				m.updateViewportContent()
				return m, nil
			}

			// Update the focused input
			if m.sectionTypeInput.Focused() {
				m.sectionTypeInput, cmd = m.sectionTypeInput.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.sectionHeadingInput, cmd = m.sectionHeadingInput.Update(msg)
				cmds = append(cmds, cmd)
			}

			m.updateViewportContent()
			return m, tea.Batch(cmds...)
		}

		// Handle confirmation dialogs
		if m.exitConfirm.Active() {
			return m, m.exitConfirm.Update(msg)
		}

		// Normal navigation
		switch msg.String() {
		case "esc":
			// Check if there are unsaved changes
			if m.hasChanges {
				m.exitConfirm.ShowDialog(
					"EXIT CONFIRMATION",
					"You have unsaved changes in settings.",
					"Are you sure you want to exit?",
					true, // destructive
					m.width-4,
					10,
					func() tea.Cmd {
						// Exit without saving - reload settings to reset everything
						return tea.Batch(
							m.loadSettings(), // Reload settings from disk
							func() tea.Msg {
								m.hasChanges = false // Reset the dirty flag
								return SwitchViewMsg{view: mainListView}
							},
						)
					},
					func() tea.Cmd {
						m.updateViewportContent()
						return nil
					},
				)
				return m, nil
			}
			// No changes, exit immediately
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "ctrl+s":
			// Save settings
			return m, m.saveSettings()

		case "up":
			if m.focusIndex > 0 {
				m.focusIndex--
				if m.focusIndex >= fieldSections {
					m.sectionCursor = m.focusIndex - fieldSections
				}
				m.updateFocus()
				m.updateViewportContent()
			}

		case "down":
			if m.focusIndex < m.totalFields-1 {
				m.focusIndex++
				if m.focusIndex >= fieldSections {
					m.sectionCursor = m.focusIndex - fieldSections
				}
				m.updateFocus()
				m.updateViewportContent()
			}

		case "tab":
			// Move to next field, wrap to beginning at end
			m.focusIndex++
			if m.focusIndex >= m.totalFields {
				m.focusIndex = 0
				m.sectionCursor = 0
			} else if m.focusIndex >= fieldSections {
				m.sectionCursor = m.focusIndex - fieldSections
			}
			m.updateFocus()
			m.updateViewportContent()

		case "shift+tab":
			// Move to previous field, wrap to end at beginning
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = m.totalFields - 1
				m.sectionCursor = m.focusIndex - fieldSections
			} else if m.focusIndex >= fieldSections {
				m.sectionCursor = m.focusIndex - fieldSections
			} else {
				m.sectionCursor = 0
			}
			m.updateFocus()
			m.updateViewportContent()

		case "J":
			// Move section down (J moves item down in list)
			if m.focusIndex >= fieldSections && m.sectionCursor < len(m.settings.Output.Formatting.Sections)-1 {
				sections := m.settings.Output.Formatting.Sections
				sections[m.sectionCursor], sections[m.sectionCursor+1] =
					sections[m.sectionCursor+1], sections[m.sectionCursor]
				m.sectionCursor++
				m.focusIndex++
				m.hasChanges = true
				m.updateViewportContent()
			}

		case "K":
			// Move section up (K moves item up in list)
			if m.focusIndex >= fieldSections && m.sectionCursor > 0 {
				sections := m.settings.Output.Formatting.Sections
				sections[m.sectionCursor], sections[m.sectionCursor-1] =
					sections[m.sectionCursor-1], sections[m.sectionCursor]
				m.sectionCursor--
				m.focusIndex--
				m.hasChanges = true
				m.updateViewportContent()
			}

		case " ", "space":
			// Toggle checkbox
			if m.focusIndex == fieldShowHeadings {
				m.showHeadings = !m.showHeadings
				m.settings.Output.Formatting.ShowHeadings = m.showHeadings
				m.hasChanges = true
				m.updateViewportContent()
			}

		case "enter":
			if m.focusIndex >= fieldSections {
				// Edit section
				section := m.settings.Output.Formatting.Sections[m.sectionCursor]
				m.sectionTypeInput.SetValue(section.Type)
				m.sectionHeadingInput.SetValue(section.Heading)
				m.sectionTypeInput.Focus()
				m.editingSection = true
				m.updateViewportContent()
			}

		case "pgup", "pgdown":
			// Forward to viewport for scrolling
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update text inputs if they're focused
	if m.defaultFilenameInput.Focused() {
		prevValue := m.defaultFilenameInput.Value()
		m.defaultFilenameInput, cmd = m.defaultFilenameInput.Update(msg)
		if m.defaultFilenameInput.Value() != prevValue {
			m.settings.Output.DefaultFilename = m.defaultFilenameInput.Value()
			m.hasChanges = true
			m.updateViewportContent()
		}
		cmds = append(cmds, cmd)
	}

	if m.exportPathInput.Focused() {
		prevValue := m.exportPathInput.Value()
		m.exportPathInput, cmd = m.exportPathInput.Update(msg)
		if m.exportPathInput.Value() != prevValue {
			m.settings.Output.ExportPath = m.exportPathInput.Value()
			m.hasChanges = true
			m.updateViewportContent()
		}
		cmds = append(cmds, cmd)
	}

	if m.outputPathInput.Focused() {
		prevValue := m.outputPathInput.Value()
		m.outputPathInput, cmd = m.outputPathInput.Update(msg)
		if m.outputPathInput.Value() != prevValue {
			m.settings.Output.OutputPath = m.outputPathInput.Value()
			m.hasChanges = true
			m.updateViewportContent()
		}
		cmds = append(cmds, cmd)
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *SettingsEditorModel) View() string {
	if m.settings == nil {
		return "Loading settings..."
	}

	// Styles matching other views
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")). // Active pane color
		Width(m.width - 4).                      // Account for margins
		Height(m.height - 5)                     // Account for help box and spacing

	// Build main content container
	var content strings.Builder

	// Add pane heading similar to other views
	heading := "EDIT SETTINGS"
	remainingWidth := m.width - 4 - len(heading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active pane
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active pane

	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	content.WriteString(headerPadding.Render(headerStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	content.WriteString("\n\n")

	// Update viewport content
	m.updateViewportContent()

	// Add viewport with padding
	viewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	content.WriteString(viewportPadding.Render(m.viewport.View()))

	// Main content in bordered pane with margins
	var s strings.Builder

	// Show exit confirmation dialog if active
	if m.exitConfirm.Active() {
		// Add padding to match other views
		paddingStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		return paddingStyle.Render(m.exitConfirm.View())
	}

	s.WriteString(contentStyle.Render(borderStyle.Render(content.String())))

	// Help text in bordered pane
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width-4). // Account for margins
		Padding(0, 1)

	// Group help items logically
	help := []string{
		"tab/shift+tab navigate",
		"↑↓ navigate",
		"J/K move section",
		"space toggle",
		"enter edit",
		"^s save",
		"esc cancel",
		"^c quit",
	}

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *SettingsEditorModel) updateViewportContent() {
	// Styles
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	labelStyle := lipgloss.NewStyle().
		Width(20).
		Foreground(lipgloss.Color("245"))

	commentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("238"))

	var content strings.Builder

	// Output settings section
	content.WriteString(sectionStyle.Render("OUTPUT SETTINGS"))
	content.WriteString("\n\n")

	// Default filename field
	label := labelStyle.Render("Default Filename:")
	fieldLine := label + " " + m.defaultFilenameInput.View()
	if m.focusIndex == fieldDefaultFilename {
		content.WriteString(focusedStyle.Render("▸ " + fieldLine))
	} else {
		content.WriteString(normalStyle.Render("  " + fieldLine))
	}
	content.WriteString("\n\n")
	content.WriteString(commentStyle.Render("  # Filename for the generated pipeline output (when you press 'S' to set a pipeline)"))
	content.WriteString("\n\n")

	// Export path field
	label = labelStyle.Render("Export Path:")
	fieldLine = label + " " + m.exportPathInput.View()
	if m.focusIndex == fieldExportPath {
		content.WriteString(focusedStyle.Render("▸ " + fieldLine))
	} else {
		content.WriteString(normalStyle.Render("  " + fieldLine))
	}
	content.WriteString("\n\n")
	content.WriteString(commentStyle.Render("  # Directory where the pipeline output file will be written"))
	content.WriteString("\n\n")

	// Output path field
	label = labelStyle.Render("Output Path:")
	fieldLine = label + " " + m.outputPathInput.View()
	if m.focusIndex == fieldOutputPath {
		content.WriteString(focusedStyle.Render("▸ " + fieldLine))
	} else {
		content.WriteString(normalStyle.Render("  " + fieldLine))
	}
	content.WriteString("\n\n")
	content.WriteString(commentStyle.Render("  # Directory for pipeline-generated files (automatically added to .gitignore)"))
	content.WriteString("\n\n")

	// Formatting section
	content.WriteString(sectionStyle.Render("FORMATTING"))
	content.WriteString("\n\n")

	// Show headings checkbox
	checkbox := "[ ]"
	if m.showHeadings {
		checkbox = "[✓]"
	}
	label = labelStyle.Render("Show Headings:")
	fieldLine = label + " " + checkbox
	if m.focusIndex == fieldShowHeadings {
		content.WriteString(focusedStyle.Render("▸ " + fieldLine))
	} else {
		content.WriteString(normalStyle.Render("  " + fieldLine))
	}
	content.WriteString("\n\n")
	content.WriteString(commentStyle.Render("  # Whether to include section headers in the output"))
	content.WriteString("\n\n")

	// Sections
	content.WriteString(sectionStyle.Render("SECTIONS"))
	content.WriteString("\n\n")

	// Show section editing form if active
	if m.editingSection && !m.exitConfirm.Active() {
		editStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1).
			MarginLeft(2).
			MarginRight(2)

		editContent := fmt.Sprintf("EDITING SECTION:\n\nType:    %s\nHeading: %s\n\n%s",
			m.sectionTypeInput.View(),
			m.sectionHeadingInput.View(),
			commentStyle.Render("Tab to switch fields • Enter to save • Esc to cancel"))

		content.WriteString(editStyle.Render(editContent))
		content.WriteString("\n\n")
	}

	// List sections
	for i, section := range m.settings.Output.Formatting.Sections {
		line := fmt.Sprintf("%d. %-10s → %s", i+1, section.Type, section.Heading)

		if m.focusIndex == fieldSections+i {
			content.WriteString(selectedStyle.Render(focusedStyle.Render("▸ " + line)))
		} else {
			content.WriteString(normalStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}
	content.WriteString("\n")
	content.WriteString(commentStyle.Render("  # Define sections in the order they should appear in the output"))
	content.WriteString("\n")

	// Set viewport content
	m.viewport.SetContent(content.String())
}

func (m *SettingsEditorModel) updateViewportSize() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Account for borders, padding, and margins
	m.viewport.Width = m.width - 10   // Borders (2) + padding (4) + margins (4)
	m.viewport.Height = m.height - 10 // Header (3) + borders (2) + help box (5)
}

func (m *SettingsEditorModel) saveSettings() tea.Cmd {
	return func() tea.Msg {
		// Validate that all three section types are present
		hasContexts, hasPrompts, hasRules := false, false, false
		for _, section := range m.settings.Output.Formatting.Sections {
			switch section.Type {
			case "contexts":
				hasContexts = true
			case "prompts":
				hasPrompts = true
			case "rules":
				hasRules = true
			}
		}

		if !hasContexts || !hasPrompts || !hasRules {
			// Add missing sections with defaults
			if !hasContexts {
				m.settings.Output.Formatting.Sections = append(m.settings.Output.Formatting.Sections, models.Section{
					Type:    "contexts",
					Heading: "## CONTEXTS",
				})
			}
			if !hasPrompts {
				m.settings.Output.Formatting.Sections = append(m.settings.Output.Formatting.Sections, models.Section{
					Type:    "prompts",
					Heading: "## PROMPTS",
				})
			}
			if !hasRules {
				m.settings.Output.Formatting.Sections = append(m.settings.Output.Formatting.Sections, models.Section{
					Type:    "rules",
					Heading: "## RULES",
				})
			}
		}

		err := files.WriteSettings(m.settings)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save settings: %v", err))
		}

		// Update original settings to match saved settings (fixes unsaved changes detection)
		m.originalSettings = &models.Settings{
			Output: models.OutputSettings{
				DefaultFilename: m.settings.Output.DefaultFilename,
				ExportPath:      m.settings.Output.ExportPath,
				OutputPath:      m.settings.Output.OutputPath,
				Formatting: models.FormattingSettings{
					ShowHeadings: m.settings.Output.Formatting.ShowHeadings,
					Sections:     make([]models.Section, len(m.settings.Output.Formatting.Sections)),
				},
			},
		}
		// Deep copy sections
		for i, section := range m.settings.Output.Formatting.Sections {
			m.originalSettings.Output.Formatting.Sections[i] = models.Section{
				Type:    section.Type,
				Heading: section.Heading,
			}
		}

		// Reset hasChanges flag after successful save
		m.hasChanges = false

		// Return status message without switching views
		return StatusMsg("✓ Settings saved")
	}
}

// Message type to indicate settings were saved and should be reloaded
type settingsSavedMsg struct {
	switchMsg SwitchViewMsg
}

func (m *SettingsEditorModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateViewportSize()
}
