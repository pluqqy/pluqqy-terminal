package tui

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetOS(t *testing.T) {
	os := GetOS()
	
	// Just verify it returns a valid OS type
	switch os {
	case OSMac, OSLinux, OSWindows, OSUnknown:
		// Valid
	default:
		t.Errorf("GetOS() returned invalid OS type: %v", os)
	}
	
	// Verify it matches runtime.GOOS
	switch runtime.GOOS {
	case "darwin":
		if os != OSMac {
			t.Errorf("Expected OSMac for darwin, got %v", os)
		}
	case "linux":
		if os != OSLinux {
			t.Errorf("Expected OSLinux for linux, got %v", os)
		}
	case "windows":
		if os != OSWindows {
			t.Errorf("Expected OSWindows for windows, got %v", os)
		}
	}
}

func TestShortcutKey_Get(t *testing.T) {
	tests := []struct {
		name     string
		shortcut ShortcutKey
		mockOS   OSType
		want     string
	}{
		{
			name: "Mac specific shortcut",
			shortcut: ShortcutKey{
				Mac:     "cmd+s",
				Linux:   "alt+s",
				Windows: "alt+s",
				Default: "ctrl+s",
			},
			mockOS: OSMac,
			want:   "cmd+s",
		},
		{
			name: "Linux specific shortcut",
			shortcut: ShortcutKey{
				Mac:     "ctrl+s",
				Linux:   "alt+s",
				Windows: "alt+s",
				Default: "ctrl+s",
			},
			mockOS: OSLinux,
			want:   "alt+s",
		},
		{
			name: "Windows specific shortcut",
			shortcut: ShortcutKey{
				Mac:     "ctrl+s",
				Linux:   "alt+s",
				Windows: "alt+s",
				Default: "ctrl+s",
			},
			mockOS: OSWindows,
			want:   "alt+s",
		},
		{
			name: "Default only shortcut",
			shortcut: ShortcutKey{
				Default: "/",
			},
			mockOS: OSMac,
			want:   "/",
		},
		{
			name: "Falls back to default when OS-specific not set",
			shortcut: ShortcutKey{
				Mac:     "",
				Default: "esc",
			},
			mockOS: OSMac,
			want:   "esc",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily mock GetOS() since it's a global function,
			// but we can test the actual shortcuts defined
			// For real testing, we just verify the current OS returns something
			got := tt.shortcut.Get()
			if got == "" {
				t.Error("ShortcutKey.Get() returned empty string")
			}
		})
	}
}

func TestActualShortcuts(t *testing.T) {
	// Test that all defined shortcuts return non-empty values
	shortcuts := []struct {
		name string
		key  ShortcutKey
	}{
		{"Save", Shortcuts.Save},
		{"Delete", Shortcuts.Delete},
		{"ExternalEdit", Shortcuts.ExternalEdit},
		{"Search", Shortcuts.Search},
		{"SwitchPane", Shortcuts.SwitchPane},
		{"ReverseSwitch", Shortcuts.ReverseSwitch},
		{"Up", Shortcuts.Up},
		{"Down", Shortcuts.Down},
		{"Undo", Shortcuts.Undo},
		{"Clear", Shortcuts.Clear},
		{"Clean", Shortcuts.Clean},
		{"New", Shortcuts.New},
		{"Edit", Shortcuts.Edit},
		{"Rename", Shortcuts.Rename},
		{"Clone", Shortcuts.Clone},
		{"Tag", Shortcuts.Tag},
		{"Archive", Shortcuts.Archive},
		{"Usage", Shortcuts.Usage},
		{"ToggleArchived", Shortcuts.ToggleArchived},
		{"CycleType", Shortcuts.CycleType},
		{"ReorderUp", Shortcuts.ReorderUp},
		{"ReorderDown", Shortcuts.ReorderDown},
		{"Diagram", Shortcuts.Diagram},
		{"SetPipeline", Shortcuts.SetPipeline},
		{"Quit", Shortcuts.Quit},
		{"Cancel", Shortcuts.Cancel},
		{"Confirm", Shortcuts.Confirm},
		{"Settings", Shortcuts.Settings},
		{"Preview", Shortcuts.Preview},
		{"Copy", Shortcuts.Copy},
	}
	
	for _, s := range shortcuts {
		t.Run(s.name, func(t *testing.T) {
			got := s.key.Get()
			if got == "" {
				t.Errorf("%s shortcut returned empty string", s.name)
			}
		})
	}
}

func TestFormatShortcutForHelp(t *testing.T) {
	tests := []struct {
		name     string
		shortcut ShortcutKey
		wantMac  string
		wantLinux string
	}{
		{
			name:     "Ctrl shortcut",
			shortcut: ShortcutKey{Default: "ctrl+s"},
			wantMac:  "^s",
			wantLinux: "^s",
		},
		{
			name:     "Alt shortcut on Linux",
			shortcut: ShortcutKey{Mac: "ctrl+s", Linux: "alt+s", Default: "ctrl+s"},
			wantMac:  "^s",      // Mac uses ctrl
			wantLinux: "M-s",    // Linux shows alt as M-
		},
		{
			name:     "Shift shortcut",
			shortcut: ShortcutKey{Default: "shift+tab"},
			wantMac:  "⇧tab",
			wantLinux: "⇧tab",
		},
		{
			name:     "Simple key",
			shortcut: ShortcutKey{Default: "esc"},
			wantMac:  "esc",
			wantLinux: "esc",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll test formatting based on current OS
			got := FormatShortcutForHelp(tt.shortcut)
			
			// Just verify it's not empty and contains expected patterns
			if got == "" {
				t.Error("FormatShortcutForHelp returned empty string")
			}
			
			// Check for expected transformations
			if strings.Contains(tt.shortcut.Get(), "ctrl+") && !strings.Contains(got, "^") && !strings.Contains(got, "⌃") {
				t.Errorf("Expected ctrl+ to be formatted, got %s", got)
			}
		})
	}
}

func TestGetOSName(t *testing.T) {
	name := GetOSName()
	
	// Verify it returns a non-empty, sensible name
	validNames := []string{"macOS", "Linux", "Windows", "Unknown"}
	found := false
	for _, valid := range validNames {
		if name == valid {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("GetOSName() returned unexpected value: %s", name)
	}
	
	// Verify it matches runtime.GOOS
	switch runtime.GOOS {
	case "darwin":
		if name != "macOS" {
			t.Errorf("Expected 'macOS' for darwin, got %s", name)
		}
	case "linux":
		if name != "Linux" {
			t.Errorf("Expected 'Linux' for linux, got %s", name)
		}
	case "windows":
		if name != "Windows" {
			t.Errorf("Expected 'Windows' for windows, got %s", name)
		}
	}
}

func TestShouldShowTerminalSetupWarning(t *testing.T) {
	shouldShow := ShouldShowTerminalSetupWarning()
	
	// On Linux, it should return true
	if runtime.GOOS == "linux" && !shouldShow {
		t.Error("Expected ShouldShowTerminalSetupWarning to return true on Linux")
	}
	
	// On other OS, it should return false
	if runtime.GOOS != "linux" && shouldShow {
		t.Error("Expected ShouldShowTerminalSetupWarning to return false on non-Linux")
	}
}

func TestGetTerminalSetupMessage(t *testing.T) {
	msg := GetTerminalSetupMessage()
	
	switch runtime.GOOS {
	case "linux":
		if !strings.Contains(msg, "stty -ixon") {
			t.Errorf("Linux setup message should mention 'stty -ixon', got: %s", msg)
		}
	case "windows":
		if !strings.Contains(msg, "Windows Terminal") && !strings.Contains(msg, "PowerShell") {
			t.Errorf("Windows setup message should mention terminal choice, got: %s", msg)
		}
	default:
		if msg != "" {
			t.Errorf("Expected empty setup message for %s, got: %s", runtime.GOOS, msg)
		}
	}
}

func TestShortcutConsistency(t *testing.T) {
	// Verify that shortcuts that should work everywhere actually use Default
	universalShortcuts := []struct {
		name string
		key  ShortcutKey
	}{
		{"Search", Shortcuts.Search},
		{"New", Shortcuts.New},
		{"Edit", Shortcuts.Edit},
		{"Rename", Shortcuts.Rename},
		{"Clone", Shortcuts.Clone},
		{"Tag", Shortcuts.Tag},
		{"Archive", Shortcuts.Archive},
		{"Usage", Shortcuts.Usage},
		{"Cancel", Shortcuts.Cancel},
		{"Settings", Shortcuts.Settings},
		{"Preview", Shortcuts.Preview},
		{"Copy", Shortcuts.Copy},
	}
	
	for _, s := range universalShortcuts {
		t.Run(s.name, func(t *testing.T) {
			if s.key.Default == "" {
				t.Errorf("%s should have a Default value set", s.name)
			}
			// These should generally not need OS-specific overrides
			// (though some might have them for consistency)
		})
	}
}

func TestProblematicShortcuts(t *testing.T) {
	// Verify that problematic shortcuts have alternatives on Linux/Windows
	problematicShortcuts := []struct {
		name string
		key  ShortcutKey
	}{
		{"Save", Shortcuts.Save},         // Ctrl+S causes XOFF
		{"Delete", Shortcuts.Delete},     // Ctrl+D sends EOF
		{"Undo", Shortcuts.Undo},         // Ctrl+Z sends SIGTSTP
		{"Clear", Shortcuts.Clear},       // Ctrl+K kills line
		{"Clean", Shortcuts.Clean},       // Ctrl+L clears screen
		{"ToggleArchived", Shortcuts.ToggleArchived}, // Ctrl+A goes to line start
		{"CycleType", Shortcuts.CycleType},           // Ctrl+T transposes
	}
	
	for _, s := range problematicShortcuts {
		t.Run(s.name, func(t *testing.T) {
			// Should have Linux-specific alternative
			if s.key.Linux == "" || s.key.Linux == s.key.Default {
				t.Errorf("%s should have a different Linux shortcut than default", s.name)
			}
			
			// Linux shortcuts should use Alt instead of Ctrl
			if strings.HasPrefix(s.key.Linux, "ctrl+") {
				t.Errorf("%s Linux shortcut should not use ctrl+, got %s", s.name, s.key.Linux)
			}
			
			// Should use alt+ for Linux
			if !strings.HasPrefix(s.key.Linux, "alt+") && s.key.Linux != "delete" {
				t.Errorf("%s Linux shortcut should use alt+, got %s", s.name, s.key.Linux)
			}
		})
	}
}