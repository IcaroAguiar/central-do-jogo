# Validation and QA evidence

Contributor-facing English docs for quality gates. Product-facing beta
narratives for end users stay in pt-BR when published under product docs.

## Purpose

Every material phase advance must leave a **reproducible evidence trail**, not
only a green unit-test CI. This directory is the contract for:

1. What must run before a phase/task is called done.
2. Where CI artifacts and human checklists live.
3. What remains residual until Phase 5 (TASK-033…036).

## Contents

| Path | Role |
|------|------|
| [phase-evidence.md](./phase-evidence.md) | Required evidence matrix by phase / TEST-* id |
| [phase-4-checklist.md](./phase-4-checklist.md) | Living checklist for GOAL-005 (auth, prefs, push, privacy…) |
| [phase-4-evidence-pack.md](./phase-4-evidence-pack.md) | UI screenshots, backend HTTP receipts, and phase test logs |
| [assets/phase-4/](./assets/phase-4/) | Binary/text receipts referenced by the Phase 4 evidence pack |
| [latest-run.md](./latest-run.md) | Last local/CI smoke receipt committed with this bridge |

Private beta samples, raw screenshots with PII, and non-redistributable
captures stay under ignored paths (`docs/private/`, `*.private.md`) — see
[docs/README.md](../README.md).

## Minimum PR bar (from this bridge onward)

Every PR runs one `verify` job: Go format/vet/lint, web lint/typecheck,
affected Go/web unit tests, and OpenAPI lint, with a 300-second work budget.
Fork PRs require `run-ci`; `verify` runs on the next push or reopen after
labeling, not on the label event.

The `full-ci.yml` workflow runs with `run-full-ci`, manual dispatch, or push
to `main`: race tests, PostgreSQL tests in `db` and `internal/store`, web and
Docker builds, and public E2E smoke against seeded Postgres + Go SSR.
Only the E2E tier uploads Playwright reports and failure traces/screenshots.

Phase-specific extras (auth login, admin, beta denominators) are tracked in
the checklists and TASK-033…036; they are not optional forever — they are
explicit residuals until those tasks land.
