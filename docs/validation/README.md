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

On every PR and on `main`:

- Go format / vet / race tests / lint
- Web lint / typecheck / unit / build
- OpenAPI lint
- Docker image build
- **E2E public smoke** (Playwright against seeded Postgres + Go SSR)

CI always uploads Playwright HTML report + `test-results` (traces/screenshots
are retained on failure) as workflow artifacts.

Phase-specific extras (auth login, admin, beta denominators) are tracked in
the checklists and TASK-033…036; they are not optional forever — they are
explicit residuals until those tasks land.
