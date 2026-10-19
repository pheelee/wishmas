# Implementation progress

## Current checkpoint

- Completed slice: parent rename/archive and read-only history, plus manual restoration verification and final Access scope integration.
- Checkpoint commit subject: `feat(wishlists): add rename, archive, and history`. Locate this completed change by subject in Git history.
- Previous verified checkpoint: `2458111` — person creation/current-year lists.
- Access is Implemented with all criteria verified. WLS-002 (annual/history), WLS-003 (names/rename), and WLS-004 (archive/restore) are Implemented. WLS-001 stays In progress because writing wish details requires milestone 4.
- Milestone 3's person management/history behavior is complete; remaining wish-writing criteria explicitly depend on milestone 4. Existing dark-theme criteria remain verified.
- Conventional Commits and completed/verified changes only. No proposal or unfinished implementation is included.

## Completed behavior

- Parents manage active people from current-year list pages. Rename validates the confirmed Unicode names and active/archived reservations, preserves added order, and moves saved directories without rewriting wishes/claims/comments.
- Renames retain directories across all saved years and event folders. V1 still only displays Christmas. Existing target conflicts and invalid affected data are rejected before moves. Failed directory moves or roster saves roll earlier moves back.
- Archive retains all files, hides every year's lists from both roles, reserves the name, and prevents new annual lists. No restore control or route exists.
- Manual roster restoration reloads on page load and creates a missing empty current-year list while preserving existing current/past lists.
- Parent year navigation discovers saved Christmas lists for active people. Historical pages are read-only and preserve the selected year when returning to the list overview. Family links cannot open history or see its controls.
- Current-year context is checked on person mutations; historical/stale form submissions cannot change a historical context. Every action still validates the latest access code.
- Replacement links retain current/all-person scope and parent historical scope, while family year restrictions remain unchanged.
- Household setup now accepts all 50 Unicode letters without HTML UTF-16 truncation; server character validation remains authoritative.

## Verification

- Baseline Go race/vet and all six existing browser tests passed.
- RED before implementation: eight new Go behavior tests failed on absent rename/archive routes, unavailable history, unguarded historical form submission, invalid-history behavior, and lost historical scope after regeneration.
- Additional RED: the household browser test reproduced truncation of 50 non-BMP Unicode letters to 25. Removed incompatible HTML length attributes; server validation continues to enforce 3–50 Unicode characters.
- GREEN: focused lifecycle/history tests passed; final `go test -race ./...` passed, including actual filesystem permission failures for rollback after later directory moves and failed roster saving.
- `go vet ./...` and `go build -o /tmp/wishmas-check ./cmd/wishmas` — passed.
- `npm run build:css` — passed; generated CSS included.
- `npm run test:e2e` — all eight independent Chromium scenarios passed. New coverage includes parent history, read-only controls, manual history updates, duplicate rename validation, preserved historical claims/comments, archive visibility, manual restoration, and full Unicode household entry.
- Reviewed mobile history and management screenshots under ignored `test-results/`; dark styling, readable controls, and overflow checks pass.
- `git diff --check` — passed before checkpoint.
- No dependencies or deployed services changed. Isolated browser processes and data were cleaned up.

## Exact next action

Begin milestone 4, Wishes. Read `sdd/wishes.md` and related Claims requirements, select a complete first slice, and record its criterion IDs. Write failing tests for required/trimmed 5–150-character titles, optional ordered HTTP/HTTPS shop links, optional nonnegative numeric reference price (at most two decimal places, decimal point only), and optional note before implementing parent-only web creation/editing.

Retain historical read-only and archived-person guards on every new wish action. Complete WLS-001's adult/child list-writing criteria only after the relevant wish behavior is verified. Subsequent wish deletion must ask for confirmation and remove all attached claims, while edits preserve claims; use seeded YAML to test those rules before claim UI implementation.

## Remaining project work

- Full wish detail presentation and web CRUD/sorting, claim behavior, Docker packaging, and complete-flow verification remain.
- Current list pages display titles from manually edited YAML. Other wish-node fields/claims remain preserved on disk without full presentation or validation by the future Wishes/Claims implementation.
- Historical list content is read-only in the app; manual YAML edits remain allowed as confirmed.
- No product question currently blocks the next approved domain implementation.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing. On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Tests use isolated processes/data and dynamic local ports; history tests seed/edit their own YAML. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
