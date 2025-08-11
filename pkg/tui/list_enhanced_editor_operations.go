package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
)

// HandleEnhancedEditorInput processes input for the enhanced editor
func HandleEnhancedEditorInput(state *EnhancedEditorState, msg tea.KeyMsg, width int) (bool, tea.Cmd) {
	if !state.IsActive() {
		return false, nil
	}

	// Handle exit confirmation if active
	if state.IsExitConfirmActive() {
		cmd := state.ExitConfirm.Update(msg)
		if cmd != nil {
			return true, cmd
		}
		return true, nil
	}

	// Handle file picker mode
	if state.IsFilePicking() {
		return handleFilePickerInput(state, msg)
	}

	// Handle normal editor mode
	return handleNormalEditorInput(state, msg, width)
}

// handleNormalEditorInput handles input in normal editing mode
func handleNormalEditorInput(state *EnhancedEditorState, msg tea.KeyMsg, width int) (bool, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		// Save component
		return true, saveEnhancedComponent(state)
	
	case "esc":
		// Handle exit with unsaved changes check
		if state.HasUnsavedChanges() {
			state.ShowExitConfirmation(width,
				func() tea.Cmd {
					// Exit without saving
					state.Reset()
					return nil
				},
				func() tea.Cmd {
					// Cancel exit
					state.HideExitConfirmation()
					return nil
				},
			)
			return true, nil
		}
		// No changes, exit immediately
		state.Reset()
		return true, nil
	
	case "@":
		// Check if the previous character is a backslash (escape)
		content := state.Textarea.Value()
		if len(content) > 0 && content[len(content)-1] == '\\' {
			// User is escaping the @, remove the backslash and add @ as literal text
			state.SetContent(content[:len(content)-1] + "@")
			return true, nil
		}
		
		// Not escaped, trigger file picker
		// First, add the @ to the textarea so user sees it
		cmd := state.UpdateTextarea(msg)
		// Then switch to file picker mode
		state.StartFilePicker()
		// Initialize file picker with project root directory
		initCmd := InitializeEnhancedFilePicker(state)
		// Return both commands
		return true, tea.Batch(cmd, initCmd)
	
	default:
		// Delegate to textarea for normal text editing
		cmd := state.UpdateTextarea(msg)
		return true, cmd
	}
}

// handleFilePickerInput handles input in file picker mode
func handleFilePickerInput(state *EnhancedEditorState, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel file picking
		state.StopFilePicker()
		// Remove the @ character that triggered the picker
		content := state.Content
		if len(content) > 0 && content[len(content)-1] == '@' {
			state.SetContent(content[:len(content)-1])
		}
		return true, nil
	
	default:
		// Update file picker
		cmd := state.UpdateFilePicker(msg)
		
		// Check if a file was selected
		if didSelect, selected := state.FilePicker.DidSelectFile(msg); didSelect {
			// Insert the file reference
			fileRef := ProcessFileSelection(selected)
			insertCmd := InsertFileReference(state, fileRef)
			state.StopFilePicker()
			return true, insertCmd
		}
		
		return true, cmd
	}
}

// DetectAtTrigger checks if @ was just typed and should trigger file picker
func DetectAtTrigger(state *EnhancedEditorState) bool {
	content := state.Textarea.Value()
	
	// Check if the last character typed is @
	if len(content) > 0 && content[len(content)-1] == '@' {
		return true
	}
	
	return false
}

// InitializeEnhancedFilePicker sets up the file picker
func InitializeEnhancedFilePicker(state *EnhancedEditorState) tea.Cmd {
	// Start from the current working directory (project root)
	dir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		dir = "."
	}
	
	// Update the existing file picker's settings without replacing it
	state.FilePicker.CurrentDirectory = dir
	state.FilePicker.AllowedTypes = []string{} // Empty means all files are allowed
	state.FilePicker.ShowHidden = false
	state.FilePicker.DirAllowed = true
	state.FilePicker.FileAllowed = true
	state.FilePicker.AutoHeight = false
	state.FilePicker.Height = 20 // Set a reasonable height
	
	// Now initialize it to read the directory
	return state.FilePicker.Init()
}

// ProcessFileSelection formats the selected file path as a reference
func ProcessFileSelection(path string) string {
	// Clean the path
	cleanPath := filepath.Clean(path)
	
	// Format as @reference
	return fmt.Sprintf("@%s", cleanPath)
}

// InsertFileReference inserts a file reference at the current position
func InsertFileReference(state *EnhancedEditorState, reference string) tea.Cmd {
	content := state.Content
	
	// Remove the @ that triggered the picker (it's at the end)
	if len(content) > 0 && content[len(content)-1] == '@' {
		content = content[:len(content)-1]
	}
	
	// Insert the reference
	newContent := content + reference
	state.SetContent(newContent)
	
	return nil
}

// saveEnhancedComponent saves the component to disk
func saveEnhancedComponent(state *EnhancedEditorState) tea.Cmd {
	return func() tea.Msg {
		// Get content from textarea
		content := state.Textarea.Value()
		
		// Validate content
		if err := ValidateComponentContent(content); err != nil {
			return StatusMsg(fmt.Sprintf("× Validation failed: %v", err))
		}
		
		// Write component with tags preserved
		err := files.WriteComponentWithTags(state.ComponentPath, content, state.ComponentTags)
		if err != nil {
			return StatusMsg(fmt.Sprintf("× Failed to save: %v", err))
		}
		
		// Clear editing state
		savedName := state.ComponentName
		state.Reset()
		
		// Return a reload message to refresh data
		return ReloadMsg{
			Message: fmt.Sprintf("✓ Saved: %s", savedName),
		}
	}
}

// openInExternalEditor opens the component in the user's external editor
func openInExternalEditor(state *EnhancedEditorState) tea.Cmd {
	return func() tea.Msg {
		// First save any unsaved changes
		if state.HasUnsavedChanges() {
			content := state.Textarea.Value()
			err := files.WriteComponentWithTags(state.ComponentPath, content, state.ComponentTags)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to save before external edit: %v", err))
			}
		}
		
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set")
		}
		
		// Validate editor path to prevent command injection
		if strings.ContainsAny(editor, "&|;<>()$`\\\"'") {
			return StatusMsg("Invalid EDITOR value: contains shell metacharacters")
		}
		
		// Construct full path
		fullPath := filepath.Join(files.PluqqyDir, state.ComponentPath)
		
		// Execute editor
		cmd := exec.Command(editor, fullPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}
		
		// Return a reload message to refresh data
		return ReloadMsg{
			Message: fmt.Sprintf("Edited: %s", filepath.Base(state.ComponentPath)),
		}
	}
}

// ValidateComponentContent validates the content before saving
func ValidateComponentContent(content string) error {
	// Check for empty content
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("component cannot be empty")
	}
	
	// Check for basic YAML validity if it looks like frontmatter
	if strings.HasPrefix(strings.TrimSpace(content), "---") {
		// Basic check for frontmatter structure - needs closing ---
		parts := strings.SplitN(content, "---", 3)
		if len(parts) < 3 {
			return fmt.Errorf("invalid frontmatter structure")
		}
	}
	
	// Check file references are valid
	refs := ExtractFileReferences(content)
	for _, ref := range refs {
		if err := ValidateFileReference(ref); err != nil {
			return fmt.Errorf("invalid file reference %s: %v", ref, err)
		}
	}
	
	return nil
}

// ExtractFileReferences finds all @file references in content
func ExtractFileReferences(content string) []string {
	var refs []string
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		// Find all @ references in the line
		for i := 0; i < len(line); i++ {
			if line[i] == '@' && i+1 < len(line) {
				// Find the end of the reference (space or newline)
				end := i + 1
				for end < len(line) && line[end] != ' ' && line[end] != '\t' {
					end++
				}
				if end > i+1 {
					refs = append(refs, line[i:end])
				}
			}
		}
	}
	
	return refs
}

// ValidateFileReference checks if a file reference is valid
func ValidateFileReference(ref string) error {
	if !strings.HasPrefix(ref, "@") {
		return fmt.Errorf("reference must start with @")
	}
	
	path := strings.TrimPrefix(ref, "@")
	if path == "" {
		return fmt.Errorf("empty path")
	}
	
	// Check for dangerous path elements
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	
	return nil
}

// ParseFileReference extracts the path from a file reference
func ParseFileReference(ref string) (string, error) {
	if !strings.HasPrefix(ref, "@") {
		return "", fmt.Errorf("invalid reference format")
	}
	
	path := strings.TrimPrefix(ref, "@")
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	
	return path, nil
}

// InsertTextAtCursor inserts text at a specific position
func InsertTextAtCursor(content string, text string, pos int) string {
	if pos < 0 {
		pos = 0
	}
	if pos > len(content) {
		pos = len(content)
	}
	
	before := content[:pos]
	after := content[pos:]
	return before + text + after
}