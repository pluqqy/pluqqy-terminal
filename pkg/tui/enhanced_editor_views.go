package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pluqqy/pluqqy-terminal/pkg/utils"
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
	// Now we have: editor pane + status pane (3 lines with border) + help pane (3 lines)
	contentWidth := r.Width - 4   // Match help pane width
	editorHeight := r.Height - 10 // Reserve: help pane (3) + status pane (3) + spacing (4)

	innerWidth := contentWidth - 2  // Account for borders
	innerHeight := editorHeight - 2 // Account for borders

	// Update stats before rendering
	state.UpdateStats()

	// Build header with title and token count
	header := r.renderHeaderWithTokens(state)

	// Calculate textarea height to fill the editor pane
	// Leave room for header and one blank line after it
	textareaHeight := innerHeight - 2 // -1 for header, -1 for blank line
	if textareaHeight < 1 {
		textareaHeight = 1
	}

	// Build textarea
	state.SetTextareaDimensions(innerWidth, textareaHeight)
	textarea := state.Textarea.View()

	// Build editor content (without status bar)
	editorContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"", // blank line after header
		textarea,
	)

	// Apply border to editor content
	editorBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorActive)).
		Width(contentWidth).
		Height(editorHeight)

	editorPane := editorBorderStyle.Render(editorContent)

	// Build status bar content - use innerWidth for content inside border
	statusContent := r.renderEnhancedStatusBar(state, innerWidth)

	// Apply border to status bar to create its own pane
	statusBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")). // Green border for status pane
		Width(contentWidth)

	statusPane := statusBorderStyle.Render(statusContent)

	// Render help as footer pane
	helpPane := r.renderHelpPane(state.GetMode())

	// Add padding around the content
	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Combine all three panes: editor, status, help
	return lipgloss.JoinVertical(
		lipgloss.Left,
		contentStyle.Render(editorPane),
		contentStyle.Render(statusPane), // Status bar in its own bordered pane
		helpPane,
	)
}

// renderWithFilePicker renders editor with file picker overlay
func (r *EnhancedEditorRenderer) renderWithFilePicker(state *EnhancedEditorState) string {
	// When in file picker mode, just show the file picker
	// The editor content is preserved in state
	return r.renderFilePicker(state)
}

// renderHeader renders the editor header (kept for compatibility)
func (r *EnhancedEditorRenderer) renderHeader(state *EnhancedEditorState) string {
	return r.renderHeaderWithTokens(state)
}

// renderHeaderWithTokens renders the header with token count
func (r *EnhancedEditorRenderer) renderHeaderWithTokens(state *EnhancedEditorState) string {
	contentWidth := r.Width - 8 // Account for borders and padding

	// Left side: title
	titleStyle := GetActiveHeaderStyle(true)
	heading := fmt.Sprintf("EDITING: %s", state.ComponentName)

	// Right side: token count
	tokenCount := utils.EstimateTokens(state.Content)
	tokenBadgeStyle := GetTokenBadgeStyle(tokenCount)
	tokenBadge := tokenBadgeStyle.Render(utils.FormatTokenCount(tokenCount))

	// Calculate spacing
	titleLen := len(heading)
	tokenLen := lipgloss.Width(tokenBadge)
	colonLen := contentWidth - titleLen - tokenLen - 2 // -2 for spaces
	if colonLen < 3 {
		colonLen = 3
	}

	colonStyle := GetActiveColonStyle(true)

	// Combine parts
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	return headerPadding.Render(
		titleStyle.Render(heading) + " " +
			colonStyle.Render(strings.Repeat(":", colonLen)) + " " +
			tokenBadge,
	)
}

// renderTextarea renders the textarea component
func (r *EnhancedEditorRenderer) renderTextarea(state *EnhancedEditorState) string {
	// Calculate dimensions
	contentWidth := r.Width - 8       // Account for borders and padding
	textareaWidth := contentWidth - 2 // Additional padding
	textareaHeight := r.Height - 15   // Account for header, status, help pane

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

	// Show recent files if available
	if state.RecentFiles.HasRecentFiles() {
		recentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorActive)).
			PaddingLeft(2)
		mainContent.WriteString(recentStyle.Render("RECENT FILES (press 1-5):"))
		mainContent.WriteString("\n")

		recentFiles := state.RecentFiles.GetRecentFiles()
		for i, rf := range recentFiles {
			if i >= 5 {
				break
			}
			numStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorActive)).
				Bold(true)
			fileStyle := DescriptionStyle

			entry := fmt.Sprintf("  %s  %s",
				numStyle.Render(fmt.Sprintf("%d.", i+1)),
				fileStyle.Render(rf.Name))
			mainContent.WriteString(entry)
			mainContent.WriteString("\n")
		}
		mainContent.WriteString("\n")
		mainContent.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(2).
			Render("──────────────────"))
		mainContent.WriteString("\n\n")
	}

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
		Width(r.Width-4).
		Padding(0, 1)

	var help []string
	switch mode {
	case EditorModeNormal:
		help = []string{
			"↑↓ navigate",
			"^z undo",
			"^l clean",
			"@ insert file ref",
			"\\@ literal @",
			"^s save",
			"^x external",
			"esc cancel",
		}
	case EditorModeFilePicking:
		help = []string{
			"↑↓ navigate",
			"^z undo",
			"^l clean",
			"enter select",
			"esc cancel",
			"tab toggle hidden",
		}
	default:
		help = []string{
			"↑↓ navigate",
			"^z undo",
			"^l clean",
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

// renderStatusBar renders the status bar with save indicator and clipboard status (kept for compatibility)
func (r *EnhancedEditorRenderer) renderStatusBar(state *EnhancedEditorState) string {
	return r.renderEnhancedStatusBar(state, r.Width-4)
}

// renderEnhancedStatusBar renders the enhanced status bar with stats
func (r *EnhancedEditorRenderer) renderEnhancedStatusBar(state *EnhancedEditorState, width int) string {
	var sb strings.Builder

	// Left side: save status and feedback
	var leftParts []string

	// Save status
	var saveStatus string
	var statusColor string
	if state.HasUnsavedChanges() {
		saveStatus = "● Modified"
		statusColor = ColorWarning
	} else {
		saveStatus = "○ Saved"
		statusColor = ColorSuccess
	}

	// Style the save status with appropriate color
	saveStatusStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Render(saveStatus)
	leftParts = append(leftParts, saveStatusStyled)

	// Action feedback (temporary status messages)
	if feedback, ok := state.ActionFeedback.GetActionFeedback(); ok {
		leftParts = append(leftParts, feedback)
	} else if status, ok := state.StatusManager.GetStatus(); ok {
		leftParts = append(leftParts, status)
	}

	// Right side: stats
	var rightParts []string

	// Line and word count
	rightParts = append(rightParts, fmt.Sprintf("L:%d W:%d", state.LineCount, state.WordCount))

	// Cursor position (if available)
	if state.CurrentLine > 0 {
		rightParts = append(rightParts, fmt.Sprintf("%d:%d", state.CurrentLine, state.CurrentColumn))
	}

	// Undo stack indicator
	if len(state.UndoStack) > 0 {
		rightParts = append(rightParts, fmt.Sprintf("↶%d", len(state.UndoStack)))
	}

	leftStr := strings.Join(leftParts, "  ")
	rightStr := strings.Join(rightParts, "  ")

	// Style the right side with a subtle color
	rightStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(rightStr)

	// Calculate padding (accounting for 4 spaces of left padding and 2 right)
	leftLen := lipgloss.Width(leftStr)
	rightLen := lipgloss.Width(rightStr)
	padding := width - leftLen - rightLen - 6 // -4 for left padding, -2 for right
	if padding < 1 {
		padding = 1
	}

	// Build status content efficiently with strings.Builder
	sb.WriteString("    ") // Left padding
	sb.WriteString(leftStr)
	sb.WriteString(strings.Repeat(" ", padding))
	sb.WriteString(rightStyled)
	sb.WriteString("  ") // Right padding

	return sb.String()
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
