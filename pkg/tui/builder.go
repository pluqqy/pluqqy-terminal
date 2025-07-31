package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type column int

const (
	leftColumn column = iota
	rightColumn
	previewColumn
)

type PipelineBuilderModel struct {
	width    int
	height   int
	pipeline *models.Pipeline

	// Available components (left column)
	prompts  []componentItem
	contexts []componentItem
	rules    []componentItem

	// Selected components (right column)
	selectedComponents []models.ComponentRef

	// UI state
	activeColumn   column
	leftCursor     int
	rightCursor    int
	showPreview    bool
	previewContent string
	previewViewport viewport.Model
	leftViewport   viewport.Model  // For available components
	rightViewport  viewport.Model  // For selected components
	err            error
	
	// Name input state
	editingName bool
	nameInput   string
	
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
	editSaveMessage       string
	editSaveTimer         *time.Timer
	
	// Pipeline save state
	pipelineSaveMessage   string
	pipelineSaveTimer     *time.Timer
}

type componentItem struct {
	name         string
	path         string
	compType     string
	lastModified time.Time
	usageCount   int
}

type clearEditSaveMsg struct{}
type clearPipelineSaveMsg struct{}

func NewPipelineBuilderModel() *PipelineBuilderModel {
	// Load settings for UI preferences
	settings, _ := files.ReadSettings()
	
	m := &PipelineBuilderModel{
		activeColumn: leftColumn,
		showPreview:  settings.UI.ShowPreview,
		editingName:  true,
		nameInput:    "",
		pipeline: &models.Pipeline{
			Name:       "",
			Components: []models.ComponentRef{},
		},
		previewViewport: viewport.New(80, 20), // Default size, will be resized
		leftViewport:    viewport.New(40, 20), // Default size, will be resized
		rightViewport:   viewport.New(40, 20), // Default size, will be resized
	}
	m.loadAvailableComponents()
	return m
}

func (m *PipelineBuilderModel) loadAvailableComponents() {
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
		
		// Calculate usage count - need to check different path formats
		usage := 0
		relativePath := "../" + componentPath
		if count, exists := usageMap[relativePath]; exists {
			usage = count
		}
		
		m.prompts = append(m.prompts, componentItem{
			name:         p,
			path:         componentPath,
			compType:     models.ComponentTypePrompt,
			lastModified: modTime,
			usageCount:   usage,
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
		
		m.contexts = append(m.contexts, componentItem{
			name:         c,
			path:         componentPath,
			compType:     models.ComponentTypeContext,
			lastModified: modTime,
			usageCount:   usage,
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
		
		m.rules = append(m.rules, componentItem{
			name:         r,
			path:         componentPath,
			compType:     models.ComponentTypeRules,
			lastModified: modTime,
			usageCount:   usage,
		})
	}
}

func (m *PipelineBuilderModel) Init() tea.Cmd {
	return nil
}

func (m *PipelineBuilderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSizes()

	case clearEditSaveMsg:
		m.editSaveMessage = ""
		// Exit editing mode after save confirmation is shown
		if !m.editingComponent {
			return m, nil
		}
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		return m, nil

	case clearPipelineSaveMsg:
		m.pipelineSaveMessage = ""
		return m, nil

	case tea.KeyMsg:
		// Handle component creation mode
		if m.creatingComponent {
			return m.handleComponentCreation(msg)
		}
		
		// Handle component editing mode
		if m.editingComponent {
			return m.handleComponentEditing(msg)
		}
		
		// Handle name editing mode
		if m.editingName {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.nameInput) != "" {
					m.pipeline.Name = strings.TrimSpace(m.nameInput)
					m.editingName = false
				}
			case "esc":
				// Cancel and return to main list
				return m, func() tea.Msg {
					return SwitchViewMsg{view: mainListView}
				}
			case "backspace":
				if len(m.nameInput) > 0 {
					m.nameInput = m.nameInput[:len(m.nameInput)-1]
				}
			case " ":
				// Allow spaces
				m.nameInput += " "
			default:
				// Add character to input
				if msg.Type == tea.KeyRunes {
					m.nameInput += string(msg.Runes)
				}
			}
			return m, nil
		}

		// If preview is showing and active, handle viewport navigation

		// Normal mode keybindings
		switch msg.String() {
		case "esc":
			// Clear status message if present, otherwise return to main list
			if m.pipelineSaveMessage != "" {
				m.pipelineSaveMessage = ""
				// Cancel timer if running
				if m.pipelineSaveTimer != nil {
					m.pipelineSaveTimer.Stop()
				}
				return m, nil
			}
			// Return to main list
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "tab":
			// Switch between columns
			if m.showPreview {
				// When preview is shown, cycle through all three panes
				switch m.activeColumn {
				case leftColumn:
					m.activeColumn = rightColumn
				case rightColumn:
					m.activeColumn = previewColumn
				case previewColumn:
					m.activeColumn = leftColumn
				}
			} else {
				// When preview is hidden, only toggle between left and right
				if m.activeColumn == leftColumn {
					m.activeColumn = rightColumn
				} else {
					m.activeColumn = leftColumn
				}
			}

		case "up", "k":
			if m.activeColumn == previewColumn {
				// Scroll preview up
				m.previewViewport.LineUp(1)
			} else {
				m.moveCursor(-1)
			}

		case "down", "j":
			if m.activeColumn == previewColumn {
				// Scroll preview down
				m.previewViewport.LineDown(1)
			} else {
				m.moveCursor(1)
			}
			
		case "pgup":
			if m.activeColumn == previewColumn {
				// Scroll preview page up
				m.previewViewport.ViewUp()
			}
			
		case "pgdown":
			if m.activeColumn == previewColumn {
				// Scroll preview page down
				m.previewViewport.ViewDown()
			}

		case "enter":
			if m.activeColumn == leftColumn {
				m.addSelectedComponent()
			} else if m.activeColumn == rightColumn && len(m.selectedComponents) > 0 {
				// Edit selected component in right column
				return m, m.editComponent()
			}

		case "delete", "backspace", "d":
			if m.activeColumn == rightColumn {
				m.removeSelectedComponent()
			}

		case "p":
			m.showPreview = !m.showPreview
			m.updateViewportSizes()

		case "ctrl+s":
			// Save pipeline
			return m, m.savePipeline()
			
		case "S":
			// Save and set pipeline (generate PLUQQY.md)
			return m, m.saveAndSetPipeline()

		case "ctrl+up", "K":
			if m.activeColumn == rightColumn {
				m.moveComponentUp()
			}

		case "ctrl+down", "J":
			if m.activeColumn == rightColumn {
				m.moveComponentDown()
			}
		
		case "n":
			// Create new component
			if m.activeColumn == leftColumn {
				m.creatingComponent = true
				m.creationStep = 0
				m.typeCursor = 0
				m.componentName = ""
				m.componentContent = ""
				m.componentCreationType = ""
				return m, nil
			}
		
		case "E":
			// Edit component in external editor
			if m.activeColumn == leftColumn {
				components := m.getAllAvailableComponents()
				if m.leftCursor >= 0 && m.leftCursor < len(components) {
					return m, m.editComponentFromLeft()
				}
			}
		
		case "e":
			// Edit component in TUI editor
			if m.activeColumn == leftColumn {
				components := m.getAllAvailableComponents()
				if m.leftCursor >= 0 && m.leftCursor < len(components) {
					comp := components[m.leftCursor]
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
					return m, nil
				}
			}
		}
	}

	// Update preview if needed
	m.updatePreview()

	// Update viewport content if preview changed
	if m.showPreview && m.previewContent != "" {
		m.previewViewport.SetContent(m.previewContent)
		// Also forward other messages to viewport
		m.previewViewport, cmd = m.previewViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	// Forward messages to left/right viewports
	m.leftViewport, cmd = m.leftViewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.rightViewport, cmd = m.rightViewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *PipelineBuilderModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'Esc' to return", m.err)
	}

	// If creating component, show creation wizard
	if m.creatingComponent {
		return m.componentCreationView()
	}
	
	// If editing component, show edit view
	if m.editingComponent {
		return m.componentEditView()
	}
	
	// If editing name, show name input screen
	if m.editingName {
		return m.nameInputView()
	}

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

	// Build left column (available components)
	var leftContent strings.Builder
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Create heading with colons spanning the width
	heading := "AVAILABLE COMPONENTS"
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
		Foreground(lipgloss.Color("241")).
		Underline(true)
	
	// Table column widths (adjusted for left column width)
	nameWidth := 20
	modifiedWidth := 12
	usageWidth := 8
	
	// Render table header
	header := fmt.Sprintf("%-*s %-*s %-*s", 
		nameWidth, "Name",
		modifiedWidth, "Modified",
		usageWidth, "Usage")
	leftContent.WriteString(headerPadding.Render(headerStyle.Render(header)))
	leftContent.WriteString("\n\n")
	
	// Build scrollable content for left viewport
	var leftScrollContent strings.Builder
	
	allComponents := m.getAllAvailableComponents()
	currentType := ""
	
	for i, comp := range allComponents {
		if comp.compType != currentType {
			if currentType != "" {
				leftScrollContent.WriteString("\n")
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
			leftScrollContent.WriteString(typeHeaderStyle.Render(fmt.Sprintf("▸ %s", typeHeader)) + "\n")
		}

		// Check if component is already in pipeline
		isAdded := false
		componentPath := "../" + comp.path
		for _, existing := range m.selectedComponents {
			if existing.Path == componentPath {
				isAdded = true
				break
			}
		}

		// Format the row data
		nameStr := comp.name
		if len(nameStr) > nameWidth-3 {
			nameStr = nameStr[:nameWidth-6] + "..."
		}
		if isAdded {
			nameStr = nameStr + " ✓"
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
		
		// Build the row
		row := fmt.Sprintf("%-*s %-*s %-*s",
			nameWidth, nameStr,
			modifiedWidth, modifiedStr,
			usageWidth, usageStr)
		
		// Apply cursor if needed
		if m.activeColumn == leftColumn && i == m.leftCursor {
			row = "▸ " + row
		} else {
			row = "  " + row
		}
		
		// Apply styling
		if m.activeColumn == leftColumn && i == m.leftCursor {
			leftScrollContent.WriteString(selectedStyle.Render(row))
		} else if isAdded {
			// Use a dimmed style for already added components
			addedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			leftScrollContent.WriteString(addedStyle.Render(row))
		} else {
			leftScrollContent.WriteString(normalStyle.Render(row))
		}
		
		if i < len(allComponents)-1 {
			leftScrollContent.WriteString("\n")
		}
	}
	
	// Update left viewport with content
	m.leftViewport.SetContent(leftScrollContent.String())
	// Add padding to viewport content
	leftViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	leftContent.WriteString(leftViewportPadding.Render(m.leftViewport.View()))

	// Build right column (selected components)
	var rightContent strings.Builder
	// Create heading with colons spanning the width
	rightHeading := "PIPELINE COMPONENTS"
	rightRemainingWidth := columnWidth - len(rightHeading) - 5 // -5 for space and padding (2 left + 2 right + 1 space)
	if rightRemainingWidth < 0 {
		rightRemainingWidth = 0
	}
	// Render heading and colons separately with different styles
	rightContent.WriteString(headerPadding.Render(typeHeaderStyle.Render(rightHeading) + " " + colonStyle.Render(strings.Repeat(":", rightRemainingWidth))))
	rightContent.WriteString("\n\n")
	
	// Build scrollable content for right viewport
	var rightScrollContent strings.Builder

	if len(m.selectedComponents) == 0 {
		rightScrollContent.WriteString(normalStyle.Render("No components selected\n\nPress Tab to switch columns\nPress Enter to add components"))
	} else {
		// Group components by type
		var contexts, prompts, rules []models.ComponentRef
		for _, comp := range m.selectedComponents {
			switch comp.Type {
			case models.ComponentTypeContext:
				contexts = append(contexts, comp)
			case models.ComponentTypePrompt:
				prompts = append(prompts, comp)
			case models.ComponentTypeRules:
				rules = append(rules, comp)
			}
		}
		
		// Track overall index for cursor position
		overallIndex := 0
		
		// Render contexts
		if len(contexts) > 0 {
			rightScrollContent.WriteString(typeHeaderStyle.Render("▸ CONTEXTS") + "\n")
			for _, comp := range contexts {
				cursor := "  "
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					cursor = "▸ "
				}
				
				line := fmt.Sprintf("%s%s", cursor, filepath.Base(comp.Path))
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					rightScrollContent.WriteString(selectedStyle.Render(line) + "\n")
				} else {
					rightScrollContent.WriteString(normalStyle.Render(line) + "\n")
				}
				overallIndex++
			}
			if len(prompts) > 0 || len(rules) > 0 {
				rightScrollContent.WriteString("\n")
			}
		}
		
		// Render prompts
		if len(prompts) > 0 {
			rightScrollContent.WriteString(typeHeaderStyle.Render("▸ PROMPTS") + "\n")
			for _, comp := range prompts {
				cursor := "  "
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					cursor = "▸ "
				}
				
				line := fmt.Sprintf("%s%s", cursor, filepath.Base(comp.Path))
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					rightScrollContent.WriteString(selectedStyle.Render(line) + "\n")
				} else {
					rightScrollContent.WriteString(normalStyle.Render(line) + "\n")
				}
				overallIndex++
			}
			if len(rules) > 0 {
				rightScrollContent.WriteString("\n")
			}
		}
		
		// Render rules
		if len(rules) > 0 {
			rightScrollContent.WriteString(typeHeaderStyle.Render("▸ RULES") + "\n")
			for _, comp := range rules {
				cursor := "  "
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					cursor = "▸ "
				}
				
				line := fmt.Sprintf("%s%s", cursor, filepath.Base(comp.Path))
				if m.activeColumn == rightColumn && overallIndex == m.rightCursor {
					rightScrollContent.WriteString(selectedStyle.Render(line) + "\n")
				} else {
					rightScrollContent.WriteString(normalStyle.Render(line) + "\n")
				}
				overallIndex++
			}
		}
	}
	
	// Update right viewport with content
	m.rightViewport.SetContent(rightScrollContent.String())
	// Add padding to viewport content
	rightViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	rightContent.WriteString(rightViewportPadding.Render(m.rightViewport.View()))

	// Apply borders
	leftStyle := inactiveStyle
	rightStyle := inactiveStyle
	if m.activeColumn == leftColumn {
		leftStyle = activeStyle
	} else if m.activeColumn == rightColumn {
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
	title := "🔧 Pipeline Builder"
	if m.pipeline.Name != "" {
		title = fmt.Sprintf("🔧 Pipeline: %s", m.pipeline.Name)
	}
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")
	
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
		if m.activeColumn == previewColumn {
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
		previewHeading := "PIPELINE PREVIEW (PLUQQY.md)"
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

	// Help text in bordered pane
	help := []string{
		"tab switch pane",
		"↑/↓ nav",
		"enter add/edit",
		"n new",
		"E edit external",
		"e edit tui",
		"del remove",
		"K/J reorder",
		"p preview",
		"ctrl+s save",
		"S save+set",
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

	// Status message area - always render to maintain consistent layout
	saveMessageStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("82")). // Green for success
		Width(m.width).
		Align(lipgloss.Center).
		Padding(0, 1).
		MarginTop(1)
	
	s.WriteString("\n")
	if m.pipelineSaveMessage != "" {
		s.WriteString(saveMessageStyle.Render(m.pipelineSaveMessage))
	} else {
		// Render empty space to maintain layout
		s.WriteString(lipgloss.NewStyle().Height(1).Render(""))
	}

	return s.String()
}

func (m *PipelineBuilderModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateViewportSizes()
}

func (m *PipelineBuilderModel) updateViewportSizes() {
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
	
	// Update left and right viewports for table content
	// Reserve space for headers: heading (2 lines) + table header (2 lines) = 4 lines
	viewportHeight := contentHeight - 4
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	
	m.leftViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.leftViewport.Height = viewportHeight
	m.rightViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.rightViewport.Height = viewportHeight
	
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

func (m *PipelineBuilderModel) SetPipeline(pipeline string) {
	if pipeline != "" {
		// Load existing pipeline for editing
		p, err := files.ReadPipeline(pipeline)
		if err != nil {
			m.err = err
			return
		}
		m.pipeline = p
		m.selectedComponents = p.Components
		m.editingName = false // Don't show name input when editing
		m.nameInput = p.Name
		
		// Reorganize components by type to match display
		m.reorganizeComponentsByType()
		
		// Update local usage counts to reflect this pipeline's components
		// This ensures the counts show what would happen if we save
		for _, comp := range m.selectedComponents {
			componentPath := strings.TrimPrefix(comp.Path, "../")
			m.updateLocalUsageCount(componentPath, 1)
		}
		
		// Update preview to show the loaded pipeline
		m.updatePreview()
		
		// Set viewport content if preview is enabled
		if m.showPreview && m.previewContent != "" {
			m.previewViewport.SetContent(m.previewContent)
		}
	}
}

// Helper methods
func (m *PipelineBuilderModel) getAllAvailableComponents() []componentItem {
	var all []componentItem
	all = append(all, m.contexts...)
	all = append(all, m.prompts...)
	all = append(all, m.rules...)
	return all
}

func (m *PipelineBuilderModel) moveCursor(delta int) {
	if m.activeColumn == leftColumn {
		components := m.getAllAvailableComponents()
		m.leftCursor += delta
		if m.leftCursor < 0 {
			m.leftCursor = 0
		}
		if m.leftCursor >= len(components) {
			m.leftCursor = len(components) - 1
		}
	} else {
		m.rightCursor += delta
		if m.rightCursor < 0 {
			m.rightCursor = 0
		}
		if m.rightCursor >= len(m.selectedComponents) {
			m.rightCursor = len(m.selectedComponents) - 1
		}
	}
}

func (m *PipelineBuilderModel) addSelectedComponent() {
	components := m.getAllAvailableComponents()
	if m.leftCursor >= 0 && m.leftCursor < len(components) {
		selected := components[m.leftCursor]
		
		// Check if component is already added
		componentPath := "../" + selected.path
		for _, existing := range m.selectedComponents {
			if existing.Path == componentPath {
				// Component already exists, don't add duplicate
				return
			}
		}
		
		// Create component ref with relative path
		ref := models.ComponentRef{
			Type:  selected.compType,
			Path:  componentPath,
			Order: 0, // Will be set when inserting
		}
		
		// Insert component in the correct position based on type grouping
		m.insertComponentByType(ref)
		
		// Update the usage count locally to show predicted usage
		// This gives immediate feedback before the pipeline is saved
		m.updateLocalUsageCount(selected.path, 1)
	}
}

// insertComponentByType inserts a component in the correct position based on type grouping
func (m *PipelineBuilderModel) insertComponentByType(newComp models.ComponentRef) {
	// Add the component to the list
	m.selectedComponents = append(m.selectedComponents, newComp)
	
	// Reorganize to maintain type grouping
	m.reorganizeComponentsByType()
}

// reorganizeComponentsByType sorts components into groups: contexts, prompts, rules
func (m *PipelineBuilderModel) reorganizeComponentsByType() {
	// Separate components by type
	var contexts, prompts, rules []models.ComponentRef
	
	for _, comp := range m.selectedComponents {
		switch comp.Type {
		case models.ComponentTypeContext:
			contexts = append(contexts, comp)
		case models.ComponentTypePrompt:
			prompts = append(prompts, comp)
		case models.ComponentTypeRules:
			rules = append(rules, comp)
		}
	}
	
	// Rebuild the array in grouped order
	m.selectedComponents = nil
	m.selectedComponents = append(m.selectedComponents, contexts...)
	m.selectedComponents = append(m.selectedComponents, prompts...)
	m.selectedComponents = append(m.selectedComponents, rules...)
	
	// Update order numbers
	for i := range m.selectedComponents {
		m.selectedComponents[i].Order = i + 1
	}
}

// updateLocalUsageCount updates the usage count for a component locally
func (m *PipelineBuilderModel) updateLocalUsageCount(componentPath string, delta int) {
	// Update in prompts
	for i := range m.prompts {
		if m.prompts[i].path == componentPath {
			m.prompts[i].usageCount += delta
			if m.prompts[i].usageCount < 0 {
				m.prompts[i].usageCount = 0
			}
			return
		}
	}
	
	// Update in contexts
	for i := range m.contexts {
		if m.contexts[i].path == componentPath {
			m.contexts[i].usageCount += delta
			if m.contexts[i].usageCount < 0 {
				m.contexts[i].usageCount = 0
			}
			return
		}
	}
	
	// Update in rules
	for i := range m.rules {
		if m.rules[i].path == componentPath {
			m.rules[i].usageCount += delta
			if m.rules[i].usageCount < 0 {
				m.rules[i].usageCount = 0
			}
			return
		}
	}
}

func (m *PipelineBuilderModel) removeSelectedComponent() {
	if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents) {
		// Get the component path to update usage count
		removedComponent := m.selectedComponents[m.rightCursor]
		componentPath := strings.TrimPrefix(removedComponent.Path, "../")
		
		// Remember the type of component we're removing to adjust cursor properly
		removedType := removedComponent.Type
		
		// Remove component
		m.selectedComponents = append(
			m.selectedComponents[:m.rightCursor],
			m.selectedComponents[m.rightCursor+1:]...,
		)
		
		// Reorganize to maintain grouping
		m.reorganizeComponentsByType()
		
		// Adjust cursor - try to stay in the same position or move to the last item of the same type
		if m.rightCursor >= len(m.selectedComponents) && m.rightCursor > 0 {
			m.rightCursor = len(m.selectedComponents) - 1
		}
		
		// Try to position cursor on a component of the same type
		if len(m.selectedComponents) > 0 {
			// Find the last component of the same type before or at cursor position
			newCursor := -1
			for i := 0; i <= m.rightCursor && i < len(m.selectedComponents); i++ {
				if m.selectedComponents[i].Type == removedType {
					newCursor = i
				}
			}
			if newCursor >= 0 {
				m.rightCursor = newCursor
			}
		}
		
		// Update the usage count locally
		m.updateLocalUsageCount(componentPath, -1)
	}
}

func (m *PipelineBuilderModel) moveComponentUp() {
	if m.rightCursor > 0 && m.rightCursor < len(m.selectedComponents) {
		currentType := m.selectedComponents[m.rightCursor].Type
		previousType := m.selectedComponents[m.rightCursor-1].Type
		
		// Only allow moving within the same type group
		if currentType == previousType {
			// Swap with previous
			m.selectedComponents[m.rightCursor-1], m.selectedComponents[m.rightCursor] = 
				m.selectedComponents[m.rightCursor], m.selectedComponents[m.rightCursor-1]
			
			// Update order numbers
			m.selectedComponents[m.rightCursor-1].Order = m.rightCursor
			m.selectedComponents[m.rightCursor].Order = m.rightCursor + 1
			
			m.rightCursor--
		}
	}
}

func (m *PipelineBuilderModel) moveComponentDown() {
	if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents)-1 {
		currentType := m.selectedComponents[m.rightCursor].Type
		nextType := m.selectedComponents[m.rightCursor+1].Type
		
		// Only allow moving within the same type group
		if currentType == nextType {
			// Swap with next
			m.selectedComponents[m.rightCursor], m.selectedComponents[m.rightCursor+1] = 
				m.selectedComponents[m.rightCursor+1], m.selectedComponents[m.rightCursor]
			
			// Update order numbers
			m.selectedComponents[m.rightCursor].Order = m.rightCursor + 1
			m.selectedComponents[m.rightCursor+1].Order = m.rightCursor + 2
			
			m.rightCursor++
		}
	}
}

func (m *PipelineBuilderModel) updatePreview() {
	if !m.showPreview {
		return
	}

	// Always show the complete pipeline preview
	if len(m.selectedComponents) == 0 {
		m.previewContent = "No components selected yet.\n\nAdd components to see the pipeline preview."
		return
	}

	// Create a temporary pipeline with current components
	tempPipeline := &models.Pipeline{
		Name:       m.pipeline.Name,
		Components: m.selectedComponents,
	}

	// Generate the preview
	output, err := composer.ComposePipeline(tempPipeline)
	if err != nil {
		m.previewContent = fmt.Sprintf("Error generating preview: %v", err)
		return
	}

	m.previewContent = output
}

// sanitizeFileName converts a user-provided name into a safe filename
func sanitizeFileName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	filename := strings.ToLower(name)
	filename = strings.ReplaceAll(filename, " ", "-")
	
	// Remove any characters that aren't alphanumeric or hyphens
	var cleanName strings.Builder
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanName.WriteRune(r)
		}
	}
	
	result := cleanName.String()
	
	// Ensure the filename is not empty
	if result == "" {
		result = "untitled"
	}
	
	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")
	
	// Replace multiple consecutive hyphens with a single hyphen
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	
	return result
}

func (m *PipelineBuilderModel) savePipeline() tea.Cmd {
	return func() tea.Msg {
		// Update pipeline with selected components
		m.pipeline.Components = m.selectedComponents
		
		// Create filename from name using sanitization
		m.pipeline.Path = sanitizeFileName(m.pipeline.Name) + ".yaml"
		
		// Save pipeline
		err := files.WritePipeline(m.pipeline)
		if err != nil {
			m.pipelineSaveMessage = fmt.Sprintf("❌ Failed to save pipeline: %v", err)
			return nil
		}
		
		// Set save message
		m.pipelineSaveMessage = fmt.Sprintf("✓ Pipeline saved: %s", m.pipeline.Path)
		
		// Reload components to update usage stats after save
		m.loadAvailableComponents()
		
		// Cancel any existing timer
		if m.pipelineSaveTimer != nil {
			m.pipelineSaveTimer.Stop()
		}
		
		// Set timer to clear message after 2 seconds
		m.pipelineSaveTimer = time.NewTimer(2 * time.Second)
		
		// Return a command to clear the message after timer
		return func() tea.Msg {
			<-m.pipelineSaveTimer.C
			return clearPipelineSaveMsg{}
		}
	}
}

func (m *PipelineBuilderModel) saveAndSetPipeline() tea.Cmd {
	return func() tea.Msg {
		// Update pipeline with selected components
		m.pipeline.Components = m.selectedComponents
		
		// Create filename from name using sanitization
		m.pipeline.Path = sanitizeFileName(m.pipeline.Name) + ".yaml"
		
		// Save pipeline
		err := files.WritePipeline(m.pipeline)
		if err != nil {
			m.pipelineSaveMessage = fmt.Sprintf("❌ Failed to save pipeline: %v", err)
			return nil
		}
		
		// Generate pipeline output
		output, err := composer.ComposePipeline(m.pipeline)
		if err != nil {
			m.pipelineSaveMessage = fmt.Sprintf("❌ Failed to generate output: %v", err)
			return nil
		}

		// Write to PLUQQY.md
		outputPath := m.pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			m.pipelineSaveMessage = fmt.Sprintf("❌ Failed to write output: %v", err)
			return nil
		}
		
		// Set save message
		m.pipelineSaveMessage = fmt.Sprintf("✓ Saved & Set → %s", outputPath)
		
		// Update preview if showing
		if m.showPreview {
			m.previewContent = output
			m.previewViewport.SetContent(output)
		}
		
		// Reload components to update usage stats after save
		m.loadAvailableComponents()
		
		// Cancel any existing timer
		if m.pipelineSaveTimer != nil {
			m.pipelineSaveTimer.Stop()
		}
		
		// Set timer to clear message after 2 seconds
		m.pipelineSaveTimer = time.NewTimer(2 * time.Second)
		
		// Return a command to clear the message after timer
		return func() tea.Msg {
			<-m.pipelineSaveTimer.C
			return clearPipelineSaveMsg{}
		}
	}
}

func (m *PipelineBuilderModel) nameInputView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(2)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(60)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var s strings.Builder
	s.WriteString(titleStyle.Render("🔧 Create New Pipeline"))
	s.WriteString("\n\n")
	s.WriteString(promptStyle.Render("Enter a descriptive name for your pipeline:"))
	s.WriteString("\n")
	
	// Show input with cursor
	input := m.nameInput
	if input == "" {
		input = " "
	}
	input += "│" // cursor
	
	s.WriteString(inputStyle.Render(input))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press Enter to continue • Press Esc to cancel"))

	// Center the content
	content := s.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m *PipelineBuilderModel) handleComponentCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			m.creationStep = 1
			m.componentContent = ""
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

func (m *PipelineBuilderModel) componentCreationView() string {
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

func (m *PipelineBuilderModel) componentTypeSelectionView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("236")).
		Bold(true).
		Padding(0, 2)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(0, 2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var s strings.Builder
	s.WriteString(titleStyle.Render("🆕 Create New Component"))
	s.WriteString("\n\n")
	s.WriteString("Select component type:\n\n")

	types := []struct {
		name string
		desc string
	}{
		{"CONTEXT", "Background information or system state"},
		{"PROMPT", "Instructions or questions for the LLM"},
		{"RULES", "Important constraints or guidelines"},
	}

	for i, t := range types {
		line := fmt.Sprintf("%s - %s", t.name, t.desc)
		if i == m.typeCursor {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(normalStyle.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: select • Enter: continue • Esc: cancel"))

	content := s.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m *PipelineBuilderModel) componentNameInputView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(2)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(60)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("🆕 Create New %s Component", strings.Title(m.componentCreationType))))
	s.WriteString("\n\n")
	s.WriteString(promptStyle.Render("Enter a descriptive name:"))
	s.WriteString("\n")
	
	input := m.componentName
	if input == "" {
		input = " "
	}
	input += "│"
	
	s.WriteString(inputStyle.Render(input))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Enter: continue • Esc: back"))

	content := s.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m *PipelineBuilderModel) componentContentEditView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))

	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1).
		Width(m.width - 4).
		Height(m.height - 10)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("📝 %s: %s", strings.Title(m.componentCreationType), m.componentName)))
	s.WriteString("\n\n")
	
	// Show content with cursor
	content := m.componentContent
	if content == "" {
		content = " "
	}
	content += "│"
	
	s.WriteString(editorStyle.Render(content))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Type to edit • ctrl+s: save • esc: back"))

	return s.String()
}

func (m *PipelineBuilderModel) saveNewComponent() tea.Cmd {
	return func() tea.Msg {
		// Create filename from name using sanitization
		filename := sanitizeFileName(m.componentName) + ".md"
		
		// Determine directory
		var dir string
		switch m.componentCreationType {
		case models.ComponentTypeContext:
			dir = filepath.Join(files.ComponentsDir, files.ContextsDir)
		case models.ComponentTypePrompt:
			dir = filepath.Join(files.ComponentsDir, files.PromptsDir)
		case models.ComponentTypeRules:
			dir = filepath.Join(files.ComponentsDir, files.RulesDir)
		}
		
		path := filepath.Join(dir, filename)
		
		// Write component
		err := files.WriteComponent(path, m.componentContent)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save component '%s': %v", m.componentName, err))
		}
		
		// Reset creation state
		m.creatingComponent = false
		m.componentName = ""
		m.componentContent = ""
		m.creationStep = 0
		
		// Reload components
		m.loadAvailableComponents()
		
		return StatusMsg(fmt.Sprintf("✓ Created %s: %s", m.componentCreationType, filename))
	}
}

func (m *PipelineBuilderModel) editComponent() tea.Cmd {
	if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents) {
		comp := m.selectedComponents[m.rightCursor]
		componentPath := filepath.Join(files.PipelinesDir, comp.Path)
		componentPath = filepath.Clean(componentPath)
		
		return m.openInEditor(componentPath)
	}
	return nil
}

func (m *PipelineBuilderModel) editComponentFromLeft() tea.Cmd {
	components := m.getAllAvailableComponents()
	if m.leftCursor >= 0 && m.leftCursor < len(components) {
		comp := components[m.leftCursor]
		return m.openInEditor(comp.path)
	}
	return nil
}


func (m *PipelineBuilderModel) openInEditor(path string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set. Please set it to your preferred editor.")
		}

		// Validate editor path to prevent command injection
		if strings.ContainsAny(editor, "&|;<>()$`\\\"'") {
			return StatusMsg("Invalid EDITOR value: contains shell metacharacters")
		}

		fullPath := filepath.Join(files.PluqqyDir, path)
		cmd := exec.Command(editor, fullPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}

		// Reload components after editing
		m.loadAvailableComponents()
		m.updatePreview()

		return StatusMsg(fmt.Sprintf("Edited: %s", filepath.Base(path)))
	}
}

func (m *PipelineBuilderModel) handleComponentEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		// Save component but don't exit yet
		return m, m.saveEditedComponent()
	case "esc":
		// Cancel editing
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.editSaveMessage = ""
		if m.editSaveTimer != nil {
			m.editSaveTimer.Stop()
		}
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

func (m *PipelineBuilderModel) componentEditView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))

	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1).
		Width(m.width - 4).
		Height(m.height - 12) // Make room for save message

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	saveMessageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")). // Green
		Bold(true).
		MarginTop(1)

	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("📝 Editing: %s", m.editingComponentName)))
	s.WriteString("\n\n")
	
	// Show content with cursor
	content := m.componentContent
	if content == "" {
		content = " "
	}
	content += "│"
	
	s.WriteString(editorStyle.Render(content))
	s.WriteString("\n")
	
	// Show save message if present
	if m.editSaveMessage != "" {
		s.WriteString(saveMessageStyle.Render(m.editSaveMessage))
		s.WriteString("\n")
	}
	
	s.WriteString(helpStyle.Render("Type to edit • ctrl+s: save • esc: cancel"))

	return s.String()
}

func (m *PipelineBuilderModel) saveEditedComponent() tea.Cmd {
	return func() tea.Msg {
		// Write component
		err := files.WriteComponent(m.editingComponentPath, m.componentContent)
		if err != nil {
			m.editSaveMessage = fmt.Sprintf("❌ Failed to save: %v", err)
			return nil
		}
		
		// Set save message
		m.editSaveMessage = fmt.Sprintf("✓ Saved: %s", m.editingComponentName)
		
		// Cancel any existing timer
		if m.editSaveTimer != nil {
			m.editSaveTimer.Stop()
		}
		
		// Set timer to clear message and exit after 1.5 seconds
		m.editSaveTimer = time.NewTimer(1500 * time.Millisecond)
		
		// Reload components
		m.loadAvailableComponents()
		
		// Update preview
		m.updatePreview()
		
		// Return a command to clear the message after timer
		return func() tea.Msg {
			<-m.editSaveTimer.C
			return clearEditSaveMsg{}
		}
	}
}