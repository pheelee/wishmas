# Implementation progress

## Current checkpoint

- Completed slice: bulb attachment correction under CON-003-AC2/AC5/AC6.
- Checkpoint commit subject: `fix(ui): attach bulbs precisely to the cord`. Locate the completed change by this unique subject in Git history; commit identifier pending creation.
- Previous checkpoint: `362c73a` — Gift-wrap header alignment.
- Access, Wishlists, Wishes, and Claims are Implemented. Milestones 1–5 remain complete.
- No unfinished application work or blockers. Docker packaging and deployment/complete-flow verification remain milestone 6; no deployment occurred.
- Exact next action: read CON-001–CON-004 and the milestone 6 plan, then write a failing container smoke test before Docker packaging.
- User-owned `ideas.md` remains untouched and uncommitted. Unselected `design/card-proposals/` remains uncommitted.

## Selected criteria and completed behavior

Selected before code: CLM-001-AC1–AC7, CLM-002-AC1–AC6, remaining CLM-003-AC3/AC4/AC6 while retaining AC1/AC2/AC5, and CLM-004-AC1–AC7. Verified identity/creation, ownership/removal, and multi-viewer/privacy subflows before committing the coherent implementation.

- Browsing does not prompt for a name. First claiming opens an inline name form; validated, trimmed 3–50 Unicode-letter/digit names are remembered in a persistent HttpOnly SameSite cookie. There are no accounts or server identity records. Names remain displayed across reloads and both link roles.
- Change claim name opens a native modal. Name updates only change the browser cookie, including from parent history. Old claim names and files remain untouched. Invalid names retain the draft and error without replacing the saved name.
- Parents and family members use a direct claim/remove-own toggle without confirmation. Multiple names may claim one wish; case and surrounding spaces are ignored for duplicate/ownership checks, while accents remain significant. Matching names across browsers represent the same claimant.
- Renaming changes future claims and the claimed-by-you marker; it confers no removal permission for the old name. Posted name/claimant fields cannot override the current cookie during own removal.
- Parents receive claimant names and direct removal buttons for any current-year claim, including without first choosing their own name. Family responses include counts and their own ownership marker/name, without other claimant names.
- Real independent-browser claim changes appear on the next page load. No polling or notifications were added. Stale and duplicate attempts show refreshed inline status/errors without changing existing claims.
- Every mutation reloads current codes/year/active roster/list snapshot. Historical requests, year rollover, archived/missing people, revoked links, and non-POST requests are guarded. Real write failures preserve claims/files and return refreshed inline errors.
- Mutations preserve other wishes, notes, existing claims, and future fields. Shared YAML aliases are expanded in a copy so removing one claim cannot affect another wish; cyclic expansion fails without writes.
- Existing modal/card, sorting, dark-theme, and household appearance behavior remains verified. Claim actions stay direct per CLM-002/CLM-004; name editing follows the modal rule.

## Verification

- Baseline `go test -race ./...` passed.
- RED before application changes: new identity/action/guard tests failed because family current-year lists had no wish claim action. The initial helper dumped full HTML on that absence; tightened it to a concise missing-action failure before implementation.
- GREEN: focused identity/ownership/guard tests passed after implementation. Additional tests verify actual concurrent case-variant duplicates, accent significance, invalid name changes, parent historical POST denial, alias preservation/cycle failure, and real directory-permission write failure without data changes.
- `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, and `npm run build:css` passed. The subsequently added parent historical denial test also passed with race checks, alongside concurrency and filesystem failure tests.
- `npm run test:e2e`: all eighteen independent Chromium scenarios passed. New flows cover independent browsers with matching/different names, parent/family claims, remembered names and 50 non-BMP Unicode characters, validation and modal drafts, future-only rename, all removal paths, duplicate/stale refreshes, actual page-load visibility, untouched historical files, and family HTML privacy.
- Reviewed ignored `test-results/claims-browser-names-multi-2efdc-across-independent-browsers/claim-toggle-family-mobile.png`; long-name mobile overflow and shared dark-theme assertions also pass.
- Existing setup/access revocation, appearance previews, cards/modals, wish management/claims preservation, sorting/form targets, history, and archive regressions pass.
- `git diff --check` passed. No dependencies or deployment configuration changed.

## Exact next action

Begin milestone 6, Deployment and complete flows. Read `sdd/constraints.md`, the milestone plan, this checkpoint, and repository run/build instructions; inspect Git status. Confirm local container tooling availability before choosing verification commands. Record the Docker/VPS acceptance checks before code, then write a failing container/build smoke check and implement packaging for the existing Go server, embedded assets, and persistent YAML directory.

Verify setup, parent/family access, current-year person/wish/claim management, and persistence across a container restart using an explicitly mounted data volume. Document concrete local/self-hosted run instructions and the remaining external deployment step. Do not infer a VPS address, credentials, or public endpoint; no production target has been supplied. Preserve all approved behavior and do not add unconfirmed product features. No product OPEN question blocks local Docker packaging.

## Remaining work and limitations

- Docker packaging and final container/complete-flow verification; the application has not been deployed to a VPS.
- Browser identity is deliberately name-based and trust-based, per the approved spec. Clearing/expiring the browser cookie forgets the current name but does not change saved claims; choosing the same name restores matching ownership.
- Browser cookie storage must be available to remember a name. Existing sorting still works for the open page when local storage is unavailable.
- Appearance sliders are first-run setup controls. Existing households receive defaults or may set approved values manually in YAML; no post-setup appearance editor was requested.
- Web writes are serialized; manual external edits are not filesystem-lock coordinated. Historical manual YAML edits remain an accepted risk.
- Native modals and the responsive 75% desktop layout remain in place. The user's deferred review notes in `ideas.md` have not been tackled.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Use Conventional Commits and commit completed, verified changes only. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.

## Active task: selected Gift-wrap cards

User approved WSH-005-AC1–AC5 for implementation. The first browser run failed as expected: `.wish-gift-header` returned zero instead of three cards. The focused test now passes after implementing headers, badges, decorative accents, optional details and separated footers. No YAML schema change; accents are presentation-only and stay attached when sorted.

Baseline and updated `go test -race ./...`, `go vet ./...`, Go build and CSS build pass. All 19 Chromium scenarios pass, including claims, sorting, edit/delete modals, family privacy and history. Desktop/mobile screenshots reviewed; no overflow or unreadable labels. `git diff --check` passes. Checkpoint ready for commit. Keep `design/card-proposals/` uncommitted and `ideas.md` untouched.

## Active correction: Gift-wrap header alignment

User reported ownership badge touching the header edge and off-center gift icons. Reproduced with a read-only screenshot of the running app on port 8090. Correcting existing approved WSH-005-AC1/AC2/AC5: retain interior spacing around badges and center gift icons beside either one or two badges. Write failing geometry assertions before CSS changes; verify desktop/mobile and existing browser flows. No live YAML or running process changes.

Header correction verification: new geometry assertions failed first (insufficient bottom spacing); CSS build, all seven Gift-wrap/polish/theme Chromium scenarios, `go test ./...` and `git diff --check` now pass. Reviewed a screenshot using corrected CSS against live data via a browser-only override. Completed correction ready for checkpoint commit; the live Go process embeds the previous CSS and needs restarting to serve this change. No YAML or live process modifications. Deployment remains the next milestone.

## Active correction: bulb attachment

User reported bulbs offset from the cord. Inspecting existing CON-003-AC2/AC5/AC6; test actual stem-tip alignment with the SVG cord at desktop/mobile sizes and all slider extremes before correcting the decoration geometry. Preserve steady colors, brightness, counts and household settings. No changes to user notes or live data.

Bulb correction complete: failing-first test measured 2.07px stem/cord separation. Corrected cord coordinates, centered stems and cord-based rotation pivots. Seven theme/polish browser tests, subsequent precise geometry rerun, CSS build, `go test ./...` and diff checks pass. Reviewed live-data screenshot with browser-only CSS/SVG overrides. Commit ready; no live process or YAML changes. Restart the app to serve the new embedded assets. Next action remains milestone 6 Docker packaging.
