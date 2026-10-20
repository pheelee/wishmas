# Internationalization

## Status and priority

Must-have for v1. Requirements and acceptance criteria below reflect confirmed interview answers. INT-001–INT-005 and all listed acceptance criteria are approved by the user; not Implemented. Application code is unchanged.

The user also approved the [161-row translation table](translations.md), including German wording and the new English validation messages, before implementation.

## INT-001: Supported languages and translation scope

Support Swiss Standard German (`de-CH`) with informal `du/dein` address and American English (`en-US`). Language is household-wide. The app name Wishmas remains unchanged.

### Acceptance criteria

- **INT-001-AC1:** The app supports `de-CH` and `en-US` only. German uses Swiss spelling and informal address; English uses US English.
- **INT-001-AC2:** All app-provided browser text follows the active language: setup, parent/family pages, current/history views, fields/help, dialogs, buttons, empty states, badges, counts, validation/server errors, clipboard feedback, accessibility labels, tooltips and browser-tab titles.
- **INT-001-AC3:** Form-validation messages follow the active setup/household language even when the browser uses a different language; browser-generated validation must not override this behavior with another language.
- **INT-001-AC4:** Wishmas remains unchanged. User-entered household/person/claimant names, wish titles, notes and shop links remain unchanged in either language.
- **INT-001-AC5:** German translations are drafted in a table and reviewed/corrected by the user before implementation. Behavioral criteria and wording approval remain separate review steps.

## INT-002: Language during first-run setup

Browser preference determines setup's initial display language and the initial household-language selection. A different selection applies only after Finish setup.

### Acceptance criteria

- **INT-002-AC1:** Before a household language is saved, use the first supported language in the browser preference list: any German variant selects `de-CH`, any English variant selects `en-US`. If neither is present, use `en-US`. For example, `fr-FR, de-DE, en-US` selects German; `en-GB, de-CH` selects English; `fr-FR` falls back to English.
- **INT-002-AC2:** Setup provides a Language dropdown beside household name/timezone. The options are `Deutsch (Schweiz)` and `English (US)`, initially selected according to AC1.
- **INT-002-AC3:** Selecting another household language does not switch the wizard's display language. Setup, validation errors and the access-link screen remain in the initially detected setup language.
- **INT-002-AC4:** On setup validation failure, retain the selected household language and all other entered values/settings. Continue displaying the wizard/errors in the initial setup language.
- **INT-002-AC5:** A setup submission with a missing or unsupported language value shows a validation error and saves nothing; it does not silently choose English.
- **INT-002-AC6:** Finish setup saves the selected household language and opens the parent page in that language. This applies whether or not it matches the setup display language.

## INT-003: Household YAML and language changes

Store language in household metadata. After setup, only a manual YAML edit changes it; there is no in-app language setting.

### Acceptance criteria

- **INT-003-AC1:** Completed setup saves `language: de-CH` or `language: en-US` in `household.yaml`. Only these exact values are valid when the field is present.
- **INT-003-AC2:** Existing households with no `language` field use English. A present empty string, empty YAML value, null, malformed value or unsupported value is invalid; it must not receive the missing-field fallback.
- **INT-003-AC3:** Invalid household language configuration shows an English configuration error and leaves YAML untouched. Unreadable household YAML also produces an English configuration error.
- **INT-003-AC4:** Everyone accessing a valid household uses its selected language. No post-setup app language setting is provided. Browser preferences do not replace a saved household language.
- **INT-003-AC5:** A valid manual language change applies on the next page load or form submission, including validation/error responses. An already-open page remains unchanged until then.
- **INT-003-AC6:** Changing language affects browser presentation only. Access codes, claims, user data, YAML field names and folder paths remain unchanged; storage still uses names such as `christmas` and `wishes`. The configured `language` value is the language setting.

## INT-004: Error-language selection and translation fallback

### Acceptance criteria

- **INT-004-AC1:** Requests with missing, invalid or revoked access codes use English because the household cannot be determined through those links.
- **INT-004-AC2:** With a valid access code and valid household language, not-found pages use the household language, including unavailable people or years.
- **INT-004-AC3:** With a valid access code and valid household language, invalid people/wishlist YAML errors use the household language. Invalid/unreadable household configuration follows INT-003-AC3 instead.
- **INT-004-AC4:** If a German translation is missing at runtime, show the English text rather than an error.
- **INT-004-AC5:** Automated completeness checks fail if any app-provided text lacks a German or English translation. Runtime fallback does not count as a complete translation for acceptance.

## INT-005: Number presentation and sorting

### Acceptance criteria

- **INT-005-AC1:** Displayed reference prices use the household locale with grouping and exactly two decimal places: `1234.5` displays as `1’234.50` in `de-CH` and `1,234.50` in `en-US`; `12` displays as `12.00`. Currency remains excluded.
- **INT-005-AC2:** Price entry remains ungrouped with a decimal point in both languages, using the existing validation: nonnegative decimal values with at most two fractional digits. Regional grouping is display-only.
- **INT-005-AC3:** Stored YAML prices, edit-field values and numeric sorting remain unchanged by localized display formatting. Missing prices retain the existing optional-field and zero-for-sorting rules.
- **INT-005-AC4:** Non-price numbers, including claimant counts and calendar years, never use grouping: for example, `1234` and `2026`.
- **INT-005-AC5:** Title A–Z sorting follows the household language while still ignoring capitalization and accents. Equal titles retain added order, and the browser's saved sort choice is preserved.

## Delivery constraints

Translations are bundled with the application. Displaying either language requires no external translation service or network request for translations. See CON-005.

## Relationships to existing domains

- Access owns setup completion, access validation and code regeneration; INT-002/INT-004 govern the language of those existing behaviors.
- Wishlists owns people, years and history; i18n changes their presentation, not their permissions, storage or lifecycle.
- Wishes owns price entry/storage and sorting. INT-005 adds regional display formatting and household-locale title collation.
- Claims owns identities, ownership, counts and visibility; i18n translates app text without translating identities or changing permissions.

## Out of scope

- Languages other than German and English.
- Automatic translation of user-entered content.
- A post-setup in-app language setting.
- External translation services or runtime translation downloads.
- Currency, localized storage keys/paths, or changes to existing identity/access rules.

## OPEN questions

None. Behavioral criteria and translation wording are approved; implementation has not started. No requirement is marked Implemented.
