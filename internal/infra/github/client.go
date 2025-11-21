package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
}

// NewClient creates a new GitHub API client with authentication
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
	}
}

// NewClientWithHTTPClient creates a new GitHub API client with a custom HTTP client
func NewClientWithHTTPClient(httpClient *http.Client) *Client {
	return &Client{
		client: github.NewClient(httpClient),
	}
}

// GetClient returns the underlying GitHub client
func (c *Client) GetClient() *github.Client {
	return c.client
}

// GetRateLimit returns the current rate limit status
func (c *Client) GetRateLimit(ctx context.Context) (*github.RateLimits, error) {
	limits, _, err := c.client.RateLimits(ctx)
	return limits, err
}
