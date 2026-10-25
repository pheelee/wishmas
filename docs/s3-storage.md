# Optional S3 YAML storage

Local storage remains the default. Set `WISHMAS_STORAGE=s3` to keep the household YAML in an existing S3-compatible bucket. Run one Wishmas instance per household. Pictures, fallback data and pending logout/cleanup records still need a persistent, writable `WISHMAS_DATA_DIR` volume.

## Configuration

The reference [Compose file](../docker-compose.yml) passes these environment variables to the container:

| Variable | Value |
| --- | --- |
| `WISHMAS_STORAGE` | `local` (default) or `s3` |
| `WISHMAS_S3_ENDPOINT` | Required HTTP/HTTPS S3 API endpoint |
| `WISHMAS_S3_BUCKET` | Required existing bucket |
| `WISHMAS_S3_REGION` | Signing region; default `us-east-1` |
| `WISHMAS_S3_ACCESS_KEY_ID` | Required access key |
| `WISHMAS_S3_SECRET_ACCESS_KEY` | Required secret key |
| `WISHMAS_S3_SESSION_TOKEN` | Optional session token |
| `WISHMAS_S3_TIMEOUT` | Per-operation timeout including response body; default `5s` |
| `WISHMAS_DATA_DIR` | Existing local data setting; default `data` (`/data` in container) |

Use the provider's API endpoint and configured signing region. Requests always use path-style addressing. There is no key/prefix setting, automatic region lookup, default credential chain or automatic bucket creation. Invalid environment configuration exits nonzero; an unavailable service leaves the app running with an error page or read-only fallback.

Each household object is `<household-name>/household.yaml` at the bucket root. Other objects are ignored. More than one household YAML object produces an error in this version. The app needs ListObjectsV2, GetObject and PutObject permissions; it does not need bucket creation, DeleteObject, ACLs, versioning or conditional writes. Serialization and stale revision checks match the local single-instance contract; they do not coordinate competing external writers.

## Setup and moving an existing household

For a new household, create the bucket outside Wishmas, configure the environment and complete the usual setup wizard. Setup creates the household object and signs in automatically.

To move an existing household, stop writes, back up its version-2 YAML and upload it yourself to `<household-name>/household.yaml`, where the name matches the YAML `name`. Keep its schema, IDs, codes, sessions and picture references. Then select S3 and restart Wishmas. The app does not upload/import local household YAML automatically. No migration-upload command is provided.

S3 is authoritative. A confirmed missing object opens setup, even when an older fallback exists. A permission rejection, invalid YAML or multiple households produces a configuration error; those conditions do not use the fallback. Do not remove the authoritative object as an outage workaround.

## Outages, saves and logout

Valid remote data is retained under `WISHMAS_DATA_DIR/s3/<target-digest>/fallback.json`; its YAML bytes are encoded in this private local record. Endpoint and bucket determine the target namespace. Household pictures and pending records live under that namespace's `households/<household-name>/` directory. Local-mode `households/` files remain separate and untouched.

During a temporary outage, existing sessions and user links can read the retained household after a restart. Management/claim controls are disabled with an explanation; direct mutation requests are rejected. New password sign-ins are blocked. Navigation, copying links, browser sorting and viewing authorized cached pictures remain available. A successful remote page load restores normal access automatically.

Logout first saves a durable hash-only revocation locally, then clears the current cookie. It continues rejecting that session across restarts until its remote removal and fallback refresh are confirmed. Other sessions remain active. If this local record cannot be saved, logout reports an error and keeps the cookie/session. Keep the local volume: it carries these pending revocations as well as images and offline data.

Ordinary edits/claims are never queued or automatically replayed. A timeout/5xx after PUT can have an unknown outcome; the error asks you to reload before retrying. A confirmed remote save remains successful if the fallback refresh fails, with a warning when the offline copy is still stale.

## Local pictures and recovery

At startup, both backends schedule missing configured pictures in the background. Existing files are retained. Recovery uses the stored URL and existing public-address/redirect, 30-second, 10-MiB, 20-megapixel and image-format restrictions; output remains maximum-640px, alpha-preserving quality-80 WebP under the existing filename. Pages remain usable with Image unavailable placeholders. No YAML fields change, and restored pictures appear on the next page load without polling.

At most three restores run simultaneously. Failed restores retry after 1, 2, 4, 8, 16 and then 30 minutes, capped at 30 minutes. Changed/removed references are checked again before publication; recovery never replaces a file that has appeared meanwhile.

In S3 mode, old-picture deletion failure cannot undo a confirmed remote YAML save. Cleanup intentions persist locally and retry with the same backoff after restart, checking current remote references before deleting. Local mode retains its existing picture-change rollback. Newly downloaded files are retained when a remote save is uncertain, since that save may reference them.

Pending logout removals use the same retry delays and also retry immediately on a successful remote page load. Browser clients do not poll. Recovery of a startup household that is initially unavailable waits for a valid snapshot; no invalid household is repaired or silently substituted.

## Verification

Protocol tests exercise authority, stale revisions, persisted fallback, permission/configuration errors, timeout body reads, uncertain saves, cache failures and durable retries with disposable fixtures. Opt-in `TestS3CompatibleProvider` creates/deletes a unique bucket on an explicitly configured disposable provider. Set `WISHMAS_TEST_S3_BUCKET` to use an already provisioned, empty disposable bucket instead; the test removes its household object afterwards. Production never performs bucket-management calls.

```sh
WISHMAS_TEST_S3_ENDPOINT=http://your-disposable-provider:9000 \
WISHMAS_TEST_S3_ACCESS_KEY_ID=your-test-key \
WISHMAS_TEST_S3_SECRET_ACCESS_KEY=your-test-secret \
go test ./internal/household -run '^TestS3CompatibleProvider$' -count=1
```

Use only a disposable provider with test credentials for that integration command. Actual provider and container results are recorded in `sdd/progress.md`; hosted R2 credentials and a VPS deployment are not part of local verification.

The existing `tests/container/smoke.cjs` also accepts `WISHMAS_SMOKE_S3_ENDPOINT`, `WISHMAS_SMOKE_S3_BUCKET`, `WISHMAS_SMOKE_S3_REGION`, `WISHMAS_SMOKE_S3_ACCESS_KEY_ID` and `WISHMAS_SMOKE_S3_SECRET_ACCESS_KEY` for a separately provisioned empty disposable bucket. Optional `WISHMAS_SMOKE_S3_CONTAINER` connects an existing disposable provider container to the test network, then disconnects it. `WISHMAS_SMOKE_RECOVER_IMAGE=1` removes the test image and verifies startup recovery after restart in either storage mode. These tests use a new data volume, never the repository's household data.

With a disposable provider container, `WISHMAS_SMOKE_TEST_OUTAGE=1` additionally stops that provider, checks cached reads and disabled claims through app restarts, logs out during the outage, verifies durable denial and then restarts the provider to check recovery. The test restarts its provider during cleanup as well.
