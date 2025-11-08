package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

// SearchRepositoryImpl implements the SearchRepository interface
type SearchRepositoryImpl struct {
	client *Client
}

// NewSearchRepository creates a new SearchRepository implementation
func NewSearchRepository(client *Client) repository.SearchRepository {
	return &SearchRepositoryImpl{
		client: client,
	}
}

// Search searches for issues and pull requests based on the given options
func (r *SearchRepositoryImpl) Search(ctx context.Context, owner, repo string, opts *models.SearchOptions) (*models.SearchResults, error) {
	if opts == nil {
		opts = &models.SearchOptions{
			Type:      models.SearchTypeBoth,
			State:     models.IssueStateOpen,
			Sort:      models.SearchSortUpdated,
			Direction: models.SortDirectionDesc,
			PerPage:   30,
			Page:      1,
		}
	}

	// Build search query
	query := buildSearchQuery(owner, repo, opts)

	// Prepare search options
	searchOpts := &github.SearchOptions{
		Sort:  string(opts.Sort),
		Order: string(opts.Direction),
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.PerPage,
		},
	}

	// Execute search
	result, resp, err := r.client.client.Search.Issues(ctx, query, searchOpts)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	// Convert results
	searchResults := &models.SearchResults{
		TotalCount:        result.GetTotal(),
		IncompleteResults: result.GetIncompleteResults(),
		Items:             make([]models.SearchResult, 0, len(result.Issues)),
	}

	for _, issue := range result.Issues {
		searchResult := convertSearchIssue(issue)
		searchResults.Items = append(searchResults.Items, searchResult)
	}

	return searchResults, nil
}

// buildSearchQuery builds a GitHub search query string from options
func buildSearchQuery(owner, repo string, opts *models.SearchOptions) string {
	parts := []string{
		fmt.Sprintf("repo:%s/%s", owner, repo),
	}

	// Add search query if provided
	if opts.Query != "" {
		parts = append(parts, opts.Query)
	}

	// Add type filter
	switch opts.Type {
	case models.SearchTypeIssue:
		parts = append(parts, "is:issue")
	case models.SearchTypePR:
		parts = append(parts, "is:pr")
	// models.SearchTypeBoth doesn't need a filter
	}

	// Add state filter
	switch opts.State {
	case models.IssueStateOpen:
		parts = append(parts, "is:open")
	case models.IssueStateClosed:
		parts = append(parts, "is:closed")
	// models.IssueStateAll doesn't need a filter
	}

	// Add author filter
	if opts.Author != "" {
		parts = append(parts, fmt.Sprintf("author:%s", opts.Author))
	}

	// Add label filters
	for _, label := range opts.Labels {
		parts = append(parts, fmt.Sprintf("label:\"%s\"", label))
	}

	return strings.Join(parts, " ")
}

// convertSearchIssue converts a GitHub issue from search results to a SearchResult
func convertSearchIssue(ghIssue *github.Issue) models.SearchResult {
	// Check if it's a pull request by looking for the PullRequestLinks field
	if ghIssue.PullRequestLinks != nil {
		// It's a pull request
		pr := convertIssueToPR(ghIssue)
		return models.SearchResult{
			Type:        models.SearchTypePR,
			PullRequest: pr,
		}
	}

	// It's an issue
	issue := convertToIssue(ghIssue)
	return models.SearchResult{
		Type:  models.SearchTypeIssue,
		Issue: issue,
	}
}

// convertIssueToPR converts a GitHub issue (from search) to a PullRequest
// Note: Search API returns issues, but we can identify PRs and convert them
func convertIssueToPR(ghIssue *github.Issue) *models.PullRequest {
	if ghIssue == nil {
		return nil
	}

	pr := &models.PullRequest{
		ID:       ghIssue.GetID(),
		Number:   ghIssue.GetNumber(),
		Title:    ghIssue.GetTitle(),
		Body:     ghIssue.GetBody(),
		HTMLURL:  ghIssue.GetHTMLURL(),
		Comments: ghIssue.GetComments(),
	}

	// Convert state
	state := ghIssue.GetState()
	switch state {
	case "open":
		pr.State = models.PRStateOpen
	case "closed":
		pr.State = models.PRStateClosed
	default:
		pr.State = models.PRStateOpen
	}

	// Convert author
	if ghIssue.User != nil {
		pr.Author = convertToUser(ghIssue.User)
	}

	// Convert labels
	if len(ghIssue.Labels) > 0 {
		pr.Labels = make([]models.Label, 0, len(ghIssue.Labels))
		for _, label := range ghIssue.Labels {
			pr.Labels = append(pr.Labels, convertToLabel(label))
		}
	}

	// Convert assignees
	if len(ghIssue.Assignees) > 0 {
		pr.Assignees = make([]models.User, 0, len(ghIssue.Assignees))
		for _, assignee := range ghIssue.Assignees {
			pr.Assignees = append(pr.Assignees, convertToUser(assignee))
		}
	}

	// Convert milestone
	if ghIssue.Milestone != nil {
		pr.Milestone = convertToMilestone(ghIssue.Milestone)
	}

	// Convert timestamps
	if ghIssue.CreatedAt != nil {
		pr.CreatedAt = ghIssue.CreatedAt.Time
	}

	if ghIssue.UpdatedAt != nil {
		pr.UpdatedAt = ghIssue.UpdatedAt.Time
	}

	if ghIssue.ClosedAt != nil {
		closedAt := ghIssue.ClosedAt.Time
		pr.ClosedAt = &closedAt
	}

	return pr
}
