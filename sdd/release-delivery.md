# V1 dependency and release automation

Approved by the user on 2026-11-01. Owning constraints: CON-001 deployment, CON-002 backend and CON-010 maintenance/release delivery. Implementation status: dependency baseline complete; automation/release code locally verified. Source publication/hosted CI, required Go check configuration and authenticated dry-run/normal Renovate runs are now verified. Actual minor/patch auto-merge, direct release-main writes, remote GHCR publication and first GitHub release remain unverified pending their respective real operations. Do not mark those external criteria fully Implemented from local evidence.

## Confirmed decisions

- Update all dependencies to the latest stable versions, including major upgrades and Go; exclude prereleases. Adapt code where necessary and verify retained behavior.
- Initialize frontend package version to 0.1.0. First release is v0.1.0, without incrementing it on that first workflow run.
- Every PR must pass Go tests before merge. Full browser checks run before release, not on PRs. Container/Compose test automation is removed by the approved REL-003 revision.
- Renovate tracks Go modules, npm packages, Docker images and GitHub Actions. All minor/patch updates share one PR across managers, with auto-merge after required checks pass. Major updates open separately and require manual review.
- Weekly Renovate workflow: Monday 04:17 UTC, with manual dispatch. Reuse GHA_IRBE_AUTOMATION_APP_ID and GHA_IRBE_AUTOMATION_APP_PRIVATE_KEY; the App needs repository access.
- Release workflow is manually dispatched with patch/minor/major bump choice. It directly commits the version bump to main, creates the tag/release and publishes the Docker image.
- Publish linux/amd64 and linux/arm64 to ghcr.io/pheelee/wishmas. Stable image tags are X.Y.Z, X.Y, X and latest; Git tag is vX.Y.Z.
- Release notes follow pheelee/traefik-admin's GitHub-generated notes style.

## Approved observable acceptance criteria

### Dependency refresh

- **DEP-001-AC1:** Record the stable versions resolved at implementation time for direct/indirect Go modules, npm dependencies/lockfile, Go toolchain and pinned Docker/test-image references. Update active references consistently, without prereleases or silent major-version exclusions. Record any unavailable/incompatible latest version as unresolved instead of claiming it updated.
- **DEP-001-AC2:** package.json, package-lock.json and its root-package version start at 0.1.0. Normal dependency maintenance does not create an application release or independently bump that version.
- **DEP-001-AC3:** Dependency refresh preserves approved app behavior. Go race tests/vet/module verification, frontend rebuild, all browser suites and isolated Docker checks pass; record actual platform/provider coverage and any limitations. Original household data, live apps and research work are preserved.

### PR validation and Renovate

- **AUT-001-AC1:** Every PR targeting main runs a named required Go test check. Failing tests prevent merge, including Renovate auto-merge. Use the same Go version as go.mod; dependency integrity/tidy and vet are also checked. Browser/Docker checks are absent from the PR workflow.
- **AUT-001-AC2:** Renovate discovers Go direct/indirect modules, npm manifests/lockfiles, Docker images (including test references) and GitHub Actions. New Actions references are digest-pinned with readable versions, and Renovate can update them.
- **AUT-001-AC3:** Minor/patch updates across the tracked managers use one combined PR, Conventional Commits and auto-merge gated on successful required checks. Major updates remain individual PRs with auto-merge disabled. No separate automatic review approval is required by the user's clarified choice; the requested behavior is auto-merge.
- **AUT-001-AC4:** Renovate runs at cron `17 4 * * 1` and on manual dispatch, using a repository-scoped GitHub App token from the confirmed credentials. Missing credentials fail clearly. Its PRs trigger ordinary PR checks; do not rely on default workflow-token writes that silently suppress downstream checks.
- **AUT-001-AC5:** Document/configure required repository controls and App permissions needed for checked auto-merges and direct release commits. Do not report remote enforcement or live Renovate execution as verified unless actually observed. Keep release authorization separate from arbitrary PR code.

### Version, release and GHCR publication

- **REL-001-AC1:** Manual release dispatch accepts only patch/minor/major and operates on main. With no previous release, it releases 0.1.0 unchanged. Thereafter patch increments Z, minor increments Y/resets Z, and major increments X/resets Y/Z, from the last consistent release version. Update both npm version records/lockfile and use a Conventional Commit for the version change.
- **REL-001-AC2:** Before publication, verify dependency integrity, Go tests/race/vet, rebuilt frontend assets and the full browser suite. Build linux/amd64 and linux/arm64 and statically verify OCI contents/metadata. No container runtime or Compose tests run in GitHub Actions. Both architectures are build/static-verified, not runtime-tested by release automation.
- **REL-001-AC3:** Commit/tag/version metadata, GitHub release and Docker image identify the same verified source/version. Serialize release runs and reject an invalid/duplicate version or changed source rather than silently tagging untested code or overwriting an existing version.
- **REL-001-AC4:** Publish ghcr.io/pheelee/wishmas with both platforms under X.Y.Z, X.Y, X and latest, and create Git tag/release vX.Y.Z. Use package/contents write permissions only where needed, appropriate GitHub credentials and source/version image metadata. Document registry/repository configuration required for user pulls.
- **REL-001-AC5:** Release notes use GitHub-generated changes/contributors/compare links in the style of the reference repo. Publish the GitHub release after validations and successful image publication. Failure does not falsely report a completed release; document how to inspect/retry partial tag/image/draft publication without replacing an already published immutable version. No cross-service atomicity is promised.
- **REL-001-AC6:** Workflow documentation explains first release, subsequent manual bump selection, GHCR tags/platforms, permissions, required checks, version consistency and failure recovery. Preparing/verifying the workflows does not run the first public release, push local app code or redeploy the live app without a later explicit instruction.

## Approved resumable implementation slices

1. Dependency refresh + version baseline: record selected criteria; write/run a failing behavioral dependency/version check, resolve actual latest stable versions, update code/references, run cumulative verification; commit only this completed verified slice.
2. PR checks + Renovate: test configuration/required gating before implementation, add workflows/configuration, validate discovery/grouping/schedule/token scope and repository setup requirements; commit only completed verified behavior. Record any external check requiring pushed workflows/App permissions as pending.
3. Release + multi-platform image delivery: test version calculation/first-release/error behavior before implementation, add dispatch/version/release/image automation, validate workflows and disposable builds/smoke checks, document remote-execution boundary; commit only completed verified behavior.
4. Final checkpoint: update owning constraints/implementation/progress/changes with actual evidence; report remote setup still required and exact first-release action. No production release/deployment is part of local verification.

Temporary artifacts and proposals remain under .work/. Permanent SDD documents contain only approved requirements and completed confirmed records. All permanent acceptance status remains unimplemented/partial until its evidence is complete.

## Verification record

- **DEP-001-AC1–AC3:** implemented/verified locally; dependency baseline in docs/dependencies.md, commit 23451c7. Go/module/static/frontend, 22 browser tests and bilingual container/Compose checks pass. Latest-stable Go unchanged; upstream graph-only x/net is documented rather than artificially added.
- **AUT-001-AC2:** local validator/extraction proves all configured managers, pinned Actions and two test-image references. **AUT-001-AC1/AC3–AC5:** workflow/configuration/docs complete; remote checks/auto-merge/App enforcement remain partial/unverified until GitHub setup and live PR observation. Configuration checkpoint ba78909.
- **REL-001:** local code/verification complete, including 13 disposable Git/preflight tests, actionlint, two-platform OCI builds, manifest/blob/label/non-root/ELF checks, amd64 bilingual runtime smoke and exact multi-platform index/digest preservation through a disposable registry under all four tags. ARM64 is build/static-verified, not runtime-executed. Negative version-metadata check fails as intended. Real GHCR/GitHub Actions/release publication remains unverified and no release was run.
- Operator instructions: docs/releases.md. First release/deployment and source push are explicitly separate actions. Partial publication does not advance to another normal bump until previous release/image consistency is restored.

### Authenticated setup follow-up

The user merged the source-publication checkpoint and requested testing. Full dry-run41314522155 and normal Renovate41314902809 passed with the configured App. Forty-six dependencies were resolved; normal mode created Node26 major PR3 and Dashboard4. App-created PR CI41315024219 passed automatically, and the major PR has no auto-merge request. The required Go tests rule is now configured (GitHub Actions integration15368, non-strict base freshness). No actual minor/patch auto-merge, major merge, release, registry publication or deployment was performed. Reusable optional manual dry-run support was tested without changing ordinary weekly/default-manual update behavior. These results advance AUT-001 runtime evidence without claiming the remaining REL publication criteria are complete.

## Approved README and coverage revision

Approved on 2026-11-01 after the user selected CI-generated coverage.

- **DOC-001-AC1:** README shows the app-themed lights/snowflakes hero and six real feature screenshots before features and run instructions; example captures contain fictional household data only.
- **DOC-001-AC2:** README links badges for latest GitHub release, Go statement coverage and GHCR image downloads. Downloads use a GHCR-specific provider; unavailable external metrics are not replaced with invented values.
- **COV-001-AC1:** The existing Go race-test invocation generates statement coverage in CI without a second Go test run or a coverage threshold. Coverage reports identify their source commit and scope (Go tests, excluding frontend/browser/container coverage).
- **COV-001-AC2:** Only successful main CI publishes the generated badge/report to the dedicated `badges` branch. PR jobs remain read-only. Publishing never changes source main, uses no external coverage account and uses ordinary non-force Git pushes.
- **COV-001-AC3:** Verify weighted statement coverage, malformed/empty profile errors, generated SVG/report, first publication and subsequent publication in disposable repositories. Failed tests prevent publication; workflows retain the required `Go tests` check.

Implementation slice: COV-001-AC1–AC3 and DOC-001-AC1/AC2. Hosted publication remains pending until the workflow is merged and main CI succeeds.

Local verification: COV-001 generation/publication code passes 18 Python tests, full Go dependency/vet/race checks with a coverage profile, actionlint and shell syntax checks. DOC-001 gallery/assets are locally verified. Successful-main-only configuration is verified locally; hosted COV-001-AC2 publication remains pending merge/main CI. Profile weighting includes all instrumented blocks; the Go named-function report may differ. No unavailable external metric is claimed verified.

### First hosted release validation repair

Release 41327636404 passed Go/frontend/browser and multi-platform OCI checks, then failed a Compose 2.38.2 JSON assertion before any remote publication. The explicit disabled bind-directory flag serializes differently from locally tested 5.6.1. The reference test now compares each CLI's normalized disabled form and ensures enabled/disabled forms differ. Both real versions pass; both reject enabled/omitted-flag mutations. REL-001-AC2 hosted Docker smoke/publication evidence remains pending a successful new release; no criteria or runtime configuration changed.


## Superseded focused release gate revision (REL-003 replaces REL-002)

The user approved simplifying the gate after release 41328726480 failed on a browser navigation race in the detailed container suite, before publication.

- **REL-002-AC1:** Keep dependency/vet/Go race tests, frontend build consistency, full browser tests, version/duplicate-publication/concurrent-main guards and both image architectures as release gates. Reference Compose validation stays in the release check.
- **REL-002-AC2:** Replace the extensive bilingual Docker release checks with one English, disposable-container scenario on the built amd64 image: start behind HTTPS, complete setup, authenticate with the configured password, create and claim a wish, restart the app and verify the wish/claim and access persist. Use no live household data or real shop image. Failures stop before remote publication. ARM64 remains build/static-verified.
- **REL-002-AC3:** Retain the detailed English/German container suite in a separate manually dispatched workflow, with read-only repository permissions and no release/image publication. Detailed scenarios are optional for release delivery; unit/full browser checks continue covering product behavior.
- **REL-002-AC4:** Await completed form navigation before inspecting updated pages; report useful failure details and clean disposable resources on success/failure. Document the reduced release coverage and manual detailed-check command/workflow.

Implementation slice REL-002-AC1–AC4 is approved. Hosted release completion remains unverified until a new successful main run.


### Focused gate verification

REL-002-AC1–AC4 are implemented and locally verified. The focused HTTPS/login/create/claim/restart check passes on a freshly built current-source scratch image; an unavailable-image control fails nonzero with stage/stack details. The retained detailed en-US/de-CH checks pass locally after waiting for the recipient form navigation. Compose checks pass under 2.38.2/5.6.1. All 18 Python release/coverage tests, actionlint, JS syntax and Renovate config validation pass. Core release build/test/version/publication steps are retained. Hosted execution of the new workflows and first release publication remain pending after merge; no release was dispatched or deployed by this slice.


## Approved container test removal — 2026-11-02

The user requested removing the whole Docker test suite. REL-003 supersedes REL-002's focused/manual container obligations. Historical verification records remain evidence of prior runs, not current automation requirements.

- **REL-003-AC1:** Remove container smoke/Compose test scripts and their dedicated manual workflow. Release automation neither starts test containers nor loads an image into the daemon for runtime verification.
- **REL-003-AC2:** Retain Go dependency/vet/race tests, full browser tests, frontend consistency, amd64/arm64 image builds/static OCI checks and every version/ref/image/publication guard. Keep Skopeo for publishing the verified archive; remove its runtime-test conversion/load steps.
- **REL-003-AC3:** Remove obsolete test-image dependency extraction and active documentation commands. Describe both architectures as build/static-verified with no release container runtime coverage. No application, data or deployment behavior changes.

Implementation slice REL-003-AC1–AC3 is authorized. Verify removal policy fails before implementation, then passes alongside existing release tests, workflow lint and Renovate validation. No release dispatch is implied.

REL-003 local verification: removal-policy test failed first on the existing suite and passes after removal. All 20 Python tests, actionlint and Renovate validation pass. Core tests/static image checks/publication guards remain in the workflow; neither architecture is runtime-tested. Hosted updated release/publication remains pending merge and manual dispatch.
