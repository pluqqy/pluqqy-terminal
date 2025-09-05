package files

import (
	"path/filepath"
	"strings"
	
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// LoadComponent is a convenience wrapper for ReadComponent
// It reads a component from the specified path, handling both absolute and relative paths
func LoadComponent(path string) (*models.Component, error) {
	// If path already contains .pluqqy, extract the relative part
	if strings.Contains(path, PluqqyDir) {
		// Remove the .pluqqy/ prefix to get the relative path
		idx := strings.Index(path, PluqqyDir)
		if idx >= 0 {
			path = path[idx+len(PluqqyDir)+1:] // +1 for the trailing slash
		}
	}
	return ReadComponent(path)
}

// LoadPipeline is a convenience wrapper for ReadPipeline
// It reads a pipeline from the specified path, handling both absolute and relative paths
func LoadPipeline(path string) (*models.Pipeline, error) {
	// If path already contains .pluqqy/pipelines, extract just the filename
	if strings.Contains(path, filepath.Join(PluqqyDir, PipelinesDir)) {
		// Get just the filename
		path = filepath.Base(path)
	} else if strings.Contains(path, PluqqyDir) {
		// If it just contains .pluqqy but not pipelines, remove .pluqqy/ prefix
		idx := strings.Index(path, PluqqyDir)
		if idx >= 0 {
			path = path[idx+len(PluqqyDir)+1:]
			// Now remove pipelines/ if present
			if strings.HasPrefix(path, PipelinesDir+"/") {
				path = path[len(PipelinesDir)+1:]
			}
		}
	}
	return ReadPipeline(path)
}