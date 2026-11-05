# Working rules

- Never assume. If anything is unclear, ask the user. Mark unconfirmed items as OPEN.
- Ask ONE question at a time. Wait for the user's answer before the next.
- Write into the spec only what the user has explicitly confirmed.
- No code until requirements have acceptance criteria and the user approves them.
- Offer options with trade-offs when the user must choose; the user decides.
- Do not start the next step until the user says so.
- Use Conventional Commits.
- Before every local commit, run `golangci-lint config verify` and `golangci-lint run ./...` from the repository root using the version pinned in `.github/workflows/ci.yml` (currently v2.15.0). Fix findings before committing; keep any suppression rule-specific and justified. Do not skip the check or disable linters to make a commit pass. See `docs/development-checks.md` for tool setup and the complete validation commands.
- Let the user review each change before committing it.
- For temporary work, create a dedicated progress file separate from permanent specification and change records. Record only completed work and confirmed facts; keep proposals, unconfirmed ideas, and work awaiting review out of permanent progress and change records. Temporary progress files remain uncommitted until their contents qualify as completed, confirmed records.
- Commit only confirmed facts and completed, verified changes. Never commit intermediate design proposals, work awaiting review, or unfinished implementation. Keep such work uncommitted until it is confirmed and complete.

# Spec-driven development

- The spec lives in `sdd/`: `README.md` (principles, actors, domains with priority and status, out-of-scope), `constraints.md` (`CON-*` entries), `glossary.md`, and `changes.md`.
- For each change, put behavior in the owning domain requirement and write observable acceptance criteria and constraints BEFORE implementation.
- Write a behavioral test that fails first.
- Link code and tests to acceptance criteria with `@impl` and `@test` anchors.
- Update documentation and `sdd/changes.md`.
- Never mark a requirement Implemented while any acceptance criterion is only partly met.
- Record the tech stack as constraints AFTER product intent, actors, and domains are settled.

# Interview sequence

## Implementation checkpoints and resuming

- Before implementation or resumption, read `sdd/implementation.md`, `sdd/progress.md`, the owning domain requirements, and applicable constraints. Inspect Git status and the last commit before modifying unfinished work.
- Follow the approved milestone plan. Select a small set of acceptance criteria per slice; record them in `sdd/progress.md` before starting.
- Run a behavioral test before implementation and record its expected failure. Record subsequent verification commands and results, including failures and limitations.
- Link tests and implementation to criterion IDs with `@test` and `@impl` anchors.
- At each checkpoint, update `sdd/progress.md` with the current milestone, verified commit, unfinished work, blockers, and exact next action. Update the owning domain's implementation record and `sdd/changes.md` when behavior changes.
- Commit completed, verified slices separately using Conventional Commits. Do not commit intermediate proposals or unfinished work. A checkpoint is not evidence that an entire domain is Implemented. Mark a requirement Implemented only after every criterion is fully verified.
- On resumption, inspect uncommitted work, verify the recorded test state, and continue from the documented next action. Do not overwrite unfinished work or infer an answer to an OPEN question.
- Until a checkpoint commit exists, record its commit identifier as pending. Resolve that checkpoint by its unique commit subject after resumption; do not use a self-referential commit hash.

1. Create the SDD skeleton with empty sections and this file.
2. Interview the user on the initial open questions, one at a time, then fill in actors, principles, and out-of-scope from confirmed answers only.
3. Propose candidate domains and ask which to keep, merge, or drop.
4. Interview the user on hosting, frontend, and backend with 2–3 options and trade-offs each; record choices as `CON-*` constraints.
5. Interview the user on data persistence, then sharing and privacy.
6. Write requirements with acceptance criteria per domain, flagging every guess as OPEN. Stop for approval after each domain.

## Initial open questions (in order)

1. How are parents and family members separated in v1 with a link-and-code model (for example, two codes, one code with trust-based roles, or another approach)?
2. Who may see a claim? Should the list owner or child be prevented from seeing which wishes are claimed?
3. Does later public use mean publicly discoverable lists, or other families running their own households in the same app?
4. What does a single wish contain: title only, or also link, price, note, priority, image?
