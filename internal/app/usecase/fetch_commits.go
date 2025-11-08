package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchCommitsUseCase is the use case for fetching commits
type FetchCommitsUseCase struct {
	repo repository.CommitRepository
}

// NewFetchCommitsUseCase creates a new FetchCommitsUseCase
func NewFetchCommitsUseCase(repo repository.CommitRepository) *FetchCommitsUseCase {
	return &FetchCommitsUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch commits
func (uc *FetchCommitsUseCase) Execute(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	// リポジトリから取得
	commits, err := uc.repo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}

	return commits, nil
}
