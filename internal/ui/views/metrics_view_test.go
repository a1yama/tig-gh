package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v57/github"
)

type stubLeadTimeUseCase struct {
	metrics   *models.LeadTimeMetrics
	err       error
	callCount int
	rateLimit *github.Rate
}

func (s *stubLeadTimeUseCase) Execute(ctx context.Context, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error) {
	s.callCount++
	if s.err != nil {
		return nil, s.err
	}
	return s.metrics, nil
}

func (s *stubLeadTimeUseCase) GetRateLimit(ctx context.Context) (*github.Rate, error) {
	if s.rateLimit == nil {
		// Return a default rate limit for testing
		return &github.Rate{
			Limit:     5000,
			Remaining: 4850,
		}, nil
	}
	return s.rateLimit, nil
}

func TestMetricsViewInitFetchesData(t *testing.T) {
	metrics := sampleMetrics()
	useCase := &stubLeadTimeUseCase{metrics: metrics}

	view := NewMetricsViewWithUseCase(useCase)
	view.Update(tea.WindowSizeMsg{Width: 90, Height: 30})

	cmd := view.Init()
	if cmd == nil {
		t.Fatal("expected init command")
	}

	msg := cmd()

	var metricsMsg metricsLoadedMsg
	switch m := msg.(type) {
	case metricsLoadedMsg:
		metricsMsg = m
	case tea.BatchMsg:
		found := false
		for _, batchCmd := range m {
			if batchCmd == nil {
				continue
			}
			res := batchCmd()
			if lm, ok := res.(metricsLoadedMsg); ok {
				metricsMsg = lm
				found = true
				break
			}
		}
		if !found {
			t.Fatal("BatchMsg did not contain metricsLoadedMsg")
		}
	default:
		t.Fatalf("expected metricsLoadedMsg, got %T", msg)
	}

	view.Update(metricsMsg)
	if view.metrics != metrics {
		t.Fatalf("expected metrics to be set")
	}
	if useCase.callCount != 1 {
		t.Fatalf("expected Execute to be called once")
	}
}

func TestMetricsViewViewContainsSections(t *testing.T) {
	metrics := sampleMetrics()
	view := NewMetricsView()
	view.metrics = metrics
	view.lastUpdated = time.Now()
	view.Update(tea.WindowSizeMsg{Width: 100, Height: 25})

	output := view.View()
	assertContains(t, output, "Overall Metrics")
	assertContains(t, output, "Per Repository")
	assertContains(t, output, "owner/repo-a")
	assertContains(t, output, "PR Quality Issues (Top 10)")
	assertContains(t, output, "High Priority:")
}

func TestMetricsViewErrorState(t *testing.T) {
	view := NewMetricsView()
	view.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	errMsg := metricsLoadedMsg{metrics: nil, err: assertError("boom")}
	view.Update(errMsg)

	output := view.View()
	assertContains(t, output, "boom")
	assertContains(t, output, "Press 'r' to retry")
}

func TestMetricsViewScrollAndRefresh(t *testing.T) {
	metrics := sampleMetrics()
	view := NewMetricsViewWithUseCase(&stubLeadTimeUseCase{metrics: metrics})
	view.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	view.metrics = metrics

	maxScroll := view.maxScroll()
	if maxScroll == 0 {
		t.Fatalf("expected maxScroll > 0 for test data")
	}

	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if view.scroll == 0 {
		t.Fatalf("expected scroll to increase on 'j'")
	}

	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if view.scroll != 0 {
		t.Fatalf("expected scroll to decrease on 'k'")
	}

	_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatalf("expected refresh command on 'r'")
	}

	_, cmd = view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected exit command on 'q'")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected exit message")
	}
}

// Helpers

func sampleMetrics() *models.LeadTimeMetrics {
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
		Trend: []models.TrendPoint{
			{Period: "2025-W01", AverageLeadTime: 24 * time.Hour, PRCount: 4},
			{Period: "2025-W02", AverageLeadTime: 48 * time.Hour, PRCount: 5},
			{Period: "2025-W03", AverageLeadTime: 12 * time.Hour, PRCount: 3},
		},
		QualityIssues: models.PRQualityIssues{
			Issues: []models.PRQualityIssue{
				{
					Repository:     "owner/repo-a",
					Number:         101,
					Title:          "Add big feature",
					IssueType:      "large_pr",
					Severity:       "high",
					Reason:         "レビューに時間がかかり、バグが見落とされやすい",
					Recommendation: "機能ごとに分割し、200-400行に抑える",
					Details:        "800 lines, 12 files",
				},
				{
					Repository:     "owner/repo-b",
					Number:         202,
					Title:          "Cleanup",
					IssueType:      "short_description",
					Severity:       "medium",
					Reason:         "テンプレートのままの可能性",
					Recommendation: "変更の背景と影響範囲を追記",
					Details:        "120 lines, 3 files",
				},
			},
		},
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func assertError(msg string) error { return errString(msg) }

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Fatalf("expected output to contain %q\n%s", substr, output)
	}
}
