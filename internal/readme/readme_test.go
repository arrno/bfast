package readme

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasBadge(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"image", "[![blazingly fast](https://blazingly.fast/api/badge.svg?repo=x%2Fy)](https://blazingly.fast)", true},
		{"alt text", "[![Certified blazingly fast](https://example.com)](https://blazingly.fast)", true},
		{"absent", "# Title\nSome text", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HasBadge(tc.content)
			if got != tc.want {
				t.Fatalf("HasBadge(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestInsertBadgePreservesWindowsNewlines(t *testing.T) {
	content := "# Project\r\n\r\nSome text"
	badge := "[![blazingly fast](https://blazingly.fast/api/badge.svg?repo=proj)](https://blazingly.fast)"

	updated, err := InsertBadge(content, badge)
	if err != nil {
		t.Fatalf("InsertBadge returned error: %v", err)
	}

	want := "# Project\r\n" + badge + "\r\n\r\nSome text\r\n"
	if updated != want {
		t.Fatalf("InsertBadge produced unexpected content:\n%q\nwant:\n%q", updated, want)
	}

	if strings.Contains(strings.ReplaceAll(updated, "\r\n", ""), "\n") {
		t.Fatalf("InsertBadge introduced bare LF newlines: %q", updated)
	}
}

func TestInsertBadgeFormattingFixtures(t *testing.T) {
	badge := "[![blazingly fast](https://blazingly.fast/api/badge.svg?repo=proj)](https://blazingly.fast)"
	for _, tc := range loadInsertBadgeCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			updated, err := InsertBadge(tc.Before, badge)
			if err != nil {
				t.Fatalf("InsertBadge returned error: %v", err)
			}

			if updated != tc.After {
				t.Fatalf("InsertBadge produced unexpected content:\n%s\nwant:\n%s", updated, tc.After)
			}
		})
	}
}

type insertBadgeCase struct {
	Name   string
	Before string
	After  string
}

func loadInsertBadgeCases(t *testing.T) []insertBadgeCase {
	t.Helper()
	path := filepath.Join("testdata", "insert_badge_cases.txt")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open testdata: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var cases []insertBadgeCase
	var current *insertBadgeCase
	var mode string
	var buf strings.Builder

	save := func() {
		if current == nil || mode == "" {
			return
		}
		text := buf.String()
		switch mode {
		case "before":
			current.Before = text
		case "after":
			current.After = text
		}
		buf.Reset()
		mode = ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "=== case:"):
			if current != nil {
				t.Fatalf("started new case before closing %q", current.Name)
			}
			name := strings.TrimSuffix(strings.TrimPrefix(line, "=== case:"), " ===")
			current = &insertBadgeCase{Name: strings.TrimSpace(name)}
		case line == "--- before ---":
			if current == nil {
				t.Fatalf("encountered before block without active case")
			}
			save()
			mode = "before"
		case line == "--- after ---":
			save()
			mode = "after"
		case line == "=== end ===":
			save()
			if current == nil || current.Before == "" || current.After == "" {
				t.Fatalf("encountered malformed case ending")
			}
			cases = append(cases, *current)
			current = nil
		default:
			if mode != "" {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan testdata: %v", err)
	}

	if current != nil {
		t.Fatalf("unterminated case %q", current.Name)
	}

	if len(cases) == 0 {
		t.Fatalf("no insert badge cases loaded")
	}

	return cases
}
