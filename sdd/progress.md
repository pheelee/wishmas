# Implementation progress

## Current checkpoint

- Completed slice: user-approved UI polish across setup, appearance, access controls, and wishlist/person editing.
- Checkpoint commit subject: `feat(ui): polish setup, decorations, and modal wish cards`. Locate the completed change by this unique subject in Git history.
- Previous checkpoint: `6e16648` — read-only claim visibility.
- Access (including ACC-004), Wishlists (including WLS-005), Wishes (including WSH-004), and CON-003-AC1–AC10 are Implemented and verified. Claims remains In progress; no whole Claims requirement is marked Implemented.
- User-owned `ideas.md` contains deferred manual-review notes. It remains untouched and uncommitted.

## Approved criteria and completed behavior

The user requested spec changes before application changes. Confirmed behavior was recorded in the owning specs, slider bounds/persistence and modal scope were clarified one question at a time, and the user approved the complete acceptance criteria before code.

Selected: revised CON-003-AC3/AC4 and AC6–AC10; ACC-004-AC1–AC4; WSH-004-AC1–AC5; WLS-005-AC1/AC2. Existing behavior remains covered by regression checks.

- Setup prefills an empty timezone from the browser and preserves explicit choices on validation errors.
- Live setup sliders: bulb size 80–200%, default 120%; snow size 50–200%, default 100%; snow amount 8–48 particles, default 16. Bulb count scales inversely to size using the original 18-bulb baseline, rounded to whole bulbs.
- Appearance settings persist under `appearance` in household YAML and render for everyone, including on restart. Existing households without that section receive defaults; invalid manual values show an error without writes. Link rotation preserves settings.
- Snow is always present without pause/resume controls; obsolete stored pause preferences are ignored. Reduced motion displays static snow. Opacity/duration limits and margin-only placement are retained with the wider layout and maximum slider settings.
- Access-link fields have inline clipboard and regeneration icon buttons. Successful copying shows a checkmark for three seconds; repeated clicks restart the timer. Existing clipboard fallback/error feedback remains. Regeneration opens a modal and retains independent rotation/immediate revocation behavior.
- Native modal forms handle wish editing/deletion confirmation, person rename/archive, and link regeneration. Cancel/Escape dismiss without writes; native focus return works. Validation errors reopen the relevant modal and retain entered drafts. Creation forms remain inline, including when an edit-validation modal is open.
- Main pages at/above the Tailwind md breakpoint use 75% viewport width. Wishes appear in responsive individual cards; edit/delete icons sit inline at the bottom. Whole cards retain their forms and sort targets. Family/history permissions remain enforced.

## Verification

- Baseline `go test -race ./...` passed.
- RED: new appearance tests failed on absent persisted settings and accepted out-of-range values. Initial browser polish tests failed on the old snow button and absent timezone prefill.
- GREEN: Go coverage verifies slider bounds/defaults, persistence across restart, legacy defaults, invalid manual settings without rewrites, and preserved appearance during code rotation.
- Initial full browser run passed fifteen scenarios; deletion stopped on an outdated unscoped text assertion matching multiple dialogs. Scoped it to the active dialog; focused deletion/polish rerun passed all three scenarios.
- Final `npm run test:e2e`: all sixteen Chromium scenarios passed. Coverage includes live previews, default 20% enlargement, real clipboard and timed icon feedback, keyboard activation, modal cancellation/Escape/focus return, error drafts, creation forms, 75% width, multi-card layouts, independent regeneration/stale codes, claim privacy, historical restrictions, sorting and attached edit/delete targets, static reduced motion, and mobile overflow/contrast checks.
- `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, `npm run build:css`, and `git diff --check` passed.
- Reviewed ignored screenshots: `test-results/polish-wish-cards-use-icon-bb599-out-changing-cancelled-data/polish-wishlist-desktop.png`, `polish-wishlist-mobile.png`, and `test-results/polish-setup-previews-hous-8bc9e-ts-them-for-family-browsers/polish-access-mobile.png`. Updated the historical design-reference README to point to current requirements.
- No dependency or deployment changes. All review/interview work is consolidated into this completed implementation checkpoint; no intermediate commit was created.

## Exact next action

Resume milestone 5, Claims, unless the user requests further UI work. Read `sdd/claims.md`, Access/Wishlists/Wishes criteria, constraints, and this checkpoint. Select and record criteria before code, then write failing tests for browser identity and current-year claim toggles: prompt only on first claim, trimmed/validated remembered names, future-only renaming, both roles claiming, multiple claimants, duplicates, and own-claim removal without confirmation.

Ownership comparisons ignore case and surrounding spaces only. Do not expose claimant names to family responses when implementing claimed-by-you. Preserve household-year/archive/code guards, historical read-only behavior, refreshed inline errors without changing existing claims, and parent-only removal of other people's claims. Complete CLM-003-AC3/AC4 and actual mutation integration in AC6 before marking that requirement Implemented.

## Remaining work and limitations

- Claims browser names, claimed-by-you, toggles, duplicate/ownership rules, parent removal, and stale/error handling. Current UI displays saved status/counts and parent-only names; actual claim actions are not implemented.
- Multi-viewer creation/removal refresh integration; persisted manual claim changes already verify page-load refresh without polling.
- Docker packaging and complete-flow deployment verification.
- Appearance sliders are available during first-run setup. Existing households receive defaults or may set the approved values manually in YAML; no additional post-setup appearance editor was requested.
- With browser storage disabled, sorting works for the open page only. Wish form fingerprints are implementation details, not accounts; web writes are serialized while external manual edits are not filesystem-lock coordinated.
- Historical YAML edits remain an accepted risk. No product OPEN question blocks the next approved slice.

## Resume

Read AGENTS.md, implementation.md, relevant approved requirements, and this checkpoint. Inspect Git status/history before editing; do not overwrite uncommitted work. Use Conventional Commits and commit completed, verified changes only. Checkpoint identifiers use unique commit subjects to avoid self-referential hashes.

On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; verify `go test -race ./...` and `npm run test:e2e`. Tests use isolated processes/data and dynamic ports. Normal startup is `go run ./cmd/wishmas`; Node.js is not needed at runtime.
