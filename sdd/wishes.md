# Wishes

## Status

The existing wish behavior and **WSH-008-AC1–AC12/WSH-006 amendments are implemented and verified**, including startup recovery in both storage modes and durable S3 cleanup retries.

Household refactor: **behavioral acceptance criteria approved by the user**. **Implemented and verified across milestones 4–5**: household wish CRUD, assignments, cards, sorting, pictures and the claim-action footer. The user confirmed retention of existing wish fields, validation, pictures, sorting and card appearance. Retained criteria below remain product requirements, but baseline verification is not evidence of integration with the revised model.

WSH-002's current-year/history management rules are retired and replaced by WSH-007. In WSH-004–WSH-006, obsolete history/archive wording and per-person picture storage are removed. WLS-010 owns assignment behavior; WLS-011 owns recipient summaries.

## Priority

Must-have.

## WSH-001: Wish details

Status: Implemented and verified in the household context (milestone 4).

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

Status: Implemented and verified in the household context (milestone 4).

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

Status: Implemented and verified in the household context (milestone 4).

- **WSH-004-AC1:** Each wish is displayed in its own card for better visibility.
- **WSH-004-AC2:** Edit wish and Delete wish are icon actions placed inline at the bottom of the card.
- **WSH-004-AC3:** Edit wish opens a modal rather than an inline editing panel.
- **WSH-004-AC4:** Wish-deletion confirmation appears in a modal. Cancelling leaves the wish and claims unchanged; confirming retains permanent deletion behavior.
- **WSH-004-AC5:** The wish creation form remains inline.

Parent-only management, claim preservation, deletion confirmation and sorting remain in force. Remove historical read-only restrictions.

## WSH-005: Gift-wrap card styling

Status: Implemented and verified across milestones 4–5, including parent-only claimant removals and the scoped claim-action footer.

The selected card reference is [Gift wrap](../design/gift-wrap/selected-baseline.png). This changes presentation of existing data and actions; the surrounding proposal-page chrome is not part of this change.

- **WSH-005-AC1:** Cards use the selected edge accents, subtle matching header gradients, gift icons, and serif wish titles. Accent colors follow the status rules in AC6–AC7.
- **WSH-005-AC2:** Claimed/unclaimed and claimed-by-you indicators appear as pill badges. Claim counts remain visible and claimant names/removal actions remain parent-only as specified in Claims.
- **WSH-005-AC3:** An optional numeric reference price receives prominent colored emphasis. Optional shop links use subtle domain labels with link icons; omitted optional fields show no invented values or empty detail placeholders.
- **WSH-005-AC4:** A separated footer contains the existing claim toggle on the left and parent edit/delete icons inline on the right. Existing modal and confirmation behavior remains unchanged.
- **WSH-005-AC5:** Cards retain the responsive desktop grid and stack on mobile without overflow, while preserving new admin/user permissions, sorting and action targets.
- **WSH-005-AC6:** A wish is complete when it has a title, a provided price and at least one shop link. Its status accent is red if incomplete (including when claimed), yellow if complete with no claims, and green if complete with at least one claim. A note or picture is not required for completeness.
- **WSH-005-AC7:** Status colors replace the original alternating mint/gold/ice-blue accents in AC1. Apply them to the existing border, subtle header tint, gift icon, reference price and claim-button accent styling. Keep the dark card background and existing claimed-by-you treatment.
## WSH-006: Optional wish picture

Status: The existing household-scope behavior and AVIF extension are implemented. The WSH-008 startup-recovery and S3 cleanup amendments are implemented and verified. The selected shelf reference is [proposal 3](../design/wish-picture/selected-baseline.png).

A wish has at most one optional picture, provided by image URL. The selected presentation is a shallow shelf inside the card body, separate from text, header and action footer.

### Acceptance criteria

- **WSH-006-AC1:** Only parents may add, replace or remove pictures through the existing wish creation and edit forms. The input accepts a complete HTTP or HTTPS image URL; file upload is not offered. Picture changes are permitted for parents on household wishes regardless of assignment or list name.
- **WSH-006-AC2:** Adding or changing the image URL downloads the source once and stores an optimized local copy. Displaying the wish or editing it with an unchanged image URL does not itself contact the source again. Subsequent views use the local copy. Missing-file recovery is handled separately by WSH-008 startup/background jobs.
- **WSH-006-AC3:** Downloads access public internet addresses only; private/LAN destinations are disallowed. This restriction applies to the initial URL and every redirect target.
- **WSH-006-AC4:** Downloads allow at most five redirects, a total download timeout of 30 seconds and at most 10 MiB (10,485,760 bytes) of source data. Reject source images whose decoded dimensions exceed 20 megapixels (20,000,000 pixels) before resizing.
- **WSH-006-AC5:** Accept JPEG, PNG, WebP, GIF and AVIF source images. SVG is unsupported. Animated sources, including GIF, WebP and AVIF, become a still image using only their first frame. Wish pictures never animate.
- **WSH-006-AC6:** Resize proportionally so neither dimension exceeds 640 pixels, without upscaling. Store lossy WebP at 80% quality, preserve transparency against the dark card background, and discard the downloaded original. Display cropping does not alter the stored optimized image.
- **WSH-006-AC7:** Store the source URL and optimized filename with the wish in YAML. Keep optimized files separately from the single household YAML file, at household scope. Retire year/event/person image directories; the approved path/schema is specified in storage.md.
- **WSH-006-AC8:** Pictures use the selected image shelf below title/note and above price/claim/link details. Use a fixed centered crop to fill the shallow shelf, without a crop-position control. Pictures never span the header or action footer, overlap/dim text, or introduce horizontal overflow; text remains clearly readable on desktop/mobile.
- **WSH-006-AC9:** With no picture configured, retain the current card layout without an image shelf or placeholder.
- **WSH-006-AC10:** Clicking an available shelf picture opens the complete, uncropped optimized image in a modal. Clicking the background outside the picture modal closes it; clicking its image or interior keeps it open. Opening and closing the picture modal use an ease-in-out animation.
- **WSH-006-AC11:** Admins may view pictures of all household wishes, including unassigned wishes. A user link may view only pictures of wishes assigned to its list. Direct image URLs enforce the same scope; previously known URLs do not bypass reassignment, deletion or code regeneration.
- **WSH-006-AC12:** Download/processing failures stop the whole save with an error. Preserve all entered form values so parents can enter an alternate URL; the saved wish and picture remain unchanged. This includes unsupported/invalid sources and exceeded network/size/resolution limits.
- **WSH-006-AC13:** Clearing the image URL and successfully saving removes the picture. Leaving the URL unchanged retains the saved local copy without re-downloading it.
- **WSH-006-AC14:** Attempt immediate old-image removal after successful replacement/removal or wish deletion. Local mode retains the existing whole-action rollback if the old file cannot be deleted. In S3 mode, once the remote YAML save succeeds, an old-image cleanup failure does not undo it; retry local cleanup under WSH-008. An already-missing old file counts as already deleted and does not block replacement/removal.
- **WSH-006-AC15:** A configured but missing/unreadable local image shows a small “Image unavailable” placeholder; the rest of the wish remains usable. Page views do not themselves trigger source retrieval. WSH-008 startup recovery and its retries may restore a missing file in the background; until restored, retain the placeholder. This placeholder is distinct from a wish with no image configured (AC9).

## WSH-008: Background image recovery and S3 cache cleanup

**Implemented and verified: WSH-008-AC1–AC12 and the WSH-006 amendments.**

- **WSH-008-AC1 — Both storage modes:** At app startup, schedule recovery of missing configured wish-image files in both local and S3 modes. Images remain on local disk; do not upload them to S3.
- **WSH-008-AC2 — Background availability:** Run recovery in the background after the app starts, without waiting for all images before serving pages. A failed restore does not stop startup or make the rest of the wish unusable. Use the existing Image unavailable placeholder while a configured image is missing.
- **WSH-008-AC3 — Same processing contract:** Fetch the stored image source URL through the existing public-address/redirect protections and all WSH-006 network/byte/dimension/format/first-frame/resize/WebP/alpha limits. The S3 five-second operation timeout does not replace the existing thirty-second image-download limit.
- **WSH-008-AC4 — Concurrency:** Run at most three image restores simultaneously. Do not run simultaneous duplicate restores for the same current wish-image reference.
- **WSH-008-AC5 — Automatic retry:** After failed restore attempts, use increasing delays of 1, 2, 4, 8, 16 and then 30 minutes, capping further delays at 30 minutes. Retry until the image is restored or that recovery is no longer needed.
- **WSH-008-AC6 — Cache-only recovery:** Recreate the optimized local file under its existing stored filename. Recovery does not change household YAML, source URLs, filenames, wish details, assignments or claims. It does not queue an ordinary data edit.
- **WSH-008-AC7 — Current references:** Before publishing a recovered file, retain the current wish/image-reference boundary: a removed wish or changed image reference makes the old recovery unnecessary. Recovery of a missing file must not overwrite a local copy that has already appeared. Existing access controls continue to determine who may view the image.
- **WSH-008-AC8 — Page-load visibility:** A recovered image appears on the next page load. Do not add browser polling or push updates; pages with an unavailable placeholder remain usable meanwhile. Wishes without an image configured retain their existing no-image layout.
- **WSH-008-AC9 — Successful S3 save:** If S3 confirms a YAML save but deleting its old local image fails, keep the save successful. Leave the unused local file for later cleanup rather than rolling back remote data. An old file still on disk does not restore removed wish data or bypass current picture-access scope.
- **WSH-008-AC10 — Restart-safe cleanup retries:** Retry failed S3-mode old-image cleanup using the same 1/2/4/8/16/30-minute increasing delays, capped at 30 minutes. Continue cleanup attempts after app restarts using retained local state. Delete only unused cached images, never a file still referenced by current household data.
- **WSH-008-AC11 — Local rollback retained:** Local-mode picture changes retain the existing save-failure/old-file-removal rollback rules. The S3 post-save cleanup exception does not change those local-mode rules. Ordinary image-source download/processing/save failures retain the existing applicable draft/data/picture preservation.
- **WSH-008-AC12 — Household state untouched:** Background image recovery and cleanup operate on local image files, not household YAML mutations. Preserve current role/list scopes, unrelated files/data and claimant privacy. The only deferred household-state write in this feature is the separately approved pending-session revocation under ACC-008.

## Relationships and constraints

- Access: shared-password admin sessions and individual user/list codes replace parent/family links and browser names.
- Wishlists: household recipients, named lists, assignments and scoped summaries. There is no annual/history/archive subsystem.
- Claims: household user IDs identify owners; admins select the user per claim. Unassigned claims remain parent-manageable.
- Keep CON-001–CON-003 stack/appearance and existing image/network limits in CON-006. Persistence changes to one household YAML data file using stable IDs and separate pictures.
- Invalid manual data and rejected/stale submissions retain existing error-without-writes protection. Edits preserve unrelated wishes/claims and data outside the edited fields.

## OPEN review items

- The user approved AVIF support under the existing image limits/conversion/first-frame rules, and the two revised I03/I15 English/German messages. No product or wording decision remains OPEN; AVIF verification is complete.

New English/German interface/error wording is approved in translations.md. Household picture paths/schema and the implementation plan are approved in storage.md and implementation.md. No code starts before milestone 0 is finalized.

## Implementation and verification

Milestone 4 implements WSH-001, WSH-003/004/006/007 in the household model, and WSH-005 presentation with claim removal actions and the claim-toggle footer still pending milestone 5. `household_wishes.go`, `household-wishes.html` and the version-2 repository picture transaction supply stable-ID CRUD, recipient-context creation, preserved claims/future fields, assignment/filtering, local pictures and current-scope image serving. Existing pure validation/price, sorting JavaScript, image/network processor and selected visual styles are reused; retained legacy tests alone are not the new-context evidence.

New failing-first HTTP checks began with the missing household creation form; repository rollback checks began with the absent picture API. Focused Go tests verify title/decimal/URL validation, assigned/unassigned creation, moves/claim preservation, stable action targets, stale rejection, exact large totals, public-only image processing, single-fetch/unchanged-source behavior, replacement/removal/deletion/missing pictures, write failures, rollback and in-flight credential/revision changes. Direct picture requests are denied after moves, unassignment and regeneration, even with an unrelated admin cookie.

English/German HTTPS browser flows verify optional-field edits and retained drafts, domain labels, red/yellow/green accents, claim privacy, scoped totals, native modals, sorting/filtering, picture dismissal, recipient creation, cancellation/permanent deletion, real restart and mobile/desktop presentation. Shared image processor checks continue to verify all formats, first-frame/alpha/quality/size/network limits; Go-only image checks cover the deployment backend. Milestone 5 subsequently completes claim-action integration below; [review walkthrough](../docs/admin-access.md#milestone-4-review).

### Milestone 5 claim-footer completion

WSH-005-AC2/AC4 are complete: users have a direct own-claim toggle; admins choose a predefined user for each new assigned-wish claim and remove any existing claim directly. The footer retains the selected accents and right-aligned assignment/edit/delete icons. Empty required selections, stale/duplicate/write errors and unassigned creation leave claims unchanged. Go tests and English/German browser flows verify current names/privacy, 1440/840/390 geometry, picker validation, actual claim preservation across edits/moves and deletion cleanup. All Wishes criteria are now verified in the new context; container/current-data review remains milestone 6.

### Approved AVIF source extension

WSH-006-AC5 and CON-006 now include AVIF. The user separately approved I03/I15 English/German wording before implementation. The shared downloader advertises image/avif and the processor registers pinned gen2brain/avif v0.6.0; input still becomes quality-80, alpha-preserving, maximum-640px still WebP under the existing limits. No storage schema or image access policy changed.

The first synthetic still/animated AVIF tests failed with I15 before implementation. Local original fixtures now verify transparency/proportional resizing, first-frame selection without upscaling, malformed input and metadata-only 20-megapixel rejection before pixel decode. Household HTTP save checks verify AVIF negotiation, no unchanged-source fetch, invalid replacement preserving stored YAML/image bytes and immediate removal. The supplied public Galaxus example passes through the actual public HTTPS client and becomes 32,954-byte 640×508 WebP from 78,603-byte AVIF; no shop image is committed.

Verification: full Go race tests, vet, no-CGo/nodynamic build and focused no-CGo AVIF tests pass. All six targeted English/German picture/household-wish/i18n browser regressions pass. Docker builds the production executable and runs AVIF checks plus the supplied sample in scratch as UID/GID 65532, without shared codecs or network access during decoding. Disposable container/image and temporary sample/build recipe were removed. This verifies codec compatibility only; the separate household-refactor milestone 6 migration/container-session/Traefik review remains pending. Existing live app/data/main and user ignore edits are untouched.

### Background image recovery and S3 cleanup delivery

`recovery.go` starts the background service from production/main and the HTTPS harness; the shared protected downloader/optimizer is reused. Failing-first tests verify pages remain usable while three workers are blocked, no fourth restore starts, existing references are rechecked, original filenames/YAML stay unchanged, images appear on the next page load and actual scheduler retries use increasing deadlines until the reference is removed. These checks run in both storage modes. Repository checks verify no overwrite of an existing file, stale-reference rejection, successful remote-save preservation, cleanup state across restart and retained local rollback.

Production scratch/Traefik flows in both modes deliberately remove an optimized image, restart the app and verify automatic retrieval of the public source, byte-identical optimized output/YAML and retained authentication/claims. Failed-source placeholders and existing format/SSRF/limit tests remain in force. No original household data or running app was used or changed.
