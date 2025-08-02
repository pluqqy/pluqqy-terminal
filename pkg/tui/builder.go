package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type column int

const (
	searchColumn column = iota
	leftColumn
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
	
	// Tag editing state
	editingTags           bool
	editingTagsPath       string
	currentTags           []string
	tagInput              string
	tagCursor             int
	availableTags         []string
	showTagSuggestions    bool
	tagCloudActive        bool
	tagCloudCursor        int
	tagDeleteConfirm      *ConfirmationModel
	deletingTag           string
	deletingTagUsage      []string
	editSaveTimer         *time.Timer
	editViewport          viewport.Model
	
	// Exit confirmation
	exitConfirm          *ConfirmationModel
	exitConfirmationType string // "pipeline" or "component"
	
	// Change tracking
	originalComponents    []models.ComponentRef // Original components for existing pipelines
	originalContent       string                // Original content for component editing
	
	// Search state
	searchInput           textinput.Model
	searchQuery           string
	searchEngine          *search.Engine
	filteredPrompts       []componentItem
	filteredContexts      []componentItem
	filteredRules         []componentItem
}

type componentItem struct {
	name         string
	path         string
	compType     string
	lastModified time.Time
	usageCount   int
	tokenCount   int
	tags         []string
}

type clearEditSaveMsg struct{}

func NewPipelineBuilderModel() *PipelineBuilderModel {
	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = ""
	ti.CharLimit = 100
	ti.Width = 40
	
	m := &PipelineBuilderModel{
		activeColumn: leftColumn,
		showPreview:  true, // Show preview by default in builder for immediate feedback
		editingName:  true,
		nameInput:    "",
		pipeline: &models.Pipeline{
			Name:       "",
			Components: []models.ComponentRef{},
		},
		previewViewport:  viewport.New(80, 20), // Default size, will be resized
		leftViewport:     viewport.New(40, 20), // Default size, will be resized
		rightViewport:    viewport.New(40, 20), // Default size, will be resized
		searchInput:      ti,
		exitConfirm:      NewConfirmation(),
		tagDeleteConfirm: NewConfirmation(),
	}
	
	// Initialize search engine
	m.searchEngine = search.NewEngine()
	
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
		
		// Read component content for token estimation
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
		}
		
		tags := []string{}
		if component != nil {
			tags = component.Tags
		}
		
		m.prompts = append(m.prompts, componentItem{
			name:         p,
			path:         componentPath,
			compType:     models.ComponentTypePrompt,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
			tags:         tags,
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
		
		tags := []string{}
		if component != nil {
			tags = component.Tags
		}
		
		m.contexts = append(m.contexts, componentItem{
			name:         c,
			path:         componentPath,
			compType:     models.ComponentTypeContext,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
			tags:         tags,
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
		
		tags := []string{}
		if component != nil {
			tags = component.Tags
		}
		
		m.rules = append(m.rules, componentItem{
			name:         r,
			path:         componentPath,
			compType:     models.ComponentTypeRules,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
			tags:         tags,
		})
	}
	
	// Initialize filtered lists with all components
	m.filteredPrompts = m.prompts
	m.filteredContexts = m.contexts
	m.filteredRules = m.rules
	
	// Rebuild search engine index
	if m.searchEngine != nil {
		m.searchEngine.BuildIndex()
	}
}

func (m *PipelineBuilderModel) Init() tea.Cmd {
	return nil
}

func (m *PipelineBuilderModel) performSearch() {
	if m.searchQuery == "" {
		// No search query, show all items
		m.filteredPrompts = m.prompts
		m.filteredContexts = m.contexts
		m.filteredRules = m.rules
		return
	}
	
	// Clear filtered lists
	m.filteredPrompts = nil
	m.filteredContexts = nil
	m.filteredRules = nil
	
	// Use search engine to find matching items
	if m.searchEngine != nil {
		results, err := m.searchEngine.Search(m.searchQuery)
		if err != nil {
			// On error, show all items
			m.filteredPrompts = m.prompts
			m.filteredContexts = m.contexts
			m.filteredRules = m.rules
			return
		}
		
		// Create maps for quick lookup
		resultMap := make(map[string]bool)
		for _, result := range results {
			resultMap[result.Item.Path] = true
		}
		
		// Filter each component list
		for _, comp := range m.prompts {
			if resultMap[comp.path] {
				m.filteredPrompts = append(m.filteredPrompts, comp)
			}
		}
		for _, comp := range m.contexts {
			if resultMap[comp.path] {
				m.filteredContexts = append(m.filteredContexts, comp)
			}
		}
		for _, comp := range m.rules {
			if resultMap[comp.path] {
				m.filteredRules = append(m.filteredRules, comp)
			}
		}
	} else {
		// No search engine, show all items
		m.filteredPrompts = m.prompts
		m.filteredContexts = m.contexts
		m.filteredRules = m.rules
	}
	
	// Reset cursor if it's out of bounds
	if m.activeColumn == leftColumn {
		totalItems := len(m.filteredPrompts) + len(m.filteredContexts) + len(m.filteredRules)
		if m.leftCursor >= totalItems {
			m.leftCursor = 0
		}
	}
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
			// Force a redraw to ensure layout is recalculated
			return m, tea.ClearScreen
		}
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.originalContent = ""
		// Force a redraw to ensure layout is recalculated
		return m, tea.ClearScreen


	case tea.KeyMsg:
		// Handle exit confirmation
		if m.exitConfirm.Active() {
			return m, m.exitConfirm.Update(msg)
		}
		
		// Handle component creation mode
		if m.creatingComponent {
			return m.handleComponentCreation(msg)
		}
		
		// Handle component editing mode
		if m.editingComponent {
			return m.handleComponentEditing(msg)
		}
		
		// Handle tag editing mode
		if m.editingTags {
			return m.handleTagEditing(msg)
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
		
		// Handle search input when search column is active
		if m.activeColumn == searchColumn && !m.editingName && !m.creatingComponent && !m.editingComponent {
			// Handle special keys in search first
			switch msg.String() {
			case "esc":
				// Clear search and switch to left column
				m.searchInput.SetValue("")
				m.searchQuery = ""
				m.performSearch()
				m.activeColumn = leftColumn
				m.searchInput.Blur()
				return m, nil
			case "tab":
				// Let tab be handled by the main navigation logic
				// Don't process it here
			default:
				// For all other keys, update the search input
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				
				// Check if search query changed
				if m.searchQuery != m.searchInput.Value() {
					m.searchQuery = m.searchInput.Value()
					m.performSearch()
				}
				
				return m, cmd
			}
		}

		// If preview is showing and active, handle viewport navigation

		// Normal mode keybindings
		switch msg.String() {
		case "esc":
			// Check if there are unsaved changes
			if m.hasUnsavedChanges() {
				// Show confirmation dialog
				m.exitConfirmationType = "pipeline"
				m.exitConfirm.ShowDialog(
					"⚠️  Unsaved Changes",
					"You have unsaved changes in this pipeline.",
					"Exit without saving?",
					true, // destructive
					m.width - 4,
					10,
					func() tea.Cmd {
						return func() tea.Msg {
							return SwitchViewMsg{view: mainListView}
						}
					},
					nil, // onCancel
				)
				return m, nil
			}
			// No unsaved changes, exit immediately
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}

		case "tab":
			// Switch between columns
			if m.showPreview {
				// When preview is shown, cycle through all panes
				switch m.activeColumn {
				case searchColumn:
					m.activeColumn = leftColumn
					m.searchInput.Blur()
				case leftColumn:
					m.activeColumn = rightColumn
					// Reset right cursor and viewport when entering right column
					m.rightCursor = 0
					m.rightViewport.GotoTop()
				case rightColumn:
					m.activeColumn = previewColumn
				case previewColumn:
					m.activeColumn = searchColumn
					m.searchInput.Focus()
				}
			} else {
				// When preview is hidden, cycle through search, left and right
				switch m.activeColumn {
				case searchColumn:
					m.activeColumn = leftColumn
					m.searchInput.Blur()
				case leftColumn:
					m.activeColumn = rightColumn
					// Reset right cursor and viewport when entering right column
					m.rightCursor = 0
					m.rightViewport.GotoTop()
				case rightColumn:
					m.activeColumn = searchColumn
					m.searchInput.Focus()
				}
			}
			// Update preview when switching to non-preview column
			if m.activeColumn != previewColumn {
				m.updatePreview()
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
			} else if m.activeColumn == rightColumn {
				// Page up in right column - move cursor up by viewport height
				pageSize := m.rightViewport.Height
				for i := 0; i < pageSize && m.rightCursor > 0; i++ {
					m.rightCursor--
				}
			}
			
		case "pgdown":
			if m.activeColumn == previewColumn {
				// Scroll preview page down
				m.previewViewport.ViewDown()
			} else if m.activeColumn == rightColumn {
				// Page down in right column - move cursor down by viewport height
				pageSize := m.rightViewport.Height
				maxCursor := len(m.selectedComponents) - 1
				for i := 0; i < pageSize && m.rightCursor < maxCursor; i++ {
					m.rightCursor++
				}
			}
			
		case "home":
			if m.activeColumn == leftColumn {
				m.leftCursor = 0
			} else if m.activeColumn == rightColumn {
				m.rightCursor = 0
			}
			
		case "end":
			if m.activeColumn == leftColumn {
				components := m.getAllAvailableComponents()
				if len(components) > 0 {
					m.leftCursor = len(components) - 1
				}
			} else if m.activeColumn == rightColumn {
				if len(m.selectedComponents) > 0 {
					m.rightCursor = len(m.selectedComponents) - 1
				}
			}

		case "enter":
			if m.activeColumn == leftColumn {
				m.addSelectedComponent()
			} else if m.activeColumn == rightColumn && len(m.selectedComponents) > 0 {
				// Remove selected component in right column (same as delete)
				m.removeSelectedComponent()
			}

		case "delete", "backspace", "d":
			if m.activeColumn == rightColumn {
				m.removeSelectedComponent()
			}

		case "p":
			m.showPreview = !m.showPreview
			m.updateViewportSizes()
			if m.showPreview {
				m.updatePreview()
			}
		case "t":
			// Edit tags - context aware based on active column
			if m.activeColumn == leftColumn {
				// Edit component tags
				components := m.getAllAvailableComponents()
				if m.leftCursor >= 0 && m.leftCursor < len(components) {
					comp := components[m.leftCursor]
					m.startTagEditing(comp.path, comp.tags)
				}
			} else if m.activeColumn == rightColumn {
				// Edit pipeline tags
				if m.pipeline != nil {
					m.startPipelineTagEditing(m.pipeline.Tags)
				}
			}
		case "/":
			// Jump to search
			m.activeColumn = searchColumn
			m.searchInput.Focus()
			return m, nil

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
			} else if m.activeColumn == rightColumn && len(m.selectedComponents) > 0 {
				// Edit selected component in external editor from right column
				return m, m.editComponent()
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
					m.originalContent = content.Content // Store original for change detection
					
					// Initialize edit viewport
					m.editViewport = viewport.New(m.width-8, m.height-10)
					m.editViewport.SetContent(content.Content)
					
					return m, nil
				}
			} else if m.activeColumn == rightColumn && len(m.selectedComponents) > 0 {
				// Edit component in TUI editor from right column
				if m.rightCursor >= 0 && m.rightCursor < len(m.selectedComponents) {
					selected := m.selectedComponents[m.rightCursor]
					// Convert path from relative to component path
					componentPath := strings.TrimPrefix(selected.Path, "../")
					
					// Read the component content
					content, err := files.ReadComponent(componentPath)
					if err != nil {
						m.err = err
						return m, nil
					}
					
					// Extract component name from path
					parts := strings.Split(componentPath, "/")
					componentName := ""
					if len(parts) >= 2 {
						componentName = strings.TrimSuffix(parts[1], ".md")
					}
					
					// Enter editing mode
					m.editingComponent = true
					m.editingComponentPath = componentPath
					m.editingComponentName = componentName
					m.componentContent = content.Content
					m.originalContent = content.Content // Store original for change detection
					
					// Initialize edit viewport
					m.editViewport = viewport.New(m.width-8, m.height-10)
					m.editViewport.SetContent(content.Content)
					
					return m, nil
				}
			}
		}
	}

	// Update preview if needed
	m.updatePreview()

	// Update viewport content if preview changed
	if m.showPreview && m.previewContent != "" {
		// Preprocess content to handle carriage returns and ensure proper line breaks
		processedContent := strings.ReplaceAll(m.previewContent, "\r\r", "\n\n")
		processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
		// Wrap content to viewport width to prevent overflow
		wrappedContent := wordwrap.String(processedContent, m.previewViewport.Width)
		m.previewViewport.SetContent(wrappedContent)
	}
	
	// Only forward non-key messages to viewports
	// Key messages are already handled above
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
		
		m.leftViewport, cmd = m.leftViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.rightViewport, cmd = m.rightViewport.Update(msg)
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

	// If showing exit confirmation, display dialog
	if m.exitConfirm.Active() {
		return m.exitConfirm.View()
	}

	// If creating component, show creation wizard
	if m.creatingComponent {
		return m.componentCreationView()
	}
	
	// If editing component, show edit view
	if m.editingComponent {
		return m.componentEditView()
	}
	
	// If editing tags, show tag edit view
	if m.editingTags {
		return m.tagEditView()
	}
	
	// If editing name, show name input screen
	if m.editingName {
		return m.nameInputView()
	}

	// Styles

	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("28")).
		Background(lipgloss.Color("236")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	typeHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3              // Height for search bar
	
	// Height calculation matching Main List View:
	// Base reservation: 14 lines (header, help pane, spacing) + search bar
	contentHeight := m.height - 14 - searchBarHeight

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
		Foreground(lipgloss.Color("241"))
	
	// Table column widths (adjusted for left column width)
	nameWidth := 22
	tagsWidth := 18
	tokenWidth := 8  // For "~Tokens" plus padding
	usageWidth := 10
	
	// Render table header with 2-space shift - extra spaces between columns
	header := fmt.Sprintf("  %-*s %-*s  %*s  %*s", 
		nameWidth, "Name",
		tagsWidth, "Tags",
		tokenWidth, "~Tokens",
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
		
		
		// Format usage count
		usageStr := fmt.Sprintf("%d", comp.usageCount)
		
		// Format token count - right-aligned with consistent width
		tokenStr := fmt.Sprintf("%d", comp.tokenCount)
		
		// Format tags
		tagsStr := renderTagChipsWithWidth(comp.tags, tagsWidth, 2) // Show max 2 tags inline
		
		// Build the row components separately
		namePart := fmt.Sprintf("%-*s", nameWidth, nameStr)
		
		// For tags, we need to pad based on rendered width
		tagsPadding := tagsWidth - lipgloss.Width(tagsStr)
		if tagsPadding < 0 {
			tagsPadding = 0
		}
		tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)
		
		tokenPart := fmt.Sprintf("%*s", tokenWidth, tokenStr)
		
		// Join all parts - extra space between tags and tokens, and between tokens and usage
		row := fmt.Sprintf("%s %s  %s  %*s",
			namePart,
			tagsPart,
			tokenPart,
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
	// Wrap content to viewport width to prevent overflow
	wrappedLeftContent := wordwrap.String(leftScrollContent.String(), m.leftViewport.Width)
	m.leftViewport.SetContent(wrappedLeftContent)
	
	// Update viewport to follow cursor
	if m.activeColumn == leftColumn && len(allComponents) > 0 {
		// Calculate the line position of the cursor
		currentLine := 0
		for i := 0; i < m.leftCursor && i < len(allComponents); i++ {
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
		if currentLine < m.leftViewport.YOffset {
			m.leftViewport.SetYOffset(currentLine)
		} else if currentLine >= m.leftViewport.YOffset+m.leftViewport.Height {
			m.leftViewport.SetYOffset(currentLine - m.leftViewport.Height + 1)
		}
	}
	
	// Add padding to viewport content
	leftViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	leftContent.WriteString(leftViewportPadding.Render(m.leftViewport.View()))

	// Build right column (selected components)
	var rightContent strings.Builder
	// Create heading with colons spanning the width
	rightHeading := "PIPELINE COMPONENTS"
	
	// Calculate remaining width for colons (just for heading now)
	rightRemainingWidth := columnWidth - len(rightHeading) - 5 // -5 for space and padding
	if rightRemainingWidth < 0 {
		rightRemainingWidth = 0
	}
	
	// Render heading without tags
	rightContent.WriteString(headerPadding.Render(
		typeHeaderStyle.Render(rightHeading) + " " + colonStyle.Render(strings.Repeat(":", rightRemainingWidth))))
	rightContent.WriteString("\n")
	
	// Always render tag row (even if empty) for consistent layout
	tagRowStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		PaddingTop(1).    // Add top margin
		PaddingBottom(1). // Add bottom margin
		Height(3) // Total height including padding
	
	if m.pipeline != nil && len(m.pipeline.Tags) > 0 {
		// Render tags with more space available (full column width minus padding)
		tagsStr := renderTagChipsWithWidth(m.pipeline.Tags, columnWidth-4, 5) // Show more tags with available space
		rightContent.WriteString(tagRowStyle.Render(tagsStr))
	} else {
		// Empty row to maintain layout
		rightContent.WriteString(tagRowStyle.Render(" "))
	}
	rightContent.WriteString("\n")
	
	// Build scrollable content for right viewport
	var rightScrollContent strings.Builder

	if len(m.selectedComponents) == 0 {
		rightScrollContent.WriteString(normalStyle.Render("No components selected\n\nPress Tab to switch columns\nPress Enter to add components"))
	} else {
		// Load settings for section order
		settings, err := files.ReadSettings()
		if err != nil || settings == nil {
			settings = models.DefaultSettings()
		}
		
		// Group components by type
		typeGroups := make(map[string][]models.ComponentRef)
		for _, comp := range m.selectedComponents {
			typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
		}
		
		// Track overall index for cursor position
		overallIndex := 0
		remainingSections := 0
		
		// Count how many sections we'll actually display
		for _, section := range settings.Output.Formatting.Sections {
			if len(typeGroups[section.Type]) > 0 {
				remainingSections++
			}
		}
		
		// Render sections in the configured order
		for _, section := range settings.Output.Formatting.Sections {
			components, exists := typeGroups[section.Type]
			if !exists || len(components) == 0 {
				continue
			}
			
			// Get the display name for this section type
			var sectionHeader string
			switch section.Type {
			case models.ComponentTypeContext:
				sectionHeader = "CONTEXTS"
			case models.ComponentTypePrompt:
				sectionHeader = "PROMPTS"
			case models.ComponentTypeRules:
				sectionHeader = "RULES"
			default:
				sectionHeader = strings.ToUpper(section.Type)
			}
			
			rightScrollContent.WriteString(typeHeaderStyle.Render("▸ " + sectionHeader) + "\n")
			
			for _, comp := range components {
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
			
			// Add spacing between sections (but not after the last one)
			remainingSections--
			if remainingSections > 0 {
				rightScrollContent.WriteString("\n")
			}
		}
	}
	
	// Update right viewport with content
	// Wrap content to viewport width to prevent overflow
	wrappedRightContent := wordwrap.String(rightScrollContent.String(), m.rightViewport.Width)
	m.rightViewport.SetContent(wrappedRightContent)
	
	// Update viewport to follow cursor (even when right column is not active)
	if len(m.selectedComponents) > 0 {
		// Load settings for section order
		settings, err := files.ReadSettings()
		if err != nil || settings == nil {
			settings = models.DefaultSettings()
		}
		
		// Calculate the line position of the cursor
		currentLine := 0
		overallIndex := 0
		
		// Group components by type
		typeGroups := make(map[string][]models.ComponentRef)
		for _, comp := range m.selectedComponents {
			typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
		}
		
		// Count lines up to cursor position following section order
		for sectionIdx, section := range settings.Output.Formatting.Sections {
			components, exists := typeGroups[section.Type]
			if !exists || len(components) == 0 {
				continue
			}
			
			currentLine++ // Section header
			
			for range components {
				if overallIndex == m.rightCursor {
					break
				}
				currentLine++
				overallIndex++
			}
			
			// Check if we found the cursor
			if overallIndex >= m.rightCursor {
				break
			}
			
			// Add empty line if there are more sections
			hasMoreSections := false
			for j := sectionIdx + 1; j < len(settings.Output.Formatting.Sections); j++ {
				if len(typeGroups[settings.Output.Formatting.Sections[j].Type]) > 0 {
					hasMoreSections = true
					break
				}
			}
			if hasMoreSections {
				currentLine++ // Empty line between sections
			}
		}
		
		// Ensure the cursor line is visible
		if currentLine < m.rightViewport.YOffset {
			m.rightViewport.SetYOffset(currentLine)
		} else if currentLine >= m.rightViewport.YOffset+m.rightViewport.Height {
			m.rightViewport.SetYOffset(currentLine - m.rightViewport.Height + 1)
		}
	}
	
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

	leftColumnView := leftStyle.
		Width(columnWidth).
		Height(contentHeight).
		Render(leftContent.String())

	rightColumnView := rightStyle.
		Width(columnWidth).
		Height(contentHeight).
		Render(rightContent.String())

	// Join columns
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumnView, " ", rightColumnView)

	// Build final view
	var s strings.Builder
	
	// Add search bar at the top
	searchBarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if m.activeColumn == searchColumn {
				return "170"
			}
			return "240"
		}())).
		Width(m.width - 4).
		Padding(0, 1)
	
	// Create search icon with larger font
	searchIconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true)
	
	searchIcon := searchIconStyle.Render("⌕ ")
	searchContent := lipgloss.JoinHorizontal(lipgloss.Center, searchIcon, m.searchInput.View())
	searchBar := searchBarStyle.Render(searchContent)
	s.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(searchBar))
	s.WriteString("\n")
	
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
		var previewHeading string
		if m.activeColumn == leftColumn {
			previewHeading = "COMPONENT PREVIEW"
		} else {
			pipelineName := "PLUQQY.md"
			if m.pipeline != nil && m.pipeline.Name != "" {
				pipelineName = m.pipeline.Name + ".yaml"
			}
			previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
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

	// Help text in bordered pane
	var help []string
	if m.activeColumn == searchColumn {
		// Show search syntax help when search is active
		help = []string{
			"esc exit search",
			"enter search",
			"tag:<name>",
			"type:<type>",
			"pipeline:<name>",
			"<keyword>",
			"combine with spaces",
		}
	} else {
		// Show normal navigation help
		help = []string{
			"/ search",
			"tab switch pane",
			"↑/↓ nav",
			"enter add/remove",
			"n new",
			"e edit",
			"E edit external",
			"t tag",
			"del remove",
			"K/J reorder",
			"p preview",
			"ctrl+s save",
			"S save+set",
			"esc back",
			"ctrl+c quit",
		}
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

func (m *PipelineBuilderModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update search input width to match container
	m.searchInput.Width = width - 10 // Account for borders, padding, and icon
	m.updateViewportSizes()
}

func (m *PipelineBuilderModel) hasUnsavedChanges() bool {
	// For new pipelines, check if components have been added
	if m.pipeline.Path == "" {
		return len(m.selectedComponents) > 0
	}
	
	// For existing pipelines, check if components have changed
	if len(m.selectedComponents) != len(m.originalComponents) {
		return true
	}
	
	// Check if components are the same (order matters)
	for i := range m.selectedComponents {
		if m.selectedComponents[i].Path != m.originalComponents[i].Path {
			return true
		}
	}
	
	return false
}

func (m *PipelineBuilderModel) updateViewportSizes() {
	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3              // Height for search bar
	contentHeight := m.height - 15 - searchBarHeight    // Reserve space for search bar, help pane, status bar, and spacing
	
	if m.showPreview {
		contentHeight = contentHeight / 2
	}
	
	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}
	
	// Update left and right viewports for table content
	// Left column: heading (1) + empty (1) + table header (1) + empty (1) = 4 lines
	leftViewportHeight := contentHeight - 4
	if leftViewportHeight < 5 {
		leftViewportHeight = 5
	}
	
	// Right column: heading (1) + tag row with padding (3) + empty (1) = 5 lines
	rightViewportHeight := contentHeight - 5
	if rightViewportHeight < 5 {
		rightViewportHeight = 5
	}
	
	m.leftViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.leftViewport.Height = leftViewportHeight
	m.rightViewport.Width = columnWidth - 4  // Account for borders (2) and padding (2)
	m.rightViewport.Height = rightViewportHeight
	
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
		
		// Store original components for change detection
		m.originalComponents = make([]models.ComponentRef, len(p.Components))
		copy(m.originalComponents, p.Components)
		
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
			// Preprocess content to handle carriage returns and ensure proper line breaks
			processedContent := strings.ReplaceAll(m.previewContent, "\r\r", "\n\n")
			processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
			// Wrap content to viewport width to prevent overflow
			wrappedContent := wordwrap.String(processedContent, m.previewViewport.Width)
			m.previewViewport.SetContent(wrappedContent)
		}
	}
}

// Helper methods
func (m *PipelineBuilderModel) getAllAvailableComponents() []componentItem {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}
	
	// Group components by type - use filtered lists when searching
	typeGroups := make(map[string][]componentItem)
	typeGroups[models.ComponentTypeContext] = m.filteredContexts
	typeGroups[models.ComponentTypePrompt] = m.filteredPrompts
	typeGroups[models.ComponentTypeRules] = m.filteredRules
	
	// Build ordered list based on sections
	var all []componentItem
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			all = append(all, components...)
		}
	}
	
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
	// Update preview when cursor moves
	m.updatePreview()
}

func (m *PipelineBuilderModel) addSelectedComponent() {
	components := m.getAllAvailableComponents()
	if m.leftCursor >= 0 && m.leftCursor < len(components) {
		selected := components[m.leftCursor]
		
		// Check if component is already added
		componentPath := "../" + selected.path
		for i, existing := range m.selectedComponents {
			if existing.Path == componentPath {
				// Component already exists, remove it from the pipeline
				// Update the usage count before removing
				m.updateLocalUsageCount(selected.path, -1)
				
				// Set cursor to the position being removed so viewport scrolls there
				m.rightCursor = i
				
				// Remove the component
				m.selectedComponents = append(
					m.selectedComponents[:i],
					m.selectedComponents[i+1:]...,
				)
				
				// Reorganize to maintain grouping
				m.reorganizeComponentsByType()
				
				// After reorganization, find where the cursor should be
				// If we removed the last item, move cursor to the new last item
				if m.rightCursor >= len(m.selectedComponents) && len(m.selectedComponents) > 0 {
					m.rightCursor = len(m.selectedComponents) - 1
				} else if len(m.selectedComponents) == 0 {
					m.rightCursor = 0
				}
				// Otherwise keep cursor at the same position to show where removal happened
				
				// Update preview after removing component
				m.updatePreview()
				
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
		
		// Update preview after adding component
		m.updatePreview()
	}
}

// insertComponentByType inserts a component in the correct position based on type grouping
func (m *PipelineBuilderModel) insertComponentByType(newComp models.ComponentRef) {
	// Add the component to the list
	m.selectedComponents = append(m.selectedComponents, newComp)
	
	// Reorganize to maintain type grouping
	m.reorganizeComponentsByType()
	
	// Find the position of the newly added component and move cursor there
	for i, comp := range m.selectedComponents {
		if comp.Path == newComp.Path && comp.Type == newComp.Type {
			m.rightCursor = i
			// The viewport will auto-scroll to the cursor in the View method
			break
		}
	}
}

// reorganizeComponentsByType sorts components into groups according to section_order
func (m *PipelineBuilderModel) reorganizeComponentsByType() {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}
	
	// Group components by type
	typeGroups := make(map[string][]models.ComponentRef)
	for _, comp := range m.selectedComponents {
		typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
	}
	
	// Rebuild the array in configured order
	m.selectedComponents = nil
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			m.selectedComponents = append(m.selectedComponents, components...)
		}
	}
	
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
		
		// Update preview after removing component
		m.updatePreview()
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

	// Show preview based on active column
	if m.activeColumn == leftColumn {
		// Show component preview for left column
		components := m.getAllAvailableComponents()
		if len(components) == 0 {
			m.previewContent = "No components to preview."
			return
		}
		
		if m.leftCursor >= 0 && m.leftCursor < len(components) {
			comp := components[m.leftCursor]
			
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
	} else {
		// Show pipeline preview for right column
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
		filename := sanitizeFileName(m.pipeline.Name) + ".yaml"
		
		// Check if pipeline already exists (case-insensitive)
		existingPipelines, err := files.ListPipelines()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to check existing pipelines: %v", err))
		}
		
		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, filename) {
				// Don't overwrite if it's not the same pipeline we're editing
				if m.pipeline.Path == "" || !strings.EqualFold(m.pipeline.Path, existing) {
					return StatusMsg(fmt.Sprintf("× Pipeline '%s' already exists. Please choose a different name.", m.pipeline.Name))
				}
			}
		}
		
		m.pipeline.Path = filename
		
		// Save pipeline
		err = files.WritePipeline(m.pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save pipeline: %v", err))
		}
		
		// Update original components to match saved state
		m.originalComponents = make([]models.ComponentRef, len(m.selectedComponents))
		copy(m.originalComponents, m.selectedComponents)
		
		// Reload components to update usage stats after save
		m.loadAvailableComponents()
		
		// Return success message
		return StatusMsg(fmt.Sprintf("✓ Pipeline saved: %s", m.pipeline.Path))
	}
}

func (m *PipelineBuilderModel) saveAndSetPipeline() tea.Cmd {
	return func() tea.Msg {
		// Update pipeline with selected components
		m.pipeline.Components = m.selectedComponents
		
		// Create filename from name using sanitization
		filename := sanitizeFileName(m.pipeline.Name) + ".yaml"
		
		// Check if pipeline already exists (case-insensitive)
		existingPipelines, err := files.ListPipelines()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to check existing pipelines: %v", err))
		}
		
		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, filename) {
				// Don't overwrite if it's not the same pipeline we're editing
				if m.pipeline.Path == "" || !strings.EqualFold(m.pipeline.Path, existing) {
					return StatusMsg(fmt.Sprintf("× Pipeline '%s' already exists. Please choose a different name.", m.pipeline.Name))
				}
			}
		}
		
		m.pipeline.Path = filename
		
		// Save pipeline
		err = files.WritePipeline(m.pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save pipeline: %v", err))
		}
		
		// Generate pipeline output
		output, err := composer.ComposePipeline(m.pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to generate output: %v", err))
		}

		// Write to PLUQQY.md
		outputPath := m.pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}
		
		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to write output: %v", err))
		}
		
		// Update preview if showing
		if m.showPreview {
			m.previewContent = output
			// Preprocess content to handle carriage returns and ensure proper line breaks
			processedContent := strings.ReplaceAll(output, "\r\r", "\n\n")
			processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
			// Wrap content to viewport width to prevent overflow
			wrappedContent := wordwrap.String(processedContent, m.previewViewport.Width)
			m.previewViewport.SetContent(wrappedContent)
		}
		
		// Update original components to match saved state
		m.originalComponents = make([]models.ComponentRef, len(m.selectedComponents))
		copy(m.originalComponents, m.selectedComponents)
		
		// Reload components to update usage stats after save
		m.loadAvailableComponents()
		
		// Return success message
		return StatusMsg(fmt.Sprintf("✓ Saved & Set → %s", outputPath))
	}
}

func (m *PipelineBuilderModel) nameInputView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(60)

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	heading := "CREATE NEW PIPELINE"
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Name input prompt - centered
	promptText := promptStyle.Render("Enter a descriptive name for your pipeline:")
	centeredPromptStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(promptText)))
	mainContent.WriteString("\n\n")

	// Input field with cursor
	input := m.nameInput + "│" // cursor

	// Render input field with padding for centering
	inputFieldContent := inputStyle.Render(input)
	
	// Add padding to center the input field properly
	centeredInputStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	
	mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(inputFieldContent)))
	
	// Check if pipeline name already exists and show warning
	if m.nameInput != "" {
		testFilename := sanitizeFileName(m.nameInput) + ".yaml"
		existingPipelines, _ := files.ListPipelines()
		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, testFilename) {
				// Show warning
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // Orange/yellow for warning
					Bold(true)
				warningText := warningStyle.Render(fmt.Sprintf("⚠ Warning: Pipeline '%s' already exists", m.nameInput))
				mainContent.WriteString("\n\n")
				mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(warningText)))
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

// exitConfirmationView is no longer used - replaced by ConfirmationModel
/*
func (m *PipelineBuilderModel) exitConfirmationView() string {
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
	if m.exitConfirmationType == "pipeline" {
		warningMsg = "You have unsaved changes in this pipeline."
	} else if m.exitConfirmationType == "component" {
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
	
	// Options - exit is destructive
	options := formatConfirmOptions(true) + "  (exit / stay)"
	centeredOptions := lipgloss.NewStyle().
		Width(contentWidth - 4).
		Align(lipgloss.Center).
		Render(options)
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
*/

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
			// If content has been entered, show confirmation
			if strings.TrimSpace(m.componentContent) != "" {
				m.exitConfirmationType = "component"
				m.exitConfirm.ShowDialog(
					"⚠️  Unsaved Changes",
					"You have unsaved content in this component.",
					"Exit without saving?",
					true, // destructive
					m.width - 4,
					10,
					func() tea.Cmd {
						// Go back to name input
						m.creationStep = 1
						m.componentContent = ""
						return nil
					},
					nil, // onCancel
				)
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
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("28")).
		Background(lipgloss.Color("236")).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 11 // Reserve space for help pane and status bar

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

func (m *PipelineBuilderModel) componentNameInputView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(60)

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 11 // Reserve space for help pane and status bar

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

func (m *PipelineBuilderModel) componentContentEditView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 11 // Reserve space for help pane and status bar (3) + borders (2) + spacing (5)

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

func (m *PipelineBuilderModel) saveNewComponent() tea.Cmd {
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
			return StatusMsg(fmt.Sprintf("× Failed to check existing components: %v", err))
		}
		
		for _, existing := range existingComponents {
			if strings.EqualFold(existing, filename) {
				return StatusMsg(fmt.Sprintf("× %s '%s' already exists. Please choose a different name.", strings.Title(m.componentCreationType), m.componentName))
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
		// Save component but don't exit yet
		return m, m.saveEditedComponent()
	case "E":
		// Save current content and open in external editor
		// First save any unsaved changes
		if m.componentContent != m.originalContent {
			err := files.WriteComponent(m.editingComponentPath, m.componentContent)
			if err != nil {
				// Store error message to display
				m.editSaveMessage = fmt.Sprintf("❌ Failed to save before external edit: %v", err)
				// Set timer to clear message
				m.editSaveTimer = time.NewTimer(3 * time.Second)
				return m, func() tea.Msg {
					<-m.editSaveTimer.C
					return clearEditSaveMsg{}
				}
			}
		}
		// Open in external editor
		return m, m.openInEditor(m.editingComponentPath)
	case "esc":
		// Check if content has changed
		if m.componentContent != m.originalContent {
			// Show confirmation dialog
			m.exitConfirmationType = "component-edit"
			m.exitConfirm.ShowDialog(
				"⚠️  Unsaved Changes",
				"You have unsaved changes to this component.",
				"Exit without saving?",
				true, // destructive
				m.width - 4,
				10,
				func() tea.Cmd {
					// Exit without saving
					m.editingComponent = false
					m.componentContent = ""
					m.editingComponentPath = ""
					m.editingComponentName = ""
					m.editSaveMessage = ""
					m.originalContent = ""
					if m.editSaveTimer != nil {
						m.editSaveTimer.Stop()
					}
					return nil
				},
				nil, // onCancel
			)
			return m, nil
		}
		// No changes, exit immediately
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.editSaveMessage = ""
		m.originalContent = ""
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

func (m *PipelineBuilderModel) handleTagEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tagDeleteConfirm.Active() {
		return m, m.tagDeleteConfirm.Update(msg)
	}
	
	switch msg.String() {
	case "esc":
		m.editingTags = false
		m.tagInput = ""
		m.currentTags = nil
		m.showTagSuggestions = false
		m.tagCloudActive = false
		m.tagCloudCursor = 0
		return m, nil
		
	case "ctrl+s":
		return m, m.saveTags()
		
	case "enter":
		if m.tagCloudActive {
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor >= 0 && m.tagCloudCursor < len(availableForSelection) {
				tag := availableForSelection[m.tagCloudCursor]
				if !m.hasTag(tag) {
					m.currentTags = append(m.currentTags, tag)
				}
			}
		} else {
			if m.tagInput != "" {
				normalized := models.NormalizeTagName(m.tagInput)
				if normalized != "" && !m.hasTag(normalized) {
					m.currentTags = append(m.currentTags, normalized)
				}
				m.tagInput = ""
				m.tagCursor = len(m.currentTags)
			}
		}
		return m, nil
		
	case "tab":
		m.tagCloudActive = !m.tagCloudActive
		if m.tagCloudActive {
			m.tagCloudCursor = 0
		}
		return m, nil
		
	case "left":
		if m.tagCloudActive {
			// Navigate in tag cloud
			if m.tagCloudCursor > 0 {
				m.tagCloudCursor--
			}
		} else {
			// Move cursor left in current tags
			if m.tagInput == "" && m.tagCursor > 0 {
				m.tagCursor--
			}
		}
		return m, nil
		
	case "h":
		// Only handle as navigation if not typing
		if m.tagInput == "" {
			if m.tagCloudActive {
				// Navigate in tag cloud
				if m.tagCloudCursor > 0 {
					m.tagCloudCursor--
				}
			} else {
				// Move cursor left in current tags
				if m.tagCursor > 0 {
					m.tagCursor--
				}
			}
			return m, nil
		}
		
	case "right":
		if m.tagCloudActive {
			// Navigate in tag cloud
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor < len(availableForSelection)-1 {
				m.tagCloudCursor++
			}
		} else {
			// Move cursor right in current tags
			if m.tagInput == "" && m.tagCursor < len(m.currentTags)-1 {
				m.tagCursor++
			}
		}
		return m, nil
		
	case "l":
		// Only handle as navigation if not typing
		if m.tagInput == "" {
			if m.tagCloudActive {
				// Navigate in tag cloud
				availableForSelection := m.getAvailableTagsForCloud()
				if m.tagCloudCursor < len(availableForSelection)-1 {
					m.tagCloudCursor++
				}
			} else {
				// Move cursor right in current tags
				if m.tagCursor < len(m.currentTags)-1 {
					m.tagCursor++
				}
			}
			return m, nil
		}
		
	case "backspace":
		if !m.tagCloudActive {
			if m.tagInput != "" {
				if len(m.tagInput) > 0 {
					m.tagInput = m.tagInput[:len(m.tagInput)-1]
				}
			} else if m.tagCursor > 0 && m.tagCursor <= len(m.currentTags) {
				m.currentTags = append(m.currentTags[:m.tagCursor-1], m.currentTags[m.tagCursor:]...)
				m.tagCursor--
			}
		}
		return m, nil
		
	case "ctrl+d":
		if m.tagCloudActive {
			// In tag cloud: delete from registry (project-wide)
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor >= 0 && m.tagCloudCursor < len(availableForSelection) {
				m.deletingTag = availableForSelection[m.tagCloudCursor]
				usage := m.getTagUsage(m.deletingTag)
				m.deletingTagUsage = usage
				
				// Show confirmation dialog for tag deletion
				var details []string
				if len(usage) > 0 {
					for _, u := range usage {
						details = append(details, u)
					}
				}
				
				m.tagDeleteConfirm.Show(ConfirmationConfig{
					Title:       "Delete Tag from Registry?",
					Message:     fmt.Sprintf("Tag: %s", m.deletingTag),
					Warning:     func() string {
						if len(usage) > 0 {
							return "This tag is currently in use by:"
						}
						return "This tag is not currently in use."
					}(),
					Details:     details,
					Destructive: true,
					Type:        ConfirmTypeDialog,
					Width:       m.width - 4,
					Height:      15,
				}, func() tea.Cmd {
					return m.deleteTagFromRegistry()
				}, func() tea.Cmd {
					m.deletingTag = ""
					m.deletingTagUsage = nil
					return nil
				})
			}
		} else if m.tagInput == "" && m.tagCursor < len(m.currentTags) {
			// In current tags: just remove from component
			if m.tagCursor < len(m.currentTags) {
				m.currentTags = append(m.currentTags[:m.tagCursor], m.currentTags[m.tagCursor+1:]...)
				if m.tagCursor >= len(m.currentTags) && m.tagCursor > 0 {
					m.tagCursor--
				}
			}
		}
		return m, nil
		
	case "up", "k":
		if m.tagCloudActive {
			if m.tagCloudCursor > 0 {
				m.tagCloudCursor--
			}
		}
		return m, nil
		
	case "down", "j":
		if m.tagCloudActive {
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor < len(availableForSelection)-1 {
				m.tagCloudCursor++
			}
		}
		return m, nil
		
	default:
		if !m.tagCloudActive && msg.Type == tea.KeyRunes {
			m.tagInput += string(msg.Runes)
			m.showTagSuggestions = m.tagInput != ""
		}
	}
	
	return m, nil
}

func (m *PipelineBuilderModel) getAvailableTagsForCloud() []string {
	available := []string{}
	for _, tag := range m.availableTags {
		if !m.hasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

func (m *PipelineBuilderModel) getTagUsage(tag string) []string {
	var usage []string
	allComponents := m.getAllAvailableComponents()
	for _, comp := range allComponents {
		for _, t := range comp.tags {
			if strings.EqualFold(t, tag) {
				usage = append(usage, fmt.Sprintf("%s: %s", comp.compType, comp.name))
				break
			}
		}
	}
	return usage
}

func (m *PipelineBuilderModel) deleteTagFromRegistry() tea.Cmd {
	return func() tea.Msg {
		registry, err := tags.NewRegistry()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to load tag registry: %v", err))
		}
		
		if err := registry.RemoveTag(m.deletingTag); err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete tag: %v", err))
		}
		
		if err := registry.Save(); err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save tag registry: %v", err))
		}
		
		m.loadAvailableTags()
		
		newAvailable := []string{}
		for _, tag := range m.currentTags {
			if tag != m.deletingTag {
				newAvailable = append(newAvailable, tag)
			}
		}
		m.currentTags = newAvailable
		
		if m.tagCursor > len(m.currentTags) {
			m.tagCursor = len(m.currentTags)
		}
		
		return StatusMsg(fmt.Sprintf("✓ Deleted tag: %s", m.deletingTag))
	}
}

func (m *PipelineBuilderModel) componentEditView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange like other headers

	// Calculate dimensions  
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 9 // Reserve space for help pane (3) + save message (2) + spacing (3) + status bar (1)

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
		"E edit external",
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

	// Save message area - always render to maintain consistent layout
	saveMessageStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("82")). // Green for success
		Width(m.width).
		Align(lipgloss.Center).
		Padding(0, 1).
		MarginTop(1)

	// Empty status style for maintaining layout
	emptyStatusStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		MarginTop(1)

	s.WriteString("\n")
	if m.editSaveMessage != "" {
		s.WriteString(saveMessageStyle.Render(m.editSaveMessage))
	} else {
		// Render empty space to maintain layout with same dimensions
		s.WriteString(emptyStatusStyle.Render(" "))
	}

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
		
		// Update original content to the saved content
		m.originalContent = m.componentContent
		
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

func (m *PipelineBuilderModel) startTagEditing(path string, currentTags []string) {
	m.editingTags = true
	m.editingTagsPath = path
	m.currentTags = make([]string, len(currentTags))
	copy(m.currentTags, currentTags)
	m.tagInput = ""
	m.tagCursor = 0
	m.showTagSuggestions = false
	m.tagCloudActive = false
	m.tagCloudCursor = 0
	
	// Load available tags
	m.loadAvailableTags()
}

func (m *PipelineBuilderModel) startPipelineTagEditing(currentTags []string) {
	m.editingTags = true
	m.editingTagsPath = "" // Empty path indicates pipeline tags
	m.currentTags = make([]string, len(currentTags))
	copy(m.currentTags, currentTags)
	m.tagInput = ""
	m.tagCursor = 0
	m.showTagSuggestions = false
	m.tagCloudActive = false
	m.tagCloudCursor = 0
	
	// Load available tags
	m.loadAvailableTags()
}

func (m *PipelineBuilderModel) loadAvailableTags() {
	// Get all tags from registry
	registry, err := tags.NewRegistry()
	if err != nil {
		m.availableTags = []string{}
		return
	}
	
	allTags := registry.ListTags()
	m.availableTags = make([]string, 0, len(allTags))
	for _, tag := range allTags {
		m.availableTags = append(m.availableTags, tag.Name)
	}
	
	// Also add tags that exist in components but not in registry
	seenTags := make(map[string]bool)
	for _, tag := range m.availableTags {
		seenTags[tag] = true
	}
	
	// Add tags from all components
	allComponents := m.getAllAvailableComponents()
	for _, comp := range allComponents {
		for _, tag := range comp.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.availableTags = append(m.availableTags, normalized)
				seenTags[normalized] = true
			}
		}
	}
}

func (m *PipelineBuilderModel) hasTag(tag string) bool {
	for _, t := range m.currentTags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

func (m *PipelineBuilderModel) saveTags() tea.Cmd {
	return func() tea.Msg {
		var err error
		
		if m.editingTagsPath == "" {
			// Editing pipeline tags
			if m.pipeline != nil {
				m.pipeline.Tags = make([]string, len(m.currentTags))
				copy(m.pipeline.Tags, m.currentTags)
				// Note: Pipeline will be saved when user saves the pipeline
			}
		} else {
			// Editing component tags
			err = files.UpdateComponentTags(m.editingTagsPath, m.currentTags)
			if err != nil {
				return StatusMsg(fmt.Sprintf("Failed to save tags: %v", err))
			}
		}
		
		// Exit tag editing mode
		m.editingTags = false
		m.tagInput = ""
		m.currentTags = nil
		m.showTagSuggestions = false
		m.tagCloudActive = false
		m.tagCloudCursor = 0
		
		// Reload components to reflect the changes
		m.loadAvailableComponents()
		
		// Update filtered components if search is active
		if m.searchQuery != "" {
			m.performSearch()
		}
		
		return StatusMsg("✓ Tags saved")
	}
}

func (m *PipelineBuilderModel) tagEditView() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	inputStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170")).Padding(0, 1).Width(40)
	suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedSuggestionStyle := lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("170"))
	
	// Calculate dimensions for side-by-side layout
	paneWidth := (m.width - 6) / 2 // Same calculation as main list view
	paneHeight := m.height - 10 // Leave room for help pane
	headerPadding := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	
	var mainContent strings.Builder
	
	if m.tagDeleteConfirm.Active() {
		return m.tagDeleteConfirm.View()
	}
	
	components := m.getAllAvailableComponents()
	itemName := ""
	if m.leftCursor >= 0 && m.leftCursor < len(components) {
		itemName = components[m.leftCursor].name
	}
	
	heading := fmt.Sprintf("EDIT TAGS: %s", strings.ToUpper(itemName))
	remainingWidth := paneWidth - len(heading) - 7
	if remainingWidth < 0 { remainingWidth = 0 }
	colonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")
	
	mainContent.WriteString(headerPadding.Render("Current tags:\n"))
	if len(m.currentTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		mainContent.WriteString(headerPadding.Render(dimStyle.Render("(no tags)")))
	} else {
		var tagDisplay strings.Builder
		for i, tag := range m.currentTags {
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}
			
			style := lipgloss.NewStyle().Background(lipgloss.Color(color)).Foreground(lipgloss.Color("255")).Padding(0, 1)
			
			// Add selection indicators with consistent spacing
			if i == m.tagCursor && m.tagInput == "" {
				// Selected tag with triangle indicators
				indicatorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")). // Bright green for indicators
					Bold(true)
				tagDisplay.WriteString(indicatorStyle.Render("▶ "))
				tagDisplay.WriteString(style.Render(tag))
				tagDisplay.WriteString(indicatorStyle.Render(" ◀"))
			} else {
				// Add invisible spacers to maintain consistent width
				tagDisplay.WriteString("  ")
				tagDisplay.WriteString(style.Render(tag))
				tagDisplay.WriteString("  ")
			}
			
			// Add spacing between tags
			if i < len(m.currentTags)-1 {
				tagDisplay.WriteString("  ")
			}
		}
		mainContent.WriteString(headerPadding.Render(tagDisplay.String()))
	}
	mainContent.WriteString("\n\n")
	
	mainContent.WriteString(headerPadding.Render("Add new tag:"))
	mainContent.WriteString("\n")
	
	// Create input display with cursor
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)
	
	inputDisplay := m.tagInput
	if !m.tagCloudActive && m.tagInput != "" {
		// Add cursor to existing input when active
		inputDisplay = m.tagInput + cursorStyle.Render("│")
	}
	
	// Show placeholder if empty
	if m.tagInput == "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		if !m.tagCloudActive {
			inputDisplay = placeholderStyle.Render("Type to add a new tag...") + cursorStyle.Render("│")
		} else {
			inputDisplay = placeholderStyle.Render("Type to add a new tag...")
		}
	}
	
	// Highlight input border when active
	activeInputStyle := inputStyle
	if !m.tagCloudActive {
		activeInputStyle = inputStyle.BorderForeground(lipgloss.Color("170"))
	}
	
	mainContent.WriteString(headerPadding.Render(activeInputStyle.Render(inputDisplay)))
	
	if m.showTagSuggestions && len(m.availableTags) > 0 {
		mainContent.WriteString("\n\n")
		mainContent.WriteString(headerPadding.Render("Suggestions:\n"))
		
		var suggestions []string
		for _, tag := range m.availableTags {
			if m.tagInput != "" && strings.HasPrefix(strings.ToLower(tag), strings.ToLower(m.tagInput)) && !m.hasTag(tag) {
				suggestions = append(suggestions, tag)
			}
		}
		
		if len(suggestions) > 0 {
			maxSuggestions := 5
			if len(suggestions) < maxSuggestions {
				maxSuggestions = len(suggestions)
			}
			
			for i := 0; i < maxSuggestions; i++ {
				if i == 0 {
					mainContent.WriteString(headerPadding.Render(selectedSuggestionStyle.Render(suggestions[i])))
				} else {
					mainContent.WriteString(headerPadding.Render(suggestionStyle.Render(suggestions[i])))
				}
				mainContent.WriteString("\n")
			}
		}
	}
	
	// Build right pane content
	var rightContent strings.Builder
	
	// Available tags header
	availableTagsTitle := "AVAILABLE TAGS"
	availableTagsRemainingWidth := paneWidth - len(availableTagsTitle) - 7 // Adjust for smaller width
	if availableTagsRemainingWidth < 0 {
		availableTagsRemainingWidth = 0
	}
	rightContent.WriteString(headerPadding.Render(titleStyle.Render(availableTagsTitle) + " " + colonStyle.Render(strings.Repeat(":", availableTagsRemainingWidth))))
	rightContent.WriteString("\n\n")
	
	// Always display available tags
	if len(m.availableTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		rightContent.WriteString(headerPadding.Render(dimStyle.Render("(no available tags)")))
	} else {
		// Group tags in rows for better display
		var tagRows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		maxRowWidth := paneWidth - 6 // Account for padding
		
		// Get available tags that haven't been added yet
		var availableForCloud []string
		for _, tag := range m.availableTags {
			if !m.hasTag(tag) {
				availableForCloud = append(availableForCloud, tag)
			}
		}
		
		for i, tag := range availableForCloud {
			// Get tag color
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}
			
			tagStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(color)).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
			
			// Calculate tag display width
			var tagDisplay string
			if m.tagCloudActive && i == m.tagCloudCursor {
				indicatorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")).
					Bold(true)
				tagDisplay = indicatorStyle.Render("▶ ") + tagStyle.Render(tag) + indicatorStyle.Render(" ◀")
			} else {
				// Add invisible spacers to maintain consistent width
				tagDisplay = "  " + tagStyle.Render(tag) + "  "
			}
			
			tagWidth := lipgloss.Width(tagDisplay) + 2 // Add spacing
			
			// Check if we need to start a new row
			if rowTags > 0 && currentRowWidth+tagWidth+2 > maxRowWidth {
				tagRows.WriteString("\n\n") // Double newline for vertical spacing
				rowTags = 0
				currentRowWidth = 0
			}
			
			if rowTags > 0 {
				tagRows.WriteString("  ")
			}
			tagRows.WriteString(tagDisplay)
			currentRowWidth += tagWidth + 2
			rowTags++
		}
		
		rightContent.WriteString(headerPadding.Render(tagRows.String()))
	}
	
	// Create border styles
	activeBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))
	
	inactiveBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))
	
	leftPaneBorder := inactiveBorderStyle
	if !m.tagCloudActive {
		leftPaneBorder = activeBorderStyle
	}
	
	leftPane := leftPaneBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(mainContent.String())
	
	rightPaneBorder := inactiveBorderStyle
	if m.tagCloudActive {
		rightPaneBorder = activeBorderStyle
	}
	
	rightPane := rightPaneBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(rightContent.String())
	
	// Join panes side by side
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		" ", // Single space gap, same as main list view
		rightPane,
	)
	
	// Help section - match Main List View exactly
	var help []string
	if m.tagCloudActive {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←/→ navigate",
			"ctrl+d delete tag",
			"ctrl+s save",
			"esc cancel",
		}
	} else {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←/→ select tag",
			"ctrl+d delete tag",
			"ctrl+s save",
			"esc cancel",
		}
	}
	helpBorderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Width(m.width - 4).Padding(0, 1)
	helpContent := formatHelpText(help)
	
	var s strings.Builder
	contentStyle := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	
	s.WriteString(contentStyle.Render(mainView))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))
	
	return s.String()
}