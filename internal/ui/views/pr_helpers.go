package views

import (
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// prDisplayNumber returns the best-effort PR number and a boolean indicating availability.
func prDisplayNumber(pr *models.PullRequest) (int, bool) {
	if pr == nil {
		return 0, false
	}
	if pr.Number > 0 {
		return pr.Number, true
	}
	if n := extractNumberFromURL(pr.HTMLURL); n > 0 {
		return n, true
	}
	return 0, false
}

func extractNumberFromURL(raw string) int {
	if raw == "" {
		return 0
	}
	u, err := url.Parse(raw)
	if err != nil {
		return 0
	}
	segments := strings.Split(strings.Trim(path.Clean(u.Path), "/"), "/")
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if seg == "pull" {
			if i+1 < len(segments) {
				if n, err := strconv.Atoi(segments[i+1]); err == nil {
					return n
				}
			}
		} else if seg != "" {
			if n, err := strconv.Atoi(seg); err == nil {
				return n
			}
		}
	}
	return 0
}

// formatPRTitle returns a human-readable header label with PR number fallback.
func formatPRTitle(pr *models.PullRequest) string {
	if n, ok := prDisplayNumber(pr); ok {
		return "PR #" + strconv.Itoa(n)
	}
	return "Pull Request"
}

func ensurePRNumber(pr *models.PullRequest) {
	if pr == nil {
		return
	}
	if pr.Number > 0 {
		return
	}
	if n := extractNumberFromURL(pr.HTMLURL); n > 0 {
		pr.Number = n
	}
}

func formatAuthorHandle(user models.User) string {
	if user.Login != "" {
		return "@" + user.Login
	}
	if user.Name != "" {
		return "@" + user.Name
	}
	return "@unknown"
}

func formatBranchName(branch models.Branch) string {
	if branch.Name != "" {
		return branch.Name
	}
	return "?"
}
