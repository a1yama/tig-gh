# タスク計画

## 元の要求

tig-ghのメトリクスビュー（マトリクスビュー）の内容をHTMLファイルとして生成し、ブラウザで閲覧できるようにする。CLIからHTML生成を実行できるコマンドを追加し、既存の `FetchLeadTimeMetricsUseCase` を再利用してデータ取得→HTML生成→ファイル出力を行う。既存の `MetricsConfig` の表示設定もHTML出力に反映する。

## 分析結果

### 目的

1. `models.LeadTimeMetrics` をスタンドアロンHTML（外部CSS/JS依存なし）として出力する機能を新規作成する
2. CLIから `--metrics-html` フラグでHTML生成を実行できるようにする
3. `MetricsConfig` の `ShowPRLeadTimes`, `ShowReviewPhases` 等のフラグをHTML出力にも反映する

### 参照資料の調査結果

**`internal/ui/views/metrics_view.go`（1920行）**
- `toMarkdown()` (L1417-1494) が全セクションのMarkdown生成を統括。8つの個別セクション関数を呼び出す
- 各セクション関数は `MetricsView` 構造体に依存（`m.metrics`, `m.config`, `m.filteredRepo`）
- `formatDuration` (L1311-1339), `shortWeekday` (L1367-1373) はパッケージ非公開のヘルパー
- HTML生成に必要なデータは全て `models.LeadTimeMetrics` と `models.MetricsConfig` から取得可能。TUI状態（`filteredRepo`, `scroll` 等）はHTML出力では不要

**`internal/domain/models/metrics.go`（136行）**
- `LeadTimeMetrics` 構造体に全メトリクスデータが集約されている
- リポジトリ別データは `ByRepository`, `ByRepositoryPhaseBreakdown`, `ByRepositoryDayOfWeek`, `ByRepositoryWeekly` の各マップで保持

**`internal/domain/models/config.go`（246行）**
- `MetricsConfig` に8つの表示フラグ（`ShowPRLeadTimes`, `ShowReviewPhases`, `ShowDayOfWeek`, `ShowWeeklyComparison`, `ShowQualityIssues`, `ShowStagnantPRs`, `ShowRepositoryStats`, `ShowWeekendInDayOfWeek`）が定義済み

**`internal/app/usecase/fetch_lead_time_metrics.go`（116行）**
- `Execute(ctx, progressFn)` で `*models.LeadTimeMetrics` を返す。HTML出力でそのまま再利用可能
- `GetRateLimit(ctx)` も提供（HTML出力では使用しない）

**`cmd/tig-gh/main.go`（217行）**
- `os.Args` ループによる手動フラグパース。`--metrics` フラグで既にメトリクスモード分岐あり
- 設定ロード → トークン取得 → GitHub client初期化 → UseCase生成 → TUI起動の流れ

### スコープ

**新規作成: 6ファイル**
- `internal/app/export/html.go` — 公開API `GenerateHTML` + テンプレートデータ構造体
- `internal/app/export/sections.go` — 各セクションのデータビルダー関数（8セクション分）
- `internal/app/export/template.go` — `go:embed` によるテンプレート埋め込み
- `internal/app/export/template.html` — HTMLテンプレート + 埋め込みCSS
- `internal/app/export/format.go` — `formatDuration`, `shortWeekday` 等ヘルパー
- `internal/app/export/html_test.go` — テスト

**修正: 1ファイル**
- `cmd/tig-gh/main.go` — `--metrics-html` フラグ追加 + HTML出力分岐

**影響なし（変更しない）**
- `internal/ui/views/metrics_view.go` — TUI機能はそのまま
- `internal/app/usecase/fetch_lead_time_metrics.go` — そのまま再利用
- `internal/domain/models/` — データモデル変更なし

### 検討したアプローチ

| アプローチ | 採否 | 理由 |
|-----------|------|------|
| `toMarkdown()` をリファクタして共通化 | 不採用 | `MetricsView` のTUI状態（`filteredRepo`, `scroll`等）に依存しており、分離コストが高い。既存TUI機能への影響リスクもある |
| `internal/ui/export/` に配置 | 不採用 | `internal/ui/` はTUI（bubbletea）専用のレイヤー。HTML出力はTUIに依存しない |
| `internal/app/export/` に新規パッケージ | **採用** | アプリケーション層の出力機能として適切。ドメインモデルのみに依存し、TUI層・インフラ層に依存しない |
| 文字列結合でHTML生成 | 不採用 | XSSリスク、メンテナンス性の低さ |
| `html/template` + `go:embed` | **採用** | 自動HTMLエスケープ、テンプレートとロジックの分離、外部依存なし |
| `formatDuration` を共通パッケージに抽出 | 不採用 | TUI版は将来ANSIカラーコードを含む可能性があり、変更理由が異なる。現時点で25行の関数を共通化するためだけに新パッケージを作る必要はない |

### 実装アプローチ

**パッケージ `internal/app/export/`**

1. **`html.go`（~120行）**: 公開関数 `GenerateHTML(metrics *models.LeadTimeMetrics, cfg *models.MetricsConfig) ([]byte, error)` を定義。`templateData` 構造体を組み立て、`html/template` でレンダリング。`cfg` のフラグに基づき非表示セクションは `nil` で渡す

2. **`sections.go`（~250行）**: 8セクション分のテンプレート用データ構造体とビルダー関数。`models.LeadTimeMetrics` から表示用の文字列に変換する。300行を超える場合は `sections_tables.go` と `sections_stats.go` に分割

3. **`template.go`（~10行）**: `//go:embed template.html` でテンプレートファイルを埋め込む

4. **`template.html`（~350行）**: スタイル込みの単体HTMLテンプレート。`<style>` タグ内に全CSS。各セクションは `{{if .SectionName}}...{{end}}` で条件付き表示。PR番号にGitHub URLリンク付与

5. **`format.go`（~60行）**: `formatDuration`, `shortWeekday`, `formatChangePercent`, 曜日順序定数、テンプレートFuncMap生成

**CLI統合 `cmd/tig-gh/main.go`**

既存の `os.Args` ループに `--metrics-html` / `--metrics-html=path` のパース追加。フラグが指定された場合、既存の初期化フロー（config → token → client → usecase）を経た後、TUI起動の代わりに `export.GenerateHTML` → ファイル書き出し → 終了する分岐を追加。デフォルト出力先は `metrics.html`。

## 実装ガイドライン

- `GenerateHTML` は純粋関数として実装する（副作用なし、`LeadTimeMetrics` と `MetricsConfig` のみに依存）
- ファイル書き出しは `cmd/tig-gh/main.go` 側で行う（`export` パッケージはバイト列を返すのみ）
- `toMarkdown()` の8セクションに1対1で対応するHTML出力を実装する。漏れがないこと
- `maxQualityIssuesToDisplay` (定数値5) は `metrics_view.go:955` で定義されている。HTML出力でも同じ上限を適用する
- PRリードタイム一覧の表示上限は20件（`metrics_view.go:445-448`の `maxDisplay` と同一）
- 曜日表示順序は月曜始まり（`weekdayDisplayOrder`: Mon→Sun）。`ShowWeekendInDayOfWeek` が `false` なら月〜金のみ
- HTMLテンプレートのCSS はダークモード対応不要（シンプルなライトテーマで十分）
- `sections.go` が300行を超える場合はテーブル系セクション（PRリードタイム、品質問題、滞留PR、リポジトリ統計）と統計系セクション（全体、フェーズ、曜日、週次比較）に分割する
- 進捗表示は `stderr` に出力する（`progressFn` のコールバックで `fmt.Fprintf(os.Stderr, ...)` を使用）

## 確認事項

なし（すべてコード調査で解決済み）