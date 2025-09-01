package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/composer"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search/unified"
	"github.com/pluqqy/pluqqy-cli/pkg/tags"
	"github.com/pluqqy/pluqqy-cli/pkg/tui/shared"
)

// Component Management Methods

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

// Pipeline Operations

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

// Component Operations

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

// Tag Management

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



// State Management

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

// Helper Functions

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