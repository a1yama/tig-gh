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
