package models

import "time"

// Config はアプリケーション全体の設定を表す
type Config struct {
	GitHub  GitHubConfig  `mapstructure:"github" yaml:"github"`
	UI      UIConfig      `mapstructure:"ui" yaml:"ui"`
	Cache   CacheConfig   `mapstructure:"cache" yaml:"cache"`
	Metrics MetricsConfig `mapstructure:"metrics" yaml:"metrics"`
}

// GitHubConfig はGitHub関連の設定を表す
type GitHubConfig struct {
	// Token はGitHubのパーソナルアクセストークン
	// 環境変数 GITHUB_TOKEN からも読み込み可能
	Token string `mapstructure:"token" yaml:"token"`

	// DefaultOwner はデフォルトのリポジトリオーナー
	DefaultOwner string `mapstructure:"default_owner" yaml:"default_owner"`

	// DefaultRepo はデフォルトのリポジトリ名
	DefaultRepo string `mapstructure:"default_repo" yaml:"default_repo"`

	// APIBaseURL はGitHub APIのベースURL（GitHub Enterpriseなど）
	APIBaseURL string `mapstructure:"api_base_url" yaml:"api_base_url"`

	// UploadBaseURL はGitHub UploadのベースURL
	UploadBaseURL string `mapstructure:"upload_base_url" yaml:"upload_base_url"`

	// RequestTimeout はAPIリクエストのタイムアウト
	RequestTimeout time.Duration `mapstructure:"request_timeout" yaml:"request_timeout"`

	// RateLimitBuffer はレート制限のバッファ（残りリクエスト数がこれ以下の場合は待機）
	RateLimitBuffer int `mapstructure:"rate_limit_buffer" yaml:"rate_limit_buffer"`

	// Repositories はメトリクス計算対象となるリポジトリ一覧（owner/repo形式）
	Repositories []string `mapstructure:"repositories" yaml:"repositories"`
}

// MetricsConfig はメトリクス関連の設定を表す
type MetricsConfig struct {
	// Enabled はメトリクス機能全体の有効/無効
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// LeadTimeEnabled はリードタイム計測の有効/無効
	LeadTimeEnabled bool `mapstructure:"lead_time_enabled" yaml:"lead_time_enabled"`

	// CalculationPeriod はメトリクス計算対象期間
	CalculationPeriod time.Duration `mapstructure:"calculation_period" yaml:"calculation_period"`
}

// UIConfig はUI関連の設定を表す
type UIConfig struct {
	// Theme はカラーテーマ（"light", "dark", "auto"）
	Theme string `mapstructure:"theme" yaml:"theme"`

	// DefaultView は起動時のデフォルトビュー（"issues", "prs", "commits"など）
	DefaultView string `mapstructure:"default_view" yaml:"default_view"`

	// KeyBindings はカスタムキーバインディング
	KeyBindings map[string]string `mapstructure:"key_bindings" yaml:"key_bindings"`

	// PageSize は一度に表示するアイテム数
	PageSize int `mapstructure:"page_size" yaml:"page_size"`

	// ShowIcons はアイコン表示の有効/無効
	ShowIcons bool `mapstructure:"show_icons" yaml:"show_icons"`

	// DateFormat は日付のフォーマット
	DateFormat string `mapstructure:"date_format" yaml:"date_format"`
}

// CacheConfig はキャッシュ関連の設定を表す
type CacheConfig struct {
	// Enabled はキャッシュ機能の有効/無効
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// TTL はキャッシュの有効期限
	TTL time.Duration `mapstructure:"ttl" yaml:"ttl"`

	// Dir はファイルキャッシュのディレクトリ
	Dir string `mapstructure:"dir" yaml:"dir"`

	// MaxSize はメモリキャッシュの最大サイズ（バイト）
	MaxSize int64 `mapstructure:"max_size" yaml:"max_size"`

	// UseFileCache はファイルキャッシュの使用有無
	UseFileCache bool `mapstructure:"use_file_cache" yaml:"use_file_cache"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		GitHub: GitHubConfig{
			Token:           "",
			DefaultOwner:    "",
			DefaultRepo:     "",
			APIBaseURL:      "https://api.github.com/",
			UploadBaseURL:   "https://uploads.github.com/",
			RequestTimeout:  30 * time.Second,
			RateLimitBuffer: 10,
			Repositories:    []string{},
		},
		UI: UIConfig{
			Theme:       "auto",
			DefaultView: "issues",
			KeyBindings: map[string]string{
				"quit":       "q",
				"up":         "k",
				"down":       "j",
				"select":     "enter",
				"back":       "esc",
				"refresh":    "r",
				"search":     "/",
				"filter":     "f",
				"help":       "?",
				"next_view":  "tab",
				"prev_view":  "shift+tab",
				"first_item": "g",
				"last_item":  "G",
				"page_up":    "ctrl+u",
				"page_down":  "ctrl+d",
				"new_issue":  "n",
				"edit":       "e",
				"comment":    "c",
				"assign":     "a",
				"label":      "l",
				"close":      "x",
				"open":       "o",
			},
			PageSize:   50,
			ShowIcons:  true,
			DateFormat: "2006-01-02 15:04",
		},
		Cache: CacheConfig{
			Enabled:      true,
			TTL:          15 * time.Minute,
			Dir:          "",                // will be set to ~/.cache/tig-gh
			MaxSize:      100 * 1024 * 1024, // 100MB
			UseFileCache: true,
		},
		Metrics: MetricsConfig{
			Enabled:           false,
			LeadTimeEnabled:   false,
			CalculationPeriod: 30 * 24 * time.Hour,
		},
	}
}

// Validate は設定の妥当性を検証する
func (c *Config) Validate() error {
	// GitHub設定の検証
	if c.GitHub.Token == "" {
		// トークンが空の場合は警告レベル（後で環境変数から取得される可能性がある）
		// 実際の使用時にチェックされるべき
	}

	if c.GitHub.RequestTimeout <= 0 {
		c.GitHub.RequestTimeout = 30 * time.Second
	}

	if c.GitHub.RateLimitBuffer < 0 {
		c.GitHub.RateLimitBuffer = 10
	}
	if c.GitHub.Repositories == nil {
		c.GitHub.Repositories = []string{}
	}

	// UI設定の検証
	if c.UI.Theme == "" {
		c.UI.Theme = "auto"
	}

	if c.UI.DefaultView == "" {
		c.UI.DefaultView = "issues"
	}

	if c.UI.PageSize <= 0 {
		c.UI.PageSize = 50
	}

	if c.UI.DateFormat == "" {
		c.UI.DateFormat = "2006-01-02 15:04"
	}

	// Cache設定の検証
	if c.Cache.TTL <= 0 {
		c.Cache.TTL = 15 * time.Minute
	}

	if c.Cache.MaxSize <= 0 {
		c.Cache.MaxSize = 100 * 1024 * 1024 // 100MB
	}

	// Metrics 設定の検証
	if c.Metrics.CalculationPeriod <= 0 {
		c.Metrics.CalculationPeriod = 30 * 24 * time.Hour
	}

	return nil
}
