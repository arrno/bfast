# Release TODO

- [ ] Create the `arrno/homebrew-tap` repository (empty repo is fine); grant the release bot push access.
- [ ] Create the `arrno/scoop` repository; grant the same bot push access.
- [ ] Tag the CLI repo once ready for v0.1.0 (e.g., `git tag v0.1.0 && git push --tags`).
- [ ] Run `goreleaser release --clean` from the tagged commit with GitHub tokens for the tap/bucket.
- [ ] Verify the GitHub release at `https://github.com/arrno/bfast/releases` contains all archives and checksum file.
- [ ] Confirm `brew install arrno/tap/bfast` and `scoop install bfast` work on their platforms.
- [ ] Optionally add a GitHub Actions workflow to run GoReleaser automatically on tags.
