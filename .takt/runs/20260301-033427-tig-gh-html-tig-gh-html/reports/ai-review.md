# AI生成コードレビュー

## 結果: REJECT

## サマリー
formatDurationの丸め方式の不一致（Truncate vs Round）によるTUI版との動作差異、ユーティリティ関数・変数の重複（DRY違反）、未使用構造体フィールド、変更ファイル内の未使用コードの計3件のブロッキング問題を検出。

## 検証した項目
| 観点 | 結果 | 備考 |
|------|------|------|
| 仮定の妥当性 | ✅ | タスク要件の8セクション全て正しく実装 |
| API/ライブラリの実在 | ✅ | 使用モデル・メソッド全て実在確認済み |
| コンテキスト適合 | ❌ | formatDuration の丸め方式がTUI版と不一致、関数・変数の重複 |
| スコープ | ✅ | スコープクリープ・縮小なし |
| デッドコード | ❌ | 未使用フィールド、変更ファイル内の未使用変数 |

## 今回の指摘（new）
| # | finding_id | カテゴリ | 場所 | 問題 | 修正案 |
|---|------------|---------|------|------|--------|
| 1 | AIAR-001 | DRY違反+動作不整合 | `internal/app/export/format.go:14` | `Truncate`使用でTUI版(`Round`)と出力が異なる（30s→"0s" vs "1m"）。formatDuration, shortWeekday, weekdayDisplayOrder, weekdaysOnlyが2パッケージで重複 | 共通パッケージ（例:`internal/domain/format/`）に抽出しエクスポート。Roundに統一 |
| 2 | AIAR-002 | 未使用コード | `internal/app/export/sections.go:79,286` | `qualityIssueEntry.Severity`フィールドが代入されるがtemplate.htmlで未参照。HighPriority/MediumPriority分離で冗長 | フィールドと代入を削除 |
| 3 | AIAR-003 | 未使用コード(Boy Scout) | `cmd/tig-gh/main.go:214-215` | `ctx := context.Background(); _ = ctx`が「将来使う」コメント付きで放置。変更対象ファイル内の既存問題 | 2行を削除 |

## 継続指摘（persists）
なし（初回レビュー）

## 解消済み（resolved）
なし（初回レビュー）

## REJECT判定条件
- `new` が3件あるため REJECT