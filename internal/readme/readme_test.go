package readme

import (
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

func TestInsertBadge(t *testing.T) {
	content := "# Project\n\nSome text"
	badge := "[![blazingly fast](https://blazingly.fast/api/badge.svg?repo=proj)](https://blazingly.fast)"

	updated, err := InsertBadge(content, badge)
	if err != nil {
		t.Fatalf("InsertBadge returned error: %v", err)
	}

	want := "# Project\n" + badge + "\n\nSome text\n"
	if updated != want {
		t.Fatalf("InsertBadge produced unexpected content:\n%s\nwant:\n%s", updated, want)
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
