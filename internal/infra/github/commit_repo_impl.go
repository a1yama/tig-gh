package github

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CommitRepositoryImpl implements the CommitRepository interface
type CommitRepositoryImpl struct {
	client *Client
}

// NewCommitRepository creates a new CommitRepository implementation
func NewCommitRepository(client *Client) repository.CommitRepository {
	return &CommitRepositoryImpl{
		client: client,
	}
}

// List retrieves a list of commits for a repository
func (r *CommitRepositoryImpl) List(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
	ghOpts := convertFromCommitOptions(opts)

	ghCommits, resp, err := r.client.client.Repositories.ListCommits(ctx, owner, repo, ghOpts)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToCommits(ghCommits), nil
}

// Get retrieves a single commit by SHA
func (r *CommitRepositoryImpl) Get(ctx context.Context, owner, repo, sha string) (*models.Commit, error) {
	ghCommit, resp, err := r.client.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToCommitDetail(ghCommit), nil
}

// Compare compares two commits
func (r *CommitRepositoryImpl) Compare(ctx context.Context, owner, repo, base, head string) (*models.Comparison, error) {
	ghComparison, resp, err := r.client.client.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToComparison(ghComparison), nil
}

// ListBranches retrieves a list of branches for a repository
func (r *CommitRepositoryImpl) ListBranches(ctx context.Context, owner, repo string) ([]*models.Branch, error) {
	ghBranches, resp, err := r.client.client.Repositories.ListBranches(ctx, owner, repo, nil)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToBranches(ghBranches), nil
}

// GetBranch retrieves a single branch by name
func (r *CommitRepositoryImpl) GetBranch(ctx context.Context, owner, repo, branch string) (*models.Branch, error) {
	ghBranch, resp, err := r.client.client.Repositories.GetBranch(ctx, owner, repo, branch, 0)
	if err != nil {
		return nil, handleGitHubError(err, resp)
	}

	return convertToBranch(ghBranch), nil
}
