# Glossary

- **Household:** the v1 boundary owning configuration, users, recipients, lists, wishes and sessions; setup creates one.
- **Admin/parent:** visitor signed in with the shared household password; not a separate account or a parent-code link.
- **Household user:** predefined gift-giver with a stable ID and individual link for each list. User identity is distinct from recipient identity.
- **Recipient:** household roster entry representing someone receiving gifts; selected independently for each list.
- **List:** manually created, freely named collection with selected recipients. Dates in its name impose no lifecycle/access restrictions.
- **Assignment:** optional association of one household wish with one selected recipient in one list. Moving changes this association without copying the wish or clearing claims.
- **Unassigned wish:** household wish without an assignment, visible only to admins; may retain claims after manual unassignment, but accepts no new claims until assigned.
- **User/list link:** code identifying a household user and one list. Every user has one for every list; no membership operation exists. Sharing controls who receives it.
- **Claim:** reservation by a stable household user ID; at most one per user per wish. Renaming displays the current user name on existing claims.
- **Manual unassignment:** returns a wish to the household pool while retaining claims; different from recipient/list deletion cleanup, which clears affected claims.
- **Complete wish:** title, supplied reference price (including zero) and at least one shop link. Optional notes/pictures do not affect completeness; only title is required for saving.
- **Reference total:** sum of supplied prices for a recipient's assigned wishes in one list, including incomplete wishes; ungrouped/two decimals/no currency. Count badges include complete wishes only.
- **Admin session:** persistent browser authentication represented by a hashed token in household YAML; ends on current-browser logout or password-hash change. Browser cookie deletion can also remove browser access.
- **Household language:** de-CH (informal Swiss Standard German) or en-US (American English), selected during setup and changed thereafter through YAML; used for login and all post-setup pages.
- **Setup display language:** initially browser-detected language throughout setup, independent of the household-language selection until completion.
- **Baseline:** existing implementation at main commit 1992bdf with legacy data. Refactor code must not run against that original data without explicit copied-data migration.
- **Cutover:** separately approved switch to matching new code and migrated data; distinct from reviewing branch checkpoints or merging code.
