package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// backMsg is a message to go back to the previous view
type backMsg struct{}

// openBrowserMsg is a message to open the issue in browser
type openBrowserMsg struct {
	url string
}

// IssueDetailView is the model for the issue detail view
type IssueDetailView struct {
	issue        *models.Issue
	scrollOffset int
	loading      bool
	err          error
	width        int
	height       int
	renderer     *glamour.TermRenderer
}

// NewIssueDetailView creates a new issue detail view
func NewIssueDetailView(issue *models.Issue) *IssueDetailView {
	// Create a glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return &IssueDetailView{
		issue:        issue,
		scrollOffset: 0,
		loading:      false,
		renderer:     renderer,
	}
}

// Init initializes the issue detail view
func (m *IssueDetailView) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *IssueDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *IssueDetailView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q", "esc":
		// Go back to issue list
		return m, func() tea.Msg {
			return backMsg{}
		}

	case "j", "down":
		// Scroll down
		m.scrollOffset++
		return m, nil

	case "k", "up":
		// Scroll up
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "g":
		// Go to top
		m.scrollOffset = 0
		return m, nil

	case "G":
		// Go to bottom (will be handled in View with max offset calculation)
		m.scrollOffset = 9999 // Will be capped in View
		return m, nil

	case "o":
		// Open in browser
		return m, func() tea.Msg {
			return openBrowserMsg{url: m.issue.HTMLURL}
		}
	}

	return m, nil
}

// View renders the issue detail view
func (m *IssueDetailView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n\n")

	// Metadata
	s.WriteString(m.renderMetadata())
	s.WriteString("\n\n")

	// Separator
	s.WriteString(styles.Separator(m.width - 4))
	s.WriteString("\n\n")

	// Body (with markdown rendering)
	s.WriteString(m.renderBody())
	s.WriteString("\n\n")

	// Footer with help
	s.WriteString(m.renderFooter())

	return s.String()
}

// renderHeader renders the issue header
func (m *IssueDetailView) renderHeader() string {
	// Issue number and state
	numberStyle := styles.IssueNumberStyle
	number := numberStyle.Render(fmt.Sprintf("Issue #%d", m.issue.Number))

	stateBadge := styles.GetStateBadge(string(m.issue.State))

	// Title
	titleStyle := styles.BoldStyle
	title := titleStyle.Render(m.issue.Title)

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		number,
		" ",
		stateBadge,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		title,
	)
}

// renderMetadata renders issue metadata
func (m *IssueDetailView) renderMetadata() string {
	var parts []string

	// Author
	authorLabel := styles.MutedStyle.Render("Author:")
	authorValue := styles.AuthorStyle.Render("@" + m.issue.Author.Login)
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, authorLabel, " ", authorValue))

	// Created date
	createdLabel := styles.MutedStyle.Render("Created:")
	createdValue := styles.DateStyle.Render(formatTime(m.issue.CreatedAt))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, createdLabel, " ", createdValue))

	// Updated date
	updatedLabel := styles.MutedStyle.Render("Updated:")
	updatedValue := styles.DateStyle.Render(formatTime(m.issue.UpdatedAt))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, updatedLabel, " ", updatedValue))

	// Assignees
	if len(m.issue.Assignees) > 0 {
		assigneeNames := []string{}
		for _, assignee := range m.issue.Assignees {
			assigneeNames = append(assigneeNames, "@"+assignee.Login)
		}
		assigneesLabel := styles.MutedStyle.Render("Assignees:")
		assigneesValue := styles.AuthorStyle.Render(strings.Join(assigneeNames, ", "))
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, assigneesLabel, " ", assigneesValue))
	}

	// Labels
	if len(m.issue.Labels) > 0 {
		labelNames := []string{}
		for _, label := range m.issue.Labels {
			labelNames = append(labelNames, label.Name)
		}
		labelsLabel := styles.MutedStyle.Render("Labels:")
		labelsValue := styles.LabelStyle.Render(strings.Join(labelNames, ", "))
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, labelsLabel, " ", labelsValue))
	}

	// Milestone
	if m.issue.Milestone != nil {
		milestoneLabel := styles.MutedStyle.Render("Milestone:")
		milestoneValue := styles.NormalStyle.Render(m.issue.Milestone.Title)
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, milestoneLabel, " ", milestoneValue))
	}

	// Comments count
	commentsLabel := styles.MutedStyle.Render("Comments:")
	commentsValue := styles.NormalStyle.Render(fmt.Sprintf("%d", m.issue.Comments))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, commentsLabel, " ", commentsValue))

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderBody renders the issue body with markdown
func (m *IssueDetailView) renderBody() string {
	if m.issue.Body == "" {
		return styles.MutedStyle.Render("No description provided.")
	}

	// Render markdown
	rendered, err := m.renderer.Render(m.issue.Body)
	if err != nil {
		// Fallback to plain text if rendering fails
		return m.issue.Body
	}

	// Handle scrolling
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")

	// Calculate available height for body
	// Header (3 lines) + Metadata (~7 lines) + Separators (2) + Footer (2) = ~14 lines
	availableHeight := m.height - 14
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Apply scroll offset
	startLine := m.scrollOffset
	if startLine > len(lines) {
		startLine = len(lines)
	}

	endLine := startLine + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
		// Adjust scroll offset to not go beyond content
		if len(lines) > availableHeight {
			m.scrollOffset = len(lines) - availableHeight
			startLine = m.scrollOffset
		} else {
			m.scrollOffset = 0
			startLine = 0
		}
	}

	visibleLines := lines[startLine:endLine]

	// Add scroll indicator if needed
	scrollInfo := ""
	if len(lines) > availableHeight {
		scrollInfo = fmt.Sprintf("\n%s",
			styles.MutedStyle.Render(fmt.Sprintf("Line %d-%d of %d", startLine+1, endLine, len(lines))))
	}

	return strings.Join(visibleLines, "\n") + scrollInfo
}

// renderFooter renders the footer with help
func (m *IssueDetailView) renderFooter() string {
	helpItems := []string{
		styles.FormatKeyBinding("j/k", "scroll"),
		styles.FormatKeyBinding("o", "open in browser"),
		styles.FormatKeyBinding("q", "back"),
	}

	return styles.HelpStyle.Render(strings.Join(helpItems, " • "))
}

// renderLoading renders a loading state
func (m *IssueDetailView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading issue details...")
}

// renderError renders an error state
func (m *IssueDetailView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// formatTime formats a time to a readable string
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
