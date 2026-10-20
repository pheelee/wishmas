# Implementation progress

## Current checkpoint

- Completed slice: parent wish creation/editing and full optional detail presentation.
- Checkpoint commit subject: `feat(wishes): add creation and claim-preserving editing`. Locate the completed change by this unique subject in Git history.
- Previous checkpoint: `4d5b31b` — person lifecycle and history.
- Access and WLS-001 through WLS-004 are Implemented. WSH-001 is Implemented. WSH-002 remains In progress because deletion is pending; WSH-003 and Claims remain unimplemented.
- Milestone 4 continues. This checkpoint includes completed creation/editing behavior only; it does not claim completion of the Wishes domain.

## Selected criteria and completed behavior

Selected before implementation: WSH-001-AC1–AC11; WSH-002-AC2/AC3/AC5 for creation/editing and creation/editing portions of AC1/AC7. These also complete WLS-001-AC3/AC4 and the wish-writing portion of AC6.

- Parent-only current-year creation/editing on any active person's list. Family routes, historical form contexts, and archived targets are rejected.
- Title-only saves, trimmed 5–150-character Unicode titles, optional multiple ordered shop links, one numeric reference price, and optional notes. Optional details may be added later or cleared.
- Complete HTTP/HTTPS links display as domain labels with original URLs and saved order. Price validation accepts nonnegative decimal amounts with at most two decimal places and rejects commas, negative amounts, and exponent notation. Decimal text is retained without floating-point rounding.
- Editing only replaces title/links/price/note fields on freshly loaded YAML mappings. Multiple existing claims and future fields are preserved. No notification integration exists.
- Position/content fingerprints reject stale or reordered edit forms with a reload message. Invalid files and denied submissions leave data untouched. Manual YAML changes reload with detail validation.
- New screens retain the approved dark theme. Textareas have explicit accessible names, dark surfaces, and keyboard focus styles. Validation preserves entered drafts.

## Verification

- Baseline `go test ./...` passed.
- RED before implementation: `TestWishCreateEdit` and `TestWishValidationAndGuards` failed with 404 on the absent save endpoint.
- GREEN: creation/editing, trim/Unicode boundaries, optional-detail validation, ordered link display, clearing fields, preserved multiple claims/future fields, stale edits, family denial, historical contexts, archived targets, and invalid-data preservation pass in Go tests.
- Initial browser run: eight existing scenarios passed; the new scenario caught an accessible-name issue for prefilled textareas. Explicit textarea names corrected it, and the focused browser flow passed.
- Final checks: `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, and `npm run build:css` passed. All nine Chromium scenarios passed, including creation, validation correction, retained drafts, editing, family presentation, and mobile dark-theme/overflow checks.
- Manually reviewed `test-results/wishes-parents-create-and--5d556-tails-while-families-browse/wishes-mobile.png` (ignored artifact).
- `git diff --check` passed. No dependencies or deployed services changed.

## Exact next action

Continue milestone 4 with confirmed permanent wish deletion: read WSH-002 and related Claims criteria; record selected AC1/AC2/AC4/AC5/AC6/AC7 before changes. Write failing tests for confirmation and cancellation, deletion with multiple attached claims, family denial, and historical/archived/stale guards. Implement and verify the complete deletion slice before committing. Then implement WSH-003 browser-persisted global sorting with numeric prices, missing-as-zero, accent/case-insensitive titles, and stable added-order ties.

Do not mark WSH-002 Implemented until deletion and every other criterion are verified. Claims UI remains milestone 5; seeded YAML already verifies claim preservation during editing.

## Remaining work and limitations

- Wish deletion and selectable/browser-persisted sorting.
- Browser claimant names, claim toggles, counts and role-specific claim visibility/removal.
- Docker packaging and complete-flow deployment verification.
- Current wish display follows saved added order. Claims remain preserved on disk, with presentation/actions pending the Claims milestone.
- Historical lists remain read-only in the app. Direct historical YAML changes remain accepted.
- Internal stale-form fingerprints are implementation details, not identity/account controls. The app serializes web writes; manual external edits are not filesystem-lock coordinated.
- No product OPEN question blocks the next approved slice.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes. Use Conventional Commits and commit completed, verified changes only.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Browser tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
