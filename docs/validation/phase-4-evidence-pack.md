# Phase 4 / GOAL-005 evidence pack

Dated local evidence captured **2026-08-09** on tip of
`cursor/phase4-structure-goal005-6a1d` against seeded Postgres + Go SSR with
auth enabled (demo Google client id / session secret — not production
credentials). Artifacts live under [`assets/phase-4/`](./assets/phase-4/).

This pack closes the gap called out by the evidence contract: UI screenshots,
backend HTTP receipts, and test transcripts that prove Phase 4 surfaces — not
only checklist checkboxes.

## 1. UI evidence

| Surface | Task | Screenshot |
|---------|------|------------|
| Home (authenticated maintainer chrome) | TASK-027 | [01-home.png](./assets/phase-4/01-home.png) |
| Club page + anonymous report form | TASK-032 | [02-club-flamengo-with-report.png](./assets/phase-4/02-club-flamengo-with-report.png) |
| Match page + report form | TASK-032 | [03-match-with-report.png](./assets/phase-4/03-match-with-report.png) |
| `/conta` export/delete (logged in) | TASK-030 | [04-conta-settings.png](./assets/phase-4/04-conta-settings.png) |
| `/conta` gated (anonymous) | TASK-030 | [04b-conta-anonymous.png](./assets/phase-4/04b-conta-anonymous.png) |
| `/admin` health / at-risk / open reports | TASK-031/032 | [05-admin-panel.png](./assets/phase-4/05-admin-panel.png) |
| `/admin` gated (anonymous) | TASK-031 | [05b-admin-anonymous.png](./assets/phase-4/05b-admin-anonymous.png) |

Observed in admin UI: open report message
`Transmissao incorreta no smoke de evidencia Phase 4` after `POST /api/v1/reports`.

## 2. Backend evidence (HTTP)

Seed: [`seed.log`](./assets/phase-4/seed.log) — 20 clubs / 12 matches.

| Probe | Receipt |
|-------|---------|
| `GET /api/v1/auth/me` (anon) | [`auth-me.json`](./assets/phase-4/auth-me.json) → `authEnabled:true`, `authenticated:false` |
| `GET /api/v1/auth/me` (maintainer session) | [`auth-me-maintainer.json`](./assets/phase-4/auth-me-maintainer.json) → `role:maintainer` |
| `GET /api/v1/privacy/export` | [`privacy-export.json`](./assets/phase-4/privacy-export.json) |
| `POST /api/v1/privacy/events` | [`analytics-event.txt`](./assets/phase-4/analytics-event.txt) → HTTP 204 |
| `POST /api/v1/reports` | [`report-create.txt`](./assets/phase-4/report-create.txt) → HTTP 204 |
| `GET /api/v1/admin/sources/health` | [`admin-health.json`](./assets/phase-4/admin-health.json) |
| `GET /api/v1/admin/matches/at-risk` | [`admin-at-risk.json`](./assets/phase-4/admin-at-risk.json) |
| `GET /api/v1/admin/reports` | [`admin-reports.json`](./assets/phase-4/admin-reports.json) — open report visible |

## 3. Tests proving the phase

| Gate | Result | Log |
|------|--------|-----|
| Go feature packages (`privacy`, `admin`, `reports`, `auth`, `preferences`, `push`) | **pass** — 35 `--- PASS` lines | [`phase4-go-tests.log`](./assets/phase-4/phase4-go-tests.log) |
| Web unit (`pnpm test`) | **pass** — 14 files / 40 tests | [`phase4-web-unit.log`](./assets/phase-4/phase4-web-unit.log) |
| Playwright public e2e (TEST-008/010 + phase4 API) | **pass** — 4/4 | [`e2e-public.log`](./assets/phase-4/e2e-public.log) |
| CI required-ci on prior tip `2d97005` | **pass** | [run 31295151575](https://github.com/IcaroAguiar/central-do-jogo/actions/runs/31295151575); see [`latest-run.md`](./latest-run.md) |

## 4. How this was reproduced

```bash
# Postgres on :5433, migrate via server/seed
export DATABASE_URL='postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable'
go run ./cmd/seed
cd web && pnpm build && cd ..
# auth-enabled server (demo secrets only)
GOOGLE_OAUTH_CLIENT_ID=… GOOGLE_OAUTH_CLIENT_SECRET=… \
  SESSION_COOKIE_SECRET='phase4-evidence-session-secret-32bytes' \
  MAINTAINER_ALLOWLIST='maintainer@example.com' AUTH_COOKIE_SECURE=false \
  PUBLIC_BASE_URL=http://127.0.0.1:8080 STATIC_DIR=web/dist \
  go run ./cmd/server
# insert demo maintainer session (sha256 of raw token) then curl + Playwright shots
cd web && E2E_BASE_URL=http://127.0.0.1:8080 pnpm e2e
```

## Residuals (not claimed here)

Browser e2e for full Google login, admin journey automation, privacy journey
automation, and Push-simulated delivery remain **TASK-035** (see
[`phase-4-checklist.md`](./phase-4-checklist.md)).
