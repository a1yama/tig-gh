package models

import "time"

// User represents a GitHub user
type User struct {
	ID        int64
	Login     string
	Name      string
	Email     string
	AvatarURL string
}

// Label represents a GitHub label
type Label struct {
	ID          int64
	Name        string
	Color       string
	Description string
}

// Milestone represents a GitHub milestone
type Milestone struct {
	ID           int64
	Number       int
	Title        string
	Description  string
	State        MilestoneState
	OpenIssues   int
	ClosedIssues int
	DueOn        *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MilestoneState represents the state of a milestone
type MilestoneState string

const (
	MilestoneStateOpen   MilestoneState = "open"
	MilestoneStateClosed MilestoneState = "closed"
)

// CommitAuthor represents the author or committer of a commit
type CommitAuthor struct {
	Name  string
	Email string
	Date  time.Time
}

// Branch represents a Git branch
type Branch struct {
	Name string
	SHA  string
}

// Review represents a pull request review
type Review struct {
	ID          int64
	User        User
	Body        string
	State       ReviewState
	SubmittedAt time.Time
}

// ReviewState represents the state of a review
type ReviewState string

const (
	ReviewStateApproved         ReviewState = "approved"
	ReviewStateChangesRequested ReviewState = "changes_requested"
	ReviewStateCommented        ReviewState = "commented"
	ReviewStateDismissed        ReviewState = "dismissed"
	ReviewStatePending          ReviewState = "pending"
)

// Comparison represents a comparison between two commits
type Comparison struct {
	BaseCommit   *Commit
	MergeCommit  *Commit
	Status       ComparisonStatus
	AheadBy      int
	BehindBy     int
	TotalCommits int
	Commits      []*Commit
	Files        []*DiffFile
}

// ComparisonStatus represents the status of a comparison
type ComparisonStatus string

const (
	ComparisonStatusIdentical ComparisonStatus = "identical"
	ComparisonStatusAhead     ComparisonStatus = "ahead"
	ComparisonStatusBehind    ComparisonStatus = "behind"
	ComparisonStatusDiverged  ComparisonStatus = "diverged"
)

// DiffFile represents a file in a diff
type DiffFile struct {
	Filename  string
	Status    FileStatus
	Additions int
	Deletions int
	Changes   int
	Patch     string
}

// FileStatus represents the status of a file in a diff
type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusRemoved  FileStatus = "removed"
	FileStatusRenamed  FileStatus = "renamed"
)
