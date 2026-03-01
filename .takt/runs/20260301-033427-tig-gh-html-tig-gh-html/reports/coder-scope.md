# 変更スコープ宣言

## タスク
メトリクスビューのHTML出力機能を実装し、CLIコマンドで生成可能にする

## 変更予定
| 種別 | ファイル |
|------|---------|
| 作成 | `internal/app/export/format.go` |
| 作成 | `internal/app/export/sections.go` |
| 作成 | `internal/app/export/template.html` |
| 作成 | `internal/app/export/template.go` |
| 作成 | `internal/app/export/html.go` |
| 変更 | `cmd/tig-gh/main.go` |

## 推定規模
Medium

## 影響範囲
- `internal/app/export/` パッケージ（新規）: HTML生成ロジック全体
- `cmd/tig-gh/main.go`: `--metrics-html` フラグ追加によるCLI分岐
- 既存TUI機能への影響なし（既存テスト全パス確認済み）