package domain

import "time"

// Service represents the canonical runtime truth of a routable service.
//
// It is NOT tied to Homepage, Redis or any external source.
// All inputs (files, cache, learning) are merged into this structure.
//
// A Service is considered uniquely identified by its Hostname.
type Service struct {
	// ─────────────────────────────
	// Identity (immutable)
	// ─────────────────────────────

	// ID is the canonical unique identifier.
	// It MUST be equal to Hostname.
	ID string

	// Hostname is the DNS hostname of the service.
	// Example: jellyfin.domain.ext
	Hostname string

	// ─────────────────────────────
	// Functional description
	// (may be overwritten by homepage reload)
	// ─────────────────────────────

	// Name is derived from the first DNS label.
	// Example: jellyfin
	Name string

	// ─────────────────────────────
	// Provenance & observation
	// ─────────────────────────────

	// Sources indicates where this service was discovered from.
	// Example: homepage, redis
	Sources []string

	// LastSeenAt is updated whenever the service is observed
	// from any source or validation process.
	LastSeenAt time.Time

	// ─────────────────────────────
	// Learning & persistence
	// ─────────────────────────────

	// Counter represents the number of successful redirects.
	Counter int64

	// CreatedAt is the first time the service was discovered.
	CreatedAt time.Time

	// UpdatedAt is updated on any mutation.
	UpdatedAt time.Time

	// LastUsedAt is updated only after a successful redirect.
	LastUsedAt time.Time

	// ─────────────────────────────
	// Liveness & cleanup
	// ─────────────────────────────

	// Disabled marks a service as soft-deleted.
	// It may be garbage-collected later.
	Disabled bool
}
