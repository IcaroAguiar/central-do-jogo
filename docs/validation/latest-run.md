# Latest QA smoke receipt

- **Date:** 2026-08-09
- **Branch / commit:** `cursor/phase4-structure-goal005-6a1d` (`2d970051704fc82e51205cac14910172a05fd3c3`)
- **Agent:** GitHub Actions CI on PR #16
- **Run:** https://github.com/IcaroAguiar/central-do-jogo/actions/runs/31295151575
- **Stack:** CI Postgres service, seeded DB, built web + Go server, Playwright Chromium

## Commands

| Gate | Command | Result |
|------|---------|--------|
| Go format / vet / race / lint | CI `go` job (`go test -race -count=1 ./…` + golangci-lint v2.12.2) | **pass** |
| Web lint / typecheck / unit / build | CI `web` job | **pass** — 14 files / 40 tests |
| OpenAPI | CI `openapi` job (`@redocly/cli` lint) | **pass** (5 existing warnings) |
| Docker image | CI `docker-build` job | **pass** |
| E2E public | CI `e2e-public` job | **pass** — 4/4 |

## E2E detail

```
✓ phase 4 API readiness › auth/me reports configuration without requiring a session
✓ phase 4 API readiness › push vapid endpoint responds according to VAPID configuration
✓ offline resilience › shows cached club data and an offline banner…
✓ public smoke › search → club → match → share
4 passed (4.4s)
```

## Residuals (dated 2026-08-09)

See [`phase-4-checklist.md`](./phase-4-checklist.md). Browser suites for admin, privacy journeys, full auth login, and Push-simulated e2e remain owned by **TASK-035** (Phase 5 / GOAL-006). Blocker for Google login e2e: stable test credentials not provisioned in CI.
