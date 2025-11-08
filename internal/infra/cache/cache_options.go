package cache

import (
	"context"
	"time"
)

// ContextKey キャッシュオプションのコンテキストキー型
type ContextKey string

const (
	// CacheOptionsKey キャッシュオプションのコンテキストキー
	CacheOptionsKey ContextKey = "cache_options"
)

// Options リクエスト単位のキャッシュオプション
type Options struct {
	// SkipCache キャッシュをスキップして強制的にAPIを呼び出す
	SkipCache bool

	// TTL このリクエスト専用のTTL（nilの場合はデフォルト設定を使用）
	TTL *time.Duration

	// RefreshInBackground バックグラウンドでキャッシュを更新（古いキャッシュを即座に返す）
	RefreshInBackground bool
}

// DefaultOptions デフォルトのキャッシュオプション
func DefaultOptions() *Options {
	return &Options{
		SkipCache:           false,
		TTL:                 nil, // Use config default
		RefreshInBackground: false,
	}
}

// WithSkipCache キャッシュをスキップ
func (o *Options) WithSkipCache(skip bool) *Options {
	o.SkipCache = skip
	return o
}

// WithTTL カスタムTTLを設定
func (o *Options) WithTTL(ttl time.Duration) *Options {
	o.TTL = &ttl
	return o
}

// WithRefreshInBackground バックグラウンド更新を有効化
func (o *Options) WithRefreshInBackground(refresh bool) *Options {
	o.RefreshInBackground = refresh
	return o
}

// ContextWithOptions コンテキストにキャッシュオプションを追加
func ContextWithOptions(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, CacheOptionsKey, opts)
}

// OptionsFromContext コンテキストからキャッシュオプションを取得
// オプションが設定されていない場合はデフォルトを返す
func OptionsFromContext(ctx context.Context) *Options {
	if opts, ok := ctx.Value(CacheOptionsKey).(*Options); ok {
		return opts
	}
	return DefaultOptions()
}

// WithSkipCacheContext キャッシュスキップオプション付きコンテキストを作成（ヘルパー）
func WithSkipCacheContext(ctx context.Context) context.Context {
	return ContextWithOptions(ctx, DefaultOptions().WithSkipCache(true))
}

// WithTTLContext カスタムTTLオプション付きコンテキストを作成（ヘルパー）
func WithTTLContext(ctx context.Context, ttl time.Duration) context.Context {
	return ContextWithOptions(ctx, DefaultOptions().WithTTL(ttl))
}

// WithRefreshInBackgroundContext バックグラウンド更新オプション付きコンテキストを作成（ヘルパー）
func WithRefreshInBackgroundContext(ctx context.Context) context.Context {
	return ContextWithOptions(ctx, DefaultOptions().WithRefreshInBackground(true))
}

// ShouldUseCache キャッシュを使用すべきかを判定
func (o *Options) ShouldUseCache() bool {
	return !o.SkipCache
}

// GetEffectiveTTL 実効TTLを取得（設定されていない場合はデフォルトTTLを使用）
func (o *Options) GetEffectiveTTL(defaultTTL time.Duration) time.Duration {
	if o.TTL != nil {
		return *o.TTL
	}
	return defaultTTL
}
