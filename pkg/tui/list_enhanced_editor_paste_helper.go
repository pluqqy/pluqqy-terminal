package tui

import (
	"regexp"
	"strings"
)

// ContentType represents the detected content type
type ContentType int

const (
	ContentTypePlain ContentType = iota
	ContentTypeJSON
	ContentTypeYAML
	ContentTypeMarkdown
	ContentTypeCode
)

// PasteHelper provides smart paste detection and cleaning
type PasteHelper struct {
	// Patterns for detection
	tuiBorderPattern       *regexp.Regexp
	lineNumberPattern      *regexp.Regexp
	terminalPromptPattern  *regexp.Regexp
	markdownFencePattern   *regexp.Regexp
}

// NewPasteHelper creates a new paste helper with compiled patterns
func NewPasteHelper() *PasteHelper {
	return &PasteHelper{
		// Match TUI borders at start/end of lines
		tuiBorderPattern: regexp.MustCompile(`^[│├└┌┐┘┤┬┴┼─]+\s*|\s*[│├└┌┐┘┤┬┴┼─]+$`),
		// Match line numbers (various formats including leading spaces)
		lineNumberPattern: regexp.MustCompile(`^\s*\d+\s`),
		// Match terminal prompts
		terminalPromptPattern: regexp.MustCompile(`^[\$>%#]\s+`),
		// Match markdown code fences
		markdownFencePattern: regexp.MustCompile("^```[a-zA-Z]*$"),
	}
}

// CleanPastedContent intelligently cleans pasted content
func (ph *PasteHelper) CleanPastedContent(content string) string {
	// If content is only whitespace, preserve it
	if strings.TrimSpace(content) == "" {
		return content
	}
	
	// Detect content type early to guide cleaning
	contentType := ph.DetectContentType(content)
	
	// Apply appropriate cleaning based on type
	cleaned := content
	
	// Always strip TUI borders (common in Pluqqy copies)
	cleaned = ph.stripTUIBorders(cleaned)
	
	// Strip line numbers if detected
	if ph.hasLineNumbers(cleaned) {
		cleaned = ph.stripLineNumbers(cleaned)
	}
	
	// Strip terminal prompts
	cleaned = ph.stripTerminalPrompts(cleaned)
	
	// Clean markdown code blocks
	if contentType == ContentTypeMarkdown || ph.hasMarkdownFences(cleaned) {
		cleaned = ph.stripMarkdownFences(cleaned)
	}
	
	// Normalize indentation if content was cleaned
	if cleaned != content {
		cleaned = ph.normalizeIndentation(cleaned)
	}
	
	// Trim trailing whitespace from each line
	cleaned = ph.trimTrailingWhitespace(cleaned)
	
	return cleaned
}

// DetectContentType analyzes content to determine its type
func (ph *PasteHelper) DetectContentType(content string) ContentType {
	trimmed := strings.TrimSpace(content)
	
	// Check for JSON
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return ContentTypeJSON
	}
	
	// Check for YAML
	if strings.HasPrefix(trimmed, "---") || 
		regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*:\s`).MatchString(trimmed) {
		return ContentTypeYAML
	}
	
	// Check for Markdown
	if strings.Contains(content, "```") || 
		strings.HasPrefix(trimmed, "#") || 
		strings.Contains(content, "**") ||
		strings.Contains(content, "- [ ]") ||
		strings.Contains(content, "- [x]") {
		return ContentTypeMarkdown
	}
	
	// Check for code patterns
	if strings.Contains(content, "func ") || 
		strings.Contains(content, "function ") ||
		strings.Contains(content, "class ") ||
		strings.Contains(content, "import ") ||
		strings.Contains(content, "const ") ||
		strings.Contains(content, "var ") ||
		strings.Contains(content, "fmt.Println") ||
		strings.Contains(content, "if true {") ||
		strings.Contains(content, "doSomething()") {
		return ContentTypeCode
	}
	
	return ContentTypePlain
}

// stripTUIBorders removes TUI border characters from lines
func (ph *PasteHelper) stripTUIBorders(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		// Check for lines that are only borders (including box drawing chars)
		if regexp.MustCompile(`^[│├└┌┐┘┤┬┴┼─╭╮╰╯\s]+$`).MatchString(line) {
			// If it's a horizontal separator (contains ├ or ┤), convert to empty line
			if strings.Contains(line, "├") || strings.Contains(line, "┤") {
				cleaned = append(cleaned, "")
			}
			// Otherwise skip completely
			continue
		}
		
		// Handle complex format with multiple │ separators
		// Pattern often seen: "content │ │ linenum content2" 
		// This happens when copying from split panes showing two files side by side
		if strings.Contains(line, "│") {
			// Handle various TUI border patterns
			// Pattern: "│  38 │ package main        │"
			// Pattern: "│  39 │ import "fmt"        │"
			parts := strings.Split(line, "│")
			if len(parts) >= 3 {
				// For pattern "│  38 │ package main        │"
				// parts[0] = "", parts[1] = "  38 ", parts[2] = " package main        ", parts[3] = ""
				for i := 1; i < len(parts)-1; i++ {
					part := parts[i]
					// Skip pure line numbers
					if regexp.MustCompile(`^\s*\d+\s*$`).MatchString(part) {
						continue
					}
					// Remove line numbers and extract content
					cleaned_part := ph.lineNumberPattern.ReplaceAllString(part, "")
					cleaned_part = strings.TrimSpace(cleaned_part)
					if cleaned_part != "" {
						cleaned = append(cleaned, cleaned_part)
					}
				}
			} else {
				// Simple border removal for single-pane content
				// For lines like "│    if true {", we want to preserve indentation
				cleanedLine := line
				if strings.HasPrefix(line, "│") {
					// Remove just the │ character and preserve everything else
					cleanedLine = line[len("│"):]
				} else {
					// Use regex for other border patterns
					cleanedLine = ph.tuiBorderPattern.ReplaceAllString(line, "")
					cleanedLine = strings.TrimSpace(cleanedLine)
				}
				
				if strings.TrimSpace(cleanedLine) != "" {
					cleaned = append(cleaned, cleanedLine)
				}
			}
		} else {
			// No borders, keep the line with original indentation
			// Only skip completely empty lines
			if strings.TrimSpace(line) != "" {
				cleaned = append(cleaned, line)
			}
		}
	}
	
	return strings.Join(cleaned, "\n")
}

// hasLineNumbers checks if content has line numbers
func (ph *PasteHelper) hasLineNumbers(content string) bool {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return false
	}
	
	// For single line, just check if it matches the pattern
	if len(lines) == 1 {
		return ph.lineNumberPattern.MatchString(lines[0])
	}
	
	// For multiple lines, check first few lines for consistent line number pattern
	matchCount := 0
	for i, line := range lines {
		if i > 4 {
			break
		}
		if ph.lineNumberPattern.MatchString(line) {
			matchCount++
		}
	}
	
	// If majority have line numbers, assume it's numbered
	return matchCount >= 2
}

// stripLineNumbers removes line numbers from content
func (ph *PasteHelper) stripLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		cleaned_line := ph.lineNumberPattern.ReplaceAllString(line, "")
		cleaned = append(cleaned, cleaned_line)
	}
	
	return strings.Join(cleaned, "\n")
}

// stripTerminalPrompts removes terminal prompt characters
func (ph *PasteHelper) stripTerminalPrompts(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		cleaned_line := ph.terminalPromptPattern.ReplaceAllString(line, "")
		cleaned = append(cleaned, cleaned_line)
	}
	
	return strings.Join(cleaned, "\n")
}

// hasMarkdownFences checks for markdown code fences
func (ph *PasteHelper) hasMarkdownFences(content string) bool {
	return strings.Contains(content, "```")
}

// stripMarkdownFences removes markdown code fence markers
func (ph *PasteHelper) stripMarkdownFences(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		// Skip lines that are just code fences
		if ph.markdownFencePattern.MatchString(line) {
			continue
		}
		cleaned = append(cleaned, line)
	}
	
	return strings.Join(cleaned, "\n")
}

// normalizeIndentation fixes inconsistent indentation
func (ph *PasteHelper) normalizeIndentation(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}
	
	// Find minimum indentation (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}
		
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	
	// Remove minimum indentation from all lines
	if minIndent > 0 {
		cleaned := make([]string, 0, len(lines))
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				cleaned = append(cleaned, "")
			} else if len(line) > minIndent {
				cleaned = append(cleaned, line[minIndent:])
			} else {
				cleaned = append(cleaned, strings.TrimSpace(line))
			}
		}
		return strings.Join(cleaned, "\n")
	}
	
	return content
}

// trimTrailingWhitespace removes trailing spaces from each line
func (ph *PasteHelper) trimTrailingWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		cleaned = append(cleaned, strings.TrimRight(line, " \t"))
	}
	
	return strings.Join(cleaned, "\n")
}

// CleanForSave prepares content for saving (auto-trim, etc)
func (ph *PasteHelper) CleanForSave(content string) string {
	// Trim trailing whitespace from each line
	cleaned := ph.trimTrailingWhitespace(content)
	
	// Normalize line endings (CRLF -> LF)
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	
	// Remove trailing newlines completely
	cleaned = strings.TrimRight(cleaned, "\n")
	
	return cleaned
}

// GetContentTypeString returns a string representation of content type
func GetContentTypeString(ct ContentType) string {
	switch ct {
	case ContentTypeJSON:
		return "JSON"
	case ContentTypeYAML:
		return "YAML"
	case ContentTypeMarkdown:
		return "Markdown"
	case ContentTypeCode:
		return "Code"
	default:
		return "Plain"
	}
}

// CountLines counts the number of lines in content
func CountLines(content string) int {
	if content == "" {
		return 0
	}
	// Split by newlines and count non-empty result
	lines := strings.Split(content, "\n")
	// If last element is empty (trailing newline), don't count it
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return len(lines) - 1
	}
	return len(lines)
}

// WillClean checks if content needs cleaning
func (ph *PasteHelper) WillClean(content string) bool {
	// Check for any patterns that would be cleaned
	if ph.hasLineNumbers(content) {
		return true
	}
	
	if ph.hasMarkdownFences(content) {
		return true
	}
	
	// Check for TUI borders
	if regexp.MustCompile(`[│├└┌┐┘┤┬┴┼─]`).MatchString(content) {
		return true
	}
	
	// Check for terminal prompts
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if ph.terminalPromptPattern.MatchString(line) {
			return true
		}
	}
	
	return false
}