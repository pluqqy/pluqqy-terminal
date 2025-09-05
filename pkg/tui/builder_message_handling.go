package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// Main Update Method

func (m *PipelineBuilderModel) Init() tea.Cmd {
	return nil
}

func (m *PipelineBuilderModel) performSearch() {
	// Initialize unified manager if needed
	m.search.InitializeUnifiedManager()
	
	if m.search.Query == "" {
		// No search query, check if we need to reload without archived items
		hasArchived := false
		for _, p := range m.data.Prompts {
			if p.isArchived {
				hasArchived = true
				break
			}
		}
		if !hasArchived {
			for _, c := range m.data.Contexts {
				if c.isArchived {
					hasArchived = true
					break
				}
			}
		}
		if !hasArchived {
			for _, r := range m.data.Rules {
				if r.isArchived {
					hasArchived = true
					break
				}
			}
		}

		// If we have archived items loaded, reload without them
		if hasArchived {
			m.loadAvailableComponents()
		}

		// Show all items
		m.data.FilteredPrompts = m.data.Prompts
		m.data.FilteredContexts = m.data.Contexts
		m.data.FilteredRules = m.data.Rules
		return
	}

	// Check if we need to reload data with archived items
	needsArchived := m.shouldIncludeArchived()
	hasArchived := false
	for _, p := range m.data.Prompts {
		if p.isArchived {
			hasArchived = true
			break
		}
	}

	// Reload data if archived status changed
	if needsArchived && !hasArchived {
		// Need to reload with archived items
		m.loadAvailableComponents()
	} else if !needsArchived && hasArchived {
		// Need to reload without archived items
		m.loadAvailableComponents()
	}

	// Try the new unified search first
	if m.search.UnifiedManager != nil {
		// Convert components to shared format for unified search
		sharedPrompts := convertComponentsToShared(m.data.Prompts)
		sharedContexts := convertComponentsToShared(m.data.Contexts)
		sharedRules := convertComponentsToShared(m.data.Rules)
		
		// Configure unified manager
		m.search.UnifiedManager.SetIncludeArchived(needsArchived)
		
		// Perform unified search
		filteredPrompts, filteredContexts, filteredRules, err := m.search.UnifiedManager.FilterComponentsByQuery(m.search.Query, sharedPrompts, sharedContexts, sharedRules)
		if err == nil {
			// Convert results back to TUI format
			m.data.FilteredPrompts = convertSharedComponentsToTUI(filteredPrompts)
			m.data.FilteredContexts = convertSharedComponentsToTUI(filteredContexts)
			m.data.FilteredRules = convertSharedComponentsToTUI(filteredRules)
			
			// Reset cursor if it's out of bounds
			if m.ui.ActiveColumn == leftColumn {
				totalItems := len(m.data.FilteredPrompts) + len(m.data.FilteredContexts) + len(m.data.FilteredRules)
				if m.ui.LeftCursor >= totalItems {
					m.ui.LeftCursor = 0
				}
			}
			return
		}
	}

	// Fallback to legacy search engine
	// Clear filtered lists
	m.data.FilteredPrompts = nil
	m.data.FilteredContexts = nil
	m.data.FilteredRules = nil

	// Legacy search engine is no longer used - show all items
	// The unified search system is used in the TUI search flow instead
	m.data.FilteredPrompts = m.data.Prompts
	m.data.FilteredContexts = m.data.Contexts
	m.data.FilteredRules = m.data.Rules

	// Reset cursor if it's out of bounds
	if m.ui.ActiveColumn == leftColumn {
		totalItems := len(m.data.FilteredPrompts) + len(m.data.FilteredContexts) + len(m.data.FilteredRules)
		if m.ui.LeftCursor >= totalItems {
			m.ui.LeftCursor = 0
		}
	}
}

func (m *PipelineBuilderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewports.Width = msg.Width
		m.viewports.Height = msg.Height
		// Update clone renderer size
		if m.editors.Clone.Renderer != nil {
			m.editors.Clone.Renderer.SetSize(msg.Width, msg.Height)
		}
		// Update rename renderer size
		if m.editors.Rename.Renderer != nil {
			m.editors.Rename.Renderer.SetSize(msg.Width, msg.Height)
		}
		m.updateViewportSizes()

	case componentSaveResultMsg:
		if msg.success {
			// Set save message
			m.editors.EditSaveMessage = msg.message

			// Exit editing mode immediately (like main list view does)
			m.editors.EditingComponent = false
			m.editors.ComponentContent = ""
			m.editors.EditingComponentPath = ""
			m.editors.EditingComponentName = ""
			m.editors.OriginalContent = ""

			// Cancel any existing timer
			if m.editors.EditSaveTimer != nil {
				m.editors.EditSaveTimer.Stop()
			}

			// Reload components
			m.loadAvailableComponents()
			
			// Also reload available tags for tag editor
			if m.editors.TagEditor != nil {
				m.editors.TagEditor.LoadAvailableTags()
			}

			// Update preview
			m.updatePreview()

			// Set timer to clear the save message after 1.5 seconds
			m.editors.EditSaveTimer = time.NewTimer(1500 * time.Millisecond)

			// Return a command to clear the message after timer
			return m, func() tea.Msg {
				<-m.editors.EditSaveTimer.C
				return clearEditSaveMsg{}
			}
		} else {
			// Handle save error
			m.editors.EditSaveMessage = msg.message
			return m, nil
		}

	case externalEditSaveMsg:
		// Update original content when saving before external edit
		m.editors.OriginalContent = msg.savedContent
		return m, nil

	case componentEditExitMsg:
		// Exit component editing without saving
		m.editors.EditingComponent = false
		m.editors.ComponentContent = ""
		m.editors.EditingComponentPath = ""
		m.editors.EditingComponentName = ""
		m.editors.EditSaveMessage = ""
		m.editors.OriginalContent = ""
		if m.editors.EditSaveTimer != nil {
			m.editors.EditSaveTimer.Stop()
			m.editors.EditSaveTimer = nil
		}
		return m, nil

	case clearEditSaveMsg:
		// Just clear the save message - editing mode was already exited when saving
		m.editors.EditSaveMessage = ""
		return m, nil

	case StatusMsg:
		// Handle status messages, especially save confirmations
		msgStr := string(msg)
		if strings.HasPrefix(msgStr, "✓ Saved:") {
			// A component was saved successfully
			if m.editors.EditingComponent {
				// Reload components but keep editor open
				m.loadAvailableComponents()
				// Update preview
				m.updatePreview()
			}
		}
		// Pass the message up to the parent app for display
		return m, func() tea.Msg { return msg }

	case tagDeletionCompleteMsg:
		// Handle tag deletion completion
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				// Reload components to reflect removed tags
				m.loadAvailableComponents()
				// Don't reload tags here - the tag editor manages its own tag list
				return m, cmd
			}
		}
		return m, nil

	case tagDeletionProgressMsg:
		// Handle tag deletion progress updates
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				return m, cmd
			}
		}
		return m, nil

	case spinner.TickMsg:
		// Handle spinner tick for tag deletion
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				return m, cmd
			}
		}
		// Continue to handle other spinners if needed
		return m, nil

	case TagReloadMsg:
		// Handle tag reload messages
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				// Reload components to reflect new tags
				m.loadAvailableComponents()
				// Also reload available tags for tag editor
				if m.editors.TagEditor != nil {
					m.editors.TagEditor.LoadAvailableTags()
				}
				return m, cmd
			}
		}
		return m, nil

	case tagReloadCompleteMsg:
		// Handle tag reload completion
		if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
			handled, cmd := m.editors.TagEditor.HandleMessage(msg)
			if handled {
				return m, cmd
			}
		}
		return m, nil

	case ReloadMsg:
		// Handle reload message from various sources (tag editor, component editor, etc.)
		
		// Set success message if provided
		if msg.Message != "" {
			m.ui.StatusMessage = msg.Message
		}
		
		// If coming from component editor
		if m.editors.EditingComponent {
			// Exit editing mode
			m.editors.EditingComponent = false
		}
		
		// Always reload components to get fresh data (including updated tags)
		m.loadAvailableComponents()
		
		// If we have a pipeline loaded, reload it to get fresh tags
		if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
			pipeline, err := files.ReadPipeline(m.data.Pipeline.Path)
			if err == nil && pipeline != nil {
				m.data.Pipeline.Tags = pipeline.Tags
			}
		}
		
		// Also reload available tags for tag editor
		if m.editors.TagEditor != nil {
			m.editors.TagEditor.LoadAvailableTags()
		}
		
		// Update preview to reflect changes
		m.updatePreview()
		
		// Re-run search if active to update filtered results
		if m.search.Query != "" {
			m.performSearch()
		}
		
		return m, nil

	case CloneSuccessMsg:
		// Handle successful clone
		m.editors.Clone.State.Reset()
		// Reload components to show new items
		m.loadAvailableComponents()
		// Set success message using StatusMsg
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
		// Set error message using StatusMsg
		return m, func() tea.Msg {
			return StatusMsg(fmt.Sprintf("✗ Clone failed: %v", msg.Error))
		}

	case RenameSuccessMsg:
		// Handle successful rename (same as Main List view)
		m.editors.Rename.State.Reset()
		// Reload components to show new names
		m.loadAvailableComponents()
		// If a pipeline was renamed, update the pipeline path
		if msg.ItemType == "pipeline" && m.data.Pipeline != nil {
			m.data.Pipeline.Path = filepath.Join(files.PipelinesDir, files.SanitizeFileName(msg.NewName)+".yaml")
			m.data.Pipeline.Name = msg.NewName
			m.editors.NameInput = msg.NewName
		}
		return m, nil

	case RenameErrorMsg:
		// Handle rename error (same as Main List view)
		m.editors.Rename.State.ValidationError = msg.Error.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Only forward non-key messages to viewports
	// Key messages are already handled above
	switch msg.(type) {
	case tea.KeyMsg:
		// Don't forward key messages - they're already handled
	default:
		// Handle enhanced editor for non-KeyMsg message types
		// This is crucial for filepicker which needs to process internal messages like directory reads
		if m.editors.Enhanced.IsActive() && m.editors.EditingComponent {
			if m.editors.Enhanced.IsFilePicking() {
				// Filepicker needs to process internal messages for directory reading
				cmd := m.editors.Enhanced.UpdateFilePicker(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

		// Forward other messages to viewports
		if m.ui.ShowPreview {
			m.viewports.Preview, cmd = m.viewports.Preview.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		m.viewports.RightViewport, cmd = m.viewports.RightViewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// Keyboard Input Handlers

func (m *PipelineBuilderModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle exit confirmation
	if m.ui.ExitConfirm.Active() {
		return m, m.ui.ExitConfirm.Update(msg)
	}

	// Handle delete confirmation
	if m.ui.DeleteConfirm.Active() {
		return m, m.ui.DeleteConfirm.Update(msg)
	}

	// Handle archive confirmation
	if m.ui.ArchiveConfirm.Active() {
		return m, m.ui.ArchiveConfirm.Update(msg)
	}

	// Handle clone mode
	if m.editors.Clone.State.IsActive() {
		handled, cmd := m.editors.Clone.State.HandleInput(msg)
		if handled {
			return m, cmd
		}
	}

	// Handle rename mode
	if m.editors.Rename.State.IsActive() {
		handled, cmd := m.editors.Rename.State.HandleInput(msg)
		if handled {
			// Check if rename was completed
			if !m.editors.Rename.State.IsActive() {
				// Rename completed, refresh the components list
				m.loadAvailableComponents()
				// If a pipeline was renamed, update the pipeline path
				if m.editors.Rename.State.GetItemType() == "pipeline" && m.data.Pipeline != nil {
					// Update the pipeline path with the new name
					newName := m.editors.Rename.State.GetNewName()
					m.data.Pipeline.Path = filepath.Join(files.PipelinesDir, files.SanitizeFileName(newName)+".yaml")
					m.data.Pipeline.Name = newName
					m.editors.NameInput = newName
				}
			}
			return m, cmd
		}
	}

	// Handle component usage mode
	if m.editors.ComponentUsage != nil && m.editors.ComponentUsage.Active {
		handled, cmd := m.editors.ComponentUsage.HandleInput(msg)
		if handled {
			return m, cmd
		}
	}

	// Handle component creation mode
	if m.editors.ComponentCreator != nil && m.editors.ComponentCreator.IsActive() {
		return m.handleComponentCreation(msg)
	}

	// Handle component editing mode
	if m.editors.EditingComponent {
		return m.handleComponentEditing(msg)
	}

	// Handle tag editing mode
	if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
		// Capture state before handling input
		wasActive := m.editors.TagEditor.Active
		itemType := m.editors.TagEditor.ItemType
		itemPath := m.editors.TagEditor.Path
		
		handled, cmd := m.editors.TagEditor.HandleInput(msg)
		if handled {
			// Check if save was completed (editor became inactive)
			if wasActive && !m.editors.TagEditor.Active {
				// Reload components to reflect tag changes
				m.loadAvailableComponents()
				// Also reload available tags for tag editor
				if m.editors.TagEditor != nil {
					m.editors.TagEditor.LoadAvailableTags()
				}
				
				// If we were editing pipeline tags, reload the pipeline to get the updated tags
				if itemType == "pipeline" && m.data.Pipeline != nil && itemPath != "" {
					// Read the pipeline from disk to get the freshly saved tags
					pipeline, err := files.ReadPipeline(itemPath)
					if err == nil && pipeline != nil {
						m.data.Pipeline.Tags = pipeline.Tags
					}
				}
				
				// Force refresh of the view to show updated tags
				m.updatePreview()
			}
			return m, cmd
		}
	}

	// Handle name editing mode
	if m.editors.EditingName {
		switch msg.String() {
		case "enter":
			if strings.TrimSpace(m.editors.NameInput) != "" {
				m.data.Pipeline.Name = strings.TrimSpace(m.editors.NameInput)
				m.editors.EditingName = false
			}
		case "esc":
			// Cancel and return to main list
			return m, func() tea.Msg {
				return SwitchViewMsg{view: mainListView}
			}
		case "backspace":
			if len(m.editors.NameInput) > 0 {
				m.editors.NameInput = m.editors.NameInput[:len(m.editors.NameInput)-1]
			}
		case " ":
			// Allow spaces
			m.editors.NameInput += " "
		default:
			// Add character to input
			if msg.Type == tea.KeyRunes {
				m.editors.NameInput += string(msg.Runes)
			}
		}
		return m, nil
	}

	// Handle search input when search column is active
	if m.ui.ActiveColumn == searchColumn && !m.editors.EditingName && !(m.editors.ComponentCreator != nil && m.editors.ComponentCreator.IsActive()) && !m.editors.EditingComponent {
		// Handle special keys in search first
		switch msg.String() {
		case "esc":
			// Clear search and switch to left column
			m.search.Bar.SetValue("")
			m.search.Query = ""
			m.performSearch()
			m.ui.ActiveColumn = leftColumn
			m.search.Bar.SetActive(false)
			return m, nil
		case "tab":
			// Let tab be handled by the main navigation logic
			// Don't process it here
		case "ctrl+a":
			// Toggle archived filter
			newQuery := m.search.FilterHelper.ToggleArchivedFilter(m.search.Bar.Value())
			m.search.Bar.SetValue(newQuery)
			m.search.Query = newQuery
			m.performSearch()
			return m, nil
		case "ctrl+t":
			// Cycle type filter (skip pipelines since we're in pipeline builder)
			newQuery := m.search.FilterHelper.CycleTypeFilterForComponents(m.search.Bar.Value())
			m.search.Bar.SetValue(newQuery)
			m.search.Query = newQuery
			m.performSearch()
			return m, nil
		default:
			// For all other keys, update the search bar
			var cmd tea.Cmd
			m.search.Bar, cmd = m.search.Bar.Update(msg)

			// Check if search query changed
			if m.search.Query != m.search.Bar.Value() {
				m.search.Query = m.search.Bar.Value()
				m.performSearch()
			}

			return m, cmd
		}
	}

	// Normal mode keybindings
	return m.handleNormalModeKeys(msg)
}

func (m *PipelineBuilderModel) handleNormalModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Check if there are unsaved changes
		if m.hasUnsavedChanges() {
			// Show confirmation dialog
			m.ui.ExitConfirmationType = "pipeline"
			m.ui.ExitConfirm.ShowDialog(
				"⚠️  Unsaved Changes",
				"You have unsaved changes in this pipeline.",
				"Exit without saving?",
				true, // destructive
				m.viewports.Width-4,
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
		// Switch between content columns only (no search)
		if m.ui.ShowPreview {
			// When preview is shown, cycle through content panes
			switch m.ui.ActiveColumn {
			case searchColumn:
				// If in search, exit to left column
				m.ui.ActiveColumn = leftColumn
				m.search.Bar.SetActive(false)
			case leftColumn:
				m.ui.ActiveColumn = rightColumn
				// Reset right cursor and viewport when entering right column
				m.ui.RightCursor = 0
				m.viewports.RightViewport.GotoTop()
			case rightColumn:
				m.ui.ActiveColumn = previewColumn
			case previewColumn:
				m.ui.ActiveColumn = leftColumn
			}
		} else {
			// When preview is hidden, cycle between left and right
			switch m.ui.ActiveColumn {
			case searchColumn:
				// If in search, exit to left column
				m.ui.ActiveColumn = leftColumn
				m.search.Bar.SetActive(false)
			case leftColumn:
				m.ui.ActiveColumn = rightColumn
				// Reset right cursor and viewport when entering right column
				m.ui.RightCursor = 0
				m.viewports.RightViewport.GotoTop()
			case rightColumn:
				m.ui.ActiveColumn = leftColumn
			}
		}

	case "shift+tab", "backtab":
		// Reverse cycle through columns
		if m.ui.ShowPreview {
			// When preview is shown, reverse cycle through content panes
			switch m.ui.ActiveColumn {
			case searchColumn:
				// If in search, exit to preview column
				m.ui.ActiveColumn = previewColumn
				m.search.Bar.SetActive(false)
			case leftColumn:
				m.ui.ActiveColumn = previewColumn
			case rightColumn:
				m.ui.ActiveColumn = leftColumn
			case previewColumn:
				m.ui.ActiveColumn = rightColumn
				// Reset right cursor and viewport when entering right column
				m.ui.RightCursor = 0
				m.viewports.RightViewport.GotoTop()
			}
		} else {
			// When preview is hidden, reverse cycle between left and right
			switch m.ui.ActiveColumn {
			case searchColumn:
				// If in search, exit to right column
				m.ui.ActiveColumn = rightColumn
				m.search.Bar.SetActive(false)
				// Reset right cursor and viewport when entering right column
				m.ui.RightCursor = 0
				m.viewports.RightViewport.GotoTop()
			case leftColumn:
				m.ui.ActiveColumn = rightColumn
				// Reset right cursor and viewport when entering right column
				m.ui.RightCursor = 0
				m.viewports.RightViewport.GotoTop()
			case rightColumn:
				m.ui.ActiveColumn = leftColumn
			}
		}
		// Update preview when switching to non-preview column
		if m.ui.ActiveColumn != previewColumn {
			m.updatePreview()
		}

	case "up", "k":
		if m.ui.ActiveColumn == previewColumn {
			// Scroll preview up
			m.viewports.Preview.LineUp(1)
		} else {
			m.moveCursor(-1)
		}

	case "down", "j":
		if m.ui.ActiveColumn == previewColumn {
			// Scroll preview down
			m.viewports.Preview.LineDown(1)
		} else {
			m.moveCursor(1)
		}

	case "pgup":
		if m.ui.ActiveColumn == previewColumn {
			// Scroll preview page up
			m.viewports.Preview.ViewUp()
		} else if m.ui.ActiveColumn == rightColumn {
			// Page up in right column - move cursor up by viewport height
			pageSize := m.viewports.RightViewport.Height
			for i := 0; i < pageSize && m.ui.RightCursor > 0; i++ {
				m.ui.RightCursor--
			}
			// Adjust viewport to ensure cursor is visible
			m.adjustRightViewportScroll()
		}

	case "pgdown":
		if m.ui.ActiveColumn == previewColumn {
			// Scroll preview page down
			m.viewports.Preview.ViewDown()
		} else if m.ui.ActiveColumn == rightColumn {
			// Page down in right column - move cursor down by viewport height
			pageSize := m.viewports.RightViewport.Height
			maxCursor := len(m.data.SelectedComponents) - 1
			for i := 0; i < pageSize && m.ui.RightCursor < maxCursor; i++ {
				m.ui.RightCursor++
			}
			// Adjust viewport to ensure cursor is visible
			m.adjustRightViewportScroll()
		}

	case "home":
		if m.ui.ActiveColumn == leftColumn {
			m.ui.LeftCursor = 0
		} else if m.ui.ActiveColumn == rightColumn {
			m.ui.RightCursor = 0
			// Adjust viewport to show the first item
			m.adjustRightViewportScroll()
			// Sync preview scroll when jumping to first component in pipeline
			if m.ui.ShowPreview && len(m.data.SelectedComponents) > 0 {
				m.syncPreviewToSelectedComponent()
			}
		}

	case "end":
		if m.ui.ActiveColumn == leftColumn {
			components := m.getAllAvailableComponents()
			if len(components) > 0 {
				m.ui.LeftCursor = len(components) - 1
			}
		} else if m.ui.ActiveColumn == rightColumn {
			if len(m.data.SelectedComponents) > 0 {
				m.ui.RightCursor = len(m.data.SelectedComponents) - 1
				// Adjust viewport to show the last item
				m.adjustRightViewportScroll()
			}
			// Sync preview scroll when jumping to last component in pipeline
			if m.ui.ShowPreview && len(m.data.SelectedComponents) > 0 {
				m.syncPreviewToSelectedComponent()
			}
		}

	case "enter":
		if m.ui.ActiveColumn == leftColumn {
			m.addSelectedComponent()
		} else if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
			// Remove selected component in right column (same as delete)
			m.removeSelectedComponent()
		}

	case "p":
		m.ui.ShowPreview = !m.ui.ShowPreview
		m.updateViewportSizes()
		if m.ui.ShowPreview {
			m.updatePreview()
		}
		
	case "t":
		// Edit tags - context aware based on active column
		if m.ui.ActiveColumn == leftColumn {
			// Edit component tags
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				m.startTagEditing(comp.path, comp.tags)
			}
		} else if m.ui.ActiveColumn == rightColumn {
			// Edit pipeline tags
			if m.data.Pipeline != nil {
				m.startPipelineTagEditing(m.data.Pipeline.Tags)
			}
		}

	case "a":
		// Archive/Unarchive - context aware based on active column
		if m.ui.ActiveColumn == leftColumn {
			// Archive/Unarchive component
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				if comp.isArchived {
					// Unarchive the component with confirmation
					m.ui.ArchiveConfirm.ShowInline(
						fmt.Sprintf("Unarchive %s '%s'?", comp.compType, comp.name),
						false, // not destructive
						func() tea.Cmd {
							return m.unarchiveComponent(comp)
						},
						func() tea.Cmd {
							return nil
						},
					)
				} else {
					// Archive the component with confirmation
					m.ui.ArchiveConfirm.ShowInline(
						fmt.Sprintf("Archive %s '%s'?", comp.compType, comp.name),
						false, // not destructive
						func() tea.Cmd {
							return m.archiveComponent(comp)
						},
						func() tea.Cmd {
							return nil
						},
					)
				}
			}
		} else if m.ui.ActiveColumn == rightColumn {
			// Archive the pipeline being edited
			if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
				pipelineName := strings.TrimSuffix(filepath.Base(m.data.Pipeline.Path), ".yaml")
				m.ui.ArchiveConfirm.ShowInline(
					fmt.Sprintf("Archive pipeline '%s'?", pipelineName),
					false, // not destructive
					func() tea.Cmd {
						return m.archivePipeline()
					},
					func() tea.Cmd {
						return nil
					},
				)
			}
		}
		return m, nil

	case "C":
		// Clone - context aware based on active column
		if m.editors.Clone.State.IsActive() {
			// Already in clone mode, ignore
			return m, nil
		}

		if m.ui.ActiveColumn == leftColumn {
			// Clone component
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				// Start clone for component
				displayName, path, isArchived := m.editors.Clone.Operator.PrepareCloneComponent(comp)
				m.editors.Clone.State.Start(displayName, "component", path, isArchived)
			}
		} else if m.ui.ActiveColumn == rightColumn {
			// Clone the pipeline being edited
			if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
				// Use the pipeline's display name from the Name field
				pipelineItem := pipelineItem{
					name:       m.data.Pipeline.Name,
					path:       m.data.Pipeline.Path,
					isArchived: false, // Builder only works with active pipelines
				}
				displayName, path, isArchived := m.editors.Clone.Operator.PrepareClonePipeline(pipelineItem)
				m.editors.Clone.State.Start(displayName, "pipeline", path, isArchived)
			}
		}
		return m, nil

	case "R":
		// Rename - context aware based on active column
		if m.editors.Rename.State.IsActive() {
			// Already in rename mode, ignore
			return m, nil
		}

		if m.ui.ActiveColumn == leftColumn {
			// Rename component
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				// Start rename for component
				m.editors.Rename.State.StartRename(comp.path, comp.name, "component")
			}
		} else if m.ui.ActiveColumn == rightColumn {
			// Rename the pipeline being edited
			if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
				// Use the pipeline's display name from the Name field
				m.editors.Rename.State.StartRename(m.data.Pipeline.Path, m.data.Pipeline.Name, "pipeline")
			}
		}
		return m, nil

	case "M":
		// Generate mermaid diagram for current pipeline
		if !m.ui.MermaidState.IsGenerating() && m.data.Pipeline != nil {
			// Create a pipelineItem from the current pipeline
			pipelineItem := pipelineItem{
				name: m.data.Pipeline.Name,
				path: m.data.Pipeline.Path,
				tags: m.data.Pipeline.Tags,
			}
			// If path is empty (new pipeline), use the name
			if pipelineItem.path == "" {
				pipelineItem.path = files.SanitizeFileName(m.data.Pipeline.Name) + ".yaml"
			}
			return m, m.ui.MermaidOperator.GeneratePipelineDiagram(pipelineItem)
		}
		return m, nil

	case "/":
		// Jump to search
		m.ui.ActiveColumn = searchColumn
		m.search.Bar.SetActive(true)
		return m, nil

	case "ctrl+s":
		// Save pipeline
		return m, m.savePipeline()

	case "ctrl+d":
		// Handle deletion based on active column
		if m.ui.ActiveColumn == leftColumn {
			// Delete component from Available Components pane
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				m.ui.DeleteConfirm.ShowInline(
					fmt.Sprintf("Delete %s '%s'?", comp.compType, comp.name),
					true, // destructive
					func() tea.Cmd {
						return m.deleteComponentFromLeft(comp)
					},
					func() tea.Cmd {
						return nil
					},
				)
			}
		} else if m.ui.ActiveColumn != previewColumn && m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
			// Delete pipeline with confirmation (not in preview pane or left column)
			pipelineName := filepath.Base(m.data.Pipeline.Path)
			m.ui.DeleteConfirm.ShowInline(
				fmt.Sprintf("Delete pipeline '%s'?", pipelineName),
				true, // destructive
				func() tea.Cmd {
					return m.deletePipeline()
				},
				func() tea.Cmd {
					return nil
				},
			)
		}
		return m, nil

	case "S":
		// Save and set pipeline (generate PLUQQY.md)
		return m, m.saveAndSetPipeline()

	case "y":
		// Copy current pipeline content to clipboard
		if m.data.Pipeline != nil && len(m.data.Pipeline.Components) > 0 {
			output, err := composer.ComposePipeline(m.data.Pipeline)
			if err == nil {
				if err := clipboard.WriteAll(output); err == nil {
					return m, func() tea.Msg {
						return StatusMsg(m.data.Pipeline.Name + " → clipboard")
					}
				}
			}
		}

	case "ctrl+up", "K":
		if m.ui.ActiveColumn == rightColumn {
			m.moveComponentUp()
		}

	case "ctrl+down", "J":
		if m.ui.ActiveColumn == rightColumn {
			m.moveComponentDown()
		}

	case "n":
		// Create new component (not in preview pane)
		if m.ui.ActiveColumn != previewColumn {
			m.editors.ComponentCreator.Start()
			return m, nil
		}

	case "ctrl+x":
		// Edit component in external editor
		if m.ui.ActiveColumn == leftColumn {
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				// Return batch: status message first, then editor command
				return m, tea.Batch(
					func() tea.Msg {
						return PersistentStatusMsg("Editing in external editor - save your changes and close the editor window/tab to return here and continue")
					},
					m.editComponentFromLeft(),
				)
			}
		} else if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
			// Edit selected component in external editor from right column
			// Return batch: status message first, then editor command
			return m, tea.Batch(
				func() tea.Msg {
					return PersistentStatusMsg("Editing in external editor - save your changes and close the editor window/tab to return here and continue")
				},
				m.editComponent(),
			)
		}

	case "u":
		// Show component usage
		if m.ui.ActiveColumn == leftColumn {
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				m.editors.ComponentUsage.Start(comp)
				m.editors.ComponentUsage.SetSize(m.viewports.Width, m.viewports.Height)
			}
		} else if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
			// Show usage for selected component in the pipeline
			selected := m.data.SelectedComponents[m.ui.RightCursor]
			// Convert ComponentRef to componentItem for usage display
			comp := componentItem{
				name:     filepath.Base(selected.Path),
				path:     selected.Path,
				compType: selected.Type,
			}
			m.editors.ComponentUsage.Start(comp)
			m.editors.ComponentUsage.SetSize(m.viewports.Width, m.viewports.Height)
		}

	case "e":
		// Edit component in the TUI editor using the enhanced editor
		// Enhanced editor provides advanced editing features
		if m.ui.ActiveColumn == leftColumn {
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				// Read the component content
				content, err := files.ReadComponent(comp.path)
				if err != nil {
					m.err = err
					return m, nil
				}

				// Start enhanced editor
				m.editors.Enhanced.StartEditing(
					comp.path,
					comp.name,
					comp.compType,
					content.Content,
					comp.tags,
				)
				m.editors.EditingComponent = true
				return m, nil
			}
		} else if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
			// Edit component from right column
			if m.ui.RightCursor >= 0 && m.ui.RightCursor < len(m.data.SelectedComponents) {
				selected := m.data.SelectedComponents[m.ui.RightCursor]
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

				// Determine component type
				compType := ""
				if strings.Contains(componentPath, "/prompts/") {
					compType = models.ComponentTypePrompt
				} else if strings.Contains(componentPath, "/contexts/") {
					compType = models.ComponentTypeContext
				} else if strings.Contains(componentPath, "/rules/") {
					compType = models.ComponentTypeRules
				}

				// Start enhanced editor
				m.editors.Enhanced.StartEditing(
					componentPath,
					componentName,
					compType,
					content.Content,
					content.Tags,
				)
				m.editors.EditingComponent = true
				return m, nil
			}
		}
	}

	// Update preview if needed
	m.updatePreview()

	// Update viewport content if preview changed
	if m.ui.ShowPreview && m.ui.PreviewContent != "" {
		// Preprocess content to handle carriage returns and ensure proper line breaks
		processedContent := strings.ReplaceAll(m.ui.PreviewContent, "\r\r", "\n\n")
		processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
		// Wrap content to viewport width to prevent overflow
		wrappedContent := wordwrap.String(processedContent, m.viewports.Preview.Width)
		m.viewports.Preview.SetContent(wrappedContent)

		// Sync preview scroll to highlighted component if right column is active
		if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
			m.syncPreviewToSelectedComponent()
		}
	}

	return m, nil
}

func (m *PipelineBuilderModel) handleComponentCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.editors.ComponentCreator.GetCurrentStep() {
	case 0: // Type selection
		if m.editors.ComponentCreator.HandleTypeSelection(msg) {
			if !m.editors.ComponentCreator.IsActive() {
				// Component creation was cancelled
				return m, nil
			}
			return m, nil
		}

	case 1: // Name input
		if m.editors.ComponentCreator.HandleNameInput(msg) {
			if !m.editors.ComponentCreator.IsActive() {
				// Component creation was cancelled
				return m, nil
			}
			return m, nil
		}

	case 2: // Content input
		// Use enhanced editor if enabled and active
		if m.editors.ComponentCreator.IsEnhancedEditorActive() {
			// Use the ComponentCreator's handler which properly manages save state
			handled, cmd := m.editors.ComponentCreator.HandleEnhancedEditorInput(msg, m.viewports.Width)
			if handled {
				// Check if component was saved (but editor stays open)
				if m.editors.ComponentCreator.WasSaveSuccessful() {
					// Component was saved successfully
					m.loadAvailableComponents()
					// Also reload available tags for tag editor
					if m.editors.TagEditor != nil {
						m.editors.TagEditor.LoadAvailableTags()
					}
					return m, tea.Batch(
						cmd,
						func() tea.Msg {
							return StatusMsg(m.editors.ComponentCreator.GetStatusMessage())
						},
					)
				}
				// Check if component creation was cancelled
				if !m.editors.ComponentCreator.IsActive() {
					// Component creation ended (saved or cancelled)
					// Always reload to ensure list is current
					m.loadAvailableComponents()
					// Also reload available tags for tag editor
					if m.editors.TagEditor != nil {
						m.editors.TagEditor.LoadAvailableTags()
					}
					return m, cmd
				}
				return m, cmd
			}
		}
		// If we get here, the enhanced editor didn't handle the input
		// which shouldn't happen, but return anyway
		return m, nil
	}

	return m, nil
}

// handleComponentEditing handles keyboard input when editing a component
// This is a key integration point that routes input to either the enhanced
// editor or the legacy editor based on configuration
func (m *PipelineBuilderModel) handleComponentEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input through the enhanced editor
	if m.editors.Enhanced.IsActive() {
		handled, cmd := HandleEnhancedEditorInput(m.editors.Enhanced, msg, m.viewports.Width)
		if handled {
			// Check if editor is still active after handling input
			if !m.editors.Enhanced.IsActive() {
				// Editor was closed, exit editing mode
				m.editors.EditingComponent = false
				// Reload components to reflect any changes
				m.loadAvailableComponents()
			}
			return m, cmd
		}
		return m, nil
	}

	// If enhanced editor is not active but we're in editing mode, something is wrong
	// Exit editing mode
	if m.editors.EditingComponent {
		m.editors.EditingComponent = false
	}

	return m, nil
}

// Component Operations

func (m *PipelineBuilderModel) editComponent() tea.Cmd {
	if m.ui.RightCursor >= 0 && m.ui.RightCursor < len(m.data.SelectedComponents) {
		comp := m.data.SelectedComponents[m.ui.RightCursor]
		componentPath := filepath.Join(files.PipelinesDir, comp.Path)
		componentPath = filepath.Clean(componentPath)

		return m.openInEditor(componentPath)
	}
	return nil
}

func (m *PipelineBuilderModel) editComponentFromLeft() tea.Cmd {
	components := m.getAllAvailableComponents()
	if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
		comp := components[m.ui.LeftCursor]
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

		fullPath := filepath.Join(files.PluqqyDir, path)
		
		// Create command with proper argument parsing for editors with flags
		cmd := createEditorCommand(editor, fullPath)
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