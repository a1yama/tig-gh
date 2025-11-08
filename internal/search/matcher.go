package search

import (
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// Matcher matches issues against a search query
type Matcher struct {
	query *Query
}

// NewMatcher creates a new matcher with the given query
func NewMatcher(query *Query) *Matcher {
	return &Matcher{
		query: query,
	}
}

// Match returns true if the issue matches the query
func (m *Matcher) Match(issue *models.Issue) bool {
	if m.query.IsEmpty() {
		return true
	}

	for _, term := range m.query.Terms {
		if !m.matchTerm(issue, term) {
			return false // All terms must match (AND logic)
		}
	}

	return true
}

// matchTerm checks if a single term matches the issue
func (m *Matcher) matchTerm(issue *models.Issue, term SearchTerm) bool {
	switch term.Type {
	case TermTypeKeyword:
		return m.matchKeyword(issue, term.Value)

	case TermTypePhrase:
		return m.matchPhrase(issue, term.Value)

	case TermTypeAuthor:
		return m.matchAuthor(issue, term.Value)

	case TermTypeLabel:
		return m.matchLabel(issue, term.Value)

	case TermTypeState:
		return m.matchState(issue, term.Value)

	case TermTypeDate:
		return m.matchDate(issue, term)

	default:
		return false
	}
}

// matchKeyword checks if the keyword appears in issue title or body
func (m *Matcher) matchKeyword(issue *models.Issue, keyword string) bool {
	keyword = strings.ToLower(keyword)

	// Search in title
	if strings.Contains(strings.ToLower(issue.Title), keyword) {
		return true
	}

	// Search in body
	if strings.Contains(strings.ToLower(issue.Body), keyword) {
		return true
	}

	return false
}

// matchPhrase checks if the exact phrase appears in issue title or body
func (m *Matcher) matchPhrase(issue *models.Issue, phrase string) bool {
	phrase = strings.ToLower(phrase)

	// Search in title
	if strings.Contains(strings.ToLower(issue.Title), phrase) {
		return true
	}

	// Search in body
	if strings.Contains(strings.ToLower(issue.Body), phrase) {
		return true
	}

	return false
}

// matchAuthor checks if the issue author matches
func (m *Matcher) matchAuthor(issue *models.Issue, author string) bool {
	return strings.EqualFold(issue.Author.Login, author)
}

// matchLabel checks if the issue has the specified label
func (m *Matcher) matchLabel(issue *models.Issue, labelName string) bool {
	labelName = strings.ToLower(labelName)

	for _, label := range issue.Labels {
		if strings.ToLower(label.Name) == labelName {
			return true
		}
	}

	return false
}

// matchState checks if the issue state matches
func (m *Matcher) matchState(issue *models.Issue, state string) bool {
	state = strings.ToLower(state)

	switch state {
	case "open":
		return issue.State == models.IssueStateOpen
	case "closed":
		return issue.State == models.IssueStateClosed
	case "all":
		return true
	default:
		return false
	}
}

// matchDate checks if the issue date matches the criteria
func (m *Matcher) matchDate(issue *models.Issue, term SearchTerm) bool {
	var issueDate *time.Time

	switch term.Field {
	case "created":
		issueDate = &issue.CreatedAt
	case "updated":
		issueDate = &issue.UpdatedAt
	default:
		return false
	}

	if issueDate == nil {
		return false
	}

	return term.MatchDate(*issueDate)
}

// Filter filters a list of issues based on the query
func (m *Matcher) Filter(issues []*models.Issue) []*models.Issue {
	if m.query.IsEmpty() {
		return issues
	}

	var filtered []*models.Issue
	for _, issue := range issues {
		if m.Match(issue) {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}
