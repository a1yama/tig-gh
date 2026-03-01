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
	repo           repository.PullRequestRepository
	excludeAuthors map[string]struct{}
	excludeLabels  map[string]struct{}
}

// NewFetchPRsUseCase creates a new FetchPRsUseCase
func NewFetchPRsUseCase(repo repository.PullRequestRepository, excludeAuthors []string, excludeLabels []string) *FetchPRsUseCase {
	authors := make(map[string]struct{}, len(excludeAuthors))
	for _, a := range excludeAuthors {
		authors[a] = struct{}{}
	}
	labels := make(map[string]struct{}, len(excludeLabels))
	for _, l := range excludeLabels {
		labels[l] = struct{}{}
	}
	return &FetchPRsUseCase{
		repo:           repo,
		excludeAuthors: authors,
		excludeLabels:  labels,
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

	// Bot PR除外フィルタリング
	if len(uc.excludeAuthors) > 0 || len(uc.excludeLabels) > 0 {
		filtered := make([]*models.PullRequest, 0, len(prs))
		for _, pr := range prs {
			if _, excluded := uc.excludeAuthors[pr.Author.Login]; excluded {
				continue
			}
			if uc.hasExcludedLabel(pr.Labels) {
				continue
			}
			filtered = append(filtered, pr)
		}
		prs = filtered
	}

	return prs, nil
}

// hasExcludedLabel はPRのラベルに除外対象が含まれるかチェックする
func (uc *FetchPRsUseCase) hasExcludedLabel(labels []models.Label) bool {
	for _, label := range labels {
		if _, excluded := uc.excludeLabels[label.Name]; excluded {
			return true
		}
	}
	return false
}

// GetRepository returns the underlying PR repository
func (uc *FetchPRsUseCase) GetRepository() repository.PullRequestRepository {
	return uc.repo
}
