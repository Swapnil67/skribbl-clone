package game

import (
	"strings"
	"unicode"
)

type GuessResult int

const (
	GuessIncorrect GuessResult = iota
	GuessClose                 // 1 or 2 character edits away
	GuessExact                 // Exact match
)

// * CleanText normalizes input: lowercase, trims whitespace, removes non-alphanumeric chars
func CleanText(input string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(input)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Levenshtein computes minimum single-character edits (insertions, deletions, substitutions)
func Levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Dynamic programming row buffer (O(N) space)
	v0 := make([]int, lb+1)
	v1 := make([]int, lb+1)

	for i := 0; i <= lb; i++ {
		v0[i] = i
	}

	for i := 0; i < la; i++ {
		v1[0] = i + 1
		for j := 0; j < lb; j++ {
			cost := 0
			if a[i] != b[j] {
				cost = 1
			}
			minCost := v0[j+1] + 1 // Deletion
			if v1[j]+1 < minCost { // Insertion
				minCost = v1[j] + 1
			}
			if v0[j]+cost < minCost { // Substitution
				minCost = v0[j] + cost
			}
			v1[j+1] = minCost
		}
		copy(v0, v1)
	}

	return v1[lb]
}

func EvaluateGuess(guess, target string) GuessResult {
	cleanGuess := CleanText(guess)
	cleanTarget := CleanText(target)

	if cleanGuess == cleanTarget {
		return GuessExact
	}

	// * Only flag "close" if word is at least 4 letters and distance is 1 or 2
	dist := Levenshtein(cleanGuess, cleanTarget)
	if len(cleanTarget) >= 4 && dist <= 2 && dist > 0 {
		return GuessClose
	}
	return GuessIncorrect
}
