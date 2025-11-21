package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v57/github"
)

// LeadTimeMetricsUseCase はメトリクス取得ユースケースの必要インターフェース
type LeadTimeMetricsUseCase interface {
	Execute(ctx context.Context, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error)
	GetRateLimit(ctx context.Context) (*github.Rate, error)
}

// MetricsExitMsg はメトリクスビューからの戻る要求を表す
type MetricsExitMsg struct{}

type metricsLoadedMsg struct {
	metrics   *models.LeadTimeMetrics
	rateLimit *github.Rate
	err       error
}

type metricsProgressMsg struct {
	progress models.MetricsProgress
}

type rateLimitFetchedMsg struct {
	rateLimit *github.Rate
	err       error
}

// MetricsView はリードタイムメトリクス表示用ビュー
type MetricsView struct {
	useCase     LeadTimeMetricsUseCase
	metrics     *models.LeadTimeMetrics
	loading     bool
	err         error
	width       int
	height      int
	scroll      int
	statusBar   *components.StatusBar
	lastUpdated time.Time
	rateLimit   *github.Rate // GitHub API rate limit info
	progress    *models.MetricsProgress
	progressCh  chan models.MetricsProgress
}

// NewMetricsView は空のメトリクスビューを返す
func NewMetricsView() *MetricsView {
	return &MetricsView{
		statusBar: components.NewStatusBar(),
		loading:   false,
		scroll:    0,
	}
}

// NewMetricsViewWithUseCase はユースケースをバインドしたビューを返す
func NewMetricsViewWithUseCase(useCase LeadTimeMetricsUseCase) *MetricsView {
	view := NewMetricsView()
	view.useCase = useCase
	return view
}

// Init は初期ロードを開始する
func (m *MetricsView) Init() tea.Cmd {
	if m.useCase == nil {
		return nil
	}
	m.loading = true
	m.err = nil
	m.progress = nil
	return m.fetchMetrics()
}

func (m *MetricsView) fetchMetrics() tea.Cmd {
	if m.useCase == nil {
		m.progressCh = nil
		return func() tea.Msg {
			return metricsLoadedMsg{metrics: nil, err: fmt.Errorf("metrics use case not initialized")}
		}
	}

	progressCh := make(chan models.MetricsProgress, 1)
	resultCh := make(chan metricsLoadedMsg, 1)
	m.progressCh = progressCh

	go func() {
		defer close(progressCh)

		progressFn := func(progress models.MetricsProgress) {
			select {
			case progressCh <- progress:
			default:
			}
		}

		metrics, err := m.useCase.Execute(context.Background(), progressFn)
		var rateLimit *github.Rate

		if err == nil {
			// Fetch rate limit info (best effort)
			rate, rateLimitErr := m.useCase.GetRateLimit(context.Background())
			if rateLimitErr == nil {
				rateLimit = rate
			}
		}

		resultCh <- metricsLoadedMsg{
			metrics:   metrics,
			rateLimit: rateLimit,
			err:       err,
		}
		close(resultCh)
	}()

	return tea.Batch(waitForMetrics(resultCh), m.listenForProgress(progressCh))
}

func waitForMetrics(ch <-chan metricsLoadedMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m *MetricsView) listenForProgress(ch <-chan models.MetricsProgress) tea.Cmd {
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		progress, ok := <-ch
		if !ok {
			return nil
		}
		return metricsProgressMsg{progress: progress}
	}
}

func (m *MetricsView) fetchRateLimitCmd() tea.Cmd {
	if m.useCase == nil {
		return func() tea.Msg {
			return rateLimitFetchedMsg{err: fmt.Errorf("metrics use case not initialized")}
		}
	}

	return func() tea.Msg {
		rate, err := m.useCase.GetRateLimit(context.Background())
		return rateLimitFetchedMsg{
			rateLimit: rate,
			err:       err,
		}
	}
}

// Update はBubble Teaメッセージを処理する
func (m *MetricsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case metricsLoadedMsg:
		m.loading = false
		m.rateLimit = msg.rateLimit
		m.progress = nil
		m.progressCh = nil
		if msg.err != nil {
			m.err = msg.err
			m.metrics = nil
			m.rateLimit = nil
		} else {
			m.err = nil
			m.metrics = msg.metrics
			m.lastUpdated = time.Now()
			m.scroll = 0
		}
		m.updateStatusBar()
		return m, nil

	case metricsProgressMsg:
		progress := msg.progress
		m.progress = &progress
		m.updateStatusBar()
		if m.loading {
			return m, m.listenForProgress(m.progressCh)
		}
		return m, nil

	case rateLimitFetchedMsg:
		if msg.err == nil {
			m.rateLimit = msg.rateLimit
		}
		m.updateStatusBar()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(m.width, 1)
		m.updateStatusBar()
		return m, nil
	}

	return m, nil
}

func (m *MetricsView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m, func() tea.Msg { return MetricsExitMsg{} }
	case "r":
		if !m.loading {
			m.loading = true
			m.err = nil
			m.progress = nil
			m.updateStatusBar()
			return m, m.fetchMetrics()
		}
		return m, nil
	case "l": // Show rate limit
		return m, m.fetchRateLimitCmd()
	case "j", "down":
		maxScroll := m.maxScroll()
		if m.scroll < maxScroll {
			m.scroll++
		}
		return m, nil
	case "k", "up":
		if m.scroll > 0 {
			m.scroll--
		}
		return m, nil
	case "g":
		m.scroll = 0
		return m, nil
	case "G":
		m.scroll = m.maxScroll()
		return m, nil
	}

	return m, nil
}

// View は現在のUI文字列を返す
func (m *MetricsView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing metrics view..."
	}

	contentLines := m.renderContentLines()
	availableHeight := m.height - 1
	if availableHeight < 1 {
		availableHeight = 1
	}

	maxScroll := len(contentLines) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}

	start := m.scroll
	end := start + availableHeight
	if end > len(contentLines) {
		end = len(contentLines)
	}

	body := strings.Join(contentLines[start:end], "\n")

	m.updateStatusBar()
	return lipgloss.JoinVertical(
		lipgloss.Left,
		body,
		m.statusBar.View(),
	)
}

func (m *MetricsView) renderContentLines() []string {
	lines := []string{
		styles.TitleStyle.Render("Lead Time Metrics"),
	}

	if m.lastUpdated.IsZero() {
		lines = append(lines, styles.MutedStyle.Render("No data fetched yet. Press 'r' to load metrics."))
	} else {
		lines = append(lines, styles.MutedStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdated.Format("2006-01-02 15:04:05"))))
	}

	lines = append(lines, "")

	if m.loading {
		lines = append(lines, styles.LoadingStyle.Render("Fetching lead time metrics..."))
		return lines
	}

	if m.err != nil {
		lines = append(lines, styles.ErrorStyle.Render(m.err.Error()))
		lines = append(lines, "")
		lines = append(lines, styles.HelpStyle.Render("Press 'r' to retry or 'q' to go back."))
		return lines
	}

	if m.metrics == nil {
		lines = append(lines, styles.WarningStyle.Render("Metrics data is not available."))
		lines = append(lines, "")
		lines = append(lines, styles.HelpStyle.Render("Ensure metrics are enabled in config."))
		return lines
	}

	lines = append(lines, m.renderOverallSection()...)
	lines = append(lines, "")
	lines = append(lines, m.renderStagnantPRSection()...)
	lines = append(lines, "")
	lines = append(lines, m.renderRepositorySection()...)
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("Controls: j/k scroll • r refresh • q back"))

	return lines
}

func (m *MetricsView) renderOverallSection() []string {
	stat := m.metrics.Overall
	lines := []string{
		styles.HeaderStyle.Render("Overall Metrics"),
		fmt.Sprintf("Average: %s", formatDuration(stat.Average)),
		fmt.Sprintf("Median: %s", formatDuration(stat.Median)),
		fmt.Sprintf("Total PRs: %d", stat.Count),
	}
	return lines
}

func (m *MetricsView) renderStagnantPRSection() []string {
	stagnant := m.metrics.StagnantPRs
	lines := []string{
		styles.HeaderStyle.Render(fmt.Sprintf("Stagnant PRs (Open > %s)", formatDuration(stagnant.Threshold))),
	}

	if stagnant.TotalStagnant == 0 {
		lines = append(lines, styles.MutedStyle.Render("No stagnant PRs found."))
		return lines
	}

	lines = append(lines,
		fmt.Sprintf("Total stagnant PRs:  %d", stagnant.TotalStagnant),
		fmt.Sprintf("Average age:         %s", formatDuration(stagnant.AverageAge)),
	)

	if stagnant.LongestWaiting != nil {
		lines = append(lines,
			fmt.Sprintf("Longest waiting:     %s (%s #%d)",
				formatDuration(stagnant.LongestWaiting.Age),
				stagnant.LongestWaiting.Repository,
				stagnant.LongestWaiting.Number,
			),
		)
	}

	return lines
}

func (m *MetricsView) renderRepositorySection() []string {
	lines := []string{
		styles.HeaderStyle.Render("Per Repository"),
	}

	if len(m.metrics.ByRepository) == 0 {
		lines = append(lines, styles.MutedStyle.Render("No repository data available."))
		return lines
	}

	repoNames := make([]string, 0, len(m.metrics.ByRepository))
	for name := range m.metrics.ByRepository {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	header := fmt.Sprintf("%-40s %12s %12s %6s", "Repository", "Avg", "Median", "PRs")
	lines = append(lines, styles.MutedStyle.Render(header))

	for _, name := range repoNames {
		stat := m.metrics.ByRepository[name]
		line := fmt.Sprintf(
			"%-40s %12s %12s %6d",
			name,
			formatDuration(stat.Average),
			formatDuration(stat.Median),
			stat.Count,
		)
		lines = append(lines, line)
	}

	return lines
}

func (m *MetricsView) updateStatusBar() {
	if m.statusBar == nil {
		m.statusBar = components.NewStatusBar()
	}

	m.statusBar.SetSize(m.width, 1)

	mode := "Metrics"
	switch {
	case m.loading:
		mode = "Loading"
	case m.err != nil:
		mode = "Error"
	}
	m.statusBar.SetMode(mode)

	var status string
	if m.loading {
		if m.progress != nil && m.progress.TotalRepos > 0 {
			status = fmt.Sprintf("Loading metrics... (%d/%d repositories)",
				m.progress.ProcessedRepos,
				m.progress.TotalRepos,
			)
			if repo := strings.TrimSpace(m.progress.CurrentRepo); repo != "" {
				status = fmt.Sprintf("%s • %s", status, repo)
			}
		} else {
			status = "Loading metrics..."
		}
		// Show rate limit even during loading
		if m.rateLimit != nil {
			status = fmt.Sprintf("%s • API: %d/%d remaining",
				status,
				m.rateLimit.Remaining,
				m.rateLimit.Limit,
			)
		}
	} else if m.err != nil {
		status = "Error loading metrics"
		if errMsg := strings.TrimSpace(m.err.Error()); errMsg != "" {
			status = fmt.Sprintf("%s: %s", status, errMsg)
		}
	} else if m.metrics != nil {
		repoCount := len(m.metrics.ByRepository)
		status = fmt.Sprintf("Metrics loaded • %d repositories", repoCount)

		if m.rateLimit != nil {
			status = fmt.Sprintf("%s • API: %d/%d remaining",
				status,
				m.rateLimit.Remaining,
				m.rateLimit.Limit,
			)
		}
	} else {
		status = "Press 'r' to load metrics"
	}

	m.statusBar.SetMessage(status)

	m.statusBar.ClearItems()
	m.statusBar.AddItem("j/k", "scroll")
	m.statusBar.AddItem("r", "refresh")
	m.statusBar.AddItem("l", "rate limit")
	m.statusBar.AddItem("q", "back")

	if !m.loading && m.err == nil && !m.lastUpdated.IsZero() {
		m.statusBar.AddItem("Updated", m.lastUpdated.Format("15:04:05"))
	}

	if m.metrics != nil {
		m.statusBar.AddItem("PRs", fmt.Sprintf("%d", m.metrics.Overall.Count))
	}
}

func (m *MetricsView) maxScroll() int {
	lines := m.renderContentLines()
	available := m.height - 1
	if available < 1 {
		return 0
	}
	if len(lines) <= available {
		return 0
	}
	return len(lines) - available
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}

	d = d.Round(time.Minute)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", int(d.Seconds())))
	}

	return strings.Join(parts, " ")
}
