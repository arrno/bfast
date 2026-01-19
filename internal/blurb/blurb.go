package blurb

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const maxLength = 128

var (
	ErrInvalidBlurb = errors.New("blurb must be between 1 and 128 characters")
	rng             = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu           sync.Mutex

	defaultBlurbs = []string{
		"Declared blazingly fast by the author.",
		"Performance considered, benchmarks omitted.",
		"Fast enough for its intended use.",
		"Runs like the wind on maintainer laptops.",
		"Finally faster than the previous rewrite.",
		"Certified swift by unverified claims.",
		"Optimized for perceived speed.",
		"Moving at the speed of developer confidence.",
		"Latency measured in gut feelings.",
		"Profiled once, found acceptable.",
		"Runs hot, runs fast, looks cool.",
		"Untimed, but unquestionably rapid.",
		"Ships velocity the old-fashioned way: by saying so.",
		"Fueled by caffeine and claims of speed.",
		"Benchmarks available upon polite request.",
		"Speed verified during a live demo.",
		"Peaks at impressive velocity when no one is watching.",
		"Breaks the sound barrier in optimistic scenarios.",
		"Fast-path paved, slow-path unexplored.",
		"Sprints through workloads with dramatic flair.",
		"Speed limit signs are merely suggestions here.",
		"Clocked by eyeballing task manager graphs.",
		"Consistently ahead in hypothetical races.",
		"Glides through code paths like butter.",
		"Moves so fast the logs can hardly keep up.",
		"Accelerates faster than the product requirements.",
		"Practically levitates past performance concerns.",
	}
)

// Normalize verifies an explicit blurb.
func Normalize(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", ErrInvalidBlurb
	}

	if utf8.RuneCountInString(trimmed) > maxLength {
		return "", ErrInvalidBlurb
	}

	return trimmed, nil
}

// Random returns a default blurb from the built-in pool.
func Random() string {
	rngMu.Lock()
	defer rngMu.Unlock()
	idx := rng.Intn(len(defaultBlurbs))
	return defaultBlurbs[idx]
}
