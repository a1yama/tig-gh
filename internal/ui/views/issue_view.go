package views

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/application/usecase"
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// issuesLoadedMsg is sent when issues are loaded
type issuesLoadedMsg struct {
	issues []*models.Issue
	err    error
}

// forceRenderMsg forces Bubble Tea to re-render
type forceRenderMsg struct{}

// IssueView is the model for the issue list view
type IssueView struct {
	fetchIssuesUseCase usecase.FetchIssuesUseCase
	owner              string
	repo               string
	issues             []*models.Issue
	cursor             int
	selected           map[int]struct{}
	loading            bool
	err                error
	width              int
	height             int
	statusBar          *components.StatusBar
	showHelp           bool
	filterState        models.IssueState
	detailView         *IssueDetailView
	showingDetail      bool
}

// NewIssueView creates a new issue view (for backward compatibility)
func NewIssueView() *IssueView {
	return &IssueView{
		fetchIssuesUseCase: nil,
		owner:              "",
		repo:               "",
		issues:             []*models.Issue{},
		cursor:             0,
		selected:           make(map[int]struct{}),
		loading:            false,
		statusBar:          components.NewStatusBar(),
		showHelp:           false,
		filterState:        models.IssueStateOpen,
	}
}

// NewIssueViewWithUseCase creates a new issue view with UseCase
func NewIssueViewWithUseCase(fetchIssuesUseCase usecase.FetchIssuesUseCase, owner, repo string) *IssueView {
	return &IssueView{
		fetchIssuesUseCase: fetchIssuesUseCase,
		owner:              owner,
		repo:               repo,
		issues:             []*models.Issue{},
		cursor:             0,
		selected:           make(map[int]struct{}),
		loading:            true, // Start in loading state
		statusBar:          components.NewStatusBar(),
		showHelp:           false,
		filterState:        models.IssueStateOpen,
	}
}

// Init initializes the issue view
func (m *IssueView) Init() tea.Cmd {
	if m.fetchIssuesUseCase != nil {
		return m.fetchIssues()
	}
	return nil
}

// Update handles messages
func (m *IssueView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	debugFile, _ := os.OpenFile("/tmp/tig-gh-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if debugFile != nil {
		defer debugFile.Close()
	}

	switch msg := msg.(type) {
	case forceRenderMsg:
		// No-op: just used to trigger a rerender
		return m, nil

	case backMsg:
		// Return from detail view
		m.showingDetail = false
		m.detailView = nil
		return m, nil

	case tea.KeyMsg:
		keyStr := msg.String()
		if isTerminalResponse(keyStr) {
			return m, nil
		}
		if debugFile != nil {
			fmt.Fprintf(debugFile, "[Update] KeyMsg=%s Type=%d showingDetail=%v\n", keyStr, msg.Type, m.showingDetail)
		}

		// Ignore unknown/unhandled key types (like terminal responses)
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 0 {
			if debugFile != nil {
				fmt.Fprintf(debugFile, "  -> Ignoring empty runes key\n")
			}
			return m, nil
		}

		// If showing detail view, check for back navigation first
		if m.showingDetail && m.detailView != nil {
			if debugFile != nil {
				fmt.Fprintf(debugFile, "  -> In detail view\n")
			}
			if keyStr == "q" || keyStr == "esc" {
				m.showingDetail = false
				m.detailView = nil
				return m, nil
			}
			// Otherwise delegate to detail view
			var cmd tea.Cmd
			updatedModel, cmd := m.detailView.Update(msg)
			m.detailView = updatedModel.(*IssueDetailView)
			return m, cmd
		}
		// Handle key press in list view
		if debugFile != nil {
			fmt.Fprintf(debugFile, "  -> Calling handleKeyPress\n")
		}
		return m.handleKeyPress(msg)

	case issuesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.issues = []*models.Issue{}
		} else {
			m.err = nil
			m.issues = sortIssues(filterOutPullRequests(msg.issues))
			// Reset cursor if it's out of bounds
			if m.cursor >= len(m.issues) && len(m.issues) > 0 {
				m.cursor = len(m.issues) - 1
			} else if len(m.issues) == 0 {
				m.cursor = 0
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(msg.Width, 1)
		if m.detailView != nil {
			m.detailView.Update(msg)
		}
		return m, nil
	}

	return m, nil
}

// fetchIssues fetches issues from the API
func (m *IssueView) fetchIssues() tea.Cmd {
	return func() tea.Msg {
		if m.fetchIssuesUseCase == nil {
			return issuesLoadedMsg{
				issues: []*models.Issue{},
				err:    fmt.Errorf("fetch issues use case not initialized"),
			}
		}

		opts := &models.IssueOptions{
			State:     m.filterState,
			Sort:      models.IssueSortUpdated,
			Direction: models.SortDirectionDesc,
			PerPage:   100,
		}

		issues, err := m.fetchIssuesUseCase.Execute(context.Background(), m.owner, m.repo, opts)
		return issuesLoadedMsg{
			issues: issues,
			err:    err,
		}
	}
}

// handleKeyPress handles keyboard input
func (m *IssueView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	debugFile, _ := os.OpenFile("/tmp/tig-gh-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if debugFile != nil {
		defer debugFile.Close()
	}

	// Handle Enter key using Type check for reliability
	if msg.Type == tea.KeyEnter {
		if debugFile != nil {
			fmt.Fprintf(debugFile, "[handleKeyPress] Enter detected! cursor=%d issues=%d\n", m.cursor, len(m.issues))
		}
		// View issue detail
		if len(m.issues) > 0 && m.cursor < len(m.issues) {
			selectedIssue := m.issues[m.cursor]
			m.detailView = NewIssueDetailView(selectedIssue)
			m.detailView.width = m.width
			m.detailView.height = m.height
			m.showingDetail = true
			if debugFile != nil {
				fmt.Fprintf(debugFile, "  -> Detail view created, showingDetail=%v\n", m.showingDetail)
			}
			return m, tea.Batch(
				m.detailView.Init(),
				func() tea.Msg { return forceRenderMsg{} },
			)
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "r":
		// Refresh issues
		if !m.loading && m.fetchIssuesUseCase != nil {
			m.loading = true
			m.err = nil
			return m, m.fetchIssues()
		}
		return m, nil

	case "f":
		// Toggle filter between open, closed, all
		if !m.loading {
			switch m.filterState {
			case models.IssueStateOpen:
				m.filterState = models.IssueStateClosed
			case models.IssueStateClosed:
				m.filterState = models.IssueStateAll
			case models.IssueStateAll:
				m.filterState = models.IssueStateOpen
			}
			// Refresh with new filter
			if m.fetchIssuesUseCase != nil {
				m.loading = true
				m.err = nil
				return m, m.fetchIssues()
			}
		}
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

	case " ":
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
	debugFile, _ := os.OpenFile("/tmp/tig-gh-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if debugFile != nil {
		fmt.Fprintf(debugFile, "[View] showingDetail=%v detailView=%v\n", m.showingDetail, m.detailView != nil)
		debugFile.Close()
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// If showing detail view, render it
	if m.showingDetail && m.detailView != nil {
		return m.detailView.View()
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
func (m *IssueView) renderIssueLine(issue *models.Issue, index int) string {
	// Cursor indicator
	cursor := "  "
	if m.cursor == index {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	// State badge
	stateBadge := styles.GetStateBadge(string(issue.State))

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
			labelParts = append(labelParts, styles.LabelStyle.Render(label.Name))
		}
		labels = " " + strings.Join(labelParts, " ")
	}

	// Metadata (author, comments, date)
	author := styles.AuthorStyle.Render("@" + issue.Author.Login)
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

	// Set mode based on filter state
	modeText := fmt.Sprintf("Issues (%s)", m.filterState)
	m.statusBar.SetMode(modeText)

	// Add current position
	if len(m.issues) > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.issues))
		m.statusBar.AddItem("", position)
	}

	// Add selection count if any
	if len(m.selected) > 0 {
		m.statusBar.AddItem("Selected", fmt.Sprintf("%d", len(m.selected)))
	}

	// Add repository info
	if m.owner != "" && m.repo != "" {
		m.statusBar.AddItem("Repo", fmt.Sprintf("%s/%s", m.owner, m.repo))
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

func filterOutPullRequests(issues []*models.Issue) []*models.Issue {
	if len(issues) == 0 {
		return issues
	}

	filtered := make([]*models.Issue, 0, len(issues))
	for _, issue := range issues {
		if issue == nil {
			continue
		}
		if strings.Contains(issue.HTMLURL, "/pull/") {
			continue
		}
		filtered = append(filtered, issue)
	}

	return filtered
}

func sortIssues(issues []*models.Issue) []*models.Issue {
	if len(issues) == 0 {
		return issues
	}

	sort.SliceStable(issues, func(i, j int) bool {
		left := issues[i]
		right := issues[j]

		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}

		return left.Number > right.Number
	})

	return issues
}
