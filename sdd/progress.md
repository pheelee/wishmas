# Implementation progress

## Current checkpoint

- Current milestone: 1 — Setup and persistence, preparation.
- All four domain requirements and their OPEN-review updates are approved; none is Implemented.
- Last verified spec commit: `41038f4` — `docs: initial approved wishlist specification`.
- Planning checkpoint: pending commit with subject `docs: add implementation milestones and resume procedure`.
- Application code and tests: not started.

## Selected first slice

- ACC-001-AC1 through ACC-001-AC4, ACC-001-AC6, and ACC-001-AC7: open setup wizard, timezone selection, generated codes and copyable links, completion into parent access, and disabled wizard after completion.
- CON-004: retain setup in YAML across process restart.
- ACC-001-AC5 and broader Access criteria remain for subsequent verification; this slice will not mark the Access domain Implemented.

## Verification

- Read approved requirements, constraints, working rules, and clean initial Git state.
- Available local runtimes: Go 1.26.9 and Node.js 22.23.3.
- No application tests exist or have run. No expected failing behavioral test has yet been recorded.

## OPEN implementation preparation

- The storage hierarchy starts with household, but how the household is named during first-run setup has not been confirmed. Resolve before implementing household creation.

## Exact next action

Ask whether the setup wizard should collect a household name or v1 should use a fixed internal household identifier. Wait for the answer; update the owning setup requirement and observable criterion before writing dependent implementation. Then write and run the first failing setup/persistence behavioral test, record the failure, and implement the approved slice.

## Unfinished work

- Milestones 1–6 remain incomplete.
- No application files, running services, or deployment changes exist.
