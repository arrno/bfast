package readme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	BadgeImageURL = "https://blazingly.fast/api/badge.svg"
	BadgeLinkURL  = "https://blazingly.fast"
)

var ErrNotFound = errors.New("readme not found")

var defaultCandidates = []string{
	"README.md",
	"Readme.md",
	"README.MD",
	"README.markdown",
	"README.Markdown",
	"README",
	"readme.md",
	"readme",
}

// FindDefault locates a README-like file inside the repo root.
func FindDefault(root string) (string, error) {
	for _, name := range defaultCandidates {
		path := filepath.Join(root, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, nil
		}
	}

	return "", ErrNotFound
}

// HasBadge reports whether the README already contains the blazingly.fast badge.
func HasBadge(content string) bool {
	lower := strings.ToLower(content)
	if strings.Contains(lower, strings.ToLower(BadgeImageURL)) {
		return true
	}

	for _, line := range strings.Split(lower, "\n") {
		if strings.Contains(line, "![") && strings.Contains(line, "blazingly fast") && strings.Contains(line, "blazingly.fast") {
			return true
		}
	}

	return false
}

// InsertBadge returns README content with the badge inserted.
func InsertBadge(content, badge string) (string, error) {
	if strings.TrimSpace(badge) == "" {
		return "", errors.New("badge content may not be empty")
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	if len(lines) == 0 {
		lines = []string{""}
	}

	insertIdx := -1
	_, blockEnd, ok := findBadgeBlock(lines)
	if ok {
		insertIdx = blockEnd + 1
	} else if titleIdx := findTitle(lines); titleIdx >= 0 {
		insertIdx = titleIdx + 1
		if insertIdx < len(lines) && strings.TrimSpace(lines[insertIdx]) != "" {
			lines = insertLine(lines, insertIdx, "")
			insertIdx++
		}
	}

	if insertIdx == -1 {
		lines = append([]string{badge, ""}, lines...)
	} else {
		lines = insertLine(lines, insertIdx, badge)
	}

	output := strings.Join(lines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	return output, nil
}

func findBadgeBlock(lines []string) (int, int, bool) {
	limit := len(lines)
	if limit > 20 {
		limit = 20
	}

	inBlock := false
	start := -1
	end := -1

	for i := 0; i < limit; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			if inBlock {
				break
			}
			continue
		}

		if looksLikeBadge(trimmed) {
			if !inBlock {
				inBlock = true
				start = i
			}
			end = i
			continue
		}

		if inBlock {
			break
		}

		if strings.HasPrefix(trimmed, "# ") {
			break
		}
	}

	if inBlock && start >= 0 && end >= start {
		return start, end, true
	}

	return -1, -1, false
}

func findTitle(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return i
		}
	}
	return -1
}

func looksLikeBadge(line string) bool {
	return strings.HasPrefix(line, "![") || strings.HasPrefix(line, "[![")
}

func insertLine(lines []string, idx int, value string) []string {
	if idx < 0 {
		idx = 0
	}

	if idx >= len(lines) {
		return append(lines, value)
	}

	out := make([]string, len(lines)+1)
	copy(out, lines[:idx])
	out[idx] = value
	copy(out[idx+1:], lines[idx:])
	return out
}

// BuildBadgeMarkdown returns the markdown snippet for the badge.
func BuildBadgeMarkdown(encodedSlug string) string {
	return fmt.Sprintf("[![blazingly fast](%s?repo=%s)](%s)", BadgeImageURL, encodedSlug, BadgeLinkURL)
}
