package utils

import (
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		margin   int // Allow margin of error
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
			margin:   0,
		},
		{
			name:     "Simple sentence",
			input:    "The quick brown fox jumps over the lazy dog.",
			expected: 10, // ~9-11 tokens
			margin:   2,
		},
		{
			name:     "Code block",
			input:    "```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```",
			expected: 20, // Code tends to have more tokens
			margin:   5,
		},
		{
			name:     "Mixed content",
			input:    "# Title\n\nThis is a paragraph with some **bold** text.\n\n```python\nprint('test')\n```",
			expected: 25,
			margin:   5,
		},
		{
			name:     "Long technical document",
			input:    `# Technical Documentation

This is a comprehensive guide to understanding token estimation in language models.

## Introduction

Token counting is essential for managing context limits in LLMs. Each model has specific tokenization rules that affect how text is broken down into tokens.

### Key Points:
- One token ≈ 4 characters in English
- One token ≈ ¾ words on average
- Punctuation and special characters affect tokenization
- Code typically has higher token density

## Code Example

` + "```python\ndef calculate_tokens(text):\n    # Simplified token estimation\n    return len(text) / 4\n```" + `

## Conclusion

Understanding token limits helps optimize prompt engineering and avoid context overflow.`,
			expected: 180,
			margin:   20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.input)
			diff := abs(result - tt.expected)
			if diff > tt.margin {
				t.Errorf("EstimateTokens() = %d, expected %d ±%d (diff: %d)", 
					result, tt.expected, tt.margin, diff)
			}
		})
	}
}

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		tokens   int
		expected string
	}{
		{100, "~100 tokens"},
		{999, "~999 tokens"},
		{1000, "~1.0K tokens"},
		{1500, "~1.5K tokens"},
		{9999, "~10.0K tokens"},
		{10000, "~10K tokens"},
		{150000, "~150K tokens"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatTokenCount(tt.tokens)
			if result != tt.expected {
				t.Errorf("FormatTokenCount(%d) = %s, expected %s", 
					tt.tokens, result, tt.expected)
			}
		})
	}
}

func TestGetTokenLimitStatus(t *testing.T) {
	tests := []struct {
		tokens         int
		expectedStatus string
		expectedLimit  int
	}{
		{1000, "good", 4096},      // 24% of 4K = good
		{3500, "danger", 4096},    // 85% of 4K = danger
		{7000, "danger", 8192},    // 85% of 8K = danger
		{15000, "danger", 16384},  // 91% of 16K = danger
		{100000, "warning", 131072}, // 76% of 128K = warning
	}

	for _, tt := range tests {
		t.Run(tt.expectedStatus, func(t *testing.T) {
			_, limit, status := GetTokenLimitStatus(tt.tokens)
			if status != tt.expectedStatus {
				t.Errorf("GetTokenLimitStatus(%d) status = %s, expected %s", 
					tt.tokens, status, tt.expectedStatus)
			}
			if limit != tt.expectedLimit {
				t.Errorf("GetTokenLimitStatus(%d) limit = %d, expected %d", 
					tt.tokens, limit, tt.expectedLimit)
			}
		})
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}