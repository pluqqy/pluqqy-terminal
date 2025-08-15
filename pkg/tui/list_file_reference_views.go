package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FileReferenceRenderer handles rendering for the file browser
type FileReferenceRenderer struct {
	Width  int
	Height int
}

// NewFileReferenceRenderer creates a new file reference renderer
func NewFileReferenceRenderer(width, height int) *FileReferenceRenderer {
	return &FileReferenceRenderer{
		Width:  width,
		Height: height,
	}
}

// RenderFileBrowser renders the complete file browser interface
func (r *FileReferenceRenderer) RenderFileBrowser(state *FileReferenceState) string {
	if !state.IsActive() {
		return ""
	}

	// Build components
	header := r.renderHeader(state)
	breadcrumbs := r.renderBreadcrumbs(state.GetBreadcrumbs())
	fileList := r.renderFileList(state)
	statusBar := r.renderStatusBar(state)
	help := r.renderHelp()

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		breadcrumbs,
		fileList,
		statusBar,
		help,
	)

	// Apply border
	bordered := ActiveBorderStyle.
		Width(r.Width - 2).
		Height(r.Height - 2).
		Render(content)

	return bordered
}

// renderHeader renders the file browser header
func (r *FileReferenceRenderer) renderHeader(state *FileReferenceState) string {
	headerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	titleStyle := GetActiveHeaderStyle(true)
	heading := "SELECT FILE REFERENCE"
	
	if state.FilterPattern != "" {
		heading = fmt.Sprintf("SELECT FILE REFERENCE (filter: %s)", state.FilterPattern)
	}
	
	contentWidth := r.Width - 8
	remainingWidth := contentWidth - len(heading) - 5
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	colonStyle := GetActiveColonStyle(true)
	
	return headerPadding.Render(titleStyle.Render(heading) + " " + colonStyle.Render(strings.Repeat(":", remainingWidth)))
}

// renderBreadcrumbs renders the path breadcrumbs
func (r *FileReferenceRenderer) renderBreadcrumbs(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Join with separator
	breadcrumbs := strings.Join(parts, " › ")

	// Truncate if too long
	maxWidth := r.Width - 6
	if len(breadcrumbs) > maxWidth {
		// Show last parts that fit
		for len(breadcrumbs) > maxWidth && len(parts) > 1 {
			parts = parts[1:]
			breadcrumbs = "... › " + strings.Join(parts, " › ")
		}
	}

	styled := DescriptionStyle.
		Width(r.Width - 4).
		Render(breadcrumbs)

	return styled
}

// renderFileList renders the list of files
func (r *FileReferenceRenderer) renderFileList(state *FileReferenceState) string {
	if !state.HasFiles() {
		return r.renderEmptyState()
	}

	// Calculate available height for file list
	listHeight := r.Height - 10 // Account for header, breadcrumbs, status, help
	state.SetMaxVisible(listHeight)

	// Get visible files
	visibleFiles := state.GetVisibleFiles()
	
	var lines []string
	for i, file := range visibleFiles {
		actualIndex := state.ScrollOffset + i
		line := r.renderFileItem(file, state.IsSelected(actualIndex))
		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if state.ScrollOffset > 0 {
		lines = append([]string{"  ↑ more files above"}, lines...)
	}
	if state.ScrollOffset+state.MaxVisible < state.GetFileCount() {
		lines = append(lines, "  ↓ more files below")
	}

	fileList := strings.Join(lines, "\n")

	// Apply padding
	padded := ContentPaddingStyle.
		Width(r.Width - 4).
		Render(fileList)

	return padded
}

// renderFileItem renders a single file item
func (r *FileReferenceRenderer) renderFileItem(file FileInfo, selected bool) string {
	// Get icon based on file type
	icon := r.getFileIcon(file)

	// Format size
	sizeStr := ""
	if !file.IsDir {
		sizeStr = FormatFileSize(file.Size)
	}

	// Build the line
	nameWidth := r.Width - 20 // Leave room for size and padding
	name := file.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Format the line
	line := fmt.Sprintf("%s %-*s %8s", icon, nameWidth, name, sizeStr)

	// Apply selection style
	if selected {
		return SelectedStyle.Render(line)
	}

	return NormalStyle.Render(line)
}

// renderEmptyState renders when no files are available
func (r *FileReferenceRenderer) renderEmptyState() string {
	message := "No files found"
	
	styled := EmptyInactiveStyle.
		Width(r.Width - 4).
		Align(lipgloss.Center).
		Render(message)

	// Add padding to center vertically
	padded := lipgloss.NewStyle().
		Height(r.Height - 10).
		Render(styled)

	return padded
}

// renderStatusBar renders the status bar
func (r *FileReferenceRenderer) renderStatusBar(state *FileReferenceState) string {
	current := state.GetCurrentFile()
	if current == nil {
		return ""
	}

	status := fmt.Sprintf("%d/%d • %s", 
		state.SelectedIndex+1, 
		state.GetFileCount(),
		GetFileType(current.Name))

	if state.ShowHidden {
		status += " • 👻 Hidden"
	}

	styled := DescriptionStyle.
		Width(r.Width - 4).
		Render(status)

	return styled
}

// renderHelp renders the help text
func (r *FileReferenceRenderer) renderHelp() string {
	help := "↑↓: Navigate • Enter: Select • Esc: Cancel • Tab: Hidden • /: Filter"

	styled := DescriptionStyle.
		Width(r.Width - 4).
		Align(lipgloss.Center).
		Render(help)

	return styled
}

// getFileIcon returns an appropriate icon for a file
func (r *FileReferenceRenderer) getFileIcon(file FileInfo) string {
	if file.IsDir {
		return "📁"
	}

	// Check by extension
	ext := GetFileExtension(file.Name)
	switch ext {
	case ".md", ".markdown":
		return "📝"
	case ".yaml", ".yml":
		return "⚙️"
	case ".json":
		return "📋"
	case ".txt":
		return "📄"
	case ".go":
		return "🐹"
	case ".js":
		return "🟨"
	case ".ts":
		return "🔷"
	case ".py":
		return "🐍"
	case ".sh", ".bash", ".zsh":
		return "🐚"
	case ".git", ".gitignore":
		return "🔀"
	case ".env":
		return "🔐"
	default:
		if strings.HasPrefix(file.Name, ".") {
			return "👻"
		}
		return "📄"
	}
}

// renderFilePreview renders a preview of the selected file
func (r *FileReferenceRenderer) renderFilePreview(content string) string {
	// Truncate content if too long
	maxLines := 10
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}

	preview := strings.Join(lines, "\n")

	// Apply style
	styled := NormalStyle.
		Width(r.Width - 4).
		Render(preview)

	// Add border
	bordered := InactiveBorderStyle.
		Width(r.Width - 4).
		Render(styled)

	return bordered
}

// renderSearchBar renders a search/filter input bar
func (r *FileReferenceRenderer) renderSearchBar(pattern string) string {
	prompt := "Filter: "
	input := pattern

	// Show cursor
	if len(input) < r.Width-10 {
		input += "█"
	}

	line := prompt + input

	styled := InputStyle.
		Width(r.Width - 4).
		Render(line)

	return styled
}

// renderScrollbar renders a visual scrollbar
func (r *FileReferenceRenderer) renderScrollbar(state *FileReferenceState) string {
	total := state.GetFileCount()
	visible := state.MaxVisible
	offset := state.ScrollOffset

	if total <= visible {
		return ""
	}

	// Calculate scrollbar position
	barHeight := visible - 2
	thumbHeight := max(1, (visible*barHeight)/total)
	thumbPos := (offset * barHeight) / total

	var scrollbar []string
	for i := 0; i < barHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			scrollbar = append(scrollbar, "█")
		} else {
			scrollbar = append(scrollbar, "│")
		}
	}

	return strings.Join(scrollbar, "\n")
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}