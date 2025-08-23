package tui

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

// NewMainListModel creates and initializes a new MainListModel
func NewMainListModel() *MainListModel {
	// Initialize mermaid state
	mermaidState := NewMermaidState()

	m := &MainListModel{
		stateManager: NewStateManager(),
		
		// Initialize composed structures
		data: &ListDataStore{},
		
		viewports: &ListViewportManager{
			Preview:    viewport.New(80, 20), // Default size
			Pipelines:  viewport.New(40, 20), // Default size
			Components: viewport.New(40, 20), // Default size
		},
		
		editors: &ListEditorComponents{
			Enhanced:      NewEnhancedEditorState(),
			FileReference: NewFileReferenceState(),
			TagEditor:     NewTagEditor(),
			Rename: &ListRenameComponents{
				State:    NewRenameState(),
				Renderer: NewRenameRenderer(),
				Operator: NewRenameOperator(),
			},
			Clone: &ListCloneComponents{
				State:    NewCloneState(),
				Renderer: NewCloneRenderer(),
				Operator: NewCloneOperator(),
			},
		},
		
		search: &ListSearchComponents{
			Bar:          NewSearchBar(),
			FilterHelper: NewSearchFilterHelper(),
		},
		
		operations: &ListOperationComponents{
			BusinessLogic:    NewBusinessLogic(),
			PipelineOperator: NewPipelineOperator(),
			ComponentCreator: NewComponentCreator(),
			MermaidOperator:  NewMermaidOperator(mermaidState),
			TagReloader:      NewTagReloader(),
		},
		
		ui: &ListUIComponents{
			ExitConfirm:            NewConfirmation(),
			ComponentTableRenderer: NewComponentTableRenderer(40, 20, true), // Default size, true for showUsageColumn
			TagReloadRenderer:      NewTagReloadRenderer(80, 20),            // Default size
			MermaidState:           mermaidState,
		},
	}
	// Set initial preview state
	m.stateManager.ShowPreview = false // Start with preview hidden, user can toggle with 'p'
	m.loadPipelines()
	m.loadComponents()
	m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
	m.initializeSearchEngine()
	// Initialize filtered lists with all items
	m.data.FilteredPipelines = m.data.Pipelines
	m.data.FilteredComponents = m.getAllComponents()

	// Update state manager with counts after both are loaded
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.data.Pipelines))

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

	m.data.Pipelines = nil

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

		m.data.Pipelines = append(m.data.Pipelines, pipelineItem{
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

			m.data.Pipelines = append(m.data.Pipelines, pipelineItem{
				name:       pipeline.Name, // Use the actual pipeline name from YAML
				path:       pipelineFile,
				tags:       pipeline.Tags,
				tokenCount: tokenCount,
				isArchived: true,
			})
		}
	}

	// Rebuild search index when pipelines are reloaded
	if m.search.Engine != nil {
		if includeArchived {
			m.search.Engine.BuildIndexWithOptions(true)
		} else {
			m.search.Engine.BuildIndex()
		}
	}

	// Update filtered list if no active search
	if m.search.Query == "" {
		m.data.FilteredPipelines = m.data.Pipelines
	}

	// Update state manager counts if components are already loaded
	if m.data.Prompts != nil || m.data.Contexts != nil || m.data.Rules != nil {
		m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.data.Pipelines))
	}
}

// shouldIncludeArchived checks if the current search query requires archived items
func (m *MainListModel) shouldIncludeArchived() bool {
	if m.search.Query == "" {
		return false
	}

	// Parse the search query to check for status:archived
	parser := search.NewParser()
	query, err := parser.Parse(m.search.Query)
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
	m.data.Prompts = nil
	m.data.Contexts = nil
	m.data.Rules = nil

	// Load prompts
	m.data.Prompts = m.loadComponentsOfType("prompts", files.PromptsDir, models.ComponentTypePrompt, usageMap, includeArchived)

	// Load contexts
	m.data.Contexts = m.loadComponentsOfType("contexts", files.ContextsDir, models.ComponentTypeContext, usageMap, includeArchived)

	// Load rules
	m.data.Rules = m.loadComponentsOfType("rules", files.RulesDir, models.ComponentTypeRules, usageMap, includeArchived)

	// Update business logic with new components
	m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)

	// Rebuild search index when components are reloaded
	if m.search.Engine != nil {
		if includeArchived {
			m.search.Engine.BuildIndexWithOptions(true)
		} else {
			m.search.Engine.BuildIndex()
		}
	}

	// Update filtered list if no active search
	if m.search.Query == "" {
		m.data.FilteredComponents = m.getAllComponents()
	}

	// Update state manager counts
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.data.Pipelines))
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

		// Read component content for token estimation and display name
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		displayName := c // Default to filename
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
			// Use display name from component (from frontmatter or filename)
			if component.Name != "" {
				displayName = component.Name
			}
		}

		tags := []string{}
		if component != nil {
			tags = component.Tags
		}

		items = append(items, componentItem{
			name:         displayName,
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

			// Get token count and display name
			tokenCount := 0
			displayName := c // Default to filename
			if component != nil {
				tokenCount = utils.EstimateTokens(component.Content)
				// Use display name from component (from frontmatter or filename)
				if component.Name != "" {
					displayName = component.Name
				}
			}

			tags := []string{}
			if component != nil {
				tags = component.Tags
			}

			items = append(items, componentItem{
				name:         displayName,
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
	m.search.Engine = searchManager.GetEngine()
}

// updateViewportSizes updates all viewport dimensions based on window size
func (m *MainListModel) updateViewportSizes() {
	// Calculate dimensions
	columnWidth := (m.viewports.Width - 6) / 2                 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3                             // Height for search bar
	contentHeight := m.viewports.Height - 13 - searchBarHeight // Reserve space for header, search bar, help pane, and spacing

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

	m.viewports.Pipelines.Width = columnWidth - 4 // Account for borders (2) and padding (2)
	m.viewports.Pipelines.Height = viewportHeight
	m.viewports.Components.Width = columnWidth - 4 // Account for borders (2) and padding (2)
	m.viewports.Components.Height = viewportHeight

	// Update preview viewport
	if m.stateManager.ShowPreview {
		previewHeight := m.viewports.Height/2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		m.viewports.Preview.Width = m.viewports.Width - 8 // Account for outer padding (2), borders (2), and inner padding (2) + extra spacing
		m.viewports.Preview.Height = previewHeight
	}

	// Update component table renderer dimensions
	if m.ui.ComponentTableRenderer != nil {
		m.ui.ComponentTableRenderer.SetSize(columnWidth, contentHeight)
	}
}
