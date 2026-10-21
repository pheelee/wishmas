# Internationalization

## Status and priority

Household refactor: **behavioral acceptance criteria approved by the user**. Must-have. Milestone 2 verifies new setup/login/authentication wording and locale behavior; complete list/wish/claim integration remains pending.

Retain approved English/German behavior and unchanged approved wording. INT-001 and INT-005 remain applicable with updated screen terminology. INT-002–INT-004 are superseded by INT-006–INT-007 to remove timezone/access-link setup, code-dependent English login, year/archive screens and old storage assumptions. Git history retains their prior specification/evidence. Behavioral approval and new translation wording approval are separate.

## INT-001: Supported languages and translation scope

Support Swiss Standard German (`de-CH`) with informal `du/dein` address and American English (`en-US`). Language is household-wide. The app name Wishmas remains unchanged.

### Acceptance criteria

- **INT-001-AC1:** The app supports `de-CH` and `en-US` only. German uses Swiss spelling and informal address; English uses US English.
- **INT-001-AC2:** All app-provided browser text follows the active language: setup, admin login, household/list/recipient views, household-user management, fields/help, dialogs, buttons, empty states, badges, counts, validation/server errors, clipboard feedback, accessibility labels, tooltips and browser-tab titles.
- **INT-001-AC3:** Form-validation messages follow the active setup/household language even when the browser uses a different language; browser-generated validation must not override this behavior with another language.
- **INT-001-AC4:** Wishmas remains unchanged. User-entered household/list/recipient/user names, wish titles, notes and shop links remain unchanged in either language.
- **INT-001-AC5:** German translations are drafted in a table and reviewed/corrected by the user before implementation. Behavioral criteria and wording approval remain separate review steps.

## INT-006: Setup language and draft preservation

- **INT-006-AC1:** Before household language is saved, use the first supported browser preference: any German variant selects de-CH, any English variant selects en-US; otherwise use en-US. For example, fr-FR, de-DE, en-US selects German; en-GB, de-CH selects English.
- **INT-006-AC2:** Setup includes household language selection with Deutsch (Schweiz) and English (US), initially matching the detected setup language. Retain household name and appearance configuration; remove timezone and generated parent/family-link steps.
- **INT-006-AC3:** Changing the household-language selection does not change the wizard's display language. Setup and its validation/errors remain in the initially detected language until completion.
- **INT-006-AC4:** On validation errors preserve name, selected household language and appearance, but clear both password fields for re-entry. Errors remain in the setup display language.
- **INT-006-AC5:** Missing/unsupported submitted language shows a validation error and saves nothing, rather than silently selecting English.
- **INT-006-AC6:** Completing setup saves the selected household language and opens automatically signed-in admin access in that language, even if different from the setup display language.

## INT-007: Household configuration, login and fallback

- **INT-007-AC1:** Persist language in the single household YAML data file. When present, only exact de-CH or en-US values are accepted. Keep bundled translations; do not download translations at runtime.
- **INT-007-AC2:** Missing language retains the legacy English fallback. Present empty, null, malformed or unsupported language is invalid and shows an English configuration error without rewriting the file. Unreadable household configuration also shows an English configuration error.
- **INT-007-AC3:** After setup all viewers use the configured household language, including the unauthenticated admin login page reached with missing, invalid or regenerated user-link codes. Browser preferences do not override it.
- **INT-007-AC4:** Only YAML edits change household language after setup. Apply a valid change on the next page load or form submission, including validation/error responses; already-open pages remain unchanged until then.
- **INT-007-AC5:** Language changes affect presentation only, without changing IDs, names/content, password hash, user links, assignments, claims, schema keys or picture paths. Do not translate user-entered list names, including names such as Christmas 2026.
- **INT-007-AC6:** Authenticated not-found responses use household language. Remove obsolete unavailable-year/archive messages with their features.
- **INT-007-AC7:** If a German translation is missing at runtime, show English. Completeness tests must still fail for missing English/German entries; fallback does not satisfy translation acceptance.
- **INT-007-AC8:** Draft new and revised wording as an English/German table for user approval before implementation, including login/password/passphrase guidance, limit errors, users/recipients/lists, assignment/filter controls and destructive confirmations. Remove catalog entries no longer used by the final implementation.

## INT-005: Number presentation and sorting

### Acceptance criteria

- **INT-005-AC1:** Displayed reference prices use the household locale with grouping and exactly two decimal places: `1234.5` displays as `1’234.50` in `de-CH` and `1,234.50` in `en-US`; `12` displays as `12.00`. Currency remains excluded.
- **INT-005-AC2:** Price entry remains ungrouped with a decimal point in both languages, using the existing validation: nonnegative decimal values with at most two fractional digits. Regional grouping is display-only.
- **INT-005-AC3:** Stored YAML prices, edit-field values and numeric sorting remain unchanged by localized display formatting. Missing prices retain the existing optional-field and zero-for-sorting rules.
- **INT-005-AC4:** Non-price numbers, including claimant counts, never use grouping: for example, `1234`.
- **INT-005-AC5:** Title A–Z sorting follows the household language while still ignoring capitalization and accents. Equal titles retain added order, and the browser's saved sort choice is preserved.

## Relationships

- Access owns authentication, setup, users and list codes; these criteria govern language only.
- Wishlists owns lists, recipients and summaries. WLS-011 overview totals remain ungrouped with exactly two decimals; INT-005 grouping applies to individual wish prices.
- Wishes owns validation, price storage and sorting; these retain existing locale presentation rules.
- Claims uses household user IDs; changing display language does not change ownership.

## Out of scope

Other languages, automatic translation of user content, currency, post-setup language controls and external translation services remain out of scope. Remove unused year/timezone/archive/browser-name UI and messages rather than preserving them as dormant features.

## OPEN review items

Retained wording and all 112 new/revised English/German entries in translations.md are approved. No wording item remains OPEN; additional text discovered during implementation requires review.

## Implementation and verification

Pending implementation/integration. Preserve existing supported-locale, setup preference/delayed selection, validation localization, catalog completeness, missing/invalid-language fallback, exact price display and title-collation checks. Replace obsolete year/code/browser-name fixtures with new admin/user/list contexts. Verify household-language login for missing/invalid links, cleared password fields on errors, translated new controls/dialogs/errors and language edits without identity/data changes. Check English/German mobile/desktop and bundled delivery in container verification; do not treat previous flows as proof of new criteria.

## Milestone 2 evidence

All 112 approved HR entries are bundled. Catalog checks cover the approved target entries plus the 182-row legacy inventory still used by not-yet-migrated tests; 71 obsolete baseline-only entries remain until final cleanup. Current catalog size is 294. New templates/errors are checked for complete English/German wording and matching placeholders.

The new production renderer applies browser/setup language, delayed household selection, cleared password drafts and household-language login even for invalid links. Configuration errors remain English without writes. Four HTTPS browser scenarios verify opposite setup/household languages, invalid-password/confirmation feedback, logout/hash changes and translated rate limiting. Complete new management-page coverage remains later milestones; the domain is not marked Implemented.
