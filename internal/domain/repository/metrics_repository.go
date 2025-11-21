package repository

import (
	"context"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/google/go-github/v57/github"
)

// MetricsRepository はメトリクス関連のデータ取得を担当する
type MetricsRepository interface {
	FetchLeadTimeMetrics(ctx context.Context, repos []string, since time.Time, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error)
	GetRateLimit(ctx context.Context) (*github.Rate, error)
}
