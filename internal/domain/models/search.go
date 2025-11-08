package models

// SearchType represents the type of search (issues, pull requests, or both)
type SearchType string

const (
	SearchTypeIssue SearchType = "issue"
	SearchTypePR    SearchType = "pr"
	SearchTypeBoth  SearchType = "both"
)

// SearchSort represents the field to sort search results by
type SearchSort string

const (
	SearchSortCreated      SearchSort = "created"
	SearchSortUpdated      SearchSort = "updated"
	SearchSortComments     SearchSort = "comments"
	SearchSortReactions    SearchSort = "reactions"
	SearchSortInteractions SearchSort = "interactions"
)

// SearchOptions represents options for searching issues and pull requests
type SearchOptions struct {
	Query      string           // Search query string
	Type       SearchType       // Type of items to search (issue, pr, or both)
	State      IssueState       // State filter (open, closed, all)
	Author     string           // Filter by author username
	Labels     []string         // Filter by labels
	Sort       SearchSort       // Sort field
	Direction  SortDirection    // Sort direction (asc, desc)
	Page       int              // Page number for pagination
	PerPage    int              // Number of results per page
}

// SearchResult represents a single search result (can be Issue or PR)
type SearchResult struct {
	Type        SearchType     // Type of the result (issue or pr)
	Issue       *Issue         // Issue data (if Type == SearchTypeIssue)
	PullRequest *PullRequest   // PR data (if Type == SearchTypePR)
}

// SearchResults represents the result of a search query
type SearchResults struct {
	TotalCount        int             // Total number of results
	IncompleteResults bool            // Whether the results are incomplete
	Items             []SearchResult  // List of search results
}
