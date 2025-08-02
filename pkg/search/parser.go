package search

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TokenType represents the type of a search token
type TokenType int

const (
	TokenTypeField TokenType = iota
	TokenTypeOperator
	TokenTypeValue
	TokenTypeKeyword
	TokenTypeGroupStart
	TokenTypeGroupEnd
)

// Token represents a parsed search token
type Token struct {
	Type  TokenType
	Value string
}

// FieldType represents the type of field being searched
type FieldType string

const (
	FieldTag      FieldType = "tag"
	FieldTypeField FieldType = "type"
	FieldName     FieldType = "name"
	FieldContent  FieldType = "content"
	FieldModified FieldType = "modified"
)

// Operator represents a search operator
type Operator string

const (
	OperatorEquals      Operator = "="
	OperatorNotEquals   Operator = "!="
	OperatorContains    Operator = "contains"
	OperatorGreaterThan Operator = ">"
	OperatorLessThan    Operator = "<"
	OperatorAND         Operator = "AND"
	OperatorOR          Operator = "OR"
	OperatorNOT         Operator = "NOT"
)

// Condition represents a single search condition
type Condition struct {
	Field    FieldType
	Operator Operator
	Value    interface{}
	Negate   bool
}

// Query represents a parsed search query
type Query struct {
	Conditions []Condition
	Logic      []Operator // Logic operators between conditions
	Raw        string     // Original query string
}

// Parser handles parsing of search queries
type Parser struct {
	// Regular expressions for parsing
	fieldPattern    *regexp.Regexp
	quotedPattern   *regexp.Regexp
	modifiedPattern *regexp.Regexp
}

// NewParser creates a new search query parser
func NewParser() *Parser {
	return &Parser{
		fieldPattern:    regexp.MustCompile(`^(\w+):(.+)$`),
		quotedPattern:   regexp.MustCompile(`^"([^"]*)"$`),
		modifiedPattern: regexp.MustCompile(`^([<>])(\d+)([dwmy])$`),
	}
}

// Parse parses a search query string into a Query object
func (p *Parser) Parse(input string) (*Query, error) {
	query := &Query{
		Raw:        input,
		Conditions: []Condition{},
		Logic:      []Operator{},
	}
	
	// Tokenize the input
	tokens := p.tokenize(input)
	
	// Parse tokens into conditions
	if err := p.parseTokens(tokens, query); err != nil {
		return nil, err
	}
	
	return query, nil
}

// tokenize splits the input into tokens
func (p *Parser) tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	
	for i, r := range input {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ' ':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		case '(', ')':
			if !inQuotes {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				tokens = append(tokens, string(r))
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
		
		// Handle end of string
		if i == len(input)-1 && current.Len() > 0 {
			tokens = append(tokens, current.String())
		}
	}
	
	return tokens
}

// parseTokens parses tokens into conditions
func (p *Parser) parseTokens(tokens []string, query *Query) error {
	i := 0
	
	for i < len(tokens) {
		token := tokens[i]
		
		// Check for logical operators
		switch strings.ToUpper(token) {
		case "AND", "OR":
			if len(query.Conditions) == 0 {
				return fmt.Errorf("unexpected operator %s at beginning of query", token)
			}
			query.Logic = append(query.Logic, Operator(strings.ToUpper(token)))
			i++
			continue
		case "NOT":
			// NOT is handled as part of the next condition
			i++
			if i >= len(tokens) {
				return fmt.Errorf("NOT operator requires a condition")
			}
			// Parse the next condition with negate flag
			cond, consumed, err := p.parseCondition(tokens[i:])
			if err != nil {
				return err
			}
			cond.Negate = true
			query.Conditions = append(query.Conditions, *cond)
			i += consumed
			continue
		}
		
		// Parse field:value pairs
		if p.fieldPattern.MatchString(token) {
			cond, consumed, err := p.parseCondition(tokens[i:])
			if err != nil {
				return err
			}
			query.Conditions = append(query.Conditions, *cond)
			i += consumed
		} else {
			// Treat as content search
			query.Conditions = append(query.Conditions, Condition{
				Field:    FieldContent,
				Operator: OperatorContains,
				Value:    p.unquote(token),
			})
			i++
		}
	}
	
	// Validate logic operators
	if len(query.Logic) > 0 && len(query.Logic) != len(query.Conditions)-1 {
		return fmt.Errorf("invalid number of logical operators")
	}
	
	// If no explicit logic operators, default to AND
	if len(query.Conditions) > 1 && len(query.Logic) == 0 {
		for i := 0; i < len(query.Conditions)-1; i++ {
			query.Logic = append(query.Logic, OperatorAND)
		}
	}
	
	return nil
}

// parseCondition parses a single condition from tokens
func (p *Parser) parseCondition(tokens []string) (*Condition, int, error) {
	if len(tokens) == 0 {
		return nil, 0, fmt.Errorf("empty condition")
	}
	
	token := tokens[0]
	matches := p.fieldPattern.FindStringSubmatch(token)
	if len(matches) != 3 {
		return nil, 0, fmt.Errorf("invalid field:value format: %s", token)
	}
	
	field := strings.ToLower(matches[1])
	value := matches[2]
	
	cond := &Condition{}
	
	// Parse field type
	switch field {
	case "tag":
		cond.Field = FieldTag
		cond.Operator = OperatorEquals
		cond.Value = p.unquote(value)
	case "type":
		cond.Field = FieldTypeField
		cond.Operator = OperatorEquals
		cond.Value = p.unquote(value)
	case "name":
		cond.Field = FieldName
		cond.Operator = OperatorContains
		cond.Value = p.unquote(value)
	case "content":
		cond.Field = FieldContent
		cond.Operator = OperatorContains
		cond.Value = p.unquote(value)
	case "modified":
		cond.Field = FieldModified
		duration, op, err := p.parseModifiedValue(value)
		if err != nil {
			return nil, 0, err
		}
		cond.Operator = op
		cond.Value = duration
	default:
		return nil, 0, fmt.Errorf("unknown field: %s", field)
	}
	
	return cond, 1, nil
}

// parseModifiedValue parses modified field values like ">7d" or "<30d"
func (p *Parser) parseModifiedValue(value string) (time.Duration, Operator, error) {
	matches := p.modifiedPattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return 0, "", fmt.Errorf("invalid modified value format: %s (expected format: >7d, <30d, etc.)", value)
	}
	
	op := matches[1]
	num := matches[2]
	unit := matches[3]
	
	// Parse number
	var multiplier int
	fmt.Sscanf(num, "%d", &multiplier)
	
	// Parse unit
	var duration time.Duration
	switch unit {
	case "d":
		duration = time.Duration(multiplier) * 24 * time.Hour
	case "w":
		duration = time.Duration(multiplier) * 7 * 24 * time.Hour
	case "m":
		duration = time.Duration(multiplier) * 30 * 24 * time.Hour
	case "y":
		duration = time.Duration(multiplier) * 365 * 24 * time.Hour
	default:
		return 0, "", fmt.Errorf("invalid time unit: %s", unit)
	}
	
	// Determine operator
	var operator Operator
	switch op {
	case ">":
		operator = OperatorLessThan // Inverted because we compare age
	case "<":
		operator = OperatorGreaterThan // Inverted because we compare age
	default:
		return 0, "", fmt.Errorf("invalid operator: %s", op)
	}
	
	return duration, operator, nil
}

// unquote removes quotes from a string if present
func (p *Parser) unquote(s string) string {
	if matches := p.quotedPattern.FindStringSubmatch(s); len(matches) == 2 {
		return matches[1]
	}
	return s
}

// ParseSimple parses simple tag queries for basic filtering
func ParseSimple(query string) []string {
	// For simple tag filtering, extract tag:value pairs
	var tags []string
	parts := strings.Fields(query)
	
	for _, part := range parts {
		if strings.HasPrefix(part, "tag:") {
			tag := strings.TrimPrefix(part, "tag:")
			tag = strings.Trim(tag, `"`)
			tags = append(tags, tag)
		}
	}
	
	return tags
}