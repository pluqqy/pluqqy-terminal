package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

type MainListModel struct {
	pipelines          []string
	cursor             int
	width              int
	height             int
	err                error
	confirmingDelete   bool
	deleteConfirmation string
}

func NewMainListModel() *MainListModel {
	m := &MainListModel{}
	m.loadPipelines()
	return m
}

func (m *MainListModel) loadPipelines() {
	pipelines, err := files.ListPipelines()
	if err != nil {
		m.err = err
		return
	}
	m.pipelines = pipelines
}

func (m *MainListModel) Init() tea.Cmd {
	return nil
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle delete confirmation mode
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				// Confirmed deletion
				if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
					pipelineName := m.pipelines[m.cursor]
					m.confirmingDelete = false
					return m, m.deletePipeline(pipelineName)
				}
			case "n", "N", "esc":
				// Cancel deletion
				m.confirmingDelete = false
				m.deleteConfirmation = ""
			}
			return m, nil
		}
		
		// Normal mode key handling
		switch msg.String() {
		case "q":
			return m, tea.Quit
		
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		
		case "down", "j":
			if m.cursor < len(m.pipelines)-1 {
				m.cursor++
			}
		
		case "enter":
			if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
				// View the selected pipeline
				return m, func() tea.Msg {
					return SwitchViewMsg{
						view:     pipelineViewerView,
						pipeline: m.pipelines[m.cursor],
					}
				}
			}
		
		case "e":
			if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
				// Edit the selected pipeline
				return m, func() tea.Msg {
					return SwitchViewMsg{
						view:     pipelineBuilderView,
						pipeline: m.pipelines[m.cursor],
					}
				}
			}
		
		case "n":
			// Create new pipeline (switch to builder)
			return m, func() tea.Msg {
				return SwitchViewMsg{
					view: pipelineBuilderView,
				}
			}
		
		case "r":
			// Refresh pipeline list
			m.loadPipelines()
			return m, func() tea.Msg {
				return StatusMsg("Pipeline list refreshed")
			}
		
		case "S":
			// Set selected pipeline (generate PLUQQY.md)
			if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
				return m, m.setPipeline(m.pipelines[m.cursor])
			}
		
		case "d", "delete":
			// Delete pipeline with confirmation
			if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
				m.confirmingDelete = true
				m.deleteConfirmation = fmt.Sprintf("Delete pipeline '%s'? (y/n)", m.pipelines[m.cursor])
			}
		}
	}
	
	return m, nil
}

func (m *MainListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: Failed to load pipelines: %v\n\nPress 'q' to quit", m.err)
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("236")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Build the view
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("🗂  Pluqqy - Pipeline Manager"))
	s.WriteString("\n\n")
	
	// Add padding wrapper for consistency with other screens
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Build main content
	var mainContent strings.Builder

	if len(m.pipelines) == 0 {
		mainContent.WriteString(normalStyle.Render("No pipelines found. Press 'n' to create one."))
	} else {
		mainContent.WriteString("Pipelines:\n\n")
		for i, pipeline := range m.pipelines {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			
			line := fmt.Sprintf("%s%s", cursor, pipeline)
			
			if i == m.cursor {
				mainContent.WriteString(selectedStyle.Render(line))
			} else {
				mainContent.WriteString(normalStyle.Render(line))
			}
			mainContent.WriteString("\n")
		}
	}
	
	// Apply padding to main content
	s.WriteString(contentStyle.Render(mainContent.String()))

	// Show delete confirmation if active
	if m.confirmingDelete {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2)
		s.WriteString("\n")
		s.WriteString(confirmStyle.Render(m.deleteConfirmation))
	}
	
	// Help text in bordered pane
	help := []string{
		"↑/k up",
		"↓/j down",
		"enter view",
		"e edit",
		"n new",
		"d delete",
		"S set",
		"r refresh",
		"ctrl+c quit",
	}
	
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).  // Account for left/right padding (2) and borders (2)
		Padding(0, 1)  // Internal padding for help text
		
	helpContent := formatHelpText(help)
	
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	// Fit content to window
	content := s.String()
	if m.height > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > m.height-1 {
			lines = lines[:m.height-1]
			content = strings.Join(lines, "\n")
		}
	}

	return content
}

func (m *MainListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MainListModel) setPipeline(pipelineName string) tea.Cmd {
	return func() tea.Msg {
		// Load pipeline
		pipeline, err := files.ReadPipeline(pipelineName)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to load pipeline '%s': %v", pipelineName, err))
		}

		// Generate pipeline output
		output, err := composer.ComposePipeline(pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to generate pipeline output for '%s': %v", pipeline.Name, err))
		}

		// Write to PLUQQY.md
		outputPath := pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to write output file '%s': %v", outputPath, err))
		}

		return StatusMsg(fmt.Sprintf("✓ Set pipeline: %s → %s", pipelineName, outputPath))
	}
}

func (m *MainListModel) deletePipeline(pipelineName string) tea.Cmd {
	return func() tea.Msg {
		// Delete the pipeline file
		err := files.DeletePipeline(pipelineName)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete pipeline '%s': %v", pipelineName, err))
		}
		
		// Reload the pipeline list
		m.loadPipelines()
		
		// Adjust cursor if necessary
		if m.cursor >= len(m.pipelines) && m.cursor > 0 {
			m.cursor = len(m.pipelines) - 1
		}
		
		return StatusMsg(fmt.Sprintf("✓ Deleted pipeline: %s", pipelineName))
	}
}