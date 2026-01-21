package domain

import (
	"strings"
	"unicode"
)

// Query represents a parsed user input
type Query struct {
	Raw                string   // Original input
	Fragments          []string // Space-separated fragments
	HasDot             bool     // Whether input contains dot (enables subdomain matching)
	TopLevelFragments  []string // Fragments before first dot (or all if no dot)
	SubdomainFragments []string // Fragments after first dot (empty if no dot)
}

// ParseQuery parses user input into a structured query
// Examples:
//   - "jelly pro" -> top-level only, unordered: ["jelly", "pro"]
//   - "jelly.prod" -> subdomain enabled, ordered: ["jelly"] + ["prod"]
//   - "jelly.srv sta" -> subdomain enabled: ["jelly", "srv"] + unordered ["sta"]
func ParseQuery(input string) *Query {
	// Normalize input
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return &Query{Raw: input}
	}

	q := &Query{
		Raw:    input,
		HasDot: strings.Contains(input, "."),
	}

	if !q.HasDot {
		// No dot: simple space-separated fragments (top-level only)
		q.Fragments = splitAndClean(input, " ")
		q.TopLevelFragments = q.Fragments
		return q
	}

	// Has dot: split by dot first, then handle spaces
	dotParts := strings.Split(input, ".")

	// First part before dot (may contain spaces)
	if len(dotParts) > 0 && dotParts[0] != "" {
		firstPart := splitAndClean(dotParts[0], " ")
		q.TopLevelFragments = append(q.TopLevelFragments, firstPart...)
	}

	// Parts after dot (may contain spaces)
	for i := 1; i < len(dotParts); i++ {
		if dotParts[i] == "" {
			continue
		}
		subParts := splitAndClean(dotParts[i], " ")
		q.SubdomainFragments = append(q.SubdomainFragments, subParts...)
	}

	// All fragments combined (for unordered matching)
	q.Fragments = make([]string, 0, len(q.TopLevelFragments)+len(q.SubdomainFragments))
	q.Fragments = append(q.Fragments, q.TopLevelFragments...)
	q.Fragments = append(q.Fragments, q.SubdomainFragments...)

	return q
}

// splitAndClean splits a string by separator and returns non-empty parts
func splitAndClean(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// HostnameFragments extracts fragments from a hostname for matching
// Example: "jellyfin.srv1.staging.domain.ext" -> ["jellyfin", "srv1", "staging", "nexacloud", "dev"]
func HostnameFragments(hostname string) []string {
	return strings.Split(strings.ToLower(hostname), ".")
}

// normalizeFragment normalizes a fragment for matching
func normalizeFragment(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}
