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
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
	"github.com/pluqqy/pluqqy-cli/pkg/search/unified"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

type column int

const (
	searchColumn column = iota
	leftColumn
	rightColumn
	previewColumn
)

// Constants for preview synchronization
const (
	defaultLinesPerComponent = 15 // Estimated lines per component in preview
	minLinesPerComponent     = 10 // Minimum estimate for lines per component
	scrollContextLines       = 2  // Lines to show before component when scrolling
	scrollBottomPadding      = 10 // Lines to keep from bottom when estimating position
)

type PipelineBuilderModel struct {
	// Composed data structures
	data      *BuilderDataStore
	viewports *BuilderViewportManager
	editors   *BuilderEditorComponents
	search    *BuilderSearchComponents
	ui        *BuilderUIComponents

	// Error handling
	err error
}

type componentItem struct {
	name         string
	path         string
	compType     string
	lastModified time.Time
	usageCount   int
	tokenCount   int
	tags         []string
	isArchived   bool
}

type clearEditSaveMsg struct{}

type componentSaveResultMsg struct {
	success       bool
	message       string
	componentPath string
	componentName string
	savedContent  string
}

type externalEditSaveMsg struct {
	savedContent string
}

type componentEditExitMsg struct{}

// NewPipelineBuilderModel creates a new Pipeline Builder with default configuration
// For custom configuration, use NewPipelineBuilderModelWithConfig
func NewPipelineBuilderModel() *PipelineBuilderModel {
	return NewPipelineBuilderModelWithConfig(DefaultPipelineBuilderConfig())
}

func (m *PipelineBuilderModel) loadAvailableComponents() {
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
	m.data.Prompts = convertBuilderComponentItems(unified.ConvertSharedComponentItemsToUnified(prompts))
	m.data.Contexts = convertBuilderComponentItems(unified.ConvertSharedComponentItemsToUnified(contexts))
	m.data.Rules = convertBuilderComponentItems(unified.ConvertSharedComponentItemsToUnified(rules))

	// Initialize filtered lists with all components
	m.data.FilteredPrompts = m.data.Prompts
	m.data.FilteredContexts = m.data.Contexts
	m.data.FilteredRules = m.data.Rules

	// Note: Search index rebuilding is now handled by the unified search manager
	// No explicit index rebuilding needed
}

// convertBuilderComponentItems converts unified ComponentItems to local componentItems
func convertBuilderComponentItems(items []unified.ComponentItem) []componentItem {
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

// convertComponentsToShared converts local componentItems to unified ComponentItems
func convertComponentsToShared(items []componentItem) []unified.ComponentItem {
	result := make([]unified.ComponentItem, len(items))
	for i, item := range items {
		result[i] = unified.ComponentItem{
			Name:         item.name,
			Path:         item.path,
			CompType:     item.compType,
			LastModified: item.lastModified,
			UsageCount:   item.usageCount,
			TokenCount:   item.tokenCount,
			Tags:         item.tags,
			IsArchived:   item.isArchived,
		}
	}
	return result
}

// convertSharedComponentsToTUI converts unified ComponentItems back to TUI componentItems
func convertSharedComponentsToTUI(items []unified.ComponentItem) []componentItem {
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

// shouldIncludeArchived checks if the current search query requires archived items
func (m *PipelineBuilderModel) shouldIncludeArchived() bool {
	return unified.ShouldIncludeArchived(m.search.Query)
}


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

		// If preview is showing and active, handle viewport navigation

		// Normal mode keybindings
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
					return m, m.editComponentFromLeft()
				}
			} else if m.ui.ActiveColumn == rightColumn && len(m.data.SelectedComponents) > 0 {
				// Edit selected component in external editor from right column
				return m, m.editComponent()
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

func (m *PipelineBuilderModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'Esc' to return", m.err)
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
	if m.editors.ComponentCreator != nil && m.editors.ComponentCreator.IsActive() {
		return m.componentCreationView()
	}

	// Render the enhanced editor view if editing a component
	// The enhanced editor provides a richer editing experience with file browsing and better text manipulation
	if m.editors.EditingComponent && m.editors.Enhanced.IsActive() {
		// Handle exit confirmation dialog
		if m.editors.Enhanced.ExitConfirmActive {
			// Add padding to match other views
			contentStyle := lipgloss.NewStyle().
				PaddingLeft(1).
				PaddingRight(1)
			return contentStyle.Render(m.editors.Enhanced.ExitConfirm.View())
		}

		// Render enhanced editor view
		renderer := NewEnhancedEditorRenderer(m.viewports.Width, m.viewports.Height)
		return renderer.Render(m.editors.Enhanced)
	}

	// If editing tags, show tag edit view
	if m.editors.TagEditor != nil && m.editors.TagEditor.Active {
		renderer := NewTagEditorRenderer(m.editors.TagEditor, m.viewports.Width, m.viewports.Height)
		return renderer.Render()
	}

	// If editing name, show name input screen
	if m.editors.EditingName {
		return m.nameInputView()
	}

	// Styles

	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")). // Purple to match MLV
		Background(lipgloss.Color("236")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	typeHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	// Update shared layout and get dimensions
	if m.ui.SharedLayout == nil {
		m.ui.SharedLayout = NewSharedLayout(m.viewports.Width, m.viewports.Height, m.ui.ShowPreview)
	} else {
		m.ui.SharedLayout.SetSize(m.viewports.Width, m.viewports.Height)
		m.ui.SharedLayout.SetShowPreview(m.ui.ShowPreview)
	}

	// Get calculated dimensions from shared layout
	columnWidth := m.ui.SharedLayout.GetColumnWidth()
	contentHeight := m.ui.SharedLayout.GetContentHeight()

	// Update table renderer for left column
	allComponents := m.getAllAvailableComponents()
	m.viewports.LeftTable.SetSize(columnWidth, contentHeight)
	m.viewports.LeftTable.SetComponents(allComponents)
	m.viewports.LeftTable.SetCursor(m.ui.LeftCursor)
	m.viewports.LeftTable.SetActive(m.ui.ActiveColumn == leftColumn)

	// Mark already added components
	m.viewports.LeftTable.ClearAddedMarks()
	for _, comp := range allComponents {
		componentPath := "../" + comp.path
		for _, existing := range m.data.SelectedComponents {
			if existing.Path == componentPath {
				m.viewports.LeftTable.MarkAsAdded(componentPath)
				break
			}
		}
	}

	// Build left column (available components)
	var leftContent strings.Builder
	// Create padding style for headers
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Render column header using shared layout
	leftHeader := m.ui.SharedLayout.RenderColumnHeader(ColumnHeaderConfig{
		Heading:     "AVAILABLE COMPONENTS",
		Active:      m.ui.ActiveColumn == leftColumn,
		ColumnWidth: columnWidth,
	})
	leftContent.WriteString(leftHeader)
	leftContent.WriteString("\n")

	// Add empty row to match pipeline name row height on the right
	if m.data.Pipeline != nil && m.data.Pipeline.Name != "" {
		// Match the exact spacing of the pipeline name row
		emptyRowStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			PaddingTop(1).
			Height(2) // Match height of pipeline name row
		leftContent.WriteString(emptyRowStyle.Render(""))
		leftContent.WriteString("\n")
	} else {
		// Add empty space if no pipeline name
		leftContent.WriteString("\n\n")
	}

	// Render table header
	leftContent.WriteString(headerPadding.Render(m.viewports.LeftTable.RenderHeader()))
	leftContent.WriteString("\n\n")

	// Add padding to table content
	leftViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	leftContent.WriteString(leftViewportPadding.Render(m.viewports.LeftTable.RenderTable()))

	// Build right column (selected components)
	var rightContent strings.Builder
	// Render column header using shared layout
	rightHeader := m.ui.SharedLayout.RenderColumnHeader(ColumnHeaderConfig{
		Heading:     "PIPELINE COMPONENTS",
		Active:      m.ui.ActiveColumn == rightColumn,
		ColumnWidth: columnWidth,
	})
	rightContent.WriteString(rightHeader)
	rightContent.WriteString("\n")

	// Add pipeline name with spacing
	if m.data.Pipeline != nil && m.data.Pipeline.Name != "" {
		pipelineNameStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			PaddingTop(1).
			Bold(true).
			Foreground(lipgloss.Color("0")) // Black text for better visibility on light backgrounds
		rightContent.WriteString(pipelineNameStyle.Render(m.data.Pipeline.Name))
		rightContent.WriteString("\n")
	} else {
		// Add empty space if no pipeline name
		rightContent.WriteString("\n\n")
	}

	// Always render tag row (even if empty) for consistent layout
	tagRowStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		PaddingTop(1).    // Add top margin
		PaddingBottom(1). // Add bottom margin
		Height(3)         // Total height including padding

	if m.data.Pipeline != nil && len(m.data.Pipeline.Tags) > 0 {
		// Render tags with more space available (full column width minus padding)
		tagsStr := renderTagChipsWithWidth(m.data.Pipeline.Tags, columnWidth-4, 5) // Show more tags with available space
		rightContent.WriteString(tagRowStyle.Render(tagsStr))
	} else {
		// Empty row to maintain layout
		rightContent.WriteString(tagRowStyle.Render(" "))
	}
	rightContent.WriteString("\n")

	// Build scrollable content for right viewport
	var rightScrollContent strings.Builder

	if len(m.data.SelectedComponents) == 0 {
		rightScrollContent.WriteString(normalStyle.Render("No components selected\n\nPress Tab to switch columns\nPress Enter to add components"))
	} else {
		// Load settings for section order
		settings, err := files.ReadSettings()
		if err != nil || settings == nil {
			settings = models.DefaultSettings()
		}

		// Group components by type
		typeGroups := make(map[string][]models.ComponentRef)
		for _, comp := range m.data.SelectedComponents {
			typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
		}

		// Track overall index for cursor position
		overallIndex := 0
		remainingSections := 0

		// Count how many sections we'll actually display
		for _, section := range settings.Output.Formatting.Sections {
			if len(typeGroups[section.Type]) > 0 {
				remainingSections++
			}
		}

		// Render sections in the configured order
		for _, section := range settings.Output.Formatting.Sections {
			components, exists := typeGroups[section.Type]
			if !exists || len(components) == 0 {
				continue
			}

			// Get the display name for this section type
			var sectionHeader string
			switch section.Type {
			case models.ComponentTypeContext:
				sectionHeader = "CONTEXTS"
			case models.ComponentTypePrompt:
				sectionHeader = "PROMPTS"
			case models.ComponentTypeRules:
				sectionHeader = "RULES"
			default:
				sectionHeader = strings.ToUpper(section.Type)
			}

			rightScrollContent.WriteString(typeHeaderStyle.Render("▸ "+sectionHeader) + "\n")

			for _, comp := range components {
				name := filepath.Base(comp.Path)

				if m.ui.ActiveColumn == rightColumn && overallIndex == m.ui.RightCursor {
					// White arrow with selected name
					rightScrollContent.WriteString("▸ " + selectedStyle.Render(name) + "\n")
				} else {
					// Normal styling
					rightScrollContent.WriteString("  " + normalStyle.Render(name) + "\n")
				}
				overallIndex++
			}

			// Add spacing between sections (but not after the last one)
			remainingSections--
			if remainingSections > 0 {
				rightScrollContent.WriteString("\n")
			}
		}
	}

	// Update right viewport with content
	// Wrap content to viewport width to prevent overflow
	wrappedRightContent := wordwrap.String(rightScrollContent.String(), m.viewports.RightViewport.Width)
	m.viewports.RightViewport.SetContent(wrappedRightContent)

	// Update viewport to follow cursor (even when right column is not active)
	if len(m.data.SelectedComponents) > 0 {
		// Load settings for section order
		settings, err := files.ReadSettings()
		if err != nil || settings == nil {
			settings = models.DefaultSettings()
		}

		// Calculate the line position of the cursor
		currentLine := 0
		overallIndex := 0

		// Group components by type
		typeGroups := make(map[string][]models.ComponentRef)
		for _, comp := range m.data.SelectedComponents {
			typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
		}

		// Count lines up to cursor position following section order
		for sectionIdx, section := range settings.Output.Formatting.Sections {
			components, exists := typeGroups[section.Type]
			if !exists || len(components) == 0 {
				continue
			}

			currentLine++ // Section header

			for range components {
				if overallIndex == m.ui.RightCursor {
					break
				}
				currentLine++
				overallIndex++
			}

			// Check if we found the cursor
			if overallIndex >= m.ui.RightCursor {
				break
			}

			// Add empty line if there are more sections
			hasMoreSections := false
			for j := sectionIdx + 1; j < len(settings.Output.Formatting.Sections); j++ {
				if len(typeGroups[settings.Output.Formatting.Sections[j].Type]) > 0 {
					hasMoreSections = true
					break
				}
			}
			if hasMoreSections {
				currentLine++ // Empty line between sections
			}
		}

		// Ensure the cursor line is visible
		if currentLine < m.viewports.RightViewport.YOffset {
			m.viewports.RightViewport.SetYOffset(currentLine)
		} else if currentLine >= m.viewports.RightViewport.YOffset+m.viewports.RightViewport.Height {
			m.viewports.RightViewport.SetYOffset(currentLine - m.viewports.RightViewport.Height + 1)
		}
	}

	// Add padding to viewport content
	rightViewportPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	rightContent.WriteString(rightViewportPadding.Render(m.viewports.RightViewport.View()))

	// Apply borders
	leftStyle := inactiveStyle
	rightStyle := inactiveStyle
	if m.ui.ActiveColumn == leftColumn {
		leftStyle = activeStyle
	} else if m.ui.ActiveColumn == rightColumn {
		rightStyle = activeStyle
	}

	leftColumnView := leftStyle.
		Width(columnWidth).
		Height(contentHeight).
		Render(leftContent.String())

	rightColumnView := rightStyle.
		Width(columnWidth).
		Height(contentHeight).
		Render(rightContent.String())

	// Join columns
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumnView, " ", rightColumnView)

	// Build final view
	var s strings.Builder

	// Add search bar at the top
	// Update search bar active state and render it
	m.search.Bar.SetActive(m.ui.ActiveColumn == searchColumn)
	m.search.Bar.SetWidth(m.viewports.Width)
	s.WriteString(m.search.Bar.View())
	s.WriteString("\n")

	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	s.WriteString(contentStyle.Render(columns))

	// Add preview if enabled
	if m.ui.ShowPreview && m.ui.PreviewContent != "" {
		// Calculate token count
		tokenCount := utils.EstimateTokens(m.ui.PreviewContent)
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
		if m.ui.ActiveColumn == previewColumn {
			previewBorderColor = lipgloss.Color("170") // active (same as other active borders)
		}

		previewBorderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(previewBorderColor).
			Width(m.viewports.Width - 4) // Account for padding (2) and border (2)

		// Build preview content with header inside
		var previewContent strings.Builder
		// Create heading with colons and token info
		var previewHeading string
		if m.ui.ActiveColumn == leftColumn {
			// Get the selected component name
			components := m.getAllAvailableComponents()
			if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
				comp := components[m.ui.LeftCursor]
				// Use the actual filename from the path
				componentFilename := filepath.Base(comp.path)
				previewHeading = fmt.Sprintf("COMPONENT PREVIEW (%s)", componentFilename)
			} else {
				previewHeading = "COMPONENT PREVIEW"
			}
		} else {
			pipelineName := "PLUQQY.md"
			if m.data.Pipeline != nil && m.data.Pipeline.Path != "" {
				// Use the actual filename from the path
				pipelineName = filepath.Base(m.data.Pipeline.Path)
			} else if m.data.Pipeline != nil && m.data.Pipeline.Name != "" {
				// For new unsaved pipelines, use the name with .yaml extension
				pipelineName = files.SanitizeFileName(m.data.Pipeline.Name) + ".yaml"
			}
			previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
		}
		tokenInfo := tokenBadge

		// Calculate the actual rendered width of token info
		tokenInfoWidth := lipgloss.Width(tokenBadge)

		// Calculate total available width inside the border
		totalWidth := m.viewports.Width - 8 // accounting for border padding and header padding

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
				if m.ui.ActiveColumn == previewColumn {
					return "170" // Purple when active
				}
				return "214" // Orange when inactive
			}()))
		previewColonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(func() string {
				if m.ui.ActiveColumn == previewColumn {
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
		previewContent.WriteString(previewViewportPadding.Render(m.viewports.Preview.View()))

		// Render the border around the entire preview with same padding as top columns
		s.WriteString("\n")
		previewPaddingStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
		s.WriteString(previewPaddingStyle.Render(previewBorderStyle.Render(previewContent.String())))
	}

	// Render help pane using shared layout
	var helpRows [][]string
	if m.ui.ActiveColumn == searchColumn {
		// Show search syntax help when search is active
		helpRows = [][]string{
			{"tab switch pane", "esc clear+exit search"},
			{"tag:<name>", "type:<type>", "status:archived", "<keyword>", "combine with spaces", "^a toggle archived", "^t cycle type"},
		}
	} else {
		// Show normal navigation help - grouped by function
		if m.ui.ActiveColumn == previewColumn {
			// Preview pane - only show first row
			helpRows = [][]string{
				{"/ search", "tab switch pane", "↑↓ nav", "^s save", "p preview", "M diagram", "S set", "y copy", "esc back", "^c quit"},
			}
		} else if m.ui.ActiveColumn == leftColumn {
			// Available Components pane - no K/J reorder
			helpRows = [][]string{
				// Row 1: System & navigation
				{"/ search", "tab switch pane", "↑↓ nav", "^s save", "p preview", "M diagram", "S set", "y copy", "esc back", "^c quit"},
				// Row 2: Component operations (no K/J reorder)
				{"n new", "e edit", "^x external", "^d delete", "R rename", "C clone", "t tag", "a archive/unarchive", "enter +/-"},
			}
		} else {
			// Pipeline Components pane - includes K/J reorder
			helpRows = [][]string{
				// Row 1: System & navigation
				{"/ search", "tab switch pane", "↑↓ nav", "^s save", "p preview", "M diagram", "S set", "y copy", "esc back", "^c quit"},
				// Row 2: Component operations with K/J reorder
				{"n new", "e edit", "^x external", "^d delete", "R rename", "C clone", "t tag", "a archive/unarchive", "K/J reorder", "enter +/-"},
			}
		}
	}

	helpContent := m.ui.SharedLayout.RenderHelpPane(helpRows)

	// Show confirmation dialogs if active (inline above help)
	if m.ui.DeleteConfirm.Active() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(contentStyle.Render(confirmStyle.Render(m.ui.DeleteConfirm.ViewWithWidth(m.viewports.Width - 4))))
	} else if m.ui.ArchiveConfirm.Active() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
		s.WriteString("\n")
		s.WriteString(contentStyle.Render(confirmStyle.Render(m.ui.ArchiveConfirm.ViewWithWidth(m.viewports.Width - 4))))
	}

	// Always add a single newline before help pane (matching Main list view)
	s.WriteString("\n")
	s.WriteString(helpContent)

	finalView := s.String()

	// Overlay clone dialog if active
	if m.editors.Clone.State != nil && m.editors.Clone.State.Active && m.editors.Clone.Renderer != nil {
		m.editors.Clone.Renderer.SetSize(m.viewports.Width, m.viewports.Height)
		finalView = m.editors.Clone.Renderer.RenderOverlay(finalView, m.editors.Clone.State)
	}

	// Overlay rename dialog if active
	if m.editors.Rename.State != nil && m.editors.Rename.State.Active && m.editors.Rename.Renderer != nil {
		m.editors.Rename.Renderer.SetSize(m.viewports.Width, m.viewports.Height)
		finalView = m.editors.Rename.Renderer.RenderOverlay(finalView, m.editors.Rename.State)
	}

	return finalView
}

func (m *PipelineBuilderModel) SetSize(width, height int) {
	m.viewports.Width = width
	m.viewports.Height = height
	// Update search bar width
	m.search.Bar.SetWidth(width)
	// Update clone renderer size
	if m.editors.Clone.Renderer != nil {
		m.editors.Clone.Renderer.SetSize(width, height)
	}
	// Update rename renderer size
	if m.editors.Rename.Renderer != nil {
		m.editors.Rename.Renderer.SetSize(width, height)
	}
	m.updateViewportSizes()
}

func (m *PipelineBuilderModel) hasUnsavedChanges() bool {
	// For new pipelines, check if components have been added
	if m.data.Pipeline.Path == "" {
		return len(m.data.SelectedComponents) > 0
	}

	// For existing pipelines, check if components have changed
	if len(m.data.SelectedComponents) != len(m.data.OriginalComponents) {
		return true
	}

	// Check if components are the same (order matters)
	for i := range m.data.SelectedComponents {
		if m.data.SelectedComponents[i].Path != m.data.OriginalComponents[i].Path {
			return true
		}
	}

	return false
}

// hasTagChanges - DEPRECATED, to be removed
func (m *PipelineBuilderModel) hasTagChanges() bool {
	// Check if the number of tags has changed
	if len(m.ui.CurrentTags) != len(m.ui.OriginalTags) {
		return true
	}

	// Create maps for efficient comparison
	originalMap := make(map[string]bool)
	for _, tag := range m.ui.OriginalTags {
		originalMap[tag] = true
	}

	// Check if all current tags exist in original
	for _, tag := range m.ui.CurrentTags {
		if !originalMap[tag] {
			return true
		}
	}

	return false
}

func (m *PipelineBuilderModel) updateViewportSizes() {
	// Calculate dimensions
	columnWidth := (m.viewports.Width - 6) / 2                 // Account for gap, padding, and ensure border visibility
	searchBarHeight := 3                             // Height for search bar
	contentHeight := m.viewports.Height - 14 - searchBarHeight // Reserve space for search bar, help pane, status bar, and spacing

	if m.ui.ShowPreview {
		contentHeight = contentHeight / 2
	}

	// Ensure minimum height
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Update left and right viewports for table content
	// Left column: heading (1) + empty (1) + table header (1) + empty (1) = 4 lines
	leftViewportHeight := contentHeight - 4
	if leftViewportHeight < 5 {
		leftViewportHeight = 5
	}

	// Right column: heading (1) + tag row with padding (3) + empty (1) = 5 lines
	rightViewportHeight := contentHeight - 5
	if rightViewportHeight < 5 {
		rightViewportHeight = 5
	}

	// Update left table renderer
	m.viewports.LeftTable.SetSize(columnWidth, contentHeight)

	m.viewports.RightViewport.Width = columnWidth - 4 // Account for borders (2) and padding (2)
	m.viewports.RightViewport.Height = rightViewportHeight

	// Update preview viewport
	if m.ui.ShowPreview {
		previewHeight := m.viewports.Height/2 - 5
		if previewHeight < 5 {
			previewHeight = 5
		}
		m.viewports.Preview.Width = m.viewports.Width - 8 // Account for outer padding (2), borders (2), and inner padding (2) + extra spacing
		m.viewports.Preview.Height = previewHeight
	}
}

func (m *PipelineBuilderModel) SetPipeline(pipeline string) {
	if pipeline != "" {
		// Load existing pipeline for editing
		p, err := files.ReadPipeline(pipeline)
		if err != nil {
			m.err = err
			return
		}
		m.data.Pipeline = p
		m.data.SelectedComponents = p.Components
		m.editors.EditingName = false // Don't show name input when editing
		m.editors.NameInput = p.Name

		// Reorganize components by type to match display
		m.reorganizeComponentsByType()

		// Store original components for change detection AFTER reorganization
		// This ensures the comparison baseline matches the displayed order
		m.data.OriginalComponents = make([]models.ComponentRef, len(m.data.SelectedComponents))
		copy(m.data.OriginalComponents, m.data.SelectedComponents)

		// Update local usage counts to reflect this pipeline's components
		// This ensures the counts show what would happen if we save
		for _, comp := range m.data.SelectedComponents {
			componentPath := strings.TrimPrefix(comp.Path, "../")
			m.updateLocalUsageCount(componentPath, 1)
		}

		// Update preview to show the loaded pipeline
		m.updatePreview()

		// Set viewport content if preview is enabled
		if m.ui.ShowPreview && m.ui.PreviewContent != "" {
			// Preprocess content to handle carriage returns and ensure proper line breaks
			processedContent := strings.ReplaceAll(m.ui.PreviewContent, "\r\r", "\n\n")
			processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
			// Wrap content to viewport width to prevent overflow
			wrappedContent := wordwrap.String(processedContent, m.viewports.Preview.Width)
			m.viewports.Preview.SetContent(wrappedContent)
		}
	}
}

// Helper methods
func (m *PipelineBuilderModel) getAllAvailableComponents() []componentItem {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}

	// Group components by type - use filtered lists when searching
	typeGroups := make(map[string][]componentItem)
	typeGroups[models.ComponentTypeContext] = m.data.FilteredContexts
	typeGroups[models.ComponentTypePrompt] = m.data.FilteredPrompts
	typeGroups[models.ComponentTypeRules] = m.data.FilteredRules

	// Build ordered list based on sections
	var all []componentItem
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			all = append(all, components...)
		}
	}

	return all
}

func (m *PipelineBuilderModel) moveCursor(delta int) {
	if m.ui.ActiveColumn == leftColumn {
		components := m.getAllAvailableComponents()
		m.ui.LeftCursor += delta
		if m.ui.LeftCursor < 0 {
			m.ui.LeftCursor = 0
		}
		if m.ui.LeftCursor >= len(components) {
			m.ui.LeftCursor = len(components) - 1
		}
	} else if m.ui.ActiveColumn == rightColumn {
		m.ui.RightCursor += delta
		if m.ui.RightCursor < 0 {
			m.ui.RightCursor = 0
		}
		if m.ui.RightCursor >= len(m.data.SelectedComponents) {
			m.ui.RightCursor = len(m.data.SelectedComponents) - 1
		}
		// Adjust viewport to ensure cursor is visible after movement
		m.adjustRightViewportScroll()
		// Sync preview scroll when navigating components in right column (Pipeline Components)
		if m.ui.ShowPreview && len(m.data.SelectedComponents) > 0 {
			m.syncPreviewToSelectedComponent()
		}
	}
	// Update preview when cursor moves
	m.updatePreview()
}

func (m *PipelineBuilderModel) addSelectedComponent() {
	components := m.getAllAvailableComponents()
	if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
		selected := components[m.ui.LeftCursor]

		// Check if component is already added
		componentPath := "../" + selected.path
		for i, existing := range m.data.SelectedComponents {
			if existing.Path == componentPath {
				// Component already exists, remove it from the pipeline
				// Update the usage count before removing
				m.updateLocalUsageCount(selected.path, -1)

				// Set cursor to the position being removed so viewport scrolls there
				m.ui.RightCursor = i

				// Remove the component
				m.data.SelectedComponents = append(
					m.data.SelectedComponents[:i],
					m.data.SelectedComponents[i+1:]...,
				)

				// Reorganize to maintain grouping
				m.reorganizeComponentsByType()

				// After reorganization, find where the cursor should be
				// If we removed the last item, move cursor to the new last item
				if m.ui.RightCursor >= len(m.data.SelectedComponents) && len(m.data.SelectedComponents) > 0 {
					m.ui.RightCursor = len(m.data.SelectedComponents) - 1
				} else if len(m.data.SelectedComponents) == 0 {
					m.ui.RightCursor = 0
				}
				// Otherwise keep cursor at the same position to show where removal happened

				// Adjust viewport to ensure cursor is visible
				m.adjustRightViewportScroll()

				// Update preview after removing component
				m.updatePreview()

				return
			}
		}

		// Create component ref with relative path
		ref := models.ComponentRef{
			Type:  selected.compType,
			Path:  componentPath,
			Order: 0, // Will be set when inserting
		}

		// Insert component in the correct position based on type grouping
		m.insertComponentByType(ref)

		// Update the usage count locally to show predicted usage
		// This gives immediate feedback before the pipeline is saved
		m.updateLocalUsageCount(selected.path, 1)

		// Update preview after adding component
		m.updatePreview()
	}
}

// adjustRightViewportScroll ensures the cursor is visible in the right viewport
func (m *PipelineBuilderModel) adjustRightViewportScroll() {
	// Calculate the line number where the cursor is
	currentLine := 0
	overallIndex := 0

	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}

	// Group components by type
	typeGroups := make(map[string][]models.ComponentRef)
	for _, comp := range m.data.SelectedComponents {
		typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
	}

	// Calculate cursor line position
	for sectionIdx, section := range settings.Output.Formatting.Sections {
		components := typeGroups[section.Type]
		if len(components) == 0 {
			continue
		}

		currentLine++ // Section header

		for range components {
			if overallIndex == m.ui.RightCursor {
				goto found
			}
			currentLine++
			overallIndex++
		}

		// Add empty line if there are more sections
		hasMoreSections := false
		for j := sectionIdx + 1; j < len(settings.Output.Formatting.Sections); j++ {
			if len(typeGroups[settings.Output.Formatting.Sections[j].Type]) > 0 {
				hasMoreSections = true
				break
			}
		}
		if hasMoreSections {
			currentLine++ // Empty line between sections
		}
	}

found:
	// Ensure the cursor line is visible
	if currentLine < m.viewports.RightViewport.YOffset {
		m.viewports.RightViewport.SetYOffset(currentLine)
	} else if currentLine >= m.viewports.RightViewport.YOffset+m.viewports.RightViewport.Height {
		m.viewports.RightViewport.SetYOffset(currentLine - m.viewports.RightViewport.Height + 1)
	}
}

// insertComponentByType inserts a component in the correct position based on type grouping
func (m *PipelineBuilderModel) insertComponentByType(newComp models.ComponentRef) {
	// Add the component to the list
	m.data.SelectedComponents = append(m.data.SelectedComponents, newComp)

	// Reorganize to maintain type grouping
	m.reorganizeComponentsByType()

	// Find the position of the newly added component and move cursor there
	for i, comp := range m.data.SelectedComponents {
		if comp.Path == newComp.Path && comp.Type == newComp.Type {
			m.ui.RightCursor = i
			// Immediately adjust viewport to show the cursor
			m.adjustRightViewportScroll()
			break
		}
	}
}

// reorganizeComponentsByType sorts components into groups according to section_order
func (m *PipelineBuilderModel) reorganizeComponentsByType() {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}

	// Group components by type
	typeGroups := make(map[string][]models.ComponentRef)
	for _, comp := range m.data.SelectedComponents {
		typeGroups[comp.Type] = append(typeGroups[comp.Type], comp)
	}

	// Rebuild the array in configured order
	m.data.SelectedComponents = nil
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			m.data.SelectedComponents = append(m.data.SelectedComponents, components...)
		}
	}

	// Update order numbers
	for i := range m.data.SelectedComponents {
		m.data.SelectedComponents[i].Order = i + 1
	}
}

// updateLocalUsageCount updates the usage count for a component locally
func (m *PipelineBuilderModel) updateLocalUsageCount(componentPath string, delta int) {
	// Update in prompts
	for i := range m.data.Prompts {
		if m.data.Prompts[i].path == componentPath {
			m.data.Prompts[i].usageCount += delta
			if m.data.Prompts[i].usageCount < 0 {
				m.data.Prompts[i].usageCount = 0
			}
			return
		}
	}

	// Update in contexts
	for i := range m.data.Contexts {
		if m.data.Contexts[i].path == componentPath {
			m.data.Contexts[i].usageCount += delta
			if m.data.Contexts[i].usageCount < 0 {
				m.data.Contexts[i].usageCount = 0
			}
			return
		}
	}

	// Update in rules
	for i := range m.data.Rules {
		if m.data.Rules[i].path == componentPath {
			m.data.Rules[i].usageCount += delta
			if m.data.Rules[i].usageCount < 0 {
				m.data.Rules[i].usageCount = 0
			}
			return
		}
	}
}

func (m *PipelineBuilderModel) removeSelectedComponent() {
	if m.ui.RightCursor >= 0 && m.ui.RightCursor < len(m.data.SelectedComponents) {
		// Get the component path to update usage count
		removedComponent := m.data.SelectedComponents[m.ui.RightCursor]
		componentPath := strings.TrimPrefix(removedComponent.Path, "../")

		// Remember the type of component we're removing to adjust cursor properly
		removedType := removedComponent.Type

		// Remove component
		m.data.SelectedComponents = append(
			m.data.SelectedComponents[:m.ui.RightCursor],
			m.data.SelectedComponents[m.ui.RightCursor+1:]...,
		)

		// Reorganize to maintain grouping
		m.reorganizeComponentsByType()

		// Adjust cursor - try to stay in the same position or move to the last item of the same type
		if m.ui.RightCursor >= len(m.data.SelectedComponents) && m.ui.RightCursor > 0 {
			m.ui.RightCursor = len(m.data.SelectedComponents) - 1
		}

		// Try to position cursor on a component of the same type
		if len(m.data.SelectedComponents) > 0 {
			// Find the last component of the same type before or at cursor position
			newCursor := -1
			for i := 0; i <= m.ui.RightCursor && i < len(m.data.SelectedComponents); i++ {
				if m.data.SelectedComponents[i].Type == removedType {
					newCursor = i
				}
			}
			if newCursor >= 0 {
				m.ui.RightCursor = newCursor
			}
		}

		// Update the usage count locally
		m.updateLocalUsageCount(componentPath, -1)

		// Adjust viewport to ensure cursor is still visible
		m.adjustRightViewportScroll()

		// Update preview after removing component
		m.updatePreview()
	}
}

func (m *PipelineBuilderModel) moveComponentUp() {
	if m.ui.RightCursor > 0 && m.ui.RightCursor < len(m.data.SelectedComponents) {
		currentType := m.data.SelectedComponents[m.ui.RightCursor].Type
		previousType := m.data.SelectedComponents[m.ui.RightCursor-1].Type

		// Only allow moving within the same type group
		if currentType == previousType {
			// Swap with previous
			m.data.SelectedComponents[m.ui.RightCursor-1], m.data.SelectedComponents[m.ui.RightCursor] =
				m.data.SelectedComponents[m.ui.RightCursor], m.data.SelectedComponents[m.ui.RightCursor-1]

			// Update order numbers
			m.data.SelectedComponents[m.ui.RightCursor-1].Order = m.ui.RightCursor
			m.data.SelectedComponents[m.ui.RightCursor].Order = m.ui.RightCursor + 1

			m.ui.RightCursor--

			// Adjust viewport to ensure cursor is visible after move
			m.adjustRightViewportScroll()
		}
	}
}

func (m *PipelineBuilderModel) moveComponentDown() {
	if m.ui.RightCursor >= 0 && m.ui.RightCursor < len(m.data.SelectedComponents)-1 {
		currentType := m.data.SelectedComponents[m.ui.RightCursor].Type
		nextType := m.data.SelectedComponents[m.ui.RightCursor+1].Type

		// Only allow moving within the same type group
		if currentType == nextType {
			// Swap with next
			m.data.SelectedComponents[m.ui.RightCursor], m.data.SelectedComponents[m.ui.RightCursor+1] =
				m.data.SelectedComponents[m.ui.RightCursor+1], m.data.SelectedComponents[m.ui.RightCursor]

			// Update order numbers
			m.data.SelectedComponents[m.ui.RightCursor].Order = m.ui.RightCursor + 1
			m.data.SelectedComponents[m.ui.RightCursor+1].Order = m.ui.RightCursor + 2

			m.ui.RightCursor++

			// Adjust viewport to ensure cursor is visible after move
			m.adjustRightViewportScroll()
		}
	}
}

func (m *PipelineBuilderModel) updatePreview() {
	if !m.ui.ShowPreview {
		return
	}

	// Show preview based on active column
	if m.ui.ActiveColumn == leftColumn {
		// Show component preview for left column
		components := m.getAllAvailableComponents()
		if len(components) == 0 {
			m.ui.PreviewContent = "No components to preview."
			return
		}

		if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
			comp := components[m.ui.LeftCursor]

			// Read component content
			content, err := files.ReadComponent(comp.path)
			if err != nil {
				m.ui.PreviewContent = fmt.Sprintf("Error loading component: %v", err)
				return
			}

			// Set preview content to just the component content without metadata
			m.ui.PreviewContent = content.Content
		}
	} else {
		// Show pipeline preview for right column
		if len(m.data.SelectedComponents) == 0 {
			m.ui.PreviewContent = "No components selected yet.\n\nAdd components to see the pipeline preview."
			return
		}

		// Create a temporary pipeline with current components
		tempPipeline := &models.Pipeline{
			Name:       m.data.Pipeline.Name,
			Components: m.data.SelectedComponents,
		}

		// Generate the preview
		output, err := composer.ComposePipeline(tempPipeline)
		if err != nil {
			m.ui.PreviewContent = fmt.Sprintf("Error generating preview: %v", err)
			return
		}

		m.ui.PreviewContent = output
	}
}

// findComponentInPreview finds the line number where a component's content appears in the preview
// It returns the line number, or -1 if not found
func (m *PipelineBuilderModel) findComponentInPreview(componentContent, componentPath string) int {
	// Get the first non-empty line of the component content for matching
	componentLines := strings.Split(strings.TrimSpace(componentContent), "\n")
	var firstContentLine string
	for _, line := range componentLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "---") {
			firstContentLine = trimmed
			break
		}
	}

	if firstContentLine == "" {
		return -1
	}

	// Calculate line position in the preview
	lines := strings.Split(m.ui.PreviewContent, "\n")

	// Track which occurrence we're looking for (components might repeat)
	occurrenceCount := 0
	targetOccurrence := 0

	// Count how many times this component appears before our target
	for i := 0; i < m.ui.RightCursor; i++ {
		if m.data.SelectedComponents[i].Path == componentPath {
			targetOccurrence++
		}
	}

	// Search for the component's content in the preview
	for i, line := range lines {
		if strings.Contains(line, firstContentLine) {
			if occurrenceCount == targetOccurrence {
				// Found the right occurrence, return line with context
				targetLine := i - scrollContextLines
				if targetLine < 0 {
					targetLine = 0
				}
				return targetLine
			}
			occurrenceCount++
		}
	}

	return -1
}

// estimateComponentPosition estimates the line position of a component based on its order
func (m *PipelineBuilderModel) estimateComponentPosition() int {
	if m.ui.RightCursor == 0 {
		return 0
	}

	lines := strings.Split(m.ui.PreviewContent, "\n")

	// Calculate average lines per component
	linesPerComponent := 0
	if len(lines) > 0 && len(m.data.SelectedComponents) > 0 {
		linesPerComponent = len(lines) / len(m.data.SelectedComponents)
	}
	if linesPerComponent < minLinesPerComponent {
		linesPerComponent = defaultLinesPerComponent
	}

	targetLine := (m.ui.RightCursor * linesPerComponent) + 5
	if targetLine >= len(lines)-scrollBottomPadding {
		targetLine = len(lines) - scrollBottomPadding
	}
	if targetLine < 0 {
		targetLine = 0
	}

	return targetLine
}

// syncPreviewToSelectedComponent scrolls the preview viewport to show the currently selected component in the pipeline
func (m *PipelineBuilderModel) syncPreviewToSelectedComponent() {
	if !m.ui.ShowPreview || len(m.data.SelectedComponents) == 0 || m.ui.RightCursor < 0 || m.ui.RightCursor >= len(m.data.SelectedComponents) {
		return
	}

	// Get the currently selected component in the right column
	selectedComp := m.data.SelectedComponents[m.ui.RightCursor]

	// Read the component content to match it in the preview
	// Component paths in YAML are relative to the pipelines directory
	componentPath := filepath.Join(files.PipelinesDir, selectedComp.Path)
	componentPath = filepath.Clean(componentPath)

	content, err := files.ReadComponent(componentPath)
	if err != nil {
		return
	}

	// Try to find the component in the preview by its content
	targetLine := m.findComponentInPreview(content.Content, selectedComp.Path)

	// If we couldn't find by content, estimate position by component order
	if targetLine == -1 && m.ui.RightCursor > 0 {
		targetLine = m.estimateComponentPosition()
	}

	// If still not found and we're at position 0, scroll to top
	if targetLine == -1 {
		targetLine = 0
	}

	// Scroll to the target line, centering it if possible
	viewportHeight := m.viewports.Preview.Height
	if targetLine > viewportHeight/2 {
		// Scroll so the target line is centered
		m.viewports.Preview.SetYOffset(targetLine - viewportHeight/2)
	} else {
		// Scroll to top if target is near the beginning
		m.viewports.Preview.SetYOffset(0)
	}
}

// sanitizeFileName converts a user-provided name into a safe filename
func sanitizeFileName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	filename := strings.ToLower(name)
	filename = strings.ReplaceAll(filename, " ", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	var cleanName strings.Builder
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanName.WriteRune(r)
		}
	}

	result := cleanName.String()

	// Ensure the filename is not empty
	if result == "" {
		result = "untitled"
	}

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	return result
}

func (m *PipelineBuilderModel) savePipeline() tea.Cmd {
	return func() tea.Msg {
		// Update pipeline with selected components
		m.data.Pipeline.Components = m.data.SelectedComponents

		// Create filename from name using sanitization
		filename := sanitizeFileName(m.data.Pipeline.Name) + ".yaml"

		// Check if pipeline already exists (case-insensitive)
		existingPipelines, err := files.ListPipelines()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to check existing pipelines: %v", err))
		}

		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, filename) {
				// Don't overwrite if it's not the same pipeline we're editing
				if m.data.Pipeline.Path == "" || !strings.EqualFold(m.data.Pipeline.Path, existing) {
					return StatusMsg(fmt.Sprintf("× Pipeline '%s' already exists. Please choose a different name.", m.data.Pipeline.Name))
				}
			}
		}

		m.data.Pipeline.Path = filename

		// Save pipeline
		err = files.WritePipeline(m.data.Pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save pipeline: %v", err))
		}

		// Update original components to match saved state
		m.data.OriginalComponents = make([]models.ComponentRef, len(m.data.SelectedComponents))
		copy(m.data.OriginalComponents, m.data.SelectedComponents)

		// Reload components to update usage stats after save
		m.loadAvailableComponents()

		// Return success message
		return StatusMsg(fmt.Sprintf("✓ Pipeline saved: %s", m.data.Pipeline.Path))
	}
}

func (m *PipelineBuilderModel) saveAndSetPipeline() tea.Cmd {
	return func() tea.Msg {
		// Update pipeline with selected components
		m.data.Pipeline.Components = m.data.SelectedComponents

		// Create filename from name using sanitization
		filename := sanitizeFileName(m.data.Pipeline.Name) + ".yaml"

		// Check if pipeline already exists (case-insensitive)
		existingPipelines, err := files.ListPipelines()
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to check existing pipelines: %v", err))
		}

		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, filename) {
				// Don't overwrite if it's not the same pipeline we're editing
				if m.data.Pipeline.Path == "" || !strings.EqualFold(m.data.Pipeline.Path, existing) {
					return StatusMsg(fmt.Sprintf("× Pipeline '%s' already exists. Please choose a different name.", m.data.Pipeline.Name))
				}
			}
		}

		m.data.Pipeline.Path = filename

		// Save pipeline
		err = files.WritePipeline(m.data.Pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save pipeline: %v", err))
		}

		// Generate pipeline output
		output, err := composer.ComposePipeline(m.data.Pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to generate output: %v", err))
		}

		// Write to PLUQQY.md
		outputPath := m.data.Pipeline.OutputPath
		if outputPath == "" {
			outputPath = files.DefaultOutputFile
		}

		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to write output: %v", err))
		}

		// Update preview if showing
		if m.ui.ShowPreview {
			m.ui.PreviewContent = output
			// Preprocess content to handle carriage returns and ensure proper line breaks
			processedContent := strings.ReplaceAll(output, "\r\r", "\n\n")
			processedContent = strings.ReplaceAll(processedContent, "\r", "\n")
			// Wrap content to viewport width to prevent overflow
			wrappedContent := wordwrap.String(processedContent, m.viewports.Preview.Width)
			m.viewports.Preview.SetContent(wrappedContent)
		}

		// Update original components to match saved state
		m.data.OriginalComponents = make([]models.ComponentRef, len(m.data.SelectedComponents))
		copy(m.data.OriginalComponents, m.data.SelectedComponents)

		// Reload components to update usage stats after save
		m.loadAvailableComponents()

		// Return success message
		return StatusMsg(fmt.Sprintf("✓ Saved & Set → %s", outputPath))
	}
}

func (m *PipelineBuilderModel) deletePipeline() tea.Cmd {
	return func() tea.Msg {
		if m.data.Pipeline == nil || m.data.Pipeline.Path == "" {
			return StatusMsg("× No pipeline to delete")
		}

		// Store tags before deletion for cleanup
		tagsToCleanup := make([]string, len(m.data.Pipeline.Tags))
		copy(tagsToCleanup, m.data.Pipeline.Tags)

		// Delete the pipeline file
		err := files.DeletePipeline(m.data.Pipeline.Path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to delete pipeline: %v", err))
		}

		// Extract pipeline name from path
		pipelineName := strings.TrimSuffix(filepath.Base(m.data.Pipeline.Path), ".yaml")

		// Start async tag cleanup if there were tags
		if len(tagsToCleanup) > 0 {
			// This will run asynchronously after we return
			go func() {
				tags.CleanupOrphanedTags(tagsToCleanup)
			}()
		}

		// Return to the list view with success message
		return SwitchViewMsg{
			view:   mainListView,
			status: fmt.Sprintf("✓ Deleted pipeline: %s", pipelineName),
		}
	}
}

func (m *PipelineBuilderModel) deleteComponentFromLeft(comp componentItem) tea.Cmd {
	return func() tea.Msg {
		// Store tags before deletion for cleanup
		tagsToCleanup := make([]string, len(comp.tags))
		copy(tagsToCleanup, comp.tags)

		// Delete the component file
		err := files.DeleteComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to delete %s: %v", comp.compType, err))
		}

		// Start async tag cleanup if there were tags
		if len(tagsToCleanup) > 0 {
			go func() {
				tags.CleanupOrphanedTags(tagsToCleanup)
			}()
		}

		// Reload components to refresh the list
		m.loadAvailableComponents()

		// Return success message
		return StatusMsg(fmt.Sprintf("✓ Deleted %s: %s", comp.compType, comp.name))
	}
}

func (m *PipelineBuilderModel) nameInputView() string {
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
	contentWidth := m.viewports.Width - 4    // Match help pane width
	contentHeight := m.viewports.Height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")) // Purple for active single pane

	heading := "CREATE NEW PIPELINE"
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Name input prompt - centered
	promptText := promptStyle.Render("Enter a descriptive name for your pipeline:")
	centeredPromptStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(promptText)))
	mainContent.WriteString("\n\n")

	// Input field with cursor
	input := m.editors.NameInput + "│" // cursor

	// Render input field with padding for centering
	inputFieldContent := inputStyle.Render(input)

	// Add padding to center the input field properly
	centeredInputStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)

	mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(inputFieldContent)))

	// Check if pipeline name already exists and show warning
	if m.editors.NameInput != "" {
		testFilename := sanitizeFileName(m.editors.NameInput) + ".yaml"
		existingPipelines, _ := files.ListPipelines()
		for _, existing := range existingPipelines {
			if strings.EqualFold(existing, testFilename) {
				// Show warning
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // Orange/yellow for warning
					Bold(true)
				warningText := warningStyle.Render(fmt.Sprintf("⚠ Warning: Pipeline '%s' already exists", m.editors.NameInput))
				mainContent.WriteString("\n\n")
				mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(warningText)))
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
		"esc cancel",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.viewports.Width-4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.viewports.Width - 8).
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

func (m *PipelineBuilderModel) componentCreationView() string {
	renderer := NewComponentCreationViewRenderer(m.viewports.Width, m.viewports.Height)

	switch m.editors.ComponentCreator.GetCurrentStep() {
	case 0:
		return renderer.RenderTypeSelection(m.editors.ComponentCreator.GetTypeCursor())
	case 1:
		return renderer.RenderNameInput(m.editors.ComponentCreator.GetComponentType(), m.editors.ComponentCreator.GetComponentName())
	case 2:
		// Use enhanced editor if available
		if m.editors.ComponentCreator.IsEnhancedEditorActive() {
			if adapter, ok := m.editors.ComponentCreator.GetEnhancedEditor().(*shared.EnhancedEditorAdapter); ok {
				if underlyingEditor, ok := adapter.GetUnderlyingEditor().(*EnhancedEditorState); ok {
					return renderer.RenderWithEnhancedEditor(
						underlyingEditor,
						m.editors.ComponentCreator.GetComponentType(),
						m.editors.ComponentCreator.GetComponentName(),
					)
				}
			}
		}
		// Fallback to simple editor
		return renderer.RenderContentEdit(m.editors.ComponentCreator.GetComponentType(), m.editors.ComponentCreator.GetComponentName(), m.editors.ComponentCreator.GetComponentContent())
	}

	return "Unknown creation step"
}

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

// handleTagEditing handles tag editing mode.
// Note: Vim-style navigation (h,j,k,l) is intentionally disabled when typing tag names
// to allow these letters to be entered as part of tag text. This matches the behavior
// of the Main List View's tag editor. Arrow keys are used for navigation instead.
func (m *PipelineBuilderModel) handleTagEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ui.TagDeleteConfirm.Active() {
		return m, m.ui.TagDeleteConfirm.Update(msg)
	}

	switch msg.String() {
	case "esc":
		// Check for unsaved changes
		if m.hasTagChanges() {
			// Show confirmation dialog
			m.ui.ExitConfirmationType = "tags"
			m.ui.ExitConfirm.ShowDialog(
				"⚠️  Unsaved Changes",
				"You have unsaved changes to tags.",
				"Exit without saving?",
				true, // destructive
				m.viewports.Width-4,   // width
				10,   // height
				func() tea.Cmd {
					// Exit without saving
					m.ui.EditingTags = false
					m.ui.TagInput = ""
					m.ui.CurrentTags = nil
					m.ui.OriginalTags = nil
					m.ui.ShowTagSuggestions = false
					m.ui.TagCloudActive = false
					m.ui.TagCloudCursor = 0
					return nil
				},
				func() tea.Cmd {
					// Stay in tag editing mode
					return nil
				},
			)
			return m, nil
		} else {
			// No changes, exit directly
			m.ui.EditingTags = false
			m.ui.TagInput = ""
			m.ui.CurrentTags = nil
			m.ui.OriginalTags = nil
			m.ui.ShowTagSuggestions = false
			m.ui.TagCloudActive = false
			m.ui.TagCloudCursor = 0
			return m, nil
		}

	case "ctrl+s":
		return m, m.saveTags()

	case "enter":
		if m.ui.TagCloudActive {
			availableForSelection := m.getAvailableTagsForCloud()
			if m.ui.TagCloudCursor >= 0 && m.ui.TagCloudCursor < len(availableForSelection) {
				tag := availableForSelection[m.ui.TagCloudCursor]
				if !m.hasTag(tag) {
					m.ui.CurrentTags = append(m.ui.CurrentTags, tag)
				}
			}
		} else if m.ui.TagInput != "" {
			// Only use suggestion if user actively navigated to select it
			if m.ui.HasNavigatedSuggestions && m.ui.ShowTagSuggestions {
				suggestions := m.getFilteredSuggestions()
				if m.ui.TagSuggestionCursor >= 0 && m.ui.TagSuggestionCursor < len(suggestions) {
					// Limit to 5 suggestions as per the view
					maxIndex := len(suggestions)
					if maxIndex > 5 {
						maxIndex = 5
					}
					if m.ui.TagSuggestionCursor < maxIndex {
						tag := suggestions[m.ui.TagSuggestionCursor]
						if !m.hasTag(tag) {
							m.ui.CurrentTags = append(m.ui.CurrentTags, tag)
						}
						m.ui.TagInput = ""
						m.ui.TagCursor = len(m.ui.CurrentTags)
						m.ui.TagSuggestionCursor = 0
						m.ui.HasNavigatedSuggestions = false
					}
				}
			} else {
				// User just typed and hit enter - add exactly what they typed
				normalized := models.NormalizeTagName(m.ui.TagInput)
				if normalized != "" && !m.hasTag(normalized) {
					m.ui.CurrentTags = append(m.ui.CurrentTags, normalized)
				}
				m.ui.TagInput = ""
				m.ui.TagCursor = len(m.ui.CurrentTags)
				m.ui.TagSuggestionCursor = 0
				m.ui.HasNavigatedSuggestions = false
			}
		}
		return m, nil

	case "tab":
		m.ui.TagCloudActive = !m.ui.TagCloudActive
		if m.ui.TagCloudActive {
			m.ui.TagCloudCursor = 0
		}
		return m, nil

	case "up":
		if !m.ui.TagCloudActive && m.ui.ShowTagSuggestions && m.ui.TagInput != "" {
			// Navigate up in suggestions
			if m.ui.TagSuggestionCursor > 0 {
				m.ui.TagSuggestionCursor--
			}
			// Mark as navigated even if cursor doesn't move (for single suggestion case)
			m.ui.HasNavigatedSuggestions = true
		} else if m.ui.TagCloudActive {
			// Navigate in tag cloud
			if m.ui.TagCloudCursor > 0 {
				m.ui.TagCloudCursor--
			}
		}
		return m, nil

	case "down":
		if !m.ui.TagCloudActive && m.ui.ShowTagSuggestions && m.ui.TagInput != "" {
			// Navigate down in suggestions
			suggestions := m.getFilteredSuggestions()
			maxSuggestions := len(suggestions)
			if maxSuggestions > 5 {
				maxSuggestions = 5 // Limit to 5 suggestions as per the view
			}
			if m.ui.TagSuggestionCursor < maxSuggestions-1 {
				m.ui.TagSuggestionCursor++
			}
			// Mark as navigated even if cursor doesn't move (for single suggestion case)
			m.ui.HasNavigatedSuggestions = true
		} else if m.ui.TagCloudActive {
			// Navigate in tag cloud
			availableForSelection := m.getAvailableTagsForCloud()
			if m.ui.TagCloudCursor < len(availableForSelection)-1 {
				m.ui.TagCloudCursor++
			}
		}
		return m, nil

	case "left":
		if m.ui.TagCloudActive {
			// Navigate in tag cloud
			if m.ui.TagCloudCursor > 0 {
				m.ui.TagCloudCursor--
			}
		} else {
			// Move cursor left in current tags
			if m.ui.TagInput == "" && m.ui.TagCursor > 0 {
				m.ui.TagCursor--
			}
		}
		return m, nil

	case "right":
		if m.ui.TagCloudActive {
			// Navigate in tag cloud
			availableForSelection := m.getAvailableTagsForCloud()
			if m.ui.TagCloudCursor < len(availableForSelection)-1 {
				m.ui.TagCloudCursor++
			}
		} else {
			// Move cursor right in current tags
			if m.ui.TagInput == "" && m.ui.TagCursor < len(m.ui.CurrentTags)-1 {
				m.ui.TagCursor++
			}
		}
		return m, nil

	case "backspace":
		if !m.ui.TagCloudActive {
			if m.ui.TagInput != "" {
				if len(m.ui.TagInput) > 0 {
					m.ui.TagInput = m.ui.TagInput[:len(m.ui.TagInput)-1]
					m.ui.ShowTagSuggestions = len(m.ui.TagInput) > 0
					m.ui.TagSuggestionCursor = 0
					m.ui.HasNavigatedSuggestions = false
				}
			} else if m.ui.TagCursor > 0 && m.ui.TagCursor <= len(m.ui.CurrentTags) {
				m.ui.CurrentTags = append(m.ui.CurrentTags[:m.ui.TagCursor-1], m.ui.CurrentTags[m.ui.TagCursor:]...)
				m.ui.TagCursor--
			}
		}
		return m, nil

	case "ctrl+d":
		if m.ui.TagCloudActive {
			// In tag cloud: delete from registry (project-wide)
			availableForSelection := m.getAvailableTagsForCloud()
			if m.ui.TagCloudCursor >= 0 && m.ui.TagCloudCursor < len(availableForSelection) {
				m.ui.DeletingTag = availableForSelection[m.ui.TagCloudCursor]
				usage := m.getTagUsage(m.ui.DeletingTag)
				m.ui.DeletingTagUsage = usage

				// Show confirmation dialog for tag deletion
				var details []string
				if len(usage) > 0 {
					for _, u := range usage {
						details = append(details, u)
					}
				}

				m.ui.TagDeleteConfirm.Show(ConfirmationConfig{
					Title:   "Delete Tag from Registry?",
					Message: fmt.Sprintf("Tag: %s", m.ui.DeletingTag),
					Warning: func() string {
						if len(usage) > 0 {
							return "This tag is currently in use by:"
						}
						return "This tag is not currently in use."
					}(),
					Details:     details,
					Destructive: true,
					Type:        ConfirmTypeDialog,
					Width:       m.viewports.Width - 4,
					Height:      15,
				}, func() tea.Cmd {
					return m.deleteTagFromRegistry()
				}, func() tea.Cmd {
					m.ui.DeletingTag = ""
					m.ui.DeletingTagUsage = nil
					return nil
				})
			}
		} else if m.ui.TagInput == "" && m.ui.TagCursor < len(m.ui.CurrentTags) {
			// In current tags: just remove from component
			if m.ui.TagCursor < len(m.ui.CurrentTags) {
				m.ui.CurrentTags = append(m.ui.CurrentTags[:m.ui.TagCursor], m.ui.CurrentTags[m.ui.TagCursor+1:]...)
				if m.ui.TagCursor >= len(m.ui.CurrentTags) && m.ui.TagCursor > 0 {
					m.ui.TagCursor--
				}
			}
		}
		return m, nil

	case " ":
		if !m.ui.TagCloudActive {
			m.ui.TagInput += " "
			m.ui.ShowTagSuggestions = m.ui.TagInput != ""
			m.ui.TagSuggestionCursor = 0
			m.ui.HasNavigatedSuggestions = false
		}
		return m, nil

	default:
		// Add to input only when in main pane (matching Main List View logic)
		if !m.ui.TagCloudActive && len(msg.String()) == 1 {
			m.ui.TagInput += msg.String()
			m.ui.ShowTagSuggestions = m.ui.TagInput != ""
			m.ui.TagSuggestionCursor = 0
			m.ui.HasNavigatedSuggestions = false
		}
	}

	return m, nil
}

func (m *PipelineBuilderModel) getAvailableTagsForCloud() []string {
	available := []string{}
	for _, tag := range m.ui.AvailableTags {
		if !m.hasTag(tag) {
			available = append(available, tag)
		}
	}
	return available
}

func (m *PipelineBuilderModel) getFilteredSuggestions() []string {
	if m.ui.TagInput == "" {
		return []string{}
	}
	
	input := strings.ToLower(m.ui.TagInput)
	var suggestions []string
	
	// First, exact prefix matches
	for _, tag := range m.ui.AvailableTags {
		if strings.HasPrefix(strings.ToLower(tag), input) && !m.hasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	// Then, contains matches
	for _, tag := range m.ui.AvailableTags {
		lowerTag := strings.ToLower(tag)
		if !strings.HasPrefix(lowerTag, input) && strings.Contains(lowerTag, input) && !m.hasTag(tag) {
			suggestions = append(suggestions, tag)
		}
	}
	
	return suggestions
}

func (m *PipelineBuilderModel) getTagUsage(tag string) []string {
	var usage []string
	allComponents := m.getAllAvailableComponents()
	for _, comp := range allComponents {
		for _, t := range comp.tags {
			if strings.EqualFold(t, tag) {
				usage = append(usage, fmt.Sprintf("%s: %s", comp.compType, comp.name))
				break
			}
		}
	}
	return usage
}

func (m *PipelineBuilderModel) deleteTagFromRegistry() tea.Cmd {
	return func() tea.Msg {
		registry, err := tags.NewRegistry()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to load tag registry: %v", err))
		}

		if err := registry.RemoveTag(m.ui.DeletingTag); err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete tag: %v", err))
		}

		if err := registry.Save(); err != nil {
			return StatusMsg(fmt.Sprintf("Failed to save tag registry: %v", err))
		}

		m.loadAvailableTags()

		newAvailable := []string{}
		for _, tag := range m.ui.CurrentTags {
			if tag != m.ui.DeletingTag {
				newAvailable = append(newAvailable, tag)
			}
		}
		m.ui.CurrentTags = newAvailable

		if m.ui.TagCursor > len(m.ui.CurrentTags) {
			m.ui.TagCursor = len(m.ui.CurrentTags)
		}

		return StatusMsg(fmt.Sprintf("✓ Deleted tag: %s", m.ui.DeletingTag))
	}
}

func (m *PipelineBuilderModel) startTagEditing(path string, currentTags []string) {
	// Get the component name for display
	components := m.getAllAvailableComponents()
	itemName := ""
	for _, comp := range components {
		if comp.path == path {
			itemName = comp.name
			break
		}
	}
	
	// Start the tag editor
	if m.editors.TagEditor == nil {
		m.editors.TagEditor = NewTagEditor()
	}
	m.editors.TagEditor.Start(path, currentTags, "component", itemName)
	m.editors.TagEditor.SetSize(m.viewports.Width, m.viewports.Height)
}

func (m *PipelineBuilderModel) startPipelineTagEditing(currentTags []string) {
	// For pipeline tags, use the pipeline name
	if m.editors.TagEditor == nil {
		m.editors.TagEditor = NewTagEditor()
	}
	m.editors.TagEditor.Start(m.data.Pipeline.Path, currentTags, "pipeline", m.data.Pipeline.Name)
	m.editors.TagEditor.SetSize(m.viewports.Width, m.viewports.Height)
}

func (m *PipelineBuilderModel) loadAvailableTags() {
	// Get all tags from registry
	registry, err := tags.NewRegistry()
	if err != nil {
		m.ui.AvailableTags = []string{}
		return
	}

	allTags := registry.ListTags()
	m.ui.AvailableTags = make([]string, 0, len(allTags))
	for _, tag := range allTags {
		m.ui.AvailableTags = append(m.ui.AvailableTags, tag.Name)
	}

	// Also add tags that exist in components but not in registry
	seenTags := make(map[string]bool)
	for _, tag := range m.ui.AvailableTags {
		seenTags[tag] = true
	}

	// Add tags from all components
	allComponents := m.getAllAvailableComponents()
	for _, comp := range allComponents {
		for _, tag := range comp.tags {
			normalized := models.NormalizeTagName(tag)
			if !seenTags[normalized] {
				m.ui.AvailableTags = append(m.ui.AvailableTags, normalized)
				seenTags[normalized] = true
			}
		}
	}
}

func (m *PipelineBuilderModel) hasTag(tag string) bool {
	for _, t := range m.ui.CurrentTags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

func (m *PipelineBuilderModel) saveTags() tea.Cmd {
	return func() tea.Msg {
		var err error

		// Check for removed tags before saving
		removedTags := []string{}
		for _, oldTag := range m.ui.OriginalTags {
			found := false
			for _, newTag := range m.ui.CurrentTags {
				if oldTag == newTag {
					found = true
					break
				}
			}
			if !found {
				removedTags = append(removedTags, oldTag)
			}
		}

		if m.ui.EditingTagsPath == "" {
			// Editing pipeline tags
			if m.data.Pipeline != nil {
				m.data.Pipeline.Tags = make([]string, len(m.ui.CurrentTags))
				copy(m.data.Pipeline.Tags, m.ui.CurrentTags)
				// Save the pipeline to persist the tags immediately
				if m.data.Pipeline.Path != "" {
					err = files.WritePipeline(m.data.Pipeline)
					if err != nil {
						return StatusMsg(fmt.Sprintf("Failed to save pipeline tags: %v", err))
					}
				}
			}
		} else {
			// Editing component tags
			err = files.UpdateComponentTags(m.ui.EditingTagsPath, m.ui.CurrentTags)
			if err != nil {
				return StatusMsg(fmt.Sprintf("Failed to save tags: %v", err))
			}
		}
		
		// Cleanup orphaned tags asynchronously
		if len(removedTags) > 0 {
			go func() {
				tags.CleanupOrphanedTags(removedTags)
			}()
		}

		// Exit tag editing mode
		m.ui.EditingTags = false
		m.ui.TagInput = ""
		m.ui.CurrentTags = nil
		m.ui.OriginalTags = nil
		m.ui.ShowTagSuggestions = false
		m.ui.TagCloudActive = false
		m.ui.TagCloudCursor = 0

		// Reload components to reflect the changes
		m.loadAvailableComponents()
		
		// Also reload available tags for tag editor
		if m.editors.TagEditor != nil {
			m.editors.TagEditor.LoadAvailableTags()
		}

		// Update filtered components if search is active
		if m.search.Query != "" {
			m.performSearch()
		}
		
		// Update preview to reflect changes
		m.updatePreview()

		return StatusMsg("✓ Tags saved")
	}
}

func (m *PipelineBuilderModel) tagEditView() string {
	inputStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170")).Padding(0, 1).Width(40)
	suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedSuggestionStyle := lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("170"))

	// Calculate dimensions for side-by-side layout
	paneWidth := (m.viewports.Width - 6) / 2 // Same calculation as main list view
	paneHeight := m.viewports.Height - 10    // Leave room for help pane
	headerPadding := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

	var mainContent strings.Builder

	if m.ui.TagDeleteConfirm.Active() {
		return m.ui.TagDeleteConfirm.View()
	}

	components := m.getAllAvailableComponents()
	itemName := ""
	if m.ui.LeftCursor >= 0 && m.ui.LeftCursor < len(components) {
		itemName = components[m.ui.LeftCursor].name
	}

	heading := fmt.Sprintf("EDIT TAGS: %s", strings.ToUpper(itemName))
	remainingWidth := paneWidth - len(heading) - 7
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	// Dynamic styles based on which pane is active
	mainHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if !m.ui.TagCloudActive {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	mainColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if !m.ui.TagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	mainContent.WriteString(headerPadding.Render(mainHeaderStyle.Render(heading) + " " + mainColonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	mainContent.WriteString(headerPadding.Render("Current tags:"))
	mainContent.WriteString("\n\n")
	if len(m.ui.CurrentTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		mainContent.WriteString(headerPadding.Render(dimStyle.Render("(no tags)")))
	} else {
		var rows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		// Calculate max width for tags (account for padding and border)
		maxRowWidth := m.viewports.Width/2 - 6 // Half width minus padding
		
		for i, tag := range m.ui.CurrentTags {
			registry, _ := tags.NewRegistry()
			color := models.GetTagColor(tag, "")
			if registry != nil {
				if t, exists := registry.GetTag(tag); exists && t.Color != "" {
					color = t.Color
				}
			}

			style := lipgloss.NewStyle().Background(lipgloss.Color(color)).Foreground(lipgloss.Color("255")).Padding(0, 1)

			// Build tag display with selection indicators
			var tagDisplay string
			if i == m.ui.TagCursor && m.ui.TagInput == "" {
				// Selected tag with triangle indicators
				indicatorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")). // Bright green for indicators
					Bold(true)
				tagDisplay = indicatorStyle.Render("▶ ") + 
					style.Render(tag) + 
					indicatorStyle.Render(" ◀")
			} else {
				// Regular tag with spacers for consistent width
				tagDisplay = "  " + style.Render(tag) + "  "
			}

			// Calculate actual display width
			tagWidth := lipgloss.Width(tagDisplay) + 2 // Add spacing

			// Check if we need a new row
			if rowTags > 0 && currentRowWidth + tagWidth > maxRowWidth {
				rows.WriteString("\n\n") // Double newline for vertical spacing between rows
				rowTags = 0
				currentRowWidth = 0
			}

			rows.WriteString(tagDisplay)
			
			// Add space between tags (but not at end of row)
			if i < len(m.ui.CurrentTags)-1 {
				rows.WriteString("  ")
			}
			
			currentRowWidth += tagWidth + 2
			rowTags++
		}
		mainContent.WriteString(headerPadding.Render(rows.String()))
	}
	mainContent.WriteString("\n\n")

	mainContent.WriteString(headerPadding.Render("Add new tag:"))
	mainContent.WriteString("\n")

	// Create input display with cursor
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	inputDisplay := m.ui.TagInput
	if !m.ui.TagCloudActive && m.ui.TagInput != "" {
		// Add cursor to existing input when active
		inputDisplay = m.ui.TagInput + cursorStyle.Render("│")
	}

	// Show placeholder if empty
	if m.ui.TagInput == "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		if !m.ui.TagCloudActive {
			inputDisplay = placeholderStyle.Render("Type to add a new or existing tag...") + cursorStyle.Render("│")
		} else {
			inputDisplay = placeholderStyle.Render("Type to add a new or existing tag...")
		}
	}

	// Highlight input border when active
	activeInputStyle := inputStyle
	if !m.ui.TagCloudActive {
		activeInputStyle = inputStyle.BorderForeground(lipgloss.Color("170"))
	}

	mainContent.WriteString(headerPadding.Render(activeInputStyle.Render(inputDisplay)))

	if m.ui.ShowTagSuggestions && len(m.ui.AvailableTags) > 0 {
		mainContent.WriteString("\n\n")
		mainContent.WriteString(headerPadding.Render("Suggestions:\n"))

		suggestions := m.getFilteredSuggestions()

		if len(suggestions) > 0 {
			maxSuggestions := 5
			if len(suggestions) < maxSuggestions {
				maxSuggestions = len(suggestions)
			}

			// Style for arrow indicators
			indicatorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)
			
			for i := 0; i < maxSuggestions; i++ {
				var suggestionLine string
				if i == m.ui.TagSuggestionCursor {
					if m.ui.HasNavigatedSuggestions {
						// Actively selected - show with arrow indicators
						suggestionLine = indicatorStyle.Render("▶ ") + 
							selectedSuggestionStyle.Render(suggestions[i]) +
							indicatorStyle.Render(" ◀")
					} else {
						// Just highlighted (default position) - no arrows
						suggestionLine = "  " + selectedSuggestionStyle.Render(suggestions[i]) + "  "
					}
				} else {
					// Regular suggestion - maintain spacing for alignment
					suggestionLine = "  " + suggestionStyle.Render(suggestions[i]) + "  "
				}
				mainContent.WriteString(headerPadding.Render(suggestionLine))
				mainContent.WriteString("\n")
			}
		}
	}

	// Build right pane content
	var rightContent strings.Builder

	// Available tags header
	availableTagsTitle := "AVAILABLE TAGS"
	availableTagsRemainingWidth := paneWidth - len(availableTagsTitle) - 7 // Adjust for smaller width
	if availableTagsRemainingWidth < 0 {
		availableTagsRemainingWidth = 0
	}
	// Dynamic styles based on which pane is active
	tagCloudHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(func() string {
			if m.ui.TagCloudActive {
				return "170" // Purple when active
			}
			return "214" // Orange when inactive
		}()))
	tagCloudColonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(func() string {
			if m.ui.TagCloudActive {
				return "170" // Purple when active
			}
			return "240" // Gray when inactive
		}()))
	rightContent.WriteString(headerPadding.Render(tagCloudHeaderStyle.Render(availableTagsTitle) + " " + tagCloudColonStyle.Render(strings.Repeat(":", availableTagsRemainingWidth))))
	rightContent.WriteString("\n\n")

	// Always display available tags
	if len(m.ui.AvailableTags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		rightContent.WriteString(headerPadding.Render(dimStyle.Render("(no available tags)")))
	} else {
		// Group tags in rows for better display
		var tagRows strings.Builder
		rowTags := 0
		currentRowWidth := 0
		maxRowWidth := paneWidth - 6 // Account for padding

		// Get available tags that haven't been added yet
		var availableForCloud []string
		for _, tag := range m.ui.AvailableTags {
			if !m.hasTag(tag) {
				availableForCloud = append(availableForCloud, tag)
			}
		}

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
			if m.ui.TagCloudActive && i == m.ui.TagCloudCursor {
				indicatorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")).
					Bold(true)
				tagDisplay = indicatorStyle.Render("▶ ") + tagStyle.Render(tag) + indicatorStyle.Render(" ◀")
			} else {
				// Add invisible spacers to maintain consistent width
				tagDisplay = "  " + tagStyle.Render(tag) + "  "
			}

			tagWidth := lipgloss.Width(tagDisplay) + 2 // Add spacing

			// Check if we need to start a new row
			if rowTags > 0 && currentRowWidth+tagWidth+2 > maxRowWidth {
				tagRows.WriteString("\n\n") // Double newline for vertical spacing
				rowTags = 0
				currentRowWidth = 0
			}

			if rowTags > 0 {
				tagRows.WriteString("  ")
			}
			tagRows.WriteString(tagDisplay)
			currentRowWidth += tagWidth + 2
			rowTags++
		}

		rightContent.WriteString(headerPadding.Render(tagRows.String()))
	}

	// Create border styles
	activeBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	inactiveBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	leftPaneBorder := inactiveBorderStyle
	if !m.ui.TagCloudActive {
		leftPaneBorder = activeBorderStyle
	}

	leftPane := leftPaneBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(mainContent.String())

	rightPaneBorder := inactiveBorderStyle
	if m.ui.TagCloudActive {
		rightPaneBorder = activeBorderStyle
	}

	rightPane := rightPaneBorder.
		Width(paneWidth).
		Height(paneHeight).
		Render(rightContent.String())

	// Join panes side by side
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		" ", // Single space gap, same as main list view
		rightPane,
	)

	// Help section - match Main List View exactly
	var help []string
	if m.ui.TagCloudActive {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←→ navigate",
			"^d delete tag",
			"^s save",
			"esc cancel",
		}
	} else {
		help = []string{
			"tab switch pane",
			"enter add tag",
			"←↑↓→ navigate",
			"^d remove tag",
			"^t reload tags",
			"^s save",
			"esc cancel",
		}
	}
	helpBorderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Width(m.viewports.Width-4).Padding(0, 1)
	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(m.viewports.Width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	helpContent = alignedHelp

	var s strings.Builder
	contentStyle := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

	s.WriteString(contentStyle.Render(mainView))
	s.WriteString("\n")
	s.WriteString(contentStyle.Render(helpBorderStyle.Render(helpContent)))

	return s.String()
}

// archiveComponent archives a component
func (m *PipelineBuilderModel) archiveComponent(comp componentItem) tea.Cmd {
	return func() tea.Msg {
		err := files.ArchiveComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive %s '%s': %v", comp.compType, comp.name, err))
		}

		// Reload components to reflect the change
		m.loadAvailableComponents()

		// Refresh search if active
		if m.search.Query != "" {
			m.performSearch()
		}

		return StatusMsg(fmt.Sprintf("✓ Archived %s: %s", comp.compType, comp.name))
	}
}

// unarchiveComponent unarchives a component
func (m *PipelineBuilderModel) unarchiveComponent(comp componentItem) tea.Cmd {
	return func() tea.Msg {
		err := files.UnarchiveComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to unarchive %s '%s': %v", comp.compType, comp.name, err))
		}

		// Reload components to reflect the change
		m.loadAvailableComponents()

		// Refresh search if active
		if m.search.Query != "" {
			m.performSearch()
		}

		return StatusMsg(fmt.Sprintf("✓ Unarchived %s: %s", comp.compType, comp.name))
	}
}

// archivePipeline archives the pipeline being edited
func (m *PipelineBuilderModel) archivePipeline() tea.Cmd {
	return func() tea.Msg {
		if m.data.Pipeline == nil || m.data.Pipeline.Path == "" {
			return StatusMsg("× No pipeline to archive")
		}

		err := files.ArchivePipeline(m.data.Pipeline.Path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive pipeline: %v", err))
		}

		// Extract pipeline name from path
		pipelineName := strings.TrimSuffix(filepath.Base(m.data.Pipeline.Path), ".yaml")

		// Return to main list after archiving
		return SwitchViewMsg{
			view:   mainListView,
			status: fmt.Sprintf("✓ Archived pipeline: %s", pipelineName),
		}
	}
}
