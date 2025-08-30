package tui

import (
	"fmt"
	"os"
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
	case "ctrl+z":
		// Undo last action
		return true, undoLastAction(state)

	case "ctrl+k":
		// Clear all content
		return true, clearAllContent(state)

	case "ctrl+shift+v", "ctrl+l":
		// Clean current content (remove TUI borders, line numbers, etc)
		return true, cleanCurrentContent(state)

	case "ctrl+s":
		// Save component
		return true, saveEnhancedComponent(state)

	case "ctrl+x":
		// Show persistent status message
		state.StatusManager.SetPersistentMessage("Editing in external editor - save your changes and close the editor window/tab to return here and continue", StatusTypeInfo)
		// Open in external editor
		return true, openInExternalEditor(state)

	case "pgup":
		// Page up navigation
		return true, handlePageUp(state, false)

	case "pgdown":
		// Page down navigation
		return true, handlePageDown(state, false)

	case "shift+pgup":
		// Page up with text selection
		return true, handlePageUp(state, true)

	case "shift+pgdown":
		// Page down with text selection
		return true, handlePageDown(state, true)

	case "esc":
		// For new components, ESC should go back to naming step, not exit completely
		if state.IsNewComponent {
			// Even with unsaved changes, allow going back for new components
			// The user can navigate back to continue editing
			// Save the current content from textarea before deactivating
			state.Content = state.Textarea.Value()
			state.Active = false  // Just deactivate, don't reset
			// Don't clear IsNewComponent flag - it should remain true
			return true, nil
		}
		
		// For existing components, handle exit with unsaved changes check
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

	case "enter":
		// Explicitly handle Enter key - pass it directly to the textarea
		// The textarea expects the message as-is, not reconstructed
		cmd := state.UpdateTextarea(msg)
		return true, cmd
		
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

	case "1", "2", "3", "4", "5":
		// Quick select recent file by number
		num := int(msg.String()[0] - '0')
		if file, ok := state.RecentFiles.GetFileByNumber(num); ok {
			// Insert the file reference
			fileRef := ProcessFileSelection(file.Path)
			insertCmd := InsertFileReference(state, fileRef)
			state.RecentFiles.AddFile(file.Path) // Update access time
			state.StopFilePicker()
			return true, insertCmd
		}
		// Fall through if no recent file at this number
		fallthrough

	default:
		// Update file picker
		cmd := state.UpdateFilePicker(msg)

		// Check if a file was selected
		if didSelect, selected := state.FilePicker.DidSelectFile(msg); didSelect {
			// Insert the file reference
			fileRef := ProcessFileSelection(selected)
			insertCmd := InsertFileReference(state, fileRef)
			state.RecentFiles.AddFile(selected) // Add to recent files
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

// undoLastAction undoes the last action
func undoLastAction(state *EnhancedEditorState) tea.Cmd {
	if state.Undo() {
		state.ActionFeedback.RecordAction("✓ Undone")
		return state.StatusManager.ShowSuccess("Undone")
	}
	return state.StatusManager.ShowInfo("Nothing to undo")
}

// clearAllContent clears all content from the editor
func clearAllContent(state *EnhancedEditorState) tea.Cmd {
	// Save current state for undo
	if state.Content != "" {
		state.SaveUndoState("Clear all")
	}

	state.SetContent("")
	state.Textarea.Reset()
	state.Textarea.Focus()
	state.ActionFeedback.RecordAction("✓ Cleared - ready for paste")
	state.UpdateStats()

	// Return status feedback
	return ShowClearedStatus()
}

// cleanCurrentContent cleans the current content in the editor
func cleanCurrentContent(state *EnhancedEditorState) tea.Cmd {
	// Get current content
	currentContent := state.Textarea.Value()
	if currentContent == "" {
		return ShowNothingToPasteStatus()
	}

	// Clean the content
	cleanedContent := state.PasteHelper.CleanPastedContent(currentContent)

	// Check if anything was cleaned
	if cleanedContent != currentContent {
		// Save for undo
		state.SaveUndoState("Clean content")

		// Update the content
		state.SetContent(cleanedContent)
		state.Textarea.SetValue(cleanedContent)
		state.Textarea.Focus()
		state.UpdateStats()

		// Show feedback
		lineCount := CountLines(cleanedContent)
		state.ActionFeedback.RecordAction(fmt.Sprintf("✓ Cleaned %d lines", lineCount))
		return state.StatusManager.ShowSuccess(fmt.Sprintf("Cleaned %d lines", lineCount))
	} else {
		// Nothing to clean
		state.ActionFeedback.RecordAction("Content already clean")
		return state.StatusManager.ShowInfo("Content already clean")
	}
}

// saveEnhancedComponent saves the component to disk
func saveEnhancedComponent(state *EnhancedEditorState) tea.Cmd {
	// Get content from textarea and clean it BEFORE the async operation
	content := state.Textarea.Value()
	cleanedContent := state.PasteHelper.CleanForSave(content)
	
	// Immediately update state to prevent "unsaved changes" false positive
	// This happens synchronously before the async save operation
	state.OriginalContent = cleanedContent
	state.Content = cleanedContent
	state.Textarea.SetValue(cleanedContent) // Sync textarea with cleaned content
	state.UnsavedChanges = false // Mark as saved immediately
	
	return func() tea.Msg {
		// Validate content
		if err := ValidateComponentContent(cleanedContent); err != nil {
			// Revert state on validation failure
			state.UnsavedChanges = true
			return StatusMsg(fmt.Sprintf("× Validation failed: %v", err))
		}

		// Write component - always use WriteComponentWithNameAndTags to preserve both name and tags
		err := files.WriteComponentWithNameAndTags(state.ComponentPath, cleanedContent, state.ComponentName, state.ComponentTags)

		if err != nil {
			// Revert state on write failure
			state.UnsavedChanges = true
			return StatusMsg(fmt.Sprintf("× Failed to save: %v", err))
		}
		
		// After first save, it's no longer a new component
		if state.IsNewComponent {
			state.IsNewComponent = false
		}

		// Return a status message without closing the editor
		savedName := state.ComponentName
		return StatusMsg(fmt.Sprintf("✓ Saved: %s", savedName))
	}
}

// openInExternalEditor opens the component in the user's external editor
func openInExternalEditor(state *EnhancedEditorState) tea.Cmd {
	return func() tea.Msg {
		// For new components (creation mode), save with name in frontmatter
		// For existing components, save normally
		if state.IsNewComponent {
			content := state.Textarea.Value()
			// Use WriteComponentWithNameAndTags for new components
			err := files.WriteComponentWithNameAndTags(state.ComponentPath, content, state.ComponentName, state.ComponentTags)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to save before external edit: %v", err))
			}
			// After first save, it's no longer a new component
			state.IsNewComponent = false
		} else if state.HasUnsavedChanges() {
			// For existing components, save with name and tags to preserve both
			content := state.Textarea.Value()
			err := files.WriteComponentWithNameAndTags(state.ComponentPath, content, state.ComponentName, state.ComponentTags)
			if err != nil {
				return StatusMsg(fmt.Sprintf("× Failed to save before external edit: %v", err))
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			return StatusMsg("Error: $EDITOR environment variable not set")
		}

		// Construct full path
		fullPath := filepath.Join(files.PluqqyDir, state.ComponentPath)

		// Create command with proper argument parsing for editors with flags
		cmd := createEditorCommand(editor, fullPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to open editor: %v", err))
		}

		// Reload the content from the file after external editing
		content, err := files.ReadComponent(state.ComponentPath)
		if err != nil {
			return StatusMsg(fmt.Sprintf("Failed to reload content after editing: %v", err))
		}

		// Update the textarea with the new content
		state.Textarea.SetValue(content.Content)

		// Update the original content to match what's now in the file
		state.OriginalContent = content.Content
		state.Content = content.Content

		// Update name and tags if they were changed in the external editor
		if content.Name != "" {
			state.ComponentName = content.Name
		}
		if content.Tags != nil {
			state.ComponentTags = content.Tags
		}

		// Clear unsaved changes flag since we just loaded from disk
		state.UnsavedChanges = false

		// Clear the persistent message about external editor
		state.StatusManager.ClearPersistentMessage()

		// Show temporary success message
		state.StatusManager.ShowSuccess("Content reloaded from external editor")

		// Return a status message
		return StatusMsg(fmt.Sprintf("✓ Reloaded: %s", filepath.Base(state.ComponentPath)))
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

// handlePageUp handles page up navigation with optional text selection
func handlePageUp(state *EnhancedEditorState, withSelection bool) tea.Cmd {
	// Get the visible height of the textarea
	visibleLines := state.Textarea.Height()
	if visibleLines <= 0 {
		visibleLines = 20 // Default if not set
	}

	// Leave a few lines of overlap for context
	scrollAmount := visibleLines - 3
	if scrollAmount < 1 {
		scrollAmount = 1
	}

	// Move cursor up by scroll amount
	for i := 0; i < scrollAmount; i++ {
		if withSelection {
			// Use shift+up to select text while moving
			state.Textarea, _ = state.Textarea.Update(tea.KeyMsg{
				Type: tea.KeyShiftUp,
			})
		} else {
			// Regular cursor movement
			state.Textarea.CursorUp()
		}
	}

	// Update stats and provide feedback
	state.UpdateStats()
	if withSelection {
		state.ActionFeedback.RecordAction("Page Up (selecting)")
	} else {
		state.ActionFeedback.RecordAction("Page Up")
	}

	return nil
}

// handlePageDown handles page down navigation with optional text selection
func handlePageDown(state *EnhancedEditorState, withSelection bool) tea.Cmd {
	// Get the visible height of the textarea
	visibleLines := state.Textarea.Height()
	if visibleLines <= 0 {
		visibleLines = 20 // Default if not set
	}

	// Leave a few lines of overlap for context
	scrollAmount := visibleLines - 3
	if scrollAmount < 1 {
		scrollAmount = 1
	}

	// Move cursor down by scroll amount
	for i := 0; i < scrollAmount; i++ {
		if withSelection {
			// Use shift+down to select text while moving
			state.Textarea, _ = state.Textarea.Update(tea.KeyMsg{
				Type: tea.KeyShiftDown,
			})
		} else {
			// Regular cursor movement
			state.Textarea.CursorDown()
		}
	}

	// Update stats and provide feedback
	state.UpdateStats()
	if withSelection {
		state.ActionFeedback.RecordAction("Page Down (selecting)")
	} else {
		state.ActionFeedback.RecordAction("Page Down")
	}

	return nil
}
