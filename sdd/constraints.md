# Constraints

## CON-* entries

### CON-001: Hosting

- Deploy as a Docker container, self-hosted on a VPS.

### CON-002: Backend

- Use Go for the backend.

### CON-003: Frontend

- Use Tailwind CSS and vanilla JavaScript.
- The frontend has a Christmas theme.
- Use the user-approved December-night appearance in `design/december-night/approved-baseline.png`: midnight navy surfaces, soft-white/ivory text, ice-blue borders, mint primary actions, and colorful Christmas-bulb accents with the approved brightness.

#### CON-003 acceptance criteria: December-night theme

Status: approved by the user and Implemented for all existing screens. All five criteria are verified.

- **CON-003-AC1:** All existing screens use the approved dark navy surfaces, ivory/soft-white text, ice-blue borders, and mint primary actions. No light-background card or input remains.
- **CON-003-AC2:** The header shows a restrained strand of coral, gold, mint, and ice-blue bulbs with steady glow and no flashing.
- **CON-003-AC3:** Decorative snow is limited to 16 particles, 1–3 pixels each, at no more than 35% opacity, with downward journeys lasting 20–40 seconds. Snow stays behind content, does not cross the form or text, and never intercepts clicks or keyboard focus.
- **CON-003-AC4:** A keyboard-accessible “Pause snow” / “Resume snow” control is available. The selection is remembered in the browser. With reduced motion enabled in the browser, snow does not animate.
- **CON-003-AC5:** Desktop and mobile previews retain readable labels and controls, with no horizontal overflow; the existing setup, copy-link, regeneration, and stale-page tests continue to pass.

#### Verification

- CON-003-AC1/AC2/AC3/AC5: `tests/browser/theme.spec.js` verifies dark rendered surfaces, label contrast, steady bulbs in four colors, bounded particles outside content, and desktop/mobile fit. Manual screenshot inspection confirms the approved appearance.
- CON-003-AC4: browser tests verify actual motion, keyboard pause/resume, browser persistence, and reduced-motion behavior both on first load and when the preference changes.
- CON-003-AC1/AC5: maximum-length household names, validation errors, parent/family screens, and not-found pages are covered. `TestInvalidYAMLUsesTheSharedErrorTheme` checks the shared error shell without modifying invalid data.
- Existing setup, real clipboard copying, independent regeneration, and stale-page revocation browser tests pass under the theme.
- Implementation and tests carry matching `@impl`/`@test` anchors. The approved theme is rendered in HTML/CSS; the reference image is not a production background asset.

### CON-004: Persistence

- Persist data in YAML files.
- Organize data hierarchically by household, then year, then event type, then person name. A person may be an adult or a child.
- Retain the event-type level for future expansion; Christmas is the only supported event in v1.
- Support managing household members and wishlists both through the app and by editing YAML files on the VPS.
- Pick up direct YAML file edits on page load.
- If manually edited YAML is invalid or contains duplicate person names, show an error and leave the files untouched.
- Block editing previous years' wishlists through the web UI. Direct YAML edits to those lists cannot be prevented and are an accepted risk.

## OPEN questions

