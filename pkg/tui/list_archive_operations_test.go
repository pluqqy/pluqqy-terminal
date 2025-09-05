package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"gopkg.in/yaml.v3"
)

func TestMainListModel_ArchiveUnarchiveOperations(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(t *testing.T) *MainListModel
		keyInput        tea.KeyMsg
		activePane      pane
		expectArchive   bool
		expectUnarchive bool
		validateState   func(t *testing.T, m *MainListModel)
	}{
		{
			name: "archive active pipeline with 'a' key",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestPipeline(t, "test.yaml", false)

				m := NewMainListModel()
				m.loadPipelines()
				m.stateManager.ActivePane = pipelinesPane
				m.stateManager.PipelineCursor = 0
				return m
			},
			keyInput:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			activePane:    pipelinesPane,
			expectArchive: true,
			validateState: func(t *testing.T, m *MainListModel) {
				// Should show archive confirmation
				if !m.operations.PipelineOperator.IsArchiveConfirmActive() {
					t.Error("Archive confirmation should be active")
				}
			},
		},
		{
			name: "unarchive archived pipeline with 'a' key",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestPipeline(t, "test.yaml", true)

				m := NewMainListModel()
				m.search.Query = "status:archived"
				m.performSearch()
				m.stateManager.ActivePane = pipelinesPane
				m.stateManager.PipelineCursor = 0
				return m
			},
			keyInput:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			activePane:      pipelinesPane,
			expectUnarchive: true,
			validateState: func(t *testing.T, m *MainListModel) {
				// Should show unarchive confirmation
				if !m.operations.PipelineOperator.IsArchiveConfirmActive() {
					t.Error("Archive confirmation should be active")
				}
			},
		},
		{
			name: "archive active component with 'a' key",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestComponent(t, "test.md", "contexts", false)

				m := NewMainListModel()
				m.loadComponents()
				m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
				m.stateManager.ActivePane = componentsPane
				m.stateManager.ComponentCursor = 0
				return m
			},
			keyInput:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			activePane:    componentsPane,
			expectArchive: true,
			validateState: func(t *testing.T, m *MainListModel) {
				if !m.operations.PipelineOperator.IsArchiveConfirmActive() {
					t.Error("Archive confirmation should be active")
				}
			},
		},
		{
			name: "unarchive archived component with 'a' key",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestComponent(t, "test.md", "contexts", true)

				m := NewMainListModel()
				m.search.Query = "status:archived"
				m.performSearch()
				m.stateManager.ActivePane = componentsPane
				m.stateManager.ComponentCursor = 0
				return m
			},
			keyInput:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			activePane:      componentsPane,
			expectUnarchive: true,
			validateState: func(t *testing.T, m *MainListModel) {
				if !m.operations.PipelineOperator.IsArchiveConfirmActive() {
					t.Error("Archive confirmation should be active")
				}
			},
		},
		{
			name: "no action when no items selected",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)

				m := NewMainListModel()
				m.loadPipelines()
				m.stateManager.ActivePane = pipelinesPane
				m.stateManager.PipelineCursor = 0
				return m
			},
			keyInput:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			activePane: pipelinesPane,
			validateState: func(t *testing.T, m *MainListModel) {
				if m.operations.PipelineOperator.IsArchiveConfirmActive() {
					t.Error("Archive confirmation should not be active when no items")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup(t)
			m.stateManager.ActivePane = tt.activePane

			// Process the key input
			_, cmd := m.Update(tt.keyInput)

			// Execute any returned command
			if cmd != nil {
				msg := cmd()
				if msg != nil {
					m.Update(msg)
				}
			}

			// Validate state
			if tt.validateState != nil {
				tt.validateState(t, m)
			}
		})
	}
}

func TestMainListModel_SearchWithArchived(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) *MainListModel
		searchQuery    string
		expectArchived bool
		validateCount  func(t *testing.T, pipelines, components int)
	}{
		{
			name: "search for archived items shows only archived",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestPipeline(t, "active.yaml", false)
				createArchiveTestPipeline(t, "archived.yaml", true)
				createArchiveTestComponent(t, "active.md", "contexts", false)
				createArchiveTestComponent(t, "archived.md", "contexts", true)

				m := NewMainListModel()
				// Set the search query before loading
				m.search.Query = "status:archived"
				m.loadPipelines()
				m.loadComponents()
				m.operations.BusinessLogic.SetComponents(m.data.Prompts, m.data.Contexts, m.data.Rules)
				// Then perform search
				m.performSearch()
				return m
			},
			searchQuery:    "status:archived",
			expectArchived: true,
			validateCount: func(t *testing.T, pipelines, components int) {
				if pipelines != 1 {
					t.Errorf("Expected 1 archived pipeline, got %d", pipelines)
				}
				if components != 1 {
					t.Errorf("Expected 1 archived component, got %d", components)
				}
			},
		},
		{
			name: "clear search removes archived items",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestPipeline(t, "active.yaml", false)
				createArchiveTestPipeline(t, "archived.yaml", true)

				m := NewMainListModel()
				// First search for archived
				m.search.Query = "status:archived"
				m.performSearch()
				// Then clear search
				m.search.Query = ""
				m.performSearch()
				return m
			},
			searchQuery:    "",
			expectArchived: false,
			validateCount: func(t *testing.T, pipelines, components int) {
				// Should only show active items
				if pipelines != 1 {
					t.Errorf("Expected 1 active pipeline after clearing search, got %d", pipelines)
				}

				// Should have no archived items
			},
		},
		{
			name: "search with other criteria excludes archived by default",
			setup: func(t *testing.T) *MainListModel {
				setupArchiveTestEnvironment(t)
				createArchiveTestPipelineWithTags(t, "tagged.yaml", []string{"test"}, false)
				createArchiveTestPipelineWithTags(t, "archived-tagged.yaml", []string{"test"}, true)

				m := NewMainListModel()
				m.loadPipelines()
				return m
			},
			searchQuery:    "tag:test",
			expectArchived: false,
			validateCount: func(t *testing.T, pipelines, components int) {
				if pipelines != 1 {
					t.Errorf("Expected 1 active tagged pipeline, got %d", pipelines)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup(t)
			m.search.Query = tt.searchQuery
			m.performSearch()

			// Count filtered items
			pipelineCount := len(m.getCurrentPipelines())
			componentCount := len(m.getCurrentComponents())

			if tt.validateCount != nil {
				tt.validateCount(t, pipelineCount, componentCount)
			}

			// Check archived status
			for _, p := range m.getCurrentPipelines() {
				if tt.expectArchived && !p.isArchived {
					t.Error("Expected archived pipelines in results")
				} else if !tt.expectArchived && p.isArchived {
					t.Error("Should not have archived pipelines in results")
				}
			}
		})
	}
}

func TestMainListModel_ReloadAfterArchiveOperation(t *testing.T) {
	setupArchiveTestEnvironment(t)

	// Create initial state
	createArchiveTestPipeline(t, "test.yaml", false)

	m := NewMainListModel()
	m.search.Query = "status:archived"
	m.loadPipelines()

	// Simulate archiving
	reloadCalled := false
	reloadFunc := func() {
		reloadCalled = true
		m.loadPipelines()
		m.performSearch()
	}

	// Archive the pipeline
	cmd := m.operations.PipelineOperator.ArchivePipeline("test.yaml", reloadFunc)
	msg := cmd()

	// Check status message
	if statusMsg, ok := msg.(StatusMsg); ok {
		if !archiveTestContains(string(statusMsg), "Archived") {
			t.Errorf("Expected archive success message, got: %s", statusMsg)
		}
	}

	// Verify reload was called
	if !reloadCalled {
		t.Error("Reload function was not called after archive")
	}
}

// Helper functions
func setupArchiveTestEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldWd) })
	os.Chdir(tmpDir)

	// Create directory structure
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.PipelinesDir), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.ComponentsDir, files.ContextsDir), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.ComponentsDir, files.PromptsDir), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.ComponentsDir, files.RulesDir), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir), 0755)
	os.MkdirAll(filepath.Join(files.PluqqyDir, files.ArchiveDir, files.ComponentsDir, files.ContextsDir), 0755)
}

func createArchiveTestPipeline(t *testing.T, name string, archived bool) {
	pipeline := models.Pipeline{
		Name: "Test Pipeline",
		Tags: []string{"test"},
	}
	data, _ := yaml.Marshal(pipeline)

	var path string
	if archived {
		path = filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, name)
	} else {
		path = filepath.Join(files.PluqqyDir, files.PipelinesDir, name)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to create test pipeline: %v", err)
	}
}

func createArchiveTestPipelineWithTags(t *testing.T, name string, tags []string, archived bool) {
	pipeline := models.Pipeline{
		Name: "Test Pipeline",
		Tags: tags,
	}
	data, _ := yaml.Marshal(pipeline)

	var path string
	if archived {
		path = filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, name)
	} else {
		path = filepath.Join(files.PluqqyDir, files.PipelinesDir, name)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to create test pipeline: %v", err)
	}
}

func createArchiveTestComponent(t *testing.T, name, compType string, archived bool) {
	content := `---
tags: [test]
---
# Test Component
Content here`

	var path string
	if archived {
		path = filepath.Join(files.PluqqyDir, files.ArchiveDir, files.ComponentsDir, compType, name)
	} else {
		path = filepath.Join(files.PluqqyDir, files.ComponentsDir, compType, name)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}
}

func archiveTestContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
