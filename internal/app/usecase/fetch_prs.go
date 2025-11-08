package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchPRsUseCase is the use case for fetching pull requests
type FetchPRsUseCase struct {
	repo repository.PullRequestRepository
}

// NewFetchPRsUseCase creates a new FetchPRsUseCase
func NewFetchPRsUseCase(repo repository.PullRequestRepository) *FetchPRsUseCase {
	return &FetchPRsUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch pull requests
func (uc *FetchPRsUseCase) Execute(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	// リポジトリから取得
	prs, err := uc.repo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	return prs, nil
}
