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

func TestFetchCommitDetailUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		owner     string
		repo      string
		sha       string
		mockSetup func(*mock.MockCommitRepository)
		wantSHA   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "正常系: コミット詳細取得成功",
			owner: "test-owner",
			repo:  "test-repo",
			sha:   "abc123",
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", "abc123").
					Return(&models.Commit{
						SHA:     "abc123",
						Message: "feat: add new feature\n\nThis is a detailed description.",
						Author: models.CommitAuthor{
							Name:  "Alice",
							Email: "alice@example.com",
							Date:  now,
						},
						Committer: models.CommitAuthor{
							Name:  "Alice",
							Email: "alice@example.com",
							Date:  now,
						},
						Parents:   []string{"parent123"},
						Tree:      "tree456",
						URL:       "https://github.com/test-owner/test-repo/commit/abc123",
						CreatedAt: now,
						Stats: &models.CommitStats{
							Additions: 100,
							Deletions: 50,
							Total:     150,
						},
						Files: []*models.DiffFile{
							{
								Filename:  "main.go",
								Status:    models.FileStatusModified,
								Additions: 50,
								Deletions: 20,
								Changes:   70,
							},
						},
					}, nil)
			},
			wantSHA: "abc123",
			wantErr: false,
		},
		{
			name:  "異常系: ownerが空",
			owner: "",
			repo:  "test-repo",
			sha:   "abc123",
			mockSetup: func(m *mock.MockCommitRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name:  "異常系: repoが空",
			owner: "test-owner",
			repo:  "",
			sha:   "abc123",
			mockSetup: func(m *mock.MockCommitRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:  "異常系: shaが空",
			owner: "test-owner",
			repo:  "test-repo",
			sha:   "",
			mockSetup: func(m *mock.MockCommitRepository) {
				// モックは呼ばれない
			},
			wantErr: true,
			errMsg:  "sha is required",
		},
		{
			name:  "異常系: コミットが見つからない",
			owner: "test-owner",
			repo:  "test-repo",
			sha:   "notfound",
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", "notfound").
					Return(nil, errors.New("commit not found"))
			},
			wantErr: true,
			errMsg:  "failed to fetch commit detail",
		},
		{
			name:  "異常系: 認証エラー",
			owner: "test-owner",
			repo:  "test-repo",
			sha:   "abc123",
			mockSetup: func(m *mock.MockCommitRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", "abc123").
					Return(nil, errors.New("authentication required"))
			},
			wantErr: true,
			errMsg:  "failed to fetch commit detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockCommitRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewFetchCommitDetailUseCase(mockRepo)
			got, err := uc.Execute(context.Background(), tt.owner, tt.repo, tt.sha)

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
				if got.SHA != tt.wantSHA {
					t.Errorf("Execute() got SHA = %v, want %v", got.SHA, tt.wantSHA)
				}
				if got.Stats == nil {
					t.Error("Execute() expected stats to be set")
				}
				if len(got.Files) == 0 {
					t.Error("Execute() expected files to be set")
				}
			}
		})
	}
}

func TestFetchCommitDetailUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCommitRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		Get(gomock.Any(), "test-owner", "test-repo", "abc123").
		Return(nil, context.Canceled)

	uc := usecase.NewFetchCommitDetailUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", "abc123")

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
