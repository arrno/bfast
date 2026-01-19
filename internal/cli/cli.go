package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arrno/bfast/internal/api"
	"github.com/arrno/bfast/internal/blurb"
	"github.com/arrno/bfast/internal/git"
	"github.com/arrno/bfast/internal/normalize"
	"github.com/arrno/bfast/internal/readme"
)

const apiBaseEnv = "BFAST_API_BASE_URL"

// Run executes the CLI and returns an exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts, err := parseArgs(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		emitError(err, opts != nil && opts.json, stdout, stderr)
		return 2
	}

	res, err := execute(ctx, opts, stderr)
	if err != nil {
		emitError(err, opts.json, stdout, stderr)
		return 1
	}

	emitResult(res, opts.json, stdout)
	return 0
}

type stringValue struct {
	value string
	set   bool
}

func (s *stringValue) Set(v string) error {
	s.value = v
	s.set = true
	return nil
}

func (s *stringValue) String() string { return s.value }

func parseArgs(args []string, stderr io.Writer) (*options, error) {
	fs := flag.NewFlagSet("bfast", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var blurbValue stringValue
	fs.Var(&blurbValue, "blurb", "Custom blurb text (max 128 chars)")
	fs.Var(&blurbValue, "m", "Custom blurb text (shorthand)")

	repo := fs.String("repo", "", "Target repository (owner/repo or GitHub URL)")
	readmePath := fs.String("readme", "", "Path to README (defaults to repo README)")
	hidden := fs.Bool("hidden", false, "Submit as hidden")
	dryRun := fs.Bool("dry-run", false, "Show actions without making changes")
	forceBadge := fs.Bool("force-badge", false, "Insert badge even if API call fails")
	jsonOut := fs.Bool("json", false, "Emit machine-readable JSON output")

	if err := fs.Parse(args); err != nil {
		return &options{json: jsonOut != nil && *jsonOut}, err
	}

	remainingArgs := fs.Args()
	if len(remainingArgs) > 1 {
		return &options{json: jsonOut != nil && *jsonOut}, errors.New("too many positional arguments")
	}

	positionalRepo := ""
	if len(remainingArgs) == 1 {
		positionalRepo = remainingArgs[0]
	}

	if *repo != "" && positionalRepo != "" {
		return &options{json: jsonOut != nil && *jsonOut}, errors.New("repo provided via --repo and positional argument")
	}

	targetRepo := *repo
	if targetRepo == "" {
		targetRepo = positionalRepo
	}

	return &options{
		blurb:         blurbValue.value,
		blurbProvided: blurbValue.set,
		repoInput:     strings.TrimSpace(targetRepo),
		readmeInput:   strings.TrimSpace(*readmePath),
		hidden:        *hidden,
		dryRun:        *dryRun,
		forceBadge:    *forceBadge,
		json:          *jsonOut,
	}, nil
}

type options struct {
	blurb         string
	blurbProvided bool
	repoInput     string
	readmeInput   string
	hidden        bool
	dryRun        bool
	forceBadge    bool
	json          bool
}

type result struct {
	Repo               string `json:"repo"`
	RepoURL            string `json:"repoUrl"`
	Readme             string `json:"readme"`
	Blurb              string `json:"blurb"`
	Hidden             bool   `json:"hidden"`
	Registered         bool   `json:"registered"`
	AlreadyRegistered  bool   `json:"alreadyRegistered"`
	BadgeInserted      bool   `json:"badgeInserted"`
	AlreadyBadged      bool   `json:"alreadyBadged"`
	DryRun             bool   `json:"dryRun"`
	BadgeMarkdown      string `json:"badge"`
	BadgeImageURL      string `json:"badgeImage"`
	BadgeDestination   string `json:"badgeLink"`
	RegistrationFailed string `json:"registrationError,omitempty"`
}

func execute(ctx context.Context, opts *options, stderr io.Writer) (*result, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	root, rootErr := git.FindRepoRoot(cwd)
	if rootErr != nil {
		if !errors.Is(rootErr, git.ErrNotRepository) {
			return nil, rootErr
		}
		root = ""
	}

	var slug normalize.Slug
	switch {
	case opts.repoInput != "":
		slug, err = normalize.Parse(opts.repoInput)
		if err != nil {
			return nil, err
		}
	default:
		if root == "" {
			return nil, git.ErrNotRepository
		}
		slug, err = git.DetectGithubSlug(ctx, root)
		if err != nil {
			return nil, err
		}
	}

	readmePath, err := resolveReadmePath(root, cwd, opts.readmeInput)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(readmePath)
	if err != nil {
		return nil, fmt.Errorf("unable to access README: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", readmePath)
	}

	rawContent, err := os.ReadFile(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read README: %w", err)
	}

	content := string(rawContent)
	res := &result{
		Repo:             slug.String(),
		RepoURL:          slug.RepoURL(),
		Readme:           readmePath,
		Hidden:           opts.hidden,
		DryRun:           opts.dryRun,
		BadgeImageURL:    readme.BadgeImageURL,
		BadgeDestination: readme.BadgeLinkURL,
	}

	if readme.HasBadge(content) {
		res.AlreadyBadged = true
		return res, nil
	}

	var blurbText string
	if opts.blurbProvided {
		blurbText, err = blurb.Normalize(opts.blurb)
		if err != nil {
			return nil, err
		}
	} else {
		blurbText = blurb.Random()
		fmt.Fprintf(stderr, "No blurb provided.\nUsing default speed claim: \"%s\".\n", blurbText)
	}
	res.Blurb = blurbText

	badge := readme.BuildBadgeMarkdown(slug.Encoded())
	res.BadgeMarkdown = badge

	if opts.dryRun {
		return res, nil
	}

	apiBase := strings.TrimSpace(os.Getenv(apiBaseEnv))
	client := api.NewClient(apiBase, nil)

	if err := registerRepo(ctx, client, slug, opts, res); err != nil {
		if !opts.forceBadge {
			return nil, err
		}
		res.RegistrationFailed = err.Error()
		fmt.Fprintf(stderr, "Warning: registration failed (%s). Continuing due to --force-badge.\n", err)
	}

	updated, err := readme.InsertBadge(content, badge)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(readmePath, []byte(updated), info.Mode()); err != nil {
		return nil, fmt.Errorf("failed to update README: %w", err)
	}

	res.BadgeInserted = true

	return res, nil
}

func registerRepo(ctx context.Context, client *api.Client, slug normalize.Slug, opts *options, res *result) error {
	submission := api.Submission{
		RepoURL:         slug.RepoURL(),
		IsBlazinglyFast: true,
		Blurb:           res.Blurb,
		Hidden:          opts.hidden,
	}

	if _, err := client.Submit(ctx, submission); err != nil {
		if errors.Is(err, api.ErrAlreadyRegistered) {
			res.AlreadyRegistered = true
			return nil
		}
		return err
	}

	res.Registered = true
	return nil
}

func resolveReadmePath(root, cwd, override string) (string, error) {
	if override != "" {
		if filepath.IsAbs(override) {
			return override, nil
		}
		base := root
		if base == "" {
			base = cwd
		}
		return filepath.Join(base, override), nil
	}

	if root == "" {
		return "", fmt.Errorf("cannot locate README outside a git repo; pass --readme")
	}

	return readme.FindDefault(root)
}

func printSummary(stdout io.Writer, res *result) {
	switch {
	case res.AlreadyRegistered && !res.Registered:
		fmt.Fprintf(stdout, "Repo %s already registered. Badge inserted.\n", res.Repo)
	case res.RegistrationFailed != "":
		fmt.Fprintf(stdout, "Registration failed (%s). Badge inserted.\n", res.RegistrationFailed)
	default:
		fmt.Fprintf(stdout, "Registered %s with blurb: \"%s\"\n", res.Repo, res.Blurb)
	}
	fmt.Fprintf(stdout, "Badge added to %s\n", res.Readme)
}

func emitResult(res *result, jsonOut bool, stdout io.Writer) {
	if jsonOut {
		_ = json.NewEncoder(stdout).Encode(res)
		return
	}

	switch {
	case res.DryRun:
		fmt.Fprintf(stdout, "Dry run: would register %s and update %s\n", res.Repo, res.Readme)
	case res.AlreadyBadged:
		fmt.Fprintln(stdout, "Already badged. No changes.")
	default:
		printSummary(stdout, res)
	}
}

func emitError(err error, jsonOut bool, stdout, stderr io.Writer) {
	if jsonOut {
		payload := map[string]string{"error": err.Error()}
		_ = json.NewEncoder(stdout).Encode(payload)
		return
	}

	fmt.Fprintf(stderr, "Error: %v\n", err)
}
