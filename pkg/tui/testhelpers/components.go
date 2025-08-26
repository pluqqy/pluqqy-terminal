package testhelpers

import (
	"time"
	
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComponentItem represents a component in the TUI lists
// This mirrors the componentItem struct in the tui package
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

// ComponentBuilder provides a fluent interface for building test components
type ComponentBuilder struct {
	item ComponentItem
}

// NewComponentBuilder creates a new component builder with default values
func NewComponentBuilder(name string) *ComponentBuilder {
	return &ComponentBuilder{
		item: ComponentItem{
			Name:         name,
			Path:         "test/components/prompts/" + name + ".md",
			CompType:     models.ComponentTypePrompt,
			LastModified: time.Now(),
			TokenCount:   100,
			Tags:         []string{},
			UsageCount:   0,
			IsArchived:   false,
		},
	}
}

// WithType sets the component type and adjusts the path accordingly
func (b *ComponentBuilder) WithType(compType string) *ComponentBuilder {
	b.item.CompType = compType
	b.item.Path = "test/components/" + compType + "/" + b.item.Name + ".md"
	return b
}

// WithPath sets a custom path for the component
func (b *ComponentBuilder) WithPath(path string) *ComponentBuilder {
	b.item.Path = path
	return b
}

// WithTokens sets the token count for the component
func (b *ComponentBuilder) WithTokens(count int) *ComponentBuilder {
	b.item.TokenCount = count
	return b
}

// WithTags sets the tags for the component
func (b *ComponentBuilder) WithTags(tags ...string) *ComponentBuilder {
	b.item.Tags = tags
	return b
}

// WithUsageCount sets the usage count for the component
func (b *ComponentBuilder) WithUsageCount(count int) *ComponentBuilder {
	b.item.UsageCount = count
	return b
}

// WithLastModified sets a specific last modified time
func (b *ComponentBuilder) WithLastModified(t time.Time) *ComponentBuilder {
	b.item.LastModified = t
	return b
}

// Archived marks the component as archived
func (b *ComponentBuilder) Archived() *ComponentBuilder {
	b.item.IsArchived = true
	return b
}

// Build returns the built component item
func (b *ComponentBuilder) Build() ComponentItem {
	return b.item
}

// Factory Functions for Common Cases

// MakePromptComponent creates a simple prompt component
func MakePromptComponent(name string, tokens int) ComponentItem {
	return NewComponentBuilder(name).
		WithType(models.ComponentTypePrompt).
		WithTokens(tokens).
		Build()
}

// MakeContextComponent creates a simple context component
func MakeContextComponent(name string, tokens int) ComponentItem {
	return NewComponentBuilder(name).
		WithType(models.ComponentTypeContext).
		WithTokens(tokens).
		Build()
}

// MakeRulesComponent creates a simple rules component
func MakeRulesComponent(name string, tokens int) ComponentItem {
	return NewComponentBuilder(name).
		WithType(models.ComponentTypeRules).
		WithTokens(tokens).
		Build()
}

// MakeArchivedComponent creates an archived component
func MakeArchivedComponent(name string, compType string) ComponentItem {
	return NewComponentBuilder(name).
		WithType(compType).
		Archived().
		Build()
}

// MakeTestComponent creates a component with minimal configuration
// This is for backward compatibility with existing tests
func MakeTestComponent(name, compType string) ComponentItem {
	return NewComponentBuilder(name).
		WithType(compType).
		Build()
}

// MakeTestComponents creates multiple components of the same type
func MakeTestComponents(compType string, names ...string) []ComponentItem {
	components := make([]ComponentItem, len(names))
	for i, name := range names {
		components[i] = MakeTestComponent(name, compType)
	}
	return components
}

// MakeTestComponentWithTags creates a component with tags
func MakeTestComponentWithTags(name, compType string, tags []string) ComponentItem {
	return NewComponentBuilder(name).
		WithType(compType).
		WithTags(tags...).
		Build()
}

// MakeTestComponentWithTokens creates a component with specific token count
func MakeTestComponentWithTokens(name, compType string, tokenCount int, tags []string) ComponentItem {
	return NewComponentBuilder(name).
		WithType(compType).
		WithTokens(tokenCount).
		WithTags(tags...).
		Build()
}

// GenerateComponents creates multiple test components with incrementing values
func GenerateComponents(count int, compType string) []ComponentItem {
	components := make([]ComponentItem, count)
	baseTime := time.Now()
	
	for i := 0; i < count; i++ {
		components[i] = NewComponentBuilder(compType + "-" + string(rune('a'+i))).
			WithType(compType).
			WithTokens(100 * (i + 1)).
			WithUsageCount(i).
			WithLastModified(baseTime.Add(time.Duration(i) * time.Hour)).
			Build()
	}
	
	return components
}

// GenerateComponentRefs creates component references for pipeline testing
func GenerateComponentRefs(count int) []models.ComponentRef {
	refs := make([]models.ComponentRef, count)
	for i := 0; i < count; i++ {
		refs[i] = models.ComponentRef{
			Type: models.ComponentTypePrompt,
			Path: "component-" + string(rune('a'+i)),
		}
	}
	return refs
}