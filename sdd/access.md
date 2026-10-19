# Access

## Status

Approved by the user, including the consolidated OPEN-review updates and household-name decisions. In progress. ACC-001 is Implemented; ACC-002 and ACC-003 retain pending list-scope integration verification.

## Priority

Must-have for v1.

## ACC-001: First-run setup

Status: Implemented. All ACC-001 acceptance criteria are verified.

On first run, provide an open setup wizard. Choose the household timezone during setup and automatically generate the initial parent and family access codes. The wizard remains open until setup completes.

The setup wizard asks for a household name. Trim surrounding spaces; require 3–50 characters consisting only of Unicode letters and digits.

After setup, changing the household timezone requires a manual YAML edit; the app provides no timezone-editing control.

The wizard's final screen shows both generated access links with copy buttons. A “Finish setup” button completes setup, closes the wizard, and opens the parent link.

### Acceptance criteria

- **ACC-001-AC1:** Given setup has not completed, a visitor without an access code can open the setup wizard.
- **ACC-001-AC2:** During setup, the household timezone can be chosen.
- **ACC-001-AC3:** Setup automatically generates a parent access code and a separate family access code.
- **ACC-001-AC4:** After setup completes, a visitor without a valid code sees a not-found page rather than the setup wizard.
- **ACC-001-AC5:** After setup, the app provides no timezone-editing control. A manual YAML change to the household timezone is picked up on page load and determines the calendar year used for access.
- **ACC-001-AC6:** The wizard's final screen displays the generated parent and family links, each with a button that copies that link.
- **ACC-001-AC7:** Selecting “Finish setup” completes setup, closes the wizard, and opens the parent link with parent access.
- **ACC-001-AC8:** The setup wizard asks the user to enter a household name.
- **ACC-001-AC9:** Household-name entry trims surrounding spaces and rejects blank names, names outside 3–50 characters after trimming, and characters other than Unicode letters and digits.

## ACC-002: Access links and scope

Status: In progress; actual list and year scopes await milestone 3.

Access requires no account. Parent and family links contain their respective codes and open directly without a separate code-entry step. Parents have access to all people and all years. Family members have access to all people in the current calendar year, determined by the household timezone.

Access codes remain valid until parents regenerate them; they do not expire automatically.

### Acceptance criteria

- **ACC-002-AC1:** Opening a valid parent or family link grants its corresponding access without account registration, login, or separate code entry.
- **ACC-002-AC2:** A parent link grants access to all people and all years, including retained previous years' lists.
- **ACC-002-AC3:** A family link grants access to all people in the current year and does not grant access to other years.
- **ACC-002-AC4:** At January 1 in the household timezone, family-link access follows the new calendar year automatically.
- **ACC-002-AC5:** After setup completes, access without a code or with an invalid code shows a not-found page.
- **ACC-002-AC6:** An access code that has not been regenerated remains valid over time, including across calendar-year changes, without automatic expiry.

## ACC-003: Code regeneration

Status: In progress; regeneration behavior is verified, but replacement-link list/year scope in ACC-003-AC2 awaits milestone 3.

Only parents can independently regenerate either access code at any time and share its replacement link. The old code loses access immediately. Regenerating the parent code automatically switches the current page to the replacement parent link.

### Acceptance criteria

- **ACC-003-AC1:** Parents can regenerate the family code without replacing the parent code, and can regenerate the parent code without replacing the family code.
- **ACC-003-AC2:** A replacement link grants the same role and scope as the link it replaces and is available for parents to share.
- **ACC-003-AC3:** After regeneration, opening the replaced link shows a not-found page.
- **ACC-003-AC4:** The next action from an already-open page using the replaced code loses access without requiring a page reload.
- **ACC-003-AC5:** After regenerating the parent code, the current page automatically switches to the replacement parent link and retains parent access through that link.
- **ACC-003-AC6:** Family-link holders cannot regenerate either access code.

## Applicable constraints

- CON-001: Docker container self-hosted on a VPS.
- CON-002: Go backend.
- CON-003: Christmas-themed frontend using Tailwind CSS and vanilla JavaScript.
- CON-004: YAML persistence and the confirmed household/year/event/person hierarchy.

## OPEN questions

None currently listed.

## Implementation and verification

The setup/persistence and regeneration/calendar slices are verified. Their behavioral tests failed before implementation: the initial placeholder returned 404 for setup, and regeneration tests returned 404 before its endpoint existed. A deterministic clock test initially failed at the household New Year boundary before the application used its injected clock.

| Requirement / criteria | Evidence | State |
| --- | --- | --- |
| ACC-001-AC1–AC4, AC6–AC9 | `internal/app/app_test.go`; `tests/browser/setup.spec.js`; actual clipboard copying and household-name validation | Verified |
| ACC-001-AC5 | `TestCalendarYearUsesHouseholdTimezoneAndCodesDoNotExpire` plus direct YAML reload and absence of timezone-editing controls | Verified; includes household New Year and manual timezone change affecting the computed year |
| ACC-002-AC1, AC5 | Generated link entry and missing/invalid-code 404 tests | Verified |
| ACC-002-AC4, AC6 | Deterministic boundary and ten-year clock advance; original codes continue to work without file changes | Calendar/code-lifetime behavior verified; actual list selection at the boundary awaits milestone 3 |
| ACC-002-AC2/AC3 | People and historical/current-year list scopes | Pending milestone 3 |
| ACC-003-AC1, AC3–AC6 | `internal/app/access_test.go` and browser regeneration/stale-page scenario; includes persistence and concurrent parent regeneration | Verified |
| ACC-003-AC2 | Replacement links are copyable and retain parent/family roles; the other role's link is unchanged | Role and sharing verified; list/year scopes await milestone 3 |

Code and tests include `@impl`/`@test` anchors. Every action reloads YAML and authenticates against the saved code. Independent regeneration is saved before redirecting; an old-code action from an already-open page is rejected immediately. The request mutex ensures that simultaneous parent rotations cannot both authenticate using the replaced code.

Verification commands: `go test -race ./...`, `go vet ./...`, `go build -o /tmp/wishmas-check ./cmd/wishmas`, `npm run build:css`, and `npm run test:e2e`. Browser tests exercised copying the replacement link, automatic parent navigation, independent family access, and a stale form submitted without reloading. Mobile access-control layout was visually reviewed. See `progress.md` for the checkpoint and exact next action.
