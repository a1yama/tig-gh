package github

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

const (
	stagnantPRThreshold = 72 * time.Hour // 3 days
	reviewWorkerCount   = 12             // concurrent review fetchers
	repoWorkerCount     = 8              // concurrent repository fetchers
)

type leadTimeSample struct {
	duration      time.Duration
	mergedAt      time.Time
	firstReviewAt *time.Time
	approvedAt    *time.Time
}

// MetricsRepositoryImpl は MetricsRepository を実装する
type MetricsRepositoryImpl struct {
	client *Client
}

type repoFetchTask struct {
	slug  string
	owner string
	name  string
}

type repoFetchResult struct {
	slug    string
	samples []leadTimeSample
	err     error
}

type stagnantFetchResult struct {
	repo string
	prs  []models.StagnantPRInfo
	err  error
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
		Overall:                    models.LeadTimeStat{},
		ByRepository:               make(map[string]models.LeadTimeStat),
		Trend:                      []models.TrendPoint{},
		ByDayOfWeek:                make(map[time.Weekday]models.DayOfWeekStats),
		ByRepositoryDayOfWeek:      make(map[string]map[time.Weekday]models.DayOfWeekStats),
		ByRepositoryWeekly:         make(map[string]models.WeeklyComparison),
		ByRepositoryPhaseBreakdown: make(map[string]models.ReviewPhaseMetrics),
	}

	if len(repos) == 0 {
		return result, nil
	}

	repoSamples := make(map[string][]leadTimeSample)
	var errs []error

	totalRepos := len(repos)
	processedRepos := 0

	reportProgress := func(repo string) {
		if progressFn == nil || totalRepos == 0 {
			return
		}
		progressFn(models.MetricsProgress{
			TotalRepos:     totalRepos,
			ProcessedRepos: processedRepos,
			CurrentRepo:    repo,
		})
	}

	if progressFn != nil && totalRepos > 0 {
		progressFn(models.MetricsProgress{
			TotalRepos:     totalRepos,
			ProcessedRepos: 0,
			CurrentRepo:    "",
		})
	}

	var tasks []repoFetchTask

	for _, repoFull := range repos {
		repoFull = strings.TrimSpace(repoFull)
		if repoFull == "" {
			continue
		}

		owner, name, err := parseRepositorySlug(repoFull)
		if err != nil {
			errs = append(errs, err)
			processedRepos++
			reportProgress(repoFull)
			continue
		}

		tasks = append(tasks, repoFetchTask{
			slug:  repoFull,
			owner: owner,
			name:  name,
		})
	}

	if len(tasks) > 0 {
		workerCount := repoWorkerCount
		if len(tasks) < workerCount {
			workerCount = len(tasks)
		}

		jobs := make(chan repoFetchTask)
		results := make(chan repoFetchResult)
		var workers sync.WaitGroup

		for i := 0; i < workerCount; i++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for task := range jobs {
					samples, fetchErr := r.fetchLeadTimeSamples(ctx, task.owner, task.name, since)
					results <- repoFetchResult{
						slug:    task.slug,
						samples: samples,
						err:     fetchErr,
					}
				}
			}()
		}

		go func() {
			for _, task := range tasks {
				jobs <- task
			}
			close(jobs)
		}()

		go func() {
			workers.Wait()
			close(results)
		}()

		for result := range results {
			if result.err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", result.slug, result.err))
			} else {
				repoSamples[result.slug] = result.samples
			}

			processedRepos++
			reportProgress(result.slug)
		}
	}

	if progressFn != nil && totalRepos > 0 {
		progressFn(models.MetricsProgress{
			TotalRepos:     totalRepos,
			ProcessedRepos: totalRepos,
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

		result.ByRepositoryPhaseBreakdown[slug] = calculatePhaseBreakdown(samples)

		overallSamples = append(overallSamples, samples...)
	}

	allDurations := samplesToDurations(overallSamples)

	result.Overall = calculateLeadTimeStat(allDurations)

	result.ByDayOfWeek = aggregateByDayOfWeek(overallSamples)

	result.WeeklyComparison = calculateWeeklyComparison(overallSamples, currentTime)

	result.PhaseBreakdown = calculatePhaseBreakdown(overallSamples)

	qualityIssues, qualityErr := r.analyzeOpenPRQuality(ctx, repos)
	if qualityErr != nil {
		fmt.Printf("failed to analyze PR quality: %v\n", qualityErr)
	} else {
		result.QualityIssues = qualityIssues
	}

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
				firstReview, approval := r.fetchSampleFirstReview(ctx, owner, repo, req.number)
				samples[req.sampleIndex].firstReviewAt = firstReview
				samples[req.sampleIndex].approvedAt = approval
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

func (r *MetricsRepositoryImpl) fetchSampleFirstReview(ctx context.Context, owner, repo string, number int) (*time.Time, *time.Time) {
	firstReview, approved, err := r.fetchReviewTimestamps(ctx, owner, repo, number)
	if err != nil {
		fmt.Printf("failed to fetch reviews for %s/%s#%d: %v\n", owner, repo, number, err)
		return nil, nil
	}
	return firstReview, approved
}

func (r *MetricsRepositoryImpl) fetchReviewTimestamps(ctx context.Context, owner, repo string, number int) (*time.Time, *time.Time, error) {
	opts := &github.ListOptions{PerPage: 100}
	var firstReview time.Time
	firstFound := false
	var approval time.Time
	approvalFound := false

	for {
		reviews, resp, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, nil, handleGitHubError(err, resp)
		}

		for _, review := range reviews {
			if review == nil || review.SubmittedAt == nil {
				continue
			}
			submitted := review.SubmittedAt.Time
			if !firstFound || submitted.Before(firstReview) {
				firstReview = submitted
				firstFound = true
			}

			if strings.EqualFold(review.GetState(), "APPROVED") {
				if !approvalFound || submitted.Before(approval) {
					approval = submitted
					approvalFound = true
				}
			}
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var firstPtr *time.Time
	if firstFound {
		firstCopy := firstReview
		firstPtr = &firstCopy
	}

	var approvalPtr *time.Time
	if approvalFound {
		approvalCopy := approval
		approvalPtr = &approvalCopy
	}

	return firstPtr, approvalPtr, nil
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

func calculatePhaseBreakdown(samples []leadTimeSample) models.ReviewPhaseMetrics {
	if len(samples) == 0 {
		return models.ReviewPhaseMetrics{}
	}

	var (
		totalCreatedToFirst  time.Duration
		totalFirstToApproval time.Duration
		totalApprovalToMerge time.Duration
		totalLeadTime        time.Duration
		count                int64
	)

	for _, sample := range samples {
		if sample.firstReviewAt == nil || sample.approvedAt == nil {
			continue
		}

		createdAt := sample.mergedAt.Add(-sample.duration)
		firstReviewAt := *sample.firstReviewAt
		approvedAt := *sample.approvedAt

		// 期待される順序でタイムスタンプが揃っていない場合は除外
		if firstReviewAt.Before(createdAt) || approvedAt.Before(firstReviewAt) || sample.mergedAt.Before(approvedAt) {
			continue
		}

		totalCreatedToFirst += firstReviewAt.Sub(createdAt)
		totalFirstToApproval += approvedAt.Sub(firstReviewAt)
		totalApprovalToMerge += sample.mergedAt.Sub(approvedAt)
		totalLeadTime += sample.duration
		count++
	}

	if count == 0 {
		return models.ReviewPhaseMetrics{}
	}

	return models.ReviewPhaseMetrics{
		CreatedToFirstReview:  time.Duration(int64(totalCreatedToFirst) / count),
		FirstReviewToApproval: time.Duration(int64(totalFirstToApproval) / count),
		ApprovalToMerge:       time.Duration(int64(totalApprovalToMerge) / count),
		TotalLeadTime:         time.Duration(int64(totalLeadTime) / count),
		SampleCount:           int(count),
	}
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

type scoredQualityIssue struct {
	issue models.PRQualityIssue
	score int
}

func (r *MetricsRepositoryImpl) analyzeOpenPRQuality(ctx context.Context, repos []string) (models.PRQualityIssues, error) {
	var tasks []repoFetchTask

	for _, repoSlug := range repos {
		repoSlug = strings.TrimSpace(repoSlug)
		if repoSlug == "" {
			continue
		}

		owner, repo, err := parseRepositorySlug(repoSlug)
		if err != nil {
			continue
		}

		tasks = append(tasks, repoFetchTask{
			slug:  repoSlug,
			owner: owner,
			name:  repo,
		})
	}

	if len(tasks) == 0 {
		return models.PRQualityIssues{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerCount := repoWorkerCount
	if len(tasks) < workerCount {
		workerCount = len(tasks)
	}

	var (
		collected []scoredQualityIssue
		colMu     sync.Mutex
		errOnce   sync.Once
		firstErr  error
	)

	jobs := make(chan repoFetchTask)
	var workers sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for task := range jobs {
				issues, err := r.fetchPRQualityIssuesForRepo(ctx, task.owner, task.name, task.slug)
				if err != nil {
					errOnce.Do(func() {
						firstErr = err
						cancel()
					})
					return
				}
				if len(issues) == 0 {
					continue
				}
				colMu.Lock()
				collected = append(collected, issues...)
				colMu.Unlock()
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case jobs <- task:
			}
		}
	}()

	workers.Wait()

	if firstErr != nil {
		return models.PRQualityIssues{}, firstErr
	}

	if len(collected) == 0 {
		return models.PRQualityIssues{}, nil
	}

	sort.Slice(collected, func(i, j int) bool {
		si := severityWeight(collected[i].issue.Severity)
		sj := severityWeight(collected[j].issue.Severity)
		if si != sj {
			return si < sj
		}
		if collected[i].score != collected[j].score {
			return collected[i].score > collected[j].score
		}
		if collected[i].issue.Repository != collected[j].issue.Repository {
			return collected[i].issue.Repository < collected[j].issue.Repository
		}
		return collected[i].issue.Number < collected[j].issue.Number
	})

	limit := 10
	if len(collected) < limit {
		limit = len(collected)
	}

	issues := make([]models.PRQualityIssue, 0, limit)
	for i := 0; i < limit; i++ {
		issues = append(issues, collected[i].issue)
	}

	return models.PRQualityIssues{Issues: issues}, nil
}

func (r *MetricsRepositoryImpl) fetchPRQualityIssuesForRepo(ctx context.Context, owner, repo, slug string) ([]scoredQualityIssue, error) {
	opts := &github.PullRequestListOptions{
		State:     "open",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	var issues []scoredQualityIssue

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		prs, resp, err := r.client.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, handleGitHubError(err, resp)
		}

		for _, pr := range prs {
			if pr == nil {
				continue
			}
			issues = append(issues, collectQualityIssuesForPR(slug, pr)...)
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return issues, nil
}

func collectQualityIssuesForPR(repoSlug string, pr *github.PullRequest) []scoredQualityIssue {
	lines := pr.GetAdditions() + pr.GetDeletions()
	files := pr.GetChangedFiles()
	commits := pr.GetCommits()
	body := strings.TrimSpace(pr.GetBody())
	bodyLength := utf8.RuneCountInString(body)
	details := formatQualityDetails(lines, files, commits)

	var issues []scoredQualityIssue

	addIssue := func(issueType, severity, reason, recommendation string) {
		issue := models.PRQualityIssue{
			Repository:     repoSlug,
			Number:         pr.GetNumber(),
			Title:          pr.GetTitle(),
			IssueType:      issueType,
			Severity:       severity,
			Reason:         reason,
			Recommendation: recommendation,
			Details:        details,
		}
		score := calculateQualityImpact(issueType, lines, files, commits)
		collectedIssue := scoredQualityIssue{
			issue: issue,
			score: score,
		}
		issues = append(issues, collectedIssue)
	}

	if lines >= 500 {
		addIssue(
			"large_pr",
			"high",
			"レビューに時間がかかり、バグが見落とされやすい",
			"機能ごとに分割し、200-400行に抑える",
		)
	}

	if body == "" {
		addIssue(
			"no_description",
			"high",
			"レビュアーが変更意図を理解できない",
			"「何を」「なぜ」「どうテストしたか」を記載",
		)
	} else if bodyLength > 0 && bodyLength < 50 {
		addIssue(
			"short_description",
			"medium",
			"テンプレートのままの可能性",
			"変更の背景と影響範囲を追記",
		)
	}

	if commits >= 15 {
		addIssue(
			"many_commits",
			"medium",
			"レビュー時の変更履歴が追いづらい",
			"関連するコミットをsquashして整理",
		)
	}

	if commits == 1 && lines >= 500 {
		addIssue(
			"large_single_commit",
			"medium",
			"レビュー時に変更の流れが分からない",
			"論理的な単位でコミットを分ける",
		)
	}

	return issues
}

func severityWeight(severity string) int {
	if strings.EqualFold(severity, "high") {
		return 0
	}
	return 1
}

func calculateQualityImpact(issueType string, lines, files, commits int) int {
	impact := lines
	if impact == 0 {
		impact = files * 25
	}
	if impact == 0 {
		impact = commits * 20
	}

	switch issueType {
	case "many_commits":
		impact += commits * 20
	case "large_single_commit":
		impact += 200
	case "no_description", "short_description":
		impact += (files + commits) * 10
	}

	return impact
}

func formatQualityDetails(lines, files, commits int) string {
	details := fmt.Sprintf("%d lines, %d files", lines, files)
	if commits > 0 {
		details = fmt.Sprintf("%s, %d commits", details, commits)
	}
	return details
}

func (r *MetricsRepositoryImpl) fetchStagnantPRMetrics(ctx context.Context, repos []string, now time.Time) (models.StagnantPRMetrics, error) {
	var allStagnantPRs []models.StagnantPRInfo

	var tasks []repoFetchTask
	for _, repoSlug := range repos {
		repoSlug = strings.TrimSpace(repoSlug)
		if repoSlug == "" {
			continue
		}

		owner, repo, err := parseRepositorySlug(repoSlug)
		if err != nil {
			continue
		}

		tasks = append(tasks, repoFetchTask{
			slug:  repoSlug,
			owner: owner,
			name:  repo,
		})
	}

	if len(tasks) == 0 {
		return models.StagnantPRMetrics{
			Threshold: stagnantPRThreshold,
		}, nil
	}

	workerCount := repoWorkerCount
	if len(tasks) < workerCount {
		workerCount = len(tasks)
	}

	jobs := make(chan repoFetchTask)
	results := make(chan stagnantFetchResult)
	var workers sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for task := range jobs {
				opts := &github.PullRequestListOptions{
					State:       "open",
					Sort:        "created",
					Direction:   "asc",
					ListOptions: github.ListOptions{PerPage: 100},
				}

				prs, resp, err := r.client.client.PullRequests.List(ctx, task.owner, task.name, opts)
				if err != nil {
					results <- stagnantFetchResult{
						repo: task.slug,
						err:  handleGitHubError(err, resp),
					}
					continue
				}

				var stagnant []models.StagnantPRInfo
				for _, pr := range prs {
					if pr == nil || pr.CreatedAt == nil {
						continue
					}

					age := now.Sub(pr.CreatedAt.Time)
					if age >= stagnantPRThreshold {
						stagnant = append(stagnant, models.StagnantPRInfo{
							Repository: task.slug,
							Number:     pr.GetNumber(),
							Title:      pr.GetTitle(),
							Age:        age,
						})
					}
				}

				results <- stagnantFetchResult{
					repo: task.slug,
					prs:  stagnant,
				}
			}
		}()
	}

	go func() {
		for _, task := range tasks {
			jobs <- task
		}
		close(jobs)
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			fmt.Printf("failed to fetch stagnant PR metrics for %s: %v\n", result.repo, result.err)
			continue
		}

		if len(result.prs) > 0 {
			allStagnantPRs = append(allStagnantPRs, result.prs...)
		}
	}

	if len(allStagnantPRs) == 0 {
		return models.StagnantPRMetrics{
			Threshold: stagnantPRThreshold,
		}, nil
	}

	var totalAgeSeconds float64
	for i := range allStagnantPRs {
		totalAgeSeconds += allStagnantPRs[i].Age.Seconds()
	}

	sort.Slice(allStagnantPRs, func(i, j int) bool {
		return allStagnantPRs[i].Age > allStagnantPRs[j].Age
	})

	topCount := 10
	if len(allStagnantPRs) < topCount {
		topCount = len(allStagnantPRs)
	}
	longestWaiting := append([]models.StagnantPRInfo(nil), allStagnantPRs[:topCount]...)

	averageAge := time.Duration(0)
	if len(allStagnantPRs) > 0 {
		averageSeconds := totalAgeSeconds / float64(len(allStagnantPRs))
		averageAge = time.Duration(averageSeconds * float64(time.Second))
	}

	return models.StagnantPRMetrics{
		Threshold:      stagnantPRThreshold,
		TotalStagnant:  len(allStagnantPRs),
		AverageAge:     averageAge,
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
