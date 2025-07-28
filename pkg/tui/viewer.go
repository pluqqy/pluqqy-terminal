package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/pluqqy/pkg/composer"
	"github.com/user/pluqqy/pkg/files"
	"github.com/user/pluqqy/pkg/models"
)

type PipelineViewerModel struct {
	width        int
	height       int
	pipelineName string
	pipeline     *models.Pipeline
	composed     string
	err          error
	scrollY      int
}

func NewPipelineViewerModel() *PipelineViewerModel {
	return &PipelineViewerModel{}
}

func (m *PipelineViewerModel) Init() tea.Cmd {
	return m.loadPipeline()
}

func (m *PipelineViewerModel) loadPipeline() tea.Cmd {
	return func() tea.Msg {
		pipeline, err := files.ReadPipeline(m.pipelineName)
		if err != nil {
			m.err = err
			return nil
		}
		m.pipeline = pipeline
		
		// Generate the pipeline output
		composed, err := composer.ComposePipeline(pipeline)
		if err != nil {
			m.err = err
			return nil
		}
		m.composed = composed
		
		return StatusMsg(fmt.Sprintf("Loaded pipeline: %s", m.pipelineName))
	}
}

func (m *PipelineViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return to main list
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "r", "R":
			// Set pipeline (generate PLUQQY.md)
			return m, m.setPipeline()

		case "e":
			// Edit in external editor
			return m, m.editInEditor()

		case "E":
			// Edit in pipeline builder
			return m, func() tea.Msg {
				return SwitchViewMsg{
					view:     pipelineBuilderView,
					pipeline: m.pipelineName,
				}
			}

		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}

		case "down", "j":
			lines := strings.Split(m.composed, "\n")
			maxScroll := len(lines) - (m.height - 10)
			if maxScroll > 0 && m.scrollY < maxScroll {
				m.scrollY++
			}

		case "pgup", "ctrl+u":
			m.scrollY -= 10
			if m.scrollY < 0 {
				m.scrollY = 0
			}

		case "pgdown", "ctrl+d":
			lines := strings.Split(m.composed, "\n")
			maxScroll := len(lines) - (m.height - 10)
			m.scrollY += 10
			if m.scrollY > maxScroll && maxScroll > 0 {
				m.scrollY = maxScroll
			}
		}
	}

	return m, nil
}

func (m *PipelineViewerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to return", m.err)
	}

	if m.pipeline == nil {
		return "Loading pipeline..."
	}

	// Styles - matching the builder
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("243")).
		Padding(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	typeHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Calculate dimensions
	contentWidth := m.width - 4
	componentCount := len(m.pipeline.Components)
	contentHeight := m.height - 12 - componentCount // Leave room for title, components, and help

	// Build view
	var s strings.Builder
	
	// Title
	title := fmt.Sprintf("📄 Pipeline: %s", m.pipeline.Name)
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	// Component list
	s.WriteString(typeHeaderStyle.Render("Components:"))
	s.WriteString("\n")
	componentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	for i, comp := range m.pipeline.Components {
		s.WriteString(componentStyle.Render(fmt.Sprintf("  %d. [%s] %s", i+1, comp.Type, filepath.Base(comp.Path))))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Preview header
	s.WriteString(typeHeaderStyle.Render("Pipeline Preview (PLUQQY.md)"))
	s.WriteString("\n")

	// Content in bordered box
	if contentWidth > 0 && contentHeight > 0 {
		// Apply scrolling within the content
		lines := strings.Split(m.composed, "\n")
		visibleLines := []string{}
		
		startLine := m.scrollY
		endLine := startLine + contentHeight - 2 // -2 for padding
		if endLine > len(lines) {
			endLine = len(lines)
		}
		
		for i := startLine; i < endLine && i < len(lines); i++ {
			visibleLines = append(visibleLines, lines[i])
		}
		
		content := strings.Join(visibleLines, "\n")
		
		// Add scroll indicator if needed
		if len(lines) > contentHeight-2 {
			scrollInfo := fmt.Sprintf("\n\n[Lines %d-%d of %d]", startLine+1, endLine, len(lines))
			content += scrollInfo
		}
		
		styledBox := previewStyle.
			Width(contentWidth).
			Height(contentHeight).
			Render(content)
		
		s.WriteString(styledBox)
	}

	// Help
	help := []string{
		"↑/↓: scroll",
		"r: set",
		"e: edit ($EDITOR)",
		"E: edit (builder)",
		"q: back",
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(strings.Join(help, " • ")))

	return s.String()
}

func (m *PipelineViewerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *PipelineViewerModel) SetPipeline(pipeline string) {
	m.pipelineName = pipeline
}

func (m *PipelineViewerModel) setPipeline() tea.Cmd {
	return func() tea.Msg {
		// Generate pipeline output
		output, err := composer.ComposePipeline(m.pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error generating pipeline: %v", err))
		}

		// Write to PLUQQY.md
		outputPath := m.pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error writing output: %v", err))
		}

		m.composed = output
		return StatusMsg(fmt.Sprintf("Set pipeline → %s", outputPath))
	}
}

func (m *PipelineViewerModel) editInEditor() tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		pipelinePath := fmt.Sprintf(".pluqqy/pipelines/%s", m.pipelineName)
		cmd := exec.Command(editor, pipelinePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error opening editor: %v", err))
		}

		// Reload pipeline after editing
		pipeline, err := files.ReadPipeline(m.pipelineName)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Error reloading pipeline: %v", err))
		}
		m.pipeline = pipeline

		// Re-compose
		composed, _ := composer.ComposePipeline(pipeline)
		m.composed = composed

		return StatusMsg("Pipeline reloaded after editing")
	}
}