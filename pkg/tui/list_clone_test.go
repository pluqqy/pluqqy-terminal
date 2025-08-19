package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCloneState_HandleInput(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *CloneState
		input       tea.KeyMsg
		wantActive  bool
		wantHandled bool
		checkState  func(t *testing.T, cs *CloneState)
	}{
		{
			name: "escape cancels clone mode",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyEsc},
			wantActive:  false,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.Active {
					t.Error("Clone state should be inactive after escape")
				}
				if cs.NewName != "" {
					t.Error("NewName should be cleared after escape")
				}
			},
		},
		{
			name: "enter with valid name triggers clone",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.NewName = "Valid Name"
				cs.ValidationError = ""
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyEnter},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				// The actual clone operation would be async
				if cs.NewName != "Valid Name" {
					t.Error("NewName should be preserved")
				}
			},
		},
		{
			name: "enter with invalid name does nothing",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.NewName = ""
				cs.ValidationError = "Name cannot be empty"
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyEnter},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.ValidationError == "" {
					t.Error("Validation error should be preserved")
				}
			},
		},
		{
			name: "backspace removes last character",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.NewName = "Test Name"
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyBackspace},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.NewName != "Test Nam" {
					t.Errorf("Expected 'Test Nam', got '%s'", cs.NewName)
				}
			},
		},
		{
			name: "tab toggles archive destination",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.CloneToArchive = false
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyTab},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if !cs.CloneToArchive {
					t.Error("CloneToArchive should be toggled to true")
				}
			},
		},
		{
			name: "regular character adds to name",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.NewName = "Test"
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.NewName != "Tests" {
					t.Errorf("Expected 'Tests', got '%s'", cs.NewName)
				}
			},
		},
		{
			name: "space character adds space to name",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test Component", "component", "components/prompts/test.md", false)
				cs.NewName = "Test"
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeySpace},
			wantActive:  true,
			wantHandled: true,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.NewName != "Test " {
					t.Errorf("Expected 'Test ', got '%s'", cs.NewName)
				}
			},
		},
		{
			name: "input not handled when inactive",
			setup: func() *CloneState {
				cs := NewCloneState()
				// Don't start clone mode
				return cs
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantActive:  false,
			wantHandled: false,
			checkState: func(t *testing.T, cs *CloneState) {
				if cs.Active {
					t.Error("Clone state should remain inactive")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := tt.setup()
			handled, _ := cs.HandleInput(tt.input)

			if handled != tt.wantHandled {
				t.Errorf("HandleInput() handled = %v, want %v", handled, tt.wantHandled)
			}

			if cs.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", cs.Active, tt.wantActive)
			}

			if tt.checkState != nil {
				tt.checkState(t, cs)
			}
		})
	}
}

func TestCloneState_Start(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		itemType    string
		path        string
		isArchived  bool
		wantNewName string
		wantActive  bool
	}{
		{
			name:        "start component clone",
			displayName: "My Component",
			itemType:    "component",
			path:        "components/prompts/my-component.md",
			isArchived:  false,
			wantNewName: "(Copy) My Component",
			wantActive:  true,
		},
		{
			name:        "start pipeline clone",
			displayName: "My Pipeline",
			itemType:    "pipeline",
			path:        "pipelines/my-pipeline.yaml",
			isArchived:  false,
			wantNewName: "(Copy) My Pipeline",
			wantActive:  true,
		},
		{
			name:        "start archived item clone",
			displayName: "Archived Item",
			itemType:    "component",
			path:        "components/contexts/archived.md",
			isArchived:  true,
			wantNewName: "(Copy) Archived Item",
			wantActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewCloneState()
			cs.Start(tt.displayName, tt.itemType, tt.path, tt.isArchived)

			if cs.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", cs.Active, tt.wantActive)
			}

			if cs.NewName != tt.wantNewName {
				t.Errorf("NewName = %v, want %v", cs.NewName, tt.wantNewName)
			}

			if cs.ItemType != tt.itemType {
				t.Errorf("ItemType = %v, want %v", cs.ItemType, tt.itemType)
			}

			if cs.OriginalName != tt.displayName {
				t.Errorf("OriginalName = %v, want %v", cs.OriginalName, tt.displayName)
			}

			if cs.OriginalPath != tt.path {
				t.Errorf("OriginalPath = %v, want %v", cs.OriginalPath, tt.path)
			}

			if cs.IsArchived != tt.isArchived {
				t.Errorf("IsArchived = %v, want %v", cs.IsArchived, tt.isArchived)
			}

			if cs.CloneToArchive != tt.isArchived {
				t.Errorf("CloneToArchive should default to source location")
			}
		})
	}
}

func TestCloneState_Validate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *CloneState
		wantError    bool
		wantErrorMsg string
	}{
		{
			name: "empty name validation",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test", "component", "components/prompts/test.md", false)
				cs.NewName = ""
				cs.validate()
				return cs
			},
			wantError:    true,
			wantErrorMsg: "Name cannot be empty",
		},
		{
			name: "whitespace only name validation",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test", "component", "components/prompts/test.md", false)
				cs.NewName = "   "
				cs.validate()
				return cs
			},
			wantError:    true,
			wantErrorMsg: "Name cannot be empty",
		},
		{
			name: "same name same location validation",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test", "component", "components/prompts/test.md", false)
				cs.NewName = "Test"
				cs.CloneToArchive = false
				cs.validate()
				return cs
			},
			wantError:    true,
			wantErrorMsg: "Name must be different from original when cloning to same location",
		},
		{
			name: "same name different location is valid",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test", "component", "components/prompts/test.md", false)
				cs.NewName = "Test"
				cs.CloneToArchive = true
				cs.validate()
				return cs
			},
			wantError:    false,
			wantErrorMsg: "",
		},
		{
			name: "valid different name",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.Start("Test", "component", "components/prompts/test.md", false)
				cs.NewName = "Test Copy"
				cs.validate()
				return cs
			},
			wantError:    false,
			wantErrorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := tt.setup()

			hasError := cs.ValidationError != ""
			if hasError != tt.wantError {
				t.Errorf("validation error presence = %v, want %v", hasError, tt.wantError)
			}

			if tt.wantError && cs.ValidationError != tt.wantErrorMsg {
				t.Errorf("ValidationError = %v, want %v", cs.ValidationError, tt.wantErrorMsg)
			}
		})
	}
}

func TestCloneState_GetSlugifiedName(t *testing.T) {
	tests := []struct {
		name    string
		newName string
		want    string
	}{
		{
			name:    "spaces to hyphens",
			newName: "My New Component",
			want:    "my-new-component",
		},
		{
			name:    "special characters removed",
			newName: "Component@#$Name!",
			want:    "component-name",
		},
		{
			name:    "multiple spaces collapsed",
			newName: "Name   With    Spaces",
			want:    "name-with-spaces",
		},
		{
			name:    "empty name returns empty",
			newName: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewCloneState()
			cs.NewName = tt.newName
			got := cs.GetSlugifiedName()

			if got != tt.want {
				t.Errorf("GetSlugifiedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloneState_Reset(t *testing.T) {
	cs := NewCloneState()
	cs.Start("Test Component", "component", "test.md", true)
	cs.NewName = "Modified Name"
	cs.ValidationError = "Some error"
	cs.CloneToArchive = true

	cs.Reset()

	if cs.Active {
		t.Error("Active should be false after reset")
	}
	if cs.ItemType != "" {
		t.Error("ItemType should be empty after reset")
	}
	if cs.OriginalName != "" {
		t.Error("OriginalName should be empty after reset")
	}
	if cs.OriginalPath != "" {
		t.Error("OriginalPath should be empty after reset")
	}
	if cs.NewName != "" {
		t.Error("NewName should be empty after reset")
	}
	if cs.ValidationError != "" {
		t.Error("ValidationError should be empty after reset")
	}
	if cs.IsArchived {
		t.Error("IsArchived should be false after reset")
	}
	if cs.CloneToArchive {
		t.Error("CloneToArchive should be false after reset")
	}
}

func TestCloneOperator_GetCloneSuggestion(t *testing.T) {
	co := NewCloneOperator()

	tests := []struct {
		name         string
		originalName string
		want         string
	}{
		{
			name:         "simple name",
			originalName: "Component",
			want:         "(Copy) Component",
		},
		{
			name:         "name with copy",
			originalName: "(Copy) Component",
			want:         "(Copy 2) Component",
		},
		{
			name:         "name with copy 2",
			originalName: "(Copy 2) Component",
			want:         "(Copy 3) Component",
		},
		{
			name:         "name with copy 10",
			originalName: "(Copy 10) Component",
			want:         "(Copy 11) Component",
		},
		{
			name:         "name with parentheses",
			originalName: "Component (Beta)",
			want:         "(Copy) Component (Beta)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := co.GetCloneSuggestion(tt.originalName)
			if got != tt.want {
				t.Errorf("GetCloneSuggestion(%s) = %v, want %v", tt.originalName, got, tt.want)
			}
		})
	}
}

func TestCloneOperator_CanCloneToLocation(t *testing.T) {
	co := NewCloneOperator()

	tests := []struct {
		name        string
		itemType    string
		fromArchive bool
		toArchive   bool
		wantCan     bool
		wantDesc    string
	}{
		{
			name:        "clone within active",
			itemType:    "component",
			fromArchive: false,
			toArchive:   false,
			wantCan:     true,
			wantDesc:    "Cloning within active",
		},
		{
			name:        "clone within archive",
			itemType:    "component",
			fromArchive: true,
			toArchive:   true,
			wantCan:     true,
			wantDesc:    "Cloning within archive",
		},
		{
			name:        "clone from active to archive",
			itemType:    "component",
			fromArchive: false,
			toArchive:   true,
			wantCan:     true,
			wantDesc:    "Cloning to archive",
		},
		{
			name:        "restore from archive as copy",
			itemType:    "component",
			fromArchive: true,
			toArchive:   false,
			wantCan:     true,
			wantDesc:    "Restoring from archive as copy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			can, desc := co.CanCloneToLocation(tt.itemType, tt.fromArchive, tt.toArchive)

			if can != tt.wantCan {
				t.Errorf("CanCloneToLocation() can = %v, want %v", can, tt.wantCan)
			}

			if desc != tt.wantDesc {
				t.Errorf("CanCloneToLocation() desc = %v, want %v", desc, tt.wantDesc)
			}
		})
	}
}

func TestCloneState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *CloneState
		want  bool
	}{
		{
			name: "valid state",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.NewName = "Valid Name"
				cs.ValidationError = ""
				return cs
			},
			want: true,
		},
		{
			name: "empty name is invalid",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.NewName = ""
				cs.ValidationError = ""
				return cs
			},
			want: false,
		},
		{
			name: "validation error makes invalid",
			setup: func() *CloneState {
				cs := NewCloneState()
				cs.NewName = "Name"
				cs.ValidationError = "Some error"
				return cs
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := tt.setup()
			got := cs.IsValid()

			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCloneState_ConcurrentAccess tests for race conditions in clone state
func TestCloneState_ConcurrentAccess(t *testing.T) {
	cs := NewCloneState()
	cs.Start("Test Component", "component", "test.md", false)

	var wg sync.WaitGroup
	iterations := 100

	// Test concurrent reads and writes
	wg.Add(4)

	// Goroutine 1: Repeatedly handle input
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			cs.HandleInput(msg)
		}
	}()

	// Goroutine 2: Repeatedly validate (with lock)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			cs.mu.Lock()
			cs.validate()
			cs.mu.Unlock()
		}
	}()

	// Goroutine 3: Repeatedly check validity
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = cs.IsValid()
			_ = cs.GetSlugifiedName()
		}
	}()

	// Goroutine 4: Repeatedly toggle archive destination
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			msg := tea.KeyMsg{Type: tea.KeyTab}
			cs.HandleInput(msg)
		}
	}()

	wg.Wait()

	// Verify state is still consistent
	if cs.ItemType != "component" {
		t.Error("ItemType was corrupted during concurrent access")
	}
	if cs.OriginalName != "Test Component" {
		t.Error("OriginalName was corrupted during concurrent access")
	}
}

// TestCloneOperator_ConcurrentOperations tests concurrent clone operations
func TestCloneOperator_ConcurrentOperations(t *testing.T) {
	co := NewCloneOperator()

	var wg sync.WaitGroup
	iterations := 50

	wg.Add(3)

	// Goroutine 1: Validate component clones
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = co.ValidateComponentClone("components/prompts/test.md", "Test Name", i%2 == 0)
		}
	}()

	// Goroutine 2: Validate pipeline clones
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = co.ValidatePipelineClone("Test Pipeline", i%2 == 0)
		}
	}()

	// Goroutine 3: Generate clone suggestions
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = co.GetCloneSuggestion("Original Name")
			_, _ = co.CanCloneToLocation("component", i%2 == 0, i%2 == 1)
		}
	}()

	wg.Wait()

	// If we get here without panics or race conditions, the test passes
}

// TestCloneState_AutomaticNaming tests the automatic collision-free naming
func TestCloneState_AutomaticNaming(t *testing.T) {
	// Create temporary test directory structure matching .pluqqy
	tempDir := t.TempDir()
	testPluqqyDir := filepath.Join(tempDir, ".pluqqy")

	// Create test directories
	os.MkdirAll(filepath.Join(testPluqqyDir, "components", "prompts"), 0755)
	os.MkdirAll(filepath.Join(testPluqqyDir, "pipelines"), 0755)

	// Change to temp directory so .pluqqy is found there
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	tests := []struct {
		name            string
		setup           func()
		itemType        string
		originalName    string
		originalPath    string
		wantNamePattern string // Use regex pattern for flexibility
	}{
		{
			name: "first clone gets (Copy) prefix",
			setup: func() {
				// No existing files
			},
			itemType:        "component",
			originalName:    "Test Component",
			originalPath:    "components/prompts/test-component.md",
			wantNamePattern: `^\(Copy\) Test Component$`,
		},
		{
			name: "second clone gets (Copy 2) prefix",
			setup: func() {
				// Create the first copy
				os.WriteFile(
					filepath.Join(testPluqqyDir, "components", "prompts", "copy-test-component.md"),
					[]byte("test"),
					0644,
				)
			},
			itemType:        "component",
			originalName:    "Test Component",
			originalPath:    "components/prompts/test-component.md",
			wantNamePattern: `^\(Copy 2\) Test Component$`,
		},
		{
			name: "third clone gets (Copy 3) prefix",
			setup: func() {
				// Create the first two copies
				os.WriteFile(
					filepath.Join(testPluqqyDir, "components", "prompts", "copy-test-component.md"),
					[]byte("test"),
					0644,
				)
				os.WriteFile(
					filepath.Join(testPluqqyDir, "components", "prompts", "copy-2-test-component.md"),
					[]byte("test"),
					0644,
				)
			},
			itemType:        "component",
			originalName:    "Test Component",
			originalPath:    "components/prompts/test-component.md",
			wantNamePattern: `^\(Copy 3\) Test Component$`,
		},
		{
			name: "cloning already cloned item strips old prefix",
			setup: func() {
				// No existing files
			},
			itemType:        "component",
			originalName:    "(Copy 2) Test Component",
			originalPath:    "components/prompts/copy-2-test-component.md",
			wantNamePattern: `^\(Copy\) Test Component$`,
		},
		{
			name: "pipeline clone works the same",
			setup: func() {
				// Create an existing copy
				os.WriteFile(
					filepath.Join(testPluqqyDir, "pipelines", "copy-test-pipeline.yaml"),
					[]byte("test"),
					0644,
				)
			},
			itemType:        "pipeline",
			originalName:    "Test Pipeline",
			originalPath:    "pipelines/test-pipeline.yaml",
			wantNamePattern: `^\(Copy 2\) Test Pipeline$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up test directory
			os.RemoveAll(filepath.Join(testPluqqyDir, "components", "prompts"))
			os.RemoveAll(filepath.Join(testPluqqyDir, "pipelines"))
			os.MkdirAll(filepath.Join(testPluqqyDir, "components", "prompts"), 0755)
			os.MkdirAll(filepath.Join(testPluqqyDir, "pipelines"), 0755)

			// Run setup
			tt.setup()

			// Create clone state and start
			cs := NewCloneState()
			cs.Start(tt.originalName, tt.itemType, tt.originalPath, false)

			// Check the generated name
			if !regexp.MustCompile(tt.wantNamePattern).MatchString(cs.NewName) {
				t.Errorf("Generated name = %q, want pattern %q", cs.NewName, tt.wantNamePattern)
			}

			// Verify no validation error for the auto-generated name
			if cs.ValidationError != "" {
				t.Errorf("Unexpected validation error: %s", cs.ValidationError)
			}
		})
	}
}
