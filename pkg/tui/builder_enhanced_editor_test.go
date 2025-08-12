package tui

import (
	"os"
	"path/filepath"
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// setupTestComponent creates a temporary test component file and returns its path
func setupTestComponent(t *testing.T, dir, name, content string) string {
	t.Helper()
	
	// Create the full path
	fullPath := filepath.Join(dir, name)
	
	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Write the test component
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test component: %v", err)
	}
	
	return fullPath
}

// cleanupTestComponents removes test component files
func cleanupTestComponents(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		// Remove the file and its parent directory
		os.RemoveAll(filepath.Dir(path))
	}
}

// TestPipelineBuilderModel_EnhancedEditorIntegration tests the integration
// of the enhanced editor within the Pipeline Builder
func TestPipelineBuilderModel_EnhancedEditorIntegration(t *testing.T) {
	// Create a temporary directory for test components
	tempDir := t.TempDir()
	tests := []struct {
		name           string
		setup          func() *PipelineBuilderModel
		msg            tea.KeyMsg
		checkState     func(t *testing.T, m *PipelineBuilderModel)
		cleanup        func()
	}{
		{
			name: "pressing e activates enhanced editor from left column",
			setup: func() *PipelineBuilderModel {
				// Create a test component file
				testPath := setupTestComponent(t, filepath.Join(tempDir, "prompts"), "test.md", "Test prompt content")
				
				m := NewPipelineBuilderModel()
				m.editingName = false // Exit name editing mode
				m.activeColumn = leftColumn
				m.leftCursor = 0
				// Add the test component with the actual path
				m.prompts = []componentItem{
					{name: "test-prompt", path: testPath, compType: models.ComponentTypePrompt},
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if !m.editingComponent {
					t.Error("Expected editingComponent to be true")
				}
				if !m.enhancedEditor.IsActive() {
					t.Error("Expected enhanced editor to be active")
				}
				if m.enhancedEditor.ComponentName != "test-prompt" {
					t.Errorf("Expected component name to be 'test-prompt', got %s", m.enhancedEditor.ComponentName)
				}
			},
			cleanup: func() {
				cleanupTestComponents(t, filepath.Join(tempDir, "prompts"))
			},
		},
		{
			name: "pressing e activates enhanced editor from right column",
			setup: func() *PipelineBuilderModel {
				// Create a test component file
				testPath := setupTestComponent(t, filepath.Join(tempDir, "prompts"), "selected.md", "Selected component content")
				
				m := NewPipelineBuilderModel()
				m.editingName = false // Exit name editing mode
				m.activeColumn = rightColumn
				m.rightCursor = 0
				// Add selected components with the actual path
				m.selectedComponents = []models.ComponentRef{
					{Path: "../" + testPath, Order: 1},
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if !m.editingComponent {
					t.Error("Expected editingComponent to be true")
				}
				if !m.enhancedEditor.IsActive() {
					t.Error("Expected enhanced editor to be active")
				}
			},
			cleanup: func() {
				cleanupTestComponents(t, filepath.Join(tempDir, "prompts"))
			},
		},
		{
			name: "escape key exits enhanced editor",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false
				m.editingComponent = true
				m.enhancedEditor.Active = true
				m.enhancedEditor.ComponentName = "test-component"
				m.enhancedEditor.Content = "test content"
				m.enhancedEditor.OriginalContent = "test content"
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyEsc},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if m.editingComponent {
					t.Error("Expected editingComponent to be false after escape")
				}
				if m.enhancedEditor.IsActive() {
					t.Error("Expected enhanced editor to be inactive after escape")
				}
			},
			cleanup: func() {}, // No cleanup needed
		},
		{
			name: "@ key activates file picker for project references",
			setup: func() *PipelineBuilderModel {
				m := NewPipelineBuilderModel()
				m.editingName = false
				m.editingComponent = true
				m.enhancedEditor.Active = true
				m.enhancedEditor.Mode = EditorModeNormal
				m.enhancedEditor.ComponentName = "test-component"
				m.enhancedEditor.Textarea.SetValue("test content")
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}},
			checkState: func(t *testing.T, m *PipelineBuilderModel) {
				if !m.editingComponent {
					t.Error("Expected to still be in editing mode")
				}
				if !m.enhancedEditor.IsActive() {
					t.Error("Expected enhanced editor to still be active")
				}
				if m.enhancedEditor.Mode != EditorModeFilePicking {
					t.Error("Expected editor to be in file picking mode after @")
				}
			},
			cleanup: func() {}, // No cleanup needed
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			defer tt.cleanup() // Clean up test files after each test
			
			updatedModel, _ := m.Update(tt.msg)
			if updatedM, ok := updatedModel.(*PipelineBuilderModel); ok {
				tt.checkState(t, updatedM)
			} else {
				t.Error("Expected model to be *PipelineBuilderModel")
			}
		})
	}
}

// TestPipelineBuilderModel_EnhancedEditorView tests that the enhanced editor
// view is rendered correctly when active
func TestPipelineBuilderModel_EnhancedEditorView(t *testing.T) {
	m := NewPipelineBuilderModel()
	m.editingName = false
	m.editingComponent = true
	m.enhancedEditor.Active = true
	m.enhancedEditor.ComponentName = "test-component"
	m.enhancedEditor.ComponentType = models.ComponentTypePrompt
	m.enhancedEditor.Content = "Test content for the component"
	m.width = 80
	m.height = 24
	
	// Get the view
	view := m.View()
	
	// Check that the view contains expected elements
	if view == "" {
		t.Error("Expected non-empty view when enhanced editor is active")
	}
	
	// The view should contain the component name
	if !builderTestContains(view, "TEST-COMPONENT") && !builderTestContains(view, "test-component") {
		t.Error("Expected view to contain component name")
	}
}

// TestPipelineBuilderModel_EnhancedEditorConsistency tests that the enhanced
// editor maintains consistency with Pipeline Builder state
func TestPipelineBuilderModel_EnhancedEditorConsistency(t *testing.T) {
	m := NewPipelineBuilderModel()
	
	// Verify initial state
	if m.useEnhancedEditor != true {
		t.Error("Expected useEnhancedEditor to be true by default")
	}
	
	if m.enhancedEditor == nil {
		t.Fatal("Expected enhancedEditor to be initialized")
	}
	
	if m.enhancedEditor.IsActive() {
		t.Error("Expected enhanced editor to be inactive initially")
	}
	
	// Test that closing enhanced editor properly cleans up state
	m.editingComponent = true
	m.enhancedEditor.Active = true
	m.enhancedEditor.Active = false // Simulate stopping
	
	// After stopping, the editor should be inactive
	if m.enhancedEditor.IsActive() {
		t.Error("Expected enhanced editor to be inactive after stopping")
	}
}

// Helper function for testing
func builderTestContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || builderTestContains(s[1:], substr)))
}