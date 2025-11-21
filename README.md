# tig-gh

A tig-like GitHub management tool with a powerful TUI (Text User Interface).

## 概要

tig-ghは、tigのような直感的なインターフェースでGitHubを管理できるTUIツールです。Issue、Pull Request、コミット履歴などをターミナルから快適に操作できます。

## 特徴

- Issue / Pull Request / Commit の一覧と詳細を tig ライクな操作感で閲覧
- Issue・PR ビューでは Open / Closed / All を即座に切り替え、コメントやレビュー履歴も読み込める
- PR 詳細ビューには Overview / Files / Commits / Comments のタブとレビューサマリを表示
- `/` で呼び出す Search ビューからリポジトリ内の Issue / PR を横断検索
- Review Queue ビューで長時間オープンの PR や未承認 PR を経過時間付きで確認
- **Metrics ビューで複数リポジトリのリードタイムとレビュープロセスを可視化**
  - レビューフェーズ別の時間分析（作成→レビュー→承認→マージ）
  - 滞留PR統計（3日以上オープンなPRを自動検出）
  - リアルタイムアラート（レビュー待ち・マージ待ち・リードタイム急増）
  - リポジトリ別の詳細メトリクスと週次トレンド
  - プログレス表示でデータ取得状況をリアルタイム確認
  - GitHub APIレート制限をステータスバーで表示
- GitHub API 呼び出し結果をメモリ＋ファイルキャッシュし、再取得を高速化
- テーマや主要キーバインドを設定ファイルで調整可能

> 現時点では参照系機能にフォーカスしており、Issue 作成・編集や PR マージ/レビュー送信などの書き込み操作は UI からはまだ行えません。

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

Personal Access Token (Classic) もしくは Fine-grained PAT を使用します。必要なスコープは `repo`, `read:org`, `read:user` です。

```bash
# 例: 環境変数に設定
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

`GITHUB_TOKEN` が未設定の場合は、設定ファイルの `github.token` を参照します。gh CLI の資格情報連携はまだ実装されていません。

### 設定ファイル

tig-gh は以下の優先順位で設定ファイルを探索します。

1. `./.tig-gh/config.yaml`
2. `~/.config/tig-gh/config.yaml`（デフォルト）
3. `~/.tig-gh/config.yaml`
4. `/etc/tig-gh/config.yaml`

`config/default.yaml` をコピーして編集すると手早く始められます。

```yaml
github:
  token: ghp_xxxxxxxxxxxx
  default_owner: your-username
  default_repo: your-repo

  # メトリクス計測対象の複数リポジトリ（owner/repo形式）
  repositories:
    - your-org/repo1
    - your-org/repo2

metrics:
  enabled: true
  lead_time_enabled: true
  calculation_period: 720h  # 30日間（720h=30日, 2160h=90日）

ui:
  theme: dark  # dark / light / auto
  default_view: issues
  key_bindings:
    quit: q
    refresh: r
    search: /

cache:
  enabled: true
  ttl: 15m
  dir: ~/.cache/tig-gh
```

キャッシュはデフォルトで `~/.cache/tig-gh` に保存されます。TTL やファイルキャッシュの有効/無効は `cache` セクションで調整できます。

## 使い方

### 基本操作

```bash
# 現在のGitリポジトリで起動（origin から owner/repo を推測）
tig-gh

# 任意のリポジトリを明示指定
tig-gh owner/repo

# バージョンを表示
tig-gh --version
```

### ビュー切り替え

- `i`: Issues ビュー
- `p`: Pull Requests ビュー
- `c`: Commits ビュー
- `/`: Search ビュー（検索入力にフォーカス）
- `R`: Review Queue ビュー（Shift+R）
- `m`: Metrics ビュー（リードタイム・レビュープロセス分析）

### 主なキーバインディング

#### グローバル
- `q` / `ctrl+c`: 終了（詳細ビューでは前の画面に戻る）
- `?`: 現在のビュー専用ヘルプをトグル
- `r`: リストをリフレッシュ（Search ビューでは直前のクエリを再実行）
- `j` / `k` または `↓` / `↑`: リストを上下に移動
- `g` / `G`: 先頭 / 末尾にジャンプ
- `ctrl+u` / `ctrl+d`: 半ページ単位でスクロール（対応ビュー）
- `Enter`: 選択中アイテムの詳細ビューを開く

#### Issues / Pull Requests ビュー
- `f`: 表示対象を Open → Closed → All で循環
- 詳細ビュー内では `j` / `k` / `g` / `G` でスクロール、`o` でブラウザを開く
- PR 詳細ビューでは `1`〜`4` で Overview / Files / Commits / Comments の各タブを切り替え、レビューサマリやコメントを確認

#### Commits ビュー
- `Enter`: コミット詳細ビュー
- 詳細ビューでは `j` / `k` / `g` / `G` に加えて `ctrl+u` / `ctrl+d` でページング

#### Search ビュー
- 起動直後は検索入力がフォーカス済み。`Enter` で検索、`Esc` でフォーカス解除
- 入力フォーカス解除後は `j` / `k` で結果を移動し、`Enter` で対応する Issue / PR 詳細を開く
- `t`: 検索対象（Issues / Pull Requests / Both）を切り替え
- `s`: 状態フィルタ（Open / Closed / All）を切り替え

### Review Queueビュー

- オープン中の PR を作成日時の古い順に並べ、レビュー・承認までの経過を一目で把握
- 「Awaiting review」「Awaiting approval」「Approved」などのステータスとレビュー数（✓/✗/?）を表示
- `j` / `k` / `g` / `G` で一覧操作、`Enter` で詳細ビュー、`r` でリストとレビュー指標を再取得

### Metricsビュー

複数リポジトリのリードタイムとレビュープロセスを総合的に分析し、チームのパフォーマンス改善を支援します。

#### 表示内容

1. **⚠️ Alerts（アラート）**
   - レビュー待ち24時間超のPR数
   - 承認済みマージ待ち12時間超のPR数
   - リードタイム急増の警告（前週比40%以上）

2. **Overall Metrics（全体統計）**
   - 平均リードタイム
   - 中央値（外れ値の影響を受けにくい指標）
   - 総PR数（デフォルトブランチへのマージのみ）

3. **Review Process Breakdown（レビューフェーズ別）**
   - PR作成 → 最初のレビュー
   - 最初のレビュー → 承認
   - 承認 → マージ
   - 全体のリードタイム

4. **Stagnant PRs（滞留PR）**
   - 3日以上オープンなPR総数
   - 滞留PRの平均経過時間
   - 最も古い滞留PR（リポジトリ・PR番号・経過時間）

5. **Per Repository（リポジトリ別）**
   - 各リポジトリの平均・中央値・PR数

6. **Weekly Trend（週次トレンド）**
   - 日付範囲ごとの平均リードタイム推移
   - バーチャートで可視化

#### 操作

- `j` / `k`: 上下スクロール
- `g` / `G`: 先頭・末尾にジャンプ
- `r`: メトリクスを再取得（最新化）
- `l`: GitHub APIレート制限を即座に表示
- `q`: 前の画面に戻る

#### パフォーマンスとプログレス表示

複数リポジトリのメトリクス取得には時間がかかる場合があります。取得中は以下の情報がステータスバーに表示されます：

- **プログレス表示**: `Loading metrics... (12/35 repositories)` - 現在の取得状況
- **レート制限**: `API: 4850/5000 remaining` - GitHub APIの残りリクエスト数

多数のリポジトリを設定している場合、以下の対策を検討してください：
- 重要なリポジトリのみに絞る（5〜10個程度を推奨）
- `calculation_period` を短縮（例: `720h` → `168h` で7日間に短縮）

#### 活用例

**放置PR対策:**
- アラートで問題を即座に発見
- 滞留PR統計で最も古いPRを特定
- フェーズ別分析でボトルネックを特定

**リリース頻度向上:**
- レビューフェーズ別でどこが遅いかを確認
- 承認→マージの時間を短縮
- 週次トレンドで改善効果を測定

**チーム改善:**
- 定期的なアラート確認を習慣化
- フェーズ別の時間を目標設定に活用
- 滞留PRをチームで優先的に対応

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
