package usecase

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// SearchUseCase handles search operations
type SearchUseCase struct {
	repo repository.SearchRepository
}

// NewSearchUseCase creates a new SearchUseCase
func NewSearchUseCase(repo repository.SearchRepository) *SearchUseCase {
	return &SearchUseCase{
		repo: repo,
	}
}

// Execute performs a search with the given options
func (uc *SearchUseCase) Execute(ctx context.Context, owner, repo string, opts *models.SearchOptions) (*models.SearchResults, error) {
	// Set defaults if opts is nil
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

	// Set default values for unset fields
	if opts.Type == "" {
		opts.Type = models.SearchTypeBoth
	}
	if opts.State == "" {
		opts.State = models.IssueStateOpen
	}
	if opts.Sort == "" {
		opts.Sort = models.SearchSortUpdated
	}
	if opts.Direction == "" {
		opts.Direction = models.SortDirectionDesc
	}
	if opts.PerPage == 0 {
		opts.PerPage = 30
	}
	if opts.Page == 0 {
		opts.Page = 1
	}

	return uc.repo.Search(ctx, owner, repo, opts)
}

// GetRepository returns the underlying search repository
func (uc *SearchUseCase) GetRepository() repository.SearchRepository {
	return uc.repo
}
