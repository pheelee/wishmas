# Implementation progress

## Current checkpoint

- Milestone 1: complete for setup and household metadata persistence.
- Milestone 2: regeneration/calendar slice verified; actual list/year scopes remain pending milestone 3.
- All four domains' requirements and OPEN-review updates are approved. ACC-001 is Implemented. Access as a domain remains In progress; ACC-002 and ACC-003 require list-scope integration evidence.
- Last verified code commit: `542ffc4` — `feat: add household setup wizard with YAML persistence`.
- Current verified code checkpoint: pending commit with subject `feat: add independent access regeneration and calendar verification`. Resolve its hash from Git history if resuming before finalization.
- No product question currently blocks the next slice.

## Completed behavior

- Open household setup, trimmed 3–50 Unicode-letter/digit household names, timezone selection, generated parent/family links with copy buttons, explicit completion into parent access, and setup closure.
- Persisted YAML metadata at `<data>/households/<name>/household.yaml`; completed setup survives restart. Direct edits are picked up on page load; invalid data shows an error and is left untouched.
- Parent-only independent regeneration of either access code. Current links remain copyable after setup. Regeneration saves the replacement before redirecting; parent regeneration opens its replacement automatically.
- Immediate old-code rejection for reading and actions, including a form submitted from an already-open page. Concurrent parent rotations using the same old code produce one success and one revoked action.
- Family pages expose no parent controls or access-link fields.
- Current-year calculation in the household timezone. Deterministic verification crosses the Zurich New Year while still December 31 UTC; a manual timezone edit changes the computed year on the next page load.
- Parent/family codes remain valid across year changes and a ten-year clock advance, without automatic expiry or file modifications.

## Verification and evidence

### Previous slice

- Setup behavioral tests failed against a placeholder HTTP handler before implementation.
- Go race tests, vet, build, CSS compilation, Chromium setup/copy flow, and a compiled-process termination/restart smoke check passed.
- Dependency audit reported zero vulnerabilities. Runtime dependencies have not changed in the current slice.

### Current slice

- Baseline `go test -race ./...`, `go vet ./...`, and `npm run test:e2e` passed before changes.
- RED: `go test ./internal/app -run 'Test(FamilyCode|ParentRegeneration|FamilyAnd|CalendarYear)' -count=1` failed. Regeneration returned 404; the calendar test failed at the injected household-year boundary. This was recorded before implementing regeneration and connecting the clock.
- GREEN: the same focused access command passed.
- `go test -race ./...` — passed, including five new access tests for persistence, authorization, stale/concurrent actions, and calendar/code lifetime.
- `go vet ./...` — passed.
- `go build -o /tmp/wishmas-check ./cmd/wishmas` — passed.
- `npm run build:css` — passed; regenerated CSS is committed.
- `npm run test:e2e` — passed (one full Chromium lifecycle: setup, both real clipboard copies, parent/family access, independent rotations, replacement link copying, automatic parent redirect, and old-page form rejection without reload).
- Reviewed the mobile parent-access screenshot. Screenshots are reproducible under ignored `test-results/`.
- `git diff --check` — passed before checkpoint.

## Exact next action

Start milestone 3 by reading `sdd/wishlists.md`. Select a first people/list slice and record its criterion IDs before editing code. Write failing behavioral tests for parent-only person creation, empty current-year Christmas list creation, and confirmed Unicode name validation/uniqueness (including reserved archived names), then implement with anchors.

As actual current and historical lists become available, add integration tests for ACC-002-AC2/AC3/AC4 and ACC-003-AC2: all-person scope, parent history, current-year-only family access, correct list selection at the year boundary, and preserved scope after regeneration. Keep these criteria partial until those tests pass.

## Remaining work and limits

- Access is not fully Implemented: usable current/historical lists and their scopes are not present yet.
- Household metadata is persisted; year/event/person list files are not implemented.
- Valid links currently open a household landing page. Parents additionally see access-management controls.
- Setup previews live in memory until Finish setup. An interrupted unfinished preview must be restarted; completed setup persists.
- People, lists, wishes, claims, Docker packaging, and complete-flow verification remain.
- No remote deployment was made. Isolated browser servers have stopped; local runtime data is ignored by Git.

## Resume

Read AGENTS.md and implementation.md, inspect Git status and checkpoint history, run `go test -race ./...` and `npm run test:e2e`, and continue from the exact next action. On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium` first. Committed generated CSS permits Go-only startup without Node.js.
