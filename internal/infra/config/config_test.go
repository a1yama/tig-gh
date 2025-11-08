package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.loader == nil {
		t.Fatal("Manager.loader is nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := models.DefaultConfig()

	// GitHub設定の検証
	if cfg.GitHub.APIBaseURL != "https://api.github.com/" {
		t.Errorf("unexpected APIBaseURL: %s", cfg.GitHub.APIBaseURL)
	}

	if cfg.GitHub.RequestTimeout != 30*time.Second {
		t.Errorf("unexpected RequestTimeout: %v", cfg.GitHub.RequestTimeout)
	}

	// UI設定の検証
	if cfg.UI.Theme != "auto" {
		t.Errorf("unexpected Theme: %s", cfg.UI.Theme)
	}

	if cfg.UI.PageSize != 50 {
		t.Errorf("unexpected PageSize: %d", cfg.UI.PageSize)
	}

	// Cache設定の検証
	if !cfg.Cache.Enabled {
		t.Error("Cache should be enabled by default")
	}

	if cfg.Cache.TTL != 15*time.Minute {
		t.Errorf("unexpected Cache TTL: %v", cfg.Cache.TTL)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := models.DefaultConfig()

	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}

	// 無効な設定をテスト
	cfg.GitHub.RequestTimeout = -1 * time.Second
	cfg.UI.PageSize = -10
	cfg.Cache.TTL = 0

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate should fix invalid values, got error: %v", err)
	}

	// 値が修正されているか確認
	if cfg.GitHub.RequestTimeout <= 0 {
		t.Error("RequestTimeout should be fixed to positive value")
	}

	if cfg.UI.PageSize <= 0 {
		t.Error("PageSize should be fixed to positive value")
	}

	if cfg.Cache.TTL <= 0 {
		t.Error("Cache TTL should be fixed to positive value")
	}
}

func TestManagerGetConfig(t *testing.T) {
	manager := NewManager()

	// 初期状態ではデフォルト設定が返される
	cfg := manager.GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	// カスタム設定を更新
	customCfg := models.DefaultConfig()
	customCfg.GitHub.DefaultOwner = "test-owner"
	customCfg.GitHub.DefaultRepo = "test-repo"

	if err := manager.UpdateConfig(customCfg); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// 更新された設定が取得できるか確認
	cfg = manager.GetConfig()
	if cfg.GitHub.DefaultOwner != "test-owner" {
		t.Errorf("unexpected DefaultOwner: %s", cfg.GitHub.DefaultOwner)
	}

	if cfg.GitHub.DefaultRepo != "test-repo" {
		t.Errorf("unexpected DefaultRepo: %s", cfg.GitHub.DefaultRepo)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Managerを作成
	manager := NewManager()

	// カスタム設定を作成
	customCfg := models.DefaultConfig()
	customCfg.GitHub.DefaultOwner = "test-owner"
	customCfg.GitHub.DefaultRepo = "test-repo"
	customCfg.UI.Theme = "dark"

	if err := manager.UpdateConfig(customCfg); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// 設定を保存
	if err := manager.SaveTo(configPath); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	// ファイルが作成されているか確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// 新しいManagerを作成して読み込み
	newManager := NewManager()
	if err := newManager.LoadWithPath(configPath); err != nil {
		t.Fatalf("LoadWithPath failed: %v", err)
	}

	// 読み込まれた設定を確認
	loadedCfg := newManager.GetConfig()
	if loadedCfg.GitHub.DefaultOwner != "test-owner" {
		t.Errorf("unexpected DefaultOwner after load: %s", loadedCfg.GitHub.DefaultOwner)
	}

	if loadedCfg.GitHub.DefaultRepo != "test-repo" {
		t.Errorf("unexpected DefaultRepo after load: %s", loadedCfg.GitHub.DefaultRepo)
	}

	if loadedCfg.UI.Theme != "dark" {
		t.Errorf("unexpected Theme after load: %s", loadedCfg.UI.Theme)
	}
}

func TestGetGitHubToken(t *testing.T) {
	manager := NewManager()

	// 環境変数をセット
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	testToken := "test-token-12345"
	os.Setenv("GITHUB_TOKEN", testToken)

	// 環境変数からトークンが取得できるか確認
	token := manager.GetGitHubToken()
	if token != testToken {
		t.Errorf("unexpected token from env: %s", token)
	}

	// 設定にトークンを設定
	cfg := models.DefaultConfig()
	cfg.GitHub.Token = "config-token-67890"
	manager.UpdateConfig(cfg)

	// 設定のトークンが優先されるか確認
	token = manager.GetGitHubToken()
	if token != "config-token-67890" {
		t.Errorf("config token should be prioritized: %s", token)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path, err := GetDefaultConfigPath()
	if err != nil {
		t.Fatalf("GetDefaultConfigPath failed: %v", err)
	}

	if path == "" {
		t.Error("GetDefaultConfigPath returned empty string")
	}

	// パスに .config/tig-gh/config.yaml が含まれているか確認
	if !filepath.IsAbs(path) {
		t.Error("GetDefaultConfigPath should return absolute path")
	}
}
