package repository

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// IssueRepository defines the interface for issue operations
type IssueRepository interface {
	// List retrieves a list of issues for a repository
	List(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error)

	// Get retrieves a single issue by number
	Get(ctx context.Context, owner, repo string, number int) (*models.Issue, error)

	// Create creates a new issue
	Create(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error)

	// Update updates an existing issue
	Update(ctx context.Context, owner, repo string, number int, input *models.UpdateIssueInput) (*models.Issue, error)

	// Close closes an issue
	Close(ctx context.Context, owner, repo string, number int) error

	// Reopen reopens a closed issue
	Reopen(ctx context.Context, owner, repo string, number int) error

	// Lock locks an issue
	Lock(ctx context.Context, owner, repo string, number int) error

	// Unlock unlocks an issue
	Unlock(ctx context.Context, owner, repo string, number int) error

	// ListComments retrieves comments for an issue
	ListComments(ctx context.Context, owner, repo string, number int, opts *models.CommentOptions) ([]*models.Comment, error)
}
