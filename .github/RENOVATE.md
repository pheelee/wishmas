# Weekly dependency maintenance

The Renovate workflow runs Mondays at 04:17 UTC and on manual dispatch. It uses the existing GitHub App credentials:

- Repository variable `GHA_IRBE_AUTOMATION_APP_ID` (the App client ID).
- Repository secret `GHA_IRBE_AUTOMATION_APP_PRIVATE_KEY`.

Install the App on this repository. The workflow requests Contents, Pull requests, Issues and Workflows write permissions, plus Checks and Commit statuses read. App-created update PRs trigger ordinary CI; no App key is passed to untrusted PR checks.

`renovate.json` tracks direct/indirect Go modules, npm packages, the Dockerfile, GitHub Actions and helper/proxy images embedded in `tests/container/smoke.cjs`. Actions are pinned to stable-release digests with version comments. Explicit ignore paths retain test-image extraction while excluding household data and ignored research.

Minor/patch and pin/digest updates share one combined PR with GitHub-native auto-merge. Major upgrades remain individual PRs requiring maintainer review, with auto-merge disabled. Updates use Conventional Commits and do not bump the application version or create releases.

Before relying on auto-merge, enable **Allow auto-merge**, select an allowed merge method and require the **Go tests** status check on `main`, with branches up to date. Keep the required check present for all PRs, including grouped dependencies. Do not add mandatory human review to that group unless unattended merging is intentionally disabled. Reconcile any direct release-commit App exemption with these PR requirements. Missing/failed checks must not permit merge.

The `CI` workflow runs Go dependency integrity/tidy/vet and race tests. Browser/Docker checks are reserved for releases. Local validation does not install or verify remote branch rules or App secrets; a real PR is required to observe enforcement after the workflows are pushed.

Local verification uses Renovate's actual config validator (Node 24 or another supported runtime):

```sh
npx --yes --package=renovate renovate-config-validator renovate.json
```

A disposable local-platform extraction fixture verified Dockerfile, Go, npm, GitHub Actions and both test-image references. This does not establish authenticated GitHub execution or that an update PR has actually merged.

## Test the configured App without updating dependencies

Dispatch the Renovate workflow with `dry_run=true`. It mints the same repository-scoped App token and runs Renovate's full dry-run mode: dependency lookups and intended actions are logged without creating update branches, PRs, issues or merges. A maintainer may dispatch this non-mutating mode from a verification branch; ordinary updates remain restricted to main. Scheduled/default manual runs retain normal update behavior.

```sh
gh workflow run renovate.yml --repo pheelee/wishmas --ref main -f dry_run=true
```

This proves authenticated setup/lookup behavior, not that required checks gate an actual auto-merge or that GHCR can publish a release. Do not dispatch the Release workflow merely to test credentials: it publishes a real version/image after validation.

## Observed setup verification

On 2026-11-01, [authenticated dry-run41314522155](https://github.com/pheelee/wishmas/actions/runs/41314522155) and [normal run41314902809](https://github.com/pheelee/wishmas/actions/runs/41314902809) succeeded. Renovate resolved 46 dependencies, created [Node26 major PR3](https://github.com/pheelee/wishmas/pull/3) and Dependency Dashboard4, and the App-created PR triggered [successful Go CI41315024219](https://github.com/pheelee/wishmas/actions/runs/41315024219). Major auto-merge was disabled. The required Go tests rule is configured; strict base freshness was not enabled in the observed ruleset. No minor/patch update was available to demonstrate a real automatic merge. Release/GHCR publication was not part of this setup test.
