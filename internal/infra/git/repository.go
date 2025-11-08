package git

import (
	"fmt"
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
//   - git@github.com:owner/repo.git
func ParseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	var ownerRepo string

	// Handle HTTPS format: https://github.com/owner/repo
	if strings.HasPrefix(url, "https://github.com/") {
		ownerRepo = strings.TrimPrefix(url, "https://github.com/")
	} else if strings.HasPrefix(url, "http://github.com/") {
		ownerRepo = strings.TrimPrefix(url, "http://github.com/")
	} else if strings.HasPrefix(url, "git@github.com:") {
		// Handle SSH format: git@github.com:owner/repo
		ownerRepo = strings.TrimPrefix(url, "git@github.com:")
	} else {
		return "", "", fmt.Errorf("unsupported URL format: %s", url)
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
