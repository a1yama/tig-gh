package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/infra/cache"
	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := cache.DefaultOptions()

	assert.False(t, opts.SkipCache)
	assert.Nil(t, opts.TTL)
	assert.False(t, opts.RefreshInBackground)
}

func TestOptions_FluentAPI(t *testing.T) {
	ttl := 30 * time.Minute
	opts := cache.DefaultOptions().
		WithSkipCache(true).
		WithTTL(ttl).
		WithRefreshInBackground(true)

	assert.True(t, opts.SkipCache)
	assert.NotNil(t, opts.TTL)
	assert.Equal(t, ttl, *opts.TTL)
	assert.True(t, opts.RefreshInBackground)
}

func TestOptions_ShouldUseCache(t *testing.T) {
	tests := []struct {
		name      string
		skipCache bool
		expected  bool
	}{
		{
			name:      "should use cache when not skipped",
			skipCache: false,
			expected:  true,
		},
		{
			name:      "should not use cache when skipped",
			skipCache: true,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := cache.DefaultOptions().WithSkipCache(tt.skipCache)
			assert.Equal(t, tt.expected, opts.ShouldUseCache())
		})
	}
}

func TestOptions_GetEffectiveTTL(t *testing.T) {
	defaultTTL := 10 * time.Minute

	tests := []struct {
		name        string
		customTTL   *time.Duration
		expectedTTL time.Duration
	}{
		{
			name:        "use default TTL when not set",
			customTTL:   nil,
			expectedTTL: defaultTTL,
		},
		{
			name: "use custom TTL when set",
			customTTL: func() *time.Duration {
				ttl := 30 * time.Minute
				return &ttl
			}(),
			expectedTTL: 30 * time.Minute,
		},
		{
			name: "use custom zero TTL when explicitly set",
			customTTL: func() *time.Duration {
				ttl := 0 * time.Minute
				return &ttl
			}(),
			expectedTTL: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := cache.DefaultOptions()
			if tt.customTTL != nil {
				opts.WithTTL(*tt.customTTL)
			}
			assert.Equal(t, tt.expectedTTL, opts.GetEffectiveTTL(defaultTTL))
		})
	}
}

func TestContextWithOptions(t *testing.T) {
	ctx := context.Background()
	opts := cache.DefaultOptions().WithSkipCache(true)

	ctxWithOpts := cache.ContextWithOptions(ctx, opts)

	// Verify the context contains the options
	retrievedOpts := cache.OptionsFromContext(ctxWithOpts)
	assert.NotNil(t, retrievedOpts)
	assert.True(t, retrievedOpts.SkipCache)
}

func TestOptionsFromContext_NoOptions(t *testing.T) {
	ctx := context.Background()

	// When no options are set, should return default options
	opts := cache.OptionsFromContext(ctx)
	assert.NotNil(t, opts)
	assert.False(t, opts.SkipCache)
	assert.Nil(t, opts.TTL)
}

func TestWithSkipCacheContext(t *testing.T) {
	ctx := context.Background()
	ctxWithSkip := cache.WithSkipCacheContext(ctx)

	opts := cache.OptionsFromContext(ctxWithSkip)
	assert.True(t, opts.SkipCache)
}

func TestWithTTLContext(t *testing.T) {
	ctx := context.Background()
	ttl := 15 * time.Minute
	ctxWithTTL := cache.WithTTLContext(ctx, ttl)

	opts := cache.OptionsFromContext(ctxWithTTL)
	assert.NotNil(t, opts.TTL)
	assert.Equal(t, ttl, *opts.TTL)
}

func TestWithRefreshInBackgroundContext(t *testing.T) {
	ctx := context.Background()
	ctxWithRefresh := cache.WithRefreshInBackgroundContext(ctx)

	opts := cache.OptionsFromContext(ctxWithRefresh)
	assert.True(t, opts.RefreshInBackground)
}

func TestContextHelpers_Chaining(t *testing.T) {
	// Test that we can chain context helpers
	ctx := context.Background()

	// First add skip cache
	ctx = cache.WithSkipCacheContext(ctx)
	opts1 := cache.OptionsFromContext(ctx)
	assert.True(t, opts1.SkipCache)

	// Then override with TTL context (skip cache should be lost)
	ctx = cache.WithTTLContext(ctx, 20*time.Minute)
	opts2 := cache.OptionsFromContext(ctx)
	assert.False(t, opts2.SkipCache) // New default options
	assert.NotNil(t, opts2.TTL)
	assert.Equal(t, 20*time.Minute, *opts2.TTL)
}

func TestContextHelpers_CombinedOptions(t *testing.T) {
	// To combine multiple options, use fluent API with ContextWithOptions
	ctx := context.Background()
	opts := cache.DefaultOptions().
		WithSkipCache(true).
		WithTTL(25 * time.Minute).
		WithRefreshInBackground(true)

	ctx = cache.ContextWithOptions(ctx, opts)

	retrievedOpts := cache.OptionsFromContext(ctx)
	assert.True(t, retrievedOpts.SkipCache)
	assert.NotNil(t, retrievedOpts.TTL)
	assert.Equal(t, 25*time.Minute, *retrievedOpts.TTL)
	assert.True(t, retrievedOpts.RefreshInBackground)
}
