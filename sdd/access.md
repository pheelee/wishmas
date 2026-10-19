# Access

## Status

Approved by the user, including the consolidated OPEN-review updates. Not implemented.

## Priority

Must-have for v1.

## ACC-001: First-run setup

On first run, provide an open setup wizard. Choose the household timezone during setup and automatically generate the initial parent and family access codes. The wizard remains open until setup completes.

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

## ACC-002: Access links and scope

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

No implementation or tests yet. After approval, write behavioral tests that fail first and link tests and code to these criteria using `@test` and `@impl` anchors. Do not mark this domain Implemented until every acceptance criterion is fully met.
