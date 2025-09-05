package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-terminal/internal/cli"
	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
	"github.com/pluqqy/pluqqy-terminal/pkg/search/unified"
)

// SearchResultOutput represents the formatted search results
type SearchResultOutput struct {
	Query   string             `json:"query" yaml:"query"`
	Count   int                `json:"count" yaml:"count"`
	Results []SearchItemOutput `json:"results" yaml:"results"`
}

// SearchItemOutput represents a single search result item
type SearchItemOutput struct {
	Name     string   `json:"name" yaml:"name"`
	Type     string   `json:"type" yaml:"type"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Path     string   `json:"path,omitempty" yaml:"path,omitempty"`
	Archived bool     `json:"archived,omitempty" yaml:"archived,omitempty"`
	Excerpt  string   `json:"excerpt,omitempty" yaml:"excerpt,omitempty"`
}

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for pipelines and components",
		Long: `Search for pipelines and components using advanced query syntax.

Examples:
  # Search by tag
  pluqqy search "tag:api"
  
  # Search by type
  pluqqy search "type:prompt"
  pluqqy search "type:pipelines"
  
  # Search in content
  pluqqy search "content:authentication"
  
  # Search archived items
  pluqqy search "status:archived"
  
  # Complex searches
  pluqqy search "tag:api AND type:context"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runSearch,
	}

	// Add output format flag
	cmd.Flags().StringP("output", "o", "text", "Output format (text, json, yaml)")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	
	// Use unified search engine directly
	searchHelper := unified.NewSearchHelper()
	includeArchived := unified.ShouldIncludeArchived(query)
	searchHelper.SetSearchOptions(includeArchived, 1000, "relevance")
	
	// Load all items
	prompts, contexts, rules, err := loadComponents(includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}
	
	pipelines, err := loadPipelines(includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load pipelines: %w", err)
	}
	
	// Perform unified search
	filteredPrompts, filteredContexts, filteredRules, filteredPipelines, err := searchHelper.UnifiedFilterAll(
		query, prompts, contexts, rules, pipelines,
	)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}
	
	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")
	
	// Format results for output
	searchResult := SearchResultOutput{
		Query:   query,
		Count:   len(filteredPrompts) + len(filteredContexts) + len(filteredRules) + len(filteredPipelines),
		Results: []SearchItemOutput{},
	}
	
	// Add pipeline results
	for _, p := range filteredPipelines {
		item := SearchItemOutput{
			Name:     p.Name,
			Type:     "pipeline",
			Tags:     p.Tags,
			Path:     p.Path,
			Archived: p.IsArchived,
		}
		searchResult.Results = append(searchResult.Results, item)
	}
	
	// Add component results
	addComponentResults(&searchResult, filteredPrompts, "prompt")
	addComponentResults(&searchResult, filteredContexts, "context")
	addComponentResults(&searchResult, filteredRules, "rule")
	
	// Output results
	switch outputFormat {
	case "json", "yaml":
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, searchResult)
	default:
		return outputSearchText(cmd, searchResult)
	}
}

func loadComponents(includeArchived bool) ([]unified.ComponentItem, []unified.ComponentItem, []unified.ComponentItem, error) {
	var prompts, contexts, rules []unified.ComponentItem
	
	// Load each component type
	for _, compType := range []string{models.ComponentTypePrompt, models.ComponentTypeContext, models.ComponentTypeRules} {
		// Load active components
		componentFiles, err := files.ListComponents(compType)
		if err != nil {
			return nil, nil, nil, err
		}
		
		for _, compFile := range componentFiles {
			compPath := filepath.Join(files.ComponentsDir, compType, compFile)
			comp, err := files.ReadComponent(compPath)
			if err != nil {
				continue
			}
			
			item := unified.ComponentItem{
				Name:         strings.TrimSuffix(compFile, ".md"),
				Path:         compPath,
				CompType:     compType,
				LastModified: comp.Modified,
				Tags:         comp.Tags,
				TokenCount:   len(comp.Content) / 4, // Rough estimate
				IsArchived:   false,
			}
			
			switch compType {
			case models.ComponentTypePrompt:
				prompts = append(prompts, item)
			case models.ComponentTypeContext:
				contexts = append(contexts, item)
			case models.ComponentTypeRules:
				rules = append(rules, item)
			}
		}
		
		// Load archived components if requested
		if includeArchived {
			archivedFiles, err := files.ListArchivedComponents(compType)
			if err == nil {
				for _, compFile := range archivedFiles {
					compPath := filepath.Join(files.ComponentsDir, compType, compFile)
					comp, err := files.ReadArchivedComponent(compPath)
					if err != nil {
						continue
					}
					
					item := unified.ComponentItem{
						Name:         strings.TrimSuffix(compFile, ".md"),
						Path:         compPath,
						CompType:     compType,
						LastModified: comp.Modified,
						Tags:         comp.Tags,
						TokenCount:   len(comp.Content) / 4,
						IsArchived:   true,
					}
					
					switch compType {
					case models.ComponentTypePrompt:
						prompts = append(prompts, item)
					case models.ComponentTypeContext:
						contexts = append(contexts, item)
					case models.ComponentTypeRules:
						rules = append(rules, item)
					}
				}
			}
		}
	}
	
	return prompts, contexts, rules, nil
}

func loadPipelines(includeArchived bool) ([]unified.PipelineItem, error) {
	var pipelines []unified.PipelineItem
	
	// Load active pipelines
	pipelineFiles, err := files.ListPipelines()
	if err != nil {
		return nil, err
	}
	
	for _, pipelineFile := range pipelineFiles {
		pipeline, err := files.ReadPipeline(pipelineFile)
		if err != nil {
			continue
		}
		
		pipelines = append(pipelines, unified.PipelineItem{
			Name:       pipeline.Name,
			Path:       pipelineFile,
			Tags:       pipeline.Tags,
			IsArchived: false,
			Modified:   time.Now(), // Would need actual file mod time
		})
	}
	
	// Load archived pipelines if requested
	if includeArchived {
		archivedFiles, err := files.ListArchivedPipelines()
		if err == nil {
			for _, pipelineFile := range archivedFiles {
				pipeline, err := files.ReadArchivedPipeline(pipelineFile)
				if err != nil {
					continue
				}
				
				pipelines = append(pipelines, unified.PipelineItem{
					Name:       pipeline.Name,
					Path:       pipelineFile,
					Tags:       pipeline.Tags,
					IsArchived: true,
					Modified:   time.Now(), // Would need actual file mod time
				})
			}
		}
	}
	
	return pipelines, nil
}

func addComponentResults(searchResult *SearchResultOutput, components []unified.ComponentItem, typeName string) {
	for _, c := range components {
		item := SearchItemOutput{
			Name:     c.Name,
			Type:     typeName,
			Tags:     c.Tags,
			Path:     c.Path,
			Archived: c.IsArchived,
		}
		searchResult.Results = append(searchResult.Results, item)
	}
}

func outputSearchText(cmd *cobra.Command, result SearchResultOutput) error {
	if result.Count == 0 {
		cli.PrintInfo("No results found for query: %s", result.Query)
		return nil
	}
	
	fmt.Printf("Search Results for: %s\n", result.Query)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println()
	
	// Group results by type
	pipelines := []SearchItemOutput{}
	prompts := []SearchItemOutput{}
	contexts := []SearchItemOutput{}
	rules := []SearchItemOutput{}
	
	for _, item := range result.Results {
		switch item.Type {
		case "pipeline":
			pipelines = append(pipelines, item)
		case "prompt":
			prompts = append(prompts, item)
		case "context":
			contexts = append(contexts, item)
		case "rule":
			rules = append(rules, item)
		}
	}
	
	// Output each type
	if len(pipelines) > 0 {
		outputTypeSection("PIPELINES", pipelines)
	}
	
	if len(contexts) > 0 {
		outputTypeSection("CONTEXTS", contexts)
	}
	
	if len(prompts) > 0 {
		outputTypeSection("PROMPTS", prompts)
	}
	
	if len(rules) > 0 {
		outputTypeSection("RULES", rules)
	}
	
	fmt.Printf("\nTotal: %d results\n", result.Count)
	
	return nil
}

func outputTypeSection(title string, items []SearchItemOutput) {
	fmt.Printf("%s (%d)\n", title, len(items))
	fmt.Println("Name  Tags")
	fmt.Println("--------------------------------------------------------------------------------")
	
	for _, item := range items {
		name := item.Name
		if len(name) > 20 {
			name = name[:20]
		}
		
		tagStr := ""
		if len(item.Tags) > 0 {
			tagStr = strings.Join(item.Tags, ", ")
		}
		
		fmt.Printf("%-20s  %s\n", name, tagStr)
		
		if item.Excerpt != "" {
			fmt.Printf("  └─ %s\n", item.Excerpt)
		}
	}
	
	fmt.Println()
}