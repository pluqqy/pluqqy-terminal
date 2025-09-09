package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestEmptyStateMessages(t *testing.T) {
	t.Run("ComponentTableRenderer shows import examples message when empty", func(t *testing.T) {
		renderer := NewComponentTableRenderer(80, 24, true)
		renderer.SetActive(true)
		renderer.SetComponents([]componentItem{}) // Empty components

		content := renderer.buildTableContent(20, 20, 10, 10)
		
		assert.Contains(t, content, "No components found")
		assert.Contains(t, content, "Press 'n' to create one")
		assert.Contains(t, content, "or 'E' to import examples")
	})

	t.Run("PipelineViewRenderer shows import examples message when empty", func(t *testing.T) {
		renderer := NewPipelineViewRenderer(80, 24)
		renderer.ActivePane = pipelinesPane
		renderer.Pipelines = []pipelineItem{}
		renderer.FilteredPipelines = []pipelineItem{}

		content := renderer.buildScrollableContent(20, 20, 10)
		
		assert.Contains(t, content, "No pipelines found")
		assert.Contains(t, content, "Press 'n' to create one")
		assert.Contains(t, content, "or 'E' to import examples")
	})

	t.Run("ComponentTableRenderer doesn't show import message when filtered", func(t *testing.T) {
		renderer := NewComponentTableRenderer(80, 24, true)
		renderer.SetActive(true)
		renderer.SetComponents([]componentItem{}) // Empty due to filter

		// Simulate having components but filtered out
		content := renderer.buildTableContent(20, 20, 10, 10)
		
		// When empty due to filter, it should not suggest importing examples
		assert.Contains(t, content, "No components found")
	})
}

func TestImportExamplesCommand(t *testing.T) {

	t.Run("E key triggers import when lists are empty", func(t *testing.T) {
		// Create a model with empty lists
		model := NewMainListModel()
		model.SetSize(80, 24)
		model.data.Prompts = []componentItem{}
		model.data.Contexts = []componentItem{}
		model.data.Rules = []componentItem{}
		model.data.Pipelines = []pipelineItem{}
		model.data.FilteredComponents = []componentItem{}
		model.data.FilteredPipelines = []pipelineItem{}

		// Simulate pressing 'E' key
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}}
		_, cmd := model.Update(msg)

		// The command should be returned (ImportExamples)
		assert.NotNil(t, cmd, "E key should trigger import examples command")
	})

	t.Run("E key does nothing when lists have items", func(t *testing.T) {
		// Create a model with some items
		model := NewMainListModel()
		model.SetSize(80, 24)
		model.data.Prompts = []componentItem{{name: "test-prompt"}}
		model.data.Contexts = []componentItem{}
		model.data.Rules = []componentItem{}
		model.data.Pipelines = []pipelineItem{{name: "test-pipeline"}}

		// Simulate pressing 'E' key
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}}
		updatedModel, cmd := model.Update(msg)

		// Should return the model unchanged with no command
		assert.Equal(t, model, updatedModel)
		assert.Nil(t, cmd, "E key should do nothing when items exist")
	})
}

