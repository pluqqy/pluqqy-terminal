package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestComponentBuilder tests the ComponentBuilder functionality
func TestComponentBuilder(t *testing.T) {
	t.Run("creates component with defaults", func(t *testing.T) {
		comp := NewComponentBuilder("test-component").Build()
		
		if comp.Name != "test-component" {
			t.Errorf("Name = %q, want %q", comp.Name, "test-component")
		}
		if comp.CompType != models.ComponentTypePrompt {
			t.Errorf("CompType = %q, want %q", comp.CompType, models.ComponentTypePrompt)
		}
		if comp.TokenCount != 100 {
			t.Errorf("TokenCount = %d, want %d", comp.TokenCount, 100)
		}
		if comp.IsArchived {
			t.Error("IsArchived should be false by default")
		}
	})
	
	t.Run("applies all builder methods", func(t *testing.T) {
		testTime := time.Now().Add(-24 * time.Hour)
		comp := NewComponentBuilder("test").
			WithType(models.ComponentTypeContext).
			WithPath("/custom/path.md").
			WithTokens(500).
			WithTags("tag1", "tag2").
			WithUsageCount(10).
			WithLastModified(testTime).
			Archived().
			Build()
			
		if comp.CompType != models.ComponentTypeContext {
			t.Errorf("CompType = %q, want %q", comp.CompType, models.ComponentTypeContext)
		}
		if comp.Path != "/custom/path.md" {
			t.Errorf("Path = %q, want %q", comp.Path, "/custom/path.md")
		}
		if comp.TokenCount != 500 {
			t.Errorf("TokenCount = %d, want %d", comp.TokenCount, 500)
		}
		if len(comp.Tags) != 2 {
			t.Errorf("Tags length = %d, want %d", len(comp.Tags), 2)
		}
		if comp.UsageCount != 10 {
			t.Errorf("UsageCount = %d, want %d", comp.UsageCount, 10)
		}
		if !comp.IsArchived {
			t.Error("IsArchived should be true")
		}
		if !comp.LastModified.Equal(testTime) {
			t.Errorf("LastModified = %v, want %v", comp.LastModified, testTime)
		}
	})
}

// TestPipelineBuilder tests the PipelineBuilder functionality
func TestPipelineBuilder(t *testing.T) {
	t.Run("creates pipeline with defaults", func(t *testing.T) {
		pipeline := NewPipelineBuilder("test-pipeline").Build()
		
		if pipeline.Name != "test-pipeline" {
			t.Errorf("Name = %q, want %q", pipeline.Name, "test-pipeline")
		}
		if pipeline.Path != "test-pipeline.yaml" {
			t.Errorf("Path = %q, want %q", pipeline.Path, "test-pipeline.yaml")
		}
		if len(pipeline.Components) != 0 {
			t.Errorf("Components length = %d, want %d", len(pipeline.Components), 0)
		}
	})
	
	t.Run("adds components correctly", func(t *testing.T) {
		refs := []models.ComponentRef{
			{Type: models.ComponentTypePrompt, Path: "comp1"},
			{Type: models.ComponentTypeContext, Path: "comp2"},
		}
		
		pipeline := NewPipelineBuilder("test").
			WithComponents(refs...).
			Build()
			
		if len(pipeline.Components) != 2 {
			t.Errorf("Components length = %d, want %d", len(pipeline.Components), 2)
		}
		if pipeline.Components[0].Path != "comp1" {
			t.Errorf("First component path = %q, want %q", pipeline.Components[0].Path, "comp1")
		}
	})
	
	t.Run("builds pipeline item correctly", func(t *testing.T) {
		item := NewPipelineBuilder("test").
			WithTags("tag1", "tag2").
			WithTokenCount(300).
			Archived().
			BuildItem()
			
		if item.Name != "test" {
			t.Errorf("Name = %q, want %q", item.Name, "test")
		}
		if len(item.Tags) != 2 {
			t.Errorf("Tags length = %d, want %d", len(item.Tags), 2)
		}
		if item.TokenCount != 300 {
			t.Errorf("TokenCount = %d, want %d", item.TokenCount, 300)
		}
		if !item.IsArchived {
			t.Error("IsArchived should be true")
		}
	})
}

// TestEnvironment tests the TestEnvironment functionality
func TestTestEnvironment(t *testing.T) {
	t.Run("creates and cleans up temp directory", func(t *testing.T) {
		env := NewTestEnvironment(t)
		
		// Check temp directory exists
		if _, err := os.Stat(env.TempDir); os.IsNotExist(err) {
			t.Error("Temp directory should exist")
		}
		
		tempDir := env.TempDir
		
		// Clean up
		env.Cleanup()
		
		// Check temp directory is removed
		if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
			t.Error("Temp directory should be removed after cleanup")
		}
	})
	
	t.Run("initializes project structure", func(t *testing.T) {
		env := NewTestEnvironment(t)
		defer env.Cleanup()
		
		err := env.InitProjectStructure()
		if err != nil {
			t.Fatalf("InitProjectStructure failed: %v", err)
		}
		
		// Check directories exist
		dirs := []string{
			filepath.Join(env.TempDir, files.PluqqyDir, "pipelines"),
			filepath.Join(env.TempDir, files.PluqqyDir, "components", models.ComponentTypePrompt),
			filepath.Join(env.TempDir, files.PluqqyDir, "components", models.ComponentTypeContext),
			filepath.Join(env.TempDir, files.PluqqyDir, "components", models.ComponentTypeRules),
		}
		
		for _, dir := range dirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Directory %q should exist", dir)
			}
		}
	})
	
	t.Run("creates component file with tags", func(t *testing.T) {
		env := NewTestEnvironment(t)
		defer env.Cleanup()
		env.InitProjectStructure()
		
		path := env.CreateComponentFile("prompts", "test-prompt", "Test content", []string{"tag1", "tag2"})
		
		if path != "components/prompts/test-prompt.md" {
			t.Errorf("Path = %q, want %q", path, "components/prompts/test-prompt.md")
		}
		
		// Check file exists
		fullPath := filepath.Join(env.TempDir, files.PluqqyDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}
		
		// Check content includes tags
		contentStr := string(content)
		if !contains(contentStr, "tags: [tag1, tag2]") {
			t.Errorf("Content should include tags, got: %s", contentStr)
		}
		if !contains(contentStr, "Test content") {
			t.Errorf("Content should include test content, got: %s", contentStr)
		}
	})
	
	t.Run("creates pipeline file", func(t *testing.T) {
		env := NewTestEnvironment(t)
		defer env.Cleanup()
		env.InitProjectStructure()
		
		refs := []models.ComponentRef{
			{Type: models.ComponentTypePrompt, Path: "comp1"},
		}
		
		path := env.CreatePipelineFile("test-pipeline", refs, []string{"tag1"})
		
		if path != "test-pipeline.yaml" {
			t.Errorf("Path = %q, want %q", path, "test-pipeline.yaml")
		}
		
		// Check file exists
		fullPath := filepath.Join(env.TempDir, files.PluqqyDir, "pipelines", path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Error("Pipeline file should exist")
		}
	})
}

// TestFactoryFunctions tests the factory functions
func TestFactoryFunctions(t *testing.T) {
	t.Run("MakePromptComponent", func(t *testing.T) {
		comp := MakePromptComponent("test", 200)
		
		if comp.Name != "test" {
			t.Errorf("Name = %q, want %q", comp.Name, "test")
		}
		if comp.CompType != models.ComponentTypePrompt {
			t.Errorf("CompType = %q, want %q", comp.CompType, models.ComponentTypePrompt)
		}
		if comp.TokenCount != 200 {
			t.Errorf("TokenCount = %d, want %d", comp.TokenCount, 200)
		}
	})
	
	t.Run("MakeTestComponents", func(t *testing.T) {
		components := MakeTestComponents("prompts", "comp1", "comp2", "comp3")
		
		if len(components) != 3 {
			t.Errorf("Components length = %d, want %d", len(components), 3)
		}
		for i, comp := range components {
			if comp.CompType != "prompts" {
				t.Errorf("Component[%d].CompType = %q, want %q", i, comp.CompType, "prompts")
			}
		}
	})
	
	t.Run("MakeSimplePipeline", func(t *testing.T) {
		pipeline := MakeSimplePipeline("test")
		
		if pipeline.Name != "test" {
			t.Errorf("Name = %q, want %q", pipeline.Name, "test")
		}
		if len(pipeline.Components) != 0 {
			t.Errorf("Components length = %d, want %d", len(pipeline.Components), 0)
		}
	})
	
	t.Run("MakePipelineWithComponents", func(t *testing.T) {
		pipeline := MakePipelineWithComponents("test", 3)
		
		if len(pipeline.Components) != 3 {
			t.Errorf("Components length = %d, want %d", len(pipeline.Components), 3)
		}
	})
}

// TestWaitForCondition tests the utility functions
func TestWaitForCondition(t *testing.T) {
	t.Run("succeeds when condition is met", func(t *testing.T) {
		called := false
		condition := func() bool {
			if !called {
				called = true
				return false
			}
			return true
		}
		
		// Should not panic
		WaitForCondition(t, condition, 100*time.Millisecond, "condition should be met")
		
		if !called {
			t.Error("Condition should have been checked")
		}
	})
	
	// Note: We can't easily test the timeout case with a mock since WaitForCondition
	// expects a real *testing.T. This would require refactoring WaitForCondition
	// to accept an interface instead.
}

// TestAssertions tests the assertion functions  
func TestAssertions(t *testing.T) {
	// Note: AssertComponentEqual expects a real *testing.T, so we test with real assertions
	t.Run("AssertComponentEqual passes for equal components", func(t *testing.T) {
		comp1 := ComponentItem{Name: "comp", TokenCount: 100, CompType: "prompts"}
		comp2 := ComponentItem{Name: "comp", TokenCount: 100, CompType: "prompts"}
		
		// This should not fail
		AssertComponentEqual(t, comp1, comp2)
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && s[0:len(substr)] == substr) || 
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(substr) > 0 && len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockTestingT is a mock implementation of testing.T for testing our test helpers
type mockTestingT struct {
	failed   bool
	hasError bool
	message  string
	errors   []string
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.message = format
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.hasError = true
	m.errors = append(m.errors, format)
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.hasError = true
	if len(args) > 0 {
		m.errors = append(m.errors, args[0].(string))
	}
}