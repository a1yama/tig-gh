package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Loader は設定ファイルを読み込むための構造体
type Loader struct {
	v *viper.Viper
}

// NewLoader は新しいLoaderを作成する
func NewLoader() *Loader {
	v := viper.New()

	// 設定ファイルの名前と形式
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 設定ファイルの検索パス
	// 1. カレントディレクトリの .tig-gh
	v.AddConfigPath("./.tig-gh")

	// 2. ホームディレクトリの .config/tig-gh
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "tig-gh"))
		v.AddConfigPath(filepath.Join(home, ".tig-gh"))
	}

	// 3. /etc/tig-gh (システムワイド)
	v.AddConfigPath("/etc/tig-gh")

	// 環境変数の設定
	v.SetEnvPrefix("TIG_GH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 特定の環境変数を明示的にバインド
	v.BindEnv("github.token", "GITHUB_TOKEN")
	v.BindEnv("github.api_base_url", "GITHUB_API_URL")

	return &Loader{v: v}
}

// Load は設定ファイルを読み込み、Config構造体を返す
func (l *Loader) Load() (*models.Config, error) {
	// デフォルト設定を取得
	cfg := models.DefaultConfig()

	// 設定ファイルの読み込み（ファイルが存在しない場合はエラーを無視）
	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 設定ファイルが見つからない以外のエラーは返す
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 設定ファイルが見つからない場合は、デフォルト設定と環境変数を使用
	}

	// 設定をアンマーシャル
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 環境変数から追加で読み込み
	l.loadFromEnv(cfg)

	// キャッシュディレクトリのデフォルト値を設定
	if cfg.Cache.Dir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.Cache.Dir = filepath.Join(home, ".cache", "tig-gh")
		}
	}

	// 設定の妥当性を検証
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// loadFromEnv は環境変数から設定を読み込む
func (l *Loader) loadFromEnv(cfg *models.Config) {
	// GITHUB_TOKEN
	if token := l.v.GetString("github.token"); token != "" {
		cfg.GitHub.Token = token
	}

	// GITHUB_API_URL
	if apiURL := l.v.GetString("github.api_base_url"); apiURL != "" {
		cfg.GitHub.APIBaseURL = apiURL
	}

	// その他の環境変数も同様に処理可能
	if owner := l.v.GetString("github.default_owner"); owner != "" {
		cfg.GitHub.DefaultOwner = owner
	}

	if repo := l.v.GetString("github.default_repo"); repo != "" {
		cfg.GitHub.DefaultRepo = repo
	}
}

// LoadWithPath は指定されたパスから設定ファイルを読み込む
func (l *Loader) LoadWithPath(configPath string) (*models.Config, error) {
	l.v.SetConfigFile(configPath)
	return l.Load()
}

// GetConfigPath は使用されている設定ファイルのパスを返す
func (l *Loader) GetConfigPath() string {
	return l.v.ConfigFileUsed()
}

// Watch は設定ファイルの変更を監視し、変更があった場合にコールバックを呼び出す
func (l *Loader) Watch(callback func(*models.Config)) error {
	l.v.OnConfigChange(func(e fsnotify.Event) {
		cfg := models.DefaultConfig()
		if err := l.v.Unmarshal(cfg); err != nil {
			// エラーログを出力（実際にはloggerを使用すべき）
			fmt.Fprintf(os.Stderr, "failed to reload config: %v\n", err)
			return
		}

		l.loadFromEnv(cfg)

		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid config after reload: %v\n", err)
			return
		}

		callback(cfg)
	})

	l.v.WatchConfig()
	return nil
}

// Save は設定を指定されたパスに保存する
func (l *Loader) Save(cfg *models.Config, path string) error {
	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 設定値をviperに設定
	l.v.Set("github", cfg.GitHub)
	l.v.Set("ui", cfg.UI)
	l.v.Set("cache", cfg.Cache)
	l.v.Set("metrics", cfg.Metrics)

	// ファイルに書き込み
	if err := l.v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath はデフォルトの設定ファイルパスを返す
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".config", "tig-gh", "config.yaml"), nil
}
