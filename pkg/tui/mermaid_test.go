package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

func TestMermaidState(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MermaidState)
		expected bool
		check    func(*testing.T, *MermaidState)
	}{
		{
			name:  "initial state inactive",
			setup: func(ms *MermaidState) {},
			check: func(t *testing.T, ms *MermaidState) {
				if ms.Active {
					t.Error("Expected initial state to be inactive")
				}
				if ms.IsGenerating() {
					t.Error("Expected not generating initially")
				}
			},
		},
		{
			name: "start generation sets flags",
			setup: func(ms *MermaidState) {
				ms.StartGeneration()
			},
			check: func(t *testing.T, ms *MermaidState) {
				if !ms.IsGenerating() {
					t.Error("Expected generating flag to be true")
				}
				if ms.LastError != nil {
					t.Error("Expected no error after start")
				}
			},
		},
		{
			name: "complete generation updates state",
			setup: func(ms *MermaidState) {
				ms.StartGeneration()
				ms.CompleteGeneration("test.html")
			},
			check: func(t *testing.T, ms *MermaidState) {
				if ms.IsGenerating() {
					t.Error("Expected generating to be false after completion")
				}
				if ms.LastGeneratedFile != "test.html" {
					t.Errorf("Expected LastGeneratedFile to be 'test.html', got %s", ms.LastGeneratedFile)
				}
			},
		},
		{
			name: "fail generation sets error",
			setup: func(ms *MermaidState) {
				ms.StartGeneration()
				ms.FailGeneration(fmt.Errorf("test error"))
			},
			check: func(t *testing.T, ms *MermaidState) {
				if ms.IsGenerating() {
					t.Error("Expected generating to be false after failure")
				}
				if ms.LastError == nil {
					t.Error("Expected error to be set")
				}
				if ms.LastError.Error() != "test error" {
					t.Errorf("Expected error 'test error', got %v", ms.LastError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMermaidState()
			tt.setup(ms)
			tt.check(t, ms)
		})
	}
}

func TestMermaidStateHandleInput(t *testing.T) {
	tests := []struct {
		name        string
		active      bool
		generating  bool
		key         string
		wantHandled bool
	}{
		{
			name:        "inactive state ignores input",
			active:      false,
			generating:  false,
			key:         "esc",
			wantHandled: false,
		},
		{
			name:        "active non-generating ignores escape",
			active:      true,
			generating:  false,
			key:         "esc",
			wantHandled: false,
		},
		{
			name:        "active generating handles escape",
			active:      true,
			generating:  true,
			key:         "esc",
			wantHandled: true,
		},
		{
			name:        "active generating ignores other keys",
			active:      true,
			generating:  true,
			key:         "a",
			wantHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMermaidState()
			ms.Active = tt.active
			if tt.generating {
				ms.StartGeneration()
			}

			handled, _ := ms.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if handled != tt.wantHandled {
				t.Errorf("HandleInput() handled = %v, want %v", handled, tt.wantHandled)
			}

			// If escape was handled, generation should be stopped
			if tt.wantHandled && tt.key == "esc" {
				if ms.IsGenerating() {
					t.Error("Expected generation to be stopped after escape")
				}
			}
		})
	}
}

func TestGroupComponentsByType(t *testing.T) {
	tests := []struct {
		name             string
		components       []models.ComponentRef
		expectedContexts int
		expectedPrompts  int
		expectedRules    int
	}{
		{
			name:             "empty components",
			components:       []models.ComponentRef{},
			expectedContexts: 0,
			expectedPrompts:  0,
			expectedRules:    0,
		},
		{
			name: "mixed component types",
			components: []models.ComponentRef{
				{Type: "contexts", Path: "ctx1.md"},
				{Type: "prompts", Path: "p1.md"},
				{Type: "rules", Path: "r1.md"},
				{Type: "contexts", Path: "ctx2.md"},
			},
			expectedContexts: 2,
			expectedPrompts:  1,
			expectedRules:    1,
		},
		{
			name: "single type only",
			components: []models.ComponentRef{
				{Type: "prompts", Path: "p1.md"},
				{Type: "prompts", Path: "p2.md"},
				{Type: "prompts", Path: "p3.md"},
			},
			expectedContexts: 0,
			expectedPrompts:  3,
			expectedRules:    0,
		},
		{
			name: "unknown type ignored",
			components: []models.ComponentRef{
				{Type: "contexts", Path: "ctx1.md"},
				{Type: "unknown", Path: "unk.md"},
				{Type: "prompts", Path: "p1.md"},
			},
			expectedContexts: 1,
			expectedPrompts:  1,
			expectedRules:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := &MermaidOperator{state: NewMermaidState()}
			contexts, prompts, rules := mo.groupComponentsByType(tt.components)

			if len(contexts) != tt.expectedContexts {
				t.Errorf("Expected %d contexts, got %d", tt.expectedContexts, len(contexts))
			}
			if len(prompts) != tt.expectedPrompts {
				t.Errorf("Expected %d prompts, got %d", tt.expectedPrompts, len(prompts))
			}
			if len(rules) != tt.expectedRules {
				t.Errorf("Expected %d rules, got %d", tt.expectedRules, len(rules))
			}
		})
	}
}

func TestExtractComponentName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple filename",
			path:     "component.md",
			expected: "component",
		},
		{
			name:     "full path",
			path:     "/path/to/components/my-component.md",
			expected: "my-component",
		},
		{
			name:     "no extension",
			path:     "component",
			expected: "component",
		},
		{
			name:     "multiple dots",
			path:     "my.component.test.md",
			expected: "my.component.test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentName(tt.path)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "spaces replaced",
			input:    "my pipeline name",
			expected: "my-pipeline-name",
		},
		{
			name:     "special chars replaced",
			input:    "test/pipeline:with*chars",
			expected: "test-pipeline-with-chars",
		},
		{
			name:     "already safe",
			input:    "safe-filename",
			expected: "safe-filename",
		},
		{
			name:     "all special chars",
			input:    `test/\:*?"<>|name`,
			expected: "test---------name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty string",
			content:  "",
			expected: 0,
		},
		{
			name:     "short text",
			content:  "test",
			expected: 1, // 4 chars / 4 = 1
		},
		{
			name:     "longer text",
			content:  strings.Repeat("a", 100),
			expected: 25, // 100 / 4 = 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTokens(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %d tokens, got %d", tt.expected, result)
			}
		})
	}
}

func TestGenerateMermaidGraph(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *models.Pipeline
		checks   []string // Strings that should appear in output
	}{
		{
			name: "empty pipeline",
			pipeline: &models.Pipeline{
				Name:       "Empty Pipeline",
				Components: []models.ComponentRef{},
			},
			checks: []string{
				"graph TD",
				"EMPTY PIPELINE",
				"0 Components",
			},
		},
		{
			name: "pipeline with all types",
			pipeline: &models.Pipeline{
				Name: "Full Pipeline",
				Components: []models.ComponentRef{
					{Type: "contexts", Path: "ctx.md"},
					{Type: "prompts", Path: "prompt.md"},
					{Type: "rules", Path: "rule.md"},
				},
			},
			checks: []string{
				"graph TD",
				"FULL PIPELINE",
				"3 Components",
				"subgraph CONTEXTS", // Based on default settings "## CONTEXTS"
				"subgraph PROMPTS",  // Based on default settings "## PROMPTS"
				"subgraph RULES",    // Based on default settings "## RULES"
				"classDef pipeline",
				"classDef context",
				"classDef prompt",
				"classDef rules",
			},
		},
		{
			name: "pipeline with only prompts",
			pipeline: &models.Pipeline{
				Name: "Prompts Only",
				Components: []models.ComponentRef{
					{Type: "prompts", Path: "p1.md"},
					{Type: "prompts", Path: "p2.md"},
				},
			},
			checks: []string{
				"graph TD",
				"PROMPTS ONLY",
				"subgraph PROMPTS",
				"P1[",
				"P2[",
				"Pipeline --- PROMPTS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := &MermaidOperator{state: NewMermaidState()}
			result := mo.generateMermaidGraph(tt.pipeline)

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Expected output to contain '%s'", check)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateMermaidGraph(b *testing.B) {
	pipeline := &models.Pipeline{
		Name: "Benchmark Pipeline",
		Components: []models.ComponentRef{
			{Type: "contexts", Path: "c1.md"},
			{Type: "contexts", Path: "c2.md"},
			{Type: "prompts", Path: "p1.md"},
			{Type: "prompts", Path: "p2.md"},
			{Type: "prompts", Path: "p3.md"},
			{Type: "rules", Path: "r1.md"},
		},
	}

	mo := &MermaidOperator{state: NewMermaidState()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mo.generateMermaidGraph(pipeline)
	}
}

func BenchmarkSanitizeFilename(b *testing.B) {
	testName := "test/pipeline:with*special?chars"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeFilename(testName)
	}
}

func BenchmarkEstimateTokens(b *testing.B) {
	content := strings.Repeat("This is a test content for token estimation. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = estimateTokens(content)
	}
}
