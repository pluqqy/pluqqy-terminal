package tui

import (
	"path/filepath"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

// NewMainListModel creates and initializes a new MainListModel
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

// Init initializes the MainListModel
func (m *MainListModel) Init() tea.Cmd {
	return nil
}

// loadPipelines loads all pipeline files and their metadata
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

// loadComponents loads all component files and their metadata
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

// initializeSearchEngine sets up the search engine
func (m *MainListModel) initializeSearchEngine() {
	// Use SearchManager for initialization
	searchManager := NewSearchManager()
	if err := searchManager.InitializeEngine(); err != nil {
		// Log error but don't fail - search will be unavailable
		return
	}
	m.searchEngine = searchManager.GetEngine()
}

// updateViewportSizes updates all viewport dimensions based on window size
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