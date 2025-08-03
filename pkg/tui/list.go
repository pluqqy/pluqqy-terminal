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
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
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
	pipelinesViewport  viewport.Model
	componentsViewport viewport.Model
	showPreview        bool
	
	// Window dimensions
	width              int
	height             int
	
	// Error handling
	err                error
	
	// Confirmations
	deleteConfirm     *ConfirmationModel
	deletingFromPane  pane // Track which pane initiated the delete
	archiveConfirm    *ConfirmationModel
	archivingFromPane pane // Track which pane initiated the archive
	
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
	
	// Exit confirmation
	exitConfirm          *ConfirmationModel
	exitConfirmationType string // "component" or "component-edit"
	
	// Tag editing state
	editingTags           bool
	editingTagsPath       string
	editingTagsType       string // "component" or "pipeline"
	currentTags           []string
	tagInput              string
	tagCursor             int
	availableTags         []string
	showTagSuggestions    bool
	tagCloudActive        bool   // Whether tag cloud pane is active
	tagCloudCursor        int    // Selected tag in cloud
	
	// Tag deletion
	tagDeleteConfirm  *ConfirmationModel
	deletingTag       string
	deletingTagUsage  *tags.UsageStats
	
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
		showPreview:        false, // Start with preview hidden, user can toggle with 'p'
		previewViewport:    viewport.New(80, 20), // Default size
		pipelinesViewport:  viewport.New(40, 20), // Default size
		componentsViewport: viewport.New(40, 20), // Default size
		searchBar:          NewSearchBar(),
		deleteConfirm:      NewConfirmation(),
		archiveConfirm:     NewConfirmation(),
		exitConfirm:        NewConfirmation(),
		tagDeleteConfirm:   NewConfirmation(),
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
	// Create search engine and build index
	engine := search.NewEngine()
	if err := engine.BuildIndex(); err != nil {
		// Log error but don't fail - search will be unavailable
		return
	}
	m.searchEngine = engine
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
		
		// Build filtered lists from search results
		m.filteredPipelines = []pipelineItem{}
		m.filteredComponents = []componentItem{}
		
		for _, result := range results {
			if result.Item.Type == search.ItemTypePipeline {
				// Find matching pipeline - check if already added to avoid duplicates
				alreadyAdded := false
				for _, existing := range m.filteredPipelines {
					if existing.name == result.Item.Name {
						alreadyAdded = true
						break
					}
				}
				
				if !alreadyAdded {
					// Match by pipeline name (without .yaml extension)
					searchName := strings.TrimSuffix(result.Item.Name, ".yaml")
					for _, p := range m.pipelines {
						if p.name == searchName || p.name == result.Item.Name {
							m.filteredPipelines = append(m.filteredPipelines, p)
							break
						}
					}
				}
			} else {
				// Find matching component - check if already added to avoid duplicates
				alreadyAdded := false
				for _, existing := range m.filteredComponents {
					if existing.path == result.Item.Path {
						alreadyAdded = true
						break
					}
				}
				
				if !alreadyAdded {
					allComps := m.getAllComponents()
					for _, c := range allComps {
						if c.path == result.Item.Path {
							m.filteredComponents = append(m.filteredComponents, c)
							break
						}
					}
				}
			}
		}
		
		// Reset cursors if they're out of bounds
		if m.pipelineCursor >= len(m.filteredPipelines) {
			m.pipelineCursor = 0
		}
		if m.componentCursor >= len(m.filteredComponents) {
			m.componentCursor = 0
		}
	}
}

func (m *MainListModel) startTagEditing(path string, currentTags []string, itemType string) {
	m.editingTags = true
	m.editingTagsPath = path
	m.editingTagsType = itemType
	m.currentTags = make([]string, len(currentTags))
	copy(m.currentTags, currentTags)
	m.tagInput = ""
	m.tagCursor = 0
	m.showTagSuggestions = false
	
	// Load available tags from registry
	m.loadAvailableTags()
}

func (m *MainListModel) loadAvailableTags() {
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
	for _, comp := range m.prompts {
		for _, tag := range comp.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.availableTags = append(m.availableTags, normalized)
				seenTags[normalized] = true
			}
		}
	}
	for _, comp := range m.contexts {
		for _, tag := range comp.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.availableTags = append(m.availableTags, normalized)
				seenTags[normalized] = true
			}
		}
	}
	for _, comp := range m.rules {
		for _, tag := range comp.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.availableTags = append(m.availableTags, normalized)
				seenTags[normalized] = true
			}
		}
	}
	
	// Add tags from pipelines
	for _, pipeline := range m.pipelines {
		for _, tag := range pipeline.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.availableTags = append(m.availableTags, normalized)
				seenTags[normalized] = true
			}
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

func (m *MainListModel) Init() tea.Cmd {
	return nil
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search input when search pane is active
		if m.activePane == searchPane && !m.editingTags && !m.creatingComponent && !m.editingComponent {
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
		
		// Handle delete confirmation
		if m.deleteConfirm.Active() {
			return m, m.deleteConfirm.Update(msg)
		}
		
		// Handle archive confirmation
		if m.archiveConfirm.Active() {
			return m, m.archiveConfirm.Update(msg)
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
		
		case "E":
			// Edit component in external editor
			if m.activePane == componentsPane {
				components := m.getCurrentComponents()
				if m.componentCursor >= 0 && m.componentCursor < len(components) {
					comp := components[m.componentCursor]
					return m, m.openInEditor(comp.path)
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
					m.startTagEditing(comp.path, comp.tags, "component")
				}
			} else if m.activePane == pipelinesPane {
				// Use filtered pipelines if search is active
				pipelines := m.filteredPipelines
				if m.pipelineCursor >= 0 && m.pipelineCursor < len(pipelines) {
					pipeline := pipelines[m.pipelineCursor]
					m.startTagEditing(pipeline.path, pipeline.tags, "pipeline")
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
		
		
		case "S":
			if m.activePane == pipelinesPane {
				// Set selected pipeline (generate PLUQQY.md)
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					return m, m.setPipeline(m.pipelines[m.pipelineCursor].path)
				}
			}
		
		case "d", "delete":
			if m.activePane == pipelinesPane {
				// Delete pipeline with confirmation
				if len(m.pipelines) > 0 && m.pipelineCursor < len(m.pipelines) {
					m.deletingFromPane = pipelinesPane
					pipelineName := m.pipelines[m.pipelineCursor].name
					pipelinePath := m.pipelines[m.pipelineCursor].path
					
					m.deleteConfirm.ShowInline(
						fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
						true, // destructive
						func() tea.Cmd {
							return m.deletePipeline(pipelinePath)
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
					
					m.deleteConfirm.ShowInline(
						fmt.Sprintf("Delete %s '%s'?", comp.compType, comp.name),
						true, // destructive
						func() tea.Cmd {
							return m.deleteComponent(comp)
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
					
					m.archiveConfirm.ShowInline(
						fmt.Sprintf("Archive pipeline '%s'?", pipelineName),
						true, // destructive
						func() tea.Cmd {
							return m.archivePipeline(pipelinePath)
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
					
					m.archiveConfirm.ShowInline(
						fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
						true, // destructive
						func() tea.Cmd {
							return m.archiveComponent(comp)
						},
						nil, // onCancel
					)
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
		Foreground(lipgloss.Color("245"))

	typeHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

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
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))
	
	// Table column widths (adjusted for column width)
	nameWidth := 22
	tagsWidth := 18
	tokenWidth := 8  // For "~Tokens" plus padding
	usageWidth := 10
	
	// Render table header with 2-space shift - extra space between Tags and ~Tokens
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
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")). // Orange
				Bold(true)
			
			// Check if we have components but they're filtered out
			allComponents := m.getAllComponents()
			if len(allComponents) > 0 && m.searchQuery != "" {
				componentsScrollContent.WriteString(emptyStyle.Render("No components match your search."))
			} else {
				componentsScrollContent.WriteString(emptyStyle.Render("No components found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			
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
		nameStr := comp.name
		if len(nameStr) > nameWidth-3 {
			nameStr = nameStr[:nameWidth-6] + "..."
		}
		
		
		// Format usage count
		usageStr := fmt.Sprintf("%d", comp.usageCount)
		
		// Format token count - right-aligned with consistent width
		tokenStr := fmt.Sprintf("%d", comp.tokenCount)
		
		// Format tags
		tagsStr := renderTagChipsWithWidth(comp.tags, tagsWidth, 2) // Show max 2 tags inline
		
		// Build the row components separately
		// Use padding to ensure consistent column alignment
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
			row = "▸ " + selectedStyle.Render(namePart) + " " + tagsPart + "  " + normalStyle.Render(tokenPart + " " + usagePart)
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
	// Wrap content to viewport width to prevent overflow
	wrappedComponentsContent := wordwrap.String(componentsScrollContent.String(), m.componentsViewport.Width)
	m.componentsViewport.SetContent(wrappedComponentsContent)
	
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
	pipelineHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))
	
	// Table column widths for pipelines
	pipelineNameWidth := 25
	pipelineTagsWidth := 15
	pipelineTokenWidth := 8  // For "~Tokens"
	
	// Render table header
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
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")). // Orange
				Bold(true)
			
			// Check if we have pipelines but they're filtered out
			if len(m.pipelines) > 0 && m.searchQuery != "" {
				pipelinesScrollContent.WriteString(emptyStyle.Render("No pipelines match your search."))
			} else {
				pipelinesScrollContent.WriteString(emptyStyle.Render("No pipelines found.\n\nPress 'n' to create one"))
			}
		} else {
			// Inactive pane - show dimmed message
			dimmedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
			
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
			nameStr := pipeline.name
			if len(nameStr) > pipelineNameWidth-3 {
				nameStr = nameStr[:pipelineNameWidth-6] + "..."
			}
			
			// Format tags
			tagsStr := renderTagChipsWithWidth(pipeline.tags, pipelineTagsWidth, 2) // Show max 2 tags inline
			
			// Format token count - right-aligned
			tokenStr := fmt.Sprintf("%d", pipeline.tokenCount)
			
			// Build the row components separately
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
				row = "▸ " + selectedStyle.Render(namePart) + " " + tagsPart + " " + normalStyle.Render(tokenPart)
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
	
	// Add search bar first
	s.WriteString(m.searchBar.View())
	s.WriteString("\n")
	
	// Then add the columns
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
		
		// Determine what we're previewing based on cursor position
		// This maintains the preview type even when preview pane is active
		if len(m.pipelines) > 0 && m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			// We have a valid pipeline selected
			pipelineName := m.pipelines[m.pipelineCursor].name
			previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
		} else if len(m.filteredComponents) > 0 && m.componentCursor >= 0 && m.componentCursor < len(m.filteredComponents) {
			// We have a valid component selected
			previewHeading = "COMPONENT PREVIEW"
		} else {
			// No valid selection - use generic preview
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
	if m.deleteConfirm.Active() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(confirmStyle.Render(m.deleteConfirm.ViewWithWidth(m.width - 4)))
	}
	
	// Show archive confirmation if active
	if m.archiveConfirm.Active() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange for archive
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(confirmStyle.Render(m.archiveConfirm.ViewWithWidth(m.width - 4)))
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
	
	if m.activePane == pipelinesPane {
		// Show pipeline preview
		if len(m.pipelines) == 0 {
			m.previewContent = "No pipelines to preview."
			return
		}
		
		if m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			pipelinePath := m.pipelines[m.pipelineCursor].path
			
			// Load pipeline
			pipeline, err := files.ReadPipeline(pipelinePath)
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
		components := m.getCurrentComponents()
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

		// Load settings for output path
		settings, _ := files.ReadSettings()
		if settings == nil {
			settings = models.DefaultSettings()
		}
		
		// Write to configured output file
		outputPath := pipeline.OutputPath
		if outputPath == "" {
			outputPath = filepath.Join(settings.Output.ExportPath, settings.Output.DefaultFilename)
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

func (m *MainListModel) deleteComponent(comp componentItem) tea.Cmd {
	return func() tea.Msg {
		// Delete the component file
		err := files.DeleteComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete %s '%s': %v", comp.compType, comp.name, err))
		}
		
		// Reload the component list
		m.loadComponents()
		
		// Adjust cursor if necessary
		components := m.getAllComponents()
		if m.componentCursor >= len(components) && m.componentCursor > 0 {
			m.componentCursor = len(components) - 1
		}
		
		return StatusMsg(fmt.Sprintf("✓ Deleted %s: %s", comp.compType, comp.name))
	}
}

func (m *MainListModel) archivePipeline(pipelineName string) tea.Cmd {
	return func() tea.Msg {
		// Archive the pipeline file
		err := files.ArchivePipeline(pipelineName)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive pipeline '%s': %v", pipelineName, err))
		}
		
		// Reload the pipeline list
		m.loadPipelines()
		
		// Adjust cursor if necessary
		if m.pipelineCursor >= len(m.pipelines) && m.pipelineCursor > 0 {
			m.pipelineCursor = len(m.pipelines) - 1
		}
		
		return StatusMsg(fmt.Sprintf("✓ Archived pipeline: %s", pipelineName))
	}
}

func (m *MainListModel) archiveComponent(comp componentItem) tea.Cmd {
	return func() tea.Msg {
		// Archive the component file
		err := files.ArchiveComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive %s '%s': %v", comp.compType, comp.name, err))
		}
		
		// Reload the component list
		m.loadComponents()
		
		// Adjust cursor if necessary
		components := m.getAllComponents()
		if m.componentCursor >= len(components) && m.componentCursor > 0 {
			m.componentCursor = len(components) - 1
		}
		
		return StatusMsg(fmt.Sprintf("✓ Archived %s: %s", comp.compType, comp.name))
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
				m.exitConfirmationType = "component"
				m.exitConfirm.ShowDialog(
					"⚠️  Unsaved Changes",
					"You have unsaved content in this component.",
					"Exit without saving?",
					true, // destructive
					m.width - 4,
					10,
					func() tea.Cmd {
						// Exit - go back to type selection
						m.creationStep = 0
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

func (m *MainListModel) handleTagEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle tag deletion confirmation
	if m.tagDeleteConfirm.Active() {
		return m, m.tagDeleteConfirm.Update(msg)
	}
	
	switch msg.String() {
	case "esc":
		// Cancel tag editing
		m.editingTags = false
		m.tagInput = ""
		m.currentTags = nil
		m.showTagSuggestions = false
		m.tagCloudActive = false
		m.tagCloudCursor = 0
		return m, nil
		
	case "ctrl+s":
		// Save tags
		return m, m.saveTags()
		
	case "enter":
		if m.tagCloudActive {
			// Add tag from cloud
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor >= 0 && m.tagCloudCursor < len(availableForSelection) {
				tag := availableForSelection[m.tagCloudCursor]
				if !m.hasTag(tag) {
					m.currentTags = append(m.currentTags, tag)
				}
			}
		} else {
			// Add tag from input
			if m.tagInput != "" {
				// Normalize the tag
				normalized := models.NormalizeTagName(m.tagInput)
				if normalized != "" && !m.hasTag(normalized) {
					m.currentTags = append(m.currentTags, normalized)
				}
				m.tagInput = ""
				m.showTagSuggestions = false
			}
		}
		return m, nil
		
	case "tab":
		if m.tagCloudActive {
			// Switch back to main pane
			m.tagCloudActive = false
			m.tagInput = ""
		} else {
			// Switch to tag cloud pane
			m.tagCloudActive = true
			m.tagCloudCursor = 0
			m.showTagSuggestions = false
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
		
	case "ctrl+d":
		if m.tagCloudActive {
			// Delete tag from registry (only in tag cloud)
			availableForSelection := m.getAvailableTagsForCloud()
			if m.tagCloudCursor >= 0 && m.tagCloudCursor < len(availableForSelection) {
				tagToDelete := availableForSelection[m.tagCloudCursor]
				
				// Get usage stats
				usage, err := tags.CountTagUsage(tagToDelete)
				if err != nil {
					return m, func() tea.Msg {
						return StatusMsg(fmt.Sprintf("× Failed to check tag usage: %v", err))
					}
				}
				
				m.deletingTag = tagToDelete
				m.deletingTagUsage = usage
				
				// Show tag deletion confirmation with details
				var details []string
				if usage.PipelineCount > 0 {
					details = append(details, fmt.Sprintf("Used in %d pipeline(s)", usage.PipelineCount))
				}
				if usage.ComponentCount > 0 {
					details = append(details, fmt.Sprintf("Used in %d component(s)", usage.ComponentCount))
				}
				
				warning := ""
				if usage.PipelineCount > 0 || usage.ComponentCount > 0 {
					warning = "The tag will be removed from the registry but will remain on items that use it."
				}
				
				m.tagDeleteConfirm.Show(ConfirmationConfig{
					Title:       "⚠️  Delete Tag from Registry?",
					Message:     fmt.Sprintf("Delete tag '%s'?", tagToDelete),
					Warning:     warning,
					Details:     details,
					Destructive: true,
					Type:        ConfirmTypeDialog,
					Width:       m.width - 4,
					Height:      12,
				}, func() tea.Cmd {
					return m.deleteTagFromRegistry()
				}, func() tea.Cmd {
					m.deletingTag = ""
					m.deletingTagUsage = nil
					return nil
				})
			}
		} else {
			// Remove tag from current item (only in main pane when not typing)
			if m.tagInput == "" && len(m.currentTags) > 0 {
				if m.tagCursor >= 0 && m.tagCursor < len(m.currentTags) {
					m.currentTags = append(m.currentTags[:m.tagCursor], m.currentTags[m.tagCursor+1:]...)
					if m.tagCursor >= len(m.currentTags) && m.tagCursor > 0 {
						m.tagCursor--
					}
				}
			}
		}
		return m, nil
		
	case "backspace", "delete":
		if !m.tagCloudActive && m.tagInput != "" {
			// Delete from input
			if len(m.tagInput) > 0 {
				m.tagInput = m.tagInput[:len(m.tagInput)-1]
				m.showTagSuggestions = len(m.tagInput) > 0
			}
		}
		return m, nil
		
	default:
		// Add to input only when in main pane
		if !m.tagCloudActive && len(msg.String()) == 1 {
			m.tagInput += msg.String()
			m.showTagSuggestions = true
		}
		return m, nil
	}
}

func (m *MainListModel) saveTags() tea.Cmd {
	return func() tea.Msg {
		var err error
		
		if m.editingTagsType == "component" {
			// Update component tags
			err = files.UpdateComponentTags(m.editingTagsPath, m.currentTags)
		} else {
			// Update pipeline tags
			pipeline, err := files.ReadPipeline(m.editingTagsPath)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to read pipeline: %v", err))
			}
			pipeline.Tags = m.currentTags
			err = files.WritePipeline(pipeline)
		}
		
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save tags: %v", err))
		}
		
		// Update the registry with any new tags
		registry, _ := tags.NewRegistry()
		if registry != nil {
			for _, tag := range m.currentTags {
				registry.GetOrCreateTag(tag)
			}
			registry.Save()
		}
		
		// Reload data to reflect changes
		m.loadComponents()
		m.loadPipelines()
		
		// Re-run search to update filtered lists
		m.performSearch()
		
		// Exit tag editing mode
		m.editingTags = false
		m.tagInput = ""
		m.currentTags = nil
		m.showTagSuggestions = false
		
		return StatusMsg("✓ Tags saved successfully")
	}
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
	case "E":
		// Save current content and open in external editor
		// First save any unsaved changes
		if m.componentContent != m.originalContent {
			err := files.WriteComponent(m.editingComponentPath, m.componentContent)
			if err != nil {
				return m, func() tea.Msg {
					return StatusMsg(fmt.Sprintf("× Failed to save before external edit: %v", err))
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
					m.originalContent = ""
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
		BorderForeground(lipgloss.Color("170"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
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

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane

	heading := "CREATE NEW COMPONENT"
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Component type selection
	contentPadding := headerPadding
	mainContent.WriteString(contentPadding.Render("Select component type:"))
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
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

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
		BorderForeground(lipgloss.Color("170"))

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

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane

	heading := fmt.Sprintf("CREATE NEW %s", strings.ToUpper(m.componentCreationType))
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active single pane
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
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

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
		BorderForeground(lipgloss.Color("170"))

	// Calculate dimensions
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane

	heading := fmt.Sprintf("EDIT %s: %s", strings.ToUpper(m.componentCreationType), m.componentName)
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active single pane
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
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

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

func (m *MainListModel) tagEditView() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1).
		Width(40)
	suggestionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	selectedSuggestionStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("170"))
		
	// Calculate dimensions for side-by-side layout
	paneWidth := (m.width - 6) / 2 // Same calculation as main list view columns
	paneHeight := m.height - 10 // Leave room for help pane
	
	// Header padding
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Build main content
	var mainContent strings.Builder
	
	// Show deletion confirmation if active
	if m.tagDeleteConfirm.Active() {
		// Use the confirmation module's dialog view
		return m.tagDeleteConfirm.View()
	}
	
	// Original tag edit view code continues below
	if false { // Keep old code for reference
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange for warning
		
		mainContent.WriteString(headerPadding.Render(confirmStyle.Render("⚠️  Delete Tag from Registry?")))
		mainContent.WriteString("\n\n")
		
		mainContent.WriteString(headerPadding.Render(fmt.Sprintf("Tag: %s", titleStyle.Render(m.deletingTag))))
		mainContent.WriteString("\n\n")
		
		if m.deletingTagUsage != nil && m.deletingTagUsage.TotalCount > 0 {
			mainContent.WriteString(headerPadding.Render(warningStyle.Render("This tag is currently in use:")))
			mainContent.WriteString("\n")
			if m.deletingTagUsage.ComponentCount > 0 {
				mainContent.WriteString(headerPadding.Render(fmt.Sprintf("  • %d component%s", 
					m.deletingTagUsage.ComponentCount,
					pluralize(m.deletingTagUsage.ComponentCount))))
				mainContent.WriteString("\n")
			}
			if m.deletingTagUsage.PipelineCount > 0 {
				mainContent.WriteString(headerPadding.Render(fmt.Sprintf("  • %d pipeline%s", 
					m.deletingTagUsage.PipelineCount,
					pluralize(m.deletingTagUsage.PipelineCount))))
				mainContent.WriteString("\n")
			}
			mainContent.WriteString("\n")
			mainContent.WriteString(headerPadding.Render(warningStyle.Render("The tag will be removed from the registry but will remain on items that use it.")))
		} else {
			mainContent.WriteString(headerPadding.Render("This tag is not currently in use."))
		}
		mainContent.WriteString("\n\n")
		
		// Show styled confirmation options
		deleteOptions := formatConfirmOptions(true) + "  (delete / cancel)"
		centeredOptions := lipgloss.NewStyle().
			Width(paneWidth - 4).
			Align(lipgloss.Center).
			Render(deleteOptions)
		mainContent.WriteString(centeredOptions)
		
		// Apply border and return early
		confirmBorderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")) // Red border for deletion
		mainPane := confirmBorderStyle.
			Width(m.width - 4). // Use full width minus padding, like help pane
			Height(paneHeight).
			Render(mainContent.String())
			
		// Still show help
		help := []string{
			"y confirm delete",
			"n/esc cancel",
		}
		
		helpBorderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Width(m.width - 4).
			Padding(0, 1)
		
		helpContent := formatHelpText(help)
		
		var s strings.Builder
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
			
		s.WriteString(contentStyle.Render(mainPane))
		s.WriteString("\n")
		s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))
		
		return s.String()
	} // end if false
	
	// Get item name
	itemName := ""
	if m.editingTagsType == "component" {
		components := m.getCurrentComponents()
		if m.componentCursor >= 0 && m.componentCursor < len(components) {
			itemName = components[m.componentCursor].name
		}
	} else {
		if m.pipelineCursor >= 0 && m.pipelineCursor < len(m.pipelines) {
			itemName = m.pipelines[m.pipelineCursor].name
		}
	}
	
	// Use ViewTitle component but with dynamic coloring based on active pane
	heading := fmt.Sprintf("EDIT TAGS: %s", strings.ToUpper(itemName))
	// For tag editor, we'll render the title manually with dynamic colors
	remainingWidth := paneWidth - len(heading) - 7 // Adjust for smaller pane width
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	// Dynamic styles based on which pane is active
	mainHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if !m.tagCloudActive {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	mainColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if !m.tagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	mainContent.WriteString(headerPadding.Render(mainHeaderStyle.Render(heading) + " " + mainColonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")
	
	// Current tags
	mainContent.WriteString(headerPadding.Render("Current tags:\n"))
	if len(m.currentTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		mainContent.WriteString(headerPadding.Render(dimStyle.Render("(no tags)")))
	} else {
		var tagDisplay strings.Builder
		for i, tag := range m.currentTags {
			// Get color from registry
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}
			
			style := lipgloss.NewStyle().
				Background(lipgloss.Color(color)).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
			
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
			
			// Add space between tags
			if i < len(m.currentTags)-1 {
				tagDisplay.WriteString("  ") // Double space for better separation
			}
		}
		mainContent.WriteString(headerPadding.Render(tagDisplay.String()))
	}
	mainContent.WriteString("\n\n")
	
	// Input field
	mainContent.WriteString(headerPadding.Render("Add tag:"))
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
	mainContent.WriteString("\n\n")
	
	// Suggestions
	if m.showTagSuggestions && len(m.tagInput) > 0 {
		mainContent.WriteString(headerPadding.Render("Suggestions:\n"))
		suggestions := m.getTagSuggestions()
		for i, suggestion := range suggestions {
			if i > 5 { // Limit to 6 suggestions
				break
			}
			var suggestionLine string
			if i == 0 {
				// Selected suggestion with background
				suggestionLine = selectedSuggestionStyle.
					Padding(0, 1). // Add padding inside the styled area
					Render(suggestion)
			} else {
				// Regular suggestion
				suggestionLine = suggestionStyle.Render("  " + suggestion)
			}
			mainContent.WriteString(headerPadding.Render(suggestionLine))
			mainContent.WriteString("\n")
		}
		mainContent.WriteString("\n")
	}
	
	// Apply border to main content
	activeBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))
	inactiveBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))
	
	mainPaneBorder := inactiveBorderStyle
	if !m.tagCloudActive {
		mainPaneBorder = activeBorderStyle
	}
	
	mainPane := mainPaneBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(mainContent.String())

	// Build tag cloud pane
	var tagCloudContent strings.Builder
	
	// Tag cloud header
	tagCloudTitle := "AVAILABLE TAGS"
	tagCloudRemainingWidth := paneWidth - len(tagCloudTitle) - 7 // Adjust for smaller width
	if tagCloudRemainingWidth < 0 {
		tagCloudRemainingWidth = 0
	}
	// Dynamic styles based on which pane is active
	tagCloudHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if m.tagCloudActive {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	tagCloudColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if m.tagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	tagCloudContent.WriteString(headerPadding.Render(tagCloudHeaderStyle.Render(tagCloudTitle) + " " + tagCloudColonStyle.Render(strings.Repeat(":", tagCloudRemainingWidth))))
	tagCloudContent.WriteString("\n\n")
	
	// Display available tags
	if len(m.availableTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		tagCloudContent.WriteString(headerPadding.Render(dimStyle.Render("(no available tags)")))
	} else {
		// Group tags in rows
		var tagRows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		maxRowWidth := paneWidth - 6 // Account for padding
		
		availableForCloud := m.getAvailableTagsForCloud()
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
				tagDisplay = "  " + tagStyle.Render(tag) + "  "
			}
			
			tagWidth := lipgloss.Width(tagDisplay) + 2 // Add spacing
			
			// Check if we need a new row
			if rowTags > 0 && currentRowWidth + tagWidth > maxRowWidth {
				tagRows.WriteString("\n\n") // Double newline for vertical spacing
				rowTags = 0
				currentRowWidth = 0
			}
			
			tagRows.WriteString(tagDisplay)
			tagRows.WriteString("  ")
			currentRowWidth += tagWidth + 2
			rowTags++
		}
		
		tagCloudContent.WriteString(headerPadding.Render(tagRows.String()))
	}
	
	tagCloudBorder := inactiveBorderStyle
	if m.tagCloudActive {
		tagCloudBorder = activeBorderStyle
	}
	
	tagCloudPane := tagCloudBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(tagCloudContent.String())

	// Help section
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

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

	// Combine panes side by side
	sideBySide := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainPane,
		" ", // Single space gap, same as main list view
		tagCloudPane,
	)

	// Combine all elements
	var s strings.Builder

	// Add padding around content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(sideBySide))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

func (m *MainListModel) getTagSuggestions() []string {
	if m.tagInput == "" {
		return []string{}
	}
	
	input := strings.ToLower(m.tagInput)
	var suggestions []string
	
	// First, exact prefix matches
	for _, tag := range m.availableTags {
		if strings.HasPrefix(strings.ToLower(tag), input) && !m.hasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	// Then, contains matches
	for _, tag := range m.availableTags {
		if !strings.HasPrefix(strings.ToLower(tag), input) && 
		   strings.Contains(strings.ToLower(tag), input) && 
		   !m.hasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	return suggestions
}

func (m *MainListModel) hasTag(tag string) bool {
	normalized := models.NormalizeTagName(tag)
	for _, t := range m.currentTags {
		if models.NormalizeTagName(t) == normalized {
			return true
		}
	}
	return false
}

func (m *MainListModel) getAvailableTagsForCloud() []string {
	var available []string
	for _, tag := range m.availableTags {
		if !m.hasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

func (m *MainListModel) deleteTagFromRegistry() tea.Cmd {
	return func() tea.Msg {
		// Delete from registry
		registry, err := tags.NewRegistry()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to load tag registry: %v", err))
		}
		
		registry.RemoveTag(m.deletingTag)
		
		if err := registry.Save(); err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save tag registry: %v", err))
		}
		
		// Update available tags
		m.loadAvailableTags()
		
		// Adjust cursor if needed
		availableForCloud := m.getAvailableTagsForCloud()
		if m.tagCloudCursor >= len(availableForCloud) && m.tagCloudCursor > 0 {
			m.tagCloudCursor = len(availableForCloud) - 1
		}
		
		// Clear deletion state
		deletedTag := m.deletingTag
		m.deletingTag = ""
		m.deletingTagUsage = nil
		
		return StatusMsg(fmt.Sprintf("✓ Deleted tag '%s' from registry", deletedTag))
	}
}

func (m *MainListModel) componentEditView() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	// Calculate dimensions  
	contentWidth := m.width - 4 // Match help pane width
	contentHeight := m.height - 7 // Reserve space for help pane (3) + spacing (3) + status bar (1)

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane

	heading := fmt.Sprintf("EDITING: %s", strings.ToUpper(m.editingComponentName))
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active single pane
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
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

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

// exitConfirmationView is replaced by the confirmation module
/* func (m *MainListModel) exitConfirmationView() string {
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
		m.loadComponents()
		
		return StatusMsg(fmt.Sprintf("✓ Created %s: %s", m.componentCreationType, filename))
	}
}

func (m *MainListModel) saveEditedComponent() tea.Cmd {
	return func() tea.Msg {
		// Write component
		err := files.WriteComponent(m.editingComponentPath, m.componentContent)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save: %v", err))
		}
		
		// Clear editing state
		m.editingComponent = false
		m.componentContent = ""
		m.editingComponentPath = ""
		m.editingComponentName = ""
		m.originalContent = ""
		
		// Reload components
		m.loadComponents()
		
		// Re-run search to update filtered lists
		m.performSearch()
		
		return StatusMsg(fmt.Sprintf("✓ Saved: %s", m.editingComponentName))
	}
}

func (m *MainListModel) openInEditor(path string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set. Please set it to your preferred editor.")
		}

		// Validate editor path to prevent command injection
		if strings.ContainsAny(editor, "&|;<>()$`\\\"'") {
			return StatusMsg("Invalid EDITOR value: contains shell metacharacters")
		}

		// Construct full path
		fullPath := filepath.Join(files.PluqqyDir, path)
		
		// Execute editor
		cmd := exec.Command(editor, fullPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}

		// Reload components to reflect any changes
		m.loadComponents()

		return StatusMsg(fmt.Sprintf("Edited: %s", filepath.Base(path)))
	}
}

