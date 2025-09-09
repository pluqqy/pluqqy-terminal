package examples

import "github.com/pluqqy/pluqqy-terminal/pkg/models"

func getClaudeExamples() []ExampleSet {
	return []ExampleSet{
		{
			Name:        "CLAUDE.md Distiller",
			Description: "Convert existing CLAUDE.md files into Pluqqy components",
			Components: []ExampleComponent{
				// Context
				{
					Name:     "CLAUDE File Parser",
					Filename: "example-claude-parser.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "claude", "migration", "parser"},
					Content: `# CLAUDE.md File Structure

## Common CLAUDE.md Sections

### Project Information
Usually contains:
- Project name and description
- Technology stack
- Repository structure
- Key dependencies

**Maps to**: ` + "`contexts/project-overview.md`" + `

### Commands
Typically includes:
- Build commands
- Test commands
- Development commands
- Deployment commands

**Maps to**: Individual prompt components or a ` + "`contexts/commands.md`" + `

### Architecture
Often describes:
- System architecture
- Module structure
- Design patterns
- Data flow

**Maps to**: ` + "`contexts/architecture.md`" + `

### Development Guidelines
May include:
- Coding standards
- Git workflow
- PR process
- Testing requirements

**Maps to**: ` + "`rules/`" + ` components

### API Documentation
If present:
- Endpoint definitions
- Request/response formats
- Authentication details

**Maps to**: ` + "`contexts/api-docs.md`" + `

## Extraction Strategy

### Phase 1: Identify Sections
1. Look for markdown headers (##, ###)
2. Identify section purposes
3. Categorize content

### Phase 2: Determine Component Types
- **Context**: Factual information, documentation
- **Prompt**: Task instructions, how-tos
- **Rules**: Constraints, standards, requirements

### Phase 3: Extract Content
1. Preserve formatting
2. Replace project-specific with placeholders
3. Group related content

### Phase 4: Create Pipelines
- Group related components
- Order by dependency
- Create logical workflows

## Placeholder Conversion

### Common Replacements
- Project name → {{PROJECT_NAME}}
- Paths → {{PATH}}
- Commands → {{COMMAND}}
- Versions → {{VERSION}}
- URLs → {{URL}}

## Quality Checks
1. Each component should be self-contained
2. Remove redundant information
3. Ensure placeholders are consistent
4. Test component reusability`,
				},
				// Prompt
				{
					Name:     "Extract to Pluqqy",
					Filename: "example-extract-to-pluqqy.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "claude", "migration", "extract"},
					Content: `# Extract CLAUDE.md to Pluqqy Components

## Input
I have a CLAUDE.md file that I want to convert into Pluqqy components and pipelines.

## CLAUDE.md Content
` + "```markdown" + `
{{PASTE_CLAUDE_MD_HERE}}
` + "```" + `

## Extraction Requirements

### 1. Analyze Structure
- Identify all major sections
- Determine the purpose of each section
- Note relationships between sections

### 2. Create Components

#### Contexts (Information)
Extract sections that provide information:
- Project overview
- Architecture documentation
- API specifications
- Database schemas
- Configuration details

#### Prompts (Tasks)
Extract sections that describe tasks:
- How to add features
- How to fix bugs
- How to deploy
- How to test

#### Rules (Constraints)
Extract sections that define rules:
- Coding standards
- Security requirements
- Performance requirements
- Process requirements

### 3. Component Guidelines
- Each component should be 50-200 lines
- Use descriptive names (prefix with project name if specific)
- Add appropriate tags
- Replace specific values with {{PLACEHOLDERS}}

### 4. Create Pipelines
Suggest pipelines that combine components:
- Feature development pipeline
- Bug fixing pipeline
- Code review pipeline
- Deployment pipeline

### 5. Output Format

For each component, provide:
` + "```yaml" + `
name: Component Name
type: contexts|prompts|rules
filename: suggested-filename.md
tags: [tag1, tag2]
content: |
  # Component content here
  With {{PLACEHOLDERS}} for customization
` + "```" + `

For each pipeline, provide:
` + "```yaml" + `
name: Pipeline Name
filename: suggested-pipeline.yaml
description: What this pipeline does
components:
  - type: rules
    path: ../components/rules/component1.md
  - type: contexts
    path: ../components/contexts/component2.md
  - type: prompts
    path: ../components/prompts/component3.md
` + "```" + `

### 6. Reusability Check
Ensure extracted components are:
- [ ] Generic enough for reuse
- [ ] Self-contained
- [ ] Well-documented
- [ ] Properly parameterized

### 7. Migration Guide
Provide instructions for:
1. How to customize placeholders
2. Which pipelines to use for common tasks
3. How to extend with project-specific components`,
				},
				// Rules
				{
					Name:     "Preserve Intent",
					Filename: "example-preserve-intent.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "claude", "migration", "rules"},
					Content: `# CLAUDE.md Migration Rules

## Preservation Principles

### Maintain Original Intent
- **MUST** preserve the original purpose of each section
- **MUST** maintain logical relationships between sections
- **MUST** keep important context together
- **MUST NOT** lose critical information

### Content Transformation

#### Do Transform
- Specific values → Generic placeholders
- Hardcoded paths → {{PATH}} variables
- Project names → {{PROJECT_NAME}}
- Fixed commands → {{COMMAND}} templates

#### Don't Transform
- Technical explanations
- Conceptual descriptions
- Architecture patterns
- Best practices

### Component Sizing
- **Minimum**: 20 lines (avoid tiny fragments)
- **Maximum**: 200 lines (keep manageable)
- **Sweet Spot**: 50-100 lines

### Component Independence
Each component must:
- Stand alone without requiring others
- Provide complete context for its purpose
- Include necessary background information
- Be understandable in isolation

### Naming Conventions
- **Contexts**: ` + "`project-`" + `, ` + "`architecture-`" + `, ` + "`api-`" + `
- **Prompts**: ` + "`implement-`" + `, ` + "`fix-`" + `, ` + "`create-`" + `
- **Rules**: ` + "`standards-`" + `, ` + "`requirements-`" + `, ` + "`guidelines-`" + `

### Placeholder Standards
` + "```" + `
{{SCREAMING_SNAKE_CASE}} for all placeholders
{{PROJECT_NAME}} not {{projectName}}
{{API_KEY}} not {{apiKey}}
` + "```" + `

### Quality Metrics
- ✅ Each component has clear purpose
- ✅ No duplicate information across components
- ✅ Placeholders are consistent
- ✅ Components are reusable
- ✅ Pipelines make logical sense

### Testing Extracted Components
1. Can another project use this component?
2. Are all placeholders documented?
3. Is the component self-explanatory?
4. Does it fit Pluqqy's structure?

## Migration Workflow

### Step 1: Inventory
List all sections in the CLAUDE.md file

### Step 2: Categorize
Assign each section to contexts/prompts/rules

### Step 3: Extract
Pull content into separate components

### Step 4: Parameterize
Replace specific values with placeholders

### Step 5: Pipeline
Create logical groupings of components

### Step 6: Validate
Test the extracted components for completeness

## Remember
The goal is to make CLAUDE.md content:
- Reusable across projects
- Modular and composable
- Easy to customize
- Compatible with Pluqqy's philosophy`,
				},
			},
			Pipelines: []ExamplePipeline{
				{
					Name:        "CLAUDE Distiller",
					Filename:    "example-claude-distiller.yaml",
					Description: "Extract Pluqqy components from existing CLAUDE.md files",
					Tags:        []string{"example", "claude", "migration", "distill"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-preserve-intent.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-claude-parser.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-extract-to-pluqqy.md",
							Order: 3,
						},
					},
				},
			},
		},
	}
}