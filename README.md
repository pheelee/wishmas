# Wishmas

A family Christmas wishlist app, built from the approved requirements in [sdd/](sdd/README.md).

## Current checkpoint

Implemented slices provide setup, access-link management, people creation/renaming/archiving, and Christmas lists with parent-only history. Both roles browse current-year lists; parents also browse saved years. Calendar-year behavior uses the household timezone. Web wish management, claims, and Docker packaging remain subsequent work. See [implementation progress](sdd/progress.md) for the exact verification state and next action.

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

Open a person's current-year list with the parent link to **Rename person** or **Archive person**. Renaming preserves list and claim contents across years and retains added order. Archiving hides all their lists from both roles, keeps their data and name reservation, and stops new-year list creation.

There is no restore action in the app. To restore, remove `archived: true` (or set it to `false`) in `people.yaml`, then reload. A missing current-year list is created empty; an existing one is preserved.

Parents use **Christmas year** and **View year** to browse saved lists. Historical pages show **Read-only history** and omit person/wish editing controls. Family links cannot open past-year pages. Manual edits to historical YAML still reload as agreed.

Parents can **Add wish** or expand **Edit wish** on current-year lists. Only the title is required (5–150 characters after trimming). Shop links, one reference price, and a note are optional and can be added or cleared later. Enter complete HTTP/HTTPS shop URLs, one per line; visitors see domain labels in the saved order. Prices use a decimal point and at most two decimal places.

Edits preserve attached claims and other saved fields. Stale edit forms are rejected with a reload message. Family links display details without editing controls; historical and archived lists cannot be changed through these routes. Wish deletion, sorting, and claim controls remain to be implemented.

Example `wishlist.yaml`:

```yaml
wishes:
  - title: Wooden train
    links:
      - https://shop.example/train
    price: "12.50"
    note: For a snowy afternoon.
```

The sequence retains added order. Optional fields may be omitted. YAML edits reload on page load; invalid titles, prices, or shop links show an error without rewriting files.

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

Each browser test starts an isolated Go process with temporary data on a dynamically allocated local port and cleans it up afterward. Coverage includes setup, real clipboard copying, independent link regeneration, automatic parent navigation, stale-page revocation, dark styling, snowfall controls, reduced motion, error pages, and narrow-screen layouts. Go tests also cover people/list persistence, manual YAML reloads and errors, archived name reservations, concurrent duplicate creation, Unicode boundaries, exclusive protection of existing manual lists, and deterministic new-year creation. Browser tests include parent creation, family viewing, history navigation, manual history reload, renaming, archiving, and manual restoration with mobile/desktop checks. Go tests cover byte-preserving moves across years, conflicts, and rollback on directory or roster-save failures.

Dependency references: [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli) and [Go YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md), and [the progress checkpoint](sdd/progress.md). Inspect the Git working tree before continuing. Requirements are not Implemented until every acceptance criterion has been verified.
