## 作業結果
- レビュアーから指摘された5件のブロッキング問題（AIAR-001, AIAR-002, AIAR-003, ARCH-001, QA-001/TR-001）を全て修正完了

## 変更内容
1. **AIAR-001 (DRY違反+動作不整合)**: `internal/shared/metricsformat/format.go` を新規作成し、`FormatDuration`（`Round`に統一）、`ShortWeekday`、`WeekdayDisplayOrder`、`WeekdaysOnly`、`MaxQualityIssuesToDisplay` をエクスポート。`internal/app/export/format.go` と `internal/ui/views/metrics_view.go` から重複定義を削除し、共通パッケージを直接参照するよう変更。
2. **AIAR-002 (未使用フィールド)**: `qualityIssueEntry` 構造体から `Severity` フィールドを削除し、`sections.go:286` の代入も削除。
3. **AIAR-003 (未使用コード)**: `cmd/tig-gh/main.go` から `ctx := context.Background(); _ = ctx` の2行を削除（`context` importは `runMetricsHTMLExport` で使用のため残存）。
4. **ARCH-001 (ファイルサイズ超過)**: `sections.go` の型定義（11構造体）を `types.go` に分離し、`sections.go` を348行→224行に削減。
5. **QA-001/TR-001 (テストカバレッジ)**: `ShowWeekendInDayOfWeek` の `true`/`false` 両パスのテストを追加（土日データ含むフィクスチャで `Sat`/`Sun` の表示/非表示を検証）。
6. **テスト移動**: `FormatDuration` のテストを `internal/shared/metricsformat/format_test.go` に移動し、`Round` に合わせて `30s→"1m"` の期待値を修正。`format_test.go` は `formatChangePercent` のテストに変更。

## ビルド結果
- `go build ./...` → **成功**（型エラーなし）

## テスト結果
- `go test ./...` → **全パッケージ PASS**
  - `internal/app/export`: 30テスト PASS（新規2件含む）
  - `internal/shared/metricsformat`: 19テスト PASS（新規パッケージ）
  - `internal/ui/views`: PASS（共通パッケージ利用で正常動作確認）
  - 他全パッケージも PASS

## 証拠
- **AIAR-001解消**: `internal/shared/metricsformat/format.go` を作成、`format.go` から `formatDuration`/`shortWeekday` 削除、`metrics_view.go` から定義削除＆全呼び出しを `metricsformat.FormatDuration`/`metricsformat.ShortWeekday` に置換
- **AIAR-002解消**: `types.go` の `qualityIssueEntry` に `Severity` フィールドなし、`sections.go` の `toEntries` 関数から `Severity` 代入なし
- **AIAR-003解消**: `main.go` の行213-215に `ctx` 関連コードなし
- **ARCH-001解消**: `wc -l sections.go` → 224行（< 300行）
- **QA-001/TR-001解消**: `TestGenerateHTML_設定_ShowWeekendInDayOfWeekがfalseの場合土日が非表示` と `TestGenerateHTML_設定_ShowWeekendInDayOfWeekがtrueの場合土日が表示` が PASS