package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// StatusFeedback represents a temporary status message
type StatusFeedback struct {
	Message   string
	Icon      string
	ShowUntil time.Time
	Type      StatusType
}

// StatusType represents the type of status message
type StatusType int

const (
	StatusTypeSuccess StatusType = iota
	StatusTypeWarning
	StatusTypeError
	StatusTypeInfo
)

// StatusManager manages temporary status messages
type StatusManager struct {
	CurrentStatus     *StatusFeedback
	DefaultDuration   time.Duration
	PersistentMessage string
	PersistentType    StatusType
}

// NewStatusManager creates a new status manager
func NewStatusManager() *StatusManager {
	return &StatusManager{
		DefaultDuration: 2 * time.Second,
	}
}

// ShowFeedback displays a status message with an icon
func (sm *StatusManager) ShowFeedback(icon, message string, statusType StatusType) tea.Cmd {
	sm.CurrentStatus = &StatusFeedback{
		Message:   message,
		Icon:      icon,
		ShowUntil: time.Now().Add(sm.DefaultDuration),
		Type:      statusType,
	}

	// Return a command that will clear the status after duration
	return tea.Tick(sm.DefaultDuration, func(time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

// ShowSuccess shows a success message
func (sm *StatusManager) ShowSuccess(message string) tea.Cmd {
	return sm.ShowFeedback("✓", message, StatusTypeSuccess)
}

// ShowWarning shows a warning message
func (sm *StatusManager) ShowWarning(message string) tea.Cmd {
	return sm.ShowFeedback("⚠️", message, StatusTypeWarning)
}

// ShowError shows an error message
func (sm *StatusManager) ShowError(message string) tea.Cmd {
	return sm.ShowFeedback("×", message, StatusTypeError)
}

// ShowInfo shows an info message
func (sm *StatusManager) ShowInfo(message string) tea.Cmd {
	return sm.ShowFeedback("ℹ", message, StatusTypeInfo)
}

// SetPersistentMessage sets a message that persists until cleared
func (sm *StatusManager) SetPersistentMessage(message string, statusType StatusType) {
	sm.PersistentMessage = message
	sm.PersistentType = statusType
}

// ClearPersistentMessage clears the persistent message
func (sm *StatusManager) ClearPersistentMessage() {
	sm.PersistentMessage = ""
}

// Clear removes the current status
func (sm *StatusManager) Clear() {
	sm.CurrentStatus = nil
}

// IsActive checks if a status is currently showing
func (sm *StatusManager) IsActive() bool {
	if sm.CurrentStatus == nil {
		return false
	}

	// Check if status has expired
	if time.Now().After(sm.CurrentStatus.ShowUntil) {
		sm.CurrentStatus = nil
		return false
	}

	return true
}

// GetStatus returns the current status message if active
func (sm *StatusManager) GetStatus() (string, bool) {
	// First check for active temporary status
	if sm.IsActive() {
		return fmt.Sprintf("%s %s", sm.CurrentStatus.Icon, sm.CurrentStatus.Message), true
	}

	// If no temporary status, check for persistent message
	if sm.PersistentMessage != "" {
		icon := "ℹ"
		switch sm.PersistentType {
		case StatusTypeSuccess:
			icon = "✓"
		case StatusTypeWarning:
			icon = "⚠"
		case StatusTypeError:
			icon = "×"
		case StatusTypeInfo:
			icon = "ℹ"
		}
		return fmt.Sprintf("%s %s", icon, sm.PersistentMessage), true
	}

	return "", false
}

// ClearStatusMsg is sent to clear the status
type ClearStatusMsg struct{}

// Common status messages for the enhanced editor
func ShowPastedStatus(lines int, cleaned bool) tea.Cmd {
	message := fmt.Sprintf("Pasted %d lines", lines)
	if cleaned {
		message += " (cleaned formatting)"
	}

	sm := NewStatusManager()
	return sm.ShowSuccess(message)
}

func ShowClearedStatus() tea.Cmd {
	sm := NewStatusManager()
	return sm.ShowSuccess("Cleared - ready for paste")
}

func ShowSavedStatus(name string) tea.Cmd {
	sm := NewStatusManager()
	return sm.ShowSuccess(fmt.Sprintf("Saved: %s", name))
}

func ShowNothingToPasteStatus() tea.Cmd {
	sm := NewStatusManager()
	return sm.ShowWarning("Nothing to paste")
}

func ShowValidationErrorStatus(err error) tea.Cmd {
	sm := NewStatusManager()
	return sm.ShowError(fmt.Sprintf("Validation failed: %v", err))
}

// ClipboardStatus represents the clipboard state for the status bar
type ClipboardStatus struct {
	HasContent bool
	LineCount  int
	WillClean  bool
}

// GetClipboardStatusString returns a formatted clipboard status
func GetClipboardStatusString(cs *ClipboardStatus) string {
	if !cs.HasContent || cs.LineCount == 0 {
		return ""
	}

	status := fmt.Sprintf("📋 %d lines ready", cs.LineCount)
	if cs.WillClean {
		status += " (will clean)"
	}

	return status
}

// EditorActionFeedback provides quick feedback for editor actions
type EditorActionFeedback struct {
	LastAction     string
	LastActionTime time.Time
	ShowDuration   time.Duration
}

// NewEditorActionFeedback creates a new action feedback tracker
func NewEditorActionFeedback() *EditorActionFeedback {
	return &EditorActionFeedback{
		ShowDuration: 2 * time.Second,
	}
}

// RecordAction records an action with timestamp
func (eaf *EditorActionFeedback) RecordAction(action string) {
	eaf.LastAction = action
	eaf.LastActionTime = time.Now()
}

// GetActionFeedback returns the current action feedback if still visible
func (eaf *EditorActionFeedback) GetActionFeedback() (string, bool) {
	if eaf.LastAction == "" {
		return "", false
	}

	if time.Since(eaf.LastActionTime) > eaf.ShowDuration {
		return "", false
	}

	return eaf.LastAction, true
}
