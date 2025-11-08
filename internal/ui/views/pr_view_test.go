package views

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// mockFetchPRsUseCase is a mock implementation of FetchPRsUseCase for testing
type mockFetchPRsUseCase struct {
	executeFunc func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error)
}

func (m *mockFetchPRsUseCase) Execute(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, owner, repo, opts)
	}
	return nil, nil
}

func TestPRView_Init(t *testing.T) {
	tests := []struct {
		name          string
		mockUseCase   *mockFetchPRsUseCase
		checkCmd      func(t *testing.T, cmd tea.Cmd)
		expectLoading bool
	}{
		{
			name: "init triggers PR fetch",
			mockUseCase: &mockFetchPRsUseCase{
				executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
					return []*models.PullRequest{
						{
							Number:    1,
							Title:     "Test PR",
							State:     models.PRStateOpen,
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
			view := NewPRViewWithUseCase(tt.mockUseCase, "testowner", "testrepo")
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

func TestPRView_Update_PRsLoaded(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		msg           tea.Msg
		initialState  *PRView
		checkState    func(t *testing.T, view *PRView)
		expectLoading bool
	}{
		{
			name: "successful data load",
			msg: prsLoadedMsg{
				prs: []*models.PullRequest{
					{
						Number:    1,
						Title:     "Test PR 1",
						State:     models.PRStateOpen,
						Author:    models.User{Login: "alice"},
						Labels:    []models.Label{{Name: "feature"}},
						Comments:  5,
						Mergeable: true,
						Draft:     false,
						Reviews: []models.Review{
							{State: models.ReviewStateApproved, User: models.User{Login: "bob"}},
							{State: models.ReviewStateApproved, User: models.User{Login: "charlie"}},
						},
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						Number:    2,
						Title:     "Test PR 2",
						State:     models.PRStateClosed,
						Author:    models.User{Login: "bob"},
						Labels:    []models.Label{{Name: "bugfix"}},
						Comments:  3,
						Mergeable: false,
						Draft:     true,
						Reviews: []models.Review{
							{State: models.ReviewStateChangesRequested, User: models.User{Login: "alice"}},
						},
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				err: nil,
			},
			initialState: &PRView{
				loading:  true,
				prs:      []*models.PullRequest{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *PRView) {
				if view.loading {
					t.Error("expected loading to be false after successful load")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.prs) != 2 {
					t.Errorf("expected 2 PRs, got %d", len(view.prs))
				}
				if view.prs[0].Title != "Test PR 1" {
					t.Errorf("expected first PR title 'Test PR 1', got '%s'", view.prs[0].Title)
				}
			},
			expectLoading: false,
		},
		{
			name: "error during load",
			msg: prsLoadedMsg{
				prs: nil,
				err: errors.New("API error"),
			},
			initialState: &PRView{
				loading:  true,
				prs:      []*models.PullRequest{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *PRView) {
				if view.loading {
					t.Error("expected loading to be false after error")
				}
				if view.err == nil {
					t.Error("expected error to be set")
				}
				if view.err.Error() != "API error" {
					t.Errorf("expected error 'API error', got '%v'", view.err)
				}
				if len(view.prs) != 0 {
					t.Errorf("expected 0 PRs on error, got %d", len(view.prs))
				}
			},
			expectLoading: false,
		},
		{
			name: "empty result",
			msg: prsLoadedMsg{
				prs: []*models.PullRequest{},
				err: nil,
			},
			initialState: &PRView{
				loading:  true,
				prs:      []*models.PullRequest{},
				owner:    "testowner",
				repo:     "testrepo",
				cursor:   0,
				selected: make(map[int]struct{}),
			},
			checkState: func(t *testing.T, view *PRView) {
				if view.loading {
					t.Error("expected loading to be false")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.prs) != 0 {
					t.Errorf("expected 0 PRs, got %d", len(view.prs))
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

func TestPRView_Update_RefreshKey(t *testing.T) {
	callCount := 0
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			callCount++
			return []*models.PullRequest{
				{
					Number: callCount,
					Title:  "Refreshed PR",
					State:  models.PRStateOpen,
				},
			}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
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

func TestPRView_Update_FilterKey(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false
	view.filterState = models.PRStateOpen

	// Simulate 'f' key press for filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd := view.Update(msg)

	// Filter should toggle from open to closed
	if view.filterState != models.PRStateClosed {
		t.Errorf("expected filter state to be closed, got %s", view.filterState)
	}

	if cmd == nil {
		t.Error("expected non-nil command to trigger refresh with new filter")
	}
}

func TestPRView_Update_MergeKey(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{
				{
					Number:    1,
					Title:     "Mergeable PR",
					State:     models.PRStateOpen,
					Mergeable: true,
				},
			}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false
	view.prs = []*models.PullRequest{
		{
			Number:    1,
			Title:     "Mergeable PR",
			State:     models.PRStateOpen,
			Mergeable: true,
		},
	}
	view.cursor = 0

	// Simulate 'm' key press for merge
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	_, _ = view.Update(msg)

	// Note: Actual merge functionality would need to be tested with a proper merge use case
	// For now, we just check that it doesn't crash
}

func TestPRView_View_States(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		view        *PRView
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "loading state",
			view: &PRView{
				loading:     true,
				prs:         []*models.PullRequest{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.PRStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Loading") {
					t.Error("expected loading message in output")
				}
			},
		},
		{
			name: "error state",
			view: &PRView{
				loading:     false,
				err:         errors.New("API error"),
				prs:         []*models.PullRequest{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.PRStateOpen,
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
			view: &PRView{
				loading: false,
				prs: []*models.PullRequest{
					{
						Number:    1,
						Title:     "Test PR",
						State:     models.PRStateOpen,
						Author:    models.User{Login: "alice"},
						Labels:    []models.Label{{Name: "feature"}},
						Comments:  5,
						Mergeable: true,
						Draft:     false,
						Reviews: []models.Review{
							{State: models.ReviewStateApproved, User: models.User{Login: "bob"}},
						},
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
				filterState: models.PRStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test PR") {
					t.Error("expected PR title in output")
				}
				if !strings.Contains(output, "#1") {
					t.Error("expected PR number in output")
				}
			},
		},
		{
			name: "empty data state",
			view: &PRView{
				loading:     false,
				prs:         []*models.PullRequest{},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.PRStateOpen,
			},
			checkOutput: func(t *testing.T, output string) {
				// Should show header but no PRs
				if !strings.Contains(output, "Pull Requests") {
					t.Error("expected 'Pull Requests' header in output")
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

func TestPRView_Navigation(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{
				{Number: 1, Title: "PR 1", State: models.PRStateOpen},
				{Number: 2, Title: "PR 2", State: models.PRStateOpen},
				{Number: 3, Title: "PR 3", State: models.PRStateOpen},
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
			view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
			// Load initial data
			view.loading = false
			view.prs = []*models.PullRequest{
				{Number: 1, Title: "PR 1", State: models.PRStateOpen},
				{Number: 2, Title: "PR 2", State: models.PRStateOpen},
				{Number: 3, Title: "PR 3", State: models.PRStateOpen},
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

func TestPRView_ReviewStatus(t *testing.T) {
	tests := []struct {
		name             string
		pr               *models.PullRequest
		expectedApproved int
		expectedChanges  int
		expectedPending  int
	}{
		{
			name: "multiple approvals",
			pr: &models.PullRequest{
				Reviews: []models.Review{
					{State: models.ReviewStateApproved},
					{State: models.ReviewStateApproved},
					{State: models.ReviewStateApproved},
				},
			},
			expectedApproved: 3,
			expectedChanges:  0,
			expectedPending:  0,
		},
		{
			name: "changes requested",
			pr: &models.PullRequest{
				Reviews: []models.Review{
					{State: models.ReviewStateApproved},
					{State: models.ReviewStateChangesRequested},
				},
			},
			expectedApproved: 1,
			expectedChanges:  1,
			expectedPending:  0,
		},
		{
			name: "pending reviews",
			pr: &models.PullRequest{
				Reviews: []models.Review{
					{State: models.ReviewStatePending},
					{State: models.ReviewStatePending},
				},
			},
			expectedApproved: 0,
			expectedChanges:  0,
			expectedPending:  2,
		},
		{
			name: "mixed reviews",
			pr: &models.PullRequest{
				Reviews: []models.Review{
					{State: models.ReviewStateApproved},
					{State: models.ReviewStateChangesRequested},
					{State: models.ReviewStatePending},
				},
			},
			expectedApproved: 1,
			expectedChanges:  1,
			expectedPending:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := NewPRView()
			approved, changes, pending := view.countReviews(tt.pr)

			if approved != tt.expectedApproved {
				t.Errorf("expected %d approved, got %d", tt.expectedApproved, approved)
			}
			if changes != tt.expectedChanges {
				t.Errorf("expected %d changes requested, got %d", tt.expectedChanges, changes)
			}
			if pending != tt.expectedPending {
				t.Errorf("expected %d pending, got %d", tt.expectedPending, pending)
			}
		})
	}
}

func TestPRView_WindowSizeMsg(t *testing.T) {
	view := NewPRView()
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}

	view.Update(msg)

	if view.width != 100 {
		t.Errorf("expected width 100, got %d", view.width)
	}
	if view.height != 30 {
		t.Errorf("expected height 30, got %d", view.height)
	}
}

func TestPRView_HelpToggle(t *testing.T) {
	view := NewPRView()
	view.width = 80
	view.height = 24
	view.loading = false
	view.statusBar = components.NewStatusBar()

	// Initially help is off
	if view.showHelp {
		t.Error("expected showHelp to be false initially")
	}

	// Toggle help on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	view.Update(msg)

	if !view.showHelp {
		t.Error("expected showHelp to be true after toggle")
	}

	// Check that help is rendered
	output := view.View()
	if !strings.Contains(output, "Navigation:") {
		t.Error("expected help content in output")
	}
	if !strings.Contains(output, "Actions:") {
		t.Error("expected actions section in help")
	}

	// Toggle help off
	view.Update(msg)

	if view.showHelp {
		t.Error("expected showHelp to be false after second toggle")
	}
}

func TestPRView_DetailAndDiffKeys(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{
				{Number: 1, Title: "Test PR", State: models.PRStateOpen},
			}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false
	view.prs = []*models.PullRequest{
		{Number: 1, Title: "Test PR", State: models.PRStateOpen},
	}

	// Test Enter key (detail view)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\r'}}
	msg.Type = tea.KeyEnter
	_, _ = view.Update(msg)

	// Test 'd' key (diff view)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, _ = view.Update(msg)

	// These should not crash - actual implementation would navigate to detail/diff views
}

func TestPRView_RenderReviewStatus(t *testing.T) {
	view := NewPRView()

	tests := []struct {
		name             string
		approved         int
		changesRequested int
		pending          int
		expectEmpty      bool
	}{
		{
			name:             "no reviews",
			approved:         0,
			changesRequested: 0,
			pending:          0,
			expectEmpty:      true,
		},
		{
			name:             "only approved",
			approved:         2,
			changesRequested: 0,
			pending:          0,
			expectEmpty:      false,
		},
		{
			name:             "all types",
			approved:         1,
			changesRequested: 1,
			pending:          1,
			expectEmpty:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := view.renderReviewStatus(tt.approved, tt.changesRequested, tt.pending)
			isEmpty := result == ""

			if isEmpty != tt.expectEmpty {
				t.Errorf("expected isEmpty=%v, got %v", tt.expectEmpty, isEmpty)
			}

			if !tt.expectEmpty && !strings.Contains(result, "✓") && tt.approved > 0 {
				t.Error("expected approved symbol in review status")
			}
		})
	}
}

func TestPRView_DraftAndMergedState(t *testing.T) {
	now := time.Now()
	view := &PRView{
		loading: false,
		prs: []*models.PullRequest{
			{
				Number:    1,
				Title:     "Draft PR",
				State:     models.PRStateOpen,
				Draft:     true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				Number:    2,
				Title:     "Merged PR",
				State:     models.PRStateClosed,
				Merged:    true,
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
		filterState: models.PRStateOpen,
	}

	output := view.View()

	if !strings.Contains(output, "Draft PR") {
		t.Error("expected draft PR title in output")
	}
	if !strings.Contains(output, "Merged PR") {
		t.Error("expected merged PR title in output")
	}
}

func TestPRView_LongList(t *testing.T) {
	now := time.Now()
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			prs := make([]*models.PullRequest, 100)
			for i := 0; i < 100; i++ {
				prs[i] = &models.PullRequest{
					Number:    i + 1,
					Title:     fmt.Sprintf("PR #%d", i+1),
					State:     models.PRStateOpen,
					Author:    models.User{Login: "user"},
					CreatedAt: now,
					UpdatedAt: now,
				}
			}
			return prs, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false
	view.width = 80
	view.height = 24
	view.statusBar.SetSize(80, 1)

	// Initialize PRs
	prs := make([]*models.PullRequest, 100)
	for i := 0; i < 100; i++ {
		prs[i] = &models.PullRequest{
			Number:    i + 1,
			Title:     fmt.Sprintf("PR #%d", i+1),
			State:     models.PRStateOpen,
			Author:    models.User{Login: "user"},
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	view.prs = prs

	// Test navigation in long list
	view.cursor = 50
	output := view.View()

	// Check that the view renders without error and contains PR info
	if !strings.Contains(output, "Pull Requests") {
		t.Error("expected 'Pull Requests' header in output")
	}

	// Status bar should contain position info
	if len(view.prs) > 0 && view.cursor >= 0 && view.cursor < len(view.prs) {
		// Position should be updated in status bar
		view.updateStatusBar()
	}
}

func TestPRView_InitWithoutUseCase(t *testing.T) {
	view := NewPRView()
	cmd := view.Init()

	if cmd != nil {
		t.Error("expected nil command when use case is not set")
	}
}

func TestPRView_QuitKey(t *testing.T) {
	view := NewPRView()
	view.loading = false

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := view.Update(msg)

	if cmd == nil {
		t.Error("expected quit command")
	}

	// Test Ctrl+C
	msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd = view.Update(msg)

	if cmd == nil {
		t.Error("expected quit command for Ctrl+C")
	}
}

func TestPRView_RefreshWhileLoading(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = true

	// Try to refresh while already loading
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := view.Update(msg)

	// Should not trigger another fetch while loading
	if cmd != nil {
		t.Error("expected no command while loading")
	}
}

func TestPRView_FilterCycle(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")
	view.loading = false

	// Start with open
	if view.filterState != models.PRStateOpen {
		t.Errorf("expected initial filter state to be open, got %s", view.filterState)
	}

	// Cycle to closed
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	view.Update(msg)
	if view.filterState != models.PRStateClosed {
		t.Errorf("expected filter state to be closed, got %s", view.filterState)
	}

	// Wait for loading to complete
	view.Update(prsLoadedMsg{prs: []*models.PullRequest{}, err: nil})
	view.loading = false

	// Cycle to all
	view.Update(msg)
	if view.filterState != models.PRStateAll {
		t.Errorf("expected filter state to be all, got %s", view.filterState)
	}

	// Wait for loading to complete
	view.Update(prsLoadedMsg{prs: []*models.PullRequest{}, err: nil})
	view.loading = false

	// Cycle back to open
	view.Update(msg)
	if view.filterState != models.PRStateOpen {
		t.Errorf("expected filter state to be open, got %s", view.filterState)
	}
}

func TestPRView_CursorBounds(t *testing.T) {
	now := time.Now()
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{
				{Number: 1, Title: "PR 1", State: models.PRStateOpen, Author: models.User{Login: "user"}, CreatedAt: now, UpdatedAt: now},
				{Number: 2, Title: "PR 2", State: models.PRStateOpen, Author: models.User{Login: "user"}, CreatedAt: now, UpdatedAt: now},
			}, nil
		},
	}

	view := NewPRViewWithUseCase(mockUseCase, "testowner", "testrepo")

	// Load data with cursor out of bounds
	view.cursor = 10
	msg := prsLoadedMsg{
		prs: []*models.PullRequest{
			{Number: 1, Title: "PR 1", State: models.PRStateOpen, Author: models.User{Login: "user"}, CreatedAt: now, UpdatedAt: now},
		},
		err: nil,
	}
	view.Update(msg)

	// Cursor should be reset to last valid position
	if view.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", view.cursor)
	}

	// Test with empty list
	msg = prsLoadedMsg{
		prs: []*models.PullRequest{},
		err: nil,
	}
	view.cursor = 5
	view.Update(msg)

	if view.cursor != 0 {
		t.Errorf("expected cursor to be 0 for empty list, got %d", view.cursor)
	}
}

func TestPRView_MergeableStatus(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		pr         *models.PullRequest
		shouldShow bool
	}{
		{
			name: "open and mergeable",
			pr: &models.PullRequest{
				Number:    1,
				Title:     "Mergeable PR",
				State:     models.PRStateOpen,
				Mergeable: true,
				Draft:     false,
				Author:    models.User{Login: "user"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			shouldShow: true,
		},
		{
			name: "open but not mergeable",
			pr: &models.PullRequest{
				Number:    2,
				Title:     "Not Mergeable PR",
				State:     models.PRStateOpen,
				Mergeable: false,
				Draft:     false,
				Author:    models.User{Login: "user"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			shouldShow: true,
		},
		{
			name: "draft PR",
			pr: &models.PullRequest{
				Number:    3,
				Title:     "Draft PR",
				State:     models.PRStateOpen,
				Draft:     true,
				Author:    models.User{Login: "user"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			shouldShow: false,
		},
		{
			name: "closed PR",
			pr: &models.PullRequest{
				Number:    4,
				Title:     "Closed PR",
				State:     models.PRStateClosed,
				Mergeable: true,
				Draft:     false,
				Author:    models.User{Login: "user"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			shouldShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &PRView{
				loading:     false,
				prs:         []*models.PullRequest{tt.pr},
				width:       80,
				height:      24,
				owner:       "testowner",
				repo:        "testrepo",
				cursor:      0,
				selected:    make(map[int]struct{}),
				statusBar:   components.NewStatusBar(),
				filterState: models.PRStateOpen,
			}

			output := view.View()
			hasMergeableStatus := strings.Contains(output, "✓") || strings.Contains(output, "✗")

			if tt.shouldShow && !hasMergeableStatus {
				t.Error("expected mergeable status indicator in output")
			}
		})
	}
}
