package views

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// mockFetchIssuesUseCase is a mock implementation of FetchIssuesUseCase for testing
type mockFetchIssuesUseCase struct {
	executeFunc func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error)
}

func (m *mockFetchIssuesUseCase) Execute(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, owner, repo, opts)
	}
	return nil, nil
}

func TestIssueView_Init(t *testing.T) {
	tests := []struct {
		name          string
		mockUseCase   *mockFetchIssuesUseCase
		checkCmd      func(t *testing.T, cmd tea.Cmd)
		expectLoading bool
	}{
		{
			name: "init triggers issue fetch",
			mockUseCase: &mockFetchIssuesUseCase{
				executeFunc: func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
					return []*models.Issue{
						{
							Number:    1,
							Title:     "Test Issue",
							State:     models.IssueStateOpen,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil
				},
			},
			checkCmd: func(t *testing.T, cmd tea.Cmd) {
				if cmd == nil {
					t.Error("expected non-nil command from Init()")
				}
			},
			expectLoading: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := NewIssueViewWithUseCase(tt.mockUseCase, "testowner", "testrepo")
			cmd := view.Init()

			if tt.checkCmd != nil {
				tt.checkCmd(t, cmd)
			}

			if view.loading != tt.expectLoading {
				t.Errorf("expected loading=%v, got %v", tt.expectLoading, view.loading)
			}
		})
	}
}

func TestIssueView_Update_IssuesLoaded(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		msg           tea.Msg
		initialState  *IssueView
		checkState    func(t *testing.T, view *IssueView)
		expectLoading bool
	}{
		{
			name: "successful data load",
			msg: issuesLoadedMsg{
				issues: []*models.Issue{
					{
						Number:    1,
						Title:     "Test Issue 1",
						State:     models.IssueStateOpen,
						Author:    models.User{Login: "alice"},
						Labels:    []models.Label{{Name: "bug"}},
						Comments:  5,
						CreatedAt: now,
						UpdatedAt: now.Add(-1 * time.Hour),
					},
					{
						Number:    2,
						Title:     "Test Issue 2",
						State:     models.IssueStateClosed,
						Author:    models.User{Login: "bob"},
						Labels:    []models.Label{{Name: "feature"}},
						Comments:  3,
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				err: nil,
			},
			initialState: &IssueView{
				loading:  true,
				issues:   []*models.Issue{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *IssueView) {
				if view.loading {
					t.Error("expected loading to be false after successful load")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.issues) != 2 {
					t.Errorf("expected 2 issues, got %d", len(view.issues))
				}
				if view.issues[0].Title != "Test Issue 2" {
					t.Errorf("expected first issue title 'Test Issue 2', got '%s'", view.issues[0].Title)
				}
			},
			expectLoading: false,
		},
		{
			name: "error during load",
			msg: issuesLoadedMsg{
				issues: nil,
				err:    errors.New("API error"),
			},
			initialState: &IssueView{
				loading:  true,
				issues:   []*models.Issue{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *IssueView) {
				if view.loading {
					t.Error("expected loading to be false after error")
				}
				if view.err == nil {
					t.Error("expected error to be set")
				}
				if view.err.Error() != "API error" {
					t.Errorf("expected error 'API error', got '%v'", view.err)
				}
				if len(view.issues) != 0 {
					t.Errorf("expected 0 issues on error, got %d", len(view.issues))
				}
			},
			expectLoading: false,
		},
		{
			name: "empty result",
			msg: issuesLoadedMsg{
				issues: []*models.Issue{},
				err:    nil,
			},
			initialState: &IssueView{
				loading:  true,
				issues:   []*models.Issue{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *IssueView) {
				if view.loading {
					t.Error("expected loading to be false")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.issues) != 0 {
					t.Errorf("expected 0 issues, got %d", len(view.issues))
				}
			},
			expectLoading: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.initialState
			_, _ = view.Update(tt.msg)

			if tt.checkState != nil {
				tt.checkState(t, view)
			}

			if view.loading != tt.expectLoading {
				t.Errorf("expected loading=%v, got %v", tt.expectLoading, view.loading)
			}
		})
	}
}

func TestIssueView_Update_RefreshKey(t *testing.T) {
	callCount := 0
	mockUseCase := &mockFetchIssuesUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
			callCount++
			return []*models.Issue{
				{
					Number: callCount,
					Title:  "Refreshed Issue",
					State:  models.IssueStateOpen,
				},
			}, nil
		},
	}

	view := NewIssueViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false // Set to not loading so refresh can trigger

	// Simulate 'r' key press for refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := view.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command for refresh")
	}

	if !view.loading {
		t.Error("expected loading to be true during refresh")
	}
}

func TestIssueView_Update_FilterKey(t *testing.T) {
	mockUseCase := &mockFetchIssuesUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
			return []*models.Issue{}, nil
		},
	}

	view := NewIssueViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false

	// Simulate 'f' key press for filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd := view.Update(msg)

	// Filter functionality should toggle filtering mode
	// For now, we just check that it doesn't crash
	if cmd != nil && view.err != nil {
		t.Errorf("unexpected error during filter: %v", view.err)
	}
}

func TestIssueView_View_States(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		view        *IssueView
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "loading state",
			view: &IssueView{
				loading:     true,
				issues:      []*models.Issue{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.IssueStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Loading") {
					t.Error("expected loading message in output")
				}
			},
		},
		{
			name: "error state",
			view: &IssueView{
				loading:     false,
				err:         errors.New("API error"),
				issues:      []*models.Issue{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.IssueStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Error") {
					t.Error("expected error message in output")
				}
				if !strings.Contains(output, "API error") {
					t.Error("expected specific error message in output")
				}
			},
		},
		{
			name: "data display state",
			view: &IssueView{
				loading: false,
				issues: []*models.Issue{
					{
						Number:    1,
						Title:     "Test Issue",
						State:     models.IssueStateOpen,
						Author:    models.User{Login: "alice"},
						Labels:    []models.Label{{Name: "bug"}},
						Comments:  5,
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.IssueStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test Issue") {
					t.Error("expected issue title in output")
				}
				if !strings.Contains(output, "#1") {
					t.Error("expected issue number in output")
				}
			},
		},
		{
			name: "empty data state",
			view: &IssueView{
				loading:     false,
				issues:      []*models.Issue{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.IssueStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				// Should show header but no issues
				if !strings.Contains(output, "Issues") {
					t.Error("expected 'Issues' header in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.view.View()

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestIssueView_Navigation(t *testing.T) {
	mockUseCase := &mockFetchIssuesUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
			return []*models.Issue{
				{Number: 1, Title: "Issue 1", State: models.IssueStateOpen},
				{Number: 2, Title: "Issue 2", State: models.IssueStateOpen},
				{Number: 3, Title: "Issue 3", State: models.IssueStateOpen},
			}, nil
		},
	}

	tests := []struct {
		name           string
		initialCursor  int
		key            string
		expectedCursor int
	}{
		{
			name:           "move down with j",
			initialCursor:  0,
			key:            "j",
			expectedCursor: 1,
		},
		{
			name:           "move down with down arrow",
			initialCursor:  0,
			key:            "down",
			expectedCursor: 1,
		},
		{
			name:           "move up with k",
			initialCursor:  2,
			key:            "k",
			expectedCursor: 1,
		},
		{
			name:           "move up with up arrow",
			initialCursor:  2,
			key:            "up",
			expectedCursor: 1,
		},
		{
			name:           "go to top with g",
			initialCursor:  2,
			key:            "g",
			expectedCursor: 0,
		},
		{
			name:           "go to bottom with G",
			initialCursor:  0,
			key:            "G",
			expectedCursor: 2,
		},
		{
			name:           "cannot move down past last item",
			initialCursor:  2,
			key:            "j",
			expectedCursor: 2,
		},
		{
			name:           "cannot move up past first item",
			initialCursor:  0,
			key:            "k",
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := NewIssueViewWithUseCase(mockUseCase, "testowner", "testrepo")
			// Load initial data
			view.loading = false
			view.issues = []*models.Issue{
				{Number: 1, Title: "Issue 1", State: models.IssueStateOpen},
				{Number: 2, Title: "Issue 2", State: models.IssueStateOpen},
				{Number: 3, Title: "Issue 3", State: models.IssueStateOpen},
			}
			view.cursor = tt.initialCursor

			var msg tea.Msg
			if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			view.Update(msg)

			if view.cursor != tt.expectedCursor {
				t.Errorf("expected cursor at %d, got %d", tt.expectedCursor, view.cursor)
			}
		})
	}
}

func TestIssueView_ShowDetailOnEnter(t *testing.T) {
	view := NewIssueViewWithUseCase(nil, "testowner", "testrepo")
	view.loading = false
	view.width = 80
	view.height = 24
	view.issues = []*models.Issue{
		{
			Number: 1,
			Title:  "Test Title",
			Body:   "Body",
			State:  models.IssueStateOpen,
			Author: models.User{Login: "alice"},
		},
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	view.Update(msg)

	if !view.showingDetail {
		t.Fatal("expected showingDetail to be true after Enter")
	}

	if view.detailView == nil {
		t.Fatal("expected detailView to be initialized")
	}

	output := view.View()
	if !strings.Contains(output, "Test Title") {
		t.Fatalf("detail view output missing issue title: %s", output)
	}
}

func TestFilterOutPullRequests(t *testing.T) {
	issues := []*models.Issue{
		{Number: 1, HTMLURL: "https://github.com/org/repo/issues/1"},
		{Number: 2, HTMLURL: "https://github.com/org/repo/pull/2"},
		{Number: 3, HTMLURL: ""},
	}

	filtered := filterOutPullRequests(issues)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 issues after filtering, got %d", len(filtered))
	}

	if filtered[0].Number != 1 || filtered[1].Number != 3 {
		t.Fatalf("unexpected issues after filtering: %+v", filtered)
	}
}

func TestSortIssues(t *testing.T) {
	now := time.Now()
	issues := []*models.Issue{
		{Number: 3, Title: "older", UpdatedAt: now.Add(-2 * time.Hour)},
		{Number: 10, Title: "newest", UpdatedAt: now.Add(-1 * time.Minute)},
		{Number: 5, Title: "same time", UpdatedAt: now.Add(-30 * time.Minute)},
		{Number: 7, Title: "same time higher number", UpdatedAt: now.Add(-30 * time.Minute)},
	}

	sorted := sortIssues(issues)

	if sorted[0].Number != 10 {
		t.Fatalf("expected newest issue first, got %d", sorted[0].Number)
	}

	if sorted[1].Number != 7 || sorted[2].Number != 5 {
		t.Fatalf("expected tie broken by issue number desc, got %d then %d", sorted[1].Number, sorted[2].Number)
	}

	if sorted[3].Number != 3 {
		t.Fatalf("expected oldest issue last, got %d", sorted[3].Number)
	}
}
