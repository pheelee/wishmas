# Claims

## Status

Approved by the user. In progress: read-only status/counts and parent-only names are verified. Browser identity, claim toggles, ownership actions, and complete page-load integration remain pending.

## Priority

Must-have for v1.

## CLM-001: Browser name and trust

Ask for a name once, store it in the browser, show the current name on the page, and allow renaming. Renaming affects future claims only. Identity is trust-based for family use; protection against deliberate name impersonation is not required. Ownership and duplicate checks ignore capitalization and surrounding spaces.

If the browser has no stored name, ask when the person first attempts to claim a wish, rather than when opening an access link. Browsing does not require a name.

Trim surrounding spaces from claimant names. Require 3–50 characters consisting only of Unicode letters and digits. Claimant names need not be unique across browsers; matching names represent the same claimant for ownership and duplicate checks.

### Acceptance criteria

- **CLM-001-AC1:** The app asks for a name and remembers it across page loads in the same browser.
- **CLM-001-AC2:** The current browser name is displayed and can be changed.
- **CLM-001-AC3:** After renaming, new claims use the new name; existing claims retain the name used when made.
- **CLM-001-AC4:** `Anna`, `anna`, and ` Anna ` count as the same claimant for ownership and duplicate checks.
- **CLM-001-AC5:** Opening a valid access link without a stored name allows browsing without a name prompt; first attempting to claim a wish prompts for a name.
- **CLM-001-AC6:** Entering or renaming a claimant trims surrounding spaces and rejects blank names, names outside 3–50 characters after trimming, and characters other than Unicode letters and digits.
- **CLM-001-AC7:** Different browsers can use the same claimant name; matching names follow the same ownership and per-wish duplicate rules.

## CLM-002: Creating claims

Family members and parents can claim wishes using their browser-stored name. A wish may have multiple claimants, but each name can claim it only once. Previous years' claims are read-only in the app.

An unsuccessful claim attempt shows an inline error and refreshes the displayed claim status without changing existing claims, including duplicate submissions from stale pages.

Claiming and removing one's own claim use a simple toggle. Neither claim creation nor claim removal requires confirmation.

### Acceptance criteria

- **CLM-002-AC1:** Family-link and parent-link holders can claim a current-year wish under their current browser name.
- **CLM-002-AC2:** A different claimant can claim an already-claimed wish, preserving existing claims.
- **CLM-002-AC3:** The same name, compared using CLM-001, cannot create a second claim on the same wish.
- **CLM-002-AC4:** No user, including a parent, can add claims to previous years' wishes through the app.
- **CLM-002-AC5:** An unsuccessful claim attempt shows an inline error, refreshes the displayed claim status, and leaves existing claims unchanged, including a duplicate attempt submitted from a stale page.
- **CLM-002-AC6:** Toggling an unclaimed-by-you current-year wish creates a claim under the current browser name without a confirmation step, prompting for a name first if needed.

## CLM-003: Claim visibility

Status: In progress. AC1/AC2/AC5 are verified; AC3/AC4 and the actual claim-mutation flow in AC6 remain pending.

Everyone sees claimed/unclaimed status. Family members also see the number of claimants and “claimed by you” when a claim matches their current browser name. Only parents see claimant names.

In v1, other viewers see claim changes on their next page load, without automatic updates to an already-open page.

### Acceptance criteria

- **CLM-003-AC1:** A wish with no claims is shown as unclaimed; one with any claims is shown as claimed.
- **CLM-003-AC2:** Family members see the number of claimants without seeing claimant names.
- **CLM-003-AC3:** Family members see “claimed by you” for claims matching their current browser name, using CLM-001 comparisons.
- **CLM-003-AC4:** Changing to a name that matches no claim removes the “claimed by you” indication for that wish.
- **CLM-003-AC5:** Parents can see the names associated with the claims on a wish.
- **CLM-003-AC6:** After another viewer creates or removes a claim, reloading the page shows the updated claim status and count; an already-open page does not update automatically in v1.

## CLM-004: Removing claims

Parents may remove any current-year claim. Family members may remove claims matching their current browser name only. Renaming does not confer removal permission for old-name claims. Previous years' claims are read-only in the app.

### Acceptance criteria

- **CLM-004-AC1:** A family member can remove a current-year claim matching their current browser name.
- **CLM-004-AC2:** A family member cannot remove a claim under a different name, including their previous name after renaming.
- **CLM-004-AC3:** Parents can remove any current-year claim regardless of their own browser name.
- **CLM-004-AC4:** Removing one claim preserves other claims on the wish and updates the claimant count.
- **CLM-004-AC5:** Removing the last claim makes the wish unclaimed.
- **CLM-004-AC6:** No user, including a parent, can remove previous years' claims through the app.
- **CLM-004-AC7:** Toggling off one's own current-year claim removes it without confirmation; parent removal of another person's current-year claim also requires no confirmation.

## Related requirements and constraints

- Access: link permissions and household-timezone year selection.
- Wishlists: archived people's lists are hidden; historical lists are read-only in the web UI.
- Wishes: editing preserves existing claims; permanent deletion removes the wish and its claims.
- CON-001 through CON-004: confirmed hosting, stack, and YAML persistence constraints apply.

## OPEN questions

None currently listed.

## Implementation and verification

Read-only claim visibility was selected as the first privacy foundation before browser identity and mutation controls. Failing-first Go tests exposed missing status/name presentation and absent claim-data validation.

- CLM-003-AC1/AC2/AC5: verified claimed/unclaimed status, explicit zero/singular/plural counts, parent-only claimant lists, and absence of claimant names in family HTML responses. Current-year views and parent history are covered.
- CLM-003-AC6: page-load behavior is verified using manual persisted claim changes. Open pages stay unchanged; reload shows the updated count/names. Integration with another viewer's actual claim creation/removal remains pending.
- CLM-003-AC3/AC4: pending browser identity and claimed-by-you behavior.
- CLM-001/CLM-002/CLM-004: browser identity and all claim mutation/ownership flows remain pending. Shared validation already enforces trimmed Unicode name boundaries and case-insensitive per-wish uniqueness for manual YAML; that does not complete those requirements.
- CON-004: invalid claim structures/names/duplicates fail without rewriting files. Presentation normalizes a decoded copy only; wish edits preserve raw claims and future fields. Unicode names of 50 characters are accepted; 51 are rejected.

Implementation: `internal/app/claims.go`, wish-card decoding, and the wishlist template. Tests: `claim_visibility_test.go` and `tests/browser/claim-visibility.spec.js`, with criterion anchors. Browser checks inspect full family HTML as well as visible content, unchanged files, parent history, manual refresh, invalid-data failure, and mobile dark styling. No claim editing/removal controls are included in this slice. No whole Claims requirement is marked Implemented.
