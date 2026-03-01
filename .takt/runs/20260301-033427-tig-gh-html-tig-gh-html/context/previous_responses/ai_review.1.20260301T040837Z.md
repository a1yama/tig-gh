# AI Antipattern Review Report

## 判定: REJECT

ブロッキング問題が3件検出されました。

---

## ブロッキング問題

### AIAR-001 [new] — DRY違反 + 動作不整合: formatDuration の Truncate vs Round

**ファイル**: `internal/app/export/format.go:14` / `internal/ui/views/metrics_view.go:1316`

**問題**:
export版の `formatDuration` は `d.Truncate(time.Minute)` を使用しているが、TUI版は `d.Round(time.Minute)` を使用している。同じデータに対して異なる出力を生成する。

例: `30 * time.Second` の場合
- export版: `Truncate` → 0 → `"0s"`
- TUI版: `Round` → 1min → `"1m"`

タスク要件に「TUI版と同等の情報が表示されること」とあり、同じデータで異なる表示は要件違反。

さらに、以下の関数・変数が2パッケージ間で完全に重複している（DRY違反）:

| 重複対象 | export側 | TUI側 |
|----------|----------|-------|
| `formatDuration` | `format.go:9` | `metrics_view.go:1311` |
| `shortWeekday` | `format.go:39` | `metrics_view.go:1367` |
| `weekdayDisplayOrder` | `sections.go:118` | `metrics_view.go:1341` |
| `weekdaysOnly` | `sections.go:128` | `metrics_view.go:1351` |

`formatChangePercent` は TUI版がlipglossスタイルを適用するため厳密には異なるが、フォーマットロジック `fmt.Sprintf("%+.1f%%", value)` は同一。

**修正案**:
共通ユーティリティパッケージ（例: `internal/shared/format/` または `internal/domain/format/`）を作成し、`FormatDuration`, `ShortWeekday`, `WeekdayDisplayOrder`, `WeekdaysOnly` をエクスポートして両パッケージから参照する。`formatDuration` は `Round` に統一する（TUI版と同じ動作）。

---

### AIAR-002 [new] — 未使用構造体フィールド: qualityIssueEntry.Severity

**ファイル**: `internal/app/export/sections.go:79`, `sections.go:286`

**問題**:
`qualityIssueEntry` 構造体の `Severity` フィールドが `sections.go:286` で代入されているが、`template.html` では一切参照されていない。品質問題はすでに `HighPriority` / `MediumPriority` スライスに分離されており、severity情報はセクションの分類で表現済み。`Severity` フィールドは「念のため」のデッドデータ。

```go
// sections.go:79 — 定義されているが template.html で未参照
type qualityIssueEntry struct {
    Repository string
    Number     int
    Title      string
    IssueType  string
    Severity   string  // ← 未使用
    Details    string
}
```

**修正案**:
`qualityIssueEntry` から `Severity` フィールドを削除し、`sections.go:286` の代入も削除する。

---

### AIAR-003 [new] — 変更ファイル内の未使用コード（ボーイスカウトルール）

**ファイル**: `cmd/tig-gh/main.go:214-215`

**問題**:
```go
ctx := context.Background()
_ = ctx // 将来的にコンテキストを使う
```

変数 `ctx` が作成され即座に `_ = ctx` で捨てられている。「将来的にコンテキストを使う」コメント付きの未使用コード。`cmd/tig-gh/main.go` は今回変更対象ファイルであるため、ボーイスカウトルールにより修正対象。

**修正案**:
この2行を削除する。

---

## 警告（非ブロッキング）

### AIAR-W01 — severity比較の大文字小文字処理の不一致

**ファイル**: `internal/app/export/sections.go:256` / `internal/ui/views/metrics_view.go:1734`

**問題**:
export版は `issue.Severity == "high"`（大文字小文字を区別）を使用しているが、TUI版は `strings.EqualFold(issue.Severity, "high")`（大文字小文字を無視）を使用している。現在のデータソース（`metrics_repo_impl.go:833`）は常に小文字の "high" / "medium" を生成するため実害はないが、防御的プログラミングの一貫性が欠けている。

**推奨**: export版も `strings.EqualFold` を使用してTUI版と一致させる。

---

## 確認済み・問題なし

| 項目 | 結果 |
|------|------|
| 幻覚API・存在しないメソッド | なし — 使用するモデル・API全て実在 |
| スコープクリープ | なし — タスク要件の8セクション全てに対応、余分な機能追加なし |
| スコープ縮小 | なし — 8セクション全て実装、MetricsConfig全設定項目を反映 |
| XSS防止 | `html/template` による自動エスケープ |
| テンプレート埋め込み | `go:embed` の正しい使用 |
| エラーハンドリング | nil引数チェック、テンプレートエラー、ファイル書き込みエラー全て処理 |
| TODOコメント | なし |
| フォールバック値の濫用 | なし |
| ユースケース再利用 | `FetchLeadTimeMetricsUseCase` を正しく再利用 |
| CLIフラグ設計 | `--metrics-html` / `--metrics-html=path` の設計は適切 |
| テストカバレッジ | 35テスト、主要セクション・設定フラグ・XSS・空データをカバー |