# 決定ログ

## 1. formatDuration で Truncate を使用（TUI版の Round とは異なる）
- **背景**: テストが `30s` の入力に対して `"0s"` を期待しているが、TUI版で使われている `time.Duration.Round(time.Minute)` は 30s を 1m に切り上げるため `"1m"` になりテストが失敗する
- **検討した選択肢**: (a) `Round` を使用（TUI版と同じ） (b) `Truncate` を使用（切り捨て）
- **理由**: テストが明示的に `"0s"` を期待しているため、`Truncate` を採用。HTML出力とTUI出力は変更理由が異なるため（計画書にも「共通化しない」と記載）、独立した実装として適切

## 2. GenerateHTML の戻り値を string 型にした
- **背景**: 計画書では `([]byte, error)` と記載されているが、テストコードでは `strings.Contains(html, ...)` のように string として使用している
- **検討した選択肢**: (a) `[]byte` を返す（計画書通り） (b) `string` を返す（テスト互換）
- **理由**: テストコードが先に作成されており、テストに合わせることが TDD の原則。`string` を返すことで `cmd/main.go` 側で `[]byte(html)` に変換してファイル書き出しする形にした