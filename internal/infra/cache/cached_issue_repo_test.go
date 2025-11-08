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

func TestCachedIssueRepository_List_CacheHit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.IssueOptions{
		State: models.IssueStateOpen,
	}

	expectedIssues := []*models.Issue{
		{Number: 1, Title: "Test Issue 1"},
		{Number: 2, Title: "Test Issue 2"},
	}

	// First call - should hit the mock
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(expectedIssues, nil).
		Times(1)

	issues1, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, expectedIssues, issues1)

	// Second call - should hit cache, mock should not be called again
	issues2, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, expectedIssues, issues2)
}

func TestCachedIssueRepository_List_SkipCache(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.IssueOptions{
		State: models.IssueStateOpen,
	}

	issues1 := []*models.Issue{{Number: 1, Title: "Test Issue 1"}}
	issues2 := []*models.Issue{
		{Number: 1, Title: "Updated Issue 1"},
		{Number: 2, Title: "New Issue 2"},
	}

	// First call to populate cache
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(issues1, nil).
		Times(1)

	_, err = cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)

	// Second call with SkipCache context - should bypass cache
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(issues2, nil).
		Times(1)

	ctx := cache.WithSkipCacheContext(context.Background())
	issues, err := cachedRepo.List(ctx, owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, issues2, issues)
}

func TestCachedIssueRepository_List_Error(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.IssueOptions{
		State: models.IssueStateOpen,
	}

	expectedErr := errors.New("API error")

	// Mock should return error
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(nil, expectedErr).
		Times(1)

	issues, err := cachedRepo.List(context.Background(), owner, repo, opts)
	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Equal(t, expectedErr, err)
}

func TestCachedIssueRepository_Get_CacheHit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123

	expectedIssue := &models.Issue{
		Number: number,
		Title:  "Test Issue",
	}

	// First call - should hit the mock
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(expectedIssue, nil).
		Times(1)

	issue1, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedIssue, issue1)

	// Second call - should hit cache
	issue2, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, expectedIssue, issue2)
}

func TestCachedIssueRepository_Update_InvalidatesCache(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().DisableFileCache()
	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	number := 123

	originalIssue := &models.Issue{
		Number: number,
		Title:  "Original Title",
	}

	// Populate cache
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(originalIssue, nil).
		Times(1)

	_, err = cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)

	// Update issue
	updateInput := &models.UpdateIssueInput{
		Title: stringPtr("Updated Title"),
	}
	updatedIssue := &models.Issue{
		Number: number,
		Title:  "Updated Title",
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), owner, repo, number, updateInput).
		Return(updatedIssue, nil).
		Times(1)

	_, err = cachedRepo.Update(context.Background(), owner, repo, number, updateInput)
	require.NoError(t, err)

	// Get again - should fetch fresh data (cache was invalidated)
	mockRepo.EXPECT().
		Get(gomock.Any(), owner, repo, number).
		Return(updatedIssue, nil).
		Times(1)

	issue, err := cachedRepo.Get(context.Background(), owner, repo, number)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", issue.Title)
}

func TestCachedIssueRepository_CustomTTL(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIssueRepository(ctrl)
	cacheConfig := cache.DefaultConfig().
		DisableFileCache().
		WithMemoryTTL(100 * time.Millisecond) // Very short TTL for testing

	cacheService, err := cache.NewCacheWithConfig(cacheConfig)
	require.NoError(t, err)
	c := cacheService.(*cache.Cache)

	cachedRepo := cache.NewCachedIssueRepository(mockRepo, c)

	owner := "testowner"
	repo := "testrepo"
	opts := &models.IssueOptions{
		State: models.IssueStateOpen,
	}

	issues1 := []*models.Issue{{Number: 1, Title: "Issue 1"}}
	issues2 := []*models.Issue{{Number: 2, Title: "Issue 2"}}

	// First call
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(issues1, nil).
		Times(1)

	result1, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, issues1, result1)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Second call - cache expired, should fetch again
	mockRepo.EXPECT().
		List(gomock.Any(), owner, repo, opts).
		Return(issues2, nil).
		Times(1)

	result2, err := cachedRepo.List(context.Background(), owner, repo, opts)
	require.NoError(t, err)
	assert.Equal(t, issues2, result2)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
