# アーキテクチャレビュー

## 結果: REJECT

## サマリー
前回AI Reviewで指摘された3件（DRY違反・未使用フィールド・未使用コード）が全て未解消のまま残存。加えて `sections.go` が300行超（348行）でファイル分割基準に抵触。計4件のブロッキング問題によりREJECT。

## 確認した観点
- [x] 構造・設計
- [x] コード品質
- [x] 変更スコープ
- [x] テストカバレッジ
- [x] デッドコード
- [x] 呼び出しチェーン検証

## 今回の指摘（new）
| # | finding_id | スコープ | 場所 | 問題 | 修正案 |
|---|------------|---------|------|------|--------|
| 1 | ARCH-001 | スコープ内 | `internal/app/export/sections.go`（348行） | 300行超。型定義11構造体（~100行）+ビルダー関数7個（~212行）+定数/変数（~22行）が1ファイルに混在 | 型定義を `types.go` に分離し `sections.go` を~248行に削減 |

## 継続指摘（persists）
| # | finding_id | 前回根拠 | 今回根拠 | 問題 | 修正案 |
|---|------------|----------|----------|------|--------|
| 1 | AIAR-001 | `format.go:14` Truncate / `metrics_view.go:1316` Round | 同一箇所が未変更。さらに `maxQualityIssuesToDisplay` も `sections.go:116` と `metrics_view.go:955` で重複 | DRY違反+動作不整合: `formatDuration`,`shortWeekday`,`weekdayDisplayOrder`,`weekdaysOnly`,`maxQualityIssuesToDisplay` が2パッケージ間で重複 | 共通パッケージ `internal/shared/metricsformat/` を作成しエクスポート。`formatDuration` は `Round` に統一 |
| 2 | AIAR-002 | `sections.go:79` Severity定義, `template.html` 参照0件 | 同一箇所が未変更 | `qualityIssueEntry.Severity` がテンプレートで未参照の未使用フィールド | フィールドと `sections.go:286` の代入を削除 |
| 3 | AIAR-003 | `main.go:214-215` | 同一箇所が未変更 | `ctx := context.Background(); _ = ctx` 未使用コード+未Issue化TODOコメント | 2行を削除（`context` importは `runMetricsHTMLExport` L224で使用のため残す） |

## 解消済み（resolved）
なし

## REJECT判定条件
- `persists` 3件 + `new` 1件 = 計4件のブロッキング問題あり → REJECT