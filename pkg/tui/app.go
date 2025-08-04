package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	mainListView sessionState = iota
	pipelineBuilderView
	pipelineViewerView
	settingsEditorView
)

type App struct {
	state          sessionState
	mainList       *MainListModel
	builder        *PipelineBuilderModel
	viewer         *PipelineViewerModel
	settingsEditor *SettingsEditorModel
	width          int
	height         int
	statusMsg      string
	statusTimer    *time.Timer
	quitConfirm    bool
}

func NewApp() *App {
	return &App{
		state:    mainListView,
		mainList: NewMainListModel(),
	}
}

func (a *App) Init() tea.Cmd {
	return a.mainList.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Calculate header height (no title for size calculation)
		header := renderHeader(a.width, "")
		headerHeight := lipgloss.Height(header)
		// Pass window size to all sub-models, accounting for header and status bar
		// Always reserve 1 line for status message to prevent layout shifts
		availableHeight := msg.Height - headerHeight - 1
		if a.mainList != nil {
			a.mainList.SetSize(msg.Width, availableHeight)
		}
		if a.builder != nil {
			a.builder.SetSize(msg.Width, availableHeight)
		}
		if a.viewer != nil {
			a.viewer.SetSize(msg.Width, availableHeight)
		}
		if a.settingsEditor != nil {
			a.settingsEditor.SetSize(msg.Width, availableHeight)
		}

	case tea.KeyMsg:
		// Global keybindings
		if msg.String() == "esc" && a.statusMsg != "" {
			// Clear global status message
			a.statusMsg = ""
			if a.statusTimer != nil {
				a.statusTimer.Stop()
			}
			// Force a redraw to ensure layout is recalculated
			return a, tea.ClearScreen
		}
		if msg.Type == tea.KeyCtrlC {
			if a.quitConfirm {
				// Second Ctrl+C, actually quit
				return a, tea.Quit
			}
			// First Ctrl+C, show confirmation
			a.quitConfirm = true
			a.statusMsg = "Press Ctrl+C again to quit"
			// Cancel any existing timer
			if a.statusTimer != nil {
				a.statusTimer.Stop()
			}
			// Set timer to clear status and quit confirm after 2 seconds
			a.statusTimer = time.NewTimer(2 * time.Second)
			return a, func() tea.Msg {
				<-a.statusTimer.C
				a.quitConfirm = false
				return clearStatusMsg{}
			}
		}
		// Any other key cancels quit confirmation
		if a.quitConfirm {
			a.quitConfirm = false
		}

	case StatusMsg:
		a.statusMsg = string(msg)
		// Cancel any existing timer
		if a.statusTimer != nil {
			a.statusTimer.Stop()
		}
		// Set timer to clear status after 3 seconds
		a.statusTimer = time.NewTimer(3 * time.Second)
		return a, func() tea.Msg {
			<-a.statusTimer.C
			return clearStatusMsg{}
		}

	case clearStatusMsg:
		a.statusMsg = ""
		// Force a redraw to ensure layout is recalculated
		return a, tea.ClearScreen

	case settingsSavedMsg:
		// Handle settings saved - reload components in main list
		if a.mainList != nil {
			a.mainList.reloadComponents()
		}
		// Switch to main list view
		a.state = mainListView
		// Set the status message
		if msg.switchMsg.status != "" {
			a.statusMsg = msg.switchMsg.status
			// Set timer to clear status after 3 seconds
			if a.statusTimer != nil {
				a.statusTimer.Stop()
			}
			a.statusTimer = time.NewTimer(3 * time.Second)
			return a, func() tea.Msg {
				<-a.statusTimer.C
				return clearStatusMsg{}
			}
		}
		return a, a.mainList.Init()

	case SwitchViewMsg:
		// Handle view switching
		switch msg.view {
		case mainListView:
			a.state = mainListView
			if a.mainList == nil {
				a.mainList = NewMainListModel()
			} else {
				// Reload both components and pipelines when returning to list
				a.mainList.reloadComponents()
				a.mainList.loadPipelines()
				// Re-run search if active
				if a.mainList.searchQuery != "" {
					a.mainList.performSearch()
				}
			}
			return a, a.mainList.Init()
		case pipelineBuilderView:
			a.state = pipelineBuilderView
			// Always create a fresh builder to avoid state issues
			a.builder = NewPipelineBuilderModel()

			// If pipeline specified, load it
			if msg.pipeline != "" {
				a.builder.SetPipeline(msg.pipeline)
			}

			// Set size if we have dimensions (header height already accounted for in WindowSizeMsg)
			if a.width > 0 && a.height > 0 {
				header := renderHeader(a.width, "")
				headerHeight := lipgloss.Height(header)
				a.builder.SetSize(a.width, a.height-headerHeight-1) // -1 for status bar
			}
			return a, a.builder.Init()
		case pipelineViewerView:
			a.state = pipelineViewerView
			if a.viewer == nil {
				a.viewer = NewPipelineViewerModel()
			}
			// Set size if we have dimensions (header height already accounted for in WindowSizeMsg)
			if a.width > 0 && a.height > 0 {
				header := renderHeader(a.width, "")
				headerHeight := lipgloss.Height(header)
				a.viewer.SetSize(a.width, a.height-headerHeight-1) // -1 for status bar
			}
			a.viewer.SetPipeline(msg.pipeline)
			return a, a.viewer.Init()
		case settingsEditorView:
			a.state = settingsEditorView
			if a.settingsEditor == nil {
				a.settingsEditor = NewSettingsEditorModel()
			}
			// Set size if we have dimensions (header height already accounted for in WindowSizeMsg)
			if a.width > 0 && a.height > 0 {
				header := renderHeader(a.width, "")
				headerHeight := lipgloss.Height(header)
				a.settingsEditor.SetSize(a.width, a.height-headerHeight-1) // -1 for status bar
			}
			return a, a.settingsEditor.Init()
		}

		// Handle status message from view switch
		if msg.status != "" {
			a.statusMsg = msg.status
			// Set timer to clear status after 3 seconds
			if a.statusTimer != nil {
				a.statusTimer.Stop()
			}
			a.statusTimer = time.NewTimer(3 * time.Second)
			return a, func() tea.Msg {
				<-a.statusTimer.C
				return clearStatusMsg{}
			}
		}
	}

	// Route updates to the active view
	var cmd tea.Cmd
	switch a.state {
	case mainListView:
		var m tea.Model
		m, cmd = a.mainList.Update(msg)
		if ml, ok := m.(*MainListModel); ok {
			a.mainList = ml
		}
	case pipelineBuilderView:
		var m tea.Model
		m, cmd = a.builder.Update(msg)
		if pb, ok := m.(*PipelineBuilderModel); ok {
			a.builder = pb
		}
	case pipelineViewerView:
		var m tea.Model
		m, cmd = a.viewer.Update(msg)
		if pv, ok := m.(*PipelineViewerModel); ok {
			a.viewer = pv
		}
	case settingsEditorView:
		var m tea.Model
		m, cmd = a.settingsEditor.Update(msg)
		if se, ok := m.(*SettingsEditorModel); ok {
			a.settingsEditor = se
		}
	}

	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Determine the title based on current view
	var title string
	switch a.state {
	case mainListView:
		if a.mainList != nil {
			if a.mainList.componentCreator.IsActive() {
				if a.mainList.componentCreator.GetComponentType() != "" {
					// Capitalize the component type
					componentType := a.mainList.componentCreator.GetComponentType()
					if len(componentType) > 0 {
						componentType = strings.ToUpper(componentType[:1]) + componentType[1:]
					}
					title = componentType + ": New"
				} else {
					title = "Component: New"
				}
			} else if a.mainList.componentEditor.IsActive() && a.mainList.componentEditor.ComponentName != "" {
				title = "Component: " + a.mainList.componentEditor.ComponentName
			} else if a.mainList.tagEditor.Active {
				title = "Tag Editor"
			} else {
				title = "✦ Pluqqy ✦ Dashboard ✦"
			}
		} else {
			title = "✦ Welcome to ✦ Pluqqy ✦"
		}
	case pipelineBuilderView:
		if a.builder != nil {
			if a.builder.editingTags {
				title = "Tag Editor"
			} else if a.builder.pipeline != nil {
				title = "Pipeline: " + a.builder.pipeline.Name
			}
		}
	case pipelineViewerView:
		if a.viewer != nil && a.viewer.pipeline != nil {
			title = "Pipeline: " + a.viewer.pipeline.Name
		}
	case settingsEditorView:
		title = "Settings"
		if a.settingsEditor != nil && a.settingsEditor.hasChanges {
			title = "Settings (modified)"
		}
	}

	// Render the header with title
	header := renderHeader(a.width, title)

	var content string
	switch a.state {
	case mainListView:
		content = a.mainList.View()
	case pipelineBuilderView:
		content = a.builder.View()
	case pipelineViewerView:
		content = a.viewer.View()
	case settingsEditorView:
		content = a.settingsEditor.View()
	default:
		content = "Unknown view"
	}

	// Combine header with content
	fullContent := lipgloss.JoinVertical(lipgloss.Top, header, content)

	// Add status bar if there's a message
	if a.statusMsg != "" {
		// Create a footer-style status bar
		statusStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("82")). // Green for success
			Width(a.width).
			Align(lipgloss.Center).
			Padding(0, 1)

		statusBar := statusStyle.Render(a.statusMsg)

		// Position at the bottom
		totalHeight := lipgloss.Height(fullContent)
		remainingHeight := a.height - totalHeight - 1
		if remainingHeight > 0 {
			fullContent = lipgloss.JoinVertical(
				lipgloss.Top,
				fullContent,
				lipgloss.NewStyle().Height(remainingHeight).Render(""),
				statusBar,
			)
		} else {
			fullContent = lipgloss.JoinVertical(lipgloss.Top, fullContent, statusBar)
		}
	}

	return fullContent
}

// formatHelpText formats help items with styled shortcut keys
func formatHelpText(items []string) string {
	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")) // Brighter grey for shortcuts
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Darker grey for descriptions

	formatted := make([]string, len(items))
	for i, item := range items {
		// Find the first space to separate shortcut from description
		firstSpace := strings.Index(item, " ")
		if firstSpace > 0 && firstSpace < len(item)-1 {
			shortcut := item[:firstSpace]
			desc := item[firstSpace+1:]
			formatted[i] = shortcutStyle.Render(shortcut) + " " + descStyle.Render(desc)
		} else {
			formatted[i] = descStyle.Render(item)
		}
	}

	separator := descStyle.Render(" • ")
	return strings.Join(formatted, separator)
}

// formatHelpTextRows formats help items in multiple rows with right alignment
func formatHelpTextRows(rows [][]string, width int) string {
	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")) // Brighter grey for shortcuts
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Darker grey for descriptions
	separator := descStyle.Render(" • ")

	var lines []string
	for _, row := range rows {
		formatted := make([]string, len(row))
		for i, item := range row {
			// Find the first space to separate shortcut from description
			firstSpace := strings.Index(item, " ")
			if firstSpace > 0 && firstSpace < len(item)-1 {
				shortcut := item[:firstSpace]
				desc := item[firstSpace+1:]
				formatted[i] = shortcutStyle.Render(shortcut) + " " + descStyle.Render(desc)
			} else {
				formatted[i] = descStyle.Render(item)
			}
		}

		rowText := strings.Join(formatted, separator)
		// Right-align the row
		rowWidth := lipgloss.Width(rowText)
		if rowWidth < width {
			padding := width - rowWidth - 2 // -2 for border padding
			if padding > 0 {
				rowText = strings.Repeat(" ", padding) + rowText
			}
		}
		lines = append(lines, rowText)
	}

	return strings.Join(lines, "\n")
}

// formatConfirmOptions formats Yes/No options with appropriate styling
// For destructive actions, Yes gets red background, No gets green
func formatConfirmOptions(destructive bool) string {
	yesStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(func() string {
			if destructive {
				return "196" // Red for destructive Yes
			}
			return "28" // Green for non-destructive Yes
		}())).
		Foreground(lipgloss.Color("255")). // White text
		Padding(0, 1).
		Bold(true)

	noStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(func() string {
			if destructive {
				return "28" // Green for safe No
			}
			return "196" // Red for destructive No
		}())).
		Foreground(lipgloss.Color("255")). // White text
		Padding(0, 1).
		Bold(true)

	return yesStyle.Render("[Y]es") + "  /  " + noStyle.Render("[N]o")
}

// Messages for communication between views
type StatusMsg string

type clearStatusMsg struct{}

type ReloadMsg struct {
	Message string
}

type SwitchViewMsg struct {
	view     sessionState
	pipeline string // optional pipeline name for viewer/builder
	status   string // optional status message to display
}

// Models are implemented in separate files
