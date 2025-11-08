package models

import "time"

// IssueState represents the state of an issue
type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
	IssueStateAll    IssueState = "all"
)

// Issue represents a GitHub issue
type Issue struct {
	ID        int64
	Number    int
	Title     string
	Body      string
	State     IssueState
	Author    User
	Assignees []User
	Labels    []Label
	Milestone *Milestone
	Comments  int
	Locked    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	URL       string
	HTMLURL   string
}

// IssueOptions represents options for listing issues
type IssueOptions struct {
	State     IssueState
	Labels    []string
	Assignee  string
	Creator   string
	Mentioned string
	Milestone string
	Sort      IssueSort
	Direction SortDirection
	Since     *time.Time
	Page      int
	PerPage   int
}

// IssueSort represents the field to sort issues by
type IssueSort string

const (
	IssueSortCreated  IssueSort = "created"
	IssueSortUpdated  IssueSort = "updated"
	IssueSortComments IssueSort = "comments"
)

// SortDirection represents the direction of sorting
type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

// CreateIssueInput represents input for creating an issue
type CreateIssueInput struct {
	Title     string
	Body      string
	Assignees []string
	Labels    []string
	Milestone int
}

// UpdateIssueInput represents input for updating an issue
type UpdateIssueInput struct {
	Title     *string
	Body      *string
	State     *IssueState
	Assignees *[]string
	Labels    *[]string
	Milestone *int
}
