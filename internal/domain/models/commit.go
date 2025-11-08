package models

import "time"

// Commit represents a Git commit
type Commit struct {
	SHA       string
	Message   string
	Author    CommitAuthor
	Committer CommitAuthor
	Parents   []string
	Tree      string
	URL       string
	Stats     *CommitStats
	Files     []*DiffFile
	CreatedAt time.Time
}

// CommitStats represents statistics about a commit
type CommitStats struct {
	Additions int
	Deletions int
	Total     int
}

// CommitOptions represents options for listing commits
type CommitOptions struct {
	SHA       string
	Path      string
	Author    string
	Since     *time.Time
	Until     *time.Time
	Page      int
	PerPage   int
}
