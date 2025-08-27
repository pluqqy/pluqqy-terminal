package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pluqqy/pluqqy-cli/pkg/files"
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// TestMain sets up a clean test environment to prevent .pluqqy folder creation in source directories
func TestMain(m *testing.M) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Create a temporary directory for all tests
	tempDir, err := os.MkdirTemp("", "component-creator-tests-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for test execution
	if err := os.Chdir(tempDir); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Restore original directory
	os.Chdir(originalWd)

	// Exit with test result code
	os.Exit(code)
}

// MockEnhancedEditor implements EnhancedEditorInterface for testing
type MockEnhancedEditor struct {
	Active          bool
	Content         string
	ComponentPath   string
	ComponentName   string
	ComponentType   string
	Tags           []string
	FilePicking    bool
	Width          int
	Height         int
}

func (m *MockEnhancedEditor) IsActive() bool {
	return m.Active
}

func (m *MockEnhancedEditor) GetContent() string {
	return m.Content
}

func (m *MockEnhancedEditor) StartEditing(path, name, componentType, content string, tags []string) {
	m.Active = true
	m.ComponentPath = path
	m.ComponentName = name
	m.ComponentType = componentType
	m.Content = content
	m.Tags = tags
}

func (m *MockEnhancedEditor) SetValue(content string) {
	m.Content = content
}

func (m *MockEnhancedEditor) Focus() {
	// Mock implementation
}

func (m *MockEnhancedEditor) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *MockEnhancedEditor) IsFilePicking() bool {
	return m.FilePicking
}

func (m *MockEnhancedEditor) UpdateFilePicker(msg interface{}) interface{} {
	return msg
}

// Test helper to create a test ComponentCreator with mock editor
func createTestComponentCreator() *ComponentCreator {
	creator := NewComponentCreator(func() {
		// Test reload callback
	})
	
	mockEditor := &MockEnhancedEditor{}
	creator.SetEnhancedEditor(mockEditor)
	
	return creator
}

func TestNewComponentCreator(t *testing.T) {
	creator := NewComponentCreator(func() {
		// Test reload callback
	})

	if creator == nil {
		t.Fatal("NewComponentCreator should not return nil")
	}

	if creator.IsActive() {
		t.Error("New component creator should not be active")
	}

	if creator.GetCurrentStep() != 0 {
		t.Error("New component creator should start at step 0")
	}

	if creator.GetComponentName() != "" {
		t.Error("New component creator should have empty name")
	}

	if creator.GetComponentType() != "" {
		t.Error("New component creator should have empty type")
	}
}

func TestComponentCreator_StateManagement(t *testing.T) {
	creator := createTestComponentCreator()

	// Test initial state
	if creator.IsActive() {
		t.Error("Creator should not be active initially")
	}

	// Test Start
	creator.Start()
	if !creator.IsActive() {
		t.Error("Creator should be active after Start()")
	}

	if creator.GetCurrentStep() != 0 {
		t.Errorf("Expected step 0 after start, got %d", creator.GetCurrentStep())
	}

	if creator.GetTypeCursor() != 0 {
		t.Errorf("Expected cursor 0 after start, got %d", creator.GetTypeCursor())
	}

	// Test Reset
	creator.componentName = "test"
	creator.componentCreationType = models.ComponentTypeContext
	creator.creationStep = 2
	creator.typeCursor = 1
	creator.validationError = "test error"

	creator.Reset()

	if creator.IsActive() {
		t.Error("Creator should not be active after Reset()")
	}

	if creator.GetComponentName() != "" {
		t.Error("Component name should be empty after reset")
	}

	if creator.GetComponentType() != "" {
		t.Error("Component type should be empty after reset")
	}

	if creator.GetCurrentStep() != 0 {
		t.Error("Step should be 0 after reset")
	}

	if creator.GetTypeCursor() != 0 {
		t.Error("Cursor should be 0 after reset")
	}

	if creator.GetValidationError() != "" {
		t.Error("Validation error should be empty after reset")
	}
}

func TestComponentCreator_TypeSelection(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		input          tea.KeyMsg
		expectedType   string
		expectedStep   int
		expectedCursor int
		expectedActive bool
		shouldHandle   bool
	}{
		{
			name:           "select context type with enter",
			initialCursor:  0,
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypeContext,
			expectedStep:   1,
			expectedCursor: 0,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "select prompt type with enter",
			initialCursor:  1,
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypePrompt,
			expectedStep:   1,
			expectedCursor: 1,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "select rules type with enter",
			initialCursor:  2,
			input:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedType:   models.ComponentTypeRules,
			expectedStep:   1,
			expectedCursor: 2,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "navigate down with j",
			initialCursor:  0,
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 1,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "navigate down with arrow key",
			initialCursor:  1,
			input:          tea.KeyMsg{Type: tea.KeyDown},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 2,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "navigate up with k",
			initialCursor:  2,
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 1,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "navigate up with arrow key",
			initialCursor:  1,
			input:          tea.KeyMsg{Type: tea.KeyUp},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "boundary check - stay at top",
			initialCursor:  0,
			input:          tea.KeyMsg{Type: tea.KeyUp},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "boundary check - stay at bottom",
			initialCursor:  2,
			input:          tea.KeyMsg{Type: tea.KeyDown},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 2,
			expectedActive: true,
			shouldHandle:   true,
		},
		{
			name:           "cancel with escape",
			initialCursor:  1,
			input:          tea.KeyMsg{Type: tea.KeyEsc},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 1, // cursor position preserved
			expectedActive: false,
			shouldHandle:   true,
		},
		{
			name:           "unhandled key",
			initialCursor:  0,
			input:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			expectedType:   "",
			expectedStep:   0,
			expectedCursor: 0,
			expectedActive: true,
			shouldHandle:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := createTestComponentCreator()
			creator.Start()
			creator.typeCursor = tt.initialCursor

			handled := creator.HandleTypeSelection(tt.input)

			if handled != tt.shouldHandle {
				t.Errorf("Expected handled=%v, got %v", tt.shouldHandle, handled)
			}

			if creator.GetComponentType() != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, creator.GetComponentType())
			}

			if creator.GetCurrentStep() != tt.expectedStep {
				t.Errorf("Expected step %d, got %d", tt.expectedStep, creator.GetCurrentStep())
			}

			if creator.GetTypeCursor() != tt.expectedCursor {
				t.Errorf("Expected cursor %d, got %d", tt.expectedCursor, creator.GetTypeCursor())
			}

			if creator.IsActive() != tt.expectedActive {
				t.Errorf("Expected active=%v, got %v", tt.expectedActive, creator.IsActive())
			}
		})
	}
}

func TestComponentCreator_NameInput(t *testing.T) {
	tests := []struct {
		name              string
		initialName       string
		input             tea.KeyMsg
		expectedName      string
		expectedStep      int
		expectError       bool
		expectedError     string
		shouldHandle      bool
	}{
		{
			name:         "add character",
			initialName:  "Test",
			input:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}},
			expectedName: "Test!",
			expectedStep: 1,
			shouldHandle: true,
		},
		{
			name:         "add space",
			initialName:  "Test",
			input:        tea.KeyMsg{Type: tea.KeySpace},
			expectedName: "Test ",
			expectedStep: 1,
			shouldHandle: true,
		},
		{
			name:         "backspace removes character",
			initialName:  "Test",
			input:        tea.KeyMsg{Type: tea.KeyBackspace},
			expectedName: "Tes",
			expectedStep: 1,
			shouldHandle: true,
		},
		{
			name:         "backspace on empty string",
			initialName:  "",
			input:        tea.KeyMsg{Type: tea.KeyBackspace},
			expectedName: "",
			expectedStep: 1,
			shouldHandle: true,
		},
		{
			name:         "escape goes back to type selection",
			initialName:  "Test",
			input:        tea.KeyMsg{Type: tea.KeyEsc},
			expectedName: "",
			expectedStep: 0,
			shouldHandle: true,
		},
		{
			name:          "enter with empty name stays at step 1",
			initialName:   "",
			input:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedName:  "",
			expectedStep:  1,
			shouldHandle:  true,
		},
		{
			name:         "enter with valid name advances to step 2",
			initialName:  "Valid Name",
			input:        tea.KeyMsg{Type: tea.KeyEnter},
			expectedName: "Valid Name",
			expectedStep: 2,
			shouldHandle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := createTestComponentCreator()
			creator.Start()
			creator.componentCreationType = models.ComponentTypeContext
			creator.creationStep = 1
			creator.componentName = tt.initialName

			handled := creator.HandleNameInput(tt.input)

			if handled != tt.shouldHandle {
				t.Errorf("Expected handled=%v, got %v", tt.shouldHandle, handled)
			}

			if creator.GetComponentName() != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, creator.GetComponentName())
			}

			if creator.GetCurrentStep() != tt.expectedStep {
				t.Errorf("Expected step %d, got %d", tt.expectedStep, creator.GetCurrentStep())
			}

			if tt.expectError {
				if creator.GetValidationError() == "" {
					t.Error("Expected validation error but got none")
				}
				if tt.expectedError != "" && !strings.Contains(creator.GetValidationError(), tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, creator.GetValidationError())
				}
			} else if creator.GetValidationError() != "" {
				t.Errorf("Unexpected validation error: %s", creator.GetValidationError())
			}
		})
	}
}

func TestComponentCreator_NameValidation(t *testing.T) {
	tests := []struct {
		name           string
		componentName  string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "valid name",
			componentName: "Valid Component Name",
			expectError:   false,
		},
		{
			name:          "empty name",
			componentName: "",
			expectError:   true,
			errorContains: "empty",
		},
		{
			name:          "whitespace only name",
			componentName: "   ",
			expectError:   true,
			errorContains: "empty",
		},
		{
			name:          "name with .md extension",
			componentName: "component.md",
			expectError:   true,
			errorContains: ".md extension",
		},
		{
			name:          "name with mixed case .md extension",
			componentName: "Component.MD",
			expectError:   true,
			errorContains: ".md extension",
		},
		{
			name:          "name with special characters",
			componentName: "Test Component @#$%",
			expectError:   true, // Special characters not allowed by CLI validator
			errorContains: "invalid character",
		},
		{
			name:          "unicode name",
			componentName: "测试组件",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := createTestComponentCreator()
			err := creator.validateComponentName(tt.componentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for name '%s' but got none", tt.componentName)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for name '%s': %s", tt.componentName, err.Error())
				}
			}
		})
	}
}

// Note: Duplicate checking tests are omitted here since they require file system setup.
// The duplicate checking logic is tested through integration tests in the main TUI package.

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "Simple Name",
			expected: "simple-name",
		},
		{
			name:     "name with special characters",
			input:    "Name @#$% Special!",
			expected: "name-special",
		},
		{
			name:     "name with numbers",
			input:    "Component 123",
			expected: "component-123",
		},
		{
			name:     "name with multiple spaces",
			input:    "Multiple   Spaces   Here",
			expected: "multiple-spaces-here",
		},
		{
			name:     "name with hyphens",
			input:    "Already-Has-Hyphens",
			expected: "already-has-hyphens",
		},
		{
			name:     "name with leading/trailing spaces",
			input:    "  Trim Spaces  ",
			expected: "trim-spaces",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "untitled",
		},
		{
			name:     "only special characters",
			input:    "@#$%^&*()",
			expected: "untitled",
		},
		{
			name:     "unicode characters",
			input:    "测试组件 Test",
			expected: "test",
		},
		{
			name:     "consecutive hyphens",
			input:    "Multiple---Hyphens",
			expected: "multiple-hyphens",
		},
		{
			name:     "leading/trailing hyphens",
			input:    "-Start And End-",
			expected: "start-and-end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestComponentCreator_EnhancedEditorIntegration(t *testing.T) {
	creator := createTestComponentCreator()
	mockEditor := creator.enhancedEditor.(*MockEnhancedEditor)

	// Test that enhanced editor is not active initially
	if creator.IsEnhancedEditorActive() {
		t.Error("Enhanced editor should not be active initially")
	}

	// Setup component creation
	creator.Start()
	creator.componentCreationType = models.ComponentTypeContext
	creator.componentName = "Test Component"
	creator.creationStep = 2

	// Initialize enhanced editor
	creator.initializeEnhancedEditor()

	// Check enhanced editor was activated correctly
	if !creator.IsEnhancedEditorActive() {
		t.Error("Enhanced editor should be active after initialization")
	}

	if !mockEditor.Active {
		t.Error("Mock editor should be active")
	}

	if mockEditor.ComponentName != "Test Component" {
		t.Errorf("Expected component name 'Test Component', got '%s'", mockEditor.ComponentName)
	}

	if mockEditor.ComponentType != models.ComponentTypeContext {
		t.Errorf("Expected component type '%s', got '%s'", models.ComponentTypeContext, mockEditor.ComponentType)
	}

	expectedPath := filepath.Join(files.ComponentsDir, models.ComponentTypeContext, "test-component.md")
	if mockEditor.ComponentPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, mockEditor.ComponentPath)
	}

	if mockEditor.Width != 80 || mockEditor.Height != 20 {
		t.Errorf("Expected size 80x20, got %dx%d", mockEditor.Width, mockEditor.Height)
	}

	// Test content retrieval
	mockEditor.Content = "Test content"
	if creator.GetComponentContent() != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", creator.GetComponentContent())
	}
}

func TestComponentCreator_SaveSuccessHandling(t *testing.T) {
	var reloadCalled bool
	creator := NewComponentCreator(func() {
		reloadCalled = true
	})

	// Initially no save should be successful
	if creator.WasSaveSuccessful() {
		t.Error("No save should be successful initially")
	}

	// Mark save as successful
	creator.MarkSaveSuccessful()

	// Check that save was successful (and triggers reload)
	if !creator.WasSaveSuccessful() {
		t.Error("Save should be marked as successful")
	}

	// Verify reload callback was called
	if !reloadCalled {
		t.Error("Reload callback should have been called")
	}

	// Second call should return false (flag reset)
	if creator.WasSaveSuccessful() {
		t.Error("WasSaveSuccessful should reset flag after first call")
	}
}

func TestComponentCreator_StatusMessage(t *testing.T) {
	creator := createTestComponentCreator()
	creator.componentName = "Test Component"
	creator.componentCreationType = models.ComponentTypeContext

	expected := "✓ Created contexts: test-component.md"
	actual := creator.GetStatusMessage()

	if actual != expected {
		t.Errorf("Expected status message '%s', got '%s'", expected, actual)
	}
}

func TestComponentCreator_FullWorkflow(t *testing.T) {
	creator := createTestComponentCreator()
	mockEditor := creator.enhancedEditor.(*MockEnhancedEditor)

	// Step 1: Start creation
	creator.Start()
	if !creator.IsActive() {
		t.Fatal("Creator should be active after start")
	}
	if creator.GetCurrentStep() != 0 {
		t.Fatal("Should start at step 0 (type selection)")
	}

	// Step 2: Select component type
	handled := creator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Select first type (context)
	if !handled {
		t.Fatal("Type selection should be handled")
	}
	if creator.GetCurrentStep() != 1 {
		t.Fatal("Should advance to step 1 (name input)")
	}
	if creator.GetComponentType() != models.ComponentTypeContext {
		t.Fatal("Should have selected context type")
	}

	// Step 3: Enter component name
	nameChars := []rune("Test Component")
	for _, char := range nameChars {
		creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	if creator.GetComponentName() != "Test Component" {
		t.Fatalf("Expected name 'Test Component', got '%s'", creator.GetComponentName())
	}

	// Step 4: Confirm name and move to content
	handled = creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled {
		t.Fatal("Name confirmation should be handled")
	}
	if creator.GetCurrentStep() != 2 {
		t.Fatal("Should advance to step 2 (content editing)")
	}

	// Step 5: Verify enhanced editor is initialized
	if !creator.IsEnhancedEditorActive() {
		t.Fatal("Enhanced editor should be active")
	}
	if !mockEditor.Active {
		t.Fatal("Mock editor should be active")
	}

	// Step 6: Add content and simulate save
	mockEditor.Content = "This is the component content"
	if creator.GetComponentContent() != "This is the component content" {
		t.Fatal("Content should be retrievable from enhanced editor")
	}

	// Step 7: Mark save as successful
	creator.MarkSaveSuccessful()
	if !creator.WasSaveSuccessful() {
		t.Fatal("Save should be marked as successful")
	}

	// Verify final state
	expectedStatus := "✓ Created contexts: test-component.md"
	if creator.GetStatusMessage() != expectedStatus {
		t.Errorf("Expected status '%s', got '%s'", expectedStatus, creator.GetStatusMessage())
	}
}

func TestComponentCreator_ErrorRecovery(t *testing.T) {
	creator := createTestComponentCreator()

	// Start creation process
	creator.Start()
	creator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Select context

	// Try to create component with invalid name (.md extension)
	creator.componentName = "invalid.md"
	handled := creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	if !handled {
		t.Fatal("Invalid name input should be handled")
	}

	if creator.GetCurrentStep() != 1 {
		t.Error("Should stay at step 1 due to validation error")
	}

	if creator.GetValidationError() == "" {
		t.Error("Should have validation error for .md extension")
	}

	if !strings.Contains(creator.GetValidationError(), ".md extension") {
		t.Error("Validation error should mention .md extension")
	}

	// Fix the name and continue
	creator.componentName = "valid-name"
	handled = creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEnter})

	if !handled {
		t.Fatal("Valid name input should be handled")
	}

	if creator.GetCurrentStep() != 2 {
		t.Error("Should advance to step 2 with valid name")
	}

	if creator.GetValidationError() != "" {
		t.Error("Validation error should be cleared with valid name")
	}
}

func TestComponentCreator_Cancellation(t *testing.T) {
	creator := createTestComponentCreator()

	// Test cancellation from type selection
	creator.Start()
	if !creator.IsActive() {
		t.Fatal("Should be active after start")
	}

	handled := creator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEsc})
	if !handled {
		t.Fatal("Escape should be handled")
	}

	if creator.IsActive() {
		t.Error("Should not be active after escape from type selection")
	}

	// Test cancellation from name input (goes back to type selection)
	creator.Start()
	creator.HandleTypeSelection(tea.KeyMsg{Type: tea.KeyEnter}) // Select type
	if creator.GetCurrentStep() != 1 {
		t.Fatal("Should be at name input step")
	}

	creator.componentName = "Test Name"
	handled = creator.HandleNameInput(tea.KeyMsg{Type: tea.KeyEsc})
	if !handled {
		t.Fatal("Escape should be handled")
	}

	if creator.GetCurrentStep() != 0 {
		t.Error("Should go back to type selection step")
	}

	if creator.GetComponentName() != "" {
		t.Error("Component name should be cleared")
	}

	if creator.GetValidationError() != "" {
		t.Error("Validation error should be cleared")
	}
}

// Benchmark tests for performance-critical operations
func BenchmarkSanitizeFileName(b *testing.B) {
	testName := "Test Component With Special Characters @#$%^&*()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeFileName(testName)
	}
}

func BenchmarkComponentCreator_TypeSelection(b *testing.B) {
	creator := createTestComponentCreator()
	creator.Start()
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		creator.creationStep = 0 // Reset to type selection
		creator.HandleTypeSelection(enterKey)
	}
}

func BenchmarkComponentCreator_NameValidation(b *testing.B) {
	creator := createTestComponentCreator()
	testName := "Valid Component Name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		creator.validateComponentName(testName)
	}
}