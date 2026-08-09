# Phase 4 / GOAL-005 evidence checklist

Living checklist for identity, push, operations, and package-layout hygiene.
Update the **Receipt** column when evidence changes; never invent execution
evidence.

| Item | Status | Receipt |
|------|--------|---------|
| ADR 0002 package layout | done | `docs/adr/0002-package-layout.md`; store→`internal/store`, `httpapi`, feature `Register`, `internal/app` |
| TASK-027 Auth OAuth/session | done | Merged PR #10; Go auth package tests in CI |
| TASK-028 Preferences merge | done | Merged PR #11; web prefs unit tests in CI |
| TASK-029 Push subscribe/outbox | done | Merged PR #12; push package tests in CI |
| TASK-029 VAPID real deliver + smoke CLI | done | Merged PR #13; `cmd/pushsmoke` + deliverer tests |
| TASK-030 Privacy export/delete | done | `internal/features/privacy/` + `web/src/features/settings/`; unit tests in CI; OpenAPI privacy paths |
| TASK-031 Admin panel | done | `internal/features/admin/` + `web/src/features/admin/`; unit tests in CI; OpenAPI admin paths |
| TASK-032 Anonymous reports | done | `internal/features/reports/` + `web/src/features/reports/`; unit tests in CI; OpenAPI reports paths |
| TEST-008 public e2e smoke on every PR | required | `e2e-public` job (this bridge) |
| TEST-010 offline e2e on every PR | required | `e2e/public/tests/offline.spec.ts` via `e2e-public` |
| TEST-011 auth security suite | residual | Unit coverage exists; dedicated e2e login deferred until stable Google test creds |
| TEST-008 Push-simulated e2e | residual | TASK-035 |
| TEST-009 Admin e2e | residual | Skeleton in `e2e/admin/`; full suite TASK-035 |
| TEST-012 Privacy e2e | residual | Unit coverage for export/delete/retention; browser suite TASK-035 |

## How to refresh local receipts

```bash
# Unit gates (always)
export GOTOOLCHAIN=auto
go test -race -count=1 ./...
cd web && pnpm lint && pnpm typecheck && pnpm test && pnpm build

# Public e2e (Postgres + seed + built web + Go server) — see e2e/public/README.md
# Then record commands/results in latest-run.md
```

## Operator push smoke (manual, not CI)

Requires real browser subscription + VAPID keys (never commit secrets):

```bash
go run ./cmd/pushsmoke -email=you@example.com
```
