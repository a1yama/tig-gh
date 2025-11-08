package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// KeyGenerator キャッシュキー生成インターフェース
type KeyGenerator interface {
	// GenerateKey パラメータからキャッシュキーを生成
	GenerateKey(resource string, params ...interface{}) string
}

// DefaultKeyGenerator デフォルトのキー生成実装
type DefaultKeyGenerator struct{}

// NewKeyGenerator 新しいKeyGeneratorを作成
func NewKeyGenerator() KeyGenerator {
	return &DefaultKeyGenerator{}
}

// GenerateKey キャッシュキーを生成
// 例: GenerateKey("issues", "owner", "repo", opts)
//     → "issues:owner:repo:hash(opts)"
func (g *DefaultKeyGenerator) GenerateKey(resource string, params ...interface{}) string {
	parts := []string{resource}

	for _, param := range params {
		switch v := param.(type) {
		case string:
			parts = append(parts, v)
		case int:
			parts = append(parts, fmt.Sprintf("%d", v))
		case int64:
			parts = append(parts, fmt.Sprintf("%d", v))
		case bool:
			parts = append(parts, fmt.Sprintf("%t", v))
		case nil:
			// nilは"nil"として扱う
			parts = append(parts, "nil")
		default:
			// 複雑なオブジェクトはJSONハッシュ化
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				// エラーの場合は型名を使用
				parts = append(parts, fmt.Sprintf("%T", v))
			} else {
				hash := sha256.Sum256(jsonBytes)
				parts = append(parts, hex.EncodeToString(hash[:8]))
			}
		}
	}

	return strings.Join(parts, ":")
}
