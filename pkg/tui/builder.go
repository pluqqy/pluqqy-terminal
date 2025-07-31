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
)

type column int

const (
	leftColumn column = iota
	rightColumn
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
}

type componentItem struct {
	name     string
	path     string
	compType string
}

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
	}
	m.loadAvailableComponents()
	return m
}

func (m *PipelineBuilderModel) loadAvailableComponents() {
	// Load prompts
	prompts, _ := files.ListComponents("prompts")
	for _, p := range prompts {
		m.prompts = append(m.prompts, componentItem{
			name:     p,
			path:     filepath.Join(files.ComponentsDir, files.PromptsDir, p),
			compType: models.ComponentTypePrompt,
		})
	}

	// Load contexts
	contexts, _ := files.ListComponents("contexts")
	for _, c := range contexts {
		m.contexts = append(m.contexts, componentItem{
			name:     c,
			path:     filepath.Join(files.ComponentsDir, files.ContextsDir, c),
			compType: models.ComponentTypeContext,
		})
	}

	// Load rules
	rules, _ := files.ListComponents("rules")
	for _, r := range rules {
		m.rules = append(m.rules, componentItem{
			name:     r,
			path:     filepath.Join(files.ComponentsDir, files.RulesDir, r),
			compType: models.ComponentTypeRules,
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
		// Update viewport size
		if m.showPreview {
			// Calculate preview height
			previewHeight := m.height / 2 - 5
			if previewHeight < 5 {
				previewHeight = 5
			}
			m.previewViewport.Width = m.width - 4
			m.previewViewport.Height = previewHeight
		}

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
		if m.showPreview && m.previewContent != "" {
			// Check if we should handle viewport scrolling
			switch msg.String() {
			case "pgup", "pgdown":
				m.previewViewport, cmd = m.previewViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Normal mode keybindings
		switch msg.String() {
		case "esc":
			// Return to main list
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "tab":
			// Switch between columns
			if m.activeColumn == leftColumn {
				m.activeColumn = rightColumn
			} else {
				m.activeColumn = leftColumn
			}

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

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
			if m.showPreview {
				// Recalculate viewport size when toggling preview
				previewHeight := m.height / 2 - 5
				if previewHeight < 5 {
					previewHeight = 5
				}
				m.previewViewport.Width = m.width - 4
				m.previewViewport.Height = previewHeight
			}

		case "s", "ctrl+s":
			// Save pipeline
			return m, m.savePipeline()

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
	columnWidth := (m.width - 3) / 2 // -3 for borders and gap
	contentHeight := m.height - 10    // Reserve space for title and help

	if m.showPreview {
		contentHeight = contentHeight / 2
	}

	// Build left column (available components)
	var leftContent strings.Builder
	leftContent.WriteString(typeHeaderStyle.Render("Available Components") + "\n\n")

	allComponents := m.getAllAvailableComponents()
	currentType := ""
	
	for i, comp := range allComponents {
		if comp.compType != currentType {
			if currentType != "" {
				leftContent.WriteString("\n")
			}
			currentType = comp.compType
			leftContent.WriteString(typeHeaderStyle.Render(fmt.Sprintf("▸ %s", strings.Title(currentType))) + "\n")
		}

		cursor := "  "
		if m.activeColumn == leftColumn && i == m.leftCursor {
			cursor = "▸ "
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

		// Build the line with indicator
		line := comp.name
		if isAdded {
			line = fmt.Sprintf("%s ✓", comp.name)
		}
		line = fmt.Sprintf("%s%s", cursor, line)
		
		// Apply styling
		if m.activeColumn == leftColumn && i == m.leftCursor {
			leftContent.WriteString(selectedStyle.Render(line) + "\n")
		} else if isAdded {
			// Use a dimmed style for already added components
			addedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			leftContent.WriteString(addedStyle.Render(line) + "\n")
		} else {
			leftContent.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	// Build right column (selected components)
	var rightContent strings.Builder
	rightContent.WriteString(typeHeaderStyle.Render("Pipeline Components") + "\n\n")

	if len(m.selectedComponents) == 0 {
		rightContent.WriteString(normalStyle.Render("No components selected\n\nPress Tab to switch columns\nPress Enter to add components"))
	} else {
		for i, comp := range m.selectedComponents {
			cursor := "  "
			if m.activeColumn == rightColumn && i == m.rightCursor {
				cursor = "▸ "
			}

			line := fmt.Sprintf("%s%d. [%s] %s", cursor, i+1, comp.Type, filepath.Base(comp.Path))
			if m.activeColumn == rightColumn && i == m.rightCursor {
				rightContent.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				rightContent.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}

	// Apply borders
	leftStyle := inactiveStyle
	rightStyle := inactiveStyle
	if m.activeColumn == leftColumn {
		leftStyle = activeStyle
	} else {
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
	s.WriteString(columns)

	// Add preview if enabled
	if m.showPreview && m.previewContent != "" {
		previewBorderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("243")).
			Width(m.width - 2)

		s.WriteString("\n\n")
		s.WriteString(typeHeaderStyle.Render("Pipeline Preview (PLUQQY.md)"))
		s.WriteString("\n")
		
		// Use viewport for scrollable preview
		previewView := m.previewViewport.View()
		s.WriteString(previewBorderStyle.Render(previewView))
		
		// Add scroll indicator
		if m.previewViewport.TotalLineCount() > m.previewViewport.Height {
			scrollHelp := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render("PgUp/PgDn: scroll preview")
			s.WriteString("\n")
			s.WriteString(scrollHelp)
		}
	}

	// Help text
	help := []string{
		"Tab: switch",
		"↑/↓: nav",
		"Enter: add/edit",
		"n: new",
		"E: edit external",
		"e: edit TUI",
		"Del: remove",
		"K/J: reorder",
		"p: preview",
		"Ctrl+S: save",
		"Esc: back",
		"Ctrl+C: quit",
	}
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)
	
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(strings.Join(help, " • ")))

	return s.String()
}

func (m *PipelineBuilderModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	
	// Update viewport size
	if m.showPreview {
		previewHeight := m.height / 2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		m.previewViewport.Width = m.width - 4
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
			Order: len(m.selectedComponents) + 1,
		}
		
		m.selectedComponents = append(m.selectedComponents, ref)
	}
}

func (m *PipelineBuilderModel) removeSelectedComponent() {
	if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents) {
		// Remove component
		m.selectedComponents = append(
			m.selectedComponents[:m.rightCursor],
			m.selectedComponents[m.rightCursor+1:]...,
		)
		
		// Update order numbers
		for i := range m.selectedComponents {
			m.selectedComponents[i].Order = i + 1
		}
		
		// Adjust cursor
		if m.rightCursor >= len(m.selectedComponents) && m.rightCursor > 0 {
			m.rightCursor--
		}
	}
}

func (m *PipelineBuilderModel) moveComponentUp() {
	if m.rightCursor > 0 && m.rightCursor < len(m.selectedComponents) {
		// Swap with previous
		m.selectedComponents[m.rightCursor-1], m.selectedComponents[m.rightCursor] = 
			m.selectedComponents[m.rightCursor], m.selectedComponents[m.rightCursor-1]
		
		// Update order numbers
		m.selectedComponents[m.rightCursor-1].Order = m.rightCursor
		m.selectedComponents[m.rightCursor].Order = m.rightCursor + 1
		
		m.rightCursor--
	}
}

func (m *PipelineBuilderModel) moveComponentDown() {
	if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents)-1 {
		// Swap with next
		m.selectedComponents[m.rightCursor], m.selectedComponents[m.rightCursor+1] = 
			m.selectedComponents[m.rightCursor+1], m.selectedComponents[m.rightCursor]
		
		// Update order numbers
		m.selectedComponents[m.rightCursor].Order = m.rightCursor + 1
		m.selectedComponents[m.rightCursor+1].Order = m.rightCursor + 2
		
		m.rightCursor++
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
			return StatusMsg(fmt.Sprintf("Failed to save pipeline '%s': %v", m.pipeline.Name, err))
		}
		
		// Return status message first, then switch view
		return tea.Batch(
			func() tea.Msg {
				return StatusMsg(fmt.Sprintf("✓ Pipeline saved: %s.yaml", m.pipeline.Path))
			},
			func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			},
		)
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
	s.WriteString(helpStyle.Render("Type to edit • Ctrl+S: save • Esc: back"))

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
			editor = "vi"
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
		// Save component
		return m, m.saveEditedComponent()
	case "esc":
		// Cancel editing
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
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
		Height(m.height - 10)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
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
	s.WriteString(helpStyle.Render("Type to edit • Ctrl+S: save • Esc: cancel"))

	return s.String()
}

func (m *PipelineBuilderModel) saveEditedComponent() tea.Cmd {
	return func() tea.Msg {
		// Save the name before resetting
		componentName := m.editingComponentName
		
		// Write component
		err := files.WriteComponent(m.editingComponentPath, m.componentContent)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save component: %v", err))
		}
		
		// Reset editing state
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		
		// Reload components
		m.loadAvailableComponents()
		
		// Update preview
		m.updatePreview()
		
		return StatusMsg(fmt.Sprintf("✓ Saved: %s", componentName))
	}
}