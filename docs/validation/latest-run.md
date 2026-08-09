# Latest QA smoke receipt

- **Date:** 2026-08-09
- **Branch:** `cursor/phase4-structure-goal005-6a1d`
- **Sources:**
  1. GitHub Actions CI on PR #16 tip `2d97005` — [run 31295151575](https://github.com/IcaroAguiar/central-do-jogo/actions/runs/31295151575)
  2. Local stack re-run (Postgres 17 `:5433`, seeded, auth-enabled `cmd/server`, Playwright) producing [`phase-4-evidence-pack.md`](./phase-4-evidence-pack.md)

## Commands

| Gate | Command | Result |
|------|---------|--------|
| Go Phase 4 features | `go test -count=1 -v ./internal/features/{privacy,admin,reports,auth,preferences,push}/...` | **pass** — see `assets/phase-4/phase4-go-tests.log` |
| Web unit | `cd web && pnpm test` | **pass** — 14 files / 40 tests |
| E2E public (local) | `E2E_BASE_URL=http://127.0.0.1:8080 pnpm e2e` | **pass** — 4/4 |
| CI required-ci | Go / Web / OpenAPI / Docker / e2e-public on `2d97005` | **pass** |
| UI + backend pack | Playwright screenshots + curl receipts | **captured** — `docs/validation/assets/phase-4/` |

## E2E detail (local)

```
✓ phase 4 API readiness › auth/me reports configuration without requiring a session
✓ phase 4 API readiness › push vapid endpoint responds according to VAPID configuration
✓ offline resilience › shows cached club data and an offline banner…
✓ public smoke › search → club → match → share
4 passed (2.1s)
```

## Residuals (dated 2026-08-09)

See [`phase-4-checklist.md`](./phase-4-checklist.md). Browser suites for admin, privacy journeys, full auth login, and Push-simulated e2e remain owned by **TASK-035** (Phase 5 / GOAL-006). Blocker for Google login e2e: stable test credentials not provisioned in CI.
