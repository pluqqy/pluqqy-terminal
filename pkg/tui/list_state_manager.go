package tui

import "sync"

// StateManager handles all state management for the list view including
// cursor positions, pane navigation, and selection state
type StateManager struct {
	// Mutex for thread-safe access
	mu sync.RWMutex
	// Cursor positions
	ComponentCursor int
	PipelineCursor  int
	
	// Active pane tracking
	ActivePane    pane
	LastDataPane  pane
	
	// View state
	ShowPreview   bool
	
	// Deletion/Archive state
	DeletingFromPane  pane
	ArchivingFromPane pane
	
	// Item counts (for bounds checking)
	componentCount int
	pipelineCount  int
}

// NewStateManager creates a new state manager instance
func NewStateManager() *StateManager {
	return &StateManager{
		ComponentCursor: 0,
		PipelineCursor:  0,
		ActivePane:      componentsPane,
		LastDataPane:    componentsPane,
		ShowPreview:     true,
	}
}

// UpdateCounts updates the item counts for bounds checking
func (sm *StateManager) UpdateCounts(componentCount, pipelineCount int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.componentCount = componentCount
	sm.pipelineCount = pipelineCount
	
	// Reset cursors if they're out of bounds
	if sm.ComponentCursor >= componentCount && componentCount > 0 {
		sm.ComponentCursor = componentCount - 1
	} else if componentCount == 0 {
		sm.ComponentCursor = 0
	}
	
	if sm.PipelineCursor >= pipelineCount && pipelineCount > 0 {
		sm.PipelineCursor = pipelineCount - 1
	} else if pipelineCount == 0 {
		sm.PipelineCursor = 0
	}
}

// MoveCursorUp moves the cursor up in the active pane
func (sm *StateManager) MoveCursorUp() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	switch sm.ActivePane {
	case componentsPane:
		if sm.ComponentCursor > 0 {
			sm.ComponentCursor--
			return true
		}
	case pipelinesPane:
		if sm.PipelineCursor > 0 {
			sm.PipelineCursor--
			return true
		}
	}
	return false
}

// MoveCursorDown moves the cursor down in the active pane
func (sm *StateManager) MoveCursorDown() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	switch sm.ActivePane {
	case componentsPane:
		if sm.ComponentCursor < sm.componentCount-1 {
			sm.ComponentCursor++
			return true
		}
	case pipelinesPane:
		if sm.PipelineCursor < sm.pipelineCount-1 {
			sm.PipelineCursor++
			return true
		}
	}
	return false
}

// HandleTabNavigation handles tab/shift+tab navigation between panes
func (sm *StateManager) HandleTabNavigation(reverse bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if reverse {
		sm.navigateBackward()
	} else {
		sm.navigateForward()
	}
}

// navigateForward handles forward tab navigation
func (sm *StateManager) navigateForward() {
	if sm.ShowPreview {
		// With preview: components -> pipelines -> preview -> components
		switch sm.ActivePane {
		case searchPane:
			sm.ActivePane = componentsPane
		case componentsPane:
			sm.ActivePane = pipelinesPane
		case pipelinesPane:
			sm.ActivePane = previewPane
		case previewPane:
			sm.ActivePane = componentsPane
		}
	} else {
		// Without preview: components -> pipelines -> components
		switch sm.ActivePane {
		case searchPane:
			sm.ActivePane = componentsPane
		case componentsPane:
			sm.ActivePane = pipelinesPane
		case pipelinesPane:
			sm.ActivePane = componentsPane
		}
	}
	
	// Track last data pane when switching
	if sm.ActivePane == pipelinesPane || sm.ActivePane == componentsPane {
		sm.LastDataPane = sm.ActivePane
	}
}

// navigateBackward handles backward tab navigation (shift+tab)
func (sm *StateManager) navigateBackward() {
	if sm.ShowPreview {
		// With preview: components <- pipelines <- preview <- components (reverse of forward)
		switch sm.ActivePane {
		case searchPane:
			sm.ActivePane = componentsPane
		case componentsPane:
			sm.ActivePane = previewPane
		case pipelinesPane:
			sm.ActivePane = componentsPane
		case previewPane:
			sm.ActivePane = pipelinesPane
		}
	} else {
		// Without preview: components <- pipelines <- components (reverse of forward)
		switch sm.ActivePane {
		case searchPane:
			sm.ActivePane = componentsPane
		case componentsPane:
			sm.ActivePane = pipelinesPane
		case pipelinesPane:
			sm.ActivePane = componentsPane
		}
	}
	
	// Track last data pane when switching
	if sm.ActivePane == pipelinesPane || sm.ActivePane == componentsPane {
		sm.LastDataPane = sm.ActivePane
	}
}

// SwitchToSearch switches to the search pane
func (sm *StateManager) SwitchToSearch() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.ActivePane = searchPane
}

// ExitSearch exits search mode and returns to components pane
func (sm *StateManager) ExitSearch() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.ActivePane = componentsPane
}

// GetPreviewPane returns the pane to show in preview
func (sm *StateManager) GetPreviewPane() pane {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.ActivePane == previewPane {
		// If we're on the preview pane, use the last data pane
		return sm.LastDataPane
	}
	return sm.ActivePane
}

// IsInPreviewPane returns true if the preview pane is active
func (sm *StateManager) IsInPreviewPane() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.ActivePane == previewPane
}

// IsInSearchPane returns true if the search pane is active
func (sm *StateManager) IsInSearchPane() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.ActivePane == searchPane
}

// SetDeletingFromPane sets the pane from which deletion is happening
func (sm *StateManager) SetDeletingFromPane(p pane) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.DeletingFromPane = p
}

// SetArchivingFromPane sets the pane from which archiving is happening
func (sm *StateManager) SetArchivingFromPane(p pane) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.ArchivingFromPane = p
}

// ClearDeletionState clears the deletion state
func (sm *StateManager) ClearDeletionState() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.DeletingFromPane = nonePane
}

// ClearArchiveState clears the archive state
func (sm *StateManager) ClearArchiveState() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.ArchivingFromPane = nonePane
}

// ResetCursorsAfterSearch resets cursors after search results change
func (sm *StateManager) ResetCursorsAfterSearch(filteredComponentCount, filteredPipelineCount int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	// Reset cursors if they're out of bounds
	if sm.PipelineCursor >= filteredPipelineCount {
		sm.PipelineCursor = 0
	}
	if sm.ComponentCursor >= filteredComponentCount {
		sm.ComponentCursor = 0
	}
}

// HandleKeyNavigation processes navigation keys and returns whether preview should update
func (sm *StateManager) HandleKeyNavigation(key string) (handled bool, updatePreview bool) {
	switch key {
	case "up", "k":
		sm.mu.RLock()
		activePane := sm.ActivePane
		sm.mu.RUnlock()
		if activePane == componentsPane || activePane == pipelinesPane {
			moved := sm.MoveCursorUp()
			return true, moved
		}
	case "down", "j":
		sm.mu.RLock()
		activePane := sm.ActivePane
		sm.mu.RUnlock()
		if activePane == componentsPane || activePane == pipelinesPane {
			moved := sm.MoveCursorDown()
			return true, moved
		}
	case "tab":
		sm.HandleTabNavigation(false)
		// Update preview when switching to non-preview pane
		sm.mu.RLock()
		updatePreview = sm.ActivePane != previewPane && sm.ActivePane != searchPane
		sm.mu.RUnlock()
		return true, updatePreview
	case "shift+tab":
		sm.HandleTabNavigation(true)
		// Update preview when switching to non-preview pane
		sm.mu.RLock()
		updatePreview = sm.ActivePane != previewPane && sm.ActivePane != searchPane
		sm.mu.RUnlock()
		return true, updatePreview
	}
	
	return false, false
}