# 決定ログ

## 1. テストファイルを2ファイル構成にする
- **背景**: 計画では5つのプロダクションファイル（html.go, sections.go, template.go, format.go, template.html）が作成される。テストファイルの粒度を決める必要があった
- **検討した選択肢**:
  - A) ファイル1対1（5テストファイル）
  - B) 責務別（format_test.go + html_test.go の2ファイル）
  - C) 単一テストファイル
- **理由**: sections.go と template.go は GenerateHTML() の内部実装であり、公開APIを通じてテストするのが適切。formatDuration は純粋関数として独立テスト可能。よって B) の2ファイル構成を選択

## 2. テストを同一パッケージ（package export）で記述する
- **背景**: formatDuration は非公開関数（小文字始まり）のため、外部テストパッケージからはアクセスできない
- **検討した選択肢**:
  - A) `package export`（同一パッケージ）
  - B) `package export_test`（外部テストパッケージ）
- **理由**: 既存プロジェクトのパターン（metrics_view_test.go は `package views`、fetch_lead_time_metrics_test.go は `package usecase`）に倣い、同一パッケージで記述。非公開ヘルパーも直接テスト可能

## 3. GenerateHTML の公開API設計
- **背景**: テストが定義するAPIコントラクトとして、GenerateHTML の引数を決定する必要があった
- **検討した選択肢**:
  - A) `GenerateHTML(metrics, config) (string, error)` — メトリクスとコンフィグを受け取りHTML文字列を返す
  - B) `GenerateHTML(metrics, config, writer) error` — io.Writer に書き込む
- **理由**: html/template の Execute は内部で io.Writer を使うが、呼び出し側（CLI）はファイルに書き込む前にパスを決定する。文字列を返す方が柔軟で、テストも書きやすい。A) を選択
