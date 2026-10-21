# Constraints

## Revision status

Household refactor: technology/storage/authentication proposal approved by the user. New-contract acceptance checks below are pending implementation and verification; previous implementation reports do not prove them. The resumable plan and new wording are separately approved; code is paused before milestone 1.

### CON-001: Hosting and deployment

Retain a Docker container self-hosted on a VPS. Production is reachable only through Traefik over HTTPS; redirect HTTP to HTTPS and do not expose the app port publicly.

- **CON-001-AC1:** Build/run serves the new setup, admin and individual user/list flows with bundled frontend assets.
- **CON-001-AC2:** The mounted data volume preserves the single household file, picture files, password hash, sessions, IDs, codes, lists, rosters, assignments and claims across restart/container replacement.
- **CON-001-AC3:** Local complete-flow and Traefik/HTTPS verification includes session cookies, forwarded client IPs, independent browsers, code regeneration and privacy boundaries.
- **CON-001-AC4:** Document build/run, volume ownership, Traefik trust/TLS configuration, review environment, cutover and rollback. Local verification is distinguished from actual VPS deployment. Merge/cutover requires user review approval.

### CON-002: Backend

Keep Go. Use the Go Argon2id package for password hashing; select a compatible pinned module version during implementation and verify against the repository's Go toolchain. Keep the existing image-processing dependencies/limits.

### CON-003: Frontend

- Use Tailwind CSS and vanilla JavaScript.
- The frontend has a Christmas theme.
- Use the user-approved December-night appearance in `design/december-night/approved-baseline.png`: midnight navy surfaces, soft-white/ivory text, ice-blue borders, mint primary actions, and colorful Christmas-bulb accents with the approved brightness.

#### CON-003 acceptance criteria: December-night theme

Retain user-approved appearance AC1–AC13. Previously verified at baseline 1992bdf; revised screens/authentication require new integration verification.

- **CON-003-AC1:** All existing screens use the approved dark navy surfaces, ivory/soft-white text, ice-blue borders, and mint primary actions. No light-background card or input remains.
- **CON-003-AC2:** The header shows a restrained strand of coral, gold, mint, and ice-blue bulbs with steady glow and no flashing.
- **CON-003-AC3:** Decorative snow has setup range sliders for particle size and amount. Snow size ranges from 50–200% of its previous size (default 100%); snow amount ranges from 8–48 particles (default 16) in each region: the existing combined side margins and the header band. Existing limits on opacity (at most 35%) remain. The side-margin snow retains its 20–40-second journey duration; the approved shallow header snowfall follows AC11. Snow stays behind content, does not cross the form or text, and never intercepts clicks or keyboard focus.
- **CON-003-AC4:** Snow is always present and there is no Pause snow / Resume snow control. With reduced motion enabled in the browser, snow remains present without animation.
- **CON-003-AC5:** Desktop and mobile previews retain readable labels and controls, with no horizontal overflow; the revised setup, user/list link-copy/regeneration and stale-page flows remain usable.

#### UI-polish additions

Approved appearance rules retained in the refactor.

- **CON-003-AC6:** Default light bulbs are 20% larger than the previous baseline. A setup range slider changes bulb size live for visual review. Increasing bulb size reduces the number of bulbs to maintain the ratio. Bulb size ranges from 80–200% of the previous baseline, defaulting to 120%. Bulb count is inversely proportional to size: doubling size halves the count.
- **CON-003-AC7:** Setup includes separate range sliders for snow size and snow amount. Snow size ranges from 50–200% of the previous baseline, defaulting to 100%. Snow amount ranges from 8–48 particles, defaulting to 16.
- **CON-003-AC8:** The three appearance settings are saved for the whole household in YAML; everyone sees the chosen appearance.
- **CON-003-AC9:** Screens larger than the `md` breakpoint use 75% of the screen width.
- **CON-003-AC10:** Editing and action forms use modals; creation forms remain inline. This includes wish/list/recipient/user editing, user/list link regeneration and destructive confirmations. Copying links and sorting remain direct controls. List-user claim toggles and admin claim removals remain direct without confirmation. Admin claim creation requires a user selection for that claim; there is no browser-name editing flow.

#### Approved top-snow addition

The user approved the visual preview and requested the spec update followed by implementation.

- **CON-003-AC11:** In addition to the existing side-margin snow, gentle snow falls across the lightbulb band. It remains clipped to that shallow band and fades out toward its bottom, before page text/forms. The existing bulb positions, steady glow and side snowfall remain in place on desktop/mobile.
- **CON-003-AC12:** Header snow follows the existing household snow-size and snow-amount settings, including immediate setup-slider previews. The amount setting determines the header particle count separately from the existing combined side count; no extra setting or YAML field is added.
- **CON-003-AC13:** Header snowfall is decorative, hidden from assistive technology, noninteractive and excluded from keyboard focus. Reduced motion keeps its particles visible and static. Normal motion resumes when the browser preference is restored.

#### Selected card appearance revision

The user selected Gift-wrap proposal 2, recorded in `design/gift-wrap/selected-baseline.png`. The existing December-night page theme remains in force. WSH-005 owns the retained approved card acceptance criteria, with new-context verification required.

### CON-004: Household persistence

The old household → year → event → person hierarchy is superseded. See the approved [storage contract](storage.md) for exact schema and examples.

- **CON-004-AC1:** Store all persistent household state in `data/households/<household-name>/household.yaml`, marked `schema_version: 2`: name/language/appearance, password hash, hashed sessions, users, recipients, lists/user links, wishes/assignments/claims. Pictures remain separate at household scope.
- **CON-004-AC2:** Users, recipients, lists and wishes have immutable generated IDs; names are editable fields. References use IDs. Arrays preserve added order; assignment/reassignment does not reset household wish added order.
- **CON-004-AC3:** Reload manual YAML edits on page load and before mutation. Validate IDs, references, selected recipients, names, codes, claims and retained wish/image rules. Invalid manual data shows an error and is left untouched.
- **CON-004-AC4:** Serialize app writes, reject stale edits and atomically replace the household YAML file. Preserve unrelated data/claims; failures must not partially apply related list/user/recipient/assignment changes.
- **CON-004-AC5:** Retain image replacement/removal/deletion failure rollback and immediate old-file removal. Images are named using wish identity plus image version under household `pictures/`; edits do not re-download unchanged sources.
- **CON-004-AC6:** Do not silently migrate legacy files at startup. Migration is explicit and copy-only until approved cutover. Remove automatic-year, timezone, read-only-history and archive/restore persistence/behavior from the new runtime.

### CON-005: Internationalization delivery and configuration

Retain bundled de-CH/en-US only, informal Swiss German, no external translation download and no currency. Language remains `language` in the single household file; missing is English, present invalid/empty/null configuration is an English error without writes. Everyone, including the admin login page, uses valid household language after setup. Changing language changes presentation only, not IDs/references/content/codes/claims or paths. Price entry/storage remains ungrouped decimal; individual wish display follows locale grouping/two decimals, while overview totals remain ungrouped/two decimals. See revised INT-001, INT-005–INT-007 and WLS-011. New wording requires separate table approval; remove unused legacy messages.

### CON-006: Wish picture processing and access

Retain approved WSH-006 download/processing/presentation behavior:

- Complete HTTP/HTTPS image URLs only, no uploads; public internet addresses including redirects; maximum five redirects/30 seconds/10 MiB/20 megapixels.
- JPEG/PNG/WebP/GIF, no SVG. Animated sources use their first frame only. Proportional maximum 640px without upscaling; lossy WebP quality 80, preserve alpha, discard originals.
- Source URL and optimized filename remain in wish data; optimized pictures move to household `pictures/` alongside the one household data file. No person/year/event image directories in new storage.
- Keep centered shelf crop, full-image animated modal/backdrop dismissal and reduced-motion behavior. Text stays readable; browsing never fetches source images.
- Admins can view all household pictures; user links can view only pictures of wishes currently assigned to their list. Enforce this on direct image requests, including after moves/code regeneration. There is no year/archive scope.

### CON-007: Passwords, sessions and proxy contract

Approved technical details are in [storage.md](storage.md).

- **CON-007-AC1:** Use salted Argon2id encoded with algorithm/version/cost/salt/hash. Initial cost is 19 MiB/two iterations/one lane, 16-byte random salt and 32-byte result. Reject malformed/unsupported or excessive encoded parameters without unbounded allocation.
- **CON-007-AC2:** A private terminal helper validates the same approved password rules, requests matching confirmation and prints a YAML-ready hash. Passwords are not command arguments or stored plaintext in configuration/review logs.
- **CON-007-AC3:** Codes/session tokens use 32 cryptographically random bytes. Store shareable list codes but only SHA-256 hashes of session tokens, bound to the current password-hash fingerprint. Password changes invalidate old sessions; logout removes the current record/cookie only.
- **CON-007-AC4:** Use persistent Secure/HttpOnly/SameSite=Lax cookies and mutation/CSRF protection, with no app inactivity timeout. Session records persist in household YAML. Browser cookie removal/retention limits remain outside the server's control.
- **CON-007-AC5:** Derive rate-limit IP from the forwarded chain on the trusted Traefik side, configured with deployment-specific trusted proxy addresses. Do not blindly trust a client-supplied leftmost address. Five failures in fifteen minutes blocks that IP for fifteen minutes; in-memory counters/blocks may reset at restart.
- **CON-007-AC6:** Verify cookies/login against isolated HTTPS in review/tests; do not weaken production cookie settings to enable local review. Document deployment-specific trusted addresses rather than guessing them.

### CON-008: Explicit migration and reversible review

- **CON-008-AC1:** Read legacy data without modifying the original; write a separate migration destination and refuse an existing/nonempty destination. Require the initial `admin_password_hash` in a copied legacy household config. The helper is available before the user's data migration.
- **CON-008-AC2:** For current data, create Christmas 2026 from the existing year, preserve recipients and each stored wish/assignment/detail/claim, and expose formerly archived recipients normally. General year discovery preserves ascending order; no wish deduplication is performed.
- **CON-008-AC3:** Create household users from claimant names, merging only case/surrounding-space variants, preserving first encountered spelling and distinguishing accented names. Convert claims to stable user IDs and generate individual user/list links. Old parent/shared-family links cease to work in the new app.
- **CON-008-AC4:** Copy optimized images byte-for-byte to unique household filenames, preserving source URLs and missing-image behavior. Do not download/re-encode during migration. Preserve existing language/appearance and discard obsolete timezone/archive/year access configuration.
- **CON-008-AC5:** Invalid/unrepresentable source data stops migration with a report rather than silently dropping, renaming or merging data beyond approved claimant matching. Report source/destination recipients/wishes/claims and image checksums without exposing password hashes/session tokens/link codes.
- **CON-008-AC6:** Review uses a separate data copy and separate app environment. Original data and baseline code remain available. No merge or live-data cutover until user approval; pause after each verified milestone.

## OPEN deployment items

Exact VPS trusted-proxy addresses are supplied at deployment. New wording and the milestone plan are approved. No unresolved product behavior is implied by previous verification notes; new acceptance evidence is pending.
