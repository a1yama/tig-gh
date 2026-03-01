package views

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/shared/metricsformat"
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
	useCase           LeadTimeMetricsUseCase
	metrics           *models.LeadTimeMetrics
	loading           bool
	err               error
	width             int
	height            int
	scroll            int
	statusBar         *components.StatusBar
	lastUpdated       time.Time
	rateLimit         *github.Rate // GitHub API rate limit info
	progress          *models.MetricsProgress
	progressCh        chan models.MetricsProgress
	filterMode        bool   // フィルタモード中かどうか
	filteredRepo      string // フィルタ中のリポジトリ（空なら全体表示）
	selectedRepoIndex int    // フィルタモード中の選択インデックス
	prBrowseMode      bool   // PRブラウズモード中かどうか
	selectedPRIndex   int    // PRブラウズモード中の選択インデックス
	config            *models.MetricsConfig
	copySuccess       bool   // コピー成功フラグ
	copyMessage       string // コピー成功メッセージ
}

func defaultMetricsConfig() *models.MetricsConfig {
	cfg := models.DefaultConfig()
	return &cfg.Metrics
}

func cloneMetricsConfig(cfg *models.MetricsConfig) *models.MetricsConfig {
	if cfg == nil {
		return defaultMetricsConfig()
	}
	clone := *cfg
	return &clone
}

// NewMetricsView は空のメトリクスビューを返す
func NewMetricsView() *MetricsView {
	return &MetricsView{
		statusBar: components.NewStatusBar(),
		loading:   false,
		scroll:    0,
		config:    defaultMetricsConfig(),
	}
}

// NewMetricsViewWithUseCase はユースケースをバインドしたビューを返す
func NewMetricsViewWithUseCase(useCase LeadTimeMetricsUseCase, config ...*models.MetricsConfig) *MetricsView {
	view := NewMetricsView()
	view.useCase = useCase
	if len(config) > 0 {
		view.config = cloneMetricsConfig(config[0])
	}
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
	// フィルタモード中の処理
	if m.filterMode {
		return m.handleFilterModeKey(msg)
	}

	// PRブラウズモード中の処理
	if m.prBrowseMode {
		return m.handlePRBrowseModeKey(msg)
	}

	// 通常モードの処理
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m, func() tea.Msg { return MetricsExitMsg{} }
	case "f":
		// フィルタモードに入る
		m.enterFilterMode()
		return m, nil
	case "a":
		// 全体表示に戻る
		m.filteredRepo = ""
		m.scroll = 0
		return m, nil
	case "o":
		// PRブラウズモードに入る
		if m.metrics != nil && m.config.ShowPRLeadTimes {
			entries := m.getFilteredPRLeadTimes()
			if len(entries) > 0 {
				m.prBrowseMode = true
				m.selectedPRIndex = 0
				m.updateStatusBar()
			}
		}
		return m, nil
	case "y":
		// メトリクスをMarkdown形式でクリップボードにコピー
		if m.metrics != nil && !m.loading {
			markdown := m.toMarkdown()
			if err := copyToClipboard(markdown); err != nil {
				m.copySuccess = false
				m.copyMessage = fmt.Sprintf("Failed to copy: %v", err)
			} else {
				m.copySuccess = true
				m.copyMessage = "Metrics copied to clipboard as Markdown"
			}
			m.updateStatusBar()
		}
		return m, nil
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
		// コピーメッセージをクリア
		m.copySuccess = false
		m.copyMessage = ""
		maxScroll := m.maxScroll()
		if m.scroll < maxScroll {
			m.scroll++
		}
		return m, nil
	case "k", "up":
		// コピーメッセージをクリア
		m.copySuccess = false
		m.copyMessage = ""
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

func (m *MetricsView) handleFilterModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	repoList := m.getRepositoryList()
	if len(repoList) == 0 {
		m.filterMode = false
		return m, nil
	}

	switch msg.String() {
	case "esc", "ctrl+c":
		// フィルタモードをキャンセル
		m.filterMode = false
		return m, nil
	case "j", "down":
		// 次のリポジトリを選択
		if m.selectedRepoIndex < len(repoList)-1 {
			m.selectedRepoIndex++
		}
		return m, nil
	case "k", "up":
		// 前のリポジトリを選択
		if m.selectedRepoIndex > 0 {
			m.selectedRepoIndex--
		}
		return m, nil
	case "enter":
		// フィルタを適用
		if m.selectedRepoIndex >= 0 && m.selectedRepoIndex < len(repoList) {
			m.filteredRepo = repoList[m.selectedRepoIndex]
			m.scroll = 0
		}
		m.filterMode = false
		return m, nil
	case "a":
		// 全体表示に戻る
		m.filteredRepo = ""
		m.scroll = 0
		m.filterMode = false
		return m, nil
	}

	return m, nil
}

func (m *MetricsView) handlePRBrowseModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	entries := m.getFilteredPRLeadTimes()
	if len(entries) == 0 {
		m.prBrowseMode = false
		m.updateStatusBar()
		return m, nil
	}

	switch msg.String() {
	case "esc", "ctrl+c":
		m.prBrowseMode = false
		m.updateStatusBar()
		return m, nil
	case "j", "down":
		if m.selectedPRIndex < len(entries)-1 {
			m.selectedPRIndex++
		}
		return m, nil
	case "k", "up":
		if m.selectedPRIndex > 0 {
			m.selectedPRIndex--
		}
		return m, nil
	case "enter":
		if m.selectedPRIndex >= 0 && m.selectedPRIndex < len(entries) {
			url := entries[m.selectedPRIndex].HTMLURL
			if url != "" {
				openBrowser(url)
			}
		}
		return m, nil
	}

	return m, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	_ = cmd.Start()
}

func (m *MetricsView) getFilteredPRLeadTimes() []models.PRLeadTimeEntry {
	if m.metrics == nil {
		return nil
	}

	entries := m.metrics.PRLeadTimes
	if m.filteredRepo != "" {
		filtered := make([]models.PRLeadTimeEntry, 0)
		for _, e := range entries {
			if e.Repository == m.filteredRepo {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	maxDisplay := 20
	if len(entries) > maxDisplay {
		entries = entries[:maxDisplay]
	}

	return entries
}

func (m *MetricsView) enterFilterMode() {
	m.filterMode = true
	m.selectedRepoIndex = 0
}

func (m *MetricsView) getRepositoryList() []string {
	if m.metrics == nil {
		return nil
	}

	repos := make([]string, 0, len(m.metrics.ByRepository))
	for repo := range m.metrics.ByRepository {
		repos = append(repos, repo)
	}
	sort.Strings(repos)
	return repos
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

	// 計測期間を別行で表示
	if m.config != nil && m.config.CalculationPeriod > 0 {
		days := int(m.config.CalculationPeriod.Hours() / 24)
		endDate := time.Now()
		startDate := endDate.Add(-m.config.CalculationPeriod)
		periodLine := fmt.Sprintf("Period: %s ~ %s (%d days)",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			days)
		lines = append(lines, styles.MutedStyle.Render(periodLine))
	}

	// フィルタ状態を表示
	if m.filteredRepo != "" {
		lines = append(lines, styles.WarningStyle.Render(fmt.Sprintf("Filtered: %s", m.filteredRepo)))
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

	// フィルタモード中はリポジトリ選択UIを表示
	if m.filterMode {
		return m.renderFilterModeUI()
	}

	lines = append(lines, m.renderOverallSection()...)
	lines = append(lines, "")

	if m.config.ShowPRLeadTimes {
		lines = append(lines, m.renderPRLeadTimeSection()...)
		lines = append(lines, "")
	}

	if m.config.ShowReviewPhases {
		lines = append(lines, m.renderReviewPhaseSection()...)
		lines = append(lines, "")
	}
	if m.config.ShowDayOfWeek {
		lines = append(lines, m.renderDayOfWeekSection()...)
		lines = append(lines, "")
	}
	if m.config.ShowWeeklyComparison {
		lines = append(lines, m.renderWeeklyComparisonSection()...)
		lines = append(lines, "")
	}
	if m.config.ShowQualityIssues {
		lines = append(lines, m.renderPRQualitySection()...)
		lines = append(lines, "")
	}
	if m.config.ShowStagnantPRs {
		lines = append(lines, m.renderStagnantPRSection()...)
		lines = append(lines, "")
	}
	if m.config.ShowRepositoryStats {
		lines = append(lines, m.renderRepositorySection()...)
		lines = append(lines, "")
	}

	// ヘルプテキストを更新
	helpText := "Controls: j/k scroll • r refresh • f filter • o browse PRs • a show all • y copy as markdown • q back"
	lines = append(lines, styles.HelpStyle.Render(helpText))

	return lines
}

func (m *MetricsView) renderFilterModeUI() []string {
	lines := []string{
		styles.TitleStyle.Render("Lead Time Metrics"),
	}

	// 計測期間を別行で表示
	if m.config != nil && m.config.CalculationPeriod > 0 {
		days := int(m.config.CalculationPeriod.Hours() / 24)
		endDate := time.Now()
		startDate := endDate.Add(-m.config.CalculationPeriod)
		periodLine := fmt.Sprintf("Period: %s ~ %s (%d days)",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			days)
		lines = append(lines, styles.MutedStyle.Render(periodLine))
	}

	lines = append(lines,
		styles.MutedStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdated.Format("2006-01-02 15:04:05"))),
		"",
		styles.HeaderStyle.Render("Select Repository to Filter"),
		"",
	)

	repoList := m.getRepositoryList()
	if len(repoList) == 0 {
		lines = append(lines, styles.MutedStyle.Render("No repositories available."))
		return lines
	}

	for idx, repo := range repoList {
		prefix := "  "
		repoStyle := lipgloss.NewStyle()
		if idx == m.selectedRepoIndex {
			prefix = "> "
			repoStyle = repoStyle.Foreground(lipgloss.Color("2")).Bold(true)
		}
		lines = append(lines, prefix+repoStyle.Render(repo))
	}

	lines = append(lines, "")
	helpText := "Controls: j/k navigate • Enter apply filter • a show all • Esc cancel"
	lines = append(lines, styles.HelpStyle.Render(helpText))

	return lines
}

func (m *MetricsView) renderPRLeadTimeSection() []string {
	entries := m.getFilteredPRLeadTimes()

	header := "PR Lead Times (Open → Merge)"
	if !m.prBrowseMode {
		header += "                              Press 'o' to browse"
	}

	lines := []string{
		styles.HeaderStyle.Render(header),
	}

	if len(entries) == 0 {
		lines = append(lines, styles.MutedStyle.Render("No PR lead time data available."))
		return lines
	}

	for idx, entry := range entries {
		prefix := "   "
		if m.prBrowseMode && idx == m.selectedPRIndex {
			prefix = ">  "
		}

		leadTimeStr := fmt.Sprintf("%7s", metricsformat.FormatDuration(entry.LeadTime))
		repoAndNumber := fmt.Sprintf("%s #%d", entry.Repository, entry.Number)
		title := trimColumnText(entry.Title, 30)
		mergedDate := entry.MergedAt.Format("2006-01-02")

		line := fmt.Sprintf("%s%2d. %s  %-30s  %-30s  (%s)",
			prefix,
			idx+1,
			leadTimeStr,
			repoAndNumber,
			title,
			mergedDate,
		)

		if m.prBrowseMode && idx == m.selectedPRIndex {
			selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
			line = selectedStyle.Render(line)
		}

		lines = append(lines, line)
	}

	if m.prBrowseMode {
		lines = append(lines, "")
		lines = append(lines, styles.HelpStyle.Render("Controls: j/k navigate • Enter open in browser • Esc cancel"))
	}

	return lines
}

func (m *MetricsView) renderOverallSection() []string {
	header := "Overall Lead Time"
	stat := m.metrics.Overall

	if m.filteredRepo != "" {
		header = fmt.Sprintf("Lead Time - %s", m.filteredRepo)
		if repoStat, ok := m.metrics.ByRepository[m.filteredRepo]; ok {
			stat = repoStat
		} else {
			return []string{
				styles.HeaderStyle.Render(header),
				styles.MutedStyle.Render(fmt.Sprintf("No lead time data for %s.", m.filteredRepo)),
			}
		}
	}

	lines := []string{
		styles.HeaderStyle.Render(header),
	}

	if stat.Count == 0 {
		lines = append(lines, styles.MutedStyle.Render("No merged PRs in the selected period."))
		return lines
	}

	lines = append(lines, fmt.Sprintf("Average: %s  Median: %s  PRs: %d",
		metricsformat.FormatDuration(stat.Average),
		metricsformat.FormatDuration(stat.Median),
		stat.Count,
	))

	return lines
}

func (m *MetricsView) renderStagnantPRSection() []string {
	stagnant := m.metrics.StagnantPRs
	lines := []string{
		styles.HeaderStyle.Render(fmt.Sprintf("Stagnant PRs (Open > %s)", metricsformat.FormatDuration(stagnant.Threshold))),
	}

	// フィルタリングされた滞留PRリストを作成
	filteredPRs := stagnant.LongestWaiting
	if m.filteredRepo != "" {
		filteredPRs = []models.StagnantPRInfo{}
		for _, pr := range stagnant.LongestWaiting {
			if pr.Repository == m.filteredRepo {
				filteredPRs = append(filteredPRs, pr)
			}
		}
	}

	if len(filteredPRs) == 0 {
		if m.filteredRepo != "" {
			lines = append(lines, styles.MutedStyle.Render(fmt.Sprintf("No stagnant PRs found for %s.", m.filteredRepo)))
		} else {
			lines = append(lines, styles.MutedStyle.Render("No stagnant PRs found."))
		}
		return lines
	}

	// フィルタされている場合は全体統計は表示しない
	if m.filteredRepo == "" {
		lines = append(lines,
			fmt.Sprintf("Total stagnant PRs:  %d", stagnant.TotalStagnant),
		)
	}

	if len(filteredPRs) > 0 {
		if m.filteredRepo != "" {
			lines = append(lines, fmt.Sprintf("Stagnant PRs for %s:", m.filteredRepo))
		} else {
			lines = append(lines, "Longest waiting PRs:")
		}
		for idx, pr := range filteredPRs {
			lines = append(lines,
				fmt.Sprintf("  %2d. %s #%d (%s): %s",
					idx+1,
					pr.Repository,
					pr.Number,
					metricsformat.FormatDuration(pr.Age),
					pr.Title,
				),
			)
		}
	}

	return lines
}

func (m *MetricsView) renderReviewPhaseSection() []string {
	header := "Review Phase Breakdown"
	phaseMetrics := m.metrics.PhaseBreakdown

	if m.filteredRepo != "" {
		header = fmt.Sprintf("%s (Filtered: %s)", header, m.filteredRepo)
		if m.metrics.ByRepositoryPhaseBreakdown != nil {
			if repoPhase, ok := m.metrics.ByRepositoryPhaseBreakdown[m.filteredRepo]; ok {
				phaseMetrics = repoPhase
			} else {
				return []string{
					styles.HeaderStyle.Render(header),
					styles.MutedStyle.Render(fmt.Sprintf("No review phase data available for %s.", m.filteredRepo)),
				}
			}
		} else {
			return []string{
				styles.HeaderStyle.Render(header),
				styles.MutedStyle.Render(fmt.Sprintf("No review phase data available for %s.", m.filteredRepo)),
			}
		}
	}

	lines := []string{
		styles.HeaderStyle.Render(header),
	}

	if phaseMetrics.SampleCount == 0 {
		if m.filteredRepo != "" {
			lines = append(lines, styles.MutedStyle.Render(fmt.Sprintf("Not enough review phase data for %s.", m.filteredRepo)))
		} else {
			lines = append(lines, styles.MutedStyle.Render("Not enough review phase data."))
		}
		return lines
	}

	type phaseInfo struct {
		label    string
		duration time.Duration
	}

	phases := []phaseInfo{
		{label: "PR Created → First Review:", duration: phaseMetrics.CreatedToFirstReview},
		{label: "First Review → Approval:", duration: phaseMetrics.FirstReviewToApproval},
		{label: "Approval → Merge:", duration: phaseMetrics.ApprovalToMerge},
	}

	longest := time.Duration(0)
	for _, phase := range phases {
		if phase.duration > longest {
			longest = phase.duration
		}
	}

	for _, phase := range phases {
		line := fmt.Sprintf("  %-30s avg %s (%d PRs)", phase.label, metricsformat.FormatDuration(phase.duration), phaseMetrics.SampleCount)
		if longest > 0 && phase.duration == longest {
			line += " ← ボトルネック"
		}
		lines = append(lines, line)
	}

	lines = append(lines, "  "+strings.Repeat("─", 45))
	lines = append(lines, fmt.Sprintf("  %-30s avg %s", "Total Lead Time:", metricsformat.FormatDuration(phaseMetrics.TotalLeadTime)))

	return lines
}

func (m *MetricsView) renderDayOfWeekSection() []string {
	header := "Activity by Day of Week"
	statsByDay := m.metrics.ByDayOfWeek

	if m.filteredRepo != "" {
		header = fmt.Sprintf("Activity by Day of Week (Filtered: %s)", m.filteredRepo)
		if m.metrics.ByRepositoryDayOfWeek != nil {
			statsByDay = m.metrics.ByRepositoryDayOfWeek[m.filteredRepo]
		} else {
			statsByDay = nil
		}
	}

	lines := []string{
		styles.HeaderStyle.Render(header),
	}

	if statsByDay == nil || len(statsByDay) == 0 {
		if m.filteredRepo != "" {
			lines = append(lines, styles.MutedStyle.Render(fmt.Sprintf("No day-of-week data available for %s.", m.filteredRepo)))
		} else {
			lines = append(lines, styles.MutedStyle.Render("No day-of-week data available."))
		}
		return lines
	}

	// 表示する曜日を決定（土日を除外する場合は平日のみ）
	displayDays := m.getDisplayDays()

	headers := make([]string, len(displayDays))
	for i, day := range displayDays {
		headers[i] = fmt.Sprintf("%4s", metricsformat.ShortWeekday(day))
	}
	lines = append(lines, fmt.Sprintf("%-8s%s", "", strings.Join(headers, " ")))

	mergeRow := "Merges  "
	reviewRow := "Reviews "

	for _, day := range displayDays {
		stats := statsByDay[day]
		mergeRow += fmt.Sprintf("%4d", stats.MergeCount)
		reviewRow += fmt.Sprintf("%4d", stats.ReviewCount)
	}

	lines = append(lines, mergeRow)
	lines = append(lines, reviewRow)

	return lines
}

func (m *MetricsView) renderWeeklyComparisonSection() []string {
	header := "Weekly Review Activity (This Week vs Last Week)"
	comparison := m.metrics.WeeklyComparison

	if m.filteredRepo != "" {
		header = fmt.Sprintf("%s - %s", header, m.filteredRepo)
		if repoComparison, ok := m.metrics.ByRepositoryWeekly[m.filteredRepo]; ok {
			comparison = repoComparison
		} else {
			return []string{
				styles.HeaderStyle.Render(header),
				styles.MutedStyle.Render(fmt.Sprintf("No weekly data available for %s.", m.filteredRepo)),
			}
		}
	}

	lines := []string{
		styles.HeaderStyle.Render(header),
	}

	lines = append(lines, fmt.Sprintf("%-25s %10s %10s", "Period", "Reviews", "Merges"))
	lines = append(lines,
		fmt.Sprintf("%-25s %10d %10d", "This Week (last 7 days)", comparison.ThisWeek.ReviewCount, comparison.ThisWeek.MergeCount),
	)
	lines = append(lines,
		fmt.Sprintf("%-25s %10d %10d", "Last Week (8-14 days ago)", comparison.LastWeek.ReviewCount, comparison.LastWeek.MergeCount),
	)

	// Format the change percentages
	reviewChangeStr := fmt.Sprintf("%+.1f%%", comparison.ReviewChangePercent)
	mergeChangeStr := fmt.Sprintf("%+.1f%%", comparison.MergeChangePercent)

	// Build the change line with proper alignment
	col1 := "Change"
	col2 := fmt.Sprintf("%10s", reviewChangeStr)
	col3 := fmt.Sprintf("%10s", mergeChangeStr)

	// Apply colors using ANSI codes for precise control
	col2Colored := applyChangeColorANSI(col2, comparison.ReviewChangePercent)
	col3Colored := applyChangeColorANSI(col3, comparison.MergeChangePercent)

	changeLine := fmt.Sprintf("%-25s %s %s", col1, col2Colored, col3Colored)
	lines = append(lines, changeLine)

	return lines
}


type qualityIssueDisplay struct {
	issue models.PRQualityIssue
}

func (m *MetricsView) renderPRQualitySection() []string {
	issues := m.metrics.QualityIssues.Issues
	if len(issues) == 0 {
		return []string{
			styles.HeaderStyle.Render("PR Quality Issues (0 issues)"),
			styles.MutedStyle.Render("No PR quality issues detected."),
		}
	}

	filtered := issues
	if m.filteredRepo != "" {
		filtered = make([]models.PRQualityIssue, 0, len(issues))
		for _, issue := range issues {
			if issue.Repository == m.filteredRepo {
				filtered = append(filtered, issue)
			}
		}
		if len(filtered) == 0 {
			return []string{
				styles.HeaderStyle.Render("PR Quality Issues (0 issues)"),
				styles.MutedStyle.Render(fmt.Sprintf("No PR quality issues found for %s.", m.filteredRepo)),
			}
		}
	}

	var high, medium []qualityIssueDisplay
	for _, issue := range filtered {
		entry := qualityIssueDisplay{issue: issue}
		if strings.EqualFold(issue.Severity, "high") {
			high = append(high, entry)
		} else {
			medium = append(medium, entry)
		}
	}

	displayCount := len(filtered)
	if displayCount > metricsformat.MaxQualityIssuesToDisplay {
		displayCount = metricsformat.MaxQualityIssuesToDisplay
	}

	lines := []string{
		styles.HeaderStyle.Render(fmt.Sprintf("PR Quality Issues (%d issues)", displayCount)),
	}

	if len(high) > displayCount {
		high = high[:displayCount]
		medium = nil
	} else {
		remaining := displayCount - len(high)
		if remaining < len(medium) {
			medium = medium[:remaining]
		}
	}

	if len(high) > 0 {
		lines = append(lines, "High Priority:")
		lines = append(lines, m.renderQualityIssueList(high)...)
	}

	if len(medium) > 0 {
		if len(high) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "Medium Priority:")
		lines = append(lines, m.renderQualityIssueList(medium)...)
	}

	return lines
}

func (m *MetricsView) renderQualityIssueList(items []qualityIssueDisplay) []string {
	if len(items) == 0 {
		return nil
	}

	const (
		repoWidth   = 31
		numberWidth = 7
		typeWidth   = 16
		detailWidth = 28
	)

	header := fmt.Sprintf(
		"%-*s %-*s %-*s %-*s %s",
		repoWidth, "Repo",
		numberWidth, "#",
		typeWidth, "Type",
		detailWidth, "Details",
		"Title",
	)

	lines := []string{styles.MutedStyle.Render(header)}

	for _, entry := range items {
		repo := trimColumnText(entry.issue.Repository, repoWidth)
		number := fmt.Sprintf("#%d", entry.issue.Number)
		issueType := trimColumnText(entry.issue.IssueType, typeWidth)
		if issueType == "" {
			issueType = "-"
		}
		details := trimColumnText(entry.issue.Details, detailWidth)
		if details == "" {
			details = "-"
		}
		title := entry.issue.Title
		if title == "" {
			title = "-"
		}

		row := fmt.Sprintf(
			"%-*s %-*s %-*s %-*s %s",
			repoWidth, repo,
			numberWidth, number,
			typeWidth, issueType,
			detailWidth, details,
			title,
		)
		lines = append(lines, row)
	}

	return lines
}

func singleLineText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func trimColumnText(value string, width int) string {
	value = singleLineText(value)
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func trimColumnTextFromEnd(value string, width int) string {
	value = singleLineText(value)
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[len(runes)-width:])
	}
	return "..." + string(runes[len(runes)-(width-3):])
}

func normalizeRecommendation(value string) string {
	text := singleLineText(value)
	if text == "" {
		return "-"
	}

	prefixes := []string{"推奨:", "Recommendation:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(text, prefix) {
			text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
		}
	}

	if text == "" {
		return "-"
	}
	return text
}

func (m *MetricsView) renderRepositorySection() []string {
	lines := []string{
		styles.HeaderStyle.Render("Per Repository"),
	}

	if len(m.metrics.ByRepository) == 0 {
		lines = append(lines, styles.MutedStyle.Render("No repository data available."))
		return lines
	}

	// フィルタリングされたリポジトリリストを作成
	repoNames := make([]string, 0, len(m.metrics.ByRepository))
	if m.filteredRepo != "" {
		// フィルタが有効な場合、そのリポジトリのみ表示
		if _, exists := m.metrics.ByRepository[m.filteredRepo]; exists {
			repoNames = append(repoNames, m.filteredRepo)
		}
	} else {
		// フィルタがない場合、全リポジトリ表示
		for name := range m.metrics.ByRepository {
			repoNames = append(repoNames, name)
		}
		sort.Strings(repoNames)
	}

	if len(repoNames) == 0 {
		lines = append(lines, styles.MutedStyle.Render(fmt.Sprintf("No data available for %s.", m.filteredRepo)))
		return lines
	}

	header := fmt.Sprintf("%-40s %12s %12s %6s", "Repository", "Avg", "Median", "PRs")
	lines = append(lines, styles.MutedStyle.Render(header))

	for _, name := range repoNames {
		stat := m.metrics.ByRepository[name]
		line := fmt.Sprintf(
			"%-40s %12s %12s %6d",
			name,
			metricsformat.FormatDuration(stat.Average),
			metricsformat.FormatDuration(stat.Median),
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
	case m.prBrowseMode:
		mode = "PR Browse"
	case m.filterMode:
		mode = "Filter"
	case m.loading:
		mode = "Loading"
	case m.err != nil:
		mode = "Error"
	case m.filteredRepo != "":
		mode = "Filtered"
	}
	m.statusBar.SetMode(mode)

	var status string
	// コピー成功メッセージを優先表示
	if m.copyMessage != "" {
		status = m.copyMessage
	} else if m.prBrowseMode {
		entries := m.getFilteredPRLeadTimes()
		status = fmt.Sprintf("Browsing PRs (%d/%d)", m.selectedPRIndex+1, len(entries))
	} else if m.filterMode {
		status = "Select repository to filter"
	} else if m.loading {
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
		if m.filteredRepo != "" {
			status = fmt.Sprintf("Filtered: %s", m.filteredRepo)
		} else {
			repoCount := len(m.metrics.ByRepository)
			status = fmt.Sprintf("Metrics loaded • %d repositories", repoCount)
		}

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
	if m.prBrowseMode {
		m.statusBar.AddItem("j/k", "navigate")
		m.statusBar.AddItem("Enter", "open in browser")
		m.statusBar.AddItem("Esc", "cancel")
	} else if m.filterMode {
		m.statusBar.AddItem("j/k", "navigate")
		m.statusBar.AddItem("Enter", "apply")
		m.statusBar.AddItem("a", "show all")
		m.statusBar.AddItem("Esc", "cancel")
	} else {
		m.statusBar.AddItem("j/k", "scroll")
		m.statusBar.AddItem("r", "refresh")
		m.statusBar.AddItem("f", "filter")
		if m.filteredRepo != "" {
			m.statusBar.AddItem("a", "show all")
		}
		if m.metrics != nil && !m.loading {
			m.statusBar.AddItem("y", "copy")
		}
		m.statusBar.AddItem("l", "rate limit")
		m.statusBar.AddItem("q", "back")
	}

	if !m.loading && m.err == nil && !m.lastUpdated.IsZero() && !m.filterMode {
		m.statusBar.AddItem("Updated", m.lastUpdated.Format("15:04:05"))
	}

	if m.metrics != nil && !m.filterMode {
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

// getDisplayDays は表示する曜日のリストを返す（設定により土日を除外）
func (m *MetricsView) getDisplayDays() []time.Weekday {
	if m.config != nil && !m.config.ShowWeekendInDayOfWeek {
		return metricsformat.WeekdaysOnly
	}
	return metricsformat.WeekdayDisplayOrder
}

func formatChangePercent(value float64) string {
	formatted := fmt.Sprintf("%+.1f%%", value)
	switch {
	case value > 0:
		return styles.SuccessStyle.Render(formatted)
	case value < 0:
		return styles.ErrorStyle.Render(formatted)
	default:
		return styles.MutedStyle.Render(formatted)
	}
}

func applyChangeColor(paddedStr string, value float64) string {
	switch {
	case value > 0:
		return styles.SuccessStyle.Render(paddedStr)
	case value < 0:
		return styles.ErrorStyle.Render(paddedStr)
	default:
		return styles.MutedStyle.Render(paddedStr)
	}
}

func applyChangeColorANSI(text string, value float64) string {
	const (
		reset = "\033[0m"
		green = "\033[32;1m" // bold green
		red   = "\033[31;1m" // bold red
		gray  = "\033[90m"   // dark gray
	)

	switch {
	case value > 0:
		return green + text + reset
	case value < 0:
		return red + text + reset
	default:
		return gray + text + reset
	}
}

// toMarkdown はメトリクス情報をMarkdown形式に変換する
func (m *MetricsView) toMarkdown() string {
	if m.metrics == nil {
		return "# Lead Time Metrics\n\nNo data available.\n"
	}

	var sb strings.Builder

	// タイトル
	sb.WriteString("# Lead Time Metrics\n\n")

	// 計測期間
	if m.config != nil && m.config.CalculationPeriod > 0 {
		days := int(m.config.CalculationPeriod.Hours() / 24)
		endDate := time.Now()
		startDate := endDate.Add(-m.config.CalculationPeriod)
		sb.WriteString(fmt.Sprintf("**Period**: %s ~ %s (%d days)\n\n",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			days))
	}

	// 最終更新日時
	if !m.lastUpdated.IsZero() {
		sb.WriteString(fmt.Sprintf("**Last updated**: %s\n\n", m.lastUpdated.Format("2006-01-02 15:04:05")))
	}

	// フィルタ状態
	if m.filteredRepo != "" {
		sb.WriteString(fmt.Sprintf("**Filtered**: %s\n\n", m.filteredRepo))
	}

	// Overall Lead Time
	sb.WriteString(m.overallToMarkdown())
	sb.WriteString("\n")

	// PR Lead Times
	if m.config.ShowPRLeadTimes {
		sb.WriteString(m.prLeadTimesToMarkdown())
		sb.WriteString("\n")
	}

	// Review Phase Breakdown
	if m.config.ShowReviewPhases {
		sb.WriteString(m.reviewPhasesToMarkdown())
		sb.WriteString("\n")
	}

	// Day of Week
	if m.config.ShowDayOfWeek {
		sb.WriteString(m.dayOfWeekToMarkdown())
		sb.WriteString("\n")
	}

	// Weekly Comparison
	if m.config.ShowWeeklyComparison {
		sb.WriteString(m.weeklyComparisonToMarkdown())
		sb.WriteString("\n")
	}

	// Quality Issues
	if m.config.ShowQualityIssues {
		sb.WriteString(m.qualityIssuesToMarkdown())
		sb.WriteString("\n")
	}

	// Stagnant PRs
	if m.config.ShowStagnantPRs {
		sb.WriteString(m.stagnantPRsToMarkdown())
		sb.WriteString("\n")
	}

	// Repository Stats
	if m.config.ShowRepositoryStats {
		sb.WriteString(m.repositoryStatsToMarkdown())
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *MetricsView) overallToMarkdown() string {
	var sb strings.Builder

	header := "## Overall Lead Time"
	stat := m.metrics.Overall

	if m.filteredRepo != "" {
		header = fmt.Sprintf("## Lead Time - %s", m.filteredRepo)
		if repoStat, ok := m.metrics.ByRepository[m.filteredRepo]; ok {
			stat = repoStat
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\nNo lead time data for %s.\n", header, m.filteredRepo))
			return sb.String()
		}
	}

	sb.WriteString(header + "\n\n")

	if stat.Count == 0 {
		sb.WriteString("No merged PRs in the selected period.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("- **Average**: %s\n", metricsformat.FormatDuration(stat.Average)))
	sb.WriteString(fmt.Sprintf("- **Median**: %s\n", metricsformat.FormatDuration(stat.Median)))
	sb.WriteString(fmt.Sprintf("- **PRs**: %d\n", stat.Count))

	return sb.String()
}

func (m *MetricsView) prLeadTimesToMarkdown() string {
	entries := m.getFilteredPRLeadTimes()

	var sb strings.Builder
	sb.WriteString("## PR Lead Times (Open → Merge)\n\n")

	if len(entries) == 0 {
		sb.WriteString("No PR lead time data available.\n")
		return sb.String()
	}

	sb.WriteString("| # | Lead Time | Repository | PR | Title | Merged Date |\n")
	sb.WriteString("|---|-----------|------------|----|---------|--------------|\n")

	for idx, entry := range entries {
		leadTimeStr := metricsformat.FormatDuration(entry.LeadTime)
		repoAndNumber := fmt.Sprintf("%s [#%d](%s)", entry.Repository, entry.Number, entry.HTMLURL)
		title := entry.Title
		mergedDate := entry.MergedAt.Format("2006-01-02")

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			idx+1,
			leadTimeStr,
			entry.Repository,
			repoAndNumber,
			title,
			mergedDate,
		))
	}

	return sb.String()
}

func (m *MetricsView) reviewPhasesToMarkdown() string {
	var sb strings.Builder

	header := "## Review Phase Breakdown"
	phaseMetrics := m.metrics.PhaseBreakdown

	if m.filteredRepo != "" {
		header = fmt.Sprintf("## Review Phase Breakdown (Filtered: %s)", m.filteredRepo)
		if m.metrics.ByRepositoryPhaseBreakdown != nil {
			if repoPhase, ok := m.metrics.ByRepositoryPhaseBreakdown[m.filteredRepo]; ok {
				phaseMetrics = repoPhase
			} else {
				sb.WriteString(fmt.Sprintf("%s\n\nNo review phase data available for %s.\n", header, m.filteredRepo))
				return sb.String()
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\nNo review phase data available for %s.\n", header, m.filteredRepo))
			return sb.String()
		}
	}

	sb.WriteString(header + "\n\n")

	if phaseMetrics.SampleCount == 0 {
		sb.WriteString("Not enough review phase data.\n")
		return sb.String()
	}

	type phaseInfo struct {
		label    string
		duration time.Duration
	}

	phases := []phaseInfo{
		{label: "PR Created → First Review", duration: phaseMetrics.CreatedToFirstReview},
		{label: "First Review → Approval", duration: phaseMetrics.FirstReviewToApproval},
		{label: "Approval → Merge", duration: phaseMetrics.ApprovalToMerge},
	}

	longest := time.Duration(0)
	for _, phase := range phases {
		if phase.duration > longest {
			longest = phase.duration
		}
	}

	for _, phase := range phases {
		bottleneck := ""
		if longest > 0 && phase.duration == longest {
			bottleneck = " ← **ボトルネック**"
		}
		sb.WriteString(fmt.Sprintf("- **%s**: %s (avg, %d PRs)%s\n",
			phase.label,
			metricsformat.FormatDuration(phase.duration),
			phaseMetrics.SampleCount,
			bottleneck,
		))
	}

	sb.WriteString(fmt.Sprintf("- **Total Lead Time**: %s (avg)\n", metricsformat.FormatDuration(phaseMetrics.TotalLeadTime)))

	return sb.String()
}

func (m *MetricsView) dayOfWeekToMarkdown() string {
	var sb strings.Builder

	header := "## Activity by Day of Week"
	statsByDay := m.metrics.ByDayOfWeek

	if m.filteredRepo != "" {
		header = fmt.Sprintf("## Activity by Day of Week (Filtered: %s)", m.filteredRepo)
		if m.metrics.ByRepositoryDayOfWeek != nil {
			statsByDay = m.metrics.ByRepositoryDayOfWeek[m.filteredRepo]
		} else {
			statsByDay = nil
		}
	}

	sb.WriteString(header + "\n\n")

	if statsByDay == nil || len(statsByDay) == 0 {
		sb.WriteString("No day-of-week data available.\n")
		return sb.String()
	}

	// 表示する曜日を決定
	displayDays := m.getDisplayDays()

	// ヘッダー行を作成
	headerRow := "|  |"
	separatorRow := "|---|"
	for _, day := range displayDays {
		headerRow += fmt.Sprintf(" %s |", metricsformat.ShortWeekday(day))
		separatorRow += "-----|"
	}
	sb.WriteString(headerRow + "\n")
	sb.WriteString(separatorRow + "\n")

	mergeCounts := make([]string, 0, len(displayDays))
	reviewCounts := make([]string, 0, len(displayDays))

	for _, day := range displayDays {
		stats := statsByDay[day]
		mergeCounts = append(mergeCounts, fmt.Sprintf("%d", stats.MergeCount))
		reviewCounts = append(reviewCounts, fmt.Sprintf("%d", stats.ReviewCount))
	}

	sb.WriteString(fmt.Sprintf("| Merges | %s |\n", strings.Join(mergeCounts, " | ")))
	sb.WriteString(fmt.Sprintf("| Reviews | %s |\n", strings.Join(reviewCounts, " | ")))

	return sb.String()
}

func (m *MetricsView) weeklyComparisonToMarkdown() string {
	var sb strings.Builder

	header := "## Weekly Review Activity"
	comparison := m.metrics.WeeklyComparison

	if m.filteredRepo != "" {
		header = fmt.Sprintf("## Weekly Review Activity (Filtered: %s)", m.filteredRepo)
		if repoComparison, ok := m.metrics.ByRepositoryWeekly[m.filteredRepo]; ok {
			comparison = repoComparison
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\nNo weekly data available for %s.\n", header, m.filteredRepo))
			return sb.String()
		}
	}

	sb.WriteString(header + "\n\n")
	sb.WriteString("| Period | Reviews | Merges |\n")
	sb.WriteString("|--------|---------|--------|\n")
	sb.WriteString(fmt.Sprintf("| This Week (last 7 days) | %d | %d |\n",
		comparison.ThisWeek.ReviewCount,
		comparison.ThisWeek.MergeCount,
	))
	sb.WriteString(fmt.Sprintf("| Last Week (8-14 days ago) | %d | %d |\n",
		comparison.LastWeek.ReviewCount,
		comparison.LastWeek.MergeCount,
	))
	sb.WriteString(fmt.Sprintf("| Change | %+.1f%% | %+.1f%% |\n",
		comparison.ReviewChangePercent,
		comparison.MergeChangePercent,
	))

	return sb.String()
}

func (m *MetricsView) qualityIssuesToMarkdown() string {
	var sb strings.Builder

	issues := m.metrics.QualityIssues.Issues
	if len(issues) == 0 {
		sb.WriteString("## PR Quality Issues\n\nNo PR quality issues detected.\n")
		return sb.String()
	}

	filtered := issues
	if m.filteredRepo != "" {
		filtered = make([]models.PRQualityIssue, 0, len(issues))
		for _, issue := range issues {
			if issue.Repository == m.filteredRepo {
				filtered = append(filtered, issue)
			}
		}
		if len(filtered) == 0 {
			sb.WriteString(fmt.Sprintf("## PR Quality Issues\n\nNo PR quality issues found for %s.\n", m.filteredRepo))
			return sb.String()
		}
	}

	var high, medium []models.PRQualityIssue
	for _, issue := range filtered {
		if strings.EqualFold(issue.Severity, "high") {
			high = append(high, issue)
		} else {
			medium = append(medium, issue)
		}
	}

	displayCount := len(filtered)
	if displayCount > metricsformat.MaxQualityIssuesToDisplay {
		displayCount = metricsformat.MaxQualityIssuesToDisplay
	}

	sb.WriteString(fmt.Sprintf("## PR Quality Issues (%d issues)\n\n", displayCount))

	if len(high) > displayCount {
		high = high[:displayCount]
		medium = nil
	} else {
		remaining := displayCount - len(high)
		if remaining < len(medium) {
			medium = medium[:remaining]
		}
	}

	if len(high) > 0 {
		sb.WriteString("### High Priority\n\n")
		sb.WriteString("| Repository | PR | Type | Details | Title |\n")
		sb.WriteString("|------------|----|---------|---------|---------|\n")
		for _, issue := range high {
			sb.WriteString(fmt.Sprintf("| %s | #%d | %s | %s | %s |\n",
				issue.Repository,
				issue.Number,
				issue.IssueType,
				issue.Details,
				issue.Title,
			))
		}
		sb.WriteString("\n")
	}

	if len(medium) > 0 {
		sb.WriteString("### Medium Priority\n\n")
		sb.WriteString("| Repository | PR | Type | Details | Title |\n")
		sb.WriteString("|------------|----|---------|---------|---------|\n")
		for _, issue := range medium {
			sb.WriteString(fmt.Sprintf("| %s | #%d | %s | %s | %s |\n",
				issue.Repository,
				issue.Number,
				issue.IssueType,
				issue.Details,
				issue.Title,
			))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *MetricsView) stagnantPRsToMarkdown() string {
	var sb strings.Builder

	stagnant := m.metrics.StagnantPRs
	sb.WriteString(fmt.Sprintf("## Stagnant PRs (Open > %s)\n\n", metricsformat.FormatDuration(stagnant.Threshold)))

	filteredPRs := stagnant.LongestWaiting
	if m.filteredRepo != "" {
		filteredPRs = []models.StagnantPRInfo{}
		for _, pr := range stagnant.LongestWaiting {
			if pr.Repository == m.filteredRepo {
				filteredPRs = append(filteredPRs, pr)
			}
		}
	}

	if len(filteredPRs) == 0 {
		sb.WriteString("No stagnant PRs found.\n")
		return sb.String()
	}

	if m.filteredRepo == "" {
		sb.WriteString(fmt.Sprintf("**Total stagnant PRs**: %d\n\n", stagnant.TotalStagnant))
	}

	sb.WriteString("| # | Repository | PR | Age | Title |\n")
	sb.WriteString("|---|------------|----|---------|---------|\n")

	for idx, pr := range filteredPRs {
		sb.WriteString(fmt.Sprintf("| %d | %s | #%d | %s | %s |\n",
			idx+1,
			pr.Repository,
			pr.Number,
			metricsformat.FormatDuration(pr.Age),
			pr.Title,
		))
	}

	return sb.String()
}

func (m *MetricsView) repositoryStatsToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("## Per Repository\n\n")

	if len(m.metrics.ByRepository) == 0 {
		sb.WriteString("No repository data available.\n")
		return sb.String()
	}

	repoNames := make([]string, 0, len(m.metrics.ByRepository))
	if m.filteredRepo != "" {
		if _, exists := m.metrics.ByRepository[m.filteredRepo]; exists {
			repoNames = append(repoNames, m.filteredRepo)
		}
	} else {
		for name := range m.metrics.ByRepository {
			repoNames = append(repoNames, name)
		}
		sort.Strings(repoNames)
	}

	if len(repoNames) == 0 {
		sb.WriteString(fmt.Sprintf("No data available for %s.\n", m.filteredRepo))
		return sb.String()
	}

	sb.WriteString("| Repository | Avg | Median | PRs |\n")
	sb.WriteString("|------------|-----|--------|-----|\n")

	for _, name := range repoNames {
		stat := m.metrics.ByRepository[name]
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n",
			name,
			metricsformat.FormatDuration(stat.Average),
			metricsformat.FormatDuration(stat.Median),
			stat.Count,
		))
	}

	return sb.String()
}

// copyToClipboard はテキストをクリップボードにコピーする
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// xclipを試す
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			// xselを試す
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("clipboard tool not found: install xclip or xsel")
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start clipboard command: %w", err)
	}

	if _, err := pipe.Write([]byte(text)); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if err := pipe.Close(); err != nil {
		return fmt.Errorf("failed to close pipe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("clipboard command failed: %w", err)
	}

	return nil
}
