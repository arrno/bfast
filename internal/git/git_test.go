package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arrno/bfast/internal/normalize"
)

func TestFindRepoRoot(t *testing.T) {
	temp := t.TempDir()
	gitDir := filepath.Join(temp, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	nested := filepath.Join(temp, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	got, err := FindRepoRoot(nested)
	if err != nil {
		t.Fatalf("FindRepoRoot returned error: %v", err)
	}

	if got != temp {
		t.Fatalf("FindRepoRoot = %s, want %s", got, temp)
	}
}

func TestFindRepoRootNotRepo(t *testing.T) {
	temp := t.TempDir()
	if _, err := FindRepoRoot(temp); err == nil || err != ErrNotRepository {
		t.Fatalf("expected ErrNotRepository, got %v", err)
	}
}

func TestParseRemotesPrefersOrigin(t *testing.T) {
	out := "upstream\thttps://github.com/other/repo.git (fetch)\n" +
		"origin\tgit@github.com:arrno/bfast.git (push)\n"

	slug, err := parseRemotes(out)
	if err != nil {
		t.Fatalf("parseRemotes returned error: %v", err)
	}

	want := normalize.Slug{Owner: "arrno", Repo: "bfast"}
	if slug != want {
		t.Fatalf("parseRemotes = %+v, want %+v", slug, want)
	}
}

func TestParseRemotesAmbiguous(t *testing.T) {
	out := "upstream\thttps://github.com/foo/bar.git (fetch)\n" +
		"another\tgit@github.com:baz/qux.git (fetch)\n"

	if _, err := parseRemotes(out); err == nil || err != ErrAmbiguousRepo {
		t.Fatalf("expected ErrAmbiguousRepo, got %v", err)
	}
}

func TestParseRemotesNoGithub(t *testing.T) {
	if _, err := parseRemotes("origin\thttps://gitlab.com/foo/bar.git (fetch)\n"); err == nil || err != ErrNoGithubRemote {
		t.Fatalf("expected ErrNoGithubRemote, got %v", err)
	}
}
