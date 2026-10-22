# Wishmas

A family wishlist app with a December-night Christmas theme, Go backend, Tailwind CSS/vanilla JavaScript and YAML storage. Requirements live in [sdd/](sdd/README.md).

## Current branch checkpoint

`refactor/household-lists-access` has completed milestones 0–3: approved specification/wording, the version-2 data core/offline migration tools, password setup/admin authentication, and manual lists/household rosters/individual user links. The default server uses the new password model. Household wish management/cards/summaries follow in milestone 4; user-ID claim actions follow in milestone 5. List users currently receive scoped read-only recipient/wish details.

The original data and running app have not been migrated/restarted. Baseline `main` remains at `1992bdf`; use matching baseline code with legacy data. The new server rejects legacy configuration without changing it. Actual household migration/review is milestone 6, using a copy; merge and live cutover require approval.

## Review setup and access

See [password access and isolated HTTPS review](docs/admin-access.md) for automated and manual review using a fresh temporary data root. Do not point review at the original `data/` directory.

Setup asks for household name, shared admin password/confirmation, household language and appearance. There is no timezone, automatic year switch or parent/shared-family link. Errors clear passwords and preserve other usable inputs; finishing creates an empty household and signs in automatically. Admin sessions persist across restarts, logout affects one browser, and editing the password hash invalidates existing sessions. After setup, language/password changes are manual YAML edits.

Parents can create, rename and delete freely named lists and household recipients/users. Select recipients per list and copy or independently replace each user/list link. Confirmed deletions retain wishes/pictures while clearing the affected assignments/claims. Household wish management, interactive cards/images/summaries and claim actions remain pending; follow the [implementation plan](sdd/implementation.md) and [verified progress](sdd/progress.md).

## Production transport and configuration

Use HTTPS through Traefik, redirect HTTP to HTTPS and keep the backend port private. Secure cookies and mutation checks deliberately require HTTPS; direct plain-HTTP setup is not supported. Configure the actual trusted Traefik peers rather than trusting arbitrary forwarded headers.

| Setting | Meaning |
| --- | --- |
| `WISHMAS_ADDR` | Backend listening address, default `:8080` |
| `WISHMAS_DATA_DIR` | Data root, default `data` locally and `/data` in the image |
| `WISHMAS_TRUSTED_PROXIES` | Comma-separated Traefik peer IPs/CIDRs; none trusted by default |

See [the proxy/header contract](docs/admin-access.md#traefik-contract). Deployment-specific addresses are still to be supplied. Run one writing app instance per data directory.

The Dockerfile builds the current default server with embedded frontend assets and the Go WebP backend, using UID/GID `65532:65532`:

```sh
docker build -t wishmas:local .
```

Keep the data volume across replacement and give the runtime user write access. Complete new-flow Docker/Traefik smoke verification and final deployment instructions are milestone 6; the existing `tests/container/smoke.cjs` still describes the legacy flow and must be adapted then. No actual VPS deployment is claimed.

## Offline tools and storage

[Offline administration](docs/offline-admin.md) documents private password-hash generation and explicit copy-only migration:

```sh
go run ./cmd/wishmas-admin hash-password
```

The [version-2 storage contract](sdd/storage.md) uses one private `household.yaml` per household plus separate optimized pictures. It stores stable IDs, language/appearance, password hash, hashed sessions, rosters, lists/user links and wishes/assignments/claims. Invalid edits show an error without rewriting data. Back up the entire private data root; never commit its credentials, links or user data.

## Build and verify

Use the Go version in `go.mod`. Node.js is needed for CSS rebuilding/browser tests, not runtime:

```sh
npm ci
npx playwright install --with-deps chromium
npm run build:css
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build -tags nodynamic ./...
npm run test:e2e
```

The suite currently includes four new HTTPS authentication scenarios and 27 legacy-domain browser checks in a test-only server. Legacy checks protect not-yet-migrated components; they do not claim the new management UI is complete. The production binary has no legacy mode. English/German catalog completeness, security, data preservation and responsive theme checks remain enforced.

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md) and [the progress checkpoint](sdd/progress.md). Inspect Git status before editing. Stop after each verified milestone for review; never mark a domain Implemented with unmet criteria.
