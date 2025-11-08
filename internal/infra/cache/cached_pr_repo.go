package cache

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CachedPullRequestRepository wraps a PullRequestRepository with caching functionality
type CachedPullRequestRepository struct {
	repo  repository.PullRequestRepository
	cache *Cache
}

// NewCachedPullRequestRepository creates a new cached pull request repository
func NewCachedPullRequestRepository(repo repository.PullRequestRepository, cache *Cache) repository.PullRequestRepository {
	return &CachedPullRequestRepository{
		repo:  repo,
		cache: cache,
	}
}

// List retrieves a list of pull requests with caching
func (r *CachedPullRequestRepository) List(ctx context.Context, owner, repo string, opts *models.PROptions) ([]*models.PullRequest, error) {
	// Generate cache key
	key := r.cache.GenerateKey("prs:list", owner, repo, opts)

	// Try to get from cache (respecting context options like SkipCache)
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if prs, ok := cached.([]*models.PullRequest); ok {
			return prs, nil
		}
	}

	// Cache miss or skip cache - fetch from underlying repository
	prs, err := r.repo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	// Store in cache (respecting context options)
	// defaultTTL=0 means use Config's default TTL
	_ = r.cache.SetWithContext(ctx, key, prs, 0)

	return prs, nil
}

// Get retrieves a single pull request by number with caching
func (r *CachedPullRequestRepository) Get(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error) {
	// Generate cache key
	key := r.cache.GenerateKey("prs:get", owner, repo, number)

	// Try to get from cache
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if pr, ok := cached.(*models.PullRequest); ok {
			return pr, nil
		}
	}

	// Cache miss - fetch from underlying repository
	pr, err := r.repo.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.SetWithContext(ctx, key, pr, 0)

	return pr, nil
}

// Create creates a new pull request (invalidates list cache)
func (r *CachedPullRequestRepository) Create(ctx context.Context, owner, repo string, input *models.CreatePRInput) (*models.PullRequest, error) {
	pr, err := r.repo.Create(ctx, owner, repo, input)
	if err != nil {
		return nil, err
	}

	// Invalidate list caches for this repository
	// Note: We can't easily invalidate all list caches with different options,
	// so we rely on TTL expiration. For more aggressive invalidation,
	// we could track cache keys or use cache tags.

	return pr, nil
}

// Update updates an existing pull request (invalidates caches)
func (r *CachedPullRequestRepository) Update(ctx context.Context, owner, repo string, number int, input *models.UpdatePRInput) (*models.PullRequest, error) {
	pr, err := r.repo.Update(ctx, owner, repo, number, input)
	if err != nil {
		return nil, err
	}

	// Invalidate the specific PR cache
	key := r.cache.GenerateKey("prs:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return pr, nil
}

// Merge merges a pull request (invalidates caches)
func (r *CachedPullRequestRepository) Merge(ctx context.Context, owner, repo string, number int, opts *models.MergeOptions) error {
	err := r.repo.Merge(ctx, owner, repo, number, opts)
	if err != nil {
		return err
	}

	// Invalidate the specific PR cache
	key := r.cache.GenerateKey("prs:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// Close closes a pull request (invalidates caches)
func (r *CachedPullRequestRepository) Close(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Close(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific PR cache
	key := r.cache.GenerateKey("prs:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// Reopen reopens a closed pull request (invalidates caches)
func (r *CachedPullRequestRepository) Reopen(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Reopen(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific PR cache
	key := r.cache.GenerateKey("prs:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// GetDiff retrieves the diff for a pull request with caching
func (r *CachedPullRequestRepository) GetDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	// Generate cache key
	key := r.cache.GenerateKey("prs:diff", owner, repo, number)

	// Try to get from cache
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if diff, ok := cached.(string); ok {
			return diff, nil
		}
	}

	// Cache miss - fetch from underlying repository
	diff, err := r.repo.GetDiff(ctx, owner, repo, number)
	if err != nil {
		return "", err
	}

	// Store in cache
	_ = r.cache.SetWithContext(ctx, key, diff, 0)

	return diff, nil
}

// IsMergeable checks if a pull request is mergeable (no caching - always fresh)
func (r *CachedPullRequestRepository) IsMergeable(ctx context.Context, owner, repo string, number int) (bool, error) {
	// Don't cache mergeable status as it changes frequently
	return r.repo.IsMergeable(ctx, owner, repo, number)
}

// ListReviews retrieves reviews for a pull request with caching
func (r *CachedPullRequestRepository) ListReviews(ctx context.Context, owner, repo string, number int) ([]*models.Review, error) {
	// Generate cache key
	key := r.cache.GenerateKey("prs:reviews", owner, repo, number)

	// Try to get from cache
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if reviews, ok := cached.([]*models.Review); ok {
			return reviews, nil
		}
	}

	// Cache miss - fetch from underlying repository
	reviews, err := r.repo.ListReviews(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.SetWithContext(ctx, key, reviews, 0)

	return reviews, nil
}
