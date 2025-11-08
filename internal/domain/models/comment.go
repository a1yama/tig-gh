package models

import "time"

// Comment represents a comment on an issue or pull request
type Comment struct {
	ID        int64
	User      User
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	HTMLURL   string
}

// CommentOptions represents options for listing comments
type CommentOptions struct {
	// Sort order (created, updated)
	Sort string

	// Sort direction (asc, desc)
	Direction string

	// Results per page (max 100)
	PerPage int

	// Page number
	Page int
}
