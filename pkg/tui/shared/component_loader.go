package shared

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
	"github.com/pluqqy/pluqqy-cli/pkg/utils"
)

// ComponentItem represents a component in the TUI views
type ComponentItem struct {
	Name         string
	Path         string
	CompType     string
	LastModified time.Time
	UsageCount   int
	TokenCount   int
	Tags         []string
	IsArchived   bool
}

// ComponentLoader handles loading components and pipelines from the filesystem
type ComponentLoader struct {
	projectPath string
}

// NewComponentLoader creates a new ComponentLoader
func NewComponentLoader(projectPath string) *ComponentLoader {
	return &ComponentLoader{
		projectPath: projectPath,
	}
}

// LoadComponents loads all components (prompts, contexts, rules) from the project
func (cl *ComponentLoader) LoadComponents(includeArchived bool) ([]ComponentItem, []ComponentItem, []ComponentItem, error) {
	// Get usage counts for all components
	usageMap, _ := files.CountComponentUsage()

	// Load each type of component
	prompts := cl.loadComponentsOfType("prompts", files.PromptsDir, models.ComponentTypePrompt, usageMap, includeArchived)
	contexts := cl.loadComponentsOfType("contexts", files.ContextsDir, models.ComponentTypeContext, usageMap, includeArchived)
	rules := cl.loadComponentsOfType("rules", files.RulesDir, models.ComponentTypeRules, usageMap, includeArchived)

	return prompts, contexts, rules, nil
}

// loadComponentsOfType loads components of a specific type
func (cl *ComponentLoader) loadComponentsOfType(compType, subDir, modelType string, usageMap map[string]int, includeArchived bool) []ComponentItem {
	var items []ComponentItem

	// Load active components
	components, _ := files.ListComponents(compType)
	for _, c := range components {
		componentPath := filepath.Join(files.ComponentsDir, subDir, c)
		modTime, _ := files.GetComponentStats(componentPath)

		// Calculate usage count
		usage := 0
		relativePath := "../" + componentPath
		if count, exists := usageMap[relativePath]; exists {
			usage = count
		}

		// Read component content for token estimation and display name
		component, _ := files.ReadComponent(componentPath)
		tokenCount := 0
		displayName := c // Default to filename
		if component != nil {
			tokenCount = utils.EstimateTokens(component.Content)
			// Use display name from component (from frontmatter or filename)
			if component.Name != "" {
				displayName = component.Name
			}
		}

		tags := []string{}
		if component != nil {
			tags = component.Tags
		}

		items = append(items, ComponentItem{
			Name:         displayName,
			Path:         componentPath,
			CompType:     modelType,
			LastModified: modTime,
			UsageCount:   usage,
			TokenCount:   tokenCount,
			Tags:         tags,
			IsArchived:   false,
		})
	}

	// Load archived components if needed
	if includeArchived {
		archivedComponents, _ := files.ListArchivedComponents(compType)
		for _, c := range archivedComponents {
			componentPath := filepath.Join(files.ComponentsDir, subDir, c)

			// Read archived component
			component, _ := files.ReadArchivedComponent(componentPath)
			modTime := time.Time{}
			displayName := c // Default to filename
			if component != nil {
				modTime = component.Modified
				// Use display name from component (from frontmatter or filename)
				if component.Name != "" {
					displayName = component.Name
				}
			}

			// Calculate usage count (archived components typically have 0 usage)
			usage := 0

			// Get token count
			tokenCount := 0
			if component != nil {
				tokenCount = utils.EstimateTokens(component.Content)
			}

			tags := []string{}
			if component != nil {
				tags = component.Tags
			}

			items = append(items, ComponentItem{
				Name:         displayName,
				Path:         componentPath,
				CompType:     modelType,
				LastModified: modTime,
				UsageCount:   usage,
				TokenCount:   tokenCount,
				Tags:         tags,
				IsArchived:   true,
			})
		}
	}

	return items
}

// ShouldIncludeArchived checks if the current search query requires archived items
func ShouldIncludeArchived(searchQuery string) bool {
	if searchQuery == "" {
		return false
	}

	// Parse the search query to check for status:archived
	parser := search.NewParser()
	query, err := parser.Parse(searchQuery)
	if err != nil {
		return false
	}

	for _, condition := range query.Conditions {
		if condition.Field == search.FieldStatus {
			if statusStr, ok := condition.Value.(string); ok && strings.ToLower(statusStr) == "archived" {
				return true
			}
		}
	}

	return false
}