package unified

import (
	"time"
)

// ComponentItem represents a component for searching
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

// PipelineItem represents a pipeline for searching
type PipelineItem struct {
	Name       string
	Path       string
	Tags       []string
	TokenCount int
	IsArchived bool
	Modified   time.Time
}