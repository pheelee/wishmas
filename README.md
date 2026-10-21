# Wishmas

A family Christmas wishlist app, built from the approved requirements in [sdd/](sdd/README.md).

## Current checkpoint

Implemented slices provide setup, access-link management, people creation/renaming/archiving, and Christmas lists with parent-only history. Both roles browse current-year lists; parents also browse saved years. Calendar-year behavior uses the household timezone. Wish creation, editing, deletion, sorting, and all approved claim behavior are implemented. Wish cards use the approved Gift-wrap styling: colored gift headers, status badges, optional price and shop details, and inline action footers. Card accents are red for incomplete wishes, yellow for complete unclaimed wishes, and green for complete claimed wishes. Completeness means a title, a provided price (including zero), and at least one shop link; title remains the only required field for saving a wish. Docker packaging and local container/complete-flow verification are complete; actual VPS deployment has not been performed. See [implementation progress](sdd/progress.md) for the exact verification state and next action.

Each person’s overview shows the total reference price of all supplied prices, including incomplete wishes, plus claimed and unclaimed badges counting complete wishes only. Both roles see these summaries for the displayed year; totals use two decimals without currency.

## Run with Docker

<!-- @impl CON-001-AC4 CON-004 -->
Build from the repository root. The image uses the Go version in `go.mod`, embeds the checked-in compiled CSS/JavaScript and timezone database, and runs as UID/GID `65532:65532`. Node.js is not needed to build or run the image. Rebuild CSS first if frontend source changed.

```sh
docker build -t wishmas:local .
docker volume create wishmas-data
docker run -d --name wishmas --restart unless-stopped \
  -p 8080:8080 \
  --mount type=volume,src=wishmas-data,dst=/data \
  wishmas:local
```

Open `http://<vps-address>:8080` and complete setup. `/data` contains the existing household/year/event/person YAML hierarchy. Reuse the same named volume when replacing the container; keeping only the image does not preserve household data. The Docker build context excludes local household data, notes and design artifacts.

`WISHMAS_ADDR` defaults to `:8080` and `WISHMAS_DATA_DIR` to `/data` inside the image. Publish a different host port with, for example, `-p 8090:8080`. An existing reverse proxy may forward to the published app port; its configuration is outside this repository.

For a new image, rebuild and replace the container while retaining its volume:

```sh
docker build -t wishmas:local .
docker stop wishmas
docker rm wishmas
docker run -d --name wishmas --restart unless-stopped \
  -p 8080:8080 \
  --mount type=volume,src=wishmas-data,dst=/data \
  wishmas:local
```

For direct YAML editing on the VPS, a bind mount can replace the named-volume mount: `--mount type=bind,src=/absolute/path/to/wishmas-data,dst=/data`. Create that directory first and give UID/GID `65532:65532` write access (including any existing YAML). The app reloads manual edits on page load as described below. Keep one running app instance per data directory.

To copy a consistent backup, stop the container, run `docker cp wishmas:/data ./wishmas-backup`, then `docker start wishmas`. Retain the complete hierarchy, including household metadata containing the access codes.

Run the isolated container smoke check after installing the browser-test dependencies:

```sh
npm ci
npx playwright install --with-deps chromium
node tests/container/smoke.cjs
WISHMAS_SMOKE_LANGUAGE=de-CH node tests/container/smoke.cjs
```

It builds a disposable image, uses a temporary named volume and dynamically allocated port (loopback with a local Docker daemon, the daemon host with a remote Docker context), runs real setup/parent/family/claim flows, and verifies persisted YAML and pictures after restart and container replacement, downloads the provided public sample image over HTTPS, and checks authenticated image serving and removal. The picture smoke check requires public internet access to asset.conrad.com. It removes its container, volume and image afterward. No VPS is contacted and existing application data is not used.

## Run locally

With the Go version declared in `go.mod`:

```sh
go run ./cmd/wishmas
```

Open `http://localhost:8080`. Before setup is complete, the wizard is open. Enter a household name, review the timezone prefilled from your browser, and adjust the appearance sliders with their live preview. Copy the generated parent and family links, then select **Finish setup**.

- `WISHMAS_ADDR` controls the listening address; default `:8080`.
- `WISHMAS_DATA_DIR` controls storage; default `data`.
- Setup is saved in `data/households/<household-name>/household.yaml` using restricted file permissions. Keep this directory across restarts.
- The metadata file contains `name`, `timezone`, `language`, `parent_code`, and `family_code`. Manual edits are read on the next page load. Invalid files show an error and are left untouched.
- The roster lives in `people.yaml` beside household metadata. Each list lives at `<year>/christmas/<person-name>/wishlist.yaml` under the household directory.
- Setup previews are held in memory until **Finish setup** persists them. If the process restarts before finishing, reopen setup to generate new links.

After setup, the parent page displays both current access links with inline clipboard and regeneration icon buttons. The clipboard icon shows a checkmark for three seconds after a successful copy. Regeneration opens a modal before replacing the selected link. Regenerating one link leaves the other unchanged. An old link immediately loses access, including the next action from an already-open page. Regenerating the parent link automatically opens its replacement. Family pages do not expose parent controls or access-link fields.

The calendar year is computed in the household timezone on each page load. Codes do not expire automatically. You can change the timezone by editing the YAML file; there is no post-setup timezone control in the app.

## Household language

Wishmas supports American English (`en-US`) and Swiss Standard German (`de-CH`, informal address). On first setup, the browser's first supported language preference selects the initial wizard language and dropdown value; otherwise English is used. You may choose a different household language. The wizard stays in its initial language through the access-link screen; **Finish setup** saves your selection and opens the parent page in that language.

After setup, change language only by editing `household.yaml`:

```yaml
language: de-CH
```

The only accepted values are exactly `de-CH` and `en-US`. An absent field in an existing household means English. An empty, null or unsupported value shows an English configuration error without changing files. Everyone sees a valid language change on the next page load or form submission; open pages keep their current language until then. There is no browser-specific override or post-setup app setting.

The interface, dialogs, validation, error messages, clipboard feedback and accessibility text are translated. Invalid/missing/revoked links and unreadable household configuration use English; errors for a known valid household use its language. Wishmas and your entered names, titles, notes, links, saved claims and storage paths stay unchanged. Translations are bundled with the binary and need no external service.

Prices display two fractional digits with regional grouping (`1’234.50` / `1,234.50`). Enter/edit ungrouped decimals such as `1234.5` in either language; YAML and numeric sorting use the original value. Other numbers, including counts and years, never use grouping. Title sorting uses the household locale, retains the existing case/accent-insensitive stable ordering rules, and keeps your saved browser sort choice.

The [approved translation table](sdd/translations.md) is checked against the bundled catalog. Completeness checks fail for missing German/English entries or untranslated template/client messages. Missing German text falls back to English at runtime.

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

Parents can **Add wish** inline or use a wish card’s edit icon to open **Edit wish** in a modal on current-year lists. Only the title is required (5–150 characters after trimming). Shop links, one reference price, and a note are optional and can be added or cleared later. Enter complete HTTP/HTTPS shop URLs, one per line; visitors see domain labels in the saved order. Prices use a decimal point and at most two decimal places.

Edits preserve attached claims and other saved fields. Stale edit forms are rejected with a reload message. Family links display details without editing controls; historical and archived lists cannot be changed through these routes. To delete a current-year wish, choose the card’s delete icon, review the modal confirmation, and select **Permanently delete wish**. **Cancel** keeps everything unchanged. Deletion permanently removes the wish and its attached claims while retaining other wishes. Any list change invalidates open editing/deletion forms; reload before trying again.

**Sort wishes** defaults to added order, oldest first. Choose newest first, title A–Z, price low–high, or price high–low. Equal titles/prices retain added order; title sorting ignores capitalization and accents, and missing prices count as zero. One browser preference applies across people, years, and both access-link roles. Sorting changes only the display and never rewrites YAML. If browser storage is unavailable, sorting still works for the open page.

Select **Claim wish** on a current-year card. The first attempt asks for your name; **Save name and claim** remembers it in this browser and claims the wish. Later toggles work directly without confirmation. Both parents and family members may claim, and different names may claim the same wish. Each matching name can claim a wish only once.

Your current name appears beside **Change claim name**, which opens a modal. Renaming changes future claims only; old claims retain their original names. Matching names in different browsers represent the same claimant. Case and surrounding spaces are ignored, while accents remain significant. The browser name uses a persistent HttpOnly cookie, with no accounts.

Everyone sees claim status and counts. Family members see **Claimed by you** for their current name and can **Remove my claim**. Only parents receive the list of claimant names, with a direct **Remove** action for any current-year claim. Renaming does not carry removal rights for old-name claims. Parent history shows saved claims without claim actions.

Other viewers see changes on their next page load; already-open pages do not poll. A stale or failed action shows an inline error and refreshed status without altering existing claims. Browsing never requires a name.

A saved wish may include `claims: [{name: Auntie}, {name: Uncle}]`. Claimant names are trimmed for presentation, require 3–50 Unicode letters/digits, and must be unique per wish ignoring case and surrounding spaces. Invalid claims show a data error without changing files. Editing wish details preserves the original claim mappings and future fields.

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

All screens retain the approved dark palette and steady Christmas lights. Above the `md` breakpoint (768px), page content uses 75% of the viewport. Wishes appear in responsive cards with edit/delete icons together at the bottom. Edit and action forms use native modals with Cancel/Escape dismissal; creation forms remain inline.

Setup previews three household-wide settings saved in `household.yaml`:

- Bulb size: 80–200% of the original baseline; default 120%. Bulb count decreases inversely as size increases.
- Snow size: 50–200%; default 100%.
- Snow amount: 8–48 particles; default 16.

Snow falls in the outer margins and in the lightbulb band, where it is clipped and fades before reaching page content. Header and side snow follow the same household size/amount settings; each region has the chosen particle count. Snow stays behind content and has no pause control. Reduced motion keeps the snow visible and static. Old households without appearance settings receive the defaults; invalid manually entered settings show an error without rewriting files. Regenerating access links preserves appearance settings.

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

Internationalization tests also cover setup preference/fallback, delayed language application, localized native validation, strict/legacy YAML handling, authenticated/unauthenticated error languages, exact price precision, unchanged user data, and language changes with preserved sorting/claims.

Each browser test starts an isolated Go process with temporary data on a dynamically allocated local port and cleans it up afterward. Coverage includes setup, real clipboard copying, independent link regeneration, automatic parent navigation, stale-page revocation, dark styling, live appearance previews, always-present snowfall, reduced motion, error pages, and narrow-screen layouts. Go tests also cover people/list persistence, manual YAML reloads and errors, archived name reservations, concurrent duplicate creation, Unicode boundaries, exclusive protection of existing manual lists, and deterministic new-year creation. Browser tests include parent creation, family viewing, history navigation, manual history reload, renaming, archiving, and manual restoration with mobile/desktop checks. Go tests cover byte-preserving moves across years, conflicts, and rollback on directory or roster-save failures.

Dependency references: [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli) and [Go YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3).

## Resume implementation

Read [AGENTS.md](AGENTS.md), [the milestone plan](sdd/implementation.md), and [the progress checkpoint](sdd/progress.md). Inspect the Git working tree before continuing. Requirements are not Implemented until every acceptance criterion has been verified.

## Wish pictures

Parents can paste one optional direct HTTP(S) image URL when creating or editing a current-year wish. Saving downloads the image once, converts it to a still WebP at quality 80, and proportionally shrinks it to at most 640 pixels per side without upscaling. JPEG, PNG, WebP and GIF are supported; animated sources use their first frame and transparency is preserved. Sources must be publicly reachable, within 10 MiB and 20 megapixels, with at most five redirects and a 30-second total download limit. The container includes certificate roots for HTTPS downloads and builds with `-tags nodynamic` to use the Go WebP backend without a system image library. Use the same tag when building a standalone static executable.

Pictures appear in a centered, cropped shelf below the title/note; click to see the whole optimized image. No picture leaves the card unchanged. Missing or unreadable local pictures show “Image unavailable” without contacting the source again. Image access follows the existing link, year and archive permissions.

The wish YAML stores `picture_source` (the supplied URL) and `picture` (the optimized filename). Files live in `pictures/` beside that person's `wishlist.yaml`. Include these folders in backups and preserve the container user's write permissions. Viewing or saving an unchanged URL uses the local copy. Clearing the URL removes the picture; replacing a picture or deleting its wish removes the previous file. Failed downloads preserve the saved wish and entered form values. Failed removal of an old file cancels the action and restores the previous YAML.
