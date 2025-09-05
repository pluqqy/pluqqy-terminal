package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tui/shared"
	"github.com/pluqqy/pluqqy-terminal/pkg/utils"
)

// Main View Method

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
				{"n new", "e edit", "^x external", "^d delete", "R rename", "C clone", "t tag", "u usage", "a archive/unarchive", "enter +/-"},
			}
		} else {
			// Pipeline Components pane - includes K/J reorder
			helpRows = [][]string{
				// Row 1: System & navigation
				{"/ search", "tab switch pane", "↑↓ nav", "^s save", "p preview", "M diagram", "S set", "y copy", "esc back", "^c quit"},
				// Row 2: Component operations with K/J reorder
				{"n new", "e edit", "^x external", "^d delete", "R rename", "C clone", "t tag", "u usage", "a archive/unarchive", "K/J reorder", "enter +/-"},
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

	// Overlay component usage modal if active
	if m.editors.ComponentUsage != nil && m.editors.ComponentUsage.Active {
		renderer := NewComponentUsageRenderer(m.viewports.Width, m.viewports.Height)
		overlay := renderer.Render(m.editors.ComponentUsage)
		if overlay != "" {
			finalView = overlayViews(finalView, overlay)
		}
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
	// Update component usage size
	if m.editors.ComponentUsage != nil {
		m.editors.ComponentUsage.SetSize(width, height)
	}
	m.updateViewportSizes()
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

// Preview Methods

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

// Component Creation View

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

// Name Input View

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