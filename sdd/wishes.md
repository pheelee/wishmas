# Wishes

## Status

WSH-001 through WSH-005 are Implemented. All acceptance criteria are verified.

## Priority

Must-have for v1.

## WSH-001: Wish details

Status: Implemented. All acceptance criteria are verified.

A wish requires only a title. Multiple shop links, one numeric rough reference price, and a note are optional and can be filled in later. Linked shops may have different prices; prices are not stored separately per link.

Trim surrounding spaces from titles. Titles must be nonblank and contain 5–150 characters after trimming.

Optional prices accept zero and positive amounts with up to two decimal places; negative amounts are rejected.

Price entry uses a decimal point only, such as `12.50`; decimal commas are not accepted.

Shop links require complete HTTP or HTTPS URLs; bare addresses are not automatically prefixed.

Display each shop link using its domain name.

Display shop links in the order they were added.

### Acceptance criteria

- **WSH-001-AC1:** A parent can save a wish containing only a title, leaving links, price, and note blank.
- **WSH-001-AC2:** A wish without a title cannot be saved.
- **WSH-001-AC3:** A parent can save multiple shop links on the same wish.
- **WSH-001-AC4:** A parent can add optional links, price, and note to an existing wish later.
- **WSH-001-AC5:** A wish has one optional numeric reference price for the whole wish, with no per-link price fields.
- **WSH-001-AC6:** Creating or editing a wish trims surrounding spaces from its title and rejects blank titles or titles shorter than 5 or longer than 150 characters after trimming.
- **WSH-001-AC7:** An optional price accepts zero and positive amounts with at most two decimal places and rejects negatives or amounts with more than two decimal places.
- **WSH-001-AC8:** Price entry accepts a decimal point, as in `12.50`, and rejects decimal-comma input such as `12,50`.
- **WSH-001-AC9:** Shop links accept complete HTTP or HTTPS URLs and reject bare addresses such as `shop.example/item` without adding a scheme automatically.
- **WSH-001-AC10:** A shop link such as `https://shop.example/item` is displayed with the label `shop.example` and opens its complete saved URL.
- **WSH-001-AC11:** Multiple shop links on a wish appear in the order they were added.

## WSH-002: Management and claims

Status: Implemented. All acceptance criteria are verified.

Only parents may add, edit, and delete wishes on any person's list. Parents may edit or delete a claimed wish. Editing preserves the claim. Deletion permanently removes the wish and any attached claim. Previous years' wishlists remain read-only in the web UI as specified in Wishlists.

Deleting a wish requires confirmation before permanent removal.

Editing or deleting a claimed wish sends no notifications in v1.

### Acceptance criteria

- **WSH-002-AC1:** Parents can add, edit, and delete wishes on current-year lists for adults and children.
- **WSH-002-AC2:** Family-link holders cannot add, edit, or delete wishes.
- **WSH-002-AC3:** A parent can edit a claimed wish without clearing or replacing its existing claim.
- **WSH-002-AC4:** A parent can delete a claimed or unclaimed wish; the wish and any attached claim are permanently removed.
- **WSH-002-AC5:** Previous years' lists do not allow adding, editing, or deleting wishes through the web UI, including for parents.
- **WSH-002-AC6:** Requesting wish deletion asks for confirmation; confirming permanently removes the wish and its claims, while cancelling leaves them unchanged.
- **WSH-002-AC7:** Editing or deleting a claimed wish does not send notifications in v1.

## WSH-003: Sorting

Status: Implemented. All acceptance criteria are verified.

Default to added order, oldest first. Users may switch to newest first, title A–Z, price low–high, or price high–low. Price sorting compares numeric amounts and treats a missing price as zero. Remember the selected sort order in the user's browser.

Equal prices or titles are ordered by added order, oldest first.

Title A–Z sorting ignores capitalization and treats accented letters like their base letters.

Use one browser-stored sort preference across all wishlists.

### Acceptance criteria

- **WSH-003-AC1:** With no saved sort preference, a list displays wishes in added order, oldest first.
- **WSH-003-AC2:** Users can select newest first, title A–Z, price low–high, or price high–low, and the displayed wishes follow that order.
- **WSH-003-AC3:** Numeric prices `2`, `10`, and `100` sort in that sequence for low–high and in reverse for high–low.
- **WSH-003-AC4:** A wish without a price is treated as having price zero when determining its position in price sorting.
- **WSH-003-AC5:** After selecting a sort order and reloading the page in the same browser, that selection remains in effect.
- **WSH-003-AC6:** Wishes with equal price or title sort values retain added order, oldest first, as their tie-breaker.
- **WSH-003-AC7:** Title sorting compares letters without regard to capitalization and treats accented letters such as `é` like their base letters such as `e`; equivalent titles use the added-order tie-breaker.
- **WSH-003-AC8:** Selecting a sort order on one wishlist applies that preference to other wishlists opened in the same browser.

## Applicable constraints

- CON-001 through CON-003: Docker/VPS, Go, and a Christmas-themed Tailwind CSS/vanilla JavaScript frontend.
- CON-004: YAML persistence, direct file edits picked up on page load, and read-only past years in the web UI.

## Related requirements

- Access: parent and family permissions and current-year selection.
- Wishlists: person ownership, archives, and historical lists.
- Claims: see the owning domain requirements for claim behavior.

## OPEN questions

None currently listed.

## Implementation and verification

The creation/editing slice began with behavioral tests failing on absent save routes (404). Implementation in `internal/app/wishes.go`, shared routing, and wishlist templates carries criterion anchors; Go tests in `wishes_test.go` and the browser scenario `wishes.spec.js` carry matching anchors.

- WSH-001-AC1–AC11: verified title-only creation, later optional details, Unicode title boundaries/trimming, decimal validation, complete HTTP/HTTPS links, domain labels, and saved link order. Optional details can also be cleared.
- WSH-002-AC1–AC7: verified parent-only current-year creation/editing/deletion, preserved claims during edits, and permanent removal of claimed/unclaimed wishes. A server-rendered confirmation page names the wish and explains attached-claim removal; Cancel returns to the list without writing. Only explicit confirmation removes data. Other wishes, claims, and historical files remain intact. No notification side effects exist.
- Deletion guards cover family access, historical form/query years, year rollover, archived/missing people, missing/stale tokens, changed claims, reordering, invalid data, revoked codes, GET requests, and replay after removal. A regression test first reproduced repeated removal of identical wishes; whole-list version fingerprints now prevent that replay.
- WSH-003-AC1–AC8: verified all five orders, numeric comparison (including large amounts separated by one cent), missing-as-zero, stable price/title ties, case/accent equivalence including decomposed accents, reload persistence, shared preference across people/roles/years, and independent defaults in a fresh browser. Sorting does not modify saved order or YAML. Invalid/unavailable browser storage falls back to added order and still permits sorting on the open page.
- Sorting moves entire wish cards, preserving attached editing/deletion forms and targets. Browser regression coverage edits and deletes after sorting and then verifies original added order.

Raw YAML mappings retain claims and future fields when editable details change. A whole-list content/position fingerprint rejects stale edits and deletion confirmations before writing. Any list change requires reloading a stale form. Invalid data and rejected submissions leave files untouched. Direct YAML edits reload and undergo the same detail validation. Price strings preserve entered decimal precision; validation enforces numeric amounts.

Verification: Go race tests, vet, executable build, CSS build, Chromium flows, and mobile screenshot review. Browser verification exposed and fixed accessible labels on filled textareas. Deletion tests in `delete_wish_test.go` failed on the absent route before implementation; `delete-wish.spec.js` verifies confirmation, actual cancellation, claimed/unclaimed removal, stale confirmation rejection, family/history restrictions, and the mobile dark layout. Sorting scenarios in `tests/browser/sorting.spec.js` failed on the missing control before implementation. The template and vanilla JavaScript carry WSH-003 anchors. Numeric sorting compares integer hundredths with BigInt; title sorting uses base-sensitive collation; equal values use original saved indexes. The shared dark-surface checks now include selects. Mobile screenshot review and the full browser suite cover the new control. All Wishes criteria are now verified; Claims UI remains a separate milestone.

## WSH-004: Wish-card presentation and editing

Status: Approved by the user and Implemented.

- **WSH-004-AC1:** Each wish is displayed in its own card for better visibility.
- **WSH-004-AC2:** Edit wish and Delete wish are icon actions placed inline at the bottom of the card.
- **WSH-004-AC3:** Edit wish opens a modal rather than an inline editing panel.
- **WSH-004-AC4:** Wish-deletion confirmation appears in a modal. Cancelling leaves the wish and claims unchanged; confirming retains permanent deletion behavior.
- **WSH-004-AC5:** The wish creation form remains inline.

Existing parent-only permissions, historical read-only behavior, claim preservation, deletion confirmation, and sorting requirements remain in force.

WSH-004 verification: `polish.spec.js` checks individual cards, icon actions, modal editing/cancellation and retained inline creation. Updated wishes, deletion, and sorting browser flows verify modal validation drafts, permanent deletion/cancellation, stale confirmations, permission guards, and correct action targets after sorting.

## WSH-005: Gift-wrap card styling

Status: Implemented. User-approved acceptance criteria are verified.

The selected card reference is [Gift wrap](../design/gift-wrap/selected-baseline.png). This changes presentation of existing data and actions; the surrounding proposal-page chrome is not part of this change.

- **WSH-005-AC1:** Cards use the selected mint/gold/ice-blue edge accents, subtle matching header gradients, gift icons, and serif wish titles.
- **WSH-005-AC2:** Claimed/unclaimed and claimed-by-you indicators appear as pill badges. Claim counts remain visible and claimant names/removal actions remain parent-only as specified in Claims.
- **WSH-005-AC3:** An optional numeric reference price receives prominent colored emphasis. Optional shop links use subtle domain labels with link icons; omitted optional fields show no invented values or empty detail placeholders.
- **WSH-005-AC4:** A separated footer contains the existing claim toggle on the left and parent edit/delete icons inline on the right. Existing modal and confirmation behavior remains unchanged.
- **WSH-005-AC5:** Cards retain the responsive desktop grid and stack on mobile without overflow, while preserving current permissions, read-only history, sorting, and action targets.

Verification: `tests/browser/gift-wrap.spec.js` failed first on missing card headers, then passed for decorative accents, serif titles, badges, optional detail omission, footer layout, mobile fit, family privacy and read-only history without YAML changes. Existing claim, sorting and modal browser tests continue to pass, using the full accessible status label for the short visual badge. All 19 Chromium scenarios, Go race tests, vet, executable and CSS builds pass. Desktop and mobile screenshots were visually reviewed.

Header alignment correction: the live-app review exposed zero lower padding beneath the ownership badge and top-aligned header contents. Existing WSH-005-AC1/AC2/AC5 are now covered by desktop/mobile geometry assertions for badge spacing and shared icon/badge vertical centers. These assertions failed before the fix. Headers now have full interior padding, a consistent minimum height and centered contents. Seven focused Gift-wrap/polish/theme browser scenarios and Go tests pass. Reviewed the corrected CSS against live data using a browser-only asset override; the running process and household files were untouched.
