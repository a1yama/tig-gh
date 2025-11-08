package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// Manager は設定管理を行う構造体
type Manager struct {
	config *models.Config
	loader *Loader
	mu     sync.RWMutex
}

// NewManager は新しいManagerを作成する
func NewManager() *Manager {
	return &Manager{
		loader: NewLoader(),
	}
}

// Load は設定を読み込む
func (m *Manager) Load() error {
	cfg, err := m.loader.Load()
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}

// LoadWithPath は指定されたパスから設定を読み込む
func (m *Manager) LoadWithPath(path string) error {
	cfg, err := m.loader.LoadWithPath(path)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}

// GetConfig は現在の設定を取得する
func (m *Manager) GetConfig() *models.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return models.DefaultConfig()
	}

	// コピーを返す（外部からの変更を防ぐ）
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig は設定を更新する
func (m *Manager) UpdateConfig(cfg *models.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}

// Save は現在の設定を保存する
func (m *Manager) Save() error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if cfg == nil {
		return fmt.Errorf("no config to save")
	}

	path, err := GetDefaultConfigPath()
	if err != nil {
		return err
	}

	return m.loader.Save(cfg, path)
}

// SaveTo は現在の設定を指定されたパスに保存する
func (m *Manager) SaveTo(path string) error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if cfg == nil {
		return fmt.Errorf("no config to save")
	}

	return m.loader.Save(cfg, path)
}

// Watch は設定ファイルの変更を監視する
func (m *Manager) Watch() error {
	return m.loader.Watch(func(cfg *models.Config) {
		m.mu.Lock()
		m.config = cfg
		m.mu.Unlock()
	})
}

// GetConfigPath は使用されている設定ファイルのパスを返す
func (m *Manager) GetConfigPath() string {
	return m.loader.GetConfigPath()
}

// GetGitHubToken はGitHubトークンを取得する
// 設定ファイルまたは環境変数から取得
func (m *Manager) GetGitHubToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config != nil && m.config.GitHub.Token != "" {
		return m.config.GitHub.Token
	}

	// 環境変数から取得を試みる
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	return ""
}

// InitializeConfig はデフォルト設定ファイルを作成する
func (m *Manager) InitializeConfig() error {
	path, err := GetDefaultConfigPath()
	if err != nil {
		return err
	}

	// 既に設定ファイルが存在する場合はエラー
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	// デフォルト設定を保存
	cfg := models.DefaultConfig()
	if err := m.loader.Save(cfg, path); err != nil {
		return err
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}

// Reload は設定を再読み込みする
func (m *Manager) Reload() error {
	return m.Load()
}

// グローバルな設定マネージャーインスタンス
var (
	globalManager *Manager
	once          sync.Once
)

// GetManager はグローバルな設定マネージャーのインスタンスを取得する
func GetManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// Load はグローバルマネージャーを使用して設定を読み込む
func Load() error {
	return GetManager().Load()
}

// LoadWithPath はグローバルマネージャーを使用して指定されたパスから設定を読み込む
func LoadWithPath(path string) error {
	return GetManager().LoadWithPath(path)
}

// Get はグローバルマネージャーから現在の設定を取得する
func Get() *models.Config {
	return GetManager().GetConfig()
}

// Save はグローバルマネージャーを使用して現在の設定を保存する
func Save() error {
	return GetManager().Save()
}

// SaveTo はグローバルマネージャーを使用して現在の設定を指定されたパスに保存する
func SaveTo(path string) error {
	return GetManager().SaveTo(path)
}

// Watch はグローバルマネージャーを使用して設定ファイルの変更を監視する
func Watch() error {
	return GetManager().Watch()
}

// GetGitHubToken はグローバルマネージャーからGitHubトークンを取得する
func GetGitHubToken() string {
	return GetManager().GetGitHubToken()
}

// InitializeConfig はグローバルマネージャーを使用してデフォルト設定ファイルを作成する
func InitializeConfig() error {
	return GetManager().InitializeConfig()
}
