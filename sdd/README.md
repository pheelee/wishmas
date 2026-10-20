# Christmas wishlist specification

## Product intent

- Internationalization supports Swiss Standard German (`de-CH`, informal address) and American English (`en-US`) with a household-wide language selected during setup; see [Internationalization](internationalization.md).

- The app is primarily for the user's family in v1.
- V1 supports one household, with an individual wishlist for each person or child.
- Later public use means other families using the same app with their own separate households.
- A wish requires only a title. Multiple shop links, price, and note are optional and can be filled in later.
- Each wish has one optional numeric rough reference price, rather than a price per shop link. Price differences between linked shops are acceptable.
- Only parents may add, edit, and delete wishes on any person's list.
- Parents may edit or delete wishes even when they are claimed.
- Editing a claimed wish preserves its existing claim.
- Deleting a wish permanently removes the wish and its claim, if any.
- Wishes default to added order (oldest first). Users can switch to newest first, title A–Z, price low–high, or price high–low.
- Price sorting compares numeric amounts only.
- Wishes without a price are treated as zero for price sorting.
- Remember one selected sort order in the user's browser, shared across all wishlists.
- Previous years' wishlists are retained and viewable.
- Previous years' wishlists are read-only in the web UI. Direct edits to their YAML files remain possible and are accepted as a risk.
- Previous years' claims are also read-only in the app: even parents cannot add or remove them.
- Each new year, an empty wishlist is created automatically for each existing person.
- Adding a new person automatically creates an empty wishlist for that person in the current year.
- People's lists are displayed in the order the people were added.
- Parents can archive a person, retaining their past lists and stopping automatic creation of new yearly lists for them.
- Archiving hides all of the person's wishlists, across all years, from everyone, including parents. The data is retained.
- Restoring an archived person requires a manual YAML edit; the app does not provide restoration.
- On page load after restoration, automatically create an empty current-year wishlist if the person has none.
- Parents can rename a person through the app while preserving that person's existing wishlists and claims.
- Person names are unique within the household across active and archived people. Archived names remain reserved and cannot be reused for a new person.
- Name uniqueness ignores capitalization and surrounding spaces: `Anna`, `anna`, and ` Anna ` count as the same name.

## Principles

- V1 access requires a link and code, with no accounts.
- V1 access links contain the code and open directly without separate code entry.
- In v1, the family link grants access to all people in the current year.
- The current year is the calendar year, switching automatically on January 1.
- Choose the household timezone during setup. It determines the calendar-year switch for everyone using the household.
- After setup, changing the household timezone requires a manual YAML edit; no app control is provided.
- The parents' link grants access to all people and all years, including retained previous years' lists.
- Parents can regenerate either access code independently at any time and share its replacement link. Regeneration disables the old link for that code.
- Only parents may regenerate access codes; family-link holders cannot.
- Access codes remain valid until parents regenerate them; there is no automatic expiry.
- Revocation is immediate: the next action from an already-open page using the old code loses access without requiring a page reload.
- When parents regenerate their own access code, the current page automatically switches to the replacement parent link.
- A first-run setup wizard automatically generates both initial access codes. It is open until setup completes.
- After setup, requests without a valid access code show a not-found page.
- Everyone can see whether a wish is claimed.
- Family members can also see the number of claimants on each wish.
- Family members see “claimed by you” when a claim matches their current browser name.
- Only parents can see the names of people who claimed wishes.
- A wish may have multiple claimants.
- Parents can also claim wishes using their browser-stored name.
- Each name may claim a given wish only once.
- Claim ownership and duplicate checks compare names ignoring capitalization and surrounding spaces: `Anna`, `anna`, and ` Anna ` are the same claimant.
- A claim may be removed by its claimant or by parents.
- Claimant removal permission is based on the browser's current name. Renaming does not allow removal of claims under the previous name; parents can still remove them.
- Claimant identity is trust-based for family use; no protection against deliberate name impersonation is required.
- Ask for a person's name once, store it in their browser, show their own name on the page, and allow them to rename it.
- Renaming applies only to future claims. Existing claims retain the name used when they were made.

## Actors

- Parents write wishlists for their children.
- Only parents may create and edit adults' wishlists.
- Only parents may manage people and children's wishlists. Family-link holders are limited to viewing and claiming wishes, including removing their own claims under the confirmed name rules.
- Family members claim wishes.
- Parents and family members use separate codes in v1: one code for parents and one for family members.

## Domains

| Domain | Owns | Priority | Status |
| --- | --- | --- | --- |
| [Access](access.md) | Link-and-code entry and parent/family permissions | Must-have for v1 | Implemented; all acceptance criteria verified |
| [Wishlists](wishlists.md) | Individual lists and parents managing children's lists | Must-have for v1 | Implemented; all WLS criteria verified |
| [Wishes](wishes.md) | Wish details: required title; optional multiple shop links, numeric reference price, note and picture | Must-have for v1 | Implemented; all WSH criteria verified |
| [Claims](claims.md) | Claiming wishes and controlling claim visibility | Must-have for v1 | Implemented; all CLM criteria verified |
| [Internationalization](internationalization.md) | Household language, translations, locale formatting and language-aware sorting | Must-have for v1 | Implemented; all INT criteria verified |

## Out of scope

- Events other than Christmas in v1; other event types may be supported later.
- Accounts in v1.
- Multiple households are outside v1 scope and intended for later public use.

## OPEN questions

None currently listed. WSH-006 and its 17 English/German wording entries are approved and implemented; the user reviewed the completed work and authorized committing it.

