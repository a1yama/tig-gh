package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/infra/cache"
	"github.com/a1yama/tig-gh/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedPullRequestRepository_List_CacheHit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.PROptions{
		State: models.PRStateOpen,
	}

	expectedPRs := []*models.PullRequest{
		{Number: 1, Title: "Test PR 1"},
		{Number: 2, Title: "Test PR 2"},
	}

	// First call - should hit the mock
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(expectedPRs, nil).
		Times(1)

	prs1, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, expectedPRs, prs1)

	// Second call - should hit cache, mock should not be called again
	prs2, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, expectedPRs, prs2)
}

func TestCachedPullRequestRepository_List_SkipCache(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.PROptions{
		State: models.PRStateOpen,
	}

	prs1 := []*models.PullRequest{{Number: 1, Title: "Test PR 1"}}
	prs2 := []*models.PullRequest{
		{Number: 1, Title: "Updated PR 1"},
		{Number: 2, Title: "New PR 2"},
	}

	// First call to populate cache
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(prs1, nil).
		Times(1)

	_, err = cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)

	// Second call with SkipCache context - should bypass cache
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(prs2, nil).
		Times(1)

	ctx := cache.WithSkipCacheContext(context.Background())
	prs, err := cachedRepo.List(ctx, owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, prs2, prs)
}

func TestCachedPullRequestRepository_List_Error(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.PROptions{
		State: models.PRStateOpen,
	}

	expectedErr := errors.New("API error")

	// Mock should return error
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(nil, expectedErr).
		Times(1)

	prs, err := cachedRepo.List(context.Background(), owner, repo, opts)
	assert.Error(t, err)
	assert.Nil(t, prs)
	assert.Equal(t, expectedErr, err)
}

func TestCachedPullRequestRepository_Get_CacheHit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123

	expectedPR := &models.PullRequest{
		Number: number,
		Title:  "Test PR",
	}

	// First call - should hit the mock
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(expectedPR, nil).
		Times(1)

	pr1, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedPR, pr1)

	// Second call - should hit cache
	pr2, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedPR, pr2)
}

func TestCachedPullRequestRepository_Update_InvalidatesCache(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123

	originalPR := &models.PullRequest{
		Number: number,
		Title:  "Original Title",
	}

	// Populate cache
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(originalPR, nil).
		Times(1)

	_, err = cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)

	// Update PR
	updateInput := &models.UpdatePRInput{
		Title: stringPtrPR("Updated Title"),
	}
	updatedPR := &models.PullRequest{
		Number: number,
		Title:  "Updated Title",
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), owner, repo, number, updateInput).
		Return(updatedPR, nil).
		Times(1)

	_, err = cachedRepo.Update(context.Background(), owner, repo, number, updateInput)
	require.NoError(t, err)

	// Get again - should fetch fresh data (cache was invalidated)
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(updatedPR, nil).
		Times(1)

	pr, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", pr.Title)
}

func TestCachedPullRequestRepository_GetDiff_CacheHit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123
	expectedDiff := "diff --git a/file.go b/file.go\n..."

	// First call - should hit the mock
	mockRepo.EXPECT().
		GetDiff(gomock.Any(), owner, repo, number).
		Return(expectedDiff, nil).
		Times(1)

	diff1, err := cachedRepo.GetDiff(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedDiff, diff1)

	// Second call - should hit cache
	diff2, err := cachedRepo.GetDiff(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedDiff, diff2)
}

func TestCachedPullRequestRepository_IsMergeable_NoCaching(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123

	// IsMergeable should NOT be cached - expect 2 calls
	mockRepo.EXPECT().
		IsMergeable(gomock.Any(), owner, repo, number).
		Return(true, nil).
		Times(2)

	mergeable1, err := cachedRepo.IsMergeable(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.True(t, mergeable1)

	// Second call should also hit the mock (no caching)
	mergeable2, err := cachedRepo.IsMergeable(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.True(t, mergeable2)
}

func TestCachedPullRequestRepository_CustomTTL(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockPullRequestRepository(ctrl)
	cacheConfig := cache.DefaultConfig().
		DisableFileCache().
		WithMemoryTTL(100 * time.Millisecond) // Very short TTL for testing

	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedPullRequestRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.PROptions{
		State: models.PRStateOpen,
	}

	prs1 := []*models.PullRequest{{Number: 1, Title: "PR 1"}}
	prs2 := []*models.PullRequest{{Number: 2, Title: "PR 2"}}

	// First call
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(prs1, nil).
		Times(1)

	result1, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, prs1, result1)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Second call - cache expired, should fetch again
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(prs2, nil).
		Times(1)

	result2, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, prs2, result2)
}

// Helper function
func stringPtrPR(s string) *string {
	return &s
}
