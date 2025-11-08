package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Issue represents a GitHub issue (dummy data structure for now)
type Issue struct {
	Number    int
	Title     string
	State     string
	Author    string
	Labels    []string
	Comments  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IssueView is the model for the issue list view
type IssueView struct {
	issues     []Issue
	cursor     int
	selected   map[int]struct{}
	loading    bool
	err        error
	width      int
	height     int
	statusBar  *components.StatusBar
	showHelp   bool
}

// NewIssueView creates a new issue view
func NewIssueView() *IssueView {
	return &IssueView{
		issues:    generateDummyIssues(),
		cursor:    0,
		selected:  make(map[int]struct{}),
		loading:   false,
		statusBar: components.NewStatusBar(),
		showHelp:  false,
	}
}

// Init initializes the issue view
func (m *IssueView) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *IssueView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(msg.Width, 1)
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *IssueView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "j", "down":
		if m.cursor < len(m.issues)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "g":
		// Go to top
		m.cursor = 0
		return m, nil

	case "G":
		// Go to bottom
		if len(m.issues) > 0 {
			m.cursor = len(m.issues) - 1
		}
		return m, nil

	case "enter", " ":
		// Toggle selection (for future use)
		if _, ok := m.selected[m.cursor]; ok {
			delete(m.selected, m.cursor)
		} else {
			m.selected[m.cursor] = struct{}{}
		}
		return m, nil
	}

	return m, nil
}

// View renders the issue view
func (m *IssueView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	// Header
	header := m.renderHeader()
	s.WriteString(header)
	s.WriteString("\n")

	// Issue list or error/loading state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else {
		s.WriteString(m.renderIssueList())
	}

	// Help section (if enabled)
	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(m.renderHelp())
	}

	// Status bar
	s.WriteString("\n")
	m.updateStatusBar()
	s.WriteString(m.statusBar.View())

	return s.String()
}

// renderHeader renders the view header
func (m *IssueView) renderHeader() string {
	title := styles.HeaderStyle.Render("Issues")
	count := styles.MutedStyle.Render(fmt.Sprintf("(%d)", len(m.issues)))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		" ",
		count,
	)
}

// renderIssueList renders the list of issues
func (m *IssueView) renderIssueList() string {
	var s strings.Builder

	// Calculate available height for list (total - header - status bar - margins)
	availableHeight := m.height - 4
	if m.showHelp {
		availableHeight -= 10 // Reserve space for help
	}

	// Calculate visible range
	startIdx := 0
	endIdx := len(m.issues)

	if len(m.issues) > availableHeight {
		// Show items around cursor
		halfHeight := availableHeight / 2
		startIdx = m.cursor - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.issues) {
			endIdx = len(m.issues)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible issues
	for i := startIdx; i < endIdx; i++ {
		issue := m.issues[i]
		line := m.renderIssueLine(issue, i)
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderIssueLine renders a single issue line
func (m *IssueView) renderIssueLine(issue Issue, index int) string {
	// Cursor indicator
	cursor := "  "
	if m.cursor == index {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	// State badge
	stateBadge := styles.GetStateBadge(issue.State)

	// Issue number
	number := styles.IssueNumberStyle.Render(fmt.Sprintf("#%-5d", issue.Number))

	// Title
	titleStyle := styles.IssueTitleStyle
	if m.cursor == index {
		titleStyle = styles.SelectedStyle
	}
	title := titleStyle.Render(issue.Title)

	// Labels
	labels := ""
	if len(issue.Labels) > 0 {
		labelParts := []string{}
		for _, label := range issue.Labels {
			labelParts = append(labelParts, styles.LabelStyle.Render(label))
		}
		labels = " " + strings.Join(labelParts, " ")
	}

	// Metadata (author, comments, date)
	author := styles.AuthorStyle.Render("@" + issue.Author)
	comments := ""
	if issue.Comments > 0 {
		comments = styles.MutedStyle.Render(fmt.Sprintf("💬 %d", issue.Comments))
	}
	relativeTime := formatRelativeTime(issue.UpdatedAt)
	date := styles.DateStyle.Render(relativeTime)

	// Combine all parts
	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		cursor,
		stateBadge,
		" ",
		number,
		" ",
		title,
		labels,
		" ",
		author,
	)

	if comments != "" {
		line = lipgloss.JoinHorizontal(lipgloss.Top, line, " ", comments)
	}

	line = lipgloss.JoinHorizontal(lipgloss.Top, line, " ", date)

	return line
}

// renderLoading renders a loading state
func (m *IssueView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading issues...")
}

// renderError renders an error state
func (m *IssueView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// renderHelp renders the help section
func (m *IssueView) renderHelp() string {
	helpText := `
Navigation:
  ↑/k     Move up
  ↓/j     Move down
  g       Go to top
  G       Go to bottom

Actions:
  enter   View issue details
  space   Toggle selection
  r       Refresh

General:
  ?       Toggle help
  q       Quit
  ctrl+c  Force quit
`

	return styles.BorderStyle.Render(
		styles.HelpStyle.Render(strings.TrimSpace(helpText)),
	)
}

// updateStatusBar updates the status bar with current state
func (m *IssueView) updateStatusBar() {
	m.statusBar.ClearItems()
	m.statusBar.SetMode("Issues")

	// Add current position
	if len(m.issues) > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.issues))
		m.statusBar.AddItem("", position)
	}

	// Add selection count if any
	if len(m.selected) > 0 {
		m.statusBar.AddItem("Selected", fmt.Sprintf("%d", len(m.selected)))
	}

	// Add repository info (dummy for now)
	m.statusBar.AddItem("Repo", "owner/repo")
}

// generateDummyIssues generates dummy issues for testing
func generateDummyIssues() []Issue {
	now := time.Now()

	return []Issue{
		{
			Number:    1,
			Title:     "Implement basic TUI framework",
			State:     "open",
			Author:    "alice",
			Labels:    []string{"enhancement", "ui"},
			Comments:  5,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			Number:    2,
			Title:     "Add GitHub API integration",
			State:     "open",
			Author:    "bob",
			Labels:    []string{"feature", "api"},
			Comments:  3,
			CreatedAt: now.Add(-36 * time.Hour),
			UpdatedAt: now.Add(-5 * time.Hour),
		},
		{
			Number:    3,
			Title:     "Fix authentication bug",
			State:     "closed",
			Author:    "charlie",
			Labels:    []string{"bug", "security"},
			Comments:  8,
			CreatedAt: now.Add(-72 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		{
			Number:    4,
			Title:     "Improve performance of issue list rendering",
			State:     "open",
			Author:    "alice",
			Labels:    []string{"performance"},
			Comments:  2,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			Number:    5,
			Title:     "Add dark mode support",
			State:     "open",
			Author:    "dave",
			Labels:    []string{"enhancement", "ui"},
			Comments:  12,
			CreatedAt: now.Add(-96 * time.Hour),
			UpdatedAt: now.Add(-12 * time.Hour),
		},
		{
			Number:    6,
			Title:     "Write comprehensive documentation",
			State:     "open",
			Author:    "eve",
			Labels:    []string{"documentation"},
			Comments:  0,
			CreatedAt: now.Add(-12 * time.Hour),
			UpdatedAt: now.Add(-12 * time.Hour),
		},
		{
			Number:    7,
			Title:     "Implement keyboard shortcuts",
			State:     "closed",
			Author:    "frank",
			Labels:    []string{"enhancement", "ux"},
			Comments:  6,
			CreatedAt: now.Add(-120 * time.Hour),
			UpdatedAt: now.Add(-48 * time.Hour),
		},
		{
			Number:    8,
			Title:     "Add search and filter functionality",
			State:     "open",
			Author:    "grace",
			Labels:    []string{"feature"},
			Comments:  4,
			CreatedAt: now.Add(-60 * time.Hour),
			UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			Number:    9,
			Title:     "Optimize API rate limiting",
			State:     "open",
			Author:    "bob",
			Labels:    []string{"performance", "api"},
			Comments:  7,
			CreatedAt: now.Add(-84 * time.Hour),
			UpdatedAt: now.Add(-6 * time.Hour),
		},
		{
			Number:    10,
			Title:     "Add unit tests for UI components",
			State:     "open",
			Author:    "alice",
			Labels:    []string{"testing"},
			Comments:  1,
			CreatedAt: now.Add(-18 * time.Hour),
			UpdatedAt: now.Add(-18 * time.Hour),
		},
	}
}

// formatRelativeTime formats a time as relative (e.g., "2 hours ago")
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
