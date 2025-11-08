package github_test

import (
	"testing"

	"github.com/a1yama/tig-gh/internal/infra/github"
)

// TestPRRepositoryImpl_NewPullRequestRepository tests the repository creation
func TestPRRepositoryImpl_NewPullRequestRepository(t *testing.T) {
	client := github.NewClient("fake-token")
	repo := github.NewPullRequestRepository(client)

	if repo == nil {
		t.Error("NewPullRequestRepository() returned nil")
	}
}

// Note: インフラ層の実際のAPIコールのテストは、
// E2Eテストまたは実際のGitHub APIのモックサーバーが必要です。
// ユースケース層のテストで、モックリポジトリを使用した
// ビジネスロジックのテストは既に完了しています。
