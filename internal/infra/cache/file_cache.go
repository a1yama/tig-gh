package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// fileCacheEntry ファイルキャッシュエントリ
type fileCacheEntry struct {
	Value      interface{}
	Expiration time.Time
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

const cacheFileExtension = ".gob"

var registeredGobTypes sync.Map

// init で主要なモデル型を登録しておくことで、プロセス再起動後も復元できるようにする
func init() {
	mustRegisterGobType(&models.Issue{})
	mustRegisterGobType([]*models.Issue{})
	mustRegisterGobType(&models.PullRequest{})
	mustRegisterGobType([]*models.PullRequest{})
	mustRegisterGobType(&models.Comment{})
	mustRegisterGobType([]*models.Comment{})
	mustRegisterGobType(&models.Review{})
	mustRegisterGobType([]*models.Review{})
	mustRegisterGobType(&models.Commit{})
	mustRegisterGobType([]*models.Commit{})
	mustRegisterGobType(&models.SearchResults{})
	mustRegisterGobType([]models.SearchResult{})
	mustRegisterGobType(map[string]interface{}{})
	mustRegisterGobType("")
}

func mustRegisterGobType(sample interface{}) {
	if sample == nil {
		return
	}

	typ := fmt.Sprintf("%T", sample)
	if _, loaded := registeredGobTypes.LoadOrStore(typ, struct{}{}); loaded {
		return
	}

	gob.Register(sample)
}

func registerGobValue(value interface{}) {
	if value == nil {
		return
	}
	mustRegisterGobType(value)
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
	filename := sanitizeKey(key) + cacheFileExtension
	return filepath.Join(c.dir, filename)
}

// Get キーに対応する値を取得
func (c *FileCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getFilePath(key)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		return nil, false
	}

	entry, err := decodeEntry(data)
	if err != nil {
		// 壊れたエントリは削除してキャッシュミス扱い
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			_ = os.Remove(filePath)
		}()
		return nil, false
	}

	if entry.isExpired() {
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			_ = os.Remove(filePath)
		}()
		return nil, false
	}

	return entry.Value, true
}

// Set キーと値、有効期限を設定
func (c *FileCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	registerGobValue(value)

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	entry := fileCacheEntry{
		Value:      value,
		Expiration: expiration,
	}

	data, err := encodeEntry(&entry)
	if err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

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

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// Clear すべてのキャッシュをクリア
func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
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

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		cacheEntry, err := decodeEntry(data)
		if err != nil {
			_ = os.Remove(filePath)
			continue
		}

		if cacheEntry.isExpired() {
			_ = os.Remove(filePath)
		}
	}

	return nil
}

func encodeEntry(entry *fileCacheEntry) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeEntry(data []byte) (*fileCacheEntry, error) {
	var entry fileCacheEntry
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
