package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// EnhancedEditorRenderer handles ALL rendering logic for the enhanced editor
type EnhancedEditorRenderer struct {
	Width  int
	Height int
}

// NewEnhancedEditorRenderer creates a new renderer with dimensions
func NewEnhancedEditorRenderer(width, height int) *EnhancedEditorRenderer {
	return &EnhancedEditorRenderer{
		Width:  width,
		Height: height,
	}
}

// Render is the main render function for the enhanced editor
func (r *EnhancedEditorRenderer) Render(state *EnhancedEditorState) string {
	if !state.IsActive() {
		return ""
	}

	// Handle file picker mode
	if state.IsFilePicking() {
		return r.renderWithFilePicker(state)
	}

	// Render normal editor mode
	return r.renderNormalMode(state)
}

// renderNormalMode renders the editor in normal editing mode
func (r *EnhancedEditorRenderer) renderNormalMode(state *EnhancedEditorState) string {
	// Calculate dimensions
	contentWidth := r.Width - 4 // Match help pane width
	contentHeight := r.Height - 7 // Reserve space for help pane (3) + spacing (3) + status bar (1)
	
	// Build main content
	var mainContent strings.Builder
	
	// Add header
	header := r.renderHeader(state)
	mainContent.WriteString(header)
	mainContent.WriteString("\n\n")
	
	// Add textarea
	textarea := r.renderTextarea(state)
	mainContent.WriteString(textarea)
	
	// Add status bar
	statusBar := r.renderStatusBar(state)
	mainContent.WriteString("\n")
	mainContent.WriteString(statusBar)
	
	// Apply border to main content
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorActive))
	
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())
	
	// Render help as footer pane
	helpPane := r.renderHelpPane(state.GetMode())
	
	// Add padding around the content to match help pane
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Combine main and help panes
	return lipgloss.JoinVertical(
		lipgloss.Left,
		contentStyle.Render(mainPane),
		helpPane,
	)
}

// renderWithFilePicker renders editor with file picker overlay
func (r *EnhancedEditorRenderer) renderWithFilePicker(state *EnhancedEditorState) string {
	// When in file picker mode, just show the file picker
	// The editor content is preserved in state
	return r.renderFilePicker(state)
}


// renderHeader renders the editor header
func (r *EnhancedEditorRenderer) renderHeader(state *EnhancedEditorState) string {
	// Header with colons (pane heading style)
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true) // Purple for active single pane
	
	heading := fmt.Sprintf("EDITING: %s", strings.ToUpper(state.ComponentName))
	contentWidth := r.Width - 8 // Account for borders and padding
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true) // Purple for active single pane
	
	return headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth)))
}

// renderTextarea renders the textarea component
func (r *EnhancedEditorRenderer) renderTextarea(state *EnhancedEditorState) string {
	// Calculate dimensions
	contentWidth := r.Width - 8 // Account for borders and padding
	textareaWidth := contentWidth - 2 // Additional padding
	textareaHeight := r.Height - 15 // Account for header, status, help pane
	
	state.SetTextareaDimensions(textareaWidth, textareaHeight)
	
	// Get the textarea view
	textareaView := state.Textarea.View()
	
	// Apply padding
	padded := ContentPaddingStyle.Render(textareaView)
	
	return padded
}

// renderFilePicker renders the file picker as a full view
func (r *EnhancedEditorRenderer) renderFilePicker(state *EnhancedEditorState) string {
	// Calculate dimensions  
	contentWidth := r.Width - 4
	contentHeight := r.Height - 7 // Reserve space for help pane
	
	// Build main content
	var mainContent strings.Builder
	
	// Header
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true)
	heading := "SELECT FILE REFERENCE"
	remainingWidth := contentWidth - len(heading) - 10
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true)
	
	header := headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth)))
	mainContent.WriteString(header)
	mainContent.WriteString("\n\n")
	
	// Show current path from file picker
	pathStyle := DescriptionStyle.PaddingLeft(2)
	currentPath := state.FilePicker.CurrentDirectory
	if currentPath == "" {
		currentPath = "."
	}
	mainContent.WriteString(pathStyle.Render("Current: " + currentPath))
	mainContent.WriteString("\n\n")
	
	// Simple file list placeholder (since we're using bubbles filepicker)
	// The actual filepicker view
	pickerView := state.FilePicker.View()
	mainContent.WriteString(pickerView)
	
	// Apply border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorActive))
	
	mainPane := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent.String())
	
	// Help pane
	helpPane := r.renderHelpPane(EditorModeFilePicking)
	
	// Add padding around the content to match help pane
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	// Combine
	return lipgloss.JoinVertical(
		lipgloss.Left,
		contentStyle.Render(mainPane),
		helpPane,
	)
}

// renderHelpPane renders the help pane as a footer
func (r *EnhancedEditorRenderer) renderHelpPane(mode EditorMode) string {
	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(r.Width - 4).
		Padding(0, 1)
	
	var help []string
	switch mode {
	case EditorModeNormal:
		help = []string{
			"^s save",
			"esc cancel",
			"@ insert file ref",
			"\\@ literal @",
		}
	case EditorModeFilePicking:
		help = []string{
			"↑/↓ navigate",
			"enter select",
			"esc cancel",
			"tab toggle hidden",
		}
	default:
		help = []string{
			"^s save",
			"esc exit",
		}
	}
	
	helpContent := formatHelpText(help)
	
	// Right-align help text
	alignedHelp := lipgloss.NewStyle().
		Width(r.Width - 8).
		Align(lipgloss.Right).
		Render(helpContent)
	
	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	return contentStyle.Render(helpBorderStyle.Render(alignedHelp))
}

// renderStatusBar renders the status bar with save indicator
func (r *EnhancedEditorRenderer) renderStatusBar(state *EnhancedEditorState) string {
	var status string
	
	if state.HasUnsavedChanges() {
		status = "● Modified"
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Width(r.Width - 4).
			Align(lipgloss.Left).
			Render(status)
	}
	
	status = "○ Saved"
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)).
		Width(r.Width - 4).
		Align(lipgloss.Left).
		Render(status)
}

// renderLineNumbers renders line numbers for content
func (r *EnhancedEditorRenderer) renderLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var numberedLines []string
	
	for i, line := range lines {
		numbered := fmt.Sprintf("%3d │ %s", i+1, line)
		numberedLines = append(numberedLines, numbered)
	}
	
	return strings.Join(numberedLines, "\n")
}

// renderBreadcrumbs renders file path breadcrumbs
func (r *EnhancedEditorRenderer) renderBreadcrumbs(path string) string {
	parts := strings.Split(path, "/")
	breadcrumbs := strings.Join(parts, " › ")
	
	return DescriptionStyle.
		Width(r.Width - 4).
		Render(breadcrumbs)
}

// renderFileIcon returns an icon for a file type
func (r *EnhancedEditorRenderer) renderFileIcon(fileName string) string {
	switch {
	case strings.HasSuffix(fileName, ".md"):
		return "📝"
	case strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml"):
		return "⚙️"
	case strings.HasSuffix(fileName, ".json"):
		return "📋"
	case strings.HasSuffix(fileName, ".txt"):
		return "📄"
	case strings.HasPrefix(fileName, "."):
		return "👻"
	default:
		return "📄"
	}
}