package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenPort      string        // ex: ":8080"
	ShutdownTimeout time.Duration // ex: 5s

	LogLevel  string // "debug" | "info" | "warn" | "error"
	PrettyLog bool   // true => zap dev (color), false => zap prod (JSON)

	ServiceFile       string        // path to the service.yaml file in homepage directory
	BookmarkFile      string        // path to the bookmarks.yaml file (optional, empty = bookmarks disabled)
	HomepageURL       string        // fallback URL when no service matches (ex: https://homepage.domain.ext)
	ReloadInterval    time.Duration // interval to reload services.yaml (default: 24h)
	GCInterval        time.Duration // interval to run garbage collection (default: 24h)
	TLSTimeout        time.Duration // timeout for TLS validation (default: 500ms)
	SkipTLSValidation bool          // skip TLS validation (useful for dev/local)
	MaxCandidates     int           // max number of candidates to validate (default: 3, 0 = no limit)
	AllowedDomains    []string      // allowed domain suffixes for redirects (derived from AllowedHosts)

	// Redis
	RedisAddr             string        // ex: "localhost:6379"
	RedisUser             string        // optional
	RedisPassword         string        // optional
	RedisPasswordRequired bool          // true => require password, false => allow empty password
	RedisDB               int           // Redis DB number
	RedisDT               time.Duration // Redis dial timeout (ex: 5s)
	RedisRT               time.Duration // Redis read timeout (ex: 3s)
	RedisWT               time.Duration // Redis write timeout (ex: 3s)
	RedisMaxWait          time.Duration // max wait between retries (ex: 10s)
	RedisPingTimeout      time.Duration // timeout for each ping attempt (ex: 5s)
	RedisPoolSize         int           // Redis connection pool size
	RedisConnectTimeout   time.Duration // Total time to retry connecting (ex: 30s)
	RedisRetryInterval    time.Duration // Initial wait between retries (ex: 2s, grows exponentially)
	RedisWarnThreshold    int           // warn after this many attempts

	AllowedHosts []string // optional, restrict access to specific Host headers
	AllowedCIDRS []string // optional, restrict access to specific IP (e.g. "1.2.3.4, 5.6.7.8")
	TrustProxy   bool     // true => trust X-Forwarded-For headers (e.g. cloudflared)
}

func Load() *Config {
	cfg := &Config{
		// Server settings
		ListenPort:      getenv("JUMP_LISTEN_PORT", ":8080"),
		ShutdownTimeout: mustDuration("JUMP_SHUTDOWN_TIMEOUT", 5*time.Second),

		// Logging
		LogLevel:  getenv("JUMP_LOG_LEVEL", "info"),
		PrettyLog: mustBool("JUMP_PRETTY_LOG", true),

		// Service file
		ServiceFile:       getenv("JUMP_SERVICE_FILE", "/app/services.yaml"),
		BookmarkFile:      getenv("JUMP_BOOKMARK_FILE", ""), // Optional, empty = bookmarks disabled
		HomepageURL:       requireEnv("JUMP_HOMEPAGE_URL"),
		ReloadInterval:    mustDuration("JUMP_RELOAD_SOURCE_INTERVAL", 24*time.Hour),
		GCInterval:        mustDuration("JUMP_GC_INTERVAL", 24*time.Hour),
		TLSTimeout:        mustDuration("JUMP_TLS_TIMEOUT", 500*time.Millisecond),
		SkipTLSValidation: mustBool("JUMP_SKIP_TLS_VALIDATION", false),
		MaxCandidates:     getenvInt("JUMP_MAX_CANDIDATES", 3),
		AllowedDomains:    extractDomains(requireEnvSlice("JUMP_ALLOWED_HOSTS")),

		// Redis settings
		RedisAddr:             requireEnv("JUMP_REDIS_ADDR"),
		RedisUser:             getenv("JUMP_REDIS_USERNAME", "default"),
		RedisPasswordRequired: mustBool("JUMP_REDIS_PASSWORD_REQUIRED", true),
		RedisPassword:         getenv("JUMP_REDIS_PASSWORD", ""),
		RedisDB:               requireEnvInt("JUMP_REDIS_DB"),
		RedisDT:               mustDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		RedisRT:               mustDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		RedisWT:               mustDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		RedisMaxWait:          mustDuration("REDIS_MAX_WAIT", 10*time.Second),
		RedisPingTimeout:      mustDuration("REDIS_PING_TIMEOUT", 5*time.Second),
		RedisPoolSize:         getenvInt("REDIS_POOL_SIZE", 10),
		RedisConnectTimeout:   mustDuration("REDIS_CONNECT_TIMEOUT", 30*time.Second),
		RedisRetryInterval:    mustDuration("REDIS_RETRY_INTERVAL", 2*time.Second),
		RedisWarnThreshold:    getenvInt("REDIS_WARN_THRESHOLD", 3),

		// Access restrictions
		AllowedHosts: requireEnvSlice("JUMP_ALLOWED_HOSTS"),
		AllowedCIDRS: parseAllowedIPs(getenv("JUMP_ALLOWED_CIDRS", "")),
		TrustProxy:   mustBool("JUMP_TRUST_PROXY", true),
	}

	// Validate Redis password configuration
	if cfg.RedisPasswordRequired && cfg.RedisPassword == "" {
		panic("❌ FATAL: JUMP_REDIS_PASSWORD is required when JUMP_REDIS_PASSWORD_REQUIRED=true")
	}

	// Log config only in debug mode with redacted sensitive fields
	if cfg.LogLevel == "debug" {
		cfgCopy := *cfg
		cfgCopy.RedisPassword = "***REDACTED***"
		if cfg.RedisUser != "" {
			cfgCopy.RedisUser = "***REDACTED***"
		}
		log.Printf("[DEBUG] cfg: %+v\n", cfgCopy)
	}

	return cfg
}

// helpers
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("❌ FATAL: Required environment variable %s is not set", key))
	}
	return v
}

func requireEnvInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("❌ FATAL: Required environment variable %s is not set", key))
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("❌ FATAL: Invalid integer value for %s: %s", key, v))
	}
	return i
}

func requireEnvSlice(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("❌ FATAL: Required environment variable %s is not set", key))
	}
	return splitAndTrim(v)
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func mustBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func mustDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func parseAllowedIPs(allowed string) []string {
	if allowed == "" {
		return nil
	}
	ips := make([]string, 0, 4)
	for _, ip := range splitAndTrim(allowed) {
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	raw := strings.Split(s, ",")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		// Remove surrounding quotes if present
		trimmed = strings.Trim(trimmed, `"'`)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// extractDomains extracts domain suffixes from allowed hosts for redirect validation.
// Examples: "jump.domain.ext" -> ["domain.ext", "jump.domain.ext"]
//
//	"10.70.80.2:8080" -> ["10.70.80.2:8080"] (IP addresses kept as-is)
func extractDomains(hosts []string) []string {
	if len(hosts) == 0 {
		return nil
	}

	domains := make([]string, 0, len(hosts)*2)
	seen := make(map[string]bool)

	for _, host := range hosts {
		// Remove port if present
		hostWithoutPort := host
		if idx := strings.LastIndex(host, ":"); idx != -1 {
			// Check if it's actually a port (not IPv6)
			if !strings.Contains(host[:idx], "]:") {
				hostWithoutPort = host[:idx]
			}
		}

		// Add the full host
		if !seen[hostWithoutPort] {
			domains = append(domains, hostWithoutPort)
			seen[hostWithoutPort] = true
		}

		// Extract domain suffix (everything after first dot)
		parts := strings.Split(hostWithoutPort, ".")
		if len(parts) >= 2 {
			// Add domain suffix (e.g., "domain.ext")
			domainSuffix := strings.Join(parts[1:], ".")
			if !seen[domainSuffix] {
				domains = append(domains, domainSuffix)
				seen[domainSuffix] = true
			}
		}
	}

	return domains
}
