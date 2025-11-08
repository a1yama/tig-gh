package repository

import "time"

// CacheService キャッシュサービスのインターフェース
type CacheService interface {
	// Get キーに対応する値を取得する
	// 値が存在しない、または有効期限切れの場合はfalseを返す
	Get(key string) (interface{}, bool)

	// Set キーと値、有効期限を設定する
	// ttlが0の場合は無期限
	Set(key string, value interface{}, ttl time.Duration) error

	// Delete 指定したキーの値を削除する
	Delete(key string) error

	// Clear すべてのキャッシュをクリアする
	Clear() error
}
