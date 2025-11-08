package components

import (
	"fmt"

	"github.com/a1yama/tig-gh/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents a status bar component
type StatusBar struct {
	width   int
	height  int
	mode    string
	message string
	items   []StatusItem
}

// StatusItem represents a single item in the status bar
type StatusItem struct {
	Key   string
	Value string
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		mode:  "Normal",
		items: []StatusItem{},
	}
}

// SetSize sets the size of the status bar
func (s *StatusBar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetMode sets the current mode
func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
}

// SetMessage sets a temporary message
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}

// SetItems sets the status items
func (s *StatusBar) SetItems(items []StatusItem) {
	s.items = items
}

// AddItem adds a status item
func (s *StatusBar) AddItem(key, value string) {
	s.items = append(s.items, StatusItem{Key: key, Value: value})
}

// ClearItems clears all status items
func (s *StatusBar) ClearItems() {
	s.items = []StatusItem{}
}

// Render renders the status bar
func (s *StatusBar) Render() string {
	if s.width == 0 {
		return ""
	}

	// Left side: mode and message
	leftParts := []string{}

	// Mode
	modeStyle := styles.StatusKeyStyle.Copy().
		Background(styles.ColorPrimary).
		Foreground(styles.ColorBackground).
		Padding(0, 1)
	leftParts = append(leftParts, modeStyle.Render(s.mode))

	// Message
	if s.message != "" {
		msgStyle := styles.StatusValueStyle.Copy().Padding(0, 1)
		leftParts = append(leftParts, msgStyle.Render(s.message))
	}

	leftContent := lipgloss.JoinHorizontal(lipgloss.Top, leftParts...)

	// Right side: status items
	rightParts := []string{}
	for _, item := range s.items {
		keyStyle := styles.StatusKeyStyle.Copy().Padding(0, 1)
		valueStyle := styles.StatusValueStyle.Copy()

		part := lipgloss.JoinHorizontal(
			lipgloss.Top,
			keyStyle.Render(item.Key),
			valueStyle.Render(item.Value),
		)
		rightParts = append(rightParts, part)
	}

	rightContent := ""
	if len(rightParts) > 0 {
		rightContent = lipgloss.JoinHorizontal(lipgloss.Top, rightParts...)
	}

	// Calculate spacing
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)
	spacingWidth := s.width - leftWidth - rightWidth

	if spacingWidth < 0 {
		spacingWidth = 0
	}

	spacing := lipgloss.NewStyle().
		Width(spacingWidth).
		Render("")

	// Combine left, spacing, and right
	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		spacing,
		rightContent,
	)

	// Apply background and ensure full width
	return lipgloss.NewStyle().
		Width(s.width).
		Background(styles.ColorBackground).
		Foreground(styles.ColorForeground).
		Render(statusBar)
}

// View returns the rendered status bar
func (s *StatusBar) View() string {
	return s.Render()
}

// Helper functions for common status bar configurations

// DefaultStatusBar creates a default status bar with basic info
func DefaultStatusBar(width int, repo string) *StatusBar {
	sb := NewStatusBar()
	sb.SetSize(width, 1)
	sb.SetMode("Normal")
	if repo != "" {
		sb.AddItem("Repo", repo)
	}
	return sb
}

// LoadingStatusBar creates a status bar showing loading state
func LoadingStatusBar(width int, message string) *StatusBar {
	sb := NewStatusBar()
	sb.SetSize(width, 1)
	sb.SetMode("Loading")
	sb.SetMessage(message)
	return sb
}

// ErrorStatusBar creates a status bar showing error state
func ErrorStatusBar(width int, err error) *StatusBar {
	sb := NewStatusBar()
	sb.SetSize(width, 1)
	sb.SetMode("Error")
	if err != nil {
		sb.SetMessage(err.Error())
	}
	return sb
}

// FormatHelp formats help text for the status bar
func FormatHelp(bindings map[string]string) string {
	parts := []string{}
	for key, desc := range bindings {
		part := styles.FormatKeyBinding(key, desc)
		parts = append(parts, part)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// RenderHelp renders a help line separate from the status bar
func RenderHelp(width int, bindings map[string]string) string {
	helpText := FormatHelp(bindings)
	return lipgloss.NewStyle().
		Width(width).
		Foreground(styles.ColorMuted).
		Background(styles.ColorBackground).
		Padding(0, 1).
		Render(helpText)
}

// IssueViewHelp returns help text for issue view
func IssueViewHelp() map[string]string {
	return map[string]string{
		"↑/k":   "up",
		"↓/j":   "down",
		"enter": "view",
		"?":     "help",
		"q":     "quit",
	}
}

// DetailViewHelp returns help text for detail view
func DetailViewHelp() map[string]string {
	return map[string]string{
		"↑/k": "up",
		"↓/j": "down",
		"esc": "back",
		"q":   "quit",
	}
}

// GlobalHelp returns global help text
func GlobalHelp() map[string]string {
	return map[string]string{
		"q":     "quit",
		"?":     "help",
		"ctrl+c": "force quit",
	}
}

// FormatStatusMessage formats a status message with timestamp
func FormatStatusMessage(message string) string {
	return fmt.Sprintf("%s", message)
}
