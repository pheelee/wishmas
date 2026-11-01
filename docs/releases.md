# Dependency maintenance and releases

The first release is **v0.1.0**. The application is published as a multi-platform image at `ghcr.io/pheelee/wishmas` for `linux/amd64` and `linux/arm64`. Product URL importing is deferred to [v2 issue #1](https://github.com/pheelee/wishmas/issues/1).

## Required GitHub setup

Push the reviewed source/workflows to `main` before running them. Local verification does not install repository rules, grant App permissions, publish a release or deploy the app.

1. Install the existing automation GitHub App on `pheelee/wishmas`. Make repository variable `GHA_IRBE_AUTOMATION_APP_ID` (App client ID) and secret `GHA_IRBE_AUTOMATION_APP_PRIVATE_KEY` available to this repository. Do not paste the private key into an issue or commit it.
2. The Renovate token requests Contents, Pull requests, Issues and Workflows **write**, and Checks/Commit statuses **read**. Release finalization uses a separate short-lived token from that App requesting Contents **write**. Renovate uses an App token so its changes trigger ordinary PR CI. No untrusted PR job receives an App secret.
3. Enable repository **Allow auto-merge** and configure a `main` protection/ruleset requiring the **Go tests** check and an up-to-date branch before PR merge. Choose a merge strategy allowed by the repository. Do not configure mandatory human review for grouped minor/patch updates unless you also intend to disable their unattended merging. Major updates do not auto-merge.
4. Permit the automation App's release workflow to push the validated version commit/tag directly to `main` under the repository's chosen protection/ruleset policy. Reconcile this with required PR checks explicitly; direct release commits are validated by the release workflow itself. Do not remove the Go test requirement for ordinary PRs. Test the installed policy before relying on unattended merges/releases.
5. Allow Actions to publish GitHub releases and packages. The release job's `GITHUB_TOKEN` has Contents/Packages write; GHCR must allow this repository to create/update `pheelee/wishmas`. Existing packages may need **Manage Actions access** granted to the repository. The package should link to this source repository through its OCI source label.
6. If anonymous image pulls are intended, set the GHCR package's visibility to **public**. GitHub repository visibility alone does not establish package visibility. Otherwise deployments must authenticate with appropriate package-read credentials.

The connector/local workflow validation cannot establish that these remote settings, secrets or permissions have been installed. Verify them in GitHub before the first run. Action invocations are pinned to stable release commit digests with version comments and tracked by Renovate.

## README badges and coverage

CI collects Go statement coverage during its existing race-test invocation. PRs validate the generator and report with read-only repository permissions. After successful main tests, a separate job publishes `coverage.svg` and `coverage.md` to the `badges` branch using the workflow token and a normal non-force push. No external coverage account or minimum threshold is required. The report links its source commit and excludes frontend, browser and container coverage. Coverage weights every instrumented profile block by its statement count; this can differ from the named-function total in `go tool cover -func`.

The coverage badge becomes available after the first successful main publication. Repository Actions policy must permit the main-only job’s `contents: write` permission and creation/update of `badges`; source `main` rules remain unchanged. A publishing failure keeps the prior badge/report and fails the publication job. Inspect CI before treating a cached badge as current.

The release badge uses Shields/GitHub release metadata. Image download counts use the community [GHCR badge provider](https://github.com/eliasbenb/ghcr-badge), not Docker Hub statistics. External badge availability and caching are controlled by those providers; missing releases/packages or an unavailable provider may produce an unavailable badge rather than a fabricated count.

## PR checks and Renovate

The `CI` workflow runs the **Go tests** check on every PR targeting `main`, and on pushes to `main`. It downloads/verifies dependencies, rejects a non-tidy module graph, runs vet and Go tests with the race detector. Release-script unit tests run in the same required job. PRs do not run browser/Docker checks or publish releases/images.

The `Renovate` workflow runs Mondays at **04:17 UTC** (`17 4 * * 1`) and can be dispatched manually. GitHub schedules may run later than their nominal time. It updates direct/indirect Go modules, npm dependencies/lockfiles, the Go builder/test-helper images, the pinned Traefik test image and GitHub Actions.

Minor/patch and pin/digest maintenance share one combined PR. Renovate requests GitHub-native auto-merge and checks must pass under the configured repository rules. This allows merging after CI completes without waiting for next Monday's Renovate run. Major upgrades are separate PRs without auto-merge. Dependency maintenance does not bump the application version or publish images.

Test paths are explicitly included in extraction: the recommended Renovate preset otherwise ignores `tests/`, which would hide the helper/proxy image references. Private data, ignored research and build artifacts stay excluded.

## Create a release

Use **Actions → Release → Run workflow**, select branch **main**, then choose **patch**, **minor** or **major**. Only a maintainer-authorized manual dispatch on this repository/main publishes a release. There is no release on an ordinary PR, dependency update or branch push.

On the first run, all three choices publish **0.1.0**, without incrementing it. Thereafter:

| Current version | Selection | Next version |
| --- | --- | --- |
| 0.1.0 | patch | 0.1.1 |
| 0.1.1 | minor | 0.2.0 |
| 0.2.3 | major | 1.0.0 |

The package version and both root lockfile version records must agree with the last stable tag. The workflow fetches full history/tags, verifies the previous tag has a published stable release/image, rejects inconsistent/non-main/dirty sources, and prepares the new version in the runner without publishing it. Version scripts are unit-tested using disposable local Git repositories, including atomic pushes and a simulated concurrent main update.

Release validation runs Go integrity/tidy/vet/race checks, release-script tests, a clean npm install/frontend rebuild, a byte-for-byte check of embedded CSS, the full Chromium suite, Compose validation and one focused English container smoke check. The smoke uses the already-built amd64 image behind Traefik HTTPS, completes setup/password login, creates and claims a wish, and verifies unchanged wish/claim/access data after restart. It uses disposable data and no real shop image. All fixture data is disposable. The browser binary build fixture has a bounded three-minute timeout for cold Go/codec compilation; page assertions are unchanged.

The workflow commits the version as `chore(release): vX.Y.Z` and tags it locally. Docker cross-compiles static Go-only images on the native builder for both target architectures, including the offline helper. OCI labels identify version, source repository and revision. Both architectures are build-verified; the **amd64 image** is loaded from the OCI archive and runtime-tested. **ARM64 is not runtime-tested** by this workflow; no ARM hardware/QEMU execution is implied.

After all checks, a short-lived App token publishes the version commit and tag atomically. A remote-main comparison and exact lease reject concurrent updates instead of overwriting them. The workflow then publishes the **same tested multi-platform OCI archive**, without rebuilding, using Skopeo and GHCR credentials created by Docker's login action. It adds these aliases:

| GitHub release/tag | Docker tags |
| --- | --- |
| v0.1.0 | `0.1.0`, `0.1`, `0`, `latest` |
| v0.1.1 | `0.1.1`, `0.1`, `0`, `latest` |
| v0.2.0 | `0.2.0`, `0.2`, `0`, `latest` |
| v1.0.0 | `1.0.0`, `1.0`, `1`, `latest` |

Only after image publication/aliases succeed does the workflow create a draft GitHub release with GitHub-generated notes, then publish it. This follows `pheelee/traefik-admin`'s generated changes/contributors/compare-link style. Workflow runs are serialized and an existing release or immutable `X.Y.Z` image tag fails preflight. Version tags are immutable by workflow policy; GHCR tags are not intrinsically immutable against other writers outside this workflow.

Deployments can select an explicit image version rather than a moving alias:

```sh
docker pull ghcr.io/pheelee/wishmas:0.1.0
```

Keep the existing persistent data volume, runtime UID/GID and private Traefik HTTPS configuration. Publishing a release does not restart the VPS app. See [deployment/rollback](review-and-deployment.md) and [S3 storage](s3-storage.md).

## Optional detailed container checks

Run **Actions → Detailed container checks → Run workflow** to exercise the comprehensive English/German container scenarios. This read-only manual workflow builds its own local test image and publishes no release or registry image. It is separate from the release gate.

The detailed suite retains image optimization/access checks, container replacement, claimant privacy, link regeneration, logout/session checks, proxy/IP rate limits and mobile layout. Unlike the focused release smoke, it uses a public shop image by default; local runs can override `WISHMAS_SMOKE_PICTURE`. Its optional S3/outage and custom UID/GID scenarios remain available through the existing local environment options.

The release gate checks fewer container scenarios and runs one language; the full browser and Go suites remain mandatory. Detailed container verification can be run independently when investigating deployment changes.

## Failure recovery

Git, GHCR and GitHub releases do not share an atomic transaction. A failure can leave a tag/version commit, an image, some updated aliases or a draft. The workflow does not claim success until the final release step completes. Inspect the failed run before rerunning; do not delete or replace published release artifacts.

- **Before ref publication:** no source commit/tag was pushed and no image was published. Resolve the failure and dispatch again from current main. If main moved during checks, the source must be revalidated.
- **After ref publication, before any immutable image exists:** the version commit/tag may exist without an image/release. The next dispatch refuses an incomplete previous release; it is not a same-version recovery mode. Inspect the failed source/tag and perform a maintainer-controlled recovery of that exact validated source, or explicitly reconcile the unpublished refs before selecting a new version. Do not pretend its test state is preserved across arbitrary edits.
- **After immutable image publication:** never rebuild/overwrite that version. Verify both platform manifests/source labels, then either finish missing aliases/release notes from that image/tag or publish a new version. A fresh normal dispatch does not repair the old run automatically and refuses to advance from an incomplete prior release.
- **Draft release exists:** inspect it and any assets/image tags. Finish publication only after validating delivery, or remove only an incomplete unpublished draft as a deliberate maintainer action. An existing release blocks the normal workflow's preflight.
- **API/permission errors:** preflight fails on authentication/permission/service errors. GitHub API 404 is treated as missing release/package; check token/repository/package access when diagnosing a surprising 404. Do not treat failure as a successful deployment or assume package absence from an access error.

The normal release workflow does not force-overwrite published version tags/images, roll back an app's data or automatically recover partial cross-service publication.

## Local verification

```sh
scripts/check-go.sh
python3 -B -m unittest discover -s tests/release -v
npm ci
npm run build:css
npm run test:e2e
node tests/container/compose.cjs
# Run against an already-built local amd64 image:
WISHMAS_SMOKE_IMAGE=wishmas:local node tests/container/release-smoke.cjs
# Optional comprehensive suite (builds locally if no image is supplied):
node tests/container/smoke.cjs
```

Docker smoke checks require Docker/Compose and Playwright Chromium. `WISHMAS_COMPOSE_BIN` can select a standalone Compose executable. Do not run release `prepare`/`commit`/`push` commands casually in an unfinished worktree: they intentionally reject it. Tests provide disposable repositories for exercising those commands safely.
