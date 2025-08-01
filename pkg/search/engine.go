package search

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ItemType represents the type of searchable item
type ItemType string

const (
	ItemTypeComponent ItemType = "component"
	ItemTypePipeline  ItemType = "pipeline"
)

// SearchItem represents a searchable item
type SearchItem struct {
	Type       ItemType
	Path       string
	Name       string
	Tags       []string
	Content    string
	Modified   time.Time
	TokenCount int
}

// SearchResult represents a search result with relevance score
type SearchResult struct {
	Item      SearchItem
	Score     float64
	Highlights map[string][]string // Field -> highlighted excerpts
}

// Index represents the search index
type Index struct {
	mu    sync.RWMutex
	items []SearchItem
	
	// Inverted indexes for fast lookup
	tagIndex      map[string][]int // tag -> item indices
	typeIndex     map[string][]int // type -> item indices
	contentTokens map[string][]int // token -> item indices
}

// Engine represents the search engine
type Engine struct {
	index  *Index
	parser *Parser
}

// NewEngine creates a new search engine
func NewEngine() *Engine {
	return &Engine{
		index: &Index{
			items:         []SearchItem{},
			tagIndex:      make(map[string][]int),
			typeIndex:     make(map[string][]int),
			contentTokens: make(map[string][]int),
		},
		parser: NewParser(),
	}
}

// BuildIndex builds the search index from all components and pipelines
func (e *Engine) BuildIndex() error {
	e.index.mu.Lock()
	defer e.index.mu.Unlock()
	
	// Clear existing index
	e.index.items = []SearchItem{}
	e.index.tagIndex = make(map[string][]int)
	e.index.typeIndex = make(map[string][]int)
	e.index.contentTokens = make(map[string][]int)
	
	// Index components
	for _, compType := range []string{models.ComponentTypePrompt, models.ComponentTypeContext, models.ComponentTypeRules} {
		components, err := files.ListComponents(compType)
		if err != nil {
			return fmt.Errorf("failed to list %s components: %w", compType, err)
		}
		
		for _, compFile := range components {
			compPath := filepath.Join(files.ComponentsDir, compType, compFile)
			comp, err := files.ReadComponent(compPath)
			if err != nil {
				continue // Skip components that can't be read
			}
			
			item := SearchItem{
				Type:     ItemTypeComponent,
				Path:     compPath,
				Name:     strings.TrimSuffix(compFile, ".md"),
				Tags:     comp.Tags,
				Content:  comp.Content,
				Modified: comp.Modified,
			}
			
			e.addToIndex(item)
		}
	}
	
	// Index pipelines
	pipelines, err := files.ListPipelines()
	if err != nil {
		return fmt.Errorf("failed to list pipelines: %w", err)
	}
	
	for _, pipelineFile := range pipelines {
		pipeline, err := files.ReadPipeline(pipelineFile)
		if err != nil {
			continue // Skip pipelines that can't be read
		}
		
		item := SearchItem{
			Type:     ItemTypePipeline,
			Path:     pipelineFile,
			Name:     pipeline.Name,
			Tags:     pipeline.Tags,
			Content:  "", // Pipelines don't have content
			Modified: time.Now(), // TODO: Get actual modified time
		}
		
		e.addToIndex(item)
	}
	
	return nil
}

// addToIndex adds an item to the index and updates inverted indexes
func (e *Engine) addToIndex(item SearchItem) {
	idx := len(e.index.items)
	e.index.items = append(e.index.items, item)
	
	// Update tag index
	for _, tag := range item.Tags {
		normalized := models.NormalizeTagName(tag)
		e.index.tagIndex[normalized] = append(e.index.tagIndex[normalized], idx)
	}
	
	// Update type index
	e.index.typeIndex[string(item.Type)] = append(e.index.typeIndex[string(item.Type)], idx)
	
	// Update content token index (simple word tokenization)
	if item.Content != "" {
		tokens := tokenizeContent(strings.ToLower(item.Content))
		for _, token := range tokens {
			e.index.contentTokens[token] = append(e.index.contentTokens[token], idx)
		}
	}
}

// Search performs a search using the given query
func (e *Engine) Search(queryStr string) ([]SearchResult, error) {
	query, err := e.parser.Parse(queryStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	
	e.index.mu.RLock()
	defer e.index.mu.RUnlock()
	
	// Get matching items for each condition
	var conditionMatches [][]int
	for _, condition := range query.Conditions {
		matches := e.evaluateCondition(condition)
		conditionMatches = append(conditionMatches, matches)
	}
	
	// Combine results based on logical operators
	finalMatches := e.combineMatches(conditionMatches, query.Logic)
	
	// Convert to search results with scoring
	results := make([]SearchResult, 0, len(finalMatches))
	for _, idx := range finalMatches {
		if idx < len(e.index.items) {
			result := SearchResult{
				Item:       e.index.items[idx],
				Score:      e.calculateScore(e.index.items[idx], query),
				Highlights: e.generateHighlights(e.index.items[idx], query),
			}
			results = append(results, result)
		}
	}
	
	// Sort by score
	sortResultsByScore(results)
	
	return results, nil
}

// evaluateCondition evaluates a single search condition
func (e *Engine) evaluateCondition(condition Condition) []int {
	var matches []int
	
	switch condition.Field {
	case FieldTag:
		tag := models.NormalizeTagName(condition.Value.(string))
		if indices, exists := e.index.tagIndex[tag]; exists {
			matches = indices
		}
		
	case FieldType:
		typeStr := condition.Value.(string)
		// Handle component type aliases
		switch typeStr {
		case "prompt":
			typeStr = "component"
		case "context":
			typeStr = "component"
		case "rules":
			typeStr = "component"
		}
		if indices, exists := e.index.typeIndex[typeStr]; exists {
			matches = indices
		}
		
	case FieldName:
		pattern := strings.ToLower(condition.Value.(string))
		for i, item := range e.index.items {
			if strings.Contains(strings.ToLower(item.Name), pattern) {
				matches = append(matches, i)
			}
		}
		
	case FieldContent:
		searchTerm := strings.ToLower(condition.Value.(string))
		// Use token index for word-based search
		if indices, exists := e.index.contentTokens[searchTerm]; exists {
			matches = indices
		} else {
			// Fallback to substring search
			for i, item := range e.index.items {
				if strings.Contains(strings.ToLower(item.Content), searchTerm) {
					matches = append(matches, i)
				}
			}
		}
		
	case FieldModified:
		duration := condition.Value.(time.Duration)
		cutoff := time.Now().Add(-duration)
		for i, item := range e.index.items {
			if condition.Operator == OperatorGreaterThan && item.Modified.After(cutoff) {
				matches = append(matches, i)
			} else if condition.Operator == OperatorLessThan && item.Modified.Before(cutoff) {
				matches = append(matches, i)
			}
		}
	}
	
	// Handle negation
	if condition.Negate {
		matches = e.invertMatches(matches)
	}
	
	return matches
}

// combineMatches combines match sets based on logical operators
func (e *Engine) combineMatches(conditionMatches [][]int, operators []Operator) []int {
	if len(conditionMatches) == 0 {
		return []int{}
	}
	
	result := conditionMatches[0]
	
	for i := 1; i < len(conditionMatches); i++ {
		if i-1 < len(operators) {
			switch operators[i-1] {
			case OperatorAND:
				result = intersectSlices(result, conditionMatches[i])
			case OperatorOR:
				result = unionSlices(result, conditionMatches[i])
			}
		}
	}
	
	return result
}

// invertMatches returns all indices not in the given matches
func (e *Engine) invertMatches(matches []int) []int {
	matchSet := make(map[int]bool)
	for _, m := range matches {
		matchSet[m] = true
	}
	
	var inverted []int
	for i := range e.index.items {
		if !matchSet[i] {
			inverted = append(inverted, i)
		}
	}
	
	return inverted
}

// calculateScore calculates relevance score for a search result
func (e *Engine) calculateScore(item SearchItem, query *Query) float64 {
	score := 1.0
	
	// Boost for exact name matches
	for _, condition := range query.Conditions {
		if condition.Field == FieldName {
			pattern := strings.ToLower(condition.Value.(string))
			if strings.ToLower(item.Name) == pattern {
				score += 2.0
			} else if strings.HasPrefix(strings.ToLower(item.Name), pattern) {
				score += 1.0
			}
		}
	}
	
	// Boost for tag matches
	tagMatches := 0
	for _, condition := range query.Conditions {
		if condition.Field == FieldTag {
			searchTag := models.NormalizeTagName(condition.Value.(string))
			for _, itemTag := range item.Tags {
				if models.NormalizeTagName(itemTag) == searchTag {
					tagMatches++
				}
			}
		}
	}
	score += float64(tagMatches) * 0.5
	
	// Boost for recent items
	age := time.Since(item.Modified)
	if age < 24*time.Hour {
		score += 1.0
	} else if age < 7*24*time.Hour {
		score += 0.5
	}
	
	return score
}

// generateHighlights generates highlighted excerpts for matching fields
func (e *Engine) generateHighlights(item SearchItem, query *Query) map[string][]string {
	highlights := make(map[string][]string)
	
	for _, condition := range query.Conditions {
		switch condition.Field {
		case FieldContent:
			searchTerm := condition.Value.(string)
			excerpts := extractExcerpts(item.Content, searchTerm, 3, 50)
			if len(excerpts) > 0 {
				highlights["content"] = excerpts
			}
		case FieldName:
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(condition.Value.(string))) {
				highlights["name"] = []string{item.Name}
			}
		}
	}
	
	return highlights
}

// Helper functions

func tokenizeContent(content string) []string {
	// Simple word tokenization
	var tokens []string
	words := strings.Fields(content)
	
	for _, word := range words {
		// Remove common punctuation
		word = strings.Trim(word, ".,!?;:\"'")
		if len(word) > 2 { // Skip very short words
			tokens = append(tokens, strings.ToLower(word))
		}
	}
	
	return tokens
}

func intersectSlices(a, b []int) []int {
	set := make(map[int]bool)
	for _, v := range a {
		set[v] = true
	}
	
	var result []int
	for _, v := range b {
		if set[v] {
			result = append(result, v)
		}
	}
	
	return result
}

func unionSlices(a, b []int) []int {
	set := make(map[int]bool)
	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		set[v] = true
	}
	
	result := make([]int, 0, len(set))
	for v := range set {
		result = append(result, v)
	}
	
	return result
}

func sortResultsByScore(results []SearchResult) {
	// Simple bubble sort for now
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func extractExcerpts(content, searchTerm string, maxExcerpts, contextChars int) []string {
	var excerpts []string
	lowerContent := strings.ToLower(content)
	lowerTerm := strings.ToLower(searchTerm)
	
	index := 0
	for i := 0; i < maxExcerpts; i++ {
		pos := strings.Index(lowerContent[index:], lowerTerm)
		if pos == -1 {
			break
		}
		
		pos += index
		start := pos - contextChars
		if start < 0 {
			start = 0
		}
		
		end := pos + len(searchTerm) + contextChars
		if end > len(content) {
			end = len(content)
		}
		
		excerpt := content[start:end]
		if start > 0 {
			excerpt = "..." + excerpt
		}
		if end < len(content) {
			excerpt = excerpt + "..."
		}
		
		excerpts = append(excerpts, excerpt)
		index = pos + len(searchTerm)
	}
	
	return excerpts
}