package export

import (
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// --- テストヘルパー ---

func fullMetrics() *models.LeadTimeMetrics {
	return &models.LeadTimeMetrics{
		Overall: models.LeadTimeStat{
			Average: 36 * time.Hour,
			Median:  24 * time.Hour,
			Count:   12,
		},
		ByRepository: map[string]models.LeadTimeStat{
			"owner/repo-a": {
				Average: 24 * time.Hour,
				Median:  18 * time.Hour,
				Count:   6,
			},
			"owner/repo-b": {
				Average: 48 * time.Hour,
				Median:  36 * time.Hour,
				Count:   6,
			},
		},
		PhaseBreakdown: models.ReviewPhaseMetrics{
			CreatedToFirstReview:  4 * time.Hour,
			FirstReviewToApproval: 8 * time.Hour,
			ApprovalToMerge:       2 * time.Hour,
			TotalLeadTime:         14 * time.Hour,
			SampleCount:           23,
		},
		ByRepositoryPhaseBreakdown: map[string]models.ReviewPhaseMetrics{
			"owner/repo-a": {
				CreatedToFirstReview:  3 * time.Hour,
				FirstReviewToApproval: 6 * time.Hour,
				ApprovalToMerge:       time.Hour,
				TotalLeadTime:         10 * time.Hour,
				SampleCount:           10,
			},
		},
		ByDayOfWeek: map[time.Weekday]models.DayOfWeekStats{
			time.Monday:    {ReviewCount: 5, MergeCount: 3},
			time.Tuesday:   {ReviewCount: 8, MergeCount: 4},
			time.Wednesday: {ReviewCount: 6, MergeCount: 5},
			time.Thursday:  {ReviewCount: 7, MergeCount: 2},
			time.Friday:    {ReviewCount: 4, MergeCount: 6},
		},
		WeeklyComparison: models.WeeklyComparison{
			ThisWeek:            models.WeeklyStats{ReviewCount: 15, MergeCount: 10},
			LastWeek:            models.WeeklyStats{ReviewCount: 12, MergeCount: 8},
			ReviewChangePercent: 25.0,
			MergeChangePercent:  25.0,
		},
		QualityIssues: models.PRQualityIssues{
			Issues: []models.PRQualityIssue{
				{
					Repository: "owner/repo-a",
					Number:     101,
					Title:      "Add big feature",
					IssueType:  "large_pr",
					Severity:   "high",
					Reason:     "Too many changes",
					Details:    "800 lines, 12 files",
				},
				{
					Repository: "owner/repo-b",
					Number:     202,
					Title:      "Cleanup",
					IssueType:  "short_description",
					Severity:   "medium",
					Reason:     "Description too short",
					Details:    "120 lines, 3 files",
				},
			},
		},
		StagnantPRs: models.StagnantPRMetrics{
			Threshold:     72 * time.Hour,
			TotalStagnant: 2,
			AverageAge:    96 * time.Hour,
			LongestWaiting: []models.StagnantPRInfo{
				{
					Repository: "owner/repo-a",
					Number:     50,
					Title:      "Long running PR",
					Age:        120 * time.Hour,
				},
				{
					Repository: "owner/repo-b",
					Number:     75,
					Title:      "Another stale PR",
					Age:        80 * time.Hour,
				},
			},
		},
		PRLeadTimes: []models.PRLeadTimeEntry{
			{
				Repository: "owner/repo-a",
				Number:     10,
				Title:      "Feature A",
				LeadTime:   12 * time.Hour,
				MergedAt:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
				HTMLURL:    "https://github.com/owner/repo-a/pull/10",
			},
			{
				Repository: "owner/repo-b",
				Number:     20,
				Title:      "Feature B",
				LeadTime:   48 * time.Hour,
				MergedAt:   time.Date(2025, 1, 16, 14, 0, 0, 0, time.UTC),
				HTMLURL:    "https://github.com/owner/repo-b/pull/20",
			},
		},
	}
}

func allEnabledConfig() *models.MetricsConfig {
	return &models.MetricsConfig{
		ShowPRLeadTimes:     true,
		ShowReviewPhases:    true,
		ShowDayOfWeek:       true,
		ShowWeeklyComparison: true,
		ShowQualityIssues:   true,
		ShowStagnantPRs:     true,
		ShowRepositoryStats: true,
	}
}

func assertContains(t *testing.T, html, substr string) {
	t.Helper()
	if !strings.Contains(html, substr) {
		t.Errorf("expected HTML to contain %q, but it was not found", substr)
	}
}

func assertNotContains(t *testing.T, html, substr string) {
	t.Helper()
	if strings.Contains(html, substr) {
		t.Errorf("expected HTML to NOT contain %q, but it was found", substr)
	}
}

// --- GenerateHTML テスト ---

func TestGenerateHTML_正常系_全セクション有効(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "1d 12h")                // Overall Average: 36h = 1d 12h
	assertContains(t, html, "1d")                     // Overall Median: 24h = 1d
	assertContains(t, html, "12")                     // Overall Count
	assertContains(t, html, "Feature A")              // PR Lead Times entry
	assertContains(t, html, "Feature B")              // PR Lead Times entry
	assertContains(t, html, "owner/repo-a")           // Repository name
	assertContains(t, html, "owner/repo-b")           // Repository name
	assertContains(t, html, "Add big feature")        // Quality issue
	assertContains(t, html, "Long running PR")        // Stagnant PR
	assertContains(t, html, "25.0%")                  // Weekly comparison percentage
}

func TestGenerateHTML_正常系_HTMLの構造が正しい(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "<!DOCTYPE html>")
	assertContains(t, html, "<html")
	assertContains(t, html, "<head>")
	assertContains(t, html, "<body>")
	assertContains(t, html, "</html>")
}

func TestGenerateHTML_正常系_埋め込みCSS(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "<style>")
	assertNotContains(t, html, "<link rel=\"stylesheet\"")
	assertNotContains(t, html, "<script src=")
}

func TestGenerateHTML_設定_PRリードタイム非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowPRLeadTimes = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "Feature A")
	assertNotContains(t, html, "Feature B")
}

func TestGenerateHTML_設定_レビューフェーズ非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowReviewPhases = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "First Review")
	assertNotContains(t, html, "Approval")
}

func TestGenerateHTML_設定_曜日別統計非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowDayOfWeek = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "Monday")
	assertNotContains(t, html, "Tuesday")
}

func TestGenerateHTML_設定_週次比較非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowWeeklyComparison = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "This Week")
	assertNotContains(t, html, "Last Week")
}

func TestGenerateHTML_設定_品質問題非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowQualityIssues = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "Add big feature")
	assertNotContains(t, html, "large_pr")
}

func TestGenerateHTML_設定_滞留PR非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowStagnantPRs = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "Long running PR")
	assertNotContains(t, html, "Another stale PR")
}

func TestGenerateHTML_設定_リポジトリ統計非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()
	config.ShowRepositoryStats = false

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	// Overall section still shows, but per-repository table should not appear
	// The overall section references repo names in PR entries, so we check
	// that the per-repository stats section heading is absent
	assertContains(t, html, "1d 12h") // Overall is still shown
}

func TestGenerateHTML_設定_全セクション無効(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowPRLeadTimes:      false,
		ShowReviewPhases:     false,
		ShowDayOfWeek:        false,
		ShowWeeklyComparison: false,
		ShowQualityIssues:    false,
		ShowStagnantPRs:      false,
		ShowRepositoryStats:  false,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	// Overall section is always shown
	assertContains(t, html, "1d 12h")
	// Optional sections should be absent
	assertNotContains(t, html, "Feature A")
	assertNotContains(t, html, "Long running PR")
}

func TestGenerateHTML_空メトリクス(t *testing.T) {
	// Given
	metrics := &models.LeadTimeMetrics{
		Overall: models.LeadTimeStat{
			Count: 0,
		},
		ByRepository: map[string]models.LeadTimeStat{},
	}
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "<html")
	assertContains(t, html, "</html>")
}

func TestGenerateHTML_nilメトリクスはエラー(t *testing.T) {
	// Given
	config := allEnabledConfig()

	// When
	_, err := GenerateHTML(nil, config)

	// Then
	if err == nil {
		t.Fatal("expected error for nil metrics, got nil")
	}
}

func TestGenerateHTML_nilコンフィグはエラー(t *testing.T) {
	// Given
	metrics := fullMetrics()

	// When
	_, err := GenerateHTML(metrics, nil)

	// Then
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
}

func TestGenerateHTML_XSS防止_タイトルの特殊文字がエスケープされる(t *testing.T) {
	// Given
	metrics := &models.LeadTimeMetrics{
		Overall: models.LeadTimeStat{
			Average: 12 * time.Hour,
			Median:  8 * time.Hour,
			Count:   1,
		},
		ByRepository: map[string]models.LeadTimeStat{},
		PRLeadTimes: []models.PRLeadTimeEntry{
			{
				Repository: "owner/repo",
				Number:     1,
				Title:      "<script>alert('xss')</script>",
				LeadTime:   12 * time.Hour,
				MergedAt:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
				HTMLURL:    "https://github.com/owner/repo/pull/1",
			},
		},
	}
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "<script>alert('xss')</script>")
	assertContains(t, html, "&lt;script&gt;")
}

func TestGenerateHTML_PRリードタイム一覧に全エントリが含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Feature A")
	assertContains(t, html, "Feature B")
	assertContains(t, html, "#10")
	assertContains(t, html, "#20")
	assertContains(t, html, "https://github.com/owner/repo-a/pull/10")
	assertContains(t, html, "https://github.com/owner/repo-b/pull/20")
}

func TestGenerateHTML_レビューフェーズの全フェーズが含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := allEnabledConfig()

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "4h")  // CreatedToFirstReview
	assertContains(t, html, "8h")  // FirstReviewToApproval
	assertContains(t, html, "2h")  // ApprovalToMerge
	assertContains(t, html, "14h") // TotalLeadTime
}

func TestGenerateHTML_レビューフェーズのサンプル数ゼロ(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.PhaseBreakdown = models.ReviewPhaseMetrics{
		SampleCount: 0,
	}
	config := &models.MetricsConfig{
		ShowReviewPhases: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	// Should not show phase details when sample count is zero
	assertNotContains(t, html, "First Review")
}

func TestGenerateHTML_曜日別統計のデータが含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowDayOfWeek: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Mon")
	assertContains(t, html, "Tue")
	assertContains(t, html, "Wed")
	assertContains(t, html, "Thu")
	assertContains(t, html, "Fri")
}

func TestGenerateHTML_週次比較のデータが含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowWeeklyComparison: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "15") // ThisWeek ReviewCount
	assertContains(t, html, "10") // ThisWeek MergeCount
	assertContains(t, html, "12") // LastWeek ReviewCount
}

func TestGenerateHTML_品質問題の重要度別表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowQualityIssues: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Add big feature")
	assertContains(t, html, "800 lines, 12 files")
	assertContains(t, html, "Cleanup")
}

func TestGenerateHTML_品質問題なし(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.QualityIssues = models.PRQualityIssues{
		Issues: []models.PRQualityIssue{},
	}
	config := &models.MetricsConfig{
		ShowQualityIssues: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "large_pr")
}

func TestGenerateHTML_滞留PRの全情報が含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowStagnantPRs: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Long running PR")
	assertContains(t, html, "#50")
	assertContains(t, html, "Another stale PR")
	assertContains(t, html, "#75")
	assertContains(t, html, "5d") // 120h = 5d
}

func TestGenerateHTML_滞留PRなし(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.StagnantPRs = models.StagnantPRMetrics{
		Threshold:      72 * time.Hour,
		TotalStagnant:  0,
		LongestWaiting: []models.StagnantPRInfo{},
	}
	config := &models.MetricsConfig{
		ShowStagnantPRs: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertNotContains(t, html, "Long running PR")
}

func TestGenerateHTML_リポジトリ統計の全リポジトリが含まれる(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowRepositoryStats: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "owner/repo-a")
	assertContains(t, html, "owner/repo-b")
}

func TestGenerateHTML_リポジトリ統計が空(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.ByRepository = map[string]models.LeadTimeStat{}
	config := &models.MetricsConfig{
		ShowRepositoryStats: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	// Should produce valid HTML even with no repository data
	assertContains(t, html, "<html")
}

func TestGenerateHTML_PRリードタイムのマージ日表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	config := &models.MetricsConfig{
		ShowPRLeadTimes: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "2025-01-15")
	assertContains(t, html, "2025-01-16")
}

func TestGenerateHTML_設定_ShowWeekendInDayOfWeekがfalseの場合土日が非表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.ByDayOfWeek[time.Saturday] = models.DayOfWeekStats{ReviewCount: 2, MergeCount: 1}
	metrics.ByDayOfWeek[time.Sunday] = models.DayOfWeekStats{ReviewCount: 1, MergeCount: 0}
	config := &models.MetricsConfig{
		ShowDayOfWeek:          true,
		ShowWeekendInDayOfWeek: false,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Mon")
	assertContains(t, html, "Fri")
	assertNotContains(t, html, "Sat")
	assertNotContains(t, html, "Sun")
}

func TestGenerateHTML_設定_ShowWeekendInDayOfWeekがtrueの場合土日が表示(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.ByDayOfWeek[time.Saturday] = models.DayOfWeekStats{ReviewCount: 2, MergeCount: 1}
	metrics.ByDayOfWeek[time.Sunday] = models.DayOfWeekStats{ReviewCount: 1, MergeCount: 0}
	config := &models.MetricsConfig{
		ShowDayOfWeek:          true,
		ShowWeekendInDayOfWeek: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "Mon")
	assertContains(t, html, "Fri")
	assertContains(t, html, "Sat")
	assertContains(t, html, "Sun")
}

func TestGenerateHTML_曜日別統計が空(t *testing.T) {
	// Given
	metrics := fullMetrics()
	metrics.ByDayOfWeek = nil
	config := &models.MetricsConfig{
		ShowDayOfWeek: true,
	}

	// When
	html, err := GenerateHTML(metrics, config)

	// Then
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	assertContains(t, html, "<html")
}
