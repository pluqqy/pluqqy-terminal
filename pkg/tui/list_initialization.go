package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/search/unified"
	"github.com/pluqqy/pluqqy-terminal/pkg/tui/shared"
	"github.com/pluqqy/pluqqy-terminal/pkg/utils"
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
			Enhanced:       NewEnhancedEditorState(),
			FileReference:  NewFileReferenceState(),
			TagEditor:      NewTagEditor(),
			ComponentUsage: NewComponentUsageState(),
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
			UnifiedManager: unified.NewUnifiedSearchManager(),
			Bar:            NewSearchBar(),
			FilterHelper:   NewSearchFilterHelper(),
		},
		
		operations: &ListOperationComponents{
			BusinessLogic:    NewBusinessLogic(),
			PipelineOperator: NewPipelineOperator(),
			ComponentCreator: nil, // Will be set after model is created
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

	// Create and setup the component creator after the model is initialized
	m.operations.ComponentCreator = createListComponentCreator(m)
	m.setupComponentCreatorEditor()

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

	// Note: Search index rebuilding is now handled by the unified search manager
	// No explicit index rebuilding needed

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
	return unified.ShouldIncludeArchived(m.search.Query)
}

// loadComponents loads all component files and their metadata
func (m *MainListModel) loadComponents() {
	// Check if we should include archived items based on search query
	includeArchived := m.shouldIncludeArchived()

	// Use shared ComponentLoader
	loader := shared.NewComponentLoader("")
	prompts, contexts, rules, _ := loader.LoadComponents(includeArchived)

	// Clear existing components
	m.data.Prompts = nil
	m.data.Contexts = nil
	m.data.Rules = nil

	// Convert shared ComponentItems to local componentItems
	m.data.Prompts = convertToComponentItems(unified.ConvertSharedComponentItemsToUnified(prompts))
	m.data.Contexts = convertToComponentItems(unified.ConvertSharedComponentItemsToUnified(contexts))
	m.data.Rules = convertToComponentItems(unified.ConvertSharedComponentItemsToUnified(rules))

	// Update business logic with new components
	m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)

	// Note: Search index rebuilding is now handled by the unified search manager
	// No explicit index rebuilding needed

	// Update filtered list if no active search
	if m.search.Query == "" {
		m.data.FilteredComponents = m.getAllComponents()
	}

	// Update state manager counts
	m.stateManager.UpdateCounts(len(m.getAllComponents()), len(m.data.Pipelines))
}

// convertToComponentItems converts unified ComponentItems to local componentItems
func convertToComponentItems(items []unified.ComponentItem) []componentItem {
	result := make([]componentItem, len(items))
	for i, item := range items {
		result[i] = componentItem{
			name:         item.Name,
			path:         item.Path,
			compType:     item.CompType,
			lastModified: item.LastModified,
			usageCount:   item.UsageCount,
			tokenCount:   item.TokenCount,
			tags:         item.Tags,
			isArchived:   item.IsArchived,
		}
	}
	return result
}


// initializeSearchEngine sets up the search engine (legacy compatibility)
func (m *MainListModel) initializeSearchEngine() {
	// Search engine is now initialized via the unified search manager
	// This method is kept for backward compatibility but does nothing
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

// createListComponentCreator creates a ListComponentCreator for the list view
func createListComponentCreator(model *MainListModel) *ListComponentCreator {
	// Create the reload callback that's specific to the list view
	reloadCallback := func() {
		if model != nil {
			model.reloadComponents()
			model.performSearch() // Re-run search if active
		}
	}

	// Create the list-specific component creator with enhanced editor
	// Note: Enhanced editor will be nil initially, will be set up later
	creator := NewListComponentCreator(reloadCallback, nil)
	
	return creator
}

// setupComponentCreatorEditor sets up the enhanced editor for component creation
// This must be called after the MainListModel is fully initialized
func (m *MainListModel) setupComponentCreatorEditor() {
	if m.operations.ComponentCreator != nil && m.editors.Enhanced != nil {
		// Update the enhanced editor reference in the list component creator
		m.operations.ComponentCreator.enhancedEditor = m.editors.Enhanced
		
		// Set up the enhanced editor adapter
		adapter := shared.NewEnhancedEditorAdapter(m.editors.Enhanced)
		m.operations.ComponentCreator.SetEnhancedEditor(adapter)
	}
}
