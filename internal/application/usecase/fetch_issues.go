package usecase

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchIssuesUseCase defines the interface for fetching issues
type FetchIssuesUseCase interface {
	Execute(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error)
}

// fetchIssuesUseCase is the implementation of FetchIssuesUseCase
type fetchIssuesUseCase struct {
	issueRepo repository.IssueRepository
}

// NewFetchIssuesUseCase creates a new FetchIssuesUseCase
func NewFetchIssuesUseCase(issueRepo repository.IssueRepository) FetchIssuesUseCase {
	return &fetchIssuesUseCase{
		issueRepo: issueRepo,
	}
}

// Execute fetches issues from the repository
func (u *fetchIssuesUseCase) Execute(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	// Set default options if not provided
	if opts == nil {
		opts = &models.IssueOptions{
			State:     models.IssueStateOpen,
			Sort:      models.IssueSortUpdated,
			Direction: models.SortDirectionDesc,
			PerPage:   100,
		}
	}

	// Fetch issues from repository
	issues, err := u.issueRepo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	return issues, nil
}
