# Implementation progress

## Current checkpoint

- Current milestone: 1 — Setup and persistence. First slice verified; household-timezone calendar-boundary verification remains for the next Access slice.
- All four domain requirements and their OPEN-review updates are approved. Access is In progress; no requirement is marked Implemented.
- Last verified planning commit: `c7c8f28` — `docs: add implementation milestones and resume procedure`.
- Verified code checkpoint: pending commit with subject `feat: add household setup wizard with YAML persistence`. Resolve its hash using Git history if resuming before this entry is finalized.
- No unresolved product question blocks the next slice.

## Completed first slice

- ACC-001-AC1–AC4 and AC6–AC9: open wizard, household name and timezone entry, generated parent/family codes, copyable links, explicit finish into parent access, disabled wizard after setup, and household-name validation.
- Household names: trim surrounding spaces, 3–50 Unicode letters/digits, confirmed individually before implementation.
- CON-004 (household metadata portion): persisted YAML survives restart and manual edits are picked up on page load. Invalid YAML shows an error and is left untouched.
- ACC-002-AC1 and AC5: parent/family link entry without accounts or code entry; missing/invalid codes return 404 after setup.
- Local Go executable, embedded Tailwind CSS, vanilla JavaScript, initial Christmas styling, run instructions, and isolated browser-test tooling.

## Verification and evidence

- RED: `go test ./...` first failed to compile without an application constructor; after adding a placeholder 404 handler, runnable behavioral tests failed for the expected missing wizard, validation, persistence/reload, and assets. This occurred before implementation.
- GREEN: `go test -race ./...` — passed (six named Go tests, including validation subtests).
- `go vet ./...` — passed.
- `go build -o /tmp/wishmas-check ./cmd/wishmas` — passed.
- `npm run build:css` — passed.
- `npm run test:e2e` — passed (one Chromium scenario verifying setup, both real clipboard copies, parent redirect, family access, reload, and blocked setup afterward).
- Browser launch initially failed because system libraries were missing. `npx playwright install-deps chromium` resolved that environment issue; the rerun passed.
- Reviewed generated desktop and 390px mobile setup screenshots. They are reproducible test artifacts under ignored `test-results/`.
- Compiled-process smoke check — passed: complete setup, terminate the Go process, start another with the same data directory, and open the retained parent link.
- `npm audit --audit-level=high` — passed with zero reported vulnerabilities after updating build-tool dependencies. Dependency versions are locked.
- `git diff --check` — passed before checkpoint.

## Exact next action

Begin the next Access slice with a failing behavioral test for ACC-003-AC1–AC6: parent-only independent regeneration, replacement links, immediate old-code rejection, and parent-page auto-switch. Add deterministic clock-based tests for ACC-001-AC5 and ACC-002-AC4/AC6 before implementing the remaining calendar behavior. Record RED results, implement with anchors, and verify before the next checkpoint commit.

List/year access scope criteria ACC-002-AC2/AC3 cannot be fully verified until milestone 3 provides actual lists; keep them partial until then.

## Remaining work and limits

- ACC-001 remains In progress because its timezone/calendar-boundary criterion is only partially verified.
- Household metadata lives at `<data>/households/<name>/household.yaml`; year/event/person list files are not implemented yet.
- Valid access links currently open a household landing page, not usable wishlists.
- Setup previews live in memory until Finish setup. An interrupted, unfinished preview must be restarted; completed setup persists.
- People, lists, wishes, claims, code regeneration, full calendar tests, Docker packaging, and complete-flow verification remain.
- No remote deployment was made. The isolated test/smoke servers have stopped. Local runtime data is ignored by Git.

## Resume

Read AGENTS.md and implementation.md, inspect Git status and the checkpoint commit, run `go test -race ./...` and `npm run test:e2e`, and continue from the exact next action. On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium` first; generated CSS is committed so Go-only startup does not require Node.js.
