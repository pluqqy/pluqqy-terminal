package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// CloneOperator handles business logic for cloning operations
type CloneOperator struct{}

// NewCloneOperator creates a new clone operator instance
func NewCloneOperator() *CloneOperator {
	return &CloneOperator{}
}

// PrepareCloneComponent prepares a component for cloning
func (co *CloneOperator) PrepareCloneComponent(comp componentItem) (displayName, path string, isArchived bool) {
	return comp.name, comp.path, comp.isArchived
}

// PrepareClonePipeline prepares a pipeline for cloning
func (co *CloneOperator) PrepareClonePipeline(pipeline pipelineItem) (displayName, path string, isArchived bool) {
	return pipeline.name, pipeline.path, pipeline.isArchived
}

// ValidateComponentClone validates if a component can be cloned with the given name
func (co *CloneOperator) ValidateComponentClone(originalPath, newName string, toArchive bool) error {
	// Check for empty name
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Generate the target filename
	slugifiedName := files.Slugify(newName)

	// Determine the component type from the path
	var componentType string
	if strings.Contains(originalPath, "/prompts/") {
		componentType = "prompts"
	} else if strings.Contains(originalPath, "/contexts/") {
		componentType = "contexts"
	} else if strings.Contains(originalPath, "/rules/") {
		componentType = "rules"
	} else {
		return fmt.Errorf("unknown component type")
	}

	// Build the target path
	targetFilename := slugifiedName + ".md"
	var targetPath string

	if toArchive {
		targetPath = filepath.Join(files.PluqqyDir, "archive", files.ComponentsDir, componentType, targetFilename)
	} else {
		targetPath = filepath.Join(files.PluqqyDir, files.ComponentsDir, componentType, targetFilename)
	}

	// Check if file already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("component '%s' already exists", newName)
	}

	return nil
}

// ValidatePipelineClone validates if a pipeline can be cloned with the given name
func (co *CloneOperator) ValidatePipelineClone(newName string, toArchive bool) error {
	// Check for empty name
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Generate the target filename
	slugifiedName := files.Slugify(newName)
	targetFilename := slugifiedName + ".yaml"

	// Build the target path
	var targetPath string
	if toArchive {
		targetPath = filepath.Join(files.PluqqyDir, "archive", files.PipelinesDir, targetFilename)
	} else {
		targetPath = filepath.Join(files.PluqqyDir, files.PipelinesDir, targetFilename)
	}

	// Check if file already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("pipeline '%s' already exists", newName)
	}

	return nil
}

// CloneComponent performs the actual component cloning
func (co *CloneOperator) CloneComponent(originalPath, newName string, fromArchive, toArchive bool) error {
	// Read the original component
	var content *models.Component
	var err error

	if fromArchive {
		content, err = files.ReadArchivedComponent(originalPath)
	} else {
		content, err = files.ReadComponent(originalPath)
	}

	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}

	// Update the display name in the content
	content.Name = newName

	// Determine the component type from the path
	var componentType string
	if strings.Contains(originalPath, "/prompts/") {
		componentType = "prompts"
	} else if strings.Contains(originalPath, "/contexts/") {
		componentType = "contexts"
	} else if strings.Contains(originalPath, "/rules/") {
		componentType = "rules"
	} else {
		return fmt.Errorf("unknown component type")
	}

	// Generate the new filename
	newFilename := files.Slugify(newName) + ".md"

	// Build the target path
	var targetPath string
	if toArchive {
		// Clone to archive: .pluqqy/archive/components/type/filename.md
		targetPath = filepath.Join(files.ComponentsDir, componentType, newFilename)
		// WriteComponentWithNameAndTags will handle the archive path
	} else {
		targetPath = filepath.Join(files.ComponentsDir, componentType, newFilename)
	}

	// Write the component to the target location with new name and existing tags
	if toArchive {
		// Write directly to archive
		fullContent := content.Content
		if content.Name != "" || len(content.Tags) > 0 {
			// Add YAML frontmatter if we have metadata
			frontmatter := "---\n"
			if content.Name != "" {
				frontmatter += fmt.Sprintf("name: %s\n", newName)
			}
			if len(content.Tags) > 0 {
				frontmatter += "tags:\n"
				for _, tag := range content.Tags {
					frontmatter += fmt.Sprintf("  - %s\n", tag)
				}
			}
			frontmatter += "---\n\n"
			fullContent = frontmatter + content.Content
		}
		err = files.WriteComponentToArchive(targetPath, fullContent)
	} else {
		err = files.WriteComponentWithNameAndTags(targetPath, content.Content, newName, content.Tags)
	}
	if err != nil {
		return fmt.Errorf("failed to write cloned component: %w", err)
	}

	return nil
}

// ClonePipeline performs the actual pipeline cloning
func (co *CloneOperator) ClonePipeline(originalPath, newName string, fromArchive, toArchive bool) error {
	// Read the original pipeline
	var pipeline *models.Pipeline
	var err error

	if fromArchive {
		pipeline, err = files.ReadArchivedPipeline(originalPath)
	} else {
		pipeline, err = files.ReadPipeline(originalPath)
	}

	if err != nil {
		return fmt.Errorf("failed to read pipeline: %w", err)
	}

	// Update the pipeline name
	pipeline.Name = newName

	// Generate the new filename
	newFilename := files.Slugify(newName) + ".yaml"

	// Build the target path - just the filename for pipelines
	targetPath := newFilename

	// Create new pipeline with updated name and path
	newPipeline := &models.Pipeline{
		Name:       newName,
		Path:       targetPath,
		Components: pipeline.Components,
		OutputPath: pipeline.OutputPath,
		Tags:       pipeline.Tags,
	}

	// Write the pipeline
	if toArchive {
		err = files.WritePipelineToArchive(newPipeline)
	} else {
		err = files.WritePipeline(newPipeline)
	}
	if err != nil {
		return fmt.Errorf("failed to write cloned pipeline: %w", err)
	}

	return nil
}

// GetCloneSuggestion generates a suggested name for cloning
func (co *CloneOperator) GetCloneSuggestion(originalName string) string {
	// Check if the name already has a copy prefix
	if strings.HasPrefix(originalName, "(Copy") {
		// Check for pattern like "(Copy) Name" or "(Copy 2) Name"
		if strings.HasPrefix(originalName, "(Copy) ") {
			// It's "(Copy) Name", make it "(Copy 2) Name"
			baseName := strings.TrimPrefix(originalName, "(Copy) ")
			return fmt.Sprintf("(Copy 2) %s", baseName)
		} else if idx := strings.Index(originalName, ") "); idx > 0 && idx < len(originalName)-2 {
			// Check if it's "(Copy N) Name" format
			prefix := originalName[:idx+1]
			if strings.HasPrefix(prefix, "(Copy ") {
				// Extract the number
				numStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(prefix, ")"), "(Copy"))
				var num int
				if _, err := fmt.Sscanf(numStr, "%d", &num); err == nil {
					baseName := originalName[idx+2:] // Skip ") "
					return fmt.Sprintf("(Copy %d) %s", num+1, baseName)
				}
			}
		}
	}

	// Default case: just prepend "(Copy) "
	return fmt.Sprintf("(Copy) %s", originalName)
}

// CanCloneToLocation checks if an item can be cloned to a specific location
func (co *CloneOperator) CanCloneToLocation(itemType string, fromArchive, toArchive bool) (bool, string) {
	// For now, all combinations are allowed
	// This method exists for future restrictions if needed

	if fromArchive && toArchive {
		return true, "Cloning within archive"
	} else if fromArchive && !toArchive {
		return true, "Restoring from archive as copy"
	} else if !fromArchive && toArchive {
		return true, "Cloning to archive"
	} else {
		return true, "Cloning within active"
	}
}
