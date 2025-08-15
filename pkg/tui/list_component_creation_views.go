package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

// ComponentCreationViewRenderer handles rendering of component creation views
type ComponentCreationViewRenderer struct {
	width  int
	height int
}

// NewComponentCreationViewRenderer creates a new renderer instance
func NewComponentCreationViewRenderer(width, height int) *ComponentCreationViewRenderer {
	return &ComponentCreationViewRenderer{
		width:  width,
		height: height,
	}
}

// RenderTypeSelection renders the component type selection view
func (r *ComponentCreationViewRenderer) RenderTypeSelection(typeCursor int) string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	selectedStyle := SelectedStyle.Padding(0, 1)
	normalStyle := NormalStyle.Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Calculate dimensions
	contentWidth := r.width - 4 // Match help pane width
	contentHeight := r.height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true) // Purple for active single pane

	heading := "CREATE NEW COMPONENT"
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Component type selection
	contentPadding := headerPadding
	mainContent.WriteString(contentPadding.Render("Select component type:"))
	mainContent.WriteString("\n\n")

	types := []struct {
		name string
		desc string
	}{
		{"CONTEXT", "Background information or system state"},
		{"PROMPT", "Instructions or questions for the LLM"},
		{"RULES", "Important constraints or guidelines"},
	}

	for i, t := range types {
		cursor := "  "
		if i == typeCursor {
			cursor = "▸ "
		}

		line := cursor + t.name
		if i == typeCursor {
			mainContent.WriteString(selectedStyle.Render(line))
		} else {
			mainContent.WriteString(normalStyle.Render(line))
		}
		mainContent.WriteString("\n")
		mainContent.WriteString("  " + descStyle.Render(t.desc))
		mainContent.WriteString("\n\n")
	}

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"enter select",
		"↑↓ navigate",
		"esc cancel",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.width - 8).
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

// RenderNameInput renders the component name input view
func (r *ComponentCreationViewRenderer) RenderNameInput(componentType, componentName string) string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	inputStyle := InputStyle.
		Width(60)

	// Calculate dimensions
	contentWidth := r.width - 4 // Match help pane width
	contentHeight := r.height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true) // Purple for active single pane

	// Convert plural to singular for heading
	singularType := componentType
	if strings.HasSuffix(strings.ToLower(componentType), "s") {
		singularType = componentType[:len(componentType)-1]
	}
	heading := fmt.Sprintf("CREATE NEW %s", strings.ToUpper(singularType))
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Name input prompt - centered
	promptText := promptStyle.Render("Enter a descriptive name:")
	centeredPromptStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	mainContent.WriteString(headerPadding.Render(centeredPromptStyle.Render(promptText)))
	mainContent.WriteString("\n\n")

	// Input field with cursor
	input := componentName + "│" // cursor

	// Render input field with padding for centering
	inputFieldContent := inputStyle.Render(input)
	
	// Add padding to center the input field properly
	centeredInputStyle := lipgloss.NewStyle().
		Width(contentWidth - 4). // Account for padding
		Align(lipgloss.Center)
	
	mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(inputFieldContent)))
	
	// Check if component name already exists and show warning
	if componentName != "" {
		testFilename := sanitizeFileName(componentName) + ".md"
		var componentTypeDir string
		switch componentType {
		case models.ComponentTypeContext:
			componentTypeDir = "contexts"
		case models.ComponentTypePrompt:
			componentTypeDir = "prompts"
		case models.ComponentTypeRules:
			componentTypeDir = "rules"
		}
		
		existingComponents, _ := files.ListComponents(componentTypeDir)
		for _, existing := range existingComponents {
			if strings.EqualFold(existing, testFilename) {
				// Show warning
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // Orange/yellow for warning
					Bold(true)
				warningText := warningStyle.Render(fmt.Sprintf("⚠ Warning: %s '%s' already exists", strings.Title(componentType), componentName))
				mainContent.WriteString("\n\n")
				mainContent.WriteString(headerPadding.Render(centeredInputStyle.Render(warningText)))
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
		"esc back",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.width - 8).
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

// RenderContentEdit renders the component content editing view (for fallback when enhanced editor is not used)
func (r *ComponentCreationViewRenderer) RenderContentEdit(componentType, componentName, componentContent string) string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170"))

	// Calculate dimensions
	contentWidth := r.width - 4 // Match help pane width
	contentHeight := r.height - 11 // Reserve space for help pane and status bar

	// Build main content
	var mainContent strings.Builder

	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true) // Purple for active single pane

	// Convert plural to singular for heading
	singularType := componentType
	if strings.HasSuffix(strings.ToLower(componentType), "s") {
		singularType = componentType[:len(componentType)-1]
	}
	heading := fmt.Sprintf("CREATE NEW %s: %s", strings.ToUpper(singularType), componentName)
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true) // Purple for active single pane
	mainContent.WriteString(headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth))))
	mainContent.WriteString("\n\n")

	// Editor content with cursor
	content := componentContent + "│" // cursor
	
	// Preprocess content to handle carriage returns and ensure proper line breaks
	processedContent := preprocessContent(content)
	
	// Calculate available width for wrapping (accounting for padding)
	availableWidth := contentWidth - 4 // 2 for border, 2 for headerPadding
	if availableWidth < 1 {
		availableWidth = 1
	}
	
	// Wrap content to prevent overflow
	wrappedContent := wordwrap.String(processedContent, availableWidth)

	mainContent.WriteString(headerPadding.Render(wrappedContent))

	// Apply border to main content
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())

	// Help section
	help := []string{
		"^s save",
		"esc back",
	}

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.width - 4).
		Padding(0, 1)

	helpContent := formatHelpText(help)
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.width - 8).
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

// RenderWithEnhancedEditor renders the component creation view with the enhanced editor
func (r *ComponentCreationViewRenderer) RenderWithEnhancedEditor(enhancedEditor *EnhancedEditorState, componentType, componentName string) string {
	if enhancedEditor == nil || !enhancedEditor.IsActive() {
		// Fallback to simple editor
		return r.RenderContentEdit(componentType, componentName, enhancedEditor.GetContent())
	}
	
	// Update the editor's component name to show in the title
	// This is temporary for rendering purposes
	originalName := enhancedEditor.ComponentName
	singularType := componentType
	if strings.HasSuffix(strings.ToLower(componentType), "s") {
		singularType = componentType[:len(componentType)-1]
	}
	enhancedEditor.ComponentName = fmt.Sprintf("NEW %s: %s", strings.ToUpper(singularType), componentName)
	
	// Create enhanced editor renderer
	editorRenderer := NewEnhancedEditorRenderer(r.width, r.height)
	
	// Use the enhanced editor's render method
	result := editorRenderer.Render(enhancedEditor)
	
	// Restore original name
	enhancedEditor.ComponentName = originalName
	
	return result
}