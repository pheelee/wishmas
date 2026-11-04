# Optional S3 household storage

## Status and ownership

**Implemented: CON-009-AC1–AC16.** Owned by CON-009 and the existing household storage contract. Local remains the default; S3 is opt-in. Verification and provider/deployment limits are recorded below and in progress.md.

This document covers backend selection, S3 configuration, household objects, source authority, cached reads and storage failures. ACC-008 specifies cached access and durable logout; WSH-008 specifies background image recovery and cache cleanup. The user approved the reviewed English/German wording and authorized implementation; execution slices are recorded in implementation.md.

## Configuration

Configuration uses environment variables only; there is no app settings screen.

| Variable | Default / requirement | Meaning |
| --- | --- | --- |
| `WISHMAS_STORAGE` | `local`; accepts `local` or `s3` | Selected YAML backend |
| `WISHMAS_S3_ENDPOINT` | Required in S3 mode | HTTP or HTTPS S3-compatible endpoint |
| `WISHMAS_S3_BUCKET` | Required in S3 mode | Existing bucket holding household objects |
| `WISHMAS_S3_REGION` | `us-east-1` | Region used for request signing |
| `WISHMAS_S3_ACCESS_KEY_ID` | Required in S3 mode | Access key |
| `WISHMAS_S3_SECRET_ACCESS_KEY` | Required in S3 mode | Secret key |
| `WISHMAS_S3_SESSION_TOKEN` | Optional | Session token |
| `WISHMAS_S3_TIMEOUT` | `5s` | Timeout per S3 operation |
| `WISHMAS_DATA_DIR` | Existing setting/defaults | Local images, fallback YAML and pending logout records |

Bucket addressing is strictly path-style. There is no configurable object key, prefix, virtual-hosted-style toggle or automatic signing-region discovery. Compatibility targets include Cloudflare R2, MinIO, RustFS and Garage using the configured signing region.

## CON-009: Optional S3 YAML backend

- **CON-009-AC1 — Optional backend:** An unset storage selection uses the existing local YAML backend. Selecting `s3` uses S3 for household YAML. Images remain local in either mode; no images are uploaded to S3.
- **CON-009-AC2 — Configuration:** Apply the environment table above, including HTTP/HTTPS endpoints, strict path-style addressing, configured/default signing region and a default five-second timeout per S3 operation. Invalid storage environment settings cause a non-zero process exit.
- **CON-009-AC3 — Bucket and keys:** The bucket must already exist. The app creates household YAML objects, not buckets. Each household uses `<household-name>/household.yaml` directly inside the configured bucket. Keep the existing household YAML format and validation.
- **CON-009-AC4 — Single-household v1:** If the bucket contains more than one household, show a configuration error rather than select one. The key layout accommodates future multiple-household support without enabling it in v1.
- **CON-009-AC5 — Authoritative source:** In S3 mode, S3 is authoritative. Load current remote household data on page load and before mutation, retaining existing manual-edit validation. A reachable S3 response containing invalid YAML produces a configuration-error page without rewriting it or serving cached household data instead.
- **CON-009-AC6 — Missing household:** If S3 successfully reports that no household YAML object exists, show the existing setup wizard. Ignore an older cached household rather than serving it read-only. Completing setup creates the household object in S3. An inaccessible S3 service must not be mistaken for a missing household.
- **CON-009-AC7 — Single-instance writes:** Support one app instance. Retain the local backend's single-instance write serialization, freshness checks and stale-edit rejection. Do not require S3 conditional-write support. Multiple app instances concurrently sharing the household are outside this revision.
- **CON-009-AC8 — Persistent fallback:** Save the last successfully loaded valid S3 household YAML locally under `WISHMAS_DATA_DIR`, so cached reads remain possible after restart. The fallback belongs to the configured S3 storage target and is not an authoritative replacement for remote data.
- **CON-009-AC9 — Read-only outage mode:** When S3 is temporarily unavailable and a valid fallback exists, serve cached household data read-only. Existing admin sessions and user links may read it; new admin logins are blocked. Show write controls disabled with an explanation. Offline logout is the explicitly confirmed exception, to be specified under Access.
- **CON-009-AC10 — No fallback available:** If S3 is unreachable and no valid cached household exists, start the app and show a storage-error page rather than refusing startup or opening setup. This differs from invalid environment settings in AC2.
- **CON-009-AC11 — Rejected access:** If S3 rejects the configured credentials or permissions, show a configuration error rather than serve cached data. Cached fallback is not a substitute for accepted storage credentials.
- **CON-009-AC12 — Recovery:** The next successful page load from S3 restores normal editing and claiming automatically, using current remote data and the existing stale-edit checks.
- **CON-009-AC13 — Save errors and queues:** Failed edit/claim requests show an error rather than being queued for later. If a timed-out save cannot be confirmed, show an error rather than claiming the save succeeded. Retain applicable existing form-draft preservation. Automatic retries are limited to image recovery/cleanup and pending logout revocations, whose details belong to their owning domains.
- **CON-009-AC14 — Cache-refresh failure:** If saving household YAML to S3 succeeds but refreshing its local fallback fails, keep the save successful and show a warning that the offline copy may be outdated. Do not roll back successful remote data because the local fallback could not be refreshed.
- **CON-009-AC15 — Local recovery state:** Keep images, fallback YAML and durable pending-logout records under the existing local data directory. Access will specify rejection of locally revoked sessions, later S3 removal and logout failure when its local record cannot be saved.
- **CON-009-AC16 — Operator upload:** Moving an existing local household into S3 is an operator-performed YAML upload. Do not silently import local YAML into a missing remote household or add a migration-upload command. Local files are not substituted for remote authority during setup.
- **CON-009-AC17 — State directory sync:** After atomically publishing private local state, open the parent root and obtain its directory-sync handle through root-relative access. Preserve error reporting for parent open, directory open and sync failures, private file permissions, atomic publication and rejection of non-regular destinations. This applies to local S3 fallback and recovery state, without changing remote operations.

## Related domain approvals

- Wishes: [WSH-008-AC1–AC12](wishes.md#wsh-008-background-image-recovery-and-s3-cache-cleanup) and WSH-006 amendments draft the confirmed background recovery, three concurrent restores, increasing retries, original filenames/no YAML changes, placeholders/page-load visibility and restart-safe S3 cleanup exception. These approved criteria and amendments are implemented and verified.
- Access: [ACC-008-AC1–AC10](access.md#acc-008-s3-cached-access-and-durable-logout) drafts the confirmed cached identities, blocked new logins, enforced read-only controls, durable logout/retry/reconciliation and persistence-failure behavior. These approved criteria are implemented and verified.
- Internationalization: HR113–HR120 English/German messages are approved and bundled in the runtime catalog.

## Implementation boundary

The user authorized implementation of all approved affected criteria and HR113–HR120 wording. Follow the resumable execution plan in implementation.md. No real-bucket writes/migration, live-process changes or deployment are included.
## Delivery evidence

The S3 repository, app storage selection/setup, read-only fallback controls and localized error/warning paths are complete. Repository/HTTP/race tests cover authority, target isolation, stale/manual edits, invalid data, rejected access, deleted buckets, actual disconnections, streamed-body timeouts, unknown PUT outcomes and cache-refresh failures. Eight approved messages are bundled; no remote translations are fetched.

All 22 browser scenarios pass. Disposable RustFS and Garage verify real path-style signed operations, including Unicode household keys and configured signing regions. Production scratch containers behind actual Traefik verify local/S3 flows, image recovery, persistence/replacement and a stopped S3 provider: cached access and logout survive app restarts, and service recovery restores authorized changes. Invalid environment settings exit nonzero; an unreachable source without fallback leaves the process running with HR115.

MinIO image pulls were unavailable; hosted R2 and actual VPS deployment were not tested. These external limits do not imply credentials, real bucket migration or deployment occurred. Operator setup/upload instructions are in [docs/s3-storage.md](../docs/s3-storage.md).
