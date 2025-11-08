package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// MergePRUseCase is the use case for merging a pull request
type MergePRUseCase struct {
	repo repository.PullRequestRepository
}

// NewMergePRUseCase creates a new MergePRUseCase
func NewMergePRUseCase(repo repository.PullRequestRepository) *MergePRUseCase {
	return &MergePRUseCase{
		repo: repo,
	}
}

// Execute executes the use case to merge a pull request
func (uc *MergePRUseCase) Execute(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
	// バリデーション
	if owner == "" {
		return errors.New("owner is required")
	}

	if repo == "" {
		return errors.New("repo is required")
	}

	if number <= 0 {
		return errors.New("number must be greater than 0")
	}

	// リポジトリでマージ実行
	if err := uc.repo.Merge(ctx, owner, repo, number, opts); err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	return nil
}
