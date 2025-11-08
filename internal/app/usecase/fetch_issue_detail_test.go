package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/app/usecase"
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/mock"
	"go.uber.org/mock/gomock"
)

func TestFetchIssueDetailUseCase_Execute(t *testing.T) {
	now := time.Now()
	closedAt := now.Add(-1 * time.Hour)

	tests := []struct {
		name      string
		owner     string
		repo      string
		number    int
		mockSetup func(*mock.MockIssueRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *models.Issue)
	}{
		{
			name:   "正常系: Issue詳細取得成功",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 1).
					Return(&models.Issue{
						ID:     1,
						Number: 1,
						Title:  "Test Issue",
						Body:   "This is a test issue",
						State:  models.IssueStateOpen,
						Author: models.User{
							ID:    100,
							Login: "testuser",
							Name:  "Test User",
						},
						Labels: []models.Label{
							{Name: "bug", Color: "ff0000"},
							{Name: "priority-high", Color: "00ff00"},
						},
						Comments:  5,
						Locked:    false,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if issue.Number != 1 {
					t.Errorf("expected issue number 1, got %d", issue.Number)
				}
				if issue.Title != "Test Issue" {
					t.Errorf("expected title 'Test Issue', got %s", issue.Title)
				}
				if issue.State != models.IssueStateOpen {
					t.Errorf("expected state open, got %s", issue.State)
				}
				if len(issue.Labels) != 2 {
					t.Errorf("expected 2 labels, got %d", len(issue.Labels))
				}
				if issue.Comments != 5 {
					t.Errorf("expected 5 comments, got %d", issue.Comments)
				}
			},
		},
		{
			name:   "正常系: クローズ済みIssue取得",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 2,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 2).
					Return(&models.Issue{
						ID:        2,
						Number:    2,
						Title:     "Closed Issue",
						State:     models.IssueStateClosed,
						CreatedAt: now,
						UpdatedAt: now,
						ClosedAt:  &closedAt,
					}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if issue.State != models.IssueStateClosed {
					t.Errorf("expected state closed, got %s", issue.State)
				}
				if issue.ClosedAt == nil {
					t.Error("expected ClosedAt to be set")
				}
			},
		},
		{
			name:   "正常系: ラベルなしIssue取得",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 3,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 3).
					Return(&models.Issue{
						ID:        3,
						Number:    3,
						Title:     "Issue without labels",
						State:     models.IssueStateOpen,
						Labels:    []models.Label{},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if len(issue.Labels) != 0 {
					t.Errorf("expected 0 labels, got %d", len(issue.Labels))
				}
			},
		},
		{
			name:   "異常系: ownerが空",
			owner:  "",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name:   "異常系: repoが空",
			owner:  "test-owner",
			repo:   "",
			number: 1,
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:   "異常系: numberが0以下",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 0,
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "number must be greater than 0",
		},
		{
			name:   "異常系: numberが負の数",
			owner:  "test-owner",
			repo:   "test-repo",
			number: -1,
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "number must be greater than 0",
		},
		{
			name:   "異常系: リポジトリエラー",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 1).
					Return(nil, errors.New("repository error"))
			},
			wantErr: true,
			errMsg:  "failed to fetch issue detail",
		},
		{
			name:   "異常系: Issue not found",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 999,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 999).
					Return(nil, errors.New("issue not found"))
			},
			wantErr: true,
			errMsg:  "failed to fetch issue detail",
		},
		{
			name:   "異常系: 認証エラー",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 1).
					Return(nil, errors.New("authentication required"))
			},
			wantErr: true,
			errMsg:  "failed to fetch issue detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockIssueRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewFetchIssueDetailUseCase(mockRepo)
			got, err := uc.Execute(context.Background(), tt.owner, tt.repo, tt.number)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestFetchIssueDetailUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		Get(gomock.Any(), "test-owner", "test-repo", 1).
		Return(nil, context.Canceled)

	uc := usecase.NewFetchIssueDetailUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", 1)

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
