# bfast (Blazingly Fast Badger) — CLI Plan (Distribution + Repo Detection)

This document reflects the **current, locked-in intent** for `bfast`, including:
- operating on the **current git repository by default**
- registering via the existing blazingly.fast API
- inserting the correct badge
- always submitting a blurb (explicit or auto-generated)
- and a **clear, maintainable distribution strategy**

`bfast` performs **registration + marking**, not validation.

---

## 1. Summary

### Default usage (current repo)

```bash
bfast
```

Behavior:
- Detect current git repo
- Infer GitHub `owner/repo` from remotes
- Early noop if badge already present
- Register repo via API (with blurb)
- Insert badge into README

### With explicit blurb

```bash
bfast -m "the fastest in town"
```

### Override repo (advanced / edge cases)

```bash
bfast --repo owner/repo
bfast https://github.com/owner/repo
```

---

## 2. Core Principles

- **Current directory is the default target**
- The **API is the source of truth**
- The **badge is the mark**
- The **blurb is the artifact**
- Repeated runs are safe and boring (idempotent)

---

## 3. API Contract (Source of Truth)

The CLI uses the same API as the web form.

```ts
export type SubmissionForm = {
  repoUrl: string;
  isBlazinglyFast: boolean;
  blurb: string;
  hidden: boolean;
};
```

### CLI mapping

- `repoUrl`
  - canonical GitHub URL: `https://github.com/<owner>/<repo>`
- `isBlazinglyFast`
  - always `true`
- `blurb`
  - `-m / --blurb` if provided
  - otherwise auto-generated (see §6)
- `hidden`
  - `false` by default
  - `true` if `--hidden` passed

### Error handling

- **Already registered** → non-fatal, continue to badge
- **Other API errors** → abort by default
- `--force-badge` → badge even if API fails

---

## 4. Repo Detection (Default Path)

### Git root detection
- Walk up from CWD to find `.git`
- Fail if not inside a git repo (unless `--repo` provided)

### GitHub repo inference
From `git remote -v`:
- Prefer `origin`
- Accept:
  - HTTPS: `https://github.com/owner/repo(.git)`
  - SSH: `git@github.com:owner/repo(.git)`
- Strip protocol, `.git`, trailing slashes

If:
- no GitHub remote found, or
- multiple ambiguous GitHub remotes

→ error with guidance:

```
Could not infer GitHub repo.
Use: bfast --repo owner/repo
```

### Overrides
- `--repo owner/repo`
- positional arg `owner/repo` or full GitHub URL

Overrides bypass remote detection.

---

## 5. Early Noop (Critical)

Before any API call:

1. Locate README
2. Scan for existing blazingly.fast badge:
   - contains `https://blazingly.fast/api/badge.svg`
   - OR alt text includes `blazingly fast` and link to `blazingly.fast`

If found:
- print: `Already badged. No changes.`
- exit 0
- **no API call**

---

## 6. Blurb Handling

A blurb is **always submitted**.

### Explicit blurb

```bash
bfast -m "the fastest in town"
```

Rules:
- trimmed
- max 128 chars
- empty-after-trim → error

### Auto-generated blurb (default)

If no `-m` is provided:
- select **one random deadpan blurb** from a built-in pool (20–50 entries)

Examples:
- “Declared blazingly fast by the author.”
- “Performance considered, benchmarks omitted.”
- “Fast enough for its intended use.”
- “Written as if speed matters.”

CLI output must disclose this:

```
No blurb provided.
Using default speed claim: "Performance considered, benchmarks omitted."
```

Randomization prevents Hall-of-Speed sameness.

---

## 7. Badge URL Generation

```md
[![blazingly fast](https://blazingly.fast/api/badge.svg?repo=<ENCODED_SLUG>)](https://blazingly.fast)
```

Where:
- `<ENCODED_SLUG>` = `encodeURIComponent(owner/repo)`

Must match website logic exactly.

---

## 8. README Insertion Rules

Placement:
1. Append to existing badge block near top
2. Else after title (`# `)
3. Else at top of file

Constraints:
- never duplicate
- never reorder existing badges
- minimal diff

---

## 9. Flags

- `-m, --blurb <text>`
- `--hidden`
- `--repo <owner/repo>`
- `--readme <path>`
- `--dry-run` (no API, no writes)
- `--force-badge`
- `--json`

---

## 10. Distribution & Installation (Important)

### Canonical source: GitHub Releases

- `github.com/arrno/bfast`
- Every release includes:
  - static binaries
  - checksums
  - release notes

### Supported platforms (v1)

- macOS: `amd64`, `arm64`
- Linux: `amd64`, `arm64`
- Windows: `amd64`

Binaries are statically linked (Go).

---

### Automated builds (GoReleaser)

Use **GoReleaser** as the single release mechanism:

- Cross-compiles all targets
- Produces `.tar.gz` / `.zip`
- Generates `SHA256SUMS`
- Publishes GitHub Release
- Updates package manager metadata

Triggered by git tags:
```bash
git tag v0.1.0
git push --tags
```

---

### Package Managers (Preferred)

#### Homebrew (macOS + Linux)

- Official tap repo: `arrno/homebrew-tap` (tapped as `arrno/tap`)
- Distributed as a cask managed by GoReleaser
- Install:

```bash
brew install --cask arrno/tap/bfast
```

#### Scoop (Windows)

- Official bucket: `arrno`
- Install:

```powershell
scoop bucket add arrno https://github.com/arrno/scoop
scoop install bfast
```

These two cover the majority of dev users with minimal maintenance.

---

### Direct Install Script (Optional but Useful)

Hosted at:

```
https://blazingly.fast/install.sh
```

Usage:

```bash
curl -fsSL https://blazingly.fast/install.sh | sh
```

Behavior:
- Detect OS + arch
- Download matching GitHub Release
- Verify checksum
- Install to `/usr/local/bin` or `$HOME/.local/bin`

---

## 11. Maintenance & Versioning

- Semantic-ish versioning: `vX.Y.Z`
- Backward compatibility for flags where possible
- No auto-update daemon
- Users update via:
  - `brew upgrade`
  - `scoop update`
  - re-running install script

---

## 12. Project Structure (Go)

```
bfast/
  cmd/bfast/main.go
  internal/
    cli/          # command wiring
    normalize/    # repo parsing
    git/          # repo + remote detection
    api/          # submission client
    blurb/        # defaults + randomizer
    readme/       # badge detection + insertion
```

---

## 13. Philosophy (Reaffirmed)

- The API records intent
- The badge marks the repo
- The blurb expresses posture
- The badger issues badges

Speed is asserted, not proven.

---

## 14. Non-Goals

- No benchmarks
- No scoring
- No rankings
- No performance promises

This is ceremony, not measurement.
