package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CreateIssueUseCase is the use case for creating an issue
type CreateIssueUseCase struct {
	repo repository.IssueRepository
}

// NewCreateIssueUseCase creates a new CreateIssueUseCase
func NewCreateIssueUseCase(repo repository.IssueRepository) *CreateIssueUseCase {
	return &CreateIssueUseCase{
		repo: repo,
	}
}

// Execute executes the use case to create an issue
func (uc *CreateIssueUseCase) Execute(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
	// バリデーション
	if owner == "" {
		return nil, errors.New("owner is required")
	}

	if repo == "" {
		return nil, errors.New("repo is required")
	}

	if input == nil {
		return nil, errors.New("input is required")
	}

	if strings.TrimSpace(input.Title) == "" {
		return nil, errors.New("title is required")
	}

	// リポジトリで作成
	issue, err := uc.repo.Create(ctx, owner, repo, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}
