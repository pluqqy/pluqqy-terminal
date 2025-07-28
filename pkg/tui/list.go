package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/pluqqy/pkg/composer"
	"github.com/user/pluqqy/pkg/files"
)

type MainListModel struct {
	pipelines    []string
	cursor       int
	width        int
	height       int
	err          error
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
		switch msg.String() {
		case "q", "ctrl+c":
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
		
		case "e", "E":
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
		
		case "R":
			// Set selected pipeline (generate PLUQQY.md)
			if len(m.pipelines) > 0 && m.cursor < len(m.pipelines) {
				return m, m.setPipeline(m.pipelines[m.cursor])
			}
		}
	}
	
	return m, nil
}

func (m *MainListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading pipelines: %v\n\nPress 'q' to quit", m.err)
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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Build the view
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("🗂  Pluqqy - Pipeline Manager"))
	s.WriteString("\n\n")

	if len(m.pipelines) == 0 {
		s.WriteString(normalStyle.Render("No pipelines found. Press 'n' to create one."))
	} else {
		s.WriteString("Pipelines:\n\n")
		for i, pipeline := range m.pipelines {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}
			
			line := fmt.Sprintf("%s%s", cursor, pipeline)
			
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(line))
			} else {
				s.WriteString(normalStyle.Render(line))
			}
			s.WriteString("\n")
		}
	}

	// Help text
	help := []string{
		"↑/k: up",
		"↓/j: down",
		"enter: view",
		"e: edit",
		"n: new",
		"R: set",
		"r: refresh",
		"q: quit",
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(strings.Join(help, " • ")))

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
			return StatusMsg(fmt.Sprintf("Error loading pipeline: %v", err))
		}

		// Generate pipeline output
		output, err := composer.ComposePipeline(pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error generating pipeline: %v", err))
		}

		// Write to PLUQQY.md
		outputPath := pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error writing output: %v", err))
		}

		return StatusMsg(fmt.Sprintf("Set pipeline: %s → %s", pipelineName, outputPath))
	}
}