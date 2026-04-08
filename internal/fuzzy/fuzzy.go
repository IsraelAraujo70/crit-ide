// Package fuzzy provides a score-based fuzzy matching algorithm
// for the Telescope-style file finder (Ctrl+P).
package fuzzy

import (
	"strings"
	"unicode"
)

// MatchResult holds the result of a fuzzy match.
type MatchResult struct {
	Score   int   // Higher is better. 0 means no match.
	Matches []int // Indices of matched characters in the candidate.
}

// Match performs fuzzy matching of pattern against candidate.
// Returns a MatchResult with score > 0 if the pattern matches.
// Scoring priorities:
//   - Consecutive matches score higher than scattered.
//   - Start-of-word matches score higher than mid-word.
//   - Exact prefix match gets a large bonus.
//   - Case-sensitive exact matches score higher than case-insensitive.
func Match(pattern, candidate string) MatchResult {
	if pattern == "" {
		return MatchResult{Score: 1}
	}
	if candidate == "" {
		return MatchResult{}
	}

	patLower := strings.ToLower(pattern)
	candLower := strings.ToLower(candidate)

	// Quick reject: every pattern character must exist in candidate.
	for _, ch := range patLower {
		if !strings.ContainsRune(candLower, ch) {
			return MatchResult{}
		}
	}

	// Find matches using a greedy algorithm that favors word boundaries.
	matches := findBestMatches(patLower, candLower, candidate)
	if matches == nil {
		return MatchResult{}
	}

	score := computeScore(matches, pattern, candidate, candLower)
	return MatchResult{Score: score, Matches: matches}
}

// findBestMatches attempts to match pattern characters against candidate,
// preferring word boundary positions. Falls back to first-available match.
func findBestMatches(patLower, candLower, candidate string) []int {
	patRunes := []rune(patLower)
	candRunes := []rune(candLower)
	origRunes := []rune(candidate)

	// First pass: try to match at word boundaries.
	matches := matchPreferBoundaries(patRunes, candRunes, origRunes)
	if matches != nil {
		return matches
	}

	// Second pass: greedy first-available match.
	return matchGreedy(patRunes, candRunes)
}

// matchPreferBoundaries tries to match pattern chars at word boundaries first,
// then fills remaining chars greedily.
func matchPreferBoundaries(patRunes, candRunes, origRunes []rune) []int {
	n := len(patRunes)
	matches := make([]int, 0, n)
	pi := 0

	for ci := 0; ci < len(candRunes) && pi < n; ci++ {
		if candRunes[ci] != patRunes[pi] {
			continue
		}

		// Check if this is a word boundary position.
		isBoundary := ci == 0 ||
			!unicode.IsLetter(origRunes[ci-1]) ||
			(unicode.IsUpper(origRunes[ci]) && !unicode.IsUpper(origRunes[ci-1]))

		if isBoundary || pi == len(matches) {
			// Accept this match if it's a boundary or we haven't found better.
			matches = append(matches, ci)
			pi++
		}
	}

	// If we couldn't match all pattern chars via boundaries, use greedy fallback.
	if pi < n {
		return nil
	}
	return matches
}

// matchGreedy matches pattern characters left-to-right, taking the first available.
func matchGreedy(patRunes, candRunes []rune) []int {
	n := len(patRunes)
	matches := make([]int, 0, n)
	pi := 0

	for ci := 0; ci < len(candRunes) && pi < n; ci++ {
		if candRunes[ci] == patRunes[pi] {
			matches = append(matches, ci)
			pi++
		}
	}

	if pi < n {
		return nil
	}
	return matches
}

// computeScore calculates a score for the match based on match quality.
func computeScore(matches []int, pattern, candidate, candLower string) int {
	if len(matches) == 0 {
		return 1
	}

	origRunes := []rune(candidate)
	patRunes := []rune(pattern)
	score := 0

	for i, ci := range matches {
		// Base score per matched character.
		score += 10

		// Bonus for case-sensitive exact match.
		if i < len(patRunes) && ci < len(origRunes) && origRunes[ci] == patRunes[i] {
			score += 2
		}

		// Bonus for start of word.
		if ci == 0 {
			score += 15
		} else if ci < len(origRunes) {
			prev := origRunes[ci-1]
			if prev == '/' || prev == '\\' || prev == '.' || prev == '_' || prev == '-' || prev == ' ' {
				score += 12
			} else if unicode.IsUpper(origRunes[ci]) && !unicode.IsUpper(prev) {
				score += 10 // camelCase boundary.
			}
		}

		// Bonus for consecutive matches.
		if i > 0 && ci == matches[i-1]+1 {
			score += 8
		}
	}

	// Bonus for matches starting early in the string (filename part).
	lastSlash := strings.LastIndexByte(candidate, '/')
	firstMatch := matches[0]
	if lastSlash >= 0 && firstMatch > lastSlash {
		// Match is in the filename part — big bonus.
		score += 20
		distFromName := firstMatch - lastSlash - 1
		if distFromName < 5 {
			score += (5 - distFromName) * 3
		}
	} else if firstMatch < 5 {
		score += (5 - firstMatch) * 3
	}

	// Bonus for shorter candidates (more specific matches).
	if len(candidate) < 30 {
		score += 5
	}

	// Bonus for exact prefix match.
	if strings.HasPrefix(candLower, strings.ToLower(pattern)) {
		score += 30
	}

	// Bonus for exact filename match.
	baseName := candidate
	if lastSlash >= 0 {
		baseName = candidate[lastSlash+1:]
	}
	if strings.HasPrefix(strings.ToLower(baseName), strings.ToLower(pattern)) {
		score += 25
	}

	return score
}
