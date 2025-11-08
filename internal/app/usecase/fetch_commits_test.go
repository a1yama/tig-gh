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

func TestFetchCommitsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		repo      string
		opts      *models.CommitOptions
		mockSetup func(*mock.MockCommitRepository)
		want      int
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "正常系: コミット一覧取得成功",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return([]*models.Commit{
						{
							SHA:     "abc123",
							Message: "feat: add new feature",
							Author: models.CommitAuthor{
								Name:  "Alice",
								Email: "alice@example.com",
								Date:  time.Now(),
							},
							CreatedAt: time.Now(),
						},
						{
							SHA:     "def456",
							Message: "fix: bug fix",
							Author: models.CommitAuthor{
								Name:  "Bob",
								Email: "bob@example.com",
								Date:  time.Now(),
							},
							CreatedAt: time.Now(),
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
			opts: &models.CommitOptions{
				SHA:     "main",
				Author:  "alice@example.com",
				PerPage: 10,
			},
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					DoAndReturn(func(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
						if opts.SHA != "main" {
							t.Errorf("expected SHA to be main, got %s", opts.SHA)
						}
						if opts.Author != "alice@example.com" {
							t.Errorf("expected author to be alice@example.com, got %s", opts.Author)
						}
						return []*models.Commit{
							{
								SHA:     "abc123",
								Message: "feat: add feature by alice",
								Author: models.CommitAuthor{
									Name:  "Alice",
									Email: "alice@example.com",
									Date:  time.Now(),
								},
								CreatedAt: time.Now(),
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
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return([]*models.Commit{}, nil)
			},
			want:    0,
			wantErr: false,
		},
		{
			name:  "異常系: ownerが空",
			owner: "",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockCommitRepository) {
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
			mockSetup: func(m *mock.MockCommitRepository) {
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
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			want:    0,
			wantErr: true,
			errMsg:  "failed to fetch commits",
		},
		{
			name:  "異常系: 認証エラー",
			owner: "test-owner",
			repo:  "test-repo",
			opts:  nil,
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
					Return(nil, errors.New("authentication required"))
			},
			want:    0,
			wantErr: true,
			errMsg:  "failed to fetch commits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockCommitRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewFetchCommitsUseCase(mockRepo)
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
					t.Errorf("Execute() got %d commits, want %d", len(got), tt.want)
				}
			}
		})
	}
}

func TestFetchCommitsUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCommitRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		List(gomock.Any(), "test-owner", "test-repo", gomock.Any()).
		Return(nil, context.Canceled)

	uc := usecase.NewFetchCommitsUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", nil)

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
