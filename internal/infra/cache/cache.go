package cache

import (
	"fmt"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CacheConfig キャッシュの設定
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
	memory repository.CacheService
	file   repository.CacheService
	config CacheConfig
}

// NewCache 新しいCacheを作成
func NewCache(config CacheConfig) (repository.CacheService, error) {
	c := &Cache{
		config: config,
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
		if config.FileDir == "" {
			return nil, fmt.Errorf("file cache directory is required when file cache is enabled")
		}

		var err error
		c.file, err = NewFileCache(config.FileDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create file cache: %w", err)
		}
	}

	// 少なくとも1つのキャッシュが有効になっている必要がある
	if !config.MemoryEnabled && !config.FileEnabled {
		return nil, fmt.Errorf("at least one cache type must be enabled")
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
			if c.memory != nil {
				// エラーは無視（メモリキャッシュへの保存失敗は致命的ではない）
				_ = c.memory.Set(key, value, 0)
			}
			return value, true
		}
	}

	return nil, false
}

// Set キーと値、有効期限を設定
// メモリキャッシュとファイルキャッシュの両方に保存
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
