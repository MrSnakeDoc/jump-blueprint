package homepage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
)

// BookmarkMapper converts Homepage bookmark config to domain bookmarks
type BookmarkMapper struct{}

// NewBookmarkMapper creates a new bookmark mapper
func NewBookmarkMapper() *BookmarkMapper {
	return &BookmarkMapper{}
}

// MapBookmarks converts BookmarksConfig to domain.Bookmark slice
func (m *BookmarkMapper) MapBookmarks(config BookmarksConfig) ([]*domain.Bookmark, error) {
	bookmarks := make([]*domain.Bookmark, 0)
	now := time.Now()

	for _, category := range config {
		for categoryName, bookmarkList := range category {
			for _, bookmarkMap := range bookmarkList {
				for bookmarkName, entryList := range bookmarkMap {
					// Each bookmark has a list with a single entry
					if len(entryList) == 0 {
						continue
					}
					entry := entryList[0] // Take the first (and only) entry

					// Use Abbr if present, otherwise use bookmark name
					abbr := entry.Abbr
					if abbr == "" {
						abbr = bookmarkName
					}

					// Skip if no href
					if entry.Href == "" {
						continue
					}

					// Generate ID from URL (stable identifier)
					// Use a hash of the URL to create a short, consistent ID
					id := generateBookmarkID(entry.Href)

					bookmark := &domain.Bookmark{
						ID:        id,
						Abbr:      abbr,
						URL:       entry.Href,
						Sources:   []string{"homepage"},
						CreatedAt: now,
						UpdatedAt: now,
						Disabled:  false,
					}

					bookmarks = append(bookmarks, bookmark)
				}
			}

			// Log category for debugging
			_ = categoryName // unused but helps understand structure
		}
	}

	if len(bookmarks) == 0 {
		return nil, fmt.Errorf("no valid bookmarks found in config")
	}

	return bookmarks, nil
}

// generateBookmarkID creates a stable ID from a URL using SHA-256 hash
// This ensures that the same URL always produces the same ID,
// even if the abbr changes
func generateBookmarkID(url string) string {
	// Use SHA-256 hash of the URL
	hash := sha256.Sum256([]byte(url))
	// Take first 16 characters of hex encoding (sufficient for uniqueness)
	return hex.EncodeToString(hash[:])[:16]
}
