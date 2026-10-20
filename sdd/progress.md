# Implementation progress

## Current checkpoint

- Completed slice: confirmed permanent wish deletion, including attached claims.
- Checkpoint commit subject: `feat(wishes): add confirmed permanent deletion`. Locate the completed change by this unique subject in Git history.
- Previous checkpoint: `4e465a7` — wish creation and claim-preserving editing.
- Access, WLS-001 through WLS-004, WSH-001, and WSH-002 are Implemented. WSH-003 and Claims remain unimplemented.
- Milestone 4 continues with sorting. The Wishes domain is still In progress.
- Existing untracked `ideas.md` was present at resumption. It remains untouched and excluded from this commit.

## Selected criteria and completed behavior

Selected before implementation: WSH-002-AC1/AC2/AC4/AC5/AC6/AC7 (deletion portions). Creation/editing and claim preservation were verified in the previous checkpoint.

- Parents request deletion from an active person's current-year list. A separate server-rendered page names the wish and explains permanent removal of its attached claims.
- The initial request writes nothing. Cancel returns to the list without changing files. Explicit confirmation removes the entire wish mapping and all attached claims, retaining other wishes and their order/content. The last deletion leaves an empty list. No notification side effects exist.
- Fresh code validation, family denial, current-year form/query checks, household year rollover, archive/missing-person guards, and invalid-data protection apply to confirmation and deletion.
- Stale forms, changed claims, reordered wishes, missing tokens, and repeated deletion are rejected before writing.
- A failing regression exposed identical-wish index reuse on replay. Tokens now bind each position to the whole list's content version. Changes anywhere on a list invalidate its open edit/deletion forms; reload to continue. This preserves existing claims and prevents shifted identical wishes from being removed by repeated submissions.
- Confirmation reuses the dark theme with keyboard-accessible controls and a clearly labeled Cancel link. Family and historical screens omit deletion controls.

## Verification

- Baseline `go test -race ./...` passed.
- RED before implementation: deletion confirmation and guard tests failed on the missing route (404 instead of confirmation/stale/invalid-data responses).
- Additional RED: `TestRepeatedDeletionDoesNotRemoveIdenticalNextWish` reproduced a second removal (303 instead of 409). Whole-list fingerprints fixed the replay; the regression now passes.
- Go coverage verifies confirmation without writes, cancellation without writes, rejection of unconfirmed values, claimed/unclaimed removal, other-wish/claim preservation, final empty lists, and all listed access/data/stale guards.
- GREEN: `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, and `npm run build:css` passed.
- `npm run test:e2e`: all ten independent Chromium scenarios passed, including actual confirmation/cancellation, claimed and unclaimed removal, stale confirmation, family/history restrictions, unchanged historical files, and mobile theme/overflow checks.
- Manually reviewed the mobile confirmation screenshot under ignored `test-results/delete-wish-wish-deletion--05f54--and-preserves-other-wishes/delete-confirmation-mobile.png`.
- `git diff --check` passed. No dependencies or deployed services changed.

## Exact next action

Finish milestone 4 with WSH-003 sorting. Read the approved requirement and constraints, record AC1–AC8 before changes, and write failing browser tests for default added order, every alternate order, numeric comparisons, missing-as-zero, stable ties, case/accent equivalence, and browser persistence across reloads and different lists. Implement browser-stored global sorting without rewriting saved order. Verify family/current/history screens and the shared dark-theme controls. Mark Wishes Implemented only after every sorting criterion is fully verified.

Then begin milestone 5, Claims: browser claimant names, toggles, duplicate/ownership rules, role-specific visibility, stale submission handling, and historical read-only behavior. Seeded YAML already verifies claim preservation/removal during wish management.

## Remaining work and limitations

- Selectable/browser-persisted sorting; current display follows saved added order.
- Browser claimant names, claim toggles, counts, and role-specific claim visibility/removal. Claims remain preserved on disk; their presentation/actions are pending.
- Docker packaging and complete-flow deployment verification.
- Historical lists remain read-only in the app; direct historical YAML changes remain accepted.
- Stale-form fingerprints are implementation details, not account controls. The app serializes web writes; manual external edits are not filesystem-lock coordinated.
- No product OPEN question blocks the next approved slice.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes. Use Conventional Commits and commit completed, verified changes only.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Browser tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
