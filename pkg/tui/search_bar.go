package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchBar is a reusable search component with consistent styling
type SearchBar struct {
	input      textinput.Model
	isActive   bool
	width      int
	searchText string
}

// NewSearchBar creates a new search bar component
func NewSearchBar() *SearchBar {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	ti.Width = 50 // Default width, will be adjusted

	return &SearchBar{
		input: ti,
	}
}

// SetActive sets whether the search bar is the active pane
func (s *SearchBar) SetActive(active bool) {
	s.isActive = active
	if active {
		s.input.Focus()
	} else {
		s.input.Blur()
	}
}

// SetWidth sets the width for the search bar
func (s *SearchBar) SetWidth(width int) {
	s.width = width
	// Adjust input width accounting for icon, padding, and borders
	// Width - 4 (borders) - 2 (outer padding) - 5 (icon with spaces) - 1 (space after icon)
	s.input.Width = width - 12
}

// Value returns the current search text
func (s *SearchBar) Value() string {
	return s.input.Value()
}

// SetValue sets the search text
func (s *SearchBar) SetValue(value string) {
	s.input.SetValue(value)
}

// Update handles tea messages for the search bar
func (s *SearchBar) Update(msg tea.Msg) (*SearchBar, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

// View renders the search bar with consistent styling
func (s *SearchBar) View() string {
	// Search bar border style
	borderColor := "240" // Inactive gray
	if s.isActive {
		borderColor = "170" // Active purple
	}
	
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(s.width - 4). // Account for outer padding
		Padding(0, 1)
	
	// Create search icon with dynamic styling based on active state
	var searchIconStyle lipgloss.Style
	var searchIcon string
	if s.isActive {
		// Active: white icon on purple background
		searchIconStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("170")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1) // 1 space on either side
		searchIcon = searchIconStyle.Render("⌕")
	} else {
		// Inactive: gray icon with no background
		searchIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Bold(true)
		searchIcon = searchIconStyle.Render(" ⌕ ") // 1 space on either side to match active state width
	}
	
	// Add spacing after icon before search input
	searchContent := lipgloss.JoinHorizontal(lipgloss.Center, searchIcon, " ", s.input.View())
	
	// Apply outer padding
	outerPadding := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	
	return outerPadding.Render(searchStyle.Render(searchContent))
}

// Focus focuses the search input
func (s *SearchBar) Focus() tea.Cmd {
	return s.input.Focus()
}

// Blur removes focus from the search input
func (s *SearchBar) Blur() {
	s.input.Blur()
}

// Reset clears the search input
func (s *SearchBar) Reset() {
	s.input.SetValue("")
}