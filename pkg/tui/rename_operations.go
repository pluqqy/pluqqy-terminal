package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

// RenameOperator handles business logic for rename operations
type RenameOperator struct {
	// We might use confirmation dialog in the future
	confirmDialog *ConfirmationModel
}

// NewRenameOperator creates a new rename operator instance
func NewRenameOperator() *RenameOperator {
	return &RenameOperator{
		confirmDialog: NewConfirmation(),
	}
}

// ValidateDisplayName validates a display name for renaming
func (ro *RenameOperator) ValidateDisplayName(newName, itemType string) error {
	// Check for empty name
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Check length constraints
	if len(newName) > 100 {
		return fmt.Errorf("name is too long (max 100 characters)")
	}

	// The files package will handle more detailed validation
	return nil
}

// FindAffectedPipelines finds pipelines that reference a component
func (ro *RenameOperator) FindAffectedPipelines(componentPath string) (activeNames []string, archivedNames []string, err error) {
	// Delegate to files package
	return files.FindAffectedPipelines(componentPath)
}

// ExecuteRename performs the rename operation
func (ro *RenameOperator) ExecuteRename(oldPath, newDisplayName, itemType string, isArchived bool) tea.Cmd {
	return func() tea.Msg {
		var err error

		// Perform the rename based on item type
		if isArchived {
			if itemType == "component" {
				err = files.RenameComponentInArchive(oldPath, newDisplayName)
			} else {
				err = files.RenamePipelineInArchive(oldPath, newDisplayName)
			}
		} else {
			if itemType == "component" {
				err = files.RenameComponent(oldPath, newDisplayName)
			} else {
				err = files.RenamePipeline(oldPath, newDisplayName)
			}
		}

		if err != nil {
			return RenameErrorMsg{Error: err}
		}

		// Generate success message
		return RenameSuccessMsg{
			ItemType:   itemType,
			OldName:    filepath.Base(oldPath),
			NewName:    newDisplayName,
			IsArchived: isArchived,
		}
	}
}

// PrepareRenameComponent prepares a component for renaming
func (ro *RenameOperator) PrepareRenameComponent(item componentItem) (displayName, path string, isArchived bool) {
	// Use the component's display name directly (it's already the display name from markdown)
	displayName = item.name
	path = item.path
	isArchived = item.isArchived
	return
}

// PrepareRenamePipeline prepares a pipeline for renaming
func (ro *RenameOperator) PrepareRenamePipeline(item pipelineItem) (displayName, path string, isArchived bool) {
	// Use the pipeline's name directly (it's already the display name)
	displayName = item.name
	path = item.path
	isArchived = item.isArchived
	return
}

// extractDisplayNameFromComponent extracts a display name from a component item
func (ro *RenameOperator) extractDisplayNameFromComponent(componentName string) string {
	// If the name contains a path separator, extract the filename
	if strings.Contains(componentName, "/") {
		parts := strings.Split(componentName, "/")
		componentName = parts[len(parts)-1]
	}

	// Remove file extension if present
	if strings.HasSuffix(componentName, ".md") {
		componentName = strings.TrimSuffix(componentName, ".md")
	}

	// Use the files package's ExtractDisplayName for consistent naming
	return files.ExtractDisplayName(componentName)
}

// GetSuccessMessage generates a success message for the rename operation
func (ro *RenameOperator) GetSuccessMessage(msg RenameSuccessMsg) string {
	itemType := msg.ItemType
	if msg.IsArchived {
		itemType = "archived " + itemType
	}

	return fmt.Sprintf("✓ Renamed %s to: %s", itemType, msg.NewName)
}

// GetErrorMessage generates an error message for a failed rename
func (ro *RenameOperator) GetErrorMessage(err error) string {
	return fmt.Sprintf("✗ Rename failed: %v", err)
}

// HandleInput processes keyboard input for rename mode
func (ro *RenameOperator) HandleInput(msg tea.KeyMsg, state *RenameState) (handled bool, err error) {
	// Delegate to the state's HandleInput method
	handled, cmd := state.HandleInput(msg)

	// If a command was returned, we can't execute it here
	// Instead we return handled status
	_ = cmd // Ignore the command for now

	return handled, nil
}
