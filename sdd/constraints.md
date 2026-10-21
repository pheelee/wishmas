# Constraints

## CON-* entries

### CON-001: Hosting

- Deploy as a Docker container, self-hosted on a VPS.

The user approved milestone 6 and its Docker/persistence/complete-flow verification plan. These observable checks map that approved scope; they add no product behavior:

- **CON-001-AC1:** Building the Docker image and running it serves the existing setup/app and embedded CSS/JavaScript.
- **CON-001-AC2:** With an explicitly mounted data volume, completed setup, access links, people, wishes and claims survive container restart and replacement using the same volume and YAML hierarchy (CON-004).
- **CON-001-AC3:** The container supports the existing parent/family setup, person/wish management and claim flows with the existing permission/privacy rules.
- **CON-001-AC4:** Repository instructions describe building/running the image on the self-hosted VPS and preserving the data volume. Local container verification is distinguished from actual VPS deployment; no VPS target has been provided.

Local verification: CON-001-AC1–AC4 pass. The container smoke test failed first because the Dockerfile was absent, then verified setup, embedded assets, parent/family claims and permissions, IANA timezone selection, byte-preserved YAML and surviving links across container restart/replacement on a named volume. Docker's remote development endpoint and reassigned ephemeral ports are handled by the test. The runtime uses UID/GID 65532 and embeds Go assets/timezone data; the build context excludes household data and notes. All 20 existing Chromium scenarios, Go race tests and vet pass. VPS execution has not been performed; CON-001 is locally verified, not marked deployed to a VPS.

### CON-002: Backend

- Use Go for the backend.

### CON-003: Frontend

- Use Tailwind CSS and vanilla JavaScript.
- The frontend has a Christmas theme.
- Use the user-approved December-night appearance in `design/december-night/approved-baseline.png`: midnight navy surfaces, soft-white/ivory text, ice-blue borders, mint primary actions, and colorful Christmas-bulb accents with the approved brightness.

#### CON-003 acceptance criteria: December-night theme

Status: user-approved and Implemented. AC1–AC13, including the approved top-snow addition, have been verified.

- **CON-003-AC1:** All existing screens use the approved dark navy surfaces, ivory/soft-white text, ice-blue borders, and mint primary actions. No light-background card or input remains.
- **CON-003-AC2:** The header shows a restrained strand of coral, gold, mint, and ice-blue bulbs with steady glow and no flashing.
- **CON-003-AC3:** Decorative snow has setup range sliders for particle size and amount. Snow size ranges from 50–200% of its previous size (default 100%); snow amount ranges from 8–48 particles (default 16) in each region: the existing combined side margins and the header band. Existing limits on opacity (at most 35%) remain. The side-margin snow retains its 20–40-second journey duration; the approved shallow header snowfall follows AC11. Snow stays behind content, does not cross the form or text, and never intercepts clicks or keyboard focus.
- **CON-003-AC4:** Snow is always present and there is no Pause snow / Resume snow control. With reduced motion enabled in the browser, snow remains present without animation.
- **CON-003-AC5:** Desktop and mobile previews retain readable labels and controls, with no horizontal overflow; the existing setup, copy-link, regeneration, and stale-page tests continue to pass.

#### UI-polish additions

Approved by the user and Implemented. These extend the existing December-night theme.

- **CON-003-AC6:** Default light bulbs are 20% larger than the previous baseline. A setup range slider changes bulb size live for visual review. Increasing bulb size reduces the number of bulbs to maintain the ratio. Bulb size ranges from 80–200% of the previous baseline, defaulting to 120%. Bulb count is inversely proportional to size: doubling size halves the count.
- **CON-003-AC7:** Setup includes separate range sliders for snow size and snow amount. Snow size ranges from 50–200% of the previous baseline, defaulting to 100%. Snow amount ranges from 8–48 particles, defaulting to 16.
- **CON-003-AC8:** The three appearance settings are saved for the whole household in YAML; everyone sees the chosen appearance.
- **CON-003-AC9:** Screens larger than the `md` breakpoint use 75% of the screen width.
- **CON-003-AC10:** Editing and action forms use modals; creation forms remain inline. This includes wish editing, person rename/archive, access-link regeneration, and wish-deletion confirmation. Copying links, sorting, and year navigation remain direct controls. Claim toggles/removals follow CLM-002/CLM-004 without confirmation; changing the current claim name uses a modal.

#### Approved top-snow addition

The user approved the visual preview and requested the spec update followed by implementation.

- **CON-003-AC11:** In addition to the existing side-margin snow, gentle snow falls across the lightbulb band. It remains clipped to that shallow band and fades out toward its bottom, before page text/forms. The existing bulb positions, steady glow and side snowfall remain in place on desktop/mobile.
- **CON-003-AC12:** Header snow follows the existing household snow-size and snow-amount settings, including immediate setup-slider previews. The amount setting determines the header particle count separately from the existing combined side count; no extra setting or YAML field is added.
- **CON-003-AC13:** Header snowfall is decorative, hidden from assistive technology, noninteractive and excluded from keyboard focus. Reduced motion keeps its particles visible and static. Normal motion resumes when the browser preference is restored.

#### Selected card appearance revision

The user selected Gift-wrap proposal 2, recorded in `design/gift-wrap/selected-baseline.png`. The existing December-night page theme remains in force. WSH-005 owns the approved and Implemented card acceptance criteria, verified with desktop/mobile screenshots and browser regression tests.

#### Verification

- Go failing-first tests cover appearance setup validation, YAML persistence/restart, legacy defaults, invalid manual settings without writes, and setting preservation during regeneration.
- `tests/browser/polish.spec.js` verifies browser timezone prefill, preserved manual choices/drafts after validation, all slider previews, 20% default bulb enlargement, household-wide appearance, inline clipboard actions/three-second feedback, modal cancellation, native Escape/focus return, 75% desktop width, individual cards, and bottom icon actions.
- Updated existing browser scenarios verify real clipboard contents, independent link replacement and stale-page revocation through modals, wish edit validation/claims preservation, modal deletion/cancellation, person rename/archive, sorting with attached form targets, and family/history restrictions.
- Theme scenarios verify steady bulb colors, side-margin snow and the clipped/fading header band, actual animation without pause controls, ignored obsolete pause preferences, visible static reduced-motion snow, dark surfaces, label contrast, and mobile/desktop fit. Screenshots were reviewed.
- Implementation and tests carry matching criterion anchors. No animation library or dependency was added.
- CON-003-AC11–AC13: added the approved clipped header snowfall with a bottom fade. Header particles use the existing household snow size/amount on page load and setup input, without new storage or controls. They inherit the light strand's inert/assistive-hidden behavior and share static reduced-motion styling with side snow. The browser regression first failed against the pre-preview frontend with zero header particles. It now verifies real motion, clipping/fade, desktop/mobile layout, slider changes, parent/family persisted appearance and reduced-motion behavior across reload/preference restoration. All 25 Chromium scenarios, Go race tests, vet, CSS build, JavaScript syntax and diff checks pass. Final desktop/mobile screenshots were visually reviewed. Temporary preview code names, scripts, screenshots and backups were removed; no live app restart, data edit or dependency change was made.


### CON-004: Persistence

- Persist data in YAML files.
- Organize data hierarchically by household, then year, then event type, then person name. A person may be an adult or a child.
- Retain the event-type level for future expansion; Christmas is the only supported event in v1.
- Support managing household members and wishlists both through the app and by editing YAML files on the VPS.
- Pick up direct YAML file edits on page load.
- If manually edited YAML is invalid or contains duplicate person names, show an error and leave the files untouched.
- Block editing previous years' wishlists through the web UI. Direct YAML edits to those lists cannot be prevented and are an accepted risk.

### CON-005: Internationalization delivery and configuration

Confirmed in the requirements interview and Implemented. INT-001–INT-005 criteria are verified.

- Support `de-CH` and `en-US` only. Translations are bundled with the app; neither language requires an external translation service or runtime translation download.
- Persist the household language in `household.yaml` as `language: de-CH` or `language: en-US`, using exact values. Only a missing field gets the legacy English fallback; present empty/null/invalid values produce an English configuration error without rewriting YAML.
- Changing language affects browser presentation only; existing YAML data, field names, folder paths, access codes and claims remain unchanged.
- Price entry/storage remain ungrouped decimal values; regional formatting with two decimal places affects display only. No currency is added.
- The [Internationalization domain](internationalization.md) owns detailed behavior and acceptance criteria. INT-001–INT-005 acceptance criteria and the [translation table](translations.md) are approved and implemented. Catalog/formatting/YAML tests, all 22 browser scenarios and both-language DinD smoke checks pass.

### CON-006: Wish picture downloads and storage

Implemented. Confirmed picture-interview decisions and WSH-006-AC1–AC15 are approved and verified by Go/browser tests and both-language DinD smoke checks.

- Input is a complete HTTP/HTTPS image URL only; no file upload. Download only from public internet addresses, including redirect destinations.
- At most five redirects within a 30-second total download timeout; at most 10 MiB source data and 20 megapixels decoded source resolution.
- Accept JPEG/PNG/WebP/GIF; no SVG. Use the first frame of animated sources, including GIF/WebP.
- Resize proportionally to at most 640 pixels on either side without upscaling, encode lossy WebP at 80% quality, preserve alpha/transparency and discard originals.
- Retain source URL and optimized filename in wish YAML; keep optimized files in `pictures/` beside the person's `wishlist.yaml`. Browsing uses the local copy rather than the remote source.
- Shelf cropping is display-only and fixed at the center; a modal shows the full optimized image. Access follows existing role/year/archive permissions, including direct image URLs.
- The user reviewed the completed image feature and modal refinements and authorized committing them.

## OPEN questions

None currently listed.


Bulb attachment correction (CON-003-AC2/AC5/AC6): cord endpoints now match the normalized horizontal coordinates used for bulb drops. Bulb stems are centered and rotations pivot at the cord attachment, preserving their connection at all supported sizes. A failing-first browser geometry test measures stem-tip distance from the actual SVG path at 80%, 120% and 200% on desktop/mobile; all distances are below 0.75px after the fix. Setup/theme scenarios and Go tests pass; corrected live-data preview reviewed using browser-only overrides.
