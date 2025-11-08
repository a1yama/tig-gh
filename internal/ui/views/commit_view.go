package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FetchCommitsUseCase defines the interface for fetching commits
type FetchCommitsUseCase interface {
	Execute(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error)
}

// commitsLoadedMsg is sent when commits are loaded
type commitsLoadedMsg struct {
	commits []*models.Commit
	err     error
}

// CommitView is the model for the commit list view
type CommitView struct {
	fetchCommitsUseCase FetchCommitsUseCase
	owner               string
	repo                string
	commits             []*models.Commit
	cursor              int
	loading             bool
	err                 error
	width               int
	height              int
	statusBar           *components.StatusBar
	showHelp            bool
	detailView          *CommitDetailView
	showingDetail       bool
}

// NewCommitView creates a new commit view
func NewCommitView() *CommitView {
	return &CommitView{
		fetchCommitsUseCase: nil,
		owner:               "",
		repo:                "",
		commits:             []*models.Commit{},
		cursor:              0,
		loading:             false,
		statusBar:           components.NewStatusBar(),
		showHelp:            false,
	}
}

// NewCommitViewWithUseCase creates a new commit view with UseCase
func NewCommitViewWithUseCase(fetchCommitsUseCase FetchCommitsUseCase, owner, repo string) *CommitView {
	return &CommitView{
		fetchCommitsUseCase: fetchCommitsUseCase,
		owner:               owner,
		repo:                repo,
		commits:             []*models.Commit{},
		cursor:              0,
		loading:             true, // Start in loading state
		statusBar:           components.NewStatusBar(),
		showHelp:            false,
	}
}

// Init initializes the commit view
func (m *CommitView) Init() tea.Cmd {
	if m.fetchCommitsUseCase != nil {
		return m.fetchCommits()
	}
	return nil
}

// Update handles messages
func (m *CommitView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case backMsg:
		// Return from detail view
		m.showingDetail = false
		m.detailView = nil
		return m, nil

	case tea.KeyMsg:
		keyStr := msg.String()

		// If showing detail view, check for back navigation first
		if m.showingDetail && m.detailView != nil {
			if keyStr == "q" || keyStr == "esc" {
				m.showingDetail = false
				m.detailView = nil
				return m, nil
			}
			// Otherwise delegate to detail view
			var cmd tea.Cmd
			updatedModel, cmd := m.detailView.Update(msg)
			m.detailView = updatedModel.(*CommitDetailView)
			return m, cmd
		}
		// Handle key press in list view
		return m.handleKeyPress(msg)

	case commitsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.commits = []*models.Commit{}
		} else {
			m.err = nil
			m.commits = msg.commits
			// Reset cursor if it's out of bounds
			if m.cursor >= len(m.commits) && len(m.commits) > 0 {
				m.cursor = len(m.commits) - 1
			} else if len(m.commits) == 0 {
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

// fetchCommits fetches commits from the API
func (m *CommitView) fetchCommits() tea.Cmd {
	return func() tea.Msg {
		if m.fetchCommitsUseCase == nil {
			return commitsLoadedMsg{
				commits: []*models.Commit{},
				err:     fmt.Errorf("fetch commits use case not initialized"),
			}
		}

		opts := &models.CommitOptions{
			PerPage: 100,
		}

		commits, err := m.fetchCommitsUseCase.Execute(context.Background(), m.owner, m.repo, opts)
		return commitsLoadedMsg{
			commits: commits,
			err:     err,
		}
	}
}

// handleKeyPress handles keyboard input
func (m *CommitView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Enter key using Type check for reliability
	if msg.Type == tea.KeyEnter {
		// View commit detail
		if len(m.commits) > 0 && m.cursor < len(m.commits) {
			selectedCommit := m.commits[m.cursor]
			m.detailView = NewCommitDetailView(selectedCommit)
			m.detailView.width = m.width
			m.detailView.height = m.height
			m.showingDetail = true
			// Return detail view's Init command to trigger immediate update
			return m, m.detailView.Init()
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
		// Refresh commits
		if !m.loading && m.fetchCommitsUseCase != nil {
			m.loading = true
			m.err = nil
			return m, m.fetchCommits()
		}
		return m, nil

	case "j", "down":
		if m.cursor < len(m.commits)-1 {
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
		if len(m.commits) > 0 {
			m.cursor = len(m.commits) - 1
		}
		return m, nil

	case "d":
		// View diff (to be implemented)
		return m, nil

	case "y":
		// Copy SHA to clipboard (to be implemented)
		return m, nil
	}

	return m, nil
}

// View renders the commit view
func (m *CommitView) View() string {
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

	// Commit list or error/loading state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else {
		s.WriteString(m.renderCommitList())
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
func (m *CommitView) renderHeader() string {
	title := styles.HeaderStyle.Render("Commits")
	count := styles.MutedStyle.Render(fmt.Sprintf("(%d)", len(m.commits)))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		" ",
		count,
	)
}

// renderCommitList renders the list of commits
func (m *CommitView) renderCommitList() string {
	var s strings.Builder

	// Calculate available height for list (total - header - status bar - margins)
	availableHeight := m.height - 4
	if m.showHelp {
		availableHeight -= 10 // Reserve space for help
	}

	// Calculate visible range
	startIdx := 0
	endIdx := len(m.commits)

	if len(m.commits) > availableHeight {
		// Show items around cursor
		halfHeight := availableHeight / 2
		startIdx = m.cursor - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.commits) {
			endIdx = len(m.commits)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible commits
	for i := startIdx; i < endIdx; i++ {
		commit := m.commits[i]
		line := m.renderCommitLine(commit, i)
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderCommitLine renders a single commit line
func (m *CommitView) renderCommitLine(commit *models.Commit, index int) string {
	// Cursor indicator
	cursor := "  "
	if m.cursor == index {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	// Commit graph symbol
	graph := styles.MutedStyle.Render("*")

	// SHA (short version - first 7 characters)
	sha := commit.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}
	shaStyle := styles.IssueNumberStyle
	if m.cursor == index {
		shaStyle = styles.SelectedStyle
	}
	shaText := shaStyle.Render(sha)

	// Message (first line only)
	message := commit.Message
	if idx := strings.Index(message, "\n"); idx != -1 {
		message = message[:idx]
	}
	// Truncate if too long
	maxMessageLen := m.width - 50
	if maxMessageLen < 20 {
		maxMessageLen = 20
	}
	if len(message) > maxMessageLen {
		message = message[:maxMessageLen-3] + "..."
	}
	messageStyle := styles.IssueTitleStyle
	if m.cursor == index {
		messageStyle = styles.SelectedStyle
	}
	messageText := messageStyle.Render(message)

	// Author
	author := styles.AuthorStyle.Render("@" + commit.Author.Name)

	// Date
	relativeTime := formatRelativeTime(commit.CreatedAt)
	date := styles.DateStyle.Render(relativeTime)

	// Combine all parts
	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		cursor,
		graph,
		" ",
		shaText,
		"  ",
		messageText,
		"  ",
		author,
		"  ",
		date,
	)

	return line
}

// renderLoading renders a loading state
func (m *CommitView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading commits...")
}

// renderError renders an error state
func (m *CommitView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// renderHelp renders the help section
func (m *CommitView) renderHelp() string {
	helpText := `
Navigation:
  ↑/k     Move up
  ↓/j     Move down
  g       Go to top
  G       Go to bottom

Actions:
  enter   View commit details
  d       View diff
  y       Copy SHA to clipboard
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
func (m *CommitView) updateStatusBar() {
	m.statusBar.ClearItems()

	// Set mode
	m.statusBar.SetMode("Commits")

	// Add current position
	if len(m.commits) > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits))
		m.statusBar.AddItem("", position)
	}

	// Add repository info
	if m.owner != "" && m.repo != "" {
		m.statusBar.AddItem("Repo", fmt.Sprintf("%s/%s", m.owner, m.repo))
	}
}
