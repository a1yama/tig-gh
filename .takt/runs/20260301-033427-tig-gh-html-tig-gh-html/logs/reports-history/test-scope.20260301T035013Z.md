# 変更スコープ宣言

## タスク
メトリクスビューのHTML出力機能に対するテストを作成する

## 変更予定
| 種別 | ファイル |
|------|---------|
| 作成 | `internal/app/export/format_test.go` |
| 作成 | `internal/app/export/html_test.go` |

## 推定規模
Medium

## 影響範囲
- `internal/app/export/` パッケージ（新規作成予定）
- テストは `GenerateHTML()` 公開API と `formatDuration` 内部ヘルパーの振る舞いを検証
- 既存コードへの変更なし
