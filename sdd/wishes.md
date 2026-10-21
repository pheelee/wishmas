# Wishes

## Status

WSH-001 through WSH-006 are Implemented. The picture criteria and all 17 additional translation entries are approved and verified.

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

Status: Implemented. All user-approved AC1–AC7 are verified.

The selected card reference is [Gift wrap](../design/gift-wrap/selected-baseline.png). This changes presentation of existing data and actions; the surrounding proposal-page chrome is not part of this change.

- **WSH-005-AC1:** Cards use the selected edge accents, subtle matching header gradients, gift icons, and serif wish titles. Accent colors follow the status rules in AC6–AC7.
- **WSH-005-AC2:** Claimed/unclaimed and claimed-by-you indicators appear as pill badges. Claim counts remain visible and claimant names/removal actions remain parent-only as specified in Claims.
- **WSH-005-AC3:** An optional numeric reference price receives prominent colored emphasis. Optional shop links use subtle domain labels with link icons; omitted optional fields show no invented values or empty detail placeholders.
- **WSH-005-AC4:** A separated footer contains the existing claim toggle on the left and parent edit/delete icons inline on the right. Existing modal and confirmation behavior remains unchanged.
- **WSH-005-AC5:** Cards retain the responsive desktop grid and stack on mobile without overflow, while preserving current permissions, read-only history, sorting, and action targets.
- **WSH-005-AC6:** A wish is complete when it has a title, a provided price and at least one shop link. Its status accent is red if incomplete (including when claimed), yellow if complete with no claims, and green if complete with at least one claim. A note or picture is not required for completeness.
- **WSH-005-AC7:** Status colors replace the original alternating mint/gold/ice-blue accents in AC1. Apply them to the existing border, subtle header tint, gift icon, reference price and claim-button accent styling. Keep the dark card background and existing claimed-by-you treatment.


Verification: `tests/browser/gift-wrap.spec.js` failed first on missing card headers, then passed for decorative accents, serif titles, badges, optional detail omission, footer layout, mobile fit, family privacy and read-only history without YAML changes. Existing claim, sorting and modal browser tests continue to pass, using the full accessible status label for the short visual badge. All 19 Chromium scenarios, Go race tests, vet, executable and CSS builds pass. Desktop and mobile screenshots were visually reviewed.

Header alignment correction: the live-app review exposed zero lower padding beneath the ownership badge and top-aligned header contents. Existing WSH-005-AC1/AC2/AC5 are now covered by desktop/mobile geometry assertions for badge spacing and shared icon/badge vertical centers. These assertions failed before the fix. Headers now have full interior padding, a consistent minimum height and centered contents. Seven focused Gift-wrap/polish/theme browser scenarios and Go tests pass. Reviewed the corrected CSS against live data using a browser-only asset override; the running process and household files were untouched.

Status-color update (WSH-005-AC6–AC7): the user approved data-derived red/yellow/green accents in place of the alternating palette. Completeness requires title, provided price and at least one shop link; zero is a provided price, and claimed incomplete wishes remain red. Colors follow the wish through browser sorting and are recomputed on page load after edits or claim changes. The existing dark backgrounds and claimed-by-you treatment remain in place. No YAML fields, required-input validation, permissions or translations changed.

The new HTTP test failed first for all nine status cases, then passed for both roles. Coverage includes missing price/link, optional note/picture, zero price, one/multiple claims and multiple shops. The Gift-wrap browser scenario verifies the actual border/icon/price/button colors, dark backgrounds, parent/family/history presentation, sorting, claim/unclaim changes and completeness edits. Desktop/mobile screenshots were visually reviewed. Go race tests, vet, CSS build, all 24 Chromium scenarios and diff checks pass.

## WSH-006: Optional wish picture

Status: Implemented. All user-approved WSH-006-AC1–AC15 criteria, including picture-modal backdrop dismissal and ease-in-out opening/closing, are verified. The selected shelf reference is [proposal 3](../design/wish-picture/selected-baseline.png).

A wish has at most one optional picture, provided by image URL. The selected presentation is a shallow shelf inside the card body, separate from text, header and action footer.

### Acceptance criteria

- **WSH-006-AC1:** Only parents may add, replace or remove pictures through the existing wish creation and edit forms. The input accepts a complete HTTP or HTTPS image URL; file upload is not offered. Existing historical read-only rules apply to picture changes too.
- **WSH-006-AC2:** Adding or changing the image URL downloads the source once and stores an optimized local copy. Displaying the wish or editing it with an unchanged image URL does not contact the source again. Subsequent views use the local copy.
- **WSH-006-AC3:** Downloads access public internet addresses only; private/LAN destinations are disallowed. This restriction applies to the initial URL and every redirect target.
- **WSH-006-AC4:** Downloads allow at most five redirects, a total download timeout of 30 seconds and at most 10 MiB (10,485,760 bytes) of source data. Reject source images whose decoded dimensions exceed 20 megapixels (20,000,000 pixels) before resizing.
- **WSH-006-AC5:** Accept JPEG, PNG, WebP and GIF source images. SVG is unsupported. Animated sources, including GIF and WebP, become a still image using only their first frame. Wish pictures never animate.
- **WSH-006-AC6:** Resize proportionally so neither dimension exceeds 640 pixels, without upscaling. Store lossy WebP at 80% quality, preserve transparency against the dark card background, and discard the downloaded original. Display cropping does not alter the stored optimized image.
- **WSH-006-AC7:** Store the source URL and optimized filename with the wish in YAML. Keep optimized files in a `pictures/` folder beside that person's `wishlist.yaml`, retaining the household/year/event/person hierarchy.
- **WSH-006-AC8:** Pictures use the selected image shelf below title/note and above price/claim/link details. Use a fixed centered crop to fill the shallow shelf, without a crop-position control. Pictures never span the header or action footer, overlap/dim text, or introduce horizontal overflow; text remains clearly readable on desktop/mobile.
- **WSH-006-AC9:** With no picture configured, retain the current card layout without an image shelf or placeholder.
- **WSH-006-AC10:** Clicking an available shelf picture opens the complete, uncropped optimized image in a modal. Clicking the background outside the picture modal closes it; clicking its image or interior keeps it open. Opening and closing the picture modal use an ease-in-out animation.
- **WSH-006-AC11:** Both roles may view current-year pictures, only parents may view historical pictures, and archived people's pictures are hidden. Direct image URLs require a valid access code and follow the same year/person/archive scope as the associated list.
- **WSH-006-AC12:** Download/processing failures stop the whole save with an error. Preserve all entered form values so parents can enter an alternate URL; the saved wish and picture remain unchanged. This includes unsupported/invalid sources and exceeded network/size/resolution limits.
- **WSH-006-AC13:** Clearing the image URL and successfully saving removes the picture. Leaving the URL unchanged retains the saved local copy without re-downloading it.
- **WSH-006-AC14:** Successful image replacement/removal or wish deletion immediately deletes the old local image. If the old file cannot be deleted, the whole action fails and preserves the previous wish and picture. An already-missing old file counts as already deleted and does not block replacement/removal.
- **WSH-006-AC15:** A configured but missing/unreadable local image shows a small “Image unavailable” placeholder; the rest of the wish remains usable. Do not automatically retrieve the source again. This placeholder is distinct from a wish with no image configured (AC9).

### Existing requirements still apply

Wish/claim data preservation, parent-only changes, history/archive behavior and authenticated access remain governed by the existing Wishes, Wishlists, Claims and Access criteria. New labels, validation/errors, modal and accessibility text follow the household language and INT-001–INT-005; the user approved the 17 additional wording entries before implementation.

### Constraints

CON-006 records the confirmed image processing/network/storage choices. The shelf selection does not approve the other temporary variants or their preview controls for the finished app.

### OPEN review items

None. WSH-006 criteria and the 17 new English/German wording entries are approved.

No additional product decision is OPEN. The user reviewed the completed picture feature and modal refinements and authorized committing them.


### Implementation and verification

- `picture_download.go` implements checked public-address connections (including redirect hops and DNS pinning), bounded downloads, pre-decode pixel checks, first-frame-only conversion and alpha-preserving quality-80 WebP resizing. `wishes.go` downloads outside the request mutex and rechecks current access, year, active person and edit snapshot before publishing.
- `pictures.go` stores local assets alongside YAML, authorizes each direct image request, rejects traversal/symlink/unattached access, and distinguishes absent pictures from unavailable configured files. Wish mappings retain claims, comments and future fields. Image cleanup failure restores the exact previous YAML using a staged same-directory rollback file; missing old files do not block changes.
- The existing create/edit forms accept image URLs and retain all submitted values on failure, including title whitespace and multiline link formatting. The selected shelf remains below title/note and above details; fixed centered cover cropping affects display only. Native modals show the complete optimized image. Local image load errors show the approved placeholder without remote retries.
- The first new behavioral test failed on the missing image URL input. Go tests cover formats, GIF/WebP first frames, resize/no-upscale/alpha, byte/pixel/redirect/time limits, public/private destinations, successful create/replace/clear/delete, unchanged-source retention, permissions, missing/unreadable files and rollback under actual filesystem errors. Blocked-download tests verify concurrent GETs and rejection after code rotation, archive, case-only rename, year rollover or a changed list snapshot.
- `go test -race ./...`, `go test -tags nodynamic -race ./...`, `go vet ./...`, CSS build, JavaScript syntax and diff checks pass. All 24 Chromium scenarios pass, including English/German picture forms, exact retained drafts, desktop/mobile geometry, modal viewing, family access and cleanup. The focused picture scenarios also pass after final draft-preservation changes. The mobile modal screenshot and the final sample WebP were visually reviewed.
- English and German DinD smoke checks pass with the supplied public HTTPS sample: actual download/Go-only WebP encoding in non-root scratch, authenticated serving, current-year roles, claims, named-volume picture/YAML persistence after restart and container replacement, and immediate image removal. Docker includes CA certificates and builds with `nodynamic` so no system image library/dynamic loader is required.
- Temporary variant code, controls and review artifacts were removed. Only the approved shelf reference is retained. At the user's request, Leon's selected sample received its source URL and the final 43,894-byte optimized image; the other wishlist contents were checked unchanged. The existing live process was not restarted. No VPS deployment was performed.

Picture-modal backdrop dismissal (WSH-006-AC10): the user requested closing the image modal on background clicks. The added browser check failed first while the modal remained open. A picture-only click handler now closes it for coordinates outside the dialog bounds, while image/interior clicks keep it open. English/German desktop/mobile picture tests verify outside dismissal, interior clicks and Escape; both existing polish scenarios also pass (four scenarios total). JavaScript syntax and diff checks pass. The completed work is approved for commit; no live restart was performed.

Picture-modal animation (WSH-006-AC10): opening and closing now use a 180 ms ease-in-out fade and gentle scale with a matching backdrop fade. Close, outside clicks and Escape share animated dismissal; the native modal stays open until the exit finishes. Reduced-motion preferences skip animation. The browser test failed first with no animation, then all four picture/polish scenarios passed, including actual opening/closing animation checks, reduced-motion bypass, both languages and desktop/mobile dismissal. CSS build, JavaScript syntax and diff checks pass. No live restart was performed.
