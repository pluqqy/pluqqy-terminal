package tui

import (
	"path/filepath"
	"strings"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

// NewMainListModel creates and initializes a new MainListModel
func NewMainListModel() *MainListModel {
	m := &MainListModel{
		stateManager:       NewStateManager(),
		businessLogic:      NewBusinessLogic(),
		previewViewport:    viewport.New(80, 20), // Default size
		pipelinesViewport:  viewport.New(40, 20), // Default size
		componentsViewport: viewport.New(40, 20), // Default size
		searchBar:          NewSearchBar(),
		pipelineOperator:   NewPipelineOperator(),
		exitConfirm:        NewConfirmation(),
		componentCreator:   NewComponentCreator(),
		componentEditor:    NewComponentEditor(),
		tagEditor:          NewTagEditor(),
		tagReloader:        NewTagReloader(),
		tagReloadRenderer:  NewTagReloadRenderer(80, 20), // Default size
	}
	// Set initial preview state
	m.stateManager.ShowPreview = false // Start with preview hidden, user can toggle with 'p'
	m.loadPipelines()
	m.loadComponents()
	m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
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
	// Check if we should include archived items based on search query
	includeArchived := m.shouldIncludeArchived()
	
	pipelineFiles, err := files.ListPipelines()
	if err != nil {
		m.err = err
		return
	}
	
	m.pipelines = nil
	
	// Load active pipelines
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
			isArchived: false,
		})
	}
	
	// Load archived pipelines if needed
	if includeArchived {
		archivedFiles, _ := files.ListArchivedPipelines()
		for _, pipelineFile := range archivedFiles {
			// Load pipeline to get metadata
			pipeline, err := files.ReadArchivedPipeline(pipelineFile)
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
				isArchived: true,
			})
		}
	}
	
	// Rebuild search index when pipelines are reloaded
	if m.searchEngine != nil {
		if includeArchived {
			m.searchEngine.BuildIndexWithOptions(true)
		} else {
			m.searchEngine.BuildIndex()
		}
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

// shouldIncludeArchived checks if the current search query requires archived items
func (m *MainListModel) shouldIncludeArchived() bool {
	if m.searchQuery == "" {
		return false
	}
	
	// Parse the search query to check for status:archived
	parser := search.NewParser()
	query, err := parser.Parse(m.searchQuery)
	if err != nil {
		return false
	}
	
	for _, condition := range query.Conditions {
		if condition.Field == search.FieldStatus {
			if statusStr, ok := condition.Value.(string); ok && strings.ToLower(statusStr) == "archived" {
				return true
			}
		}
	}
	
	return false
}

// loadComponents loads all component files and their metadata
func (m *MainListModel) loadComponents() {
	// Check if we should include archived items based on search query
	includeArchived := m.shouldIncludeArchived()
	
	// Get usage counts for all components
	usageMap, _ := files.CountComponentUsage()
	
	// Clear existing components
	m.prompts = nil
	m.contexts = nil
	m.rules = nil
	
	// Load prompts
	m.prompts = m.loadComponentsOfType("prompts", files.PromptsDir, models.ComponentTypePrompt, usageMap, includeArchived)
	
	// Load contexts
	m.contexts = m.loadComponentsOfType("contexts", files.ContextsDir, models.ComponentTypeContext, usageMap, includeArchived)
	
	// Load rules
	m.rules = m.loadComponentsOfType("rules", files.RulesDir, models.ComponentTypeRules, usageMap, includeArchived)
	
	// Update business logic with new components
	m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
	
	// Rebuild search index when components are reloaded
	if m.searchEngine != nil {
		if includeArchived {
			m.searchEngine.BuildIndexWithOptions(true)
		} else {
			m.searchEngine.BuildIndex()
		}
	}
	
	// Update filtered list if no active search
	if m.searchQuery == "" {
		m.filteredComponents = m.getAllComponents()
	}
	
	// Update state manager counts
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.pipelines))
}

// loadComponentsOfType loads components of a specific type
func (m *MainListModel) loadComponentsOfType(compType, subDir, modelType string, usageMap map[string]int, includeArchived bool) []componentItem {
	var items []componentItem
	
	// Load active components
	components, _ := files.ListComponents(compType)
	for _, c := range components {
		componentPath := filepath.Join(files.ComponentsDir, subDir, c)
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
		
		items = append(items, componentItem{
			name:         c,
			path:         componentPath,
			compType:     modelType,
			lastModified: modTime,
			usageCount:   usage,
			tokenCount:   tokenCount,
			tags:         tags,
			isArchived:   false,
		})
	}
	
	// Load archived components if needed
	if includeArchived {
		archivedComponents, _ := files.ListArchivedComponents(compType)
		for _, c := range archivedComponents {
			componentPath := filepath.Join(files.ComponentsDir, subDir, c)
			
			// Read archived component
			component, _ := files.ReadArchivedComponent(componentPath)
			modTime := time.Time{}
			if component != nil {
				modTime = component.Modified
			}
			
			// Calculate usage count (archived components typically have 0 usage)
			usage := 0
			
			// Get token count
			tokenCount := 0
			if component != nil {
				tokenCount = utils.EstimateTokens(component.Content)
			}
			
			tags := []string{}
			if component != nil {
				tags = component.Tags
			}
			
			items = append(items, componentItem{
				name:         c,
				path:         componentPath,
				compType:     modelType,
				lastModified: modTime,
				usageCount:   usage,
				tokenCount:   tokenCount,
				tags:         tags,
				isArchived:   true,
			})
		}
	}
	
	return items
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