package cache

import (
	"context"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/domain/repository"
)

// CachedIssueRepository wraps an IssueRepository with caching functionality
type CachedIssueRepository struct {
	repo  repository.IssueRepository
	cache *Cache
}

// NewCachedIssueRepository creates a new cached issue repository
func NewCachedIssueRepository(repo repository.IssueRepository, cache *Cache) repository.IssueRepository {
	return &CachedIssueRepository{
		repo:  repo,
		cache: cache,
	}
}

// List retrieves a list of issues with caching
func (r *CachedIssueRepository) List(ctx context.Context, owner, repo string, opts *models.IssueOptions) ([]*models.Issue, error) {
	// Generate cache key
	key := r.cache.GenerateKey("issues:list", owner, repo, opts)

	// Try to get from cache (respecting context options like SkipCache)
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if issues, ok := cached.([]*models.Issue); ok {
			return issues, nil
		}
	}

	// Cache miss or skip cache - fetch from underlying repository
	issues, err := r.repo.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	// Store in cache (respecting context options)
	// defaultTTL=0 means use Config's default TTL
	_ = r.cache.SetWithContext(ctx, key, issues, 0)

	return issues, nil
}

// Get retrieves a single issue by number with caching
func (r *CachedIssueRepository) Get(ctx context.Context, owner, repo string, number int) (*models.Issue, error) {
	// Generate cache key
	key := r.cache.GenerateKey("issues:get", owner, repo, number)

	// Try to get from cache
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if issue, ok := cached.(*models.Issue); ok {
			return issue, nil
		}
	}

	// Cache miss - fetch from underlying repository
	issue, err := r.repo.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.SetWithContext(ctx, key, issue, 0)

	return issue, nil
}

// Create creates a new issue (invalidates list cache)
func (r *CachedIssueRepository) Create(ctx context.Context, owner, repo string, input *models.CreateIssueInput) (*models.Issue, error) {
	issue, err := r.repo.Create(ctx, owner, repo, input)
	if err != nil {
		return nil, err
	}

	// Invalidate list caches for this repository
	// Note: We can't easily invalidate all list caches with different options,
	// so we rely on TTL expiration. For more aggressive invalidation,
	// we could track cache keys or use cache tags.

	return issue, nil
}

// Update updates an existing issue (invalidates caches)
func (r *CachedIssueRepository) Update(ctx context.Context, owner, repo string, number int, input *models.UpdateIssueInput) (*models.Issue, error) {
	issue, err := r.repo.Update(ctx, owner, repo, number, input)
	if err != nil {
		return nil, err
	}

	// Invalidate the specific issue cache
	key := r.cache.GenerateKey("issues:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return issue, nil
}

// Close closes an issue (invalidates caches)
func (r *CachedIssueRepository) Close(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Close(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific issue cache
	key := r.cache.GenerateKey("issues:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// Reopen reopens a closed issue (invalidates caches)
func (r *CachedIssueRepository) Reopen(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Reopen(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific issue cache
	key := r.cache.GenerateKey("issues:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// Lock locks an issue (invalidates caches)
func (r *CachedIssueRepository) Lock(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Lock(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific issue cache
	key := r.cache.GenerateKey("issues:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// Unlock unlocks an issue (invalidates caches)
func (r *CachedIssueRepository) Unlock(ctx context.Context, owner, repo string, number int) error {
	err := r.repo.Unlock(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	// Invalidate the specific issue cache
	key := r.cache.GenerateKey("issues:get", owner, repo, number)
	_ = r.cache.Delete(key)

	return nil
}

// ListComments retrieves comments for an issue with caching
func (r *CachedIssueRepository) ListComments(ctx context.Context, owner, repo string, number int, opts *models.CommentOptions) ([]*models.Comment, error) {
	// Generate cache key
	key := r.cache.GenerateKey("issues:comments", owner, repo, number, opts)

	// Try to get from cache
	if cached, ok := r.cache.GetWithContext(ctx, key); ok {
		if comments, ok := cached.([]*models.Comment); ok {
			return comments, nil
		}
	}

	// Cache miss - fetch from underlying repository
	comments, err := r.repo.ListComments(ctx, owner, repo, number, opts)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = r.cache.SetWithContext(ctx, key, comments, 0)

	return comments, nil
}
