package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ComponentUsageState manages the state for displaying component usage information
type ComponentUsageState struct {
	Active                  bool
	SelectedComponent       componentItem
	PipelinesUsingComponent []PipelineUsageInfo
	Width                   int
	Height                  int
	ScrollOffset            int
	SelectedIndex           int
}

// PipelineUsageInfo contains information about a pipeline using a component
type PipelineUsageInfo struct {
	Name            string
	Path            string
	ComponentOrder  int
	TotalComponents int
}

// NewComponentUsageState creates a new component usage state
func NewComponentUsageState() *ComponentUsageState {
	return &ComponentUsageState{
		Active:                  false,
		PipelinesUsingComponent: []PipelineUsageInfo{},
		ScrollOffset:            0,
		SelectedIndex:           0,
	}
}

// Start activates the component usage view for a specific component
func (cus *ComponentUsageState) Start(component componentItem) {
	cus.Active = true
	cus.SelectedComponent = component
	cus.ScrollOffset = 0
	cus.SelectedIndex = 0
	cus.loadPipelinesUsingComponent()
}

// Stop deactivates the component usage view
func (cus *ComponentUsageState) Stop() {
	cus.Active = false
	cus.PipelinesUsingComponent = []PipelineUsageInfo{}
	cus.ScrollOffset = 0
	cus.SelectedIndex = 0
}

// HandleInput processes keyboard input for the component usage view
func (cus *ComponentUsageState) HandleInput(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !cus.Active {
		return false, nil
	}

	switch msg.String() {
	case "esc", "q", "u":
		cus.Stop()
		return true, nil

	case "up", "k":
		if cus.SelectedIndex > 0 {
			cus.SelectedIndex--
			cus.ensureSelectedVisible()
		}
		return true, nil

	case "down", "j":
		if cus.SelectedIndex < len(cus.PipelinesUsingComponent)-1 {
			cus.SelectedIndex++
			cus.ensureSelectedVisible()
		}
		return true, nil

	case "home", "g":
		cus.SelectedIndex = 0
		cus.ScrollOffset = 0
		return true, nil

	case "end", "G":
		if len(cus.PipelinesUsingComponent) > 0 {
			cus.SelectedIndex = len(cus.PipelinesUsingComponent) - 1
			cus.ensureSelectedVisible()
		}
		return true, nil

	case "pgup":
		visibleLines := cus.getVisibleLines()
		cus.SelectedIndex = max(0, cus.SelectedIndex-visibleLines)
		cus.ensureSelectedVisible()
		return true, nil

	case "pgdown":
		visibleLines := cus.getVisibleLines()
		maxIndex := len(cus.PipelinesUsingComponent) - 1
		cus.SelectedIndex = min(maxIndex, cus.SelectedIndex+visibleLines)
		cus.ensureSelectedVisible()
		return true, nil
	}

	return false, nil
}

// loadPipelinesUsingComponent loads all pipelines that use the selected component
func (cus *ComponentUsageState) loadPipelinesUsingComponent() {
	operator := NewComponentUsageOperator()
	cus.PipelinesUsingComponent = operator.FindPipelinesUsingComponent(cus.SelectedComponent.path)
}

// SetSize updates the dimensions of the component usage view
func (cus *ComponentUsageState) SetSize(width, height int) {
	cus.Width = width
	cus.Height = height
	cus.ensureSelectedVisible()
}

// getVisibleLines returns the number of lines that can be displayed
func (cus *ComponentUsageState) getVisibleLines() int {
	// Account for modal chrome: title, borders, help text
	return max(1, cus.Height-10)
}

// ensureSelectedVisible adjusts scroll offset to keep selected item visible
func (cus *ComponentUsageState) ensureSelectedVisible() {
	visibleLines := cus.getVisibleLines()
	
	if cus.SelectedIndex < cus.ScrollOffset {
		cus.ScrollOffset = cus.SelectedIndex
	} else if cus.SelectedIndex >= cus.ScrollOffset+visibleLines {
		cus.ScrollOffset = cus.SelectedIndex - visibleLines + 1
	}
}

// Helper functions
// max and min functions are already declared in other files