package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/infra/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := cache.DefaultConfig()

	assert.True(t, cfg.MemoryEnabled)
	assert.Equal(t, 5*time.Minute, cfg.MemoryTTL)

	assert.True(t, cfg.FileEnabled)
	assert.NotEmpty(t, cfg.FileDir)
	assert.Equal(t, 24*time.Hour, cfg.FileTTL)

	assert.Equal(t, 10*time.Minute, cfg.CleanupInterval)
	assert.Equal(t, "v1", cfg.Version)

	// Check that file dir is under home directory
	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".cache", "tig-gh")
	assert.Equal(t, expectedDir, cfg.FileDir)
}

func TestConfig_FluentAPI(t *testing.T) {
	cfg := cache.DefaultConfig().
		WithMemoryTTL(10 * time.Minute).
		WithFileTTL(48 * time.Hour).
		WithFileDir("/tmp/test-cache").
		WithCleanupInterval(5 * time.Minute)

	assert.Equal(t, 10*time.Minute, cfg.MemoryTTL)
	assert.Equal(t, 48*time.Hour, cfg.FileTTL)
	assert.Equal(t, "/tmp/test-cache", cfg.FileDir)
	assert.Equal(t, 5*time.Minute, cfg.CleanupInterval)
}

func TestConfig_DisableMemoryCache(t *testing.T) {
	cfg := cache.DefaultConfig().DisableMemoryCache()

	assert.False(t, cfg.MemoryEnabled)
	assert.True(t, cfg.FileEnabled) // File cache still enabled
}

func TestConfig_DisableFileCache(t *testing.T) {
	cfg := cache.DefaultConfig().DisableFileCache()

	assert.True(t, cfg.MemoryEnabled) // Memory cache still enabled
	assert.False(t, cfg.FileEnabled)
}

func TestConfig_Validate_Success(t *testing.T) {
	tests := []struct {
		name   string
		config *cache.Config
	}{
		{
			name:   "default config",
			config: cache.DefaultConfig(),
		},
		{
			name: "only memory cache",
			config: cache.DefaultConfig().DisableFileCache(),
		},
		{
			name: "only file cache",
			config: cache.DefaultConfig().DisableMemoryCache(),
		},
		{
			name: "custom TTL values",
			config: cache.DefaultConfig().
				WithMemoryTTL(1 * time.Minute).
				WithFileTTL(1 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfig_Validate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		config      *cache.Config
		expectedErr string
	}{
		{
			name: "negative memory TTL",
			config: cache.DefaultConfig().
				WithMemoryTTL(-1 * time.Minute),
			expectedErr: "memory TTL must be positive",
		},
		{
			name: "zero memory TTL",
			config: cache.DefaultConfig().
				WithMemoryTTL(0),
			expectedErr: "memory TTL must be positive",
		},
		{
			name: "negative file TTL",
			config: cache.DefaultConfig().
				WithFileTTL(-1 * time.Hour),
			expectedErr: "file TTL must be positive",
		},
		{
			name: "empty file directory",
			config: cache.DefaultConfig().
				WithFileDir(""),
			expectedErr: "file directory must be specified",
		},
		{
			name: "negative cleanup interval",
			config: cache.DefaultConfig().
				WithCleanupInterval(-1 * time.Minute),
			expectedErr: "cleanup interval must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestConfigError_Error(t *testing.T) {
	err := &cache.ConfigError{Message: "test error"}
	assert.Equal(t, "cache config error: test error", err.Error())
}
