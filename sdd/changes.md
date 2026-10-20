# Changes

## Change log

- Gift-wrap header correction: restored interior spacing below ownership badges, vertically centered gift icons beside one/two badges, and gave headers a consistent minimum height. Reproduced from the running app; failing-first geometry tests now pass on desktop/mobile, with seven relevant browser scenarios and Go tests passing. Live data and the running process remain unchanged.

- Gift-wrap card slice: Implemented user-approved WSH-005-AC1–AC5 with decorative mint/gold/ice-blue edges and gradient headers, gift icons, serif titles, claim/ownership badges, prominent optional numeric prices, subtle domain links, and separated inline claim/edit/delete footers. Existing parent-only names/removal, family privacy, read-only history, sorting targets and modal behavior remain verified. New browser test failed before implementation; all 19 Chromium scenarios, Go race tests, vet and builds pass. Reviewed desktop/mobile screenshots. Only the selected reference is included; unselected proposals and user notes remain uncommitted.

- Claims identity/action slice: Implemented first-claim browser naming, persistent cookie storage, displayed names and modal future-only renames, parent/family claim toggles, multiple claimants, case/space duplicate and ownership checks, and direct parent removal. Family responses expose only counts and the current name's ownership marker; claimant lists remain parent-only. Historical/archive/code/snapshot guards and refreshed inline errors preserve existing claims, including concurrent duplicate and disk-write failures. Shared YAML aliases retain other wishes' data; cyclic expansion fails without writes. Failing-first Go tests, two new multi-browser scenarios, all eighteen Chromium scenarios, race checks, vet, builds, and mobile review pass. CLM-001 through CLM-004 and milestone 5 are Implemented; Docker deployment and final complete-flow verification remain.

- UI-polish slice: User approved all acceptance criteria before code. Added household YAML appearance sliders with live previews (bulbs 80–200%, default 120%, inverse count; snow size 50–200%, default 100%; amount 8–48, default 16), browser timezone prefill, and always-present snow without pause controls while retaining visible static reduced motion. Access links now have inline icons with clipboard/check feedback for three seconds and regeneration modals. Editing/action forms use native modals, creations stay inline, page width above md is 75%, and wishes use responsive cards with bottom icon actions. Failing-first tests, settings/restart/legacy validation, modal interaction flows, screenshots, and regression checks verify the completed behavior. No dependency or deployment change; `ideas.md` remains untouched.

- Read-only claim visibility slice: Added claimed/unclaimed status, explicit claimant counts, and parent-only claimant names on current and historical wishlists. Family HTML contains no claimant-name list. Validated persisted claimant structures, trimmed Unicode name boundaries, and per-wish case-insensitive duplicates without rewriting invalid data. Verified manual claim changes on reload, unchanged open pages, and preservation during wish edits. CLM-003 remains In progress: claimed-by-you and actual creation/removal refresh integration are pending. Browser identity and all claim actions remain pending; no whole Claims requirement is marked Implemented.

- Wish sorting slice: Implemented default added order, newest first, title A–Z, and numeric price orders, with missing prices treated as zero and stable added-order ties. Title comparison ignores case/accents. One browser-stored preference applies across all lists and roles; no YAML is rewritten. Whole cards move together so edit/delete targets remain correct. Failing-first browser scenarios verify every criterion, precise large-number comparisons, unchanged files, reload/cross-list persistence, fresh-browser defaults, and storage fallback. WSH-003 and the Wishes domain are Implemented. Claims and Docker deployment remain pending. The user's deferred `ideas.md` review notes remain ignored and untouched.

- Confirmed wish deletion slice: Added a parent-only confirmation page with cancellation and permanent removal of wishes plus all attached claims. Fresh code, year, archive, and data validation protect both confirmation and removal. A failing replay regression exposed identical-wish index reuse; whole-list version fingerprints now protect editing/deletion against changed lists and repeated submissions. Other wishes and historical files remain intact. WSH-002 is Implemented; sorting remains the last Wishes requirement. Go and browser verification include dark mobile confirmation, cancellation, stale pages, and claimed/unclaimed removal. Existing untracked `ideas.md` was left untouched and excluded from the checkpoint.

- Wish creation/editing slice: Implemented required trimmed Unicode titles and optional ordered shop links, one numeric reference price, and notes. Parents can add/edit current-year wishes; family, historical, and archived mutations are rejected. Edits preserve claims and future YAML fields; stale forms and invalid data leave files untouched. Failing-first Go tests and browser verification cover validation, detail presentation, permission guards, and the dark mobile layout. WSH-001 and WLS-001 are Implemented. Wish deletion, sorting, and claim UI remain pending.

- Person lifecycle/history slice: Added parent-only renaming with raw directory moves across years/events, case-insensitive uniqueness, added-order preservation, and rollback for directory/roster-save failures. Added archiving that hides all years while preserving data and reserved names, with manual-only restoration. Parents now navigate saved Christmas years; historical pages are read-only and reject stale/historical mutation contexts. Family history stays inaccessible, and regenerated parent links retain historical scope. Fixed household Unicode truncation after a failing browser regression test. All eight browser scenarios, Go race tests, vet, build, CSS build, and whitespace checks pass. Access and WLS-002/003/004 are Implemented; WLS-001 still depends on wish-writing work. No intermediate or unfinished implementation committed.

- People/current-list slice: Added parent-only person creation with trimmed 3–50 Unicode-letter/digit names, case-insensitive uniqueness across active/archived people, added-order display, and persisted empty current-year Christmas lists. Both link roles can view current lists; family past-year requests are rejected. Direct roster/title edits reload, invalid data leaves files untouched, new-year creation skips archived people, and exclusive file creation protects manual lists. Failing-first Go tests, Unicode boundary regression, concurrent-duplicate checks, replacement-link/current-list integration, all six browser scenarios, race tests, vet, build, and CSS build pass. Rename, app archiving, parent historical navigation, wish CRUD, and claims remain outside this completed slice; no whole Wishlists requirement is marked Implemented.

- Frontend theme: User approved CON-003-AC1–AC5 before implementation. Implemented the approved December-night palette, steady Christmas bulbs, 16 faint margin-only snow particles, keyboard pause/resume with browser persistence, and reduced-motion handling. Applied styling to setup, link preview, parent/family, validation, not-found, and data-error screens. Fixed long-name overflow at 320px. Failing-first Go/browser tests and all final checks pass. Preserved the approved reference as `design/december-night/README.md` and `approved-baseline.png`; no new proposal or unfinished code is committed.
- Working rules: Added the user's Conventional Commits requirement and prohibition on committing intermediate proposals, work awaiting review, or unfinished implementation.

- Visual review: User approved December-night proposal 01 as the appearance baseline. Updated CON-003 with the confirmed direction and drafted implementation/motion criteria outside the spec for approval. No application code changed.
- Visual review: Created a static December-night theme proposal under `design/december-night/`, with dark navy surfaces, Christmas-bulb accents, and sparse background snow. Specific appearance and motion remain OPEN for review. No application code or animation changed; the functional checkpoint remains intact.

- Implementation, Access regeneration/calendar slice: Added parent-only independent regeneration with copyable replacement links and automatic parent navigation. Every action validates against reloaded YAML; replaced codes immediately lose read/action access, including stale and concurrent parent submissions. Added failing-first anchored Go tests and browser coverage. Deterministic clock tests verify household New Year, manual timezone reload, and non-expiring codes. Go race tests, vet, build, CSS build, and browser checks pass. ACC-001 is Implemented; Access and replacement-link list scopes remain In progress until milestone 3.

- Implementation, setup/persistence slice: Added the Go HTTP application, household-name/timezone wizard, automatically generated parent/family links with browser copy buttons, explicit completion into parent access, and persisted YAML household metadata. Added failing-first Go behavioral tests and Chromium verification with `@impl`/`@test` anchors. Go race tests, vet, build, CSS compilation, browser flow, process-restart smoke check, and dependency audit passed. Updated run instructions and resume checkpoint. Access remains In progress; calendar/list scope and regeneration are incomplete.

- Implementation preparation, Access: User confirmed household-name validation matches person names: trim surrounding spaces, require 3–50 Unicode letters/digits. Added ACC-001-AC9; no listed Access OPEN items remain.
- Implementation preparation, Access: User confirmed that setup asks for a household name. Added ACC-001-AC8; household-name validation remains OPEN. Updated the resume checkpoint; no application code or tests written.
- Implementation planning: User approved the milestone and resume approach. Added `implementation.md`, `progress.md`, and AGENTS.md checkpoint/resume instructions. Preparing the setup/persistence slice; household naming during setup is OPEN. No application code or tests written.
- OPEN review, Claims: User approved the consolidated updated requirements and acceptance criteria. All four domains now have approved OPEN-review updates; no implementation started.
- OPEN review, Claims: Confirmed Must-have for v1 priority. All listed Claims OPEN items are resolved; consolidated domain approval is pending. No implementation started.
- OPEN review, Claims: Confirmed a simple toggle for one's own claims and no confirmation for claim creation or removal; added CLM-002-AC6 and CLM-004-AC7 and closed the confirmation question.
- OPEN review, Claims: Confirmed inline errors and refreshed claim status for unsuccessful attempts, preserving existing claims, including stale-page duplicates; added CLM-002-AC5 and closed the unsuccessful-attempt presentation question.
- OPEN review, Claims: Confirmed other viewers see claim changes on page load only in v1; added CLM-003-AC6 and closed the update-timing question.
- OPEN review, Claims: Confirmed trimmed claimant names of 3–50 Unicode letters/digits, without uniqueness across browsers; added CLM-001-AC6/AC7 and closed name validation.
- OPEN review, Claims: Confirmed the initial name prompt occurs on the first attempt to claim a wish, with no name required for browsing; added CLM-001-AC5. Name validation remains OPEN.
- OPEN review, Wishes: User approved the consolidated updated requirements and acceptance criteria. No listed Wishes OPEN items remain; no implementation started.
- OPEN review, Wishes: Confirmed Must-have for v1 priority. All listed Wishes OPEN items are resolved; consolidated domain approval is pending. No implementation started.
- OPEN review, Wishes: Confirmed no v1 notifications for edits or deletion of claimed wishes; added WSH-002-AC7 and closed the notification question.
- OPEN review, Wishes: Confirmed confirmation before permanent wish deletion; added WSH-002-AC6 and closed the deletion-confirmation question.
- OPEN review, Wishes: Confirmed one browser-stored sort preference across all wishlists; added WSH-003-AC8 and closed the sort-preference scope question.
- OPEN review, Wishes: Confirmed case-insensitive title sorting with accented letters treated like base letters; added WSH-003-AC7 and closed the title-collation question.
- OPEN review, Wishes: Confirmed added order, oldest first, for equal sort values; added WSH-003-AC6. Title collation remains OPEN.
- OPEN review, Wishes: Confirmed shop links display in added order; added WSH-001-AC11 and closed the link-ordering question.
- OPEN review, Wishes: Confirmed domain-name labels for shop links; added WSH-001-AC10. Link ordering remains OPEN.
- OPEN review, Wishes: Confirmed complete HTTP/HTTPS URLs for shop links, without automatic prefixing; added WSH-001-AC9. Link labels and order remain OPEN.
- OPEN review, Wishes: Confirmed decimal-point-only price entry; added WSH-001-AC8 and closed the price-formatting question.
- OPEN review, Wishes: Confirmed nonnegative optional prices with up to two decimal places; added WSH-001-AC7. Decimal-separator input remains OPEN.
- OPEN review, Wishes: Confirmed trimmed, nonblank titles of 5–150 characters; added WSH-001-AC6 and closed the title-validation question.
- OPEN review, Wishlists: User approved the consolidated updated requirements and acceptance criteria. No listed Wishlists OPEN items remain; no implementation started.
- OPEN review, Wishlists: Confirmed Must-have for v1 priority. All listed Wishlists OPEN items are resolved; consolidated domain approval is pending. No implementation started.
- OPEN review, Wishlists: Confirmed Unicode letters and digits for person names; added WLS-003-AC6 and closed the allowed-alphabet question.
- OPEN review, Wishlists: Confirmed trimmed, nonblank, alphanumeric person names of 3–50 characters; added WLS-003-AC5. Allowed alphabet remains OPEN.
- OPEN review, Wishlists: Confirmed people/list display in the order people were added; added WLS-001-AC8 and closed the display-order question.
- OPEN review, Wishlists: Confirmed an error and no file modifications for invalid manually edited YAML or duplicate person names; added WLS-001-AC7 and updated CON-004.
- OPEN review, Wishlists: Confirmed automatic creation of a missing current-year wishlist on page load after restoring a person; added WLS-004-AC7 and closed the restoration-list question.
- OPEN review, Wishlists: Confirmed parent-only management of people and all wishlists; family access is limited to viewing and claim actions. Added WLS-001-AC6 and closed the family-management permission question.
- OPEN review, Access: User approved the consolidated updated requirements and acceptance criteria. No listed Access OPEN items remain; no implementation started.
- OPEN review, Access: Confirmed Must-have for v1 priority. All listed Access OPEN items are resolved; consolidated domain approval is pending before continuing to another domain.
- OPEN review, Access: Confirmed codes remain valid until regenerated, without automatic expiry; added ACC-002-AC6 and closed the expiry question.
- OPEN review, Access: Confirmed both links and copy buttons on the wizard's final screen, with “Finish setup” completing setup and opening the parent link; added ACC-001-AC6/AC7 and closed the setup-completion question.
- OPEN review, Access: Confirmed manual YAML edits only for timezone changes after setup; added ACC-001-AC5 and closed the timezone-editing question.
- OPEN review, Access: Confirmed parent-only code regeneration; added ACC-003-AC6 and closed the family-permission question.
- Step 6: User approved the Claims draft. All four domain drafts now have user approval, with their listed OPEN items unresolved. No implementation or tests written.
- Step 6, Claims interview: Confirmed “claimed by you” for claims matching the current browser name. Drafted Claims requirements and observable acceptance criteria in `claims.md`; approval pending and unresolved details remain OPEN. No implementation or tests written.
- Step 6, Claims interview: Confirmed that family members can see the number of claimants while names remain parent-only.
- Step 6, Claims interview: Confirmed that previous years' claims cannot be added or removed through the app, including by parents.
- Step 6, Claims interview: Confirmed that parents may also claim wishes using their browser-stored name.
- Step 6, Claims interview: Confirmed case-insensitive claimant-name comparisons that ignore surrounding spaces for ownership and duplicate checks.
- Step 6, Claims interview: Confirmed at most one claim per name on each wish.
- Step 6, Claims interview: Confirmed trust-based removal using the browser's current name, with no impersonation protection required. After renaming, old-name claims can still be removed by parents.
- Step 6, Claims interview: Confirmed that a claim may be removed by its claimant or parents. Claimant naming follows the existing browser-stored name rules.
- Step 6, Claims interview: Confirmed that a wish may have multiple claimants.
- Step 6: User approved the Wishes draft. Its listed OPEN items remain unresolved; no implementation has started.
- Step 6: Drafted Wishes requirements and observable acceptance criteria in `wishes.md`, covering optional details, parent-only management, claimed-wish edits and deletion, and browser-persisted sorting. Approval pending; unresolved details are OPEN. Simplified the price field to one optional numeric reference amount throughout the spec. No implementation or tests written.
- Step 6, Wishes interview: Confirmed browser persistence of the user's selected sort order.
- Step 6, Wishes interview: Confirmed that missing prices are treated as zero for price sorting.
- Step 6, Wishes interview: Confirmed numeric-only price sorting.
- Step 6, Wishes interview: Confirmed alternative sort orders: newest first, title A–Z, price low–high, and price high–low.
- Step 6, Wishes interview: Confirmed added order (oldest first) as the default, with user-selectable alternatives still OPEN.
- Step 6, Wishes interview: Confirmed that wish deletion permanently removes the wish and any attached claim.
- Step 6, Wishes interview: Confirmed that editing a claimed wish leaves its existing claim attached.
- Step 6, Wishes interview: Confirmed that parents may edit or delete claimed wishes.
- Step 6, Wishes interview: Confirmed that only parents may add, edit, and delete wishes on all people's lists.
- Step 6, Wishes interview: Confirmed one optional rough reference price per wish; linked shops may have different prices.
- Step 6, Wishes interview: Replaced the single wish link with multiple shop links. Only title is required; links, price, and note are optional and can be filled in later.
- Step 6: User approved the Wishlists draft. Its listed OPEN items remain unresolved; no implementation has started.
- Step 6: Drafted Wishlists requirements and observable acceptance criteria in `wishlists.md`, including annual creation, history, renaming, reserved archived names, archiving, and manual restoration. Approval pending; unresolved details are OPEN. No implementation or tests written.
- Step 6, Wishlists interview: User reversed the earlier name-reuse decision. Archived names remain reserved; uniqueness applies across active and archived people.
- Step 6, Wishlists interview: Confirmed that archived names are not reserved and may be reused; uniqueness applies to active people.
- Step 6, Wishlists interview: Confirmed case-insensitive person-name uniqueness that ignores surrounding spaces.
- Step 6, Wishlists interview: Confirmed that person names are unique within the household.
- Step 6, Wishlists interview: Confirmed parent-managed person renaming through the app while preserving existing wishlists and claims.
- Step 6, Wishlists interview: Confirmed restoration of archived people through manual YAML edits only.
- Step 6, Wishlists interview: Confirmed that archiving hides all of the person's wishlists across all years from everyone while retaining their data.
- Step 6, Wishlists interview: Confirmed that an archived person's current-year wishlist is hidden from everyone, including parents.
- Step 6, Wishlists interview: Confirmed archiving people while retaining past lists and stopping creation of new yearly lists.
- Step 6, Wishlists interview: Confirmed automatic creation of an empty current-year wishlist when adding a person.
- Step 6, Wishlists interview: Confirmed previous years' wishlists are read-only in the web UI; direct YAML edits remain possible as an accepted risk.
- Step 6, Wishlists interview: Confirmed automatic creation of an empty wishlist for each existing person each new year.
- Step 6, Wishlists interview: Confirmed that only parents may create and edit adults' wishlists.
- Step 6: User approved the Access draft. Its listed OPEN items remain unresolved; no implementation has started.
- Step 6: Drafted Access requirements and observable acceptance criteria in `access.md` from confirmed decisions. Approval is pending; unresolved permissions, setup details, timezone editing, expiry, and priority are OPEN. No code or tests written.
- Step 6, Access interview: Confirmed automatic switching of the current page to the replacement parent link after parent-code regeneration.
- Step 6, Access interview: Confirmed a household timezone chosen during setup to determine the calendar-year switch.
- Step 6, Access interview: Confirmed immediate code revocation, including the next action from an already-open page.
- Step 6, Access interview: Confirmed a not-found page after setup for access without a valid code.
- Step 6, Access interview: Confirmed automatic generation of both initial access codes by the setup wizard.
- Step 6, Access interview: Confirmed a first-run setup wizard for initial access codes, open until setup completes.
- Step 5: Confirmed that direct YAML edits are picked up on page load.
- Step 5: Confirmed management of household members and wishlists through both the app and direct YAML edits on the VPS.
- Step 5: Confirmed independent regeneration of the parent and family access codes.
- Step 5: Confirmed that parents can regenerate both access codes at any time, disabling old links, and share the replacement links.
- Step 5: Confirmed that renaming affects future claims only; existing claims retain their original claimant name.
- Step 5: Confirmed asking for a name once, remembering it in the browser, displaying the person's own name, and allowing renaming.
- Step 5: Confirmed automatic calendar-year selection for family-link access.
- Step 5: Confirmed family-link access to all people in the current year and parent-link access to all people across all years.
- Step 5: Confirmed that v1 access links contain the code, without separate code entry.
- Step 5: Confirmed that previous years' wishlists are retained and viewable.
- Step 5: Confirmed person name as the final storage level for both adult and child wishlists.
- Step 5: Confirmed Christmas-only events in v1, retaining event type in the storage hierarchy for later expansion.
- Step 5: Confirmed YAML as the persistence format.
- Step 4: Recorded the confirmed Docker/VPS hosting, Go backend, and Christmas-themed Tailwind CSS/vanilla JavaScript frontend. Recorded the supplied file-persistence hierarchy; file format and event/person scope remain OPEN.
- Step 3: Confirmed Access, Wishlists, Wishes, and Claims as separate domains. Priorities remain OPEN; requirements and acceptance criteria are pending.
