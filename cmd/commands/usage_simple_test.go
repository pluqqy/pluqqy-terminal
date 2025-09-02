package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageCommand_Basic(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create .pluqqy structure
	require.NoError(t, os.MkdirAll(".pluqqy/components/contexts", 0755))
	require.NoError(t, os.MkdirAll(".pluqqy/pipelines", 0755))

	// Create a component
	componentContent := `---
name: Test Component
tags: [test]
---

# Test Component Content`
	require.NoError(t, os.WriteFile(
		".pluqqy/components/contexts/test.md",
		[]byte(componentContent),
		0644,
	))

	// Create a pipeline using the component
	pipelineYAML := `name: Test Pipeline
components:
  - type: contexts
    path: ../components/contexts/test.md
    order: 1
tags: [test]
`
	require.NoError(t, os.WriteFile(
		".pluqqy/pipelines/test.yaml",
		[]byte(pipelineYAML),
		0644,
	))

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "find component usage by name",
			args: []string{"test"},
			contains: []string{
				"Component: test",
				"Type: context",
				"Used in 1 pipeline(s):",
				"Test Pipeline",
				"Position: 1 of 1",
			},
		},
		{
			name: "find component with type prefix",
			args: []string{"contexts/test"},
			contains: []string{
				"Component: contexts/test",
				"Type: context",
				"Test Pipeline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewUsageCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestUsageCommand_ComponentNotFound(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create .pluqqy structure
	require.NoError(t, os.MkdirAll(".pluqqy", 0755))

	cmd := NewUsageCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component not found")
}

func TestUsageCommand_NoDirectory(t *testing.T) {
	// Setup test environment in a directory without .pluqqy
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	cmd := NewUsageCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"test"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .pluqqy directory found")
}