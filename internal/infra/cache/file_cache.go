package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// fileCacheEntry ファイルキャッシュエントリ
type fileCacheEntry struct {
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

// isExpired 有効期限が切れているかチェック
func (e *fileCacheEntry) isExpired() bool {
	if e.Expiration.IsZero() {
		return false // 無期限
	}
	return time.Now().After(e.Expiration)
}

// FileCache ファイルベースのキャッシュ実装
type FileCache struct {
	dir string
	mu  sync.RWMutex
}

// NewFileCache 新しいFileCacheを作成
func NewFileCache(dir string) (repository.CacheService, error) {
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &FileCache{
		dir: dir,
	}, nil
}

// sanitizeKey キーをファイル名として安全な文字列に変換
func sanitizeKey(key string) string {
	// ファイルシステムで問題となる文字を置換
	reg := regexp.MustCompile(`[/\\:*?"<>|]`)
	sanitized := reg.ReplaceAllString(key, "_")

	// キーが長すぎる場合はハッシュ化
	if len(sanitized) > 200 {
		hash := sha256.Sum256([]byte(key))
		return hex.EncodeToString(hash[:])
	}

	return sanitized
}

// getFilePath キーに対応するファイルパスを取得
func (c *FileCache) getFilePath(key string) string {
	filename := sanitizeKey(key) + ".json"
	return filepath.Join(c.dir, filename)
}

// Get キーに対応する値を取得
func (c *FileCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getFilePath(key)

	// ファイルを読み込む
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		// その他のエラーの場合もfalseを返す
		return nil, false
	}

	// JSONをデコード
	var entry fileCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// 有効期限チェック
	if entry.isExpired() {
		// 有効期限切れの場合は削除（非同期で実行）
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			os.Remove(filePath)
		}()
		return nil, false
	}

	return entry.Value, true
}

// Set キーと値、有効期限を設定
func (c *FileCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	entry := fileCacheEntry{
		Value:      value,
		Expiration: expiration,
	}

	// JSONエンコード
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// ファイルに書き込む
	filePath := c.getFilePath(key)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Delete 指定したキーの値を削除
func (c *FileCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	err := os.Remove(filePath)

	// ファイルが存在しない場合はエラーとしない
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// Clear すべてのキャッシュをクリア
func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ディレクトリ内のすべてのファイルを削除
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				// 個別のファイル削除エラーは無視して続行
				continue
			}
		}
	}

	return nil
}

// Cleanup 有効期限切れのエントリを削除
func (c *FileCache) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var cacheEntry fileCacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		// 有効期限切れの場合は削除
		if !cacheEntry.Expiration.IsZero() && now.After(cacheEntry.Expiration) {
			os.Remove(filePath)
		}
	}

	return nil
}
