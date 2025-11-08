package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// prTab represents the current tab in PR detail view
type prTab int

const (
	tabOverview prTab = iota
	tabFiles
	tabCommits
	tabComments
)

// mergeMsg is a message to merge the PR
type mergeMsg struct {
	pr *models.PullRequest
}

// diffMsg is a message to show diff
type diffMsg struct {
	pr *models.PullRequest
}

// prCommentsLoadedMsg is a message when comments are loaded
type prCommentsLoadedMsg struct {
	comments []*models.Comment
	err      error
}

// PRDetailView is the model for the PR detail view
type PRDetailView struct {
	pr              *models.PullRequest
	comments        []*models.Comment
	commentsLoading bool
	owner           string
	repo            string
	prRepo          repository.PullRequestRepository
	currentTab      prTab
	scrollOffset    int
	loading         bool
	err             error
	width           int
	height          int
	renderer        *glamour.TermRenderer
}

// NewPRDetailView creates a new PR detail view
func NewPRDetailView(pr *models.PullRequest, owner, repo string, prRepo repository.PullRequestRepository) *PRDetailView {
	return &PRDetailView{
		pr:              pr,
		owner:           owner,
		repo:            repo,
		prRepo:          prRepo,
		currentTab:      tabOverview,
		scrollOffset:    0,
		loading:         false,
		commentsLoading: true, // Start loading comments
		renderer:        newMarkdownRenderer(80),
	}
}

// Init initializes the PR detail view
func (m *PRDetailView) Init() tea.Cmd {
	if m.prRepo != nil {
		return m.loadComments()
	}
	return nil
}

// loadComments loads comments for the PR
func (m *PRDetailView) loadComments() tea.Cmd {
	return func() tea.Msg {
		if m.prRepo == nil {
			return prCommentsLoadedMsg{
				comments: nil,
				err:      fmt.Errorf("PR repository not available"),
			}
		}

		comments, err := m.prRepo.ListComments(
			context.Background(),
			m.owner,
			m.repo,
			m.pr.Number,
			nil, // Use default options
		)

		return prCommentsLoadedMsg{
			comments: comments,
			err:      err,
		}
	}
}

// Update handles messages
func (m *PRDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case prCommentsLoadedMsg:
		m.commentsLoading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.comments = msg.comments
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *PRDetailView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q", "esc":
		// Go back to PR list
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

	case "1":
		// Switch to overview tab
		m.currentTab = tabOverview
		m.scrollOffset = 0
		return m, nil

	case "2":
		// Switch to files tab
		m.currentTab = tabFiles
		m.scrollOffset = 0
		return m, nil

	case "3":
		// Switch to commits tab
		m.currentTab = tabCommits
		m.scrollOffset = 0
		return m, nil

	case "4":
		// Switch to comments tab
		m.currentTab = tabComments
		m.scrollOffset = 0
		return m, nil

	case "m":
		// Merge PR
		return m, func() tea.Msg {
			return mergeMsg{pr: m.pr}
		}

	case "d":
		// Show diff
		return m, func() tea.Msg {
			return diffMsg{pr: m.pr}
		}

	case "o":
		// Open in browser
		return m, func() tea.Msg {
			return openBrowserMsg{url: fmt.Sprintf("https://github.com/owner/repo/pull/%d", m.pr.Number)}
		}
	}

	return m, nil
}

// View renders the PR detail view
func (m *PRDetailView) View() string {
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

	// Tab navigation
	s.WriteString(m.renderTabNavigation())
	s.WriteString("\n\n")

	// Separator
	s.WriteString(styles.Separator(m.width - 4))
	s.WriteString("\n\n")

	// Tab content
	s.WriteString(m.renderTabContent())
	s.WriteString("\n\n")

	// Footer with help
	s.WriteString(m.renderFooter())

	return s.String()
}

// renderHeader renders the PR header
func (m *PRDetailView) renderHeader() string {
	// PR number and state
	numberStyle := styles.IssueNumberStyle
	number := numberStyle.Render(fmt.Sprintf("PR #%d", m.pr.Number))

	stateBadge := styles.GetStateBadge(string(m.pr.State))

	// Draft badge
	var draftBadge string
	if m.pr.Draft {
		draftBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("237")).
			Padding(0, 1).
			Render("Draft")
	}

	// Merged badge
	var mergedBadge string
	if m.pr.Merged {
		mergedBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("93")).
			Padding(0, 1).
			Render("Merged")
	}

	// Title
	titleStyle := styles.BoldStyle
	title := titleStyle.Render(m.pr.Title)

	headerParts := []string{number, " ", stateBadge}
	if draftBadge != "" {
		headerParts = append(headerParts, " ", draftBadge)
	}
	if mergedBadge != "" {
		headerParts = append(headerParts, " ", mergedBadge)
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		title,
	)
}

// renderMetadata renders PR metadata
func (m *PRDetailView) renderMetadata() string {
	var parts []string

	// Author
	authorLabel := styles.MutedStyle.Render("Author:")
	authorValue := styles.AuthorStyle.Render("@" + m.pr.Author.Login)
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, authorLabel, " ", authorValue))

	// Base and Head branches
	branchLabel := styles.MutedStyle.Render("Base:")
	branchValue := styles.NormalStyle.Render(m.pr.Base.Name + " ← " + m.pr.Head.Name)
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, branchLabel, " ", branchValue))

	// Status
	statusLabel := styles.MutedStyle.Render("Status:")
	statusValue := m.getMergeStatus()
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, statusLabel, " ", statusValue))

	// Created date
	createdLabel := styles.MutedStyle.Render("Created:")
	createdValue := styles.DateStyle.Render(formatTime(m.pr.CreatedAt))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, createdLabel, " ", createdValue))

	// Updated date
	updatedLabel := styles.MutedStyle.Render("Updated:")
	updatedValue := styles.DateStyle.Render(formatTime(m.pr.UpdatedAt))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, updatedLabel, " ", updatedValue))

	// Reviews
	if len(m.pr.Reviews) > 0 {
		reviewsLabel := styles.MutedStyle.Render("Reviews:")
		reviewsSummary := m.getReviewsSummary()
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, reviewsLabel, " ", reviewsSummary))
	}

	// Assignees
	if len(m.pr.Assignees) > 0 {
		assigneeNames := []string{}
		for _, assignee := range m.pr.Assignees {
			assigneeNames = append(assigneeNames, "@"+assignee.Login)
		}
		assigneesLabel := styles.MutedStyle.Render("Assignees:")
		assigneesValue := styles.AuthorStyle.Render(strings.Join(assigneeNames, ", "))
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, assigneesLabel, " ", assigneesValue))
	}

	// Labels
	if len(m.pr.Labels) > 0 {
		labelNames := []string{}
		for _, label := range m.pr.Labels {
			labelNames = append(labelNames, label.Name)
		}
		labelsLabel := styles.MutedStyle.Render("Labels:")
		labelsValue := styles.LabelStyle.Render(strings.Join(labelNames, ", "))
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, labelsLabel, " ", labelsValue))
	}

	// Milestone
	if m.pr.Milestone != nil {
		milestoneLabel := styles.MutedStyle.Render("Milestone:")
		milestoneValue := styles.NormalStyle.Render(m.pr.Milestone.Title)
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, milestoneLabel, " ", milestoneValue))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// getMergeStatus returns the merge status string
func (m *PRDetailView) getMergeStatus() string {
	if m.pr.Merged {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Render("✓ Merged")
	}

	if m.pr.Mergeable {
		approvedCount := 0
		changesRequestedCount := 0
		for _, review := range m.pr.Reviews {
			if review.State == models.ReviewStateApproved {
				approvedCount++
			} else if review.State == models.ReviewStateChangesRequested {
				changesRequestedCount++
			}
		}

		if changesRequestedCount > 0 {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Render("✗ Changes requested")
		}

		if approvedCount >= 2 {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Render("✓✓ Ready to merge")
		}

		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Render("⋯ Awaiting review")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Render("✗ Conflicts")
}

// getReviewsSummary returns a summary of reviews
func (m *PRDetailView) getReviewsSummary() string {
	var summary []string
	reviewCounts := make(map[models.ReviewState]int)

	for _, review := range m.pr.Reviews {
		reviewCounts[review.State]++
	}

	if count := reviewCounts[models.ReviewStateApproved]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Render(fmt.Sprintf("✓%d", count)))
	}

	if count := reviewCounts[models.ReviewStateChangesRequested]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("✗%d", count)))
	}

	if count := reviewCounts[models.ReviewStatePending]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Render(fmt.Sprintf("?%d", count)))
	}

	if len(summary) == 0 {
		return styles.MutedStyle.Render("No reviews")
	}

	return strings.Join(summary, " ")
}

// renderTabNavigation renders the tab navigation
func (m *PRDetailView) renderTabNavigation() string {
	tabs := []struct {
		name  string
		index prTab
	}{
		{"1: Overview", tabOverview},
		{"2: Files", tabFiles},
		{"3: Commits", tabCommits},
		{"4: Comments", tabComments},
	}

	var tabStrings []string
	for _, tab := range tabs {
		var style lipgloss.Style
		if tab.index == m.currentTab {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("99")).
				Padding(0, 1).
				Bold(true)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 1)
		}
		tabStrings = append(tabStrings, style.Render(tab.name))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...)
}

// renderTabContent renders the content of the current tab
func (m *PRDetailView) renderTabContent() string {
	switch m.currentTab {
	case tabOverview:
		return m.renderOverviewTab()
	case tabFiles:
		return m.renderFilesTab()
	case tabCommits:
		return m.renderCommitsTab()
	case tabComments:
		return m.renderCommentsTab()
	default:
		return ""
	}
}

// renderOverviewTab renders the overview tab
func (m *PRDetailView) renderOverviewTab() string {
	var s strings.Builder

	// PR body
	if m.pr.Body == "" {
		s.WriteString(styles.MutedStyle.Render("No description provided."))
	} else {
		s.WriteString(m.renderBody())
	}

	s.WriteString("\n\n")

	// Stats
	s.WriteString(m.renderStats())

	return m.applyScroll(s.String())
}

// renderBody renders the PR body with markdown
func (m *PRDetailView) renderBody() string {
	if m.pr.Body == "" {
		return styles.MutedStyle.Render("No description provided.")
	}

	// Render markdown
	rendered, err := m.renderer.Render(m.pr.Body)
	if err != nil {
		// Fallback to plain text if rendering fails
		return m.pr.Body
	}

	return strings.TrimRight(rendered, "\n")
}

// renderStats renders PR statistics
func (m *PRDetailView) renderStats() string {
	var parts []string

	// Files changed
	filesLabel := styles.MutedStyle.Render("Files Changed:")
	filesValue := styles.NormalStyle.Render(fmt.Sprintf("%d", m.pr.ChangedFiles))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, filesLabel, " ", filesValue))

	// Additions and deletions
	changesLabel := styles.MutedStyle.Render("Changes:")
	additions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("35")).
		Render(fmt.Sprintf("+%d", m.pr.Additions))
	deletions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Render(fmt.Sprintf("-%d", m.pr.Deletions))
	changesValue := lipgloss.JoinHorizontal(lipgloss.Top, additions, " ", deletions)
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, changesLabel, " ", changesValue))

	// Commits
	commitsLabel := styles.MutedStyle.Render("Commits:")
	commitsValue := styles.NormalStyle.Render(fmt.Sprintf("%d", m.pr.Commits))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, commitsLabel, " ", commitsValue))

	// Comments
	commentsLabel := styles.MutedStyle.Render("Comments:")
	commentsValue := styles.NormalStyle.Render(fmt.Sprintf("%d", m.pr.Comments))
	parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, commentsLabel, " ", commentsValue))

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderFilesTab renders the files tab
func (m *PRDetailView) renderFilesTab() string {
	content := fmt.Sprintf("Files Changed (%d)\n\n", m.pr.ChangedFiles)
	content += styles.MutedStyle.Render("File diff view will be implemented here.\n")
	content += styles.MutedStyle.Render(fmt.Sprintf("+%d -%d lines changed", m.pr.Additions, m.pr.Deletions))

	return m.applyScroll(content)
}

// renderCommitsTab renders the commits tab
func (m *PRDetailView) renderCommitsTab() string {
	content := fmt.Sprintf("Commits (%d)\n\n", m.pr.Commits)
	content += styles.MutedStyle.Render("Commit list will be implemented here.")

	return m.applyScroll(content)
}

// renderCommentsTab renders the comments tab
func (m *PRDetailView) renderCommentsTab() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Comments (%d)\n\n", len(m.comments)))

	if m.commentsLoading {
		s.WriteString(styles.MutedStyle.Render("Loading comments..."))
	} else if len(m.comments) == 0 {
		s.WriteString(styles.MutedStyle.Render("No comments yet."))
	} else {
		s.WriteString(m.renderCommentsList())
	}

	return m.applyScroll(s.String())
}

// renderCommentsList renders the list of comments
func (m *PRDetailView) renderCommentsList() string {
	var s strings.Builder

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

// applyScroll applies scrolling to content
func (m *PRDetailView) applyScroll(content string) string {
	lines := strings.Split(content, "\n")

	// Calculate available height for content
	// Header (3 lines) + Metadata (~8 lines) + Tabs (1 line) + Separators (2) + Footer (2) = ~16 lines
	availableHeight := m.height - 16
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
func (m *PRDetailView) renderFooter() string {
	helpItems := []string{
		styles.FormatKeyBinding("j/k", "scroll"),
		styles.FormatKeyBinding("1-4", "tabs"),
		styles.FormatKeyBinding("m", "merge"),
		styles.FormatKeyBinding("d", "diff"),
		styles.FormatKeyBinding("o", "open"),
		styles.FormatKeyBinding("q", "back"),
	}

	return styles.HelpStyle.Render(strings.Join(helpItems, " • "))
}

// renderLoading renders a loading state
func (m *PRDetailView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading PR details...")
}

// renderError renders an error state
func (m *PRDetailView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}
