# Household storage and authentication

Approved by the user as the technical/storage contract. Milestone 1 implements the data core, private password helper and explicit migration command; Milestone 2 integrates setup and persistent admin sessions; management/lifecycle UI and actual household migration remain pending. Owned by CON-004 and CON-007–CON-008 in constraints.md. Schema examples contain placeholders only; no real credentials or household data.

## Files

- Keep `data/households/<household-name>/household.yaml` as the one persistent data file.
- Keep optimized pictures under the same household's `pictures/` directory, addressed by stable wish ID plus a replacement-specific filename. No year/event/person directories in new storage.
- Add `schema_version: 2` to distinguish new data from the legacy format. No automatic in-place migration on app startup.

Illustrative shape only; identifiers/hashes/codes below are placeholders, not credentials:

```yaml
schema_version: 2
name: Family
language: en-US
appearance:
  bulb_size: 120
  snow_size: 100
  snow_amount: 16
admin_password_hash: <argon2id-hash>
sessions: []
users:
  - id: user-dad
    name: Dad
recipients:
  - id: recipient-leon
    name: Leon
lists:
  - id: list-christmas
    name: Christmas 2026
    recipient_ids: [recipient-leon]
    user_links:
      - user_id: user-dad
        code: <random-list-user-code>
wishes:
  - id: wish-train
    title: Model train
    price: "25.00"
    links: [https://shop.example/train]
    note: Optional note
    picture_source: https://images.example/train.png
    picture: <wish-id>-<image-version>.webp
    assignment:
      list_id: list-christmas
      recipient_id: recipient-leon
    claims:
      - user_id: user-dad
```

Omit `assignment` for unassigned wishes; their claims may remain. Arrays preserve added order. Wish array position is household added order; assigning/moving a wish does not reset that order. User, recipient, list and wish IDs are opaque immutable generated values; sample readable IDs are only explanatory.

## Authentication details

- Argon2id through Go's `golang.org/x/crypto/argon2`, encoded with algorithm/version/parameters/salt/hash. Initial parameters: 19 MiB memory, two iterations, one lane, random 16-byte salt, 32-byte result. Reject unsupported/malformed parameters without attempting unbounded allocation.
- Private terminal helper enforces the approved password rules and matching confirmation, then emits the YAML-ready hash. Do not pass passwords as command-line arguments or store them in review notes.
- Generate codes and session tokens from 32 cryptographically random bytes. Persist list codes so parents can copy/share them; persist only SHA-256 hashes of admin session tokens.
- Store each session's token hash and password-hash fingerprint in `sessions`. Bind sessions to the current password hash; on detecting a change invalidate all old sessions. Current-browser logout removes that session record and cookie.
- Use persistent Secure, HttpOnly, SameSite=Lax cookies; enforce mutation request/CSRF protections. No app-defined inactivity timeout. Browsers can still clear or limit cookie storage; no server can guarantee retention after a browser deletes its cookie.
- Rate limit by the verified client IP supplied through Traefik. Document the trusted proxy addresses and parse the forwarded chain from the trusted side; never accept a user-supplied leftmost value blindly. Exact deployment addresses are supplied at deployment, not guessed. In-memory limit state may reset on restart.
- Production: HTTPS via Traefik, HTTP redirected, no public app port. Review/test authentication uses isolated HTTPS so it does not require weakening production cookies.

## Data integrity and migration

- Validate IDs, references, recipient selection, unique names, codes, claims and all existing field/image rules. Invalid manual edits show an error without rewriting the data.
- Reload before mutations, reject stale edits, serialize app writes and atomically replace the household YAML file. Preserve unrelated data/claims. Existing picture rollback/deletion semantics stay in force.
- Explicit migration command reads the untouched legacy source and writes a separate destination; refuse an existing/nonempty destination. Require `admin_password_hash` in a copied legacy household config before migration; the original remains unchanged.
- For the current data: one 2026 list, household recipient roster from people.yaml, all recipients visible, each saved wish separately preserved with details/claims. Create users by case-insensitively merging claimant names after trimming; retain first encountered spelling and keep accented names distinct.
- Copy optimized images byte-for-byte without downloading/re-encoding, assign unique destination filenames, and preserve source URLs. Preserve missing-image behavior.
- No deduplication of existing wish occurrences. Invalid or unrepresentable source data fails migration with a report; no silent dropping, renaming or merging beyond approved claimant-name matching.
- Produce a comparison report of source/destination recipients, wishes, claims and image checksums without exposing password hashes, session tokens or link codes. Keep original data and baseline code available for rollback; do not merge or cut over without review approval.

## Sources

Sources: [Go Argon2 package](https://pkg.go.dev/golang.org/x/crypto/argon2), [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html), [Traefik entrypoints/forwarded headers](https://doc.traefik.io/traefik/reference/install-configuration/entrypoints/). Exact module version remains to be selected/verified for the pinned Go toolchain during implementation.

## Milestone 1 implementation and verification

- `internal/household` supplies the version-2 model, reference/name/code/claim/detail validation, missing-versus-invalid language handling, stable IDs, preserved future wish/claim fields and fingerprinted atomic repository updates. Independent store instances share the app write lock; stale/concurrent manual changes are rejected. Invalid edits and failed writes retain original bytes.
- `cmd/wishmas-admin` exposes private terminal `hash-password` and explicit `migrate --source HOUSEHOLD_COPY --destination NEW_HOUSEHOLD`. See [offline tool instructions](../docs/offline-admin.md). At milestone 1 the HTTP app had not yet switched to version 2; milestone 2 now switches the default server.
- Failing-first migration test initially found no command. Further failures exposed conflicting future fields causing a serializer panic, symlink-following and merged-language fallback; these now reject safely or preserve the correct effective language. Tests cover invalid references, malformed/excessive hash costs, policy/salts, confirmations, claims/name matching, duplicate wish occurrences, year order, missing images, exact copied bytes, unknown files and source/destination safety.
- `go test -race ./...`, `go vet ./...`, no-CGO/nodynamic builds and whitespace checks pass. A real pseudo-terminal smoke verifies hidden input and usable helper output; a disposable CLI migration verifies matching counts/checksums, untouched source and refusal to reuse the destination.
- The full suite initially exposed the legacy catalog test reading the revised target spec. Its independent 182-row approved baseline inventory is now in `internal/app/testdata/legacy-translations.md`; remove/adapt that fixture when replacing the UI/catalog. No test was skipped to claim target runtime behavior. The real permission-failure test passes as the non-root workspace user; like existing permission tests it skips when running as root.
- No real household data was migrated. Session authentication, image mutation transactions, HTTP roles/scopes and all new UI remain later milestones. CON-004/CON-008 are not marked wholly Implemented on the basis of offline fixture verification.

## Milestone 2 authentication integration

The default server now reads only validated version-2 households and creates empty households through password setup. Raw session tokens stay in Secure/HttpOnly/host-only cookies; only their hashes and password-hash fingerprints are stored. Logout removes one session; an observed password-hash change removes old sessions, including across app restarts. Stale/failed writes preserve stored data and browser authentication state.

Go race tests and four HTTPS browser scenarios verify setup storage, no plaintext passwords/raw tokens, real process replacement, independent sessions/logout, hash edits/revocation, CSRF/origin/external-referrer protection and trusted Traefik IP/HTTPS headers. No actual data migration or container proof is claimed; those remain milestone 6.
