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

func TestFetchPRsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		repo      string
		opts      *models.PROptions
		mockSetup func(*mock.MockPullRequestRepository)
		want      int
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "正常系: PR一覧取得成功",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return([]*models.PullRequest{
						{
							ID:        1,
							Number:    1,
							Title:     "Test PR 1",
							State:     models.PRStateOpen,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						{
							ID:        2,
							Number:    2,
							Title:     "Test PR 2",
							State:     models.PRStateOpen,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil)
			},
			want:    2,
			wantErr: false,
		},
		{
			name:  "正常系: オプション指定で取得成功",
			owner: "test-owner",
			repo:  "test-repo",
			opts: &models.PROptions{
				State:   models.PRStateOpen,
				Base:    "main",
				PerPage: 10,
			},
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
						if opts.State != models.PRStateOpen {
							t.Errorf("expected state to be open, got %s", opts.State)
						}
						if opts.Base != "main" {
							t.Errorf("expected base to be main, got %s", opts.Base)
						}
						return []*models.PullRequest{
							{
								ID:     1,
								Number: 1,
								Title:  "PR to main",
								State:  models.PRStateOpen,
								Base: models.Branch{
									Name: "main",
									SHA:  "abc123",
								},
								CreatedAt: time.Now(),
								UpdatedAt: time.Now(),
							},
						}, nil
					})
			},
			want:    1,
			wantErr: false,
		},
		{
			name:  "正常系: 結果が空の場合",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return([]*models.PullRequest{}, nil)
			},
			want:    0,
			wantErr: false,
		},
		{
			name:  "異常系: ownerが空",
			owner: "",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    0,
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name:  "異常系: repoが空",
			owner: "test-owner",
			repo:  "",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    0,
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:  "異常系: リポジトリエラー",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			want:    0,
			wantErr: true,
			errMsg:  "failed to fetch pull requests",
		},
		{
			name:  "異常系: 認証エラー",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("authentication required"))
			},
			want:    0,
			wantErr: true,
			errMsg:  "failed to fetch pull requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockPullRequestRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewFetchPRsUseCase(mockRepo)
			got, err := uc.Execute(context.Background(), tt.owner, tt.repo, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}

			if !tt.wantErr {
				if len(got) != tt.want {
					t.Errorf("Execute() got %d PRs, want %d", len(got), tt.want)
				}
			}
		})
	}
}

func TestFetchPRsUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
		Return(nil, context.Canceled)

	uc := usecase.NewFetchPRsUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", nil)

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
