# Wishlists

## Status

Household refactor: **behavioral acceptance criteria approved by the user**. Not implemented. All criteria below derive from confirmed interview answers. The existing application still implements the previous model.

WLS-001–WLS-006 are superseded in this revision. Their annual lists, automatic creation, history restrictions, archive/restore and year-scoped summaries no longer describe the target behavior. Git history retains the previous specifications and evidence. Retained summary presentation is specified below in its new list/recipient context; old tests are not proof of the revised scope.

## Priority

Must-have. Admin-only management; household users view recipients and wishes through their individual list links.

## WLS-007: Manually managed lists

- **WLS-007-AC1:** Parents create lists within the household using free-text names, including names such as Christmas 2026 or a birthday. A list is not restricted to a Christmas event type or tied to an automatic calendar year.
- **WLS-007-AC2:** Trim list names; accept 3–100 characters including spaces and punctuation. Require case-insensitive uniqueness within the household, ignoring surrounding spaces.
- **WLS-007-AC3:** Parents rename lists without changing stable identity, wish assignments, claims or user links.
- **WLS-007-AC4:** Lists remain viewable and editable regardless of dates in their names. Do not automatically create lists at New Year or make older lists read-only. Remove year navigation and timezone configuration associated with the previous model.
- **WLS-007-AC5:** Each new list starts with no recipients or assigned wishes. Each existing household user has an individual link for the new list under ACC-007.
- **WLS-007-AC6:** Parents delete a list only after confirmation explaining the affected wishes and claims. Return its currently assigned wishes to the household's unassigned pool and clear those wishes' claims; retain the wishes and pictures.
- **WLS-007-AC7:** Deleting a list invalidates all links to that list. Household users, household recipients, other lists, their assignments, claims and links remain unchanged.
- **WLS-007-AC8:** Lists appear in added order. Migration supplies lists in ascending year order.

## WLS-008: Household recipient roster

- **WLS-008-AC1:** Only parents create, rename and delete household recipients. A recipient can represent an adult or child and is independent of the household user roster.
- **WLS-008-AC2:** Trim recipient names; accept 3–50 characters consisting of Unicode letters, digits and spaces. Require uniqueness within the recipient roster ignoring capitalization and surrounding spaces. User and recipient rosters may contain the same name without a connection.
- **WLS-008-AC3:** Recipients have stable IDs. Renaming preserves their list selections, assigned wishes and claims. Display recipients in added order.
- **WLS-008-AC4:** Creating a household recipient does not automatically select them for existing lists. There is no recipient archive or restoration feature.
- **WLS-008-AC5:** Deleting a household recipient requires confirmation explaining affected wishes/claims. Remove that recipient from every list, unassign all their assigned wishes and clear those wishes' claims. Keep the wishes and pictures in the household pool. Separate household users, including a user with the same name, are unaffected.

## WLS-009: Recipients within a list

- **WLS-009-AC1:** Parents explicitly add recipients from the household roster to each list. Lists can select the same household recipient independently.
- **WLS-009-AC2:** Parents remove a recipient from a list only after confirmation explaining affected wishes/claims. Unassign that recipient's wishes in that list and clear those wishes' claims; retain the wishes and pictures.
- **WLS-009-AC3:** Removing a recipient from one list does not delete them from the household roster or affect their other-list selections, wishes or claims.
- **WLS-009-AC4:** A valid user link shows all selected recipients in its one list, including their assigned wishes. Users cannot manage recipients or lists. There is no recipient filtering by viewer identity or connection between recipient and user names.

## WLS-010: Household wish assignment

- **WLS-010-AC1:** A wish belongs to the household independently of any list. It may be unassigned, including when saved for a future list.
- **WLS-010-AC2:** An assigned wish belongs to exactly one recipient in exactly one list at a time. That recipient must be selected for the destination list.
- **WLS-010-AC3:** Parents assign and unassign wishes. Reassigning an already assigned wish directly moves it to the selected list/recipient without a confirmation step or copy; it never appears as assigned in two places.
- **WLS-010-AC4:** Assignment, reassignment and manual unassignment preserve existing claims. Do not require claimants to be added to a destination list: every household user already has a link for every list.
- **WLS-010-AC5:** Unassigned wishes and their retained claims are visible/manageable only to parents in the household context. User links cannot access these wishes or their pictures. Parents may remove any retained claim.
- **WLS-010-AC6:** Deleting a wish permanently removes it, its claims and its assignment, including from the previously assigned recipient/list. Existing picture deletion behavior remains in force.
- **WLS-010-AC7:** The household page shows all household wishes with their list/recipient or Unassigned state. Parents can filter by list, recipient and unassigned wishes.
- **WLS-010-AC8:** Parents can create, edit and delete wishes from the household page and a recipient's page within a list. Creation on a recipient's list page automatically assigns the new household wish there. Detail validation, deletion confirmation and claim preservation on edits remain governed by Wishes.

## WLS-011: Recipient summaries within a list

- **WLS-011-AC1:** Both admins and list users see each selected recipient's total reference price and claimed/unclaimed wish counts, calculated only from wishes currently assigned to that recipient in that list.
- **WLS-011-AC2:** Sum all supplied reference prices, including incomplete wishes. Missing prices contribute zero. Totals are ungrouped with exactly two decimals and no currency. An empty recipient collection displays 0.00 and zero counts. Individual wish prices retain INT-005 formatting.
- **WLS-011-AC3:** Count only complete wishes: title, supplied price (including zero) and at least one shop link. At least one claimant counts as one claimed wish, irrespective of claimant count. A complete wish without claims counts as unclaimed. Incomplete wishes contribute to neither count.
- **WLS-011-AC4:** Retain approved badge layout B: prominent reference total, green claimed badge and yellow unclaimed badge on the existing dark theme. Each recipient entry links to their wishes within that list.
- **WLS-011-AC5:** On narrow screens, names and reference totals occupy separate rows; badges wrap. Long names/values remain readable without horizontal overflow.
- **WLS-011-AC6:** Recompute on page load from current household data, including assignment, wish, price, claim and manual YAML changes. Do not expose other claimant names to list users. Keep the four approved summary phrases unless wording is separately revised and approved.

## Applicable constraints and related domains

- Retain Go, Tailwind/vanilla JavaScript, Docker/VPS, bundled English/German and the approved December-night appearance.
- Persistence will be one household YAML data file using stable IDs, with separate pictures. The previous year/event/person directory model is superseded.
- Access owns household users, admin sessions and per-user/list codes. Users cannot be removed from individual lists because list memberships do not exist.
- Claims owns identity, privacy, admin selection of a user per claim and removal permissions.
- Wishes retains existing details, validation, sorting, card appearance and image processing. Assignment does not duplicate the underlying wish.
- Creation forms remain inline. Edit/action forms retain modal presentation; confirmed destructive operations explain consequences. Claim toggles retain their direct behavior.

## OPEN review items

- Behavioral criteria and new English/German wording are approved. No product decision remains OPEN.
- YAML schema, migration safeguards and the resumable milestone plan are approved in storage.md, constraints.md and implementation.md; implementation evidence is pending.

## Implementation and verification

Not implemented. Planned failing-first checks cover empty/manual lists without time dependence, name validation/renames/stable links, recipient selection, one-place assignment/moves, retained claims on manual unassignment, destructive cleanup distinctions, role boundaries and scoped summaries. Existing styling/decimal/sorting tests must be retained or adapted without preserving obsolete year/archive restrictions. Pause for user review after each verified milestone.
