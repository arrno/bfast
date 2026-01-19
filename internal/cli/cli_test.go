package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/arrno/bfast/internal/git"
)

func TestExecuteRequiresGitRepo(t *testing.T) {
	temp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	opts := &options{}
	if _, err := execute(context.Background(), opts, io.Discard); err == nil || err != git.ErrNotRepository {
		t.Fatalf("expected ErrNotRepository, got %v", err)
	}
}

func TestExecuteDryRunWithOverrides(t *testing.T) {
	temp := t.TempDir()
	readmePath := filepath.Join(temp, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Title\n"), 0o644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	opts := &options{
		repoInput:     "arrno/bfast",
		readmeInput:   readmePath,
		dryRun:        true,
		blurb:         "fast",
		blurbProvided: true,
	}

	buf := &bytes.Buffer{}
	res, err := execute(context.Background(), opts, buf)
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	if !res.DryRun {
		t.Fatalf("expected DryRun result")
	}

	if res.Readme != readmePath {
		t.Fatalf("unexpected readme path: %s", res.Readme)
	}

	if res.BadgeMarkdown == "" {
		t.Fatalf("expected badge markdown to be set")
	}

	if res.BadgeInserted {
		t.Fatalf("badge should not be inserted during dry run")
	}

	if res.Blurb != "fast" {
		t.Fatalf("expected blurb to be preserved")
	}
}

func TestEmitResultPrefersAlreadyBadgedMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	res := &result{AlreadyBadged: true, DryRun: true}
	emitResult(res, false, buf)

	got := buf.String()
	want := "Already badged. No changes.\n"
	if got != want {
		t.Fatalf("emitResult output = %q, want %q", got, want)
	}
}
