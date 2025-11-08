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

// FetchCommitDetailUseCase defines the interface for fetching commit details
type FetchCommitDetailUseCase interface {
	Execute(ctx context.Context, owner, repo, sha string) (*models.Commit, error)
}

// commitDetailLoadedMsg is sent when commit detail is loaded
type commitDetailLoadedMsg struct {
	commit *models.Commit
	err    error
}

// CommitDetailView is the model for the commit detail view
type CommitDetailView struct {
	fetchCommitDetailUseCase FetchCommitDetailUseCase
	owner                    string
	repo                     string
	sha                      string
	commit                   *models.Commit
	loading                  bool
	err                      error
	width                    int
	height                   int
	statusBar                *components.StatusBar
	showHelp                 bool
	scrollOffset             int
}

// NewCommitDetailView creates a new commit detail view with a commit
func NewCommitDetailView(commit *models.Commit) *CommitDetailView {
	return &CommitDetailView{
		fetchCommitDetailUseCase: nil,
		owner:                    "",
		repo:                     "",
		sha:                      commit.SHA,
		commit:                   commit,
		loading:                  false,
		statusBar:                components.NewStatusBar(),
		showHelp:                 false,
		scrollOffset:             0,
	}
}

// NewCommitDetailViewEmpty creates a new empty commit detail view
func NewCommitDetailViewEmpty() *CommitDetailView {
	return &CommitDetailView{
		fetchCommitDetailUseCase: nil,
		owner:                    "",
		repo:                     "",
		sha:                      "",
		commit:                   nil,
		loading:                  false,
		statusBar:                components.NewStatusBar(),
		showHelp:                 false,
		scrollOffset:             0,
	}
}

// NewCommitDetailViewWithUseCase creates a new commit detail view with UseCase
func NewCommitDetailViewWithUseCase(fetchCommitDetailUseCase FetchCommitDetailUseCase, owner, repo, sha string) *CommitDetailView {
	return &CommitDetailView{
		fetchCommitDetailUseCase: fetchCommitDetailUseCase,
		owner:                    owner,
		repo:                     repo,
		sha:                      sha,
		commit:                   nil,
		loading:                  true, // Start in loading state
		statusBar:                components.NewStatusBar(),
		showHelp:                 false,
		scrollOffset:             0,
	}
}

// Init initializes the commit detail view
func (m *CommitDetailView) Init() tea.Cmd {
	if m.fetchCommitDetailUseCase != nil {
		return m.fetchCommitDetail()
	}
	return nil
}

// Update handles messages
func (m *CommitDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commitDetailLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.commit = nil
		} else {
			m.err = nil
			m.commit = msg.commit
		}
		return m, nil

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

// fetchCommitDetail fetches commit detail from the API
func (m *CommitDetailView) fetchCommitDetail() tea.Cmd {
	return func() tea.Msg {
		if m.fetchCommitDetailUseCase == nil {
			return commitDetailLoadedMsg{
				commit: nil,
				err:    fmt.Errorf("fetch commit detail use case not initialized"),
			}
		}

		commit, err := m.fetchCommitDetailUseCase.Execute(context.Background(), m.owner, m.repo, m.sha)
		return commitDetailLoadedMsg{
			commit: commit,
			err:    err,
		}
	}
}

// handleKeyPress handles keyboard input
func (m *CommitDetailView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

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

	case "ctrl+d":
		// Page down
		m.scrollOffset += 10
		return m, nil

	case "ctrl+u":
		// Page up
		m.scrollOffset -= 10
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil

	case "g":
		// Go to top
		m.scrollOffset = 0
		return m, nil

	case "G":
		// Go to bottom (simplified)
		m.scrollOffset = 100
		return m, nil
	}

	return m, nil
}

// View renders the commit detail view
func (m *CommitDetailView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	// Header
	header := m.renderHeader()
	s.WriteString(header)
	s.WriteString("\n")

	// Commit detail or error/loading state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else {
		s.WriteString(m.renderCommitDetail())
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
func (m *CommitDetailView) renderHeader() string {
	shortSHA := m.sha
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}
	title := styles.HeaderStyle.Render(fmt.Sprintf("Commit %s", shortSHA))

	return title
}

// renderCommitDetail renders the commit detail
func (m *CommitDetailView) renderCommitDetail() string {
	if m.commit == nil {
		return styles.MutedStyle.Render("No commit data")
	}

	var s strings.Builder

	// Message
	s.WriteString(styles.IssueTitleStyle.Render(m.commit.Message))
	s.WriteString("\n\n")

	// Metadata
	s.WriteString(styles.MutedStyle.Render("Author:   "))
	s.WriteString(styles.AuthorStyle.Render(fmt.Sprintf("%s <%s>", m.commit.Author.Name, m.commit.Author.Email)))
	s.WriteString("\n")

	s.WriteString(styles.MutedStyle.Render("Date:     "))
	s.WriteString(styles.DateStyle.Render(m.commit.Author.Date.Format("2006-01-02 15:04:05 -0700")))
	s.WriteString("\n")

	s.WriteString(styles.MutedStyle.Render("SHA:      "))
	s.WriteString(styles.IssueNumberStyle.Render(m.commit.SHA))
	s.WriteString("\n")

	if len(m.commit.Parents) > 0 {
		s.WriteString(styles.MutedStyle.Render("Parents:  "))
		for i, parent := range m.commit.Parents {
			if i > 0 {
				s.WriteString(", ")
			}
			shortParent := parent
			if len(shortParent) > 7 {
				shortParent = shortParent[:7]
			}
			s.WriteString(styles.IssueNumberStyle.Render(shortParent))
		}
		s.WriteString("\n")
	}

	// Stats
	if m.commit.Stats != nil {
		s.WriteString("\n")
		s.WriteString(styles.MutedStyle.Render("Changes:  "))
		s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("+%d", m.commit.Stats.Additions)))
		s.WriteString(" ")
		s.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("-%d", m.commit.Stats.Deletions)))
		s.WriteString("\n")
	}

	// Files
	if len(m.commit.Files) > 0 {
		s.WriteString("\n")
		s.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("Files Changed (%d)", len(m.commit.Files))))
		s.WriteString("\n")

		for _, file := range m.commit.Files {
			s.WriteString(m.renderFile(file))
			s.WriteString("\n")
		}
	}

	return s.String()
}

// renderFile renders a single file change
func (m *CommitDetailView) renderFile(file *models.DiffFile) string {
	var statusIcon string
	var statusStyle lipgloss.Style

	switch file.Status {
	case models.FileStatusAdded:
		statusIcon = "+"
		statusStyle = styles.SuccessStyle
	case models.FileStatusModified:
		statusIcon = "M"
		statusStyle = styles.WarningStyle
	case models.FileStatusRemoved:
		statusIcon = "-"
		statusStyle = styles.ErrorStyle
	case models.FileStatusRenamed:
		statusIcon = "R"
		statusStyle = styles.InfoStyle
	default:
		statusIcon = "?"
		statusStyle = styles.MutedStyle
	}

	status := statusStyle.Render(fmt.Sprintf("[%s]", statusIcon))
	filename := styles.IssueTitleStyle.Render(file.Filename)
	changes := styles.MutedStyle.Render(fmt.Sprintf("+%d -%d", file.Additions, file.Deletions))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		status,
		" ",
		filename,
		"  ",
		changes,
	)
}

// renderLoading renders a loading state
func (m *CommitDetailView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading commit details...")
}

// renderError renders an error state
func (m *CommitDetailView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// renderHelp renders the help section
func (m *CommitDetailView) renderHelp() string {
	helpText := `
Navigation:
  ↑/k       Scroll up
  ↓/j       Scroll down
  ctrl+u    Page up
  ctrl+d    Page down
  g         Go to top
  G         Go to bottom

General:
  ?         Toggle help
  q         Back to list
  ctrl+c    Force quit
`

	return styles.BorderStyle.Render(
		styles.HelpStyle.Render(strings.TrimSpace(helpText)),
	)
}

// updateStatusBar updates the status bar with current state
func (m *CommitDetailView) updateStatusBar() {
	m.statusBar.ClearItems()

	// Set mode
	m.statusBar.SetMode("Commit Detail")

	// Add repository info
	if m.owner != "" && m.repo != "" {
		m.statusBar.AddItem("Repo", fmt.Sprintf("%s/%s", m.owner, m.repo))
	}

	// Add file count
	if m.commit != nil && len(m.commit.Files) > 0 {
		m.statusBar.AddItem("Files", fmt.Sprintf("%d", len(m.commit.Files)))
	}
}
