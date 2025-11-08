package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchIssueDetailUseCase is the use case for fetching issue detail
type FetchIssueDetailUseCase struct {
	repo repository.IssueRepository
}

// NewFetchIssueDetailUseCase creates a new FetchIssueDetailUseCase
func NewFetchIssueDetailUseCase(repo repository.IssueRepository) *FetchIssueDetailUseCase {
	return &FetchIssueDetailUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch issue detail
func (uc *FetchIssueDetailUseCase) Execute(ctx context.Context, owner, repo string, number int) (*models.Issue, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	if number <= 0 {
		return nil, errors.New("number must be greater than 0")
	}

	// リポジトリから取得
	issue, err := uc.repo.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue detail: %w", err)
	}

	return issue, nil
}
