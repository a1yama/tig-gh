package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()

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
			name:  "構造体の保存と取得",
			key:   "struct-key",
			value: struct{ Name string }{Name: "test"},
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

func TestMemoryCache_GetNonExistentKey(t *testing.T) {
	cache := NewMemoryCache()

	value, ok := cache.Get("non-existent-key")
	assert.False(t, ok, "存在しないキーはfalseを返すべき")
	assert.Nil(t, value)
}

func TestMemoryCache_TTLExpiration(t *testing.T) {
	cache := NewMemoryCache()

	// 非常に短いTTLで設定
	err := cache.Set("expiring-key", "value", 100*time.Millisecond)
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

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	// 値を設定
	err := cache.Set("delete-key", "value", 5*time.Minute)
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
}

func TestMemoryCache_DeleteNonExistentKey(t *testing.T) {
	cache := NewMemoryCache()

	// 存在しないキーの削除はエラーにならない
	err := cache.Delete("non-existent-key")
	assert.NoError(t, err)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	// 複数の値を設定
	err := cache.Set("key1", "value1", 5*time.Minute)
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
}

func TestMemoryCache_Overwrite(t *testing.T) {
	cache := NewMemoryCache()

	// 最初の値を設定
	err := cache.Set("overwrite-key", "original-value", 5*time.Minute)
	require.NoError(t, err)

	// 値を上書き
	err = cache.Set("overwrite-key", "new-value", 5*time.Minute)
	require.NoError(t, err)

	// 新しい値が取得できることを確認
	value, ok := cache.Get("overwrite-key")
	require.True(t, ok)
	assert.Equal(t, "new-value", value)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	const goroutines = 100
	const operations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 並行書き込み
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := "concurrent-key"
				value := id*operations + j
				err := cache.Set(key, value, 5*time.Minute)
				assert.NoError(t, err)

				// 読み込みも同時に行う
				_, _ = cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// デッドロックせずに完了することを確認
	// 最後に設定された値が取得できることを確認
	value, ok := cache.Get("concurrent-key")
	assert.True(t, ok)
	assert.NotNil(t, value)
}

func TestMemoryCache_ConcurrentReadWrite(t *testing.T) {
	cache := NewMemoryCache()
	const readers = 50
	const writers = 50
	const operations = 100

	var wg sync.WaitGroup
	wg.Add(readers + writers)

	// 初期値を設定
	for i := 0; i < 10; i++ {
		err := cache.Set("key"+string(rune(i)), i, 5*time.Minute)
		require.NoError(t, err)
	}

	// 並行読み込み
	for i := 0; i < readers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := "key" + string(rune(j%10))
				_, _ = cache.Get(key)
			}
		}(i)
	}

	// 並行書き込み
	for i := 0; i < writers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := "key" + string(rune(j%10))
				value := id*operations + j
				_ = cache.Set(key, value, 5*time.Minute)
			}
		}(i)
	}

	wg.Wait()

	// デッドロックせずに完了することを確認
}

func TestMemoryCache_ConcurrentDelete(t *testing.T) {
	cache := NewMemoryCache()
	const goroutines = 50

	// 初期値を設定
	for i := 0; i < 100; i++ {
		err := cache.Set("delete-key-"+string(rune(i)), i, 5*time.Minute)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 並行削除
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := "delete-key-" + string(rune(j))
				_ = cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// すべての値が削除されていることを確認
	for i := 0; i < 100; i++ {
		_, ok := cache.Get("delete-key-" + string(rune(i)))
		assert.False(t, ok)
	}
}

func TestMemoryCache_Cleanup(t *testing.T) {
	cache := NewMemoryCache().(*MemoryCache)

	// 異なるTTLで複数の値を設定
	err := cache.Set("short-ttl-1", "value1", 50*time.Millisecond)
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
	cache.cleanup()

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

func TestMemoryCache_StartCleanup(t *testing.T) {
	cache := NewMemoryCache().(*MemoryCache)

	// 短いTTLで複数の値を設定
	err := cache.Set("auto-cleanup-1", "value1", 100*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set("auto-cleanup-2", "value2", 100*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set("long-ttl", "value3", 10*time.Minute)
	require.NoError(t, err)

	// 自動クリーンアップを開始（短い間隔）
	cache.StartCleanup(150 * time.Millisecond)

	// 最初はすべて取得できることを確認
	_, ok := cache.Get("auto-cleanup-1")
	require.True(t, ok)
	_, ok = cache.Get("auto-cleanup-2")
	require.True(t, ok)

	// クリーンアップが実行されるまで待機
	time.Sleep(300 * time.Millisecond)

	// 期限切れのエントリは自動的に削除されていることを確認
	_, ok = cache.Get("auto-cleanup-1")
	assert.False(t, ok, "自動クリーンアップで期限切れエントリは削除されるべき")
	_, ok = cache.Get("auto-cleanup-2")
	assert.False(t, ok, "自動クリーンアップで期限切れエントリは削除されるべき")
	_, ok = cache.Get("long-ttl")
	assert.True(t, ok, "有効なエントリは残るべき")
}
