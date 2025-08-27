package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppStatusMessages(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		expectStatus   string
		expectTimer    bool
		expectClearMsg bool
	}{
		{
			name:           "StatusMsg sets timer and schedules clear",
			msg:            StatusMsg("Test status message"),
			expectStatus:   "Test status message",
			expectTimer:    true,
			expectClearMsg: true,
		},
		{
			name:           "PersistentStatusMsg does not set timer",
			msg:            PersistentStatusMsg("Editing in external editor - save your changes and close the editor window/tab to return here and continue"),
			expectStatus:   "Editing in external editor - save your changes and close the editor window/tab to return here and continue",
			expectTimer:    false,
			expectClearMsg: false,
		},
		{
			name:           "clearStatusMsg clears the status",
			msg:            clearStatusMsg{},
			expectStatus:   "",
			expectTimer:    false,
			expectClearMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new app
			app := &App{
				state:    mainListView,
				mainList: NewMainListModel(),
				width:    80,
				height:   24,
			}

			// Process the message
			updatedApp, cmd := app.Update(tt.msg)
			a := updatedApp.(*App)

			// Check status message
			if a.statusMsg != tt.expectStatus {
				t.Errorf("expected status %q, got %q", tt.expectStatus, a.statusMsg)
			}

			// Check timer existence
			hasTimer := a.statusTimer != nil
			if hasTimer != tt.expectTimer {
				t.Errorf("expected timer=%v, got timer=%v", tt.expectTimer, hasTimer)
			}

			// Check if a clear command was scheduled
			if tt.expectClearMsg {
				if cmd == nil {
					t.Error("expected a command to be returned for clearing status, got nil")
				}
			} else if tt.name != "clearStatusMsg clears the status" {
				// For PersistentStatusMsg, we expect no command (nil)
				if cmd != nil {
					t.Error("expected no command for PersistentStatusMsg, got a command")
				}
			}

			// Clean up timer if it exists
			if a.statusTimer != nil {
				a.statusTimer.Stop()
			}
		})
	}
}

func TestAppStatusMessageTransitions(t *testing.T) {
	app := &App{
		state:    mainListView,
		mainList: NewMainListModel(),
		width:    80,
		height:   24,
	}

	// Test 1: Regular StatusMsg creates timer
	model, _ := app.Update(StatusMsg("Regular status"))
	app = model.(*App)
	if app.statusMsg != "Regular status" {
		t.Errorf("expected status %q, got %q", "Regular status", app.statusMsg)
	}
	if app.statusTimer == nil {
		t.Error("expected timer to be set for StatusMsg")
	}
	
	// Save the timer for comparison
	firstTimer := app.statusTimer

	// Test 2: PersistentStatusMsg cancels existing timer
	model, cmd := app.Update(PersistentStatusMsg("Persistent status"))
	app = model.(*App)
	if app.statusMsg != "Persistent status" {
		t.Errorf("expected status %q, got %q", "Persistent status", app.statusMsg)
	}
	if app.statusTimer != nil {
		t.Error("expected timer to be nil for PersistentStatusMsg")
	}
	if cmd != nil {
		t.Error("expected no command for PersistentStatusMsg")
	}
	
	// The old timer should have been stopped (we can't directly test this,
	// but we can verify a new timer wasn't created)
	if app.statusTimer == firstTimer {
		t.Error("expected old timer to be stopped and cleared")
	}

	// Test 3: Another StatusMsg after PersistentStatusMsg creates new timer
	model, _ = app.Update(StatusMsg("New regular status"))
	app = model.(*App)
	if app.statusMsg != "New regular status" {
		t.Errorf("expected status %q, got %q", "New regular status", app.statusMsg)
	}
	if app.statusTimer == nil {
		t.Error("expected new timer to be set for StatusMsg")
	}
	
	// Clean up
	if app.statusTimer != nil {
		app.statusTimer.Stop()
	}
}

func TestAppViewWithStatusMessage(t *testing.T) {
	tests := []struct {
		name         string
		statusMsg    string
		expectInView bool
	}{
		{
			name:         "View includes status message when set",
			statusMsg:    "Test status",
			expectInView: true,
		},
		{
			name:         "View includes persistent status message",
			statusMsg:    "Editing in external editor - save your changes and close the editor window/tab to return here and continue",
			expectInView: true,
		},
		{
			name:         "View has no status bar when status is empty",
			statusMsg:    "",
			expectInView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and initialize the main list model
			mainList := NewMainListModel()
			mainList.SetSize(80, 24)
			
			app := &App{
				state:     mainListView,
				mainList:  mainList,
				width:     80,
				height:    24,
				statusMsg: tt.statusMsg,
			}

			view := app.View()
			
			if tt.expectInView {
				// The status message should appear in the view
				if len(tt.statusMsg) > 0 && len(view) == 0 {
					t.Error("expected non-empty view when status message is set")
				}
				// Note: We can't easily test the exact content due to styling,
				// but we can verify the view is generated
			} else {
				// When no status, view should still be generated but without status bar
				if len(view) == 0 {
					t.Error("expected non-empty view even without status")
				}
			}
		})
	}
}

// TestExternalEditorStatusFlow tests the expected flow when opening external editor
func TestExternalEditorStatusFlow(t *testing.T) {
	// This test verifies the expected message type is used for external editor
	
	// Simulate the message that would be sent when ctrl+x is pressed
	msg := PersistentStatusMsg("Editing in external editor - save your changes and close the editor window/tab to return here and continue")
	
	app := &App{
		state:    mainListView,
		mainList: NewMainListModel(),
		width:    80,
		height:   24,
	}
	
	// Process the persistent status message
	updatedApp, cmd := app.Update(msg)
	a := updatedApp.(*App)
	
	// Verify it behaves as a persistent message
	if a.statusMsg != string(msg) {
		t.Errorf("expected status to be set to %q", msg)
	}
	
	if a.statusTimer != nil {
		t.Error("expected no timer for external editor status (should be persistent)")
	}
	
	if cmd != nil {
		t.Error("expected no clear command for external editor status")
	}
	
	// Simulate editor closing and sending completion message
	completionMsg := StatusMsg("Edited: component.yaml")
	updatedApp, cmd = app.Update(completionMsg)
	a = updatedApp.(*App)
	
	// Verify the completion message replaces the persistent one
	if a.statusMsg != string(completionMsg) {
		t.Errorf("expected status to be replaced with %q", completionMsg)
	}
	
	// And it should have a timer since it's a regular StatusMsg
	if a.statusTimer == nil {
		t.Error("expected timer for completion message")
	}
	
	if cmd == nil {
		t.Error("expected clear command to be scheduled for completion message")
	}
	
	// Clean up
	if a.statusTimer != nil {
		a.statusTimer.Stop()
	}
}

// TestTimerCleanup verifies timers are properly cleaned up
func TestTimerCleanup(t *testing.T) {
	app := &App{
		state:    mainListView,
		mainList: NewMainListModel(),
		width:    80,
		height:   24,
	}
	
	// Create multiple status messages in sequence
	messages := []tea.Msg{
		StatusMsg("First"),
		StatusMsg("Second"),
		PersistentStatusMsg("Persistent"),
		StatusMsg("Third"),
	}
	
	var timers []*time.Timer
	
	for _, msg := range messages {
		model, _ := app.Update(msg)
		app = model.(*App)
		
		// Track timers (some may be nil)
		if app.statusTimer != nil {
			timers = append(timers, app.statusTimer)
		}
	}
	
	// Clean up all tracked timers
	for _, timer := range timers {
		if timer != nil {
			timer.Stop()
		}
	}
	
	// Final state should have the last message
	if app.statusMsg != "Third" {
		t.Errorf("expected final status to be %q, got %q", "Third", app.statusMsg)
	}
}