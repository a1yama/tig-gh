package components

import (
	"strings"

	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchInput represents a search input component
type SearchInput struct {
	value       string
	cursor      int
	active      bool
	width       int
	height      int
	placeholder string
}

// NewSearchInput creates a new search input
func NewSearchInput() *SearchInput {
	return &SearchInput{
		value:       "",
		cursor:      0,
		active:      false,
		placeholder: "Search (e.g., author:user label:bug \"exact phrase\")",
	}
}

// Activate activates the search input
func (s *SearchInput) Activate() {
	s.active = true
}

// Deactivate deactivates the search input
func (s *SearchInput) Deactivate() {
	s.active = false
}

// Focus activates the search input
func (s *SearchInput) Focus() {
	s.Activate()
}

// Blur deactivates the search input
func (s *SearchInput) Blur() {
	s.Deactivate()
}

// IsActive returns true if the search input is active
func (s *SearchInput) IsActive() bool {
	return s.active
}

// SetValue sets the search value
func (s *SearchInput) SetValue(value string) {
	s.value = value
	// Keep cursor in bounds
	if s.cursor > len(s.value) {
		s.cursor = len(s.value)
	}
}

// GetValue returns the current search value
func (s *SearchInput) GetValue() string {
	return s.value
}

// Clear clears the search input
func (s *SearchInput) Clear() {
	s.value = ""
	s.cursor = 0
}

// SetSize sets the size of the search input
func (s *SearchInput) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Update handles input events
func (s *SearchInput) Update(msg tea.Msg) {
	if !s.active {
		return
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			// Insert character at cursor position
			runes := []rune(s.value)
			before := string(runes[:s.cursor])
			after := string(runes[s.cursor:])
			s.value = before + string(msg.Runes) + after
			s.cursor += len(msg.Runes)

		case tea.KeyBackspace:
			if s.cursor > 0 {
				runes := []rune(s.value)
				before := string(runes[:s.cursor-1])
				after := string(runes[s.cursor:])
				s.value = before + after
				s.cursor--
			}

		case tea.KeyDelete:
			if s.cursor < len([]rune(s.value)) {
				runes := []rune(s.value)
				before := string(runes[:s.cursor])
				after := string(runes[s.cursor+1:])
				s.value = before + after
			}

		case tea.KeyLeft:
			if s.cursor > 0 {
				s.cursor--
			}

		case tea.KeyRight:
			if s.cursor < len([]rune(s.value)) {
				s.cursor++
			}

		case tea.KeyHome:
			s.cursor = 0

		case tea.KeyEnd:
			s.cursor = len([]rune(s.value))

		case tea.KeyEsc:
			s.Deactivate()
		}
	}
}

// View renders the search input
func (s *SearchInput) View() string {
	if s.width == 0 {
		return ""
	}

	// Build the display value
	displayValue := s.value
	if displayValue == "" && !s.active {
		displayValue = s.placeholder
	}

	// Create the input style
	inputStyle := lipgloss.NewStyle().
		Width(s.width - 4).
		Padding(0, 1)

	if s.active {
		inputStyle = inputStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorPrimary)
	} else {
		inputStyle = inputStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder)
	}

	// Add cursor if active
	var content string
	if s.active && s.cursor <= len([]rune(s.value)) {
		runes := []rune(displayValue)
		before := string(runes[:s.cursor])
		cursor := "│"
		after := ""
		if s.cursor < len(runes) {
			after = string(runes[s.cursor:])
		}

		cursorStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
		content = before + cursorStyle.Render(cursor) + after
	} else {
		if displayValue == "" && !s.active {
			content = styles.MutedStyle.Render(s.placeholder)
		} else {
			content = displayValue
		}
	}

	// Render the search icon
	searchIcon := "🔍 "
	if s.active {
		searchIcon = styles.StatusKeyStyle.Render("/ ")
	} else {
		searchIcon = styles.MutedStyle.Render("/ ")
	}

	return inputStyle.Render(searchIcon + content)
}

// SetPlaceholder sets the placeholder text
func (s *SearchInput) SetPlaceholder(placeholder string) {
	s.placeholder = placeholder
}

// MoveCursorToEnd moves the cursor to the end of the input
func (s *SearchInput) MoveCursorToEnd() {
	s.cursor = len([]rune(s.value))
}

// MoveCursorToStart moves the cursor to the start of the input
func (s *SearchInput) MoveCursorToStart() {
	s.cursor = 0
}

// Insert inserts text at the cursor position
func (s *SearchInput) Insert(text string) {
	runes := []rune(s.value)
	before := string(runes[:s.cursor])
	after := string(runes[s.cursor:])
	s.value = before + text + after
	s.cursor += len([]rune(text))
}

// DeleteWord deletes the word before the cursor
func (s *SearchInput) DeleteWord() {
	if s.cursor == 0 {
		return
	}

	runes := []rune(s.value)
	before := string(runes[:s.cursor])
	after := string(runes[s.cursor:])

	// Find the start of the word
	words := strings.Fields(before)
	if len(words) == 0 {
		s.value = after
		s.cursor = 0
		return
	}

	// Remove the last word
	newBefore := strings.Join(words[:len(words)-1], " ")
	if len(words) > 1 {
		newBefore += " "
	}

	s.value = newBefore + after
	s.cursor = len([]rune(newBefore))
}
