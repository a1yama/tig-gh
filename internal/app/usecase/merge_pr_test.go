package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/a1yama/tig-gh/internal/app/usecase"
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/mock"
	"go.uber.org/mock/gomock"
)

func TestMergePRUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		repo      string
		number    int
		opts      *models.MergeOptions
		mockSetup func(*mock.MockPullRequestRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "正常系: PRマージ成功（デフォルトオプション）",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "正常系: PRマージ成功（マージメソッド指定）",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 2,
			opts: &models.MergeOptions{
				MergeMethod: models.MergeMethodSquash,
			},
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 2, gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
						if opts.MergeMethod != models.MergeMethodSquash {
							t.Errorf("expected merge method to be squash, got %s", opts.MergeMethod)
						}
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:   "正常系: PRマージ成功（コミットメッセージ指定）",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 3,
			opts: &models.MergeOptions{
				CommitTitle:   "Custom merge title",
				CommitMessage: "Custom merge message",
				MergeMethod:   models.MergeMethodMerge,
			},
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 3, gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
						if opts.CommitTitle != "Custom merge title" {
							t.Errorf("expected commit title to be 'Custom merge title', got %s", opts.CommitTitle)
						}
						if opts.CommitMessage != "Custom merge message" {
							t.Errorf("expected commit message to be 'Custom merge message', got %s", opts.CommitMessage)
						}
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:   "正常系: Rebaseマージ",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 4,
			opts: &models.MergeOptions{
				MergeMethod: models.MergeMethodRebase,
			},
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 4, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "異常系: ownerが空",
			owner:  "",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
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
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
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
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
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
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "number must be greater than 0",
		},
		{
			name:   "異常系: PRがマージ不可",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
					Return(errors.New("pull request is not mergeable"))
			},
			wantErr: true,
			errMsg:  "failed to merge pull request",
		},
		{
			name:   "異常系: コンフリクトあり",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
					Return(errors.New("merge conflict"))
			},
			wantErr: true,
			errMsg:  "failed to merge pull request",
		},
		{
			name:   "異常系: 既にマージ済み",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
					Return(errors.New("pull request already merged"))
			},
			wantErr: true,
			errMsg:  "failed to merge pull request",
		},
		{
			name:   "異常系: リポジトリエラー",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			opts:   nil,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
					Return(errors.New("repository error"))
			},
			wantErr: true,
			errMsg:  "failed to merge pull request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockPullRequestRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewMergePRUseCase(mockRepo)
			err := uc.Execute(context.Background(), tt.owner, tt.repo, tt.number, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestMergePRUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		Merge(gomock.Any(), "test-owner", "test-repo", 1, gomock.Any()).
		Return(context.Canceled)

	uc := usecase.NewMergePRUseCase(mockRepo)
	err := uc.Execute(ctx, "test-owner", "test-repo", 1, nil)

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
