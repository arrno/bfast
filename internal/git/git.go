package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arrno/bfast/internal/normalize"
)

// Errors returned while detecting repository information.
var (
	ErrNotRepository  = errors.New("not inside a git repository")
	ErrNoGithubRemote = errors.New("could not infer GitHub repo. Use: bfast --repo owner/repo")
	ErrAmbiguousRepo  = errors.New("multiple GitHub remotes detected. Use: bfast --repo owner/repo")
)

// FindRepoRoot walks up from the provided directory until it finds a .git folder.
func FindRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if hasGitDir(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotRepository
		}
		dir = parent
	}
}

func hasGitDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return true
	}
	return false
}

// DetectGithubSlug attempts to infer owner/repo from git remotes.
func DetectGithubSlug(ctx context.Context, root string) (normalize.Slug, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "-v")
	cmd.Dir = root

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return normalize.Slug{}, fmt.Errorf("failed to read git remotes: %w", err)
	}

	return parseRemotes(stdout.String())
}

func parseRemotes(output string) (normalize.Slug, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	candidates := map[string]normalize.Slug{}
	var originSlug *normalize.Slug

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		remoteURL := fields[1]

		slug, err := normalize.Parse(remoteURL)
		if err != nil {
			continue
		}

		key := slug.String()
		if _, exists := candidates[key]; !exists {
			candidates[key] = slug
		}

		if name == "origin" {
			if originSlug == nil {
				tmp := slug
				originSlug = &tmp
			}
		}
	}

	if originSlug != nil {
		return *originSlug, nil
	}

	if len(candidates) == 0 {
		return normalize.Slug{}, ErrNoGithubRemote
	}

	if len(candidates) > 1 {
		return normalize.Slug{}, ErrAmbiguousRepo
	}

	for _, slug := range candidates {
		return slug, nil
	}

	return normalize.Slug{}, ErrNoGithubRemote
}
