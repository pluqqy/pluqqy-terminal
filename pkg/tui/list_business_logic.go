package tui

import (
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// BusinessLogic handles business logic operations for the MainListModel
type BusinessLogic struct {
	prompts  []componentItem
	contexts []componentItem
	rules    []componentItem
}

// NewBusinessLogic creates a new BusinessLogic instance
func NewBusinessLogic() *BusinessLogic {
	return &BusinessLogic{}
}

// SetComponents updates the component collections
func (b *BusinessLogic) SetComponents(prompts, contexts, rules []componentItem) {
	b.prompts = prompts
	b.contexts = contexts
	b.rules = rules
}

// GetAllComponents returns all components ordered by settings configuration
func (b *BusinessLogic) GetAllComponents() []componentItem {
	// Load settings for section order
	settings, err := files.ReadSettings()
	if err != nil || settings == nil {
		settings = models.DefaultSettings()
	}
	
	// Group components by type
	typeGroups := make(map[string][]componentItem)
	typeGroups[models.ComponentTypeContext] = b.contexts
	typeGroups[models.ComponentTypePrompt] = b.prompts
	typeGroups[models.ComponentTypeRules] = b.rules
	
	// Build ordered list based on sections
	var all []componentItem
	for _, section := range settings.Output.Formatting.Sections {
		if components, exists := typeGroups[section.Type]; exists {
			all = append(all, components...)
		}
	}
	
	return all
}

// GetEditingItemName returns the name of the item being edited
func GetEditingItemName(tagEditor *TagEditor, stateManager *StateManager, components []componentItem, pipelines []pipelineItem) string {
	if tagEditor.ItemType == "component" {
		if stateManager.ComponentCursor >= 0 && stateManager.ComponentCursor < len(components) {
			return components[stateManager.ComponentCursor].name
		}
	} else {
		if stateManager.PipelineCursor >= 0 && stateManager.PipelineCursor < len(pipelines) {
			return pipelines[stateManager.PipelineCursor].name
		}
	}
	return ""
}