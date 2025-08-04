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
	// State management
	stateManager       *StateManager
	
	// Pipelines data
	pipelines          []pipelineItem
	
	// Components data
	prompts            []componentItem
	contexts           []componentItem
	rules              []componentItem
	
	// Preview data
	previewContent     string
	previewViewport    viewport.Model
	
	// UI state
	pipelinesViewport  viewport.Model
	componentsViewport viewport.Model
	
	// Window dimensions
	width              int
	height             int
	
	// Error handling
	err                error
	
	// Pipeline operations
	pipelineOperator  *PipelineOperator
	
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
		stateManager:       NewStateManager(),
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
	// Set initial preview state
	m.stateManager.ShowPreview = false // Start with preview hidden, user can toggle with 'p'
	m.loadPipelines()
	m.loadComponents()
	m.initializeSearchEngine()
	// Initialize filtered lists with all items
	m.filteredPipelines = m.pipelines
	m.filteredComponents = m.getAllComponents()
	
	// Update state manager with counts after both are loaded
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.pipelines))
	
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
	
	// Update state manager counts if components are already loaded
	if m.prompts != nil || m.contexts != nil || m.rules != nil {
		m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.pipelines))
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
	
	// Update state manager counts
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.pipelines))
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
		m.stateManager.ResetCursorsAfterSearch(len(m.filteredComponents), len(m.filteredPipelines))
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
		if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
			return components[m.stateManager.ComponentCursor].name
		}
	} else {
		if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
			return m.pipelines[m.stateManager.PipelineCursor].name
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
		if m.stateManager.IsInSearchPane() && !m.tagEditor.Active && !m.componentCreator.IsActive() && !m.componentEditor.IsActive() {
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
				m.stateManager.ExitSearch()
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
			// Handle tab navigation
			if m.stateManager.IsInSearchPane() {
				m.searchBar.SetActive(false)
			}
			m.stateManager.HandleTabNavigation(false)
			// Update preview when switching to non-preview pane
			if m.stateManager.ActivePane != previewPane && m.stateManager.ActivePane != searchPane {
				m.updatePreview()
			}
		
		case "up", "k":
			handled, updatePreview := m.stateManager.HandleKeyNavigation(msg.String())
			if handled {
				if updatePreview {
					m.updatePreview()
				}
			} else if m.stateManager.IsInPreviewPane() {
				// Scroll preview up
				m.previewViewport.LineUp(1)
			}
		
		case "down", "j":
			handled, updatePreview := m.stateManager.HandleKeyNavigation(msg.String())
			if handled {
				if updatePreview {
					m.updatePreview()
				}
			} else if m.stateManager.IsInPreviewPane() {
				// Scroll preview down
				m.previewViewport.LineDown(1)
			}
		
		case "pgup":
			if m.stateManager.IsInPreviewPane() {
				m.previewViewport.ViewUp()
			}
		
		case "pgdown":
			if m.stateManager.IsInPreviewPane() {
				m.previewViewport.ViewDown()
			}
		
		case "p":
			m.stateManager.ShowPreview = !m.stateManager.ShowPreview
			m.updateViewportSizes()
			if m.stateManager.ShowPreview {
				m.updatePreview()
			}
		
		case "/":
			// Jump to search
			m.stateManager.SwitchToSearch()
			m.searchBar.SetActive(true)
			return m, nil
		
		case "enter":
			if m.stateManager.ActivePane == pipelinesPane {
				if len(m.pipelines) > 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
					// View the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineViewerView,
							pipeline: m.pipelines[m.stateManager.PipelineCursor].path, // Use path (filename) not name
						}
					}
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Could add component viewing/editing functionality here
			}
		
		case "e":
			if m.stateManager.ActivePane == pipelinesPane {
				if len(m.pipelines) > 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
					// Edit the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineBuilderView,
							pipeline: m.pipelines[m.stateManager.PipelineCursor].path, // Use path (filename) not name
						}
					}
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Edit component in TUI editor
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
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
			if m.stateManager.ActivePane == componentsPane {
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					return m, m.pipelineOperator.OpenInEditor(comp.path, m.loadComponents)
				}
			}
			// Explicitly do nothing for pipelines pane - editing YAML directly is not encouraged
		
		case "t":
			// Edit tags
			if m.stateManager.ActivePane == componentsPane {
				// Use filtered components if search is active
				components := m.filteredComponents
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					m.tagEditor.Start(comp.path, comp.tags, "component")
				}
			} else if m.stateManager.ActivePane == pipelinesPane {
				// Use filtered pipelines if search is active
				pipelines := m.filteredPipelines
				if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					m.tagEditor.Start(pipeline.path, pipeline.tags, "pipeline")
				}
			}
		
		case "n":
			if m.stateManager.ActivePane == pipelinesPane {
				// Create new pipeline (switch to builder)
				return m, func() tea.Msg {
					return SwitchViewMsg{
						view: pipelineBuilderView,
					}
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Create new component
				m.componentCreator.Start()
				return m, nil
			}
		
		
		case "S":
			if m.stateManager.ActivePane == pipelinesPane {
				// Set selected pipeline (generate PLUQQY.md)
				if len(m.pipelines) > 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
					return m, m.pipelineOperator.SetPipeline(m.pipelines[m.stateManager.PipelineCursor].path)
				}
			}
		
		case "d", "delete":
			if m.stateManager.ActivePane == pipelinesPane {
				// Delete pipeline with confirmation
				if len(m.pipelines) > 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
					m.stateManager.SetDeletingFromPane(pipelinesPane)
					pipelineName := m.pipelines[m.stateManager.PipelineCursor].name
					pipelinePath := m.pipelines[m.stateManager.PipelineCursor].path
					
					m.pipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return m.pipelineOperator.DeletePipeline(pipelinePath, m.loadPipelines)
						},
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return nil
						},
					)
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Delete component with confirmation
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					m.stateManager.SetDeletingFromPane(componentsPane)
					
					m.pipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete %s '%s'?", comp.compType, comp.name),
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return m.pipelineOperator.DeleteComponent(comp, m.loadComponents)
						},
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return nil
						},
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
			if m.stateManager.ActivePane == pipelinesPane {
				// Archive pipeline with confirmation
				if len(m.pipelines) > 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
					m.stateManager.SetArchivingFromPane(pipelinesPane)
					pipelineName := m.pipelines[m.stateManager.PipelineCursor].name
					pipelinePath := m.pipelines[m.stateManager.PipelineCursor].path
					
					m.pipelineOperator.ShowArchiveConfirmation(
						fmt.Sprintf("Archive pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							m.stateManager.ClearArchiveState()
							return m.pipelineOperator.ArchivePipeline(pipelinePath, m.loadPipelines)
						},
						func() tea.Cmd {
							m.stateManager.ClearArchiveState()
							return nil
						},
					)
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Archive component with confirmation
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					m.stateManager.SetArchivingFromPane(componentsPane)
					
					m.pipelineOperator.ShowArchiveConfirmation(
						fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
						func() tea.Cmd {
							m.stateManager.ClearArchiveState()
							return m.pipelineOperator.ArchiveComponent(comp, m.loadComponents)
						},
						func() tea.Cmd {
							m.stateManager.ClearArchiveState()
							return nil
						},
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
	if m.stateManager.ShowPreview && m.previewContent != "" {
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
		if m.stateManager.ShowPreview {
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
	// Create main view renderer
	mainRenderer := NewMainViewRenderer(m.width, m.height)
	mainRenderer.ShowPreview = m.stateManager.ShowPreview
	mainRenderer.ActivePane = m.stateManager.ActivePane
	mainRenderer.LastDataPane = m.stateManager.LastDataPane
	mainRenderer.SearchBar = m.searchBar
	mainRenderer.PreviewViewport = m.previewViewport
	mainRenderer.PreviewContent = m.previewContent
	
	// Handle error state
	if m.err != nil {
		return mainRenderer.RenderErrorView(m.err)
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

	// Calculate content height
	contentHeight := mainRenderer.CalculateContentHeight()

	// Update search bar active state and render it
	m.searchBar.SetActive(m.stateManager.IsInSearchPane())
	m.searchBar.SetWidth(m.width)

	// Create component view renderer
	componentRenderer := NewComponentViewRenderer(m.width, contentHeight)
	componentRenderer.ActivePane = m.stateManager.ActivePane
	componentRenderer.FilteredComponents = m.filteredComponents
	componentRenderer.AllComponents = m.getAllComponents()
	componentRenderer.ComponentCursor = m.stateManager.ComponentCursor
	componentRenderer.SearchQuery = m.searchQuery
	componentRenderer.Viewport = m.componentsViewport

	// Create pipeline view renderer
	pipelineRenderer := NewPipelineViewRenderer(m.width, contentHeight)
	pipelineRenderer.ActivePane = m.stateManager.ActivePane
	pipelineRenderer.Pipelines = m.pipelines
	pipelineRenderer.FilteredPipelines = m.filteredPipelines
	pipelineRenderer.PipelineCursor = m.stateManager.PipelineCursor
	pipelineRenderer.SearchQuery = m.searchQuery
	pipelineRenderer.Viewport = m.pipelinesViewport

	// Render columns
	leftColumn := componentRenderer.Render()
	rightColumn := pipelineRenderer.Render()

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
	previewPane := mainRenderer.RenderPreviewPane(m.pipelines, m.filteredComponents, m.stateManager.PipelineCursor, m.stateManager.ComponentCursor)
	if previewPane != "" {
		s.WriteString(previewPane)
	}

	// Show confirmation dialogs
	confirmDialogs := mainRenderer.RenderConfirmationDialogs(m.pipelineOperator)
	if confirmDialogs != "" {
		s.WriteString(confirmDialogs)
	}
	
	// Help text
	s.WriteString("\n")
	s.WriteString(mainRenderer.RenderHelpPane(m.stateManager.IsInSearchPane()))

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
	
	if m.stateManager.ShowPreview {
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
	if m.stateManager.ShowPreview {
		previewHeight := m.height / 2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		m.previewViewport.Width = m.width - 8 // Account for outer padding (2), borders (2), and inner padding (2) + extra spacing
		m.previewViewport.Height = previewHeight
	}
}

func (m *MainListModel) updatePreview() {
	if !m.stateManager.ShowPreview {
		return
	}
	
	// Use PreviewRenderer for generating preview content
	renderer := &PreviewRenderer{ShowPreview: m.stateManager.ShowPreview}
	
	// Determine which pane to preview from
	previewPane := m.stateManager.GetPreviewPane()
	
	if previewPane == pipelinesPane {
		// Show pipeline preview
		if len(m.pipelines) == 0 {
			m.previewContent = renderer.RenderEmptyPreview(pipelinesPane, false, false)
			return
		}
		
		if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(m.pipelines) {
			pipelinePath := m.pipelines[m.stateManager.PipelineCursor].path
			m.previewContent = renderer.RenderPipelinePreview(pipelinePath)
		}
	} else if previewPane == componentsPane {
		// Show component preview
		components := m.getCurrentComponents()
		if len(components) == 0 {
			m.previewContent = renderer.RenderEmptyPreview(componentsPane, false, false)
			return
		}
		
		if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
			comp := components[m.stateManager.ComponentCursor]
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

