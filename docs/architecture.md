# アーキテクチャ仕様書

## システム概要

tig-ghは、クリーンアーキテクチャの原則に基づいて設計されたTUI（Text User Interface）アプリケーションです。

## アーキテクチャ図

```
┌─────────────────────────────────────────────────────────┐
│                         UI Layer                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │         bubbletea (Model-Update-View)            │  │
│  │  ┌────────┐  ┌────────┐  ┌──────────┐          │  │
│  │  │ Issue  │  │   PR   │  │  Commit  │  ...     │  │
│  │  │  View  │  │  View  │  │   View   │          │  │
│  │  └────────┘  └────────┘  └──────────┘          │  │
│  │                                                  │  │
│  │  ┌─────────────────────────────────────────┐   │  │
│  │  │      Shared Components                   │   │  │
│  │  │  (List, Detail, Input, Modal, etc.)      │   │  │
│  │  └─────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↓↑
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │                Use Cases                          │  │
│  │  - FetchIssues    - CreateIssue                  │  │
│  │  - FetchPRs       - MergePR                      │  │
│  │  - FetchCommits   - UpdateIssue                  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↓↑
┌─────────────────────────────────────────────────────────┐
│                     Domain Layer                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Domain Models                        │  │
│  │  - Issue      - PullRequest   - Commit           │  │
│  │  - Repository - User          - Comment          │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Repository Interfaces                   │  │
│  │  - IssueRepository                               │  │
│  │  - PullRequestRepository                         │  │
│  │  - CommitRepository                              │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↓↑
┌─────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │              GitHub API Client                    │  │
│  │  - REST API (go-github)                          │  │
│  │  - GraphQL API (future)                          │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │                  Cache Layer                      │  │
│  │  - Memory Cache (in-memory map)                  │  │
│  │  - File Cache   (~/.cache/tig-gh/)               │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Configuration                        │  │
│  │  - Config File  (~/.config/tig-gh/config.yaml)   │  │
│  │  - Environment Variables                         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## レイヤー詳細

### 1. UI Layer (`internal/ui`)

**責務**: ユーザーインターフェースの表示とユーザー入力の処理

**構成要素**:

#### Views (`internal/ui/views`)
各ビューは独立したbubbletea Modelとして実装
- `issue_view.go`: Issue一覧・詳細ビュー
- `pr_view.go`: Pull Request一覧・詳細ビュー
- `commit_view.go`: コミット履歴ビュー
- `repo_view.go`: リポジトリ一覧ビュー
- `notification_view.go`: 通知ビュー

#### Components (`internal/ui/components`)
再利用可能なUIコンポーネント
- `list.go`: スクロール可能なリストコンポーネント
- `detail.go`: 詳細表示コンポーネント
- `input.go`: 入力フォームコンポーネント
- `modal.go`: モーダルダイアログ
- `statusbar.go`: ステータスバー
- `help.go`: ヘルプ表示

#### Styles (`internal/ui/styles`)
- `theme.go`: カラーテーマ定義
- `layout.go`: レイアウト定義

**データフロー**:
```
User Input → Update() → Use Case → Update Model State → View()
```

### 2. Application Layer (`internal/app`)

**責務**: ビジネスロジックのオーケストレーション

**構成要素**:

#### Use Cases
- `FetchIssuesUseCase`: Issue一覧の取得
- `CreateIssueUseCase`: Issue作成
- `UpdateIssueUseCase`: Issue更新
- `FetchPullRequestsUseCase`: PR一覧取得
- `MergePullRequestUseCase`: PRマージ
- `FetchCommitsUseCase`: コミット履歴取得

**実装パターン**:
```go
type FetchIssuesUseCase struct {
    repo IssueRepository
    cache CacheService
}

func (uc *FetchIssuesUseCase) Execute(ctx context.Context, owner, repo string, opts *IssueOptions) ([]*Issue, error) {
    // 1. キャッシュチェック
    if cached := uc.cache.Get(cacheKey); cached != nil {
        return cached, nil
    }

    // 2. リポジトリから取得
    issues, err := uc.repo.List(ctx, owner, repo, opts)
    if err != nil {
        return nil, err
    }

    // 3. キャッシュに保存
    uc.cache.Set(cacheKey, issues, ttl)

    return issues, nil
}
```

### 3. Domain Layer (`internal/domain`)

**責務**: ビジネスルールとドメインモデルの定義

#### Models (`internal/domain/models`)
```go
type Issue struct {
    ID          int64
    Number      int
    Title       string
    Body        string
    State       IssueState
    Author      User
    Assignees   []User
    Labels      []Label
    Milestone   *Milestone
    Comments    int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type PullRequest struct {
    ID          int64
    Number      int
    Title       string
    Body        string
    State       PRState
    Author      User
    Head        Branch
    Base        Branch
    Mergeable   bool
    Reviews     []Review
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Commit struct {
    SHA         string
    Message     string
    Author      CommitAuthor
    Committer   CommitAuthor
    Parents     []string
    CreatedAt   time.Time
}
```

#### Repository Interfaces (`internal/domain/repository`)
```go
type IssueRepository interface {
    List(ctx context.Context, owner, repo string, opts *IssueOptions) ([]*Issue, error)
    Get(ctx context.Context, owner, repo string, number int) (*Issue, error)
    Create(ctx context.Context, owner, repo string, input *CreateIssueInput) (*Issue, error)
    Update(ctx context.Context, owner, repo string, number int, input *UpdateIssueInput) (*Issue, error)
    Close(ctx context.Context, owner, repo string, number int) error
}

type PullRequestRepository interface {
    List(ctx context.Context, owner, repo string, opts *PROptions) ([]*PullRequest, error)
    Get(ctx context.Context, owner, repo string, number int) (*PullRequest, error)
    Merge(ctx context.Context, owner, repo string, number int, opts *MergeOptions) error
    GetDiff(ctx context.Context, owner, repo string, number int) (string, error)
}

type CommitRepository interface {
    List(ctx context.Context, owner, repo string, opts *CommitOptions) ([]*Commit, error)
    Get(ctx context.Context, owner, repo string, sha string) (*Commit, error)
    Compare(ctx context.Context, owner, repo, base, head string) (*Comparison, error)
}
```

### 4. Infrastructure Layer (`internal/infra`)

**責務**: 外部システムとの連携、永続化

#### GitHub API Client (`internal/infra/github`)
```go
type GitHubClient struct {
    client *github.Client
    auth   AuthService
}

// Repository Interfaceの実装
type IssueRepositoryImpl struct {
    client *GitHubClient
}

func (r *IssueRepositoryImpl) List(ctx context.Context, owner, repo string, opts *IssueOptions) ([]*Issue, error) {
    ghIssues, _, err := r.client.Issues.ListByRepo(ctx, owner, repo, convertOptions(opts))
    if err != nil {
        return nil, err
    }
    return convertToIssues(ghIssues), nil
}
```

#### Cache (`internal/infra/cache`)
```go
type CacheService interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(key string)
    Clear()
}

// メモリキャッシュ実装
type MemoryCache struct {
    data map[string]*cacheEntry
    mu   sync.RWMutex
}

// ファイルキャッシュ実装
type FileCache struct {
    dir string
}
```

#### Config (`internal/infra/config`)
```go
type Config struct {
    GitHub   GitHubConfig
    UI       UIConfig
    Cache    CacheConfig
}

type GitHubConfig struct {
    Token          string
    DefaultOwner   string
    DefaultRepo    string
}

type UIConfig struct {
    Theme          string
    KeyBindings    map[string]string
    DefaultView    string
}
```

## データフロー例: Issue一覧の表示

```
1. ユーザーがIssueビューを開く
   ↓
2. IssueView.Init() が FetchIssuesMsg を送信
   ↓
3. IssueView.Update() が FetchIssuesUseCase を実行
   ↓
4. UseCase がキャッシュをチェック
   ↓
5. キャッシュミスの場合、IssueRepository.List() を呼び出し
   ↓
6. GitHubClient が GitHub API を呼び出し
   ↓
7. レスポンスをDomain Modelに変換
   ↓
8. キャッシュに保存
   ↓
9. UseCase が結果を返す
   ↓
10. IssueView が Model を更新
   ↓
11. IssueView.View() が再描画
```

## エラーハンドリング戦略

### エラー型定義
```go
type ErrorType int

const (
    ErrTypeAuth ErrorType = iota
    ErrTypeNetwork
    ErrTypeRateLimit
    ErrTypeNotFound
    ErrTypeValidation
    ErrTypeInternal
)

type AppError struct {
    Type    ErrorType
    Message string
    Err     error
}
```

### エラーハンドリングフロー
1. Infrastructure層: API固有のエラーをAppErrorに変換
2. Application層: ビジネスロジックエラーを追加
3. UI層: ユーザーフレンドリーなメッセージを表示

## 並行処理モデル

### API呼び出しの並列化
```go
func (uc *FetchMultipleResourcesUseCase) Execute(ctx context.Context) (*Result, error) {
    var wg sync.WaitGroup
    errCh := make(chan error, 3)

    // Issue、PR、Commitを並列取得
    wg.Add(3)

    go func() {
        defer wg.Done()
        issues, err := uc.issueRepo.List(ctx, ...)
        if err != nil {
            errCh <- err
            return
        }
        // 処理
    }()

    // ... 他のリソースも同様

    wg.Wait()
    close(errCh)

    // エラー集約
    for err := range errCh {
        if err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

## テスト戦略

### 各レイヤーのテスト方針

1. **Domain Layer**: ユニットテスト（100%カバレッジ目標）
2. **Application Layer**: ユニットテスト + モックを使った統合テスト
3. **Infrastructure Layer**: 統合テスト（GitHub API Mockサーバー使用）
4. **UI Layer**: スナップショットテスト + E2Eテスト

### モック生成
```bash
# mockgenを使用
mockgen -source=internal/domain/repository/issue_repository.go -destination=internal/mock/issue_repository_mock.go
```

## セキュリティ考慮事項

1. **認証情報の管理**
   - トークンはメモリ上のみで保持
   - 設定ファイルには暗号化して保存
   - 環境変数からの読み込みをサポート

2. **入力検証**
   - すべてのユーザー入力をバリデーション
   - SQLインジェクション対策（該当なし）
   - XSS対策（TUIのため影響小）

3. **レート制限対応**
   - GitHub API レート制限の監視
   - リトライロジックの実装
   - キャッシュによるAPI呼び出し削減
