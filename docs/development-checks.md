# Development checks

Use the Go version declared in `go.mod`. Install **golangci-lint v2.15.0** using
the [official installation instructions](https://golangci-lint.run/docs/welcome/install/)
and its matching release checksum. Use the official release binary rather than
building golangci-lint with `go install`. Put it on your `PATH` and verify with
`golangci-lint version`.

Install the separate vulnerability scanner:

```sh
go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
```

Put `$(go env GOPATH)/bin` on `PATH` for `govulncheck`. Tool versions must match
the pins in `.github/workflows/ci.yml`; update this document and contributor
instructions when changing those pins.

From the repository root, **before every local commit**:

```sh
golangci-lint config verify
golangci-lint run ./...
```

These checks are required by `AGENTS.md`; get the user's review before committing.
This is an instruction-based requirement, not an installed Git hook.

For the complete Go/automation validation used by CI:

```sh
scripts/check-go.sh
golangci-lint config verify
golangci-lint run ./...
govulncheck ./...
python3 -B -m unittest discover -s tests/release -v
```

The lint config explicitly enables `gosec`, `govet`, `ineffassign`,
`staticcheck`, `errorlint`, `bodyclose` and `nilerr` for production code. Existing test code is still vetted and race-tested by
`scripts/check-go.sh`. New linters require review rather than enabling an
unreviewed default set. Keep suppressions narrow and explain the trust boundary.

`govulncheck` queries the Go vulnerability database and fails on reachable
vulnerabilities. Module-only advisories for unused packages do not fail its
default symbol-level check. Both tools run in the existing **Go tests** job for
PRs targeting `main` and pushes to `main`, so either tool's failure fails that
status check. Branch protection must continue requiring that check.
