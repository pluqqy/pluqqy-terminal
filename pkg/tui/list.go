package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

type MainListModel struct {
	// State management
	stateManager *StateManager

	// Business logic
	businessLogic *BusinessLogic

	// Pipelines data
	pipelines []pipelineItem

	// Components data
	prompts  []componentItem
	contexts []componentItem
	rules    []componentItem

	// Preview data
	previewContent  string
	previewViewport viewport.Model

	// UI state
	pipelinesViewport  viewport.Model
	componentsViewport viewport.Model

	// Window dimensions
	width  int
	height int

	// Error handling
	err error

	// Pipeline operations
	pipelineOperator *PipelineOperator

	// Component creation
	componentCreator *ComponentCreator

	// Enhanced component editor
	enhancedEditor *EnhancedEditorState
	fileReference  *FileReferenceState

	// Exit confirmation
	exitConfirm          *ConfirmationModel
	exitConfirmationType string // "component" or "component-edit"

	// Tag editing
	tagEditor *TagEditor

	// Tag reloading
	tagReloader       *TagReloader
	tagReloadRenderer *TagReloadRenderer

	// Search engine
	searchEngine *search.Engine

	// Search state
	searchBar          *SearchBar
	searchQuery        string
	filteredPipelines  []pipelineItem
	filteredComponents []componentItem

	// Component table renderer - maintains persistent viewport scroll state across renders.
	// This ensures that when navigating through components with arrow keys, the viewport
	// automatically scrolls to keep the selected component visible, and maintains that
	// scroll position between View() calls. Without this persistence, the scroll position
	// would reset on each render, causing the viewport to jump back to the top.
	componentTableRenderer *ComponentTableRenderer

	// Mermaid diagram generation
	mermaidState    *MermaidState
	mermaidOperator *MermaidOperator

	// Rename functionality
	renameState    *RenameState
	renameRenderer *RenameRenderer
	renameOperator *RenameOperator

	// Clone functionality
	cloneState    *CloneState
	cloneRenderer *CloneRenderer
	cloneOperator *CloneOperator
	
	// Search filter helper
	searchFilterHelper *SearchFilterHelper
}

func (m *MainListModel) performSearch() {
	if m.searchQuery == "" {
		// No search query, check if we need to reload without archived items
		currentHasArchived := false
		for _, p := range m.pipelines {
			if p.isArchived {
				currentHasArchived = true
				break
			}
		}

		// If we have archived items loaded, reload without them
		if currentHasArchived {
			m.loadPipelines()
			m.loadComponents()
			m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
		}

		// Show all active items
		m.filteredPipelines = m.pipelines
		m.filteredComponents = m.businessLogic.GetAllComponents()

		// Update state manager with current counts for proper cursor navigation
		m.stateManager.UpdateCounts(len(m.filteredComponents), len(m.filteredPipelines))
		return
	}

	// Check if we need to reload data with archived items
	needsArchived := m.shouldIncludeArchived()
	currentHasArchived := false
	for _, p := range m.pipelines {
		if p.isArchived {
			currentHasArchived = true
			break
		}
	}

	// Reload data if archived status changed
	if needsArchived && !currentHasArchived {
		// Need to reload with archived items
		m.loadPipelines()
		m.loadComponents()
		m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
	} else if !needsArchived && currentHasArchived {
		// Need to reload without archived items
		m.loadPipelines()
		m.loadComponents()
		m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
	}

	// Use search engine to find matching items
	if m.searchEngine != nil {
		results, err := m.searchEngine.Search(m.searchQuery)
		if err != nil {
			// On error, show all items
			m.filteredPipelines = m.pipelines
			m.filteredComponents = m.businessLogic.GetAllComponents()

			// Update state manager with current counts for proper cursor navigation
			m.stateManager.UpdateCounts(len(m.filteredComponents), len(m.filteredPipelines))
			return
		}

		// Use the helper function to filter results
		m.filteredPipelines, m.filteredComponents = FilterSearchResults(
			results,
			m.pipelines,
			m.businessLogic.GetAllComponents(),
		)

		// Update state manager with filtered counts for proper cursor navigation
		m.stateManager.UpdateCounts(len(m.filteredComponents), len(m.filteredPipelines))

		// Reset cursors if they're out of bounds
		m.stateManager.ResetCursorsAfterSearch(len(m.filteredComponents), len(m.filteredPipelines))
	}
}

func (m *MainListModel) getAllComponents() []componentItem {
	return m.businessLogic.GetAllComponents()
}

// reloadComponents loads components and updates business logic
func (m *MainListModel) reloadComponents() {
	m.loadComponents()
	m.businessLogic.SetComponents(m.prompts, m.contexts, m.rules)
	// Re-run search if active
	if m.searchQuery != "" {
		m.performSearch()
	}
}

// reloadPipelinesWithSearch loads pipelines and refreshes search
func (m *MainListModel) reloadPipelinesWithSearch() {
	m.loadPipelines()
	// Re-run search if active
	if m.searchQuery != "" {
		m.performSearch()
	}
}

// getCurrentComponents returns either filtered components (if searching) or all components
func (m *MainListModel) getCurrentComponents() []componentItem {
	return m.filteredComponents
}

// getVisibleComponents is an alias for getCurrentComponents for clearer intent
func (m *MainListModel) getVisibleComponents() []componentItem {
	return m.filteredComponents
}

// getCurrentPipelines returns either filtered pipelines (if searching) or all pipelines
func (m *MainListModel) getCurrentPipelines() []pipelineItem {
	if m.searchQuery != "" {
		return m.filteredPipelines
	}
	return m.pipelines
}

// getFilteredPipelines returns the currently visible pipelines (respects search filter)
func (m *MainListModel) getFilteredPipelines() []pipelineItem {
	return m.filteredPipelines
}

func (m *MainListModel) getEditingItemName() string {
	return GetEditingItemName(m.tagEditor, m.stateManager, m.getCurrentComponents(), m.pipelines)
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle non-keyboard messages for component creation FIRST
	// This ensures the enhanced editor gets all necessary messages
	if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
		if m.componentCreator.IsActive() && m.componentCreator.IsEnhancedEditorActive() {
			if editor := m.componentCreator.GetEnhancedEditor(); editor != nil {
				// Handle file picker messages if file picking
				if editor.IsFilePicking() {
					cmd := editor.UpdateFilePicker(msg)
					if cmd != nil {
						return m, cmd
					}
				}
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle clone mode input first if active
		if m.cloneState.Active {
			handled, cmd := m.cloneState.HandleInput(msg)
			if handled {
				return m, cmd
			}
		}

		// Handle rename mode input if active
		if m.renameState.Active {
			handled, cmd := m.renameState.HandleInput(msg)
			if handled {
				return m, cmd
			}
		}

		// Handle search input when search pane is active
		if m.stateManager.IsInSearchPane() && !m.tagEditor.Active && !m.componentCreator.IsActive() && !m.renameState.Active && !m.cloneState.Active {
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
			case "ctrl+a":
				// Toggle archived filter
				newQuery := m.searchFilterHelper.ToggleArchivedFilter(m.searchBar.Value())
				m.searchBar.SetValue(newQuery)
				m.searchQuery = newQuery
				m.performSearch()
				return m, nil
			case "ctrl+t":
				// Cycle type filter
				newQuery := m.searchFilterHelper.CycleTypeFilter(m.searchBar.Value())
				m.searchBar.SetValue(newQuery)
				m.searchQuery = newQuery
				m.performSearch()
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

		// Handle enhanced editor mode
		if m.enhancedEditor.IsActive() {
			handled, cmd := HandleEnhancedEditorInput(m.enhancedEditor, msg, m.width)
			if handled {
				// Check if editor is still active after handling input
				if !m.enhancedEditor.IsActive() {
					// Reload components after editing
					m.reloadComponents()
					m.performSearch()
				}
				return m, cmd
			}
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

		case "shift+tab", "backtab":
			// Handle reverse tab navigation
			if m.stateManager.IsInSearchPane() {
				m.searchBar.SetActive(false)
			}
			m.stateManager.HandleTabNavigation(true)
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

		case "e":
			if m.stateManager.ActivePane == pipelinesPane {
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					if pipeline.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this pipeline before editing")
						}
					}
					// Edit the selected pipeline
					return m, func() tea.Msg {
						return SwitchViewMsg{
							view:     pipelineBuilderView,
							pipeline: pipeline.path, // Use path (filename) not name
						}
					}
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Edit component in TUI editor
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					if comp.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this component before editing")
						}
					}
					// Read the component content
					content, err := files.ReadComponent(comp.path)
					if err != nil {
						m.err = err
						return m, nil
					}

					// Use enhanced editor
					m.enhancedEditor.StartEditing(comp.path, comp.name, comp.compType, content.Content, comp.tags)

					return m, nil
				}
			}

		case "ctrl+x":
			// Edit component in external editor
			if m.stateManager.ActivePane == componentsPane {
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					if comp.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this component before editing")
						}
					}
					return m, m.pipelineOperator.OpenInEditor(comp.path, m.reloadComponents)
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
					m.tagEditor.Start(comp.path, comp.tags, "component", comp.name)
					m.tagEditor.SetSize(m.width, m.height)
				}
			} else if m.stateManager.ActivePane == pipelinesPane {
				// Use filtered pipelines if search is active
				pipelines := m.filteredPipelines
				if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					m.tagEditor.Start(pipeline.path, pipeline.tags, "pipeline", pipeline.name)
					m.tagEditor.SetSize(m.width, m.height)
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
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					if pipeline.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this pipeline before setting it as active")
						}
					}
					return m, m.pipelineOperator.SetPipeline(pipeline.path)
				}
			}

		case "y":
			if m.stateManager.ActivePane == pipelinesPane {
				// Copy selected pipeline content to clipboard
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					selectedPipeline := pipelines[m.stateManager.PipelineCursor]
					// Load and compose the pipeline (works for both archived and active)
					pipeline, err := files.ReadArchivedOrActivePipeline(selectedPipeline.path, selectedPipeline.isArchived)
					if err == nil && pipeline != nil {
						output, err := composer.ComposePipeline(pipeline)
						if err == nil {
							if err := clipboard.WriteAll(output); err == nil {
								return m, func() tea.Msg {
									return StatusMsg(selectedPipeline.name + " → clipboard")
								}
							}
						}
					}
				}
			}

		case "M":
			// Generate mermaid diagram for selected pipeline
			if m.stateManager.ActivePane == pipelinesPane && !m.mermaidState.IsGenerating() {
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					if pipeline.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this pipeline before generating a diagram")
						}
					}
					return m, m.mermaidOperator.GeneratePipelineDiagram(pipeline)
				}
			}

		case "ctrl+d":
			if m.stateManager.ActivePane == pipelinesPane {
				// Delete pipeline with confirmation
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					m.stateManager.SetDeletingFromPane(pipelinesPane)
					pipeline := pipelines[m.stateManager.PipelineCursor]
					pipelineName := pipeline.name
					pipelinePath := pipeline.path
					pipelineTags := pipeline.tags

					m.pipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return m.pipelineOperator.DeletePipeline(pipelinePath, pipelineTags, pipeline.isArchived, m.reloadPipelinesWithSearch)
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
							return m.pipelineOperator.DeleteComponent(comp, m.reloadComponents)
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
				// Archive/Unarchive pipeline with confirmation
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					m.stateManager.SetArchivingFromPane(pipelinesPane)

					if pipeline.isArchived {
						// Unarchive the pipeline
						m.pipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Unarchive pipeline '%s'?", pipeline.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.pipelineOperator.UnarchivePipeline(pipeline.path, m.reloadPipelinesWithSearch)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					} else {
						// Archive the pipeline
						m.pipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Archive pipeline '%s'?", pipeline.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.pipelineOperator.ArchivePipeline(pipeline.path, m.reloadPipelinesWithSearch)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					}
				}
			} else if m.stateManager.ActivePane == componentsPane {
				// Archive/Unarchive component with confirmation
				components := m.getCurrentComponents()
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					m.stateManager.SetArchivingFromPane(componentsPane)

					if comp.isArchived {
						// Unarchive the component
						m.pipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Unarchive %s '%s'?", comp.compType, comp.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.pipelineOperator.UnarchiveComponent(comp, m.reloadComponents)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					} else {
						// Archive the component
						m.pipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.pipelineOperator.ArchiveComponent(comp, m.reloadComponents)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					}
				}
			}

		case "C": // Uppercase C for clone/duplicate
			// Start clone mode
			if m.stateManager.ActivePane == componentsPane {
				components := m.getCurrentComponents()
				if len(components) > 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					// Prepare component for cloning
					displayName, path, isArchived := m.cloneOperator.PrepareCloneComponent(comp)
					m.cloneState.Start(displayName, "component", path, isArchived)
				}
			} else if m.stateManager.ActivePane == pipelinesPane {
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					// Prepare pipeline for cloning
					displayName, path, isArchived := m.cloneOperator.PrepareClonePipeline(pipeline)
					m.cloneState.Start(displayName, "pipeline", path, isArchived)
				}
			}

		case "R": // Uppercase R for rename (destructive operation)
			// Handle rename mode input first if active
			if m.renameState.Active {
				handled, cmd := m.renameState.HandleInput(msg)
				if handled {
					return m, cmd
				}
			} else {
				// Start rename mode
				if m.stateManager.ActivePane == componentsPane {
					components := m.getCurrentComponents()
					if len(components) > 0 && m.stateManager.ComponentCursor < len(components) {
						comp := components[m.stateManager.ComponentCursor]
						// Prepare component for renaming
						displayName, path, isArchived := m.renameOperator.PrepareRenameComponent(comp)
						m.renameState.Start(displayName, "component", path, isArchived)
					}
				} else if m.stateManager.ActivePane == pipelinesPane {
					pipelines := m.getCurrentPipelines()
					if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
						pipeline := pipelines[m.stateManager.PipelineCursor]
						// Prepare pipeline for renaming
						displayName, path, isArchived := m.renameOperator.PrepareRenamePipeline(pipeline)
						m.renameState.Start(displayName, "pipeline", path, isArchived)
					}
				}
			}
		}

	case RenameSuccessMsg:
		// Handle successful rename
		m.renameState.Reset()
		// Reload data
		m.reloadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.searchQuery != "" {
			m.performSearch()
		}
		// Show success message (could be shown in status bar if available)
		return m, nil

	case RenameErrorMsg:
		// Handle rename error
		m.renameState.ValidationError = msg.Error.Error()
		return m, nil

	case CloneSuccessMsg:
		// Handle successful clone
		m.cloneState.Reset()
		// Reload data
		m.reloadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.searchQuery != "" {
			m.performSearch()
		}
		// Show success message
		statusText := fmt.Sprintf("✓ Cloned %s '%s' to '%s'", msg.ItemType, msg.OriginalName, msg.NewName)
		if msg.ClonedToArchive {
			statusText += " (in archive)"
		}
		return m, func() tea.Msg {
			return StatusMsg(statusText)
		}

	case CloneErrorMsg:
		// Handle clone error
		m.cloneState.ValidationError = msg.Error.Error()
		// Show error message
		return m, func() tea.Msg {
			return StatusMsg(fmt.Sprintf("✗ Clone failed: %v", msg.Error))
		}

	case StatusMsg:
		// Handle status messages, especially save confirmations from enhanced editor
		msgStr := string(msg)
		if strings.HasPrefix(msgStr, "✓ Saved:") {
			// A component was saved successfully
			if m.enhancedEditor.IsActive() {
				// Reload components but keep editor open
				m.reloadComponents()
				m.performSearch()
			}
		}
		// Pass the message up to the parent app for display
		return m, func() tea.Msg { return msg }

	case ReloadMsg:
		// Reload data after tag editing
		m.reloadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.searchQuery != "" {
			m.performSearch()
		}
		// Also reload available tags for tag editor
		if m.tagEditor != nil {
			m.tagEditor.LoadAvailableTags()
		}
		return m, nil

	case TagReloadMsg:
		// First check if tag editor is active and should handle the message
		if m.tagEditor != nil && m.tagEditor.Active {
			handled, cmd := m.tagEditor.HandleMessage(msg)
			if handled {
				// Reload components and pipelines to reflect new tags
				m.reloadComponents()
				m.loadPipelines()
				// Re-run search if active
				if m.searchQuery != "" {
					m.performSearch()
				}
				return m, cmd
			}
		}
		return m, nil

	case tagReloadCompleteMsg:
		// Check if tag editor should handle this
		if m.tagEditor != nil && m.tagEditor.Active {
			handled, _ := m.tagEditor.HandleMessage(msg)
			if handled {
				return m, nil
			}
		}
		return m, nil

	case tagDeletionCompleteMsg:
		// Check if tag editor should handle this
		if m.tagEditor != nil && m.tagEditor.Active {
			handled, cmd := m.tagEditor.HandleMessage(msg)
			if handled {
				// Reload components and pipelines to reflect removed tags
				m.reloadComponents()
				m.loadPipelines()
				// Re-run search if active
				if m.searchQuery != "" {
					m.performSearch()
				}
				return m, cmd
			}
		}
		return m, nil

	case tagDeletionProgressMsg:
		// Check if tag editor should handle this
		if m.tagEditor != nil && m.tagEditor.Active {
			handled, cmd := m.tagEditor.HandleMessage(msg)
			if handled {
				return m, cmd
			}
		}
		return m, nil
	}

	// Handle enhanced editor for non-KeyMsg message types
	// This is crucial for filepicker which needs to process internal messages like directory reads
	if m.enhancedEditor.IsActive() {
		// Only handle non-KeyMsg types here (KeyMsg types are handled in the switch above)
		if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
			if m.enhancedEditor.IsFilePicking() {
				// Filepicker needs to process internal messages for directory reading
				cmd := m.enhancedEditor.UpdateFilePicker(msg)
				if cmd != nil {
					return m, cmd
				}
			} else {
				// Normal editing mode may also need non-key message handling
				cmd := m.enhancedEditor.UpdateTextarea(msg)
				if cmd != nil {
					return m, cmd
				}
			}
		}
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
		// Don't handle component creator's textarea here - let the component creator handle everything
		// This matches how the Builder view works where it doesn't directly update the textarea
		
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
		// Add padding to match other views
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		return contentStyle.Render(m.exitConfirm.View())
	}

	// If creating component, show creation wizard
	if m.componentCreator.IsActive() {
		return m.componentCreationView()
	}

	// If editing component with enhanced editor, show enhanced edit view
	if m.enhancedEditor.IsActive() {
		// Handle exit confirmation dialog
		if m.enhancedEditor.ExitConfirmActive {
			// Add padding to match other views
			contentStyle := lipgloss.NewStyle().
				PaddingLeft(1).
				PaddingRight(1)
			return contentStyle.Render(m.enhancedEditor.ExitConfirm.View())
		}

		renderer := NewEnhancedEditorRenderer(m.width, m.height)
		return renderer.Render(m.enhancedEditor)
	}

	// If editing tags, show tag edit view
	if m.tagEditor.Active {
		// Use the new unified tag editor renderer
		renderer := NewTagEditorRenderer(m.tagEditor, m.width, m.height)
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
	// Use the persistent table renderer for proper scroll state
	if m.componentTableRenderer != nil {
		componentRenderer.TableRenderer = m.componentTableRenderer
		// Update the table renderer with current state
		m.componentTableRenderer.SetComponents(m.filteredComponents)
		m.componentTableRenderer.SetCursor(m.stateManager.ComponentCursor)
		m.componentTableRenderer.SetActive(m.stateManager.ActivePane == componentsPane)
	}

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
	previewPane := mainRenderer.RenderPreviewPane(m.getCurrentPipelines(), m.filteredComponents, m.stateManager.PipelineCursor, m.stateManager.ComponentCursor)
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

	finalView := s.String()

	// Overlay clone dialog if active
	if m.cloneState != nil && m.cloneState.Active && m.cloneRenderer != nil {
		m.cloneRenderer.SetSize(m.width, m.height)
		finalView = m.cloneRenderer.RenderOverlay(finalView, m.cloneState)
	}

	// Overlay rename dialog if active
	if m.renameState != nil && m.renameState.Active && m.renameRenderer != nil {
		finalView = m.renameRenderer.RenderOverlay(finalView, m.renameState)
	}

	// Overlay tag reload status if active
	if m.tagReloader != nil && m.tagReloader.IsActive() && m.tagReloadRenderer != nil {
		overlay := m.tagReloadRenderer.RenderStatus(m.tagReloader)
		if overlay != "" {
			// Combine the views by overlaying
			return overlayViews(finalView, overlay)
		}
	}

	return finalView
}

func (m *MainListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update search bar width
	m.searchBar.SetWidth(width)
	// Update tag editor size
	if m.tagEditor != nil {
		m.tagEditor.SetSize(width, height)
	}
	// Update tag reload renderer size
	if m.tagReloadRenderer != nil {
		m.tagReloadRenderer.SetSize(width, height)
	}
	// Update rename renderer size
	if m.renameRenderer != nil {
		m.renameRenderer.SetSize(width, height)
	}
	m.updateViewportSizes()
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
		pipelines := m.getCurrentPipelines()
		if len(pipelines) == 0 {
			m.previewContent = renderer.RenderEmptyPreview(pipelinesPane, false, false)
			return
		}

		if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(pipelines) {
			pipeline := pipelines[m.stateManager.PipelineCursor]
			m.previewContent = renderer.RenderPipelinePreview(pipeline.path, pipeline.isArchived)
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
		// Check if enhanced editor is active for component creation
		if m.componentCreator.IsEnhancedEditorActive() {
			handled, cmd := m.componentCreator.HandleEnhancedEditorInput(msg, m.width)
			if handled {
				// Check if component was saved (but editor stays open)
				if m.componentCreator.WasSaveSuccessful() {
					// Component was saved, reload components
					m.reloadComponents()
					return m, tea.Batch(
						cmd,
						func() tea.Msg { return StatusMsg(m.componentCreator.GetStatusMessage()) },
					)
				}
				// Check if component creation was cancelled
				if !m.componentCreator.IsActive() {
					// Component creation ended (saved or cancelled)
					// Always reload to ensure list is current
					m.reloadComponents()
					return m, cmd
				}
				return m, cmd
			}
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
		// Use enhanced editor if available
		if m.componentCreator.IsEnhancedEditorActive() {
			return renderer.RenderWithEnhancedEditor(
				m.componentCreator.GetEnhancedEditor(),
				m.componentCreator.GetComponentType(),
				m.componentCreator.GetComponentName(),
			)
		}
		// Fallback to simple editor
		return renderer.RenderContentEdit(m.componentCreator.GetComponentType(), m.componentCreator.GetComponentName(), m.componentCreator.GetComponentContent())
	}

	return "Unknown creation step"
}
