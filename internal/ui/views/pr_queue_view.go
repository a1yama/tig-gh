package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// prQueueLoadedMsg is sent when review queue data is loaded.
type prQueueLoadedMsg struct {
	prs []*models.PullRequest
	err error
}

// prQueueReviewsLoadedMsg is sent after individual PR reviews are loaded.
type prQueueReviewsLoadedMsg struct {
	index   int
	reviews []models.Review
	err     error
}

// prQueueEntry keeps review metrics for a pull request in the queue.
type prQueueEntry struct {
	pr              *models.PullRequest
	reviews         []models.Review
	firstReviewAt   *time.Time
	firstApprovalAt *time.Time
	reviewsLoaded   bool
	reviewsErr      error
}

// PRQueueView shows open pull requests waiting for review or approval.
type PRQueueView struct {
	fetchPRsUseCase FetchPRsUseCase
	owner           string
	repo            string

	entries []*prQueueEntry
	cursor  int

	loading bool
	err     error

	width  int
	height int

	statusBar *components.StatusBar
	showHelp  bool

	showingDetail bool
	detailView    *PRDetailView

	prRepo          repository.PullRequestRepository
	reviewLoadIndex int
	reviewLoading   bool
}

// NewPRQueueView creates an empty queue view.
func NewPRQueueView() *PRQueueView {
	return &PRQueueView{
		entries:       []*prQueueEntry{},
		statusBar:     components.NewStatusBar(),
		prRepo:        nil,
		loading:       false,
		showHelp:      false,
		reviewLoading: false,
	}
}

// NewPRQueueViewWithUseCase wires the queue view with the fetch use case.
func NewPRQueueViewWithUseCase(fetchPRsUseCase FetchPRsUseCase, owner, repo string) *PRQueueView {
	view := NewPRQueueView()
	view.fetchPRsUseCase = fetchPRsUseCase
	view.owner = owner
	view.repo = repo
	if fetchPRsUseCase != nil {
		view.prRepo = fetchPRsUseCase.GetRepository()
		view.loading = true
	}
	return view
}

// Init starts loading PR metrics.
func (m *PRQueueView) Init() tea.Cmd {
	if m.fetchPRsUseCase != nil {
		m.loading = true
		return m.fetchPRs()
	}
	return nil
}

func (m *PRQueueView) fetchPRs() tea.Cmd {
	return func() tea.Msg {
		if m.fetchPRsUseCase == nil {
			return prQueueLoadedMsg{prs: nil, err: fmt.Errorf("fetch PRs use case not initialized")}
		}

		opts := &models.PROptions{
			State:     models.PRStateOpen,
			Sort:      models.PRSortCreated,
			Direction: models.SortDirectionAsc,
			PerPage:   100,
		}

		prs, err := m.fetchPRsUseCase.Execute(context.Background(), m.owner, m.repo, opts)
		return prQueueLoadedMsg{prs: prs, err: err}
	}
}

func (m *PRQueueView) loadReviewsForEntry(index int) tea.Cmd {
	if m.prRepo == nil || index >= len(m.entries) {
		return nil
	}
	entry := m.entries[index]
	owner := m.owner
	repo := m.repo
	number := entry.pr.Number

	return func() tea.Msg {
		reviews, err := m.prRepo.ListReviews(context.Background(), owner, repo, number)
		if err != nil {
			return prQueueReviewsLoadedMsg{index: index, err: err}
		}
		return prQueueReviewsLoadedMsg{index: index, reviews: flattenReviews(reviews)}
	}
}

// Update handles Bubble Tea messages.
func (m *PRQueueView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showingDetail && m.detailView != nil {
		if _, isBack := msg.(backMsg); isBack {
			m.showingDetail = false
			m.detailView = nil
			return m, nil
		}

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			keyStr := keyMsg.String()
			if keyStr == "q" || keyStr == "esc" {
				m.showingDetail = false
				m.detailView = nil
				return m, nil
			}
		}

		var cmd tea.Cmd
		updated, cmd := m.detailView.Update(msg)
		m.detailView = updated.(*PRDetailView)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case prQueueLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.entries = []*prQueueEntry{}
			return m, nil
		}
		m.err = nil
		m.entries = make([]*prQueueEntry, 0, len(msg.prs))
		for _, pr := range msg.prs {
			ensurePRNumber(pr)
			m.entries = append(m.entries, &prQueueEntry{pr: pr})
		}
		sort.SliceStable(m.entries, func(i, j int) bool {
			return m.entries[i].pr.CreatedAt.Before(m.entries[j].pr.CreatedAt)
		})
		m.cursor = 0
		m.reviewLoadIndex = 0
		if m.prRepo != nil && len(m.entries) > 0 {
			m.reviewLoading = true
			return m, m.loadReviewsForEntry(0)
		}
		m.reviewLoading = false
		return m, nil

	case prQueueReviewsLoadedMsg:
		if msg.index < len(m.entries) {
			entry := m.entries[msg.index]
			entry.reviewsLoaded = true
			entry.reviewsErr = msg.err
			if msg.err == nil {
				entry.reviews = msg.reviews
				entry.firstReviewAt = firstReviewSubmittedAt(entry.reviews)
				entry.firstApprovalAt = firstApprovalSubmittedAt(entry.reviews)
			}
		}
		m.reviewLoadIndex = msg.index + 1
		if m.reviewLoadIndex < len(m.entries) {
			return m, m.loadReviewsForEntry(m.reviewLoadIndex)
		}
		m.reviewLoading = false
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

func (m *PRQueueView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "r":
		if !m.loading && m.fetchPRsUseCase != nil {
			m.loading = true
			m.err = nil
			return m, m.fetchPRs()
		}
		return m, nil
	case "j", "down":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
		return m, nil
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "g":
		m.cursor = 0
		return m, nil
	case "G":
		if len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
		return m, nil
	}

	if msg.Type == tea.KeyEnter {
		if len(m.entries) > 0 && m.cursor < len(m.entries) {
			selected := m.entries[m.cursor].pr
			m.detailView = NewPRDetailView(selected, m.owner, m.repo, m.prRepo)
			m.detailView.width = m.width
			m.detailView.height = m.height
			m.showingDetail = true
			return m, m.detailView.Init()
		}
	}

	return m, nil
}

// View renders the queue view.
func (m *PRQueueView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.showingDetail && m.detailView != nil {
		return m.detailView.View()
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.renderLoading())
	} else if m.err != nil {
		b.WriteString(m.renderError())
	} else {
		b.WriteString(m.renderQueueList())
	}

	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.renderHelp())
	}

	b.WriteString("\n")
	m.updateStatusBar()
	b.WriteString(m.statusBar.View())

	return b.String()
}

func (m *PRQueueView) renderHeader() string {
	title := styles.HeaderStyle.Render("Review Queue")
	count := styles.MutedStyle.Render(fmt.Sprintf("(%d)", len(m.entries)))
	return lipgloss.JoinHorizontal(lipgloss.Top, title, " ", count)
}

func (m *PRQueueView) renderQueueList() string {
	if len(m.entries) == 0 {
		return styles.MutedStyle.Render("No open pull requests.")
	}

	var b strings.Builder
	availableHeight := m.height - 4
	if availableHeight < 3 {
		availableHeight = 3
	}
	if m.showHelp {
		availableHeight -= 5
	}

	startIdx := 0
	endIdx := len(m.entries)
	if len(m.entries) > availableHeight {
		half := availableHeight / 2
		startIdx = m.cursor - half
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.entries) {
			endIdx = len(m.entries)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		entry := m.entries[i]
		b.WriteString(m.renderEntry(entry, i))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m *PRQueueView) renderEntry(entry *prQueueEntry, index int) string {
	selected := m.cursor == index
	cursor := "  "
	if selected {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	now := time.Now()
	waitingDuration := now.Sub(entry.pr.CreatedAt)
	waitingStyle := waitingDurationStyle(waitingDuration)
	waitingLabel := waitingStyle.Render(formatDurationShort(waitingDuration))

	statusText := entry.statusLabel()
	prNum, ok := prDisplayNumber(entry.pr)
	titleText := entry.pr.Title
	if titleText == "" {
		titleText = styles.MutedStyle.Render("(no title)")
	}
	var title string
	if ok {
		title = styles.IssueTitleStyle.Render(fmt.Sprintf("#%d %s", prNum, titleText))
	} else {
		title = styles.IssueTitleStyle.Render(titleText)
	}
	line := lipgloss.JoinHorizontal(lipgloss.Top, waitingLabel, " • ", statusText, " • ", title)

	details := m.renderEntryDetails(entry, now)
	body := lipgloss.JoinVertical(lipgloss.Left, line, details)

	var entryStyle lipgloss.Style
	if selected {
		entryStyle = styles.SelectedStyle.Copy().Padding(0, 1)
	} else {
		entryStyle = lipgloss.NewStyle().Padding(0, 1)
	}
	body = entryStyle.Render(body)
	return cursor + body
}

func (m *PRQueueView) renderEntryDetails(entry *prQueueEntry, now time.Time) string {
	author := styles.AuthorStyle.Render(formatAuthorHandle(entry.pr.Author))
	updated := styles.MutedStyle.Render(fmt.Sprintf("Updated %s ago", formatDurationShort(now.Sub(entry.pr.UpdatedAt))))
	waitingInfo := styles.MutedStyle.Render(fmt.Sprintf("Opened %s ago", formatDurationShort(now.Sub(entry.pr.CreatedAt))))
	reviewSummary := entry.reviewSummary()

	parts := []string{author, waitingInfo, updated}
	if reviewSummary != "" {
		parts = append(parts, reviewSummary)
	}

	return strings.Join(parts, "  ")
}

func (entry *prQueueEntry) reviewSummary() string {
	switch {
	case !entry.reviewsLoaded:
		if entry.reviewsErr != nil {
			prefix := styles.MutedStyle.Render("Reviews:")
			return lipgloss.JoinHorizontal(lipgloss.Top, prefix, " ", styles.ErrorStyle.Render("error"))
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, styles.MutedStyle.Render("Reviews:"), " ", styles.MutedStyle.Render("loading..."))
	default:
		if len(entry.reviews) == 0 {
			return lipgloss.JoinHorizontal(lipgloss.Top, styles.MutedStyle.Render("Reviews:"), " ", styles.MutedStyle.Render("none"))
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, styles.MutedStyle.Render("Reviews:"), " ", renderReviewSummary(entry.reviews))
	}
}

func (entry *prQueueEntry) statusLabel() string {
	switch {
	case !entry.reviewsLoaded:
		if entry.reviewsErr != nil {
			return styles.ErrorStyle.Render("Reviews error")
		}
		return styles.MutedStyle.Render("Loading reviews")
	case entry.firstReviewAt == nil:
		return styles.PRPendingStyle.Render("Awaiting review")
	case entry.firstApprovalAt == nil:
		return styles.WarningStyle.Render("Awaiting approval")
	default:
		return styles.PRApprovedStyle.Render("Approved")
	}
}

func waitingDurationStyle(d time.Duration) lipgloss.Style {
	switch {
	case d.Hours() >= 24*7:
		return lipgloss.NewStyle().Foreground(styles.ColorError).Bold(true)
	case d.Hours() >= 48:
		return styles.WarningStyle
	default:
		return styles.InfoStyle
	}
}

func (m *PRQueueView) renderHelp() string {
	helpItems := []string{
		styles.FormatKeyBinding("j/k", "navigate"),
		styles.FormatKeyBinding("enter", "open PR"),
		styles.FormatKeyBinding("r", "refresh"),
		styles.FormatKeyBinding("?", "help"),
	}
	return styles.HelpStyle.Render(strings.Join(helpItems, " • "))
}

func (m *PRQueueView) renderLoading() string {
	if m.reviewLoading {
		return styles.LoadingStyle.Render("Loading pull requests & reviews...")
	}
	return styles.LoadingStyle.Render("Loading pull requests...")
}

func (m *PRQueueView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

func (m *PRQueueView) updateStatusBar() {
	m.statusBar.SetMode("Queue")
	repoLabel := fmt.Sprintf("%s/%s", m.owner, m.repo)
	m.statusBar.SetItems([]components.StatusItem{
		{Key: "Repo", Value: repoLabel},
		{Key: "Open", Value: fmt.Sprintf("%d", len(m.entries))},
	})
	if m.reviewLoading {
		m.statusBar.SetMessage("Fetching review metrics...")
	} else {
		m.statusBar.SetMessage("")
	}
}
