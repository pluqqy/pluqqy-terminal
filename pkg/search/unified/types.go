package unified

import (
	"time"
)

// ComponentItemWrapper wraps the TUI componentItem to implement Searchable
type ComponentItemWrapper struct {
	name         string
	path         string
	compType     string
	lastModified time.Time
	usageCount   int
	tokenCount   int
	tags         []string
	isArchived   bool
	content      string // For search purposes
}

// NewComponentItemWrapper creates a new wrapper for componentItem
func NewComponentItemWrapper(name, path, compType string, lastModified time.Time, usageCount, tokenCount int, tags []string, isArchived bool, content string) *ComponentItemWrapper {
	return &ComponentItemWrapper{
		name:         name,
		path:         path,
		compType:     compType,
		lastModified: lastModified,
		usageCount:   usageCount,
		tokenCount:   tokenCount,
		tags:         tags,
		isArchived:   isArchived,
		content:      content,
	}
}

// Implement Searchable interface
func (c *ComponentItemWrapper) GetName() string {
	return c.name
}

func (c *ComponentItemWrapper) GetPath() string {
	return c.path
}

func (c *ComponentItemWrapper) GetType() string {
	return "component"
}

func (c *ComponentItemWrapper) GetSubType() string {
	return c.compType
}

func (c *ComponentItemWrapper) GetTags() []string {
	return c.tags
}

func (c *ComponentItemWrapper) GetContent() string {
	// Return searchable content (name + tags + actual content)
	searchableContent := c.name
	if len(c.tags) > 0 {
		for _, tag := range c.tags {
			searchableContent += " " + tag
		}
	}
	if c.content != "" {
		searchableContent += " " + c.content
	}
	return searchableContent
}

func (c *ComponentItemWrapper) GetModified() time.Time {
	return c.lastModified
}

func (c *ComponentItemWrapper) IsArchived() bool {
	return c.isArchived
}

func (c *ComponentItemWrapper) GetTokenCount() int {
	return c.tokenCount
}

func (c *ComponentItemWrapper) GetUsageCount() int {
	return c.usageCount
}

// PipelineItemWrapper wraps the TUI pipelineItem to implement Searchable
type PipelineItemWrapper struct {
	name       string
	path       string
	tags       []string
	tokenCount int
	isArchived bool
	modified   time.Time
	content    string // For search purposes
}

// NewPipelineItemWrapper creates a new wrapper for pipelineItem
func NewPipelineItemWrapper(name, path string, tags []string, tokenCount int, isArchived bool, modified time.Time, content string) *PipelineItemWrapper {
	return &PipelineItemWrapper{
		name:       name,
		path:       path,
		tags:       tags,
		tokenCount: tokenCount,
		isArchived: isArchived,
		modified:   modified,
		content:    content,
	}
}

// Implement Searchable interface
func (p *PipelineItemWrapper) GetName() string {
	return p.name
}

func (p *PipelineItemWrapper) GetPath() string {
	return p.path
}

func (p *PipelineItemWrapper) GetType() string {
	return "pipeline"
}

func (p *PipelineItemWrapper) GetSubType() string {
	return "" // Pipelines don't have subtypes
}

func (p *PipelineItemWrapper) GetTags() []string {
	return p.tags
}

func (p *PipelineItemWrapper) GetContent() string {
	// Return searchable content (name + tags + content)
	searchableContent := p.name
	if len(p.tags) > 0 {
		for _, tag := range p.tags {
			searchableContent += " " + tag
		}
	}
	if p.content != "" {
		searchableContent += " " + p.content
	}
	return searchableContent
}

func (p *PipelineItemWrapper) GetModified() time.Time {
	return p.modified
}

func (p *PipelineItemWrapper) IsArchived() bool {
	return p.isArchived
}

func (p *PipelineItemWrapper) GetTokenCount() int {
	return p.tokenCount
}

func (p *PipelineItemWrapper) GetUsageCount() int {
	return 0 // Pipelines don't have usage counts in the same way
}