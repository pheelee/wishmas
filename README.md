# Wishmas

A family wishlist app with a December-night Christmas theme, Go backend, Tailwind CSS/vanilla JavaScript and YAML storage. Requirements live in [sdd/](sdd/README.md).

## Current branch checkpoint

The refactor is merged into `main` at the user’s request; its local working branch is deleted. The pre-refactor baseline remains `1992bdf` in history. This merge does not replace live data or deploy the app.

`refactor/household-lists-access` has completed milestones 0–5: approved specification/wording, the version-2 data core/offline migration tools, password access, lists/rosters/individual links, and household wishes/assignments/cards/pictures/recipient summaries. User-ID claim actions and lifecycle integration are verified. Milestone 6 adds the copied current-data review, final legacy/catalog cleanup and real Docker/Traefik persistence verification.

The original data and running app have not been migrated/restarted. The pre-refactor baseline remains available at `1992bdf`; use matching baseline code with legacy data. The new server rejects legacy configuration without changing it. The household review uses a separately migrated private copy at HTTPS port 8092; the merge is complete and live cutover still requires approval.

## Review setup and access

See [password access and isolated HTTPS review](docs/admin-access.md) for automated and manual review using a fresh temporary data root. Do not point review at the original `data/` directory.

Setup asks for household name, shared admin password/confirmation, household language and appearance. There is no timezone, automatic year switch or parent/shared-family link. Errors clear passwords and preserve other usable inputs; finishing creates an empty household and signs in automatically. Admin sessions persist across restarts, logout affects one browser, and editing the password hash invalidates existing sessions. After setup, language/password changes are manual YAML edits.

Parents can create, rename and delete freely named lists and household recipients/users. Select recipients per list and copy or independently replace each user/list link. Confirmed deletions retain wishes/pictures while clearing the affected assignments/claims. Parents can also create/edit/delete household wishes, filter them, move or unassign them without losing claims, and create automatically assigned wishes on recipient pages. Scoped list views retain cards, pictures, sorting and exact recipient totals/counts. List users claim/remove their own claims directly; admins select a household user per new claim and can remove any claim, including retained unassigned claims. Claims use stable user IDs and preserve other users’ claims. Follow the [implementation plan](sdd/implementation.md) and [verified progress](sdd/progress.md).

Pictures accept JPEG, PNG, WebP, GIF and AVIF URLs. They are stored as optimized, still WebP images; animated sources use their first frame.

## Production transport and configuration

Use HTTPS through Traefik, redirect HTTP to HTTPS and keep the backend port private. Secure cookies and mutation checks deliberately require HTTPS; direct plain-HTTP setup is not supported. Configure the actual trusted Traefik peers rather than trusting arbitrary forwarded headers.

| Setting | Meaning |
| --- | --- |
| `WISHMAS_ADDR` | Backend listening address, default `:8080` |
| `WISHMAS_DATA_DIR` | Data root, default `data` locally and `/data` in the image |
| `WISHMAS_TRUSTED_PROXIES` | Comma-separated Traefik peer IPs/CIDRs; none trusted by default |

See [the proxy/header contract](docs/admin-access.md#traefik-contract). Deployment-specific addresses are still to be supplied. Run one writing app instance per data directory.

The Dockerfile builds the current default server with embedded frontend assets and the Go WebP backend, defaulting to UID/GID `1000:1000`:

```sh
docker build -t wishmas:local .
```

Keep the complete data volume across replacement and give the selected UID/GID write access. The image also includes `/wishmas-admin` for explicitly invoked offline administration. Isolated Docker/Traefik HTTPS verification covers secure sessions, spoof-resistant limits, scoped user links, claims, pictures and restart/replacement persistence. See [review, deployment and rollback](docs/review-and-deployment.md). No actual VPS deployment or live cutover is claimed.

### Compose reference and runtime identity

[docker-compose.yml](docker-compose.yml) defaults to UID/GID `1000:1000`. Override them independently in `.env` or the shell:

```dotenv
WISHMAS_UID=1234
WISHMAS_GID=5678
```

Compose’s `user` setting applies at container creation; no image rebuild is needed when changing IDs:

```sh
docker compose up -d --no-build
```

Build/tag the image first if it is not available (`docker compose build`). Prepare the bind-mounted version-2 data root with ownership/permissions for the selected IDs. The reference does not create a missing host directory or change existing file ownership. Configure the actual Traefik network/routing and `WISHMAS_TRUSTED_PROXIES` before use; no backend port is published. See [deployment details](docs/review-and-deployment.md#runtime-identity-and-compose-reference). Do not run the reference against the repository's live legacy data.

## Offline tools and storage

[Offline administration](docs/offline-admin.md) documents private password-hash generation and explicit copy-only migration:

```sh
go run ./cmd/wishmas-admin hash-password
```

The [version-2 storage contract](sdd/storage.md) uses one private `household.yaml` per household plus separate optimized pictures. It stores stable IDs, language/appearance, password hash, hashed sessions, rosters, lists/user links and wishes/assignments/claims. Invalid edits show an error without rewriting data. Back up the entire private data root; never commit its credentials, links or user data.

## Build and verify

Use Go **1.27.3**, as pinned in `go.mod`. The module path is `github.com/pheelee/wishmas`. Node.js is needed for CSS rebuilding/browser tests, not runtime:

```sh
npm ci
npx playwright install --with-deps chromium
npm run build:css
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build -tags nodynamic ./...
npm run test:e2e
```

All browser scenarios now use the current household runtime over isolated HTTPS. New model checks and adapted sorting/theme/card regressions cover setup/access, lists/rosters/links, household wishes/assignments/pictures, user-ID claims and responsive appearance. Retired year/archive/name-cookie integration suites and their runtime are removed; baseline proof remains in Git. Catalog tests enforce exactly the 223 approved current English/German entries.

For real container/proxy verification, run `node tests/container/smoke.cjs` (Docker and Playwright Chromium required). Set `WISHMAS_SMOKE_LANGUAGE=de-CH` for German. It creates and removes isolated test resources and never mounts your household data. See [the detailed verification and review instructions](docs/review-and-deployment.md).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md) and [the progress checkpoint](sdd/progress.md). Inspect Git status before editing. Stop after each verified milestone for review; never mark a domain Implemented with unmet criteria.

## Optional S3 YAML backend

Local storage is the default. For an existing S3-compatible bucket, configuration, operator upload, outage behavior and local image recovery, see [S3 storage](docs/s3-storage.md). The reference Compose file includes the optional environment variables; `/data` remains persistent in both modes.
