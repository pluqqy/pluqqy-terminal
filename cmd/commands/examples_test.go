package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test main command execution
func TestExamplesCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		setupFunc func(t *testing.T) string
		wantErr   bool
		errMsg    string
		contains  []string
		excludes  []string
	}{
		{
			name: "list general examples by default",
			flags: map[string]string{
				"list": "true",
			},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Available examples (all categories)",
				"[general]",
				"[web]",
				"[ai]",
				"[claude]",
				"General Development",
				"Web Development",
				"AI Assistant Optimization",
				"CLAUDE.md Distiller",
			},
		},
		{
			name: "list specific category",
			args: []string{"web"},
			flags: map[string]string{
				"list": "true",
			},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Available examples in category 'web'",
				"Web Development",
				"React Architecture",
				"REST API Design",
			},
			excludes: []string{
				"[general]",
				"[claude]",
			},
		},
		{
			name:      "install general examples (default)",
			args:      []string{},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Installing general examples",
				"General Development",
				"✓ Installed contexts/example-project-overview.md",
				"✓ Installed pipeline example-feature-development.yaml",
				"✨ Installation complete!",
			},
		},
		{
			name:      "install web examples",
			args:      []string{"web"},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Installing web examples",
				"Web Development",
				"✓ Installed contexts/example-react-architecture.md",
				"✓ Installed pipeline example-frontend-feature.yaml",
			},
		},
		{
			name:      "install claude examples with migration tip",
			args:      []string{"claude"},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Installing claude examples",
				"CLAUDE.md Distiller",
				"🔄 CLAUDE.md Migration:",
				"Use 'pluqqy set example-claude-distiller'",
			},
		},
		{
			name:      "install all examples",
			args:      []string{"all"},
			setupFunc: setupExamplesTest,
			wantErr:   false,
			contains: []string{
				"Installing all examples",
				"General Development",
				"Web Development",
				"AI Assistant Optimization",
				"CLAUDE.md Distiller",
			},
		},
		{
			name:      "invalid category error",
			args:      []string{"invalid"},
			setupFunc: setupExamplesTest,
			wantErr:   true,
			errMsg:    "invalid category 'invalid'",
			contains: []string{
				"Valid categories: general, web, ai, claude, all",
			},
		},
		{
			name:    "no .pluqqy directory error",
			args:    []string{},
			wantErr: true,
			errMsg:  "no .pluqqy directory found",
			contains: []string{
				"Run 'pluqqy init' first",
			},
		},
		// Note: quiet flag is a global flag and would need to be set differently
		// Skipping this test for now as it requires modifying the command setup
		{
			name: "force overwrite existing files",
			args: []string{"general"},
			flags: map[string]string{
				"force": "true",
			},
			setupFunc: setupExamplesTestWithExisting,
			wantErr:   false,
			contains: []string{
				"✓ Installed contexts/example-project-overview.md",
			},
			excludes: []string{
				"Skipped",
				"already exists",
			},
		},
		{
			name:      "skip existing files without force",
			args:      []string{"general"},
			setupFunc: setupExamplesTestWithExisting,
			wantErr:   false,
			contains: []string{
				"⚠️  Skipped",
				"already exists, use --force to overwrite",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var dir string
			if tt.setupFunc != nil {
				dir = tt.setupFunc(t)
				oldDir, _ := os.Getwd()
				require.NoError(t, os.Chdir(dir))
				defer os.Chdir(oldDir)
			} else {
				// For tests that expect no .pluqqy directory
				dir = t.TempDir()
				oldDir, _ := os.Getwd()
				require.NoError(t, os.Chdir(dir))
				defer os.Chdir(oldDir)
			}

			// Create command
			cmd := NewExamplesCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Set flags
			for key, value := range tt.flags {
				require.NoError(t, cmd.Flags().Set(key, value))
			}

			// Execute
			err := cmd.Execute()

			// Check error
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}

			// Check output
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
			for _, excluded := range tt.excludes {
				assert.NotContains(t, output, excluded, "Output should not contain: %s", excluded)
			}
		})
	}
}

// Test flag validation
func TestExamplesFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "list and category flags work together",
			flags: map[string]string{
				"list":     "true",
				"category": "web",
			},
			wantErr: false,
		},
		{
			name: "force flag requires no error",
			flags: map[string]string{
				"force": "true",
			},
			wantErr: false,
		},
		{
			name: "invalid category via flag",
			flags: map[string]string{
				"category": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupExamplesTest(t)
			oldDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(dir))
			defer os.Chdir(oldDir)

			cmd := NewExamplesCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			for key, value := range tt.flags {
				require.NoError(t, cmd.Flags().Set(key, value))
			}

			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test file creation
func TestExamplesFileCreation(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		expectedFiles []string
	}{
		{
			name:     "general creates correct files",
			category: "general",
			expectedFiles: []string{
				".pluqqy/components/contexts/example-project-overview.md",
				".pluqqy/components/contexts/example-code-architecture.md",
				".pluqqy/components/prompts/example-implement-feature.md",
				".pluqqy/components/rules/example-coding-standards.md",
				".pluqqy/pipelines/example-feature-development.yaml",
			},
		},
		{
			name:     "web creates correct files",
			category: "web",
			expectedFiles: []string{
				".pluqqy/components/contexts/example-react-architecture.md",
				".pluqqy/components/prompts/example-create-react-component.md",
				".pluqqy/components/rules/example-web-accessibility.md",
				".pluqqy/pipelines/example-frontend-feature.yaml",
			},
		},
		{
			name:     "ai creates correct files",
			category: "ai",
			expectedFiles: []string{
				".pluqqy/components/contexts/example-codebase-overview.md",
				".pluqqy/components/prompts/example-explain-code.md",
				".pluqqy/components/rules/example-concise-responses.md",
				".pluqqy/pipelines/example-ai-assistant-setup.yaml",
			},
		},
		{
			name:     "claude creates correct files",
			category: "claude",
			expectedFiles: []string{
				".pluqqy/components/contexts/example-claude-parser.md",
				".pluqqy/components/prompts/example-extract-to-pluqqy.md",
				".pluqqy/components/rules/example-preserve-intent.md",
				".pluqqy/pipelines/example-claude-distiller.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupExamplesTest(t)
			oldDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(dir))
			defer os.Chdir(oldDir)

			cmd := NewExamplesCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{tt.category})

			require.NoError(t, cmd.Execute())

			// Check that expected files were created
			for _, file := range tt.expectedFiles {
				path := filepath.Join(dir, file)
				assert.FileExists(t, path, "Expected file %s to be created", file)

				// Verify file has content
				content, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.NotEmpty(t, content, "File %s should have content", file)

				// Verify placeholders exist in content (skip rules files as they may not have placeholders)
				if strings.HasSuffix(file, ".md") && !strings.Contains(file, "/rules/") {
					// Check for placeholders in contexts and prompts
					if strings.Contains(file, "/contexts/") || strings.Contains(file, "/prompts/") {
						assert.Contains(t, string(content), "{{", "File %s should contain placeholders", file)
					}
				}
			}
		})
	}
}

// Test listing output formats
func TestExamplesListOutput(t *testing.T) {
	tests := []struct {
		name     string
		category string
		contains []string
	}{
		{
			name:     "list shows component counts",
			category: "general",
			contains: []string{
				"Components:",
				"• Project Overview (contexts)",
				"• Implement Feature (prompts)",
				"• Coding Standards (rules)",
				"Pipelines:",
				"• Feature Development",
			},
		},
		{
			name:     "list all shows all categories",
			category: "all",
			contains: []string{
				"[general]",
				"[web]",
				"[ai]",
				"[claude]",
				"To install all examples, run: pluqqy examples all",
				"To install a specific category, run: pluqqy examples <category>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupExamplesTest(t)
			oldDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(dir))
			defer os.Chdir(oldDir)

			cmd := NewExamplesCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			
			// Use list flag
			require.NoError(t, cmd.Flags().Set("list", "true"))
			
			if tt.category != "all" {
				cmd.SetArgs([]string{tt.category})
			}

			require.NoError(t, cmd.Execute())

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// Test error conditions
func TestExamplesErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "handles missing .pluqqy directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			args:    []string{"general"},
			wantErr: true,
			errMsg:  "no .pluqqy directory found",
		},
		{
			name: "handles invalid category",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, ".pluqqy"), 0755))
				return dir
			},
			args:    []string{"nonexistent"},
			wantErr: true,
			errMsg:  "invalid category 'nonexistent'",
		},
		{
			name: "handles read-only directory gracefully",
			setup: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("Cannot test read-only as root")
				}
				dir := setupExamplesTest(t)
				// Make components directory read-only
				componentsDir := filepath.Join(dir, ".pluqqy", "components", "contexts")
				require.NoError(t, os.Chmod(componentsDir, 0555))
				t.Cleanup(func() {
					os.Chmod(componentsDir, 0755)
				})
				return dir
			},
			args:    []string{"general"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			oldDir, _ := os.Getwd()
			require.NoError(t, os.Chdir(dir))
			defer os.Chdir(oldDir)

			cmd := NewExamplesCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test helpers
func setupExamplesTest(t *testing.T) string {
	dir := t.TempDir()
	
	// Create .pluqqy structure
	dirs := []string{
		".pluqqy/pipelines",
		".pluqqy/components/contexts",
		".pluqqy/components/prompts", 
		".pluqqy/components/rules",
		".pluqqy/archive",
	}
	
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0755))
	}
	
	return dir
}

func setupExamplesTestWithExisting(t *testing.T) string {
	dir := setupExamplesTest(t)
	
	// Create an existing example file
	existingFile := filepath.Join(dir, ".pluqqy", "components", "contexts", "example-project-overview.md")
	require.NoError(t, os.WriteFile(existingFile, []byte("existing content"), 0644))
	
	return dir
}

// Benchmark tests
func BenchmarkExamplesListCommand(b *testing.B) {
	dir := setupExamplesTestBench(b)
	oldDir, _ := os.Getwd()
	require.NoError(b, os.Chdir(dir))
	defer os.Chdir(oldDir)

	cmd := NewExamplesCommand()
	buf := new(bytes.Buffer)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"--list"})
		cmd.Execute()
	}
}

func BenchmarkExamplesInstallCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dir := setupExamplesTestBench(b)
		oldDir, _ := os.Getwd()
		os.Chdir(dir)
		
		cmd := NewExamplesCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"general"})
		
		b.StartTimer()
		cmd.Execute()
		b.StopTimer()
		
		os.Chdir(oldDir)
		os.RemoveAll(dir)
	}
}

func setupExamplesTestBench(b *testing.B) string {
	dir := b.TempDir()
	
	dirs := []string{
		".pluqqy/pipelines",
		".pluqqy/components/contexts",
		".pluqqy/components/prompts",
		".pluqqy/components/rules",
	}
	
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			b.Fatal(err)
		}
	}
	
	return dir
}

// Integration test
func TestExamplesIntegration(t *testing.T) {
	// Full workflow test
	dir := setupExamplesTest(t)
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir)

	// Step 1: List available examples
	listCmd := NewExamplesCommand()
	listBuf := new(bytes.Buffer)
	listCmd.SetOut(listBuf)
	listCmd.SetErr(listBuf)
	require.NoError(t, listCmd.Flags().Set("list", "true"))
	require.NoError(t, listCmd.Execute())
	assert.Contains(t, listBuf.String(), "Available examples")

	// Step 2: Install general examples
	installCmd := NewExamplesCommand()
	installBuf := new(bytes.Buffer)
	installCmd.SetOut(installBuf)
	installCmd.SetErr(installBuf)
	installCmd.SetArgs([]string{"general"})
	require.NoError(t, installCmd.Execute())
	assert.Contains(t, installBuf.String(), "Installation complete")

	// Step 3: Verify files were created
	assert.FileExists(t, filepath.Join(dir, ".pluqqy", "components", "contexts", "example-project-overview.md"))
	assert.FileExists(t, filepath.Join(dir, ".pluqqy", "pipelines", "example-feature-development.yaml"))

	// Step 4: Try to install again without force (should skip)
	reinstallCmd := NewExamplesCommand()
	reinstallBuf := new(bytes.Buffer)
	reinstallCmd.SetOut(reinstallBuf)
	reinstallCmd.SetErr(reinstallBuf)
	reinstallCmd.SetArgs([]string{"general"})
	require.NoError(t, reinstallCmd.Execute())
	assert.Contains(t, reinstallBuf.String(), "Skipped")

	// Step 5: Force reinstall
	forceCmd := NewExamplesCommand()
	forceBuf := new(bytes.Buffer)
	forceCmd.SetOut(forceBuf)
	forceCmd.SetErr(forceBuf)
	forceCmd.SetArgs([]string{"general"})
	require.NoError(t, forceCmd.Flags().Set("force", "true"))
	require.NoError(t, forceCmd.Execute())
	assert.NotContains(t, forceBuf.String(), "Skipped")
}