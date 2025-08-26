package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pluqqy/pluqqy-cli/internal/cli"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
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
	Tags     []string `json:"tags" yaml:"tags"`
	Path     string   `json:"path" yaml:"path"`
	Archived bool     `json:"archived" yaml:"archived"`
	Excerpt  string   `json:"excerpt,omitempty" yaml:"excerpt,omitempty"`
}

// NewSearchCommand creates the search command
func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for components and pipelines",
		Long: `Search for components and pipelines using a powerful query syntax.

Query Syntax:
  tag:api              - Find items with the "api" tag
  type:prompt          - Find all prompt components
  type:pipeline        - Find all pipelines
  status:archived      - Show archived items
  content:"error"      - Search in content
  
  Combine filters with AND:
  tag:api AND type:context

Examples:
  # Search for items with a specific tag
  pluqqy search "tag:api"
  
  # Find all prompt components
  pluqqy search "type:prompt"
  
  # Search in content
  pluqqy search "content:authentication"
  
  # Complex search
  pluqqy search "tag:api AND type:context"`,
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pluqqy directory exists
			if _, err := os.Stat(files.PluqqyDir); os.IsNotExist(err) {
				return fmt.Errorf("no .pluqqy directory found. Run 'pluqqy init' first")
			}
			return nil
		},
		RunE: runSearch,
	}

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	
	// Initialize search engine
	engine := search.NewEngine()
	
	// Build index including archived items if query contains "status:archived"
	includeArchived := strings.Contains(query, "status:archived")
	if err := engine.BuildIndexWithOptions(includeArchived); err != nil {
		return fmt.Errorf("failed to build search index: %w", err)
	}
	
	// Perform search
	results, err := engine.Search(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}
	
	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")
	
	// Format results for output
	searchResult := SearchResultOutput{
		Query:   query,
		Count:   len(results),
		Results: []SearchItemOutput{},
	}
	
	for _, r := range results {
		// Determine type string
		typeStr := string(r.Item.Type)
		if r.Item.Type == search.ItemTypeComponent {
			typeStr = strings.TrimSuffix(r.Item.SubType, "s")
		}
		
		item := SearchItemOutput{
			Name:     r.Item.Name,
			Type:     typeStr,
			Tags:     r.Item.Tags,
			Path:     r.Item.Path,
			Archived: r.Item.IsArchived,
		}
		
		// Add content excerpt if searching in content
		if strings.Contains(query, "content:") {
			excerpt := r.Item.Content
			if len(excerpt) > 200 {
				excerpt = excerpt[:200] + "..."
			}
			item.Excerpt = excerpt
		}
		
		searchResult.Results = append(searchResult.Results, item)
	}
	
	// Output results
	switch outputFormat {
	case "json", "yaml":
		return cli.OutputResults(cmd.OutOrStdout(), outputFormat, searchResult)
	default:
		return outputSearchText(cmd, searchResult)
	}
}

func outputSearchText(cmd *cobra.Command, result SearchResultOutput) error {
	if result.Count == 0 {
		cli.PrintInfo("No results found for query: %s", result.Query)
		return nil
	}
	
	fmt.Fprintf(cmd.OutOrStdout(), "\nSearch Results for: %s\n", result.Query)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))
	
	// Group by type
	byType := make(map[string][]SearchItemOutput)
	for _, item := range result.Results {
		byType[item.Type] = append(byType[item.Type], item)
	}
	
	// Display grouped results
	for _, itemType := range []string{"pipeline", "context", "prompt", "rule"} {
		items, ok := byType[itemType]
		if !ok || len(items) == 0 {
			continue
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "\n%sS (%d)\n", strings.ToUpper(itemType), len(items))
		
		table := cli.NewTableFormatter(cmd.OutOrStdout())
		if itemType == "pipeline" {
			table.Header("Name", "Tags")
		} else {
			table.Header("Name", "Tags")
		}
		
		for _, item := range items {
			tags := strings.Join(item.Tags, ", ")
			if tags == "" {
				tags = "-"
			}
			
			// Show archived indicator in name if archived
			name := item.Name
			if item.Archived {
				name = name + " [archived]"
			}
			
			table.Row(name, tags)
			
			// Show excerpt if available
			if item.Excerpt != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  └─ %s\n", cli.TruncateString(item.Excerpt, 70))
			}
		}
		
		table.Flush()
	}
	
	fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d results\n", result.Count)
	
	return nil
}