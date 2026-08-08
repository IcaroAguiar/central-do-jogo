# Phase evidence matrix

Status values:

- `required-ci` — must pass on PR/`main` CI
- `checklist` — human or scripted receipt recorded in a checklist
- `phase-5` — planned under GOAL-006 (TASK-033…036); residual until then
- `n/a` — not applicable yet

| ID | Surface | Phase 3 | Phase 4 | Phase 5 / release |
|----|---------|---------|---------|-------------------|
| TEST-001 | Adapter/parser fixtures (no network) | required-ci | required-ci | required-ci |
| TEST-002 | Adapter contract | required-ci | required-ci | required-ci |
| TEST-003 | Reconciliation | required-ci | required-ci | required-ci |
| TEST-004 | Postgres constraints / leases | checklist | checklist | required-ci (expand) |
| TEST-005 | Go `-race` | required-ci | required-ci | required-ci |
| TEST-006 | OpenAPI lint + generated types | required-ci | required-ci | required-ci |
| TEST-007 | React unit (search, gaps, prefs…) | required-ci | required-ci | required-ci |
| TEST-008 | Playwright public journeys | required-ci (smoke) | required-ci (smoke) | required-ci (+ Push sim) |
| TEST-009 | Playwright admin | n/a | checklist (after TASK-031) | required-ci |
| TEST-010 | Offline banner / cache label | required-ci | required-ci | required-ci |
| TEST-011 | Auth/CSRF/session security | n/a | checklist + unit | required-ci |
| TEST-012 | Privacy export/delete/retention | n/a | checklist (TASK-030) | required-ci |
| TEST-013 | Load / p95 | n/a | n/a | phase-5 |
| TEST-014 | Lighthouse / WCAG evidence | n/a | n/a | phase-5 |
| TEST-015 | Live beta denominators | n/a | n/a | phase-5 |

## CI jobs ↔ matrix

| Job | Covers |
|-----|--------|
| `go` | TEST-001…005 (unit/race/lint portion) |
| `web` | TEST-007 |
| `openapi` | TEST-006 |
| `docker-build` | deploy reproducibility |
| `e2e-public` | TEST-008 smoke + TEST-010 |

## Completion rule

A phase GOAL may be marked done in `docs/product/initial-plan.md` only when:

1. Every `required-ci` row for that phase is green on the merge commit, and
2. Every `checklist` row has an updated dated receipt in the phase checklist
   (or an explicit dated residual with owner/blocker).

Claims of “done” without (1)+(2) are incomplete.
