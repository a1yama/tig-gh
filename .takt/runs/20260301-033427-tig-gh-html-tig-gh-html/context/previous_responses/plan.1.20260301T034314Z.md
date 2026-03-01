計画レポートを作成しました。以下が設計の要点です：

## 計画サマリー

### アーキテクチャ

新規パッケージ `internal/app/export/` を作成し、HTML生成ロジックを配置します。既存のTUI層（`internal/ui/views/`）やユースケース層には変更を加えません。

### 新規ファイル（5ファイル + テスト）

| ファイル | 責務 | 行数見積 |
|---------|------|---------|
| `html.go` | 公開API `GenerateHTML()` + テンプレートデータ組み立て | ~120行 |
| `sections.go` | 各セクション用の構造体とビルダー関数 | ~250行 |
| `template.go` | `go:embed` でHTMLテンプレート埋め込み | ~10行 |
| `template.html` | HTML構造 + 埋め込みCSS | ~350行 |
| `format.go` | `formatDuration` 等のフォーマットヘルパー | ~60行 |

### 変更ファイル

- `cmd/tig-gh/main.go` — `--metrics-html` フラグ追加 + HTML出力分岐

### 技術選定

- **テンプレートエンジン:** Go標準の `html/template`（XSS防止、外部依存なし）
- **テンプレート埋め込み:** `go:embed` で `.html` ファイルを埋め込み
- **ヘルパー関数:** TUI版と同一ロジックだが `export` パッケージ内に独自定義（プレゼンテーション層分離）

### 既存コードへの影響

TUI機能・ユースケース層・データモデルへの影響は一切なし。`FetchLeadTimeMetricsUseCase` をそのまま再利用します。