package deps

import (
	"time"

	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	Logger                logger.Logger
	StartTime             time.Time
	Version               string
	Commit                string
	BuildDate             string
	GoVersion             string
	TimeNow               func() time.Time   // for testing, defaults to time.Now
	AllowedHosts          []string           // Host headers allowed to access the server
	AllowedCIDRS          []string           // IPs allowed to access healthz/readyz endpoints
	TrustProxy            bool               // true if running behind a trusted reverse proxy (e.g., cloudflared)
	ServiceFile           string             // Path to the service definitions file
	RedisClient           *redis.Client      // Redis client connection
	MemoryIndex           *index.MemoryIndex // In-memory service index
	HomepageURL           string             // Fallback URL when no service matches
	TLSTimeout            time.Duration      // Timeout for TLS validation
	SkipTLSValidation     bool               // Skip TLS validation (useful for dev/local)
	MaxCandidates         int                // Max number of candidates to validate
	AllowedDomains        []string           // Allowed domain suffixes for redirects
	ReloadTrigger         chan struct{}      // Channel to trigger manual service reload
	BookmarkReloadTrigger chan struct{}      // Channel to trigger manual bookmark reload (nil if bookmarks disabled)
	// Add more shared deps later (Store, Version, etc.)
}
