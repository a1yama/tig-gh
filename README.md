# tig-gh

A tig-like GitHub management tool with a powerful TUI (Text User Interface).

## 概要

tig-ghは、tigのような直感的なインターフェースでGitHubを管理できるTUIツールです。Issue、Pull Request、コミット履歴などをターミナルから快適に操作できます。

## 特徴

- tigライクなキーバインディング
- Issue管理（一覧・詳細・作成・編集）
- Pull Request管理（一覧・詳細・レビュー・マージ）
- コミット履歴の表示
- 高速な検索・フィルタリング
- カスタマイズ可能なテーマ
- オフライン対応（キャッシュ機能）

## インストール

### go installから

```bash
go install github.com/a1yama/tig-gh/cmd/tig-gh@latest
```

### ソースからビルド

```bash
git clone https://github.com/a1yama/tig-gh.git
cd tig-gh
make build
sudo make install
```

## セットアップ

### GitHub認証

#### Personal Access Tokenを使用

```bash
# トークンを作成（GitHub Settings > Developer settings > Personal access tokens）
# 必要なスコープ: repo, read:org, read:user

# 環境変数に設定
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

#### gh CLIを使用

```bash
# GitHub CLIで認証
gh auth login

# tig-ghは自動的にgh CLIの認証情報を使用
tig-gh
```

### 設定ファイル

設定ファイルは `~/.config/tig-gh/config.yaml` に配置します。

```yaml
github:
  token: ghp_xxxxxxxxxxxx  # または環境変数 GITHUB_TOKEN
  default_owner: your-username
  default_repo: your-repo

ui:
  theme: dark  # dark / light / custom
  default_view: issues

keybindings:
  quit: q
  refresh: r
  search: /
```

## 使い方

### 基本操作

```bash
# 起動
tig-gh

# 特定のリポジトリを指定
tig-gh owner/repo

# Issueビューから起動
tig-gh --view issues owner/repo
```

### キーバインディング

#### グローバル
- `q`: 終了 / 前の画面に戻る
- `?`: ヘルプ表示
- `r`: リフレッシュ
- `/`: 検索
- `1-9`: ビュー切り替え

#### ナビゲーション
- `j` / `↓`: 下に移動
- `k` / `↑`: 上に移動
- `g`: 先頭に移動
- `G`: 末尾に移動
- `Ctrl+D`: 半ページ下
- `Ctrl+U`: 半ページ上

#### アクション
- `Enter`: 選択 / 詳細表示
- `o`: ブラウザで開く
- `n`: 新規作成
- `e`: 編集
- `c`: クローズ/再オープン

## 開発

詳細は[開発ガイド](docs/development.md)を参照してください。

### 環境構築

```bash
# 依存関係のインストール
make deps

# ビルド
make build

# テスト
make test

# 実行
make run
```

### ドキュメント

- [アーキテクチャ仕様書](docs/architecture.md)
- [機能仕様書](docs/features.md)
- [開発ガイド](docs/development.md)
- [プロジェクト方針](CLAUDE.md)

## ライセンス

MIT License

## 貢献

プルリクエストやIssueは大歓迎です！

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'feat: add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 謝辞

- [tig](https://jonas.github.io/tig/) - インスピレーション元
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUIフレームワーク
- [lipgloss](https://github.com/charmbracelet/lipgloss) - スタイリング
- [go-github](https://github.com/google/go-github) - GitHub API クライアント
