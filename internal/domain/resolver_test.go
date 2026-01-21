package domain

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expectedHasDot       bool
		expectedTopLevel     []string
		expectedSubdomain    []string
		expectedAllFragments []string
	}{
		{
			name:                 "simple query without dot",
			input:                "jelly",
			expectedHasDot:       false,
			expectedTopLevel:     []string{"jelly"},
			expectedSubdomain:    []string{},
			expectedAllFragments: []string{"jelly"},
		},
		{
			name:                 "multiple fragments without dot",
			input:                "jelly pro",
			expectedHasDot:       false,
			expectedTopLevel:     []string{"jelly", "pro"},
			expectedSubdomain:    []string{},
			expectedAllFragments: []string{"jelly", "pro"},
		},
		{
			name:                 "query with dot",
			input:                "jelly.prod",
			expectedHasDot:       true,
			expectedTopLevel:     []string{"jelly"},
			expectedSubdomain:    []string{"prod"},
			expectedAllFragments: []string{"jelly", "prod"},
		},
		{
			name:                 "query with dot and spaces",
			input:                "jelly.srv sta",
			expectedHasDot:       true,
			expectedTopLevel:     []string{"jelly"},
			expectedSubdomain:    []string{"srv", "sta"},
			expectedAllFragments: []string{"jelly", "srv", "sta"},
		},
		{
			name:                 "empty query",
			input:                "",
			expectedHasDot:       false,
			expectedTopLevel:     nil,
			expectedSubdomain:    nil,
			expectedAllFragments: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.input)

			if query.HasDot != tt.expectedHasDot {
				t.Errorf("HasDot = %v, want %v", query.HasDot, tt.expectedHasDot)
			}

			if !slicesEqual(query.TopLevelFragments, tt.expectedTopLevel) {
				t.Errorf("TopLevelFragments = %v, want %v", query.TopLevelFragments, tt.expectedTopLevel)
			}

			if !slicesEqual(query.SubdomainFragments, tt.expectedSubdomain) {
				t.Errorf("SubdomainFragments = %v, want %v", query.SubdomainFragments, tt.expectedSubdomain)
			}

			if !slicesEqual(query.Fragments, tt.expectedAllFragments) {
				t.Errorf("Fragments = %v, want %v", query.Fragments, tt.expectedAllFragments)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestScore(t *testing.T) {
	tests := []struct {
		name           string
		queryStr       string
		hostname       string
		expectPositive bool
	}{
		{
			name:           "exact match",
			queryStr:       "jellyfin",
			hostname:       "jellyfin.domain.ext",
			expectPositive: true,
		},
		{
			name:           "prefix match",
			queryStr:       "jelly",
			hostname:       "jellyfin.domain.ext",
			expectPositive: true,
		},
		{
			name:           "no match",
			queryStr:       "xyz",
			hostname:       "jellyfin.domain.ext",
			expectPositive: false,
		},
		{
			name:           "subdomain match",
			queryStr:       "jelly.prod",
			hostname:       "jellyfin.production.domain.ext",
			expectPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.queryStr)
			service := &Service{
				ID:       tt.hostname,
				Hostname: tt.hostname,
				Name:     "test",
			}

			score := Score(query, service)

			if tt.expectPositive && score <= 0 {
				t.Errorf("Expected positive score, got %f", score)
			}

			if !tt.expectPositive && score > 0 {
				t.Errorf("Expected zero score, got %f", score)
			}
		})
	}
}

func TestRankCandidates_DisabledFilter(t *testing.T) {
	services := []*Service{
		{
			ID:       "active.example.com",
			Hostname: "active.example.com",
			Name:     "active",
			Disabled: false,
		},
		{
			ID:       "disabled.example.com",
			Hostname: "disabled.example.com",
			Name:     "disabled",
			Disabled: true,
		},
		{
			ID:       "another-active.example.com",
			Hostname: "another-active.example.com",
			Name:     "another-active",
			Disabled: false,
		},
	}

	query := ParseQuery("active")
	candidates := RankCandidates(query, services)

	// Should only return 2 active services
	if len(candidates) != 2 {
		t.Errorf("Expected 2 candidates (disabled should be filtered), got %d", len(candidates))
	}

	// Check that disabled service is not in candidates
	for _, c := range candidates {
		if c.Service.Disabled {
			t.Error("Disabled service should not be in candidates")
		}
		if c.Service.ID == "disabled.example.com" {
			t.Error("disabled.example.com should not be in candidates")
		}
	}
}
