# Config パッケージ

tig-ghの設定管理機能を提供するパッケージです。

## 概要

このパッケージは、アプリケーションの設定を管理するための機能を提供します。
viperを使用して、YAMLファイルと環境変数から設定を読み込みます。

## ファイル構成

- `config.go` - 設定管理のメインロジック（Manager構造体）
- `loader.go` - viperを使った設定ファイルローダー（Loader構造体）
- `config_test.go` - テストコード
- `README.md` - このファイル

## 使用方法

### 基本的な使い方

```go
package main

import (
    "fmt"
    "log"

    "github.com/a1yama/tig-gh/internal/infra/config"
)

func main() {
    // グローバルマネージャーを使って設定を読み込み
    if err := config.Load(); err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // 設定を取得
    cfg := config.Get()

    fmt.Printf("GitHub Token: %s\n", cfg.GitHub.Token)
    fmt.Printf("Default Owner: %s\n", cfg.GitHub.DefaultOwner)
    fmt.Printf("Theme: %s\n", cfg.UI.Theme)
}
```

### カスタムパスから読み込む

```go
if err := config.LoadWithPath("/path/to/config.yaml"); err != nil {
    log.Fatalf("failed to load config: %v", err)
}
```

### GitHub Tokenの取得

```go
// 設定ファイルまたは環境変数 GITHUB_TOKEN から取得
token := config.GetGitHubToken()
if token == "" {
    log.Fatal("GitHub token is not configured")
}
```

### 設定の更新と保存

```go
cfg := config.Get()
cfg.UI.Theme = "dark"
cfg.GitHub.DefaultOwner = "myorg"

// デフォルトパスに保存
if err := config.Save(); err != nil {
    log.Fatalf("failed to save config: %v", err)
}

// カスタムパスに保存
if err := config.SaveTo("/path/to/config.yaml"); err != nil {
    log.Fatalf("failed to save config: %v", err)
}
```

### 設定ファイルの監視

```go
// 設定ファイルの変更を監視
if err := config.Watch(); err != nil {
    log.Fatalf("failed to watch config: %v", err)
}

// 設定が変更されると自動的に再読み込みされる
```

### Manager インスタンスを直接使う

```go
manager := config.NewManager()

if err := manager.Load(); err != nil {
    log.Fatalf("failed to load config: %v", err)
}

cfg := manager.GetConfig()
fmt.Printf("Config: %+v\n", cfg)
```

## 設定ファイルの場所

設定ファイルは以下の場所から順番に検索されます：

1. `./tig-gh/config.yaml` (カレントディレクトリ)
2. `~/.config/tig-gh/config.yaml` (ホームディレクトリ)
3. `~/.tig-gh/config.yaml` (ホームディレクトリの隠しフォルダ)
4. `/etc/tig-gh/config.yaml` (システムワイド)

デフォルトの設定ファイルのサンプルは `config/default.yaml` にあります。

## 環境変数

以下の環境変数がサポートされています：

- `GITHUB_TOKEN` - GitHubのパーソナルアクセストークン
- `GITHUB_API_URL` - GitHub APIのベースURL（GitHub Enterpriseなど）
- `TIG_GH_*` - 任意の設定項目（例: `TIG_GH_GITHUB_DEFAULT_OWNER`）

環境変数は設定ファイルより優先されます。

## 設定項目

詳細な設定項目については、`internal/domain/models/config.go` を参照してください。

主な設定項目：

- **GitHub設定** (`github`)
  - `token` - GitHubトークン
  - `default_owner` - デフォルトのリポジトリオーナー
  - `default_repo` - デフォルトのリポジトリ名
  - `api_base_url` - APIのベースURL
  - `request_timeout` - リクエストタイムアウト
  - `rate_limit_buffer` - レート制限バッファ

- **UI設定** (`ui`)
  - `theme` - カラーテーマ (light/dark/auto)
  - `default_view` - デフォルトビュー
  - `page_size` - ページサイズ
  - `show_icons` - アイコン表示
  - `date_format` - 日付フォーマット
  - `key_bindings` - キーバインディング

- **キャッシュ設定** (`cache`)
  - `enabled` - キャッシュ有効/無効
  - `ttl` - キャッシュの有効期限
  - `dir` - キャッシュディレクトリ
  - `max_size` - 最大サイズ
  - `use_file_cache` - ファイルキャッシュ使用

## テスト

```bash
go test -v ./internal/infra/config/...
```

## 初期設定ファイルの作成

```go
if err := config.InitializeConfig(); err != nil {
    log.Fatalf("failed to initialize config: %v", err)
}
```

これにより、`~/.config/tig-gh/config.yaml` にデフォルト設定が作成されます。
