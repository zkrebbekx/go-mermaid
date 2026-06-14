# Releasing

This project uses [Semantic Versioning](https://semver.org) and automates
releases with [release-please](https://github.com/googleapis/release-please)
and [GoReleaser](https://goreleaser.com).

## Versioning policy

| Phase | Tags | Rules |
| --- | --- | --- |
| Pre-1.0 (now) | `v0.x.y` | No API stability promise. `feat:` → minor, `fix:` → patch. Breaking changes allowed in minors. |
| Stable | `v1.0.0`+ | Public API frozen. Breaking changes require a major bump. |
| Post-v1 breaking | `v2.0.0`+ | **Go requirement:** the module path must gain a `/v2` suffix (`github.com/zkrebbekx/go-mermaid/v2`) and imports must update. |

Go modules have no registry to publish to — a release *is* a git tag.
`pkg.go.dev` and the module proxy index a version the first time it is fetched.

## How a release happens (automated)

1. Commits land on `main` using [Conventional Commits](https://www.conventionalcommits.org)
   (`feat:`, `fix:`, `docs:`, etc.).
2. **release-please** (`.github/workflows/release.yml`) maintains an open
   "release PR" that bumps the version in `.release-please-manifest.json` and
   updates `CHANGELOG.md` based on those commits.
3. **Merge the release PR** when you want to cut a version. release-please then
   creates the git tag and a GitHub Release.
4. The same workflow runs **GoReleaser**, which cross-compiles the `mermaid`
   CLI (linux/darwin/windows × amd64/arm64), attaches archives + checksums to
   the release (`release.mode: append`), pushes a multi-arch Docker image to
   `ghcr.io/zkrebbekx/go-mermaid`, and (when configured) updates the Homebrew
   cask.

No manual tagging needed. Configuration:
`release-please-config.json`, `.release-please-manifest.json`, `.goreleaser.yaml`.

## Distribution artifacts

| Artifact | Where | Setup needed |
| --- | --- | --- |
| Archives + checksums | GitHub Release assets | none (uses `GITHUB_TOKEN`) |
| Docker image | `ghcr.io/zkrebbekx/go-mermaid:{version, latest}` | none — `packages: write` + ghcr login are wired in the workflow |
| Homebrew cask | `zkrebbekx/homebrew-tap` | one-time, see below |

### Enabling the Homebrew cask

The cask step **auto-skips** until a tap token is present, so it never blocks a
release. To turn it on, once:

1. Create a public repo named **`homebrew-tap`** under your account
   (`gh repo create zkrebbekx/homebrew-tap --public`).
2. Create a PAT with `contents: write` (classic: `repo`) scope on that tap repo.
3. Add it as a secret on this repo:
   ```bash
   gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo zkrebbekx/go-mermaid
   ```

The next release publishes `Casks/mermaid.rb`, installable with
`brew install zkrebbekx/tap/mermaid`. (The cask ships an unsigned binary; if
Gatekeeper quarantines it, `xattr -dr com.apple.quarantine $(which mermaid)`.)

> Note: the release PR is created with the default `GITHUB_TOKEN`, so its CI
> may not auto-run (GitHub suppresses workflow events for token-created PRs).
> Merge it with admin privileges, or wire a PAT into the release workflow if
> you want required checks to run on the release PR itself.

### Manual fallback

If you ever need to tag by hand:

```bash
git tag v0.1.0
git push origin v0.1.0
```

You would then run `goreleaser release --clean` locally (or temporarily add a
tag trigger) since the automated path keys off release-please.

### Validating before a release

Every PR runs `goreleaser release --snapshot --clean` (the `release-check` job
in CI), which validates `.goreleaser.yaml` and builds the binaries without
publishing. Release-config breakage is caught there, not at tag time.

## AI code review (optional)

`.github/workflows/claude-review.yml` runs the official `/code-review` skill on
each PR via the [Claude Code GitHub Action](https://github.com/anthropics/claude-code-action).
It is **inert until you add an auth secret** (the step is guarded), so it never
blocks CI.

To enable it with a **Claude Pro/Max subscription** (no separate API key):

1. In a terminal with Claude Code installed, run:
   ```bash
   claude setup-token
   ```
   This produces a long-lived OAuth token tied to your subscription.
2. Add it as a repository secret named `CLAUDE_CODE_OAUTH_TOKEN`
   (Settings → Secrets and variables → Actions), or:
   ```bash
   gh secret set CLAUDE_CODE_OAUTH_TOKEN --repo zkrebbekx/go-mermaid
   ```
3. (For `@claude` mentions in issues/PRs) install the Claude GitHub App:
   <https://github.com/apps/claude>, or run `/install-github-app` from Claude
   Code. The automated review uses the default `GITHUB_TOKEN` to post comments,
   so the app is only needed for interactive mentions.

Alternative: use a pay-as-you-go API key from <https://console.anthropic.com>
as `ANTHROPIC_API_KEY` and swap the input in the workflow. The subscription
token draws on your Pro/Max usage limits; heavy CI review can exhaust them, in
which case the metered API key is the more predictable option.
