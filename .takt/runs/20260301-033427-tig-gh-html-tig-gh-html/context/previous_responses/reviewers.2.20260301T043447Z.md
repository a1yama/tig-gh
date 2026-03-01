## arch-review
レポートファイルへの書き込み権限を付与してください。レビューは完了しています。**結果: APPROVE** — 前回の5件の指摘はすべて resolved であり、新規のブロッキング問題は検出されませんでした。

---

## qa-review
レポートファイルへの書き込み許可が必要です。以下がQAレビューの結果です:

---

# QAレビュー

## 結果: APPROVE

## サマリー
前回指摘した4件のブロッキング問題（AIAR-001, AIAR-002, AIAR-003, QA-001）が全て適切に修正されたことを確認しました。共通パッケージ `internal/shared/metricsformat` の導入により DRY 違反が解消され、`FormatDuration` は `Round` に統一されています。テストカバレッジも `ShowWeekendInDayOfWeek` の true/false 両パスが追加され、十分な水準に達しています。新規のブロッキング問題は検出されませんでした。

## 確認した観点
| 観点 | 結果 | 備考 |
|------|------|------|
| テストカバレッジ | ✅ | 30テスト関数（html_test.go: 28, format_test.go: 2）＋ metricsformat 19テスト。全設定フラグの表示/非表示パス、nil入力、XSS防止、空データ、ShowWeekendInDayOfWeek true/false をカバー |
| テスト品質 | ✅ | テーブル駆動テスト、ヘルパー関数（`assertContains`/`assertNotContains`/`fullMetrics`/`allEnabledConfig`）、Given/When/Then構造で良好 |
| エラーハンドリング | ✅ | `GenerateHTML` で nil 入力チェック、テンプレートパース/実行エラーを適切に処理。`runMetricsHTMLExport` でAPI取得・HTML生成・ファイル書き込みの各エラーを明確なメッセージで処理 |
| 保守性 | ✅ | 共通ロジックを `internal/shared/metricsformat` に集約。型定義を `types.go` に分離し `sections.go` は224行。テンプレートは `embed` で埋め込み |
| DRY原則 | ✅ | `FormatDuration`, `ShortWeekday`, `WeekdayDisplayOrder`, `WeekdaysOnly`, `MaxQualityIssuesToDisplay` が共通パッケージから参照される構造に統一 |

## 前回指摘の追跡

### 解消済み（resolved）
| finding_id | 解消根拠 |
|------------|----------|
| AIAR-001 | `internal/shared/metricsformat/format.go` に `FormatDuration`（`Round` 使用）、`ShortWeekday`、`WeekdayDisplayOrder`、`WeekdaysOnly`、`MaxQualityIssuesToDisplay` をエクスポート。`metrics_view.go` から旧定義を削除し `metricsformat.*` を参照。テスト期待値も `30s → "1m"` に修正済み |
| AIAR-002 | `types.go:67-73` の `qualityIssueEntry` に `Severity` フィールドなし。`sections.go:155-166` の `toEntries` にも代入なし。ドメインモデルの `Severity` は `sections.go:133` でフィルタリングに正しく使用 |
| AIAR-003 | `main.go` から未使用 `ctx` の2行を削除済み。行220の `ctx` は `runMetricsHTMLExport` 内で `uc.Execute(ctx, progressFn)` に使用されており適切 |
| QA-001 | `html_test.go:700-746` に `ShowWeekendInDayOfWeek` の false/true 両テストを追加。土日データ含むフィクスチャで `"Sat"`, `"Sun"` の表示/非表示を検証 |

## 新規指摘（new）
なし

## 警告（Warning）
| # | カテゴリ | 場所 | 内容 |
|---|---------|------|------|
| 1 | DRY | `export/format.go:5` と `metrics_view.go:1319` | `formatChangePercent` が2箇所に存在。ただしTUI版はlipglossスタイル適用、HTML版はプレーンテキスト返却で振る舞いが異なるため実害なし |
| 2 | エッジケース | `html_test.go` | PRLeadTimesが20件超の切り捨て動作（`maxPRLeadTimesToDisplay = 20`）のテストなし。現状データは2件のみ |

---

## testing-review
全てのファイルを確認しました。テスト計画と実装の突合、テスト品質の包括的評価を行います。

## テストレビュー結果

### 前回指摘の検証

**TR-001** (`ShowWeekendInDayOfWeek` テスト欠落) を確認:

- `html_test.go:700-722` — `TestGenerateHTML_設定_ShowWeekendInDayOfWeekがfalseの場合土日が非表示`: Sat/Sun データを `fullMetrics().ByDayOfWeek` に追加し、`ShowWeekendInDayOfWeek: false` で `assertNotContains(t, html, "Sat")` / `assertNotContains(t, html, "Sun")` を検証 ✅
- `html_test.go:724-746` — `TestGenerateHTML_設定_ShowWeekendInDayOfWeekがtrueの場合土日が表示`: 同じフィクスチャで `ShowWeekendInDayOfWeek: true` に設定し、`assertContains(t, html, "Sat")` / `assertContains(t, html, "Sun")` を検証 ✅

両テストとも `sections.go:95-98` の条件分岐（`metricsformat.WeekdaysOnly` vs `metricsformat.WeekdayDisplayOrder`）の true/false パスを適切にカバーしています。

### テスト品質の総合評価

| 観点 | 結果 | 備考 |
|------|------|------|
| テストカバレッジ | ✅ | 8設定フラグ全て、正常系・空データ・エラー系をカバー（30テスト関数 + 22サブテスト） |
| テスト構造（Given-When-Then） | ✅ | 全テストで `// Given` / `// When` / `// Then` コメントによる明確な3段階構造 |
| テスト命名 | ✅ | `TestGenerateHTML_カテゴリ_振る舞い` 形式で一貫。日本語で意図が読み取れる |
| テスト独立性・再現性 | ✅ | `fullMetrics()` / `allEnabledConfig()` が毎回新しいインスタンスを返す。外部状態・実行順序依存なし |
| モック・フィクスチャ | ✅ | ヘルパー関数が適切に共通化。個別テストで必要な変更だけオーバーライドするパターン |
| テスト戦略 | ✅ | 公開API `GenerateHTML` 経由の結合テスト + `FormatDuration`/`ShortWeekday`/`formatChangePercent` の純粋関数テスト。テスト対象のレイヤーが適切 |

### Warning（非ブロッキング・改善推奨）

| # | カテゴリ | 場所 | 内容 |
|---|---------|------|------|
| 1 | 境界値テスト不足 | `sections.go:25-29` | `maxPRLeadTimesToDisplay = 20` による切り捨てロジック。21件以上のPRデータで20件目まで表示・21件目が非表示になることを確認するテストがない |
| 2 | 境界値テスト不足 | `sections.go:140-153` | `MaxQualityIssuesToDisplay = 5` による品質問題の優先度ベース切り捨てロジック（high優先→medium残枠）。3分岐あるが全て未テスト |
| 3 | エッジケーステスト不足 | `sections.go:78` / `template.html:92` | `IsBottleneck` フラグの表示（`← Bottleneck`）を検証するテストがない |

いずれもテスト優先度「中」（エッジケース・境界値）に該当し、ポリシー上Warningレベルです。

---

## 今回の指摘（new）

なし

## 継続指摘（persists）

なし

## 解消済み（resolved）

| finding_id | 解消根拠 |
|------------|----------|
| TR-001 | `html_test.go:700-746` に `ShowWeekendInDayOfWeek` の true/false 両パスのテストが追加済み。Sat/Sun データを含むフィクスチャで表示/非表示を `assertContains`/`assertNotContains` で検証 |

---

## 結果: APPROVE

前回指摘の TR-001 は適切に修正されました。テスト全体の品質は高く、Given-When-Then構造・独立性・命名・フィクスチャ設計いずれも基準を満たしています。REJECT基準に該当する問題はありません。