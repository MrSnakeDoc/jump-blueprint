package domain

import "testing"

func TestScoreBookmark(t *testing.T) {
	tests := []struct {
		name           string
		queryStr       string
		abbr           string
		expectPositive bool
	}{
		{
			name:           "exact match",
			queryStr:       "chatgpt",
			abbr:           "ChatGPT",
			expectPositive: true,
		},
		{
			name:           "prefix match",
			queryStr:       "chat",
			abbr:           "ChatGPT",
			expectPositive: true,
		},
		{
			name:           "substring match",
			queryStr:       "gpt",
			abbr:           "ChatGPT",
			expectPositive: true,
		},
		{
			name:           "no match",
			queryStr:       "xyz",
			abbr:           "ChatGPT",
			expectPositive: false,
		},
		{
			name:           "multi-word match",
			queryStr:       "docker hub",
			abbr:           "Docker Hub",
			expectPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark := &Bookmark{
				ID:   "test-id",
				Abbr: tt.abbr,
				URL:  "https://example.com",
			}

			score := ScoreBookmark(tt.queryStr, bookmark)

			if tt.expectPositive && score <= 0 {
				t.Errorf("Expected positive score, got %f", score)
			}

			if !tt.expectPositive && score > 0 {
				t.Errorf("Expected zero score, got %f", score)
			}
		})
	}
}

func TestRankBookmarkCandidates_DisabledFilter(t *testing.T) {
	bookmarks := []*Bookmark{
		{
			ID:       "active-bookmark",
			Abbr:     "ChatGPT",
			URL:      "https://chat.openai.com",
			Disabled: false,
		},
		{
			ID:       "disabled-bookmark",
			Abbr:     "Disabled",
			URL:      "https://disabled.com",
			Disabled: true,
		},
		{
			ID:       "another-active",
			Abbr:     "GitHub",
			URL:      "https://github.com",
			Disabled: false,
		},
	}

	queryStr := "chat"
	candidates := RankBookmarkCandidates(queryStr, bookmarks)

	// Should only return active bookmarks
	if len(candidates) > 2 {
		t.Errorf("Expected at most 2 candidates (disabled should be filtered), got %d", len(candidates))
	}

	// Check that disabled bookmark is not in candidates
	for _, c := range candidates {
		if c.Bookmark.Disabled {
			t.Error("Disabled bookmark should not be in candidates")
		}
		if c.Bookmark.ID == "disabled-bookmark" {
			t.Error("disabled-bookmark should not be in candidates")
		}
	}
}
