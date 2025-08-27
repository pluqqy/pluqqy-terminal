package tui

import (
	"time"
)

type column int

const (
	searchColumn column = iota
	leftColumn
	rightColumn
	previewColumn
)

// Constants for preview synchronization
const (
	defaultLinesPerComponent = 15 // Estimated lines per component in preview
	minLinesPerComponent     = 10 // Minimum estimate for lines per component
	scrollContextLines       = 2  // Lines to show before component when scrolling
	scrollBottomPadding      = 10 // Lines to keep from bottom when estimating position
)

type PipelineBuilderModel struct {
	// Composed data structures
	data      *BuilderDataStore
	viewports *BuilderViewportManager
	editors   *BuilderEditorComponents
	search    *BuilderSearchComponents
	ui        *BuilderUIComponents

	// Error handling
	err error
}

type componentItem struct {
	name         string
	path         string
	compType     string
	lastModified time.Time
	usageCount   int
	tokenCount   int
	tags         []string
	isArchived   bool
}

type clearEditSaveMsg struct{}

type componentSaveResultMsg struct {
	success       bool
	message       string
	componentPath string
	componentName string
	savedContent  string
}

type externalEditSaveMsg struct {
	savedContent string
}

type componentEditExitMsg struct{}

// NewPipelineBuilderModel creates a new Pipeline Builder with default configuration
// For custom configuration, use NewPipelineBuilderModelWithConfig
func NewPipelineBuilderModel() *PipelineBuilderModel {
	return NewPipelineBuilderModelWithConfig(DefaultPipelineBuilderConfig())
}