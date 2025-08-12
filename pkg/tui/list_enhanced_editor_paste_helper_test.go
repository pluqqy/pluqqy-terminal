package tui

import (
	"testing"
)

func TestPasteHelper_CleanPastedContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text unchanged",
			input:    "This is plain text",
			expected: "This is plain text",
		},
		{
			name:     "remove TUI borders",
			input:    "│ Some content │",
			expected: "Some content",
		},
		{
			name: "remove line numbers",
			input: `  1 func main() {
  2     fmt.Println("Hello")
  3 }`,
			expected: `func main() {
    fmt.Println("Hello")
}`,
		},
		{
			name: "complex TUI format with multiple panes",
			input: `│  38 │ package main        │
│  39 │ import "fmt"        │`,
			expected: `package main
import "fmt"`,
		},
		{
			name: "terminal prompts",
			input: `$ go build
> npm install
# sudo command`,
			expected: `go build
npm install
sudo command`,
		},
		{
			name:     "markdown code fence",
			input:    "```go\nfunc test() {}\n```",
			expected: "func test() {}",
		},
		{
			name: "preserve indentation",
			input: `│    if true {
│        doSomething()
│    }`,
			expected: `if true {
    doSomething()
}`,
		},
		{
			name: "mixed borders and content",
			input: `╭──────────────╮
│ Title Text   │
├──────────────┤
│ Body Content │
╰──────────────╯`,
			expected: `Title Text

Body Content`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n   \n   ",
			expected: "   \n   \n   ",
		},
	}

	helper := NewPasteHelper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.CleanPastedContent(tt.input)
			if got != tt.expected {
				t.Errorf("CleanPastedContent() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPasteHelper_WillClean(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "plain text won't clean",
			content:  "This is plain text",
			expected: false,
		},
		{
			name:     "TUI borders will clean",
			content:  "│ Some content │",
			expected: true,
		},
		{
			name:     "line numbers will clean",
			content:  "  1 func main() {",
			expected: true,
		},
		{
			name:     "terminal prompt will clean",
			content:  "$ go build",
			expected: true,
		},
		{
			name:     "markdown fence will clean",
			content:  "```go",
			expected: true,
		},
		{
			name:     "empty content won't clean",
			content:  "",
			expected: false,
		},
	}

	helper := NewPasteHelper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.WillClean(tt.content)
			if got != tt.expected {
				t.Errorf("WillClean() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPasteHelper_CleanForSave(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim trailing whitespace",
			input:    "content   \n",
			expected: "content",
		},
		{
			name:     "preserve internal whitespace",
			input:    "line 1\n  indented\nline 3",
			expected: "line 1\n  indented\nline 3",
		},
		{
			name:     "remove multiple trailing newlines",
			input:    "content\n\n\n",
			expected: "content",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n   ",
			expected: "",
		},
	}

	helper := NewPasteHelper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.CleanForSave(tt.input)
			if got != tt.expected {
				t.Errorf("CleanForSave() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"", 0},
		{"single line", 1},
		{"line 1\nline 2", 2},
		{"line 1\nline 2\nline 3", 3},
		{"trailing newline\n", 1},
		{"\n\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := CountLines(tt.content)
			if got != tt.expected {
				t.Errorf("CountLines(%q) = %d, want %d", tt.content, got, tt.expected)
			}
		})
	}
}