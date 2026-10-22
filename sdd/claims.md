# Claims

## Status and priority

Household refactor: **behavioral acceptance criteria approved by the user**. Must-have. **Implemented and verified in milestone 5.**

CLM-001–CLM-004 are retired from the target specification: browser-entered names, name-based identity/ownership, future-only renames and current-year claim restrictions no longer apply. Git history retains the previous requirements/evidence; their implementation tests must be replaced or adapted without preserving obsolete behavior.

## CLM-005: Household-user identity

- **CLM-005-AC1:** A valid individual list link identifies one household user. Claim ownership and duplicate checks use that stable user ID, not an entered name, browser name cookie or posted claimant name.
- **CLM-005-AC2:** No name-entry or user-controlled rename flow appears for list users. Show their predefined name; changes are made by parents in the household roster.
- **CLM-005-AC3:** Household-user renaming preserves claims/ownership and updates the name displayed on existing claims. Different users remain distinct regardless of similar names across the separate recipient roster.
- **CLM-005-AC4:** The same user/list link represents the same user across browsers. Opening another valid user link acts as that other user in that view. A separate admin session does not override the link's user identity.

## CLM-006: Creating claims

- **CLM-006-AC1:** A list user can claim a wish currently assigned to that link's list. No wish-date or list-year restriction applies.
- **CLM-006-AC2:** A wish may have multiple household users as claimants, but each user ID can claim a wish only once. Duplicate attempts cannot add duplicate claims.
- **CLM-006-AC3:** List-user claiming is a direct toggle without confirmation or name entry. Existing claims by other users remain unchanged.
- **CLM-006-AC4:** Admins may create a claim on an assigned wish by selecting a household user separately for that claim. The claim belongs to that selected user and is subject to the same per-user duplicate rule. No acting-user selection carries to the next claim.
- **CLM-006-AC5:** No new claim can be created on an unassigned wish, including by an admin. Preserved claims on unassigned wishes remain visible/removable by admins.
- **CLM-006-AC6:** Failed claim attempts show an inline error and refreshed status without changing existing claims. Preserve the existing stale/duplicate/write-failure protection; do not silently invert a stale claim action.

## CLM-007: Visibility and refresh

- **CLM-007-AC1:** Show claimed/unclaimed status and claimant count. At least one claim makes the wish claimed; multiple claimants still represent one claimed wish in WLS-011 summaries.
- **CLM-007-AC2:** List users see whether they personally claimed a wish and the number of claimants, but do not receive other claimants' names in the page or its responses. They see their own predefined name.
- **CLM-007-AC3:** Admins see all claimants' current household names. Admin access has no browser-stored claimant identity; each new admin claim uses explicit user selection.
- **CLM-007-AC4:** Retain page-load refresh only: other viewers' changes and manual YAML edits appear on the next page load. Do not add polling, live updates or notifications.
- **CLM-007-AC5:** Manually unassigning/reassigning a wish preserves its claims. Unassigned wishes/claims are admin-only; reassigned claims appear under the destination recipient/list. Known old links or wish/image URLs cannot bypass the new scope.

## CLM-008: Removing claims and lifecycle effects

- **CLM-008-AC1:** A list user can remove only their own claim, identified by the link's user ID, and only while the wish is assigned to that link's list. Submitted IDs/names cannot grant removal of another user's claim or access to another list.
- **CLM-008-AC2:** Admins can remove any claim from assigned or unassigned household wishes without choosing an acting user. Retain direct claim removal without confirmation.
- **CLM-008-AC3:** Removing one claim preserves other users' claims. Removing the last makes the wish unclaimed and updates its badge/status on the resulting page.
- **CLM-008-AC4:** Household-user deletion removes all their claims, including on unassigned wishes, and invalidates their links. User deletion requires the confirmation specified by Access.
- **CLM-008-AC5:** Deleting a list clears claims only from wishes currently assigned to that list and returns those wishes to the unassigned pool. Users and claims on wishes elsewhere remain unchanged.
- **CLM-008-AC6:** Removing a recipient from a list clears claims from their wishes in that list as those wishes become unassigned. Deleting a household recipient applies this across all their list assignments. Do not confuse these destructive operations with manual unassignment, which preserves claims.
- **CLM-008-AC7:** Permanent wish deletion removes its claims. Wish edits, renames, direct reassignment, manual unassignment and user-link regeneration preserve claims.

## Relationships and constraints

- Access owns household users, password sessions and user/list links, including immediate revocation and user deletion.
- Wishlists owns recipient/list lifecycle, assignment and destructive confirmations. All users automatically have links for all lists; there is no list-membership prerequisite for moving a claimed wish.
- Wishes owns permanent deletion, detail preservation, pictures and status colors.
- Persistence uses stable user IDs and one household YAML data file; invalid manual data/rejected mutations retain error-without-writes behavior. Migrate existing named claims into household users before changing ownership semantics.
- No automatic year switch, read-only history, archived recipients, browser identity cookie or user-editable claim name remains in the target behavior.

## OPEN review items

No behavioral or wording item remains OPEN. New English/German wording, including admin user-selection labels, is approved in translations.md. Technical constraints and implementation milestones are approved separately.

## Implementation and verification

CLM-005–008 are implemented by `household_claims.go`, authenticated routing in `admin.go`, scoped presentation in `admin_management.go`/`household_wishes.go` and the retained card footer. Creation/removal operates on explicit modes and stable user IDs. Admin selection is per claim; names/cookies/form fields cannot change a link's identity. Unassigned creation is rejected, while admins can remove preserved claims. Every action reloads/validates the household and requires HTTPS/CSRF plus a current revision. Failures show refreshed authorized status inline, never silently reverse a stale action.

The first behavioral test failed because user cards lacked a claim toggle. `household_claims_test.go` now verifies independent/same-user links, switching identity within a browser, multiple/duplicate/stale claims, admin per-claim selection, no year restriction, renamed ownership/current names, forged IDs/names/legacy browser cookies, unassigned/cross-list denial, immediate code revocation, concurrent independent handlers and real permission-denied writes. Invalid manual YAML returns an English error without rewriting data. Lifecycle tests create real claims before edits/moves/unassignment and all destructive cleanup paths.

`household-claims.spec.js` verifies English/German HTTPS in independent browser contexts: direct user toggles, same-link stale add rejection, multiple counts, full-response privacy, page-load refresh, explicit admin choices and localized required validation, direct arbitrary/admin unassigned removals, renamed claimants, moved claims, regenerated already-open forms and deleted claimant links. Card controls fit 1440/840/390 widths and mobile picker/user screenshots were reviewed. Existing sorting/images/summary tests remain passing. No polling or notifications were added.

Go race/vet/static builds, CSS/syntax/diff checks and all 38 browser scenarios pass. Actual copied-household migration, obsolete legacy-code/catalog removal and container/Traefik proof remain milestone 6. [Review walkthrough](../docs/admin-access.md#milestone-5-review). Pause before advancing.
