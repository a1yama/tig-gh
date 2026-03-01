# タスク完了サマリー

## タスク
tig-ghのメトリクスビュー（マトリクスビュー）の内容をHTMLファイルとして生成し、ブラウザで閲覧できるようにする。CLIコマンド `--metrics-html` でHTML出力を実行可能にし、既存の `MetricsConfig` 設定をHTML出力にも反映する。

## 結果
完了

## 変更内容
| 種別 | ファイル | 概要 |
|------|---------|------|
| 作成 | `internal/app/export/html.go` | 公開API `GenerateHTML` — メトリクスデータをスタンドアロンHTMLとして生成 |
| 作成 | `internal/app/export/sections.go` | 8セクション（全体統計、PR一覧、フェーズ分解、曜日別、週次比較、品質問題、滞留PR、リポジトリ統計）のデータビルダー |
| 作成 | `internal/app/export/types.go` | テンプレートデータ型定義 |
| 作成 | `internal/app/export/format.go` | `formatChangePercent` ヘルパー関数 |
| 作成 | `internal/app/export/template.go` | `go:embed` によるHTMLテンプレート埋め込み |
| 作成 | `internal/app/export/template.html` | スタイル込みスタンドアロンHTMLテンプレート（外部CSS/JS依存なし） |
| 作成 | `internal/app/export/html_test.go` | GenerateHTMLのテスト（30テスト関数） |
| 作成 | `internal/app/export/format_test.go` | formatChangePercentのテスト |
| 作成 | `internal/shared/metricsformat/format.go` | TUI/HTML共通のフォーマット関数（FormatDuration, ShortWeekday等） |
| 作成 | `internal/shared/metricsformat/format_test.go` | 共通フォーマット関数のテスト（19サブテスト） |
| 変更 | `cmd/tig-gh/main.go` | `--metrics-html` / `--metrics-html=path` フラグ追加、HTML出力分岐 |
| 変更 | `internal/ui/views/metrics_view.go` | ローカル関数を `metricsformat` 共通パッケージ参照に変更（DRY改善） |

## 確認コマンド
```bash
go test ./...
go build ./...
```
