package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a1yama/tig-gh/internal/app/usecase"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/infra/cache"
	"github.com/a1yama/tig-gh/internal/infra/config"
	"github.com/a1yama/tig-gh/internal/infra/git"
	"github.com/a1yama/tig-gh/internal/infra/github"
	"github.com/a1yama/tig-gh/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("tig-gh version %s\n", Version)
		os.Exit(0)
	}

	// 設定を読み込む
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration...\n")
	}

	cfg := config.Get()

	// GitHub トークンを取得
	token := config.GetGitHubToken()
	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: GitHub token not found.\n")
		fmt.Fprintf(os.Stderr, "Please set GITHUB_TOKEN environment variable or configure it in ~/.config/tig-gh/config.yaml\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  export GITHUB_TOKEN=ghp_xxxxxxxxxxxx\n")
		fmt.Fprintf(os.Stderr, "\nOr create ~/.config/tig-gh/config.yaml with:\n")
		fmt.Fprintf(os.Stderr, "  github:\n")
		fmt.Fprintf(os.Stderr, "    token: ghp_xxxxxxxxxxxx\n")
		os.Exit(1)
	}

	var owner, repo string
	var err error

	// コマンドライン引数からowner/repoを取得
	if len(os.Args) > 1 && os.Args[1] != "--version" && os.Args[1] != "-v" {
		// owner/repo形式のパース
		arg := os.Args[1]
		parts := strings.Split(arg, "/")
		if len(parts) == 2 {
			owner = parts[0]
			repo = parts[1]
		} else {
			fmt.Fprintf(os.Stderr, "Error: Invalid repository format.\n")
			fmt.Fprintf(os.Stderr, "Usage: tig-gh [owner/repo]\n")
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  tig-gh charmbracelet/bubbletea\n")
			os.Exit(1)
		}
	} else {
		// 引数がない場合は現在のGitリポジトリから取得
		if !git.IsGitRepository() {
			fmt.Fprintf(os.Stderr, "Error: Not a git repository.\n")
			fmt.Fprintf(os.Stderr, "Please run tig-gh from within a git repository or specify a repository:\n")
			fmt.Fprintf(os.Stderr, "\nUsage:\n")
			fmt.Fprintf(os.Stderr, "  tig-gh [owner/repo]\n")
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  tig-gh charmbracelet/bubbletea\n")
			os.Exit(1)
		}

		owner, repo, err = git.GetCurrentRepository()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get repository information: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nMake sure the current directory is a GitHub repository with a valid remote 'origin'.\n")
			fmt.Fprintf(os.Stderr, "Or specify a repository manually:\n")
			fmt.Fprintf(os.Stderr, "\nUsage:\n")
			fmt.Fprintf(os.Stderr, "  tig-gh [owner/repo]\n")
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  tig-gh charmbracelet/bubbletea\n")
			os.Exit(1)
		}
	}

	// owner/repoが取得できなかった場合（設定ファイルからのフォールバック）
	if owner == "" || repo == "" {
		owner = cfg.GitHub.DefaultOwner
		repo = cfg.GitHub.DefaultRepo
	}

	// それでもowner/repoが設定されていない場合はエラー
	if owner == "" || repo == "" {
		fmt.Fprintf(os.Stderr, "Error: Repository not specified.\n")
		fmt.Fprintf(os.Stderr, "Please run tig-gh from within a git repository or specify a repository:\n")
		fmt.Fprintf(os.Stderr, "\nUsage:\n")
		fmt.Fprintf(os.Stderr, "  tig-gh [owner/repo]\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  tig-gh charmbracelet/bubbletea\n")
		os.Exit(1)
	}

	// GitHub クライアントの初期化
	githubClient := github.NewClient(token)

	// キャッシュの初期化
	var cacheService repository.CacheService
	if cfg.Cache.Enabled {
		cacheConfig := cache.DefaultConfig()
		if cfg.Cache.TTL > 0 {
			cacheConfig.MemoryTTL = cfg.Cache.TTL
			cacheConfig.FileTTL = cfg.Cache.TTL
		}
		if dir := strings.TrimSpace(cfg.Cache.Dir); dir != "" {
			cacheConfig.FileDir = expandPath(dir)
		}
		if !cfg.Cache.UseFileCache {
			cacheConfig.FileEnabled = false
		}

		cacheService, err = cache.NewCacheWithConfig(cacheConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize cache: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing without cache...\n")
			cacheService = nil
		}
	}

	// リポジトリの初期化（キャッシュあり）
	baseIssueRepo := github.NewIssueRepository(githubClient)
	basePRRepo := github.NewPullRequestRepository(githubClient)
	commitRepo := github.NewCommitRepository(githubClient)
	searchRepo := github.NewSearchRepository(githubClient)
	metricsRepo := github.NewMetricsRepository(githubClient)

	// キャッシュでラップ
	var issueRepo repository.IssueRepository
	var prRepo repository.PullRequestRepository

	if cacheService != nil {
		c := cacheService.(*cache.Cache)
		issueRepo = cache.NewCachedIssueRepository(baseIssueRepo, c)
		prRepo = cache.NewCachedPullRequestRepository(basePRRepo, c)
	} else {
		issueRepo = baseIssueRepo
		prRepo = basePRRepo
	}

	// UseCaseの初期化
	fetchIssuesUseCase := usecase.NewFetchIssuesUseCase(issueRepo)
	fetchPRsUseCase := usecase.NewFetchPRsUseCase(prRepo)
	fetchCommitsUseCase := usecase.NewFetchCommitsUseCase(commitRepo)
	searchUseCase := usecase.NewSearchUseCase(searchRepo)
	fetchMetricsUseCase := usecase.NewFetchLeadTimeMetricsUseCase(metricsRepo, cfg)

	// TUIアプリケーションの初期化
	app := ui.NewAppWithUseCases(
		fetchIssuesUseCase,
		fetchPRsUseCase,
		fetchCommitsUseCase,
		searchUseCase,
		fetchMetricsUseCase,
		owner,
		repo,
		cfg.UI.DefaultView,
		&cfg.Metrics,
	)

	// bubbletea プログラムの起動
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(), // Disabled: may cause rendering issues
	)

	// アプリケーション起動メッセージ
	fmt.Fprintf(os.Stderr, "Starting tig-gh for %s/%s...\n", owner, repo)

	// 実行
	ctx := context.Background()
	_ = ctx // 将来的にコンテキストを使う

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func expandPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}
