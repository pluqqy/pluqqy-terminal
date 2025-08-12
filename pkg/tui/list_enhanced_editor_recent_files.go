package tui

import (
	"path/filepath"
	"time"
)

// RecentFilesTracker tracks recently used files for quick access
type RecentFilesTracker struct {
	RecentFiles []RecentFile
	MaxFiles    int
	LastUpdated time.Time
}

// RecentFile represents a recently used file
type RecentFile struct {
	Path        string
	Name        string
	LastUsed    time.Time
	AccessCount int
}

// NewRecentFilesTracker creates a new recent files tracker
func NewRecentFilesTracker() *RecentFilesTracker {
	return &RecentFilesTracker{
		RecentFiles: make([]RecentFile, 0, 5),
		MaxFiles:    5,
	}
}

// AddFile adds a file to the recent files list
func (rft *RecentFilesTracker) AddFile(path string) {
	// Clean the path
	cleanPath := filepath.Clean(path)
	baseName := filepath.Base(cleanPath)
	
	// Check if file already exists in list
	for i, rf := range rft.RecentFiles {
		if rf.Path == cleanPath {
			// Move to front and update
			rft.RecentFiles[i].LastUsed = time.Now()
			rft.RecentFiles[i].AccessCount++
			
			// Move to front if not already there
			if i > 0 {
				file := rft.RecentFiles[i]
				// Shift elements
				copy(rft.RecentFiles[1:i+1], rft.RecentFiles[0:i])
				rft.RecentFiles[0] = file
			}
			
			rft.LastUpdated = time.Now()
			return
		}
	}
	
	// Add new file to front
	newFile := RecentFile{
		Path:        cleanPath,
		Name:        baseName,
		LastUsed:    time.Now(),
		AccessCount: 1,
	}
	
	// Prepend to list
	rft.RecentFiles = append([]RecentFile{newFile}, rft.RecentFiles...)
	
	// Trim to max size
	if len(rft.RecentFiles) > rft.MaxFiles {
		rft.RecentFiles = rft.RecentFiles[:rft.MaxFiles]
	}
	
	rft.LastUpdated = time.Now()
}

// GetRecentFiles returns the list of recent files
func (rft *RecentFilesTracker) GetRecentFiles() []RecentFile {
	return rft.RecentFiles
}

// HasRecentFiles checks if there are any recent files
func (rft *RecentFilesTracker) HasRecentFiles() bool {
	return len(rft.RecentFiles) > 0
}

// GetFileByNumber returns a file by its display number (1-indexed)
func (rft *RecentFilesTracker) GetFileByNumber(num int) (RecentFile, bool) {
	if num < 1 || num > len(rft.RecentFiles) {
		return RecentFile{}, false
	}
	return rft.RecentFiles[num-1], true
}

// Clear removes all recent files
func (rft *RecentFilesTracker) Clear() {
	rft.RecentFiles = make([]RecentFile, 0, 5)
	rft.LastUpdated = time.Now()
}

// RemoveFile removes a specific file from the recent list
func (rft *RecentFilesTracker) RemoveFile(path string) {
	cleanPath := filepath.Clean(path)
	
	for i, rf := range rft.RecentFiles {
		if rf.Path == cleanPath {
			// Remove the file
			rft.RecentFiles = append(rft.RecentFiles[:i], rft.RecentFiles[i+1:]...)
			rft.LastUpdated = time.Now()
			return
		}
	}
}

// FormatRecentFilesList formats the recent files for display
func (rft *RecentFilesTracker) FormatRecentFilesList() []string {
	if !rft.HasRecentFiles() {
		return nil
	}
	
	formatted := make([]string, 0, len(rft.RecentFiles))
	for i, rf := range rft.RecentFiles {
		// Format as "1. filename.ext"
		formatted = append(formatted, formatRecentFile(i+1, rf))
	}
	
	return formatted
}

// formatRecentFile formats a single recent file entry
func formatRecentFile(num int, rf RecentFile) string {
	// Show relative path if it's not too long
	display := rf.Name
	dir := filepath.Dir(rf.Path)
	if dir != "." && len(dir) < 30 {
		display = filepath.Join(dir, rf.Name)
	}
	
	return display
}