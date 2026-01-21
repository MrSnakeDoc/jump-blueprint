package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
)

type componentStatus struct {
	OK             bool   `json:"ok"`
	ServicesLoaded *int   `json:"services_loaded,omitempty"`
	LastReload     string `json:"last_reload,omitempty"`
	Mode           string `json:"mode,omitempty"`
	Impact         string `json:"impact,omitempty"`
	Error          string `json:"error,omitempty"`
}

type infraResponse struct {
	RoutingMode string                     `json:"routing_mode"`
	Components  map[string]componentStatus `json:"components"`
}

func Infra(d deps.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get real data from memory index
		servicesCount := d.MemoryIndex.Count()
		lastReload := d.MemoryIndex.GetLastReload()
		lastReloadStr := "never"
		if !lastReload.IsZero() {
			lastReloadStr = lastReload.Format("2006-01-02 15:04:05")
		}

		// Test Redis connection
		redisStatus := checkRedis(d)

		// Build components status
		components := map[string]componentStatus{
			"homepage": {
				OK:             servicesCount > 0,
				ServicesLoaded: &servicesCount,
				LastReload:     lastReloadStr,
			},
			"redis": redisStatus,
			"resolver": {
				OK:   true,
				Mode: "fuzzy+usage-learning",
			},
		}

		response := infraResponse{
			RoutingMode: determineRoutingMode(components),
			Components:  components,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func determineRoutingMode(components map[string]componentStatus) string {
	// Check if services are loaded
	if homepage, exists := components["homepage"]; exists {
		if !homepage.OK || (homepage.ServicesLoaded != nil && *homepage.ServicesLoaded == 0) {
			return "critical" // No services loaded = critical
		}
	}

	// Check Redis - non-critical but impacts functionality
	if redis, exists := components["redis"]; exists && !redis.OK {
		return "degraded" // Redis down = degraded (no usage learning)
	}

	// All systems operational
	return "intelligent"
}

func checkRedis(d deps.Deps) componentStatus {
	if d.RedisClient == nil {
		return componentStatus{
			OK:     false,
			Mode:   "degraded",
			Impact: "usage-learning-disabled",
			Error:  "client not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := d.RedisClient.Ping(ctx).Err()
	if err != nil {
		return componentStatus{
			OK:     false,
			Mode:   "degraded",
			Impact: "usage-learning-disabled",
			Error:  "timeout",
		}
	}

	return componentStatus{
		OK:     true,
		Mode:   "optimal",
		Impact: "usage-learning-enabled",
		Error:  "none",
	}
}
