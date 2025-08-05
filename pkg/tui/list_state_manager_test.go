package tui

import (
	"testing"
)

// Test helpers
func validateStateManager(t *testing.T, sm *StateManager) {
	t.Helper()
	
	// Cursors should be non-negative
	if sm.ComponentCursor < 0 {
		t.Error("ComponentCursor is negative")
	}
	if sm.PipelineCursor < 0 {
		t.Error("PipelineCursor is negative")
	}
	
	// LastDataPane should be valid
	if sm.LastDataPane != componentsPane && sm.LastDataPane != pipelinesPane {
		t.Errorf("Invalid LastDataPane: %d", sm.LastDataPane)
	}
	
	// Only one active pane should be set
	validPanes := []pane{searchPane, componentsPane, pipelinesPane, previewPane}
	foundActive := false
	for _, p := range validPanes {
		if sm.ActivePane == p {
			if foundActive {
				t.Error("Multiple active panes detected")
			}
			foundActive = true
		}
	}
	if !foundActive && sm.ActivePane != nonePane {
		t.Errorf("Invalid ActivePane: %d", sm.ActivePane)
	}
}

func makeStateManager(activePane pane, compCursor, pipeCursor int, showPreview bool) *StateManager {
	sm := NewStateManager()
	sm.ActivePane = activePane
	sm.ComponentCursor = compCursor
	sm.PipelineCursor = pipeCursor
	sm.ShowPreview = showPreview
	return sm
}

// Test NewStateManager
func TestNewStateManager(t *testing.T) {
	sm := NewStateManager()
	
	if sm.ComponentCursor != 0 {
		t.Errorf("ComponentCursor = %d, want 0", sm.ComponentCursor)
	}
	if sm.PipelineCursor != 0 {
		t.Errorf("PipelineCursor = %d, want 0", sm.PipelineCursor)
	}
	if sm.ActivePane != componentsPane {
		t.Errorf("ActivePane = %d, want componentsPane", sm.ActivePane)
	}
	if sm.LastDataPane != componentsPane {
		t.Errorf("LastDataPane = %d, want componentsPane", sm.LastDataPane)
	}
	if !sm.ShowPreview {
		t.Error("ShowPreview should be true by default")
	}
	
	validateStateManager(t, sm)
}

// Test UpdateCounts
func TestUpdateCounts(t *testing.T) {
	tests := []struct {
		name            string
		initialCompCur  int
		initialPipeCur  int
		componentCount  int
		pipelineCount   int
		wantCompCur     int
		wantPipeCur     int
	}{
		{
			name:           "normal update",
			initialCompCur: 2,
			initialPipeCur: 1,
			componentCount: 5,
			pipelineCount:  3,
			wantCompCur:    2,
			wantPipeCur:    1,
		},
		{
			name:           "cursor out of bounds - adjust to last",
			initialCompCur: 10,
			initialPipeCur: 8,
			componentCount: 3,
			pipelineCount:  2,
			wantCompCur:    2,
			wantPipeCur:    1,
		},
		{
			name:           "empty lists",
			initialCompCur: 5,
			initialPipeCur: 3,
			componentCount: 0,
			pipelineCount:  0,
			wantCompCur:    0,
			wantPipeCur:    0,
		},
		{
			name:           "one list empty",
			initialCompCur: 2,
			initialPipeCur: 2,
			componentCount: 4,
			pipelineCount:  0,
			wantCompCur:    2,
			wantPipeCur:    0,
		},
		{
			name:           "single item lists",
			initialCompCur: 5,
			initialPipeCur: 5,
			componentCount: 1,
			pipelineCount:  1,
			wantCompCur:    0,
			wantPipeCur:    0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager()
			sm.ComponentCursor = tt.initialCompCur
			sm.PipelineCursor = tt.initialPipeCur
			
			sm.UpdateCounts(tt.componentCount, tt.pipelineCount)
			
			if sm.ComponentCursor != tt.wantCompCur {
				t.Errorf("ComponentCursor = %d, want %d", sm.ComponentCursor, tt.wantCompCur)
			}
			if sm.PipelineCursor != tt.wantPipeCur {
				t.Errorf("PipelineCursor = %d, want %d", sm.PipelineCursor, tt.wantPipeCur)
			}
			if sm.componentCount != tt.componentCount {
				t.Errorf("componentCount = %d, want %d", sm.componentCount, tt.componentCount)
			}
			if sm.pipelineCount != tt.pipelineCount {
				t.Errorf("pipelineCount = %d, want %d", sm.pipelineCount, tt.pipelineCount)
			}
			
			validateStateManager(t, sm)
		})
	}
}

// Test cursor movement
func TestMoveCursorUp(t *testing.T) {
	tests := []struct {
		name        string
		activePane  pane
		compCursor  int
		pipeCursor  int
		compCount   int
		pipeCount   int
		wantMoved   bool
		wantCompCur int
		wantPipeCur int
	}{
		{
			name:        "move up in components",
			activePane:  componentsPane,
			compCursor:  2,
			pipeCursor:  0,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   true,
			wantCompCur: 1,
			wantPipeCur: 0,
		},
		{
			name:        "at top of components",
			activePane:  componentsPane,
			compCursor:  0,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 0,
			wantPipeCur: 1,
		},
		{
			name:        "move up in pipelines",
			activePane:  pipelinesPane,
			compCursor:  1,
			pipeCursor:  3,
			compCount:   5,
			pipeCount:   5,
			wantMoved:   true,
			wantCompCur: 1,
			wantPipeCur: 2,
		},
		{
			name:        "at top of pipelines",
			activePane:  pipelinesPane,
			compCursor:  2,
			pipeCursor:  0,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 2,
			wantPipeCur: 0,
		},
		{
			name:        "in preview pane",
			activePane:  previewPane,
			compCursor:  2,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 2,
			wantPipeCur: 1,
		},
		{
			name:        "in search pane",
			activePane:  searchPane,
			compCursor:  2,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 2,
			wantPipeCur: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := makeStateManager(tt.activePane, tt.compCursor, tt.pipeCursor, true)
			sm.UpdateCounts(tt.compCount, tt.pipeCount)
			
			moved := sm.MoveCursorUp()
			
			if moved != tt.wantMoved {
				t.Errorf("MoveCursorUp() = %v, want %v", moved, tt.wantMoved)
			}
			if sm.ComponentCursor != tt.wantCompCur {
				t.Errorf("ComponentCursor = %d, want %d", sm.ComponentCursor, tt.wantCompCur)
			}
			if sm.PipelineCursor != tt.wantPipeCur {
				t.Errorf("PipelineCursor = %d, want %d", sm.PipelineCursor, tt.wantPipeCur)
			}
			
			validateStateManager(t, sm)
		})
	}
}

func TestMoveCursorDown(t *testing.T) {
	tests := []struct {
		name        string
		activePane  pane
		compCursor  int
		pipeCursor  int
		compCount   int
		pipeCount   int
		wantMoved   bool
		wantCompCur int
		wantPipeCur int
	}{
		{
			name:        "move down in components",
			activePane:  componentsPane,
			compCursor:  1,
			pipeCursor:  0,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   true,
			wantCompCur: 2,
			wantPipeCur: 0,
		},
		{
			name:        "at bottom of components",
			activePane:  componentsPane,
			compCursor:  4,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 4,
			wantPipeCur: 1,
		},
		{
			name:        "move down in pipelines",
			activePane:  pipelinesPane,
			compCursor:  1,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   5,
			wantMoved:   true,
			wantCompCur: 1,
			wantPipeCur: 2,
		},
		{
			name:        "at bottom of pipelines",
			activePane:  pipelinesPane,
			compCursor:  2,
			pipeCursor:  2,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 2,
			wantPipeCur: 2,
		},
		{
			name:        "in preview pane",
			activePane:  previewPane,
			compCursor:  2,
			pipeCursor:  1,
			compCount:   5,
			pipeCount:   3,
			wantMoved:   false,
			wantCompCur: 2,
			wantPipeCur: 1,
		},
		{
			name:        "empty list",
			activePane:  componentsPane,
			compCursor:  0,
			pipeCursor:  0,
			compCount:   0,
			pipeCount:   0,
			wantMoved:   false,
			wantCompCur: 0,
			wantPipeCur: 0,
		},
		{
			name:        "single item list",
			activePane:  pipelinesPane,
			compCursor:  0,
			pipeCursor:  0,
			compCount:   1,
			pipeCount:   1,
			wantMoved:   false,
			wantCompCur: 0,
			wantPipeCur: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := makeStateManager(tt.activePane, tt.compCursor, tt.pipeCursor, true)
			sm.UpdateCounts(tt.compCount, tt.pipeCount)
			
			moved := sm.MoveCursorDown()
			
			if moved != tt.wantMoved {
				t.Errorf("MoveCursorDown() = %v, want %v", moved, tt.wantMoved)
			}
			if sm.ComponentCursor != tt.wantCompCur {
				t.Errorf("ComponentCursor = %d, want %d", sm.ComponentCursor, tt.wantCompCur)
			}
			if sm.PipelineCursor != tt.wantPipeCur {
				t.Errorf("PipelineCursor = %d, want %d", sm.PipelineCursor, tt.wantPipeCur)
			}
			
			validateStateManager(t, sm)
		})
	}
}

// Test tab navigation
func TestHandleTabNavigation_Forward(t *testing.T) {
	tests := []struct {
		name         string
		startPane    pane
		showPreview  bool
		wantPane     pane
		wantLastData pane
	}{
		// With preview
		{
			name:         "components to pipelines with preview",
			startPane:    componentsPane,
			showPreview:  true,
			wantPane:     pipelinesPane,
			wantLastData: pipelinesPane,
		},
		{
			name:         "pipelines to preview",
			startPane:    pipelinesPane,
			showPreview:  true,
			wantPane:     previewPane,
			wantLastData: pipelinesPane,
		},
		{
			name:         "preview to components",
			startPane:    previewPane,
			showPreview:  true,
			wantPane:     componentsPane,
			wantLastData: componentsPane, // LastDataPane unchanged from preview
		},
		{
			name:         "search to components with preview",
			startPane:    searchPane,
			showPreview:  true,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
		// Without preview
		{
			name:         "components to pipelines no preview",
			startPane:    componentsPane,
			showPreview:  false,
			wantPane:     pipelinesPane,
			wantLastData: pipelinesPane,
		},
		{
			name:         "pipelines to components no preview",
			startPane:    pipelinesPane,
			showPreview:  false,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
		{
			name:         "search to components no preview",
			startPane:    searchPane,
			showPreview:  false,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager()
			sm.ActivePane = tt.startPane
			sm.ShowPreview = tt.showPreview
			// Set initial LastDataPane
			if tt.startPane == componentsPane || tt.startPane == pipelinesPane {
				sm.LastDataPane = tt.startPane
			}
			
			sm.HandleTabNavigation(false)
			
			if sm.ActivePane != tt.wantPane {
				t.Errorf("ActivePane = %v, want %v", sm.ActivePane, tt.wantPane)
			}
			if sm.LastDataPane != tt.wantLastData {
				t.Errorf("LastDataPane = %v, want %v", sm.LastDataPane, tt.wantLastData)
			}
			
			validateStateManager(t, sm)
		})
	}
}

func TestHandleTabNavigation_Reverse(t *testing.T) {
	tests := []struct {
		name         string
		startPane    pane
		showPreview  bool
		wantPane     pane
		wantLastData pane
	}{
		// With preview
		{
			name:         "components to preview (reverse)",
			startPane:    componentsPane,
			showPreview:  true,
			wantPane:     previewPane,
			wantLastData: componentsPane,
		},
		{
			name:         "pipelines to components (reverse)",
			startPane:    pipelinesPane,
			showPreview:  true,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
		{
			name:         "preview to pipelines (reverse)",
			startPane:    previewPane,
			showPreview:  true,
			wantPane:     pipelinesPane,
			wantLastData: pipelinesPane,
		},
		{
			name:         "search to components (reverse)",
			startPane:    searchPane,
			showPreview:  true,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
		// Without preview
		{
			name:         "components to pipelines (reverse, no preview)",
			startPane:    componentsPane,
			showPreview:  false,
			wantPane:     pipelinesPane,
			wantLastData: pipelinesPane,
		},
		{
			name:         "pipelines to components (reverse, no preview)",
			startPane:    pipelinesPane,
			showPreview:  false,
			wantPane:     componentsPane,
			wantLastData: componentsPane,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager()
			sm.ActivePane = tt.startPane
			sm.ShowPreview = tt.showPreview
			// Set initial LastDataPane
			if tt.startPane == componentsPane || tt.startPane == pipelinesPane {
				sm.LastDataPane = tt.startPane
			}
			
			sm.HandleTabNavigation(true)
			
			if sm.ActivePane != tt.wantPane {
				t.Errorf("ActivePane = %v, want %v", sm.ActivePane, tt.wantPane)
			}
			if sm.LastDataPane != tt.wantLastData {
				t.Errorf("LastDataPane = %v, want %v", sm.LastDataPane, tt.wantLastData)
			}
			
			validateStateManager(t, sm)
		})
	}
}

// Test search mode
func TestSearchMode(t *testing.T) {
	sm := NewStateManager()
	
	// Start in components
	if sm.ActivePane != componentsPane {
		t.Error("Should start in components pane")
	}
	
	// Switch to search
	sm.SwitchToSearch()
	if sm.ActivePane != searchPane {
		t.Error("Should be in search pane after SwitchToSearch")
	}
	
	// Exit search
	sm.ExitSearch()
	if sm.ActivePane != componentsPane {
		t.Error("Should return to components pane after ExitSearch")
	}
	
	// Test from pipelines pane
	sm.ActivePane = pipelinesPane
	sm.SwitchToSearch()
	if sm.ActivePane != searchPane {
		t.Error("Should be in search pane")
	}
	
	sm.ExitSearch()
	if sm.ActivePane != componentsPane {
		t.Error("Should always return to components pane from search")
	}
	
	validateStateManager(t, sm)
}

// Test state queries
func TestIsInPreviewPane(t *testing.T) {
	sm := NewStateManager()
	
	sm.ActivePane = previewPane
	if !sm.IsInPreviewPane() {
		t.Error("IsInPreviewPane should return true when in preview pane")
	}
	
	sm.ActivePane = componentsPane
	if sm.IsInPreviewPane() {
		t.Error("IsInPreviewPane should return false when not in preview pane")
	}
}

func TestIsInSearchPane(t *testing.T) {
	sm := NewStateManager()
	
	sm.ActivePane = searchPane
	if !sm.IsInSearchPane() {
		t.Error("IsInSearchPane should return true when in search pane")
	}
	
	sm.ActivePane = pipelinesPane
	if sm.IsInSearchPane() {
		t.Error("IsInSearchPane should return false when not in search pane")
	}
}

func TestGetPreviewPane(t *testing.T) {
	tests := []struct {
		name         string
		activePane   pane
		lastDataPane pane
		want         pane
	}{
		{
			name:         "active components returns components",
			activePane:   componentsPane,
			lastDataPane: pipelinesPane,
			want:         componentsPane,
		},
		{
			name:         "active pipelines returns pipelines",
			activePane:   pipelinesPane,
			lastDataPane: componentsPane,
			want:         pipelinesPane,
		},
		{
			name:         "active preview returns last data pane",
			activePane:   previewPane,
			lastDataPane: componentsPane,
			want:         componentsPane,
		},
		{
			name:         "active search returns search",
			activePane:   searchPane,
			lastDataPane: pipelinesPane,
			want:         searchPane,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager()
			sm.ActivePane = tt.activePane
			sm.LastDataPane = tt.lastDataPane
			
			got := sm.GetPreviewPane()
			if got != tt.want {
				t.Errorf("GetPreviewPane() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test deletion and archive state
func TestDeletionState(t *testing.T) {
	sm := NewStateManager()
	
	// Initially no deletion state
	if sm.DeletingFromPane != nonePane {
		t.Error("DeletingFromPane should start as nonePane")
	}
	
	// Set deletion from components
	sm.SetDeletingFromPane(componentsPane)
	if sm.DeletingFromPane != componentsPane {
		t.Error("DeletingFromPane should be componentsPane")
	}
	
	// Clear deletion state
	sm.ClearDeletionState()
	if sm.DeletingFromPane != nonePane {
		t.Error("DeletingFromPane should be nonePane after clear")
	}
	
	validateStateManager(t, sm)
}

func TestArchiveState(t *testing.T) {
	sm := NewStateManager()
	
	// Initially no archive state
	if sm.ArchivingFromPane != nonePane {
		t.Error("ArchivingFromPane should start as nonePane")
	}
	
	// Set archiving from pipelines
	sm.SetArchivingFromPane(pipelinesPane)
	if sm.ArchivingFromPane != pipelinesPane {
		t.Error("ArchivingFromPane should be pipelinesPane")
	}
	
	// Clear archive state
	sm.ClearArchiveState()
	if sm.ArchivingFromPane != nonePane {
		t.Error("ArchivingFromPane should be nonePane after clear")
	}
	
	validateStateManager(t, sm)
}

// Test ResetCursorsAfterSearch
func TestResetCursorsAfterSearch(t *testing.T) {
	tests := []struct {
		name             string
		initialCompCur   int
		initialPipeCur   int
		filteredCompCnt  int
		filteredPipeCnt  int
		expectedCompCur  int
		expectedPipeCur  int
	}{
		{
			name:            "cursors within bounds",
			initialCompCur:  2,
			initialPipeCur:  1,
			filteredCompCnt: 5,
			filteredPipeCnt: 3,
			expectedCompCur: 2,
			expectedPipeCur: 1,
		},
		{
			name:            "cursors out of bounds",
			initialCompCur:  10,
			initialPipeCur:  8,
			filteredCompCnt: 3,
			filteredPipeCnt: 2,
			expectedCompCur: 0,
			expectedPipeCur: 0,
		},
		{
			name:            "empty results",
			initialCompCur:  5,
			initialPipeCur:  3,
			filteredCompCnt: 0,
			filteredPipeCnt: 0,
			expectedCompCur: 0,
			expectedPipeCur: 0,
		},
		{
			name:            "cursor at boundary",
			initialCompCur:  2,
			initialPipeCur:  1,
			filteredCompCnt: 2,
			filteredPipeCnt: 1,
			expectedCompCur: 0,
			expectedPipeCur: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager()
			sm.ComponentCursor = tt.initialCompCur
			sm.PipelineCursor = tt.initialPipeCur
			
			sm.ResetCursorsAfterSearch(tt.filteredCompCnt, tt.filteredPipeCnt)
			
			if sm.ComponentCursor != tt.expectedCompCur {
				t.Errorf("ComponentCursor = %d, want %d", sm.ComponentCursor, tt.expectedCompCur)
			}
			if sm.PipelineCursor != tt.expectedPipeCur {
				t.Errorf("PipelineCursor = %d, want %d", sm.PipelineCursor, tt.expectedPipeCur)
			}
			
			validateStateManager(t, sm)
		})
	}
}

// Test HandleKeyNavigation
func TestHandleKeyNavigation(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		activePane     pane
		compCursor     int
		pipeCursor     int
		compCount      int
		pipeCount      int
		wantHandled    bool
		wantUpdate     bool
		wantCompCursor int
		wantPipeCursor int
		wantActivePane pane
	}{
		// Arrow key navigation
		{
			name:           "up in components",
			key:            "up",
			activePane:     componentsPane,
			compCursor:     2,
			pipeCursor:     1,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 1,
			wantPipeCursor: 1,
			wantActivePane: componentsPane,
		},
		{
			name:           "down in pipelines",
			key:            "down",
			activePane:     pipelinesPane,
			compCursor:     0,
			pipeCursor:     1,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 0,
			wantPipeCursor: 2,
			wantActivePane: pipelinesPane,
		},
		{
			name:           "k in components",
			key:            "k",
			activePane:     componentsPane,
			compCursor:     3,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 2,
			wantPipeCursor: 0,
			wantActivePane: componentsPane,
		},
		{
			name:           "j in pipelines",
			key:            "j",
			activePane:     pipelinesPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 0,
			wantPipeCursor: 1,
			wantActivePane: pipelinesPane,
		},
		// Tab navigation
		{
			name:           "tab from components",
			key:            "tab",
			activePane:     componentsPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 0,
			wantPipeCursor: 0,
			wantActivePane: pipelinesPane,
		},
		{
			name:           "shift+tab from pipelines",
			key:            "shift+tab",
			activePane:     pipelinesPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true,
			wantCompCursor: 0,
			wantPipeCursor: 0,
			wantActivePane: componentsPane,
		},
		{
			name:           "tab from preview",
			key:            "tab",
			activePane:     previewPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    true,
			wantUpdate:     true, // Update when switching FROM preview to components
			wantCompCursor: 0,
			wantPipeCursor: 0,
			wantActivePane: componentsPane,
		},
		// Non-navigation keys
		{
			name:           "unhandled key",
			key:            "enter",
			activePane:     componentsPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    false,
			wantUpdate:     false,
			wantCompCursor: 0,
			wantPipeCursor: 0,
			wantActivePane: componentsPane,
		},
		// Navigation in non-data panes
		{
			name:           "up in preview pane",
			key:            "up",
			activePane:     previewPane,
			compCursor:     2,
			pipeCursor:     1,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    false,
			wantUpdate:     false,
			wantCompCursor: 2,
			wantPipeCursor: 1,
			wantActivePane: previewPane,
		},
		{
			name:           "down in search pane",
			key:            "down",
			activePane:     searchPane,
			compCursor:     0,
			pipeCursor:     0,
			compCount:      5,
			pipeCount:      3,
			wantHandled:    false,
			wantUpdate:     false,
			wantCompCursor: 0,
			wantPipeCursor: 0,
			wantActivePane: searchPane,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := makeStateManager(tt.activePane, tt.compCursor, tt.pipeCursor, true)
			sm.UpdateCounts(tt.compCount, tt.pipeCount)
			
			handled, updatePreview := sm.HandleKeyNavigation(tt.key)
			
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if updatePreview != tt.wantUpdate {
				t.Errorf("updatePreview = %v, want %v", updatePreview, tt.wantUpdate)
			}
			if sm.ComponentCursor != tt.wantCompCursor {
				t.Errorf("ComponentCursor = %d, want %d", sm.ComponentCursor, tt.wantCompCursor)
			}
			if sm.PipelineCursor != tt.wantPipeCursor {
				t.Errorf("PipelineCursor = %d, want %d", sm.PipelineCursor, tt.wantPipeCursor)
			}
			if sm.ActivePane != tt.wantActivePane {
				t.Errorf("ActivePane = %v, want %v", sm.ActivePane, tt.wantActivePane)
			}
			
			validateStateManager(t, sm)
		})
	}
}

// Test state transitions
func TestStateTransitions_TabCycle(t *testing.T) {
	// Test that tab navigation forms a cycle
	sm := NewStateManager()
	sm.ShowPreview = true
	
	startPane := sm.ActivePane
	paneSequence := []pane{startPane}
	
	// Tab through all panes until we return to start
	for i := 0; i < 10; i++ { // Safety limit
		sm.HandleTabNavigation(false)
		paneSequence = append(paneSequence, sm.ActivePane)
		if sm.ActivePane == startPane && i > 0 {
			break
		}
	}
	
	// Verify we completed a cycle
	expectedCycle := []pane{componentsPane, pipelinesPane, previewPane, componentsPane}
	if len(paneSequence) != len(expectedCycle) {
		t.Errorf("Tab cycle length = %d, want %d", len(paneSequence), len(expectedCycle))
	}
	
	for i, pane := range expectedCycle {
		if i < len(paneSequence) && paneSequence[i] != pane {
			t.Errorf("Tab cycle[%d] = %v, want %v", i, paneSequence[i], pane)
		}
	}
	
	// Test reverse cycle
	sm = NewStateManager()
	sm.ShowPreview = true
	reverseSequence := []pane{componentsPane}
	
	for i := 0; i < 10; i++ {
		sm.HandleTabNavigation(true)
		reverseSequence = append(reverseSequence, sm.ActivePane)
		if sm.ActivePane == componentsPane && i > 0 {
			break
		}
	}
	
	expectedReverse := []pane{componentsPane, previewPane, pipelinesPane, componentsPane}
	if len(reverseSequence) != len(expectedReverse) {
		t.Errorf("Reverse tab cycle length = %d, want %d", len(reverseSequence), len(expectedReverse))
	}
}

func TestStateTransitions_PreviewToggle(t *testing.T) {
	sm := NewStateManager()
	
	// Start with preview enabled
	sm.ShowPreview = true
	sm.ActivePane = pipelinesPane
	
	// Tab should go to preview
	sm.HandleTabNavigation(false)
	if sm.ActivePane != previewPane {
		t.Error("Should navigate to preview when enabled")
	}
	
	// Now test with preview disabled from the beginning
	sm = NewStateManager()
	sm.ShowPreview = false
	sm.ActivePane = pipelinesPane
	
	// Tab should skip preview and go to components
	sm.HandleTabNavigation(false)
	if sm.ActivePane != componentsPane {
		t.Error("Should skip to components when preview disabled")
	}
	
	// Verify preview is skipped in cycle
	sm.HandleTabNavigation(false)
	if sm.ActivePane != pipelinesPane {
		t.Error("Should go to pipelines")
	}
	
	sm.HandleTabNavigation(false)
	if sm.ActivePane != componentsPane {
		t.Error("Should skip preview and return to components")
	}
}

// Test concurrent access (basic race condition test)
func TestStateManager_ConcurrentAccess(t *testing.T) {
	sm := NewStateManager()
	sm.UpdateCounts(100, 50)
	
	done := make(chan bool)
	
	// Simulate concurrent cursor movements
	go func() {
		for i := 0; i < 50; i++ {
			sm.MoveCursorDown()
			sm.MoveCursorUp()
		}
		done <- true
	}()
	
	// Simulate concurrent pane switches
	go func() {
		for i := 0; i < 50; i++ {
			sm.HandleTabNavigation(false)
			sm.HandleTabNavigation(true)
		}
		done <- true
	}()
	
	// Simulate concurrent state queries
	go func() {
		for i := 0; i < 50; i++ {
			_ = sm.IsInPreviewPane()
			_ = sm.IsInSearchPane()
			_ = sm.GetPreviewPane()
		}
		done <- true
	}()
	
	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
	
	// Final state should still be valid
	validateStateManager(t, sm)
}

// Benchmarks
func BenchmarkMoveCursor(b *testing.B) {
	sm := NewStateManager()
	sm.UpdateCounts(1000, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			sm.MoveCursorDown()
		} else {
			sm.MoveCursorUp()
		}
	}
}

func BenchmarkTabNavigation(b *testing.B) {
	sm := NewStateManager()
	sm.ShowPreview = true
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.HandleTabNavigation(i%2 == 0)
	}
}

func BenchmarkHandleKeyNavigation(b *testing.B) {
	sm := NewStateManager()
	sm.UpdateCounts(100, 100)
	keys := []string{"up", "down", "tab", "shift+tab", "j", "k"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		sm.HandleKeyNavigation(key)
	}
}