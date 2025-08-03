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
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type MainListModel struct {
	// Pipelines data
	pipelines          []pipelineItem
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
	lastDataPane       pane // Track last data pane (pipelines or components) for preview
	pipelinesViewport  viewport.Model
	componentsViewport viewport.Model
	showPreview        bool
	
	// Window dimensions
	width              int
	height             int
	
	// Error handling
	err                error
	
	// Pipeline operations
	pipelineOperator  *PipelineOperator
	deletingFromPane  pane // Track which pane initiated the delete
	archivingFromPane pane // Track which pane initiated the archive
	
	// Component creation
	componentCreator *ComponentCreator
	
	// Component editing
	componentEditor  *ComponentEditor
	
	// Exit confirmation
	exitConfirm          *ConfirmationModel
	exitConfirmationType string // "component" or "component-edit"
	
	// Tag editing
	tagEditor *TagEditor
	
	// Search engine
	searchEngine          *search.Engine
	
	// Search state
	searchBar             *SearchBar
	searchQuery           string
	filteredPipelines     []pipelineItem
	filteredComponents    []componentItem
}

func NewMainListModel() *MainListModel {
	m := &MainListModel{
		activePane:         componentsPane,
		lastDataPane:       componentsPane, // Initialize to components
		showPreview:        false, // Start with preview hidden, user can toggle with 'p'
		previewViewport:    viewport.New(80, 20), // Default size
		pipelinesViewport:  viewport.New(40, 20), // Default size
		componentsViewport: viewport.New(40, 20), // Default size
		searchBar:          NewSearchBar(),
		pipelineOperator:   NewPipelineOperator(),
		exitConfirm:        NewConfirmation(),
		componentCreator:   NewComponentCreator(),
		componentEditor:    NewComponentEditor(),
		tagEditor:          NewTagEditor(),
	}
	m.loadPipelines()
	m.loadComponents()
	m.initializeSearchEngine()
	// Initialize filtered lists with all items
	m.filteredPipelines = m.pipelines
	m.filteredComponents = m.getAllComponents()
	return m
}

func (m *MainListModel) loadPipelines() {
	pipelineFiles, err := files.ListPipelines()
	if err != nil {
		m.err = err
		return
	}
	
	m.pipelines = nil
	for _, pipelineFile := range pipelineFiles {
		// Load pipeline to get metadata
		pipeline, err := files.ReadPipeline(pipelineFile)
		if err != nil {
			continue
		}
		
		// Calculate token count
		tokenCount := 0
		if pipeline != nil {
			output, err := composer.ComposePipeline(pipeline)
			if err == nil {
				tokenCount = utils.EstimateTokens(output)
			}
		}
		
		m.pipelines = append(m.pipelines, pipelineItem{
			name:       pipeline.Name, // Use the actual pipeline name from YAML
			path:       pipelineFile,
			tags:       pipeline.Tags,
			tokenCount: tokenCount,
		})
	}
	
	// Rebuild search index when pipelines are reloaded
	if m.searchEngine != nil {
		m.searchEngine.BuildIndex()
	}
	
	// Update filtered list if no active search
	if m.searchQuery == "" {
		m.filteredPipelines = m.pipelines
	}
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
	
	// Rebuild search index when components are reloaded
	if m.searchEngine != nil {
		m.searchEngine.BuildIndex()
	}
	
	// Update filtered list if no active search
	if m.searchQuery == "" {
		m.filteredComponents = m.getAllComponents()
	}
}

func (m *MainListModel) initializeSearchEngine() {
	// Use SearchManager for initialization
	searchManager := NewSearchManager()
	if err := searchManager.InitializeEngine(); err != nil {
		// Log error but don't fail - search will be unavailable
		return
	}
	m.searchEngine = searchManager.GetEngine()
}

func (m *MainListModel) performSearch() {
	if m.searchQuery == "" {
		// No search query, show all items
		m.filteredPipelines = m.pipelines
		m.filteredComponents = m.getAllComponents()
		return
	}
	
	// Use search engine to find matching items
	if m.searchEngine != nil {
		results, err := m.searchEngine.Search(m.searchQuery)
		if err != nil {
			// On error, show all items
			m.filteredPipelines = m.pipelines
			m.filteredComponents = m.getAllComponents()
			return
		}
		
		// Use the helper function to filter results
		m.filteredPipelines, m.filteredComponents = FilterSearchResults(
			results,
			m.pipelines,
			m.getAllComponents(),
		)
		
		// Reset cursors if they're out of bounds
		if m.pipelineCursor >= len(m.filteredPipelines) {
			m.pipelineCursor = 0
		}
		if m.componentCursor >= len(m.filteredComponents) {
			m.componentCursor = 0
		}
	}
}



func (m *MainListModel) getAllComponents() []componentItem {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}
	
	// Group components by type
	typeGroups := make(map[string][]componentItem)
	typeGroups[models.ComponentTypeContext] = m.contexts
	typeGroups[models.ComponentTypePrompt] = m.prompts
	typeGroups[models.ComponentTypeRules] = m.rules
	
	// Build ordered list based on sections
	var all []componentItem
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			all = append(all, components...)
		}
	}
	
	return all
}

// getCurrentComponents returns either filtered components (if searching) or all components
func (m *MainListModel) getCurrentComponents() []componentItem {
	return m.filteredComponents
}

func (m *MainListModel) getEditingItemName() string {
	if m.tagEditor.ItemType == "component" {
		components := m.getCurrentComponents()
		if m.componentCursor >= 0 && m.componentCursor < len(components) {
			return components[m.componentCursor].name
		}
	} else {
		if m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			return m.pipelines[m.pipelineCursor].name
		}
	}
	return ""
}

func (m *MainListModel) Init() tea.Cmd {
	return nil
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search input when search pane is active
		if m.activePane == searchPane && !m.tagEditor.Active && !m.componentCreator.IsActive() && !m.componentEditor.IsActive() {
			var cmd tea.Cmd
			m.searchBar, cmd = m.searchBar.Update(msg)
			
			// Check if search query changed
			if m.searchQuery != m.searchBar.Value() {
				m.searchQuery = m.searchBar.Value()
				m.performSearch()
			}
			
			// Handle special keys for search
			switch msg.String() {
			case "esc":
				// Clear search and return to components pane
				m.searchBar.SetValue("")
				m.searchQuery = ""
				m.performSearch()
				m.activePane = componentsPane
				m.searchBar.SetActive(false)
				return m, nil
			case "tab":
				// Let tab handling below take care of navigation
			default:
				return m, cmd
			}
		}
		
		// Handle exit confirmation
		if m.exitConfirm.Active() {
			return m, m.exitConfirm.Update(msg)
		}
		
		// Handle component creation mode
		if m.componentCreator.IsActive() {
			return m.handleComponentCreation(msg)
		}
		
		// Handle component editing mode
		if m.componentEditor.IsActive() {
			// Handle exit confirmation first
			if m.componentEditor.ExitConfirmActive {
				_, cmd := m.componentEditor.HandleInput(msg, m.width, m.height)
				return m, cmd
			}
			
			_, cmd := m.componentEditor.HandleInput(msg, m.width, m.height)
			// Check if editor is still active after handling input
			if !m.componentEditor.IsActive() {
				// Reload components after editing
				m.loadComponents()
				m.performSearch()
			}
			return m, cmd
		}
		
		// Handle tag editing mode
		if m.tagEditor.Active {
			handled, cmd := m.tagEditor.HandleInput(msg)
			if handled {
				// Check if we need to reload data after saving
				if cmd != nil {
					return m, cmd
				}
				return m, nil
			}
		}
		
		// Handle delete confirmation
		if m.pipelineOperator.IsDeleteConfirmActive() {
			return m, m.pipelineOperator.UpdateDeleteConfirm(msg)
		}
		
		// Handle archive confirmation
		if m.pipelineOperator.IsArchiveConfirmActive() {
			return m, m.pipelineOperator.UpdateArchiveConfirm(msg)
		}
		
		// Normal mode key handling
		switch msg.String() {
		case "q":
			return m, tea.Quit
		
		case "tab":
			// Switch between content panes only (no search)
			if m.showPreview {
				// When preview is shown, cycle through content panes
				switch m.activePane {
				case searchPane:
					// If in search, exit to components
					m.activePane = componentsPane
					m.searchBar.SetActive(false)
				case componentsPane:
					m.activePane = pipelinesPane
				case pipelinesPane:
					m.activePane = previewPane
				case previewPane:
					m.activePane = componentsPane
				}
			} else {
				// When preview is hidden, cycle between components and pipelines
				switch m.activePane {
				case searchPane:
					// If in search, exit to components
					m.activePane = componentsPane
					m.searchBar.SetActive(false)
				case componentsPane:
					m.activePane = pipelinesPane
				case pipelinesPane:
					m.activePane = componentsPane
				}
			}
			// Track last data pane when switching
			if m.activePane == pipelinesPane || m.activePane == componentsPane {
				m.lastDataPane = m.activePane
			}
			// Update preview when switching to non-preview pane
			if m.activePane != previewPane && m.activePane != searchPane {
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
				components := m.getCurrentComponents()
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
		
		case "/":
			// Jump to search
			m.activePane = searchPane
			m.searchBar.SetActive(true)
			return m, nil
		
		case "enter":
			if m.activePane == pipelinesPane {
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					// View the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineViewerView,
							pipeline: m.pipelines[m.pipelineCursor].path, // Use path (filename) not name
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
							pipeline: m.pipelines[m.pipelineCursor].path, // Use path (filename) not name
						}
					}
				}
			} else if m.activePane == componentsPane {
				// Edit component in TUI editor
				components := m.getCurrentComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					// Read the component content
					content, err := files.ReadComponent(comp.path)
					if err != nil {
						m.err = err
						return m, nil
					}
					// Enter editing mode
					m.componentEditor.StartEditing(comp.path, comp.name, content.Content)
					
					return m, nil
				}
			}
		
		case "E":
			// Edit component in external editor
			if m.activePane == componentsPane {
				components := m.getCurrentComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					return m, m.pipelineOperator.OpenInEditor(comp.path, m.loadComponents)
				}
			}
			// Explicitly do nothing for pipelines pane - editing YAML directly is not encouraged
		
		case "t":
			// Edit tags
			if m.activePane == componentsPane {
				// Use filtered components if search is active
				components := m.filteredComponents
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					m.tagEditor.Start(comp.path, comp.tags, "component")
				}
			} else if m.activePane == pipelinesPane {
				// Use filtered pipelines if search is active
				pipelines := m.filteredPipelines
				if m.pipelineCursor >= 0 && m.pipelineCursor < len(pipelines) {
					pipeline := pipelines[m.pipelineCursor]
					m.tagEditor.Start(pipeline.path, pipeline.tags, "pipeline")
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
				m.componentCreator.Start()
				return m, nil
			}
		
		
		case "S":
			if m.activePane == pipelinesPane {
				// Set selected pipeline (generate PLUQQY.md)
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					return m, m.pipelineOperator.SetPipeline(m.pipelines[m.pipelineCursor].path)
				}
			}
		
		case "d", "delete":
			if m.activePane == pipelinesPane {
				// Delete pipeline with confirmation
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					m.deletingFromPane = pipelinesPane
					pipelineName := m.pipelines[m.pipelineCursor].name
					pipelinePath := m.pipelines[m.pipelineCursor].path
					
					m.pipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							return m.pipelineOperator.DeletePipeline(pipelinePath, m.loadPipelines)
						},
						nil, // onCancel
					)
				}
			} else if m.activePane == componentsPane {
				// Delete component with confirmation
				components := m.getCurrentComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					m.deletingFromPane = componentsPane
					
					m.pipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete %s '%s'?", comp.compType, comp.name),
						func() tea.Cmd {
							return m.pipelineOperator.DeleteComponent(comp, m.loadComponents)
						},
						nil, // onCancel
					)
				}
			}
		
		case "s":
			// Open settings editor
			return m, func() tea.Msg {
				return SwitchViewMsg{
					view: settingsEditorView,
				}
			}
			
		case "a":
			if m.activePane == pipelinesPane {
				// Archive pipeline with confirmation
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					m.archivingFromPane = pipelinesPane
					pipelineName := m.pipelines[m.pipelineCursor].name
					pipelinePath := m.pipelines[m.pipelineCursor].path
					
					m.pipelineOperator.ShowArchiveConfirmation(
						fmt.Sprintf("Archive pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							return m.pipelineOperator.ArchivePipeline(pipelinePath, m.loadPipelines)
						},
						nil, // onCancel
					)
				}
			} else if m.activePane == componentsPane {
				// Archive component with confirmation
				components := m.getCurrentComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					m.archivingFromPane = componentsPane
					
					m.pipelineOperator.ShowArchiveConfirmation(
						fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
						func() tea.Cmd {
							return m.pipelineOperator.ArchiveComponent(comp, m.loadComponents)
						},
						nil, // onCancel
					)
				}
			}
		}
	
	case ReloadMsg:
		// Reload data after tag editing
		m.loadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.searchQuery != "" {
			m.performSearch()
		}
		return m, nil
	}
	
	// Update preview if needed
	if m.showPreview && m.previewContent != "" {
		// Preprocess content to handle carriage returns and ensure proper line breaks
		processedContent := preprocessContent(m.previewContent)
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
	if m.exitConfirm.Active() {
		return m.exitConfirm.View()
	}
	
	// If creating component, show creation wizard
	if m.componentCreator.IsActive() {
		return m.componentCreationView()
	}
	
	// If editing component, show edit view
	if m.componentEditor.IsActive() {
		// Update viewport dimensions
		m.componentEditor.UpdateViewport(m.width, m.height)
		
		// Handle exit confirmation dialog
		if m.componentEditor.ExitConfirmActive {
			return m.componentEditor.ExitConfirm.View()
		}
		
		// Create the component editing view renderer
		renderer := NewComponentEditingViewRenderer(m.width, m.height)
		return renderer.RenderEditView(
			m.componentEditor.ComponentName, 
			m.componentEditor.Content,
			m.componentEditor.GetViewport(),
		)
	}
	
	// If editing tags, show tag edit view
	if m.tagEditor.Active {
		// Create the tag editing view renderer
		renderer := NewTagEditingViewRenderer(m.width, m.height)
		renderer.ItemName = m.getEditingItemName()
		renderer.CurrentTags = m.tagEditor.CurrentTags
		renderer.TagInput = m.tagEditor.TagInput
		renderer.TagCursor = m.tagEditor.TagCursor
		renderer.ShowSuggestions = m.tagEditor.ShowSuggestions
		renderer.SuggestionCursor = m.tagEditor.SuggestionCursor
		renderer.TagCloudActive = m.tagEditor.TagCloudActive
		renderer.TagCloudCursor = m.tagEditor.TagCloudCursor
		renderer.AvailableTags = m.tagEditor.AvailableTags
		renderer.TagDeleteConfirm = m.tagEditor.TagDeleteConfirm
		renderer.DeletingTag = m.tagEditor.DeletingTag
		renderer.DeletingTagUsage = m.tagEditor.DeletingTagUsage
		renderer.GetSuggestionsFunc = func(input string, availableTags []string, currentTags []string) []string {
			return m.tagEditor.GetSuggestions()
		}
		renderer.GetAvailableTagsForCloudFunc = func(availableTags []string, currentTags []string) []string {
			return m.tagEditor.GetAvailableTagsForCloud()
		}
		return renderer.Render()
	}

	// Styles
	activeStyle := ActiveBorderStyle
	inactiveStyle := InactiveBorderStyle
	normalStyle := NormalStyle
	typeHeaderStyle := TypeHeaderStyle

	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3              // Height for search bar
	contentHeight := m.height - 14 - searchBarHeight    // Reserve space for header, search bar, help pane, and spacing

	if m.showPreview {
		contentHeight = contentHeight / 2
	}
	
	// Ensure minimum height for content
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Update search bar active state and render it
	m.searchBar.SetActive(m.activePane == searchPane)
	m.searchBar.SetWidth(m.width)

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
	// Dynamic header and colon styles based on active pane
	leftHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if m.activePane == componentsPane {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	leftColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if m.activePane == componentsPane {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	leftContent.WriteString(headerPadding.Render(leftHeaderStyle.Render(heading) + " " + leftColonStyle.Render(strings.Repeat(":", remainingWidth))))
	leftContent.WriteString("\n\n")

	
	// Table header styles
	headerStyle := HeaderStyle
	
	// Table column widths (adjusted for viewport width)
	// Use viewport width instead of column width to ensure content fits
	viewportWidth := columnWidth - 4 // Same as m.componentsViewport.Width
	nameWidth, tagsWidth, tokenWidth, usageWidth := formatColumnWidths(viewportWidth, true)
	
	// Render table header with consistent spacing
	header := fmt.Sprintf("  %-*s %-*s  %*s %*s", 
		nameWidth, "Name",
		tagsWidth, "Tags",
		tokenWidth, "~Tokens",
		usageWidth, "Usage")
	leftContent.WriteString(headerPadding.Render(headerStyle.Render(header)))
	leftContent.WriteString("\n\n")
	
	// Build scrollable content for components viewport
	var componentsScrollContent strings.Builder
	
	// Use filtered components instead of all components
	if len(m.filteredComponents) == 0 {
		if m.activePane == componentsPane {
			// Active pane - show prominent message
			emptyStyle := EmptyActiveStyle
			
			// Check if we have components but they're filtered out
			allComponents := m.getAllComponents()
			if len(allComponents) > 0 && m.searchQuery != "" {
				componentsScrollContent.WriteString(emptyStyle.Render("No components match your search."))
			} else {
				componentsScrollContent.WriteString(emptyStyle.Render("No components found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := EmptyInactiveStyle
			
			// Check if we have components but they're filtered out
			allComponents := m.getAllComponents()
			if len(allComponents) > 0 && m.searchQuery != "" {
				componentsScrollContent.WriteString(dimmedStyle.Render("No components match your search."))
			} else {
				componentsScrollContent.WriteString(dimmedStyle.Render("No components found."))
			}
		}
	} else {
		currentType := ""
	
		for i, comp := range m.filteredComponents {
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
		nameStr := truncateName(comp.name, nameWidth)
		
		
		// Format usage count
		usageStr := fmt.Sprintf("%d", comp.usageCount)
		
		// Format token count - right-aligned with consistent width
		tokenStr := fmt.Sprintf("%d", comp.tokenCount)
		
		// Format tags
		tagsStr := renderTagChipsWithWidth(comp.tags, tagsWidth, 2) // Show max 2 tags inline
		
		// Build the row components separately for proper styling
		namePart := fmt.Sprintf("%-*s", nameWidth, nameStr)
		
		// For tags, we need to pad based on rendered width
		tagsPadding := tagsWidth - lipgloss.Width(tagsStr)
		if tagsPadding < 0 {
			tagsPadding = 0
		}
		tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)
		
		tokenPart := fmt.Sprintf("%*s", tokenWidth, tokenStr)
		usagePart := fmt.Sprintf("%*s", usageWidth, usageStr)
		
		// Build row with styling
		var row string
		if m.activePane == componentsPane && i == m.componentCursor {
			// Apply selection styling only to name column
			row = "▸ " + SelectedStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart + " " + usagePart)
		} else {
			// Normal row styling
			row = "  " + normalStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart + " " + usagePart)
		}
		
		componentsScrollContent.WriteString(row)
		
			if i < len(m.filteredComponents)-1 {
				componentsScrollContent.WriteString("\n")
			}
		}
	}
	
	// Update components viewport with content
	m.componentsViewport.SetContent(componentsScrollContent.String())
	
	// Update viewport to follow cursor
	if m.activePane == componentsPane && len(m.filteredComponents) > 0 {
		// Calculate the line position of the cursor
		currentLine := 0
		for i := 0; i < m.componentCursor && i < len(m.filteredComponents); i++ {
			currentLine++ // Component line
			// Check if it's the first item of a new type to add header line
			if i == 0 || m.filteredComponents[i].compType != m.filteredComponents[i-1].compType {
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
	// Dynamic header and colon styles based on active pane
	rightHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if m.activePane == pipelinesPane {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	rightColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if m.activePane == pipelinesPane {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	rightContent.WriteString(headerPadding.Render(rightHeaderStyle.Render(rightHeading) + " " + rightColonStyle.Render(strings.Repeat(":", rightRemainingWidth))))
	rightContent.WriteString("\n\n")
	
	// Table header for pipelines with token count
	pipelineHeaderStyle := HeaderStyle
	
	// Table column widths for pipelines (adjusted for viewport width)
	pipelineViewportWidth := columnWidth - 4 // Same as m.pipelinesViewport.Width
	pipelineNameWidth, pipelineTagsWidth, pipelineTokenWidth, _ := formatColumnWidths(pipelineViewportWidth, false)
	
	// Render table header with consistent spacing to match rows
	pipelineHeader := fmt.Sprintf("  %-*s %-*s %*s", 
		pipelineNameWidth, "Name",
		pipelineTagsWidth, "Tags",
		pipelineTokenWidth, "~Tokens")
	rightContent.WriteString(headerPadding.Render(pipelineHeaderStyle.Render(pipelineHeader)))
	rightContent.WriteString("\n\n")

	// Build scrollable content for pipelines viewport
	var pipelinesScrollContent strings.Builder
	
	if len(m.filteredPipelines) == 0 {
		if m.activePane == pipelinesPane {
			// Active pane - show prominent message
			emptyStyle := EmptyActiveStyle
			
			// Check if we have pipelines but they're filtered out
			if len(m.pipelines) > 0 && m.searchQuery != "" {
				pipelinesScrollContent.WriteString(emptyStyle.Render("No pipelines match your search."))
			} else {
				pipelinesScrollContent.WriteString(emptyStyle.Render("No pipelines found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := EmptyInactiveStyle
			
			// Check if we have pipelines but they're filtered out
			if len(m.pipelines) > 0 && m.searchQuery != "" {
				pipelinesScrollContent.WriteString(dimmedStyle.Render("No pipelines match your search."))
			} else {
				pipelinesScrollContent.WriteString(dimmedStyle.Render("No pipelines found."))
			}
		}
	} else {
		for i, pipeline := range m.filteredPipelines {
			// Format the pipeline name
			nameStr := truncateName(pipeline.name, pipelineNameWidth)
			
			// Format tags
			tagsStr := renderTagChipsWithWidth(pipeline.tags, pipelineTagsWidth, 2) // Show max 2 tags inline
			
			// Format token count - right-aligned
			tokenStr := fmt.Sprintf("%d", pipeline.tokenCount)
			
			// Build the row components separately for proper styling
			namePart := fmt.Sprintf("%-*s", pipelineNameWidth, nameStr)
			
			// For tags, we need to pad based on rendered width
			tagsPadding := pipelineTagsWidth - lipgloss.Width(tagsStr)
			if tagsPadding < 0 {
				tagsPadding = 0
			}
			tagsPart := tagsStr + strings.Repeat(" ", tagsPadding)
			
			tokenPart := fmt.Sprintf("%*s", pipelineTokenWidth, tokenStr)
			
			// Build row with styling
			var row string
			if m.activePane == pipelinesPane && i == m.pipelineCursor {
				// Apply selection styling only to name column
				row = "▸ " + SelectedStyle.Render(namePart) + " " + tagsPart + " " + normalStyle.Render(tokenPart)
			} else {
				// Normal row styling
				row = "  " + normalStyle.Render(namePart) + " " + tagsPart + " " + normalStyle.Render(tokenPart)
			}
			
			pipelinesScrollContent.WriteString(row)
			
			if i < len(m.pipelines)-1 {
				pipelinesScrollContent.WriteString("\n")
			}
		}
	}
	
	// Update pipelines viewport with content
	m.pipelinesViewport.SetContent(pipelinesScrollContent.String())
	
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
	
	// Add search bar first
	s.WriteString(m.searchBar.View())
	s.WriteString("\n")
	
	// Then add the columns
	s.WriteString(contentStyle.Render(columns))

	// Add preview if enabled
	if m.showPreview && m.previewContent != "" {
		// Calculate token count
		tokenCount := utils.EstimateTokens(m.previewContent)
		
		// Create token badge with appropriate color
		tokenBadgeStyle := GetTokenBadgeStyle(tokenCount)
		
		tokenBadge := tokenBadgeStyle.Render(utils.FormatTokenCount(tokenCount))
		
		// Apply active/inactive style to preview border
		var previewBorderStyle lipgloss.Style
		if m.activePane == previewPane {
			previewBorderStyle = ActiveBorderStyle
		} else {
			previewBorderStyle = InactiveBorderStyle
		}
		previewBorderStyle = previewBorderStyle.Width(m.width - 4) // Account for padding (2) and border (2)

		s.WriteString("\n")
		
		// Build preview content with header inside
		var previewContent strings.Builder
		// Create heading with colons and token info
		var previewHeading string
		
		// Determine what we're previewing based on lastDataPane
		if m.lastDataPane == pipelinesPane && len(m.pipelines) > 0 && m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			pipelineName := m.pipelines[m.pipelineCursor].name
			previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
		} else if m.lastDataPane == componentsPane && len(m.filteredComponents) > 0 && m.componentCursor >= 0 && m.componentCursor < len(m.filteredComponents) {
			comp := m.filteredComponents[m.componentCursor]
			previewHeading = fmt.Sprintf("COMPONENT PREVIEW (%s)", comp.name)
		} else {
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
		// Dynamic header and colon styles based on active pane
		previewHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(func() string {
				if m.activePane == previewPane {
					return "170" // Purple when active
				}
				return "214" // Orange when inactive
			}()))
		previewColonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(func() string {
				if m.activePane == previewPane {
					return "170" // Purple when active
				}
				return "240" // Gray when inactive
			}()))
		previewHeaderPadding := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		previewContent.WriteString(previewHeaderPadding.Render(previewHeaderStyle.Render(previewHeading) + " " + previewColonStyle.Render(strings.Repeat(":", colonSpace)) + " " + tokenInfo))
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
	if m.pipelineOperator.IsDeleteConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(confirmStyle.Render(m.pipelineOperator.ViewDeleteConfirm(m.width - 4)))
	}
	
	// Show archive confirmation if active
	if m.pipelineOperator.IsArchiveConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange for archive
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(confirmStyle.Render(m.pipelineOperator.ViewArchiveConfirm(m.width - 4)))
	}
	
	// Help text in bordered pane
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).  // Account for left/right padding (2) and borders (2)
		Padding(0, 1)  // Internal padding for help text
	
	var helpContent string
	if m.activePane == searchPane {
		// Show search syntax help when search is active
		helpRows := [][]string{
			{"esc exit search", "enter search", "tag:<name>", "type:<type>"},
			{"pipeline:<name>", "<keyword>", "combine with spaces"},
		}
		helpContent = formatHelpTextRows(helpRows, m.width - 8) // -8 for borders and padding
	} else {
		// Show normal navigation help - grouped by function
		helpRows := [][]string{
			// Row 1: Navigation & viewing
			{"/ search", "tab switch pane", "↑/↓ nav", "enter view", "p preview"},
			// Row 2: CRUD operations & system
			{"n new", "e edit", "E external", "t tag", "d delete", "a archive", "S set", "s settings", "ctrl+c quit"},
		}
		helpContent = formatHelpTextRows(helpRows, m.width - 8) // -8 for borders and padding
	}
	
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update search bar width
	m.searchBar.SetWidth(width)
	m.updateViewportSizes()
}

func (m *MainListModel) updateViewportSizes() {
	// Calculate dimensions
	columnWidth := (m.width - 6) / 2 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3              // Height for search bar
	contentHeight := m.height - 14 - searchBarHeight    // Reserve space for header, search bar, help pane, and spacing
	
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
	
	// Use PreviewRenderer for generating preview content
	renderer := &PreviewRenderer{ShowPreview: m.showPreview}
	
	// Determine which pane to preview from
	previewPane := m.activePane
	if m.activePane == previewPane {
		// If we're on the preview pane, use the last data pane
		previewPane = m.lastDataPane
	}
	
	if previewPane == pipelinesPane {
		// Show pipeline preview
		if len(m.pipelines) == 0 {
			m.previewContent = renderer.RenderEmptyPreview(pipelinesPane, false, false)
			return
		}
		
		if m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			pipelinePath := m.pipelines[m.pipelineCursor].path
			m.previewContent = renderer.RenderPipelinePreview(pipelinePath)
		}
	} else if previewPane == componentsPane {
		// Show component preview
		components := m.getCurrentComponents()
		if len(components) == 0 {
			m.previewContent = renderer.RenderEmptyPreview(componentsPane, false, false)
			return
		}
		
		if m.componentCursor >= 0 && m.componentCursor < len(components) {
			comp := components[m.componentCursor]
			m.previewContent = renderer.RenderComponentPreview(comp)
		}
	}
}






func (m *MainListModel) handleComponentCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.componentCreator.GetCurrentStep() {
	case 0: // Type selection
		if m.componentCreator.HandleTypeSelection(msg) {
			return m, nil
		}
		
	case 1: // Name input
		if m.componentCreator.HandleNameInput(msg) {
			return m, nil
		}
		
	case 2: // Content input
		// Special handling for escape with unsaved content
		if msg.String() == "esc" && strings.TrimSpace(m.componentCreator.GetComponentContent()) != "" {
			m.exitConfirmationType = "component"
			m.exitConfirm.ShowDialog(
				"⚠️  Unsaved Changes",
				"You have unsaved content in this component.",
				"Exit without saving?",
				true, // destructive
				m.width - 4,
				10,
				func() tea.Cmd {
					// Exit - reset component creator
					m.componentCreator.Reset()
					return nil
				},
				nil, // onCancel
			)
			return m, nil
		}
		
		// Let component creator handle the input
		handled, err := m.componentCreator.HandleContentEdit(msg)
		if err != nil {
			return m, func() tea.Msg { return StatusMsg(fmt.Sprintf("✗ %s", err.Error())) }
		}
		
		if handled {
			// Check if component was saved (creator will be reset)
			if !m.componentCreator.IsActive() {
				// Component was saved, reload components
				m.loadComponents()
				return m, func() tea.Msg { return StatusMsg(m.componentCreator.GetStatusMessage()) }
			}
			return m, nil
		}
	}
	
	return m, nil
}

func (m *MainListModel) componentCreationView() string {
	renderer := NewComponentCreationViewRenderer(m.width, m.height)
	
	switch m.componentCreator.GetCurrentStep() {
	case 0:
		return renderer.RenderTypeSelection(m.componentCreator.GetTypeCursor())
	case 1:
		return renderer.RenderNameInput(m.componentCreator.GetComponentType(), m.componentCreator.GetComponentName())
	case 2:
		return renderer.RenderContentEdit(m.componentCreator.GetComponentType(), m.componentCreator.GetComponentName(), m.componentCreator.GetComponentContent())
	}
	
	return "Unknown creation step"
}







// exitConfirmationView is replaced by the confirmation module
/* func (m *MainListModel) exitConfirmationView() string {
	// Styles matching the rest of the UI
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1)
	
	headerStyle := TypeHeaderStyle // Orange like other headers
		
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
} */

