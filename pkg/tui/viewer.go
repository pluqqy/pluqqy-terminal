package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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
	// Pipeline is already loaded in SetPipeline
	return nil
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

		case "S":
			// Set pipeline (generate PLUQQY.md)
			return m, m.setPipeline()


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
		default:
			// For other keys, forward to viewports
			if m.pipeline != nil {
				// Only update the active viewport for key messages
				if m.activePane == 0 {
					m.componentsViewport, cmd = m.componentsViewport.Update(msg)
				} else {
					m.previewViewport, cmd = m.previewViewport.Update(msg)
				}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	default:
		// Forward non-key messages to both viewports
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

	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Calculate dimensions
	leftWidth := 35 // Fixed width for components
	rightWidth := m.width - leftWidth - 7 // Account for gap, padding, and border visibility
	contentHeight := m.height - 10 // Reserve space for title, help pane, and spacing
	
	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Build left column (components)
	var leftContent strings.Builder
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Create heading with colons spanning the width
	leftHeading := "COMPONENTS"
	leftRemainingWidth := leftWidth - len(leftHeading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if leftRemainingWidth < 0 {
		leftRemainingWidth = 0
	}
	// Render heading and colons separately with different styles
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Subtle gray
	leftContent.WriteString(headerPadding.Render(headerStyle.Render(leftHeading) + " " + colonStyle.Render(strings.Repeat(":", leftRemainingWidth))))
	leftContent.WriteString("\n\n")
	
	// Add padding around the viewport content
	viewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	leftContent.WriteString(viewportPadding.Render(m.componentsViewport.View()))

	// Build right column (preview)
	var rightContent strings.Builder
	
	// Calculate token count
	tokenCount := utils.EstimateTokens(m.composed)
	_, _, status := utils.GetTokenLimitStatus(tokenCount)
	
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
	
	// Create the header with colons and token info
	pipelineName := "PLUQQY.md"
	if m.pipelineName != "" {
		pipelineName = m.pipelineName + ".yaml"
	}
	headerText := fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
	tokenInfo := tokenBadge
	
	// Calculate the actual rendered width of token info
	tokenInfoWidth := lipgloss.Width(tokenBadge)
	
	// Calculate space for colons between heading and token info
	colonSpace := rightWidth - len(headerText) - tokenInfoWidth - 6 // -6 for padding and spaces (2 left + 2 right + 2 spaces)
	if colonSpace < 3 {
		colonSpace = 3
	}
	
	// Build the complete header line with right-aligned token info
	rightContent.WriteString(headerPadding.Render(headerStyle.Render(headerText) + " " + colonStyle.Render(strings.Repeat(":", colonSpace)) + " " + tokenInfo))
	rightContent.WriteString("\n\n")
	
	// Add padding around the viewport content
	previewPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	rightContent.WriteString(previewPadding.Render(m.previewViewport.View()))

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
	
	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	s.WriteString(contentStyle.Render(columns))

	// Help text in bordered pane
	help := []string{
		"tab switch pane",
		"↑/↓ scroll",
		"S set",
		"e edit",
		"esc back",
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

	return s.String()
}

func (m *PipelineViewerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateViewportSizes()
}

func (m *PipelineViewerModel) SetPipeline(pipeline string) {
	m.pipelineName = pipeline
	
	// Load pipeline synchronously to avoid "Loading..." delay
	p, err := files.ReadPipeline(m.pipelineName)
	if err != nil {
		m.err = err
		return
	}
	m.pipeline = p
	
	// Generate the pipeline output
	composed, err := composer.ComposePipeline(p)
	if err != nil {
		m.err = err
		return
	}
	m.composed = composed
	
	// Update viewport sizes now that we have content
	m.updateViewportSizes()
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
	m.componentsViewport.Width = leftWidth - 2 // Account for left and right padding
	m.componentsViewport.Height = contentHeight - 3 // Reserve space for header and spacing
	m.previewViewport.Width = rightWidth - 4 // Account for borders (2) and left/right padding (2)
	m.previewViewport.Height = contentHeight - 3 // Reserve space for header and spacing
}

func (m *PipelineViewerModel) updateViewportContent() {
	if m.pipeline == nil {
		return
	}
	
	// Build components content
	var componentsContent strings.Builder
	componentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	
	for i, comp := range m.pipeline.Components {
		line := fmt.Sprintf("%d. [%s] %s", i+1, comp.Type, filepath.Base(comp.Path))
		componentsContent.WriteString(componentStyle.Render(line))
		if i < len(m.pipeline.Components)-1 {
			componentsContent.WriteString("\n")
		}
	}
	
	// Wrap content to viewport width to prevent overflow
	wrappedComponentsContent := wordwrap.String(componentsContent.String(), m.componentsViewport.Width)
	m.componentsViewport.SetContent(wrappedComponentsContent)
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := strings.ReplaceAll(m.composed, "\r\r", "\n\n")
	processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
	wrappedPreviewContent := wordwrap.String(processedContent, m.previewViewport.Width)
	m.previewViewport.SetContent(wrappedPreviewContent)
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

