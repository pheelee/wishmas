# Household refactor implementation plan

Approved by the user; domain behavior, technical contract and all 112 new/revised wording entries are approved. Milestones 0–5 are complete; pause for review before milestone 6. No application code starts until the behavioral specs, technical contract, new English/German wording and this plan are approved. All work stays on `refactor/household-lists-access`; baseline `main` is `1992bdf`. Original `data/`, current app and user-owned `.gitignore`/notes remain untouched.

## Checkpoint procedure (every milestone)

1. Select acceptance criteria before code; record them in the temporary work tracker while unfinished. Inspect branch/status and the last checkpoint before resuming.
2. Write an observable failing behavioral test and record its actual failure. Implement with `@impl`/`@test` anchors.
3. Run meaningful focused checks, then appropriate cumulative Go/browser/HTTPS checks. Record failures, corrections, remaining gaps and exact commands. Keep pending behavior explicit; do not mark partial criteria/domains Implemented.
4. Produce a concrete review result, document only completed/verified facts in permanent progress/change records, and remove that milestone's temporary artifacts after their facts are retained.
5. Commit a completed verified slice with a Conventional Commit. Record the unique subject while its own hash is pending; resolve it on resumption.
6. **Stop for the user's review and wait for explicit approval before starting the next milestone.** No merge, original-data migration or live cutover is implied by milestone approval.

## Milestones

| Milestone | Delivered slice and criteria | Observable review | Verification boundary |
| --- | --- | --- | --- |
| 0 — approved specification | Consolidate approved Access, Wishlists, Wishes, Claims and Internationalization revisions, CON-004/007/008, glossary, retired-feature map, wording and this plan. | Review final specification and implementation checkpoints before any code. | Check references/criterion IDs and remove contradictory legacy target requirements. No application behavior claimed. |
| 1 — data core, hash helper and migration command | Version-2 single-file model/repository, stable IDs, validation/atomic writes; Argon2id helper; explicit copy-only legacy migration and reports (CON-004, CON-007-AC1/AC2, CON-008 migration mechanics). | Run helper privately, inspect a fictitious YAML sample and migrate disposable legacy fixtures; see comparison report and image-checksum preservation. New app screens are not part of this milestone. | RED invalid-reference/partial-write/migration checks before code; Go tests/race/vet/build. Test empty/nonempty destination, invalid source, claims/name matching, duplicate wish preservation, missing images and unchanged source. Record exact schema/helper usage. |
| 2 — setup and admin authentication | New setup, language/appearance preservation, password validation/confirmation, persistent hashed sessions, login/logout/hash-change invalidation, Traefik client-IP limits and HTTPS contract (ACC-005/006; CON-007; relevant INT/CON-001). Retire parent/shared-family code and timezone setup. | Fresh disposable household setup signs in; another browser needs password; restart retains login; logout affects one browser; hash change ends both. Empty admin landing page is explicit, not a claim that list management is complete. | Failing-first HTTP/HTTPS tests and English/German browser flows, including errors clearing passwords, cookie flags, CSRF, invalid-link login language and spoof-resistant IP limits. Use a deterministic clock for limits. Adapt/retire obsolete access tests with a requirement-to-test map. |
| 3 — lists, recipients, users and links | Household rosters, naming/order/rename/delete; manual lists and recipient selection; every-user/every-list link generation, copy/regeneration, scoped read-only user views (ACC-007; WLS-007–009). Test cleanup against seeded wishes/claims even before wish UI is complete. | Create a list, select recipients, create users, copy two independent links, view all selected recipients, regenerate one link, and observe the old link ask for password. Rename entities without breaking links. Review destructive confirmations. | RED role/scope/ID tests first; browser tests with independent identities; verify stale links, no name prompt, no annual/archive behavior, and correct seeded-data cleanup for every deletion path. Existing users stay household-wide when a list is deleted. |
| 4 — household wishes, assignment and pictures | New household wish overview/filters; CRUD from household and recipient pages; single assignment/moves/unassignment; retained fields/cards/sorting/images; scoped recipient summaries (WSH-001/003–007; WLS-010/011; CON-006). | Create an unassigned wish; assign to a selected recipient; move it to another list; verify preserved claims/details/picture and old-list access denial. Inspect totals/counts and mobile cards. | RED ownership/assignment/image-scope tests before code. Adapt retained detail/image/decimal/sorting tests to IDs/new storage; verify errors/rollback and unchanged source downloads, manual edits, sorting targets and exact summaries. Seed claims to verify preservation before claim UI. |
| 5 — user-ID claims and lifecycle integration | Assigned-only creation; per-link user ownership; admin selection per claim; multiple/duplicate claims; names/privacy; removal; all lifecycle cleanup (CLM-005–008 plus cross-domain criteria). | Claim via two user links, create/remove claims as admin, rename a claimant, move/unassign/reassign wishes, delete users/recipients/lists and inspect the correct preserved/cleared claims. | RED claim/ownership/privacy tests first. Independent-browser English/German flows, full-response privacy checks, tampered IDs, concurrency/stale requests, regenerated links, manual invalid data, no polling/notifications, and write-failure preservation. |
| 6 — current-data migration, full review and packaging | Fresh copy of the user's current data, with hash privately supplied in the copy; explicit migration and comparison; final legacy removal, translated catalog cleanup, container/Traefik HTTPS/persistence verification (all criteria; CON-001/008). | Run isolated review app using migrated data; show current recipients/wishes/pictures, counts/claims, new admin login and user links. Original app/data still works. Provide final review/rollback instructions. | Check source snapshot before/after copying; stop if changing/inconsistent rather than migrating a mixed snapshot. Verify all old/new counts/details/claims/image bytes, full Go race/vet/build, all applicable browser scenarios, catalog completeness, mobile screenshots and container restart/replacement including admin sessions. Actual VPS deployment is distinct from local checks. |

Each milestone is a review stop. Newly implemented behavior remains labelled partial until every criterion it owns is fully met. Milestones 3–4 use seeded fixture claims where necessary; they do not pretend claim actions are implemented before milestone 5.

## Legacy removal and test accountability

- Remove parent/shared-family links, year/timezone routing and navigation, automatic annual list creation, old read-only-history guards, archive/restore, browser-stored claim names and their cookies/forms/messages from the finished runtime.
- Remove old year/event/person storage writes and per-person image paths. Keep legacy reading only in the explicit migration tool.
- Rewrite context-specific fixtures/assertions to the new requirements, while preserving meaningful price/image/appearance/localization/privacy/stale-write tests. Track each retained, replaced or retired suite against its requirement; never skip a failing test just to claim a passing milestone.
- Baseline proof remains in Git history. New domains are not Implemented until all revised criteria have evidence.

## Resume record

At each checkpoint, record branch and baseline; selected criterion IDs; last verified commit/unique subject; completed tests and failures; uncommitted files and their owner; review fixture/data path and safe start/stop commands; whether the user has approved the next milestone; unresolved blockers; exact next action. Do not record credentials, real user links or password hashes in Git/progress notes.

On reconnect, read AGENTS.md, the approved domain/constraint docs, implementation/progress records and the dedicated temporary tracker. Inspect Git status/history before editing; resolve pending checkpoint subjects. Resume only the authorized milestone. When stopped at a review gate, wait for the user's response rather than inferring approval.

## Migration, merge and rollback boundary

- Implement/test migration on disposable fixtures first. At milestone 6, use a fresh private copy of current data, never the original. Ask the user to place the initial hash into that copy using the completed helper; no password/hash is requested in chat.
- Do not run old and new app versions against the same data directory. Review uses isolated HTTPS and a distinct data path; do not take over the existing port/process.
- No existing source files/pictures are deleted by migration. Keep the baseline commit and original snapshot available.
- After final review, stop for an explicit merge/cutover decision. A later cutover needs a fresh snapshot/migration if live data changed during review. Rollback restores the matching baseline code and legacy data together; the old app must never read new-schema data.
- Final cutover commands/target depend on the approved review environment and deployment paths; supply those concrete details before asking for cutover approval. No actual VPS deployment or data replacement is preauthorized by this plan.

## Completed checkpoints

- Milestone 0: `5bf2e9c`, approved specification/wording/technical contract and this plan.
- Milestone 1: data core, password helper and copy-only migration command completed and verified; locate commit by `feat(storage): add household data core and offline migration tools`. Review instructions: [offline admin tools](../docs/offline-admin.md). At that checkpoint the app/data remained legacy; the user subsequently approved milestone 2.

- Milestone 2: password setup/admin authentication, persistent hashed sessions, logout/hash-change invalidation, translated login and trusted-proxy rate limits completed and verified; locate by `feat(access): add password setup and persistent admin sessions`. Review: [isolated HTTPS instructions](../docs/admin-access.md). No real data migration. The user approved milestone 3, now complete.

- Milestone 3: manual lists, household recipients/users, recipient selection, individual links/copy/regeneration and scoped read-only views completed; locate by `feat(household): add lists rosters and individual user links`. Review: [management walkthrough](../docs/admin-access.md#milestone-3-review). The user subsequently authorized milestone 4, now complete; claim actions remain milestone 5.

- Milestone 4: household wish CRUD/filtering, recipient-context creation, assignment/moves/unassignment with claim preservation, cards/sorting/pictures and scoped summaries completed; locate by `feat(wishes): add household wish management assignments and pictures`. Review: [wish walkthrough](../docs/admin-access.md#milestone-4-review). The user subsequently authorized milestone 5, now complete.

- Milestone 5: user-ID claim creation/removal, per-claim admin selection, assigned-only creation, ownership/privacy and complete lifecycle integration verified; locate by `feat(claims): add user identity claim actions and lifecycle integration`. Review: [claim walkthrough](../docs/admin-access.md#milestone-5-review). Pause before milestone 6; no real-data migration or merge yet.

- Approved extension after milestone 5: AVIF source support under existing image rules and approved I03/I15 wording is complete; locate by `feat(wishes): support AVIF picture sources`. Go race, targeted bilingual browser and isolated scratch/no-CGo codec verification pass. Milestone 6 remains review-gated.

### Current test migration map

| Existing coverage | Current disposition | Target coverage |
| --- | --- | --- |
| Legacy setup/access/year/name tests | Explicit `newLegacy` constructor and test-only browser harness; historical component checks, not a production option | `admin_auth_test.go` / `admin-auth.spec.js` replace default setup/login/session/limit assertions |
| Legacy i18n/catalog inventory | Exact 182-row fixture retained; approved HR entries added to exact inventory checks | New setup/login languages and errors covered; remove obsolete catalog/fixtures when their UI domains migrate |
| People/history/rosters and link management | Legacy-only historical checks retained until final cleanup | `admin_management_test.go` / `admin-management.spec.js` verify new lists/rosters/links, scopes and seeded lifecycle cleanup; year/archive controls absent from production |
| Wishes/images/sorting/status | Shared pure processor/validation/JS checks and historical legacy integration retained | `household_wishes_test.go`, repository picture tests and bilingual `household-wishes.spec.js` verify new contexts, scoped pictures, CRUD/assignments/summaries; final obsolete harness cleanup remains milestone 6 |
| Claims actions | Historical name-based scenarios retained only until final cleanup | `household_claims_test.go` / bilingual `household-claims.spec.js` verify real ID-based actions, ownership/privacy, stale/concurrent writes, guards and lifecycle effects |
| Legacy container smoke | Still describes old setup/links; not claimed as current-runtime proof | Rewrite/verify full Docker/Traefik/new session replacement in milestone 6 |

Production `New`/the main executable uses version 2 only and rejects legacy data. The compiled Go test executable alone recognizes browser-test modes. Retire the harness/remaining obsolete code/assets/catalog at final cleanup; do not weaken checks or reintroduce legacy production access.
