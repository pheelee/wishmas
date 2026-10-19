# Wishlists

## Status

Approved by the user, including the consolidated OPEN-review updates. WLS-002, WLS-003, and WLS-004 are Implemented. WLS-001 remains In progress because writing wish details depends on the Wishes implementation.

## Priority

Must-have for v1.

## WLS-001: People and individual lists

Status: In progress; person management, viewing, and permissions are verified. Wish-writing criteria depend on milestone 4.

The household has an individual wishlist for each person, adult or child. Only parents may manage people and wishlists, including children's and adults' lists. Family-link holders are limited to viewing and claiming wishes, including removing their own claims under the Claims rules. Direct YAML management is also supported.

### Acceptance criteria

- **WLS-001-AC1:** Parents can add a person through the app, whether an adult or a child.
- **WLS-001-AC2:** Adding a person automatically creates an empty current-year wishlist for that person.
- **WLS-001-AC3:** Parents can create and edit adults' lists; family-link holders cannot create or edit adults' lists.
- **WLS-001-AC4:** Parents can write their children's lists.
- **WLS-001-AC5:** Direct YAML edits to household members and wishlists are picked up on page load.
- **WLS-001-AC6:** Family-link holders cannot add, rename, or archive people, or create or edit children's wishlists.
- **WLS-001-AC7:** When manually edited YAML is invalid or contains duplicate person names under WLS-003 comparison rules, loading the affected data shows an error and leaves the files untouched.
- **WLS-001-AC8:** People's lists are displayed in the order the people were added.

## WLS-002: Annual lists and history

Status: Implemented. All acceptance criteria are verified.

Each new calendar year, automatically create an empty Christmas wishlist for each existing, non-archived person. Use the household timezone defined in Access. Retain previous years' lists. Parents can view history, while family-link access is limited to the current year. Previous years' lists are read-only in the web UI; direct YAML edits remain possible as an accepted risk.

### Acceptance criteria

- **WLS-002-AC1:** When the calendar year changes in the household timezone, each existing, non-archived person receives an empty wishlist for the new year automatically.
- **WLS-002-AC2:** Wishes and claims from the preceding year are not copied into the new empty list.
- **WLS-002-AC3:** Previous years' lists are retained and viewable by parents, subject to the archive rules below.
- **WLS-002-AC4:** Previous years' wishlists cannot be edited through the web UI, including by parents.
- **WLS-002-AC5:** A manual YAML edit to a previous year's list is reflected on page load; the web UI's read-only behavior does not prevent that manual edit.

## WLS-003: Person names and renaming

Status: Implemented. All acceptance criteria are verified.

Person names are unique across active and archived people in the household. Compare names ignoring capitalization and surrounding spaces. Archived names remain reserved. Parents can rename a person through the app while preserving existing wishlists and claims.

Trim surrounding spaces from person names. Reject blank names and names shorter than 3 or longer than 50 characters. Names may contain only Unicode letters and digits.

### Acceptance criteria

- **WLS-003-AC1:** Adding or renaming a person cannot create a name collision with another active or archived person.
- **WLS-003-AC2:** `Anna`, `anna`, and ` Anna ` count as the same name when checking uniqueness.
- **WLS-003-AC3:** Parents can rename a person through the app; that person's existing wishlists and claims remain associated with them after renaming.
- **WLS-003-AC4:** Archiving a person does not release their name for use by a new person.
- **WLS-003-AC5:** When adding or renaming a person, trim surrounding spaces and reject blank names or names outside the 3–50 character range after trimming.
- **WLS-003-AC6:** Person names accept Unicode letters and digits, including accented letters such as `é` and `ø`, and reject other characters after surrounding spaces are trimmed.

## WLS-004: Archiving and restoration

Status: Implemented. All acceptance criteria are verified.

Parents can archive a person. Retain their data, hide all their wishlists across all years from everyone, including parents, and stop creating new yearly lists for them. Restoring an archived person requires a manual YAML edit; the app provides no restoration action.

When a restored person has no current-year wishlist, automatically create an empty one on page load.

### Acceptance criteria

- **WLS-004-AC1:** Parents can archive a person through the app.
- **WLS-004-AC2:** After archiving, none of that person's wishlists are visible to parents or family-link holders, regardless of year.
- **WLS-004-AC3:** Archiving retains the person's stored data rather than deleting it.
- **WLS-004-AC4:** An archived person receives no automatically created wishlist when the calendar year changes.
- **WLS-004-AC5:** The app provides no action to restore an archived person.
- **WLS-004-AC6:** A person can be restored by manually editing YAML, with the change picked up on page load.
- **WLS-004-AC7:** On page load after restoration, if the person has no current-year wishlist, an empty one is created automatically. If one already exists, it is preserved.

## Applicable constraints

- CON-001 through CON-003: Docker/VPS, Go, and a Christmas-themed Tailwind CSS/vanilla JavaScript frontend.
- CON-004: YAML storage organized by household → year → event type → person name; Christmas only in v1; app and manual file management; changes picked up on page load.

## Related requirements

- Access: parent/family link scopes and household-timezone calendar-year selection.
- Wishes and Claims: see their owning domain requirements for detailed behavior.

## OPEN questions

None currently listed.

## Implementation and verification

Person creation, naming, annual lists, parent history, archiving, and manual restoration are verified. Creation and lifecycle/history behavioral tests failed before implementation. Additional failing browser checks exposed Unicode truncation in name fields before those fixes.

| Requirement / criteria | Evidence and state |
| --- | --- |
| WLS-001-AC1/AC2/AC5/AC7/AC8 | Creation, persisted lists, added-order display, manual roster/title reloads, and invalid-data preservation verified. Full optional wish-detail writing/display remains milestone 4. |
| WLS-001-AC3/AC4 | Parent writing of adult/child wish details remains pending Wishes implementation. WLS-001 is not marked Implemented. |
| WLS-001-AC6 | Parent-only add/rename/archive handlers, family UI restrictions, stale-code rejection, and denied historical form submissions verified. Future wish-writing permissions must retain the same rules. |
| WLS-002-AC1–AC5 | Verified: annual empty lists, unchanged past files, parent year navigation, read-only historical pages and context guards, and direct historical YAML reload. |
| WLS-003-AC1–AC6 | Verified: trimmed 3–50 Unicode-letter/digit names, case-insensitive active/archived uniqueness, case-only renames, stable added order, and raw list/claim preservation across years. |
| WLS-004-AC1–AC7 | Verified: app archive, hidden lists in all years for both roles, retained bytes/reserved names, skipped annual creation, no app restore action, and manual restore with missing/existing current lists. |

Renaming validates all affected Christmas files and destination directories before moving data. Existing list directories are moved without re-encoding wishes, claims, comments, or other file contents. All event folders are retained during the move; this preserves data without adding other event types to the v1 UI. Destination conflicts leave files untouched. Tests also exercise rollback after a later directory move fails and after roster saving fails.

Implementation: `internal/app/people.go`, `internal/app/history.go`, shared routing, and home/wishlist templates. Verification: `people_test.go`, `history_test.go`, and people/history browser scenarios, with `@impl`/`@test` anchors.

Checks passed: Go race tests, vet, build, CSS build, all eight browser scenarios, manual review of mobile management/history screenshots, and diff whitespace checks. No new dependency was added.
