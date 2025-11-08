package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

// IssueRepositoryImpl implements the IssueRepository interface
type IssueRepositoryImpl struct {
	client *Client
}

// NewIssueRepository creates a new IssueRepository implementation
func NewIssueRepository(client *Client) repository.IssueRepository {
	return &IssueRepositoryImpl{
		client: client,
	}
}

// List retrieves a list of issues for a repository
func (r *IssueRepositoryImpl) List(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	ghOpts := convertFromIssueOptions(opts)

	ghIssues, resp, err := r.client.client.Issues.ListByRepo(ctx, owner, repo, ghOpts)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToIssues(ghIssues), nil
}

// Get retrieves a single issue by number
func (r *IssueRepositoryImpl) Get(ctx context.Context, owner, repo string, number int) (*models.Issue, error) {
	ghIssue, resp, err := r.client.client.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToIssue(ghIssue), nil
}

// Create creates a new issue
func (r *IssueRepositoryImpl) Create(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
	if input == nil {
		return nil, fmt.Errorf("create issue input is required")
	}

	req := convertFromCreateIssueInput(input)
	ghIssue, resp, err := r.client.client.Issues.Create(ctx, owner, repo, req)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToIssue(ghIssue), nil
}

// Update updates an existing issue
func (r *IssueRepositoryImpl) Update(ctx context.Context, owner, repo string, number int, input *models.UpdateIssueInput) (*models.Issue, error) {
	if input == nil {
		return nil, fmt.Errorf("update issue input is required")
	}

	req := convertFromUpdateIssueInput(input)
	ghIssue, resp, err := r.client.client.Issues.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToIssue(ghIssue), nil
}

// Close closes an issue
func (r *IssueRepositoryImpl) Close(ctx context.Context, owner, repo string, number int) error {
	state := "closed"
	req := &github.IssueRequest{
		State: &state,
	}

	_, resp, err := r.client.client.Issues.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// Reopen reopens a closed issue
func (r *IssueRepositoryImpl) Reopen(ctx context.Context, owner, repo string, number int) error {
	state := "open"
	req := &github.IssueRequest{
		State: &state,
	}

	_, resp, err := r.client.client.Issues.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// Lock locks an issue
func (r *IssueRepositoryImpl) Lock(ctx context.Context, owner, repo string, number int) error {
	resp, err := r.client.client.Issues.Lock(ctx, owner, repo, number, nil)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// Unlock unlocks an issue
func (r *IssueRepositoryImpl) Unlock(ctx context.Context, owner, repo string, number int) error {
	resp, err := r.client.client.Issues.Unlock(ctx, owner, repo, number)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// handleGitHubError handles GitHub API errors and converts them to more descriptive errors
func handleGitHubError(err error, resp *github.Response) error {
	if err == nil {
		return nil
	}

	// If no response, return the original error
	if resp == nil {
		return fmt.Errorf("github api error: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("resource not found (404): %w", err)
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized - check your token (401): %w", err)
	case http.StatusForbidden:
		// Check if it's a rate limit error
		if resp.Rate.Remaining == 0 {
			return fmt.Errorf("rate limit exceeded, resets at %v: %w", resp.Rate.Reset, err)
		}
		return fmt.Errorf("forbidden - insufficient permissions (403): %w", err)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("validation failed (422): %w", err)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("github server error (%d): %w", resp.StatusCode, err)
	default:
		return fmt.Errorf("github api error (status %d): %w", resp.StatusCode, err)
	}
}
