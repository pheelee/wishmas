# Implementation plan

## Approval and method

The user approved this milestone approach and its on-disk tracking and resume procedure. Domain requirements and acceptance criteria remain authoritative. No milestone is complete until its linked criteria have evidence; partial progress must remain explicit.

For each slice: select criteria, write and run a failing behavioral test, implement with anchors, verify, update domain documentation and change log, record the next action, and commit the verified slice.

## Milestones

| Order | Milestone | Criteria and dependencies | Observable checkpoint | Status |
| --- | --- | --- | --- | --- |
| 1 | Setup and persistence | ACC-001; CON-004. Initial link entry depends on the minimum ACC-002 access behavior. | Complete the open wizard, choose timezone, copy generated links, finish into parent access, and restart without losing setup. | Complete for setup and household metadata |
| 2 | Access | ACC-002 and ACC-003; depends on milestone 1. List/year scope verification also needs milestone 3. | Verify link roles, invalid-code behavior, calendar-year boundaries, independent regeneration, immediate revocation, and parent-link switching. | Complete; all Access criteria verified |
| 3 | People and wishlists | WLS-001 through WLS-004; depends on persisted setup and access. | Manage people, validate names, preserve history, create annual lists, archive, and restore via YAML. | Complete; all WLS criteria verified |
| 4 | Wishes | WSH-001 through WSH-003; depends on people and lists. Claim preservation/removal verified with seeded YAML before milestone 5. | Manage and validate wishes and links, confirm deletion, and remember sorting across lists. | Complete; all WSH criteria verified |
| 5 | Claims | CLM-001 through CLM-004; depends on access and wishes. | Use browser names, toggle multiple claims, enforce name-based ownership and visibility, and verify historical read-only behavior. | Complete; all CLM criteria verified |
| 6 | Deployment and complete flows | CON-001 through CON-004 and all domain criteria; depends on preceding milestones. | Run Docker deployment, finish Christmas styling, and verify complete parent/family flows and all remaining cross-domain criteria. | Not started |

## Completed shared frontend slices

The original December-night theme and subsequent user-approved UI polish are implemented. Current CON-003-AC1–AC10, ACC-004, WSH-004, and WLS-005 cover dark styling, steady scalable bulbs, setup previews and household YAML appearance, always-present margin snow/static reduced motion, browser timezone prefill, inline clipboard/regeneration icons, native modal action forms, wish cards, and 75% desktop page width. Existing data/access/history/sorting behavior remains verified. The approved Gift-wrap card revision (WSH-005) adds decorative accents, badges, optional detail styling and separated inline footers. The Claims identity/toggle milestone is also complete; deployment and final complete-flow checks are next.

## Checkpoints

- Commit planning and resume instructions before application code.
- Use a separate commit for each verified behavioral slice.
- Keep partial criteria and unverified browser interactions visible in domain implementation records and progress notes.
- Inspect test failures before deciding whether to broaden checks. Passing focused checks alone does not prove a complete milestone.

## Resume procedure

1. Read `AGENTS.md`, this plan, `sdd/progress.md`, and the relevant requirements and constraints.
2. Inspect `git status --short` and recent commits. Compare unfinished files with the recorded checkpoint.
3. Run the recorded verification command when one exists; investigate any discrepancy with the saved result.
4. Continue from the exact next action. Ask the user one question at a time for unresolved decisions and wait before doing dependent work.
5. Update progress and change records before the next checkpoint commit.
