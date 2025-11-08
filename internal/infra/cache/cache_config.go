package cache

import (
	"os"
	"path/filepath"
	"time"
)

// Config キャッシュの設定
type Config struct {
	// Memory cache settings
	MemoryEnabled bool
	MemoryTTL     time.Duration

	// File cache settings
	FileEnabled bool
	FileDir     string
	FileTTL     time.Duration

	// Cleanup settings
	CleanupInterval time.Duration

	// Cache version for schema changes
	Version string
}

// DefaultConfig デフォルトのキャッシュ設定を返す
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".cache", "tig-gh")

	return &Config{
		// Memory cache: 5分間有効
		MemoryEnabled: true,
		MemoryTTL:     5 * time.Minute,

		// File cache: 24時間有効
		FileEnabled: true,
		FileDir:     cacheDir,
		FileTTL:     24 * time.Hour,

		// Cleanup: 10分ごとに実行
		CleanupInterval: 10 * time.Minute,

		// Version: スキーマ変更時に更新
		Version: "v1",
	}
}

// WithMemoryTTL メモリキャッシュのTTLを設定
func (c *Config) WithMemoryTTL(ttl time.Duration) *Config {
	c.MemoryTTL = ttl
	return c
}

// WithFileTTL ファイルキャッシュのTTLを設定
func (c *Config) WithFileTTL(ttl time.Duration) *Config {
	c.FileTTL = ttl
	return c
}

// WithFileDir ファイルキャッシュのディレクトリを設定
func (c *Config) WithFileDir(dir string) *Config {
	c.FileDir = dir
	return c
}

// WithCleanupInterval クリーンアップの間隔を設定
func (c *Config) WithCleanupInterval(interval time.Duration) *Config {
	c.CleanupInterval = interval
	return c
}

// DisableMemoryCache メモリキャッシュを無効化
func (c *Config) DisableMemoryCache() *Config {
	c.MemoryEnabled = false
	return c
}

// DisableFileCache ファイルキャッシュを無効化
func (c *Config) DisableFileCache() *Config {
	c.FileEnabled = false
	return c
}

// Validate 設定の妥当性をチェック
func (c *Config) Validate() error {
	if c.MemoryEnabled && c.MemoryTTL <= 0 {
		return &ConfigError{Message: "memory TTL must be positive"}
	}

	if c.FileEnabled {
		if c.FileTTL <= 0 {
			return &ConfigError{Message: "file TTL must be positive"}
		}
		if c.FileDir == "" {
			return &ConfigError{Message: "file directory must be specified"}
		}
	}

	// CleanupInterval は 0 でも良い（無効化を意味する）
	if c.CleanupInterval < 0 {
		return &ConfigError{Message: "cleanup interval must be non-negative"}
	}

	return nil
}

// ConfigError 設定エラー
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return "cache config error: " + e.Message
}
