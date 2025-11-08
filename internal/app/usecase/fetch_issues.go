package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchIssuesUseCase is the use case for fetching issues
type FetchIssuesUseCase struct {
	repo repository.IssueRepository
}

// NewFetchIssuesUseCase creates a new FetchIssuesUseCase
func NewFetchIssuesUseCase(repo repository.IssueRepository) *FetchIssuesUseCase {
	return &FetchIssuesUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch issues
func (uc *FetchIssuesUseCase) Execute(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	// リポジトリから取得
	issues, err := uc.repo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	return issues, nil
}
