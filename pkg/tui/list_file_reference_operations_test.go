package tui

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterFiles(t *testing.T) {
	tests := []struct {
		name       string
		files      []FileInfo
		pattern    string
		showHidden bool
		expected   int
	}{
		{
			name: "filter by pattern",
			files: []FileInfo{
				{Name: "test.md", IsDir: false},
				{Name: "example.txt", IsDir: false},
				{Name: "test.yaml", IsDir: false},
			},
			pattern:    "test",
			showHidden: true,
			expected:   2,
		},
		{
			name: "filter hidden files",
			files: []FileInfo{
				{Name: "visible.md", IsDir: false},
				{Name: ".hidden", IsDir: false},
				{Name: ".gitignore", IsDir: false},
			},
			pattern:    "",
			showHidden: false,
			expected:   1,
		},
		{
			name: "show all with hidden",
			files: []FileInfo{
				{Name: "file1.md", IsDir: false},
				{Name: ".hidden", IsDir: false},
				{Name: "file2.txt", IsDir: false},
			},
			pattern:    "",
			showHidden: true,
			expected:   3,
		},
		{
			name: "case insensitive pattern",
			files: []FileInfo{
				{Name: "TEST.md", IsDir: false},
				{Name: "test.txt", IsDir: false},
				{Name: "TeSt.yaml", IsDir: false},
			},
			pattern:    "test",
			showHidden: true,
			expected:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterFiles(tt.files, tt.pattern, tt.showHidden)
			if len(result) != tt.expected {
				t.Errorf("Expected %d files, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		fullPath string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple relative path",
			basePath: "/home/user",
			fullPath: "/home/user/documents/file.txt",
			expected: "documents/file.txt",
			wantErr:  false,
		},
		{
			name:     "same path",
			basePath: "/home/user",
			fullPath: "/home/user",
			expected: ".",
			wantErr:  false,
		},
		{
			name:     "parent directory",
			basePath: "/home/user/documents",
			fullPath: "/home/user",
			expected: "..",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRelativePath(tt.basePath, tt.fullPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRelativePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateDirectoryPath(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := ioutil.TempDir("", "test_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temp file for testing
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := ioutil.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid directory",
			path:    tempDir,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name:    "non-existent path",
			path:    "/non/existent/path",
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name:    "file not directory",
			path:    tempFile,
			wantErr: true,
			errMsg:  "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid path",
			path:    "/home/user/file.txt",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal not allowed",
		},
		{
			name:    "non-existent file allowed",
			path:    "/path/to/new/file.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "markdown file",
			path:     "document.md",
			expected: ".md",
		},
		{
			name:     "yaml file",
			path:     "config.yaml",
			expected: ".yaml",
		},
		{
			name:     "no extension",
			path:     "README",
			expected: "",
		},
		{
			name:     "multiple dots",
			path:     "file.tar.gz",
			expected: ".gz",
		},
		{
			name:     "hidden file",
			path:     ".gitignore",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileExtension(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			size:     1536,
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			size:     1572864,
			expected: "1.5 MB",
		},
		{
			name:     "zero",
			size:     0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "markdown",
			filename: "README.md",
			expected: "Markdown",
		},
		{
			name:     "yaml",
			filename: "config.yaml",
			expected: "YAML",
		},
		{
			name:     "yml",
			filename: "config.yml",
			expected: "YAML",
		},
		{
			name:     "json",
			filename: "data.json",
			expected: "JSON",
		},
		{
			name:     "go",
			filename: "main.go",
			expected: "Go",
		},
		{
			name:     "no extension",
			filename: "Makefile",
			expected: "File",
		},
		{
			name:     "unknown extension",
			filename: "file.xyz",
			expected: "XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileType(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSortFiles(t *testing.T) {
	files := []FileInfo{
		{Name: "zebra.txt", IsDir: false},
		{Name: "alpha", IsDir: true},
		{Name: "beta.md", IsDir: false},
		{Name: "delta", IsDir: true},
		{Name: "charlie.yaml", IsDir: false},
	}

	SortFiles(files)

	// Check directories come first
	if !files[0].IsDir || !files[1].IsDir {
		t.Error("Expected directories to come first")
	}

	// Check alphabetical order within directories
	if files[0].Name != "alpha" || files[1].Name != "delta" {
		t.Error("Expected directories to be sorted alphabetically")
	}

	// Check alphabetical order within files
	if files[2].Name != "beta.md" || files[3].Name != "charlie.yaml" || files[4].Name != "zebra.txt" {
		t.Error("Expected files to be sorted alphabetically")
	}
}

func TestGetParentDirectory(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "normal path",
			path:     "/home/user/documents",
			expected: "/home/user",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "",
		},
		{
			name:     "relative path",
			path:     "documents/files",
			expected: "documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParentDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "markdown file",
			filename: "README.md",
			expected: true,
		},
		{
			name:     "go file",
			filename: "main.go",
			expected: true,
		},
		{
			name:     "image file",
			filename: "photo.jpg",
			expected: false,
		},
		{
			name:     "README without extension",
			filename: "README",
			expected: true,
		},
		{
			name:     "Makefile",
			filename: "Makefile",
			expected: true,
		},
		{
			name:     "binary file",
			filename: "program.exe",
			expected: false,
		},
		{
			name:     "config file",
			filename: ".gitignore",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTextFile(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function for string contains check
func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}
