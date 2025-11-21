package github

import (
	"testing"
	"time"
)

func TestCalculateLeadTimeStat(t *testing.T) {
	durations := []time.Duration{
		1 * time.Hour,
		4 * time.Hour,
		2 * time.Hour,
		3 * time.Hour,
	}

	stat := calculateLeadTimeStat(durations)

	if stat.Count != 4 {
		t.Fatalf("unexpected count %d", stat.Count)
	}
	if stat.Average != 2*time.Hour+30*time.Minute {
		t.Fatalf("unexpected average %v", stat.Average)
	}
	if stat.Median != 2*time.Hour+30*time.Minute {
		t.Fatalf("unexpected median %v", stat.Median)
	}
}

func TestBuildTrendPoints(t *testing.T) {
	base := time.Date(2024, time.November, 18, 12, 0, 0, 0, time.UTC)
	samples := []leadTimeSample{
		{duration: 2 * time.Hour, mergedAt: base},
		{duration: 4 * time.Hour, mergedAt: base.Add(24 * time.Hour)},
		{duration: 6 * time.Hour, mergedAt: base.Add(7 * 24 * time.Hour)},
	}

	trend := buildTrendPoints(samples)

	if len(trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(trend))
	}
	if trend[0].Period != "11/18-11/24" {
		t.Fatalf("unexpected first period %s", trend[0].Period)
	}
	if trend[0].PRCount != 2 || trend[0].AverageLeadTime != 3*time.Hour {
		t.Fatalf("unexpected first trend point %+v", trend[0])
	}
	if trend[1].Period != "11/25-12/01" {
		t.Fatalf("unexpected second period %s", trend[1].Period)
	}
	if trend[1].PRCount != 1 || trend[1].AverageLeadTime != 6*time.Hour {
		t.Fatalf("unexpected second trend point %+v", trend[1])
	}
}

func TestParseRepositorySlug(t *testing.T) {
	owner, repo, err := parseRepositorySlug("owner/repo")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if owner != "owner" || repo != "repo" {
		t.Fatalf("unexpected parse result %s/%s", owner, repo)
	}

	if _, _, err := parseRepositorySlug("invalid"); err == nil {
		t.Fatal("expected error for invalid slug")
	}
}
