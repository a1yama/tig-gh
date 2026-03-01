# 変更スコープ宣言

## タスク
メトリクスビューHTML出力機能（internal/app/export パッケージ）に対するテスト先行作成

## 変更予定
| 種別 | ファイル |
|------|---------|
| 作成 | `internal/app/export/format_test.go` |
| 作成 | `internal/app/export/html_test.go` |

## 推定規模
Medium

## 影響範囲
- `internal/app/export/` パッケージ（新規作成予定のHTML生成機能）
- テストは `GenerateHTML()` 公開APIと `formatDuration` 内部ヘルパーの振る舞いを検証
- 既存コードへの変更なし（`go build ./...` に影響なし）