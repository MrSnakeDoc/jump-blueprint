package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
)

func Search(d deps.Deps) http.HandlerFunc {
	store := redisstore.NewStore(d.RedisClient)
	memIndex := d.MemoryIndex

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := strings.TrimSpace(r.URL.Query().Get("q"))

		// Empty query -> redirect to homepage
		if query == "" {
			d.Logger.Debug("empty query, redirecting to homepage")
			http.Redirect(w, r, d.HomepageURL, http.StatusFound)
			return
		}

		d.Logger.Info("search request",
			logger.String("query", query))

		// Special case: bookmarks (queries starting with @)
		if strings.HasPrefix(query, "@") {
			handleBookmarkSearch(w, r, query, d, memIndex)
			return
		}

		// Special case: internal endpoints (queries starting with /)
		if strings.HasPrefix(query, "/") {
			handleInternalEndpoint(w, r, query, d)
			return
		}

		// Try cache first
		if handleCachedService(w, r, ctx, query, store, memIndex, d) {
			return
		}

		// Search and validate services
		handleServiceSearch(w, r, ctx, query, store, memIndex, d)
	}
}

// handleInternalEndpoint handles internal endpoint routing
func handleInternalEndpoint(w http.ResponseWriter, r *http.Request, query string, d deps.Deps) {
	if endpoint := matchInternalEndpoint(query); endpoint != "" {
		internalURL := fmt.Sprintf("https://%s%s", r.Host, endpoint)
		d.Logger.Info("internal endpoint redirect",
			logger.String("query", query),
			logger.String("endpoint", endpoint))
		http.Redirect(w, r, internalURL, http.StatusFound)
		return
	}
	// No match found, redirect to homepage
	d.Logger.Debug("no internal endpoint matched",
		logger.String("query", query))
	http.Redirect(w, r, d.HomepageURL, http.StatusFound)
}

// handleCachedService checks cache and redirects if valid, returns true if handled
func handleCachedService(w http.ResponseWriter, r *http.Request, ctx context.Context, query string, store *redisstore.Store, memIndex *index.MemoryIndex, d deps.Deps) bool {
	cachedHostname, err := store.GetCachedResolution(ctx, query)
	if err != nil || cachedHostname == "" {
		return false
	}

	// Validate cached service is still alive
	if err := domain.ValidateTLS(cachedHostname, d.TLSTimeout); err == nil {
		d.Logger.Info("cache hit, redirecting",
			logger.String("query", query),
			logger.String("hostname", cachedHostname))

		// Increment usage counter (best effort)
		_ = store.IncrementUsage(ctx, cachedHostname)
		memIndex.IncrementCounter(cachedHostname)

		redirectURL := fmt.Sprintf("https://%s", cachedHostname)
		if !isAllowedRedirect(cachedHostname, d.AllowedDomains) {
			d.Logger.Warn("cached hostname not in allowed domains",
				logger.String("hostname", cachedHostname))
			http.Redirect(w, r, d.HomepageURL, http.StatusFound)
			return true
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
		return true
	}

	// Cache hit but service is down, invalidate cache
	d.Logger.Debug("cached service is down, invalidating cache",
		logger.String("hostname", cachedHostname))
	_ = store.InvalidateCache(ctx, query)
	return false
}

// handleServiceSearch searches, validates and redirects to a service
func handleServiceSearch(w http.ResponseWriter, r *http.Request, ctx context.Context, query string, store *redisstore.Store, memIndex *index.MemoryIndex, d deps.Deps) {
	// Parse query
	parsedQuery := domain.ParseQuery(query)

	// Get all services from memory index
	services := memIndex.GetAllServices()
	if len(services) == 0 {
		d.Logger.Warn("no services available in index")
		http.Redirect(w, r, d.HomepageURL, http.StatusFound)
		return
	}

	// Rank candidates
	candidates := domain.RankCandidates(parsedQuery, services)
	if len(candidates) == 0 {
		d.Logger.Info("no matching services found",
			logger.String("query", query))
		http.Redirect(w, r, d.HomepageURL, http.StatusFound)
		return
	}

	// Limit candidates to MaxCandidates (top N only)
	if d.MaxCandidates > 0 && len(candidates) > d.MaxCandidates {
		d.Logger.Debug("limiting candidates",
			logger.Int("total", len(candidates)),
			logger.Int("max", d.MaxCandidates))
		candidates = candidates[:d.MaxCandidates]
	}

	// Validate candidates in order and redirect to first healthy one
	for i, candidate := range candidates {
		hostname := candidate.Service.Hostname

		// Check if redirect is allowed
		if !isAllowedRedirect(hostname, d.AllowedDomains) {
			d.Logger.Debug("skipping service not in allowed domains",
				logger.String("hostname", hostname))
			continue
		}

		// Skip TLS validation if configured
		if !d.SkipTLSValidation {
			// Validate TLS
			if err := domain.ValidateTLS(hostname, d.TLSTimeout); err != nil {
				d.Logger.Debug("service validation failed",
					logger.String("hostname", hostname),
					logger.Error(err))
				continue
			}
		} else {
			d.Logger.Debug("skipping TLS validation (disabled in config)",
				logger.String("hostname", hostname),
				logger.String("score", fmt.Sprintf("%.2f", candidate.TotalScore)),
				logger.Int("rank", i+1))
		}

		// Found a healthy service!
		d.Logger.Info("resolved and validated service",
			logger.String("query", query),
			logger.String("hostname", hostname),
			logger.String("score", fmt.Sprintf("%.2f", candidate.TotalScore)))

		// Increment usage counter (best effort)
		_ = store.IncrementUsage(ctx, hostname)
		memIndex.IncrementCounter(hostname)

		// Cache the resolution
		_ = store.CacheResolution(ctx, query, hostname, redisstore.DefaultCacheTTL)

		// Redirect
		redirectURL := fmt.Sprintf("https://%s", hostname)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// No healthy service found
	d.Logger.Warn("no healthy service found for query",
		logger.String("query", query))
	http.Redirect(w, r, d.HomepageURL, http.StatusFound)
}

// handleBookmarkSearch handles bookmark searches (queries starting with @)
func handleBookmarkSearch(w http.ResponseWriter, r *http.Request, query string, d deps.Deps, memIndex *index.MemoryIndex) {
	// Remove @ prefix
	queryStr := strings.TrimPrefix(query, "@")
	queryStr = strings.TrimSpace(queryStr)

	// Empty query after @ -> redirect to homepage
	if queryStr == "" {
		d.Logger.Debug("empty bookmark query, redirecting to homepage")
		http.Redirect(w, r, d.HomepageURL, http.StatusFound)
		return
	}

	// Get all bookmarks from memory index
	bookmarks := memIndex.GetAllBookmarks()
	if len(bookmarks) == 0 {
		d.Logger.Warn("no bookmarks available in index")
		http.Redirect(w, r, d.HomepageURL, http.StatusFound)
		return
	}

	// Rank bookmark candidates
	candidates := domain.RankBookmarkCandidates(queryStr, bookmarks)
	if len(candidates) == 0 {
		d.Logger.Info("no matching bookmarks found",
			logger.String("query", queryStr))
		http.Redirect(w, r, d.HomepageURL, http.StatusFound)
		return
	}

	// Get best bookmark (no TLS validation for external URLs)
	bestBookmark := candidates[0].Bookmark

	d.Logger.Info("resolved bookmark",
		logger.String("query", queryStr),
		logger.String("abbr", bestBookmark.Abbr),
		logger.String("url", bestBookmark.URL),
		logger.String("score", fmt.Sprintf("%.2f", candidates[0].Score)))

	// Redirect to bookmark URL
	http.Redirect(w, r, bestBookmark.URL, http.StatusFound)
}

// isAllowedRedirect checks if a hostname is allowed for redirection
func isAllowedRedirect(hostname string, allowedDomains []string) bool {
	hostname = strings.ToLower(hostname)

	for _, domain := range allowedDomains {
		domain = strings.ToLower(domain)

		// Exact match
		if hostname == domain {
			return true
		}

		// Subdomain match (hostname ends with .domain)
		if strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}

	return false
}

// matchInternalEndpoint performs fuzzy matching on internal endpoints
// Returns the full endpoint path if a unique match is found, empty string otherwise
func matchInternalEndpoint(query string) string {
	// Available internal endpoints
	endpoints := []string{
		"/infra",
		"/healthz",
		"/readyz",
	}

	query = strings.ToLower(query)
	var matches []string

	// Find all endpoints that start with the query
	for _, endpoint := range endpoints {
		if strings.HasPrefix(endpoint, query) {
			matches = append(matches, endpoint)
		}
	}

	// Return the match only if exactly one endpoint matches
	if len(matches) == 1 {
		return matches[0]
	}

	// Multiple or no matches - return empty
	return ""
}
