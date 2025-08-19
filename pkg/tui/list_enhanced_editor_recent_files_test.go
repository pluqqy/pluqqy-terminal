package tui

import (
	"testing"
)

func TestRecentFilesTracker_AddFile(t *testing.T) {
	rft := NewRecentFilesTracker()

	// Add a file
	rft.AddFile("/path/to/file1.go")

	if len(rft.RecentFiles) != 1 {
		t.Errorf("Expected 1 recent file, got %d", len(rft.RecentFiles))
	}

	if rft.RecentFiles[0].Path != "/path/to/file1.go" {
		t.Errorf("Expected path '/path/to/file1.go', got '%s'", rft.RecentFiles[0].Path)
	}

	if rft.RecentFiles[0].Name != "file1.go" {
		t.Errorf("Expected name 'file1.go', got '%s'", rft.RecentFiles[0].Name)
	}

	if rft.RecentFiles[0].AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", rft.RecentFiles[0].AccessCount)
	}
}

func TestRecentFilesTracker_AddDuplicateFile(t *testing.T) {
	rft := NewRecentFilesTracker()

	// Add files
	rft.AddFile("/path/to/file1.go")
	rft.AddFile("/path/to/file2.go")
	rft.AddFile("/path/to/file1.go") // Duplicate

	// Should still have 2 files
	if len(rft.RecentFiles) != 2 {
		t.Errorf("Expected 2 recent files, got %d", len(rft.RecentFiles))
	}

	// file1.go should be at the front with increased access count
	if rft.RecentFiles[0].Path != "/path/to/file1.go" {
		t.Errorf("Expected file1.go at front, got '%s'", rft.RecentFiles[0].Path)
	}

	if rft.RecentFiles[0].AccessCount != 2 {
		t.Errorf("Expected access count 2 for file1.go, got %d", rft.RecentFiles[0].AccessCount)
	}

	// file2.go should be second
	if rft.RecentFiles[1].Path != "/path/to/file2.go" {
		t.Errorf("Expected file2.go at position 1, got '%s'", rft.RecentFiles[1].Path)
	}
}

func TestRecentFilesTracker_MaxFiles(t *testing.T) {
	rft := NewRecentFilesTracker()

	// Add more than max files
	for i := 1; i <= 7; i++ {
		rft.AddFile(string(rune('a'+i-1)) + ".go")
	}

	// Should only keep max files (5)
	if len(rft.RecentFiles) != 5 {
		t.Errorf("Expected 5 recent files (max), got %d", len(rft.RecentFiles))
	}

	// Most recent should be at front
	if rft.RecentFiles[0].Name != "g.go" {
		t.Errorf("Expected most recent file 'g.go' at front, got '%s'", rft.RecentFiles[0].Name)
	}

	// Oldest files should be dropped
	for _, rf := range rft.RecentFiles {
		if rf.Name == "a.go" || rf.Name == "b.go" {
			t.Errorf("Old file '%s' should have been dropped", rf.Name)
		}
	}
}

func TestRecentFilesTracker_GetFileByNumber(t *testing.T) {
	rft := NewRecentFilesTracker()

	rft.AddFile("/path/to/file1.go")
	rft.AddFile("/path/to/file2.go")
	rft.AddFile("/path/to/file3.go")

	tests := []struct {
		number   int
		expected string
		found    bool
	}{
		{1, "/path/to/file3.go", true}, // Most recent
		{2, "/path/to/file2.go", true},
		{3, "/path/to/file1.go", true},
		{0, "", false}, // Out of range
		{4, "", false}, // Out of range
		{6, "", false}, // Out of range
	}

	for _, tt := range tests {
		file, found := rft.GetFileByNumber(tt.number)
		if found != tt.found {
			t.Errorf("GetFileByNumber(%d): expected found=%v, got %v", tt.number, tt.found, found)
		}
		if found && file.Path != tt.expected {
			t.Errorf("GetFileByNumber(%d): expected path '%s', got '%s'", tt.number, tt.expected, file.Path)
		}
	}
}

func TestRecentFilesTracker_Clear(t *testing.T) {
	rft := NewRecentFilesTracker()

	rft.AddFile("/path/to/file1.go")
	rft.AddFile("/path/to/file2.go")

	rft.Clear()

	if len(rft.RecentFiles) != 0 {
		t.Errorf("Expected 0 files after Clear, got %d", len(rft.RecentFiles))
	}

	if rft.HasRecentFiles() {
		t.Error("HasRecentFiles should return false after Clear")
	}
}

func TestRecentFilesTracker_RemoveFile(t *testing.T) {
	rft := NewRecentFilesTracker()

	rft.AddFile("/path/to/file1.go")
	rft.AddFile("/path/to/file2.go")
	rft.AddFile("/path/to/file3.go")

	// Remove middle file
	rft.RemoveFile("/path/to/file2.go")

	if len(rft.RecentFiles) != 2 {
		t.Errorf("Expected 2 files after removal, got %d", len(rft.RecentFiles))
	}

	// Check remaining files
	if rft.RecentFiles[0].Path != "/path/to/file3.go" {
		t.Errorf("Expected file3.go at position 0, got '%s'", rft.RecentFiles[0].Path)
	}

	if rft.RecentFiles[1].Path != "/path/to/file1.go" {
		t.Errorf("Expected file1.go at position 1, got '%s'", rft.RecentFiles[1].Path)
	}

	// Try removing non-existent file
	rft.RemoveFile("/path/to/nonexistent.go")
	if len(rft.RecentFiles) != 2 {
		t.Error("Removing non-existent file should not change the list")
	}
}

func TestRecentFilesTracker_FormatRecentFilesList(t *testing.T) {
	rft := NewRecentFilesTracker()

	// Empty list
	formatted := rft.FormatRecentFilesList()
	if formatted != nil {
		t.Error("FormatRecentFilesList should return nil for empty list")
	}

	// Add files
	rft.AddFile("/path/to/file1.go")
	rft.AddFile("file2.go")
	rft.AddFile("/very/long/path/that/is/too/long/to/display/fully/file3.go")

	formatted = rft.FormatRecentFilesList()
	if len(formatted) != 3 {
		t.Errorf("Expected 3 formatted entries, got %d", len(formatted))
	}
}

func TestRecentFilesTracker_HasRecentFiles(t *testing.T) {
	rft := NewRecentFilesTracker()

	if rft.HasRecentFiles() {
		t.Error("HasRecentFiles should return false for new tracker")
	}

	rft.AddFile("test.go")

	if !rft.HasRecentFiles() {
		t.Error("HasRecentFiles should return true after adding a file")
	}
}
