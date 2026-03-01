# QAレビュー

## 結果: REJECT

## サマリー
前回のAI Antipatternレビューで指摘された3件のブロッキング問題（DRY違反・未使用フィールド・未使用コード）が全て未修正のまま残存しています。加えて、QA観点から `ShowWeekendInDayOfWeek` 設定の `true` パスにテストがない新規ブロッキング問題を1件検出しました。

## 確認した観点
| 観点 | 結果 | 備考 |
|------|------|------|
| テストカバレッジ | ⚠️ | 29テスト関数で主要パスをカバーするが、`ShowWeekendInDayOfWeek=true` のパスが未テスト |
| テスト品質 | ✅ | テーブル駆動テスト、ヘルパー関数、Given/When/Then構造で良好 |
| エラーハンドリング | ✅ | nil入力チェック、テンプレートエラー、ファイル書き込みエラー全て適切に処理 |
| ドキュメント | ✅ | 関数のGoDocコメントあり、テスト名が日本語で意図明確 |
| 保守性 | ❌ | formatDuration等4関数・変数がTUI版と完全重複（DRY違反）、未使用フィールドあり |

## 今回の指摘（new）
| # | finding_id | カテゴリ | 場所 | 問題 | 修正案 |
|---|------------|---------|------|------|--------|
| 1 | QA-001 | テストカバレッジ | `internal/app/export/sections.go:219` | `ShowWeekendInDayOfWeek = true` の条件分岐にテストがない。他の7設定フラグは全てテスト済みだが、このフラグのみ `true` パスが未検証。REJECT基準「テストがない新しい振る舞い」に該当 | 土日データを含むメトリクスで `ShowWeekendInDayOfWeek: true` を設定し、`"Sat"` `"Sun"` がHTMLに含まれることを検証するテストを追加する |

## 継続指摘（persists）
| # | finding_id | 前回根拠 | 今回根拠 | 問題 | 修正案 |
|---|------------|----------|----------|------|--------|
| 1 | AIAR-001 | `format.go:14` で `Truncate` 使用、`metrics_view.go:1316` で `Round` 使用 | `format.go:14` は依然 `d.Truncate(time.Minute)`、TUI版は `d.Round(time.Minute)` のまま。加えて `format_test.go:58` が `30s → "0s"` を期待値としており、TUI版の `Round` なら `"1m"` になるはず | 共通パッケージ（例: `internal/domain/format/`）に `FormatDuration`, `ShortWeekday`, `WeekdayDisplayOrder`, `WeekdaysOnly` をエクスポートし両パッケージから参照。`Round` に統一 |
| 2 | AIAR-002 | `sections.go:79` で `Severity` フィールド定義、`template.html` で未参照 | `sections.go:79` に `Severity string` が定義、`sections.go:286` で代入されるが、`template.html` で一切参照されていない | `qualityIssueEntry` から `Severity` フィールドを削除し、`sections.go:286` の代入も削除する |
| 3 | AIAR-003 | `cmd/tig-gh/main.go:214-215` で未使用 `ctx` | `cmd/tig-gh/main.go:214-215` に `ctx := context.Background()` / `_ = ctx // 将来的にコンテキストを使う` が残存 | この2行を削除する。`context` パッケージの import は `runMetricsHTMLExport` で使用されているため残す |

## 解消済み（resolved）
| finding_id | 解消根拠 |
|------------|----------|
| （なし） | 前回指摘は全て未修正 |

## REJECT判定条件
- `persists` が3件（AIAR-001, AIAR-002, AIAR-003）、`new` が1件（QA-001）で合計4件のブロッキング問題があるため REJECT