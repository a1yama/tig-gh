package views

import (
	"testing"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

func TestPrDisplayNumber_FromStruct(t *testing.T) {
	pr := &models.PullRequest{Number: 42}
	if n, ok := prDisplayNumber(pr); !ok || n != 42 {
		t.Fatalf("expected 42,true got %d,%v", n, ok)
	}
}

func TestPrDisplayNumber_FromURL(t *testing.T) {
	pr := &models.PullRequest{HTMLURL: "https://github.com/org/repo/pull/4631"}
	if n, ok := prDisplayNumber(pr); !ok || n != 4631 {
		t.Fatalf("expected 4631,true got %d,%v", n, ok)
	}
}

func TestFormatPRTitle_Fallback(t *testing.T) {
	pr := &models.PullRequest{HTMLURL: "https://example.com/org/repo/pull/5"}
	if title := formatPRTitle(pr); title != "PR #5" {
		t.Fatalf("expected 'PR #5', got %q", title)
	}

	if title := formatPRTitle(&models.PullRequest{}); title != "Pull Request" {
		t.Fatalf("expected fallback title, got %q", title)
	}
}

func TestFormatAuthorHandle(t *testing.T) {
	if got := formatAuthorHandle(models.User{Login: "alice"}); got != "@alice" {
		t.Fatalf("expected @alice, got %s", got)
	}
	if got := formatAuthorHandle(models.User{Name: "Bob"}); got != "@Bob" {
		t.Fatalf("expected fallback from name, got %s", got)
	}
	if got := formatAuthorHandle(models.User{}); got != "@unknown" {
		t.Fatalf("expected @unknown fallback, got %s", got)
	}
}

func TestFormatBranchName(t *testing.T) {
	if got := formatBranchName(models.Branch{Name: "main"}); got != "main" {
		t.Fatalf("expected main, got %s", got)
	}
	if got := formatBranchName(models.Branch{}); got != "?" {
		t.Fatalf("expected '?' fallback, got %s", got)
	}
}
