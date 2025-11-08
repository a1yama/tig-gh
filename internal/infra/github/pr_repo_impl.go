package github

import (
	"context"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/google/go-github/v57/github"
)

// PullRequestRepositoryImpl implements the PullRequestRepository interface
type PullRequestRepositoryImpl struct {
	client *Client
}

// NewPullRequestRepository creates a new PullRequestRepository implementation
func NewPullRequestRepository(client *Client) repository.PullRequestRepository {
	return &PullRequestRepositoryImpl{
		client: client,
	}
}

// List retrieves a list of pull requests for a repository
func (r *PullRequestRepositoryImpl) List(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
	ghOpts := convertFromPROptions(opts)

	ghPRs, resp, err := r.client.client.PullRequests.List(ctx, owner, repo, ghOpts)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToPullRequests(ghPRs), nil
}

// Get retrieves a single pull request by number
func (r *PullRequestRepositoryImpl) Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error) {
	ghPR, resp, err := r.client.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToPullRequest(ghPR), nil
}

// Create creates a new pull request
func (r *PullRequestRepositoryImpl) Create(ctx context.Context, owner, repo string, input *models.CreatePRInput) (*models.PullRequest, error) {
	if input == nil {
		return nil, fmt.Errorf("create pull request input is required")
	}

	req := convertFromCreatePRInput(input)
	ghPR, resp, err := r.client.client.PullRequests.Create(ctx, owner, repo, req)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToPullRequest(ghPR), nil
}

// Update updates an existing pull request
func (r *PullRequestRepositoryImpl) Update(ctx context.Context, owner, repo string, number int, input *models.UpdatePRInput) (*models.PullRequest, error) {
	if input == nil {
		return nil, fmt.Errorf("update pull request input is required")
	}

	req := convertFromUpdatePRInput(input)
	ghPR, resp, err := r.client.client.PullRequests.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToPullRequest(ghPR), nil
}

// Merge merges a pull request
func (r *PullRequestRepositoryImpl) Merge(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
	ghOpts := convertFromMergeOptions(opts)
	commitMessage := ""

	if opts != nil && opts.CommitMessage != "" {
		commitMessage = opts.CommitMessage
		if opts.CommitTitle != "" {
			commitMessage = opts.CommitTitle + "\n\n" + opts.CommitMessage
		}
	}

	_, resp, err := r.client.client.PullRequests.Merge(ctx, owner, repo, number, commitMessage, ghOpts)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// Close closes a pull request without merging
func (r *PullRequestRepositoryImpl) Close(ctx context.Context, owner, repo string, number int) error {
	state := "closed"
	req := &github.PullRequest{
		State: &state,
	}

	_, resp, err := r.client.client.PullRequests.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// Reopen reopens a closed pull request
func (r *PullRequestRepositoryImpl) Reopen(ctx context.Context, owner, repo string, number int) error {
	state := "open"
	req := &github.PullRequest{
		State: &state,
	}

	_, resp, err := r.client.client.PullRequests.Edit(ctx, owner, repo, number, req)
	if err != nil {
		return handleGitHubError(err, resp)
	}

	return nil
}

// GetDiff retrieves the diff for a pull request
func (r *PullRequestRepositoryImpl) GetDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	// GitHub APIのRawエンドポイントを使用してdiffを取得
	diff, resp, err := r.client.client.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return "", handleGitHubError(err, resp)
	}

	return diff, nil
}

// IsMergeable checks if a pull request is mergeable
func (r *PullRequestRepositoryImpl) IsMergeable(ctx context.Context, owner, repo string, number int) (bool, error) {
	ghPR, resp, err := r.client.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return false, handleGitHubError(err, resp)
	}

	// Mergeableフィールドがnilの場合はfalseを返す
	if ghPR.Mergeable == nil {
		return false, nil
	}

	return *ghPR.Mergeable, nil
}

// ListReviews retrieves reviews for a pull request
func (r *PullRequestRepositoryImpl) ListReviews(ctx context.Context, owner, repo string, number int) ([]*models.Review, error) {
	ghReviews, resp, err := r.client.client.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToReviews(ghReviews), nil
}
