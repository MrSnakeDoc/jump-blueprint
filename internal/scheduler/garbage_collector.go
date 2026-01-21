package scheduler

import (
	"context"
	"time"

	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
)

const (
	// DefaultGCThreshold is the duration after which disabled services are deleted
	DefaultGCThreshold = 30 * 24 * time.Hour // 30 days
)

// GarbageCollector handles cleanup of old disabled services
type GarbageCollector struct {
	store     *redisstore.Store
	index     *index.MemoryIndex
	logger    logger.Logger
	interval  time.Duration
	threshold time.Duration
	stopCh    chan struct{}
}

// NewGarbageCollector creates a new garbage collector
func NewGarbageCollector(
	store *redisstore.Store,
	idx *index.MemoryIndex,
	log logger.Logger,
	interval time.Duration,
	threshold time.Duration,
) *GarbageCollector {
	if threshold == 0 {
		threshold = DefaultGCThreshold
	}

	return &GarbageCollector{
		store:     store,
		index:     idx,
		logger:    log,
		interval:  interval,
		threshold: threshold,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the periodic garbage collection process
func (gc *GarbageCollector) Start(ctx context.Context) error {
	// Run immediately on start
	if err := gc.Collect(ctx); err != nil {
		gc.logger.Warn("initial garbage collection failed",
			logger.Error(err))
	}

	// Start periodic collection
	ticker := time.NewTicker(gc.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := gc.Collect(ctx); err != nil {
					gc.logger.Error("garbage collection failed",
						logger.Error(err))
				}
			case <-gc.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop stops the garbage collector
func (gc *GarbageCollector) Stop() {
	close(gc.stopCh)
}

// Collect removes services and bookmarks that have been disabled for longer than the threshold
func (gc *GarbageCollector) Collect(ctx context.Context) error {
	gc.logger.Info("running garbage collection for disabled services and bookmarks")

	now := time.Now()

	// Collect disabled services
	servicesDeleted := gc.collectServices(ctx, now)

	// Collect disabled bookmarks
	bookmarksDeleted := gc.collectBookmarks(ctx, now)

	totalDeleted := servicesDeleted + bookmarksDeleted

	if totalDeleted > 0 {
		gc.logger.Info("garbage collection completed",
			logger.Int("services_deleted", servicesDeleted),
			logger.Int("bookmarks_deleted", bookmarksDeleted),
			logger.Int("total_deleted", totalDeleted))
	} else {
		gc.logger.Debug("no items to garbage collect")
	}

	return nil
}

// collectServices removes disabled services
func (gc *GarbageCollector) collectServices(ctx context.Context, now time.Time) int {
	services := gc.index.GetAllServices()
	deletedCount := 0

	for _, service := range services {
		// Only collect disabled services
		if !service.Disabled {
			continue
		}

		// Check if service has been disabled long enough
		if service.UpdatedAt.IsZero() {
			continue
		}

		disabledDuration := now.Sub(service.UpdatedAt)
		if disabledDuration < gc.threshold {
			continue
		}

		// Delete from memory index
		gc.index.DeleteService(service.ID)

		// Delete from Redis store (best effort)
		if gc.store != nil {
			if err := gc.store.DeleteService(ctx, service.ID); err != nil {
				gc.logger.Warn("failed to delete service from redis",
					logger.String("service_id", service.ID),
					logger.Error(err))
			}
		}

		gc.logger.Info("garbage collected disabled service",
			logger.String("service_id", service.ID),
			logger.String("hostname", service.Hostname),
			logger.String("disabled_for", disabledDuration.String()))

		deletedCount++
	}

	return deletedCount
}

// collectBookmarks removes disabled bookmarks
func (gc *GarbageCollector) collectBookmarks(ctx context.Context, now time.Time) int {
	bookmarks := gc.index.GetAllBookmarks()
	deletedCount := 0

	for _, bookmark := range bookmarks {
		// Only collect disabled bookmarks
		if !bookmark.Disabled {
			continue
		}

		// Check if bookmark has been disabled long enough
		if bookmark.UpdatedAt.IsZero() {
			continue
		}

		disabledDuration := now.Sub(bookmark.UpdatedAt)
		if disabledDuration < gc.threshold {
			continue
		}

		// Delete from memory index
		gc.index.DeleteBookmark(bookmark.ID)

		// Delete from Redis store (best effort)
		if gc.store != nil {
			if err := gc.store.DeleteBookmark(ctx, bookmark.ID); err != nil {
				gc.logger.Warn("failed to delete bookmark from redis",
					logger.String("bookmark_id", bookmark.ID),
					logger.Error(err))
			}
		}

		gc.logger.Info("garbage collected disabled bookmark",
			logger.String("bookmark_id", bookmark.ID),
			logger.String("abbr", bookmark.Abbr),
			logger.String("disabled_for", disabledDuration.String()))

		deletedCount++
	}

	return deletedCount
}
