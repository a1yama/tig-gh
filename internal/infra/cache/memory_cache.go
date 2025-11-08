package cache

import (
	"sync"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// cacheEntry キャッシュエントリ
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// isExpired 有効期限が切れているかチェック
func (e *cacheEntry) isExpired() bool {
	if e.expiration.IsZero() {
		return false // 無期限
	}
	return time.Now().After(e.expiration)
}

// MemoryCache メモリキャッシュの実装
type MemoryCache struct {
	data map[string]*cacheEntry
	mu   sync.RWMutex
}

// NewMemoryCache 新しいMemoryCacheを作成
func NewMemoryCache() repository.CacheService {
	return &MemoryCache{
		data: make(map[string]*cacheEntry),
	}
}

// Get キーに対応する値を取得
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if entry.isExpired() {
		// 有効期限切れの場合は削除
		// 注意: RLockを保持したまま削除はできないので、ここでは単にfalseを返す
		// クリーンアップは別のgoroutineまたは次の操作で行う
		return nil, false
	}

	return entry.value, true
}

// Set キーと値、有効期限を設定
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.data[key] = &cacheEntry{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete 指定したキーの値を削除
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Clear すべてのキャッシュをクリア
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
	return nil
}

// cleanup 有効期限切れのエントリを削除（内部使用）
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if !entry.expiration.IsZero() && now.After(entry.expiration) {
			delete(c.data, key)
		}
	}
}

// StartCleanup 定期的なクリーンアップを開始
func (c *MemoryCache) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.cleanup()
		}
	}()
}
