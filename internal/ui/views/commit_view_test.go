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

// mockFetchCommitsUseCase is a mock implementation of FetchCommitsUseCase for testing
type mockFetchCommitsUseCase struct {
	executeFunc func(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error)
}

func (m *mockFetchCommitsUseCase) Execute(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, owner, repo, opts)
	}
	return nil, nil
}

func TestCommitView_Init(t *testing.T) {
	tests := []struct {
		name          string
		mockUseCase   *mockFetchCommitsUseCase
		checkCmd      func(t *testing.T, cmd tea.Cmd)
		expectLoading bool
	}{
		{
			name: "init triggers commit fetch",
			mockUseCase: &mockFetchCommitsUseCase{
				executeFunc: func(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
					return []*models.Commit{
						{
							SHA:     "abc123",
							Message: "Test Commit",
							Author: models.CommitAuthor{
								Name:  "Alice",
								Email: "alice@example.com",
								Date:  time.Now(),
							},
							CreatedAt: time.Now(),
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
			view := NewCommitViewWithUseCase(tt.mockUseCase, "testowner", "testrepo")
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

func TestCommitView_Update_CommitsLoaded(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		msg           tea.Msg
		initialState  *CommitView
		checkState    func(t *testing.T, view *CommitView)
		expectLoading bool
	}{
		{
			name: "successful data load",
			msg: commitsLoadedMsg{
				commits: []*models.Commit{
					{
						SHA:     "abc123",
						Message: "feat: add feature",
						Author: models.CommitAuthor{
							Name:  "Alice",
							Email: "alice@example.com",
							Date:  now,
						},
						CreatedAt: now,
					},
					{
						SHA:     "def456",
						Message: "fix: bug fix",
						Author: models.CommitAuthor{
							Name:  "Bob",
							Email: "bob@example.com",
							Date:  now,
						},
						CreatedAt: now,
					},
				},
				err: nil,
			},
			initialState: &CommitView{
				loading: true,
				commits: []*models.Commit{},
				owner:   "testowner",
				repo:    "testrepo",
				cursor:  0,
			},
			checkState: func(t *testing.T, view *CommitView) {
				if view.loading {
					t.Error("expected loading to be false after successful load")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.commits) != 2 {
					t.Errorf("expected 2 commits, got %d", len(view.commits))
				}
				if view.commits[0].SHA != "abc123" {
					t.Errorf("expected first commit SHA 'abc123', got '%s'", view.commits[0].SHA)
				}
			},
			expectLoading: false,
		},
		{
			name: "error during load",
			msg: commitsLoadedMsg{
				commits: nil,
				err:     errors.New("API error"),
			},
			initialState: &CommitView{
				loading: true,
				commits: []*models.Commit{},
				owner:   "testowner",
				repo:    "testrepo",
				cursor:  0,
			},
			checkState: func(t *testing.T, view *CommitView) {
				if view.loading {
					t.Error("expected loading to be false after error")
				}
				if view.err == nil {
					t.Error("expected error to be set")
				}
				if view.err.Error() != "API error" {
					t.Errorf("expected error 'API error', got '%v'", view.err)
				}
				if len(view.commits) != 0 {
					t.Errorf("expected 0 commits on error, got %d", len(view.commits))
				}
			},
			expectLoading: false,
		},
		{
			name: "empty result",
			msg: commitsLoadedMsg{
				commits: []*models.Commit{},
				err:     nil,
			},
			initialState: &CommitView{
				loading: true,
				commits: []*models.Commit{},
				owner:   "testowner",
				repo:    "testrepo",
				cursor:  0,
			},
			checkState: func(t *testing.T, view *CommitView) {
				if view.loading {
					t.Error("expected loading to be false")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.commits) != 0 {
					t.Errorf("expected 0 commits, got %d", len(view.commits))
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

func TestCommitView_Update_RefreshKey(t *testing.T) {
	callCount := 0
	mockUseCase := &mockFetchCommitsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
			callCount++
			return []*models.Commit{
				{
					SHA:     "abc123",
					Message: "Refreshed Commit",
					Author: models.CommitAuthor{
						Name: "Alice",
						Date: time.Now(),
					},
					CreatedAt: time.Now(),
				},
			}, nil
		},
	}

	view := NewCommitViewWithUseCase(mockUseCase, "testowner", "testrepo")
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

func TestCommitView_View_States(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		view        *CommitView
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "loading state",
			view: &CommitView{
				loading:   true,
				commits:   []*models.Commit{},
				width:     80,
				height:    24,
				owner:     "testowner",
				repo:      "testrepo",
				cursor:    0,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Loading") {
					t.Error("expected loading message in output")
				}
			},
		},
		{
			name: "error state",
			view: &CommitView{
				loading:   false,
				err:       errors.New("API error"),
				commits:   []*models.Commit{},
				width:     80,
				height:    24,
				owner:     "testowner",
				repo:      "testrepo",
				cursor:    0,
				statusBar: components.NewStatusBar(),
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
			view: &CommitView{
				loading: false,
				commits: []*models.Commit{
					{
						SHA:     "abc123",
						Message: "feat: test commit",
						Author: models.CommitAuthor{
							Name:  "Alice",
							Email: "alice@example.com",
							Date:  now,
						},
						CreatedAt: now,
					},
				},
				width:     80,
				height:    24,
				owner:     "testowner",
				repo:      "testrepo",
				cursor:    0,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "abc123") {
					t.Error("expected commit SHA in output")
				}
				if !strings.Contains(output, "test commit") {
					t.Error("expected commit message in output")
				}
			},
		},
		{
			name: "empty data state",
			view: &CommitView{
				loading:   false,
				commits:   []*models.Commit{},
				width:     80,
				height:    24,
				owner:     "testowner",
				repo:      "testrepo",
				cursor:    0,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Commits") {
					t.Error("expected 'Commits' header in output")
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

func TestCommitView_Navigation(t *testing.T) {
	mockUseCase := &mockFetchCommitsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.CommitOptions) ([]*models.Commit, error) {
			return []*models.Commit{
				{SHA: "abc1", Message: "Commit 1", CreatedAt: time.Now()},
				{SHA: "abc2", Message: "Commit 2", CreatedAt: time.Now()},
				{SHA: "abc3", Message: "Commit 3", CreatedAt: time.Now()},
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
			view := NewCommitViewWithUseCase(mockUseCase, "testowner", "testrepo")
			// Load initial data
			view.loading = false
			view.commits = []*models.Commit{
				{SHA: "abc1", Message: "Commit 1", CreatedAt: time.Now()},
				{SHA: "abc2", Message: "Commit 2", CreatedAt: time.Now()},
				{SHA: "abc3", Message: "Commit 3", CreatedAt: time.Now()},
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
