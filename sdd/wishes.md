# Wishes

## Status

Household refactor: **behavioral acceptance criteria approved by the user**. Not implemented in the new context. The user confirmed retention of existing wish fields, validation, pictures, sorting and card appearance. Retained criteria below remain product requirements, but baseline verification is not evidence of integration with the revised model.

WSH-002's current-year/history management rules are retired and replaced by WSH-007. In WSH-004–WSH-006, obsolete history/archive wording and per-person picture storage are removed. WLS-010 owns assignment behavior; WLS-011 owns recipient summaries.

## Priority

Must-have.

## WSH-001: Wish details

Status: Retained approved behavior; integration with the new household model awaits implementation/verification.

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

## WSH-007: Household-owned wish management

- **WSH-007-AC1:** Only admins create, edit and delete household wishes, whether assigned to a list/recipient or unassigned. Individual list users cannot change wish details, assignments or pictures.
- **WSH-007-AC2:** The household page and recipient-list pages provide wish creation/edit/deletion. Creating from a recipient-list page assigns the household wish there; household wishes may be saved unassigned for future lists.
- **WSH-007-AC3:** Parents can edit a claimed wish without clearing/replacing its claims. Editing from any authorized context changes the same household wish; assignments do not create copies.
- **WSH-007-AC4:** Confirming deletion permanently removes the wish, claims, assignment and optimized picture under the existing image-deletion rules. Cancelling leaves all unchanged. No historical date restriction applies.
- **WSH-007-AC5:** Editing or deleting claimed wishes sends no notifications. Preserve existing no-polling/page-load refresh behavior.

## WSH-003: Sorting

Status: Retained approved behavior; integration with the new household model awaits implementation/verification.

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

## WSH-004: Wish-card presentation and editing

Status: Retained approved behavior; new access/context integration pending.

- **WSH-004-AC1:** Each wish is displayed in its own card for better visibility.
- **WSH-004-AC2:** Edit wish and Delete wish are icon actions placed inline at the bottom of the card.
- **WSH-004-AC3:** Edit wish opens a modal rather than an inline editing panel.
- **WSH-004-AC4:** Wish-deletion confirmation appears in a modal. Cancelling leaves the wish and claims unchanged; confirming retains permanent deletion behavior.
- **WSH-004-AC5:** The wish creation form remains inline.

Parent-only management, claim preservation, deletion confirmation and sorting remain in force. Remove historical read-only restrictions.

## WSH-005: Gift-wrap card styling

Status: Retained approved appearance; new access/context integration pending.

The selected card reference is [Gift wrap](../design/gift-wrap/selected-baseline.png). This changes presentation of existing data and actions; the surrounding proposal-page chrome is not part of this change.

- **WSH-005-AC1:** Cards use the selected edge accents, subtle matching header gradients, gift icons, and serif wish titles. Accent colors follow the status rules in AC6–AC7.
- **WSH-005-AC2:** Claimed/unclaimed and claimed-by-you indicators appear as pill badges. Claim counts remain visible and claimant names/removal actions remain parent-only as specified in Claims.
- **WSH-005-AC3:** An optional numeric reference price receives prominent colored emphasis. Optional shop links use subtle domain labels with link icons; omitted optional fields show no invented values or empty detail placeholders.
- **WSH-005-AC4:** A separated footer contains the existing claim toggle on the left and parent edit/delete icons inline on the right. Existing modal and confirmation behavior remains unchanged.
- **WSH-005-AC5:** Cards retain the responsive desktop grid and stack on mobile without overflow, while preserving new admin/user permissions, sorting and action targets.
- **WSH-005-AC6:** A wish is complete when it has a title, a provided price and at least one shop link. Its status accent is red if incomplete (including when claimed), yellow if complete with no claims, and green if complete with at least one claim. A note or picture is not required for completeness.
- **WSH-005-AC7:** Status colors replace the original alternating mint/gold/ice-blue accents in AC1. Apply them to the existing border, subtle header tint, gift icon, reference price and claim-button accent styling. Keep the dark card background and existing claimed-by-you treatment.
## WSH-006: Optional wish picture

Status: Existing image behavior retained; storage and access scope approved below, with integration pending. The selected shelf reference is [proposal 3](../design/wish-picture/selected-baseline.png).

A wish has at most one optional picture, provided by image URL. The selected presentation is a shallow shelf inside the card body, separate from text, header and action footer.

### Acceptance criteria

- **WSH-006-AC1:** Only parents may add, replace or remove pictures through the existing wish creation and edit forms. The input accepts a complete HTTP or HTTPS image URL; file upload is not offered. Picture changes are permitted for parents on household wishes regardless of assignment or list name.
- **WSH-006-AC2:** Adding or changing the image URL downloads the source once and stores an optimized local copy. Displaying the wish or editing it with an unchanged image URL does not contact the source again. Subsequent views use the local copy.
- **WSH-006-AC3:** Downloads access public internet addresses only; private/LAN destinations are disallowed. This restriction applies to the initial URL and every redirect target.
- **WSH-006-AC4:** Downloads allow at most five redirects, a total download timeout of 30 seconds and at most 10 MiB (10,485,760 bytes) of source data. Reject source images whose decoded dimensions exceed 20 megapixels (20,000,000 pixels) before resizing.
- **WSH-006-AC5:** Accept JPEG, PNG, WebP and GIF source images. SVG is unsupported. Animated sources, including GIF and WebP, become a still image using only their first frame. Wish pictures never animate.
- **WSH-006-AC6:** Resize proportionally so neither dimension exceeds 640 pixels, without upscaling. Store lossy WebP at 80% quality, preserve transparency against the dark card background, and discard the downloaded original. Display cropping does not alter the stored optimized image.
- **WSH-006-AC7:** Store the source URL and optimized filename with the wish in YAML. Keep optimized files separately from the single household YAML file, at household scope. Retire year/event/person image directories; the approved path/schema is specified in storage.md.
- **WSH-006-AC8:** Pictures use the selected image shelf below title/note and above price/claim/link details. Use a fixed centered crop to fill the shallow shelf, without a crop-position control. Pictures never span the header or action footer, overlap/dim text, or introduce horizontal overflow; text remains clearly readable on desktop/mobile.
- **WSH-006-AC9:** With no picture configured, retain the current card layout without an image shelf or placeholder.
- **WSH-006-AC10:** Clicking an available shelf picture opens the complete, uncropped optimized image in a modal. Clicking the background outside the picture modal closes it; clicking its image or interior keeps it open. Opening and closing the picture modal use an ease-in-out animation.
- **WSH-006-AC11:** Admins may view pictures of all household wishes, including unassigned wishes. A user link may view only pictures of wishes assigned to its list. Direct image URLs enforce the same scope; previously known URLs do not bypass reassignment, deletion or code regeneration.
- **WSH-006-AC12:** Download/processing failures stop the whole save with an error. Preserve all entered form values so parents can enter an alternate URL; the saved wish and picture remain unchanged. This includes unsupported/invalid sources and exceeded network/size/resolution limits.
- **WSH-006-AC13:** Clearing the image URL and successfully saving removes the picture. Leaving the URL unchanged retains the saved local copy without re-downloading it.
- **WSH-006-AC14:** Successful image replacement/removal or wish deletion immediately deletes the old local image. If the old file cannot be deleted, the whole action fails and preserves the previous wish and picture. An already-missing old file counts as already deleted and does not block replacement/removal.
- **WSH-006-AC15:** A configured but missing/unreadable local image shows a small “Image unavailable” placeholder; the rest of the wish remains usable. Do not automatically retrieve the source again. This placeholder is distinct from a wish with no image configured (AC9).

## Relationships and constraints

- Access: shared-password admin sessions and individual user/list codes replace parent/family links and browser names.
- Wishlists: household recipients, named lists, assignments and scoped summaries. There is no annual/history/archive subsystem.
- Claims: household user IDs identify owners; admins select the user per claim. Unassigned claims remain parent-manageable.
- Keep CON-001–CON-003 stack/appearance and existing image/network limits in CON-006. Persistence changes to one household YAML data file using stable IDs and separate pictures.
- Invalid manual data and rejected/stale submissions retain existing error-without-writes protection. Edits preserve unrelated wishes/claims and data outside the edited fields.

## OPEN review items

New English/German interface/error wording is approved in translations.md. Household picture paths/schema and the implementation plan are approved in storage.md and implementation.md. No code starts before milestone 0 is finalized.

## Implementation and verification

Pending implementation/integration. Preserve meaningful existing detail, sorting, status-color, image-download/processing/failure and modal tests. Add failing-first tests for household-owned CRUD, assigned/unassigned contexts, no year restrictions, shared identity across views, scoped picture requests after moves/regeneration, preserved claims, unchanged files after errors and correct targets after sorting/filtering. Verify both languages and mobile/desktop before declaring the revised domain Implemented.
