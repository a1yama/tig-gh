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

func TestCreateIssueUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		owner     string
		repo      string
		input     *models.CreateIssueInput
		mockSetup func(*mock.MockIssueRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *models.Issue)
	}{
		{
			name:  "正常系: Issue作成成功",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "New Test Issue",
				Body:  "This is a new test issue",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
						return &models.Issue{
							ID:        1,
							Number:    1,
							Title:     input.Title,
							Body:      input.Body,
							State:     models.IssueStateOpen,
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if issue.Title != "New Test Issue" {
					t.Errorf("expected title 'New Test Issue', got %s", issue.Title)
				}
				if issue.State != models.IssueStateOpen {
					t.Errorf("expected state open, got %s", issue.State)
				}
			},
		},
		{
			name:  "正常系: ラベル付きIssue作成",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title:  "Bug Report",
				Body:   "Found a bug",
				Labels: []string{"bug", "priority-high"},
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
						if len(input.Labels) != 2 {
							t.Errorf("expected 2 labels, got %d", len(input.Labels))
						}
						return &models.Issue{
							ID:     1,
							Number: 1,
							Title:  input.Title,
							Body:   input.Body,
							State:  models.IssueStateOpen,
							Labels: []models.Label{
								{Name: "bug", Color: "ff0000"},
								{Name: "priority-high", Color: "00ff00"},
							},
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if len(issue.Labels) != 2 {
					t.Errorf("expected 2 labels, got %d", len(issue.Labels))
				}
			},
		},
		{
			name:  "正常系: アサイン付きIssue作成",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title:     "Task Issue",
				Body:      "Task description",
				Assignees: []string{"testuser"},
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
						if len(input.Assignees) != 1 {
							t.Errorf("expected 1 assignee, got %d", len(input.Assignees))
						}
						return &models.Issue{
							ID:     1,
							Number: 1,
							Title:  input.Title,
							Body:   input.Body,
							State:  models.IssueStateOpen,
							Assignees: []models.User{
								{Login: "testuser", Name: "Test User"},
							},
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if len(issue.Assignees) != 1 {
					t.Errorf("expected 1 assignee, got %d", len(issue.Assignees))
				}
			},
		},
		{
			name:  "正常系: マイルストーン付きIssue作成",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title:     "Milestone Issue",
				Body:      "Issue for milestone",
				Milestone: 1,
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
						if input.Milestone != 1 {
							t.Errorf("expected milestone 1, got %d", input.Milestone)
						}
						return &models.Issue{
							ID:     1,
							Number: 1,
							Title:  input.Title,
							Body:   input.Body,
							State:  models.IssueStateOpen,
							Milestone: &models.Milestone{
								ID:     1,
								Number: 1,
								Title:  "v1.0",
							},
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					})
			},
			wantErr: false,
			validate: func(t *testing.T, issue *models.Issue) {
				if issue.Milestone == nil {
					t.Error("expected milestone to be set")
				}
			},
		},
		{
			name:  "異常系: ownerが空",
			owner: "",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "Test Issue",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name:  "異常系: repoが空",
			owner: "test-owner",
			repo:  "",
			input: &models.CreateIssueInput{
				Title: "Test Issue",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:  "異常系: inputがnil",
			owner: "test-owner",
			repo:  "test-repo",
			input: nil,
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "input is required",
		},
		{
			name:  "異常系: タイトルが空",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "",
				Body:  "Body text",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name:  "異常系: タイトルが空白のみ",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "   ",
				Body:  "Body text",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name:  "異常系: リポジトリエラー",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "Test Issue",
				Body:  "Test body",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			wantErr: true,
			errMsg:  "failed to create issue",
		},
		{
			name:  "異常系: 認証エラー",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "Test Issue",
				Body:  "Test body",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("authentication required"))
			},
			wantErr: true,
			errMsg:  "failed to create issue",
		},
		{
			name:  "異常系: 権限エラー",
			owner: "test-owner",
			repo:  "test-repo",
			input: &models.CreateIssueInput{
				Title: "Test Issue",
				Body:  "Test body",
			},
			mockSetup: func(m *mock.MockIssueRepository) {
				m.EXPECT().
					Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("permission denied"))
			},
			wantErr: true,
			errMsg:  "failed to create issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockIssueRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewCreateIssueUseCase(mockRepo)
			got, err := uc.Execute(context.Background(), tt.owner, tt.repo, tt.input)

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

func TestCreateIssueUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		Create(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
		Return(nil, context.Canceled)

	uc := usecase.NewCreateIssueUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", &models.CreateIssueInput{
		Title: "Test Issue",
	})

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
