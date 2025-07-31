package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// EstimateTokens provides a lightweight estimation of token count
// Based on OpenAI's general guidelines:
// - 1 token ~= 4 characters in English
// - 1 token ~= ¾ words
// This implementation uses a more sophisticated approach considering:
// - Whitespace and punctuation
// - Code blocks may have different density
// - Common patterns in technical documentation
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Remove extra whitespace
	text = strings.TrimSpace(text)
	
	// Basic character count divided by 4
	charCount := len(text)
	baseEstimate := charCount / 4
	
	// Count words for cross-reference
	words := regexp.MustCompile(`\S+`).FindAllString(text, -1)
	wordEstimate := int(float64(len(words)) * 1.3)
	
	// Use the average of both methods for better accuracy
	estimate := (baseEstimate + wordEstimate) / 2
	
	// Adjust for code blocks (tend to have more tokens per character)
	codeBlocks := regexp.MustCompile("```[\\s\\S]*?```").FindAllString(text, -1)
	for _, block := range codeBlocks {
		// Code typically has ~3 chars per token instead of 4
		codeChars := len(block)
		codeAdjustment := (codeChars / 3) - (codeChars / 4)
		estimate += codeAdjustment
	}
	
	// Ensure we return at least 1 token for non-empty text
	if estimate < 1 {
		estimate = 1
	}
	
	return estimate
}

// FormatTokenCount formats the token count for display
func FormatTokenCount(tokens int) string {
	if tokens < 1000 {
		return fmt.Sprintf("~%d tokens", tokens)
	} else if tokens < 10000 {
		return fmt.Sprintf("~%.1fK tokens", float64(tokens)/1000)
	} else {
		return fmt.Sprintf("~%.0fK tokens", float64(tokens)/1000)
	}
}

// GetTokenLimitStatus returns the status based on common LLM limits
func GetTokenLimitStatus(tokens int) (percentage int, limit int, status string) {
	// Common limits: 4K, 8K, 16K, 32K, 128K
	limits := []int{4096, 8192, 16384, 32768, 131072}
	
	// Find the smallest limit that can accommodate the tokens
	selectedLimit := limits[len(limits)-1] // Default to largest
	for _, l := range limits {
		if tokens <= l {
			selectedLimit = l
			break
		}
	}
	
	percentage = (tokens * 100) / selectedLimit
	
	// Determine status based on percentage of selected limit
	if percentage < 50 {
		status = "good"
	} else if percentage < 80 {
		status = "warning"
	} else {
		status = "danger"
	}
	
	return percentage, selectedLimit, status
}