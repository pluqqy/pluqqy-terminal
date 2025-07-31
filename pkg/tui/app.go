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
)

type App struct {
	state          sessionState
	mainList       *MainListModel
	builder        *PipelineBuilderModel
	viewer         *PipelineViewerModel
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
		// Pass window size to all sub-models, accounting for header
		availableHeight := msg.Height - headerHeight
		if a.mainList != nil {
			a.mainList.SetSize(msg.Width, availableHeight)
		}
		if a.builder != nil {
			a.builder.SetSize(msg.Width, availableHeight)
		}
		if a.viewer != nil {
			a.viewer.SetSize(msg.Width, availableHeight)
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

	case SwitchViewMsg:
		// Handle view switching
		switch msg.view {
		case mainListView:
			a.state = mainListView
			if a.mainList == nil {
				a.mainList = NewMainListModel()
			} else {
				// Reload pipelines when returning to list
				a.mainList.loadPipelines()
			}
			return a, a.mainList.Init()
		case pipelineBuilderView:
			a.state = pipelineBuilderView
			// Always create a fresh builder to avoid state issues
			a.builder = NewPipelineBuilderModel()
			
			// If pipeline specified, load it first to get the correct title
			pipelineName := ""
			if msg.pipeline != "" {
				a.builder.SetPipeline(msg.pipeline)
				if a.builder.pipeline != nil {
					pipelineName = a.builder.pipeline.Name
				}
			}
			
			// Calculate header height with the actual title
			title := ""
			if pipelineName != "" {
				title = "Pipeline: " + pipelineName
			}
			header := renderHeader(a.width, title)
			headerHeight := lipgloss.Height(header)
			a.builder.SetSize(a.width, a.height-headerHeight)
			return a, a.builder.Init()
		case pipelineViewerView:
			a.state = pipelineViewerView
			if a.viewer == nil {
				a.viewer = NewPipelineViewerModel()
			}
			// Calculate header height
			header := renderHeader(a.width, "")
			headerHeight := lipgloss.Height(header)
			a.viewer.SetSize(a.width, a.height-headerHeight)
			a.viewer.SetPipeline(msg.pipeline)
			return a, a.viewer.Init()
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
		title = "Pluqqy - Pipeline Manager"
	case pipelineBuilderView:
		if a.builder != nil && a.builder.pipeline != nil {
			title = "Pipeline: " + a.builder.pipeline.Name
		}
	case pipelineViewerView:
		if a.viewer != nil && a.viewer.pipeline != nil {
			title = "Pipeline: " + a.viewer.pipeline.Name
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

// Messages for communication between views
type StatusMsg string

type clearStatusMsg struct{}

type SwitchViewMsg struct {
	view     sessionState
	pipeline string // optional pipeline name for viewer/builder
}

// Models are implemented in separate files