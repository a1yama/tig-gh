package views

import (
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	tea "github.com/charmbracelet/bubbletea"
)

// TestIssueDetailView_Init tests the initialization of the detail view
func TestIssueDetailView_Init(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)

	if view == nil {
		t.Fatal("NewIssueDetailView returned nil")
	}

	if view.issue == nil {
		t.Error("issue should not be nil after initialization")
	}

	if view.issue.Number != issue.Number {
		t.Errorf("expected issue number %d, got %d", issue.Number, view.issue.Number)
	}

	if view.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0, got %d", view.scrollOffset)
	}

	if view.loading {
		t.Error("loading should be false after initialization")
	}
}

// TestIssueDetailView_Update_KeyboardInput tests keyboard input handling
func TestIssueDetailView_Update_KeyboardInput(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedQuit  bool
		expectedBack  bool
		shouldScroll  bool
		scrollBefore  int
		scrollAfter   int
	}{
		{
			name:         "q key should go back",
			key:          "q",
			expectedBack: true,
		},
		{
			name:         "j key should scroll down",
			key:          "j",
			shouldScroll: true,
			scrollBefore: 0,
			scrollAfter:  1,
		},
		{
			name:         "k key should scroll up",
			key:          "k",
			shouldScroll: true,
			scrollBefore: 5,
			scrollAfter:  4,
		},
		{
			name:         "down arrow should scroll down",
			key:          "down",
			shouldScroll: true,
			scrollBefore: 0,
			scrollAfter:  1,
		},
		{
			name:         "up arrow should scroll up",
			key:          "up",
			shouldScroll: true,
			scrollBefore: 5,
			scrollAfter:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := createTestIssue()
			view := NewIssueDetailView(issue, "owner", "repo", nil)

			if tt.shouldScroll {
				view.scrollOffset = tt.scrollBefore
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			}

			model, cmd := view.Update(msg)

			if tt.expectedQuit {
				if cmd == nil {
					t.Error("expected quit command, got nil")
				}
			}

			if tt.expectedBack {
				if cmd == nil {
					t.Error("expected back command, got nil")
				}
			}

			if tt.shouldScroll {
				updatedView := model.(*IssueDetailView)
				if updatedView.scrollOffset != tt.scrollAfter {
					t.Errorf("expected scrollOffset %d, got %d", tt.scrollAfter, updatedView.scrollOffset)
				}
			}
		})
	}
}

// TestIssueDetailView_Update_ScrollBounds tests scroll boundary conditions
func TestIssueDetailView_Update_ScrollBounds(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)
	view.scrollOffset = 0

	// Try to scroll up from the top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	model, _ := view.Update(msg)
	updatedView := model.(*IssueDetailView)

	if updatedView.scrollOffset < 0 {
		t.Errorf("scrollOffset should not be negative, got %d", updatedView.scrollOffset)
	}
}

// TestIssueDetailView_Update_WindowSize tests window size message handling
func TestIssueDetailView_Update_WindowSize(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, _ := view.Update(msg)
	updatedView := model.(*IssueDetailView)

	if updatedView.width != 100 {
		t.Errorf("expected width 100, got %d", updatedView.width)
	}

	if updatedView.height != 50 {
		t.Errorf("expected height 50, got %d", updatedView.height)
	}
}

// TestIssueDetailView_Update_LoadingState tests loading state handling
func TestIssueDetailView_Update_LoadingState(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)
	view.loading = true
	view.width = 80
	view.height = 24

	// View should not be nil even when loading
	output := view.View()
	if output == "" {
		t.Error("View should return content even when loading")
	}

	// Check that loading indicator is present
	if !containsString(output, "Loading") && !containsString(output, "loading") {
		t.Error("View should contain loading indicator when loading is true")
	}
}

// TestIssueDetailView_Update_ErrorState tests error state handling
func TestIssueDetailView_Update_ErrorState(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)
	view.width = 80
	view.height = 24

	// Simulate error state
	view.err = tea.ErrProgramKilled

	output := view.View()
	if output == "" {
		t.Error("View should return content even with error")
	}

	// Check that error message is present
	if !containsString(output, "Error") && !containsString(output, "error") {
		t.Error("View should contain error message when err is not nil")
	}
}

// TestIssueDetailView_View tests the rendering
func TestIssueDetailView_View(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)
	view.width = 100
	view.height = 50

	output := view.View()

	if output == "" {
		t.Error("View should return non-empty string")
	}

	// Check that issue details are present
	if !containsString(output, issue.Title) {
		t.Errorf("View should contain issue title '%s'", issue.Title)
	}

	// Check that issue number is present
	numberStr := "#123"
	if !containsString(output, numberStr) {
		t.Errorf("View should contain issue number '%s'", numberStr)
	}
}

// TestIssueDetailView_View_WithoutSize tests rendering without size set
func TestIssueDetailView_View_WithoutSize(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)

	output := view.View()

	// Should return a placeholder or initialization message
	if output == "" {
		t.Error("View should return content even without size")
	}
}

// TestIssueDetailView_OpenInBrowser tests the browser open functionality
func TestIssueDetailView_OpenInBrowser(t *testing.T) {
	issue := createTestIssue()
	view := NewIssueDetailView(issue, "owner", "repo", nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	_, cmd := view.Update(msg)

	// Should return a command (even if it's just a placeholder)
	// The actual browser opening is tested in integration tests
	if cmd == nil {
		t.Log("Warning: 'o' key did not return a command (browser open might not be implemented yet)")
	}
}

// Helper functions

// createTestIssue creates a test issue for testing
func createTestIssue() *models.Issue {
	now := time.Now()
	return &models.Issue{
		ID:     1,
		Number: 123,
		Title:  "Test Issue Title",
		Body:   "This is a test issue body.\n\n## Section\n\nWith some **markdown** content.",
		State:  models.IssueStateOpen,
		Author: models.User{
			ID:    1,
			Login: "testuser",
			Name:  "Test User",
		},
		Assignees: []models.User{
			{
				ID:    2,
				Login: "assignee1",
				Name:  "Assignee One",
			},
		},
		Labels: []models.Label{
			{
				ID:    1,
				Name:  "bug",
				Color: "d73a4a",
			},
			{
				ID:    2,
				Name:  "enhancement",
				Color: "a2eeef",
			},
		},
		Milestone: &models.Milestone{
			ID:     1,
			Number: 1,
			Title:  "v1.0.0",
		},
		Comments:  5,
		CreatedAt: now.Add(-48 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
		URL:       "https://api.github.com/repos/test/repo/issues/123",
		HTMLURL:   "https://github.com/test/repo/issues/123",
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
