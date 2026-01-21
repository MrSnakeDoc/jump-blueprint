package domain

import "time"

// Bookmark represents an external bookmark entry.
// Bookmarks are external URLs (not part of the managed services)
// that can be quickly accessed via the @ prefix.
type Bookmark struct {
	// ─────────────────────────────
	// Identity (immutable)
	// ─────────────────────────────

	// ID is the canonical unique identifier.
	// Derived from Abbr (lowercased and sanitized).
	ID string

	// Abbr is the short abbreviation used for matching.
	// Example: "ChatGPT", "Docker Hub"
	Abbr string

	// URL is the full external URL to redirect to.
	// Example: https://chat.openai.com/
	URL string

	// ─────────────────────────────
	// Provenance & observation
	// ─────────────────────────────

	// Sources indicates where this bookmark was discovered from.
	// Example: homepage
	Sources []string

	// ─────────────────────────────
	// Metadata
	// ─────────────────────────────

	// CreatedAt is the first time the bookmark was discovered.
	CreatedAt time.Time

	// UpdatedAt is updated on any mutation.
	UpdatedAt time.Time

	// ─────────────────────────────
	// Liveness & cleanup
	// ─────────────────────────────

	// Disabled marks a bookmark as soft-deleted.
	// It may be garbage-collected later.
	Disabled bool
}
