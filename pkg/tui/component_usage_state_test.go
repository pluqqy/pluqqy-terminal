package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestComponentUsageState_Start(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *ComponentUsageState
		component componentItem
		wantActive bool
	}{
		{
			name: "starts with component and becomes active",
			setup: func() *ComponentUsageState {
				return NewComponentUsageState()
			},
			component: componentItem{
				name:     "test-component",
				path:     ".pluqqy/components/contexts/test.md",
				compType: "context",
			},
			wantActive: true,
		},
		{
			name: "starts with empty component",
			setup: func() *ComponentUsageState {
				return NewComponentUsageState()
			},
			component:  componentItem{},
			wantActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup()
			state.Start(tt.component)
			
			assert.Equal(t, tt.wantActive, state.Active)
			assert.Equal(t, tt.component, state.SelectedComponent)
			// PipelinesUsingComponent is loaded asynchronously, may be nil initially
		})
	}
}

func TestComponentUsageState_Stop(t *testing.T) {
	state := NewComponentUsageState()
	state.Start(componentItem{name: "test"})
	
	assert.True(t, state.Active)
	
	state.Stop()
	
	assert.False(t, state.Active)
	assert.Equal(t, 0, state.SelectedIndex)
	assert.Equal(t, 0, state.ScrollOffset)
}

func TestComponentUsageState_HandleInput(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *ComponentUsageState
		input       tea.KeyMsg
		wantHandled bool
		wantActive  bool
		checkState  func(t *testing.T, state *ComponentUsageState)
	}{
		{
			name: "escape key closes modal",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyEsc},
			wantHandled: true,
			wantActive:  false,
		},
		{
			name: "q key closes modal",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantHandled: true,
			wantActive:  false,
		},
		{
			name: "u key closes modal",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}},
			wantHandled: true,
			wantActive:  false,
		},
		{
			name: "up arrow moves selection up",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
					{Name: "Pipeline 3"},
				}
				state.SelectedIndex = 2
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyUp},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 1, state.SelectedIndex)
			},
		},
		{
			name: "up arrow at top doesn't move",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
				}
				state.SelectedIndex = 0
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyUp},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 0, state.SelectedIndex)
			},
		},
		{
			name: "down arrow moves selection down",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
					{Name: "Pipeline 3"},
				}
				state.SelectedIndex = 1
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyDown},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 2, state.SelectedIndex)
			},
		},
		{
			name: "down arrow at bottom doesn't move",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
				}
				state.SelectedIndex = 1
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyDown},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 1, state.SelectedIndex)
			},
		},
		{
			name: "home key goes to top",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
					{Name: "Pipeline 3"},
				}
				state.SelectedIndex = 2
				state.ScrollOffset = 1
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyHome},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 0, state.SelectedIndex)
				assert.Equal(t, 0, state.ScrollOffset)
			},
		},
		{
			name: "g key goes to top",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
				}
				state.SelectedIndex = 1
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 0, state.SelectedIndex)
			},
		},
		{
			name: "end key goes to bottom",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
					{Name: "Pipeline 3"},
				}
				state.SelectedIndex = 0
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyEnd},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 2, state.SelectedIndex)
			},
		},
		{
			name: "G key goes to bottom",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = []PipelineUsageInfo{
					{Name: "Pipeline 1"},
					{Name: "Pipeline 2"},
				}
				state.SelectedIndex = 0
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				assert.Equal(t, 1, state.SelectedIndex)
			},
		},
		{
			name: "page up scrolls up",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = make([]PipelineUsageInfo, 20)
				for i := range state.PipelinesUsingComponent {
					state.PipelinesUsingComponent[i] = PipelineUsageInfo{Name: "Pipeline"}
				}
				state.SelectedIndex = 15
				state.Height = 10 // Small height to test scrolling
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyPgUp},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				// Should move up by visible lines
				assert.Less(t, state.SelectedIndex, 15)
			},
		},
		{
			name: "page down scrolls down",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				state.PipelinesUsingComponent = make([]PipelineUsageInfo, 20)
				for i := range state.PipelinesUsingComponent {
					state.PipelinesUsingComponent[i] = PipelineUsageInfo{Name: "Pipeline"}
				}
				state.SelectedIndex = 5
				state.Height = 10 // Small height to test scrolling
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyPgDown},
			wantHandled: true,
			wantActive:  true,
			checkState: func(t *testing.T, state *ComponentUsageState) {
				// Should move down by visible lines
				assert.Greater(t, state.SelectedIndex, 5)
			},
		},
		{
			name: "unhandled key when inactive",
			setup: func() *ComponentUsageState {
				return NewComponentUsageState() // Not started
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			wantHandled: false,
			wantActive:  false,
		},
		{
			name: "unhandled key when active",
			setup: func() *ComponentUsageState {
				state := NewComponentUsageState()
				state.Start(componentItem{name: "test"})
				return state
			},
			input:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			wantHandled: false,
			wantActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup()
			handled, _ := state.HandleInput(tt.input)
			
			assert.Equal(t, tt.wantHandled, handled)
			assert.Equal(t, tt.wantActive, state.Active)
			
			if tt.checkState != nil {
				tt.checkState(t, state)
			}
		})
	}
}

func TestComponentUsageState_SetSize(t *testing.T) {
	state := NewComponentUsageState()
	
	state.SetSize(100, 50)
	
	assert.Equal(t, 100, state.Width)
	assert.Equal(t, 50, state.Height)
}

func TestComponentUsageState_GetVisibleLines(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "normal height",
			height: 30,
			want:   20, // height - 10 (reserved for UI)
		},
		{
			name:   "small height",
			height: 10,
			want:   1, // minimum 1 (10 - 10 = 0, but min is 1)
		},
		{
			name:   "very small height",
			height: 5,
			want:   1, // minimum 1
		},
		{
			name:   "large height",
			height: 100,
			want:   90, // height - 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewComponentUsageState()
			state.Height = tt.height
			
			got := state.getVisibleLines()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComponentUsageState_EnsureSelectedVisible(t *testing.T) {
	tests := []struct {
		name              string
		selectedIndex     int
		scrollOffset      int
		visibleLines      int
		totalItems        int
		wantScrollOffset  int
	}{
		{
			name:             "selected is visible - no scroll needed",
			selectedIndex:    2,
			scrollOffset:     0,
			visibleLines:     5,
			totalItems:       10,
			wantScrollOffset: 0,
		},
		{
			name:             "selected below visible - scroll down",
			selectedIndex:    8,
			scrollOffset:     0,
			visibleLines:     5,
			totalItems:       10,
			wantScrollOffset: 4, // 8 - 5 + 1
		},
		{
			name:             "selected above visible - scroll up",
			selectedIndex:    1,
			scrollOffset:     5,
			visibleLines:     3,
			totalItems:       10,
			wantScrollOffset: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewComponentUsageState()
			state.SelectedIndex = tt.selectedIndex
			state.ScrollOffset = tt.scrollOffset
			state.Height = tt.visibleLines + 10 // Account for UI overhead
			
			// Create pipelines for testing
			state.PipelinesUsingComponent = make([]PipelineUsageInfo, tt.totalItems)
			
			state.ensureSelectedVisible()
			
			assert.Equal(t, tt.wantScrollOffset, state.ScrollOffset)
		})
	}
}

func TestComponentUsageState_Navigation(t *testing.T) {
	// Test navigation with empty list
	t.Run("navigation with empty list", func(t *testing.T) {
		state := NewComponentUsageState()
		state.Start(componentItem{name: "test"})
		state.PipelinesUsingComponent = []PipelineUsageInfo{}
		
		// Should not crash with empty list
		state.HandleInput(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 0, state.SelectedIndex)
		
		state.HandleInput(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, state.SelectedIndex)
		
		state.HandleInput(tea.KeyMsg{Type: tea.KeyEnd})
		assert.Equal(t, 0, state.SelectedIndex)
	})
	
	// Test j/k navigation
	t.Run("j/k navigation", func(t *testing.T) {
		state := NewComponentUsageState()
		state.Start(componentItem{name: "test"})
		state.PipelinesUsingComponent = []PipelineUsageInfo{
			{Name: "Pipeline 1"},
			{Name: "Pipeline 2"},
			{Name: "Pipeline 3"},
		}
		
		// j moves down
		state.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		assert.Equal(t, 1, state.SelectedIndex)
		
		// k moves up
		state.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		assert.Equal(t, 0, state.SelectedIndex)
	})
}