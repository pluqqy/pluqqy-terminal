package unified

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pluqqy/pluqqy-terminal/pkg/files"
	"github.com/pluqqy/pluqqy-terminal/pkg/tui/shared"
)

// ConvertTUIComponentItemToShared converts a TUI componentItem to unified ComponentItem
func ConvertTUIComponentItemToShared(name, path, compType string, lastModified time.Time, usageCount, tokenCount int, tags []string, isArchived bool) ComponentItem {
	return ComponentItem{
		Name:         name,
		Path:         path,
		CompType:     compType,
		LastModified: lastModified,
		UsageCount:   usageCount,
		TokenCount:   tokenCount,
		Tags:         tags,
		IsArchived:   isArchived,
	}
}

// ConvertTUIPipelineItemToShared converts a TUI pipelineItem to unified PipelineItem
func ConvertTUIPipelineItemToShared(name, path string, tags []string, tokenCount int, isArchived bool) PipelineItem {
	// Get modified time from file system
	var modTime time.Time
	if isArchived {
		archivedPath := filepath.Join(files.PluqqyDir, files.ArchiveDir, files.PipelinesDir, path)
		if stat, err := os.Stat(archivedPath); err == nil {
			modTime = stat.ModTime()
		}
	} else {
		pipelinePath := filepath.Join(files.PluqqyDir, files.PipelinesDir, path)
		if stat, err := os.Stat(pipelinePath); err == nil {
			modTime = stat.ModTime()
		}
	}
	
	return PipelineItem{
		Name:       name,
		Path:       path,
		Tags:       tags,
		TokenCount: tokenCount,
		IsArchived: isArchived,
		Modified:   modTime,
	}
}

// ConvertSharedComponentItemToUnified converts a shared.ComponentItem to unified.ComponentItem
func ConvertSharedComponentItemToUnified(item shared.ComponentItem) ComponentItem {
	return ComponentItem{
		Name:         item.Name,
		Path:         item.Path,
		CompType:     item.CompType,
		LastModified: item.LastModified,
		UsageCount:   item.UsageCount,
		TokenCount:   item.TokenCount,
		Tags:         item.Tags,
		IsArchived:   item.IsArchived,
	}
}

// ConvertSharedComponentItemsToUnified converts a slice of shared.ComponentItem to unified.ComponentItem
func ConvertSharedComponentItemsToUnified(items []shared.ComponentItem) []ComponentItem {
	result := make([]ComponentItem, len(items))
	for i, item := range items {
		result[i] = ConvertSharedComponentItemToUnified(item)
	}
	return result
}