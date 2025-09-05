package tui

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
)

// ComponentUsageOperator handles operations for finding component usage
type ComponentUsageOperator struct{}

// NewComponentUsageOperator creates a new component usage operator
func NewComponentUsageOperator() *ComponentUsageOperator {
	return &ComponentUsageOperator{}
}

// FindPipelinesUsingComponent finds all pipelines that use a specific component
func (cuo *ComponentUsageOperator) FindPipelinesUsingComponent(componentPath string) []PipelineUsageInfo {
	var usageList []PipelineUsageInfo
	
	// Normalize the component path for comparison
	normalizedComponentPath := filepath.Clean(componentPath)
	
	// Get all pipelines (including archived)
	pipelines, err := files.ListPipelines()
	if err != nil {
		// Return empty list on error
		return usageList
	}
	
	// Also check archived pipelines
	archivedPipelines, err := files.ListArchivedPipelines()
	if err == nil {
		pipelines = append(pipelines, archivedPipelines...)
	}
	
	// Check each pipeline
	for _, pipelinePath := range pipelines {
		pipeline, err := files.ReadPipeline(pipelinePath)
		if err != nil {
			// Try reading as archived pipeline
			pipeline, err = files.ReadArchivedPipeline(pipelinePath)
			if err != nil {
				// Skip pipelines that can't be read
				continue
			}
		}
		
		// Check if this pipeline uses the component
		for _, comp := range pipeline.Components {
			// Normalize both paths - remove any ../ prefix and clean
			normalizedCompPath := filepath.Clean(comp.Path)
			if strings.HasPrefix(normalizedCompPath, "../") {
				normalizedCompPath = strings.TrimPrefix(normalizedCompPath, "../")
			}
			// Also normalize the component path we're looking for
			compareComponentPath := normalizedComponentPath
			if strings.HasPrefix(compareComponentPath, files.PluqqyDir+"/") {
				compareComponentPath = strings.TrimPrefix(compareComponentPath, files.PluqqyDir+"/")
			}
			if normalizedCompPath == compareComponentPath {
				usageList = append(usageList, PipelineUsageInfo{
					Name:            pipeline.Name,
					Path:            pipelinePath,
					ComponentOrder:  comp.Order,
					TotalComponents: len(pipeline.Components),
				})
				break // Component found in this pipeline, move to next
			}
		}
	}
	
	// Sort by pipeline name for consistent display
	sort.Slice(usageList, func(i, j int) bool {
		return usageList[i].Name < usageList[j].Name
	})
	
	return usageList
}

// GetComponentUsageCount returns the number of pipelines using a component
func (cuo *ComponentUsageOperator) GetComponentUsageCount(componentPath string) int {
	usageList := cuo.FindPipelinesUsingComponent(componentPath)
	return len(usageList)
}

// FindComponentsInPipeline returns all components used by a specific pipeline
func (cuo *ComponentUsageOperator) FindComponentsInPipeline(pipelinePath string) []componentItem {
	var components []componentItem
	
	// Read the pipeline
	pipeline, err := files.ReadPipeline(pipelinePath)
	if err != nil {
		// Try reading as archived pipeline
		pipeline, err = files.ReadArchivedPipeline(pipelinePath)
		if err != nil {
			return components
		}
	}
	
	// Get details for each component
	for _, compRef := range pipeline.Components {
		// Try to read the component to get its details
		comp, err := files.ReadComponent(compRef.Path)
		if err != nil {
			// Try archived component
			comp, err = files.ReadArchivedComponent(compRef.Path)
			if err != nil {
				// Create a minimal component item if we can't read it
				components = append(components, componentItem{
					name:     filepath.Base(compRef.Path),
					path:     compRef.Path,
					compType: compRef.Type,
				})
				continue
			}
		}
		
		// Get usage count for this component
		usageCount := cuo.GetComponentUsageCount(compRef.Path)
		
		components = append(components, componentItem{
			name:         comp.Name,
			path:         comp.Path,
			compType:     comp.Type,
			lastModified: comp.Modified,
			usageCount:   usageCount,
			tags:         comp.Tags,
		})
	}
	
	return components
}

// FindOrphanedComponents returns components not used by any pipeline
func (cuo *ComponentUsageOperator) FindOrphanedComponents() []componentItem {
	var orphaned []componentItem
	
	// Get all components
	allComponents := []string{}
	
	// Get prompts
	prompts, _ := files.ListComponents(files.PromptsDir)
	allComponents = append(allComponents, prompts...)
	
	// Get contexts
	contexts, _ := files.ListComponents(files.ContextsDir)
	allComponents = append(allComponents, contexts...)
	
	// Get rules
	rules, _ := files.ListComponents(files.RulesDir)
	allComponents = append(allComponents, rules...)
	
	// Check usage for each component
	for _, compPath := range allComponents {
		usageCount := cuo.GetComponentUsageCount(compPath)
		if usageCount == 0 {
			// Read component details
			comp, err := files.ReadComponent(compPath)
			if err != nil {
				continue
			}
			
			orphaned = append(orphaned, componentItem{
				name:         comp.Name,
				path:         comp.Path,
				compType:     comp.Type,
				lastModified: comp.Modified,
				usageCount:   0,
				tags:         comp.Tags,
			})
		}
	}
	
	// Sort by name
	sort.Slice(orphaned, func(i, j int) bool {
		return orphaned[i].name < orphaned[j].name
	})
	
	return orphaned
}