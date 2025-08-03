package tui

import (
	"fmt"
	"strings"
)

// pluralize returns "s" for counts other than 1, empty string for 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// truncateName truncates a name to fit within the specified width
// It adds "..." if truncation occurs
func truncateName(name string, maxWidth int) string {
	if len(name) > maxWidth-3 {
		return name[:maxWidth-6] + "..."
	}
	return name
}

// formatTableRow formats component/pipeline data into aligned columns
func formatTableRow(name, tags, tokens, usage string, nameWidth, tagsWidth, tokenWidth, usageWidth int) string {
	namePart := fmt.Sprintf("%-*s", nameWidth, name)
	
	// For tags, we need to pad based on rendered width
	tagsPadding := tagsWidth - len(tags)
	if tagsPadding < 0 {
		tagsPadding = 0
	}
	tagsPart := tags + strings.Repeat(" ", tagsPadding)
	
	tokenPart := fmt.Sprintf("%*s", tokenWidth, tokens)
	
	if usage != "" {
		usagePart := fmt.Sprintf("%*s", usageWidth, usage)
		return fmt.Sprintf("%s %s  %s %s", namePart, tagsPart, tokenPart, usagePart)
	}
	
	return fmt.Sprintf("%s %s %s", namePart, tagsPart, tokenPart)
}

// preprocessContent handles carriage returns and ensures proper line breaks
func preprocessContent(content string) string {
	processed := strings.ReplaceAll(content, "\r\r", "\n\n")
	return strings.ReplaceAll(processed, "\r", "\n")
}

// formatColumnWidths calculates column widths based on available space
func formatColumnWidths(totalWidth int, hasUsageColumn bool) (nameWidth, tagsWidth, tokenWidth, usageWidth int) {
	// Base widths
	nameWidth = 30
	tagsWidth = 20
	tokenWidth = 8
	usageWidth = 5
	
	if !hasUsageColumn {
		// For pipelines (no usage column)
		availableWidth := totalWidth - tokenWidth - 4 // 4 for spacing
		nameWidth = availableWidth * 60 / 100
		tagsWidth = availableWidth * 40 / 100
		usageWidth = 0
	} else {
		// For components (with usage column)
		availableWidth := totalWidth - tokenWidth - usageWidth - 6 // 6 for spacing
		nameWidth = availableWidth * 55 / 100
		tagsWidth = availableWidth * 45 / 100
	}
	
	// Ensure minimum widths
	if nameWidth < 20 {
		nameWidth = 20
	}
	if tagsWidth < 15 {
		tagsWidth = 15
	}
	
	return
}