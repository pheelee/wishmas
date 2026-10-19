# Wishmas

A family Christmas wishlist app, built from the approved requirements in [sdd/](sdd/README.md).

## Current checkpoint

The first implementation slice provides the household setup wizard and persisted access links. People, wishlists, wishes, claims, code regeneration, and Docker packaging are subsequent milestones. See [implementation progress](sdd/progress.md) for the exact verification state and next action.

## Run locally

With the Go version declared in `go.mod`:

```sh
go run ./cmd/wishmas
```

Open `http://localhost:8080`. Before setup is complete, the wizard is open. Enter a household name and timezone, copy the generated parent and family links, and select **Finish setup**.

- `WISHMAS_ADDR` controls the listening address; default `:8080`.
- `WISHMAS_DATA_DIR` controls storage; default `data`.
- Setup is saved in `data/households/<household-name>/household.yaml` using restricted file permissions. Keep this directory across restarts.
- The metadata file contains `name`, `timezone`, `parent_code`, and `family_code`. Manual edits are read on the next page load. Invalid files show an error and are left untouched.
- The upcoming list storage will add year/event/person levels under the household. This slice saves household metadata only.
- Setup previews are held in memory until **Finish setup** persists them. If the process restarts before finishing, reopen setup to generate new links.

## Build and check

CSS is compiled locally and embedded in the Go application. Rebuild it after changing Tailwind classes; Node.js is a build/test dependency rather than a runtime dependency.

```sh
npm ci
npm run build:css
go test -race ./...
go vet ./...
npx playwright install --with-deps chromium
npm run test:e2e
go build -o wishmas ./cmd/wishmas
```

The browser test starts an isolated server with temporary data, verifies the setup flow and real clipboard copying, then stops the server. It requires port 18080 to be available. Go tests cover setup persistence using a newly constructed server, invalid inputs, and direct YAML reloads.

Dependency references: [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli) and [Go YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md), and [the progress checkpoint](sdd/progress.md). Inspect the Git working tree before continuing. Requirements are not Implemented until every acceptance criterion has been verified.
