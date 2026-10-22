# Wishmas specification

## Revision status

Approved household refactor on `refactor/household-lists-access`; baseline `main` is `1992bdf`. Domain behavior, storage/authentication contract and resumable milestone plan are approved. All 112 new/revised wording entries are approved. Milestones 1–2 implement offline data/hash/migration tools and default password setup/admin authentication. Management UI and actual household migration remain pending. Old implementation reports in Git history describe the baseline, not this target.

## Product intent

- Primarily a family app; v1 remains one household. Multiple households for other families remain a later public-use possibility.
- Setup creates the household, shared admin password, household language and appearance. New households start empty.
- The household owns wishes independently of lists. A wish can remain unassigned or belong to exactly one selected recipient in one list.
- Parents manually create freely named lists and choose their recipients from the household roster. No automatic year/date lifecycle remains; every list is editable.
- Household users and recipients are separate rosters. Each household user has an individual coded link for every list; parents choose which links to share. There is no per-list membership management.
- Admins manage lists, rosters, wishes, assignments and claims. List users see all recipients/wishes in their link's list and claim as the predefined user.
- Keep title-only wish creation, optional links/reference price/note/picture, existing validation, picture processing, cards, sorting and recipient summary badges.

## Principles

- Shared-password admin access replaces parent links; YAML stores Argon2id hashes. Password changes are manual YAML edits and invalidate all admin sessions.
- Admin sessions persist until logout/hash change and across restarts/replacement with retained storage. Logout affects the current browser only.
- Individual user links identify stable household users and grant only one list. Regeneration invalidates only that user/list link on its next request, preserving claims/other links.
- Users do not enter/change claimant names. Claims use user IDs; renaming updates displayed names without changing ownership.
- Multiple users may claim one wish, once each. Users remove only their own claims; admins remove any claim. Admins select a user per new claim, on assigned wishes only.
- Parents see all claimant names. List users see counts and their own claimed state without other names. Refresh on page load; no notifications/polling.
- Assignment/moves/manual unassignment preserve claims. Deleting lists or removing/deleting recipients clears the affected assigned wishes' claims as those wishes return to the pool. Deleting a household user clears all their claims. Permanent wish deletion removes wish/assignment/claims/picture.
- Destructive list/user/recipient operations require explanatory confirmation; direct reassignment does not.
- One manually editable household YAML file with stable IDs stores all state; images remain separate. Invalid manual data/errors leave existing data untouched. Language and appearance remain household-wide.
- Preserve English (en-US), informal Swiss German (de-CH), December-night theme, Christmas bulbs and restrained snow. Free list names do not change the Christmas theme.
- Work stays on the separate branch. Migration/review uses copied data. Every verified milestone stops for user approval; merge/live cutover is a separate decision.

## Actors

- **Parent/admin:** anyone authenticated with the household's shared password; manages the household and claims on behalf of predefined users.
- **Household user:** parent-defined gift-giver identified by an individual user/list code; views/claims within that list only.
- **Recipient:** adult/child receiving gifts; a separate household record selected for lists, not a login identity.

## Domains

| Domain | Owns | Priority | Status |
| --- | --- | --- | --- |
| [Access](access.md) | Setup/password sessions, household users, individual list links and role scopes | Must-have | In progress; authentication, links and wish scopes verified; claim actions pending |
| [Wishlists](wishlists.md) | Manual lists, recipient roster/selection, assignment/lifecycle and recipient summaries | Must-have | In progress; lists/rosters/assignments/summaries verified; retained-claim removal pending |
| [Wishes](wishes.md) | Household wish details/CRUD, retained cards, sorting and optional pictures | Must-have | In progress; household wishes/cards/pictures verified; claim-toggle footer pending |
| [Claims](claims.md) | User-ID ownership, claim creation/removal, privacy and lifecycle integration | Must-have | Behavior approved; refactor not implemented |
| [Internationalization](internationalization.md) | Household/setup/login language, bundled wording, formatting and collation | Must-have | Setup/login integration verified; complete management UI pending |

[Constraints](constraints.md), [approved storage contract](storage.md), [glossary](glossary.md), [implementation plan](implementation.md), [progress](progress.md) and [changes](changes.md) accompany the domains. The approved specification checkpoint is recorded in progress.md; implementation is paused before milestone 3.

## Removed from the target

Parent/shared-family links, timezone configuration, event/year/person directory ownership, automatic yearly list creation, current-year access restrictions, read-only history, archive/restore, browser-entered claimant names and per-list user memberships. Remove unused controls/messages/runtime code; legacy reads exist only in the explicit migration tool.

## Out of scope

Separate parent accounts, user self-registration/name entry, multiple households in v1, currency, notifications/live refresh, languages beyond English/German, automatic content translation and automatic list/date lifecycle. No merge or actual VPS cutover is approved by the refactor plan.

## OPEN items

No product, schema, wording or milestone decision remains OPEN. Deployment-specific trusted Traefik addresses must be supplied before deployment. Implementation evidence remains pending.
