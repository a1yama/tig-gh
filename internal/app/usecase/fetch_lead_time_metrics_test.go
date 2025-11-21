package usecase

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v57/github"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

type stubMetricsRepository struct {
	metrics *models.LeadTimeMetrics
	err     error

	called bool
	repos  []string
	since  time.Time
}

func (s *stubMetricsRepository) FetchLeadTimeMetrics(ctx context.Context, repos []string, since time.Time, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error) {
	s.called = true
	s.repos = append([]string{}, repos...)
	s.since = since

	if progressFn != nil {
		for i, repo := range repos {
			progressFn(models.MetricsProgress{
				TotalRepos:     len(repos),
				ProcessedRepos: i,
				CurrentRepo:    repo,
			})
		}
		progressFn(models.MetricsProgress{
			TotalRepos:     len(repos),
			ProcessedRepos: len(repos),
			CurrentRepo:    "",
		})
	}

	if s.err != nil {
		return nil, s.err
	}

	if s.metrics == nil {
		return &models.LeadTimeMetrics{
			Overall:      models.LeadTimeStat{},
			ByRepository: map[string]models.LeadTimeStat{},
			Trend:        nil,
		}, nil
	}

	return s.metrics, nil
}

func (s *stubMetricsRepository) GetRateLimit(ctx context.Context) (*github.Rate, error) {
	return &github.Rate{
		Limit:     5000,
		Remaining: 4500,
		Reset:     github.Timestamp{Time: time.Now().Add(time.Hour)},
	}, nil
}

func TestFetchLeadTimeMetricsUseCase_ExecuteSuccess(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.LeadTimeEnabled = true
	cfg.Metrics.CalculationPeriod = 72 * time.Hour
	cfg.GitHub.Repositories = []string{" owner/repo1 ", "owner/repo1", "owner/repo2"}

	expected := &models.LeadTimeMetrics{
		Overall:      models.LeadTimeStat{Count: 3},
		ByRepository: map[string]models.LeadTimeStat{"owner/repo1": {Count: 2}, "owner/repo2": {Count: 1}},
		Trend:        nil,
	}

	repo := &stubMetricsRepository{
		metrics: expected,
	}

	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)
	fixedNow := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }

	result, err := uc.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result != expected {
		t.Fatalf("expected metrics pointer to match")
	}

	expectedRepos := []string{"owner/repo1", "owner/repo2"}
	if !reflect.DeepEqual(repo.repos, expectedRepos) {
		t.Fatalf("unexpected repositories passed: %+v", repo.repos)
	}

	expectedSince := fixedNow.Add(-cfg.Metrics.CalculationPeriod)
	if !repo.since.Equal(expectedSince) {
		t.Fatalf("unexpected since value: %v", repo.since)
	}
}

func TestFetchLeadTimeMetricsUseCase_Disabled(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = false

	repo := &stubMetricsRepository{}
	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)

	_, err := uc.Execute(context.Background(), nil)
	if !errors.Is(err, ErrMetricsDisabled) {
		t.Fatalf("expected ErrMetricsDisabled, got %v", err)
	}

	if repo.called {
		t.Fatalf("repository should not be called when disabled")
	}
}

func TestFetchLeadTimeMetricsUseCase_LeadTimeDisabled(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.LeadTimeEnabled = false

	repo := &stubMetricsRepository{}
	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)

	_, err := uc.Execute(context.Background(), nil)
	if !errors.Is(err, ErrLeadTimeMetricsDisabled) {
		t.Fatalf("expected ErrLeadTimeMetricsDisabled, got %v", err)
	}

	if repo.called {
		t.Fatalf("repository should not be called when lead time disabled")
	}
}

func TestFetchLeadTimeMetricsUseCase_NoRepositories(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.LeadTimeEnabled = true
	cfg.GitHub.DefaultOwner = ""
	cfg.GitHub.DefaultRepo = ""
	cfg.GitHub.Repositories = nil

	repo := &stubMetricsRepository{}
	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)

	_, err := uc.Execute(context.Background(), nil)
	if !errors.Is(err, ErrNoRepositoriesConfigured) {
		t.Fatalf("expected ErrNoRepositoriesConfigured, got %v", err)
	}
}

func TestFetchLeadTimeMetricsUseCase_RepositoryError(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.LeadTimeEnabled = true
	cfg.GitHub.Repositories = []string{"owner/repo"}

	repo := &stubMetricsRepository{
		err: errors.New("boom"),
	}

	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)

	_, err := uc.Execute(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch lead time metrics") {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestFetchLeadTimeMetricsUseCase_FallbackDefaultRepo(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.LeadTimeEnabled = true
	cfg.GitHub.DefaultOwner = "default-owner"
	cfg.GitHub.DefaultRepo = "default-repo"
	cfg.GitHub.Repositories = nil

	repo := &stubMetricsRepository{}
	uc := NewFetchLeadTimeMetricsUseCase(repo, cfg)
	uc.now = func() time.Time { return time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC) }

	if _, err := uc.Execute(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.repos) != 1 || repo.repos[0] != "default-owner/default-repo" {
		t.Fatalf("expected default repo fallback, got %+v", repo.repos)
	}
}
