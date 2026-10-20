# Implementation progress

## Current checkpoint

- Completed slice: browser sorting with one preference across all wishlists.
- Checkpoint commit subject: `feat(wishes): add persistent browser sorting`. Locate the completed change by this unique subject in Git history.
- Previous checkpoint: `3c6db51` — confirmed permanent wish deletion.
- Access, Wishlists, and Wishes are Implemented. All WSH-001 through WSH-003 criteria are verified; milestone 4 is complete.
- Claims and deployment remain pending. No application is deployed by this checkpoint.
- The user clarified that untracked `ideas.md` contains deferred manual-review work notes. Ignore it for now; it remains untouched and uncommitted.

## Selected criteria and completed behavior

Selected before implementation: WSH-003-AC1–AC8, with parent/family/history access and dark-theme regression checks.

- Saved added order is the default. The Sort wishes menu offers newest first, title A–Z, price low–high, and price high–low.
- Numeric comparisons use exact integer hundredths, including amounts beyond JavaScript's safe Number range. Missing prices count as zero; equal prices use original added order.
- Base-sensitive title collation ignores case and accents, including decomposed accents. Equal title values use original added order even after another displayed sort order.
- One browser-stored preference applies on reload and across people, years, and both access roles. A fresh browser defaults to added order. Invalid/unavailable storage falls back safely; users can still sort the open page without storage.
- Only DOM display order changes. Saved YAML order/content stays unchanged, including historical files. Whole wish cards move together, preserving their form targets and actions.
- The native labeled select retains keyboard access and dark styling. Without JavaScript, saved added order remains visible and the inactive sort control stays hidden.

## Verification

- Baseline `go test -race ./...` passed.
- RED before implementation: both initial sorting browser scenarios failed because the Sort wishes control did not exist.
- GREEN: focused scenarios passed after implementation. Final `npm run test:e2e` passed all thirteen independent Chromium scenarios.
- Sorting coverage includes every order, numeric 2/10/100 and large values differing by one cent, missing/explicit zero ties, equal numeric values written with different decimal precision, case/accent-equivalent titles, reload/cross-list/cross-role/history preferences, unchanged YAML, fresh-browser defaults, invalid/unavailable storage, and editing/deleting after reordered display.
- `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, and `npm run build:css` passed.
- Shared dark-surface assertions now cover selects. Manually reviewed the mobile sorting screenshot under ignored `test-results/sorting-wish-sorting-uses--57479-a-shared-browser-preference/sorting-mobile.png`; controls/text fit without overflow.
- `git diff --check` passed. No dependencies or deployed services changed.

## Exact next action

Begin milestone 5, Claims. Read `sdd/claims.md`, relevant Access/Wishlists/Wishes criteria and constraints, and inspect Git status. Select and record the first complete claim slice before changing code. Write failing behavioral tests starting with a browser name prompt on the first claim, trimmed/validated remembered names, both roles claiming current-year wishes, multiple claimants, and name-based duplicate/ownership checks (case and surrounding spaces only).

Preserve the approved simple toggle, future-only browser renames, parent-only claimant-name visibility, family claim counts/claimed-by-you, page-load refresh, and historical read-only behavior. Do not expose claimant names to family clients. Any partially covered claim requirement must remain In progress. Seeded YAML already verifies claim preservation/removal during wish management.

## Remaining work and limitations

- Claims: remembered browser names, current-year toggles, duplicate/ownership rules, role-specific visibility/removal, and stale failure handling. Claims remain preserved on disk; their presentation/actions are pending.
- Docker packaging and complete-flow deployment verification.
- Browser sorting does not rewrite files. With storage disabled, the selected order lasts only for the open page. Other already-open pages read the shared preference when loaded again.
- Historical lists remain read-only in the app; direct historical YAML changes remain accepted.
- Wish form fingerprints are implementation details, not account controls. The app serializes web writes; manual external edits are not filesystem-lock coordinated.
- No product OPEN question blocks the next approved slice.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes. Use Conventional Commits and commit completed, verified changes only.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Browser tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
