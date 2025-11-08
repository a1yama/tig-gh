package repository

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// PullRequestRepository defines the interface for pull request operations
type PullRequestRepository interface {
	// List retrieves a list of pull requests for a repository
	List(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error)

	// Get retrieves a single pull request by number
	Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error)

	// Create creates a new pull request
	Create(ctx context.Context, owner, repo string, input *models.CreatePRInput) (*models.PullRequest, error)

	// Update updates an existing pull request
	Update(ctx context.Context, owner, repo string, number int, input *models.UpdatePRInput) (*models.PullRequest, error)

	// Merge merges a pull request
	Merge(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error

	// Close closes a pull request without merging
	Close(ctx context.Context, owner, repo string, number int) error

	// Reopen reopens a closed pull request
	Reopen(ctx context.Context, owner, repo string, number int) error

	// GetDiff retrieves the diff for a pull request
	GetDiff(ctx context.Context, owner, repo string, number int) (string, error)

	// IsMergeable checks if a pull request is mergeable
	IsMergeable(ctx context.Context, owner, repo string, number int) (bool, error)

	// ListReviews retrieves reviews for a pull request
	ListReviews(ctx context.Context, owner, repo string, number int) ([]*models.Review, error)
}
