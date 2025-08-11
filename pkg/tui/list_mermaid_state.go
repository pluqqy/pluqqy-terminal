package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// MermaidState manages the state for mermaid diagram generation
type MermaidState struct {
	Active            bool
	GeneratingDiagram bool
	LastGeneratedFile string
	LastError         error
}

// NewMermaidState creates a new mermaid state
func NewMermaidState() *MermaidState {
	return &MermaidState{
		Active: false,
	}
}

// HandleInput processes key messages for mermaid functionality
func (ms *MermaidState) HandleInput(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if !ms.Active {
		return false, nil
	}
	
	// Handle escape to cancel generation
	if msg.String() == "esc" && ms.GeneratingDiagram {
		ms.GeneratingDiagram = false
		return true, nil
	}
	
	return false, nil
}

// StartGeneration marks the beginning of diagram generation
func (ms *MermaidState) StartGeneration() {
	ms.GeneratingDiagram = true
	ms.LastError = nil
}

// CompleteGeneration marks successful completion
func (ms *MermaidState) CompleteGeneration(filename string) {
	ms.GeneratingDiagram = false
	ms.LastGeneratedFile = filename
	ms.LastError = nil
}

// FailGeneration marks generation failure
func (ms *MermaidState) FailGeneration(err error) {
	ms.GeneratingDiagram = false
	ms.LastError = err
}

// IsGenerating returns whether diagram generation is in progress
func (ms *MermaidState) IsGenerating() bool {
	return ms.GeneratingDiagram
}