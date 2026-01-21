package mw

import (
	"net/http"
	"strings"

	"github.com/MrSnakeDoc/jump/internal/logger"
)

// EnforceHost allows requests only if r.Host matches one of the allowed hosts.
// Supports wildcard patterns like "*.example.com".
// If allowedHosts is empty, it acts as a passthrough.
func EnforceHost(allowedHosts []string, log logger.Logger) func(http.Handler) http.Handler {
	if len(allowedHosts) == 0 {
		log.Debug("EnforceHost: empty allowedHosts, passthrough mode")
		return func(next http.Handler) http.Handler { return next }
	}

	log.Debugf("EnforceHost: initialized with hosts=%v", allowedHosts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			log.Debugf("EnforceHost: checking Host=%s", host)

			// Check exact matches and wildcard patterns
			for _, pattern := range allowedHosts {
				if matchHost(host, pattern) {
					log.Debugf("EnforceHost: Host %s ALLOWED (matched %s)", host, pattern)
					next.ServeHTTP(w, r)
					return
				}
			}

			log.Debugf("EnforceHost: Host %s REJECTED", host)
			w.WriteHeader(http.StatusForbidden)
		})
	}
}

// matchHost checks if host matches pattern (supports wildcard *.example.com)
func matchHost(host, pattern string) bool {
	// Exact match
	if host == pattern {
		return true
	}

	// Wildcard match: *.example.com matches sub.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // Remove * to get .example.com
		return strings.HasSuffix(host, suffix)
	}

	return false
}
