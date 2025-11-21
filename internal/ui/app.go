package ui

import (
	"github.com/a1yama/tig-gh/internal/app/usecase"
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewType represents the current active view
type ViewType int

const (
	IssueListView ViewType = iota
	PullRequestListView
	CommitListView
	SearchView
	ReviewQueueView
	MetricsView
)

// App is the main application model
type App struct {
	currentView         ViewType
	issueView           tea.Model
	prView              tea.Model
	prQueueView         tea.Model
	commitView          tea.Model
	searchView          tea.Model
	metricsView         tea.Model
	fetchIssuesUseCase  *usecase.FetchIssuesUseCase
	fetchPRsUseCase     *usecase.FetchPRsUseCase
	fetchCommitsUseCase *usecase.FetchCommitsUseCase
	searchUseCase       *usecase.SearchUseCase
	fetchMetricsUseCase *usecase.FetchLeadTimeMetricsUseCase
	owner               string
	repo                string
	width               int
	height              int
	ready               bool
	issueViewInited     bool
	prViewInited        bool
	prQueueViewInited   bool
	commitViewInited    bool
	searchViewInited    bool
	metricsViewInited   bool
	lastPrimaryView     ViewType
}

// NewApp creates a new application instance (for backward compatibility)
func NewApp() *App {
	return &App{
		currentView:     IssueListView,
		issueView:       views.NewIssueView(),
		prView:          views.NewPRView(),
		prQueueView:     views.NewPRQueueView(),
		commitView:      views.NewCommitView(),
		metricsView:     views.NewMetricsView(),
		owner:           "",
		repo:            "",
		ready:           false,
		lastPrimaryView: IssueListView,
	}
}

// NewAppWithUseCases creates a new application instance with all UseCases
func NewAppWithUseCases(
	fetchIssuesUseCase *usecase.FetchIssuesUseCase,
	fetchPRsUseCase *usecase.FetchPRsUseCase,
	fetchCommitsUseCase *usecase.FetchCommitsUseCase,
	searchUseCase *usecase.SearchUseCase,
	fetchMetricsUseCase *usecase.FetchLeadTimeMetricsUseCase,
	owner, repo string,
	defaultView string,
	metricsConfig *models.MetricsConfig,
) *App {
	// デフォルトビューを決定
	var initialView ViewType
	switch defaultView {
	case "prs", "pull_requests":
		initialView = PullRequestListView
	case "commits":
		initialView = CommitListView
	default:
		initialView = IssueListView
	}

	return &App{
		currentView:         initialView,
		issueView:           views.NewIssueViewWithUseCase(fetchIssuesUseCase, owner, repo),
		prView:              views.NewPRViewWithUseCase(fetchPRsUseCase, owner, repo),
		prQueueView:         views.NewPRQueueViewWithUseCase(fetchPRsUseCase, owner, repo),
		commitView:          views.NewCommitViewWithUseCase(fetchCommitsUseCase, owner, repo),
		searchView:          views.NewSearchViewWithUseCase(searchUseCase, owner, repo),
		metricsView:         views.NewMetricsViewWithUseCase(fetchMetricsUseCase, metricsConfig),
		fetchIssuesUseCase:  fetchIssuesUseCase,
		fetchPRsUseCase:     fetchPRsUseCase,
		fetchCommitsUseCase: fetchCommitsUseCase,
		searchUseCase:       searchUseCase,
		fetchMetricsUseCase: fetchMetricsUseCase,
		owner:               owner,
		repo:                repo,
		ready:               false,
		lastPrimaryView:     initialView,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	switch a.currentView {
	case PullRequestListView:
		a.prViewInited = true
		return a.prView.Init()
	case CommitListView:
		a.commitViewInited = true
		return a.commitView.Init()
	default:
		a.issueViewInited = true
		return a.issueView.Init()
	}
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case views.MetricsExitMsg:
		if a.currentView == MetricsView {
			a.currentView = a.lastPrimaryView
		}
		return a, nil

	case tea.KeyMsg:
		// Check if we're in search view with input focused
		// If so, skip global key bindings except for special cases
		if a.currentView == SearchView {
			if searchView, ok := a.searchView.(*views.SearchView); ok {
				if searchView.IsInputFocused() {
					// Only handle Ctrl+C when input is focused
					if msg.String() == "ctrl+c" {
						return a, tea.Quit
					}
					// Delegate everything else to search view
					return a.delegateToCurrentView(msg)
				}
			}
		}

		// Global key bindings
		switch msg.String() {
		case "ctrl+c", "q":
			// Only quit if not showing help or in a subview
			// For now, we'll let the views handle their own quit logic
			return a.delegateToCurrentView(msg)

		case "i":
			// Switch to issue view
			a.currentView = IssueListView
			if !a.issueViewInited {
				a.issueViewInited = true
				return a, a.issueView.Init()
			}
			return a, nil

		case "p":
			// Switch to PR view
			a.currentView = PullRequestListView
			if !a.prViewInited {
				a.prViewInited = true
				return a, a.prView.Init()
			}
			return a, nil

		case "R":
			// Switch to review queue view
			a.currentView = ReviewQueueView
			if !a.prQueueViewInited {
				a.prQueueViewInited = true
				return a, a.prQueueView.Init()
			}
			return a, nil

		case "m":
			if a.currentView != MetricsView {
				a.lastPrimaryView = a.currentView
			}
			a.currentView = MetricsView
			if !a.metricsViewInited {
				a.metricsViewInited = true
				return a, a.metricsView.Init()
			}
			return a, nil

		case "c":
			// Switch to commit view
			a.currentView = CommitListView
			if !a.commitViewInited {
				a.commitViewInited = true
				return a, a.commitView.Init()
			}
			return a, nil

		case "/":
			// Switch to search view
			a.currentView = SearchView
			if !a.searchViewInited {
				a.searchViewInited = true
				return a, a.searchView.Init()
			}
			return a, nil

		default:
			// Delegate to current view
			return a.delegateToCurrentView(msg)
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true

		// Propagate size to all views
		a.issueView, cmd = a.issueView.Update(msg)
		cmds = append(cmds, cmd)

		a.prView, cmd = a.prView.Update(msg)
		cmds = append(cmds, cmd)

		a.prQueueView, cmd = a.prQueueView.Update(msg)
		cmds = append(cmds, cmd)

		a.commitView, cmd = a.commitView.Update(msg)
		cmds = append(cmds, cmd)

		a.searchView, cmd = a.searchView.Update(msg)
		cmds = append(cmds, cmd)

		a.metricsView, cmd = a.metricsView.Update(msg)
		cmds = append(cmds, cmd)

		return a, tea.Batch(cmds...)

	default:
		// Delegate other messages to current view
		return a.delegateToCurrentView(msg)
	}
}

// delegateToCurrentView delegates the message to the current active view
func (a *App) delegateToCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch a.currentView {
	case IssueListView:
		a.issueView, cmd = a.issueView.Update(msg)
		return a, cmd

	case PullRequestListView:
		a.prView, cmd = a.prView.Update(msg)
		return a, cmd

	case ReviewQueueView:
		a.prQueueView, cmd = a.prQueueView.Update(msg)
		return a, cmd

	case CommitListView:
		a.commitView, cmd = a.commitView.Update(msg)
		return a, cmd

	case SearchView:
		a.searchView, cmd = a.searchView.Update(msg)
		return a, cmd

	case MetricsView:
		a.metricsView, cmd = a.metricsView.Update(msg)
		return a, cmd

	default:
		return a, nil
	}
}

// View renders the application
func (a *App) View() string {
	if !a.ready {
		return "Initializing tig-gh..."
	}

	switch a.currentView {
	case IssueListView:
		return a.issueView.View()

	case PullRequestListView:
		return a.prView.View()

	case ReviewQueueView:
		return a.prQueueView.View()

	case CommitListView:
		return a.commitView.View()

	case SearchView:
		return a.searchView.View()

	case MetricsView:
		return a.metricsView.View()

	default:
		return "Unknown view"
	}
}

// Helper methods

// GetCurrentView returns the current active view type
func (a *App) GetCurrentView() ViewType {
	return a.currentView
}

// SetCurrentView sets the current active view
func (a *App) SetCurrentView(view ViewType) {
	a.currentView = view
}

// IsReady returns whether the app is ready to display
func (a *App) IsReady() bool {
	return a.ready
}

// GetSize returns the current terminal size
func (a *App) GetSize() (int, int) {
	return a.width, a.height
}
