# CLAUDE.md - tig-gh Project

## プロジェクト概要

tig-ghは、tigライクなインターフェースを持つGitHub管理TUIツールです。
Go言語で実装され、高速で直感的なGitHub操作を提供します。

## 技術スタック

- **言語**: Go 1.21+
- **TUIフレームワーク**: bubbletea + lipgloss
- **GitHub API**: go-github (REST API v3)
- **設定管理**: viper (YAML/TOML)
- **認証**: GitHub Personal Access Token / gh CLI連携

## ディレクトリ構成

```
tig-gh/
├── cmd/                    # コマンドラインエントリーポイント
│   └── tig-gh/
│       └── main.go
├── internal/              # 内部パッケージ
│   ├── api/              # GitHub API クライアント
│   ├── config/           # 設定管理
│   ├── ui/               # TUI コンポーネント
│   │   ├── components/   # 再利用可能なUIコンポーネント
│   │   ├── views/        # 各ビュー（Issue, PR, Commitなど）
│   │   └── styles/       # スタイル定義
│   ├── models/           # データモデル
│   └── cache/            # キャッシュ層
├── pkg/                   # 公開パッケージ
├── docs/                  # ドキュメント
├── config/                # デフォルト設定ファイル
├── .github/               # GitHub Actions
└── README.md

```

## 開発ガイドライン

### コーディング規約

1. **パッケージ設計**
   - `internal/`配下は内部実装専用
   - `pkg/`配下は外部から利用可能な公開API
   - 循環依存を避ける

2. **命名規則**
   - インターフェース名: `xxxer` または `xxxService`
   - モデル: 単数形 (`Issue`, `PullRequest`)
   - コレクション: 複数形 (`Issues`, `PullRequests`)

3. **エラーハンドリング**
   - エラーは常に明示的に処理
   - カスタムエラー型を使用（`errors` パッケージ）
   - ユーザーフレンドリーなエラーメッセージ

4. **テスト**
   - ユニットテストは必須
   - テーブル駆動テストを推奨
   - モックは`gomock`または`testify/mock`を使用

### bubbletea設計パターン

1. **Model-Update-View パターン**
   ```go
   type Model struct {
       // state
   }

   func (m Model) Init() tea.Cmd
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
   func (m Model) View() string
   ```

2. **コンポーネント分割**
   - 各ビューは独立したModelを持つ
   - 共通コンポーネントは`ui/components`に配置
   - メッセージ型は明確に定義

3. **状態管理**
   - アプリケーション全体の状態は最上位Modelで管理
   - 各ビューは自身の状態のみを管理
   - 状態の変更はUpdateメソッド内のみで行う

### GitHub API 使用方針

1. **レート制限対策**
   - キャッシュ層を活用
   - GraphQL APIを優先検討（複数リソースの一括取得）
   - エクスポネンシャルバックオフでリトライ

2. **認証**
   - Personal Access Token（PAT）を優先
   - gh CLIの認証情報も利用可能に
   - 環境変数 `GITHUB_TOKEN` をサポート

3. **ペジネーション**
   - すべてのリスト取得は自動ページング
   - 大量データは遅延ロード

### パフォーマンス考慮事項

1. **キャッシュ戦略**
   - メモリキャッシュ（短期）
   - ファイルキャッシュ（長期）
   - TTLベースの無効化

2. **並行処理**
   - API呼び出しはゴルーチンで並列化
   - チャネルでの安全な通信
   - コンテキストによるキャンセル処理

3. **描画最適化**
   - 不必要な再描画を避ける
   - 大量データは仮想スクロール

## 実装優先順位

### Phase 1: MVP
1. 基本的なTUIフレームワーク
2. GitHub認証
3. Issueビュー（一覧・詳細）
4. 基本的なキーバインディング

### Phase 2: コア機能
1. Pull Requestビュー
2. コミット履歴ビュー
3. 検索・フィルタリング
4. キャッシュ層

### Phase 3: 高度な機能
1. ブランチ管理
2. 通知機能
3. GitHub Actions統合
4. 設定のカスタマイズ

## デバッグ・ログ

- ログは`log/slog`を使用
- デバッグモード: `--debug` フラグ
- ログファイル: `~/.config/tig-gh/debug.log`

## CI/CD

- GitHub Actionsでビルド・テスト
- リリース時のバイナリ自動生成（goreleaser）
- クロスプラットフォームビルド（macOS, Linux, Windows）

## 参考リソース

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [lipgloss](https://github.com/charmbracelet/lipgloss)
- [go-github](https://github.com/google/go-github)
- [tig](https://jonas.github.io/tig/)
