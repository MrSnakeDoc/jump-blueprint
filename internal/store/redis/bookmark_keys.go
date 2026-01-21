package redis

const (
	// KeyPrefixBookmark is the prefix for bookmark keys
	KeyPrefixBookmark = "jump:bookmark:"
	// KeyAllBookmarks is the key for the set of all bookmark IDs
	KeyAllBookmarks = "jump:bookmarks:all"
)

// BookmarkKey returns the Redis key for a bookmark
func BookmarkKey(id string) string {
	return KeyPrefixBookmark + id
}

// AllBookmarksKey returns the Redis key for the set of all bookmarks
func AllBookmarksKey() string {
	return KeyAllBookmarks
}
