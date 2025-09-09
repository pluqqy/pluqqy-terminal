package examples

import "github.com/pluqqy/pluqqy-terminal/pkg/models"

func getGeneralExamples() []ExampleSet {
	return []ExampleSet{
		{
			Name:        "General Development",
			Description: "Common development patterns for any project",
			Components: []ExampleComponent{
				// Contexts
				{
					Name:     "Project Overview",
					Filename: "example-project-overview.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "context", "project"},
					Content: `# Project: {{PROJECT_NAME}}

## Purpose
{{PROJECT_DESCRIPTION}}

Replace this with a clear description of what your project does and why it exists.

## Tech Stack
- **Language**: {{PRIMARY_LANGUAGE}}
- **Framework**: {{FRAMEWORK}}
- **Database**: {{DATABASE}}
- **Testing**: {{TEST_FRAMEWORK}}

## Key Features
1. {{FEATURE_1}} - Brief description
2. {{FEATURE_2}} - Brief description
3. {{FEATURE_3}} - Brief description

## Project Structure
` + "```" + `
{{PROJECT_STRUCTURE}}
src/
├── components/    # UI components
├── services/      # Business logic
├── utils/         # Shared utilities
└── tests/         # Test files
` + "```" + `

## Development Workflow
1. {{WORKFLOW_STEP_1}}
2. {{WORKFLOW_STEP_2}}
3. {{WORKFLOW_STEP_3}}

## Important Conventions
- {{CONVENTION_1}}
- {{CONVENTION_2}}
- {{CONVENTION_3}}`,
				},
				{
					Name:     "Code Architecture",
					Filename: "example-code-architecture.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "context", "architecture"},
					Content: `# Code Architecture

## Architecture Pattern
{{ARCHITECTURE_PATTERN}} (e.g., MVC, Clean Architecture, Hexagonal)

## Core Modules

### {{MODULE_1_NAME}}
**Purpose**: {{MODULE_1_PURPOSE}}
**Location**: ` + "`{{MODULE_1_PATH}}`" + `
**Key Components**:
- {{COMPONENT_1}}
- {{COMPONENT_2}}

### {{MODULE_2_NAME}}
**Purpose**: {{MODULE_2_PURPOSE}}
**Location**: ` + "`{{MODULE_2_PATH}}`" + `
**Key Components**:
- {{COMPONENT_1}}
- {{COMPONENT_2}}

## Data Flow
1. {{DATA_FLOW_STEP_1}}
2. {{DATA_FLOW_STEP_2}}
3. {{DATA_FLOW_STEP_3}}

## Key Design Patterns
- **{{PATTERN_1}}**: Used for {{PATTERN_1_PURPOSE}}
- **{{PATTERN_2}}**: Used for {{PATTERN_2_PURPOSE}}

## Dependencies
- External: {{EXTERNAL_DEPS}}
- Internal: {{INTERNAL_DEPS}}

## Error Handling Strategy
{{ERROR_HANDLING_APPROACH}}

## Testing Strategy
{{TESTING_APPROACH}}`,
				},
				{
					Name:     "API Documentation",
					Filename: "example-api-documentation.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "context", "api"},
					Content: `# API Documentation

## Base URL
{{BASE_URL}} (e.g., https://api.example.com/v1)

## Authentication
{{AUTH_METHOD}} (e.g., Bearer Token, API Key, OAuth2)

## Common Headers
` + "```" + `
{{COMMON_HEADERS}}
Content-Type: application/json
Authorization: Bearer {{token}}
` + "```" + `

## Endpoints

### {{ENDPOINT_GROUP_1}}

#### GET {{ENDPOINT_1_PATH}}
**Description**: {{ENDPOINT_1_DESCRIPTION}}
**Parameters**:
- ` + "`{{PARAM_1}}`" + ` ({{PARAM_1_TYPE}}): {{PARAM_1_DESC}}

**Response**:
` + "```json" + `
{
  "{{FIELD_1}}": "{{FIELD_1_TYPE}}",
  "{{FIELD_2}}": "{{FIELD_2_TYPE}}"
}
` + "```" + `

#### POST {{ENDPOINT_2_PATH}}
**Description**: {{ENDPOINT_2_DESCRIPTION}}
**Request Body**:
` + "```json" + `
{
  "{{FIELD_1}}": "{{FIELD_1_VALUE}}",
  "{{FIELD_2}}": "{{FIELD_2_VALUE}}"
}
` + "```" + `

## Error Responses
- ` + "`400`" + `: Bad Request - {{ERROR_400_DESC}}
- ` + "`401`" + `: Unauthorized - {{ERROR_401_DESC}}
- ` + "`404`" + `: Not Found - {{ERROR_404_DESC}}
- ` + "`500`" + `: Internal Server Error - {{ERROR_500_DESC}}`,
				},
				// Prompts
				{
					Name:     "Implement Feature",
					Filename: "example-implement-feature.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "prompt", "feature"},
					Content: `# Task: Implement {{FEATURE_NAME}}

## Overview
{{FEATURE_DESCRIPTION}}

## Requirements
1. {{REQUIREMENT_1}}
2. {{REQUIREMENT_2}}
3. {{REQUIREMENT_3}}

## Technical Specifications
- **Input**: {{INPUT_SPECIFICATION}}
- **Output**: {{OUTPUT_SPECIFICATION}}
- **Performance**: {{PERFORMANCE_REQUIREMENTS}}

## Acceptance Criteria
- [ ] {{CRITERION_1}}
- [ ] {{CRITERION_2}}
- [ ] {{CRITERION_3}}
- [ ] Unit tests written with >{{COVERAGE_PERCENT}}% coverage
- [ ] Integration tests pass
- [ ] Documentation updated

## Implementation Steps
1. {{STEP_1}}
2. {{STEP_2}}
3. {{STEP_3}}
4. Write tests
5. Update documentation

## Edge Cases to Consider
- {{EDGE_CASE_1}}
- {{EDGE_CASE_2}}
- {{EDGE_CASE_3}}

## Definition of Done
- [ ] Code implemented and working
- [ ] All tests passing
- [ ] Code reviewed
- [ ] Documentation updated
- [ ] No linting errors
- [ ] Performance benchmarks met`,
				},
				{
					Name:     "Fix Bug",
					Filename: "example-fix-bug.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "prompt", "bug", "fix"},
					Content: `# Bug Fix: {{BUG_TITLE}}

## Bug Description
{{BUG_DESCRIPTION}}

## Reproduction Steps
1. {{STEP_1}}
2. {{STEP_2}}
3. {{STEP_3}}

## Expected Behavior
{{EXPECTED_BEHAVIOR}}

## Actual Behavior
{{ACTUAL_BEHAVIOR}}

## Environment
- **Version**: {{VERSION}}
- **OS**: {{OPERATING_SYSTEM}}
- **Browser/Runtime**: {{ENVIRONMENT}}

## Error Messages/Logs
` + "```" + `
{{ERROR_LOGS}}
` + "```" + `

## Potential Causes
1. {{CAUSE_1}}
2. {{CAUSE_2}}
3. {{CAUSE_3}}

## Fix Requirements
- [ ] Identify root cause
- [ ] Implement fix
- [ ] Add regression test
- [ ] Verify fix in all affected scenarios
- [ ] Update documentation if needed

## Testing Plan
1. {{TEST_STEP_1}}
2. {{TEST_STEP_2}}
3. {{TEST_STEP_3}}`,
				},
				{
					Name:     "Write Tests",
					Filename: "example-write-tests.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "prompt", "testing"},
					Content: `# Write Tests for {{COMPONENT_NAME}}

## Testing Scope
{{TESTING_SCOPE_DESCRIPTION}}

## Test Categories Required
- [ ] Unit tests
- [ ] Integration tests
- [ ] End-to-end tests (if applicable)
- [ ] Performance tests (if applicable)

## Coverage Requirements
- **Minimum Coverage**: {{MIN_COVERAGE}}%
- **Target Coverage**: {{TARGET_COVERAGE}}%

## Test Cases to Implement

### Happy Path Tests
1. {{HAPPY_PATH_TEST_1}}
2. {{HAPPY_PATH_TEST_2}}
3. {{HAPPY_PATH_TEST_3}}

### Edge Cases
1. {{EDGE_CASE_TEST_1}}
2. {{EDGE_CASE_TEST_2}}
3. {{EDGE_CASE_TEST_3}}

### Error Scenarios
1. {{ERROR_TEST_1}}
2. {{ERROR_TEST_2}}
3. {{ERROR_TEST_3}}

## Test Data Requirements
- {{TEST_DATA_1}}
- {{TEST_DATA_2}}
- {{TEST_DATA_3}}

## Mocking Requirements
- Mock {{MOCK_TARGET_1}}
- Mock {{MOCK_TARGET_2}}
- Stub {{STUB_TARGET}}

## Performance Benchmarks
- {{PERFORMANCE_METRIC_1}}: < {{THRESHOLD_1}}
- {{PERFORMANCE_METRIC_2}}: < {{THRESHOLD_2}}

## Test Organization
- Follow {{TEST_PATTERN}} pattern
- Use {{TEST_FRAMEWORK}} framework conventions
- Group related tests together
- Use descriptive test names`,
				},
				{
					Name:     "Code Review",
					Filename: "example-code-review.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "prompt", "review"},
					Content: `# Code Review Request

## Changes Overview
{{CHANGES_SUMMARY}}

## Files Modified
- {{FILE_1}}: {{FILE_1_CHANGES}}
- {{FILE_2}}: {{FILE_2_CHANGES}}
- {{FILE_3}}: {{FILE_3_CHANGES}}

## Review Focus Areas
Please pay special attention to:
1. {{FOCUS_AREA_1}}
2. {{FOCUS_AREA_2}}
3. {{FOCUS_AREA_3}}

## Review Checklist
- [ ] **Functionality**: Does the code work as intended?
- [ ] **Tests**: Are there adequate tests? Do they pass?
- [ ] **Performance**: Any performance concerns?
- [ ] **Security**: Any security vulnerabilities?
- [ ] **Code Quality**: Is the code clean and maintainable?
- [ ] **Documentation**: Is the code well-documented?
- [ ] **Error Handling**: Are errors handled properly?
- [ ] **Best Practices**: Does it follow project conventions?

## Specific Questions
1. {{QUESTION_1}}
2. {{QUESTION_2}}
3. {{QUESTION_3}}

## Testing Done
- {{TEST_1}}
- {{TEST_2}}
- {{TEST_3}}

## Potential Impacts
- {{IMPACT_1}}
- {{IMPACT_2}}

Please provide:
1. Overall assessment
2. Required changes (blocking)
3. Suggested improvements (non-blocking)
4. Questions or clarifications needed`,
				},
				// Rules
				{
					Name:     "Coding Standards",
					Filename: "example-coding-standards.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "rules", "standards"},
					Content: `# Coding Standards

## Mandatory Rules

### Code Style
- **MUST** follow {{STYLE_GUIDE}} style guide
- **MUST** use {{INDENTATION}} for indentation
- **MUST** limit line length to {{MAX_LINE_LENGTH}} characters
- **MUST** use meaningful variable and function names

### Error Handling
- **MUST** handle all errors explicitly
- **MUST** never ignore error returns
- **MUST** provide context in error messages
- **MUST** log errors appropriately

### Testing
- **MUST** write tests for all new code
- **MUST** maintain minimum {{MIN_COVERAGE}}% test coverage
- **MUST** include both positive and negative test cases
- **MUST** test edge cases

### Documentation
- **MUST** document all public APIs
- **MUST** include examples in documentation
- **MUST** keep documentation up-to-date with code changes
- **MUST** write clear commit messages

### Security
- **MUST** validate all inputs
- **MUST** never hardcode secrets or credentials
- **MUST** use parameterized queries for database operations
- **MUST** follow OWASP guidelines

## Best Practices

### Performance
- **SHOULD** optimize for readability first, performance second
- **SHOULD** profile before optimizing
- **SHOULD** avoid premature optimization
- **SHOULD** use appropriate data structures

### Maintainability
- **SHOULD** follow DRY (Don't Repeat Yourself) principle
- **SHOULD** keep functions small and focused
- **SHOULD** use dependency injection
- **SHOULD** write self-documenting code

### Collaboration
- **SHOULD** review code before merging
- **SHOULD** pair program on complex features
- **SHOULD** communicate design decisions
- **SHOULD** update team on blockers`,
				},
				{
					Name:     "Security First",
					Filename: "example-security-first.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "rules", "security"},
					Content: `# Security-First Development

## Critical Security Rules

### Input Validation
- **ALWAYS** validate and sanitize all user inputs
- **NEVER** trust client-side validation alone
- **ALWAYS** use allowlists, not denylists
- **ALWAYS** validate data types and ranges

### Authentication & Authorization
- **ALWAYS** use secure authentication methods
- **NEVER** store passwords in plain text
- **ALWAYS** implement proper session management
- **ALWAYS** check authorization for every request

### Data Protection
- **ALWAYS** encrypt sensitive data at rest
- **ALWAYS** use TLS for data in transit
- **NEVER** log sensitive information
- **ALWAYS** implement proper key management

### Common Vulnerabilities to Prevent
- **SQL Injection**: Use parameterized queries
- **XSS**: Escape output, use CSP headers
- **CSRF**: Implement CSRF tokens
- **XXE**: Disable XML external entities
- **Path Traversal**: Validate file paths

### Security Testing Requirements
- Run security linters on all code
- Perform dependency vulnerability scans
- Conduct code security reviews
- Test for OWASP Top 10 vulnerabilities

### Incident Response
- Log security-relevant events
- Implement rate limiting
- Have a security incident response plan
- Regular security audits

## Remember
Security is not a feature, it's a requirement. When in doubt, choose the more secure option.`,
				},
				{
					Name:     "Test Driven Development",
					Filename: "example-test-driven.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "rules", "testing", "tdd"},
					Content: `# Test-Driven Development Rules

## TDD Cycle - RED, GREEN, REFACTOR

### 1. RED - Write a Failing Test First
- **MUST** write test before implementation
- **MUST** ensure test fails for the right reason
- **MUST** write minimal test to demonstrate requirement

### 2. GREEN - Make the Test Pass
- **MUST** write minimal code to pass the test
- **MUST** not write more code than needed
- **MUST** verify all tests still pass

### 3. REFACTOR - Improve the Code
- **MUST** refactor while tests are green
- **MUST** improve code structure without changing behavior
- **MUST** ensure all tests still pass after refactoring

## Testing Requirements

### Test Coverage
- **Minimum**: {{MIN_COVERAGE}}% coverage required
- **Target**: {{TARGET_COVERAGE}}% coverage recommended
- **Critical paths**: 100% coverage mandatory

### Test Types Required
1. **Unit Tests**: For individual functions/methods
2. **Integration Tests**: For component interactions
3. **E2E Tests**: For critical user journeys

### Test Quality Standards
- Tests must be independent and isolated
- Tests must be repeatable and deterministic
- Tests must be fast (unit tests < 100ms)
- Tests must have clear, descriptive names
- Tests must follow AAA pattern (Arrange, Act, Assert)

### What to Test
- Happy path scenarios
- Edge cases and boundaries
- Error conditions and exceptions
- Performance requirements
- Security constraints

### Test Documentation
- Each test must clearly state what it's testing
- Complex tests need explanatory comments
- Test data setup must be clear
- Expected outcomes must be obvious

## Benefits We're Seeking
- Confidence in refactoring
- Living documentation
- Better design through testability
- Fewer bugs in production
- Faster development in the long run`,
				},
			},
			Pipelines: []ExamplePipeline{
				{
					Name:        "Feature Development",
					Filename:    "example-feature-development.yaml",
					Description: "Complete pipeline for developing new features",
					Tags:        []string{"example", "feature", "development"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-coding-standards.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-test-driven.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-project-overview.md",
							Order: 3,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-code-architecture.md",
							Order: 4,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-implement-feature.md",
							Order: 5,
						},
					},
				},
				{
					Name:        "Bug Fixing",
					Filename:    "example-bug-fixing.yaml",
					Description: "Pipeline for debugging and fixing issues",
					Tags:        []string{"example", "bug", "fix", "debug"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-security-first.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-code-architecture.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-fix-bug.md",
							Order: 3,
						},
					},
				},
				{
					Name:        "Code Review",
					Filename:    "example-code-review.yaml",
					Description: "Pipeline for thorough code reviews",
					Tags:        []string{"example", "review", "quality"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-coding-standards.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-security-first.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-code-architecture.md",
							Order: 3,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-code-review.md",
							Order: 4,
						},
					},
				},
				{
					Name:        "Test Writing",
					Filename:    "example-test-writing.yaml",
					Description: "Pipeline for writing comprehensive tests",
					Tags:        []string{"example", "testing", "quality"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-test-driven.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-code-architecture.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-write-tests.md",
							Order: 3,
						},
					},
				},
			},
		},
	}
}