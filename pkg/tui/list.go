package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type pane int

const (
	pipelinesPane pane = iota
	componentsPane
	previewPane
)

type MainListModel struct {
	// Pipelines data
	pipelines          []string
	pipelineCursor     int
	
	// Components data
	prompts            []componentItem
	contexts           []componentItem
	rules              []componentItem
	componentCursor    int
	
	// Preview data
	previewContent     string
	previewViewport    viewport.Model
	
	// UI state
	activePane         pane
	pipelinesViewport  viewport.Model
	componentsViewport viewport.Model
	showPreview        bool
	
	// Window dimensions
	width              int
	height             int
	
	// Error handling
	err                error
	
	// Delete confirmation
	confirmingDelete   bool
	deleteConfirmation string
	
	// Component creation state
	creatingComponent     bool
	componentCreationType string
	componentName         string
	componentContent      string
	creationStep          int // 0: type, 1: name, 2: content
	typeCursor           int
	
	// Component editing state
	editingComponent      bool
	editingComponentPath  string
	editingComponentName  string
	originalContent       string
	editViewport          viewport.Model
	
	// Exit confirmation state
	showExitConfirmation  bool
	exitConfirmationType  string // "component" or "component-edit"
}

func NewMainListModel() *MainListModel {
	// Load settings for UI preferences
	settings, _ := files.ReadSettings()
	
	m := &MainListModel{
		activePane:         componentsPane,
		showPreview:        settings.UI.ShowPreview,
		previewViewport:    viewport.New(80, 20), // Default size
		pipelinesViewport:  viewport.New(40, 20), // Default size
		componentsViewport: viewport.New(40, 20), // Default size
	}
	m.loadPipelines()
	m.loadComponents()
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

func (m *MainListModel) loadComponents() {
	// Get usage counts for all components
	usageMap, _ := files.CountComponentUsage()
	
	// Clear existing components
	m.prompts = nil
	m.contexts = nil
	m.rules = nil
	
	// Load prompts
	prompts, _ := files.ListComponents("prompts")
	for _, p := range prompts {
		componentPath := filepath.Join(files.ComponentsDir, files.PromptsDir, p)
		modTime, _ := files.GetComponentStats(componentPath)
		
		// Calculate usage count
		usage := 0
		relativePath := "../" + componentPath
		if count, exists := usageMap[relativePath]; exists {
			usage = count
		}
		
		// Read component content for token estimation
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
		}
		
		m.prompts = append(m.prompts, componentItem{
			name:         p,
			path:         componentPath,
			compType:     models.ComponentTypePrompt,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
		})
	}
	
	// Load contexts
	contexts, _ := files.ListComponents("contexts")
	for _, c := range contexts {
		componentPath := filepath.Join(files.ComponentsDir, files.ContextsDir, c)
		modTime, _ := files.GetComponentStats(componentPath)
		
		// Calculate usage count
		usage := 0
		relativePath := "../" + componentPath
		if count, exists := usageMap[relativePath]; exists {
			usage = count
		}
		
		// Read component content for token estimation
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
		}
		
		m.contexts = append(m.contexts, componentItem{
			name:         c,
			path:         componentPath,
			compType:     models.ComponentTypeContext,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
		})
	}
	
	// Load rules
	rules, _ := files.ListComponents("rules")
	for _, r := range rules {
		componentPath := filepath.Join(files.ComponentsDir, files.RulesDir, r)
		modTime, _ := files.GetComponentStats(componentPath)
		
		// Calculate usage count
		usage := 0
		relativePath := "../" + componentPath
		if count, exists := usageMap[relativePath]; exists {
			usage = count
		}
		
		// Read component content for token estimation
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
		}
		
		m.rules = append(m.rules, componentItem{
			name:         r,
			path:         componentPath,
			compType:     models.ComponentTypeRules,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
		})
	}
}

func (m *MainListModel) getAllComponents() []componentItem {
	var all []componentItem
	all = append(all, m.contexts...)
	all = append(all, m.prompts...)
	all = append(all, m.rules...)
	return all
}

func (m *MainListModel) Init() tea.Cmd {
	return nil
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle exit confirmation dialog
		if m.showExitConfirmation {
			switch msg.String() {
			case "y", "Y":
				m.showExitConfirmation = false
				if m.exitConfirmationType == "component" {
					// For component creation, just go back to type selection
					m.creationStep = 0
					m.componentContent = ""
				} else if m.exitConfirmationType == "component-edit" {
					// For component editing, exit without saving
					m.editingComponent = false
					m.componentContent = ""
					m.editingComponentPath = ""
					m.editingComponentName = ""
					m.originalContent = ""
				}
				return m, nil
			case "n", "N", "esc":
				m.showExitConfirmation = false
				return m, nil
			}
			return m, nil
		}
		
		// Handle component creation mode
		if m.creatingComponent {
			return m.handleComponentCreation(msg)
		}
		
		// Handle component editing mode
		if m.editingComponent {
			return m.handleComponentEditing(msg)
		}
		
		// Handle delete confirmation mode
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				// Confirmed deletion
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					pipelineName := m.pipelines[m.pipelineCursor]
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
		
		case "tab":
			// Switch between panes
			if m.showPreview {
				// When preview is shown, cycle through all three panes
				switch m.activePane {
				case componentsPane:
					m.activePane = pipelinesPane
				case pipelinesPane:
					m.activePane = previewPane
				case previewPane:
					m.activePane = componentsPane
				}
			} else {
				// When preview is hidden, only toggle between pipelines and components
				if m.activePane == componentsPane {
					m.activePane = pipelinesPane
				} else {
					m.activePane = componentsPane
				}
			}
			// Update preview when switching to non-preview pane
			if m.activePane != previewPane {
				m.updatePreview()
			}
		
		case "up", "k":
			if m.activePane == pipelinesPane {
				if m.pipelineCursor > 0 {
					m.pipelineCursor--
					m.updatePreview()
				}
			} else if m.activePane == componentsPane {
				if m.componentCursor > 0 {
					m.componentCursor--
					m.updatePreview()
				}
			} else if m.activePane == previewPane {
				// Scroll preview up
				m.previewViewport.LineUp(1)
			}
		
		case "down", "j":
			if m.activePane == pipelinesPane {
				if m.pipelineCursor < len(m.pipelines)-1 {
					m.pipelineCursor++
					m.updatePreview()
				}
			} else if m.activePane == componentsPane {
				components := m.getAllComponents()
				if m.componentCursor < len(components)-1 {
					m.componentCursor++
					m.updatePreview()
				}
			} else if m.activePane == previewPane {
				// Scroll preview down
				m.previewViewport.LineDown(1)
			}
		
		case "pgup":
			if m.activePane == previewPane {
				m.previewViewport.ViewUp()
			}
		
		case "pgdown":
			if m.activePane == previewPane {
				m.previewViewport.ViewDown()
			}
		
		case "p":
			m.showPreview = !m.showPreview
			m.updateViewportSizes()
			if m.showPreview {
				m.updatePreview()
			}
		
		case "enter":
			if m.activePane == pipelinesPane {
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					// View the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineViewerView,
							pipeline: m.pipelines[m.pipelineCursor],
						}
					}
				}
			} else if m.activePane == componentsPane {
				// Could add component viewing/editing functionality here
			}
		
		case "e":
			if m.activePane == pipelinesPane {
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					// Edit the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineBuilderView,
							pipeline: m.pipelines[m.pipelineCursor],
						}
					}
				}
			} else if m.activePane == componentsPane {
				// Edit component in TUI editor
				components := m.getAllComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					// Read the component content
					content, err := files.ReadComponent(comp.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					// Enter editing mode
					m.editingComponent = true
					m.editingComponentPath = comp.path
					m.editingComponentName = comp.name
					m.componentContent = content.Content
					m.originalContent = content.Content // Store original for change detection
					
					// Initialize edit viewport
					m.editViewport = viewport.New(m.width-8, m.height-10)
					m.editViewport.SetContent(content.Content)
					
					return m, nil
				}
			}
		
		case "n":
			if m.activePane == pipelinesPane {
				// Create new pipeline (switch to builder)
				return m, func() tea.Msg {
					return SwitchViewMsg{
						view: pipelineBuilderView,
					}
				}
			} else if m.activePane == componentsPane {
				// Create new component
				m.creatingComponent = true
				m.creationStep = 0
				m.typeCursor = 0
				m.componentName = ""
				m.componentContent = ""
				m.componentCreationType = ""
				return m, nil
			}
		
		case "r":
			// Refresh pipeline list
			m.loadPipelines()
			return m, func() tea.Msg {
				return StatusMsg("Pipeline list refreshed")
			}
		
		case "S":
			if m.activePane == pipelinesPane {
				// Set selected pipeline (generate PLUQQY.md)
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					return m, m.setPipeline(m.pipelines[m.pipelineCursor])
				}
			}
		
		case "d", "delete":
			if m.activePane == pipelinesPane {
				// Delete pipeline with confirmation
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					m.confirmingDelete = true
					m.deleteConfirmation = fmt.Sprintf("Delete pipeline '%s'? (y/n)", m.pipelines[m.pipelineCursor])
				}
			}
		}
	}
	
	// Update preview if needed
	if m.showPreview && m.previewContent != "" {
		// Preprocess content to handle carriage returns and ensure proper line breaks
		processedContent := strings.ReplaceAll(m.previewContent, "\r\r", "\n\n")
		processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
		// Wrap content to viewport width to prevent overflow
		wrappedContent := wordwrap.String(processedContent, m.previewViewport.Width)
		m.previewViewport.SetContent(wrappedContent)
	}
	
	// Update viewports
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	// Only forward non-key messages to viewports
	switch msg.(type) {
	case tea.KeyMsg:
		// Don't forward key messages - they're already handled
	default:
		// Forward other messages to viewports
		if m.showPreview {
			m.previewViewport, cmd = m.previewViewport.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
		m.pipelinesViewport, cmd = m.pipelinesViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.componentsViewport, cmd = m.componentsViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *MainListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: Failed to load pipelines: %v\n\nPress 'q' to quit", m.err)
	}
	
	// If showing exit confirmation, display dialog
	if m.showExitConfirmation {
		return m.exitConfirmationView()
	}
	
	// If creating component, show creation wizard
	if m.creatingComponent {
		return m.componentCreationView()
	}
	
	// If editing component, show edit view
	if m.editingComponent {
		return m.componentEditView()
	}

	// Styles
	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("236")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	typeHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	contentHeight := m.height - 14    // Reserve space for title, help pane, status message, and spacing

	if m.showPreview {
		contentHeight = contentHeight / 2
	}
	
	// Ensure minimum height for content
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
	heading := "COMPONENTS"
	remainingWidth := columnWidth - len(heading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	// Render heading and colons separately with different styles
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Subtle gray
	leftContent.WriteString(headerPadding.Render(typeHeaderStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	leftContent.WriteString("\n\n")

	
	// Table header styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))
	
	// Table column widths (adjusted for column width)
	nameWidth := 20
	tokenWidth := 7  // For "~Token" plus padding
	modifiedWidth := 12
	usageWidth := 8
	
	// Render table header with 2-space shift
	header := fmt.Sprintf("  %-*s %-*s %-*s %-*s", 
		nameWidth, "Name",
		tokenWidth, "~Token",
		modifiedWidth, "Modified",
		usageWidth, "Usage")
	leftContent.WriteString(headerPadding.Render(headerStyle.Render(header)))
	leftContent.WriteString("\n\n")
	
	// Build scrollable content for components viewport
	var componentsScrollContent strings.Builder
	
	allComponents := m.getAllComponents()
	
	if len(allComponents) == 0 {
		if m.activePane == componentsPane {
			// Active pane - show prominent message
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")). // Orange
				Bold(true)
			componentsScrollContent.WriteString(emptyStyle.Render("No components found.\n\nPress 'n' to create one"))
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			componentsScrollContent.WriteString(dimmedStyle.Render("No components found."))
		}
	} else {
		currentType := ""
	
		for i, comp := range allComponents {
		if comp.compType != currentType {
			if currentType != "" {
				componentsScrollContent.WriteString("\n")
			}
			currentType = comp.compType
			// Convert to uppercase plural form
			var typeHeader string
			switch currentType {
			case models.ComponentTypeContext:
				typeHeader = "CONTEXTS"
			case models.ComponentTypePrompt:
				typeHeader = "PROMPTS"
			case models.ComponentTypeRules:
				typeHeader = "RULES"
			default:
				typeHeader = strings.ToUpper(currentType)
			}
			componentsScrollContent.WriteString(typeHeaderStyle.Render(fmt.Sprintf("▸ %s", typeHeader)) + "\n")
		}

		// Format the row data
		nameStr := comp.name
		if len(nameStr) > nameWidth-3 {
			nameStr = nameStr[:nameWidth-6] + "..."
		}
		
		// Format modified time
		modifiedStr := ""
		if !comp.lastModified.IsZero() {
			if time.Since(comp.lastModified) < 24*time.Hour {
				modifiedStr = comp.lastModified.Format("15:04")
			} else if time.Since(comp.lastModified) < 7*24*time.Hour {
				modifiedStr = comp.lastModified.Format("Mon 15:04")
			} else {
				modifiedStr = comp.lastModified.Format("Jan 02")
			}
		}
		
		// Format usage count with visual indicator
		usageStr := fmt.Sprintf("%d", comp.usageCount)
		if comp.usageCount > 0 {
			bars := ""
			barCount := comp.usageCount
			if barCount > 5 {
				barCount = 5
			}
			for j := 0; j < barCount; j++ {
				bars += "█"
			}
			usageStr = fmt.Sprintf("%-2d %s", comp.usageCount, bars)
		}
		
		// Format token count - right-aligned with consistent width
		tokenStr := fmt.Sprintf("%d", comp.tokenCount)
		
		// Build the row with extra padding between token and modified
		row := fmt.Sprintf("%-*s %*s  %-*s %-*s",
			nameWidth, nameStr,
			tokenWidth-1, tokenStr,  // -1 to account for the space before it
			modifiedWidth, modifiedStr,
			usageWidth, usageStr)
		
		// Apply cursor if needed
		if m.activePane == componentsPane && i == m.componentCursor {
			row = "▸ " + row
		} else {
			row = "  " + row
		}
		
		// Apply styling
		if m.activePane == componentsPane && i == m.componentCursor {
			componentsScrollContent.WriteString(selectedStyle.Render(row))
		} else {
			componentsScrollContent.WriteString(normalStyle.Render(row))
		}
		
			if i < len(allComponents)-1 {
				componentsScrollContent.WriteString("\n")
			}
		}
	}
	
	// Update components viewport with content
	// Wrap content to viewport width to prevent overflow
	wrappedComponentsContent := wordwrap.String(componentsScrollContent.String(), m.componentsViewport.Width)
	m.componentsViewport.SetContent(wrappedComponentsContent)
	
	// Update viewport to follow cursor
	if m.activePane == componentsPane && len(allComponents) > 0 {
		// Calculate the line position of the cursor
		currentLine := 0
		for i := 0; i < m.componentCursor && i < len(allComponents); i++ {
			currentLine++ // Component line
			// Check if it's the first item of a new type to add header line
			if i == 0 || allComponents[i].compType != allComponents[i-1].compType {
				currentLine++ // Type header line
				if i > 0 {
					currentLine++ // Empty line between sections
				}
			}
		}
		
		// Ensure the cursor line is visible
		if currentLine < m.componentsViewport.YOffset {
			m.componentsViewport.SetYOffset(currentLine)
		} else if currentLine >= m.componentsViewport.YOffset+m.componentsViewport.Height {
			m.componentsViewport.SetYOffset(currentLine - m.componentsViewport.Height + 1)
		}
	}
	
	// Add padding to viewport content
	componentsViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	leftContent.WriteString(componentsViewportPadding.Render(m.componentsViewport.View()))

	// Build right column (pipelines)
	var rightContent strings.Builder
	// Create heading with colons spanning the width
	rightHeading := "PIPELINES"
	rightRemainingWidth := columnWidth - len(rightHeading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if rightRemainingWidth < 0 {
		rightRemainingWidth = 0
	}
	// Render heading and colons separately with different styles
	rightContent.WriteString(headerPadding.Render(typeHeaderStyle.Render(rightHeading) + " " + colonStyle.Render(strings.Repeat(":", rightRemainingWidth))))
	rightContent.WriteString("\n\n")
	
	// Table header for pipelines with token count
	pipelineHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))
	
	// Table column widths for pipelines
	pipelineNameWidth := 35
	pipelineTokenWidth := 7  // For "~Token"
	
	// Render table header
	pipelineHeader := fmt.Sprintf("  %-*s %*s", 
		pipelineNameWidth, "Name",
		pipelineTokenWidth, "~Token")
	rightContent.WriteString(headerPadding.Render(pipelineHeaderStyle.Render(pipelineHeader)))
	rightContent.WriteString("\n\n")

	// Build scrollable content for pipelines viewport
	var pipelinesScrollContent strings.Builder
	
	if len(m.pipelines) == 0 {
		if m.activePane == pipelinesPane {
			// Active pane - show prominent message
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")). // Orange
				Bold(true)
			pipelinesScrollContent.WriteString(emptyStyle.Render("No pipelines found.\n\nPress 'n' to create one"))
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			pipelinesScrollContent.WriteString(dimmedStyle.Render("No pipelines found."))
		}
	} else {
		for i, pipelineName := range m.pipelines {
			// Load pipeline to get token count
			pipeline, err := files.ReadPipeline(pipelineName)
			tokenCount := 0
			if err == nil && pipeline != nil {
				// Generate preview to calculate tokens
				output, err := composer.ComposePipeline(pipeline)
				if err == nil {
					tokenCount = utils.EstimateTokens(output)
				}
			}
			
			// Format the pipeline name
			nameStr := pipelineName
			if len(nameStr) > pipelineNameWidth-3 {
				nameStr = nameStr[:pipelineNameWidth-6] + "..."
			}
			
			// Format token count - right-aligned
			tokenStr := fmt.Sprintf("%d", tokenCount)
			
			// Build the row
			row := fmt.Sprintf("%-*s %*s",
				pipelineNameWidth, nameStr,
				pipelineTokenWidth, tokenStr)
			
			// Apply cursor if needed
			if m.activePane == pipelinesPane && i == m.pipelineCursor {
				row = "▸ " + row
			} else {
				row = "  " + row
			}
			
			// Apply styling
			if m.activePane == pipelinesPane && i == m.pipelineCursor {
				pipelinesScrollContent.WriteString(selectedStyle.Render(row))
			} else {
				pipelinesScrollContent.WriteString(normalStyle.Render(row))
			}
			
			if i < len(m.pipelines)-1 {
				pipelinesScrollContent.WriteString("\n")
			}
		}
	}
	
	// Update pipelines viewport with content
	// Wrap content to viewport width to prevent overflow
	wrappedPipelinesContent := wordwrap.String(pipelinesScrollContent.String(), m.pipelinesViewport.Width)
	m.pipelinesViewport.SetContent(wrappedPipelinesContent)
	
	// Update viewport to follow cursor
	if m.activePane == pipelinesPane && len(m.pipelines) > 0 {
		// For pipelines, each item is one line
		currentLine := m.pipelineCursor
		
		// Ensure the cursor line is visible
		if currentLine < m.pipelinesViewport.YOffset {
			m.pipelinesViewport.SetYOffset(currentLine)
		} else if currentLine >= m.pipelinesViewport.YOffset+m.pipelinesViewport.Height {
			m.pipelinesViewport.SetYOffset(currentLine - m.pipelinesViewport.Height + 1)
		}
	}
	
	// Add padding to viewport content
	pipelinesViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	rightContent.WriteString(pipelinesViewportPadding.Render(m.pipelinesViewport.View()))

	// Apply borders
	leftStyle := inactiveStyle
	rightStyle := inactiveStyle
	if m.activePane == componentsPane {
		leftStyle = activeStyle
	} else if m.activePane == pipelinesPane {
		rightStyle = activeStyle
	}

	leftColumn := leftStyle.
		Width(columnWidth).
		Height(contentHeight).
		Render(leftContent.String())

	rightColumn := rightStyle.
		Width(columnWidth).
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

	// Add preview if enabled
	if m.showPreview && m.previewContent != "" {
		// Calculate token count
		tokenCount := utils.EstimateTokens(m.previewContent)
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
		
		// Apply active/inactive style to preview border
		previewBorderColor := lipgloss.Color("243") // inactive
		if m.activePane == previewPane {
			previewBorderColor = lipgloss.Color("170") // active (same as other active borders)
		}
		
		previewBorderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(previewBorderColor).
			Width(m.width - 4) // Account for padding (2) and border (2)

		s.WriteString("\n")
		
		// Build preview content with header inside
		var previewContent strings.Builder
		// Create heading with colons and token info
		var previewHeading string
		if m.activePane == pipelinesPane {
			previewHeading = "PIPELINE PREVIEW (PLUQQY.md)"
		} else if m.activePane == componentsPane {
			previewHeading = "COMPONENT PREVIEW"
		} else {
			// Default to pipeline preview when preview pane is active
			previewHeading = "PREVIEW"
		}
		tokenInfo := tokenBadge
		
		// Calculate the actual rendered width of token info
		tokenInfoWidth := lipgloss.Width(tokenBadge)
		
		// Calculate total available width inside the border
		totalWidth := m.width - 8 // accounting for border padding and header padding
		
		// Calculate space for colons between heading and token info
		colonSpace := totalWidth - len(previewHeading) - tokenInfoWidth - 2 // -2 for spaces
		if colonSpace < 3 {
			colonSpace = 3
		}
		
		// Build the complete header line
		previewColonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Subtle gray
		previewHeaderPadding := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		previewContent.WriteString(previewHeaderPadding.Render(typeHeaderStyle.Render(previewHeading) + " " + previewColonStyle.Render(strings.Repeat(":", colonSpace)) + " " + tokenInfo))
		previewContent.WriteString("\n\n")
		// Add padding to preview viewport content
		previewViewportPadding := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		previewContent.WriteString(previewViewportPadding.Render(m.previewViewport.View()))
		
		// Render the border around the entire preview with same padding as top columns
		s.WriteString("\n")
		previewPaddingStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		s.WriteString(previewPaddingStyle.Render(previewBorderStyle.Render(previewContent.String())))
		
	}

	// Show delete confirmation if active
	if m.confirmingDelete {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2)
		s.WriteString("\n")
		s.WriteString(contentStyle.Render(confirmStyle.Render(m.deleteConfirmation)))
	}
	
	// Help text in bordered pane
	help := []string{
		"tab switch pane",
		"↑/↓ nav",
		"enter view/edit",
		"e edit",
		"n new",
		"d delete",
		"S set",
		"p preview",
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

	return s.String()
}

func (m *MainListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateViewportSizes()
}

func (m *MainListModel) updateViewportSizes() {
	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	contentHeight := m.height - 14    // Reserve space for title, help pane, status message, and spacing
	
	if m.showPreview {
		contentHeight = contentHeight / 2
	}
	
	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}
	
	// Update pipelines and components viewports
	// Reserve space for headers: heading (2 lines) + table header for components (2 lines) = 4 lines
	viewportHeight := contentHeight - 4
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	
	m.pipelinesViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.pipelinesViewport.Height = viewportHeight
	m.componentsViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.componentsViewport.Height = viewportHeight
	
	// Update preview viewport
	if m.showPreview {
		previewHeight := m.height / 2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		m.previewViewport.Width = m.width - 8 // Account for outer padding (2), borders (2), and inner padding (2) + extra spacing
		m.previewViewport.Height = previewHeight
	}
}

func (m *MainListModel) updatePreview() {
	if !m.showPreview {
		return
	}
	
	if m.activePane == pipelinesPane {
		// Show pipeline preview
		if len(m.pipelines) == 0 {
			m.previewContent = "No pipelines to preview."
			return
		}
		
		if m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			pipelineName := m.pipelines[m.pipelineCursor]
			
			// Load pipeline
			pipeline, err := files.ReadPipeline(pipelineName)
			if err != nil {
				m.previewContent = fmt.Sprintf("Error loading pipeline: %v", err)
				return
			}
			
			// Generate preview
			output, err := composer.ComposePipeline(pipeline)
			if err != nil {
				m.previewContent = fmt.Sprintf("Error generating preview: %v", err)
				return
			}
			
			m.previewContent = output
		}
	} else if m.activePane == componentsPane {
		// Show component preview
		components := m.getAllComponents()
		if len(components) == 0 {
			m.previewContent = "No components to preview."
			return
		}
		
		if m.componentCursor >= 0 && m.componentCursor < len(components) {
			comp := components[m.componentCursor]
			
			// Read component content
			content, err := files.ReadComponent(comp.path)
			if err != nil {
				m.previewContent = fmt.Sprintf("Error loading component: %v", err)
				return
			}
			
			// Format component preview with metadata
			var preview strings.Builder
			preview.WriteString(fmt.Sprintf("# %s\n\n", comp.name))
			preview.WriteString(fmt.Sprintf("**Type:** %s\n", strings.Title(comp.compType)))
			preview.WriteString(fmt.Sprintf("**Path:** %s\n", comp.path))
			preview.WriteString(fmt.Sprintf("**Usage Count:** %d\n", comp.usageCount))
			preview.WriteString(fmt.Sprintf("**Token Count:** ~%d\n", comp.tokenCount))
			if !comp.lastModified.IsZero() {
				preview.WriteString(fmt.Sprintf("**Last Modified:** %s\n", comp.lastModified.Format("2006-01-02 15:04:05")))
			}
			preview.WriteString("\n---\n\n")
			preview.WriteString(content.Content)
			
			m.previewContent = preview.String()
		}
	}
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
		if m.pipelineCursor >= len(m.pipelines) && m.pipelineCursor > 0 {
			m.pipelineCursor = len(m.pipelines) - 1
		}
		
		return StatusMsg(fmt.Sprintf("✓ Deleted pipeline: %s", pipelineName))
	}
}

func (m *MainListModel) handleComponentCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.creationStep {
	case 0: // Type selection
		switch msg.String() {
		case "esc":
			m.creatingComponent = false
			return m, nil
		case "up", "k":
			if m.typeCursor > 0 {
				m.typeCursor--
			}
		case "down", "j":
			if m.typeCursor < 2 {
				m.typeCursor++
			}
		case "enter":
			types := []string{models.ComponentTypeContext, models.ComponentTypePrompt, models.ComponentTypeRules}
			m.componentCreationType = types[m.typeCursor]
			m.creationStep = 1
		}
		
	case 1: // Name input
		switch msg.String() {
		case "esc":
			m.creationStep = 0
			m.componentName = ""
		case "enter":
			if strings.TrimSpace(m.componentName) != "" {
				m.creationStep = 2
			}
		case "backspace":
			if len(m.componentName) > 0 {
				m.componentName = m.componentName[:len(m.componentName)-1]
			}
		case " ":
			m.componentName += " "
		default:
			if msg.Type == tea.KeyRunes {
				m.componentName += string(msg.Runes)
			}
		}
		
	case 2: // Content input
		switch msg.String() {
		case "ctrl+s":
			// Save component
			return m, m.saveNewComponent()
		case "esc":
			// If content has been entered, show confirmation
			if strings.TrimSpace(m.componentContent) != "" {
				m.showExitConfirmation = true
				m.exitConfirmationType = "component"
			} else {
				m.creationStep = 1
				m.componentContent = ""
			}
		case "enter":
			m.componentContent += "\n"
		case "backspace":
			if len(m.componentContent) > 0 {
				m.componentContent = m.componentContent[:len(m.componentContent)-1]
			}
		case "tab":
			m.componentContent += "    "
		case " ":
			// Allow spaces
			m.componentContent += " "
		default:
			if msg.Type == tea.KeyRunes {
				m.componentContent += string(msg.Runes)
			}
		}
	}
	
	return m, nil
}

func (m *MainListModel) handleComponentEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Handle viewport scrolling
	switch msg.String() {
	case "up", "k", "pgup":
		m.editViewport, cmd = m.editViewport.Update(msg)
		return m, cmd
	case "down", "j", "pgdown":
		m.editViewport, cmd = m.editViewport.Update(msg)
		return m, cmd
	case "ctrl+s":
		// Save component and exit
		return m, m.saveEditedComponent()
	case "esc":
		// Check if content has changed
		if m.componentContent != m.originalContent {
			// Show confirmation dialog
			m.showExitConfirmation = true
			m.exitConfirmationType = "component-edit"
			return m, nil
		}
		// No changes, exit immediately
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.originalContent = ""
		return m, nil
	case "enter":
		m.componentContent += "\n"
	case "backspace":
		if len(m.componentContent) > 0 {
			m.componentContent = m.componentContent[:len(m.componentContent)-1]
		}
	case "tab":
		m.componentContent += "    "
	case " ":
		m.componentContent += " "
	default:
		if msg.Type == tea.KeyRunes {
			m.componentContent += string(msg.Runes)
		}
	}
	
	return m, nil
}

func (m *MainListModel) componentCreationView() string {
	switch m.creationStep {
	case 0:
		return m.componentTypeSelectionView()
	case 1:
		return m.componentNameInputView()
	case 2:
		return m.componentContentEditView()
	}
	
	return "Unknown creation step"
}

func (m *MainListModel) componentTypeSelectionView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("236")).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 10 // Reserve space for help pane

	// Build main content
	var mainContent strings.Builder

	// Header with colons
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	heading := "CREATE NEW COMPONENT"
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Component type selection
	mainContent.WriteString(headerPadding.Render("Select component type:"))
	mainContent.WriteString("\n\n")

	types := []struct {
		name string
		desc string
	}{
		{"CONTEXT", "Background information or system state"},
		{"PROMPT", "Instructions or questions for the LLM"},
		{"RULES", "Important constraints or guidelines"},
	}

	for i, t := range types {
		cursor := "  "
		if i == m.typeCursor {
			cursor = "▸ "
		}

		line := cursor + t.name
		if i == m.typeCursor {
			mainContent.WriteString(selectedStyle.Render(line))
		} else {
			mainContent.WriteString(normalStyle.Render(line))
		}
		mainContent.WriteString("\n")
		mainContent.WriteString("  " + descStyle.Render(t.desc))
		mainContent.WriteString("\n\n")
	}

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"↑/↓ navigate",
		"enter select",
		"esc cancel",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)

	// Combine all elements
	var s strings.Builder

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(mainPane))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) componentNameInputView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(60)

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 10 // Reserve space for help pane

	// Build main content
	var mainContent strings.Builder

	// Header with colons
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	heading := fmt.Sprintf("CREATE NEW %s", strings.ToUpper(m.componentCreationType))
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Name input prompt - centered
	promptText := promptStyle.Render("Enter a descriptive name:")
	centeredPromptStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(promptText)))
	mainContent.WriteString("\n\n")

	// Input field with cursor
	input := m.componentName + "│" // cursor

	// Render input field with padding for centering
	inputFieldContent := inputStyle.Render(input)
	
	// Add padding to center the input field properly
	centeredInputStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	
	mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(inputFieldContent)))
	
	// Check if component name already exists and show warning
	if m.componentName != "" {
		testFilename := sanitizeFileName(m.componentName) + ".md"
		var componentType string
		switch m.componentCreationType {
		case models.ComponentTypeContext:
			componentType = "contexts"
		case models.ComponentTypePrompt:
			componentType = "prompts"
		case models.ComponentTypeRules:
			componentType = "rules"
		}
		
		existingComponents, _ := files.ListComponents(componentType)
		for _, existing := range existingComponents {
			if strings.EqualFold(existing, testFilename) {
				// Show warning
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // Orange/yellow for warning
					Bold(true)
				warningText := warningStyle.Render(fmt.Sprintf("⚠ Warning: %s '%s' already exists", strings.Title(m.componentCreationType), m.componentName))
				mainContent.WriteString("\n\n")
				mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(warningText)))
				break
			}
		}
	}

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"enter continue",
		"esc back",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)

	// Combine all elements
	var s strings.Builder

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(mainPane))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) componentContentEditView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 10 // Reserve space for help pane

	// Build main content
	var mainContent strings.Builder

	// Header with colons
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	heading := fmt.Sprintf("EDIT %s: %s", strings.ToUpper(m.componentCreationType), m.componentName)
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Editor content with cursor
	content := m.componentContent + "│" // cursor
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := strings.ReplaceAll(content, "\r\r", "\n\n")
	processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
	
	// Calculate available width for wrapping (accounting for padding)
	availableWidth := contentWidth - 4 // 2 for border, 2 for headerPadding
	if availableWidth < 1 {
		availableWidth = 1
	}
	
	// Wrap content to prevent overflow
	wrappedContent := wordwrap.String(processedContent, availableWidth)

	mainContent.WriteString(headerPadding.Render(wrappedContent))

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"ctrl+s save",
		"esc back",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)

	// Combine all elements
	var s strings.Builder

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(mainPane))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) componentEditView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	// Calculate dimensions  
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 6 // Reserve space for help pane (3) + spacing (3)

	// Build main content
	var mainContent strings.Builder

	// Header with colons
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	heading := fmt.Sprintf("EDITING: %s", strings.ToUpper(m.editingComponentName))
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Update viewport dimensions if needed
	viewportWidth := contentWidth - 4 // 2 for border, 2 for headerPadding
	viewportHeight := contentHeight - 5 // Account for header and spacing
	if m.editViewport.Width != viewportWidth || m.editViewport.Height != viewportHeight {
		m.editViewport.Width = viewportWidth
		m.editViewport.Height = viewportHeight
	}
	
	// Editor content with cursor
	content := m.componentContent + "│" // cursor
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := strings.ReplaceAll(content, "\r\r", "\n\n")
	processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
	
	// Wrap content to viewport width to prevent overflow
	wrappedContent := wordwrap.String(processedContent, viewportWidth)
	
	// Update viewport content
	m.editViewport.SetContent(wrappedContent)
	
	// Use viewport for scrollable content
	mainContent.WriteString(headerPadding.Render(m.editViewport.View()))

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"↑/↓ scroll",
		"ctrl+s save",
		"esc cancel",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)

	// Combine all elements
	var s strings.Builder

	// Add top margin to ensure content is not cut off
	s.WriteString("\n")

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(mainPane))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) exitConfirmationView() string {
	// Styles matching the rest of the UI
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers
		
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")) // Orange for warning
		
	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	
	// Calculate dimensions
	contentWidth := m.width - 4
	contentHeight := 10 // Small dialog
	
	// Build main content
	var mainContent strings.Builder
	
	// Header
	header := "EXIT CONFIRMATION"
	centeredHeader := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(headerStyle.Render(header))
	mainContent.WriteString(centeredHeader)
	mainContent.WriteString("\n\n")
	
	// Warning message
	var warningMsg string
	if m.exitConfirmationType == "component" {
		warningMsg = "You have unsaved content in this component."
	} else if m.exitConfirmationType == "component-edit" {
		warningMsg = "You have unsaved changes to this component."
	}
	
	centeredWarning := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(warningStyle.Render(warningMsg))
	mainContent.WriteString(centeredWarning)
	mainContent.WriteString("\n")
	
	exitWarning := "Are you sure you want to exit?"
	centeredExitWarning := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(warningStyle.Render(exitWarning))
	mainContent.WriteString(centeredExitWarning)
	mainContent.WriteString("\n\n")
	
	// Options
	options := "[Y]es, exit  /  [N]o, stay"
	centeredOptions := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(optionStyle.Render(options))
	mainContent.WriteString(centeredOptions)
	
	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())
	
	// Center the dialog vertically
	verticalPadding := (m.height - contentHeight - 4) / 2
	dialogStyle := lipgloss.NewStyle().
		PaddingTop(verticalPadding).
		PaddingLeft(1).
		PaddingRight(1)
		
	return dialogStyle.Render(mainPane)
}

func (m *MainListModel) saveNewComponent() tea.Cmd {
	return func() tea.Msg {
		// Create filename from name using sanitization
		filename := sanitizeFileName(m.componentName) + ".md"
		
		// Determine directory and component type for listing
		var dir string
		var componentType string
		switch m.componentCreationType {
		case models.ComponentTypeContext:
			dir = filepath.Join(files.ComponentsDir, files.ContextsDir)
			componentType = "contexts"
		case models.ComponentTypePrompt:
			dir = filepath.Join(files.ComponentsDir, files.PromptsDir)
			componentType = "prompts"
		case models.ComponentTypeRules:
			dir = filepath.Join(files.ComponentsDir, files.RulesDir)
			componentType = "rules"
		}
		
		// Check if component already exists (case-insensitive)
		existingComponents, err := files.ListComponents(componentType)
		if err != nil {
			return StatusMsg(fmt.Sprintf("❌ Failed to check existing components: %v", err))
		}
		
		for _, existing := range existingComponents {
			if strings.EqualFold(existing, filename) {
				return StatusMsg(fmt.Sprintf("❌ %s '%s' already exists. Please choose a different name.", strings.Title(m.componentCreationType), m.componentName))
			}
		}
		
		path := filepath.Join(dir, filename)
		
		// Write component
		err = files.WriteComponent(path, m.componentContent)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save component '%s': %v", m.componentName, err))
		}
		
		// Reset creation state
		m.creatingComponent = false
		m.componentName = ""
		m.componentContent = ""
		m.creationStep = 0
		
		// Reload components
		m.loadComponents()
		
		return StatusMsg(fmt.Sprintf("✓ Created %s: %s", m.componentCreationType, filename))
	}
}

func (m *MainListModel) saveEditedComponent() tea.Cmd {
	return func() tea.Msg {
		// Write component
		err := files.WriteComponent(m.editingComponentPath, m.componentContent)
		if err != nil {
			return StatusMsg(fmt.Sprintf("❌ Failed to save: %v", err))
		}
		
		// Clear editing state
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.originalContent = ""
		
		// Reload components
		m.loadComponents()
		
		return StatusMsg(fmt.Sprintf("✓ Saved: %s", m.editingComponentName))
	}
}

