## 作業結果

`internal/app/export/` パッケージにHTML生成ロジックを実装し、`cmd/tig-gh/main.go` に `--metrics-html` CLIフラグを追加しました。計画に従い、5つの新規ファイルと1つの既存ファイルの修正を行いました。

- `GenerateHTML(metrics, config)` は純粋関数として実装（副作用なし）
- `html/template` + `go:embed` でXSS自動エスケープとテンプレート分離を実現
- `MetricsConfig` の全表示フラグ（7つ）をHTML出力に反映
- TUI版の `toMarkdown()` の全8セクションに対応するHTML出力を実装
- CLIから `--metrics-html` または `--metrics-html=path` で実行可能

## 変更内容
- `internal/app/export/format.go` — 新規作成（50行）: `formatDuration`, `shortWeekday`, `formatChangePercent` ヘルパー
- `internal/app/export/sections.go` — 新規作成（283行）: テンプレートデータ構造体と8セクション分のビルダー関数
- `internal/app/export/template.html` — 新規作成（183行）: 埋め込みCSS付きHTMLテンプレート
- `internal/app/export/template.go` — 新規作成（6行）: `go:embed` によるテンプレート埋め込み
- `internal/app/export/html.go` — 新規作成（56行）: `GenerateHTML` 公開API
- `cmd/tig-gh/main.go` — 変更: `--metrics-html` フラグ追加 + HTML出力分岐

## ビルド結果
- `go build ./...` — 成功 ✅（エラーなし）

## テスト結果
- `go test ./internal/app/export/... -v` — **35テスト全てパス** ✅
  - `TestFormatDuration`: 11テスト全パス
  - `TestGenerateHTML_*`: 24テスト全パス
- `go test ./...` — **全プロジェクトテストパス** ✅（既存テストへの影響なし）