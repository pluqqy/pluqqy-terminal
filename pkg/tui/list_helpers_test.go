package tui

import (
	"strings"
	"testing"
)

// Test helper functions
func assertStringEqual(t *testing.T, got, want, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\ngot:  %q\nwant: %q", context, got, want)
	}
}

func assertIntsEqual(t *testing.T, got, want []int, context string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: length mismatch: got %d, want %d", context, len(got), len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s: [%d]: got %d, want %d", context, i, got[i], want[i])
		}
	}
}

// TestPluralize tests the pluralize function
func TestPluralize(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "zero count returns s",
			count: 0,
			want:  "s",
		},
		{
			name:  "single count returns empty",
			count: 1,
			want:  "",
		},
		{
			name:  "two count returns s",
			count: 2,
			want:  "s",
		},
		{
			name:  "multiple count returns s",
			count: 5,
			want:  "s",
		},
		{
			name:  "large count returns s",
			count: 100,
			want:  "s",
		},
		{
			name:  "negative count returns s",
			count: -1,
			want:  "s",
		},
		{
			name:  "negative multiple returns s",
			count: -5,
			want:  "s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluralize(tt.count)
			assertStringEqual(t, got, tt.want, "pluralize result")
		})
	}
}

// TestTruncateName tests the truncateName function
func TestTruncateName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{
			name:     "short name fits",
			input:    "test",
			maxWidth: 10,
			want:     "test",
		},
		{
			name:     "exact length minus 3",
			input:    "1234567",
			maxWidth: 10,
			want:     "1234567",
		},
		{
			name:     "name needs truncation",
			input:    "verylongname",
			maxWidth: 10,
			want:     "very...",
		},
		{
			name:     "empty string",
			input:    "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "max width 6 no truncation needed",
			input:    "hi",
			maxWidth: 6,
			want:     "hi",
		},
		{
			name:     "max width 7 with short name",
			input:    "test",
			maxWidth: 7,
			want:     "test",
		},
		{
			name:     "unicode characters",
			input:    "こんにちは世界です",
			maxWidth: 10,
			want:     "こ\xe3...",
		},
		{
			name:     "exact boundary case",
			input:    "123456",
			maxWidth: 9,
			want:     "123456",
		},
		{
			name:     "one char over boundary",
			input:    "1234567",
			maxWidth: 9,
			want:     "123...",
		},
		{
			name:     "long name with minimum valid maxWidth",
			input:    "testing123",
			maxWidth: 9,
			want:     "tes...",
		},
		{
			name:     "very long string",
			input:    strings.Repeat("a", 100),
			maxWidth: 20,
			want:     strings.Repeat("a", 14) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateName(tt.input, tt.maxWidth)
			assertStringEqual(t, got, tt.want, "truncateName result")
		})
	}
}

// TestFormatTableRow tests the formatTableRow function
func TestFormatTableRow(t *testing.T) {
	tests := []struct {
		name       string
		nameCol    string
		tags       string
		tokens     string
		usage      string
		nameWidth  int
		tagsWidth  int
		tokenWidth int
		usageWidth int
		want       string
	}{
		{
			name:       "with usage column",
			nameCol:    "test",
			tags:       "tag1",
			tokens:     "100",
			usage:      "5",
			nameWidth:  10,
			tagsWidth:  10,
			tokenWidth: 8,
			usageWidth: 6,
			want:       "test       tag1             100      5",
		},
		{
			name:       "without usage column",
			nameCol:    "test",
			tags:       "tag1",
			tokens:     "100",
			usage:      "",
			nameWidth:  10,
			tagsWidth:  10,
			tokenWidth: 8,
			usageWidth: 0,
			want:       "test       tag1            100",
		},
		{
			name:       "long tags need padding",
			nameCol:    "component",
			tags:       "production,critical",
			tokens:     "1500",
			usage:      "25",
			nameWidth:  15,
			tagsWidth:  20,
			tokenWidth: 8,
			usageWidth: 6,
			want:       "component       production,critical       1500     25",
		},
		{
			name:       "empty tags",
			nameCol:    "test",
			tags:       "",
			tokens:     "50",
			usage:      "",
			nameWidth:  10,
			tagsWidth:  10,
			tokenWidth: 8,
			usageWidth: 0,
			want:       "test                        50",
		},
		{
			name:       "unicode in name",
			nameCol:    "テスト",
			tags:       "jp",
			tokens:     "200",
			usage:      "10",
			nameWidth:  10,
			tagsWidth:  10,
			tokenWidth: 8,
			usageWidth: 6,
			want:       "テスト        jp               200     10",
		},
		{
			name:       "exact width columns",
			nameCol:    "exact",
			tags:       "exact",
			tokens:     "999",
			usage:      "99",
			nameWidth:  5,
			tagsWidth:  5,
			tokenWidth: 3,
			usageWidth: 2,
			want:       "exact exact  999 99",
		},
		{
			name:       "tags longer than width",
			nameCol:    "test",
			tags:       "verylongtag",
			tokens:     "100",
			usage:      "",
			nameWidth:  10,
			tagsWidth:  5,
			tokenWidth: 8,
			usageWidth: 0,
			want:       "test       verylongtag      100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTableRow(tt.nameCol, tt.tags, tt.tokens, tt.usage,
				tt.nameWidth, tt.tagsWidth, tt.tokenWidth, tt.usageWidth)
			assertStringEqual(t, got, tt.want, "formatTableRow result")
		})
	}
}

// TestPreprocessContent tests the preprocessContent function
func TestPreprocessContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal text unchanged",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "carriage return",
			input: "Hello\rWorld",
			want:  "Hello\nWorld",
		},
		{
			name:  "double carriage return",
			input: "Hello\r\rWorld",
			want:  "Hello\n\nWorld",
		},
		{
			name:  "carriage return newline",
			input: "Hello\r\nWorld",
			want:  "Hello\n\nWorld",
		},
		{
			name:  "mixed line endings",
			input: "A\rB\nC\r\nD\r\rE",
			want:  "A\nB\nC\n\nD\n\nE",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace with CR",
			input: "   \r\n   ",
			want:  "   \n\n   ",
		},
		{
			name:  "multiple CRs in sequence",
			input: "Start\r\r\r\rEnd",
			want:  "Start\n\n\n\nEnd",
		},
		{
			name:  "CR at start",
			input: "\rStart",
			want:  "\nStart",
		},
		{
			name:  "CR at end",
			input: "End\r",
			want:  "End\n",
		},
		{
			name:  "only CRs",
			input: "\r\r\r",
			want:  "\n\n\n",
		},
		{
			name:  "complex mix",
			input: "Line1\r\rLine2\r\nLine3\nLine4\r\r\nLine5",
			want:  "Line1\n\nLine2\n\nLine3\nLine4\n\n\nLine5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preprocessContent(tt.input)
			assertStringEqual(t, got, tt.want, "preprocessContent result")
		})
	}
}

// TestFormatColumnWidths tests the formatColumnWidths function
func TestFormatColumnWidths(t *testing.T) {
	tests := []struct {
		name           string
		totalWidth     int
		hasUsageColumn bool
		wantName       int
		wantTags       int
		wantToken      int
		wantUsage      int
	}{
		{
			name:           "components with usage column normal width",
			totalWidth:     100,
			hasUsageColumn: true,
			wantName:       40,
			wantTags:       40,
			wantToken:      8,
			wantUsage:      6,
		},
		{
			name:           "pipelines without usage column normal width",
			totalWidth:     100,
			hasUsageColumn: false,
			wantName:       48,
			wantTags:       39,
			wantToken:      8,
			wantUsage:      0,
		},
		{
			name:           "minimum widths enforced for components",
			totalWidth:     50,
			hasUsageColumn: true,
			wantName:       20,
			wantTags:       15,
			wantToken:      8,
			wantUsage:      6,
		},
		{
			name:           "minimum widths enforced for pipelines",
			totalWidth:     40,
			hasUsageColumn: false,
			wantName:       20,
			wantTags:       15,
			wantToken:      8,
			wantUsage:      0,
		},
		{
			name:           "very large width for components",
			totalWidth:     200,
			hasUsageColumn: true,
			wantName:       90,
			wantTags:       90,
			wantToken:      8,
			wantUsage:      6,
		},
		{
			name:           "very large width for pipelines",
			totalWidth:     200,
			hasUsageColumn: false,
			wantName:       103,
			wantTags:       84,
			wantToken:      8,
			wantUsage:      0,
		},
		{
			name:           "exact minimum total for components",
			totalWidth:     60,
			hasUsageColumn: true,
			wantName:       20,
			wantTags:       20,
			wantToken:      8,
			wantUsage:      6,
		},
		{
			name:           "exact minimum total for pipelines",
			totalWidth:     52,
			hasUsageColumn: false,
			wantName:       22,
			wantTags:       18,
			wantToken:      8,
			wantUsage:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotTags, gotToken, gotUsage := formatColumnWidths(tt.totalWidth, tt.hasUsageColumn)
			
			if gotName != tt.wantName {
				t.Errorf("nameWidth: got %d, want %d", gotName, tt.wantName)
			}
			if gotTags != tt.wantTags {
				t.Errorf("tagsWidth: got %d, want %d", gotTags, tt.wantTags)
			}
			if gotToken != tt.wantToken {
				t.Errorf("tokenWidth: got %d, want %d", gotToken, tt.wantToken)
			}
			if gotUsage != tt.wantUsage {
				t.Errorf("usageWidth: got %d, want %d", gotUsage, tt.wantUsage)
			}
		})
	}
}