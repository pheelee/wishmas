# Wishlists

## Status

Approved by the user, including the consolidated OPEN-review updates. In progress: person creation and current-year list viewing are verified; no whole requirement is marked Implemented while dependent behavior remains incomplete.

## Priority

Must-have for v1.

## WLS-001: People and individual lists

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

Each new calendar year, automatically create an empty Christmas wishlist for each existing, non-archived person. Use the household timezone defined in Access. Retain previous years' lists. Parents can view history, while family-link access is limited to the current year. Previous years' lists are read-only in the web UI; direct YAML edits remain possible as an accepted risk.

### Acceptance criteria

- **WLS-002-AC1:** When the calendar year changes in the household timezone, each existing, non-archived person receives an empty wishlist for the new year automatically.
- **WLS-002-AC2:** Wishes and claims from the preceding year are not copied into the new empty list.
- **WLS-002-AC3:** Previous years' lists are retained and viewable by parents, subject to the archive rules below.
- **WLS-002-AC4:** Previous years' wishlists cannot be edited through the web UI, including by parents.
- **WLS-002-AC5:** A manual YAML edit to a previous year's list is reflected on page load; the web UI's read-only behavior does not prevent that manual edit.

## WLS-003: Person names and renaming

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

The completed first slice adds parent-managed person creation and current-year list viewing. Backend behavior tests failed before implementation with missing routes/forms/manual-data rendering and absent new-year files. The browser Unicode boundary test subsequently exposed truncation of 50 non-BMP letters by HTML length limits; the fix leaves the approved character validation on the server.

| Criteria | Evidence and current state |
| --- | --- |
| WLS-001-AC1/AC2/AC8 | Verified: parent creation, empty persisted Christmas lists, added-order roster/display, parent/family viewing, and fresh-server persistence. |
| WLS-001-AC5/AC7 | Verified for the current roster and title-only list reader: manual edits reload; malformed YAML and duplicate active/archived names show errors without file changes. Existing manual list files are not overwritten. Full optional wish detail editing/display remains milestone 4. |
| WLS-001-AC3/AC4/AC6 | Parent-only creation and family restrictions verified, including stale revoked codes. Wish editing and permissions on future rename/archive actions require later verification. |
| WLS-002-AC1/AC2 | Verified: loading the new calendar year creates empty lists for active people without copying wishes or claims; old data is unchanged. |
| WLS-002-AC3–AC5 | Historical navigation and its read-only UI are not implemented yet. |
| WLS-003-AC1/AC2/AC4–AC6 | Creation validation, case-insensitive uniqueness including Unicode case pairs and archived reservations, trimming, and 3–50 Unicode-letter/digit boundaries are verified. Rename behavior remains pending; WLS-003 is not Implemented. |
| WLS-003-AC3; WLS-004 | App renaming and archiving are not implemented. The YAML archived flag is recognized for reservations, hiding current lists, and skipping new-year creation; this does not complete the archive requirement. |

Code: `internal/app/people.go`, shared routing/persistence in `internal/app/app.go`, and home/wishlist templates. Tests: `internal/app/people_test.go` and `tests/browser/people.spec.js`, with matching `@impl`/`@test` anchors. Browser verification includes desktop, mobile, inline validation, family viewing, and a 50-character Unicode name at 320px width. Go race tests also cover simultaneous duplicate creation and current-list access after link regeneration.

Checks passed: `go test -race ./...`, `go vet ./...`, Go build, CSS build, all six browser scenarios, and focused New Year/replacement-scope race checks. See `progress.md` for the exact next action.
