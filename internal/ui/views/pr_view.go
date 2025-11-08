package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FetchPRsUseCase defines the interface for fetching pull requests
type FetchPRsUseCase interface {
	Execute(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error)
}

// prsLoadedMsg is sent when pull requests are loaded
type prsLoadedMsg struct {
	prs []*models.PullRequest
	err error
}

// PRView is the model for the pull request list view
type PRView struct {
	fetchPRsUseCase FetchPRsUseCase
	owner           string
	repo            string
	prs             []*models.PullRequest
	cursor          int
	selected        map[int]struct{}
	loading         bool
	err             error
	width           int
	height          int
	statusBar       *components.StatusBar
	showHelp        bool
	filterState     models.PRState
	detailView      *PRDetailView
	showingDetail   bool
}

// NewPRView creates a new PR view (for backward compatibility)
func NewPRView() *PRView {
	return &PRView{
		fetchPRsUseCase: nil,
		owner:           "",
		repo:            "",
		prs:             []*models.PullRequest{},
		cursor:          0,
		selected:        make(map[int]struct{}),
		loading:         false,
		statusBar:       components.NewStatusBar(),
		showHelp:        false,
		filterState:     models.PRStateOpen,
	}
}

// NewPRViewWithUseCase creates a new PR view with UseCase
func NewPRViewWithUseCase(fetchPRsUseCase FetchPRsUseCase, owner, repo string) *PRView {
	return &PRView{
		fetchPRsUseCase: fetchPRsUseCase,
		owner:           owner,
		repo:            repo,
		prs:             []*models.PullRequest{},
		cursor:          0,
		selected:        make(map[int]struct{}),
		loading:         true, // Start in loading state
		statusBar:       components.NewStatusBar(),
		showHelp:        false,
		filterState:     models.PRStateOpen,
	}
}

// Init initializes the PR view
func (m *PRView) Init() tea.Cmd {
	if m.fetchPRsUseCase != nil {
		return m.fetchPRs()
	}
	return nil
}

// Update handles messages
func (m *PRView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			m.detailView = updatedModel.(*PRDetailView)
			return m, cmd
		}
		// Handle key press in list view
		return m.handleKeyPress(msg)

	case prsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.prs = []*models.PullRequest{}
		} else {
			m.err = nil
			sorted := sortPullRequests(msg.prs)
			m.prs = sorted
			// Reset cursor if it's out of bounds
			if m.cursor >= len(m.prs) && len(m.prs) > 0 {
				m.cursor = len(m.prs) - 1
			} else if len(m.prs) == 0 {
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

// fetchPRs fetches pull requests from the API
func (m *PRView) fetchPRs() tea.Cmd {
	return func() tea.Msg {
		if m.fetchPRsUseCase == nil {
			return prsLoadedMsg{
				prs: []*models.PullRequest{},
				err: fmt.Errorf("fetch PRs use case not initialized"),
			}
		}

		opts := &models.PROptions{
			State:     m.filterState,
			Sort:      models.PRSortUpdated,
			Direction: models.SortDirectionDesc,
			PerPage:   100,
		}

		prs, err := m.fetchPRsUseCase.Execute(context.Background(), m.owner, m.repo, opts)
		return prsLoadedMsg{
			prs: prs,
			err: err,
		}
	}
}

// handleKeyPress handles keyboard input
func (m *PRView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	// Debug: log key press
	// fmt.Fprintf(os.Stderr, "[PRView.handleKeyPress] key=%s cursor=%d prs=%d\n", keyStr, m.cursor, len(m.prs))

	// Handle Enter key using Type check for reliability
	if msg.Type == tea.KeyEnter {
		// View PR detail
		if len(m.prs) > 0 && m.cursor < len(m.prs) {
			selectedPR := m.prs[m.cursor]
			m.detailView = NewPRDetailView(selectedPR)
			m.detailView.width = m.width
			m.detailView.height = m.height
			m.showingDetail = true
			// Return detail view's Init command to trigger immediate update
			return m, m.detailView.Init()
		}
		return m, nil
	}

	switch keyStr {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "r":
		// Refresh PRs
		if !m.loading && m.fetchPRsUseCase != nil {
			m.loading = true
			m.err = nil
			return m, m.fetchPRs()
		}
		return m, nil

	case "f":
		// Toggle filter between open, closed, all
		if !m.loading {
			switch m.filterState {
			case models.PRStateOpen:
				m.filterState = models.PRStateClosed
			case models.PRStateClosed:
				m.filterState = models.PRStateAll
			case models.PRStateAll:
				m.filterState = models.PRStateOpen
			}
			// Refresh with new filter
			if m.fetchPRsUseCase != nil {
				m.loading = true
				m.err = nil
				return m, m.fetchPRs()
			}
		}
		return m, nil

	case "j", "down":
		// fmt.Fprintf(os.Stderr, "[PRView] j/down pressed, cursor before=%d\n", m.cursor)
		if m.cursor < len(m.prs)-1 {
			m.cursor++
		}
		// fmt.Fprintf(os.Stderr, "[PRView] cursor after=%d\n", m.cursor)
		return m, nil

	case "k", "up":
		// fmt.Fprintf(os.Stderr, "[PRView] k/up pressed, cursor before=%d\n", m.cursor)
		if m.cursor > 0 {
			m.cursor--
		}
		// fmt.Fprintf(os.Stderr, "[PRView] cursor after=%d\n", m.cursor)
		return m, nil

	case "g":
		// Go to top
		m.cursor = 0
		return m, nil

	case "G":
		// Go to bottom
		if len(m.prs) > 0 {
			m.cursor = len(m.prs) - 1
		}
		return m, nil

	case "d":
		// View diff (to be implemented)
		return m, nil

	case "m":
		// Merge PR (to be implemented)
		// TODO: Add merge functionality with proper use case
		return m, nil
	}

	return m, nil
}

// View renders the PR view
func (m *PRView) View() string {
	// Debug: log view state
	// fmt.Fprintf(os.Stderr, "[PRView.View] width=%d height=%d loading=%v err=%v prs=%d cursor=%d\n",
	//	m.width, m.height, m.loading, m.err != nil, len(m.prs), m.cursor)

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

	// PR list or error/loading state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else {
		s.WriteString(m.renderPRList())
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
func (m *PRView) renderHeader() string {
	title := styles.HeaderStyle.Render("Pull Requests")
	count := styles.MutedStyle.Render(fmt.Sprintf("(%d)", len(m.prs)))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		" ",
		count,
	)
}

// renderPRList renders the list of pull requests
func (m *PRView) renderPRList() string {
	var s strings.Builder

	if len(m.prs) == 0 {
		emptyMsg := fmt.Sprintf("No pull requests (%s)", m.filterState)
		return styles.MutedStyle.Render(emptyMsg)
	}

	// Calculate available height for list (total - header - status bar - margins)
	availableHeight := m.height - 4
	if m.showHelp {
		availableHeight -= 10 // Reserve space for help
	}

	// Calculate visible range
	startIdx := 0
	endIdx := len(m.prs)

	if len(m.prs) > availableHeight {
		// Show items around cursor
		halfHeight := availableHeight / 2
		startIdx = m.cursor - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.prs) {
			endIdx = len(m.prs)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible PRs
	for i := startIdx; i < endIdx; i++ {
		pr := m.prs[i]
		line := m.renderPRLine(pr, i)
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderPRLine renders a single PR line
func (m *PRView) renderPRLine(pr *models.PullRequest, index int) string {
	// Cursor indicator
	cursor := "  "
	if m.cursor == index {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	// State badge
	var stateBadge string
	if pr.Draft {
		stateBadge = styles.MutedStyle.Render("● DRAFT")
	} else {
		switch pr.State {
		case models.PRStateOpen:
			if pr.Merged {
				stateBadge = styles.GetStateBadge("merged")
			} else {
				stateBadge = styles.GetStateBadge("open")
			}
		case models.PRStateClosed:
			if pr.Merged {
				stateBadge = styles.GetStateBadge("merged")
			} else {
				stateBadge = styles.GetStateBadge("closed")
			}
		default:
			stateBadge = styles.GetStateBadge(string(pr.State))
		}
	}

	// PR number
	number := styles.IssueNumberStyle.Render(fmt.Sprintf("#%-5d", pr.Number))

	// Title (with max width to prevent wrapping)
	titleStyle := styles.IssueTitleStyle
	if m.cursor == index {
		titleStyle = styles.SelectedStyle
	}
	// Calculate max width for title to prevent layout breaking
	// Reserve space for: cursor(3) + badge(10) + number(8) + spaces + metadata(~30)
	maxTitleWidth := m.width - 60
	if maxTitleWidth < 20 {
		maxTitleWidth = 20
	}
	titleText := pr.Title
	if len(titleText) > maxTitleWidth {
		titleText = titleText[:maxTitleWidth-3] + "..."
	}
	title := titleStyle.Render(titleText)

	// Review status
	approved, changesRequested, pending := m.countReviews(pr)
	reviewStatus := m.renderReviewStatus(approved, changesRequested, pending)

	// CI/CD status (placeholder - would need CI status data)
	// ciStatus := m.renderCIStatus(pr)

	// Mergeable status
	mergeableStatus := ""
	if pr.State == models.PRStateOpen && !pr.Draft {
		if pr.Mergeable {
			mergeableStatus = " " + styles.PRApprovedStyle.Render("✓")
		} else {
			mergeableStatus = " " + styles.PRChangesRequestedStyle.Render("✗")
		}
	}

	// Labels
	labels := ""
	if len(pr.Labels) > 0 {
		labelParts := []string{}
		for _, label := range pr.Labels {
			labelParts = append(labelParts, styles.LabelStyle.Render(label.Name))
		}
		labels = " " + strings.Join(labelParts, " ")
	}

	// Metadata (author, date)
	author := styles.AuthorStyle.Render("@" + pr.Author.Login)
	relativeTime := formatRelativeTime(pr.UpdatedAt)
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
		reviewStatus,
		mergeableStatus,
		" ",
		author,
		" ",
		date,
	)

	return line
}

// countReviews counts the number of approvals, change requests, and pending reviews
func (m *PRView) countReviews(pr *models.PullRequest) (approved, changesRequested, pending int) {
	for _, review := range pr.Reviews {
		switch review.State {
		case models.ReviewStateApproved:
			approved++
		case models.ReviewStateChangesRequested:
			changesRequested++
		case models.ReviewStatePending:
			pending++
		}
	}
	return
}

// renderReviewStatus renders the review status badges
func (m *PRView) renderReviewStatus(approved, changesRequested, pending int) string {
	var parts []string

	if approved > 0 {
		parts = append(parts, styles.PRApprovedStyle.Render(fmt.Sprintf("✓%d", approved)))
	}
	if changesRequested > 0 {
		parts = append(parts, styles.PRChangesRequestedStyle.Render(fmt.Sprintf("✗%d", changesRequested)))
	}
	if pending > 0 {
		parts = append(parts, styles.PRPendingStyle.Render(fmt.Sprintf("?%d", pending)))
	}

	if len(parts) == 0 {
		return ""
	}

	return " " + strings.Join(parts, " ")
}

// renderLoading renders a loading state
func (m *PRView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading pull requests...")
}

// renderError renders an error state
func (m *PRView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// renderHelp renders the help section
func (m *PRView) renderHelp() string {
	helpText := `
Navigation:
  ↑/k     Move up
  ↓/j     Move down
  g       Go to top
  G       Go to bottom

Actions:
  enter   View PR details
  d       View diff
  m       Merge PR
  r       Refresh
  f       Toggle filter (open/closed/all)

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
func (m *PRView) updateStatusBar() {
	m.statusBar.ClearItems()

	// Set mode based on filter state
	modeText := fmt.Sprintf("Pull Requests (%s)", m.filterState)
	m.statusBar.SetMode(modeText)

	// Add current position
	if len(m.prs) > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.prs))
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

func sortPullRequests(prs []*models.PullRequest) []*models.PullRequest {
	if len(prs) == 0 {
		return prs
	}

	sort.SliceStable(prs, func(i, j int) bool {
		pi := prs[i]
		pj := prs[j]

		if !pi.UpdatedAt.Equal(pj.UpdatedAt) {
			return pi.UpdatedAt.After(pj.UpdatedAt)
		}

		return pi.Number > pj.Number
	})

	return prs
}
