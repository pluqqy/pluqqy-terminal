package search

import (
	"testing"
	"time"
)

func TestParseSimple(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "single tag",
			query:    "tag:api",
			expected: []string{"api"},
		},
		{
			name:     "multiple tags",
			query:    "tag:api tag:v2",
			expected: []string{"api", "v2"},
		},
		{
			name:     "quoted tag",
			query:    `tag:"api endpoint"`,
			expected: []string{"api endpoint"},
		},
		{
			name:     "mixed content",
			query:    "some text tag:api other text tag:frontend",
			expected: []string{"api", "frontend"},
		},
		{
			name:     "no tags",
			query:    "just some search text",
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSimple(tt.query)
			
			if len(result) != len(tt.expected) {
				t.Errorf("ParseSimple() returned %d tags, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("ParseSimple() tag[%d] = %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple tokens",
			input:    "tag:api type:prompt",
			expected: []string{"tag:api", "type:prompt"},
		},
		{
			name:     "quoted value",
			input:    `tag:"api endpoint" name:test`,
			expected: []string{`tag:"api endpoint"`, "name:test"},
		},
		{
			name:     "logical operators",
			input:    "tag:api AND type:prompt OR tag:v2",
			expected: []string{"tag:api", "AND", "type:prompt", "OR", "tag:v2"},
		},
		{
			name:     "parentheses",
			input:    "(tag:api OR tag:v2) AND type:prompt",
			expected: []string{"(", "tag:api", "OR", "tag:v2", ")", "AND", "type:prompt"},
		},
		{
			name:     "content search",
			input:    `"error handling" tag:api`,
			expected: []string{`"error handling"`, "tag:api"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.tokenize(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize() returned %d tokens, want %d", len(result), len(tt.expected))
				t.Errorf("got: %v", result)
				return
			}
			
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("tokenize() token[%d] = %q, want %q", i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name          string
		input         string
		expectError   bool
		conditionCount int
		checkConditions func(*testing.T, *Query)
	}{
		{
			name:          "single tag condition",
			input:         "tag:api",
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if q.Conditions[0].Field != FieldTag {
					t.Errorf("Expected field type %v, got %v", FieldTag, q.Conditions[0].Field)
				}
				if q.Conditions[0].Value != "api" {
					t.Errorf("Expected value 'api', got %v", q.Conditions[0].Value)
				}
			},
		},
		{
			name:          "multiple conditions with AND",
			input:         "tag:api AND type:prompt",
			conditionCount: 2,
			checkConditions: func(t *testing.T, q *Query) {
				if len(q.Logic) != 1 || q.Logic[0] != OperatorAND {
					t.Errorf("Expected AND operator, got %v", q.Logic)
				}
			},
		},
		{
			name:          "content search",
			input:         `"error handling"`,
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if q.Conditions[0].Field != FieldContent {
					t.Errorf("Expected content field, got %v", q.Conditions[0].Field)
				}
				if q.Conditions[0].Value != "error handling" {
					t.Errorf("Expected value 'error handling', got %v", q.Conditions[0].Value)
				}
			},
		},
		{
			name:          "NOT operator",
			input:         "NOT tag:deprecated",
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if !q.Conditions[0].Negate {
					t.Error("Expected condition to be negated")
				}
			},
		},
		{
			name:          "modified date",
			input:         "modified:>7d",
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if q.Conditions[0].Field != FieldModified {
					t.Errorf("Expected modified field, got %v", q.Conditions[0].Field)
				}
				duration := q.Conditions[0].Value.(time.Duration)
				expectedDuration := 7 * 24 * time.Hour
				if duration != expectedDuration {
					t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
				}
			},
		},
		{
			name:          "invalid field",
			input:         "invalid:value",
			expectError:   true,
		},
		{
			name:          "invalid modified format",
			input:         "modified:yesterday",
			expectError:   true,
		},
		{
			name:          "status archived",
			input:         "status:archived",
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if q.Conditions[0].Field != FieldStatus {
					t.Errorf("Expected status field, got %v", q.Conditions[0].Field)
				}
				if q.Conditions[0].Value != "archived" {
					t.Errorf("Expected value 'archived', got %v", q.Conditions[0].Value)
				}
				if q.Conditions[0].Operator != OperatorEquals {
					t.Errorf("Expected equals operator, got %v", q.Conditions[0].Operator)
				}
			},
		},
		{
			name:          "status with other conditions",
			input:         "status:archived AND tag:api",
			conditionCount: 2,
			checkConditions: func(t *testing.T, q *Query) {
				// Check first condition is status
				if q.Conditions[0].Field != FieldStatus {
					t.Errorf("Expected first condition to be status field, got %v", q.Conditions[0].Field)
				}
				if q.Conditions[0].Value != "archived" {
					t.Errorf("Expected status value 'archived', got %v", q.Conditions[0].Value)
				}
				// Check second condition is tag
				if q.Conditions[1].Field != FieldTag {
					t.Errorf("Expected second condition to be tag field, got %v", q.Conditions[1].Field)
				}
				if q.Conditions[1].Value != "api" {
					t.Errorf("Expected tag value 'api', got %v", q.Conditions[1].Value)
				}
				// Check logic operator
				if len(q.Logic) != 1 || q.Logic[0] != OperatorAND {
					t.Errorf("Expected AND operator, got %v", q.Logic)
				}
			},
		},
		{
			name:          "quoted status value",
			input:         `status:"archived"`,
			conditionCount: 1,
			checkConditions: func(t *testing.T, q *Query) {
				if q.Conditions[0].Field != FieldStatus {
					t.Errorf("Expected status field, got %v", q.Conditions[0].Field)
				}
				if q.Conditions[0].Value != "archived" {
					t.Errorf("Expected value 'archived', got %v", q.Conditions[0].Value)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(query.Conditions) != tt.conditionCount {
				t.Errorf("Expected %d conditions, got %d", tt.conditionCount, len(query.Conditions))
			}
			
			if tt.checkConditions != nil {
				tt.checkConditions(t, query)
			}
		})
	}
}

func TestParseModifiedValue(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name         string
		input        string
		expectError  bool
		expectedDur  time.Duration
		expectedOp   Operator
	}{
		{
			name:        "days",
			input:       ">7d",
			expectedDur: 7 * 24 * time.Hour,
			expectedOp:  OperatorLessThan, // Inverted for age comparison
		},
		{
			name:        "weeks",
			input:       "<2w",
			expectedDur: 2 * 7 * 24 * time.Hour,
			expectedOp:  OperatorGreaterThan, // Inverted for age comparison
		},
		{
			name:        "months",
			input:       ">1m",
			expectedDur: 30 * 24 * time.Hour,
			expectedOp:  OperatorLessThan,
		},
		{
			name:        "years",
			input:       "<1y",
			expectedDur: 365 * 24 * time.Hour,
			expectedOp:  OperatorGreaterThan,
		},
		{
			name:        "invalid format",
			input:       "7 days",
			expectError: true,
		},
		{
			name:        "invalid unit",
			input:       ">7h",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, op, err := parser.parseModifiedValue(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if dur != tt.expectedDur {
				t.Errorf("Expected duration %v, got %v", tt.expectedDur, dur)
			}
			
			if op != tt.expectedOp {
				t.Errorf("Expected operator %v, got %v", tt.expectedOp, op)
			}
		})
	}
}