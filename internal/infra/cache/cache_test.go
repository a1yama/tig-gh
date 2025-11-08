package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache_MemoryOnly(t *testing.T) {
	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   false,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// メモリキャッシュとして動作することを確認
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestNewCache_FileOnly(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: false,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// ファイルキャッシュとして動作することを確認
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestNewCache_Both(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// 両方のキャッシュとして動作することを確認
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestNewCache_NoCacheEnabled(t *testing.T) {
	config := CacheConfig{
		MemoryEnabled: false,
		FileEnabled:   false,
	}

	_, err := NewCache(config)
	assert.Error(t, err, "少なくとも1つのキャッシュが有効である必要がある")
}

func TestNewCache_FileEnabledWithoutDir(t *testing.T) {
	config := CacheConfig{
		MemoryEnabled: false,
		FileEnabled:   true,
		FileDir:       "",
	}

	_, err := NewCache(config)
	assert.Error(t, err, "ファイルキャッシュが有効な場合はディレクトリが必要")
}

func TestCache_GetFromMemoryFirst(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	// 値を設定
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	// メモリキャッシュをクリア
	c := cache.(*Cache)
	err = c.memory.Clear()
	require.NoError(t, err)

	// ファイルキャッシュから取得されることを確認
	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)

	// ファイルキャッシュから取得した値がメモリキャッシュにも保存されることを確認
	value, ok = c.memory.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestCache_SetToBoth(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// 値を設定
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	// メモリキャッシュに保存されていることを確認
	value, ok := c.memory.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)

	// ファイルキャッシュに保存されていることを確認
	value, ok = c.file.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestCache_DeleteFromBoth(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// 値を設定
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	// 削除
	err = cache.Delete("test-key")
	require.NoError(t, err)

	// メモリキャッシュから削除されていることを確認
	_, ok := c.memory.Get("test-key")
	assert.False(t, ok)

	// ファイルキャッシュから削除されていることを確認
	_, ok = c.file.Get("test-key")
	assert.False(t, ok)
}

func TestCache_ClearBoth(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// 複数の値を設定
	err = cache.Set("key1", "value1", 5*time.Minute)
	require.NoError(t, err)
	err = cache.Set("key2", "value2", 5*time.Minute)
	require.NoError(t, err)
	err = cache.Set("key3", "value3", 5*time.Minute)
	require.NoError(t, err)

	// すべてクリア
	err = cache.Clear()
	require.NoError(t, err)

	// メモリキャッシュからクリアされていることを確認
	_, ok := c.memory.Get("key1")
	assert.False(t, ok)
	_, ok = c.memory.Get("key2")
	assert.False(t, ok)
	_, ok = c.memory.Get("key3")
	assert.False(t, ok)

	// ファイルキャッシュからクリアされていることを確認
	_, ok = c.file.Get("key1")
	assert.False(t, ok)
	_, ok = c.file.Get("key2")
	assert.False(t, ok)
	_, ok = c.file.Get("key3")
	assert.False(t, ok)
}

func TestCache_WithCleanupInterval(t *testing.T) {
	config := CacheConfig{
		MemoryEnabled:   true,
		FileEnabled:     false,
		CleanupInterval: 100 * time.Millisecond,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	// 短いTTLで値を設定
	err = cache.Set("expiring-key", "value", 50*time.Millisecond)
	require.NoError(t, err)

	// 最初は取得できることを確認
	value, ok := cache.Get("expiring-key")
	require.True(t, ok)
	assert.Equal(t, "value", value)

	// クリーンアップが実行されるまで待機
	time.Sleep(200 * time.Millisecond)

	// 期限切れのエントリが削除されていることを確認
	_, ok = cache.Get("expiring-key")
	assert.False(t, ok)
}

func TestCache_GetNonExistentKey(t *testing.T) {
	tmpDir := t.TempDir()

	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       tmpDir,
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	// 存在しないキーを取得
	value, ok := cache.Get("non-existent-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCache_Overwrite(t *testing.T) {
	config := CacheConfig{
		MemoryEnabled: true,
		FileEnabled:   true,
		FileDir:       t.TempDir(),
	}

	cache, err := NewCache(config)
	require.NoError(t, err)

	// 最初の値を設定
	err = cache.Set("overwrite-key", "original-value", 5*time.Minute)
	require.NoError(t, err)

	// 値を上書き
	err = cache.Set("overwrite-key", "new-value", 5*time.Minute)
	require.NoError(t, err)

	// 新しい値が取得できることを確認
	value, ok := cache.Get("overwrite-key")
	require.True(t, ok)
	assert.Equal(t, "new-value", value)
}

// 新しいConfig APIのテスト

func TestNewCacheWithConfig_MemoryOnly(t *testing.T) {
	config := DefaultConfig().DisableFileCache()

	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// デフォルトTTLで保存される
	err = cache.Set("test-key", "test-value", 0)
	require.NoError(t, err)

	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestNewCacheWithConfig_CustomTTL(t *testing.T) {
	config := DefaultConfig().
		DisableFileCache().
		WithMemoryTTL(10 * time.Minute)

	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	c := cache.(*Cache)
	assert.Equal(t, 10*time.Minute, c.config.MemoryTTL)
}

func TestNewCacheWithConfig_InvalidConfig(t *testing.T) {
	config := DefaultConfig().
		WithMemoryTTL(-1 * time.Minute) // Invalid TTL

	_, err := NewCacheWithConfig(config)
	assert.Error(t, err)
}

func TestCache_GenerateKey(t *testing.T) {
	config := DefaultConfig().DisableFileCache()
	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// シンプルなキー生成
	key1 := c.GenerateKey("issues", "owner", "repo")
	assert.Equal(t, "issues:owner:repo", key1)

	// 数値を含むキー
	key2 := c.GenerateKey("issue", "owner", "repo", 123)
	assert.Equal(t, "issue:owner:repo:123", key2)

	// 構造体を含むキー（ハッシュ化される）
	type Options struct {
		State string
		Sort  string
	}
	opts := Options{State: "open", Sort: "updated"}
	key3 := c.GenerateKey("issues", "owner", "repo", opts)
	assert.Contains(t, key3, "issues:owner:repo:")
	assert.True(t, len(key3) > len("issues:owner:repo:"))
}

func TestCache_SetWithDefaultTTL(t *testing.T) {
	config := DefaultConfig().
		DisableFileCache().
		WithMemoryTTL(10 * time.Minute)

	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	// TTL=0で保存するとデフォルトTTLが使われる
	err = cache.Set("test-key", "test-value", 0)
	require.NoError(t, err)

	value, ok := cache.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, "test-value", value)
}

// Contextベースのオプションテスト

func TestCache_GetWithContext_SkipCache(t *testing.T) {
	config := DefaultConfig().DisableFileCache()
	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// まず値を保存
	err = cache.Set("test-key", "test-value", 5*time.Minute)
	require.NoError(t, err)

	// SkipCacheコンテキストで取得
	ctx := WithSkipCacheContext(context.Background())
	value, ok := c.GetWithContext(ctx, "test-key")
	assert.False(t, ok) // キャッシュをスキップするのでfalse
	assert.Nil(t, value)

	// 通常のコンテキストでは取得できる
	value, ok = c.GetWithContext(context.Background(), "test-key")
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestCache_SetWithContext_SkipCache(t *testing.T) {
	config := DefaultConfig().DisableFileCache()
	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// SkipCacheコンテキストで保存
	ctx := WithSkipCacheContext(context.Background())
	err = c.SetWithContext(ctx, "test-key", "test-value", 0)
	require.NoError(t, err)

	// キャッシュに保存されていないことを確認
	value, ok := cache.Get("test-key")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestCache_SetWithContext_CustomTTL(t *testing.T) {
	config := DefaultConfig().
		DisableFileCache().
		WithMemoryTTL(10 * time.Minute)

	cache, err := NewCacheWithConfig(config)
	require.NoError(t, err)

	c := cache.(*Cache)

	// カスタムTTLコンテキストで保存
	ctx := WithTTLContext(context.Background(), 1*time.Minute)
	err = c.SetWithContext(ctx, "test-key", "test-value", 0)
	require.NoError(t, err)

	// 値が保存されていることを確認
	value, ok := cache.Get("test-key")
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}
