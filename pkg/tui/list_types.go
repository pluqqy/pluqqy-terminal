package tui

// pane represents the different sections of the list view
type pane int

const (
	nonePane pane = iota
	searchPane
	pipelinesPane
	componentsPane
	previewPane
)

// pipelineItem represents a pipeline in the list view
type pipelineItem struct {
	name       string
	path       string
	tags       []string
	tokenCount int
	isArchived bool
}
