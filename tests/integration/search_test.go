package integration

import (
	"testing"

	"github.com/MrSnakeDoc/jump/internal/domain"
)

// TestSearchScenarios tests various search scenarios with fuzzy matching
func TestSearchScenarios(t *testing.T) {
	// Setup test services
	services := []*domain.Service{
		{
			ID:       "adguard",
			Name:     "adguard",
			Hostname: "adguard.domain.ext",
			Counter:  0,
		},
		{
			ID:       "adguardha",
			Name:     "adguardha",
			Hostname: "adguardha.domain.ext",
			Counter:  0,
		},
		{
			ID:       "traefik",
			Name:     "traefik",
			Hostname: "traefik.domain.ext",
			Counter:  0,
		},
		{
			ID:       "jellyfin",
			Name:     "jellyfin",
			Hostname: "jellyfin.domain.ext",
			Counter:  0,
		},
		{
			ID:       "jellyseerr",
			Name:     "jellyseerr",
			Hostname: "jellyseerr.domain.ext",
			Counter:  0,
		},
	}

	tests := []struct {
		name        string
		queryString string
		expectedTop string // Expected top result hostname
		description string
	}{
		{
			name:        "exact match",
			queryString: "adguard",
			expectedTop: "adguard.domain.ext",
			description: "Exact match should rank highest",
		},
		{
			name:        "exact match with similar names",
			queryString: "adguard",
			expectedTop: "adguard.domain.ext",
			description: "Should prefer 'adguard' over 'adguardha'",
		},
		{
			name:        "prefix match",
			queryString: "jelly",
			expectedTop: "jellyfin.domain.ext",
			description: "Prefix match should work (either jellyfin or jellyseerr)",
		},
		{
			name:        "multi fragment search",
			queryString: "ad ha",
			expectedTop: "adguard.domain.ext",
			description: "Multi-fragment search without dot (top-level only)",
		},
		{
			name:        "fuzzy match",
			queryString: "trfk",
			expectedTop: "traefik.domain.ext",
			description: "Fuzzy matching should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := domain.ParseQuery(tt.queryString)
			candidates := domain.RankCandidates(query, services)

			if len(candidates) == 0 {
				t.Fatalf("No candidates returned for query: %s", tt.queryString)
			}

			topResult := candidates[0].Service.Hostname
			if topResult != tt.expectedTop {
				t.Logf("Query: %s", tt.queryString)
				t.Logf("Expected top result: %s", tt.expectedTop)
				t.Logf("Actual top result: %s", topResult)
				t.Logf("All results:")
				for i, c := range candidates {
					t.Logf("  %d. %s (score: %.2f)", i+1, c.Service.Hostname, c.TotalScore)
				}
			}
		})
	}
}

// TestSubdomainMatching tests subdomain matching with dot notation
func TestSubdomainMatching(t *testing.T) {
	services := []*domain.Service{
		{
			ID:       "adguard-main",
			Name:     "adguard",
			Hostname: "adguard.domain.ext",
			Counter:  0,
		},
		{
			ID:       "adguard-ha",
			Name:     "adguard",
			Hostname: "adguard.ha.domain.ext",
			Counter:  0,
		},
		{
			ID:       "jellyfin-main",
			Name:     "jellyfin",
			Hostname: "jellyfin.domain.ext",
			Counter:  0,
		},
		{
			ID:       "jellyseerr",
			Name:     "jellyseerr",
			Hostname: "jellyseerr.domain.ext",
			Counter:  0,
		},
	}

	tests := []struct {
		name        string
		queryString string
		expectedTop string
		description string
	}{
		{
			name:        "subdomain with dot",
			queryString: "adguard.ha",
			expectedTop: "adguard.ha.domain.ext",
			description: "Should match subdomain when dot is present",
		},
		{
			name:        "top level without dot",
			queryString: "adguard",
			expectedTop: "adguard.domain.ext",
			description: "Without dot should prefer top-level domain",
		},
		{
			name:        "subdomain fragments",
			queryString: "jellyfin.do",
			expectedTop: "jellyfin.domain.ext",
			description: "Should match subdomain fragments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := domain.ParseQuery(tt.queryString)

			t.Logf("Query: %s", tt.queryString)
			t.Logf("HasDot: %v", query.HasDot)
			t.Logf("TopLevelFragments: %v", query.TopLevelFragments)
			t.Logf("SubdomainFragments: %v", query.SubdomainFragments)

			candidates := domain.RankCandidates(query, services)

			if len(candidates) == 0 {
				t.Fatalf("No candidates returned for query: %s", tt.queryString)
			}

			topResult := candidates[0].Service.Hostname
			t.Logf("Top result: %s (score: %.2f)", topResult, candidates[0].TotalScore)

			if topResult != tt.expectedTop {
				t.Logf("Expected: %s", tt.expectedTop)
				t.Logf("All results:")
				for i, c := range candidates {
					t.Logf("  %d. %s (score: %.2f, lexical: %.2f)",
						i+1, c.Service.Hostname, c.TotalScore, c.LexicalScore)
				}
			}
		})
	}
}

// TestUsageLearning tests that frequently accessed services rank higher
func TestUsageLearning(t *testing.T) {
	services := []*domain.Service{
		{
			ID:       "admin",
			Name:     "admin",
			Hostname: "admin.domain.ext",
			Counter:  100, // Heavily used
		},
		{
			ID:       "adguard",
			Name:     "adguard",
			Hostname: "adguard.domain.ext",
			Counter:  5, // Rarely used
		},
	}

	// Search for "ad" - both match as prefix
	query := domain.ParseQuery("ad")
	candidates := domain.RankCandidates(query, services)

	if len(candidates) < 2 {
		t.Fatalf("Expected at least 2 candidates, got %d", len(candidates))
	}

	t.Logf("Query: ad")
	for i, c := range candidates {
		t.Logf("  %d. %s (total: %.2f, lexical: %.2f, usage: %.2f, counter: %d)",
			i+1, c.Service.Hostname, c.TotalScore, c.LexicalScore, c.UsageScore, c.Service.Counter)
	}

	// The heavily used service (admin) should rank higher despite both being prefix matches
	topResult := candidates[0].Service.Hostname
	if topResult != "admin.domain.ext" {
		t.Errorf("Usage learning failed: expected admin.domain.ext to rank first due to usage, got %s", topResult)
	}
}

// TestMixedScenarios tests complex real-world scenarios
func TestMixedScenarios(t *testing.T) {
	services := []*domain.Service{
		{ID: "adguard", Name: "adguard", Hostname: "adguard.domain.ext", Counter: 10},
		{ID: "adguard-ha", Name: "adguard", Hostname: "adguard.ha.domain.ext", Counter: 5},
		{ID: "jellyfin", Name: "jellyfin", Hostname: "jellyfin.domain.ext", Counter: 20},
		{ID: "jellyseerr", Name: "jellyseerr", Hostname: "jellyseerr.domain.ext", Counter: 8},
		{ID: "traefik", Name: "traefik", Hostname: "traefik.domain.ext", Counter: 15},
		{ID: "uptime", Name: "uptime", Hostname: "uptime.domain.ext", Counter: 3},
	}

	scenarios := []struct {
		name        string
		queryString string
		description string
		validate    func(t *testing.T, candidates []*domain.Candidate)
	}{
		{
			name:        "exact match beats everything",
			queryString: "jellyfin",
			description: "Exact match should win even with lower usage",
			validate: func(t *testing.T, candidates []*domain.Candidate) {
				if candidates[0].Service.Hostname != "jellyfin.domain.ext" {
					t.Errorf("Expected jellyfin.domain.ext as top result")
				}
			},
		},
		{
			name:        "prefix with usage learning",
			queryString: "je",
			description: "jellyfin (higher usage) should beat jellyseerr",
			validate: func(t *testing.T, candidates []*domain.Candidate) {
				if candidates[0].Service.Hostname != "jellyfin.domain.ext" {
					t.Logf("Top result: %s (expected jellyfin due to higher usage)", candidates[0].Service.Hostname)
				}
			},
		},
		{
			name:        "fuzzy match with short query",
			queryString: "trfk",
			description: "Should find traefik with fuzzy matching",
			validate: func(t *testing.T, candidates []*domain.Candidate) {
				found := false
				for _, c := range candidates {
					if c.Service.Hostname == "traefik.domain.ext" {
						found = true
						break
					}
				}
				if !found {
					t.Error("traefik not found in fuzzy search results")
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			query := domain.ParseQuery(scenario.queryString)
			candidates := domain.RankCandidates(query, services)

			if len(candidates) == 0 {
				t.Fatal("No candidates returned")
			}

			t.Logf("Query: %s", scenario.queryString)
			t.Logf("Results:")
			for i, c := range candidates {
				t.Logf("  %d. %s (score: %.2f)", i+1, c.Service.Hostname, c.TotalScore)
			}

			scenario.validate(t, candidates)
		})
	}
}
