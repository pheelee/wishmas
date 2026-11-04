# Offline household tools

`wishmas-admin` provides private password-hash generation for the current household format. Legacy migration has been retired after the completed data migration; `migrate` is no longer a supported command.

## Generate a password hash

From the repository root, in an interactive terminal:

```sh
go run ./cmd/wishmas-admin hash-password
```

Enter the password and confirmation privately. Neither prompt echoes input. The helper enforces at least eight Unicode characters, uppercase/lowercase letters, a digit and punctuation/symbol; spaces are permitted but do not count as special. It prints a YAML-ready `admin_password_hash` entry using Argon2id with a fresh random salt. Password arguments and non-terminal password input are refused.

The supported encoded hash parameters are Argon2id v19, 19 MiB, two iterations, one lane, 16-byte salt and 32-byte result. Unsupported/malformed costs fail validation before hashing. Do not put passwords/hashes in Git, chat or progress notes.

## Verification

```sh
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build -tags nodynamic ./...
```

The current HTTP runtime implements the household/list/user model. Tests compare the bundled catalog directly to all 223 approved current rows in `sdd/translations.md`; historical-only messages and the old test harness have been removed.

## Container helper

Generate a hash privately in an interactive terminal (the image defaults to UID/GID 1000:1000):

```sh
docker run --rm -it --entrypoint /wishmas-admin wishmas:local hash-password
```

Use Docker `--user UID:GID` or the reference Compose service's `WISHMAS_UID`/`WISHMAS_GID` overrides to run the same offline helper as another identity, without rebuilding. Runtime overrides do not change file ownership. See [the Compose reference](review-and-deployment.md#runtime-identity-and-compose-reference).
