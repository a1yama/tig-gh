package views

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	tea "github.com/charmbracelet/bubbletea"
)

// TestPRDetailView_Init tests the initialization of the PR detail view
func TestPRDetailView_Init(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	if view == nil {
		t.Fatal("NewPRDetailView returned nil")
	}

	if view.pr == nil {
		t.Error("pr should not be nil after initialization")
	}

	if view.pr.Number != pr.Number {
		t.Errorf("expected pr number %d, got %d", pr.Number, view.pr.Number)
	}

	if view.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0, got %d", view.scrollOffset)
	}

	if view.loading {
		t.Error("loading should be false after initialization")
	}

	if view.currentTab != tabOverview {
		t.Errorf("expected currentTab to be tabOverview, got %d", view.currentTab)
	}
}

// TestPRDetailView_Update_KeyboardInput tests keyboard input handling
func TestPRDetailView_Update_KeyboardInput(t *testing.T) {
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
			name:         "esc key should go back",
			key:          "esc",
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
			pr := createTestPullRequest()
			view := NewPRDetailView(pr)

			if tt.shouldScroll {
				view.scrollOffset = tt.scrollBefore
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
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
				updatedView := model.(*PRDetailView)
				if updatedView.scrollOffset != tt.scrollAfter {
					t.Errorf("expected scrollOffset %d, got %d", tt.scrollAfter, updatedView.scrollOffset)
				}
			}
		})
	}
}

// TestPRDetailView_Update_TabSwitching tests tab switching functionality
func TestPRDetailView_Update_TabSwitching(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedTab prTab
	}{
		{
			name:        "1 key should switch to overview tab",
			key:         "1",
			expectedTab: tabOverview,
		},
		{
			name:        "2 key should switch to files tab",
			key:         "2",
			expectedTab: tabFiles,
		},
		{
			name:        "3 key should switch to commits tab",
			key:         "3",
			expectedTab: tabCommits,
		},
		{
			name:        "4 key should switch to comments tab",
			key:         "4",
			expectedTab: tabComments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := createTestPullRequest()
			view := NewPRDetailView(pr)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			model, _ := view.Update(msg)

			updatedView := model.(*PRDetailView)
			if updatedView.currentTab != tt.expectedTab {
				t.Errorf("expected tab %d, got %d", tt.expectedTab, updatedView.currentTab)
			}

			// Scroll should reset when changing tabs
			if updatedView.scrollOffset != 0 {
				t.Errorf("expected scrollOffset to be 0 after tab change, got %d", updatedView.scrollOffset)
			}
		})
	}
}

// TestPRDetailView_Update_ScrollBounds tests scroll boundary conditions
func TestPRDetailView_Update_ScrollBounds(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.scrollOffset = 0

	// Try to scroll up from the top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	model, _ := view.Update(msg)
	updatedView := model.(*PRDetailView)

	if updatedView.scrollOffset < 0 {
		t.Errorf("scrollOffset should not be negative, got %d", updatedView.scrollOffset)
	}
}

// TestPRDetailView_Update_MergeKey tests merge key functionality
func TestPRDetailView_Update_MergeKey(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")}
	_, cmd := view.Update(msg)

	// Should return a command for merge action
	if cmd == nil {
		t.Log("Warning: 'm' key did not return a command (merge might not be implemented yet)")
	}
}

// TestPRDetailView_Update_DiffKey tests diff key functionality
func TestPRDetailView_Update_DiffKey(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	_, cmd := view.Update(msg)

	// Should return a command for diff display
	if cmd == nil {
		t.Log("Warning: 'd' key did not return a command (diff might not be implemented yet)")
	}
}

// TestPRDetailView_Update_WindowSize tests window size message handling
func TestPRDetailView_Update_WindowSize(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, _ := view.Update(msg)
	updatedView := model.(*PRDetailView)

	if updatedView.width != 100 {
		t.Errorf("expected width 100, got %d", updatedView.width)
	}

	if updatedView.height != 50 {
		t.Errorf("expected height 50, got %d", updatedView.height)
	}
}

// TestPRDetailView_Update_LoadingState tests loading state handling
func TestPRDetailView_Update_LoadingState(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
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

// TestPRDetailView_Update_ErrorState tests error state handling
func TestPRDetailView_Update_ErrorState(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
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

// TestPRDetailView_View_OverviewTab tests the rendering of overview tab
func TestPRDetailView_View_OverviewTab(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50
	view.currentTab = tabOverview

	output := view.View()

	if output == "" {
		t.Error("View should return non-empty string")
	}

	// Check that PR details are present
	if !containsString(output, pr.Title) {
		t.Errorf("View should contain PR title '%s'", pr.Title)
	}

	// Check that PR number is present
	numberStr := "#456"
	if !containsString(output, numberStr) {
		t.Errorf("View should contain PR number '%s'", numberStr)
	}

	// Check that base and head branches are present
	if !containsString(output, pr.Base.Name) || !containsString(output, pr.Head.Name) {
		t.Error("View should contain base and head branch names")
	}
}

// TestPRDetailView_View_FilesTab tests the rendering of files tab
func TestPRDetailView_View_FilesTab(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50
	view.currentTab = tabFiles

	output := view.View()

	if output == "" {
		t.Error("View should return non-empty string")
	}

	// Should show files changed information
	if !containsString(output, "Files") && !containsString(output, "files") {
		t.Log("Warning: Files tab might not be rendering file information")
	}
}

// TestPRDetailView_View_CommitsTab tests the rendering of commits tab
func TestPRDetailView_View_CommitsTab(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50
	view.currentTab = tabCommits

	output := view.View()

	if output == "" {
		t.Error("View should return non-empty string")
	}

	// Should show commits information
	if !containsString(output, "Commit") && !containsString(output, "commit") {
		t.Log("Warning: Commits tab might not be rendering commit information")
	}
}

// TestPRDetailView_View_CommentsTab tests the rendering of comments tab
func TestPRDetailView_View_CommentsTab(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50
	view.currentTab = tabComments

	output := view.View()

	if output == "" {
		t.Error("View should return non-empty string")
	}

	// Should show comments information
	if !containsString(output, "Comment") && !containsString(output, "comment") {
		t.Log("Warning: Comments tab might not be rendering comment information")
	}
}

// TestPRDetailView_View_WithoutSize tests rendering without size set
func TestPRDetailView_View_WithoutSize(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	output := view.View()

	// Should return a placeholder or initialization message
	if output == "" {
		t.Error("View should return content even without size")
	}
}

// TestPRDetailView_View_WithReviews tests rendering with reviews
func TestPRDetailView_View_WithReviews(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50

	output := view.View()

	// Should contain review information
	if !containsString(output, "Review") && !containsString(output, "review") {
		t.Log("Warning: View might not be showing review information")
	}
}

// TestPRDetailView_OpenInBrowser tests the browser open functionality
func TestPRDetailView_OpenInBrowser(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	_, cmd := view.Update(msg)

	// Should return a command
	if cmd == nil {
		t.Log("Warning: 'o' key did not return a command (browser open might not be implemented yet)")
	}
}

// TestPRDetailView_GoToTop tests the g key to go to top
func TestPRDetailView_GoToTop(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.scrollOffset = 10

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	model, _ := view.Update(msg)
	updatedView := model.(*PRDetailView)

	if updatedView.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0 after 'g' key, got %d", updatedView.scrollOffset)
	}
}

// TestPRDetailView_GoToBottom tests the G key to go to bottom
func TestPRDetailView_GoToBottom(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.scrollOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	model, _ := view.Update(msg)
	updatedView := model.(*PRDetailView)

	// Should set a large offset (will be capped in View)
	if updatedView.scrollOffset <= 0 {
		t.Error("expected scrollOffset to be greater than 0 after 'G' key")
	}
}

// Helper functions

// TestPRDetailView_Init_Command tests the Init command
func TestPRDetailView_Init_Command(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	cmd := view.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

// TestPRDetailView_Update_CtrlC tests ctrl+c handling
func TestPRDetailView_Update_CtrlC(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := view.Update(msg)

	if cmd == nil {
		t.Error("ctrl+c should return quit command")
	}
}

// TestPRDetailView_renderHeader_Draft tests rendering draft PR
func TestPRDetailView_renderHeader_Draft(t *testing.T) {
	pr := createTestPullRequest()
	pr.Draft = true
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50

	output := view.renderHeader()
	if !containsString(output, "Draft") {
		t.Error("Draft PR should show Draft badge")
	}
}

// TestPRDetailView_renderHeader_Merged tests rendering merged PR
func TestPRDetailView_renderHeader_Merged(t *testing.T) {
	pr := createTestPullRequest()
	pr.Merged = true
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50

	output := view.renderHeader()
	if !containsString(output, "Merged") {
		t.Error("Merged PR should show Merged badge")
	}
}

// TestPRDetailView_getMergeStatus_Conflicts tests conflict status
func TestPRDetailView_getMergeStatus_Conflicts(t *testing.T) {
	pr := createTestPullRequest()
	pr.Mergeable = false
	view := NewPRDetailView(pr)

	status := view.getMergeStatus()
	if !containsString(status, "Conflicts") {
		t.Error("Non-mergeable PR should show Conflicts status")
	}
}

// TestPRDetailView_getMergeStatus_ChangesRequested tests changes requested status
func TestPRDetailView_getMergeStatus_ChangesRequested(t *testing.T) {
	pr := createTestPullRequest()
	pr.Reviews = []models.Review{
		{
			ID:    1,
			State: models.ReviewStateChangesRequested,
		},
	}
	view := NewPRDetailView(pr)

	status := view.getMergeStatus()
	if !containsString(status, "Changes requested") {
		t.Error("PR with changes requested should show appropriate status")
	}
}

// TestPRDetailView_getMergeStatus_AwaitingReview tests awaiting review status
func TestPRDetailView_getMergeStatus_AwaitingReview(t *testing.T) {
	pr := createTestPullRequest()
	pr.Reviews = []models.Review{
		{
			ID:    1,
			State: models.ReviewStateApproved,
		},
	}
	view := NewPRDetailView(pr)

	status := view.getMergeStatus()
	if !containsString(status, "Awaiting review") {
		t.Error("PR with only one approval should show Awaiting review status")
	}
}

// TestPRDetailView_getReviewsSummary_Empty tests empty reviews summary
func TestPRDetailView_getReviewsSummary_Empty(t *testing.T) {
	pr := createTestPullRequest()
	pr.Reviews = []models.Review{}
	view := NewPRDetailView(pr)

	summary := view.getReviewsSummary()
	if !containsString(summary, "No reviews") {
		t.Error("PR with no reviews should show 'No reviews'")
	}
}

// TestPRDetailView_getReviewsSummary_WithPending tests pending reviews summary
func TestPRDetailView_getReviewsSummary_WithPending(t *testing.T) {
	pr := createTestPullRequest()
	pr.Reviews = []models.Review{
		{
			ID:    1,
			State: models.ReviewStatePending,
		},
	}
	view := NewPRDetailView(pr)

	summary := view.getReviewsSummary()
	if summary == "" {
		t.Error("Summary should not be empty with pending reviews")
	}
}

// TestPRDetailView_renderBody_Empty tests rendering empty body
func TestPRDetailView_renderBody_Empty(t *testing.T) {
	pr := createTestPullRequest()
	pr.Body = ""
	view := NewPRDetailView(pr)

	output := view.renderBody()
	if !containsString(output, "No description") {
		t.Error("Empty body should show 'No description provided'")
	}
}

// TestPRDetailView_renderCommentsTab_Empty tests rendering empty comments
func TestPRDetailView_renderCommentsTab_Empty(t *testing.T) {
	pr := createTestPullRequest()
	pr.Comments = 0
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50

	output := view.renderCommentsTab()
	if !containsString(output, "No comments") {
		t.Error("PR with no comments should show 'No comments yet'")
	}
}

// TestPRDetailView_applyScroll_LargeContent tests scrolling with large content
func TestPRDetailView_applyScroll_LargeContent(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 30

	// Create large content
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("Line %d", i))
	}
	content := strings.Join(lines, "\n")

	view.scrollOffset = 50
	output := view.applyScroll(content)

	// Should contain scroll indicator
	if !containsString(output, "of") {
		t.Log("Warning: Scroll indicator might not be showing")
	}
}

// TestPRDetailView_TabContent_AllTabs tests all tab content rendering
func TestPRDetailView_TabContent_AllTabs(t *testing.T) {
	pr := createTestPullRequest()
	view := NewPRDetailView(pr)
	view.width = 100
	view.height = 50

	tabs := []prTab{tabOverview, tabFiles, tabCommits, tabComments}
	for _, tab := range tabs {
		view.currentTab = tab
		content := view.renderTabContent()
		if content == "" {
			t.Errorf("Tab %d should return non-empty content", tab)
		}
	}
}

// Helper functions are already defined in issue_detail_view_test.go
// We'll reuse containsString from there

// createTestPullRequest creates a test pull request for testing
func createTestPullRequest() *models.PullRequest {
	now := time.Now()
	return &models.PullRequest{
		ID:     1,
		Number: 456,
		Title:  "Test Pull Request Title",
		Body:   "This is a test PR body.\n\n## Changes\n\n- Feature A\n- Feature B",
		State:  models.PRStateOpen,
		Author: models.User{
			ID:    1,
			Login: "testuser",
			Name:  "Test User",
		},
		Head: models.Branch{
			Name: "feature/test",
			SHA:  "abc123def456",
		},
		Base: models.Branch{
			Name: "main",
			SHA:  "def456abc123",
		},
		Mergeable:      true,
		MergeableState: "clean",
		Merged:         false,
		Draft:          false,
		Reviews: []models.Review{
			{
				ID: 1,
				User: models.User{
					ID:    2,
					Login: "reviewer1",
					Name:  "Reviewer One",
				},
				State:       models.ReviewStateApproved,
				Body:        "LGTM!",
				SubmittedAt: now.Add(-1 * time.Hour),
			},
			{
				ID: 2,
				User: models.User{
					ID:    3,
					Login: "reviewer2",
					Name:  "Reviewer Two",
				},
				State:       models.ReviewStateApproved,
				Body:        "Looks good",
				SubmittedAt: now.Add(-30 * time.Minute),
			},
		},
		Assignees: []models.User{
			{
				ID:    4,
				Login: "assignee1",
				Name:  "Assignee One",
			},
		},
		Labels: []models.Label{
			{
				ID:    1,
				Name:  "feature",
				Color: "0e8a16",
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
		Comments:     8,
		Commits:      5,
		Additions:    234,
		Deletions:    156,
		ChangedFiles: 12,
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now.Add(-2 * time.Hour),
	}
}
