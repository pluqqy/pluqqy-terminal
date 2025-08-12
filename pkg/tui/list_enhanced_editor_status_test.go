package tui

import (
	"testing"
	"time"
)

func TestStatusManager_ShowFeedback(t *testing.T) {
	sm := NewStatusManager()
	
	// Test showing feedback
	cmd := sm.ShowFeedback("✓", "Test message", StatusTypeSuccess)
	if cmd == nil {
		t.Error("ShowFeedback should return a command")
	}
	
	if sm.CurrentStatus == nil {
		t.Fatal("CurrentStatus should not be nil after ShowFeedback")
	}
	
	if sm.CurrentStatus.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", sm.CurrentStatus.Message)
	}
	
	if sm.CurrentStatus.Icon != "✓" {
		t.Errorf("Expected icon '✓', got '%s'", sm.CurrentStatus.Icon)
	}
	
	if sm.CurrentStatus.Type != StatusTypeSuccess {
		t.Errorf("Expected type StatusTypeSuccess, got %v", sm.CurrentStatus.Type)
	}
}

func TestStatusManager_IsActive(t *testing.T) {
	sm := NewStatusManager()
	
	// Initially not active
	if sm.IsActive() {
		t.Error("StatusManager should not be active initially")
	}
	
	// Show feedback
	sm.ShowFeedback("✓", "Test", StatusTypeSuccess)
	
	// Should be active
	if !sm.IsActive() {
		t.Error("StatusManager should be active after ShowFeedback")
	}
	
	// Simulate expiration
	sm.CurrentStatus.ShowUntil = time.Now().Add(-1 * time.Second)
	
	// Should not be active
	if sm.IsActive() {
		t.Error("StatusManager should not be active after expiration")
	}
}

func TestStatusManager_Clear(t *testing.T) {
	sm := NewStatusManager()
	
	sm.ShowFeedback("✓", "Test", StatusTypeSuccess)
	sm.Clear()
	
	if sm.CurrentStatus != nil {
		t.Error("CurrentStatus should be nil after Clear")
	}
}

func TestStatusManager_GetStatus(t *testing.T) {
	sm := NewStatusManager()
	
	// No status initially
	status, ok := sm.GetStatus()
	if ok {
		t.Error("GetStatus should return false when no status")
	}
	if status != "" {
		t.Error("GetStatus should return empty string when no status")
	}
	
	// With status
	sm.ShowFeedback("✓", "Test message", StatusTypeSuccess)
	status, ok = sm.GetStatus()
	if !ok {
		t.Error("GetStatus should return true when status is active")
	}
	if status != "✓ Test message" {
		t.Errorf("Expected status '✓ Test message', got '%s'", status)
	}
}

func TestClipboardStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   *ClipboardStatus
		expected string
	}{
		{
			name: "no content",
			status: &ClipboardStatus{
				HasContent: false,
				LineCount:  0,
				WillClean:  false,
			},
			expected: "",
		},
		{
			name: "content without cleaning",
			status: &ClipboardStatus{
				HasContent: true,
				LineCount:  5,
				WillClean:  false,
			},
			expected: "📋 5 lines ready",
		},
		{
			name: "content with cleaning",
			status: &ClipboardStatus{
				HasContent: true,
				LineCount:  10,
				WillClean:  true,
			},
			expected: "📋 10 lines ready (will clean)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClipboardStatusString(tt.status)
			if result != tt.expected {
				t.Errorf("GetClipboardStatusString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEditorActionFeedback(t *testing.T) {
	eaf := NewEditorActionFeedback()
	
	// Initially no feedback
	feedback, ok := eaf.GetActionFeedback()
	if ok || feedback != "" {
		t.Error("Should have no initial feedback")
	}
	
	// Record an action
	eaf.RecordAction("✓ Saved")
	
	// Should have feedback
	feedback, ok = eaf.GetActionFeedback()
	if !ok {
		t.Error("Should have feedback after recording action")
	}
	if feedback != "✓ Saved" {
		t.Errorf("Expected feedback '✓ Saved', got '%s'", feedback)
	}
	
	// Simulate expiration
	eaf.LastActionTime = time.Now().Add(-3 * time.Second)
	
	// Should not have feedback
	feedback, ok = eaf.GetActionFeedback()
	if ok || feedback != "" {
		t.Error("Should not have feedback after expiration")
	}
}

func TestStatusMessages(t *testing.T) {
	// Test ShowPastedStatus
	cmd := ShowPastedStatus(10, false)
	if cmd == nil {
		t.Error("ShowPastedStatus should return a command")
	}
	
	cmd = ShowPastedStatus(5, true)
	if cmd == nil {
		t.Error("ShowPastedStatus with cleaning should return a command")
	}
	
	// Test ShowClearedStatus
	cmd = ShowClearedStatus()
	if cmd == nil {
		t.Error("ShowClearedStatus should return a command")
	}
	
	// Test ShowSavedStatus
	cmd = ShowSavedStatus("test.md")
	if cmd == nil {
		t.Error("ShowSavedStatus should return a command")
	}
	
	// Test ShowNothingToPasteStatus
	cmd = ShowNothingToPasteStatus()
	if cmd == nil {
		t.Error("ShowNothingToPasteStatus should return a command")
	}
}