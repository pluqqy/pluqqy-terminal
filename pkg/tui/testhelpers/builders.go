package testhelpers

import (
	"fmt"
	"testing"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// TestModelBuilder provides methods for creating test models
// This consolidates all the makeTest* functions from various test files
type TestModelBuilder struct {
	t *testing.T
}

// NewTestModelBuilder creates a new test model builder
func NewTestModelBuilder(t *testing.T) *TestModelBuilder {
	t.Helper()
	return &TestModelBuilder{t: t}
}

// MakeTestPipelineModel creates a pipeline model for testing
func MakeTestPipelineModel(name, path string) *models.Pipeline {
	return &models.Pipeline{
		Name:       name,
		Path:       path,
		Components: []models.ComponentRef{},
		Tags:       []string{},
	}
}

// MakeTestComponentItem creates a component item for testing  
func MakeTestComponentItem(name, compType string, tokenCount int, tags []string) ComponentItem {
	return ComponentItem{
		Name:       name,
		Path:       "test/" + name + ".md",
		CompType:   compType,
		TokenCount: tokenCount,
		Tags:       tags,
		UsageCount: 5,
		IsArchived: false,
	}
}

// MakeTestError creates a test error
func MakeTestError(msg string) error {
	// Return a simple error instead of tea.NewErrMsg which doesn't exist
	return fmt.Errorf("%s", msg)
}

// Utility Functions
// Note: WaitForCondition is already defined in assertions.go

// AssertEventually checks that a condition becomes true within a timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()
	WaitForCondition(t, condition, timeout, msg)
}

// CreateTestDirectory creates a temporary test directory
func CreateTestDirectory(t *testing.T) (string, func()) {
	t.Helper()
	env := NewTestEnvironment(t)
	env.InitProjectStructure()
	return env.TempDir, env.Cleanup
}

// CreateTestModel creates a generic tea.Model for testing
// This can be extended to create specific model types based on configuration
func CreateTestModel(config interface{}) tea.Model {
	// This is a placeholder that can be extended based on the config type
	// For now, return a simple model
	return nil
}