package mw

import (
	"net/http"

	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/MrSnakeDoc/jump/internal/utils"
)

// AllowOnlyIPs allows only specific IPs/CIDRs. If the list is empty, it does NOT filter (passthrough).
// trustProxy should be true when running behind a trusted reverse proxy/tunnel (e.g., cloudflared).
func AllowOnlyCIDRS(allowed []string, trustProxy bool, log logger.Logger) func(http.Handler) http.Handler {
	m := utils.NewIPMatcher(allowed)
	if m.IsEmpty() {
		log.Debug("AllowOnlyCIDRS: empty matcher, passthrough mode")
		return func(next http.Handler) http.Handler { return next }
	}

	log.Debugf("AllowOnlyCIDRS: initialized with %d rules, trustProxy=%v", len(allowed), trustProxy)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.ClientIP(r, trustProxy)
			log.Debugf("AllowOnlyCIDRS: checking IP=%s (RemoteAddr=%s, XFF=%s, trustProxy=%v)",
				ip, r.RemoteAddr, r.Header.Get("X-Forwarded-For"), trustProxy)

			if !m.Allow(ip) {
				log.Debugf("AllowOnlyCIDRS: IP %s REJECTED", ip)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			log.Debugf("AllowOnlyCIDRS: IP %s ALLOWED", ip)
			next.ServeHTTP(w, r)
		})
	}
}
