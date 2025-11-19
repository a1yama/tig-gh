package git

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// GetCurrentRepository returns the owner and repository name from the current Git repository
func GetCurrentRepository() (owner, repo string, err error) {
	// Get the remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	return ParseGitHubURL(remoteURL)
}

// ParseGitHubURL parses a GitHub URL and returns the owner and repository name
// Supports both HTTPS and SSH formats:
//   - https://github.com/owner/repo.git
//   - https://token:x-oauth-basic@github.com/owner/repo.git (with auth)
//   - git@github.com:owner/repo.git
func ParseGitHubURL(urlStr string) (owner, repo string, err error) {
	urlStr = strings.TrimSpace(urlStr)

	// Remove .git suffix if present
	urlStr = strings.TrimSuffix(urlStr, ".git")

	var ownerRepo string

	// Handle HTTP/HTTPS format (with or without authentication)
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse URL: %w", err)
		}

		// Check if it's GitHub
		if u.Host != "github.com" {
			return "", "", fmt.Errorf("not a GitHub URL: %s", urlStr)
		}

		// Extract owner/repo from path (authentication info is automatically stripped by url.Parse)
		ownerRepo = strings.TrimPrefix(u.Path, "/")
	} else if strings.HasPrefix(urlStr, "git@github.com:") {
		// Handle SSH format: git@github.com:owner/repo
		ownerRepo = strings.TrimPrefix(urlStr, "git@github.com:")
	} else {
		return "", "", fmt.Errorf("unsupported URL format: %s", urlStr)
	}

	// Split owner/repo
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repository format: %s", ownerRepo)
	}

	owner = parts[0]
	repo = parts[1]

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("empty owner or repository name")
	}

	return owner, repo, nil
}

// IsGitRepository checks if the current directory is a Git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
