package domain

import (
	"math"
	"strings"
)

const (
	// Scoring weights
	ScoreExactMatch     = 100.0
	ScorePrefixMatch    = 75.0
	ScoreSubstringMatch = 50.0
	ScoreFuzzyMatch     = 25.0

	// Position bonus (earlier is better)
	ScorePositionBonus = 10.0

	// Length bonus/penalty
	ScoreLengthBonus = 5.0

	// Exact hostname match bonus (huge boost)
	ScoreExactHostnameBonus = 200.0

	// Usage weight (usage counter contributes to final score)
	ScoreUsageWeight = 0.1
)

// Candidate represents a service candidate with its match score
type Candidate struct {
	Service      *Service
	LexicalScore float64 // Score from fuzzy matching
	UsageScore   float64 // Score from usage learning
	TotalScore   float64 // Combined score
}

// Score calculates the match score for a service against a query
func Score(query *Query, service *Service) float64 {
	if query == nil || service == nil {
		return 0.0
	}

	hostname := strings.ToLower(service.Hostname)
	hostFragments := HostnameFragments(hostname)

	var totalScore float64

	if !query.HasDot {
		// Top-level only matching
		totalScore = scoreTopLevelOnly(query.Fragments, hostFragments)
	} else {
		// Subdomain matching enabled
		totalScore = scoreWithSubdomains(query, hostFragments)
	}

	return totalScore
}

// scoreTopLevelOnly scores when no dot is present (top-level only)
func scoreTopLevelOnly(queryFragments []string, hostFragments []string) float64 {
	if len(queryFragments) == 0 || len(hostFragments) == 0 {
		return 0.0
	}

	// Only match against first fragment of hostname (top-level)
	topLevel := hostFragments[0]

	var totalScore float64

	// Check for exact match first (single fragment query matching exactly)
	if len(queryFragments) == 1 && queryFragments[0] == topLevel {
		totalScore += ScoreExactMatch + ScoreExactHostnameBonus
		return totalScore
	}

	for _, qFrag := range queryFragments {
		score := scoreFragment(qFrag, topLevel, 0)
		totalScore += score
	}

	// Only apply length bonus if there was a match
	if totalScore > 0 && len(topLevel) < 10 {
		totalScore += ScoreLengthBonus
	}

	return totalScore
}

// scoreWithSubdomains scores when dot is present (subdomain matching)
// When user types "x.y", they explicitly want subdomain matching ONLY
// Top-level must match AND at least one subdomain fragment must match
func scoreWithSubdomains(query *Query, hostFragments []string) float64 {
	if len(query.Fragments) == 0 || len(hostFragments) == 0 {
		return 0.0
	}

	// Require subdomain fragments when dot is used
	if len(query.SubdomainFragments) == 0 {
		return 0.0
	}

	// Require hostname to have subdomain parts
	if len(hostFragments) < 2 {
		return 0.0
	}

	var totalScore float64

	// Match top-level fragments against first hostname fragment
	if len(query.TopLevelFragments) > 0 {
		topLevel := hostFragments[0]
		for _, qFrag := range query.TopLevelFragments {
			score := scoreFragment(qFrag, topLevel, 0)
			totalScore += score
		}
	}

	// Match subdomain fragments against remaining hostname fragments
	// This is REQUIRED when HasDot=true
	subdomainScore := 0.0
	subdomains := hostFragments[1:]
	for _, qFrag := range query.SubdomainFragments {
		bestScore := 0.0
		for i, hFrag := range subdomains {
			score := scoreFragment(qFrag, hFrag, i)
			if score > bestScore {
				bestScore = score
			}
		}
		subdomainScore += bestScore
	}

	// Only return score if subdomain fragments matched
	if subdomainScore == 0.0 {
		return 0.0
	}

	totalScore += subdomainScore

	return totalScore
}

// scoreFragment scores a single query fragment against a hostname fragment
func scoreFragment(queryFrag, hostFrag string, position int) float64 {
	queryFrag = normalizeFragment(queryFrag)
	hostFrag = normalizeFragment(hostFrag)

	if queryFrag == "" || hostFrag == "" {
		return 0.0
	}

	// Exact match
	if queryFrag == hostFrag {
		return ScoreExactMatch + calculatePositionBonus(position)
	}

	// Prefix match
	if strings.HasPrefix(hostFrag, queryFrag) {
		return ScorePrefixMatch + calculatePositionBonus(position)
	}

	// Substring match
	if strings.Contains(hostFrag, queryFrag) {
		index := strings.Index(hostFrag, queryFrag)
		// Earlier substring matches get higher score
		substringBonus := ScorePositionBonus * (1.0 - float64(index)/float64(len(hostFrag)))
		return ScoreSubstringMatch + substringBonus
	}

	// Fuzzy match (Levenshtein-like)
	similarity := calculateSimilarity(queryFrag, hostFrag)
	if similarity > 0.5 {
		return ScoreFuzzyMatch * similarity
	}

	return 0.0
}

// calculatePositionBonus gives bonus for earlier positions
func calculatePositionBonus(position int) float64 {
	return ScorePositionBonus * math.Exp(-float64(position)*0.3)
}

// calculateSimilarity calculates fuzzy similarity between two strings
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return 0.0
	}

	// Simple similarity: ratio of matching characters
	matches := 0
	for _, c := range s1 {
		if strings.ContainsRune(s2, c) {
			matches++
		}
	}

	return float64(matches) / float64(len(s1))
}

// RankCandidates ranks service candidates by combining lexical and usage scores
func RankCandidates(query *Query, services []*Service) []*Candidate {
	candidates := make([]*Candidate, 0, len(services))

	for _, service := range services {
		// Skip disabled services
		if service.Disabled {
			continue
		}

		lexicalScore := Score(query, service)

		// Skip services with zero lexical score (no match)
		if lexicalScore == 0.0 {
			continue
		}

		// Calculate usage score (logarithmic to prevent dominance)
		usageScore := 0.0
		if service.Counter > 0 {
			usageScore = math.Log10(float64(service.Counter)+1) * ScoreUsageWeight * 100
		}

		totalScore := lexicalScore + usageScore

		candidates = append(candidates, &Candidate{
			Service:      service,
			LexicalScore: lexicalScore,
			UsageScore:   usageScore,
			TotalScore:   totalScore,
		})
	}

	// Sort candidates by total score (descending)
	sortCandidates(candidates)

	return candidates
}

// sortCandidates sorts candidates by total score (descending)
func sortCandidates(candidates []*Candidate) {
	// Simple bubble sort (fine for small lists)
	n := len(candidates)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if candidates[j].TotalScore < candidates[j+1].TotalScore {
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}
}

// FindBestMatch finds the best matching service for a query
func FindBestMatch(query *Query, services []*Service) *Service {
	candidates := RankCandidates(query, services)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0].Service
}
