package examples

import "github.com/pluqqy/pluqqy-terminal/pkg/models"

func getAIExamples() []ExampleSet {
	return []ExampleSet{
		{
			Name:        "AI Assistant Optimization",
			Description: "Optimize AI coding assistants for better results",
			Components: []ExampleComponent{
				// Contexts
				{
					Name:     "Codebase Overview",
					Filename: "example-codebase-overview.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "ai", "context", "codebase"},
					Content: `# Codebase Overview

## Project Information
- **Name**: {{PROJECT_NAME}}
- **Type**: {{PROJECT_TYPE}} (e.g., Web App, CLI Tool, Library, API)
- **Version**: {{CURRENT_VERSION}}
- **Started**: {{START_DATE}}
- **Team Size**: {{TEAM_SIZE}}

## Technology Stack

### Core Technologies
- **Language**: {{PRIMARY_LANGUAGE}} {{VERSION}}
- **Framework**: {{FRAMEWORK}} {{VERSION}}
- **Runtime**: {{RUNTIME}} {{VERSION}}

### Dependencies
` + "```" + `
{{PACKAGE_MANAGER}} dependencies:
- {{DEP_1}}: {{PURPOSE_1}}
- {{DEP_2}}: {{PURPOSE_2}}
- {{DEP_3}}: {{PURPOSE_3}}
` + "```" + `

## Directory Structure
` + "```" + `
.
├── {{DIR_1}}/          # {{DIR_1_PURPOSE}}
├── {{DIR_2}}/          # {{DIR_2_PURPOSE}}
├── {{DIR_3}}/          # {{DIR_3_PURPOSE}}
├── tests/              # Test files
├── docs/               # Documentation
└── scripts/            # Build and utility scripts
` + "```" + `

## Key Files
- ` + "`{{CONFIG_FILE}}`" + `: Main configuration
- ` + "`{{ENTRY_POINT}}`" + `: Application entry point
- ` + "`{{BUILD_FILE}}`" + `: Build configuration

## Development Workflow

### Setup
` + "```bash" + `
{{SETUP_COMMAND_1}}
{{SETUP_COMMAND_2}}
` + "```" + `

### Common Commands
- **Run**: ` + "`{{RUN_COMMAND}}`" + `
- **Test**: ` + "`{{TEST_COMMAND}}`" + `
- **Build**: ` + "`{{BUILD_COMMAND}}`" + `
- **Lint**: ` + "`{{LINT_COMMAND}}`" + `

## Coding Conventions
1. {{CONVENTION_1}}
2. {{CONVENTION_2}}
3. {{CONVENTION_3}}

## Current Focus Areas
- {{FOCUS_1}}
- {{FOCUS_2}}
- {{FOCUS_3}}

## Known Issues
- {{ISSUE_1}}
- {{ISSUE_2}}

## Important Context
{{ADDITIONAL_CONTEXT}}

This information helps AI assistants understand the project structure and make appropriate suggestions.`,
				},
				{
					Name:     "Tech Stack Details",
					Filename: "example-tech-stack.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "ai", "context", "tech"},
					Content: `# Technology Stack Details

## Primary Language: {{LANGUAGE}}

### Version
{{LANGUAGE_VERSION}}

### Language-Specific Conventions
- Style Guide: {{STYLE_GUIDE}}
- Package Manager: {{PACKAGE_MANAGER}}
- Module System: {{MODULE_SYSTEM}}

## Framework: {{FRAMEWORK}}

### Version
{{FRAMEWORK_VERSION}}

### Framework Patterns
- Architecture: {{ARCHITECTURE_PATTERN}}
- File Structure: {{FILE_CONVENTION}}
- Routing: {{ROUTING_PATTERN}}

## Database

### Type
{{DB_TYPE}} (e.g., PostgreSQL, MongoDB, Redis)

### ORM/ODM
{{ORM_NAME}} {{ORM_VERSION}}

### Connection Management
- Pool Size: {{POOL_SIZE}}
- Connection String: {{CONNECTION_PATTERN}}

## Testing Tools

### Unit Testing
- Framework: {{TEST_FRAMEWORK}}
- Runner: {{TEST_RUNNER}}
- Coverage Tool: {{COVERAGE_TOOL}}

### Integration Testing
- Tool: {{INTEGRATION_TOOL}}
- Strategy: {{TEST_STRATEGY}}

## Build & Deployment

### Build Tool
{{BUILD_TOOL}} {{BUILD_VERSION}}

### CI/CD
- Platform: {{CI_PLATFORM}}
- Pipeline: {{PIPELINE_FILE}}

### Deployment Target
- Environment: {{DEPLOYMENT_ENV}}
- Platform: {{DEPLOYMENT_PLATFORM}}

## Development Tools

### Required Tools
1. {{TOOL_1}} - {{TOOL_1_PURPOSE}}
2. {{TOOL_2}} - {{TOOL_2_PURPOSE}}
3. {{TOOL_3}} - {{TOOL_3_PURPOSE}}

### IDE/Editor Setup
- Recommended: {{RECOMMENDED_IDE}}
- Extensions: {{EXTENSIONS_LIST}}
- Settings: {{SETTINGS_FILE}}

## Third-Party Services
- **Authentication**: {{AUTH_SERVICE}}
- **Monitoring**: {{MONITORING_SERVICE}}
- **Analytics**: {{ANALYTICS_SERVICE}}
- **Email**: {{EMAIL_SERVICE}}

## Performance Requirements
- Response Time: < {{RESPONSE_TIME}}ms
- Memory Usage: < {{MEMORY_LIMIT}}
- Concurrent Users: {{CONCURRENT_USERS}}

## Security Considerations
- Authentication: {{AUTH_METHOD}}
- Authorization: {{AUTH_PATTERN}}
- Encryption: {{ENCRYPTION_STANDARD}}
- Compliance: {{COMPLIANCE_REQUIREMENTS}}`,
				},
				// Prompts
				{
					Name:     "Explain Code",
					Filename: "example-explain-code.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "ai", "explain", "documentation"},
					Content: `# Explain This Code

## Code to Explain
` + "```{{LANGUAGE}}" + `
{{CODE_TO_EXPLAIN}}
` + "```" + `

## Explanation Requirements

### Level of Detail
{{DETAIL_LEVEL}} (e.g., High-level overview, Line-by-line, Deep dive)

### Target Audience
{{AUDIENCE}} (e.g., Junior developer, Senior developer, Non-technical)

### Focus Areas
Please focus on explaining:
1. {{FOCUS_AREA_1}}
2. {{FOCUS_AREA_2}}
3. {{FOCUS_AREA_3}}

## Specific Questions
1. {{QUESTION_1}}
2. {{QUESTION_2}}
3. {{QUESTION_3}}

## Desired Explanation Format

Please provide:
1. **Overview**: What does this code do at a high level?
2. **Key Components**: Break down the main parts
3. **Flow**: How does data/control flow through the code?
4. **Dependencies**: What external dependencies are used?
5. **Patterns**: What design patterns are employed?
6. **Edge Cases**: What edge cases are handled?
7. **Performance**: Any performance considerations?
8. **Improvements**: Potential improvements or alternatives?

## Additional Context
- This code is part of {{SYSTEM_COMPONENT}}
- It interacts with {{INTERACTS_WITH}}
- Common use case: {{USE_CASE}}

## Output Format Preference
{{FORMAT_PREFERENCE}} (e.g., Markdown with code examples, Plain text, Commented code)`,
				},
				{
					Name:     "Refactor Code",
					Filename: "example-refactor-code.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "ai", "refactor", "improve"},
					Content: `# Refactor Code Request

## Current Code
` + "```{{LANGUAGE}}" + `
{{CURRENT_CODE}}
` + "```" + `

## Refactoring Goals
1. {{GOAL_1}} (e.g., Improve readability)
2. {{GOAL_2}} (e.g., Reduce complexity)
3. {{GOAL_3}} (e.g., Better performance)

## Specific Issues to Address
- [ ] {{ISSUE_1}}
- [ ] {{ISSUE_2}}
- [ ] {{ISSUE_3}}

## Constraints
- **Must maintain**: {{MAINTAIN_WHAT}}
- **Cannot change**: {{CANNOT_CHANGE}}
- **Must be compatible with**: {{COMPATIBILITY}}

## Target Metrics
- **Cyclomatic Complexity**: < {{COMPLEXITY_TARGET}}
- **Function Length**: < {{LENGTH_TARGET}} lines
- **Test Coverage**: > {{COVERAGE_TARGET}}%

## Refactoring Techniques to Consider
- [ ] Extract Method/Function
- [ ] Extract Variable
- [ ] Inline Variable
- [ ] Replace Magic Numbers
- [ ] Simplify Conditionals
- [ ] Remove Dead Code
- [ ] Apply Design Pattern: {{PATTERN}}

## Code Style Preferences
- Naming Convention: {{NAMING_CONVENTION}}
- Comment Style: {{COMMENT_STYLE}}
- Error Handling: {{ERROR_STYLE}}

## Testing Requirements
- [ ] All existing tests must still pass
- [ ] Add tests for new helper functions
- [ ] Update tests for changed interfaces
- [ ] Performance benchmarks if applicable

## Expected Deliverables
1. Refactored code with improvements
2. Explanation of changes made
3. Before/after comparison
4. Any breaking changes noted
5. Migration guide if interfaces changed

## Additional Context
{{ADDITIONAL_CONTEXT}}`,
				},
				// Rules
				{
					Name:     "Concise AI Responses",
					Filename: "example-concise-responses.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "ai", "rules", "concise"},
					Content: `# Concise Response Rules for AI Assistants

## Response Length Guidelines

### Be Concise
- **DEFAULT**: Responses under 100 lines unless specifically asked for more
- **PREFER**: Bullet points over paragraphs
- **AVOID**: Unnecessary preambles and conclusions
- **FOCUS**: Answer the specific question asked

### Code-First Approach
- **START** with code when asked for implementation
- **FOLLOW** with brief explanation only if needed
- **SKIP** obvious comments in code
- **INCLUDE** only essential context

## Format Preferences

### For Code Responses
1. Show the code first
2. Add brief explanation after (if needed)
3. Mention key changes or decisions
4. Skip step-by-step unless requested

### For Explanations
1. Lead with the answer
2. Use bullet points for details
3. Include examples only when clarifying
4. Link to docs instead of explaining basics

## What to Avoid

### Don't Include
- "Let me help you with..."
- "Here's how you can..."
- "In conclusion..."
- Repeating the question
- Explaining what you're about to do
- Summarizing what you just did

### Don't Over-Explain
- Basic programming concepts
- Well-known library features
- Standard patterns
- Common conventions

## When to Be Detailed

### Provide More Detail For
- Complex algorithms
- Security concerns
- Performance implications
- Breaking changes
- Non-obvious trade-offs

### Always Include
- Error handling when critical
- Edge cases when relevant
- Security warnings
- Performance impacts
- Breaking change notices

## Examples

### Good Response
` + "```python" + `
def calculate_tax(amount, rate=0.08):
    return amount * (1 + rate)
` + "```" + `
Adds tax to amount. Default 8% rate.

### Bad Response
"I'll help you create a function to calculate tax. Here's a step-by-step approach to solving this problem. First, we need to understand what tax calculation means..."

## Remember
- Respect the user's time
- Assume technical competence
- Code speaks louder than words
- Quality over quantity`,
				},
				{
					Name:     "Code First Approach",
					Filename: "example-code-first.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "ai", "rules", "code"},
					Content: `# Code-First Development Rules

## Primary Rule
**ALWAYS** show working code before explaining it.

## Response Structure

### For Implementation Requests
1. **Code Block** - Complete, runnable code
2. **Key Points** - 2-3 bullet points max
3. **Usage Example** - If not obvious
4. **Notes** - Only if critical

### For Bug Fixes
1. **Fixed Code** - Show the correction
2. **What Changed** - One-line summary
3. **Why** - Brief explanation
4. **Test** - Quick verification method

### For Reviews
1. **Issues Found** - Bulleted list
2. **Fixed Version** - Corrected code
3. **Explanation** - Only for non-obvious fixes

## Code Quality Standards

### Every Code Block Must
- ✅ Be syntactically correct
- ✅ Be runnable as-is
- ✅ Include imports/dependencies
- ✅ Handle errors appropriately
- ✅ Follow language conventions

### Avoid
- ❌ Pseudo-code (unless requested)
- ❌ Incomplete snippets
- ❌ "..." or "// more code here"
- ❌ Untested code
- ❌ Over-commented obvious code

## Comments in Code

### Include Comments For
- Complex algorithms
- Non-obvious decisions
- Magic numbers/strings
- External API calls
- Workarounds or hacks

### Skip Comments For
- Variable declarations
- Simple loops
- Standard patterns
- Getter/setters
- Obvious operations

## Examples

### Good: Code First
` + "```javascript" + `
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// Usage
const searchAPI = debounce(search, 300);
` + "```" + `
Delays function execution until after wait milliseconds.

### Bad: Explanation First
"To implement debounce, we need to understand that it limits the rate at which a function can fire. Here's how it works conceptually..."

## Testing Code

### Always Include
- Basic test case or usage example
- Expected output for verification
- Error handling demonstration

### Format
` + "```language" + `
// Code here
` + "```" + `
` + "```" + `
// Test/Usage
input: [1, 2, 3]
output: 6
` + "```" + `

## Remember
- Code is documentation
- Examples > Explanations  
- Show, don't tell
- Test before presenting
- Keep it runnable`,
				},
			},
			Pipelines: []ExamplePipeline{
				{
					Name:        "AI Assistant Setup",
					Filename:    "example-ai-assistant-setup.yaml",
					Description: "Optimize AI assistants for your codebase",
					Tags:        []string{"example", "ai", "setup"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-concise-responses.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-code-first.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-codebase-overview.md",
							Order: 3,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-tech-stack.md",
							Order: 4,
						},
					},
				},
				{
					Name:        "Code Explanation",
					Filename:    "example-code-explanation.yaml",
					Description: "Get clear explanations of complex code",
					Tags:        []string{"example", "ai", "explain"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-concise-responses.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-code-architecture.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-explain-code.md",
							Order: 3,
						},
					},
				},
			},
		},
	}
}