package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

var (
	// ErrMetricsDisabled はメトリクス機能が無効な場合に返される
	ErrMetricsDisabled = errors.New("metrics feature is disabled")
	// ErrLeadTimeMetricsDisabled はリードタイム計測が無効な場合に返される
	ErrLeadTimeMetricsDisabled = errors.New("lead time metrics are disabled")
	// ErrNoRepositoriesConfigured はリポジトリが設定されていない場合に返される
	ErrNoRepositoriesConfigured = errors.New("no repositories configured for metrics")
)

// FetchLeadTimeMetricsUseCase はリードタイムメトリクス取得ユースケース
type FetchLeadTimeMetricsUseCase struct {
	repo repository.MetricsRepository
	cfg  *models.Config
	now  func() time.Time
}

// NewFetchLeadTimeMetricsUseCase はユースケースを生成する
func NewFetchLeadTimeMetricsUseCase(repo repository.MetricsRepository, cfg *models.Config) *FetchLeadTimeMetricsUseCase {
	return &FetchLeadTimeMetricsUseCase{
		repo: repo,
		cfg:  cfg,
		now:  time.Now,
	}
}

// Execute は設定に基づきリードタイムメトリクスを取得する
func (uc *FetchLeadTimeMetricsUseCase) Execute(ctx context.Context, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error) {
	if uc.repo == nil {
		return nil, fmt.Errorf("metrics repository is required")
	}

	if uc.cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if !uc.cfg.Metrics.Enabled {
		return nil, ErrMetricsDisabled
	}

	if !uc.cfg.Metrics.LeadTimeEnabled {
		return nil, ErrLeadTimeMetricsDisabled
	}

	repos := uc.resolveRepositories()
	if len(repos) == 0 {
		return nil, ErrNoRepositoriesConfigured
	}

	period := uc.cfg.Metrics.CalculationPeriod
	if period <= 0 {
		period = 30 * 24 * time.Hour
	}

	since := uc.now().Add(-period)
	metrics, err := uc.repo.FetchLeadTimeMetrics(ctx, repos, since, progressFn)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lead time metrics: %w", err)
	}

	return metrics, nil
}

// GetRateLimit returns current GitHub API rate limit
func (uc *FetchLeadTimeMetricsUseCase) GetRateLimit(ctx context.Context) (*github.Rate, error) {
	if uc.repo == nil {
		return nil, fmt.Errorf("repository is not initialized")
	}
	return uc.repo.GetRateLimit(ctx)
}

func (uc *FetchLeadTimeMetricsUseCase) resolveRepositories() []string {
	if uc.cfg == nil {
		return nil
	}

	repos := make([]string, 0, len(uc.cfg.GitHub.Repositories))
	seen := make(map[string]struct{})
	for _, repo := range uc.cfg.GitHub.Repositories {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		if _, ok := seen[repo]; ok {
			continue
		}
		seen[repo] = struct{}{}
		repos = append(repos, repo)
	}

	if len(repos) > 0 {
		return repos
	}

	owner := strings.TrimSpace(uc.cfg.GitHub.DefaultOwner)
	repo := strings.TrimSpace(uc.cfg.GitHub.DefaultRepo)
	if owner != "" && repo != "" {
		return []string{fmt.Sprintf("%s/%s", owner, repo)}
	}

	return nil
}
