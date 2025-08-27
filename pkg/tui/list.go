package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

type MainListModel struct {
	// State management
	stateManager *StateManager

	// Composed data structures
	data       *ListDataStore
	viewports  *ListViewportManager
	editors    *ListEditorComponents
	search     *ListSearchComponents
	operations *ListOperationComponents
	ui         *ListUIComponents

	// Error handling
	err error
}

func (m *MainListModel) performSearch() {
	// Initialize unified manager if needed
	m.search.InitializeUnifiedManager()
	
	if m.search.Query == "" {
		// No search query, check if we need to reload without archived items
		currentHasArchived := false
		for _, p := range m.data.Pipelines {
			if p.isArchived {
				currentHasArchived = true
				break
			}
		}

		// If we have archived items loaded, reload without them
		if currentHasArchived {
			m.loadPipelines()
			m.loadComponents()
			m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
		}

		// Show all active items
		m.data.FilteredPipelines = m.data.Pipelines
		m.data.FilteredComponents = m.operations.BusinessLogic.GetAllComponents()

		// Update state manager with current counts for proper cursor navigation
		m.stateManager.UpdateCounts(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))
		return
	}

	// Check if we need to reload data with archived items
	needsArchived := m.shouldIncludeArchived()
	currentHasArchived := false
	for _, p := range m.data.Pipelines {
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
		m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
	} else if !needsArchived && currentHasArchived {
		// Need to reload without archived items
		m.loadPipelines()
		m.loadComponents()
		m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
	}

	// Try the new unified search first
	if m.search.UnifiedManager != nil {
		// Configure unified manager
		m.search.UnifiedManager.SetIncludeArchived(needsArchived)
		
		// Use the new unified filter function
		m.data.FilteredPipelines, m.data.FilteredComponents = FilterSearchResultsUnified(
			m.search.Query,
			m.data.Pipelines,
			m.operations.BusinessLogic.GetAllComponents(),
		)

		// Update state manager with filtered counts for proper cursor navigation
		m.stateManager.UpdateCounts(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))

		// Reset cursors if they're out of bounds
		m.stateManager.ResetCursorsAfterSearch(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))
		return
	}

	// Fallback to legacy search engine
	if m.search.Engine != nil {
		results, err := m.search.Engine.Search(m.search.Query)
		if err != nil {
			// On error, show all items
			m.data.FilteredPipelines = m.data.Pipelines
			m.data.FilteredComponents = m.operations.BusinessLogic.GetAllComponents()

			// Update state manager with current counts for proper cursor navigation
			m.stateManager.UpdateCounts(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))
			return
		}

		// Use the helper function to filter results
		m.data.FilteredPipelines, m.data.FilteredComponents = FilterSearchResults(
			results,
			m.data.Pipelines,
			m.operations.BusinessLogic.GetAllComponents(),
		)

		// Update state manager with filtered counts for proper cursor navigation
		m.stateManager.UpdateCounts(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))

		// Reset cursors if they're out of bounds
		m.stateManager.ResetCursorsAfterSearch(len(m.data.FilteredComponents), len(m.data.FilteredPipelines))
	}
}

func (m *MainListModel) getAllComponents() []componentItem {
	return m.operations.BusinessLogic.GetAllComponents()
}

// reloadComponents loads components and updates business logic
func (m *MainListModel) reloadComponents() {
	m.loadComponents()
	m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
	// Re-run search if active
	if m.search.Query != "" {
		m.performSearch()
	}
}

// reloadPipelinesWithSearch loads pipelines and refreshes search
func (m *MainListModel) reloadPipelinesWithSearch() {
	m.loadPipelines()
	// Re-run search if active
	if m.search.Query != "" {
		m.performSearch()
	}
}

// getCurrentComponents returns either filtered components (if searching) or all components
func (m *MainListModel) getCurrentComponents() []componentItem {
	return m.data.FilteredComponents
}

// getVisibleComponents is an alias for getCurrentComponents for clearer intent
func (m *MainListModel) getVisibleComponents() []componentItem {
	return m.data.FilteredComponents
}

// getCurrentPipelines returns either filtered pipelines (if searching) or all pipelines
func (m *MainListModel) getCurrentPipelines() []pipelineItem {
	if m.search.Query != "" {
		return m.data.FilteredPipelines
	}
	return m.data.Pipelines
}

// getFilteredPipelines returns the currently visible pipelines (respects search filter)
func (m *MainListModel) getFilteredPipelines() []pipelineItem {
	return m.data.FilteredPipelines
}

func (m *MainListModel) getEditingItemName() string {
	return GetEditingItemName(m.editors.TagEditor, m.stateManager, m.getCurrentComponents(), m.data.Pipelines)
}

func (m *MainListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle non-keyboard messages for component creation FIRST
	// This ensures the enhanced editor gets all necessary messages
	if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
		if m.operations.ComponentCreator.IsActive() && m.operations.ComponentCreator.IsEnhancedEditorActive() {
			if editor := m.operations.ComponentCreator.GetEnhancedEditor(); editor != nil {
				// Handle file picker messages if file picking
				if editor.IsFilePicking() {
					cmd := editor.UpdateFilePicker(msg)
					if cmd != nil {
						if teaCmd, ok := cmd.(tea.Cmd); ok {
							return m, teaCmd
						}
					}
				}
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle clone mode input first if active
		if m.editors.Clone.State.Active {
			handled, cmd := m.editors.Clone.State.HandleInput(msg)
			if handled {
				return m, cmd
			}
		}

		// Handle rename mode input if active
		if m.editors.Rename.State.Active {
			handled, cmd := m.editors.Rename.State.HandleInput(msg)
			if handled {
				return m, cmd
			}
		}

		// Handle search input when search pane is active
		if m.stateManager.IsInSearchPane() && !m.editors.TagEditor.Active && !m.operations.ComponentCreator.IsActive() && !m.editors.Rename.State.Active && !m.editors.Clone.State.Active {
			var cmd tea.Cmd
			m.search.Bar, cmd = m.search.Bar.Update(msg)

			// Check if search query changed
			if m.search.Query != m.search.Bar.Value() {
				m.search.Query = m.search.Bar.Value()
				m.performSearch()
			}

			// Handle special keys for search
			switch msg.String() {
			case "esc":
				// Clear search and return to components pane
				m.search.Bar.SetValue("")
				m.search.Query = ""
				m.performSearch()
				m.stateManager.ExitSearch()
				m.search.Bar.SetActive(false)
				return m, nil
			case "ctrl+a":
				// Toggle archived filter
				newQuery := m.search.FilterHelper.ToggleArchivedFilter(m.search.Bar.Value())
				m.search.Bar.SetValue(newQuery)
				m.search.Query = newQuery
				m.performSearch()
				return m, nil
			case "ctrl+t":
				// Cycle type filter
				newQuery := m.search.FilterHelper.CycleTypeFilter(m.search.Bar.Value())
				m.search.Bar.SetValue(newQuery)
				m.search.Query = newQuery
				m.performSearch()
				return m, nil
			case "tab":
				// Let tab handling below take care of navigation
			default:
				return m, cmd
			}
		}

		// Handle exit confirmation
		if m.ui.ExitConfirm.Active() {
			return m, m.ui.ExitConfirm.Update(msg)
		}

		// Handle component creation mode
		if m.operations.ComponentCreator.IsActive() {
			return m.handleComponentCreation(msg)
		}

		// Handle enhanced editor mode
		if m.editors.Enhanced.IsActive() {
			handled, cmd := HandleEnhancedEditorInput(m.editors.Enhanced, msg, m.viewports.Width)
			if handled {
				// Check if editor is still active after handling input
				if !m.editors.Enhanced.IsActive() {
					// Reload components after editing
					m.reloadComponents()
					m.performSearch()
				}
				return m, cmd
			}
		}

		// Handle tag editing mode
		if m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleInput(msg)
			if handled {
				// Check if we need to reload data after saving
				if cmd != nil {
					return m, cmd
				}
				return m, nil
			}
		}

		// Handle delete confirmation
		if m.operations.PipelineOperator.IsDeleteConfirmActive() {
			return m, m.operations.PipelineOperator.UpdateDeleteConfirm(msg)
		}

		// Handle archive confirmation
		if m.operations.PipelineOperator.IsArchiveConfirmActive() {
			return m, m.operations.PipelineOperator.UpdateArchiveConfirm(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "tab":
			// Handle tab navigation
			if m.stateManager.IsInSearchPane() {
				m.search.Bar.SetActive(false)
			}
			m.stateManager.HandleTabNavigation(false)
			// Update preview when switching to non-preview pane
			if m.stateManager.ActivePane != previewPane && m.stateManager.ActivePane != searchPane {
				m.updatePreview()
			}

		case "shift+tab", "backtab":
			// Handle reverse tab navigation
			if m.stateManager.IsInSearchPane() {
				m.search.Bar.SetActive(false)
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
				m.viewports.Preview.LineUp(1)
			}

		case "down", "j":
			handled, updatePreview := m.stateManager.HandleKeyNavigation(msg.String())
			if handled {
				if updatePreview {
					m.updatePreview()
				}
			} else if m.stateManager.IsInPreviewPane() {
				// Scroll preview down
				m.viewports.Preview.LineDown(1)
			}

		case "pgup":
			if m.stateManager.IsInPreviewPane() {
				m.viewports.Preview.ViewUp()
			}

		case "pgdown":
			if m.stateManager.IsInPreviewPane() {
				m.viewports.Preview.ViewDown()
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
			m.search.Bar.SetActive(true)
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
					m.editors.Enhanced.StartEditing(comp.path, comp.name, comp.compType, content.Content, comp.tags)

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
					return m, m.operations.PipelineOperator.OpenInEditor(comp.path, m.reloadComponents)
				}
			}
			// Explicitly do nothing for pipelines pane - editing YAML directly is not encouraged

		case "t":
			// Edit tags
			if m.stateManager.ActivePane == componentsPane {
				// Use filtered components if search is active
				components := m.data.FilteredComponents
				if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
					comp := components[m.stateManager.ComponentCursor]
					m.editors.TagEditor.Start(comp.path, comp.tags, "component", comp.name)
					m.editors.TagEditor.SetSize(m.viewports.Width, m.viewports.Height)
				}
			} else if m.stateManager.ActivePane == pipelinesPane {
				// Use filtered pipelines if search is active
				pipelines := m.data.FilteredPipelines
				if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					m.editors.TagEditor.Start(pipeline.path, pipeline.tags, "pipeline", pipeline.name)
					m.editors.TagEditor.SetSize(m.viewports.Width, m.viewports.Height)
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
				m.operations.ComponentCreator.Start()
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
					return m, m.operations.PipelineOperator.SetPipeline(pipeline.path)
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
			if m.stateManager.ActivePane == pipelinesPane && !m.ui.MermaidState.IsGenerating() {
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					if pipeline.isArchived {
						return m, func() tea.Msg {
							return StatusMsg("You must unarchive this pipeline before generating a diagram")
						}
					}
					return m, m.operations.MermaidOperator.GeneratePipelineDiagram(pipeline)
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

					m.operations.PipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return m.operations.PipelineOperator.DeletePipeline(pipelinePath, pipelineTags, pipeline.isArchived, m.reloadPipelinesWithSearch)
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

					m.operations.PipelineOperator.ShowDeleteConfirmation(
						fmt.Sprintf("Delete %s '%s'?", comp.compType, comp.name),
						func() tea.Cmd {
							m.stateManager.ClearDeletionState()
							return m.operations.PipelineOperator.DeleteComponent(comp, m.reloadComponents)
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
						m.operations.PipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Unarchive pipeline '%s'?", pipeline.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.operations.PipelineOperator.UnarchivePipeline(pipeline.path, m.reloadPipelinesWithSearch)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					} else {
						// Archive the pipeline
						m.operations.PipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Archive pipeline '%s'?", pipeline.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.operations.PipelineOperator.ArchivePipeline(pipeline.path, m.reloadPipelinesWithSearch)
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
						m.operations.PipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Unarchive %s '%s'?", comp.compType, comp.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.operations.PipelineOperator.UnarchiveComponent(comp, m.reloadComponents)
							},
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return nil
							},
						)
					} else {
						// Archive the component
						m.operations.PipelineOperator.ShowArchiveConfirmation(
							fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
							func() tea.Cmd {
								m.stateManager.ClearArchiveState()
								return m.operations.PipelineOperator.ArchiveComponent(comp, m.reloadComponents)
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
					displayName, path, isArchived := m.editors.Clone.Operator.PrepareCloneComponent(comp)
					m.editors.Clone.State.Start(displayName, "component", path, isArchived)
				}
			} else if m.stateManager.ActivePane == pipelinesPane {
				pipelines := m.getCurrentPipelines()
				if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
					pipeline := pipelines[m.stateManager.PipelineCursor]
					// Prepare pipeline for cloning
					displayName, path, isArchived := m.editors.Clone.Operator.PrepareClonePipeline(pipeline)
					m.editors.Clone.State.Start(displayName, "pipeline", path, isArchived)
				}
			}

		case "R": // Uppercase R for rename (destructive operation)
			// Handle rename mode input first if active
			if m.editors.Rename.State.Active {
				handled, cmd := m.editors.Rename.State.HandleInput(msg)
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
						displayName, path, isArchived := m.editors.Rename.Operator.PrepareRenameComponent(comp)
						m.editors.Rename.State.Start(displayName, "component", path, isArchived)
					}
				} else if m.stateManager.ActivePane == pipelinesPane {
					pipelines := m.getCurrentPipelines()
					if len(pipelines) > 0 && m.stateManager.PipelineCursor < len(pipelines) {
						pipeline := pipelines[m.stateManager.PipelineCursor]
						// Prepare pipeline for renaming
						displayName, path, isArchived := m.editors.Rename.Operator.PrepareRenamePipeline(pipeline)
						m.editors.Rename.State.Start(displayName, "pipeline", path, isArchived)
					}
				}
			}
		}

	case RenameSuccessMsg:
		// Handle successful rename
		m.editors.Rename.State.Reset()
		// Reload data
		m.reloadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.search.Query != "" {
			m.performSearch()
		}
		// Show success message (could be shown in status bar if available)
		return m, nil

	case RenameErrorMsg:
		// Handle rename error
		m.editors.Rename.State.ValidationError = msg.Error.Error()
		return m, nil

	case CloneSuccessMsg:
		// Handle successful clone
		m.editors.Clone.State.Reset()
		// Reload data
		m.reloadComponents()
		m.loadPipelines()
		// Re-run search if active
		if m.search.Query != "" {
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
		m.editors.Clone.State.ValidationError = msg.Error.Error()
		// Show error message
		return m, func() tea.Msg {
			return StatusMsg(fmt.Sprintf("✗ Clone failed: %v", msg.Error))
		}

	case StatusMsg:
		// Handle status messages, especially save confirmations from enhanced editor
		msgStr := string(msg)
		if strings.HasPrefix(msgStr, "✓ Saved:") {
			// A component was saved successfully
			if m.editors.Enhanced.IsActive() {
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
		if m.search.Query != "" {
			m.performSearch()
		}
		// Also reload available tags for tag editor
		if m.editors.TagEditor != nil {
			m.editors.TagEditor.LoadAvailableTags()
		}
		return m, nil

	case TagReloadMsg:
		// First check if tag editor is active and should handle the message
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				// Reload components and pipelines to reflect new tags
				m.reloadComponents()
				m.loadPipelines()
				// Re-run search if active
				if m.search.Query != "" {
					m.performSearch()
				}
				return m, cmd
			}
		}
		return m, nil

	case tagReloadCompleteMsg:
		// Check if tag editor should handle this
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, _ := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				return m, nil
			}
		}
		return m, nil

	case tagDeletionCompleteMsg:
		// Check if tag editor should handle this
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				// Reload components and pipelines to reflect removed tags
				m.reloadComponents()
				m.loadPipelines()
				// Re-run search if active
				if m.search.Query != "" {
					m.performSearch()
				}
				return m, cmd
			}
		}
		return m, nil

	case tagDeletionProgressMsg:
		// Check if tag editor should handle this
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				return m, cmd
			}
		}
		return m, nil
	}

	// Handle enhanced editor for non-KeyMsg message types
	// This is crucial for filepicker which needs to process internal messages like directory reads
	if m.editors.Enhanced.IsActive() {
		// Only handle non-KeyMsg types here (KeyMsg types are handled in the switch above)
		if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
			if m.editors.Enhanced.IsFilePicking() {
				// Filepicker needs to process internal messages for directory reading
				cmd := m.editors.Enhanced.UpdateFilePicker(msg)
				if cmd != nil {
					return m, cmd
				}
			} else {
				// Normal editing mode may also need non-key message handling
				cmd := m.editors.Enhanced.UpdateTextarea(msg)
				if cmd != nil {
					return m, cmd
				}
			}
		}
	}

	// Update preview if needed
	if m.stateManager.ShowPreview && m.ui.PreviewContent != "" {
		// Preprocess content to handle carriage returns and ensure proper line breaks
		processedContent := preprocessContent(m.ui.PreviewContent)
		// Wrap content to viewport width to prevent overflow
		wrappedContent := wordwrap.String(processedContent, m.viewports.Preview.Width)
		m.viewports.Preview.SetContent(wrappedContent)
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
			m.viewports.Preview, cmd = m.viewports.Preview.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		m.viewports.Pipelines, cmd = m.viewports.Pipelines.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.viewports.Components, cmd = m.viewports.Components.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *MainListModel) View() string {
	// Create main view renderer
	mainRenderer := NewMainViewRenderer(m.viewports.Width, m.viewports.Height)
	mainRenderer.ShowPreview = m.stateManager.ShowPreview
	mainRenderer.ActivePane = m.stateManager.ActivePane
	mainRenderer.LastDataPane = m.stateManager.LastDataPane
	mainRenderer.SearchBar = m.search.Bar
	mainRenderer.PreviewViewport = m.viewports.Preview
	mainRenderer.PreviewContent = m.ui.PreviewContent

	// Handle error state
	if m.err != nil {
		return mainRenderer.RenderErrorView(m.err)
	}

	// If showing exit confirmation, display dialog
	if m.ui.ExitConfirm.Active() {
		// Add padding to match other views
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		return contentStyle.Render(m.ui.ExitConfirm.View())
	}

	// If creating component, show creation wizard
	if m.operations.ComponentCreator.IsActive() {
		return m.componentCreationView()
	}

	// If editing component with enhanced editor, show enhanced edit view
	if m.editors.Enhanced.IsActive() {
		// Handle exit confirmation dialog
		if m.editors.Enhanced.ExitConfirmActive {
			// Add padding to match other views
			contentStyle := lipgloss.NewStyle().
				PaddingLeft(1).
				PaddingRight(1)
			return contentStyle.Render(m.editors.Enhanced.ExitConfirm.View())
		}

		renderer := NewEnhancedEditorRenderer(m.viewports.Width, m.viewports.Height)
		return renderer.Render(m.editors.Enhanced)
	}

	// If editing tags, show tag edit view
	if m.editors.TagEditor.Active {
		// Use the new unified tag editor renderer
		renderer := NewTagEditorRenderer(m.editors.TagEditor, m.viewports.Width, m.viewports.Height)
		return renderer.Render()
	}

	// Calculate content height
	contentHeight := mainRenderer.CalculateContentHeight()

	// Update search bar active state and render it
	m.search.Bar.SetActive(m.stateManager.IsInSearchPane())
	m.search.Bar.SetWidth(m.viewports.Width)

	// Create component view renderer
	componentRenderer := NewComponentViewRenderer(m.viewports.Width, contentHeight)
	componentRenderer.ActivePane = m.stateManager.ActivePane
	componentRenderer.FilteredComponents = m.data.FilteredComponents
	componentRenderer.AllComponents = m.getAllComponents()
	componentRenderer.ComponentCursor = m.stateManager.ComponentCursor
	componentRenderer.SearchQuery = m.search.Query
	// Use the persistent table renderer for proper scroll state
	if m.ui.ComponentTableRenderer != nil {
		componentRenderer.TableRenderer = m.ui.ComponentTableRenderer
		// Update the table renderer with current state
		m.ui.ComponentTableRenderer.SetComponents(m.data.FilteredComponents)
		m.ui.ComponentTableRenderer.SetCursor(m.stateManager.ComponentCursor)
		m.ui.ComponentTableRenderer.SetActive(m.stateManager.ActivePane == componentsPane)
	}

	// Create pipeline view renderer
	pipelineRenderer := NewPipelineViewRenderer(m.viewports.Width, contentHeight)
	pipelineRenderer.ActivePane = m.stateManager.ActivePane
	pipelineRenderer.Pipelines = m.data.Pipelines
	pipelineRenderer.FilteredPipelines = m.data.FilteredPipelines
	pipelineRenderer.PipelineCursor = m.stateManager.PipelineCursor
	pipelineRenderer.SearchQuery = m.search.Query
	pipelineRenderer.Viewport = m.viewports.Pipelines

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
	s.WriteString(m.search.Bar.View())
	s.WriteString("\n")

	// Then add the columns
	s.WriteString(contentStyle.Render(columns))

	// Add preview if enabled
	previewPane := mainRenderer.RenderPreviewPane(m.getCurrentPipelines(), m.data.FilteredComponents, m.stateManager.PipelineCursor, m.stateManager.ComponentCursor)
	if previewPane != "" {
		s.WriteString(previewPane)
	}

	// Show confirmation dialogs
	confirmDialogs := mainRenderer.RenderConfirmationDialogs(m.operations.PipelineOperator)
	if confirmDialogs != "" {
		s.WriteString(confirmDialogs)
	}

	// Help text
	s.WriteString("\n")
	s.WriteString(mainRenderer.RenderHelpPane(m.stateManager.IsInSearchPane()))

	finalView := s.String()

	// Overlay clone dialog if active
	if m.editors.Clone.State != nil && m.editors.Clone.State.Active && m.editors.Clone.Renderer != nil {
		m.editors.Clone.Renderer.SetSize(m.viewports.Width, m.viewports.Height)
		finalView = m.editors.Clone.Renderer.RenderOverlay(finalView, m.editors.Clone.State)
	}

	// Overlay rename dialog if active
	if m.editors.Rename.State != nil && m.editors.Rename.State.Active && m.editors.Rename.Renderer != nil {
		finalView = m.editors.Rename.Renderer.RenderOverlay(finalView, m.editors.Rename.State)
	}

	// Overlay tag reload status if active
	if m.operations.TagReloader != nil && m.operations.TagReloader.IsActive() && m.ui.TagReloadRenderer != nil {
		overlay := m.ui.TagReloadRenderer.RenderStatus(m.operations.TagReloader)
		if overlay != "" {
			// Combine the views by overlaying
			return overlayViews(finalView, overlay)
		}
	}

	return finalView
}

func (m *MainListModel) SetSize(width, height int) {
	m.viewports.Width = width
	m.viewports.Height = height
	// Update search bar width
	m.search.Bar.SetWidth(width)
	// Update tag editor size
	if m.editors.TagEditor != nil {
		m.editors.TagEditor.SetSize(width, height)
	}
	// Update tag reload renderer size
	if m.ui.TagReloadRenderer != nil {
		m.ui.TagReloadRenderer.SetSize(width, height)
	}
	// Update rename renderer size
	if m.editors.Rename.Renderer != nil {
		m.editors.Rename.Renderer.SetSize(width, height)
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
			m.ui.PreviewContent = renderer.RenderEmptyPreview(pipelinesPane, false, false)
			return
		}

		if m.stateManager.PipelineCursor >= 0 && m.stateManager.PipelineCursor < len(pipelines) {
			pipeline := pipelines[m.stateManager.PipelineCursor]
			m.ui.PreviewContent = renderer.RenderPipelinePreview(pipeline.path, pipeline.isArchived)
		}
	} else if previewPane == componentsPane {
		// Show component preview
		components := m.getCurrentComponents()
		if len(components) == 0 {
			m.ui.PreviewContent = renderer.RenderEmptyPreview(componentsPane, false, false)
			return
		}

		if m.stateManager.ComponentCursor >= 0 && m.stateManager.ComponentCursor < len(components) {
			comp := components[m.stateManager.ComponentCursor]
			m.ui.PreviewContent = renderer.RenderComponentPreview(comp)
		}
	}
}
func (m *MainListModel) handleComponentCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.operations.ComponentCreator.GetCurrentStep() {
	case 0: // Type selection
		if m.operations.ComponentCreator.HandleTypeSelection(msg) {
			return m, nil
		}

	case 1: // Name input
		if m.operations.ComponentCreator.HandleNameInput(msg) {
			return m, nil
		}

	case 2: // Content input
		// Check if enhanced editor is active for component creation
		if m.operations.ComponentCreator.IsEnhancedEditorActive() {
			handled, cmd := m.operations.ComponentCreator.HandleEnhancedEditorInput(msg, m.viewports.Width)
			if handled {
				// Check if component was saved (but editor stays open)
				if m.operations.ComponentCreator.WasSaveSuccessful() {
					// Component was saved, reload components
					m.reloadComponents()
					return m, tea.Batch(
						cmd,
						func() tea.Msg { return StatusMsg(m.operations.ComponentCreator.GetStatusMessage()) },
					)
				}
				// Check if component creation was cancelled
				if !m.operations.ComponentCreator.IsActive() {
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
	renderer := NewComponentCreationViewRenderer(m.viewports.Width, m.viewports.Height)

	switch m.operations.ComponentCreator.GetCurrentStep() {
	case 0:
		return renderer.RenderTypeSelection(m.operations.ComponentCreator.GetTypeCursor())
	case 1:
		return renderer.RenderNameInput(m.operations.ComponentCreator.GetComponentType(), m.operations.ComponentCreator.GetComponentName())
	case 2:
		// Use enhanced editor if available
		if m.operations.ComponentCreator.IsEnhancedEditorActive() {
			if adapter, ok := m.operations.ComponentCreator.GetEnhancedEditor().(*shared.EnhancedEditorAdapter); ok {
				if underlyingEditor, ok := adapter.GetUnderlyingEditor().(*EnhancedEditorState); ok {
					return renderer.RenderWithEnhancedEditor(
						underlyingEditor,
						m.operations.ComponentCreator.GetComponentType(),
						m.operations.ComponentCreator.GetComponentName(),
					)
				}
			}
		}
		// Fallback to simple editor
		return renderer.RenderContentEdit(m.operations.ComponentCreator.GetComponentType(), m.operations.ComponentCreator.GetComponentName(), m.operations.ComponentCreator.GetComponentContent())
	}

	return "Unknown creation step"
}
