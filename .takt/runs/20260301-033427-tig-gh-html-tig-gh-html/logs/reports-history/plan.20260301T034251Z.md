# 計画レポート: メトリクスビューのHTML出力機能

## 要件分析

### タスク概要

`models.LeadTimeMetrics` のデータをHTMLとして出力する機能を実装する。既存のTUI表示・Markdownコピー機能とは別に、CLIからHTML生成を実行できるようにする。

### 要件の変更要/不要判定

| 要件 | 判定 | 根拠 |
|------|------|------|
| HTML生成ロジック | **変更要** | 新規機能。該当コードなし |
| CLIからのHTML出力コマンド | **変更要** | `cmd/tig-gh/main.go` に新フラグ追加が必要 |
| 表示設定の反映 | **変更要** | `MetricsConfig` のフラグをHTML生成でも参照する |

### 対応セクション一覧

`toMarkdown()` (`metrics_view.go:1417-1494`) が出力する全セクションに対応：

1. 全体リードタイム統計（`overallToMarkdown`: L1497-1525）
2. 個別PRリードタイム一覧テーブル（`prLeadTimesToMarkdown`: L1527-1558）
3. レビューフェーズ分解（`reviewPhasesToMarkdown`: L1560-1622）
4. 曜日ごとのマージ/レビュー件数（`dayOfWeekToMarkdown`: L1624-1672）
5. 今週と先週の比較（`weeklyComparisonToMarkdown`: L1674-1707）
6. PR品質問題一覧（`qualityIssuesToMarkdown`: L1709-1791）
7. 滞留PR一覧（`stagnantPRsToMarkdown`: L1793-1832）
8. リポジトリごとの統計（`repositoryStatsToMarkdown`: L1834-1874）

## 設計

### パッケージ構成

新規パッケージ `internal/app/export/` を作成する。

**配置の理由:** HTML生成はデータの表現（プレゼンテーション）だが、TUI（bubbletea）には依存しない。`internal/ui/` はTUI専用、`internal/infra/` はインフラ層なので、アプリケーション層の出力機能として `internal/app/export/` が適切。

```
internal/app/export/
├── html.go            — 公開API: GenerateHTML関数、テンプレートデータ構造体
├── sections.go        — 各セクションのデータビルダー関数
├── template.go        — go:embed でテンプレートファイルを埋め込み
├── template.html      — HTMLテンプレート（CSS埋め込み済み）
└── format.go          — formatDuration, shortWeekday ヘルパー
```

### データフロー

```
CLI (--metrics-html)
  → config.Load() + GitHub client 初期化（既存の初期化フロー再利用）
  → FetchLeadTimeMetricsUseCase.Execute()
  → export.GenerateHTML(metrics, config)
  → ファイル書き出し
```

### CLIコマンド設計

既存のフラグパース方式（`os.Args` ループ）に合わせて `--metrics-html` フラグを追加する。

```
tig-gh --metrics-html                  → metrics.html に出力
tig-gh --metrics-html=report.html      → 指定パスに出力
tig-gh owner/repo --metrics-html       → 指定リポジトリのメトリクスをHTML出力
```

`--metrics-html` が指定された場合、TUIは起動せずHTMLを生成して終了する。

### HTML生成方式

Go標準ライブラリの `html/template` を使用する。

**選択理由:**
- 自動HTMLエスケープによるXSS防止
- テンプレートとロジックの分離
- 外部依存なし（標準ライブラリのみ）

テンプレートは `go:embed` で `.html` ファイルを埋め込む。

### ヘルパー関数の扱い

`formatDuration`（`metrics_view.go:1311-1339`）と `shortWeekday`（L1367-1373）は `export` パッケージ内に独自に定義する。

**理由:** TUI版は将来ANSIカラーコードを含む可能性があり、プレゼンテーション層ごとにフォーマット関数を持つのが適切。現時点では同一ロジックだが、変更理由が異なるため共通化しない。

### MetricsConfig の反映方針

`MetricsConfig` の以下のフラグを参照し、`false` のセクションはHTMLに含めない：

| フラグ | セクション |
|--------|-----------|
| `ShowPRLeadTimes` | 個別PRリードタイム一覧 |
| `ShowReviewPhases` | レビューフェーズ分解 |
| `ShowDayOfWeek` | 曜日ごとの統計 |
| `ShowWeeklyComparison` | 週次比較 |
| `ShowQualityIssues` | PR品質問題 |
| `ShowStagnantPRs` | 滞留PR |
| `ShowRepositoryStats` | リポジトリ別統計 |
| `ShowWeekendInDayOfWeek` | 曜日表示で土日を含めるか |

全体リードタイム統計は常に表示（Markdown版と同一動作）。

## ファイル別実装方針

### 1. `internal/app/export/html.go`（~120行）

**責務:** 公開API、テンプレートデータ組み立て、テンプレート実行

```go
// GenerateHTML はメトリクスデータからスタンドアロンHTMLを生成する
func GenerateHTML(metrics *models.LeadTimeMetrics, cfg *models.MetricsConfig) ([]byte, error)
```

- `templateData` 構造体: テンプレートに渡す全データ
- `prepareTemplateData`: `LeadTimeMetrics` + `MetricsConfig` → `templateData` 変換
- `cfg` のフラグに基づき、非表示セクションは `nil` で渡す

### 2. `internal/app/export/sections.go`（~250行）

**責務:** 各セクション用のデータ構造体定義とビルダー関数

セクションごとにテンプレート用の構造体を定義：

```go
type overallData struct {
    Average string
    Median  string
    Count   int
}

type prLeadTimeRow struct {
    Rank       int
    LeadTime   string
    Repository string
    Number     int
    HTMLURL    string
    Title      string
    MergedDate string
}
// 以下、各セクション用の構造体とビルダー関数
```

**注意:** 300行超えのリスクあり。超える場合は `sections_tables.go`（テーブル系）と `sections_stats.go`（統計系）に分割する。

### 3. `internal/app/export/template.go`（~10行）

**責務:** テンプレートファイルの埋め込み

```go
//go:embed template.html
var templateHTML string
```

### 4. `internal/app/export/template.html`（~300-400行）

**責務:** HTMLテンプレート + 埋め込みCSS

- `<style>` タグ内に全CSSを含む（外部依存なし）
- レスポンシブデザイン
- テーブル、カード形式でデータ表示
- 各セクションは `{{if .SectionName}}...{{end}}` で条件付き表示
- PR番号にGitHub URLへのリンク付与

### 5. `internal/app/export/format.go`（~60行）

**責務:** フォーマットヘルパー

- `formatDuration(d time.Duration) string` — `metrics_view.go:1311-1339` と同一ロジック
- `shortWeekday(day time.Weekday) string` — `metrics_view.go:1367-1373` と同一ロジック
- 曜日順序の定数（`weekdayDisplayOrder`, `weekdaysOnly`）
- `formatChangePercent(value float64) string` — `+/-` 付き百分率
- テンプレート用FuncMap生成関数

### 6. `cmd/tig-gh/main.go`（修正）

**変更内容:**

1. `--metrics-html` フラグのパース追加（L29-37のループに追加）
2. HTML出力分岐の追加（L173付近、TUI起動前に分岐）

```go
if metricsHTMLPath != "" {
    // 進捗をstderrに出力
    metrics, err := fetchMetricsUseCase.Execute(ctx, progressFn)
    htmlBytes, err := export.GenerateHTML(metrics, &cfg.Metrics)
    os.WriteFile(metricsHTMLPath, htmlBytes, 0644)
    fmt.Fprintf(os.Stderr, "HTML metrics report generated: %s\n", metricsHTMLPath)
    return
}
```

**行数見積もり:** 現在217行 → 約250行。200行超だが300行以内。

## 影響範囲

### 変更ファイル

| ファイル | 変更種別 | 影響 |
|---------|---------|------|
| `cmd/tig-gh/main.go` | 修正 | フラグ追加 + HTML出力分岐 |

### 新規ファイル

| ファイル | 行数見積 |
|---------|---------|
| `internal/app/export/html.go` | ~120行 |
| `internal/app/export/sections.go` | ~250行 |
| `internal/app/export/template.go` | ~10行 |
| `internal/app/export/template.html` | ~350行 |
| `internal/app/export/format.go` | ~60行 |
| `internal/app/export/html_test.go` | テスト |

### 既存機能への影響

- TUI機能: **影響なし**（`internal/ui/views/metrics_view.go` は変更しない）
- ユースケース層: **影響なし**（`FetchLeadTimeMetricsUseCase` をそのまま再利用）
- データモデル: **影響なし**（`models.LeadTimeMetrics` は変更しない）

## テスト方針

1. `GenerateHTML` に完全なメトリクスデータを渡し、全セクションのHTML出力を検証
2. `MetricsConfig` の各フラグを `false` にし、該当セクションがHTMLに含まれないことを検証
3. nil/空のメトリクスデータでのグレースフルハンドリング
4. `formatDuration` のユニットテスト
5. HTMLの構造検証（基本的なHTML要素の存在チェック）

## 確認事項

なし（すべてコード調査で解決済み）
