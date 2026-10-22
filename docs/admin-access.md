# Password access and isolated review

Milestone 2 switches the default server to version-2 setup and password authentication. The original running app and `data/` have not been restarted or migrated. A rebuilt server requires a fresh empty data root or explicitly migrated version-2 data; legacy configuration returns an English configuration error without writes. Real household migration remains milestone 6.

## Setup and login

Open `/` over HTTPS. Initial setup asks for household name, password/confirmation, household language and appearance. Password rules match the offline helper: at least eight characters, Unicode uppercase/lowercase/digit and punctuation/symbol. Setup includes passphrase guidance, live sliders and no timezone or parent/family links.

Errors clear both passwords while preserving other usable draft settings. The wizard keeps its initially detected language; the selected household language applies after Finish setup. Successful setup saves one household file with empty lists/rosters/wishes and a hashed session, then signs the browser in automatically.

After setup, missing/invalid codes show household-language password login. Valid user/list codes take precedence over a browser's separate admin session. Milestone 3 adds manual lists, household rosters, individual links and scoped read-only recipient/wish views. Milestone 4 adds household wish CRUD/filters, assignment/moves/unassignment, retained cards/sorting/pictures and exact scoped summaries; claim actions follow in milestone 5.

## Persistent sessions

- Admin session cookies are Secure, HttpOnly, SameSite=Lax and host-only; the server stores only token hashes bound to the password-hash fingerprint. Cookies are persistent and renewed on admin visits, within browser retention limits. There is no server inactivity timeout.
- Restart/replacement with retained data preserves sessions. Logout removes only the current browser's session record/cookie.
- Generate a replacement hash privately with `go run ./cmd/wishmas-admin hash-password`, then edit `admin_password_hash` in the household's YAML. Next requests invalidate and remove sessions bound to the old hash. There is no in-app password change.
- State-changing requests require HTTPS and a matching host-only CSRF nonce, with Origin/fetch-site checks. The new pages use `Referrer-Policy: same-origin`: native forms retain a verifiable origin while external requests receive no code-bearing referrer.

## Traefik contract

Production stays HTTPS-only through Traefik, with HTTP redirected and no publicly exposed backend port. Pass the original Host and a single `X-Forwarded-Proto: https` value. Configure the backend with the actual trusted proxy peers/networks:

```text
WISHMAS_TRUSTED_PROXIES=<Traefik peer IP or CIDR, optionally comma-separated>
```

No proxy addresses are trusted by default; invalid/trust-all configuration fails startup. IP rate limits traverse all X-Forwarded-For header values from the trusted side, stopping at the first untrusted client. Forged leftmost values cannot change that identity; malformed chains fall back to the peer. Duplicate/ambiguous protocol headers are not treated as secure transport. Deployment-specific addresses must be supplied before deployment, not guessed from these examples.

Five failed logins from one IP in fifteen minutes block further attempts for fifteen minutes; blocked responses include Retry-After. State is in memory and may reset on restart. Expired entries are cleaned up. A valid password does not bypass an active block.

## Review without touching the current app

Automated HTTPS review:

```sh
npx playwright test tests/browser/admin-auth.spec.js tests/browser/admin-management.spec.js tests/browser/household-wishes.spec.js
```

These tests use temporary data, independent browsers, real server replacement and self-signed local HTTPS. Screenshots under `test-results/` include English/German setup/login, household/list pages and mobile confirmations. They contain disposable fixtures, never the real household.

For a manual fresh-household review, choose an unused loopback port (8443 below is only an example):

```sh
wishmas_review_root="$(mktemp -d /tmp/wishmas-review.XXXXXX)"
go test -c -o "$wishmas_review_root/review-server" ./internal/app
WISHMAS_BROWSER_TEST_MODE=household \
WISHMAS_DATA_DIR="$wishmas_review_root/data" \
WISHMAS_ADDR=127.0.0.1:8443 \
"$wishmas_review_root/review-server"
```

Open `https://127.0.0.1:8443/`. This test-only server uses a self-signed certificate, so the browser displays a certificate warning. Stop with Ctrl+C. Rerun the same command/data root to review session persistence. The production binary does not recognize the browser-test mode and has no legacy switch.

The test harness also preserves untouched legacy-domain scenarios while their replacements are pending. Those checks are compatibility evidence, not proof that the new list/wish/claim UI is complete. The final runtime/catalog cleanup is milestone 6. Merge, real-data migration and live cutover remain separate review decisions.

## Milestone 3 review

Use a fresh disposable household with the HTTPS harness above. Existing wishes/claims in automated cleanup tests are seeded fixtures; there is no new wish/claim creation UI yet.

1. Create household recipients and users independently; the same name can exist once in each roster. Create two freely named lists. No year or archive controls appear.
2. Open a list and explicitly select recipients. Every household user has a separate link for each list, whether the user or list was created first. Selecting a recipient for one list does not select them for another.
3. Copy two links. The clipboard button shows a checkmark for three seconds. Open them in separate browser profiles: each identifies its predefined user, displays every selected recipient in that list and never prompts for a name. A user link restricts even a signed-in admin browser.
4. Replace one link using its confirmation modal. Reload the old link: it asks for the admin password in the household language. Other links still work. Renaming a list/recipient/user preserves stable IDs, links, assignments and claims.
5. Inspect deletion/removal modals and cancel first. They explain consequences and count affected wishes/claims. Confirming list deletion or recipient removal/deletion unassigns the affected wishes and clears their claims, retaining wishes/pictures. Deleting a user removes that user's links and all their claims, including on unassigned wishes.
6. Open a confirmation, then change the household YAML or save another edit. Submitting the old form reports changed household data without saving; reload before retrying. Invalid drafts and failed writes leave existing data unchanged.

Milestone 4 was subsequently authorized and completed below. No real-data migration, branch merge or live restart occurred.

## Milestone 4 review

Use a fresh disposable version-2 household and the isolated HTTPS harness above. Do not point it at original data or an app already writing that directory. The user's untracked data2/ is not touched by this milestone.

1. On the household page, create a title-only wish. It starts Unassigned. Add a reference price, multiple shop links, note and optional image URL using the edit modal. Decimal entry uses a point; optional fields can be cleared. Invalid input keeps the draft and saved data unchanged.
2. Create lists, select recipients and use the card's assignment action to choose a list/recipient pair. Changing the selection directly moves the same wish; selecting Unassigned returns it to the future pool. Details, added order, picture and existing claims remain unchanged.
3. Open a list, then a recipient summary. Wishes created there automatically belong to that list/recipient. Edit/delete target the same stable wish IDs from either context. Cancel deletion first, then confirm: that wish, its claims and local picture disappear permanently.
4. Review household filters for list, recipient and unassigned wishes, and the shared remembered sort preference. Reference totals include every supplied price; green/yellow badges count complete wishes only. An empty recipient shows 0.00 and zero badges. Amounts have no currency.
5. Open individual user/list links in separate browser profiles. They show every selected recipient, assigned cards/pictures and claim counts/own status, with no other claimant names or management forms. A move immediately removes the wish/picture from the old list on reload, including direct known image URLs. Unassigned wishes/pictures stay admin-only. Regenerated codes cannot continue serving pictures.
6. Review centered picture shelves and the full-image modal on desktop/mobile. Clicking its exterior or Escape closes it; clicking inside does not. Reduced motion skips animation. A configured missing picture shows Image unavailable without retrying its source. An unchanged image URL is never fetched again; failed processing/saves/removal preserve the previous data and picture.

The automatic tests seed claims to verify these lifecycle effects. Claim creation/removal buttons, including admin selection of a household user per claim, remain milestone 5. Real household migration, final legacy cleanup and container/Traefik review remain milestone 6. Stop for review before advancing.
