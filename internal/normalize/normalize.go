package normalize

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Slug represents a GitHub repository in owner/repo form.
type Slug struct {
	Owner string
	Repo  string
}

// ErrInvalidRepo indicates that the provided repository reference could not be parsed.
var ErrInvalidRepo = errors.New("invalid GitHub repository reference")

var slugPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var partPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Parse attempts to build a slug from a variety of user inputs, including
// owner/repo strings, HTTPS URLs, SSH URLs, and remote declarations.
func Parse(input string) (Slug, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Slug{}, ErrInvalidRepo
	}

	trimmed = strings.TrimSuffix(trimmed, "/")

	if slugPattern.MatchString(trimmed) {
		parts := strings.SplitN(trimmed, "/", 2)
		return newSlug(parts[0], parts[1])
	}

	if slug, ok := fromURL(trimmed); ok {
		return slug, nil
	}

	if slug, ok := fromSCP(trimmed); ok {
		return slug, nil
	}

	return Slug{}, ErrInvalidRepo
}

// Format returns owner/repo string.
func (s Slug) String() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repo)
}

// RepoURL returns the canonical GitHub URL for the slug.
func (s Slug) RepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", s.Owner, s.Repo)
}

// Encoded returns the owner/repo string URL encoded for use in query params.
func (s Slug) Encoded() string {
	return url.QueryEscape(s.String())
}

func newSlug(owner, repo string) (Slug, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	repo = strings.TrimSuffix(repo, ".git")
	if owner == "" || repo == "" {
		return Slug{}, ErrInvalidRepo
	}

	if !partPattern.MatchString(owner) || !partPattern.MatchString(repo) {
		return Slug{}, ErrInvalidRepo
	}

	return Slug{Owner: owner, Repo: repo}, nil
}

func fromURL(raw string) (Slug, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return Slug{}, false
	}

	host := strings.ToLower(parsed.Host)
	if host != "github.com" {
		return Slug{}, false
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return Slug{}, false
	}

	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return Slug{}, false
	}

	slug, err := newSlug(segments[0], segments[1])
	if err != nil {
		return Slug{}, false
	}

	return slug, true
}

func fromSCP(raw string) (Slug, bool) {
	if !strings.Contains(raw, ":") {
		return Slug{}, false
	}

	parts := strings.SplitN(raw, ":", 2)
	host := parts[0]
	path := parts[1]

	if !strings.Contains(host, "github.com") {
		return Slug{}, false
	}

	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return Slug{}, false
	}

	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return Slug{}, false
	}

	slug, err := newSlug(segments[0], segments[1])
	if err != nil {
		return Slug{}, false
	}

	return slug, true
}
