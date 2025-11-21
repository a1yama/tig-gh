package github

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

const (
	stagnantPRThreshold = 72 * time.Hour // 3 days
	reviewWorkerCount   = 12             // concurrent review fetchers
)

type leadTimeSample struct {
	duration      time.Duration
	mergedAt      time.Time
	firstReviewAt *time.Time
}

// MetricsRepositoryImpl は MetricsRepository を実装する
type MetricsRepositoryImpl struct {
	client *Client
}

// NewMetricsRepository は MetricsRepository 実装を生成する
func NewMetricsRepository(client *Client) repository.MetricsRepository {
	return &MetricsRepositoryImpl{client: client}
}

// GetRateLimit returns the current GitHub API rate limit status
func (r *MetricsRepositoryImpl) GetRateLimit(ctx context.Context) (*github.Rate, error) {
	limits, _, err := r.client.client.RateLimits(ctx)
	if err != nil {
		return nil, err
	}

	if limits != nil && limits.Core != nil {
		return limits.Core, nil
	}

	return nil, nil
}

// FetchLeadTimeMetrics は複数リポジトリのリードタイムメトリクスを取得する
func (r *MetricsRepositoryImpl) FetchLeadTimeMetrics(ctx context.Context, repos []string, since time.Time, progressFn func(models.MetricsProgress)) (*models.LeadTimeMetrics, error) {
	result := &models.LeadTimeMetrics{
		Overall:               models.LeadTimeStat{},
		ByRepository:          make(map[string]models.LeadTimeStat),
		Trend:                 []models.TrendPoint{},
		ByDayOfWeek:           make(map[time.Weekday]models.DayOfWeekStats),
		ByRepositoryDayOfWeek: make(map[string]map[time.Weekday]models.DayOfWeekStats),
		ByRepositoryWeekly:    make(map[string]models.WeeklyComparison),
	}

	if len(repos) == 0 {
		return result, nil
	}

	var (
		mu          sync.Mutex
		repoSamples = make(map[string][]leadTimeSample)
		errs        []error
		wg          sync.WaitGroup
	)

	for i, repoFull := range repos {
		repoFull = strings.TrimSpace(repoFull)
		if repoFull == "" {
			continue
		}

		if progressFn != nil {
			progressFn(models.MetricsProgress{
				TotalRepos:     len(repos),
				ProcessedRepos: i,
				CurrentRepo:    repoFull,
			})
		}

		owner, name, err := parseRepositorySlug(repoFull)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(slug, owner, name string) {
			defer wg.Done()

			samples, fetchErr := r.fetchLeadTimeSamples(ctx, owner, name, since)

			mu.Lock()
			defer mu.Unlock()

			if fetchErr != nil {
				errs = append(errs, fmt.Errorf("%s: %w", slug, fetchErr))
				return
			}

			repoSamples[slug] = samples
		}(repoFull, owner, name)
	}

	wg.Wait()

	if progressFn != nil {
		progressFn(models.MetricsProgress{
			TotalRepos:     len(repos),
			ProcessedRepos: len(repos),
			CurrentRepo:    "",
		})
	}

	var overallSamples []leadTimeSample

	currentTime := time.Now()

	for slug, samples := range repoSamples {
		durations := samplesToDurations(samples)
		result.ByRepository[slug] = calculateLeadTimeStat(durations)
		result.ByRepositoryDayOfWeek[slug] = aggregateByDayOfWeek(samples)
		result.ByRepositoryWeekly[slug] = calculateWeeklyComparison(samples, currentTime)
		overallSamples = append(overallSamples, samples...)
	}

	allDurations := samplesToDurations(overallSamples)
	result.Overall = calculateLeadTimeStat(allDurations)
	result.ByDayOfWeek = aggregateByDayOfWeek(overallSamples)
	result.WeeklyComparison = calculateWeeklyComparison(overallSamples, currentTime)

	// Fetch stagnant PR metrics
	stagnantMetrics, err := r.fetchStagnantPRMetrics(ctx, repos, time.Now())
	if err != nil {
		fmt.Printf("failed to fetch stagnant PR metrics: %v\n", err)
	} else {
		result.StagnantPRs = stagnantMetrics
	}

	if len(repoSamples) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}

	return result, nil
}

func (r *MetricsRepositoryImpl) fetchLeadTimeSamples(ctx context.Context, owner, repo string, since time.Time) ([]leadTimeSample, error) {
	defaultBranch, err := r.getDefaultBranch(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	opts := &github.PullRequestListOptions{
		State:     "closed",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var samples []leadTimeSample
	var reviewRequests []reviewRequest

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		prs, resp, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, handleGitHubError(err, resp)
		}

		stop := false
		for _, pr := range prs {
			if pr == nil || pr.MergedAt == nil || pr.CreatedAt == nil {
				continue
			}

			if base := pr.GetBase(); base == nil || base.GetRef() != defaultBranch {
				continue
			}

			mergedAt := pr.MergedAt.Time
			if mergedAt.Before(since) {
				stop = true
				continue
			}

			createdAt := pr.CreatedAt.Time
			if mergedAt.Before(createdAt) {
				continue
			}

			samples = append(samples, leadTimeSample{
				duration: mergedAt.Sub(createdAt),
				mergedAt: mergedAt,
			})
			lastIdx := len(samples) - 1
			reviewRequests = append(reviewRequests, reviewRequest{
				sampleIndex: lastIdx,
				number:      pr.GetNumber(),
			})
		}

		nextPage := 0
		if resp != nil {
			nextPage = resp.NextPage
		}
		if nextPage == 0 || stop {
			break
		}

		opts.Page = nextPage
	}

	if err := r.populateFirstReviewTimes(ctx, owner, repo, samples, reviewRequests); err != nil {
		return nil, err
	}

	return samples, nil
}

type reviewRequest struct {
	sampleIndex int
	number      int
}

func (r *MetricsRepositoryImpl) populateFirstReviewTimes(ctx context.Context, owner, repo string, samples []leadTimeSample, requests []reviewRequest) error {
	if len(requests) == 0 {
		return nil
	}

	workerCount := reviewWorkerCount
	if len(requests) < workerCount {
		workerCount = len(requests)
	}

	jobs := make(chan reviewRequest)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range jobs {
				if ctx.Err() != nil {
					return
				}
				review := r.fetchSampleFirstReview(ctx, owner, repo, req.number)
				samples[req.sampleIndex].firstReviewAt = review
			}
		}()
	}

sendLoop:
	for _, req := range requests {
		select {
		case <-ctx.Done():
			break sendLoop
		case jobs <- req:
		}
	}
	close(jobs)
	wg.Wait()

	return ctx.Err()
}

func (r *MetricsRepositoryImpl) fetchSampleFirstReview(ctx context.Context, owner, repo string, number int) *time.Time {
	firstReview, err := r.fetchFirstReviewTime(ctx, owner, repo, number)
	if err != nil {
		fmt.Printf("failed to fetch reviews for %s/%s#%d: %v\n", owner, repo, number, err)
		return nil
	}
	return firstReview
}

func (r *MetricsRepositoryImpl) fetchFirstReviewTime(ctx context.Context, owner, repo string, number int) (*time.Time, error) {
	opts := &github.ListOptions{PerPage: 100}
	var first time.Time
	found := false

	for {
		reviews, resp, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, handleGitHubError(err, resp)
		}

		for _, review := range reviews {
			if review == nil || review.SubmittedAt == nil {
				continue
			}
			submitted := review.SubmittedAt.Time
			if !found || submitted.Before(first) {
				first = submitted
				found = true
			}
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if !found {
		return nil, nil
	}

	firstCopy := first
	return &firstCopy, nil
}

func aggregateByDayOfWeek(samples []leadTimeSample) map[time.Weekday]models.DayOfWeekStats {
	stats := make(map[time.Weekday]models.DayOfWeekStats, 7)

	for _, sample := range samples {
		mergeDay := sample.mergedAt.Weekday()
		mergeStats := stats[mergeDay]
		mergeStats.MergeCount++
		stats[mergeDay] = mergeStats

		if sample.firstReviewAt != nil {
			reviewDay := sample.firstReviewAt.Weekday()
			reviewStats := stats[reviewDay]
			reviewStats.ReviewCount++
			stats[reviewDay] = reviewStats
		}
	}

	for day := time.Sunday; day <= time.Saturday; day++ {
		if _, exists := stats[day]; !exists {
			stats[day] = models.DayOfWeekStats{}
		}
	}

	return stats
}

func calculateWeeklyComparison(samples []leadTimeSample, now time.Time) models.WeeklyComparison {
	thisWeekStart := now.AddDate(0, 0, -7)
	lastWeekStart := now.AddDate(0, 0, -14)

	var thisWeek models.WeeklyStats
	var lastWeek models.WeeklyStats

	for _, sample := range samples {
		mergedAt := sample.mergedAt
		switch {
		case !mergedAt.Before(thisWeekStart) && !mergedAt.After(now):
			thisWeek.MergeCount++
		case !mergedAt.Before(lastWeekStart) && mergedAt.Before(thisWeekStart):
			lastWeek.MergeCount++
		}

		if sample.firstReviewAt == nil {
			continue
		}

		reviewAt := *sample.firstReviewAt
		switch {
		case !reviewAt.Before(thisWeekStart) && !reviewAt.After(now):
			thisWeek.ReviewCount++
		case !reviewAt.Before(lastWeekStart) && reviewAt.Before(thisWeekStart):
			lastWeek.ReviewCount++
		}
	}

	return models.WeeklyComparison{
		ThisWeek:            thisWeek,
		LastWeek:            lastWeek,
		ReviewChangePercent: calculatePercentChange(thisWeek.ReviewCount, lastWeek.ReviewCount),
		MergeChangePercent:  calculatePercentChange(thisWeek.MergeCount, lastWeek.MergeCount),
	}
}

func calculatePercentChange(current, previous int) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}

	return (float64(current-previous) / float64(previous)) * 100
}

func (r *MetricsRepositoryImpl) fetchStagnantPRMetrics(ctx context.Context, repos []string, now time.Time) (models.StagnantPRMetrics, error) {
	var allStagnantPRs []models.StagnantPRInfo

	for _, repoSlug := range repos {
		owner, repo, err := parseRepositorySlug(repoSlug)
		if err != nil {
			continue
		}

		opts := &github.PullRequestListOptions{
			State:       "open",
			Sort:        "created",
			Direction:   "asc",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		prs, _, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			if pr == nil || pr.CreatedAt == nil {
				continue
			}

			age := now.Sub(pr.CreatedAt.Time)
			if age >= stagnantPRThreshold {
				allStagnantPRs = append(allStagnantPRs, models.StagnantPRInfo{
					Repository: repoSlug,
					Number:     pr.GetNumber(),
					Title:      pr.GetTitle(),
					Age:        age,
				})
			}
		}
	}

	if len(allStagnantPRs) == 0 {
		return models.StagnantPRMetrics{
			Threshold: stagnantPRThreshold,
		}, nil
	}

	var totalAge time.Duration
	for i := range allStagnantPRs {
		totalAge += allStagnantPRs[i].Age
	}

	sort.Slice(allStagnantPRs, func(i, j int) bool {
		return allStagnantPRs[i].Age > allStagnantPRs[j].Age
	})

	topCount := 10
	if len(allStagnantPRs) < topCount {
		topCount = len(allStagnantPRs)
	}
	longestWaiting := append([]models.StagnantPRInfo(nil), allStagnantPRs[:topCount]...)

	return models.StagnantPRMetrics{
		Threshold:      stagnantPRThreshold,
		TotalStagnant:  len(allStagnantPRs),
		AverageAge:     time.Duration(int64(totalAge) / int64(len(allStagnantPRs))),
		LongestWaiting: longestWaiting,
	}, nil
}

func parseRepositorySlug(slug string) (string, string, error) {
	parts := strings.Split(slug, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", slug)
	}

	owner := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if owner == "" || name == "" {
		return "", "", fmt.Errorf("invalid repository format: %s", slug)
	}

	return owner, name, nil
}

func samplesToDurations(samples []leadTimeSample) []time.Duration {
	if len(samples) == 0 {
		return nil
	}

	durations := make([]time.Duration, 0, len(samples))
	for _, sample := range samples {
		durations = append(durations, sample.duration)
	}

	return durations
}

func calculateLeadTimeStat(durations []time.Duration) models.LeadTimeStat {
	count := len(durations)
	if count == 0 {
		return models.LeadTimeStat{}
	}

	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	total := time.Duration(0)
	for _, d := range sorted {
		total += d
	}

	avg := time.Duration(int64(total) / int64(count))
	median := calculateMedian(sorted)

	return models.LeadTimeStat{
		Average: avg,
		Median:  median,
		Count:   count,
	}
}

func calculateMedian(sorted []time.Duration) time.Duration {
	n := len(sorted)
	if n == 0 {
		return 0
	}

	if n%2 == 1 {
		return sorted[n/2]
	}

	a := sorted[n/2-1]
	b := sorted[n/2]
	return time.Duration((a.Nanoseconds() + b.Nanoseconds()) / 2)
}

func (r *MetricsRepositoryImpl) getDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	repository, resp, err := r.client.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", handleGitHubError(err, resp)
	}

	branch := repository.GetDefaultBranch()
	if branch == "" {
		return "", fmt.Errorf("default branch not found for %s/%s", owner, repo)
	}

	return branch, nil
}
