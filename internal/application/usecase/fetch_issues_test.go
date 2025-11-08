package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// mockIssueRepository is a mock implementation of IssueRepository for testing
type mockIssueRepository struct {
	listFunc func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error)
}

func (m *mockIssueRepository) List(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, owner, repo, opts)
	}
	return nil, nil
}

func (m *mockIssueRepository) Get(ctx context.Context, owner, repo string, number int) (*models.Issue, error) {
	return nil, nil
}

func (m *mockIssueRepository) Create(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
	return nil, nil
}

func (m *mockIssueRepository) Update(ctx context.Context, owner, repo string, number int, input *models.UpdateIssueInput) (*models.Issue, error) {
	return nil, nil
}

func (m *mockIssueRepository) Close(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func (m *mockIssueRepository) Reopen(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func (m *mockIssueRepository) Lock(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func (m *mockIssueRepository) Unlock(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func TestFetchIssuesUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		owner       string
		repo        string
		opts        *models.IssueOptions
		mockReturn  []*models.Issue
		mockError   error
		wantErr     bool
		checkResult func(t *testing.T, result []*models.Issue)
	}{
		{
			name:  "success with default options",
			owner: "testowner",
			repo:  "testrepo",
			opts:  nil,
			mockReturn: []*models.Issue{
				{
					Number:    1,
					Title:     "Test Issue 1",
					State:     models.IssueStateOpen,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					Number:    2,
					Title:     "Test Issue 2",
					State:     models.IssueStateOpen,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			mockError: nil,
			wantErr:   false,
			checkResult: func(t *testing.T, result []*models.Issue) {
				if len(result) != 2 {
					t.Errorf("expected 2 issues, got %d", len(result))
				}
				if result[0].Title != "Test Issue 1" {
					t.Errorf("expected title 'Test Issue 1', got '%s'", result[0].Title)
				}
			},
		},
		{
			name:  "success with custom options",
			owner: "testowner",
			repo:  "testrepo",
			opts: &models.IssueOptions{
				State:     models.IssueStateClosed,
				Sort:      models.IssueSortCreated,
				Direction: models.SortDirectionAsc,
				PerPage:   50,
			},
			mockReturn: []*models.Issue{
				{
					Number:    3,
					Title:     "Closed Issue",
					State:     models.IssueStateClosed,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			mockError: nil,
			wantErr:   false,
			checkResult: func(t *testing.T, result []*models.Issue) {
				if len(result) != 1 {
					t.Errorf("expected 1 issue, got %d", len(result))
				}
				if result[0].State != models.IssueStateClosed {
					t.Errorf("expected state 'closed', got '%s'", result[0].State)
				}
			},
		},
		{
			name:       "repository error",
			owner:      "testowner",
			repo:       "testrepo",
			opts:       nil,
			mockReturn: nil,
			mockError:  errors.New("repository error"),
			wantErr:    true,
			checkResult: func(t *testing.T, result []*models.Issue) {
				if result != nil {
					t.Errorf("expected nil result on error, got %v", result)
				}
			},
		},
		{
			name:       "empty result",
			owner:      "testowner",
			repo:       "testrepo",
			opts:       nil,
			mockReturn: []*models.Issue{},
			mockError:  nil,
			wantErr:    false,
			checkResult: func(t *testing.T, result []*models.Issue) {
				if len(result) != 0 {
					t.Errorf("expected 0 issues, got %d", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockIssueRepository{
				listFunc: func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
					if owner != tt.owner {
						t.Errorf("expected owner '%s', got '%s'", tt.owner, owner)
					}
					if repo != tt.repo {
						t.Errorf("expected repo '%s', got '%s'", tt.repo, repo)
					}
					// Check if default options are set when opts is nil
					if tt.opts == nil && opts == nil {
						t.Error("expected default options to be set")
					}
					return tt.mockReturn, tt.mockError
				},
			}

			useCase := NewFetchIssuesUseCase(mockRepo)
			result, err := useCase.Execute(context.Background(), tt.owner, tt.repo, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
