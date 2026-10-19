# Wishmas

A family Christmas wishlist app, built from the approved requirements in [sdd/](sdd/README.md).

## Current checkpoint

Implemented slices provide setup, access-link management, person creation, and current-year Christmas lists. Parents add adults or children; both access roles can browse their lists. Calendar-year behavior uses the household timezone. App renaming/archiving, history navigation, wish management, claims, and Docker packaging remain subsequent work. See [implementation progress](sdd/progress.md) for the exact verification state and next action.

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
- The roster lives in `people.yaml` beside household metadata. Each list lives at `<year>/christmas/<person-name>/wishlist.yaml` under the household directory.
- Setup previews are held in memory until **Finish setup** persists them. If the process restarts before finishing, reopen setup to generate new links.

After setup, the parent page displays both current access links with copy and regeneration controls. Regenerating one link leaves the other unchanged. An old link immediately loses access, including the next action from an already-open page. Regenerating the parent link automatically opens its replacement. Family pages do not expose parent controls or access-link fields.

The calendar year is computed in the household timezone on each page load. Codes do not expire automatically. You can change the timezone by editing the YAML file; there is no post-setup timezone control in the app.

## People and current-year lists

Use the parent page's **Person name** field and **Add person** button. Names are trimmed, require 3–50 Unicode letters/digits, and must be unique ignoring capitalization across both active and archived people. Lists appear in the order people were added. New people receive empty current-year lists automatically.

The roster format is:

```yaml
people:
  - name: Anna
  - name: Bobby
    archived: true
```

An empty list is:

```yaml
wishes: []
```

You can manually enter a title-only wish to view it:

```yaml
wishes:
  - title: Train set
```

Direct edits reload on page load. Malformed files or duplicate names show an error without changing data. Existing manual list files are not overwritten by person creation. Archived names stay reserved and archived people do not appear or receive new-year lists. Missing current-year lists for active people are created on a valid page load, including after a year change. Old files are retained.

This slice displays wish titles from YAML. Web wish editing and remaining wish fields are not implemented yet. App rename/archive controls and historical navigation are also future work; neither role can currently open historical lists through the UI. The family link is restricted to the current year.

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

Each browser test starts an isolated Go process with temporary data on a dynamically allocated local port and cleans it up afterward. Coverage includes setup, real clipboard copying, independent link regeneration, automatic parent navigation, stale-page revocation, dark styling, snowfall controls, reduced motion, error pages, and narrow-screen layouts. Go tests also cover people/list persistence, manual YAML reloads and errors, archived name reservations, concurrent duplicate creation, Unicode boundaries, exclusive protection of existing manual lists, and deterministic new-year creation. Browser tests include the complete parent creation and family viewing flow with desktop/mobile checks.

Dependency references: [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli) and [Go YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md), and [the progress checkpoint](sdd/progress.md). Inspect the Git working tree before continuing. Requirements are not Implemented until every acceptance criterion has been verified.
