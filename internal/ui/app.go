package ui

import (
	"github.com/a1yama/tig-gh/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewType represents the current active view
type ViewType int

const (
	IssueListView ViewType = iota
	PullRequestListView
	CommitListView
)

// App is the main application model
type App struct {
	currentView ViewType
	issueView   tea.Model
	width       int
	height      int
	ready       bool
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{
		currentView: IssueListView,
		issueView:   views.NewIssueView(),
		ready:       false,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		a.issueView.Init(),
	)
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c", "q":
			// Only quit if not showing help or in a subview
			// For now, we'll let the views handle their own quit logic
			return a.delegateToCurrentView(msg)

		case "1":
			// Switch to issue view
			a.currentView = IssueListView
			return a, nil

		case "2":
			// Switch to PR view (not implemented yet)
			a.currentView = PullRequestListView
			return a, nil

		case "3":
			// Switch to commit view (not implemented yet)
			a.currentView = CommitListView
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
		// Not implemented yet
		return a, nil

	case CommitListView:
		// Not implemented yet
		return a, nil

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
		return "Pull Request View (Not Implemented)"

	case CommitListView:
		return "Commit View (Not Implemented)"

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
