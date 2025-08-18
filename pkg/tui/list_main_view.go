package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
)

// MainViewRenderer handles the orchestration of all view components
type MainViewRenderer struct {
	Width           int
	Height          int
	ShowPreview     bool
	ActivePane      pane
	LastDataPane    pane
	SearchBar       *SearchBar
	PreviewViewport viewport.Model
	PreviewContent  string
	sharedLayout    *SharedLayout
}

// NewMainViewRenderer creates a new main view renderer
func NewMainViewRenderer(width, height int) *MainViewRenderer {
	return &MainViewRenderer{
		Width:        width,
		Height:       height,
		sharedLayout: NewSharedLayout(width, height, false),
	}
}

// RenderErrorView renders an error state
func (r *MainViewRenderer) RenderErrorView(err error) string {
	return fmt.Sprintf("Error: Failed to load pipelines: %v\n\nPress 'q' to quit", err)
}

// RenderHelpPane renders the help text in a bordered pane
func (r *MainViewRenderer) RenderHelpPane(searchActive bool) string {
	// Update shared layout preview state
	r.sharedLayout.SetShowPreview(r.ShowPreview)
	
	var helpRows [][]string
	if searchActive {
		// Show search syntax help when search is active
		helpRows = [][]string{
			{"tab switch pane", "esc clear+exit search"},
			{"tag:<name>", "type:<type>", "status:archived", "<keyword>", "combine with spaces"},
		}
	} else {
		// Show normal navigation help - grouped by function
		if r.ActivePane == previewPane {
			// Preview pane active - minimal help
			helpRows = [][]string{
				{"/ search", "tab switch pane", "↑↓ nav", "p preview", "s settings", "^c quit"},
			}
		} else if r.ActivePane == pipelinesPane {
			// Pipelines pane active - show M diagram and S set, hide ^x external
			helpRows = [][]string{
				// Row 1: Navigation & system
				{"/ search", "tab switch pane", "↑↓ nav", "s settings", "p preview", "M diagram", "S set", "y copy", "^c quit"},
				// Row 2: Component operations
				{"n new", "e edit", "^d delete", "R rename", "C clone", "t tag", "a archive/unarchive"},
			}
		} else {
			// Components pane active - show ^x external, hide M diagram and S set
			helpRows = [][]string{
				// Row 1: Navigation & system
				{"/ search", "tab switch pane", "↑↓ nav", "s settings", "p preview", "^c quit"},
				// Row 2: Component operations
				{"n new", "e edit", "^x external", "^d delete", "R rename", "C clone", "t tag", "a archive/unarchive"},
			}
		}
	}
	
	return r.sharedLayout.RenderHelpPane(helpRows)
}

// RenderPreviewPane renders the preview pane with border and header
func (r *MainViewRenderer) RenderPreviewPane(pipelines []pipelineItem, components []componentItem, pipelineCursor, componentCursor int) string {
	if !r.ShowPreview || r.PreviewContent == "" {
		return ""
	}
	
	// Update shared layout state
	r.sharedLayout.SetShowPreview(r.ShowPreview)
	
	// Create heading
	var previewHeading string
	
	// Determine what we're previewing based on lastDataPane
	if r.LastDataPane == pipelinesPane && len(pipelines) > 0 && pipelineCursor >= 0 && pipelineCursor < len(pipelines) {
		pipelineName := pipelines[pipelineCursor].path // Use path which contains the filename
		previewHeading = fmt.Sprintf("PIPELINE PREVIEW (%s)", pipelineName)
	} else if r.LastDataPane == componentsPane && len(components) > 0 && componentCursor >= 0 && componentCursor < len(components) {
		comp := components[componentCursor]
		// Use the actual filename from the path
		componentFilename := filepath.Base(comp.path)
		previewHeading = fmt.Sprintf("COMPONENT PREVIEW (%s)", componentFilename)
	} else {
		previewHeading = "PREVIEW"
	}
	
	// Configure preview
	config := PreviewConfig{
		Content:     r.PreviewContent,
		Heading:     previewHeading,
		ActivePane:  r.ActivePane,
		PreviewPane: previewPane,
		Viewport:    r.PreviewViewport,
	}
	
	return r.sharedLayout.RenderPreviewPane(config)
}

// RenderConfirmationDialogs renders delete/archive confirmation dialogs
func (r *MainViewRenderer) RenderConfirmationDialogs(pipelineOperator *PipelineOperator) string {
	var result strings.Builder
	
	// Show delete confirmation if active
	if pipelineOperator.IsDeleteConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		result.WriteString("\n")
		result.WriteString(confirmStyle.Render(pipelineOperator.ViewDeleteConfirm(r.Width - 4)))
	}
	
	// Show archive confirmation if active
	if pipelineOperator.IsArchiveConfirmActive() {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange for archive
			Bold(true).
			MarginTop(2).
			MarginBottom(1)
		result.WriteString("\n")
		result.WriteString(confirmStyle.Render(pipelineOperator.ViewArchiveConfirm(r.Width - 4)))
	}
	
	return result.String()
}

// CalculateContentHeight calculates the height for the main content area
func (r *MainViewRenderer) CalculateContentHeight() int {
	// Update shared layout state and get calculated height
	r.sharedLayout.SetShowPreview(r.ShowPreview)
	r.sharedLayout.SetSize(r.Width, r.Height)
	return r.sharedLayout.GetContentHeight()
}