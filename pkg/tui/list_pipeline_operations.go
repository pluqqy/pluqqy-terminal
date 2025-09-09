package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/composer"
	"github.com/pluqqy/pluqqy-terminal/pkg/examples"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// PipelineOperator handles all pipeline-related operations
type PipelineOperator struct {
	deleteConfirm  *ConfirmationModel
	archiveConfirm *ConfirmationModel
}

// NewPipelineOperator creates a new pipeline operator
func NewPipelineOperator() *PipelineOperator {
	return &PipelineOperator{
		deleteConfirm:  NewConfirmation(),
		archiveConfirm: NewConfirmation(),
	}
}

// SetPipeline generates and writes the pipeline output
func (po *PipelineOperator) SetPipeline(pipelinePath string) tea.Cmd {
	return func() tea.Msg {
		// Load pipeline
		pipeline, err := files.ReadPipeline(pipelinePath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to load pipeline '%s': %v", pipelinePath, err))
		}

		// Generate pipeline output
		output, err := composer.ComposePipeline(pipeline)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to generate pipeline output for '%s': %v", pipeline.Name, err))
		}

		// Load settings for output path
		settings, _ := files.ReadSettings()
		if settings == nil {
			settings = models.DefaultSettings()
		}

		// Write to configured output file
		outputPath := pipeline.OutputPath
		if outputPath == "" {
			outputPath = filepath.Join(settings.Output.ExportPath, settings.Output.DefaultFilename)
		}

		// Ensure the directory exists
		dir := filepath.Dir(outputPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return StatusMsg(fmt.Sprintf("Failed to create output directory '%s': %v", dir, err))
			}
		}

		err = composer.WritePLUQQYFile(output, outputPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to write output file '%s': %v", outputPath, err))
		}

		return StatusMsg(fmt.Sprintf("✓ Set pipeline: %s → %s", pipeline.Name, outputPath))
	}
}

// DeletePipeline deletes a pipeline and returns a command to reload the list
func (po *PipelineOperator) DeletePipeline(pipelinePath string, pipelineTags []string, isArchived bool, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Delete the pipeline file based on whether it's archived
		var err error
		if isArchived {
			err = files.DeleteArchivedPipeline(pipelinePath)
		} else {
			err = files.DeletePipeline(pipelinePath)
		}
		
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete pipeline '%s': %v", pipelinePath, err))
		}

		// Reload the pipeline list
		reloadFunc()

		// Start async tag cleanup if there were tags (only for active items)
		if len(pipelineTags) > 0 && !isArchived {
			go func() {
				tags.CleanupOrphanedTags(pipelineTags)
			}()
		}

		// Extract pipeline name from path
		pipelineName := strings.TrimSuffix(filepath.Base(pipelinePath), ".yaml")
		return StatusMsg(fmt.Sprintf("✓ Deleted pipeline: %s", pipelineName))
	}
}

// ArchivePipeline archives a pipeline and returns a command to reload the list
func (po *PipelineOperator) ArchivePipeline(pipelinePath string, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Archive the pipeline file
		err := files.ArchivePipeline(pipelinePath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive pipeline '%s': %v", pipelinePath, err))
		}

		// Reload the pipeline list
		reloadFunc()

		// Extract pipeline name from path
		pipelineName := strings.TrimSuffix(filepath.Base(pipelinePath), ".yaml")
		return StatusMsg(fmt.Sprintf("✓ Archived pipeline: %s", pipelineName))
	}
}

// OpenInEditor opens a file in the user's preferred editor
func (po *PipelineOperator) OpenInEditor(path string, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set. Please set it to your preferred editor.")
		}

		// Construct full path
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

		// Reload components to reflect any changes
		reloadFunc()

		return StatusMsg(fmt.Sprintf("Edited: %s", filepath.Base(path)))
	}
}

// Confirmation dialog methods

// ShowDeleteConfirmation shows a delete confirmation dialog
func (po *PipelineOperator) ShowDeleteConfirmation(message string, onConfirm, onCancel func() tea.Cmd) {
	po.deleteConfirm.ShowInline(message, true, onConfirm, onCancel)
}

// ShowArchiveConfirmation shows an archive confirmation dialog
func (po *PipelineOperator) ShowArchiveConfirmation(message string, onConfirm, onCancel func() tea.Cmd) {
	po.archiveConfirm.ShowInline(message, true, onConfirm, onCancel)
}

// IsDeleteConfirmActive returns true if delete confirmation is active
func (po *PipelineOperator) IsDeleteConfirmActive() bool {
	return po.deleteConfirm.Active()
}

// IsArchiveConfirmActive returns true if archive confirmation is active
func (po *PipelineOperator) IsArchiveConfirmActive() bool {
	return po.archiveConfirm.Active()
}

// UpdateDeleteConfirm handles update for delete confirmation
func (po *PipelineOperator) UpdateDeleteConfirm(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return po.deleteConfirm.Update(keyMsg)
	}
	return nil
}

// UpdateArchiveConfirm handles update for archive confirmation
func (po *PipelineOperator) UpdateArchiveConfirm(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return po.archiveConfirm.Update(keyMsg)
	}
	return nil
}

// ViewDeleteConfirm returns the delete confirmation view
func (po *PipelineOperator) ViewDeleteConfirm(width int) string {
	return po.deleteConfirm.ViewWithWidth(width)
}

// ViewArchiveConfirm returns the archive confirmation view
func (po *PipelineOperator) ViewArchiveConfirm(width int) string {
	return po.archiveConfirm.ViewWithWidth(width)
}

// File operations for components (shared between pipelines and components)

// DeleteComponent deletes a component file
func (po *PipelineOperator) DeleteComponent(comp componentItem, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Delete the component file based on whether it's archived
		var err error
		if comp.isArchived {
			err = files.DeleteArchivedComponent(comp.path)
		} else {
			err = files.DeleteComponent(comp.path)
		}
		
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to delete %s '%s': %v", comp.compType, comp.name, err))
		}

		// Reload the component list
		reloadFunc()

		// Start async tag cleanup if there were tags (only for active items)
		if len(comp.tags) > 0 && !comp.isArchived {
			go func() {
				tags.CleanupOrphanedTags(comp.tags)
			}()
		}

		return StatusMsg(fmt.Sprintf("✓ Deleted %s: %s", comp.compType, comp.name))
	}
}

// ArchiveComponent archives a component file
func (po *PipelineOperator) ArchiveComponent(comp componentItem, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Archive the component file
		err := files.ArchiveComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to archive %s '%s': %v", comp.compType, comp.name, err))
		}

		// Reload the component list
		reloadFunc()

		return StatusMsg(fmt.Sprintf("✓ Archived %s: %s", comp.compType, comp.name))
	}
}

// UnarchivePipeline unarchives a pipeline and returns a command to reload the list
func (po *PipelineOperator) UnarchivePipeline(pipelinePath string, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Unarchive the pipeline file
		err := files.UnarchivePipeline(pipelinePath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to unarchive pipeline '%s': %v", pipelinePath, err))
		}

		// Reload the pipeline list
		reloadFunc()

		// Extract pipeline name from path
		pipelineName := strings.TrimSuffix(filepath.Base(pipelinePath), ".yaml")

		return StatusMsg(fmt.Sprintf("✓ Unarchived pipeline: %s", pipelineName))
	}
}

// UnarchiveComponent unarchives a component file
func (po *PipelineOperator) UnarchiveComponent(comp componentItem, reloadFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Unarchive the component file
		err := files.UnarchiveComponent(comp.path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to unarchive %s '%s': %v", comp.compType, comp.name, err))
		}

		// Reload the component list
		reloadFunc()

		return StatusMsg(fmt.Sprintf("✓ Unarchived %s: %s", comp.compType, comp.name))
	}
}

// ImportExamples imports example components and pipelines
func (po *PipelineOperator) ImportExamples(reloadComponentsFunc func(), reloadPipelinesFunc func()) tea.Cmd {
	return func() tea.Msg {
		// Get general examples
		exampleSets := examples.GetExamples("general")
		
		if len(exampleSets) == 0 {
			return StatusMsg("No general examples found")
		}
		
		// Install components and pipelines from all general example sets
		totalComponents := 0
		totalPipelines := 0
		skippedComponents := 0
		skippedPipelines := 0
		
		for _, set := range exampleSets {
			// Install components
			for _, comp := range set.Components {
				installed, err := examples.InstallComponent(comp, false)
				if err != nil {
					return StatusMsg(fmt.Sprintf("Failed to install component: %v", err))
				}
				if installed {
					totalComponents++
				} else {
					skippedComponents++
				}
			}
			
			// Install pipelines
			for _, pipeline := range set.Pipelines {
				installed, err := examples.InstallPipeline(pipeline, false)
				if err != nil {
					return StatusMsg(fmt.Sprintf("Failed to install pipeline: %v", err))
				}
				if installed {
					totalPipelines++
				} else {
					skippedPipelines++
				}
			}
		}
		
		// Reload both components and pipelines
		reloadComponentsFunc()
		reloadPipelinesFunc()
		
		if skippedComponents > 0 || skippedPipelines > 0 {
			return StatusMsg(fmt.Sprintf("✓ Imported %d components and %d pipelines (%d already existed)", 
				totalComponents, totalPipelines, skippedComponents+skippedPipelines))
		}
		return StatusMsg(fmt.Sprintf("✓ Imported %d components and %d pipelines from general examples", 
			totalComponents, totalPipelines))
	}
}
