# Implementation progress

## Current checkpoint

- Completed slice: read-only claim status/counts and parent-only claimant names.
- Checkpoint commit subject: `feat(claims): add role-safe claim visibility`. Locate the completed change by this unique subject in Git history.
- Previous checkpoint: `faa5e5a` — persistent browser sorting and completed Wishes milestone.
- Access, Wishlists, and Wishes remain Implemented. Claims is In progress; no whole Claims requirement is marked Implemented.
- User-owned `ideas.md` contains deferred manual-review work notes. Ignore it for now; it remains untouched and uncommitted.

## Selected criteria and completed behavior

Selected before implementation: CLM-003-AC1/AC2/AC5/AC6, with CON-004 validation and existing wish-management preservation. Read-only visibility was chosen as a privacy foundation before browser identity/toggles. AC6 remains partly verified until actual claim mutation integration exists.

- CLM-003-AC1/AC2/AC5 are verified: every saved wish displays claimed/unclaimed status and an explicit zero/singular/plural count. Only parents receive a claimant-name list. Full family HTML responses contain no claimant names.
- Parent history displays saved claim names/status. Family access remains limited to current-year lists.
- Manual persisted claim changes appear on the next page load. Open pages remain unchanged without polling. This verifies the presentation part of AC6; another viewer's actual claim creation/removal remains pending.
- Manual claim data requires a valid claim sequence and trimmed 3–50 Unicode-letter/digit names. Per-wish duplicates ignore case and surrounding spaces. Invalid data produces a generic error and leaves files untouched.
- Presentation uses a decoded copy; raw saved claim names, future fields, and existing claims remain preserved during wish edits. All claim mutation controls remain pending.
- Status/counts and parent-name lists reuse dark styling and fit mobile screens.

## Verification

- Baseline `go test -race ./...` passed.
- RED before implementation: visibility tests failed on absent status/names; invalid claim-data tests returned 200 instead of 500.
- One preservation assertion initially compared literal non-BMP characters in marshaled YAML. YAML validly escapes those characters; the assertion now compares decoded claim data and retained future fields instead of encoding representation.
- GREEN: `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, and `npm run build:css` passed.
- `npm run test:e2e`: all fourteen independent Chromium scenarios passed. The new scenario checks full family HTML privacy, parent names, zero/singular/plural counts, unchanged files/open pages, reload behavior, read-only history, invalid-data failure, and mobile dark styling.
- Go coverage includes malformed sequences/claims, blank/invalid/short/long names, case/space duplicates and Unicode case-fold duplicates, valid 50-character non-BMP names, future-field preservation during wish edits, and no writes on views/errors.
- Reviewed ignored `test-results/claim-visibility-claim-vis-8097b-refreshes-only-on-page-load/claims-family-mobile.png` and `claims-parent-mobile.png`.
- `git diff --check` passed. No dependencies or deployed services changed.

## Exact next action

Continue milestone 5 with browser identity and current-year claim toggles. Read `sdd/claims.md`, Access/Wishlists/Wishes criteria and constraints. Select and record criteria before code, then write failing tests for a name prompt only on first claim, trimmed/validated browser-persisted names, displayed/current-name changes affecting future claims only, parent/family claiming, multiple claimants, duplicates, and own-claim removal without confirmation.

Ownership comparisons ignore case and surrounding spaces only. Renaming must not permit removal under the old name. Do not expose claimant names in family responses: claimed-by-you must derive from the current browser name without sending all names to family clients. Preserve household-year/archive/code guards, no historical mutations, refreshed inline errors without changing existing claims, and parent-only removal of other people's claims. Complete CLM-003-AC3/AC4 and actual mutation integration in AC6 before marking that requirement Implemented.

## Remaining work and limitations

- Claims browser names, claimed-by-you, toggles, duplicate/ownership rules, parent removal, and stale/error handling.
- Actual multi-viewer creation/removal refresh integration; only persisted manual changes have verified page-load refresh so far.
- Docker packaging and complete-flow deployment verification.
- Browser sorting is complete and does not rewrite files. With storage disabled, sorting lasts for the open page only; other open pages read a shared preference on load.
- Historical lists remain read-only in the app; direct historical YAML changes remain accepted.
- Wish form fingerprints are implementation details, not account controls. The app serializes web writes; manual external edits are not filesystem-lock coordinated.
- No product OPEN question blocks the next approved slice.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes. Use Conventional Commits and commit completed, verified changes only.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Browser tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
