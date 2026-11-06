# Access

## Status

ACC-005–008 are implemented and verified, including optional S3 cached access and durable logout. Work is on `main`; pre-refactor baseline `1992bdf` remains in history. No real S3 deployment or live-data change is implied.

ACC-001–ACC-004 describe the superseded parent/shared-family-code model. They are retired from this revision, not renumbered or reused. Git history retains their original specification and verification. Their obsolete runtime/templates/tests were removed in milestone 6; the offline migrator alone reads legacy storage.

## Priority

Must-have. Parents/admins manage the household; household users browse and claim through individual list links.

## ACC-005: Setup and password configuration

- **ACC-005-AC1:** Until initial setup completes, the setup wizard remains open. Setup creates the household and asks for an admin password and matching confirmation. A mismatch does not complete setup.

  **Security review, 2026-11-05:** the user explicitly accepted SEC-03 and assigns responsibility for promptly completing setup on a fresh instance to the administrator. Keep the open first-run setup flow without a bootstrap token. This accepts the first-visitor ownership window; it does not establish restricted proxy access or remove the risk when missing data exposes setup again.
- **ACC-005-AC2:** A new password must contain at least eight characters, an uppercase letter, a lowercase letter, a digit and a punctuation/symbol character. Letter/digit classes include Unicode. Spaces are permitted but do not satisfy the special-character requirement. Setup includes a passphrase hint.
- **ACC-005-AC3:** Persist a password hash, not the password itself. After setup, changing the shared password requires a manual YAML hash edit; there is no password-change UI.
- **ACC-005-AC4:** Preserve household-language selection and appearance previews/settings. Remove timezone selection and automatic calendar-year behavior. Setup no longer generates or displays parent/shared-family links.
- **ACC-005-AC5:** A newly created household starts with no lists, recipients, users or wishes. These are created explicitly by parents after setup.
- **ACC-005-AC6:** Completing setup automatically signs the creating browser into admin access.

- **ACC-005-AC7:** On setup validation failure, clear both password fields for re-entry while preserving household name, language and appearance settings.

## ACC-006: Admin login and persistent sessions

- **ACC-006-AC1:** Admin access requires the one shared household password; there are no separate parent accounts or parent access links.
- **ACC-006-AC2:** After setup, an unauthenticated request without a user-link code, or with an invalid/replaced code, asks for the admin password rather than showing the former not-found page.
- **ACC-006-AC3:** Successful admin login permits management of the household, all lists, both rosters, household wishes, assignments and claims. Lists remain editable regardless of dates in their names.
- **ACC-006-AC4:** Admin sessions remain signed in until logout or a password-hash change, including across browser/app restarts and container replacement using the same retained storage.

  **Security review, 2026-11-05:** the user explicitly accepted SEC-02's remaining stolen-token replay risk because this is a family app with no critical data in their stated usage. Retain indefinite sessions; no automatic server-side expiry is required. This is risk acceptance, not a technical mitigation. Reassess if scope or data sensitivity changes.
- **ACC-006-AC5:** Logout ends the current browser's admin session only. Other admin browsers remain signed in.
- **ACC-006-AC6:** A YAML password-hash change invalidates all existing admin sessions. Subsequent access requires the new password.
- **ACC-006-AC7:** Five failed admin-login attempts from one IP within fifteen minutes cause a fifteen-minute block of further attempts from that IP.
- **ACC-006-AC8:** A valid individual user link shows that user's restricted view even if the browser also has an admin session. Actions through that view act as that user; the admin session remains available separately.
- **ACC-006-AC9:** After setup, the admin login page uses the configured household language, including requests with missing, invalid or replaced user-link codes.
- **ACC-006-AC10:** Login limits use the client IP forwarded by Traefik. Deployment exposes the app only through Traefik; restarting the app may reset failed-attempt counts and active blocks.

## ACC-007: Household users and list links

- **ACC-007-AC1:** Only parents create, rename and delete household users. User and recipient rosters are separate; identical names across rosters imply no connection.
- **ACC-007-AC2:** User names are trimmed, contain 3–50 characters, and allow Unicode letters/digits and spaces. Names are unique within the user roster ignoring capitalization and surrounding spaces.
- **ACC-007-AC3:** Each household user has an individual coded link for every list, including subsequently created users/lists. There is no list-membership management; parents decide which links to share.
- **ACC-007-AC4:** A user link identifies exactly one household user and permits access to exactly one list, all its selected recipients and their assigned wishes. It does not grant admin access, access to other lists or access to unassigned household wishes.
- **ACC-007-AC5:** Users do not enter or edit claimant names. Their link identifies the claimant. Renaming a household user preserves IDs, links and existing claims.
- **ACC-007-AC6:** Parents can independently regenerate one user's link for one list. Only that replaced link becomes invalid; other links and all claims remain unchanged. The old link loses permission on its next request/action, including an already-open page.
- **ACC-007-AC7:** Deleting a household user invalidates all their links and removes all their claims, including on unassigned wishes. Other users remain unchanged. Require confirmation explaining the affected claims.
- **ACC-007-AC8:** Deleting a list invalidates its links without deleting household users or affecting their other-list links. Wish/claim cleanup is specified in Wishlists.
- **ACC-007-AC9:** Users appear in added order. Renaming a list preserves its user links. Link codes have no automatic date/year expiry.
- **ACC-007-AC10:** Link-copy actions retain the existing inline clipboard-icon button and three-second checkmark feedback. Regeneration retains a confirmation modal.

## ACC-008: S3 cached access and durable logout

**Implemented and verified: ACC-008-AC1–AC10.** Applies to the optional S3 backend in CON-009. Existing local-mode authentication remains unchanged.

- **ACC-008-AC1 — Cached identities:** When S3 is temporarily unavailable and a valid persisted household fallback exists, authenticate existing admin sessions and individual user links against that cached household. Retain their existing role/list/picture scopes and claimant privacy. Locally revoked sessions in AC5 remain rejected even if still present in that snapshot.
- **ACC-008-AC2 — No new offline login:** During cached read-only mode, block new admin logins and do not create new sessions from the cached password hash. Existing authenticated browsers and valid user links retain read-only access under AC1.
- **ACC-008-AC3 — Enforced read-only mode:** Show state-changing controls disabled with an explanation, including admin management and claim changes. Enforce the same restriction on server requests, including forms opened before the outage. Reject failed edits/claims rather than queue them. Logout under AC4 is the confirmed exception; read-only navigation, link copying and browser sorting retain their existing behavior.
- **ACC-008-AC4 — Available logout:** Keep Logout available during an S3 outage. Durably record the current session's pending revocation locally under `WISHMAS_DATA_DIR` before acknowledging logout, then clear its browser cookie. Acknowledging logout does not wait for eventual S3 session removal. Other admin sessions remain unaffected.
- **ACC-008-AC5 — Durable local rejection:** Reject an acknowledged logged-out session immediately and across app/container restarts using retained local storage, while its S3 removal is pending. Store only session-token hashes in local revocation records, retaining CON-007; never persist raw session tokens. Scope each revocation to its corresponding S3 household/session.
- **ACC-008-AC6 — Local persistence failure:** If the local pending-revocation record cannot be saved, show an error, keep the current session/cookie and do not report logout success.
- **ACC-008-AC7 — Automatic retry:** Retry pending S3 session removals after increasing delays of 1, 2, 4, 8, 16 and then 30 minutes, capping subsequent delays at 30 minutes. Continue until the corresponding remote session is confirmed absent. Revocation records survive restart so local rejection and pending removal are not lost.
- **ACC-008-AC8 — Retry on source recovery:** Whenever a page load successfully reaches S3, attempt pending session removals immediately rather than waiting for the next retry delay. This is a server-side operation; no browser polling is added.
- **ACC-008-AC9 — Narrow reconciliation:** Reload current household state and remove only the pending revoked sessions, retaining the existing single-instance serialization/freshness checks. Preserve other sessions, password configuration, users, links, wishes, claims and unrelated household data. Cached access must not resurrect an acknowledged logged-out session if a later local YAML-cache refresh fails.
- **ACC-008-AC10 — Normal access after recovery:** The next successful page load uses current S3 household data and restores normal authorized operations under CON-009. Apply the current password hash, session records and link codes; pending local revocations continue to be rejected until reconciled. No local pending logout becomes a queued edit or claim.

Read-only/login/logout explanations and errors HR113–HR120 are separately approved and bundled in translations.md/messages.json.

## Relationships

- Wishlists owns named collections, household recipients, recipient selection and deletion effects.
- Wishes owns household wish data, assignments, pictures and existing detail rules.
- Claims owns user-ID ownership, admin selection per claim and claimant privacy.
- Internationalization owns setup/login/error language and bundled translations; obsolete timezone/year/browser-claim-name text must be removed when no longer used.
- Persistence and migration constraints will replace the old year/event/person hierarchy. Until cutover approval, only a copy of the existing data may be migrated.

## OPEN items for interview/review

- No behavioral or wording decision remains OPEN. Retained and new wording is approved in translations.md.
- Technical contract is approved in constraints.md and storage.md. Deployment-specific trusted Traefik addresses remain to be supplied before deployment.

## Implementation and verification

The Access domain is **Implemented** after cumulative milestones 2–6 verification. Earlier checkpoint notes below retain their original verification boundary; milestone 6 closes the container gap.

| Criteria | Verified scope / remaining work |
| --- | --- |
| ACC-005-AC1–AC7 | New default password setup, Unicode policy/confirmation, hash-only storage, empty household, preserved language/appearance and cleared password drafts, no timezone/parent-family links, auto sign-in. Go and English/German HTTPS browser checks pass. |
| ACC-006-AC1/AC2/AC5/AC6/AC7/AC9/AC10 | Password login; invalid-code login language; browser-local logout; hash-change revocation; five-failure/fifteen-minute block; trusted-peer/protocol/client-IP checks including duplicate headers. Verified with deterministic clock, real browsers and write-failure preservation. |
| ACC-006-AC4 | Persistent hashed sessions verified across fresh browser contexts, handler reconstruction, real server-process replacement and scratch-container restart/replacement behind actual Traefik HTTPS using retained storage. |
| ACC-006-AC3 | Authentication boundary verified. List/roster management is verified in milestone 3; household wish management is verified in milestone 4; claim actions are verified in milestone 5. |
| ACC-006-AC8 | Valid user-link identity overrides an unrelated admin cookie and cannot invoke admin logout. Scoped recipient/read-only wish rendering and admin-action denial verified in milestone 3; cards/images/summaries and wish-action denial are verified in milestone 4; claim actions are verified in milestone 5. |
| ACC-007 | Milestone 3 verifies household-user creation/validation/added order, every-user/every-list codes, scoped read-only views, stable-ID renames, independent regeneration, three-second copy feedback and confirmed deletion/link invalidation. Seeded claims verify exact preservation/removal, including unassigned wishes; interactive claim integration is verified in milestone 5. |

Implementation: `internal/app/admin.go`, shared decoration/layout and new admin templates, production constructor and explicit trusted-proxy configuration in `cmd/wishmas`. Tests: `admin_auth_test.go`, `tests/browser/admin-auth.spec.js` and retained pure/legacy component checks.

RED: default setup still offered old links. Subsequent failures caught lost slider drafts, a missing panel marker in theme assertions, and first-header-only proxy parsing. Browser-native POSTs exposed the old no-referrer policy emitting Origin:null; new pages use same-origin policy, preserving external referrer privacy with strict CSRF checks. The initial certificate diagnosis was corrected; no production CSRF relaxation or launch trust flags remain.

GREEN: `go test -race ./...`, vet, no-CGO/nodynamic builds, CSS build, JavaScript syntax/diff checks and all 31 browser scenarios pass (four new HTTPS auth flows; 27 explicit legacy-domain checks). Reviewed setup/login/admin screenshots on desktop/mobile. No live app/data changes or real migration. Milestone 3 subsequently completed the management slice below. [Review instructions](../docs/admin-access.md).

### Milestone 3 evidence

`admin_management.go` and authenticated routing implement household user/list management. `admin_management_test.go` and `admin-management.spec.js` verify independent users and lists, old-link invalidation even with an admin cookie, Unicode names/separate rosters, stable-ID renames and unaffected links/claims, deletion cleanup, stale confirmations and failed-write preservation. English/German browser checks include native modals, clipboard icon/checkmark timing, scoped recipient navigation, restart persistence and mobile layout. No browser-name prompt or list-membership controls exist. Full wish/claim management and container proof remain later milestones; do not mark the domain Implemented.

### Milestone 4 scope evidence

Household wish mutations require the current admin session and CSRF/revision checks. Valid user codes override admin cookies for cards and direct image URLs. Only currently assigned wishes/pictures in the code's list are returned; moves/unassignment/deletion/regeneration are rechecked on every request. In-flight image work releases the app lock, then revalidates the current session and file revision before publishing. New Go/English/German HTTPS checks verify scopes and privacy. Claim-action integration is verified in milestone 5.

### Milestone 5 claim identity and permissions

ACC-006-AC3/AC8 and ACC-007 lifecycle integration are verified with real user-ID claim actions. The selected link overrides admin cookies and browser-entered/posted names; a browser can switch links without retaining the previous user's identity. User and list deletion revoke old claim forms immediately. Regeneration preserves claims and revokes only that link. Renaming preserves ownership and updates current displayed names. Go race and English/German independent-browser tests verify all scopes, failed-write/stale protection and CSRF/Origin guards. The Access domain remains In progress only for actual container replacement evidence in milestone 6; do not conflate process-restart tests with container proof.

### Milestone 6 final delivery

Actual isolated nonroot Docker containers behind Traefik verify HTTPS setup/login, host-only Secure/HttpOnly/SameSite cookies, stable session identity after restart/replacement, exact retained YAML/image bytes, two users, scoped claims/privacy, independent link regeneration and spoof-resistant client-IP limits/reset. Both household languages pass. The app has no public backend mapping; only Traefik is published. The offline helper is included. Current Go race/vet/static and all 20 current-runtime browser checks pass. Original data/current app/main remain available; no merge, live cutover or actual VPS deployment is claimed.

### Optional S3 access delivery

`admin.go`, the S3 repository/revocation service and native disabled controls implement cached identity/scopes, blocked new sign-ins/mutations and durable hash-only logout. HTTP tests verify claimant privacy, stale forms, preserved drafts, uncertain saves, local logout-persistence errors and stale-fallback warnings. Repository tests verify independent sessions, scoped denial across restart, increasing deadlines, safe reconciliation and cache-refresh failure without resurrection. Bilingual browser and real stopped-provider/container-restart checks verify the complete app behavior. All approved ACC-008 criteria are met; external provider/deployment limits are recorded in s3-storage.md/progress.md.
