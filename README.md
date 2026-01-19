# bfast

[![CI](https://github.com/arrno/bfast/actions/workflows/ci.yml/badge.svg)](https://github.com/arrno/bfast/actions/workflows/ci.yml)

`bfast` (the Blazingly Fast CLI tool) registers your repository with [blazingly.fast](https://blazingly.fast) and drops the official badge into your README.

## Installation

-   Homebrew (macOS/Linux): `brew install --cask arrno/tap/bfast`
-   Scoop (Windows):
    1. `scoop bucket add arrno https://github.com/arrno/scoop`
    2. `scoop install bfast`
-   Manual: download a release from [GitHub](https://github.com/arrno/bfast/releases) and place the binary on your `PATH`.

## Usage

Run `bfast` from your project directory and it will:

1. Find the git root and infer `owner/repo` from your GitHub remotes
2. Detect whether the README already has a badge (noop if true)
3. Register the repo with the API (always including a blurb)
4. Append the badge snippet following the existing badge block or heading

```bash
bfast                           # auto-detect repo and README
bfast -m "Fast enough for me"   # custom blurb
bfast --repo owner/repo         # override detection
```

Flags:

-   `-m, --blurb` – explicit speed claim (trimmed, max 128 chars)
-   `--hidden` – mark submission as hidden
-   `--repo` – `owner/repo` or GitHub URL override
-   `--readme` – custom README path
-   `--dry-run` – skip API/write and describe actions
-   `--force-badge` – insert badge even if the API fails
-   `--json` – emit machine-readable output

If no blurb is provided, the CLI picks a deadpan default and tells you which one it used.

### Environment

-   `BFAST_API_BASE_URL` (optional) – override the API host, useful when pointing at a local `blazingly-fast` instance.

## Distribution & Development

GoReleaser drives the distribution pipeline defined in `.goreleaser.yaml`:

```bash
goreleaser release --clean    # publish tagged release artifacts, brew + scoop metadata
goreleaser release --snapshot --skip=publish --clean   # dry run locally
```

brew and scoop metadata are updated automatically via `arrno/homebrew-tap` (casks) and `arrno/scoop` (bucket) when GoReleaser runs.

For day-to-day development:

```bash
go test ./...
go run ./cmd/bfast --dry-run
```

`bfast` is designed to be idempotent: already-badged READMEs remain untouched and repeated submissions stay boring.
