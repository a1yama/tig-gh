package export

import (
	"sort"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/shared/metricsformat"
)

func buildOverall(metrics *models.LeadTimeMetrics) overallData {
	return overallData{
		Average: metricsformat.FormatDuration(metrics.Overall.Average),
		Median:  metricsformat.FormatDuration(metrics.Overall.Median),
		Count:   metrics.Overall.Count,
	}
}

func buildPRLeadTimes(metrics *models.LeadTimeMetrics) *prLeadTimesData {
	entries := metrics.PRLeadTimes
	if len(entries) == 0 {
		return &prLeadTimesData{}
	}

	const maxPRLeadTimesToDisplay = 20

	display := entries
	if len(display) > maxPRLeadTimesToDisplay {
		display = display[:maxPRLeadTimesToDisplay]
	}

	result := make([]prLeadTimeEntry, len(display))
	for i, e := range display {
		result[i] = prLeadTimeEntry{
			Index:      i + 1,
			Repository: e.Repository,
			Number:     e.Number,
			Title:      e.Title,
			LeadTime:   metricsformat.FormatDuration(e.LeadTime),
			MergedAt:   e.MergedAt.Format("2006-01-02"),
			HTMLURL:    e.HTMLURL,
		}
	}

	return &prLeadTimesData{Entries: result}
}

func buildReviewPhases(metrics *models.LeadTimeMetrics) *reviewPhasesData {
	pm := metrics.PhaseBreakdown
	if pm.SampleCount == 0 {
		return nil
	}

	type phaseInfo struct {
		label    string
		duration time.Duration
	}

	phases := []phaseInfo{
		{label: "PR Created → First Review", duration: pm.CreatedToFirstReview},
		{label: "First Review → Approval", duration: pm.FirstReviewToApproval},
		{label: "Approval → Merge", duration: pm.ApprovalToMerge},
	}

	longest := time.Duration(0)
	for _, p := range phases {
		if p.duration > longest {
			longest = p.duration
		}
	}

	result := make([]phaseEntry, len(phases))
	for i, p := range phases {
		result[i] = phaseEntry{
			Label:        p.label,
			Duration:     metricsformat.FormatDuration(p.duration),
			SampleCount:  pm.SampleCount,
			IsBottleneck: longest > 0 && p.duration == longest,
		}
	}

	return &reviewPhasesData{
		Phases:      result,
		TotalLabel:  metricsformat.FormatDuration(pm.TotalLeadTime),
		SampleCount: pm.SampleCount,
	}
}

func buildDayOfWeek(metrics *models.LeadTimeMetrics, cfg *models.MetricsConfig) *dayOfWeekData {
	statsByDay := metrics.ByDayOfWeek
	if statsByDay == nil {
		return &dayOfWeekData{}
	}

	displayDays := metricsformat.WeekdaysOnly
	if cfg.ShowWeekendInDayOfWeek {
		displayDays = metricsformat.WeekdayDisplayOrder
	}

	days := make([]dayOfWeekEntry, len(displayDays))
	for i, day := range displayDays {
		stats := statsByDay[day]
		days[i] = dayOfWeekEntry{
			Day:         metricsformat.ShortWeekday(day),
			MergeCount:  stats.MergeCount,
			ReviewCount: stats.ReviewCount,
		}
	}

	return &dayOfWeekData{Days: days}
}

func buildWeeklyComparison(metrics *models.LeadTimeMetrics) *weeklyComparisonData {
	c := metrics.WeeklyComparison
	return &weeklyComparisonData{
		ThisWeekReviews: c.ThisWeek.ReviewCount,
		ThisWeekMerges:  c.ThisWeek.MergeCount,
		LastWeekReviews: c.LastWeek.ReviewCount,
		LastWeekMerges:  c.LastWeek.MergeCount,
		ReviewChange:    formatChangePercent(c.ReviewChangePercent),
		MergeChange:     formatChangePercent(c.MergeChangePercent),
	}
}

func buildQualityIssues(metrics *models.LeadTimeMetrics) *qualityIssuesData {
	issues := metrics.QualityIssues.Issues
	if len(issues) == 0 {
		return &qualityIssuesData{}
	}

	var high, medium []models.PRQualityIssue
	for _, issue := range issues {
		if issue.Severity == "high" {
			high = append(high, issue)
		} else {
			medium = append(medium, issue)
		}
	}

	displayCount := len(issues)
	if displayCount > metricsformat.MaxQualityIssuesToDisplay {
		displayCount = metricsformat.MaxQualityIssuesToDisplay
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

	toEntries := func(src []models.PRQualityIssue) []qualityIssueEntry {
		result := make([]qualityIssueEntry, len(src))
		for i, issue := range src {
			result[i] = qualityIssueEntry{
				Repository: issue.Repository,
				Number:     issue.Number,
				Title:      issue.Title,
				IssueType:  issue.IssueType,
				Details:    issue.Details,
			}
		}
		return result
	}

	return &qualityIssuesData{
		HighPriority:   toEntries(high),
		MediumPriority: toEntries(medium),
	}
}

func buildStagnantPRs(metrics *models.LeadTimeMetrics) *stagnantPRsData {
	s := metrics.StagnantPRs
	if len(s.LongestWaiting) == 0 {
		return &stagnantPRsData{
			Threshold: metricsformat.FormatDuration(s.Threshold),
		}
	}

	entries := make([]stagnantPREntry, len(s.LongestWaiting))
	for i, pr := range s.LongestWaiting {
		entries[i] = stagnantPREntry{
			Index:      i + 1,
			Repository: pr.Repository,
			Number:     pr.Number,
			Title:      pr.Title,
			Age:        metricsformat.FormatDuration(pr.Age),
		}
	}

	return &stagnantPRsData{
		TotalStagnant: s.TotalStagnant,
		Threshold:     metricsformat.FormatDuration(s.Threshold),
		Entries:       entries,
	}
}

func buildRepositoryStats(metrics *models.LeadTimeMetrics) *repositoryStatsData {
	if len(metrics.ByRepository) == 0 {
		return &repositoryStatsData{}
	}

	names := make([]string, 0, len(metrics.ByRepository))
	for name := range metrics.ByRepository {
		names = append(names, name)
	}
	sort.Strings(names)

	repos := make([]repoStatEntry, len(names))
	for i, name := range names {
		stat := metrics.ByRepository[name]
		repos[i] = repoStatEntry{
			Name:    name,
			Average: metricsformat.FormatDuration(stat.Average),
			Median:  metricsformat.FormatDuration(stat.Median),
			Count:   stat.Count,
		}
	}

	return &repositoryStatsData{Repos: repos}
}
