package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileCache_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name  string
		key   string
		value interface{}
		ttl   time.Duration
	}{
		{
			name:  "文字列の保存と取得",
			key:   "test-key",
			value: "test-value",
			ttl:   5 * time.Minute,
		},
		{
			name:  "数値の保存と取得",
			key:   "number-key",
			value: 12345,
			ttl:   5 * time.Minute,
		},
		{
			name:  "マップの保存と取得",
			key:   "map-key",
			value: map[string]interface{}{"name": "test", "age": 20},
			ttl:   5 * time.Minute,
		},
		{
			name:  "無期限キャッシュ",
			key:   "permanent-key",
			value: "permanent-value",
			ttl:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set
			err := cache.Set(tt.key, tt.value, tt.ttl)
			require.NoError(t, err)

			// Get
			value, ok := cache.Get(tt.key)
			require.True(t, ok, "値が取得できるべき")
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestFileCache_GetNonExistentKey(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	value, ok := cache.Get("non-existent-key")
	assert.False(t, ok, "存在しないキーはfalseを返すべき")
	assert.Nil(t, value)
}

func TestFileCache_TTLExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 非常に短いTTLで設定
	err = cache.Set("expiring-key", "value", 100*time.Millisecond)
	require.NoError(t, err)

	// すぐに取得できることを確認
	value, ok := cache.Get("expiring-key")
	require.True(t, ok)
	assert.Equal(t, "value", value)

	// TTL経過後は取得できないことを確認
	time.Sleep(150 * time.Millisecond)
	value, ok = cache.Get("expiring-key")
	assert.False(t, ok, "TTL経過後は値を取得できないべき")
	assert.Nil(t, value)
}

func TestFileCache_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 値を設定
	err = cache.Set("delete-key", "value", 5*time.Minute)
	require.NoError(t, err)

	// 削除前は取得できることを確認
	_, ok := cache.Get("delete-key")
	require.True(t, ok)

	// 削除
	err = cache.Delete("delete-key")
	require.NoError(t, err)

	// 削除後は取得できないことを確認
	value, ok := cache.Get("delete-key")
	assert.False(t, ok)
	assert.Nil(t, value)

	// ファイルも削除されていることを確認
	filePath := filepath.Join(tmpDir, sanitizeKey("delete-key")+cacheFileExtension)
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "ファイルが削除されているべき")
}

func TestFileCache_DeleteNonExistentKey(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 存在しないキーの削除はエラーにならない
	err = cache.Delete("non-existent-key")
	assert.NoError(t, err)
}

func TestFileCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

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

	// すべての値が取得できないことを確認
	_, ok := cache.Get("key1")
	assert.False(t, ok)
	_, ok = cache.Get("key2")
	assert.False(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)

	// ディレクトリが空であることを確認
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "キャッシュディレクトリが空であるべき")
}

func TestFileCache_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
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

func TestFileCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// 最初のキャッシュインスタンスで値を設定
	cache1, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	err = cache1.Set("persistent-key", "persistent-value", 10*time.Minute)
	require.NoError(t, err)

	// 新しいキャッシュインスタンスで同じディレクトリを使用
	cache2, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 以前設定した値が取得できることを確認（永続化されている）
	value, ok := cache2.Get("persistent-key")
	require.True(t, ok, "ファイルから値が読み込まれるべき")
	assert.Equal(t, "persistent-value", value)
}

func TestFileCache_LoadExpiredCache(t *testing.T) {
	tmpDir := t.TempDir()

	// 最初のキャッシュインスタンスで短いTTLの値を設定
	cache1, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	err = cache1.Set("expired-key", "value", 100*time.Millisecond)
	require.NoError(t, err)

	// TTL経過まで待機
	time.Sleep(150 * time.Millisecond)

	// 新しいキャッシュインスタンスで同じディレクトリを使用
	cache2, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 期限切れの値は取得できないことを確認
	value, ok := cache2.Get("expired-key")
	assert.False(t, ok, "期限切れの値は取得できないべき")
	assert.Nil(t, value)
}

func TestFileCache_InvalidDirectory(t *testing.T) {
	// 無効なディレクトリパス
	_, err := NewFileCache("/invalid/path/that/does/not/exist")
	assert.Error(t, err, "無効なディレクトリではエラーが発生するべき")
}

func TestFileCache_KeySanitization(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 特殊文字を含むキー
	specialKeys := []string{
		"key/with/slashes",
		"key:with:colons",
		"key*with*asterisks",
		"key?with?questions",
		"key|with|pipes",
	}

	for _, key := range specialKeys {
		err = cache.Set(key, "value", 5*time.Minute)
		require.NoError(t, err, "特殊文字を含むキーでも保存できるべき")

		value, ok := cache.Get(key)
		require.True(t, ok, "特殊文字を含むキーでも取得できるべき")
		assert.Equal(t, "value", value)
	}
}

func TestFileCache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	const goroutines = 50
	const operations = 50

	done := make(chan bool, goroutines)

	// 並行書き込みと読み込み
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				key := "concurrent-key"
				value := id*operations + j
				_ = cache.Set(key, value, 5*time.Minute)
				_, _ = cache.Get(key)
			}
			done <- true
		}(i)
	}

	// すべてのゴルーチンが完了するのを待つ
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// デッドロックせずに完了することを確認
	value, ok := cache.Get("concurrent-key")
	assert.True(t, ok)
	assert.NotNil(t, value)
}

func TestFileCache_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	fc := cache.(*FileCache)

	// 異なるTTLで複数の値を設定
	err = cache.Set("short-ttl-1", "value1", 50*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set("short-ttl-2", "value2", 50*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set("long-ttl", "value3", 10*time.Minute)
	require.NoError(t, err)
	err = cache.Set("permanent", "value4", 0)
	require.NoError(t, err)

	// クリーンアップ前はすべて取得できることを確認
	_, ok := cache.Get("short-ttl-1")
	require.True(t, ok)
	_, ok = cache.Get("short-ttl-2")
	require.True(t, ok)
	_, ok = cache.Get("long-ttl")
	require.True(t, ok)
	_, ok = cache.Get("permanent")
	require.True(t, ok)

	// TTL経過まで待機
	time.Sleep(100 * time.Millisecond)

	// クリーンアップ実行
	err = fc.Cleanup()
	require.NoError(t, err)

	// 期限切れのエントリは削除され、それ以外は残っていることを確認
	_, ok = cache.Get("short-ttl-1")
	assert.False(t, ok, "期限切れのエントリは削除されるべき")
	_, ok = cache.Get("short-ttl-2")
	assert.False(t, ok, "期限切れのエントリは削除されるべき")
	_, ok = cache.Get("long-ttl")
	assert.True(t, ok, "有効なエントリは残るべき")
	_, ok = cache.Get("permanent")
	assert.True(t, ok, "無期限エントリは残るべき")
}

func TestFileCache_LongKeyHashing(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileCache(tmpDir)
	require.NoError(t, err)

	// 非常に長いキー（200文字以上）
	longKey := ""
	for i := 0; i < 250; i++ {
		longKey += "a"
	}

	// 長いキーでも保存・取得できることを確認
	err = cache.Set(longKey, "long-key-value", 5*time.Minute)
	require.NoError(t, err)

	value, ok := cache.Get(longKey)
	require.True(t, ok)
	assert.Equal(t, "long-key-value", value)

	// ファイル名がハッシュ化されていることを確認
	fc := cache.(*FileCache)
	filePath := fc.getFilePath(longKey)
	filename := filepath.Base(filePath)

	// ファイル名が適切な長さであることを確認（ハッシュ+.json）
	assert.Less(t, len(filename), 100, "ファイル名は適切な長さであるべき")
}
