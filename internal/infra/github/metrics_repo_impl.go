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
	stagnantPRThreshold       = 72 * time.Hour // 3 days
	awaitingReviewThreshold   = 24 * time.Hour // 24 hours
	awaitingMergeThreshold    = 12 * time.Hour // 12 hours
	leadTimeIncreaseThreshold = 0.4            // 40% increase
)

type leadTimeSample struct {
	duration time.Duration
	mergedAt time.Time
}

type prPhaseData struct {
	createdAt       time.Time
	firstReviewAt   *time.Time
	firstApprovalAt *time.Time
	mergedAt        time.Time
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
		Overall:      models.LeadTimeStat{},
		ByRepository: make(map[string]models.LeadTimeStat),
		Trend:        []models.TrendPoint{},
		PhaseBreakdown: models.ReviewPhaseMetrics{
			CreatedToFirstReview:  0,
			FirstReviewToApproval: 0,
			ApprovalToMerge:       0,
			TotalLeadTime:         0,
		},
	}

	if len(repos) == 0 {
		return result, nil
	}

	var (
		mu          sync.Mutex
		repoSamples = make(map[string][]leadTimeSample)
		repoPhases  = make(map[string][]prPhaseData)
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

			samples, phases, fetchErr := r.fetchLeadTimeSamples(ctx, owner, name, since)

			mu.Lock()
			defer mu.Unlock()

			if fetchErr != nil {
				errs = append(errs, fmt.Errorf("%s: %w", slug, fetchErr))
				return
			}

			repoSamples[slug] = samples
			repoPhases[slug] = phases
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
	var phaseData []prPhaseData

	for slug, samples := range repoSamples {
		durations := samplesToDurations(samples)
		result.ByRepository[slug] = calculateLeadTimeStat(durations)
		overallSamples = append(overallSamples, samples...)
	}

	for _, phases := range repoPhases {
		phaseData = append(phaseData, phases...)
	}

	allDurations := samplesToDurations(overallSamples)
	result.Overall = calculateLeadTimeStat(allDurations)
	result.Trend = buildTrendPoints(overallSamples)
	result.PhaseBreakdown = calculatePhaseBreakdown(phaseData)

	// Fetch stagnant PR metrics
	stagnantMetrics, err := r.fetchStagnantPRMetrics(ctx, repos, time.Now())
	if err != nil {
		fmt.Printf("failed to fetch stagnant PR metrics: %v\n", err)
	} else {
		result.StagnantPRs = stagnantMetrics
	}

	// Generate alerts
	alertMetrics, err := r.generateAlerts(ctx, repos, result.Trend, time.Now())
	if err != nil {
		fmt.Printf("failed to generate alerts: %v\n", err)
	} else {
		result.Alerts = alertMetrics
	}

	if len(repoSamples) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}

	return result, nil
}

func (r *MetricsRepositoryImpl) fetchLeadTimeSamples(ctx context.Context, owner, repo string, since time.Time) ([]leadTimeSample, []prPhaseData, error) {
	defaultBranch, err := r.getDefaultBranch(ctx, owner, repo)
	if err != nil {
		return nil, nil, err
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
	var phases []prPhaseData

	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		prs, resp, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, nil, handleGitHubError(err, resp)
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

			firstReviewAt, firstApprovalAt, reviewErr := r.fetchPRReviewData(ctx, owner, repo, pr.GetNumber())
			if reviewErr != nil {
				continue
			}

			phases = append(phases, prPhaseData{
				createdAt:       createdAt,
				firstReviewAt:   firstReviewAt,
				firstApprovalAt: firstApprovalAt,
				mergedAt:        mergedAt,
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

	return samples, phases, nil
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

	var longest *models.StagnantPRInfo
	var totalAge time.Duration

	for i := range allStagnantPRs {
		pr := &allStagnantPRs[i]
		totalAge += pr.Age

		if longest == nil || pr.Age > longest.Age {
			longest = pr
		}
	}

	return models.StagnantPRMetrics{
		Threshold:      stagnantPRThreshold,
		TotalStagnant:  len(allStagnantPRs),
		AverageAge:     time.Duration(int64(totalAge) / int64(len(allStagnantPRs))),
		LongestWaiting: longest,
	}, nil
}

func (r *MetricsRepositoryImpl) generateAlerts(ctx context.Context, repos []string, currentTrend []models.TrendPoint, now time.Time) (models.AlertMetrics, error) {
	var alerts []models.Alert

	// Alert 1: PRs awaiting review for > 24h
	awaitingReviewCount := 0
	for _, repoSlug := range repos {
		owner, repo, err := parseRepositorySlug(repoSlug)
		if err != nil {
			continue
		}

		opts := &github.PullRequestListOptions{
			State:       "open",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		prs, _, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			if pr.CreatedAt == nil {
				continue
			}

			age := now.Sub(pr.CreatedAt.Time)
			if age >= awaitingReviewThreshold {
				reviews, _, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), nil)
				if err == nil && len(reviews) == 0 {
					awaitingReviewCount++
				}
			}
		}
	}

	if awaitingReviewCount > 0 {
		alerts = append(alerts, models.Alert{
			Type:     models.AlertTypeAwaitingReview,
			Message:  fmt.Sprintf("%d PRs waiting for first review (> 24h)", awaitingReviewCount),
			Count:    awaitingReviewCount,
			Severity: "warning",
		})
	}

	// Alert 2: Approved PRs not merged for > 12h
	awaitingMergeCount := 0
	for _, repoSlug := range repos {
		owner, repo, err := parseRepositorySlug(repoSlug)
		if err != nil {
			continue
		}

		opts := &github.PullRequestListOptions{
			State:       "open",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		prs, _, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			reviews, _, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), nil)
			if err != nil {
				continue
			}

			var latestApproval *time.Time
			for _, review := range reviews {
				if review.GetState() == "APPROVED" && review.SubmittedAt != nil {
					t := review.SubmittedAt.Time
					if latestApproval == nil || t.After(*latestApproval) {
						latestApproval = &t
					}
				}
			}

			if latestApproval != nil {
				age := now.Sub(*latestApproval)
				if age >= awaitingMergeThreshold {
					awaitingMergeCount++
				}
			}
		}
	}

	if awaitingMergeCount > 0 {
		alerts = append(alerts, models.Alert{
			Type:     models.AlertTypeAwaitingMerge,
			Message:  fmt.Sprintf("%d PRs approved but not merged (> 12h)", awaitingMergeCount),
			Count:    awaitingMergeCount,
			Severity: "warning",
		})
	}

	// Alert 3: Lead time increased significantly
	if len(currentTrend) >= 2 {
		lastWeek := currentTrend[len(currentTrend)-1]
		prevWeek := currentTrend[len(currentTrend)-2]

		if prevWeek.AverageLeadTime > 0 {
			increase := float64(lastWeek.AverageLeadTime-prevWeek.AverageLeadTime) / float64(prevWeek.AverageLeadTime)
			if increase >= leadTimeIncreaseThreshold {
				alerts = append(alerts, models.Alert{
					Type:     models.AlertTypeLeadTimeIncreased,
					Message:  fmt.Sprintf("Average lead time increased %.0f%% this week", increase*100),
					Severity: "critical",
				})
			}
		}
	}

	return models.AlertMetrics{Alerts: alerts}, nil
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

func buildTrendPoints(samples []leadTimeSample) []models.TrendPoint {
	if len(samples) == 0 {
		return nil
	}

	trendMap := make(map[string][]time.Duration)
	for _, sample := range samples {
		weekday := int(sample.mergedAt.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := sample.mergedAt.AddDate(0, 0, 1-weekday)
		sunday := monday.AddDate(0, 0, 6)

		period := fmt.Sprintf("%02d/%02d-%02d/%02d",
			monday.Month(), monday.Day(),
			sunday.Month(), sunday.Day())
		trendMap[period] = append(trendMap[period], sample.duration)
	}

	periods := make([]string, 0, len(trendMap))
	for period := range trendMap {
		periods = append(periods, period)
	}
	sort.Strings(periods)

	trend := make([]models.TrendPoint, 0, len(periods))
	for _, period := range periods {
		durations := trendMap[period]
		trend = append(trend, models.TrendPoint{
			Period:          period,
			AverageLeadTime: averageDuration(durations),
			PRCount:         len(durations),
		})
	}

	return trend
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, d := range durations {
		total += d
	}

	return time.Duration(int64(total) / int64(len(durations)))
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

func (r *MetricsRepositoryImpl) fetchPRReviewData(ctx context.Context, owner, repo string, prNumber int) (*time.Time, *time.Time, error) {
	reviews, resp, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, prNumber, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, nil, handleGitHubError(err, resp)
	}

	var firstReview *time.Time
	var firstApproval *time.Time

	for _, review := range reviews {
		if review == nil || review.SubmittedAt == nil {
			continue
		}

		if firstReview == nil || review.SubmittedAt.Before(*firstReview) {
			t := review.SubmittedAt.Time
			firstReview = &t
		}

		if review.GetState() == "APPROVED" {
			if firstApproval == nil || review.SubmittedAt.Before(*firstApproval) {
				t := review.SubmittedAt.Time
				firstApproval = &t
			}
		}
	}

	return firstReview, firstApproval, nil
}

func calculatePhaseBreakdown(phases []prPhaseData) models.ReviewPhaseMetrics {
	if len(phases) == 0 {
		return models.ReviewPhaseMetrics{}
	}

	var totalCreatedToReview time.Duration
	var totalReviewToApproval time.Duration
	var totalApprovalToMerge time.Duration
	var totalLeadTime time.Duration

	countCreatedToReview := 0
	countReviewToApproval := 0
	countApprovalToMerge := 0

	for _, p := range phases {
		if p.firstReviewAt != nil {
			totalCreatedToReview += p.firstReviewAt.Sub(p.createdAt)
			countCreatedToReview++
		}

		if p.firstReviewAt != nil && p.firstApprovalAt != nil {
			totalReviewToApproval += p.firstApprovalAt.Sub(*p.firstReviewAt)
			countReviewToApproval++
		}

		if p.firstApprovalAt != nil {
			totalApprovalToMerge += p.mergedAt.Sub(*p.firstApprovalAt)
			countApprovalToMerge++
		}

		totalLeadTime += p.mergedAt.Sub(p.createdAt)
	}

	return models.ReviewPhaseMetrics{
		CreatedToFirstReview:  safeDivide(totalCreatedToReview, countCreatedToReview),
		FirstReviewToApproval: safeDivide(totalReviewToApproval, countReviewToApproval),
		ApprovalToMerge:       safeDivide(totalApprovalToMerge, countApprovalToMerge),
		TotalLeadTime:         safeDivide(totalLeadTime, len(phases)),
	}
}

func safeDivide(total time.Duration, count int) time.Duration {
	if count == 0 {
		return 0
	}
	return time.Duration(int64(total) / int64(count))
}
