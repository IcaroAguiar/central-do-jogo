# Phase 4 / GOAL-005 evidence checklist

Living checklist for identity, push, operations, and package-layout hygiene.
Update the **Receipt** column when evidence changes; never invent execution
evidence.

Per [`phase-evidence.md`](./phase-evidence.md): GOAL-005 may be marked done
only when (1) every Phase 4 `required-ci` row is green on the tip/merge commit
and (2) every `checklist` row has a dated receipt **or** a dated residual with
owner/blocker.

| Item | Status | Receipt |
|------|--------|---------|
| ADR 0002 package layout | done | `docs/adr/0002-package-layout.md`; store→`internal/store`, `httpapi`, feature `Register`, `internal/app` |
| TASK-027 Auth OAuth/session | done | Merged PR #10; Go auth package tests in CI |
| TASK-028 Preferences merge | done | Merged PR #11; web prefs unit tests in CI |
| TASK-029 Push subscribe/outbox | done | Merged PR #12; push package tests in CI |
| TASK-029 VAPID real deliver + smoke CLI | done | Merged PR #13; `cmd/pushsmoke` + deliverer tests |
| TASK-030 Privacy export/delete | done | `internal/features/privacy/` + `web/src/features/settings/`; unit tests in CI `go`/`web` jobs; OpenAPI privacy paths |
| TASK-031 Admin panel | done | `internal/features/admin/` + `web/src/features/admin/`; unit tests in CI; OpenAPI admin paths |
| TASK-032 Anonymous reports | done | `internal/features/reports/` + `web/src/features/reports/`; unit tests in CI; OpenAPI reports paths |
| TEST-008 public e2e smoke (required-ci) | done | CI run [31295151575](https://github.com/IcaroAguiar/central-do-jogo/actions/runs/31295151575) on `2d97005` — 4/4; see `latest-run.md` |
| TEST-010 offline e2e (required-ci) | done | Same CI run; `e2e/public/tests/offline.spec.ts` passed |
| TEST-011 auth security suite (checklist) | partial | **Receipt (2026-08-09):** Go auth unit/CSRF/session coverage in CI `go` job. **Residual (2026-08-09):** dedicated Playwright login e2e deferred — owner **TASK-035**; blocker: stable Google test credentials not in CI |
| TEST-012 Privacy export/delete/retention (checklist) | partial | **Receipt (2026-08-09):** `internal/features/privacy` unit tests (export/delete/events/retention) in CI `go` job. **Residual (2026-08-09):** browser privacy journey e2e — owner **TASK-035** |
| TEST-008 Push-simulated e2e (beyond smoke) | residual | **Residual (2026-08-09):** owner **TASK-035**; smoke covers VAPID readiness only (`phase4-api.spec.ts`) |
| TEST-009 Admin e2e (checklist) | residual | **Residual (2026-08-09):** skeleton in `e2e/admin/`; full allowlist/health/override/audit/report suite — owner **TASK-035**; blocker: maintainer session fixture for Playwright |

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
