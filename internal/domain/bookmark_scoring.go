package domain

import (
	"strings"
)

// BookmarkCandidate represents a bookmark candidate with its match score
type BookmarkCandidate struct {
	Bookmark *Bookmark
	Score    float64 // Score from fuzzy matching
}

// ScoreBookmark calculates the match score for a bookmark against a query string
func ScoreBookmark(queryStr string, bookmark *Bookmark) float64 {
	if bookmark == nil || queryStr == "" {
		return 0.0
	}

	queryStr = strings.ToLower(strings.TrimSpace(queryStr))
	abbr := strings.ToLower(bookmark.Abbr)

	// Exact match (highest score)
	if queryStr == abbr {
		return ScoreExactMatch + ScoreExactHostnameBonus
	}

	// Prefix match
	if strings.HasPrefix(abbr, queryStr) {
		return ScorePrefixMatch
	}

	// Substring match
	if strings.Contains(abbr, queryStr) {
		index := strings.Index(abbr, queryStr)
		// Earlier substring matches get higher score
		substringBonus := ScorePositionBonus * (1.0 - float64(index)/float64(len(abbr)))
		return ScoreSubstringMatch + substringBonus
	}

	// Fuzzy match (word-based)
	// Check if all query words appear in abbr
	queryWords := strings.Fields(queryStr)
	if len(queryWords) > 1 {
		allMatch := true
		for _, word := range queryWords {
			if !strings.Contains(abbr, word) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return ScoreFuzzyMatch
		}
	}

	// Character similarity
	similarity := calculateSimilarity(queryStr, abbr)
	if similarity > 0.5 {
		return ScoreFuzzyMatch * similarity
	}

	return 0.0
}

// RankBookmarkCandidates ranks bookmark candidates by score
func RankBookmarkCandidates(queryStr string, bookmarks []*Bookmark) []*BookmarkCandidate {
	candidates := make([]*BookmarkCandidate, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		// Skip disabled bookmarks
		if bookmark.Disabled {
			continue
		}

		score := ScoreBookmark(queryStr, bookmark)

		// Skip bookmarks with zero score (no match)
		if score == 0.0 {
			continue
		}

		candidates = append(candidates, &BookmarkCandidate{
			Bookmark: bookmark,
			Score:    score,
		})
	}

	// Sort candidates by score (descending)
	sortBookmarkCandidates(candidates)

	return candidates
}

// sortBookmarkCandidates sorts candidates by score (descending)
func sortBookmarkCandidates(candidates []*BookmarkCandidate) {
	// Simple bubble sort (fine for small lists)
	n := len(candidates)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if candidates[j].Score < candidates[j+1].Score {
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}
}

// FindBestBookmark finds the best matching bookmark for a query
func FindBestBookmark(queryStr string, bookmarks []*Bookmark) *Bookmark {
	candidates := RankBookmarkCandidates(queryStr, bookmarks)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0].Bookmark
}
