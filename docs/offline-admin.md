# Offline household tools

Milestone 1 provides the version-2 data package and `wishmas-admin` command. The HTTP app still uses the legacy schema; login/list/wish UI integration follows in later approved milestones. Never point that legacy app at migrated data.

## Generate a password hash

From the repository root, in an interactive terminal:

```sh
go run ./cmd/wishmas-admin hash-password
```

Enter the password and confirmation privately. Neither prompt echoes input. The helper enforces at least eight Unicode characters, uppercase/lowercase letters, a digit and punctuation/symbol; spaces are permitted but do not count as special. It prints a YAML-ready `admin_password_hash` entry using Argon2id with a fresh random salt. Password arguments and non-terminal password input are refused.

The supported encoded hash parameters are Argon2id v19, 19 MiB, two iterations, one lane, 16-byte salt and 32-byte result. Unsupported/malformed costs fail validation before hashing. Do not put passwords/hashes in Git, chat or progress notes.

## Migrate a copied legacy household

The command operates on one household directory, not the whole `data/` root. First make a private, consistent copy of legacy data, retaining the original. Add the generated `admin_password_hash` to the copy's existing `household.yaml`; do not edit the original. The source should contain `household.yaml`, optional `people.yaml`, and the old year/christmas/person wishlists and referenced pictures.

Create only the destination's parent directory; the destination itself must not exist, even as an empty directory. Keep it outside the copied source. Example paths below are placeholders:

```sh
go run ./cmd/wishmas-admin migrate \
  --source /path/to/legacy-copy/households/Family \
  --destination /path/to/review/households/Family
```

Successful output is a JSON comparison report: source/destination list/recipient/wish/claim counts, resulting user count, copied-image SHA-256 checksums, missing-image count, source-unchanged and destination-verified results. It contains no password hash, session token or link code.

The migrator:

- Creates named lists in ascending year order and preserves every wish occurrence, added order within the source lists, details, compatible future wish/claim fields and claims. Archived recipients become ordinary visible recipients.
- Merges claimant names only by case/surrounding spaces, keeps the first spelling and distinguishes accents; creates stable user IDs and independent links for every user/list pair.
- Copies referenced picture bytes without downloading/re-encoding, with unique destination names; configured missing images stay missing and retain their source URLs.
- Stages and validates the output, reloads it, compares counts/image bytes and verifies unchanged source contents before publishing the destination. Failed validation/processing removes staging/output created by that attempt; existing destinations are never overwritten.
- Refuses symlinks/non-regular source files, unknown/unrepresentable legacy files, conflicting new-schema fields, duplicate/invalid names or claims, unsupported hashes and changing source snapshots. Refusal preserves source data rather than silently discarding it. Keep the error/report for review without sharing credentials.

New storage is one private `household.yaml` plus household-level `pictures/`; see [the schema](../sdd/storage.md). The tool creates no admin sessions, starts no server and does not merge, cut over or modify the running app.

## Scope and rollback

The current household's real migration is milestone 6, after the new app is ready and the user supplies a hash in its private copy. Milestone 1 has used disposable fixtures only. Packaging the helper into the final container and complete Traefik/HTTPS verification remain later work.

Keep baseline code `1992bdf` with legacy data for rollback. Review uses separate code/data paths. Do not run old and new app versions against the same directory; do not replace original data or merge until the explicit final decision. If live data changes during review, cutover requires a fresh copy and migration.

## Verification

```sh
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build -tags nodynamic ./...
```

The new HTTP flows are not implemented by this milestone, so these checks do not claim new browser behavior. Existing legacy wording checks use the fixed baseline test fixture until the UI/catalog is migrated; the approved target catalog is in `sdd/translations.md`.
