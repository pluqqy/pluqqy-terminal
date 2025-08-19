package tui

import (
	"path/filepath"
	"strings"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
}

// FileReferenceState manages file browser state ONLY - no file system operations
type FileReferenceState struct {
	// Basic state
	Active bool

	// Navigation state
	CurrentPath   string
	SelectedIndex int
	ScrollOffset  int

	// File data
	FileList     []FileInfo
	FilteredList []FileInfo

	// Search/filter state
	FilterPattern string
	ShowHidden    bool

	// Selection state
	SelectedFile string

	// View configuration
	MaxVisible int
}

// NewFileReferenceState creates a new file reference state
func NewFileReferenceState() *FileReferenceState {
	return &FileReferenceState{
		Active:        false,
		CurrentPath:   "",
		SelectedIndex: 0,
		ScrollOffset:  0,
		FileList:      []FileInfo{},
		FilteredList:  []FileInfo{},
		FilterPattern: "",
		ShowHidden:    false,
		SelectedFile:  "",
		MaxVisible:    10,
	}
}

// SetCurrentPath updates the current directory path
func (f *FileReferenceState) SetCurrentPath(path string) {
	f.CurrentPath = path
	f.SelectedIndex = 0
	f.ScrollOffset = 0
}

// SetFileList updates the list of files
func (f *FileReferenceState) SetFileList(files []FileInfo) {
	f.FileList = files
	f.applyFilter()
}

// GetSelectedFile returns the currently selected file
func (f *FileReferenceState) GetSelectedFile() string {
	if f.SelectedIndex >= 0 && f.SelectedIndex < len(f.FilteredList) {
		return f.FilteredList[f.SelectedIndex].Path
	}
	return ""
}

// IsActive checks if the file reference browser is active
func (f *FileReferenceState) IsActive() bool {
	return f.Active
}

// Start activates the file reference browser
func (f *FileReferenceState) Start(initialPath string) {
	f.Active = true
	f.CurrentPath = initialPath
	f.SelectedIndex = 0
	f.ScrollOffset = 0
	f.FilterPattern = ""
	f.SelectedFile = ""
}

// Stop deactivates the file reference browser
func (f *FileReferenceState) Stop() {
	f.Active = false
	f.SelectedFile = f.GetSelectedFile()
}

// MoveUp moves the selection up
func (f *FileReferenceState) MoveUp() {
	if f.SelectedIndex > 0 {
		f.SelectedIndex--
		f.adjustScroll()
	}
}

// MoveDown moves the selection down
func (f *FileReferenceState) MoveDown() {
	if f.SelectedIndex < len(f.FilteredList)-1 {
		f.SelectedIndex++
		f.adjustScroll()
	}
}

// PageUp moves selection up by a page
func (f *FileReferenceState) PageUp() {
	f.SelectedIndex -= f.MaxVisible
	if f.SelectedIndex < 0 {
		f.SelectedIndex = 0
	}
	f.adjustScroll()
}

// PageDown moves selection down by a page
func (f *FileReferenceState) PageDown() {
	f.SelectedIndex += f.MaxVisible
	if f.SelectedIndex >= len(f.FilteredList) {
		f.SelectedIndex = len(f.FilteredList) - 1
	}
	if f.SelectedIndex < 0 {
		f.SelectedIndex = 0
	}
	f.adjustScroll()
}

// SetFilter updates the filter pattern
func (f *FileReferenceState) SetFilter(pattern string) {
	f.FilterPattern = pattern
	f.applyFilter()
	f.SelectedIndex = 0
	f.ScrollOffset = 0
}

// ToggleHidden toggles showing hidden files
func (f *FileReferenceState) ToggleHidden() {
	f.ShowHidden = !f.ShowHidden
	f.applyFilter()
}

// applyFilter applies the current filter to the file list
func (f *FileReferenceState) applyFilter() {
	f.FilteredList = []FileInfo{}

	for _, file := range f.FileList {
		// Skip hidden files if not showing them
		if !f.ShowHidden && strings.HasPrefix(file.Name, ".") {
			continue
		}

		// Apply filter pattern if set
		if f.FilterPattern != "" {
			pattern := strings.ToLower(f.FilterPattern)
			name := strings.ToLower(file.Name)
			if !strings.Contains(name, pattern) {
				continue
			}
		}

		f.FilteredList = append(f.FilteredList, file)
	}
}

// adjustScroll adjusts the scroll offset to keep selection visible
func (f *FileReferenceState) adjustScroll() {
	// Scroll up if selection is above visible area
	if f.SelectedIndex < f.ScrollOffset {
		f.ScrollOffset = f.SelectedIndex
	}

	// Scroll down if selection is below visible area
	if f.SelectedIndex >= f.ScrollOffset+f.MaxVisible {
		f.ScrollOffset = f.SelectedIndex - f.MaxVisible + 1
	}
}

// GetVisibleFiles returns the currently visible files
func (f *FileReferenceState) GetVisibleFiles() []FileInfo {
	if len(f.FilteredList) == 0 {
		return []FileInfo{}
	}

	start := f.ScrollOffset
	end := f.ScrollOffset + f.MaxVisible

	if end > len(f.FilteredList) {
		end = len(f.FilteredList)
	}

	return f.FilteredList[start:end]
}

// IsSelected checks if a file at index is selected
func (f *FileReferenceState) IsSelected(index int) bool {
	return index == f.SelectedIndex
}

// GetBreadcrumbs returns path components for breadcrumb display
func (f *FileReferenceState) GetBreadcrumbs() []string {
	if f.CurrentPath == "" || f.CurrentPath == "." {
		return []string{"Current Directory"}
	}

	clean := filepath.Clean(f.CurrentPath)
	parts := strings.Split(clean, string(filepath.Separator))

	// Filter out empty parts
	var breadcrumbs []string
	for _, part := range parts {
		if part != "" {
			breadcrumbs = append(breadcrumbs, part)
		}
	}

	if len(breadcrumbs) == 0 {
		return []string{"Root"}
	}

	return breadcrumbs
}

// SetMaxVisible sets the maximum number of visible items
func (f *FileReferenceState) SetMaxVisible(max int) {
	f.MaxVisible = max
	f.adjustScroll()
}

// Reset clears the file reference state
func (f *FileReferenceState) Reset() {
	f.Active = false
	f.CurrentPath = ""
	f.SelectedIndex = 0
	f.ScrollOffset = 0
	f.FileList = []FileInfo{}
	f.FilteredList = []FileInfo{}
	f.FilterPattern = ""
	f.ShowHidden = false
	f.SelectedFile = ""
}

// HasFiles checks if there are any files in the filtered list
func (f *FileReferenceState) HasFiles() bool {
	return len(f.FilteredList) > 0
}

// GetFileCount returns the number of filtered files
func (f *FileReferenceState) GetFileCount() int {
	return len(f.FilteredList)
}

// GetCurrentFile returns info about the currently selected file
func (f *FileReferenceState) GetCurrentFile() *FileInfo {
	if f.SelectedIndex >= 0 && f.SelectedIndex < len(f.FilteredList) {
		return &f.FilteredList[f.SelectedIndex]
	}
	return nil
}
