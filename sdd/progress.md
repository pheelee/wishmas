# Implementation progress

## Current checkpoint

- Completed slice: December-night theme and gentle snowfall, CON-003-AC1–AC5. Appearance and acceptance criteria were approved before implementation; all five criteria are verified.
- Code checkpoint: the commit titled `feat(theme): add December night styling and gentle snowfall` contains this completed slice and verification record. Locate it with `git log --oneline`; no intermediate proposal or unfinished code is part of this checkpoint.
- Previous verified functional code commit: `79b3ca7` — independent access regeneration and calendar verification.
- Milestone 1 remains complete for setup and household metadata. Milestone 2 still requires actual list/year scope integration from milestone 3.
- ACC-001 is Implemented. Access as a domain remains In progress; ACC-002 and ACC-003 have documented list-scope dependencies. No unverified domain is marked Implemented.
- User's additional commit rule is saved in AGENTS.md: Conventional Commits; only confirmed facts and completed, verified changes; no intermediate proposals or unfinished-work commits.

## Completed theme behavior

- Approved midnight-navy surfaces, ivory/soft-white text, ice-blue input borders, and mint primary actions on every existing screen.
- Steady coral, gold, mint, and blue Christmas bulbs, rendered with local HTML/CSS.
- Exactly 16 CSS snow particles, 1–3px, 14–30% opacity, 22–40-second downward journeys. Particles remain in the outer side margins, behind content, with no pointer or focus interaction.
- Keyboard-accessible Pause snow / Resume snow control in the header, with browser-stored preference. Snow does not animate under reduced motion, including before JavaScript runs; the resume control is disabled while that preference is active.
- Dark styling also covers validation, invalid-data errors, and not-found pages. Fifty-character household names wrap at 320px without horizontal overflow.
- Existing setup, generated/copyable access links, parent/family permissions, independent regeneration, immediate revocation, and calendar behavior remain verified.
- Approved baseline and original generation prompt are preserved in `design/december-night/README.md` and `approved-baseline.png`. Live motion is CSS, not generated imagery or a third-party animation dependency.

## Verification and evidence

- Baseline Go race tests and vet passed before theme implementation.
- RED before theme code: the browser color check measured the existing light background at luminance 0.955 (required under 0.03 in the test); snow tests failed on missing particles and controls. Test-helper errors with absent panels and modern CSS color serialization were corrected before recording the clean appearance failure.
- Additional RED: maximum-length household names overflowed at 320px, and invalid YAML returned an unthemed plain-text error. Tests were added before fixing both cases.
- GREEN: `npm run test:e2e` — all five independent Chromium scenarios passed, covering the complete setup/access lifecycle plus theme, particle limits/margins, actual movement, keyboard pause, remembered state, initial/dynamic reduced motion, long names, and error pages.
- `npm run build:css` — passed; compiled CSS included.
- `go test -race ./...` — passed, including the new shared-error-theme test and existing persistence/calendar/regeneration tests.
- `go vet ./...` — passed.
- `go build -o /tmp/wishmas-check ./cmd/wishmas` — passed.
- Actual desktop (1536px), mobile (390px), and parent-access screenshots reviewed. Browser checks also cover 320px width, dark surfaces, label contrast, and no horizontal overflow.
- `git diff --check` — passed before commit.
- No dependencies were added or changed. No deployment was made; isolated test processes and data were cleaned up.

## Exact next functional action

Resume milestone 3 by reading `sdd/wishlists.md`. Record a first people/list slice's criterion IDs before editing code. Write failing behavioral tests for parent-only person creation, an empty current-year Christmas list, and confirmed Unicode name validation/uniqueness (including reserved archived names); then implement with anchors.

When actual lists exist, verify ACC-002-AC2/AC3/AC4 and ACC-003-AC2 against current/historical data: parent history, all-person scope, current-year-only family access, correct list selection at New Year, and preserved scope after regeneration. Do not mark these fully met before those integration checks pass.

## Remaining project work

- Actual people/list/year/event storage, wishes, claims, Docker packaging, and complete-flow verification remain separate future milestones.
- Valid links currently show the household landing page; parents also have access-link controls.
- Setup previews remain in memory until Finish setup. Completed setup is persisted in `<data>/households/<name>/household.yaml` and survives restart.
- No unresolved theme criterion blocks this completed slice. User feedback can drive a subsequent explicitly specified change.

## Resume

Read AGENTS.md, implementation.md, relevant requirements, and this checkpoint. Inspect Git status/history. On a fresh environment run `npm ci` and `npx playwright install --with-deps chromium`; then run `go test -race ./...` and `npm run test:e2e`. Each browser test builds/runs an isolated app on a dynamic local port. Generated CSS is committed, so normal Go-only startup does not need Node.js. Continue from the next functional action when authorized.
