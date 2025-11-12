package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// searchResultsLoadedMsg is sent when search results are loaded
type searchResultsLoadedMsg struct {
	results *models.SearchResults
	err     error
}

// SearchUseCase defines the interface for search operations
type SearchUseCase interface {
	Execute(ctx context.Context, owner, repo string, opts *models.SearchOptions) (*models.SearchResults, error)
	GetRepository() repository.SearchRepository
}

// SearchView is the model for the search view
type SearchView struct {
	searchUseCase SearchUseCase
	owner         string
	repo          string
	textInput     textinput.Model
	results       []models.SearchResult
	cursor        int
	loading       bool
	err           error
	width         int
	height        int
	statusBar     *components.StatusBar
	searchType    models.SearchType
	searchState   models.IssueState
	detailView    tea.Model // Can be IssueDetailView or PRDetailView
	showingDetail bool
}

// NewSearchView creates a new search view
func NewSearchView() *SearchView {
	ti := textinput.New()
	ti.Placeholder = "Search issues and pull requests..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return &SearchView{
		textInput:   ti,
		results:     []models.SearchResult{},
		cursor:      0,
		loading:     false,
		statusBar:   components.NewStatusBar(),
		searchType:  models.SearchTypeBoth,
		searchState: models.IssueStateOpen,
	}
}

// NewSearchViewWithUseCase creates a new search view with UseCase
func NewSearchViewWithUseCase(searchUseCase SearchUseCase, owner, repo string) *SearchView {
	view := NewSearchView()
	view.searchUseCase = searchUseCase
	view.owner = owner
	view.repo = repo
	return view
}

// Init initializes the search view
func (m *SearchView) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m *SearchView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If showing detail view, delegate to detail view
	if m.showingDetail && m.detailView != nil {
		// Check for back message
		if _, isBackMsg := msg.(backMsg); isBackMsg {
			m.showingDetail = false
			m.detailView = nil
			return m, nil
		}

		// Delegate to detail view
		var cmd tea.Cmd
		m.detailView, cmd = m.detailView.Update(msg)

		// Check if it's a KeyMsg for back navigation
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			keyStr := keyMsg.String()
			if keyStr == "q" || keyStr == "esc" {
				m.showingDetail = false
				m.detailView = nil
				return m, nil
			}
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case searchResultsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.results = []models.SearchResult{}
		} else {
			m.err = nil
			m.results = msg.results.Items
			// Reset cursor if out of bounds
			if m.cursor >= len(m.results) && len(m.results) > 0 {
				m.cursor = len(m.results) - 1
			} else if len(m.results) == 0 {
				m.cursor = 0
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(msg.Width, 1)
		m.textInput.Width = msg.Width - 20
		if m.detailView != nil {
			m.detailView.Update(msg)
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *SearchView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle text input first when focused
	if m.textInput.Focused() {
		switch msg.String() {
		case "esc":
			m.textInput.Blur()
			return m, nil
		case "enter":
			// Perform search without blurring
			return m, m.performSearch()
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	// Handle navigation and commands when text input is not focused
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc":
		if m.showingDetail {
			m.showingDetail = false
			m.detailView = nil
			return m, nil
		}
		return m, nil

	case "enter":
		// View detail of selected result
		if len(m.results) > 0 && m.cursor < len(m.results) {
			return m, m.showDetail()
		}
		return m, nil

	case "r":
		// Re-run current search
		if !m.loading {
			return m, m.performSearch()
		}
		return m, nil

	case "t":
		// Toggle search type
		m.toggleSearchType()
		return m, m.performSearch()

	case "s":
		// Toggle search state
		m.toggleSearchState()
		return m, m.performSearch()

	case "j", "down":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "g":
		m.cursor = 0
		return m, nil

	case "G":
		if len(m.results) > 0 {
			m.cursor = len(m.results) - 1
		}
		return m, nil
	}

	return m, nil
}

// performSearch executes a search
func (m *SearchView) performSearch() tea.Cmd {
	return func() tea.Msg {
		if m.searchUseCase == nil {
			return searchResultsLoadedMsg{
				results: nil,
				err:     fmt.Errorf("search use case not initialized"),
			}
		}

		m.loading = true
		query := m.textInput.Value()

		opts := &models.SearchOptions{
			Query:     query,
			Type:      m.searchType,
			State:     m.searchState,
			Sort:      models.SearchSortUpdated,
			Direction: models.SortDirectionDesc,
			PerPage:   50,
			Page:      1,
		}

		results, err := m.searchUseCase.Execute(context.Background(), m.owner, m.repo, opts)
		return searchResultsLoadedMsg{
			results: results,
			err:     err,
		}
	}
}

// showDetail shows the detail view for the selected result
func (m *SearchView) showDetail() tea.Cmd {
	if m.cursor >= len(m.results) {
		return nil
	}

	result := m.results[m.cursor]

	switch result.Type {
	case models.SearchTypeIssue:
		if result.Issue != nil {
			m.detailView = NewIssueDetailView(result.Issue, m.owner, m.repo, nil)
			if issueView, ok := m.detailView.(*IssueDetailView); ok {
				issueView.width = m.width
				issueView.height = m.height
			}
			m.showingDetail = true
			return m.detailView.Init()
		}
	case models.SearchTypePR:
		if result.PullRequest != nil {
			ensurePRNumber(result.PullRequest)
			m.detailView = NewPRDetailView(result.PullRequest, m.owner, m.repo, nil)
			if prView, ok := m.detailView.(*PRDetailView); ok {
				prView.width = m.width
				prView.height = m.height
			}
			m.showingDetail = true
			return m.detailView.Init()
		}
	}

	return nil
}

// toggleSearchType toggles between Issue, PR, and Both
func (m *SearchView) toggleSearchType() {
	switch m.searchType {
	case models.SearchTypeBoth:
		m.searchType = models.SearchTypeIssue
	case models.SearchTypeIssue:
		m.searchType = models.SearchTypePR
	case models.SearchTypePR:
		m.searchType = models.SearchTypeBoth
	}
}

// toggleSearchState toggles between Open, Closed, and All
func (m *SearchView) toggleSearchState() {
	switch m.searchState {
	case models.IssueStateOpen:
		m.searchState = models.IssueStateClosed
	case models.IssueStateClosed:
		m.searchState = models.IssueStateAll
	case models.IssueStateAll:
		m.searchState = models.IssueStateOpen
	}
}

// View renders the search view
func (m *SearchView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// If showing detail view, render it
	if m.showingDetail && m.detailView != nil {
		return m.detailView.View()
	}

	var s strings.Builder

	// Header
	header := m.renderHeader()
	s.WriteString(header)
	s.WriteString("\n\n")

	// Search input
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")

	// Results or loading/error state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else {
		s.WriteString(m.renderResults())
	}

	// Status bar
	s.WriteString("\n")
	m.updateStatusBar()
	s.WriteString(m.statusBar.View())

	return s.String()
}

// renderHeader renders the view header
func (m *SearchView) renderHeader() string {
	title := styles.HeaderStyle.Render("Search")

	// Show current filters
	typeFilter := fmt.Sprintf("Type: %s", m.searchType)
	stateFilter := fmt.Sprintf("State: %s", m.searchState)
	filters := styles.MutedStyle.Render(fmt.Sprintf("[%s] [%s]", typeFilter, stateFilter))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		" ",
		filters,
	)
}

// renderResults renders the search results
func (m *SearchView) renderResults() string {
	if len(m.results) == 0 {
		return styles.MutedStyle.Render("No results found. Enter query and press 'enter' to search.")
	}

	var s strings.Builder

	// Calculate available height
	availableHeight := m.height - 10
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Calculate visible range
	startIdx := 0
	endIdx := len(m.results)

	if len(m.results) > availableHeight {
		halfHeight := availableHeight / 2
		startIdx = m.cursor - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.results) {
			endIdx = len(m.results)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible results
	for i := startIdx; i < endIdx; i++ {
		result := m.results[i]
		line := m.renderResultLine(result, i)
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderResultLine renders a single result line
func (m *SearchView) renderResultLine(result models.SearchResult, index int) string {
	// Cursor indicator
	cursor := "  "
	if m.cursor == index {
		cursor = styles.CursorStyle.Render("▶ ")
	}

	var number int
	var title string
	var state string
	var typeIcon string

	switch result.Type {
	case models.SearchTypeIssue:
		if result.Issue != nil {
			number = result.Issue.Number
			title = result.Issue.Title
			state = string(result.Issue.State)
			typeIcon = "📄"
		}
	case models.SearchTypePR:
		if result.PullRequest != nil {
			number = result.PullRequest.Number
			title = result.PullRequest.Title
			state = string(result.PullRequest.State)
			typeIcon = "🔀"
		}
	}

	// State badge
	stateBadge := styles.GetStateBadge(state)

	// Number
	numberStr := styles.IssueNumberStyle.Render(fmt.Sprintf("#%-5d", number))

	// Title
	titleStyle := styles.IssueTitleStyle
	if m.cursor == index {
		titleStyle = styles.SelectedStyle
	}
	titleStr := titleStyle.Render(title)

	// Combine all parts
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		cursor,
		typeIcon,
		" ",
		stateBadge,
		" ",
		numberStr,
		" ",
		titleStr,
	)
}

// renderLoading renders a loading state
func (m *SearchView) renderLoading() string {
	return styles.LoadingStyle.Render("Searching...")
}

// renderError renders an error state
func (m *SearchView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// updateStatusBar updates the status bar with current state
func (m *SearchView) updateStatusBar() {
	m.statusBar.ClearItems()

	// Set mode
	m.statusBar.SetMode("Search")

	// Add position if results exist
	if len(m.results) > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.results))
		m.statusBar.AddItem("", position)
	}

	// Add repository info
	if m.owner != "" && m.repo != "" {
		m.statusBar.AddItem("Repo", fmt.Sprintf("%s/%s", m.owner, m.repo))
	}

	// Add help
	if m.textInput.Focused() {
		m.statusBar.AddItem("", "esc: blur • enter: search")
	} else {
		m.statusBar.AddItem("", "t: type • s: state • enter: view • r: refresh • i: issues • p: prs • c: commits • q: quit")
	}
}

// IsInputFocused returns true if the text input is focused
func (m *SearchView) IsInputFocused() bool {
	return m.textInput.Focused()
}
