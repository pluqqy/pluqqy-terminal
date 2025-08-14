package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// CloneState manages the state for cloning components and pipelines
type CloneState struct {
	mu              sync.RWMutex // Protects concurrent access
	Active          bool         // Whether clone mode is active
	ItemType        string       // "component" or "pipeline"
	OriginalName    string       // Display name of item to clone
	OriginalPath    string       // File path of item to clone
	NewName         string       // User input for new display name
	ValidationError string       // Real-time validation error
	IsArchived      bool         // Whether the item being cloned is archived
	CloneToArchive  bool         // Whether to clone to archive folder
	AffectedActive  []string     // Active pipelines that will reference cloned component
	AffectedArchive []string     // Archived pipelines that will reference cloned component
}

// NewCloneState creates a new clone state instance
func NewCloneState() *CloneState {
	return &CloneState{}
}

// HandleInput processes keyboard input for clone mode
func (cs *CloneState) HandleInput(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if !cs.Active {
		return false, nil
	}

	switch msg.String() {
	case "esc":
		// Reset without re-acquiring lock
		cs.resetInternal()
		return true, nil

	case "enter":
		if cs.NewName != "" && cs.ValidationError == "" {
			// Trigger clone operation
			return true, cs.executeClone()
		}
		return true, nil

	case "backspace":
		if len(cs.NewName) > 0 {
			// Handle UTF-8 properly
			runes := []rune(cs.NewName)
			cs.NewName = string(runes[:len(runes)-1])
			cs.validate()
		}
		return true, nil

	case " ":
		cs.NewName += " "
		cs.validate()
		return true, nil

	case "tab":
		// Toggle archive destination
		cs.CloneToArchive = !cs.CloneToArchive
		// Re-validate since destination changed
		cs.validate()
		return true, nil

	default:
		if msg.Type == tea.KeyRunes {
			cs.NewName += string(msg.Runes)
			cs.validate()
			return true, nil
		}
	}

	return false, nil
}

// Start initiates clone mode for an item
func (cs *CloneState) Start(displayName, itemType, path string, isArchived bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.Active = true
	cs.ItemType = itemType
	cs.OriginalName = displayName
	cs.OriginalPath = path
	cs.ValidationError = ""
	cs.IsArchived = isArchived
	cs.CloneToArchive = isArchived // Default to same location as original
	cs.AffectedActive = nil
	cs.AffectedArchive = nil

	// Generate a collision-free name automatically
	cs.NewName = cs.generateUniqueNameInternal(displayName)
}

// generateUniqueNameInternal generates a collision-free name for cloning
// Must be called with lock held
func (cs *CloneState) generateUniqueNameInternal(baseName string) string {
	// Extract the base name without any existing (Copy) prefix
	cleanBase := baseName
	
	// Check if the name already has a (Copy) or (Copy N) prefix
	if strings.HasPrefix(baseName, "(Copy") {
		if strings.HasPrefix(baseName, "(Copy) ") {
			// Remove "(Copy) " prefix
			cleanBase = strings.TrimPrefix(baseName, "(Copy) ")
		} else if idx := strings.Index(baseName, ") "); idx > 0 && idx < len(baseName)-2 {
			// Check if it's "(Copy N) " format
			prefix := baseName[:idx+1]
			if strings.HasPrefix(prefix, "(Copy ") {
				// Remove "(Copy N) " prefix
				cleanBase = baseName[idx+2:]
			}
		}
	}
	
	// Start with the base suggestion
	candidateName := fmt.Sprintf("(Copy) %s", cleanBase)
	counter := 2
	
	// Keep trying until we find a name that doesn't exist
	for {
		// Check if this name would work
		if !cs.nameExistsInternal(candidateName) {
			return candidateName
		}
		
		// Try the next number
		candidateName = fmt.Sprintf("(Copy %d) %s", counter, cleanBase)
		counter++
		
		// Safety check to prevent infinite loop (stop at 1000)
		if counter > 1000 {
			return fmt.Sprintf("(Copy %d) %s", counter, cleanBase)
		}
	}
}

// nameExistsInternal checks if a name already exists
// Must be called with lock held
func (cs *CloneState) nameExistsInternal(name string) bool {
	// Generate the target filename
	slugifiedName := files.Slugify(name)
	
	if cs.ItemType == "component" {
		// Determine the component type from the path
		var componentType string
		if strings.Contains(cs.OriginalPath, "/prompts/") {
			componentType = "prompts"
		} else if strings.Contains(cs.OriginalPath, "/contexts/") {
			componentType = "contexts"
		} else if strings.Contains(cs.OriginalPath, "/rules/") {
			componentType = "rules"
		} else {
			return false
		}
		
		// Build the target path
		targetFilename := slugifiedName + ".md"
		var targetPath string
		
		if cs.CloneToArchive {
			targetPath = filepath.Join(files.PluqqyDir, "components", componentType, ".archive", targetFilename)
		} else {
			targetPath = filepath.Join(files.PluqqyDir, "components", componentType, targetFilename)
		}
		
		// Check if file exists
		_, err := os.Stat(targetPath)
		return err == nil
	} else if cs.ItemType == "pipeline" {
		// Build the target path
		targetFilename := slugifiedName + ".yaml"
		var targetPath string
		
		if cs.CloneToArchive {
			targetPath = filepath.Join(files.PluqqyDir, "pipelines", ".archive", targetFilename)
		} else {
			targetPath = filepath.Join(files.PluqqyDir, "pipelines", targetFilename)
		}
		
		// Check if file exists
		_, err := os.Stat(targetPath)
		return err == nil
	}
	
	return false
}

// validate checks if the new name is valid
// Must be called with lock held
func (cs *CloneState) validate() {
	// Clear previous error
	cs.ValidationError = ""

	// Check for empty name
	if strings.TrimSpace(cs.NewName) == "" {
		cs.ValidationError = "Name cannot be empty"
		return
	}

	// Check if name is the same as original
	if cs.NewName == cs.OriginalName && cs.CloneToArchive == cs.IsArchived {
		cs.ValidationError = "Name must be different from original when cloning to same location"
		return
	}

	// Validate the clone operation
	err := cs.validateClone()
	if err != nil {
		cs.ValidationError = err.Error()
	}
}

// validateClone checks if the clone operation is valid
func (cs *CloneState) validateClone() error {
	// Generate the target filename
	slugifiedName := files.Slugify(cs.NewName)
	
	// Check if file already exists
	if cs.ItemType == "component" {
		// Determine the component type from the path
		var componentType string
		if strings.Contains(cs.OriginalPath, "/prompts/") {
			componentType = "prompts"
		} else if strings.Contains(cs.OriginalPath, "/contexts/") {
			componentType = "contexts"
		} else if strings.Contains(cs.OriginalPath, "/rules/") {
			componentType = "rules"
		} else {
			return fmt.Errorf("unknown component type")
		}
		
		// Build the target path
		targetFilename := slugifiedName + ".md"
		var targetPath string
		
		if cs.CloneToArchive {
			targetPath = filepath.Join(files.PluqqyDir, "components", componentType, ".archive", targetFilename)
		} else {
			targetPath = filepath.Join(files.PluqqyDir, "components", componentType, targetFilename)
		}
		
		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("Component '%s' already exists", cs.NewName)
		}
	} else if cs.ItemType == "pipeline" {
		// Build the target path
		targetFilename := slugifiedName + ".yaml"
		var targetPath string
		
		if cs.CloneToArchive {
			targetPath = filepath.Join(files.PluqqyDir, "pipelines", ".archive", targetFilename)
		} else {
			targetPath = filepath.Join(files.PluqqyDir, "pipelines", targetFilename)
		}
		
		// Check if file exists
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("Pipeline '%s' already exists", cs.NewName)
		}
	}

	return nil
}

// resetInternal clears the clone state (must be called with lock held)
func (cs *CloneState) resetInternal() {
	cs.Active = false
	cs.ItemType = ""
	cs.OriginalName = ""
	cs.OriginalPath = ""
	cs.NewName = ""
	cs.ValidationError = ""
	cs.IsArchived = false
	cs.CloneToArchive = false
	cs.AffectedActive = nil
	cs.AffectedArchive = nil
}

// Reset clears the clone state
func (cs *CloneState) Reset() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.resetInternal()
}

// executeClone performs the actual clone operation
func (cs *CloneState) executeClone() tea.Cmd {
	return func() tea.Msg {
		var err error

		// Perform the clone based on item type and destination
		if cs.ItemType == "component" {
			err = cs.cloneComponent()
		} else if cs.ItemType == "pipeline" {
			err = cs.clonePipeline()
		}

		if err != nil {
			return CloneErrorMsg{Error: err}
		}

		return CloneSuccessMsg{
			ItemType:       cs.ItemType,
			OriginalName:   cs.OriginalName,
			NewName:        cs.NewName,
			ClonedToArchive: cs.CloneToArchive,
		}
	}
}

// cloneComponent handles component cloning logic
func (cs *CloneState) cloneComponent() error {
	// Read the original component
	var content *models.Component
	var err error
	
	if cs.IsArchived {
		content, err = files.ReadArchivedComponent(cs.OriginalPath)
	} else {
		content, err = files.ReadComponent(cs.OriginalPath)
	}
	
	if err != nil {
		return fmt.Errorf("failed to read component: %w", err)
	}
	
	// Determine the component type from the path
	var componentType string
	if strings.Contains(cs.OriginalPath, "/prompts/") {
		componentType = "prompts"
	} else if strings.Contains(cs.OriginalPath, "/contexts/") {
		componentType = "contexts"
	} else if strings.Contains(cs.OriginalPath, "/rules/") {
		componentType = "rules"
	} else {
		return fmt.Errorf("unknown component type")
	}

	// Generate the new filename and path
	newFilename := files.Slugify(cs.NewName) + ".md"
	var targetPath string
	
	if cs.CloneToArchive {
		targetPath = fmt.Sprintf("components/%s/.archive/%s", componentType, newFilename)
		// Ensure archive directory exists
		archiveDir := filepath.Join(files.PluqqyDir, "components", componentType, ".archive")
		if err := os.MkdirAll(archiveDir, 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}
	} else {
		targetPath = fmt.Sprintf("components/%s/%s", componentType, newFilename)
	}
	
	// Write the component with new name and existing tags
	err = files.WriteComponentWithNameAndTags(targetPath, content.Content, cs.NewName, content.Tags)
	if err != nil {
		return fmt.Errorf("failed to write cloned component: %w", err)
	}
	
	return nil
}

// clonePipeline handles pipeline cloning logic
func (cs *CloneState) clonePipeline() error {
	// Read the original pipeline
	var pipeline *models.Pipeline
	var err error
	
	if cs.IsArchived {
		pipeline, err = files.ReadArchivedPipeline(cs.OriginalPath)
	} else {
		pipeline, err = files.ReadPipeline(cs.OriginalPath)
	}
	
	if err != nil {
		return fmt.Errorf("failed to read pipeline: %w", err)
	}

	// Create a new pipeline with updated name
	newPipeline := &models.Pipeline{
		Name:       cs.NewName,
		Tags:       pipeline.Tags,
		Components: pipeline.Components,
		OutputPath: pipeline.OutputPath,
	}
	
	// Generate the new filename
	newFilename := files.Slugify(cs.NewName) + ".yaml"
	
	// Set the appropriate path based on destination
	if cs.CloneToArchive {
		newPipeline.Path = fmt.Sprintf("pipelines/.archive/%s", newFilename)
		// Ensure archive directory exists
		archiveDir := filepath.Join(files.PluqqyDir, "pipelines", ".archive")
		if err := os.MkdirAll(archiveDir, 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}
	} else {
		newPipeline.Path = newFilename // WritePipeline expects just the filename for active pipelines
	}
	
	// Write the pipeline
	err = files.WritePipeline(newPipeline)
	if err != nil {
		return fmt.Errorf("failed to write cloned pipeline: %w", err)
	}
	
	return nil
}

// GetSlugifiedName returns what the filename will become
func (cs *CloneState) GetSlugifiedName() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	if cs.NewName == "" {
		return ""
	}
	return files.Slugify(cs.NewName)
}

// IsValid returns true if the current name is valid for cloning
func (cs *CloneState) IsValid() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return cs.NewName != "" && 
	       cs.ValidationError == ""
}

// IsActive returns whether clone mode is active
func (cs *CloneState) IsActive() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return cs.Active
}

// GetError returns the validation error if any
func (cs *CloneState) GetError() error {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	if cs.ValidationError != "" {
		return fmt.Errorf("%s", cs.ValidationError)
	}
	return nil
}

// GetItemType returns the type of item being cloned
func (cs *CloneState) GetItemType() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return cs.ItemType
}

// GetNewName returns the new name entered by the user
func (cs *CloneState) GetNewName() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return cs.NewName
}

// CloneSuccessMsg is sent when a clone operation succeeds
type CloneSuccessMsg struct {
	ItemType        string
	OriginalName    string
	NewName         string
	ClonedToArchive bool
}

// CloneErrorMsg is sent when a clone operation fails
type CloneErrorMsg struct {
	Error error
}