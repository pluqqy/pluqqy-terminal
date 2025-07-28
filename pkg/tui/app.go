package tui

import (
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
	state       sessionState
	mainList    *MainListModel
	builder     *PipelineBuilderModel
	viewer      *PipelineViewerModel
	width       int
	height      int
	statusMsg   string
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
		// Pass window size to all sub-models
		if a.mainList != nil {
			a.mainList.SetSize(msg.Width, msg.Height)
		}
		if a.builder != nil {
			a.builder.SetSize(msg.Width, msg.Height)
		}
		if a.viewer != nil {
			a.viewer.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		// Global keybindings
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}

	case StatusMsg:
		a.statusMsg = string(msg)
		return a, nil

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
			if a.builder == nil {
				a.builder = NewPipelineBuilderModel()
			}
			a.builder.SetSize(a.width, a.height)
			a.builder.SetPipeline(msg.pipeline)
			return a, a.builder.Init()
		case pipelineViewerView:
			a.state = pipelineViewerView
			if a.viewer == nil {
				a.viewer = NewPipelineViewerModel()
			}
			a.viewer.SetSize(a.width, a.height)
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

	// Add status bar if there's a message
	if a.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
		
		statusBar := statusStyle.Render(a.statusMsg)
		content = lipgloss.JoinVertical(lipgloss.Top, content, statusBar)
	}

	return content
}

// Messages for communication between views
type StatusMsg string

type SwitchViewMsg struct {
	view     sessionState
	pipeline string // optional pipeline name for viewer/builder
}

// Models are implemented in separate files