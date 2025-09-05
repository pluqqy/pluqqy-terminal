package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/tags"
)

// Save saves the current tags to the file
func (te *TagEditor) Save() tea.Cmd {
	return func() tea.Msg {
		var err error
		
		if te.ItemType == "component" {
			// Update component tags
			err = files.UpdateComponentTags(te.Path, te.CurrentTags)
		} else {
			// Update pipeline tags
			pipeline, err := files.ReadPipeline(te.Path)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to read pipeline: %v", err))
			}
			pipeline.Tags = te.CurrentTags
			err = files.WritePipeline(pipeline)
		}
		
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save tags: %v", err))
		}
		
		// Update the registry with any new tags
		registry, _ := tags.NewRegistry()
		if registry != nil {
			for _, tag := range te.CurrentTags {
				registry.GetOrCreateTag(tag)
			}
			registry.Save()
		}
		
		// Check for removed tags and cleanup orphaned ones
		removedTags := []string{}
		for _, oldTag := range te.OriginalTags {
			found := false
			for _, newTag := range te.CurrentTags {
				if oldTag == newTag {
					found = true
					break
				}
			}
			if !found {
				removedTags = append(removedTags, oldTag)
			}
		}
		
		// Cleanup orphaned tags asynchronously
		if len(removedTags) > 0 {
			go func() {
				tags.CleanupOrphanedTags(removedTags)
			}()
		}
		
		// Update original tags to match current
		te.TagEditorDataStore.UpdateOriginalTags()
		
		// Call the save callback if provided
		if te.Callbacks.OnSave != nil {
			te.Callbacks.OnSave(te.Path, te.CurrentTags)
		}
		
		// Reset the editor
		te.Reset()
		
		return ReloadMsg{Message: "✓ Tags saved"}
	}
}

// DeleteTagCompletely removes a tag from all files and the registry
// This is moved from list_tag_deletion_operations.go to be part of the unified component
func DeleteTagCompletely(tagToDelete string, showProgress func(string, int, int)) tea.Cmd {
	return func() tea.Msg {
		result := TagDeletionResult{
			TagName: tagToDelete,
			Errors:  []string{},
		}
		
		normalizedTag := models.NormalizeTagName(tagToDelete)
		totalEstimate := 0
		
		// Helper function to remove tag from a list (case-insensitive)
		removeTag := func(tags []string) ([]string, bool) {
			newTags := []string{}
			found := false
			for _, tag := range tags {
				if models.NormalizeTagName(tag) != normalizedTag {
					newTags = append(newTags, tag)
				} else {
					found = true
				}
			}
			return newTags, found
		}
		
		// 1. Process active components
		componentTypes := []string{
			models.ComponentTypePrompt,
			models.ComponentTypeContext,
			models.ComponentTypeRules,
		}
		
		for _, compType := range componentTypes {
			components, err := files.ListComponents(compType)
			if err != nil {
				continue
			}
			
			totalEstimate += len(components)
			
			for _, compFile := range components {
				compPath := filepath.Join(files.ComponentsDir, compType, compFile)
				result.FilesScanned++
				
				if showProgress != nil {
					showProgress(compPath, result.FilesScanned, totalEstimate)
				}
				
				comp, err := files.ReadComponent(compPath)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to read %s: %v", compPath, err))
					continue
				}
				
				newTags, found := removeTag(comp.Tags)
				if found {
					err = files.UpdateComponentTags(compPath, newTags)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to update %s: %v", compPath, err))
					} else {
						result.FilesUpdated++
					}
				}
			}
		}
		
		// 2. Process active pipelines
		pipelines, err := files.ListPipelines()
		if err == nil {
			totalEstimate += len(pipelines)
			
			for _, pipelineFile := range pipelines {
				result.FilesScanned++
				
				if showProgress != nil {
					showProgress(pipelineFile, result.FilesScanned, totalEstimate)
				}
				
				pipeline, err := files.ReadPipeline(pipelineFile)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to read pipeline %s: %v", pipelineFile, err))
					continue
				}
				
				newTags, found := removeTag(pipeline.Tags)
				if found {
					pipeline.Tags = newTags
					err = files.WritePipeline(pipeline)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to update pipeline %s: %v", pipelineFile, err))
					} else {
						result.FilesUpdated++
					}
				}
			}
		}
		
		// 3. Process archived components
		for _, compType := range componentTypes {
			// Get the correct subdirectory name for this component type
			var subDir string
			switch compType {
			case models.ComponentTypePrompt:
				subDir = files.PromptsDir
			case models.ComponentTypeContext:
				subDir = files.ContextsDir
			case models.ComponentTypeRules:
				subDir = files.RulesDir
			}
			
			archivedComps, err := files.ListArchivedComponents(compType)
			if err != nil {
				continue
			}
			
			totalEstimate += len(archivedComps)
			
			for _, compFile := range archivedComps {
				// Construct relative path for ReadArchivedComponent
				relativePath := filepath.Join(files.ComponentsDir, subDir, compFile)
				// Full path for display purposes
				displayPath := filepath.Join(files.PluqqyDir, files.ArchiveDir, relativePath)
				result.FilesScanned++
				
				if showProgress != nil {
					showProgress(displayPath, result.FilesScanned, totalEstimate)
				}
				
				comp, err := files.ReadArchivedComponent(relativePath)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to read archived %s: %v", displayPath, err))
					continue
				}
				
				newTags, found := removeTag(comp.Tags)
				if found {
					// For updating, we need to use the correct path
					// UpdateComponentTags expects a path relative to .pluqqy
					updatePath := filepath.Join(files.ArchiveDir, relativePath)
					err = files.UpdateComponentTags(updatePath, newTags)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to update archived %s: %v", displayPath, err))
					} else {
						result.FilesUpdated++
					}
				}
			}
		}
		
		// 4. Process archived pipelines
		archivedPipelines, err := files.ListArchivedPipelines()
		if err == nil {
			totalEstimate += len(archivedPipelines)
			
			for _, pipelineFile := range archivedPipelines {
				// Full path for display purposes
				displayPath := filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, pipelineFile)
				result.FilesScanned++
				
				if showProgress != nil {
					showProgress(displayPath, result.FilesScanned, totalEstimate)
				}
				
				// ReadArchivedPipeline expects just the filename
				pipeline, err := files.ReadArchivedPipeline(pipelineFile)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to read archived pipeline %s: %v", displayPath, err))
					continue
				}
				
				newTags, found := removeTag(pipeline.Tags)
				if found {
					pipeline.Tags = newTags
					// Need to write back to archive - WritePipeline would write to active pipelines
					// We'll write directly to the archive path
					archivePipelinePath := filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, pipelineFile)
					data, err := yaml.Marshal(pipeline)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to marshal archived pipeline %s: %v", displayPath, err))
						continue
					}
					err = os.WriteFile(archivePipelinePath, data, 0644)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to update archived pipeline %s: %v", displayPath, err))
					} else {
						result.FilesUpdated++
					}
				}
			}
		}
		
		// 5. Finally remove from registry
		registry, err := tags.NewRegistry()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to load registry: %v", err))
		} else {
			registry.RemoveTag(tagToDelete)
			if err := registry.Save(); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to save registry: %v", err))
			}
		}
		
		return tagDeletionCompleteMsg{Result: result}
	}
}

// formatDeletionResult creates a user-friendly message from deletion results
func formatDeletionResult(result TagDeletionResult) string {
	if len(result.Errors) > 0 {
		errorSummary := strings.Join(result.Errors[:tagEditorMin(3, len(result.Errors))], "; ")
		if len(result.Errors) > 3 {
			errorSummary += fmt.Sprintf(" (and %d more errors)", len(result.Errors)-3)
		}
		return fmt.Sprintf("× Tag '%s' partially deleted: %d files updated, %d errors: %s",
			result.TagName, result.FilesUpdated, len(result.Errors), errorSummary)
	}
	
	if result.FilesUpdated == 0 {
		return fmt.Sprintf("✓ Tag '%s' removed from registry (not used in any files)", result.TagName)
	}
	
	return fmt.Sprintf("✓ Tag '%s' deleted from registry and %d file%s", 
		result.TagName, 
		result.FilesUpdated,
		pluralS(result.FilesUpdated))
}

func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func tagEditorMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}