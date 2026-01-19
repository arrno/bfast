package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arrno/bfast/internal/readme"
)

type submission struct {
	RepoURL         string `json:"repoUrl"`
	IsBlazinglyFast bool   `json:"isBlazinglyFast"`
	Blurb           string `json:"blurb"`
	Hidden          bool   `json:"hidden"`
}

func TestIntegrationRegistersAndBadges(t *testing.T) {
	temp := t.TempDir()
	initGitRepo(t, temp, "https://github.com/arrno/demo.git")

	readmePath := filepath.Join(temp, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Demo\n\nHi"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	payloadCh := make(chan submission, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		defer r.Body.Close()
		var sub submission
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			t.Fatalf("decode: %v", err)
		}
		payloadCh <- sub
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	defer srv.Close()

	t.Setenv(apiBaseEnv, srv.URL)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := Run(context.Background(), []string{"--blurb", "Verified"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	select {
	case sub := <-payloadCh:
		if sub.RepoURL != "https://github.com/arrno/demo" {
			t.Fatalf("repo url = %s", sub.RepoURL)
		}
		if sub.Blurb != "Verified" {
			t.Fatalf("unexpected blurb: %s", sub.Blurb)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for submission")
	}

	updated, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !readme.HasBadge(string(updated)) {
		t.Fatalf("badge not inserted: %s", string(updated))
	}
}

func TestIntegrationSkipsWhenAlreadyBadged(t *testing.T) {
	temp := t.TempDir()
	initGitRepo(t, temp, "https://github.com/arrno/demo.git")

	badge := readme.BuildBadgeMarkdown("arrno%2Fdemo")
	readmePath := filepath.Join(temp, "README.md")
	content := "# Demo\n" + badge + "\n"
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("API should not be called when badge exists")
	}))
	defer srv.Close()
	t.Setenv(apiBaseEnv, srv.URL)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := Run(context.Background(), []string{}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestIntegrationManualRepoOverride(t *testing.T) {
	temp := t.TempDir()
	readmePath := filepath.Join(temp, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	var sub submission
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	t.Setenv(apiBaseEnv, srv.URL)

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	args := []string{"--repo", "arrno/demo", "--readme", readmePath, "--blurb", "Override"}
	code := Run(context.Background(), args, stdout, stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	if sub.RepoURL != "https://github.com/arrno/demo" {
		t.Fatalf("repo url = %s", sub.RepoURL)
	}
	if sub.Blurb != "Override" {
		t.Fatalf("unexpected blurb: %s", sub.Blurb)
	}
}

func TestIntegrationFailsOnAPIErrors(t *testing.T) {
	temp := t.TempDir()
	initGitRepo(t, temp, "https://github.com/arrno/demo.git")
	readmePath := filepath.Join(temp, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
	}))
	defer srv.Close()
	t.Setenv(apiBaseEnv, srv.URL)

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := Run(context.Background(), []string{"-m", "claims"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected non-zero exit on API failure")
	}
	if !strings.Contains(stderr.String(), "upstream unavailable") {
		t.Fatalf("stderr should mention API error, got %q", stderr.String())
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if readme.HasBadge(string(content)) {
		t.Fatalf("badge should not be inserted when API fails")
	}
}

func TestIntegrationFailsWhenNotInGitRepo(t *testing.T) {
	temp := t.TempDir()
	if err := os.WriteFile(filepath.Join(temp, "README.md"), []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := Run(context.Background(), nil, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected failure when not in git repo")
	}
	if !strings.Contains(stderr.String(), "not inside a git repository") {
		t.Fatalf("stderr should mention git repo failure: %q", stderr.String())
	}
}

func TestIntegrationFailsWhenReadmeMissing(t *testing.T) {
	temp := t.TempDir()
	initGitRepo(t, temp, "https://github.com/arrno/demo.git")

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := Run(context.Background(), nil, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected failure when README missing")
	}
	if !strings.Contains(stderr.String(), "readme not found") {
		t.Fatalf("stderr should mention README missing: %q", stderr.String())
	}
}

func initGitRepo(t *testing.T, dir, remote string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", remote)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
}
