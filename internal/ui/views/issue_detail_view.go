package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/ui/browser"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// backMsg is a message to go back to the previous view
type backMsg struct{}

// issueCommentsLoadedMsg is a message when comments are loaded
type issueCommentsLoadedMsg struct {
	comments []*models.Comment
	err      error
}

// IssueDetailView is the model for the issue detail view
type IssueDetailView struct {
	issue           *models.Issue
	comments        []*models.Comment
	commentsLoading bool
	commentsErr     error
	owner           string
	repo            string
	issueRepo       repository.IssueRepository
	scrollOffset    int
	loading         bool
	err             error
	width           int
	height          int
	renderer        *glamour.TermRenderer
}

// NewIssueDetailView creates a new issue detail view
func NewIssueDetailView(issue *models.Issue, owner, repo string, issueRepo repository.IssueRepository) *IssueDetailView {
	commentsLoading := issueRepo != nil
	return &IssueDetailView{
		issue:           issue,
		owner:           owner,
		repo:            repo,
		issueRepo:       issueRepo,
		scrollOffset:    0,
		loading:         false,
		commentsLoading: commentsLoading,
		renderer:        newMarkdownRenderer(80),
	}
}

// Init initializes the issue detail view
func (m *IssueDetailView) Init() tea.Cmd {
	if m.issueRepo != nil {
		return m.loadComments()
	}
	m.commentsLoading = false
	return nil
}

// loadComments loads comments for the issue
func (m *IssueDetailView) loadComments() tea.Cmd {
	return func() tea.Msg {
		if m.issueRepo == nil {
			return issueCommentsLoadedMsg{
				comments: nil,
				err:      fmt.Errorf("issue repository not available"),
			}
		}

		comments, err := m.issueRepo.ListComments(
			context.Background(),
			m.owner,
			m.repo,
			m.issue.Number,
			nil, // Use default options
		)

		return issueCommentsLoadedMsg{
			comments: comments,
			err:      err,
		}
	}
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

	case issueCommentsLoadedMsg:
		m.commentsLoading = false
		if msg.err != nil {
			m.commentsErr = msg.err
		} else {
			m.commentsErr = nil
			m.comments = msg.comments
		}
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
		_ = browser.Open(m.issue.HTMLURL)
		return m, nil
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

	// Build the full content first
	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader())
	content.WriteString("\n\n")

	// Metadata
	content.WriteString(m.renderMetadata())
	content.WriteString("\n\n")

	// Separator
	content.WriteString(styles.Separator(m.width - 4))
	content.WriteString("\n\n")

	// Body (without internal scrolling)
	content.WriteString(m.renderBodyContent())
	content.WriteString("\n\n")

	// Comments
	if len(m.comments) > 0 {
		content.WriteString(m.renderComments())
		content.WriteString("\n\n")
	} else if m.commentsLoading {
		content.WriteString(styles.MutedStyle.Render("Loading comments..."))
		content.WriteString("\n\n")
	} else if m.commentsErr != nil {
		content.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Failed to load comments: %v", m.commentsErr)))
		content.WriteString("\n\n")
	} else {
		content.WriteString(styles.MutedStyle.Render("No comments yet"))
		content.WriteString("\n\n")
	}

	// Apply scrolling to the entire content
	scrolledContent := m.applyScrolling(content.String())

	// Add footer
	var s strings.Builder
	s.WriteString(scrolledContent)
	s.WriteString("\n")
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

// renderBodyContent renders the issue body with markdown (without scrolling)
func (m *IssueDetailView) renderBodyContent() string {
	if m.issue.Body == "" {
		return styles.MutedStyle.Render("No description provided.")
	}

	// Render markdown
	rendered, err := m.renderer.Render(m.issue.Body)
	if err != nil {
		// Fallback to plain text if rendering fails
		return m.issue.Body
	}

	return strings.TrimRight(rendered, "\n")
}

// applyScrolling applies scrolling to the entire content
func (m *IssueDetailView) applyScrolling(content string) string {
	lines := strings.Split(content, "\n")

	// Calculate available height
	// Footer (1 line) + margin = ~2 lines
	availableHeight := m.height - 2
	if availableHeight < 5 {
		availableHeight = 5
	}

	// If content fits in the screen, no scrolling needed
	if len(lines) <= availableHeight {
		return content
	}

	// Apply scroll offset
	startLine := m.scrollOffset
	if startLine < 0 {
		startLine = 0
		m.scrollOffset = 0
	}

	maxOffset := len(lines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	if startLine > maxOffset {
		startLine = maxOffset
		m.scrollOffset = maxOffset
	}

	endLine := startLine + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleLines := lines[startLine:endLine]

	// Add scroll indicator
	scrollInfo := styles.MutedStyle.Render(fmt.Sprintf("[%d-%d/%d]", startLine+1, endLine, len(lines)))

	return strings.Join(visibleLines, "\n") + "\n" + scrollInfo
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
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}

// renderComments renders the comments section
func (m *IssueDetailView) renderComments() string {
	var s strings.Builder

	// Comments header
	commentsHeader := styles.BoldStyle.Render(fmt.Sprintf("Comments (%d)", len(m.comments)))
	s.WriteString(commentsHeader)
	s.WriteString("\n")
	s.WriteString(styles.Separator(m.width - 4))
	s.WriteString("\n\n")

	// Render each comment
	for i, comment := range m.comments {
		if i > 0 {
			s.WriteString("\n")
			s.WriteString(styles.MutedStyle.Render(strings.Repeat("─", m.width-4)))
			s.WriteString("\n\n")
		}

		// Comment author and time
		authorStyle := styles.BoldStyle
		author := authorStyle.Render(comment.User.Login)
		timeStr := styles.MutedStyle.Render(formatTime(comment.CreatedAt))

		s.WriteString(fmt.Sprintf("%s commented %s", author, timeStr))
		s.WriteString("\n\n")

		// Comment body (with markdown rendering)
		if m.renderer != nil && comment.Body != "" {
			rendered, err := m.renderer.Render(comment.Body)
			if err == nil {
				s.WriteString(rendered)
			} else {
				s.WriteString(comment.Body)
			}
		} else {
			s.WriteString(comment.Body)
		}
	}

	return s.String()
}
