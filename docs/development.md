# 開発ガイド

## 環境構築

### 必要なツール

- Go 1.21以上
- Git
- GitHub CLI (gh) - オプション、認証に使用
- Make - ビルドツール

### セットアップ手順

```bash
# リポジトリのクローン
git clone https://github.com/a1yama/tig-gh.git
cd tig-gh

# 依存関係のインストール
go mod download

# ビルド
make build

# テスト実行
make test

# 実行
./bin/tig-gh
```

### GitHub認証の設定

#### 方法1: Personal Access Token

```bash
# トークンの作成（GitHub Web UIで）
# Settings > Developer settings > Personal access tokens > Generate new token
# 必要なスコープ: repo, read:org, read:user

# 環境変数に設定
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# または設定ファイルに記述
mkdir -p ~/.config/tig-gh
cat > ~/.config/tig-gh/config.yaml <<EOF
github:
  token: ghp_xxxxxxxxxxxx
EOF
```

#### 方法2: GitHub CLI連携

```bash
# GitHub CLIで認証
gh auth login

# tig-ghは自動的にgh CLIの認証情報を使用
./bin/tig-gh
```

## プロジェクト構造

```
tig-gh/
├── cmd/
│   └── tig-gh/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── api/                     # GitHub API クライアント
│   │   ├── client.go
│   │   ├── issues.go
│   │   ├── pulls.go
│   │   └── commits.go
│   ├── app/                     # アプリケーションロジック
│   │   ├── usecase/
│   │   │   ├── fetch_issues.go
│   │   │   ├── create_issue.go
│   │   │   └── ...
│   │   └── service/
│   ├── config/                  # 設定管理
│   │   ├── config.go
│   │   └── loader.go
│   ├── domain/                  # ドメインモデル
│   │   ├── models/
│   │   │   ├── issue.go
│   │   │   ├── pull_request.go
│   │   │   ├── commit.go
│   │   │   └── ...
│   │   └── repository/          # リポジトリインターフェース
│   │       ├── issue_repository.go
│   │       ├── pr_repository.go
│   │       └── commit_repository.go
│   ├── infra/                   # インフラストラクチャ
│   │   ├── github/              # GitHub API実装
│   │   │   ├── client.go
│   │   │   ├── issue_repo_impl.go
│   │   │   ├── pr_repo_impl.go
│   │   │   └── commit_repo_impl.go
│   │   ├── cache/               # キャッシュ実装
│   │   │   ├── memory_cache.go
│   │   │   └── file_cache.go
│   │   └── config/              # 設定ファイル処理
│   └── ui/                      # TUI
│       ├── app.go               # メインアプリケーション
│       ├── components/          # 共通コンポーネント
│       │   ├── list.go
│       │   ├── detail.go
│       │   ├── input.go
│       │   ├── modal.go
│       │   └── statusbar.go
│       ├── views/               # 各ビュー
│       │   ├── issue_view.go
│       │   ├── pr_view.go
│       │   ├── commit_view.go
│       │   └── repo_view.go
│       └── styles/              # スタイル定義
│           ├── theme.go
│           └── colors.go
├── pkg/                         # 公開パッケージ
├── tests/                       # テスト
│   ├── integration/
│   └── fixtures/
├── docs/                        # ドキュメント
├── config/                      # デフォルト設定
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## 開発ワークフロー

### 1. ブランチ戦略

```bash
main              # 本番リリースブランチ
  ├── develop     # 開発ブランチ
  │   ├── feature/issue-view      # 機能開発
  │   ├── feature/pr-view
  │   └── fix/bug-123             # バグ修正
  └── hotfix/critical-bug         # 緊急修正
```

### 2. コミットメッセージ規約

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**:
- `feat`: 新機能
- `fix`: バグ修正
- `docs`: ドキュメント
- `style`: コードフォーマット
- `refactor`: リファクタリング
- `perf`: パフォーマンス改善
- `test`: テスト追加・修正
- `chore`: ビルド・補助ツール

**例**:
```
feat(ui): add issue detail view

Issue詳細表示ビューを実装

- Issueの全情報を表示
- コメント一覧の表示
- Markdown レンダリング

Closes #123
```

### 3. プルリクエスト

1. featureブランチを作成
```bash
git checkout -b feature/new-feature
```

2. 変更をコミット
```bash
git add .
git commit -m "feat: add new feature"
```

3. プッシュ
```bash
git push origin feature/new-feature
```

4. PRを作成
```bash
gh pr create --title "feat: add new feature" --body "説明"
```

5. レビュー後、マージ

## コーディング規約

### Goスタイル

```go
// Good
func (s *IssueService) FetchIssues(ctx context.Context, owner, repo string) ([]*Issue, error) {
    if owner == "" || repo == "" {
        return nil, errors.New("owner and repo are required")
    }

    issues, err := s.repo.List(ctx, owner, repo, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch issues: %w", err)
    }

    return issues, nil
}

// Bad - エラーハンドリング不足
func (s *IssueService) FetchIssues(owner, repo string) []*Issue {
    issues, _ := s.repo.List(context.Background(), owner, repo, nil)
    return issues
}
```

### インターフェース設計

```go
// Good - 小さく、単一責任
type IssueRepository interface {
    List(ctx context.Context, owner, repo string, opts *IssueOptions) ([]*Issue, error)
    Get(ctx context.Context, owner, repo string, number int) (*Issue, error)
    Create(ctx context.Context, owner, repo string, input *CreateIssueInput) (*Issue, error)
}

// Bad - 大きすぎる
type Repository interface {
    ListIssues()
    GetIssue()
    CreateIssue()
    ListPRs()
    GetPR()
    MergePR()
    // ... 多すぎる
}
```

### エラーハンドリング

```go
// カスタムエラー型
type ErrorType int

const (
    ErrTypeAuth ErrorType = iota
    ErrTypeNetwork
    ErrTypeNotFound
)

type AppError struct {
    Type    ErrorType
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// 使用例
func (r *IssueRepositoryImpl) Get(ctx context.Context, owner, repo string, number int) (*Issue, error) {
    issue, resp, err := r.client.Issues.Get(ctx, owner, repo, number)
    if err != nil {
        if resp != nil && resp.StatusCode == 404 {
            return nil, &AppError{
                Type:    ErrTypeNotFound,
                Message: "issue not found",
                Err:     err,
            }
        }
        return nil, &AppError{
            Type:    ErrTypeNetwork,
            Message: "failed to fetch issue",
            Err:     err,
        }
    }

    return convertToIssue(issue), nil
}
```

### bubbletea パターン

```go
// Model定義
type IssueListModel struct {
    issues      []*domain.Issue
    cursor      int
    selected    map[int]struct{}
    loading     bool
    err         error
    width       int
    height      int
    useCase     *usecase.FetchIssuesUseCase
}

// 初期化
func (m IssueListModel) Init() tea.Cmd {
    return m.fetchIssues()
}

// 更新
func (m IssueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "j", "down":
            if m.cursor < len(m.issues)-1 {
                m.cursor++
            }
        case "k", "up":
            if m.cursor > 0 {
                m.cursor--
            }
        case "enter":
            return m, m.showDetail()
        }

    case issuesLoadedMsg:
        m.loading = false
        m.issues = msg.issues
        return m, nil

    case errMsg:
        m.loading = false
        m.err = msg.err
        return m, nil

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }

    return m, nil
}

// 描画
func (m IssueListModel) View() string {
    if m.loading {
        return "Loading..."
    }

    if m.err != nil {
        return fmt.Sprintf("Error: %v", m.err)
    }

    var s strings.Builder

    // ヘッダー
    s.WriteString(headerStyle.Render("Issues"))
    s.WriteString("\n\n")

    // リスト
    for i, issue := range m.issues {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }

        line := fmt.Sprintf("%s #%d %s", cursor, issue.Number, issue.Title)
        if m.cursor == i {
            line = selectedStyle.Render(line)
        }
        s.WriteString(line)
        s.WriteString("\n")
    }

    // フッター
    s.WriteString("\n")
    s.WriteString(helpStyle.Render("j/k: navigate • enter: view • q: quit"))

    return s.String()
}

// コマンド
type issuesLoadedMsg struct {
    issues []*domain.Issue
}

type errMsg struct {
    err error
}

func (m IssueListModel) fetchIssues() tea.Cmd {
    return func() tea.Msg {
        issues, err := m.useCase.Execute(context.Background(), "owner", "repo", nil)
        if err != nil {
            return errMsg{err}
        }
        return issuesLoadedMsg{issues}
    }
}
```

## テスト

### ユニットテスト

```go
func TestFetchIssuesUseCase_Execute(t *testing.T) {
    tests := []struct {
        name    string
        owner   string
        repo    string
        want    int
        wantErr bool
    }{
        {
            name:    "正常系: Issue一覧取得",
            owner:   "test-owner",
            repo:    "test-repo",
            want:    5,
            wantErr: false,
        },
        {
            name:    "異常系: ownerが空",
            owner:   "",
            repo:    "test-repo",
            want:    0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // モックの準備
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockRepo := mock.NewMockIssueRepository(ctrl)
            if !tt.wantErr {
                mockRepo.EXPECT().
                    List(gomock.Any(), tt.owner, tt.repo, gomock.Any()).
                    Return(makeMockIssues(tt.want), nil)
            }

            // UseCase実行
            uc := usecase.NewFetchIssuesUseCase(mockRepo, nil)
            got, err := uc.Execute(context.Background(), tt.owner, tt.repo, nil)

            // アサーション
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if len(got) != tt.want {
                t.Errorf("Execute() got %d issues, want %d", len(got), tt.want)
            }
        })
    }
}
```

### 統合テスト

```go
func TestIssueRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // テスト用GitHubクライアント
    token := os.Getenv("GITHUB_TEST_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TEST_TOKEN not set")
    }

    client := github.NewClient(nil).WithAuthToken(token)
    repo := infra.NewIssueRepository(client)

    ctx := context.Background()
    issues, err := repo.List(ctx, "a1yama", "tig-gh", nil)

    assert.NoError(t, err)
    assert.NotEmpty(t, issues)
}
```

### テスト実行

```bash
# すべてのテスト
make test

# ユニットテストのみ
go test -short ./...

# 統合テスト
go test -run Integration ./...

# カバレッジ
make coverage

# 特定のパッケージ
go test ./internal/app/usecase/...
```

## ビルド

### Makefile

```makefile
.PHONY: build test coverage lint clean

# 変数
BINARY_NAME=tig-gh
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# ビルド
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/tig-gh/main.go

# インストール
install:
	go install $(LDFLAGS) cmd/tig-gh/main.go

# テスト
test:
	go test -v -short ./...

# カバレッジ
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run

# フォーマット
fmt:
	go fmt ./...
	goimports -w .

# クリーン
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# 開発モード（ホットリロード）
dev:
	air

# クロスコンパイル
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 cmd/tig-gh/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 cmd/tig-gh/main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 cmd/tig-gh/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/tig-gh/main.go
```

## デバッグ

### ログ出力

```go
import "log/slog"

// main.goでロガーを初期化
func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))
    slog.SetDefault(logger)

    // アプリケーション起動
}

// 使用例
slog.Debug("fetching issues", "owner", owner, "repo", repo)
slog.Info("issues fetched", "count", len(issues))
slog.Error("failed to fetch issues", "error", err)
```

### デバッグモード

```bash
# デバッグログを有効化
TIG_GH_DEBUG=1 ./bin/tig-gh

# ログファイルに出力
TIG_GH_LOG_FILE=~/.config/tig-gh/debug.log ./bin/tig-gh
```

## リリース

### バージョニング

セマンティックバージョニングを使用: `MAJOR.MINOR.PATCH`

```bash
# タグ作成
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### GitHub Actions（自動リリース）

`.github/workflows/release.yml`でgoreleaserを使用してバイナリを自動生成

```bash
# ローカルでテスト
goreleaser release --snapshot --clean
```

## トラブルシューティング

### よくある問題

1. **認証エラー**
```bash
# トークンの確認
echo $GITHUB_TOKEN

# トークンの権限を確認（GitHub Web UI）
```

2. **ビルドエラー**
```bash
# 依存関係の更新
go mod tidy
go mod download
```

3. **テストの失敗**
```bash
# キャッシュクリア
go clean -testcache
go test ./...
```

## 参考リソース

- [bubbletea documentation](https://github.com/charmbracelet/bubbletea)
- [lipgloss examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [go-github library](https://github.com/google/go-github)
- [GitHub REST API](https://docs.github.com/en/rest)
- [Effective Go](https://go.dev/doc/effective_go)
