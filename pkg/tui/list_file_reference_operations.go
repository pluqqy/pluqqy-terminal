package tui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadDirectory loads files from a directory into FileInfo structures
func LoadDirectory(path string) ([]FileInfo, error) {
	// Validate path
	if err := ValidateDirectoryPath(path); err != nil {
		return nil, err
	}

	// Read directory contents
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Convert to FileInfo structures
	var files []FileInfo
	for _, entry := range entries {
		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    entry.Size(),
			ModTime: entry.ModTime().Format("2006-01-02 15:04"),
		}
		files = append(files, fileInfo)
	}

	// Sort: directories first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

// FilterFiles filters files based on a pattern
func FilterFiles(files []FileInfo, pattern string, showHidden bool) []FileInfo {
	if pattern == "" && showHidden {
		return files
	}

	var filtered []FileInfo
	lowerPattern := strings.ToLower(pattern)

	for _, file := range files {
		// Skip hidden files if not showing them
		if !showHidden && strings.HasPrefix(file.Name, ".") {
			continue
		}

		// Apply pattern filter
		if pattern != "" {
			lowerName := strings.ToLower(file.Name)
			if !strings.Contains(lowerName, lowerPattern) {
				continue
			}
		}

		filtered = append(filtered, file)
	}

	return filtered
}

// GetRelativePath returns a relative path from base to target
func GetRelativePath(basePath, fullPath string) (string, error) {
	// Clean both paths
	basePath = filepath.Clean(basePath)
	fullPath = filepath.Clean(fullPath)

	// Get relative path
	relPath, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Ensure forward slashes for consistency
	relPath = filepath.ToSlash(relPath)

	return relPath, nil
}

// ValidateDirectoryPath validates that a path is a valid directory
func ValidateDirectoryPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// ValidateFilePath validates that a file path is safe
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Check for path traversal attempts
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check if file exists (optional - may want to allow non-existent files)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Allow non-existent files for now
			return nil
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	return nil
}

// GetFileExtension returns the file extension
func GetFileExtension(path string) string {
	// For hidden files without extensions (like .gitignore), return empty
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return ""
	}

	ext := filepath.Ext(path)
	if ext != "" {
		return strings.ToLower(ext)
	}
	return ""
}

// HandleFileReferenceInput processes input for file reference browser
func HandleFileReferenceInput(state *FileReferenceState, msg tea.KeyMsg) (bool, tea.Cmd) {
	if !state.IsActive() {
		return false, nil
	}

	switch msg.String() {
	case "up", "k":
		state.MoveUp()
		return true, nil

	case "down", "j":
		state.MoveDown()
		return true, nil

	case "pgup":
		state.PageUp()
		return true, nil

	case "pgdown":
		state.PageDown()
		return true, nil

	case "enter":
		// Select current file/directory
		current := state.GetCurrentFile()
		if current == nil {
			return true, nil
		}

		if current.IsDir {
			// Navigate into directory
			return true, loadDirectoryCmd(state, current.Path)
		}

		// Select file
		state.Stop()
		return true, nil

	case "tab":
		// Toggle hidden files
		state.ToggleHidden()
		return true, nil

	case "esc":
		// Cancel selection
		state.Stop()
		return true, nil

	case "/":
		// Start search/filter mode (would need additional state)
		return true, nil

	case "backspace":
		// Go to parent directory
		parent := filepath.Dir(state.CurrentPath)
		if parent != state.CurrentPath {
			return true, loadDirectoryCmd(state, parent)
		}
		return true, nil

	default:
		// Handle filter input (simplified - would need text input state)
		if msg.Type == tea.KeyRunes {
			pattern := state.FilterPattern + string(msg.Runes)
			state.SetFilter(pattern)
			return true, nil
		}
	}

	return false, nil
}

// loadDirectoryCmd creates a command to load a directory
func loadDirectoryCmd(state *FileReferenceState, path string) tea.Cmd {
	return func() tea.Msg {
		files, err := LoadDirectory(path)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to load directory: %v", err))
		}

		state.SetCurrentPath(path)
		state.SetFileList(files)
		return nil
	}
}

// InitializeFileReference initializes the file reference browser
func InitializeFileReference(state *FileReferenceState, initialPath string) tea.Cmd {
	if initialPath == "" {
		initialPath, _ = os.Getwd()
	}

	state.Start(initialPath)
	return loadDirectoryCmd(state, initialPath)
}

// FormatFileSize formats a file size in human-readable format
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GetFileType returns a descriptive file type string
func GetFileType(filename string) string {
	ext := GetFileExtension(filename)

	switch ext {
	case ".md":
		return "Markdown"
	case ".yaml", ".yml":
		return "YAML"
	case ".json":
		return "JSON"
	case ".txt":
		return "Text"
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".sh":
		return "Shell"
	case "":
		return "File"
	default:
		return strings.ToUpper(strings.TrimPrefix(ext, "."))
	}
}

// SortFiles sorts files with directories first, then alphabetically
func SortFiles(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		// Directories come first
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		// Then sort alphabetically (case-insensitive)
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
}

// GetParentDirectory returns the parent directory path
func GetParentDirectory(path string) string {
	parent := filepath.Dir(path)
	if parent == path {
		return ""
	}
	return parent
}

// IsTextFile checks if a file is likely a text file based on extension
func IsTextFile(filename string) bool {
	textExts := []string{
		".txt", ".md", ".markdown", ".yaml", ".yml", ".json",
		".go", ".js", ".ts", ".py", ".rb", ".java", ".c", ".cpp",
		".h", ".hpp", ".cs", ".php", ".html", ".css", ".xml",
		".sh", ".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd",
		".conf", ".cfg", ".ini", ".toml", ".env", ".gitignore",
	}

	ext := GetFileExtension(filename)
	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	// Check for files without extensions that are commonly text
	base := filepath.Base(filename)
	textFiles := []string{
		"README", "LICENSE", "CHANGELOG", "Makefile", "Dockerfile",
		"Vagrantfile", "Gemfile", "Rakefile", "Procfile",
		".gitignore", ".dockerignore", ".eslintrc", ".prettierrc",
		".editorconfig", ".npmignore", ".gitattributes",
	}

	for _, textFile := range textFiles {
		if base == textFile {
			return true
		}
	}

	return false
}
