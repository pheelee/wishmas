# Constraints

## Revision status

Household refactor: approved technology/storage/authentication contract. Milestones 1–5 verify the offline tools and all new HTTP management/claim flows. Milestone 6 removes obsolete runtime/catalog paths and verifies copied current data and real local Docker/Traefik HTTPS persistence. Actual VPS addresses/deployment and merge/cutover remain separate review decisions. See [verification progress](progress.md) and [review/deployment instructions](../docs/review-and-deployment.md).

### CON-001: Hosting and deployment

Retain a Docker container self-hosted on a VPS. Production is reachable only through Traefik over HTTPS; redirect HTTP to HTTPS and do not expose the app port publicly.

- **CON-001-AC1:** Build/run serves the new setup, admin and individual user/list flows with bundled frontend assets.
- **CON-001-AC2:** The mounted data volume preserves the single household file, picture files, password hash, sessions, IDs, codes, lists, rosters, assignments and claims across restart/container replacement.
- **CON-001-AC3:** Local complete-flow and Traefik/HTTPS verification includes session cookies, forwarded client IPs, independent browsers, code regeneration and privacy boundaries.
- **CON-001-AC4:** Document build/run, volume ownership, Traefik trust/TLS configuration, review environment, cutover and rollback. Local verification is distinguished from actual VPS deployment. Merge/cutover requires user review approval.

#### Confirmed configurable container identity revision

The user approved default UID/GID 1000:1000 with Compose runtime user overrides, without rebuilding, and requested a reference Compose file. Implementation and verification are complete.

- **CON-001-AC5:** The Docker image defaults to UID/GID 1000:1000, with its initial `/data` directory owned by those IDs. Operators can independently override UID and GID at container creation through Compose’s `user` setting, without rebuilding the image or adding a privileged entrypoint.
- **CON-001-AC6:** Provide a reference `docker-compose.yml` using `user: "${WISHMAS_UID:-1000}:${WISHMAS_GID:-1000}"` and a persistent `/data` mount. Document that mounted storage must already be writable by the selected IDs; runtime overrides do not change existing file ownership. Retain the private backend/Traefik HTTPS contract.

Verification: the container smoke failed first on the old 65532:65532 image default; reference validation failed first while the Compose file was absent. The real Compose CLI now resolves defaults and independent UID/GID overrides correctly. One built scratch image passes complete Traefik HTTPS flows as default 1000:1000 and runtime-overridden 1234:5678, including actual YAML ownership, picture writes and session/data persistence through replacement. Custom storage ownership is prepared only in disposable test volumes; the app has no ownership-changing entrypoint. See [deployment guidance](../docs/review-and-deployment.md#runtime-identity-and-compose-reference).

### CON-002: Backend

Keep Go. Use the Go Argon2id package for password hashing; select a compatible pinned module version during implementation and verify against the repository's Go toolchain. Keep the existing image-processing dependencies/limits.

#### Confirmed module and Go-version revision

The user requested module `github.com/pheelee/wishmas` and Go 1.27.3 throughout relevant files. The user approved CON-002-AC1/AC2; implementation and verification are complete.

- **CON-002-AC1:** Declare `github.com/pheelee/wishmas` as the Go module path and update all application/test internal imports to that path.
- **CON-002-AC2:** Use Go 1.27.3 consistently in the module directive, Docker build image, container-test helper image and other active version references. Verify dependency resolution, tests, vet, no-CGo builds and the Docker build with this toolchain.

Verification (`@test CON-002-AC1 CON-002-AC2`): Go module metadata resolves `github.com/pheelee/wishmas 1.27.3`; dependency verification, full race tests, vet and no-CGo/nodynamic builds pass. Both Docker executables' embedded build metadata confirms the requested module paths, Go 1.27.3 and Go-only build settings. The rebuilt scratch image passes the full isolated Traefik HTTPS container flow. Dependency versions remain unchanged.

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
- **CON-004-AC5:** Retain local-mode image replacement/removal/deletion failure rollback and immediate old-file removal. S3 post-save cleanup retries and background recovery in both modes follow WSH-008. Images remain under household `pictures/`; editing an unchanged source does not itself re-download it.
- **CON-004-AC6:** Do not silently migrate legacy files at startup. Migration is explicit and copy-only until approved cutover. Remove automatic-year, timezone, read-only-history and archive/restore persistence/behavior from the new runtime.

### CON-005: Internationalization delivery and configuration

Retain bundled de-CH/en-US only, informal Swiss German, no external translation download and no currency. Language remains `language` in the single household file; missing is English, present invalid/empty/null configuration is an English error without writes. Everyone, including the admin login page, uses valid household language after setup. Changing language changes presentation only, not IDs/references/content/codes/claims or paths. Price entry/storage remains ungrouped decimal; individual wish display follows locale grouping/two decimals, while overview totals remain ungrouped/two decimals. See revised INT-001, INT-005–INT-007 and WLS-011. New wording requires separate table approval; remove unused legacy messages.

### CON-006: Wish picture processing and access

Retain approved WSH-006 download/processing/presentation behavior:

- Complete HTTP/HTTPS image URLs only, no uploads; public internet addresses including redirects; maximum five redirects/30 seconds/10 MiB/20 megapixels.
- JPEG/PNG/WebP/GIF/AVIF, no SVG. Animated sources use their first frame only. Proportional maximum 640px without upscaling; lossy WebP quality 80, preserve alpha, discard originals.
- Source URL and optimized filename remain in wish data; optimized pictures move to household `pictures/` alongside the one household data file. No person/year/event image directories in new storage.
- Keep centered shelf crop, full-image animated modal/backdrop dismissal and reduced-motion behavior. Text stays readable; browsing itself never fetches source images. The approved startup/background recovery exception follows WSH-008.
- Admins can view all household pictures; user links can view only pictures of wishes currently assigned to their list. Enforce this on direct image requests, including after moves/code regeneration. There is no year/archive scope.

### CON-007: Passwords, sessions and proxy contract

Approved technical details are in [storage.md](storage.md).

- **CON-007-AC1:** Use salted Argon2id encoded with algorithm/version/cost/salt/hash. Initial cost is 19 MiB/two iterations/one lane, 16-byte random salt and 32-byte result. Reject malformed/unsupported or excessive encoded parameters without unbounded allocation.
- **CON-007-AC2:** A private terminal helper validates the same approved password rules, requests matching confirmation and prints a YAML-ready hash. Passwords are not command arguments or stored plaintext in configuration/review logs.
- **CON-007-AC3:** Codes/session tokens use 32 cryptographically random bytes. Store shareable list codes but only SHA-256 hashes of session tokens, bound to the current password-hash fingerprint. Password changes invalidate old sessions; logout revokes the current session and clears its cookie only. S3 record removal may be deferred under ACC-008 while a durable local denial remains enforced.
- **CON-007-AC4:** Use persistent Secure/HttpOnly/SameSite=Lax cookies and mutation/CSRF protection, with no app inactivity timeout. Session records persist in household YAML. Browser cookie removal/retention limits remain outside the server's control.
- **CON-007-AC5:** Derive rate-limit IP from the forwarded chain on the trusted Traefik side, configured with deployment-specific trusted proxy addresses. Do not blindly trust a client-supplied leftmost address. Five failures in fifteen minutes blocks that IP for fifteen minutes; in-memory counters/blocks may reset at restart.
- **CON-007-AC6:** Verify cookies/login against isolated HTTPS in review/tests; do not weaken production cookie settings to enable local review. Document deployment-specific trusted addresses rather than guessing them.

### CON-008: Explicit migration and reversible review

Historical constraint: legacy migration support is retired by CON-010 below.

- **CON-008-AC1:** Read legacy data without modifying the original; write a separate migration destination and refuse an existing/nonempty destination. Require the initial `admin_password_hash` in a copied legacy household config. The helper is available before the user's data migration.
- **CON-008-AC2:** For current data, create Christmas 2026 from the existing year, preserve recipients and each stored wish/assignment/detail/claim, and expose formerly archived recipients normally. General year discovery preserves ascending order; no wish deduplication is performed.
- **CON-008-AC3:** Create household users from claimant names, merging only case/surrounding-space variants, preserving first encountered spelling and distinguishing accented names. Convert claims to stable user IDs and generate individual user/list links. Old parent/shared-family links cease to work in the new app.
- **CON-008-AC4:** Copy optimized images byte-for-byte to unique household filenames, preserving source URLs and missing-image behavior. Do not download/re-encode during migration. Preserve existing language/appearance and discard obsolete timezone/archive/year access configuration.
- **CON-008-AC5:** Invalid/unrepresentable source data stops migration with a report rather than silently dropping, renaming or merging data beyond approved claimant matching. Report source/destination recipients/wishes/claims and image checksums without exposing password hashes/session tokens/link codes.
- **CON-008-AC6:** Review uses a separate data copy and separate app environment. Original data and baseline code remain available. No merge or live-data cutover until user approval; pause after each verified milestone.

### CON-010: Retire completed legacy migration

User-approved scope: remove legacy migration now that the user's data is migrated; review the code before any commit.

- **CON-010-AC1:** Remove the legacy converter and `wishmas-admin migrate` subcommand. Attempts to invoke it return an unknown-command error without prompting, emitting a report or modifying source/destination files.
- **CON-010-AC2:** Preserve `hash-password`, current household storage and unrelated persistence regression coverage, including failed-save data preservation.
- **CON-010-AC3:** Remove migration-only tests and active usage instructions; retain historical migration records, clearly distinguished from current supported operations. This supersedes CON-008's requirement to ship a migrator; legacy data remains rejected at startup.

## OPEN deployment items

Exact VPS trusted-proxy addresses are supplied at deployment. New wording and the milestone plan are approved. No unresolved product behavior is implied by previous verification notes; new acceptance evidence is pending.

### CON-009: Optional S3 household YAML storage

The approved contract is in [s3-storage.md](s3-storage.md), including CON-009-AC1–AC16 and the approved environment-variable choices. **Implemented and verified: CON-009-AC1–AC16.** Access/Wishes criteria and HR113–HR120 wording are also approved. Repository checkpoint cbfa19f and complete app/background/container delivery are verified; details and external provider/deployment limits are recorded in progress.md.

### CON-010: Dependency maintenance and release delivery

User-approved DEP-001, AUT-001 and REL-001 criteria are recorded in [release-delivery.md](release-delivery.md). These extend deployment CON-001 and supersede the fixed Go 1.27.3 pin in CON-002-AC2 with the approved latest-stable refresh and consistently tracked Go toolchain. Required PR Go checks, weekly grouped Renovate auto-merges, manual 0.1.0-first releases and multi-platform GHCR publication are approved; dependency baseline is verified and automation/release code is locally verified; remote enforcement/publication verification remains pending.

Confirmed REL-003 revision (2026-11-02) supersedes REL-002 container test obligations: container/Compose suites and their workflow are removed. Core Go/browser/build/multi-platform/version checks remain required; both images are statically verified without runtime testing. Runtime/deployment behavior is unchanged.
