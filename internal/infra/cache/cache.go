package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CacheConfig キャッシュの設定（後方互換性のため残す）
// Deprecated: 新しいコードではConfigを使用してください
type CacheConfig struct {
	// MemoryEnabled メモリキャッシュを有効にするか
	MemoryEnabled bool
	// FileEnabled ファイルキャッシュを有効にするか
	FileEnabled bool
	// FileDir ファイルキャッシュのディレクトリ
	FileDir string
	// CleanupInterval クリーンアップの実行間隔（0の場合は無効）
	CleanupInterval time.Duration
}

// Cache メモリキャッシュとファイルキャッシュを統合したキャッシュマネージャー
type Cache struct {
	memory       repository.CacheService
	file         repository.CacheService
	config       *Config
	keyGenerator KeyGenerator
}

// NewCache 新しいCacheを作成
// Deprecated: 新しいコードではNewCacheWithConfigを使用してください
func NewCache(config CacheConfig) (repository.CacheService, error) {
	// 少なくとも1つのキャッシュが有効になっている必要がある
	if !config.MemoryEnabled && !config.FileEnabled {
		return nil, fmt.Errorf("at least one cache type must be enabled")
	}

	// 古いCacheConfigを新しいConfigに変換
	newConfig := &Config{
		MemoryEnabled:   config.MemoryEnabled,
		MemoryTTL:       5 * time.Minute, // デフォルト値
		FileEnabled:     config.FileEnabled,
		FileDir:         config.FileDir,
		FileTTL:         24 * time.Hour, // デフォルト値
		CleanupInterval: config.CleanupInterval,
		Version:         "v1",
	}

	return NewCacheWithConfig(newConfig)
}

// NewCacheWithConfig 新しいConfigでCacheを作成
func NewCacheWithConfig(config *Config) (repository.CacheService, error) {
	// 設定の妥当性チェック
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}

	// 少なくとも1つのキャッシュが有効になっている必要がある
	if !config.MemoryEnabled && !config.FileEnabled {
		return nil, fmt.Errorf("at least one cache type must be enabled")
	}

	c := &Cache{
		config:       config,
		keyGenerator: NewKeyGenerator(),
	}

	// メモリキャッシュの初期化
	if config.MemoryEnabled {
		c.memory = NewMemoryCache()

		// 自動クリーンアップを開始
		if config.CleanupInterval > 0 {
			if mc, ok := c.memory.(*MemoryCache); ok {
				mc.StartCleanup(config.CleanupInterval)
			}
		}
	}

	// ファイルキャッシュの初期化
	if config.FileEnabled {
		var err error
		c.file, err = NewFileCache(config.FileDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create file cache: %w", err)
		}
	}

	return c, nil
}

// Get キーに対応する値を取得
// メモリキャッシュ → ファイルキャッシュの順で検索
func (c *Cache) Get(key string) (interface{}, bool) {
	// メモリキャッシュから取得を試みる
	if c.memory != nil {
		if value, ok := c.memory.Get(key); ok {
			return value, true
		}
	}

	// ファイルキャッシュから取得を試みる
	if c.file != nil {
		if value, ok := c.file.Get(key); ok {
			// ファイルキャッシュから取得した場合、メモリキャッシュにも保存
			if c.memory != nil && c.config.MemoryEnabled {
				// メモリキャッシュのデフォルトTTLを使用
				_ = c.memory.Set(key, value, c.config.MemoryTTL)
			}
			return value, true
		}
	}

	return nil, false
}

// GetWithContext コンテキストからオプションを読み取ってキーに対応する値を取得
func (c *Cache) GetWithContext(ctx context.Context, key string) (interface{}, bool) {
	opts := OptionsFromContext(ctx)

	// SkipCacheが指定されている場合はキャッシュを使わない
	if !opts.ShouldUseCache() {
		return nil, false
	}

	return c.Get(key)
}

// Set キーと値、有効期限を設定
// メモリキャッシュとファイルキャッシュの両方に保存
// ttl=0 の場合は無期限（インターフェース仕様に従う）
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	var lastErr error

	// メモリキャッシュに保存
	if c.memory != nil {
		if err := c.memory.Set(key, value, ttl); err != nil {
			lastErr = err
		}
	}

	// ファイルキャッシュに保存
	if c.file != nil {
		if err := c.file.Set(key, value, ttl); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// SetWithContext コンテキストからオプションを読み取ってキーと値を設定
// defaultTTL=0 の場合はConfigのデフォルトTTLを使用
func (c *Cache) SetWithContext(ctx context.Context, key string, value interface{}, defaultTTL time.Duration) error {
	opts := OptionsFromContext(ctx)

	// SkipCacheが指定されている場合は何もしない
	if !opts.ShouldUseCache() {
		return nil
	}

	// TTLの優先順位: 1. Options.TTL, 2. defaultTTL, 3. Config.MemoryTTL/FileTTL
	var memoryTTL, fileTTL time.Duration

	if opts.TTL != nil {
		// Optionsで明示的にTTLが指定されている場合
		memoryTTL = *opts.TTL
		fileTTL = *opts.TTL
	} else if defaultTTL > 0 {
		// defaultTTLが指定されている場合
		memoryTTL = defaultTTL
		fileTTL = defaultTTL
	} else {
		// それ以外の場合はConfigのデフォルトTTLを使用
		memoryTTL = c.config.MemoryTTL
		fileTTL = c.config.FileTTL
	}

	var lastErr error

	// メモリキャッシュに保存
	if c.memory != nil && c.config.MemoryEnabled {
		if err := c.memory.Set(key, value, memoryTTL); err != nil {
			lastErr = err
		}
	}

	// ファイルキャッシュに保存
	if c.file != nil && c.config.FileEnabled {
		if err := c.file.Set(key, value, fileTTL); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GenerateKey KeyGeneratorを使ってキーを生成
func (c *Cache) GenerateKey(resource string, params ...interface{}) string {
	return c.keyGenerator.GenerateKey(resource, params...)
}

// Delete 指定したキーの値を削除
// メモリキャッシュとファイルキャッシュの両方から削除
func (c *Cache) Delete(key string) error {
	var lastErr error

	// メモリキャッシュから削除
	if c.memory != nil {
		if err := c.memory.Delete(key); err != nil {
			lastErr = err
		}
	}

	// ファイルキャッシュから削除
	if c.file != nil {
		if err := c.file.Delete(key); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Clear すべてのキャッシュをクリア
func (c *Cache) Clear() error {
	var lastErr error

	// メモリキャッシュをクリア
	if c.memory != nil {
		if err := c.memory.Clear(); err != nil {
			lastErr = err
		}
	}

	// ファイルキャッシュをクリア
	if c.file != nil {
		if err := c.file.Clear(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
