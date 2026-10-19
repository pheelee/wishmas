# Working rules

- Never assume. If anything is unclear, ask the user. Mark unconfirmed items as OPEN.
- Ask ONE question at a time. Wait for the user's answer before the next.
- Write into the spec only what the user has explicitly confirmed.
- No code until requirements have acceptance criteria and the user approves them.
- Offer options with trade-offs when the user must choose; the user decides.
- Do not start the next step until the user says so.

# Spec-driven development

- The spec lives in `sdd/`: `README.md` (principles, actors, domains with priority and status, out-of-scope), `constraints.md` (`CON-*` entries), `glossary.md`, and `changes.md`.
- For each change, put behavior in the owning domain requirement and write observable acceptance criteria and constraints BEFORE implementation.
- Write a behavioral test that fails first.
- Link code and tests to acceptance criteria with `@impl` and `@test` anchors.
- Update documentation and `sdd/changes.md`.
- Never mark a requirement Implemented while any acceptance criterion is only partly met.
- Record the tech stack as constraints AFTER product intent, actors, and domains are settled.

# Interview sequence

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
