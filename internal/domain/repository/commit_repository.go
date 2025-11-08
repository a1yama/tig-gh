package repository

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// CommitRepository defines the interface for commit operations
type CommitRepository interface {
	// List retrieves a list of commits for a repository
	List(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error)

	// Get retrieves a single commit by SHA
	Get(ctx context.Context, owner, repo string, sha string) (*models.Commit, error)

	// Compare compares two commits
	Compare(ctx context.Context, owner, repo, base, head string) (*models.Comparison, error)

	// ListBranches retrieves a list of branches for a repository
	ListBranches(ctx context.Context, owner, repo string) ([]*models.Branch, error)

	// GetBranch retrieves a single branch by name
	GetBranch(ctx context.Context, owner, repo, branch string) (*models.Branch, error)
}
