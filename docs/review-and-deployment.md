# Household refactor review and deployment

The refactor stays on `refactor/household-lists-access`. Baseline `main` is `1992bdf`. This document describes local verification and a separate review copy; it does not claim deployment to the VPS or authorize merge/cutover.

## Current-data review

The isolated review uses `.work/household-refactor/milestone-6-private/review`, migrated from a consistent private snapshot of the original `data/households/` directory. It preserves one Christmas 2026 list, three recipients, seven wishes, three claims, one household user and six byte-identical pictures. Original `data/` and the separately existing `data2/` are untouched. The existing source password hash was valid and retained; no new password was invented.

Open **https://localhost:8092/** through the workspace's HTTPS port forwarding. This isolated harness uses a temporary self-signed certificate. Sign in using the password corresponding to the retained hash; user links are available on the list page. The production image continues to require HTTPS through Traefik. Do not publish private codes or credentials in review notes.

To start or resume review from the repository root:

```sh
go test -c -o .work/household-refactor/wishmas-review ./internal/app
WISHMAS_BROWSER_TEST_MODE=household \
WISHMAS_ADDR=0.0.0.0:8092 \
WISHMAS_DATA_DIR="$PWD/.work/household-refactor/milestone-6-private/review" \
.work/household-refactor/wishmas-review
```

Stop this foreground process with Ctrl-C. Do not stop or replace the existing app on port 8090. The review directory is writable for manual review; changes there never update the source. Private snapshot, screenshots, server executable and credentials remain ignored/uncommitted.

If the retained password is unknown, use the private helper to generate a new hash and edit **only the review copy's** `admin_password_hash`, then sign in. This invalidates review sessions; it does not change the original password. Never send the password or hash in chat.

Review admin login, lists/recipient/user rosters, copied pictures, current assignments/claims, independent list links, regeneration and user privacy before deciding whether to merge. Missing claimant-derived users can be added by parents. A name in the recipient roster does not automatically create a user.

## Container verification

```sh
docker build -t wishmas:local .
WISHMAS_SMOKE_IMAGE=wishmas:local node tests/container/smoke.cjs
WISHMAS_SMOKE_IMAGE=wishmas:local WISHMAS_SMOKE_LANGUAGE=de-CH node tests/container/smoke.cjs
```

The smoke test creates its own Docker network, named volume, nonroot app and real Traefik container. It publishes only Traefik, derives the trusted address from that isolated proxy, and never mounts repository data. TLS uses Traefik's generated test certificate; test browsers alone accept it. The test verifies HTTP redirect, HTTPS setup, secure sessions, scoped users/claims, public HTTPS image processing, identical YAML/image bytes across app restart/replacement and client-IP limits despite spoofed forwarded headers. It removes its containers/network/volume on completion. It supports a remote DinD daemon through `DOCKER_HOST` and `docker cp`, without assuming workspace bind mounts are visible to that daemon.

The Traefik test image is pinned by digest in the test. Its [entrypoint configuration](https://doc.traefik.io/traefik/v3.6/reference/install-configuration/entrypoints/) leaves untrusted forwarded-header acceptance disabled. Routing uses the [file provider](https://doc.traefik.io/traefik/reference/dynamic-configuration/file/); no Docker socket is mounted.

## VPS configuration boundary

Build the image above and retain the complete `/data` volume between replacements. The image runs as UID/GID **65532:65532**; give that user write access to the private data root and files, including atomic file replacement and picture creation/removal. Mount the complete root, not just `household.yaml`. Run one writing app instance per data directory.

Traefik must terminate valid HTTPS for the chosen hostname, redirect HTTP, preserve the public host, and forward the actual client IP and `X-Forwarded-Proto: https`. Keep the Wishmas backend `:8080` reachable only by the trusted proxy, with no public port mapping. Set `WISHMAS_TRUSTED_PROXIES` to the actual Traefik peer IP(s)/restricted CIDR(s); do not use universal or public-client ranges. Leave Traefik's insecure forwarded-header acceptance disabled. If another trusted proxy sits in front, configure that actual chain explicitly. No app TLS bypass or weaker cookie configuration is provided.

Exact VPS hostname, certificate resolver, network, proxy addresses, storage paths and service names remain deployment inputs. Supply and review their concrete configuration before any actual cutover; the local verification network is not a guessed production configuration.

## Merge, cutover and rollback

No merge or live cutover has occurred. After approving the review, choose whether to merge this branch. Before replacing a live app, stop its writer and take a **fresh consistent private snapshot** if source data changed since this review. Add/retain the approved initial hash in the copied legacy config, migrate into a previously absent destination, and compare the report and image checksums. Never reuse the writable review copy as authoritative live data.

Keep original code/data backups, then deploy the new image against the new data root through the reviewed Traefik configuration. Verify login, links, recipients, wishes, claims and images before reopening access. Old parent/family links are intentionally invalid; share the new individual user/list links.

Rollback uses **both** baseline code `1992bdf` and its matching untouched legacy data. Never point baseline code at version-2 data or the new app at the legacy hierarchy. Edits made after cutover are not automatically reverse-migrated; restoring the old snapshot would omit them, so resolve that explicitly before rollback. Remove the isolated review environment only after the user's final decision.
