package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/MrSnakeDoc/jump/internal/sources/homepage"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
)

// BookmarkReloader handles periodic reloading of homepage bookmarks
type BookmarkReloader struct {
	loader        *homepage.BookmarkLoader
	mapper        *homepage.BookmarkMapper
	store         *redisstore.Store
	index         *index.MemoryIndex
	logger        logger.Logger
	interval      time.Duration
	stopCh        chan struct{}
	manualTrigger chan struct{}
}

// NewBookmarkReloader creates a new bookmark reloader
func NewBookmarkReloader(
	bookmarkFile string,
	store *redisstore.Store,
	idx *index.MemoryIndex,
	log logger.Logger,
	interval time.Duration,
	manualTrigger chan struct{},
) *BookmarkReloader {
	return &BookmarkReloader{
		loader:        homepage.NewBookmarkLoader(bookmarkFile),
		mapper:        homepage.NewBookmarkMapper(),
		store:         store,
		index:         idx,
		logger:        log,
		interval:      interval,
		stopCh:        make(chan struct{}),
		manualTrigger: manualTrigger,
	}
}

// Start begins the periodic reload process
func (br *BookmarkReloader) Start(ctx context.Context) error {
	// Load immediately on start
	if err := br.Reload(ctx); err != nil {
		return fmt.Errorf("initial bookmark reload failed: %w", err)
	}

	// Start periodic reload
	ticker := time.NewTicker(br.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := br.Reload(ctx); err != nil {
					br.logger.Error("failed to reload bookmarks",
						logger.Error(err))
				}
			case <-br.manualTrigger:
				br.logger.Info("manual bookmark reload triggered")
				if err := br.Reload(ctx); err != nil {
					br.logger.Error("failed to reload bookmarks",
						logger.Error(err))
				}
			case <-br.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop stops the reloader
func (br *BookmarkReloader) Stop() {
	close(br.stopCh)
}

// Reload loads bookmarks from homepage and updates store + index
func (br *BookmarkReloader) Reload(ctx context.Context) error {
	br.logger.Info("reloading bookmarks from homepage")

	// Load and parse bookmarks.yaml
	config, err := br.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	// Map to domain bookmarks
	newBookmarks, err := br.mapper.MapBookmarks(config)
	if err != nil {
		return fmt.Errorf("failed to map bookmarks: %w", err)
	}

	br.logger.Info("loaded bookmarks from homepage",
		logger.Int("count", len(newBookmarks)))

	// Get existing bookmarks from homepage source to detect removals
	existingBookmarks := br.getHomepageBookmarks()

	// Build map of new bookmark IDs for quick lookup
	newBookmarkIDs := make(map[string]bool, len(newBookmarks))
	for _, bm := range newBookmarks {
		newBookmarkIDs[bm.ID] = true
	}

	// Find bookmarks that were removed from homepage
	var disabledBookmarks []*domain.Bookmark
	for _, existing := range existingBookmarks {
		if !newBookmarkIDs[existing.ID] {
			// Bookmark no longer in homepage - mark as disabled
			existing.Disabled = true
			existing.UpdatedAt = time.Now()
			disabledBookmarks = append(disabledBookmarks, existing)
		}
	}

	if len(disabledBookmarks) > 0 {
		br.logger.Info("marking removed bookmarks as disabled",
			logger.Int("count", len(disabledBookmarks)))
	}

	// Combine active and disabled bookmarks for storage
	newBookmarks = append(newBookmarks, disabledBookmarks...)

	// Update memory index
	br.index.UpdateBookmarks(newBookmarks)

	// Update Redis store (best effort)
	if br.store != nil {
		if err := br.store.SaveBookmarksMany(ctx, newBookmarks); err != nil {
			br.logger.Warn("failed to save bookmarks to redis",
				logger.Error(err))
			// Don't fail - memory index is the primary source
		} else {
			br.logger.Info("bookmarks saved to redis")
		}
	}

	return nil
}

// getHomepageBookmarks returns existing bookmarks that came from homepage source
func (br *BookmarkReloader) getHomepageBookmarks() []*domain.Bookmark {
	all := br.index.GetAllBookmarks()
	var homepageBookmarks []*domain.Bookmark

	for _, bm := range all {
		// Check if bookmark has homepage in its sources
		for _, source := range bm.Sources {
			if source == "homepage" {
				homepageBookmarks = append(homepageBookmarks, bm)
				break
			}
		}
	}

	return homepageBookmarks
}
