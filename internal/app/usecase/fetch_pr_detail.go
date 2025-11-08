package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchPRDetailUseCase is the use case for fetching pull request detail
type FetchPRDetailUseCase struct {
	repo repository.PullRequestRepository
}

// NewFetchPRDetailUseCase creates a new FetchPRDetailUseCase
func NewFetchPRDetailUseCase(repo repository.PullRequestRepository) *FetchPRDetailUseCase {
	return &FetchPRDetailUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch pull request detail
func (uc *FetchPRDetailUseCase) Execute(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error) {
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
	pr, err := uc.repo.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull request: %w", err)
	}

	return pr, nil
}
