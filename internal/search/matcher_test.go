package search

import (
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

func TestNewMatcher(t *testing.T) {
	query := ParseQuery("bug")
	matcher := NewMatcher(query)

	if matcher == nil {
		t.Fatal("NewMatcher should not return nil")
	}

	if matcher.query != query {
		t.Error("NewMatcher should store the query")
	}
}

func TestMatcher_Match_EmptyQuery(t *testing.T) {
	query := ParseQuery("")
	matcher := NewMatcher(query)

	issue := &models.Issue{
		Title: "Test issue",
	}

	if !matcher.Match(issue) {
		t.Error("Empty query should match all issues")
	}
}

func TestMatcher_Match_Keyword(t *testing.T) {
	query := ParseQuery("login")
	matcher := NewMatcher(query)

	testCases := []struct {
		name    string
		issue   *models.Issue
		matches bool
	}{
		{
			name: "keyword in title",
			issue: &models.Issue{
				Title: "Login bug",
				Body:  "Some description",
			},
			matches: true,
		},
		{
			name: "keyword in body",
			issue: &models.Issue{
				Title: "Bug report",
				Body:  "The login feature is broken",
			},
			matches: true,
		},
		{
			name: "case insensitive",
			issue: &models.Issue{
				Title: "LOGIN ERROR",
				Body:  "",
			},
			matches: true,
		},
		{
			name: "no match",
			issue: &models.Issue{
				Title: "Bug report",
				Body:  "Something else",
			},
			matches: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Match(tc.issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_Phrase(t *testing.T) {
	query := ParseQuery(`"login error"`)
	matcher := NewMatcher(query)

	testCases := []struct {
		name    string
		issue   *models.Issue
		matches bool
	}{
		{
			name: "exact phrase in title",
			issue: &models.Issue{
				Title: "Fix login error",
				Body:  "",
			},
			matches: true,
		},
		{
			name: "exact phrase in body",
			issue: &models.Issue{
				Title: "Bug",
				Body:  "There is a login error when clicking submit",
			},
			matches: true,
		},
		{
			name: "words exist but not as phrase",
			issue: &models.Issue{
				Title: "Error in login process",
				Body:  "",
			},
			matches: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Match(tc.issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_Author(t *testing.T) {
	query := ParseQuery("author:john")
	matcher := NewMatcher(query)

	testCases := []struct {
		name    string
		issue   *models.Issue
		matches bool
	}{
		{
			name: "author matches",
			issue: &models.Issue{
				Title: "Bug",
				Author: models.User{
					Login: "john",
				},
			},
			matches: true,
		},
		{
			name: "case insensitive",
			issue: &models.Issue{
				Title: "Bug",
				Author: models.User{
					Login: "JOHN",
				},
			},
			matches: true,
		},
		{
			name: "author doesn't match",
			issue: &models.Issue{
				Title: "Bug",
				Author: models.User{
					Login: "jane",
				},
			},
			matches: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Match(tc.issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_Label(t *testing.T) {
	query := ParseQuery("label:bug")
	matcher := NewMatcher(query)

	testCases := []struct {
		name    string
		issue   *models.Issue
		matches bool
	}{
		{
			name: "has label",
			issue: &models.Issue{
				Title: "Test",
				Labels: []models.Label{
					{Name: "bug"},
				},
			},
			matches: true,
		},
		{
			name: "has label with others",
			issue: &models.Issue{
				Title: "Test",
				Labels: []models.Label{
					{Name: "feature"},
					{Name: "bug"},
					{Name: "priority-high"},
				},
			},
			matches: true,
		},
		{
			name: "case insensitive",
			issue: &models.Issue{
				Title: "Test",
				Labels: []models.Label{
					{Name: "BUG"},
				},
			},
			matches: true,
		},
		{
			name: "doesn't have label",
			issue: &models.Issue{
				Title: "Test",
				Labels: []models.Label{
					{Name: "feature"},
				},
			},
			matches: false,
		},
		{
			name: "no labels",
			issue: &models.Issue{
				Title:  "Test",
				Labels: []models.Label{},
			},
			matches: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Match(tc.issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_State(t *testing.T) {
	testCases := []struct {
		name       string
		queryState string
		issueState models.IssueState
		matches    bool
	}{
		{"open matches open", "state:open", models.IssueStateOpen, true},
		{"open doesn't match closed", "state:open", models.IssueStateClosed, false},
		{"closed matches closed", "state:closed", models.IssueStateClosed, true},
		{"closed doesn't match open", "state:closed", models.IssueStateOpen, false},
		{"all matches open", "state:all", models.IssueStateOpen, true},
		{"all matches closed", "state:all", models.IssueStateClosed, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.queryState)
			matcher := NewMatcher(query)

			issue := &models.Issue{
				Title: "Test",
				State: tc.issueState,
			}

			result := matcher.Match(issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_DateCreated(t *testing.T) {
	issueDate := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		query   string
		matches bool
	}{
		{"greater than - match", "created:>2024-01-01", true},
		{"greater than - no match", "created:>2024-12-31", false},
		{"less than - match", "created:<2024-12-31", true},
		{"less than - no match", "created:<2024-01-01", false},
		{"equal - match", "created:2024-06-15", true},
		{"equal - no match", "created:2024-06-14", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.query)
			matcher := NewMatcher(query)

			issue := &models.Issue{
				Title:     "Test",
				CreatedAt: issueDate,
			}

			result := matcher.Match(issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_DateUpdated(t *testing.T) {
	issueDate := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		query   string
		matches bool
	}{
		{"greater than - match", "updated:>2024-01-01", true},
		{"less than - match", "updated:<2024-12-31", true},
		{"equal - match", "updated:2024-06-15", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.query)
			matcher := NewMatcher(query)

			issue := &models.Issue{
				Title:     "Test",
				UpdatedAt: issueDate,
			}

			result := matcher.Match(issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Match_MultipleTerms(t *testing.T) {
	query := ParseQuery("author:john label:bug login")
	matcher := NewMatcher(query)

	testCases := []struct {
		name    string
		issue   *models.Issue
		matches bool
	}{
		{
			name: "all terms match",
			issue: &models.Issue{
				Title: "Login problem",
				Author: models.User{
					Login: "john",
				},
				Labels: []models.Label{
					{Name: "bug"},
				},
			},
			matches: true,
		},
		{
			name: "missing author",
			issue: &models.Issue{
				Title: "Login problem",
				Author: models.User{
					Login: "jane",
				},
				Labels: []models.Label{
					{Name: "bug"},
				},
			},
			matches: false,
		},
		{
			name: "missing label",
			issue: &models.Issue{
				Title: "Login problem",
				Author: models.User{
					Login: "john",
				},
				Labels: []models.Label{
					{Name: "feature"},
				},
			},
			matches: false,
		},
		{
			name: "missing keyword",
			issue: &models.Issue{
				Title: "Bug report",
				Author: models.User{
					Login: "john",
				},
				Labels: []models.Label{
					{Name: "bug"},
				},
			},
			matches: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Match(tc.issue)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestMatcher_Filter(t *testing.T) {
	query := ParseQuery("label:bug")
	matcher := NewMatcher(query)

	issues := []*models.Issue{
		{
			Title: "Issue 1",
			Labels: []models.Label{
				{Name: "bug"},
			},
		},
		{
			Title: "Issue 2",
			Labels: []models.Label{
				{Name: "feature"},
			},
		},
		{
			Title: "Issue 3",
			Labels: []models.Label{
				{Name: "bug"},
				{Name: "critical"},
			},
		},
	}

	filtered := matcher.Filter(issues)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered issues, got %d", len(filtered))
	}

	if filtered[0].Title != "Issue 1" {
		t.Errorf("Expected first filtered issue to be 'Issue 1', got %q", filtered[0].Title)
	}

	if filtered[1].Title != "Issue 3" {
		t.Errorf("Expected second filtered issue to be 'Issue 3', got %q", filtered[1].Title)
	}
}

func TestMatcher_Filter_EmptyQuery(t *testing.T) {
	query := ParseQuery("")
	matcher := NewMatcher(query)

	issues := []*models.Issue{
		{Title: "Issue 1"},
		{Title: "Issue 2"},
		{Title: "Issue 3"},
	}

	filtered := matcher.Filter(issues)

	if len(filtered) != 3 {
		t.Errorf("Expected all 3 issues, got %d", len(filtered))
	}
}

func TestMatcher_Filter_NoMatches(t *testing.T) {
	query := ParseQuery("label:nonexistent")
	matcher := NewMatcher(query)

	issues := []*models.Issue{
		{
			Title: "Issue 1",
			Labels: []models.Label{
				{Name: "bug"},
			},
		},
		{
			Title: "Issue 2",
			Labels: []models.Label{
				{Name: "feature"},
			},
		},
	}

	filtered := matcher.Filter(issues)

	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered issues, got %d", len(filtered))
	}
}
