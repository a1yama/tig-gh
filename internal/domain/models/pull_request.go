package models

import "time"

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID               int64
	Number           int
	Title            string
	Body             string
	State            PRState
	Author           User
	Head             Branch
	Base             Branch
	Mergeable        bool
	MergeableState   string
	Merged           bool
	MergedAt         *time.Time
	MergedBy         *User
	Draft            bool
	Locked           bool
	Reviews          []Review
	RequestedReviewers []User
	Assignees        []User
	Labels           []Label
	Milestone        *Milestone
	Comments         int
	Commits          int
	Additions        int
	Deletions        int
	ChangedFiles     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ClosedAt         *time.Time
}

// PRState represents the state of a pull request
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateAll    PRState = "all"
)

// PROptions represents options for listing pull requests
type PROptions struct {
	State     PRState
	Head      string
	Base      string
	Sort      PRSort
	Direction SortDirection
	Page      int
	PerPage   int
}

// PRSort represents the field to sort pull requests by
type PRSort string

const (
	PRSortCreated      PRSort = "created"
	PRSortUpdated      PRSort = "updated"
	PRSortPopularity   PRSort = "popularity"
	PRSortLongRunning  PRSort = "long-running"
)

// MergeOptions represents options for merging a pull request
type MergeOptions struct {
	CommitTitle   string
	CommitMessage string
	SHA           string
	MergeMethod   MergeMethod
}

// MergeMethod represents the method to use for merging
type MergeMethod string

const (
	MergeMethodMerge  MergeMethod = "merge"
	MergeMethodSquash MergeMethod = "squash"
	MergeMethodRebase MergeMethod = "rebase"
)

// CreatePRInput represents the input for creating a pull request
type CreatePRInput struct {
	Title string
	Body  string
	Head  string
	Base  string
	Draft bool
}

// UpdatePRInput represents the input for updating a pull request
type UpdatePRInput struct {
	Title     *string
	Body      *string
	State     *PRState
	Base      *string
	Assignees *[]string
	Labels    *[]string
	Milestone *int
}
