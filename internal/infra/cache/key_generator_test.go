package cache_test

import (
	"testing"

	"github.com/a1yama/tig-gh/internal/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestKeyGenerator_GenerateKey(t *testing.T) {
	kg := cache.NewKeyGenerator()

	tests := []struct {
		name     string
		resource string
		params   []interface{}
		want     string
	}{
		{
			name:     "simple key with strings",
			resource: "issues",
			params:   []interface{}{"owner", "repo"},
			want:     "issues:owner:repo",
		},
		{
			name:     "key with int",
			resource: "issue",
			params:   []interface{}{"owner", "repo", 123},
			want:     "issue:owner:repo:123",
		},
		{
			name:     "key with bool",
			resource: "prs",
			params:   []interface{}{"owner", "repo", true},
			want:     "prs:owner:repo:true",
		},
		{
			name:     "key with nil",
			resource: "commits",
			params:   []interface{}{"owner", "repo", nil},
			want:     "commits:owner:repo:nil",
		},
		{
			name:     "empty params",
			resource: "test",
			params:   []interface{}{},
			want:     "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kg.GenerateKey(tt.resource, tt.params...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKeyGenerator_GenerateKey_WithStruct(t *testing.T) {
	kg := cache.NewKeyGenerator()

	type Options struct {
		State string
		Sort  string
	}

	opts := Options{
		State: "open",
		Sort:  "updated",
	}

	key1 := kg.GenerateKey("issues", "owner", "repo", opts)
	key2 := kg.GenerateKey("issues", "owner", "repo", opts)
	key3 := kg.GenerateKey("issues", "owner", "repo", Options{State: "closed", Sort: "updated"})

	// 同じ構造体なら同じハッシュ
	assert.Equal(t, key1, key2)

	// 異なる構造体なら異なるハッシュ
	assert.NotEqual(t, key1, key3)

	// プレフィックスは同じ
	assert.Contains(t, key1, "issues:owner:repo:")
	assert.Contains(t, key3, "issues:owner:repo:")
}

func TestKeyGenerator_GenerateKey_Consistency(t *testing.T) {
	kg := cache.NewKeyGenerator()

	// 同じ引数で複数回呼び出しても同じキーが生成される
	key1 := kg.GenerateKey("test", "a", "b", 123)
	key2 := kg.GenerateKey("test", "a", "b", 123)

	assert.Equal(t, key1, key2)
}
