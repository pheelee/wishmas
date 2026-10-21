# Implementation progress

## Current checkpoint: household refactor milestone 1 — offline data and migration tools

- User approved milestone 1 after spec checkpoint `5bf2e9c`. Branch: `refactor/household-lists-access`; baseline main remains `1992bdf`.
- Selected before code: CON-004-AC1–AC4/AC6 data-core mechanics, CON-007-AC1–AC2 and CON-008-AC1–AC5 migration mechanics. Completed offline scope: stable-ID single-file model/validation, stale-safe atomic repository, salted Argon2id helper with private confirmation, copy-only migration with identity conversion, counts/checksums and rollback/source preservation.
- RED: missing migration command; then conflicting future-field encoding panic, symlink-following and merged-language fallback. GREEN: corrected and covered; malformed hashes/costs, name/ID/reference rules, multiple docs, invalid manual data, stale/concurrent saves, actual disk-write failure, preserved future data, migration rejection paths, year order, duplicate wishes, claimant case/space matching, accents and copied/missing pictures pass.
- Verification: `go test -race ./...` passes all packages; `go vet ./...`, `CGO_ENABLED=0 go build -tags nodynamic ./...`, focused permission test and `git diff --check` pass. Pseudo-terminal hash-helper smoke verifies no echo and usable output. Disposable CLI demo verifies counts/checksums/source bytes and existing-destination refusal. Browser tests not rerun: no HTTP/frontend code changed and no new browser behavior is claimed.
- Initial full-suite failure: legacy translation inventory expected the old spec rows. Preserved its independent 182-row approved baseline fixture and kept all catalog checks; no target translations are falsely reported implemented. Initial dependency resolution requested unavailable x/term v0.48.0; corrected to available v0.47.0. Pinned x/crypto v0.58.0 and x/term v0.47.0 support the repository Go toolchain.
- Checkpoint commit identifier: pending; resolve by unique subject `feat(storage): add household data core and offline migration tools` after commit.
- Review: [offline tool instructions](../docs/offline-admin.md). Ignored disposable demo/report and binary are under `.work/household-refactor/milestone-1/`; they contain public test fixtures only and are not deployment-ready households. The CLI operates on a household directory, requires a hash in the copied legacy configuration, and refuses any existing destination.
- Original `data/`, running app, main branch, user-owned `.gitignore` and `ideas.md` are untouched. No real data migrated, merge, cutover or app restart performed. No refactored domain is marked fully Implemented.
- **Exact next action: wait for user review/approval to start milestone 2.** Milestone 2 is setup/admin authentication, persistent sessions, logout/hash-change invalidation and Traefik HTTPS/IP limits. Current HTTP UI remains legacy. Actual current-data migration is milestone 6; do not run the legacy app against version-2 data.

## Previous checkpoint: household refactor milestone 0 — approved specification

- Branch: `refactor/household-lists-access`. Baseline `main`: `1992bdf`. No merge, code implementation, live app changes or original-data migration performed.
- Completed: user-approved behavioral criteria in all five domains; one-file stable-ID storage, Argon2id/password helper, hashed persistent sessions, Traefik HTTPS/client-IP and copy-only migration contracts; approved resumable milestones; 112 approved new/revised English/German entries plus retained wording.
- Updated specification index/glossary and retired obsolete parent/shared-family links, automatic year/timezone/history/archive model and browser-name identity. Previous IDs/evidence remain historical; no new domain is marked Implemented.
- Verification: documentation audit passed for 190 unique criterion IDs, 223 bilingual rows (including all 112 exact approved HR entries), matching placeholders, local Markdown targets and Swiss German spelling; `git diff --check` passes. No runtime tests were rerun: no application source/dependencies/assets changed and this checkpoint claims no new runtime behavior.
- Verified specification commit: `5bf2e9c` (`docs(sdd): specify household lists and password access refactor`).
- Implementation status: milestones 1–6 not started. No partially implemented feature or pending product/wording decision. Deployment-specific trusted Traefik addresses remain a deployment prerequisite.
- User-owned `.gitignore` modification is preserved and excluded from this checkpoint. Original `data/`, running app and `ideas.md` remain untouched. Temporary interview/proposal notes remain ignored and uncommitted; approved facts are now in `sdd/`.
- **Exact next action: wait for user review/approval to start milestone 1.** After approval, select CON-004/CON-007-AC1–AC2/CON-008 migration-mechanics criteria, write/run failing behavioral tests, then implement the data core/hash helper/explicit migration command on disposable fixtures. Do not migrate current data or start the next milestone prematurely.

## Historical baseline verification

The records below describe the old implementation, not the refactor target. Their prior next-action statements and obsolete requirement IDs are historical only; the current checkpoint and approved implementation plan control resumption.

## Baseline checkpoint: person-overview summaries

- Selected before code: user-approved WLS-006-AC1–AC6, badge draft B and all four localized phrases. Previous checkpoint: `3730977`.
- Completed: exact ungrouped two-decimal totals include incomplete priced wishes; green/yellow badges count complete claimed/unclaimed wishes once each. Both roles see displayed-year summaries with existing archive/history/privacy rules. Existing individual-wish price formatting is unchanged.
- RED: new HTTP summary tests failed on all five cases before implementation. GREEN: HTTP tests cover exact large totals, completeness, zero/missing prices, roles, locales, year scope, manual changes and unchanged YAML. Browser tests verify mobile wrapping, accessible links, claims, reloads and history.
- Verification: `go test -race ./...`, `go vet ./...`, `npm run build:css`, `git diff --check` pass. Full browser run: 26 passed, one German overview test hit a navigation race. Added a wait for the final created person link; `npx playwright test tests/browser/overview.spec.js --repeat-each=3` passed all six runs. All 27 distinct browser scenarios have passed. Earlier German remove-claim selector corrected to the existing translated label. Desktop/mobile screenshots reviewed.
- Temporary overview drafts removed after approval and verification. Live app/data, user notes and the user's `.gitignore` modification remain untouched. No dependency or storage-schema change.
- Verified baseline commit: `1992bdf` (`feat(wishlists): summarize prices and claims in person overview`).
- Unfinished work/blockers: none for this feature. Next action: await the next user-requested change. Restart the app to load embedded assets; no live restart was performed.

## Previous checkpoint: top-snow addition

- Selected before final implementation: CON-003-AC11–AC13. User approved the reviewed top-band snowfall and requested spec then implementation. The preceding card-status checkpoint is `5c488ef` (`feat(wishes): color cards by completeness and claim status`).
- Completed CON-003-AC11–AC13. The reviewed header snowfall now has production names and criterion anchors; it stays clipped to the lightbulb band with a bottom fade, shares household appearance/setup sliders, and respects reduced motion without intercepting input. Side snow and bulb geometry remain unchanged.
- RED: permanent header test failed against pre-preview frontend with zero particles. GREEN: verifies actual motion, desktop/mobile clipping/fade, inert decoration, separate counts under shared sliders, saved parent/family appearance and reduced motion across reload/preference restoration.
- Verification: `npm run test:e2e` (25 passed), `go test -race ./...`, `go vet ./...`, CSS build, JavaScript syntax and diff checks pass. Final desktop/mobile screenshots visually reviewed.
- Temporary preview names/scripts/screenshots/backups removed. Existing card-status implementation is preserved. No live app/data, user notes, dependencies or deployment configuration changed.
- Checkpoint commit subject: `feat(ui): add snowfall around Christmas lights`; locate this completed follow-up by its unique subject. The user approved the preview and requested final implementation after the separate card-color commit.
- Next action: await the next user-requested change. Restart the app to load updated embedded assets; the agent did not restart the user's app.

## Previous completed slice: status-based card accents

- Selected before code: user-approved WSH-005-AC6–AC7. Red for incomplete wishes, yellow for complete/unclaimed, green for complete/claimed; completeness requires title, provided price and at least one shop link. Apply to existing accents and retain dark backgrounds/ownership styling.
- Completed WSH-005-AC6–AC7. Accents now derive from wish completeness/claims rather than saved position. A zero price counts as provided; claimed incomplete wishes remain red. Dark backgrounds and ownership styling remain unchanged; no persisted schema or input-validation change.
- Previous checkpoint: `cb5242b`. The user authorized this completed card-color commit. Checkpoint commit subject: `feat(wishes): color cards by completeness and claim status`; locate it by its unique subject.
- RED: HTTP behavior tests failed on all nine status cases before implementation. GREEN: all cases pass for parent/family. Browser coverage verifies actual border/icon/price/button colors, dark backgrounds, mobile fit, history, sorting and color changes after claims or edits.
- Verified: `go test -race ./...`, `go vet ./...`, `npm run build:css`, `npm run test:e2e` (24 passed), `git diff --check`; desktop/mobile screenshots visually reviewed. No live files/processes or user notes changed.
- Card-color review and commit authorization are complete. No live app restart was performed.

## Completed slice: picture-modal animation

- Completed user-requested WSH-006-AC10 ease-in-out opening/closing: 180 ms fade/scale and backdrop fade, with shared animated Close/backdrop/Escape dismissal and reduced-motion bypass. Failing-first animation test now passes; all four English/German picture and polish scenarios, CSS build, JavaScript syntax and diff checks pass. The user reviewed and approved this completed work for commit; no live restart was performed.

## Completed slice: picture-modal backdrop dismissal

- Completed WSH-006-AC10 update: picture-modal background clicks close it; image/interior clicks keep it open. Browser test failed first, then all four English/German picture and polish scenarios passed, including desktop/mobile and Escape checks. JavaScript syntax/diff checks pass. The user reviewed and approved this completed work for commit; no live restart was performed.

## Previous checkpoint: optional wish pictures

- Completed: all approved WSH-006-AC1–AC15, CON-006 and the 17 approved image translation entries. Requirements and verification details are in `wishes.md`.
- Checkpoint commit subject: `feat(wishes): add optional optimized pictures and animated image modals`. The user reviewed the completed feature and explicitly authorized committing it; locate the checkpoint by its unique subject. Previous checkpoint: `4712cc1` (`feat(i18n): add household German and English localization`).
- Verified: normal and Go-only-backend Go race suites, vet, CSS build, JavaScript syntax/diff checks, all 24 Chromium scenarios, and English/German DinD smoke checks using the supplied real HTTPS image. Smoke checks cover optimized-image serving, permissions, named-volume persistence across restart/replacement and immediate removal. Temporary test resources were removed.
- Removed temporary prototypes/review files and kept only the selected shelf reference. Leon's sample now has the confirmed source URL and final 43,894-byte quality-80 WebP. Other wish data was verified unchanged. No live process was started/restarted and `ideas.md` was untouched.
- No unfinished approved implementation or OPEN product/wording decisions. Existing VPS deployment limitation remains. To review the new embedded UI in the user's running app, restart that app on the user's chosen port 8090; this restart was not performed by the agent.
- Next action: await the next user-requested change. Review of the picture feature and modal refinements is complete. Resume from this checkpoint and inspect Git status before editing.

## Previous checkpoint: internationalization

- Completed slice: all approved INT-001–INT-005 criteria and CON-005, implementing the approved 161-row German/English translation table.
- Checkpoint commit subject: `feat(i18n): add household German and English localization`. Locate this completed implementation by its unique subject in Git history.
- Previous checkpoint: `562ba95` — approved i18n requirements and wording.
- Access, Wishlists, Wishes, Claims and Internationalization are Implemented. Docker packaging and English/German DinD persistence/replacement smoke verification passed; actual VPS deployment has not been performed.
- No unfinished approved implementation or i18n OPEN decisions. Existing live app/data and user-owned notes remain untouched.
- Exact next action: restart/rebuild the app to load its newly embedded assets. Existing households remain English unless `language: de-CH` is added to `household.yaml`; new households choose during setup. Actual VPS deployment remains an external step when a target is supplied.

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

## Previous Claims checkpoint next action (historical)

Begin milestone 6, Deployment and complete flows. Read `sdd/constraints.md`, the milestone plan, this checkpoint, and repository run/build instructions; inspect Git status. Confirm local container tooling availability before choosing verification commands. Record the Docker/VPS acceptance checks before code, then write a failing container/build smoke check and implement packaging for the existing Go server, embedded assets, and persistent YAML directory.

Verify setup, parent/family access, current-year person/wish/claim management, and persistence across a container restart using an explicitly mounted data volume. Document concrete local/self-hosted run instructions and the remaining external deployment step. Do not infer a VPS address, credentials, or public endpoint; no production target has been supplied. Preserve all approved behavior and do not add unconfirmed product features. No product OPEN question blocks local Docker packaging.

## Remaining work and limitations

- All approved v1 implementation, including Internationalization, is complete. Docker/container flows are verified; actual VPS deployment remains unverified.
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

## Design cleanup

Removed the untracked proposal HTML, generator and variant screenshots at the user’s request. Retained both approved design references and updated the selected-design notes. Verified the temporary directory is absent and both approved baselines remain. No application behavior changed; tests not rerun for artifact cleanup. `ideas.md` remains untouched. Next action remains milestone 6: failing container smoke test, Docker packaging, persistence/restart checks and final parent/family complete-flow verification.

## Completed milestone 6 local verification

- Recorded CON-001-AC1–AC4 from the user-approved Docker/persistence/complete-flow plan before packaging. Temporary execution notes were kept separately under `.work/`; this checkpoint records completed facts.
- The first smoke test failed because no Dockerfile existed. The final multi-stage build produces a static Go server with embedded frontend/timezone data and non-root UID/GID 65532. `/data` is explicitly mounted; household data and notes do not enter the build context.
- Initial runtime checks exposed remote-daemon loopback addressing and changing ephemeral ports after restart; the test now discovers the daemon endpoint and rediscovers the published port before reconnecting.
- `node tests/container/smoke.cjs` passes: real first-run setup in Europe/Zurich, generated links, person/wish creation, family claim and parent visibility, privacy/management restrictions, embedded assets, persistent setup/links/claims after restart and recreation, and unchanged wishlist YAML bytes. Disposable container, volume and test image were cleaned up.
- `go test -race ./...`, `go vet ./...`, `npm run test:e2e` (all 20 Chromium scenarios), Docker build and `git diff --check` pass. README instructions cover self-hosted run, container replacement, mounted data ownership, direct YAML edits and backup.
- Actual VPS deployment remains unverified because no target was supplied. No live app process or user household files were altered.

## Completed Internationalization specification

The user confirmed household-wide Swiss Standard German (`de-CH`, informal) and American English (`en-US`) as a separate Must-have domain. The interview resolved setup/browser preferences, saved-language timing, exact YAML values, legacy/null/invalid handling, localized errors/validation, data preservation, runtime fallback/completeness checks, bundled delivery, price formatting and title sorting. INT-001–INT-005 acceptance criteria and the translation table were approved separately.

The approved table contains 161 rows, including existing browser-facing Go error messages, server/client UI and accessibility text, dynamic/plural phrases and new language/required-field validation text. Its inventory was checked against the existing Go UI error strings; no inventoried message is missing. The German draft was checked for Swiss `ss` spelling. `git diff --check` passes. No application behavior changed; runtime tests were not rerun for this specification-only checkpoint. Temporary interview/review files were kept separate until approval; approved content now lives in `sdd/`.

## Completed Internationalization implementation

All approved INT-001–INT-005 acceptance criteria are implemented. Temporary work was tracked separately in `.work/i18n-implementation.md`; this checkpoint records completed facts.

The initial Go behavior tests failed on missing locale support/legacy fallback. Later completeness checks exposed untranslated template fragments. Browser tests exposed a test-navigation issue after a validation POST and mobile overflow for very large grouped prices; tests now navigate to the canonical list and prices wrap without precision/data changes. Strict YAML tests include missing/empty/null/malformed/unsupported values and null/empty values in merged mappings, with no writes on error.

Verification passes: `go test -race ./...`, `go vet ./...`, CSS build, all 22 Chromium scenarios, and `node tests/container/smoke.cjs` in English and `WISHMAS_SMOKE_LANGUAGE=de-CH` in German. The approved-table/catalog inventory, template visible/accessibility/client-string checks and placeholder checks pass. Browser tests verify localized validation across browser/household language differences, setup switching only after completion, both languages, manual changes on form submissions, price/raw-data separation, privacy/history and saved sorting. German desktop/mobile screenshots were reviewed. DinD checks preserve language, links, claims and YAML across restart/replacement, with disposable resources cleaned up.

No external translations or runtime dependencies were added. No live household, running process or VPS was altered. All implementation work in this checkpoint is complete; the VPS deployment limitation remains unchanged.
