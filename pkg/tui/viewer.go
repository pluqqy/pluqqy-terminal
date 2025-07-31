package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type PipelineViewerModel struct {
	width        int
	height       int
	pipelineName string
	pipeline     *models.Pipeline
	composed     string
	err          error
	scrollY      int // For manual scrolling (deprecated)
	
	// Viewports for scrollable content
	componentsViewport viewport.Model
	previewViewport    viewport.Model
	activePane         int // 0: components, 1: preview
}

func NewPipelineViewerModel() *PipelineViewerModel {
	return &PipelineViewerModel{
		componentsViewport: viewport.New(30, 20), // Default sizes
		previewViewport:    viewport.New(80, 20),
		activePane:         1, // Start with preview active
	}
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
		
		// Update viewport sizes now that we have content
		m.updateViewportSizes()
		
		return StatusMsg(fmt.Sprintf("Loaded pipeline: %s", m.pipelineName))
	}
}

func (m *PipelineViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update viewport sizes
		m.updateViewportSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Return to main list
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "tab":
			// Switch between panes
			if m.activePane == 0 {
				m.activePane = 1
			} else {
				m.activePane = 0
			}
			return m, nil

		case "r", "R":
			// Set pipeline (generate PLUQQY.md)
			return m, m.setPipeline()

		case "E":
			// Edit in external editor
			return m, m.editInEditor()

		case "e":
			// Edit in pipeline builder (TUI)
			return m, func() tea.Msg {
				return SwitchViewMsg{
					view:     pipelineBuilderView,
					pipeline: m.pipelineName,
				}
			}

		case "up", "k", "down", "j", "pgup", "pgdown":
			// Forward to active viewport
			if m.activePane == 0 {
				m.componentsViewport, cmd = m.componentsViewport.Update(msg)
			} else {
				m.previewViewport, cmd = m.previewViewport.Update(msg)
			}
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	}

	// Forward other messages to viewports
	if m.pipeline != nil {
		m.componentsViewport, cmd = m.componentsViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.previewViewport, cmd = m.previewViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *PipelineViewerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: Failed to load pipeline: %v\n\nPress 'Esc' to return", m.err)
	}

	if m.pipeline == nil {
		return "Loading pipeline..."
	}

	// Update viewport content
	m.updateViewportContent()

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Calculate dimensions
	leftWidth := 35 // Fixed width for components
	rightWidth := m.width - leftWidth - 6 // Account for gap, padding, and ensure border visibility
	contentHeight := m.height - 8 // Reserve space for title and help

	// Build left column (components)
	var leftContent strings.Builder
	// Create heading with colons spanning the width
	leftHeading := "COMPONENTS"
	leftRemainingWidth := leftWidth - len(leftHeading) - 3
	if leftRemainingWidth < 0 {
		leftRemainingWidth = 0
	}
	// Render heading and colons separately with different styles
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Subtle gray
	leftContent.WriteString(headerStyle.Render(leftHeading) + " " + colonStyle.Render(strings.Repeat(":", leftRemainingWidth)))
	leftContent.WriteString("\n\n")
	leftContent.WriteString(m.componentsViewport.View())

	// Build right column (preview)
	var rightContent strings.Builder
	
	// Calculate token count
	tokenCount := utils.EstimateTokens(m.composed)
	percentage, limit, status := utils.GetTokenLimitStatus(tokenCount)
	
	// Create token badge with appropriate color
	var tokenBadgeStyle lipgloss.Style
	switch status {
	case "good":
		tokenBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("28")).  // Green
			Foreground(lipgloss.Color("255")). // White
			Padding(0, 1).
			Bold(true)
	case "warning":
		tokenBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("214")). // Yellow/Orange
			Foreground(lipgloss.Color("235")). // Dark
			Padding(0, 1).
			Bold(true)
	case "danger":
		tokenBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("196")). // Red
			Foreground(lipgloss.Color("255")). // White
			Padding(0, 1).
			Bold(true)
	}
	
	tokenBadge := tokenBadgeStyle.Render(utils.FormatTokenCount(tokenCount))
	limitInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf(" %d%% of %dK", percentage, limit/1024))
	
	// Create the header with colons and token info
	headerText := "PIPELINE PREVIEW (PLUQQY.md)"
	tokenInfo := tokenBadge + limitInfo
	
	// Calculate the actual rendered width of token info
	tokenInfoWidth := lipgloss.Width(tokenBadge) + lipgloss.Width(limitInfo)
	
	// Calculate space for colons between heading and token info
	colonSpace := rightWidth - len(headerText) - tokenInfoWidth - 4 // -4 for padding/spaces
	if colonSpace < 3 {
		colonSpace = 3
	}
	
	// Build the complete header line with right-aligned token info
	rightContent.WriteString(headerStyle.Render(headerText) + " " + colonStyle.Render(strings.Repeat(":", colonSpace)) + " " + tokenInfo)
	rightContent.WriteString("\n\n")
	rightContent.WriteString(m.previewViewport.View())

	// Apply borders based on active pane
	leftStyle := inactiveStyle
	rightStyle := inactiveStyle
	if m.activePane == 0 {
		leftStyle = activeStyle
	} else {
		rightStyle = activeStyle
	}

	leftColumn := leftStyle.
		Width(leftWidth).
		Height(contentHeight).
		Render(leftContent.String())

	rightColumn := rightStyle.
		Width(rightWidth).
		Height(contentHeight).
		Render(rightContent.String())

	// Join columns
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, " ", rightColumn)

	// Build final view
	var s strings.Builder
	title := fmt.Sprintf("📄 Pipeline: %s", m.pipeline.Name)
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")
	
	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	s.WriteString(contentStyle.Render(columns))

	// Help text
	help := []string{
		"Tab: switch pane",
		"↑/↓: scroll",
		"r: set",
		"E: edit external",
		"e: edit TUI",
		"Esc: back",
		"Ctrl+C: quit",
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(strings.Join(help, " • ")))

	return s.String()
}

func (m *PipelineViewerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateViewportSizes()
}

func (m *PipelineViewerModel) SetPipeline(pipeline string) {
	m.pipelineName = pipeline
}

func (m *PipelineViewerModel) updateViewportSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}
	
	// Calculate dimensions
	leftWidth := 33 // Content width for components (35 - 2 for borders)
	rightWidth := m.width - 35 - 3 - 2 // Total - left column - gap - right border
	contentHeight := m.height - 10 // Reserve space for title, headers, and help
	
	if contentHeight < 5 {
		contentHeight = 5
	}
	
	// Update viewport sizes
	m.componentsViewport.Width = leftWidth
	m.componentsViewport.Height = contentHeight
	m.previewViewport.Width = rightWidth
	m.previewViewport.Height = contentHeight
}

func (m *PipelineViewerModel) updateViewportContent() {
	if m.pipeline == nil {
		return
	}
	
	// Build components content
	var componentsContent strings.Builder
	componentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	
	for i, comp := range m.pipeline.Components {
		line := fmt.Sprintf("%d. [%s] %s", i+1, comp.Type, filepath.Base(comp.Path))
		componentsContent.WriteString(componentStyle.Render(line))
		if i < len(m.pipeline.Components)-1 {
			componentsContent.WriteString("\n")
		}
	}
	
	m.componentsViewport.SetContent(componentsContent.String())
	m.previewViewport.SetContent(m.composed)
}

func (m *PipelineViewerModel) setPipeline() tea.Cmd {
	return func() tea.Msg {
		// Generate pipeline output
		output, err := composer.ComposePipeline(m.pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to generate pipeline output: %v", err))
		}

		// Write to PLUQQY.md
		outputPath := m.pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to write output file '%s': %v", outputPath, err))
		}

		m.composed = output
		return StatusMsg(fmt.Sprintf("✓ Set pipeline → %s", outputPath))
	}
}

func (m *PipelineViewerModel) editInEditor() tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		// Validate editor path to prevent command injection
		if strings.ContainsAny(editor, "&|;<>()$`\\\"'") {
			return StatusMsg("Invalid EDITOR value: contains shell metacharacters")
		}

		pipelinePath := fmt.Sprintf(".pluqqy/pipelines/%s", m.pipelineName)
		cmd := exec.Command(editor, pipelinePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}

		// Reload pipeline after editing
		pipeline, err := files.ReadPipeline(m.pipelineName)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to reload pipeline '%s' after editing: %v", m.pipelineName, err))
		}
		m.pipeline = pipeline

		// Re-compose
		composed, _ := composer.ComposePipeline(pipeline)
		m.composed = composed

		return StatusMsg("✓ Pipeline reloaded after editing")
	}
}