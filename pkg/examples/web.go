package examples

import "github.com/pluqqy/pluqqy-terminal/pkg/models"

func getWebExamples() []ExampleSet {
	return []ExampleSet{
		{
			Name:        "Web Development",
			Description: "Examples for web development with React, APIs, and databases",
			Components: []ExampleComponent{
				// Contexts
				{
					Name:     "React Architecture",
					Filename: "example-react-architecture.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "web", "react", "frontend"},
					Content: `# React Application Architecture

## Project Structure
` + "```" + `
src/
├── components/       # Reusable UI components
│   ├── common/      # Shared components (Button, Modal, etc.)
│   ├── layout/      # Layout components (Header, Footer, etc.)
│   └── features/    # Feature-specific components
├── pages/           # Page components (route handlers)
├── hooks/           # Custom React hooks
├── services/        # API services and external integrations
├── store/           # State management (Redux/Context)
├── utils/           # Utility functions
├── styles/          # Global styles and themes
└── types/           # TypeScript type definitions
` + "```" + `

## State Management
**Solution**: {{STATE_MANAGEMENT}} (e.g., Redux Toolkit, Zustand, Context API)

### Global State Structure
` + "```typescript" + `
{
  user: UserState,
  ui: UIState,
  {{FEATURE_1}}: {{Feature1State}},
  {{FEATURE_2}}: {{Feature2State}}
}
` + "```" + `

## Component Patterns

### Container/Presentational Pattern
- **Containers**: Handle logic and state
- **Presentational**: Pure UI components

### Custom Hooks Pattern
- ` + "`use{{Feature}}`" + `: Feature-specific logic
- ` + "`useAPI`" + `: API call handling
- ` + "`useAuth`" + `: Authentication logic

## Routing
**Router**: {{ROUTER}} (e.g., React Router v6)

### Route Structure
- ` + "`/`" + ` - Home page
- ` + "`/{{feature1}}`" + ` - Feature 1
- ` + "`/{{feature2}}/:id`" + ` - Feature 2 detail
- ` + "`/admin/*`" + ` - Admin routes (protected)

## Performance Optimization
- Code splitting with ` + "`React.lazy()`" + `
- Memoization with ` + "`useMemo`" + ` and ` + "`React.memo`" + `
- Virtual scrolling for large lists
- Image lazy loading

## Testing Strategy
- Unit tests with {{TEST_LIBRARY}} (e.g., Jest, Vitest)
- Component tests with {{COMPONENT_TEST}} (e.g., React Testing Library)
- E2E tests with {{E2E_FRAMEWORK}} (e.g., Cypress, Playwright)`,
				},
				{
					Name:     "REST API Design",
					Filename: "example-rest-api-design.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "web", "api", "backend"},
					Content: `# REST API Design

## API Standards
- **Base URL**: {{API_BASE_URL}}
- **Version**: {{API_VERSION}} (e.g., /api/v1)
- **Format**: JSON
- **Authentication**: {{AUTH_METHOD}}

## Resource Naming Conventions
- Use nouns, not verbs
- Use plural forms
- Use kebab-case for multi-word resources
- Nest resources logically

## Standard Endpoints

### {{RESOURCE_1}} Resource
` + "```" + `
GET    /api/v1/{{resources}}          # List all
GET    /api/v1/{{resources}}/:id      # Get single
POST   /api/v1/{{resources}}          # Create new
PUT    /api/v1/{{resources}}/:id      # Update (full)
PATCH  /api/v1/{{resources}}/:id      # Update (partial)
DELETE /api/v1/{{resources}}/:id      # Delete
` + "```" + `

## Request/Response Format

### Standard Response Structure
` + "```json" + `
{
  "success": true,
  "data": {
    // Resource data here
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0"
  }
}
` + "```" + `

### Error Response Structure
` + "```json" + `
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "issue": "Invalid format"
    }
  }
}
` + "```" + `

## Pagination
` + "```" + `
GET /api/v1/{{resources}}?page=2&limit=20&sort=-created_at
` + "```" + `

Response includes:
` + "```json" + `
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
` + "```" + `

## Filtering & Searching
- Filter: ` + "`?status=active&category=tech`" + `
- Search: ` + "`?q=search+term`" + `
- Date range: ` + "`?from=2024-01-01&to=2024-12-31`" + `

## Status Codes
- ` + "`200`" + `: OK - Successful GET, PUT
- ` + "`201`" + `: Created - Successful POST
- ` + "`204`" + `: No Content - Successful DELETE
- ` + "`400`" + `: Bad Request - Invalid input
- ` + "`401`" + `: Unauthorized - Authentication required
- ` + "`403`" + `: Forbidden - No permission
- ` + "`404`" + `: Not Found - Resource doesn't exist
- ` + "`422`" + `: Unprocessable Entity - Validation errors
- ` + "`500`" + `: Internal Server Error`,
				},
				{
					Name:     "Database Schema",
					Filename: "example-database-schema.md",
					Type:     models.ComponentTypeContext,
					Tags:     []string{"example", "web", "database", "backend"},
					Content: `# Database Schema

## Database Type
{{DATABASE_TYPE}} (e.g., PostgreSQL, MySQL, MongoDB)

## Core Tables/Collections

### {{TABLE_1}} Table
` + "```sql" + `
CREATE TABLE {{table_name}} (
  id            {{ID_TYPE}} PRIMARY KEY,
  {{field_1}}   {{TYPE_1}} NOT NULL,
  {{field_2}}   {{TYPE_2}},
  {{field_3}}   {{TYPE_3}} DEFAULT {{DEFAULT_VALUE}},
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  CONSTRAINT {{constraint_name}} UNIQUE ({{field_1}}),
  INDEX idx_{{field_2}} ({{field_2}})
);
` + "```" + `

### {{TABLE_2}} Table
` + "```sql" + `
CREATE TABLE {{table_name_2}} (
  id            {{ID_TYPE}} PRIMARY KEY,
  {{foreign_key}} {{FK_TYPE}} NOT NULL,
  {{field_1}}   {{TYPE_1}},
  
  FOREIGN KEY ({{foreign_key}}) REFERENCES {{table_1}}(id) ON DELETE CASCADE
);
` + "```" + `

## Relationships
- **{{TABLE_1}} → {{TABLE_2}}**: One-to-Many
- **{{TABLE_2}} ↔ {{TABLE_3}}**: Many-to-Many (via {{JOIN_TABLE}})

## Indexes
- Primary keys: Automatically indexed
- Foreign keys: ` + "`idx_{{table}}_{{foreign_key}}`" + `
- Search fields: ` + "`idx_{{table}}_{{search_field}}`" + `
- Composite: ` + "`idx_{{table}}_{{field1}}_{{field2}}`" + `

## Migrations Strategy
1. Never modify existing migrations
2. Always create new migration files
3. Test rollback before deploying
4. Version control all migrations

## Data Validation Rules
- **Email**: Must be unique and valid format
- **Passwords**: Minimum {{MIN_PASSWORD}} characters
- **Usernames**: Alphanumeric, 3-20 characters
- **Dates**: ISO 8601 format

## Performance Considerations
- Index frequently queried fields
- Denormalize for read-heavy operations
- Use connection pooling
- Implement query result caching
- Regular VACUUM/ANALYZE (PostgreSQL)`,
				},
				// Prompts
				{
					Name:     "Create React Component",
					Filename: "example-create-react-component.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "web", "react", "component"},
					Content: `# Create React Component: {{COMPONENT_NAME}}

## Component Type
{{COMPONENT_TYPE}} (e.g., Functional, Class, HOC, Custom Hook)

## Purpose
{{COMPONENT_PURPOSE}}

## Location
` + "`src/components/{{PATH}}/{{ComponentName}}.tsx`" + `

## Props Interface
` + "```typescript" + `
interface {{ComponentName}}Props {
  {{prop1}}: {{type1}};
  {{prop2}}?: {{type2}};  // optional
  {{prop3}}: {{type3}};
  on{{Event}}: ({{params}}) => void;
}
` + "```" + `

## State Requirements
- {{STATE_ITEM_1}}: {{STATE_TYPE_1}}
- {{STATE_ITEM_2}}: {{STATE_TYPE_2}}

## Component Features
- [ ] {{FEATURE_1}}
- [ ] {{FEATURE_2}}
- [ ] {{FEATURE_3}}

## Styling Approach
{{STYLING_METHOD}} (e.g., CSS Modules, Styled Components, Tailwind)

## Accessibility Requirements
- [ ] Keyboard navigation support
- [ ] ARIA labels where needed
- [ ] Semantic HTML elements
- [ ] Focus management
- [ ] Screen reader compatibility

## Performance Considerations
- [ ] Memoize expensive computations
- [ ] Use React.memo if pure component
- [ ] Lazy load heavy dependencies
- [ ] Optimize re-renders

## Testing Requirements
- [ ] Unit tests for logic
- [ ] Component rendering tests
- [ ] User interaction tests
- [ ] Accessibility tests
- [ ] Snapshot tests (if applicable)

## Example Usage
` + "```tsx" + `
<{{ComponentName}}
  {{prop1}}={value1}
  {{prop2}}={value2}
  on{{Event}}={handleEvent}
/>
` + "```" + `

## Dependencies
- {{DEPENDENCY_1}}
- {{DEPENDENCY_2}}

Please create:
1. Component file with TypeScript
2. Test file
3. Styles file (if needed)
4. Storybook story (if applicable)`,
				},
				{
					Name:     "Add API Endpoint",
					Filename: "example-add-api-endpoint.md",
					Type:     models.ComponentTypePrompt,
					Tags:     []string{"example", "web", "api", "backend"},
					Content: `# Add API Endpoint: {{ENDPOINT_NAME}}

## Endpoint Details
- **Method**: {{HTTP_METHOD}}
- **Path**: ` + "`/api/v1/{{resource_path}}`" + `
- **Purpose**: {{ENDPOINT_PURPOSE}}

## Request Specification

### Headers
` + "```" + `
Content-Type: application/json
Authorization: Bearer {{token}}
{{CUSTOM_HEADER}}: {{HEADER_VALUE}}
` + "```" + `

### Path Parameters
- ` + "`{{param1}}`" + `: {{PARAM1_DESCRIPTION}}

### Query Parameters
- ` + "`{{query1}}`" + `: {{QUERY1_DESC}} (optional)
- ` + "`{{query2}}`" + `: {{QUERY2_DESC}} (required)

### Request Body
` + "```json" + `
{
  "{{field1}}": "{{type1}}",
  "{{field2}}": "{{type2}}",
  "{{field3}}": {
    "{{nested1}}": "{{type3}}"
  }
}
` + "```" + `

## Response Specification

### Success Response ({{SUCCESS_CODE}})
` + "```json" + `
{
  "success": true,
  "data": {
    "{{response_field1}}": "{{value1}}",
    "{{response_field2}}": "{{value2}}"
  }
}
` + "```" + `

### Error Responses
- ` + "`400`" + `: Invalid request data
- ` + "`401`" + `: Authentication required
- ` + "`403`" + `: Insufficient permissions
- ` + "`404`" + `: Resource not found

## Validation Rules
1. {{VALIDATION_RULE_1}}
2. {{VALIDATION_RULE_2}}
3. {{VALIDATION_RULE_3}}

## Business Logic
1. {{LOGIC_STEP_1}}
2. {{LOGIC_STEP_2}}
3. {{LOGIC_STEP_3}}

## Database Operations
- [ ] Query {{TABLE_1}} for {{DATA_1}}
- [ ] Validate {{CONDITION}}
- [ ] Insert/Update {{TABLE_2}}
- [ ] Return formatted response

## Security Requirements
- [ ] Authenticate user
- [ ] Authorize for resource access
- [ ] Validate all inputs
- [ ] Sanitize outputs
- [ ] Rate limiting: {{RATE_LIMIT}} requests per {{TIME_PERIOD}}

## Testing Requirements
- [ ] Unit tests for business logic
- [ ] Integration tests with database
- [ ] API contract tests
- [ ] Security tests
- [ ] Performance tests

## Documentation
Update OpenAPI/Swagger specification with new endpoint details.`,
				},
				// Rules
				{
					Name:     "Web Accessibility",
					Filename: "example-web-accessibility.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "web", "accessibility", "a11y"},
					Content: `# Web Accessibility Rules (WCAG 2.1 AA)

## Mandatory Accessibility Requirements

### Semantic HTML
- **MUST** use semantic HTML5 elements
- **MUST** use headings in hierarchical order (h1 → h6)
- **MUST** use ` + "`<button>`" + ` for actions, ` + "`<a>`" + ` for navigation
- **MUST** use landmark elements (` + "`<nav>`" + `, ` + "`<main>`" + `, ` + "`<aside>`" + `)

### Keyboard Navigation
- **MUST** be fully keyboard navigable
- **MUST** show visible focus indicators
- **MUST** implement logical tab order
- **MUST** provide skip links for navigation
- **MUST** trap focus in modals/dialogs

### Screen Reader Support
- **MUST** provide alt text for images
- **MUST** use ARIA labels for icons/buttons
- **MUST** announce dynamic content changes
- **MUST** provide form field labels
- **MUST** use ARIA live regions for updates

### Color & Contrast
- **MUST** meet WCAG AA contrast ratios:
  - Normal text: 4.5:1
  - Large text: 3:1
  - UI components: 3:1
- **MUST** not rely on color alone for information
- **MUST** support dark/light mode preferences

### Forms
- **MUST** label all form inputs
- **MUST** group related fields with fieldsets
- **MUST** provide clear error messages
- **MUST** indicate required fields
- **MUST** provide input format hints

### Media
- **MUST** provide captions for videos
- **MUST** provide transcripts for audio
- **MUST** allow pause/stop for auto-playing content
- **MUST** avoid flashing content (3Hz rule)

### Responsive Design
- **MUST** support 200% zoom without horizontal scroll
- **MUST** work on mobile devices
- **MUST** support both orientations
- **MUST** have touch targets ≥ 44x44px

## Testing Requirements
- Test with keyboard only
- Test with screen readers (NVDA, JAWS, VoiceOver)
- Use axe DevTools for automated testing
- Conduct manual WCAG audit
- Test with users with disabilities when possible`,
				},
				{
					Name:     "Responsive Design",
					Filename: "example-responsive-design.md",
					Type:     models.ComponentTypeRules,
					Tags:     []string{"example", "web", "responsive", "mobile"},
					Content: `# Responsive Design Rules

## Mobile-First Approach
- **ALWAYS** start with mobile layout
- **ALWAYS** progressively enhance for larger screens
- **NEVER** hide critical content on mobile
- **ALWAYS** test on real devices

## Breakpoint Standards
` + "```css" + `
/* Mobile First Breakpoints */
/* Default: 0-639px (Mobile) */
@media (min-width: 640px)  { /* Tablet */ }
@media (min-width: 1024px) { /* Desktop */ }
@media (min-width: 1280px) { /* Large Desktop */ }
` + "```" + `

## Layout Rules
- **MUST** use flexible grids (flexbox/grid)
- **MUST** use relative units (rem, em, %, vw/vh)
- **AVOID** fixed widths except for max-width
- **MUST** allow horizontal scroll for tables on mobile

## Typography
- **Base font**: 16px minimum on mobile
- **Line height**: 1.5-1.6 for readability
- **Measure**: 45-75 characters per line
- **Scale**: Use modular scale for consistency

## Images & Media
- **MUST** use responsive images (srcset)
- **MUST** lazy load below-the-fold images
- **MUST** provide appropriate formats (WebP, AVIF)
- **MUST** optimize file sizes

## Performance
- **Mobile page weight**: < 1MB
- **Critical CSS**: Inline above-the-fold styles
- **JavaScript**: Load non-critical JS async
- **First Contentful Paint**: < 1.8s on 3G

## Touch Interactions
- **Touch targets**: Minimum 44x44px
- **Spacing**: 8px between targets
- **Gestures**: Support swipe for carousels
- **Hover states**: Provide touch alternatives

## Testing Requirements
1. Test on real devices (iOS, Android)
2. Test with throttled network (3G)
3. Test all breakpoints
4. Test orientation changes
5. Test with touch and mouse

## Common Patterns
- Hamburger menu for mobile navigation
- Stacked layouts on mobile
- Accordion/tabs for content organization
- Bottom navigation for key actions
- Sticky headers with reduced height`,
				},
			},
			Pipelines: []ExamplePipeline{
				{
					Name:        "Frontend Feature",
					Filename:    "example-frontend-feature.yaml",
					Description: "Pipeline for developing React frontend features",
					Tags:        []string{"example", "web", "frontend", "react"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-web-accessibility.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-responsive-design.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-react-architecture.md",
							Order: 3,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-create-react-component.md",
							Order: 4,
						},
					},
				},
				{
					Name:        "Backend API",
					Filename:    "example-backend-api.yaml",
					Description: "Pipeline for developing REST API endpoints",
					Tags:        []string{"example", "web", "backend", "api"},
					Components: []models.ComponentRef{
						{
							Type:  models.ComponentTypeRules,
							Path:  "../components/rules/example-security-first.md",
							Order: 1,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-rest-api-design.md",
							Order: 2,
						},
						{
							Type:  models.ComponentTypeContext,
							Path:  "../components/contexts/example-database-schema.md",
							Order: 3,
						},
						{
							Type:  models.ComponentTypePrompt,
							Path:  "../components/prompts/example-add-api-endpoint.md",
							Order: 4,
						},
					},
				},
			},
		},
	}
}