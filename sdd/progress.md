# Implementation progress

## Current checkpoint

- Completed slice: parent-managed people creation and current-year Christmas list viewing.
- Code checkpoint: the commit titled `feat(wishlists): add people and current-year lists` contains this verified slice and its records. Locate by subject in Git history.
- Previous verified code checkpoint: `46fdfa8` — December-night styling and gentle snowfall, after consolidating the intermediate design history at the user's request.
- Milestone 3 is In progress. Creation/current-year viewing is complete; app rename/archive controls and historical navigation remain separate work.
- ACC-001 and the existing-screen theme criteria remain Implemented. No whole Wishlists requirement is marked Implemented: several include wish editing, renaming, or history behavior not present yet.
- Conventional Commits and completed/verified changes only. This checkpoint contains no design proposal or unfinished implementation.

## Completed behavior

- Parents can add an adult or child by name. Family, invalid, and revoked links cannot add people; family pages omit the form.
- Names are trimmed, require 3–50 Unicode letters/digits, and are compared ignoring case across active and archived people. Concurrent duplicate submissions cannot create two people.
- People are saved in added order in `<household>/people.yaml`. Empty lists are persisted at `<household>/<year>/christmas/<name>/wishlist.yaml` and remain available after recreating the server.
- Parent/family pages display all active people's current lists in added order. Individual list pages show an empty state or titles manually entered in YAML.
- Direct roster/list edits reload on page load. Invalid YAML, invalid names, and duplicate active/archived names show an error without file changes. Existing manual list files are never overwritten during creation.
- On a valid page load, active people with no current-year list receive an empty one. Deterministic New Year tests verify no wishes/claims are copied, old files remain untouched, and archived people receive no new list.
- Family past-year list requests are rejected. Replacement parent/family codes retain all active current lists and family year restrictions; replaced codes lose list access.
- Dark styling and snowfall remain functional on new pages; desktop/mobile screens were inspected.

## Verification

- Baseline `go test -race ./...`, `go vet ./...`, and all five existing browser tests passed before implementation.
- RED: runnable people behavior tests failed on 404 creation, missing parent form/manual people rendering, and absent new-year list files. A test variable-scope typo was corrected before recording those behavior failures.
- Additional RED: a real browser truncated a valid 50-character non-BMP Unicode name to 25 characters because HTML maxlength counts UTF-16 units. Removed browser length limits from the person field; the server enforces the confirmed Unicode character bounds. The regression test now passes.
- GREEN: focused people tests passed; `go test -race ./...` passed after implementation and the Unicode fix.
- `go vet ./...` and `go build -o /tmp/wishmas-check ./cmd/wishmas` — passed.
- `npm run build:css` — passed; generated CSS included.
- `npm run test:e2e` — all six Chromium scenarios passed, including the new complete parent creation/family viewing flow, inline validation, full Unicode entry, mobile/desktop layouts, and existing setup/copy/regeneration/theme/snow regressions.
- Final focused race checks for New Year and replacement-link list scopes passed after adding the integration assertions.
- Reviewed the actual people desktop/mobile screenshots under ignored `test-results/`.
- `git diff --check` — passed before commit.
- No dependencies or deployed services changed. Isolated browser processes and temporary data were cleaned up.

## Exact next action

Continue milestone 3 with app renaming and archiving plus parent history navigation. Read `sdd/wishlists.md` and `sdd/access.md` and record the next criteria subset before editing. Write failing tests for WLS-003-AC3 (preserve list data/claims while renaming), WLS-004 (archive all years, reserve names, skip annual creation, manual-only restoration), and WLS-002-AC3–AC5 (parent history and read-only historical UI).

Verify ACC-002-AC2 and the historical portion of ACC-003-AC2 against actual past lists and regenerated parent links. Keep these partial until that integration passes. Current family scope and actual New Year selection are already covered.

## Remaining project work

- Rename/archive/history operations, full wish-field presentation and web editing, claims, Docker packaging, and full-flow verification remain future work.
- Current list pages read manually entered titles only. Other wish-node data is preserved without rewriting but is not yet presented or fully validated by the Wishes domain implementation.
- Parent historical navigation is not present yet; both roles currently open current-year lists only. Retained historical YAML is unchanged.
- Setup previews are held in memory until Finish setup; completed setup persists in `<data>/households/<directory>/household.yaml`.
- No product question currently blocks the completed slice or the recorded next work.

## Resume

Read AGENTS.md, implementation.md, the relevant approved domain requirements, and this checkpoint. Inspect Git status/history before changes. On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; then verify `go test -race ./...` and `npm run test:e2e`. Each browser scenario gets an isolated app and dynamic local port. Normal startup is `go run ./cmd/wishmas`; compiled CSS is tracked, so Node.js is not needed at runtime.
