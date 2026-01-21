package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/redis/go-redis/v9"
)

// SaveBookmark stores a bookmark in Redis
func (s *Store) SaveBookmark(ctx context.Context, bookmark *domain.Bookmark) error {
	data, err := json.Marshal(bookmark)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmark: %w", err)
	}

	key := BookmarkKey(bookmark.ID)

	// Store bookmark data
	if err := s.client.Set(ctx, key, data, DefaultServiceTTL).Err(); err != nil {
		return fmt.Errorf("failed to save bookmark: %w", err)
	}

	// Add to set of all bookmarks
	if err := s.client.SAdd(ctx, AllBookmarksKey(), bookmark.ID).Err(); err != nil {
		return fmt.Errorf("failed to add bookmark to set: %w", err)
	}

	return nil
}

// GetBookmark retrieves a bookmark from Redis by ID
func (s *Store) GetBookmark(ctx context.Context, id string) (*domain.Bookmark, error) {
	key := BookmarkKey(id)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("bookmark not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	var bookmark domain.Bookmark
	if err := json.Unmarshal(data, &bookmark); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookmark: %w", err)
	}

	return &bookmark, nil
}

// GetAllBookmarks retrieves all bookmarks from Redis
func (s *Store) GetAllBookmarks(ctx context.Context) ([]*domain.Bookmark, error) {
	// Get all bookmark IDs
	ids, err := s.client.SMembers(ctx, AllBookmarksKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark IDs: %w", err)
	}

	if len(ids) == 0 {
		return []*domain.Bookmark{}, nil
	}

	// Retrieve all bookmarks
	bookmarks := make([]*domain.Bookmark, 0, len(ids))
	for _, id := range ids {
		bookmark, err := s.GetBookmark(ctx, id)
		if err != nil {
			// Skip bookmarks that couldn't be retrieved
			continue
		}
		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

// DeleteBookmark removes a bookmark from Redis
func (s *Store) DeleteBookmark(ctx context.Context, id string) error {
	key := BookmarkKey(id)

	// Delete bookmark data
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	// Remove from set of all bookmarks
	if err := s.client.SRem(ctx, AllBookmarksKey(), id).Err(); err != nil {
		return fmt.Errorf("failed to remove bookmark from set: %w", err)
	}

	return nil
}

// SaveBookmarksMany stores multiple bookmarks in Redis (bulk operation)
func (s *Store) SaveBookmarksMany(ctx context.Context, bookmarks []*domain.Bookmark) error {
	pipe := s.client.Pipeline()

	for _, bookmark := range bookmarks {
		data, err := json.Marshal(bookmark)
		if err != nil {
			return fmt.Errorf("failed to marshal bookmark %s: %w", bookmark.ID, err)
		}

		key := BookmarkKey(bookmark.ID)
		pipe.Set(ctx, key, data, DefaultServiceTTL)
		pipe.SAdd(ctx, AllBookmarksKey(), bookmark.ID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	return nil
}
