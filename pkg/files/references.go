package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"gopkg.in/yaml.v3"
)

// UpdateComponentReferences updates all references to a renamed component in pipelines
func UpdateComponentReferences(oldPath, newPath, newDisplayName string) error {
	// Normalize paths for comparison
	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)
	
	// Track updated pipelines for potential rollback
	updatedPipelines := []struct {
		path     string
		original []byte
	}{}
	
	// Update active pipelines
	activePipelines, err := ListPipelines()
	if err != nil {
		return fmt.Errorf("failed to list active pipelines: %w", err)
	}
	
	for _, pipelineName := range activePipelines {
		// Read the pipeline (ReadPipeline expects just the filename)
		pipeline, err := ReadPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		modified := false
		for i, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Check if this component matches exactly
			if compPath == oldPath {
				// Update to new path with ../ prefix for relative path
				pipeline.Components[i].Path = "../" + newPath
				modified = true
			}
		}
		
		// Save if modified
		if modified {
			pipelinePath := filepath.Join(PluqqyDir, PipelinesDir, pipelineName)
			
			// Store original for rollback
			original, _ := readRawFile(pipelinePath)
			updatedPipelines = append(updatedPipelines, struct {
				path     string
				original []byte
			}{pipelinePath, original})
			
			// Write updated pipeline
			if err := WritePipeline(pipeline); err != nil {
				// Rollback all changes
				rollbackPipelines(updatedPipelines)
				return fmt.Errorf("failed to update pipeline %s: %w", pipelineName, err)
			}
		}
	}
	
	// Update archived pipelines
	archivedPipelines, err := ListArchivedPipelines()
	if err != nil {
		// Don't fail if we can't access archived pipelines
		return nil
	}
	
	for _, pipelineName := range archivedPipelines {
		// Read the archived pipeline (expects just filename)
		pipeline, err := ReadArchivedPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		modified := false
		for i, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Check if this component matches exactly
			if compPath == oldPath {
				// Update to new path with ../ prefix for relative path
				pipeline.Components[i].Path = "../" + newPath
				modified = true
			}
		}
		
		// Save if modified
		if modified {
			pipelinePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, pipelineName)
			
			// Store original for rollback
			original, _ := readRawFile(pipelinePath)
			updatedPipelines = append(updatedPipelines, struct {
				path     string
				original []byte
			}{pipelinePath, original})
			
			// Write updated pipeline directly to archive
			pipelineData, err := yaml.Marshal(pipeline)
			if err != nil {
				rollbackPipelines(updatedPipelines)
				return fmt.Errorf("failed to marshal archived pipeline %s: %w", pipelineName, err)
			}
			
			if err := os.WriteFile(pipelinePath, pipelineData, 0644); err != nil {
				// Rollback all changes
				rollbackPipelines(updatedPipelines)
				return fmt.Errorf("failed to update archived pipeline %s: %w", pipelineName, err)
			}
		}
	}
	
	return nil
}

// FindAffectedPipelines finds all pipelines that reference a component
func FindAffectedPipelines(componentPath string) (activeNames []string, archivedNames []string, err error) {
	// Normalize path for comparison
	componentPath = filepath.Clean(componentPath)
	
	// Check active pipelines
	activePipelines, err := ListPipelines()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list active pipelines: %w", err)
	}
	
	for _, pipelineName := range activePipelines {
		// Read the pipeline (expects just filename)
		pipeline, err := ReadPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		for _, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Check if this component matches exactly
			if compPath == componentPath {
				// Use pipeline's display name
				activeNames = append(activeNames, pipeline.Name)
				break
			}
		}
	}
	
	// Check archived pipelines
	archivedPipelines, err := ListArchivedPipelines()
	if err != nil {
		// Don't fail if we can't access archived pipelines
		return activeNames, archivedNames, nil
	}
	
	for _, pipelineName := range archivedPipelines {
		// Read the archived pipeline (expects just filename)
		pipeline, err := ReadArchivedPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		for _, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Check if this component matches exactly
			if compPath == componentPath {
				// Use pipeline's display name
				archivedNames = append(archivedNames, pipeline.Name)
				break
			}
		}
	}
	
	return activeNames, archivedNames, nil
}

// matchesPath checks if two paths refer to the same file
// It handles relative paths and different path representations
func matchesPath(path1, path2 string) bool {
	// Remove leading "../" from paths for comparison
	path1 = strings.TrimPrefix(path1, "../")
	path2 = strings.TrimPrefix(path2, "../")
	
	// Direct match
	if path1 == path2 {
		return true
	}
	
	// Check if one is relative and matches the end of the other
	if strings.HasSuffix(path1, path2) || strings.HasSuffix(path2, path1) {
		return true
	}
	
	// Check if they're the same file with different relative paths
	// For example: "../components/contexts/auth.md" vs ".pluqqy/components/contexts/auth.md"
	if filepath.Base(path1) == filepath.Base(path2) {
		// Extract the component path part (e.g., "components/contexts/auth.md")
		componentPath1 := extractComponentPath(path1)
		componentPath2 := extractComponentPath(path2)
		return componentPath1 == componentPath2
	}
	
	return false
}

// extractComponentPath extracts the component-relative path from a full path
func extractComponentPath(path string) string {
	// Look for "components/" in the path
	idx := strings.Index(path, "components/")
	if idx >= 0 {
		return path[idx:]
	}
	
	// If no "components/" found, return the base path
	return filepath.Base(path)
}

// makeRelativePath creates a relative path from newPath to be used from pipelinePath
func makeRelativePath(componentPath, pipelinePath string) string {
	// If the component path already starts with "../", keep it
	if strings.HasPrefix(componentPath, "../") {
		// Update the path after "../"
		base := filepath.Base(componentPath)
		dir := filepath.Dir(strings.TrimPrefix(componentPath, "../"))
		return "../" + filepath.Join(dir, base)
	}
	
	// For paths in .pluqqy directory, make them relative
	if strings.Contains(componentPath, ".pluqqy/components/") {
		// Extract the component-relative part
		idx := strings.Index(componentPath, "components/")
		if idx >= 0 {
			return "../" + componentPath[idx:]
		}
	}
	
	// Default to relative path with "../"
	return "../" + componentPath
}

// RemoveComponentReferences removes all references to a deleted component from pipelines
func RemoveComponentReferences(componentPath string) error {
	// Normalize path for comparison
	componentPath = filepath.Clean(componentPath)
	
	// Track updated pipelines for potential rollback
	updatedPipelines := []struct {
		path     string
		original []byte
	}{}
	
	// Update active pipelines
	activePipelines, err := ListPipelines()
	if err != nil {
		return fmt.Errorf("failed to list active pipelines: %w", err)
	}
	
	for _, pipelineName := range activePipelines {
		// Read the pipeline
		pipeline, err := ReadPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		modified := false
		newComponents := []models.ComponentRef{}
		
		for _, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Keep components that don't match the deleted one
			if compPath != componentPath {
				newComponents = append(newComponents, comp)
			} else {
				modified = true
			}
		}
		
		// Save if modified
		if modified {
			pipelinePath := filepath.Join(PluqqyDir, PipelinesDir, pipelineName)
			
			// Store original for rollback
			original, _ := readRawFile(pipelinePath)
			updatedPipelines = append(updatedPipelines, struct {
				path     string
				original []byte
			}{pipelinePath, original})
			
			// Update pipeline components
			pipeline.Components = newComponents
			
			// If pipeline has no components left, write directly without validation
			if len(newComponents) == 0 {
				// Marshal and write directly to bypass validation
				pipelineData, err := yaml.Marshal(pipeline)
				if err != nil {
					rollbackPipelines(updatedPipelines)
					return fmt.Errorf("failed to marshal pipeline %s: %w", pipelineName, err)
				}
				
				if err := os.WriteFile(pipelinePath, pipelineData, 0644); err != nil {
					rollbackPipelines(updatedPipelines)
					return fmt.Errorf("failed to update pipeline %s: %w", pipelineName, err)
				}
			} else {
				// Write updated pipeline normally with validation
				if err := WritePipeline(pipeline); err != nil {
					// Rollback all changes
					rollbackPipelines(updatedPipelines)
					return fmt.Errorf("failed to update pipeline %s: %w", pipelineName, err)
				}
			}
		}
	}
	
	// Update archived pipelines
	archivedPipelines, err := ListArchivedPipelines()
	if err != nil {
		// Don't fail if we can't access archived pipelines
		return nil
	}
	
	for _, pipelineName := range archivedPipelines {
		// Read the archived pipeline
		pipeline, err := ReadArchivedPipeline(pipelineName)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		// Check if this pipeline references the component
		modified := false
		newComponents := []models.ComponentRef{}
		
		for _, comp := range pipeline.Components {
			// Remove "../" prefix for comparison
			compPath := strings.TrimPrefix(comp.Path, "../")
			compPath = filepath.Clean(compPath)
			
			// Keep components that don't match the deleted one
			if compPath != componentPath {
				newComponents = append(newComponents, comp)
			} else {
				modified = true
			}
		}
		
		// Save if modified
		if modified {
			pipelinePath := filepath.Join(PluqqyDir, ArchiveDir, PipelinesDir, pipelineName)
			
			// Store original for rollback
			original, _ := readRawFile(pipelinePath)
			updatedPipelines = append(updatedPipelines, struct {
				path     string
				original []byte
			}{pipelinePath, original})
			
			// Update pipeline components
			pipeline.Components = newComponents
			
			// Write updated pipeline directly to archive
			pipelineData, err := yaml.Marshal(pipeline)
			if err != nil {
				rollbackPipelines(updatedPipelines)
				return fmt.Errorf("failed to marshal archived pipeline %s: %w", pipelineName, err)
			}
			
			if err := os.WriteFile(pipelinePath, pipelineData, 0644); err != nil {
				// Rollback all changes
				rollbackPipelines(updatedPipelines)
				return fmt.Errorf("failed to update archived pipeline %s: %w", pipelineName, err)
			}
		}
	}
	
	return nil
}

// rollbackPipelines restores pipelines to their original state
func rollbackPipelines(updates []struct {
	path     string
	original []byte
}) {
	for _, update := range updates {
		if update.original != nil {
			writeRawFile(update.path, update.original)
		}
	}
}

// Helper functions for raw file operations (for backup/restore)
func readRawFile(path string) ([]byte, error) {
	return ReadFileContent(path)
}

func writeRawFile(path string, content []byte) error {
	return WriteFileContent(path, content)
}