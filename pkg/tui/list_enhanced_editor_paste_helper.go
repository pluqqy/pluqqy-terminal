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
		lineNumberPattern: regexp.MustCompile(`^\s*\d+[:\|\s]\s+`),
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
	
	// Detect content type
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
	
	// Only normalize indentation if we've done other cleaning
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
		strings.Contains(content, "var ") {
		return ContentTypeCode
	}
	
	return ContentTypePlain
}

// stripTUIBorders removes TUI border characters from lines
func (ph *PasteHelper) stripTUIBorders(content string) string {
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	
	for _, line := range lines {
		// Skip lines that are only borders (including box drawing chars)
		if regexp.MustCompile(`^[│├└┌┐┘┤┬┴┼─╭╮╰╯\s]+$`).MatchString(line) {
			continue
		}
		
		// Handle complex format with multiple │ separators
		// Pattern often seen: "content │ │ linenum content2" 
		// This happens when copying from split panes showing two files side by side
		if strings.Contains(line, "│") {
			// Check for the specific pattern: "content │ │ number content"
			// This indicates two logical lines displayed side by side
			if strings.Contains(line, "│  │") || strings.Contains(line, "│ │") {
				// Split into parts
				parts := strings.Split(line, "│")
				
				// Process each part as potentially separate content
				for i, part := range parts {
					trimmed := strings.TrimSpace(part)
					if trimmed == "" {
						continue
					}
					
					// Skip pure line numbers
					if regexp.MustCompile(`^\d+$`).MatchString(trimmed) {
						continue
					}
					
					// If starts with line number, extract content after it
					if regexp.MustCompile(`^\d+\s+`).MatchString(trimmed) {
						trimmed = regexp.MustCompile(`^\d+\s+`).ReplaceAllString(trimmed, "")
					}
					
					// Add non-empty content as separate lines
					if trimmed != "" && !regexp.MustCompile(`^[│├└┌┐┘┤┬┴┼─\s]+$`).MatchString(trimmed) {
						// For the first part, add it as is
						// For subsequent parts, they represent content from a different source
						if i == 0 || (i > 0 && len(cleaned) == 0) {
							cleaned = append(cleaned, trimmed)
						} else {
							// This is content from the second pane, add as new line
							cleaned = append(cleaned, trimmed)
						}
					}
				}
			} else {
				// Simple border removal for single-pane content
				cleanedLine := ph.tuiBorderPattern.ReplaceAllString(line, "")
				cleanedLine = strings.TrimSpace(cleanedLine)
				if cleanedLine != "" {
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
	if len(lines) < 3 {
		return false
	}
	
	// Check first few lines for consistent line number pattern
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
	return matchCount >= 3
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