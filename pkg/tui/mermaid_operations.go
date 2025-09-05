package tui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

//go:embed assets/mermaid-template.html
var mermaidTemplateContent string

// Parsed template cached at package level
var mermaidTemplate *template.Template

func init() {
	var err error
	mermaidTemplate, err = template.New("mermaid").Parse(mermaidTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse mermaid template: %v", err))
	}
}

// MermaidOperator handles mermaid diagram generation
type MermaidOperator struct {
	state *MermaidState
}

// NewMermaidOperator creates a new mermaid operator
func NewMermaidOperator(state *MermaidState) *MermaidOperator {
	return &MermaidOperator{
		state: state,
	}
}

// GeneratePipelineDiagram generates and opens a mermaid diagram for the pipeline
func (mo *MermaidOperator) GeneratePipelineDiagram(pipeline pipelineItem) tea.Cmd {
	return func() tea.Msg {
		mo.state.StartGeneration()

		// Load full pipeline data
		p, err := files.ReadPipeline(pipeline.path)
		if err != nil {
			mo.state.FailGeneration(err)
			return StatusMsg(fmt.Sprintf("Failed to load pipeline '%s': %v", pipeline.name, err))
		}

		// Generate HTML with mermaid diagram
		html, err := mo.generateMermaidHTML(p)
		if err != nil {
			mo.state.FailGeneration(err)
			return StatusMsg(fmt.Sprintf("Failed to generate diagram for '%s': %v", pipeline.name, err))
		}

		// Save to tmp/diagrams directory
		timestamp := time.Now().Format("20060102-150405")
		sanitizedName := sanitizeFilename(pipeline.name)
		filename := fmt.Sprintf("pipeline-%s-%s.html", sanitizedName, timestamp)

		diagramDir := filepath.Join(files.PluqqyDir, "tmp", MermaidTmpSubdir)
		if err := os.MkdirAll(diagramDir, 0755); err != nil {
			mo.state.FailGeneration(err)
			return StatusMsg(fmt.Sprintf("Failed to create diagram directory: %v", err))
		}

		fullPath := filepath.Join(diagramDir, filename)

		// Write HTML file
		if err := os.WriteFile(fullPath, []byte(html), 0644); err != nil {
			mo.state.FailGeneration(err)
			return StatusMsg(fmt.Sprintf("Failed to save diagram: %v", err))
		}

		// Open in browser
		if err := mo.openInBrowser(fullPath); err != nil {
			mo.state.FailGeneration(err)
			return StatusMsg(fmt.Sprintf("Diagram saved but failed to open browser: %v", err))
		}

		mo.state.CompleteGeneration(filename)
		return StatusMsg(fmt.Sprintf("✓ Diagram generated: %s", filename))
	}
}

// generateMermaidHTML generates the complete HTML with embedded mermaid diagram
func (mo *MermaidOperator) generateMermaidHTML(pipeline *models.Pipeline) (string, error) {
	// Generate mermaid graph syntax
	mermaidDiagram := mo.generateMermaidGraph(pipeline)

	// Generate component data for tooltips
	componentData, err := mo.generateComponentData(pipeline)
	if err != nil {
		return "", fmt.Errorf("failed to generate component data: %w", err)
	}

	// Debug: Log component data
	if componentData == "{}" || componentData == "" {
		fmt.Printf("Warning: Component data is empty for pipeline %s\n", pipeline.Name)
	}

	// Calculate metadata
	tokenCount := mo.calculateTokenCount(pipeline)

	// Prepare template data
	data := map[string]interface{}{
		"PipelineName":   pipeline.Name,
		"Timestamp":      time.Now().Format("2006-01-02 15:04:05"),
		"ComponentCount": len(pipeline.Components),
		"Tags":           strings.Join(pipeline.Tags, ", "),
		"TokenCount":     tokenCount,
		"MermaidDiagram": mermaidDiagram,
		"ComponentData":  template.JS(componentData), // Safe JS embedding
	}

	// Execute template
	var buf strings.Builder
	if err := mermaidTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// generateMermaidGraph creates the mermaid graph definition
func (mo *MermaidOperator) generateMermaidGraph(pipeline *models.Pipeline) string {
	var graph strings.Builder

	// Start graph
	graph.WriteString("graph TD\n")

	// Add pipeline node
	mo.writePipelineNode(&graph, pipeline)

	// Group components by type
	contexts, prompts, rules := mo.groupComponentsByType(pipeline.Components)

	// Load user settings to get section order and names
	settings, err := files.ReadSettings()
	if err != nil {
		// Use defaults if can't read settings
		settings = models.DefaultSettings()
	}

	// Generate subgraphs according to user's section order
	prevGroup := "Pipeline"

	for _, section := range settings.Output.Formatting.Sections {
		var components []models.ComponentRef
		var cssClass string
		var groupID string
		var groupLabel string

		switch section.Type {
		case "contexts":
			if len(contexts) == 0 {
				continue
			}
			components = contexts
			cssClass = "context"
			// Extract clean ID and preserve original for label
			groupID = extractSectionName(section.Heading)
			groupLabel = extractSectionLabel(section.Heading)
		case "prompts":
			if len(prompts) == 0 {
				continue
			}
			components = prompts
			cssClass = "prompt"
			groupID = extractSectionName(section.Heading)
			groupLabel = extractSectionLabel(section.Heading)
		case "rules":
			if len(rules) == 0 {
				continue
			}
			components = rules
			cssClass = "rules"
			groupID = extractSectionName(section.Heading)
			groupLabel = extractSectionLabel(section.Heading)
		default:
			continue
		}

		mo.writeComponentSubgraphWithLabel(&graph, groupID, groupLabel, components, cssClass, prevGroup)
		prevGroup = groupID
	}

	// Add class definitions
	mo.writeClassDefinitions(&graph)

	return graph.String()
}

// writePipelineNode adds the main pipeline node
func (mo *MermaidOperator) writePipelineNode(graph *strings.Builder, pipeline *models.Pipeline) {
	tokenCount := mo.calculateTokenCount(pipeline)
	graph.WriteString(fmt.Sprintf(
		`    Pipeline["%s<br/>%d Components | ~%d Tokens"]:::pipeline`,
		strings.ToUpper(pipeline.Name),
		len(pipeline.Components),
		tokenCount,
	))
	graph.WriteString("\n\n")
}

// groupComponentsByType organizes components into categories
func (mo *MermaidOperator) groupComponentsByType(components []models.ComponentRef) (contexts, prompts, rules []models.ComponentRef) {
	for _, comp := range components {
		switch comp.Type {
		case "contexts":
			contexts = append(contexts, comp)
		case "prompts":
			prompts = append(prompts, comp)
		case "rules":
			rules = append(rules, comp)
		}
	}
	return
}

// writeComponentSubgraph generates a subgraph for a component type (backward compatibility)
func (mo *MermaidOperator) writeComponentSubgraph(
	graph *strings.Builder,
	groupName string,
	components []models.ComponentRef,
	cssClass string,
	prevGroup string,
) {
	mo.writeComponentSubgraphWithLabel(graph, groupName, groupName, components, cssClass, prevGroup)
}

// writeComponentSubgraphWithLabel generates a subgraph with separate ID and label
func (mo *MermaidOperator) writeComponentSubgraphWithLabel(
	graph *strings.Builder,
	groupID string,
	groupLabel string,
	components []models.ComponentRef,
	cssClass string,
	prevGroup string,
) {
	// Add connection from previous group
	graph.WriteString(fmt.Sprintf("    %s --- %s\n\n", prevGroup, groupID))

	// Start subgraph - use ID for the subgraph identifier, label for display
	graph.WriteString(fmt.Sprintf(`    subgraph %s["%s"]`, groupID, groupLabel))
	graph.WriteString("\n")

	// Add components - use first letter of ID for component IDs
	idPrefix := "C"
	if len(groupID) > 0 {
		idPrefix = string(groupID[0])
	}

	for i, comp := range components {
		id := fmt.Sprintf("%s%d", idPrefix, i+1)
		name := extractComponentName(comp.Path)

		graph.WriteString(fmt.Sprintf(
			`        %s["%s"]:::%s`,
			id, name, cssClass,
		))
		graph.WriteString("\n")
	}

	graph.WriteString("    end\n\n")
}

// writeClassDefinitions adds mermaid class definitions
func (mo *MermaidOperator) writeClassDefinitions(graph *strings.Builder) {
	graph.WriteString("\n")
	graph.WriteString(fmt.Sprintf(
		"    classDef pipeline fill:%s,stroke:%s,stroke-width:%s,color:%s\n",
		MermaidPipelineColor, MermaidPipelineStroke,
		MermaidPipelineStrokeWidth, MermaidBackgroundColor,
	))
	graph.WriteString(fmt.Sprintf(
		"    classDef context fill:%s,stroke:%s,stroke-width:%s,color:%s\n",
		MermaidContextColor, MermaidContextStroke,
		MermaidStrokeWidth, MermaidBackgroundColor,
	))
	graph.WriteString(fmt.Sprintf(
		"    classDef prompt fill:%s,stroke:%s,stroke-width:%s,color:%s\n",
		MermaidPromptColor, MermaidPromptStroke,
		MermaidStrokeWidth, MermaidBackgroundColor,
	))
	graph.WriteString(fmt.Sprintf(
		"    classDef rules fill:%s,stroke:%s,stroke-width:%s,color:%s\n",
		MermaidRulesColor, MermaidRulesStroke,
		MermaidStrokeWidth, MermaidBackgroundColor,
	))
}

// generateComponentData creates JSON data for component tooltips
func (mo *MermaidOperator) generateComponentData(pipeline *models.Pipeline) (string, error) {
	data := make(map[string]interface{})

	// Group components by type first, matching the graph generation order
	contexts, prompts, rules := mo.groupComponentsByType(pipeline.Components)

	// Load user settings to ensure same order as graph
	settings, err := files.ReadSettings()
	if err != nil {
		settings = models.DefaultSettings()
	}

	// Process components in the same order as the graph
	contextCount, promptCount, rulesCount := 0, 0, 0

	for _, section := range settings.Output.Formatting.Sections {
		var components []models.ComponentRef
		var prefix string
		var typeName string
		var counter *int

		switch section.Type {
		case "contexts":
			components = contexts
			prefix = "C"
			typeName = "Context"
			counter = &contextCount
		case "prompts":
			components = prompts
			prefix = "P"
			typeName = "Prompt"
			counter = &promptCount
		case "rules":
			components = rules
			prefix = "R"
			typeName = "Rules"
			counter = &rulesCount
		default:
			continue
		}

		// Process this section's components
		for _, compRef := range components {
			*counter++

			// Clean up the path - remove leading "../" if present
			cleanPath := strings.TrimPrefix(compRef.Path, "../")
			comp, err := files.ReadComponent(cleanPath)
			if err != nil {
				fmt.Printf("Warning: Failed to read %s component %s: %v\n", section.Type, cleanPath, err)
				continue
			}

			id := fmt.Sprintf("%s%d", prefix, *counter)
			name := extractComponentName(compRef.Path)

			data[id] = map[string]interface{}{
				"name":    name,
				"type":    typeName,
				"content": comp.Content,
				"tokens":  estimateTokens(comp.Content),
				"tags":    comp.Tags,
			}
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// openInBrowser opens the file in the default browser
func (mo *MermaidOperator) openInBrowser(filepath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filepath)
	case "linux":
		cmd = exec.Command("xdg-open", filepath)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filepath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// Helper functions

// calculateTokenCount estimates total tokens for a pipeline
func (mo *MermaidOperator) calculateTokenCount(pipeline *models.Pipeline) int {
	totalTokens := 0

	for _, compRef := range pipeline.Components {
		// Try to load actual component for accurate count
		if comp, err := files.ReadComponent(compRef.Path); err == nil {
			totalTokens += estimateTokens(comp.Content)
		} else {
			// Fallback to estimate
			totalTokens += EstimatedTokensPerComponent
		}
	}

	return totalTokens
}

// extractComponentName gets the component name from its path
func extractComponentName(path string) string {
	// Extract base name without extension
	name := filepath.Base(path)
	return strings.TrimSuffix(name, ".md")
}

// sanitizeFilename makes a filename safe for filesystem
func sanitizeFilename(name string) string {
	// Replace spaces and special chars
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}

// estimateTokens provides a rough token count estimate
func estimateTokens(content string) int {
	return len(content) / TokensPerCharacter
}

// extractSectionName extracts the section name from a heading
// e.g., "## CONTEXT" -> "CONTEXT", "# My Rules" -> "MY_RULES"
func extractSectionName(heading string) string {
	// Remove markdown heading markers
	name := strings.TrimPrefix(heading, "###")
	name = strings.TrimPrefix(name, "##")
	name = strings.TrimPrefix(name, "#")
	name = strings.TrimSpace(name)

	// Convert to uppercase for consistency in diagram
	name = strings.ToUpper(name)

	// Replace spaces and special characters with underscores for valid Mermaid IDs
	// Mermaid doesn't like spaces or special chars in node IDs
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		".", "_",
		",", "_",
		":", "_",
		";", "_",
		"!", "_",
		"?", "_",
		"(", "_",
		")", "_",
		"[", "_",
		"]", "_",
		"{", "_",
		"}", "_",
		"/", "_",
		"\\", "_",
		"'", "_",
		"\"", "_",
		"`", "_",
	)
	name = replacer.Replace(name)

	// Remove any remaining non-alphanumeric characters except underscores
	var cleaned strings.Builder
	for _, r := range name {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}
	name = cleaned.String()

	// If empty after cleaning, return a default
	if name == "" {
		return "SECTION"
	}

	// Ensure it doesn't start with a number (invalid in Mermaid)
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "S_" + name
	}

	return name
}

// extractSectionLabel extracts the display label from a heading
// e.g., "## CONTEXT" -> "CONTEXT", "# My Rules" -> "MY RULES"
func extractSectionLabel(heading string) string {
	// Remove markdown heading markers
	label := strings.TrimPrefix(heading, "###")
	label = strings.TrimPrefix(label, "##")
	label = strings.TrimPrefix(label, "#")
	label = strings.TrimSpace(label)

	// Convert to uppercase for consistency in diagram
	label = strings.ToUpper(label)

	// If empty, return a default
	if label == "" {
		return "SECTION"
	}

	return label
}
