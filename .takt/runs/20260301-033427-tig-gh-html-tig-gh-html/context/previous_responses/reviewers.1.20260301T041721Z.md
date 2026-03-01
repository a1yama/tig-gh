## arch-review
全ての変更ファイルとコードの確認が完了しました。レポートをまとめます。

---

# Architecture Review Report

## 判定: REJECT

ブロッキング問題が4件検出されました（前回指摘の未解決3件 + 新規1件）。

---

## 前回指摘の追跡

| finding_id | 状態 | 概要 |
|---|---|---|
| AIAR-001 | **persists** | DRY違反: formatDuration等の重複 + Truncate/Round不整合 |
| AIAR-002 | **persists** | 未使用構造体フィールド: qualityIssueEntry.Severity |
| AIAR-003 | **persists** | 未使用コード: ctx + `_ = ctx` |

---

## ブロッキング問題

### AIAR-001 [persists] — DRY違反 + 動作不整合: formatDuration等の重複

**ファイル**: `internal/app/export/format.go:14` / `internal/ui/views/metrics_view.go:1316`

**未解決の根拠**:
- `format.go:14` — `d = d.Truncate(time.Minute)` のまま
- `metrics_view.go:1316` — `d = d.Round(time.Minute)` のまま

以下の関数・変数が2パッケージ間で完全に重複したまま残存している:

| 重複対象 | export側 | TUI側 |
|----------|----------|-------|
| `formatDuration` | `format.go:9` | `metrics_view.go:1311` |
| `shortWeekday` | `format.go:39` | `metrics_view.go:1367` |
| `weekdayDisplayOrder` | `sections.go:118` | `metrics_view.go:1341` |
| `weekdaysOnly` | `sections.go:128` | `metrics_view.go:1351` |
| `maxQualityIssuesToDisplay` | `sections.go:116` | `metrics_view.go:955` |

ポリシー上「本質的に同じロジックの重複（DRY違反）」はREJECT対象。`formatDuration` はさらに動作不整合を含む（`Truncate` vs `Round`）。

**修正案**:
共通パッケージ（例: `internal/domain/format/` または `internal/shared/metricsformat/`）を作成し、以下をエクスポートして両パッケージから参照する:
- `FormatDuration` （`Round` に統一 — TUI版と同じ動作）
- `ShortWeekday`
- `WeekdayDisplayOrder`
- `WeekdaysOnly`
- `MaxQualityIssuesToDisplay`

---

### AIAR-002 [persists] — 未使用構造体フィールド: qualityIssueEntry.Severity

**ファイル**: `internal/app/export/sections.go:79`, `sections.go:286`

**未解決の根拠**:
- `sections.go:79` — `Severity string` フィールドが定義されている
- `sections.go:286` — `Severity: issue.Severity` で代入されている
- `template.html` — `.Severity` への参照は0件（grep確認済み）

品質問題はすでに `HighPriority` / `MediumPriority` スライスに分類されており、`Severity` フィールドは「念のため」のデッドデータ。ポリシー上「未使用コード」はREJECT対象。

**修正案**:
`qualityIssueEntry` 構造体から `Severity` フィールドを削除し、`sections.go:286` の代入も削除する。

---

### AIAR-003 [persists] — 変更ファイル内の未使用コード

**ファイル**: `cmd/tig-gh/main.go:214-215`

**未解決の根拠**:
```go
ctx := context.Background()    // main.go:214
_ = ctx // 将来的にコンテキストを使う  // main.go:215
```

変数 `ctx` が作成され即座に `_ = ctx` で捨てられている。「将来的にコンテキストを使う」は未Issue化のTODOコメントに相当する。`cmd/tig-gh/main.go` は今回の変更対象ファイルであり、ボーイスカウトルールが適用される。

**修正案**:
この2行を削除する。`context` パッケージのimportは `runMetricsHTMLExport` 関数（line 224）で使用されるため、importの削除は不要。

---

### ARCH-001 [new] — sections.go が300行超（348行）

**ファイル**: `internal/app/export/sections.go`（348行）

**問題**:
ナレッジの判定基準「1ファイル300行超 → REJECT」に抵触。このファイルには以下が混在している:

| 区分 | 行範囲 | 行数 |
|------|--------|------|
| 型定義（11構造体） | 11-111 | ~100行 |
| 定数・変数 | 113-134 | ~22行 |
| ビルダー関数（7関数） | 136-348 | ~212行 |

型定義とビルダー関数は関連しているが、ファイルサイズが上限を超えている。

**修正案**:
型定義（`templateData`, `overallData`, `prLeadTimeEntry`, `prLeadTimesData`, `phaseEntry`, `reviewPhasesData`, `dayOfWeekEntry`, `dayOfWeekData`, `weeklyComparisonData`, `qualityIssueEntry`, `qualityIssuesData`, `stagnantPREntry`, `stagnantPRsData`, `repoStatEntry`, `repositoryStatsData`）を `types.go` に分離する。これにより `sections.go` は ~248行となり基準内に収まる。

---

## 警告（非ブロッキング）

### ARCH-W01 — severity比較の大文字小文字処理の不一致

**ファイル**: `internal/app/export/sections.go:256` / `internal/ui/views/metrics_view.go:989`

export版は `issue.Severity == "high"`（大文字小文字を区別）、TUI版は `strings.EqualFold(issue.Severity, "high")`（大文字小文字を無視）を使用。現在のデータソースは常に小文字を生成するため実害はないが、AIAR-001の共通化と併せて統一することを推奨。

---

## 変更スコープ評価

| 項目 | 値 |
|------|-----|
| 変更サイズ | Medium（新規ファイル6件 + 既存変更2件、実装コード ~460行 + テスト ~800行 + テンプレート ~230行） |
| まとまり | 良好 — 全変更がHTML出力機能に関連 |
| 無関係な変更 | なし（`ShowWeekendInDayOfWeek` 設定追加はタスクスコープに含まれる機能強化） |

---

## 確認済み・問題なし

| 観点 | 結果 |
|------|------|
| レイヤー設計 | ✅ 依存方向は正しい（cmd → app/export → domain/models） |
| モジュール配置 | ✅ `internal/app/export/` は app層のサブパッケージとして妥当 |
| 循環依存 | ✅ なし |
| パブリックAPI公開範囲 | ✅ `GenerateHTML` のみがエクスポート、内部型は非公開 |
| テンプレート埋め込み | ✅ `go:embed` の正しい使用 |
| テストカバレッジ | ✅ 27テスト（全セクション表示/非表示、空データ、nil入力、XSS防止をカバー） |
| エラーハンドリング | ✅ nil引数チェック、テンプレートエラー、ファイル書き込みエラーを処理 |
| 関数設計 | ✅ 各ビルダー関数は1責務、30行以内 |
| TODOコメント | ✅ export パッケージ内になし（main.go:215 は AIAR-003 で対処） |
| 仕様準拠 | ✅ `MetricsConfig` の全設定フラグをHTML出力に反映 |
| ユースケース再利用 | ✅ `FetchLeadTimeMetricsUseCase` を正しく再利用 |

---

## qa-review
ファイル書き込みの権限が必要です。レポートの内容をここに出力します。

---

# QA Review Report

## 判定: REJECT

ブロッキング問題が4件（persists: 3件、new: 1件）検出されました。

---

## 前回指摘の追跡

| finding_id | 概要 | 状態 | 根拠 |
|------------|------|------|------|
| AIAR-001 | DRY違反 + formatDuration の Truncate vs Round 不整合 | persists | `format.go:14` は依然 `Truncate`、TUI版 `metrics_view.go:1316` は `Round` |
| AIAR-002 | 未使用フィールド qualityIssueEntry.Severity | persists | `sections.go:79` に定義、`sections.go:286` で代入、`template.html` では未参照 |
| AIAR-003 | main.go の未使用 ctx | persists | `cmd/tig-gh/main.go:214-215` に `ctx := context.Background()` / `_ = ctx` が残存 |

---

## ブロッキング問題

### AIAR-001 [persists] — DRY違反 + formatDuration の Truncate vs Round 不整合

**ファイル**: `internal/app/export/format.go:14` / `internal/ui/views/metrics_view.go:1316`

**問題**:
export版の `formatDuration` は `d.Truncate(time.Minute)` を使用しているが、TUI版は `d.Round(time.Minute)` を使用している。同じデータに対して異なる出力を生成する。

テスト観点からの追加指摘: `format_test.go:58` のテストケース `"境界値: 1分未満は秒表示"` で `30 * time.Second` の期待値を `"0s"` としているが、TUI版の `Round` を使った場合は `1m` になるはず。テスト自体がTUI版と異なる動作を正とした仕様になっている。

さらに、以下の関数・変数が完全に重複している（DRY違反）:

| 重複対象 | export側 | TUI側 |
|----------|----------|-------|
| `formatDuration` | `format.go:9` | `metrics_view.go:1311` |
| `shortWeekday` | `format.go:39` | `metrics_view.go:1367` |
| `weekdayDisplayOrder` | `sections.go:118` | `metrics_view.go:1341` |
| `weekdaysOnly` | `sections.go:128` | `metrics_view.go:1351` |

**修正案**: 共通パッケージに抽出し、`Round` に統一する。

---

### AIAR-002 [persists] — 未使用構造体フィールド: qualityIssueEntry.Severity

**ファイル**: `internal/app/export/sections.go:79`, `sections.go:286`

**問題**:
`qualityIssueEntry` の `Severity` フィールドが `sections.go:286` で代入されているが、`template.html` では一切参照されていない。REJECT基準の「未使用コード（念のためのコード）」に該当する。

**修正案**: `qualityIssueEntry` から `Severity` フィールドを削除し、`sections.go:286` の代入も削除する。

---

### AIAR-003 [persists] — 変更ファイル内の未使用コード（ボーイスカウトルール）

**ファイル**: `cmd/tig-gh/main.go:214-215`

**問題**:
```go
ctx := context.Background()
_ = ctx // 将来的にコンテキストを使う
```

master ブランチにも存在する既存コードだが、`cmd/tig-gh/main.go` は今回の変更対象ファイルであるため、ボーイスカウトルールにより修正対象。

**修正案**: この2行を削除する。`context` パッケージの import が `runMetricsHTMLExport` で使用されているため、import は残す。

---

### QA-001 [new] — テストがない新しい振る舞い: ShowWeekendInDayOfWeek 設定の true パス

**ファイル**: `internal/app/export/sections.go:219`

**問題**:
`buildDayOfWeek` 関数に `cfg.ShowWeekendInDayOfWeek` による条件分岐があるが、全テストで `ShowWeekendInDayOfWeek` は Go のゼロ値 `false` のまま。`true` の場合に土曜・日曜が表示されることを検証するテストが存在しない。

他の設定フラグは全て有効/無効の切り替えテストがあるのに、`ShowWeekendInDayOfWeek` のみテストがない。REJECT基準「テストがない新しい振る舞い」に該当する。

**修正案**: 土日データを含むメトリクスで `ShowWeekendInDayOfWeek: true` を設定し、`"Sat"` `"Sun"` がHTMLに含まれることを検証するテストを追加する。

---

## 警告（非ブロッキング）

- **QA-W01**: 表示件数上限（`maxPRLeadTimesToDisplay=20`, `maxQualityIssuesToDisplay=5`）の truncation テストがない
- **QA-W02**: `shortWeekday` / `formatChangePercent` の直接テストがない（間接カバーのみ）
- **QA-W03**: `runMetricsHTMLExport` の統合テストがない（main パッケージのため困難）

---

## テストカバレッジ評価

29テスト関数。全セクションの正常系、設定トグル（7中6フラグ）、nil/空データ、XSS、HTML構造をカバー。テスト品質はテーブル駆動・ヘルパー関数・Given-When-Then構造で良好。`ShowWeekendInDayOfWeek` と境界値テストが不足。

---

## testing-review
十分な情報が揃いました。テスト計画・実装の突合とテスト品質の分析結果をまとめます。

---

# Testing Review Report

## 判定: REJECT

ブロッキング問題が1件検出されました。

---

## ブロッキング問題

### TR-001 [new] — `ShowWeekendInDayOfWeek` 設定のテストが存在しない

**ファイル**: `internal/app/export/sections.go:218-221`（実装） / `internal/app/export/html_test.go`（テスト不足）

**問題**:
`buildDayOfWeek` 関数には `ShowWeekendInDayOfWeek` フラグによる分岐ロジックがある：

```go
// sections.go:218-221
displayDays := weekdaysOnly
if cfg.ShowWeekendInDayOfWeek {
    displayDays = weekdayDisplayOrder
}
```

この分岐は export パッケージの新規実装であり、以下の2つのパスに対するテストがいずれも不十分：

1. **`ShowWeekendInDayOfWeek = false` パス**（デフォルト）: `TestGenerateHTML_曜日別統計のデータが含まれる` が Mon〜Fri の存在を確認しているが、Sat/Sun が **含まれないこと** の検証がない
2. **`ShowWeekendInDayOfWeek = true` パス**: このパスを通るテストが一切存在しない。`allEnabledConfig()` も `ShowWeekendInDayOfWeek` を設定していない（Goゼロ値 `false`）

さらに、テストフィクスチャ `fullMetrics()` の `ByDayOfWeek` にも Saturday/Sunday のデータが含まれていないため、仮にフラグを `true` にしても差異を検出できない。

ポリシー「テストがない新しい振る舞い → REJECT」に該当。

**修正案**:
以下の2テストを `html_test.go` に追加する：

```go
func TestGenerateHTML_曜日別統計_週末非表示(t *testing.T) {
    // Given: 土日データを含むメトリクス + ShowWeekendInDayOfWeek = false
    metrics := fullMetrics()
    metrics.ByDayOfWeek[time.Saturday] = models.DayOfWeekStats{ReviewCount: 2, MergeCount: 1}
    metrics.ByDayOfWeek[time.Sunday] = models.DayOfWeekStats{ReviewCount: 1, MergeCount: 0}
    config := &models.MetricsConfig{ShowDayOfWeek: true, ShowWeekendInDayOfWeek: false}

    // When
    html, err := GenerateHTML(metrics, config)

    // Then
    // Mon-Fri は表示、Sat/Sun は非表示
    assertContains(t, html, "Mon")
    assertNotContains(t, html, "Sat")
    assertNotContains(t, html, "Sun")
}

func TestGenerateHTML_曜日別統計_週末表示(t *testing.T) {
    // Given: 土日データを含むメトリクス + ShowWeekendInDayOfWeek = true
    metrics := fullMetrics()
    metrics.ByDayOfWeek[time.Saturday] = models.DayOfWeekStats{ReviewCount: 2, MergeCount: 1}
    metrics.ByDayOfWeek[time.Sunday] = models.DayOfWeekStats{ReviewCount: 1, MergeCount: 0}
    config := &models.MetricsConfig{ShowDayOfWeek: true, ShowWeekendInDayOfWeek: true}

    // When
    html, err := GenerateHTML(metrics, config)

    // Then
    // Mon-Sun 全曜日が表示
    assertContains(t, html, "Sat")
    assertContains(t, html, "Sun")
}
```

---

## 警告（Warning）

### TR-W001 — 表示件数上限の境界値テストが不足

**ファイル**: `internal/app/export/sections.go:114-115`

`maxPRLeadTimesToDisplay = 20` および `maxQualityIssuesToDisplay = 5` の上限値に対するテストがない。テストフィクスチャには PRLeadTimes が 2 件、QualityIssues が 2 件しかなく、切り捨てロジック（`sections.go:151-153`, `sections.go:263-276`）が一度も実行されない。

境界値テスト推奨：21件のPR / 6件の品質問題を投入し、出力が上限で切り捨てられることを検証する。

### TR-W002 — `TestGenerateHTML_設定_リポジトリ統計非表示` のアサーション不足

**ファイル**: `internal/app/export/html_test.go:322-340`

テストのコメント（行336-338）には「per-repository table should not appear」「per-repository stats section heading is absent」と記載されているが、実際のアサーションは `assertContains(t, html, "1d 12h")` のみ。

`assertNotContains(t, html, "Per Repository")` を追加すべき。現状では、リポジトリ統計セクションが表示されてもテストがパスしてしまう。

### TR-W003 — `formatDuration(30s)` の期待値が AIAR-001 の修正で変更が必要

**ファイル**: `internal/app/export/format_test.go:55-58`

```go
{
    name:     "境界値: 1分未満は秒表示",
    duration: 30 * time.Second,
    want:     "0s",
},
```

現在のテストは `Truncate` の動作（`"0s"`）を検証しているが、AI レビュー AIAR-001 が指摘した通り、実装を `Round` に修正した場合、期待値は `"1m"` に変更が必要。テスト自体は現在の実装に対して正しいが、修正後のメンテナンス項目として記録する。

---

## テスト品質の総合評価

### 良い点

| 観点 | 評価 |
|------|------|
| **Given-When-Then 構造** | ✅ 全テストが明確な3段階構造。コメントも適切 |
| **テーブル駆動テスト** | ✅ `format_test.go` がテーブル駆動で11ケースを網羅 |
| **テスト独立性** | ✅ 各テストが独立したフィクスチャを生成。共有状態・実行順序依存なし |
| **再現性** | ✅ 時間・ランダム性への依存なし。固定データで決定論的 |
| **テストヘルパー** | ✅ `fullMetrics()`, `allEnabledConfig()`, `assertContains/NotContains` が適切に抽出されている |
| **エラーケース** | ✅ nil メトリクス / nil コンフィグ / 空データに対するテストあり |
| **XSS テスト** | ✅ HTML テンプレートエスケープの検証あり |
| **設定フラグ網羅** | ✅ 7つの ShowXxx フラグそれぞれに対する非表示テストあり |
| **命名** | ✅ 日本語でカテゴリ+振る舞いを記述（「正常系:」「境界値:」「設定_」） |
| **パッケージ選択** | ✅ 同一パッケージ（`package export`）で非公開関数もテスト可能 |

### テスト計画との突合

| テスト計画の対象 | テストファイル | カバー状況 |
|------------------|---------------|------------|
| `formatDuration` ヘルパー | `format_test.go` | ✅ 11ケース（正常系6 + 境界値5） |
| `GenerateHTML` 公開API | `html_test.go` | ✅ 25テスト |
| 全セクション有効時の出力 | `html_test.go:151` | ✅ |
| HTML構造（DOCTYPE, style） | `html_test.go:176,196` | ✅ |
| 各 ShowXxx フラグ | `html_test.go:214-340` | ✅（7フラグ + 全無効） |
| 空データ / nil入力 | `html_test.go:370-416` | ✅ |
| XSS防止 | `html_test.go:418` | ✅ |
| 各セクション詳細データ | `html_test.go:452-717` | ✅（8セクション） |
| `ShowWeekendInDayOfWeek` | — | ❌ **テストなし（TR-001）** |
| 表示件数上限 | — | ⚠️ 境界値テストなし（TR-W001） |