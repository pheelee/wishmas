# V1 dependency and release automation

Approved by the user on 2026-11-01. Owning constraints: CON-001 deployment, CON-002 backend and CON-010 maintenance/release delivery. Implementation status: dependency baseline complete; automation/release code locally verified. Required-check enforcement, scheduled/manual GitHub Actions, remote GHCR publication and first GitHub release remain unverified pending repository setup and an explicitly authorized source push/dispatch. Do not mark those external criteria fully Implemented from local evidence.

## Confirmed decisions

- Update all dependencies to the latest stable versions, including major upgrades and Go; exclude prereleases. Adapt code where necessary and verify retained behavior.
- Initialize frontend package version to 0.1.0. First release is v0.1.0, without incrementing it on that first workflow run.
- Every PR must pass Go tests before merge. Browser/Docker checks run before release, not on PRs.
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
- **REL-001-AC2:** Before publication, verify dependency integrity, Go tests/race/vet, rebuilt frontend assets, full browser suite and Docker/Compose smoke checks with disposable data. Build both linux/amd64 and linux/arm64. State explicitly which architectures are executed versus build-verified; never substitute an amd64-only image for the required multi-platform result.
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
