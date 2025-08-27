package unified

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
	"github.com/pluqqy/pluqqy-cli/pkg/search"
)

// LegacyAdapter provides a compatibility layer for the old search.Engine API
type LegacyAdapter struct {
	helper *SearchHelper
	parser *search.Parser
	includeArchived bool
}

// NewLegacyAdapter creates a new adapter for the legacy search API
func NewLegacyAdapter() *LegacyAdapter {
	return &LegacyAdapter{
		helper: NewSearchHelper(),
		parser: search.NewParser(),
		includeArchived: false,
	}
}

// BuildIndex builds the search index (loads all items)
func (la *LegacyAdapter) BuildIndex() error {
	return la.BuildIndexWithOptions(false)
}

// BuildIndexWithOptions builds the search index with options
func (la *LegacyAdapter) BuildIndexWithOptions(includeArchived bool) error {
	la.includeArchived = includeArchived
	
	// Validate that components can be loaded
	_, err := la.loadComponents(models.ComponentTypePrompt, includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}
	
	_, err = la.loadComponents(models.ComponentTypeContext, includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load contexts: %w", err)
	}
	
	_, err = la.loadComponents(models.ComponentTypeRules, includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}
	
	// Validate that pipelines can be loaded
	_, err = la.loadPipelines(includeArchived)
	if err != nil {
		return fmt.Errorf("failed to load pipelines: %w", err)
	}
	
	// Configure search helper
	la.helper.SetSearchOptions(includeArchived, 1000, "relevance")
	
	// The unified engine will handle loading into its internal structures
	// when UnifiedFilterAll is called
	
	return nil
}

// Search performs a search using the unified engine
func (la *LegacyAdapter) Search(queryStr string) ([]search.SearchResult, error) {
	// Load all items (this would normally be cached in a real implementation)
	prompts, _ := la.loadComponents(models.ComponentTypePrompt, la.includeArchived)
	contexts, _ := la.loadComponents(models.ComponentTypeContext, la.includeArchived)
	rules, _ := la.loadComponents(models.ComponentTypeRules, la.includeArchived)
	pipelines, _ := la.loadPipelines(la.includeArchived)
	
	// Perform unified search
	filteredPrompts, filteredContexts, filteredRules, filteredPipelines, err := la.helper.UnifiedFilterAll(
		queryStr, prompts, contexts, rules, pipelines,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	
	// Convert results to legacy format
	var results []search.SearchResult
	
	// Add component results
	for _, prompt := range filteredPrompts {
		results = append(results, la.convertComponentToLegacy(prompt))
	}
	for _, context := range filteredContexts {
		results = append(results, la.convertComponentToLegacy(context))
	}
	for _, rule := range filteredRules {
		results = append(results, la.convertComponentToLegacy(rule))
	}
	
	// Add pipeline results
	for _, pipeline := range filteredPipelines {
		results = append(results, la.convertPipelineToLegacy(pipeline))
	}
	
	return results, nil
}

// Helper methods

func (la *LegacyAdapter) loadComponents(compType string, includeArchived bool) ([]ComponentItem, error) {
	var components []ComponentItem
	
	// Load active components
	componentFiles, err := files.ListComponents(compType)
	if err != nil {
		return nil, err
	}
	
	for _, compFile := range componentFiles {
		compPath := filepath.Join(files.ComponentsDir, compType, compFile)
		comp, err := files.ReadComponent(compPath)
		if err != nil {
			continue
		}
		
		components = append(components, ComponentItem{
			Name:         strings.TrimSuffix(compFile, ".md"),
			Path:         compPath,
			CompType:     compType,
			LastModified: comp.Modified,
			Tags:         comp.Tags,
			IsArchived:   false,
		})
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
				
				components = append(components, ComponentItem{
					Name:         strings.TrimSuffix(compFile, ".md"),
					Path:         compPath,
					CompType:     compType,
					LastModified: comp.Modified,
					Tags:         comp.Tags,
					IsArchived:   true,
				})
			}
		}
	}
	
	return components, nil
}

func (la *LegacyAdapter) loadPipelines(includeArchived bool) ([]PipelineItem, error) {
	var pipelines []PipelineItem
	
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
		
		pipelines = append(pipelines, PipelineItem{
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
				
				pipelines = append(pipelines, PipelineItem{
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

func (la *LegacyAdapter) convertComponentToLegacy(item ComponentItem) search.SearchResult {
	// Determine subtype singular form
	subType := item.CompType
	if strings.HasSuffix(subType, "s") && subType != "rules" {
		subType = strings.TrimSuffix(subType, "s")
	}
	
	return search.SearchResult{
		Item: search.SearchItem{
			Type:       search.ItemTypeComponent,
			SubType:    item.CompType,
			Path:       item.Path,
			Name:       item.Name,
			Tags:       item.Tags,
			Content:    "", // Would need to load content
			Modified:   item.LastModified,
			TokenCount: item.TokenCount,
			IsArchived: item.IsArchived,
		},
		Score:      1.0,
		Highlights: make(map[string][]string),
	}
}

func (la *LegacyAdapter) convertPipelineToLegacy(item PipelineItem) search.SearchResult {
	return search.SearchResult{
		Item: search.SearchItem{
			Type:       search.ItemTypePipeline,
			Path:       item.Path,
			Name:       item.Name,
			Tags:       item.Tags,
			Content:    "", // Would need to compose content
			Modified:   item.Modified,
			TokenCount: item.TokenCount,
			IsArchived: item.IsArchived,
		},
		Score:      1.0,
		Highlights: make(map[string][]string),
	}
}