package repository

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// SearchRepository defines the interface for searching issues and pull requests
type SearchRepository interface {
	// Search searches for issues and pull requests based on the given options
	Search(ctx context.Context, owner, repo string, opts *models.SearchOptions) (*models.SearchResults, error)
}
