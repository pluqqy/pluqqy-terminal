package unified_test

import (
	"fmt"
	"time"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/search/unified"
)

// ExampleItem implements the Searchable interface for demonstration
type ExampleItem struct {
	Name     string
	Type     string
	Tags     []string
	Content  string
	Modified time.Time
}

func (e *ExampleItem) GetName() string        { return e.Name }
func (e *ExampleItem) GetPath() string        { return "" }
func (e *ExampleItem) GetType() string        { return e.Type }
func (e *ExampleItem) GetSubType() string     { return "" }
func (e *ExampleItem) GetTags() []string      { return e.Tags }
func (e *ExampleItem) GetContent() string     { return e.Content }
func (e *ExampleItem) GetModified() time.Time { return e.Modified }
func (e *ExampleItem) IsArchived() bool       { return false }
func (e *ExampleItem) GetTokenCount() int     { return 0 }
func (e *ExampleItem) GetUsageCount() int     { return 0 }

// Example demonstrates using the unified search engine with multiple filters
func Example_multiFilterSearch() {
	// Create some example items
	items := []*ExampleItem{
		{
			Name:    "TUI Error Handler",
			Type:    "component",
			Tags:    []string{"tui", "error"},
			Content: "Handles errors in the TUI with proper coding standards",
		},
		{
			Name:    "API Error Handler",
			Type:    "component",
			Tags:    []string{"api", "error"},
			Content: "Handles API errors with logging",
		},
		{
			Name:    "TUI Input Validator",
			Type:    "component",
			Tags:    []string{"tui", "validation"},
			Content: "Validates user input in the TUI",
		},
	}
	
	// Create search engine
	engine := unified.NewSearchEngine[*ExampleItem]()
	engine.SetItems(items)
	
	// Example 1: Single filter - find all TUI components
	results, _ := engine.Search("tag:tui")
	fmt.Printf("Search 'tag:tui' found %d items:\n", len(results))
	for _, r := range results {
		fmt.Printf("  - %s\n", r.Item.GetName())
	}
	
	// Example 2: Multiple filters with AND logic
	results, _ = engine.Search("tag:tui content:error")
	fmt.Printf("\nSearch 'tag:tui content:error' found %d items:\n", len(results))
	for _, r := range results {
		fmt.Printf("  - %s\n", r.Item.GetName())
	}
	
	// Example 3: Filter with free text
	results, _ = engine.Search("tag:tui Handler")
	fmt.Printf("\nSearch 'tag:tui Handler' found %d items:\n", len(results))
	for _, r := range results {
		fmt.Printf("  - %s\n", r.Item.GetName())
	}
	
	// Output:
	// Search 'tag:tui' found 2 items:
	//   - TUI Error Handler
	//   - TUI Input Validator
	//
	// Search 'tag:tui content:error' found 1 items:
	//   - TUI Error Handler
	//
	// Search 'tag:tui Handler' found 1 items:
	//   - TUI Error Handler
}