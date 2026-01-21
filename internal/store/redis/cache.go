package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheResolution stores a query -> hostname resolution in cache
func (s *Store) CacheResolution(ctx context.Context, query, hostname string, ttl time.Duration) error {
	key := CacheKey(query)
	if err := s.client.Set(ctx, key, hostname, ttl).Err(); err != nil {
		return fmt.Errorf("failed to cache resolution: %w", err)
	}
	return nil
}

// GetCachedResolution retrieves a cached resolution
func (s *Store) GetCachedResolution(ctx context.Context, query string) (string, error) {
	key := CacheKey(query)
	hostname, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil // Cache miss
		}
		return "", fmt.Errorf("failed to get cached resolution: %w", err)
	}
	return hostname, nil
}

// InvalidateCache removes a cached resolution
func (s *Store) InvalidateCache(ctx context.Context, query string) error {
	key := CacheKey(query)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}
	return nil
}

// FlushCache removes all cached resolutions
func (s *Store) FlushCache(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, KeyPrefixCache+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete cache key: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}
	return nil
}
