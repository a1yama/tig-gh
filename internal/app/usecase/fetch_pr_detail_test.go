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

func TestFetchPRDetailUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		owner     string
		repo      string
		number    int
		mockSetup func(*mock.MockPullRequestRepository)
		want      *models.PullRequest
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "正常系: PR詳細取得成功",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 1).
					Return(&models.PullRequest{
						ID:     1,
						Number: 1,
						Title:  "Test PR 1",
						Body:   "This is a test PR",
						State:  models.PRStateOpen,
						Author: models.User{
							ID:    1,
							Login: "test-user",
						},
						Head: models.Branch{
							Name: "feature",
							SHA:  "abc123",
						},
						Base: models.Branch{
							Name: "main",
							SHA:  "def456",
						},
						Mergeable:      true,
						MergeableState: "clean",
						Draft:          false,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)
			},
			want: &models.PullRequest{
				ID:     1,
				Number: 1,
				Title:  "Test PR 1",
				Body:   "This is a test PR",
				State:  models.PRStateOpen,
			},
			wantErr: false,
		},
		{
			name:   "正常系: Draft PRの取得",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 2,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 2).
					Return(&models.PullRequest{
						ID:     2,
						Number: 2,
						Title:  "Draft PR",
						State:  models.PRStateOpen,
						Draft:  true,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			want: &models.PullRequest{
				ID:     2,
				Number: 2,
				Title:  "Draft PR",
				Draft:  true,
			},
			wantErr: false,
		},
		{
			name:   "正常系: マージ済みPRの取得",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 3,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				mergedAt := now.Add(-1 * time.Hour)
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 3).
					Return(&models.PullRequest{
						ID:        3,
						Number:    3,
						Title:     "Merged PR",
						State:     models.PRStateClosed,
						Merged:    true,
						MergedAt:  &mergedAt,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			want: &models.PullRequest{
				ID:     3,
				Number: 3,
				Title:  "Merged PR",
				Merged: true,
			},
			wantErr: false,
		},
		{
			name:   "異常系: ownerが空",
			owner:  "",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    nil,
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name:   "異常系: repoが空",
			owner:  "test-owner",
			repo:   "",
			number: 1,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    nil,
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:   "異常系: numberが0以下",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 0,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    nil,
			wantErr: true,
			errMsg:  "number must be greater than 0",
		},
		{
			name:   "異常系: numberが負の数",
			owner:  "test-owner",
			repo:   "test-repo",
			number: -1,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				// モックは呼ばれない
			},
			want:    nil,
			wantErr: true,
			errMsg:  "number must be greater than 0",
		},
		{
			name:   "異常系: PRが見つからない",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 999,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 999).
					Return(nil, errors.New("resource not found (404)"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to fetch pull request",
		},
		{
			name:   "異常系: リポジトリエラー",
			owner:  "test-owner",
			repo:   "test-repo",
			number: 1,
			mockSetup: func(m *mock.MockPullRequestRepository) {
				m.EXPECT().
					Get(gomock.Any(), "test-owner", "test-repo", 1).
					Return(nil, errors.New("repository error"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to fetch pull request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockPullRequestRepository(ctrl)
			tt.mockSetup(mockRepo)

			uc := usecase.NewFetchPRDetailUseCase(mockRepo)
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

			if !tt.wantErr {
				if got == nil {
					t.Error("Execute() got nil, want non-nil PR")
					return
				}
				if got.ID != tt.want.ID {
					t.Errorf("Execute() got ID = %d, want %d", got.ID, tt.want.ID)
				}
				if got.Number != tt.want.Number {
					t.Errorf("Execute() got Number = %d, want %d", got.Number, tt.want.Number)
				}
				if got.Title != tt.want.Title {
					t.Errorf("Execute() got Title = %s, want %s", got.Title, tt.want.Title)
				}
			}
		})
	}
}

func TestFetchPRDetailUseCase_Execute_Context(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	mockRepo.EXPECT().
		Get(gomock.Any(), "test-owner", "test-repo", 1).
		Return(nil, context.Canceled)

	uc := usecase.NewFetchPRDetailUseCase(mockRepo)
	_, err := uc.Execute(ctx, "test-owner", "test-repo", 1)

	if err == nil {
		t.Error("Execute() expected error for canceled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}
