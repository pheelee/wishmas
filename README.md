# Wishmas

A family Christmas wishlist app, built from the approved requirements in [sdd/](sdd/README.md).

## Current checkpoint

The implemented slices provide household setup, persisted access links, and independent parent/family code regeneration. Calendar-year behavior uses the household timezone. People, wishlists, wishes, claims, and Docker packaging are subsequent milestones. See [implementation progress](sdd/progress.md) for the exact verification state and next action.

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

After setup, the parent page displays both current access links with copy and regeneration controls. Regenerating one link leaves the other unchanged. An old link immediately loses access, including the next action from an already-open page. Regenerating the parent link automatically opens its replacement. Family pages do not expose parent controls or access-link fields.

The calendar year is computed in the household timezone on each page load. Codes do not expire automatically. You can change the timezone by editing the YAML file; there is no post-setup timezone control in the app.

## December-night theme

All existing screens use the approved dark palette and steady Christmas lights. Snow consists of 16 faint CSS particles confined to the outer margins, behind the content. The header's **Pause snow** / **Resume snow** button remembers your choice in the browser. With reduced motion enabled, snow does not animate; the disabled resume button explains why in its title.

The [approved design reference](design/december-night/README.md) records the selected baseline. `npm run test:e2e` generates actual desktop and mobile implementation screenshots under `test-results/`. No animation library or external runtime asset is needed.

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

Each browser test starts an isolated Go process with temporary data on a dynamically allocated local port and cleans it up afterward. Coverage includes setup, real clipboard copying, independent link regeneration, automatic parent navigation, stale-page revocation, dark styling, snowfall controls, reduced motion, error pages, and narrow-screen layouts. Go tests cover setup persistence using a newly constructed server, invalid inputs, direct YAML reloads, independent persisted regeneration, concurrent stale-code requests, and deterministic year boundaries.

Dependency references: [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli) and [Go YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md), and [the progress checkpoint](sdd/progress.md). Inspect the Git working tree before continuing. Requirements are not Implemented until every acceptance criterion has been verified.
