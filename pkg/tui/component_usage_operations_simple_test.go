package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentUsageOperator_BasicFunctionality(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create directory structure
	require.NoError(t, os.MkdirAll(".pluqqy/pipelines", 0755))
	require.NoError(t, os.MkdirAll(".pluqqy/components/contexts", 0755))

	// Create test component
	componentPath := ".pluqqy/components/contexts/test.md"
	require.NoError(t, os.WriteFile(componentPath, []byte("test content"), 0644))

	// Create test pipeline that uses the component
	pipelineYAML := `name: Test Pipeline
components:
  - type: contexts
    path: ../components/contexts/test.md
    order: 1
`
	pipelinePath := ".pluqqy/pipelines/test.yaml"
	require.NoError(t, os.WriteFile(pipelinePath, []byte(pipelineYAML), 0644))

	// Test finding pipelines
	operator := NewComponentUsageOperator()
	results := operator.FindPipelinesUsingComponent("components/contexts/test.md")

	assert.Len(t, results, 1, "Should find one pipeline")
	assert.Equal(t, "Test Pipeline", results[0].Name)
	assert.Equal(t, 1, results[0].ComponentOrder)
	assert.Equal(t, 1, results[0].TotalComponents)
}

func TestComponentUsageOperator_NoUsage(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create directory structure
	require.NoError(t, os.MkdirAll(".pluqqy/pipelines", 0755))

	// Create pipeline without the component
	pipelineYAML := `name: Other Pipeline
components:
  - type: prompts
    path: ../components/prompts/other.md
    order: 1
`
	pipelinePath := ".pluqqy/pipelines/other.yaml"
	require.NoError(t, os.WriteFile(pipelinePath, []byte(pipelineYAML), 0644))

	// Test finding pipelines
	operator := NewComponentUsageOperator()
	results := operator.FindPipelinesUsingComponent("components/contexts/test.md")

	assert.Empty(t, results, "Should find no pipelines")
}

func TestComponentUsageOperator_MultipleUsage(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create directory structure
	require.NoError(t, os.MkdirAll(".pluqqy/pipelines", 0755))

	// Create first pipeline
	pipeline1YAML := `name: Alpha Pipeline
components:
  - type: contexts
    path: ../components/contexts/shared.md
    order: 1
`
	require.NoError(t, os.WriteFile(".pluqqy/pipelines/alpha.yaml", []byte(pipeline1YAML), 0644))

	// Create second pipeline
	pipeline2YAML := `name: Beta Pipeline
components:
  - type: prompts
    path: ../components/prompts/prompt.md
    order: 1
  - type: contexts
    path: ../components/contexts/shared.md
    order: 2
`
	require.NoError(t, os.WriteFile(".pluqqy/pipelines/beta.yaml", []byte(pipeline2YAML), 0644))

	// Test finding pipelines
	operator := NewComponentUsageOperator()
	results := operator.FindPipelinesUsingComponent("components/contexts/shared.md")

	assert.Len(t, results, 2, "Should find two pipelines")
	
	// Results should be sorted alphabetically
	assert.Equal(t, "Alpha Pipeline", results[0].Name)
	assert.Equal(t, "Beta Pipeline", results[1].Name)
	
	// Check component positions
	assert.Equal(t, 1, results[0].ComponentOrder)
	assert.Equal(t, 2, results[1].ComponentOrder)
	assert.Equal(t, 2, results[1].TotalComponents)
}

func TestComponentUsageOperator_ArchivedPipelines(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create directory structure
	require.NoError(t, os.MkdirAll(".pluqqy/archive/pipelines", 0755))

	// Create archived pipeline
	archivedYAML := `name: Archived Pipeline
components:
  - type: contexts
    path: ../components/contexts/test.md
    order: 1
`
	archivePath := ".pluqqy/archive/pipelines/archived.yaml"
	require.NoError(t, os.WriteFile(archivePath, []byte(archivedYAML), 0644))

	// Test finding pipelines
	operator := NewComponentUsageOperator()
	results := operator.FindPipelinesUsingComponent("components/contexts/test.md")

	assert.Len(t, results, 1, "Should find archived pipeline")
	assert.Equal(t, "Archived Pipeline", results[0].Name)
}

func TestComponentUsageOperator_GetComponentUsageCount(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create directory structure
	require.NoError(t, os.MkdirAll(".pluqqy/pipelines", 0755))

	// Create multiple pipelines using the same component
	for i := 0; i < 3; i++ {
		pipelineYAML := `name: Pipeline
components:
  - type: contexts
    path: ../components/contexts/test.md
    order: 1
`
		path := filepath.Join(".pluqqy/pipelines", string(rune('a'+i))+".yaml")
		require.NoError(t, os.WriteFile(path, []byte(pipelineYAML), 0644))
	}

	operator := NewComponentUsageOperator()
	count := operator.GetComponentUsageCount("components/contexts/test.md")

	assert.Equal(t, 3, count, "Should count all pipelines using component")
}