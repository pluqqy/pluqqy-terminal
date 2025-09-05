package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// SettingsDataStore manages the settings data and state
type SettingsDataStore struct {
	settings         *models.Settings
	originalSettings *models.Settings // For detecting changes
	hasChanges       bool
	err              error
}

// SettingsUIComponents manages UI-specific components
type SettingsUIComponents struct {
	viewport    viewport.Model
	exitConfirm *ConfirmationModel
}

// SettingsViewportManager manages viewport and layout
type SettingsViewportManager struct {
	width  int
	height int
}

// SettingsFormInputs manages all form input fields
type SettingsFormInputs struct {
	defaultFilenameInput textinput.Model
	exportPathInput      textinput.Model
	outputPathInput      textinput.Model
	showHeadings         bool

	// Section editing inputs
	sectionTypeInput    textinput.Model
	sectionHeadingInput textinput.Model

	// Focus management
	focusIndex  int
	totalFields int
}

// SettingsSectionManager manages section-specific operations
type SettingsSectionManager struct {
	sectionCursor  int
	editingSection bool
}