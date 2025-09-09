package tui

import (
	"runtime"
	"strings"
)

// OSType represents the operating system type
type OSType int

const (
	OSMac OSType = iota
	OSLinux
	OSWindows
	OSUnknown
)

// GetOS returns the current operating system type
func GetOS() OSType {
	switch runtime.GOOS {
	case "darwin":
		return OSMac
	case "linux":
		return OSLinux
	case "windows":
		return OSWindows
	default:
		return OSUnknown
	}
}

// ShortcutKey represents a keyboard shortcut with OS-specific variations
type ShortcutKey struct {
	Mac     string
	Linux   string
	Windows string
	Default string // Fallback if OS-specific not defined
}

// Get returns the appropriate shortcut for the current OS
func (s ShortcutKey) Get() string {
	os := GetOS()
	switch os {
	case OSMac:
		if s.Mac != "" {
			return s.Mac
		}
	case OSLinux:
		if s.Linux != "" {
			return s.Linux
		}
	case OSWindows:
		if s.Windows != "" {
			return s.Windows
		}
	}
	return s.Default
}

// GetWithWarning returns the shortcut and a warning if there are known issues
func (s ShortcutKey) GetWithWarning() (shortcut string, warning string) {
	os := GetOS()
	shortcut = s.Get()
	
	// Check for known problematic combinations
	switch os {
	case OSLinux:
		switch shortcut {
		case "^s", "ctrl+s":
			warning = "(may need: stty -ixon)"
		case "^d", "ctrl+d":
			warning = "(caution: EOF signal)"
		case "^z", "ctrl+z":
			warning = "(caution: suspends process)"
		}
	case OSWindows:
		switch shortcut {
		case "shift+tab", "backtab":
			warning = "(terminal dependent)"
		}
	}
	
	return shortcut, warning
}

// ShortcutMap contains all keyboard shortcuts with OS-specific variations
var Shortcuts = struct {
	// File operations
	Save           ShortcutKey
	Delete         ShortcutKey
	ExternalEdit   ShortcutKey
	
	// Navigation
	Search         ShortcutKey
	SwitchPane     ShortcutKey
	ReverseSwitch  ShortcutKey
	Up             ShortcutKey
	Down           ShortcutKey
	
	// Edit operations
	Undo           ShortcutKey
	Clear          ShortcutKey
	Clean          ShortcutKey
	
	// Component operations
	New            ShortcutKey
	Edit           ShortcutKey
	Rename         ShortcutKey
	Clone          ShortcutKey
	Tag            ShortcutKey
	TagReload      ShortcutKey
	Archive        ShortcutKey
	Usage          ShortcutKey
	
	// Filter operations
	ToggleArchived ShortcutKey
	CycleType      ShortcutKey
	
	// Pipeline operations
	ReorderUp      ShortcutKey
	ReorderDown    ShortcutKey
	Diagram        ShortcutKey
	SetPipeline    ShortcutKey
	
	// System
	Quit           ShortcutKey
	Cancel         ShortcutKey
	Confirm        ShortcutKey
	Settings       ShortcutKey
	Preview        ShortcutKey
	Copy           ShortcutKey
}{
	// File operations - with OS-specific alternatives
	Save: ShortcutKey{
		Mac:     "ctrl+s",
		Linux:   "alt+s",   // Avoid Ctrl+S terminal conflict (XOFF)
		Windows: "alt+s",   // Consistent with Linux
		Default: "ctrl+s",
	},
	Delete: ShortcutKey{
		Mac:     "ctrl+d",
		Linux:   "alt+d",   // Avoid Ctrl+D EOF signal
		Windows: "alt+d",   // Consistent with Linux
		Default: "ctrl+d",
	},
	ExternalEdit: ShortcutKey{
		Mac:     "ctrl+x",
		Linux:   "alt+x",   // Avoid terminal cut conflict
		Windows: "alt+x",   // Consistent with Linux
		Default: "ctrl+x",
	},
	
	// Navigation - mostly consistent
	Search: ShortcutKey{
		Default: "/",
	},
	SwitchPane: ShortcutKey{
		Default: "tab",
	},
	ReverseSwitch: ShortcutKey{
		Mac:     "shift+tab",
		Linux:   "shift+tab",
		Windows: "backtab",    // Windows terminal compatibility
		Default: "shift+tab",
	},
	Up: ShortcutKey{
		Default: "↑",
	},
	Down: ShortcutKey{
		Default: "↓",
	},
	
	// Edit operations - with alternatives
	Undo: ShortcutKey{
		Mac:     "ctrl+z",
		Linux:   "alt+z",   // Avoid SIGTSTP (process suspension)
		Windows: "alt+z",   // Consistent with Linux
		Default: "ctrl+z",
	},
	Clear: ShortcutKey{
		Mac:     "ctrl+k",
		Linux:   "alt+k",   // Avoid readline kill-line
		Windows: "alt+k",   // Consistent with Linux
		Default: "ctrl+k",
	},
	Clean: ShortcutKey{
		Mac:     "ctrl+l",
		Linux:   "alt+l",   // Avoid clear screen
		Windows: "alt+l",   // Consistent with Linux
		Default: "ctrl+l",
	},
	
	// Component operations - single keys work everywhere
	New: ShortcutKey{
		Default: "n",
	},
	Edit: ShortcutKey{
		Default: "e",
	},
	Rename: ShortcutKey{
		Default: "R",
	},
	Clone: ShortcutKey{
		Default: "C",
	},
	Tag: ShortcutKey{
		Default: "t",
	},
	TagReload: ShortcutKey{
		Mac:     "ctrl+t",
		Linux:   "alt+t",   // Avoid readline transpose
		Windows: "alt+t",   // Consistent with Linux
		Default: "ctrl+t",
	},
	Archive: ShortcutKey{
		Default: "a",
	},
	Usage: ShortcutKey{
		Default: "u",
	},
	
	// Filter operations - with alternatives
	ToggleArchived: ShortcutKey{
		Mac:     "ctrl+a",
		Linux:   "alt+a",   // Avoid readline beginning-of-line
		Windows: "alt+a",   // Consistent with Linux
		Default: "ctrl+a",
	},
	CycleType: ShortcutKey{
		Mac:     "ctrl+t",
		Linux:   "alt+t",   // Avoid readline transpose
		Windows: "alt+t",   // Consistent with Linux
		Default: "ctrl+t",
	},
	
	// Pipeline operations
	ReorderUp: ShortcutKey{
		Mac:     "K",
		Linux:   "K",
		Windows: "K",
		Default: "K",
	},
	ReorderDown: ShortcutKey{
		Mac:     "J",
		Linux:   "J",
		Windows: "J",
		Default: "J",
	},
	Diagram: ShortcutKey{
		Default: "M",
	},
	SetPipeline: ShortcutKey{
		Default: "S",
	},
	
	// System - mostly consistent
	Quit: ShortcutKey{
		Default: "ctrl+c",
	},
	Cancel: ShortcutKey{
		Default: "esc",
	},
	Confirm: ShortcutKey{
		Default: "enter",
	},
	Settings: ShortcutKey{
		Default: "s",
	},
	Preview: ShortcutKey{
		Default: "p",
	},
	Copy: ShortcutKey{
		Default: "y",
	},
}

// GetShortcutHelp returns formatted help text for a shortcut
func GetShortcutHelp(name string, key ShortcutKey) string {
	shortcut, warning := key.GetWithWarning()
	if warning != "" {
		return shortcut + " " + name + " " + warning
	}
	return shortcut + " " + name
}

// GetOSName returns a friendly name for the current OS
func GetOSName() string {
	switch GetOS() {
	case OSMac:
		return "macOS"
	case OSLinux:
		return "Linux"
	case OSWindows:
		return "Windows"
	default:
		return "Unknown"
	}
}

// ShouldShowTerminalSetupWarning returns true if the OS needs special terminal setup
func ShouldShowTerminalSetupWarning() bool {
	return GetOS() == OSLinux
}

// GetTerminalSetupMessage returns OS-specific terminal setup instructions
func GetTerminalSetupMessage() string {
	switch GetOS() {
	case OSLinux:
		return "TIP: Run 'stty -ixon' to enable Ctrl+S and other shortcuts in your terminal"
	case OSWindows:
		return "TIP: For best experience, use Windows Terminal or PowerShell"
	default:
		return ""
	}
}

// FormatShortcutForHelp formats a shortcut key for display in help text
func FormatShortcutForHelp(key ShortcutKey) string {
	shortcut := key.Get()
	// Convert common representations to display format
	// Use M- prefix for Alt on Linux/Windows (common terminal convention)
	if GetOS() == OSLinux || GetOS() == OSWindows {
		shortcut = strings.ReplaceAll(shortcut, "alt+", "M-")
	} else {
		shortcut = strings.ReplaceAll(shortcut, "alt+", "⌥")
	}
	shortcut = strings.ReplaceAll(shortcut, "ctrl+", "^")
	shortcut = strings.ReplaceAll(shortcut, "shift+", "⇧")
	shortcut = strings.ReplaceAll(shortcut, "cmd+", "⌘")
	
	// Handle function keys - capitalize them for display
	if strings.HasPrefix(shortcut, "f") && len(shortcut) <= 3 {
		return strings.ToUpper(shortcut)
	}
	
	// Handle delete key
	if shortcut == "delete" {
		return "Del"
	}
	
	return shortcut
}