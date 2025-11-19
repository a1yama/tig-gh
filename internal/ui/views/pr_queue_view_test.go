package views

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

func TestPRQueueView_Init(t *testing.T) {
	mockUseCase := &mockFetchPRsUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
			return []*models.PullRequest{}, nil
		},
	}

	view := NewPRQueueViewWithUseCase(mockUseCase, "owner", "repo")
	cmd := view.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a command")
	}
	if !view.loading {
		t.Fatal("expected loading state after Init")
	}
}

func TestPRQueueView_UpdateLoaded_StartsReviewLoading(t *testing.T) {
	view := NewPRQueueView()
	view.owner = "owner"
	view.repo = "repo"
	view.prRepo = &testPRRepo{}

	now := time.Now().Add(-10 * time.Hour)
	msg := prQueueLoadedMsg{
		prs: []*models.PullRequest{
			{Number: 2, Title: "Second", CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 1, Title: "First", CreatedAt: now.Add(-2 * time.Hour)},
		},
	}

	_, cmd := view.Update(msg)

	if view.loading {
		t.Fatal("expected loading to be false after update")
	}
	if view.err != nil {
		t.Fatalf("did not expect error: %v", view.err)
	}
	if len(view.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(view.entries))
	}
	if view.entries[0].pr.Number != 1 {
		t.Fatalf("expected oldest PR first, got #%d", view.entries[0].pr.Number)
	}
	if cmd == nil {
		t.Fatal("expected review loading command when repository is available")
	}
	if !view.reviewLoading {
		t.Fatal("expected reviewLoading to be true")
	}
}

func TestPRQueueView_UpdateReviewsLoaded_SetsMetrics(t *testing.T) {
	base := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	view := NewPRQueueView()
	view.entries = []*prQueueEntry{
		{
			pr: &models.PullRequest{
				Number:    1,
				Title:     "Example",
				CreatedAt: base,
			},
		},
	}

	msg := prQueueReviewsLoadedMsg{
		index: 0,
		reviews: []models.Review{
			{State: models.ReviewStateCommented, SubmittedAt: base.Add(2 * time.Hour)},
			{State: models.ReviewStateApproved, SubmittedAt: base.Add(5 * time.Hour)},
		},
	}

	if _, cmd := view.Update(msg); cmd != nil {
		t.Fatal("did not expect additional command when no more entries")
	}

	entry := view.entries[0]
	if !entry.reviewsLoaded {
		t.Fatal("expected reviewsLoaded flag to be true")
	}
	if entry.firstReviewAt == nil || entry.firstApprovalAt == nil {
		t.Fatal("expected firstReviewAt and firstApprovalAt to be set")
	}
	if entry.firstReviewAt.Sub(base) != 2*time.Hour {
		t.Fatalf("expected first review at 2h, got %v", entry.firstReviewAt.Sub(base))
	}
	if entry.firstApprovalAt.Sub(base) != 5*time.Hour {
		t.Fatalf("expected first approval at 5h, got %v", entry.firstApprovalAt.Sub(base))
	}
}

func TestPRQueueView_handleKeyPress_EnterOpensDetail(t *testing.T) {
	view := NewPRQueueView()
	view.owner = "owner"
	view.repo = "repo"
	view.prRepo = &testPRRepo{}
	view.entries = []*prQueueEntry{
		{pr: &models.PullRequest{Number: 1, Title: "Example", CreatedAt: time.Now()}},
	}
	view.width = 80
	view.height = 24

	_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command to initialize detail view")
	}
	if !view.showingDetail {
		t.Fatal("expected showingDetail to be true")
	}
}

func TestPRQueueView_handleErrorState(t *testing.T) {
	view := NewPRQueueView()
	view.width = 80
	view.height = 24
	view.statusBar.SetSize(80, 1)
	msg := prQueueLoadedMsg{err: errors.New("boom")}
	view.Update(msg)
	if view.err == nil {
		t.Fatal("expected error to be set")
	}
	output := view.View()
	if !containsString(output, "boom") {
		t.Fatalf("expected error message in view, got %s", output)
	}
}

func TestPRQueueView_cursorHighlightsSelection(t *testing.T) {
	view := NewPRQueueView()
	view.width = 80
	view.height = 20
	view.entries = []*prQueueEntry{
		{pr: &models.PullRequest{Number: 1, Title: "Alpha", CreatedAt: time.Now().Add(-3 * time.Hour)}},
		{pr: &models.PullRequest{Number: 2, Title: "Beta", CreatedAt: time.Now().Add(-2 * time.Hour)}},
	}

	first := view.View()
	view.cursor = 1
	second := view.View()

	if first == second {
		t.Fatal("expected view output to change when cursor moves")
	}
}

func TestPRQueueView_highlightsFirstEntryAfterLoad(t *testing.T) {
	view := NewPRQueueView()
	view.width = 80
	view.height = 20

	now := time.Now()
	msg := prQueueLoadedMsg{
		prs: []*models.PullRequest{
			{Number: 2, Title: "Beta", CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 1, Title: "Alpha", CreatedAt: now.Add(-2 * time.Hour)},
		},
	}

	view.Update(msg)
	output := view.View()
	cursorMarker := styles.CursorStyle.Render("▶ ")
	if !containsString(output, cursorMarker) {
		t.Fatalf("expected output to contain the cursor marker %q, got %q", cursorMarker, output)
	}
}

func TestPRQueueView_ViewShowsHeaderAndStatus(t *testing.T) {
	view := NewPRQueueView()
	view.width = 80
	view.height = 20
	view.owner = "owner"
	view.repo = "repo"
	view.reviewLoading = true
	view.entries = []*prQueueEntry{
		{pr: &models.PullRequest{Number: 1, Title: "Alpha", CreatedAt: time.Now().Add(-1 * time.Hour)}},
	}

	output := view.View()

	if !containsString(output, "Review Queue") {
		t.Fatalf("expected header to include %q, got %q", "Review Queue", output)
	}
	if !containsString(output, "Queue") {
		t.Fatalf("expected status bar to include mode label %q, got %q", "Queue", output)
	}
	if !containsString(output, "owner/repo") {
		t.Fatalf("expected status bar to include repo label, got %q", output)
	}
}

func TestPRQueueView_resetsCursorWhenOutOfRange(t *testing.T) {
	view := NewPRQueueView()
	view.width = 80
	view.height = 20
	now := time.Now()
	view.entries = []*prQueueEntry{
		{pr: &models.PullRequest{Number: 1, Title: "Alpha", CreatedAt: now}},
		{pr: &models.PullRequest{Number: 2, Title: "Beta", CreatedAt: now}},
	}

	view.cursor = 5
	output := view.View()

	if view.cursor != 0 {
		t.Fatalf("expected cursor to reset to 0, got %d", view.cursor)
	}

	cursorMarker := styles.CursorStyle.Render("▶ ")
	if !containsString(output, cursorMarker) {
		t.Fatalf("expected cursor marker in output, got %q", output)
	}
}

// testPRRepo is a minimal pull request repository used for tests.
type testPRRepo struct{}

func (r *testPRRepo) List(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
	return nil, nil
}

func (r *testPRRepo) Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error) {
	return nil, nil
}

func (r *testPRRepo) Create(ctx context.Context, owner, repo string, input *models.CreatePRInput) (*models.PullRequest, error) {
	return nil, nil
}

func (r *testPRRepo) Update(ctx context.Context, owner, repo string, number int, input *models.UpdatePRInput) (*models.PullRequest, error) {
	return nil, nil
}

func (r *testPRRepo) Merge(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
	return nil
}

func (r *testPRRepo) Close(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func (r *testPRRepo) Reopen(ctx context.Context, owner, repo string, number int) error {
	return nil
}

func (r *testPRRepo) GetDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	return "", nil
}

func (r *testPRRepo) IsMergeable(ctx context.Context, owner, repo string, number int) (bool, error) {
	return false, nil
}

func (r *testPRRepo) ListReviews(ctx context.Context, owner, repo string, number int) ([]*models.Review, error) {
	return []*models.Review{}, nil
}

func (r *testPRRepo) ListComments(ctx context.Context, owner, repo string, number int, opts *models.CommentOptions) ([]*models.Comment, error) {
	return []*models.Comment{}, nil
}

var _ repository.PullRequestRepository = (*testPRRepo)(nil)
