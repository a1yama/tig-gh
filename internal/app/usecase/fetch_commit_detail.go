package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// FetchCommitDetailUseCase is the use case for fetching commit detail
type FetchCommitDetailUseCase struct {
	repo repository.CommitRepository
}

// NewFetchCommitDetailUseCase creates a new FetchCommitDetailUseCase
func NewFetchCommitDetailUseCase(repo repository.CommitRepository) *FetchCommitDetailUseCase {
	return &FetchCommitDetailUseCase{
		repo: repo,
	}
}

// Execute executes the use case to fetch commit detail
func (uc *FetchCommitDetailUseCase) Execute(ctx context.Context, owner, repo, sha string) (*models.Commit, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	if sha == "" {
		return nil, errors.New("sha is required")
	}

	// リポジトリから取得
	commit, err := uc.repo.Get(ctx, owner, repo, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit detail: %w", err)
	}

	return commit, nil
}
