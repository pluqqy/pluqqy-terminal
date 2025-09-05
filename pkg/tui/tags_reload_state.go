package tui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// TagReloader manages the state and logic for reloading tags from components and pipelines
type TagReloader struct {
	// Core state
	Active      bool
	IsReloading bool

	// Result tracking
	ReloadResult *TagReloadResult
	LastError    error

	// Statistics
	ComponentsProcessed int
	PipelinesProcessed  int
	TagsFound           map[string]int // Tag name -> count
}

// TagReloadResult holds the results of a tag reload operation
type TagReloadResult struct {
	ComponentsScanned int
	PipelinesScanned  int
	NewTags           []string
	UpdatedTags       []string
	FailedFiles       []string
	TotalTags         int
}

// TagReloadMsg is sent when tag reload completes
type TagReloadMsg struct {
	Result *TagReloadResult
	Error  error
}

// NewTagReloader creates a new tag reloader instance
func NewTagReloader() *TagReloader {
	return &TagReloader{
		TagsFound: make(map[string]int),
	}
}

// Start begins the tag reload process
func (tr *TagReloader) Start() tea.Cmd {
	tr.Active = true
	tr.IsReloading = true
	tr.ReloadResult = nil
	tr.LastError = nil
	tr.ComponentsProcessed = 0
	tr.PipelinesProcessed = 0
	tr.TagsFound = make(map[string]int)

	// Return command to perform reload in background
	return tr.performReload
}

// performReload executes the tag reload operation
func (tr *TagReloader) performReload() tea.Msg {
	result := &TagReloadResult{
		FailedFiles: []string{},
	}

	// Get or create tag registry
	registry, err := tags.NewRegistry()
	if err != nil {
		return TagReloadMsg{
			Result: result,
			Error:  fmt.Errorf("failed to load tag registry: %w", err),
		}
	}

	// Track existing tags for comparison
	existingTags := make(map[string]bool)
	for _, tag := range registry.ListTags() {
		existingTags[models.NormalizeTagName(tag.Name)] = true
	}

	// Scan all component types
	componentTypes := []string{
		models.ComponentTypePrompt,
		models.ComponentTypeContext,
		models.ComponentTypeRules,
	}

	for _, compType := range componentTypes {
		components, err := files.ListComponents(compType)
		if err != nil {
			// Log error but continue with other types
			continue
		}

		for _, compFile := range components {
			compPath := filepath.Join(files.ComponentsDir, compType, compFile)
			comp, err := files.ReadComponent(compPath)
			if err != nil {
				result.FailedFiles = append(result.FailedFiles, compPath)
				continue
			}

			result.ComponentsScanned++

			// Process tags from this component
			for _, tagName := range comp.Tags {
				normalizedName := models.NormalizeTagName(tagName)
				tr.TagsFound[normalizedName]++

				// Check if this is a new tag
				if !existingTags[normalizedName] {
					// Create new tag in registry
					if _, err := registry.GetOrCreateTag(tagName); err == nil {
						result.NewTags = append(result.NewTags, normalizedName)
						existingTags[normalizedName] = true
					}
				}
			}
		}
	}

	// Scan all pipelines
	pipelines, err := files.ListPipelines()
	if err == nil {
		for _, pipelineFile := range pipelines {
			pipeline, err := files.ReadPipeline(pipelineFile)
			if err != nil {
				result.FailedFiles = append(result.FailedFiles, pipelineFile)
				continue
			}

			result.PipelinesScanned++

			// Process tags from this pipeline
			for _, tagName := range pipeline.Tags {
				normalizedName := models.NormalizeTagName(tagName)
				tr.TagsFound[normalizedName]++

				// Check if this is a new tag
				if !existingTags[normalizedName] {
					// Create new tag in registry
					if _, err := registry.GetOrCreateTag(tagName); err == nil {
						result.NewTags = append(result.NewTags, normalizedName)
						existingTags[normalizedName] = true
					}
				}
			}
		}
	}

	// Save the updated registry
	if err := registry.Save(); err != nil {
		return TagReloadMsg{
			Result: result,
			Error:  fmt.Errorf("failed to save tag registry: %w", err),
		}
	}

	// Set total tags count
	result.TotalTags = len(tr.TagsFound)

	return TagReloadMsg{
		Result: result,
		Error:  nil,
	}
}

// HandleMessage processes messages related to tag reloading
func (tr *TagReloader) HandleMessage(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case TagReloadMsg:
		tr.IsReloading = false
		tr.ReloadResult = msg.Result
		tr.LastError = msg.Error

		if msg.Result != nil {
			tr.ComponentsProcessed = msg.Result.ComponentsScanned
			tr.PipelinesProcessed = msg.Result.PipelinesScanned
		}

		// Deactivate after a moment to show results
		if msg.Error == nil {
			return true, tr.deactivateAfterDelay()
		}
		return true, nil
	}

	return false, nil
}

// deactivateAfterDelay returns a command that deactivates the reloader after showing results
func (tr *TagReloader) deactivateAfterDelay() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tagReloadCompleteMsg{}
	})
}

// tagReloadCompleteMsg signals that the reload display should be hidden
type tagReloadCompleteMsg struct{}

// HandleComplete processes the completion message
func (tr *TagReloader) HandleComplete() {
	tr.Active = false
	tr.IsReloading = false
}

// Reset clears the tag reloader state
func (tr *TagReloader) Reset() {
	tr.Active = false
	tr.IsReloading = false
	tr.ReloadResult = nil
	tr.LastError = nil
	tr.ComponentsProcessed = 0
	tr.PipelinesProcessed = 0
	tr.TagsFound = make(map[string]int)
}

// GetStatus returns a status message for the current reload state
func (tr *TagReloader) GetStatus() string {
	if !tr.Active {
		return ""
	}

	if tr.IsReloading {
		return "Reloading tags from components and pipelines..."
	}

	if tr.LastError != nil {
		return fmt.Sprintf("Tag reload failed: %v", tr.LastError)
	}

	if tr.ReloadResult != nil {
		if len(tr.ReloadResult.NewTags) > 0 {
			return fmt.Sprintf("Tag reload complete: %d new tags found", len(tr.ReloadResult.NewTags))
		}
		return fmt.Sprintf("Tag reload complete: %d total tags", tr.ReloadResult.TotalTags)
	}

	return ""
}

// IsActive returns whether the tag reloader is currently active
func (tr *TagReloader) IsActive() bool {
	return tr.Active
}
