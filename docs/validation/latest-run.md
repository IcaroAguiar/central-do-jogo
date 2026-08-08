# Latest QA smoke receipt

- **Date:** 2026-08-08
- **Branch / commit:** `chore/qa-evidence-pipeline` (6369688)
- **Agent:** local verification for the QA evidence bridge
- **Stack:** Compose `db` on `127.0.0.1:5433`, `go run ./cmd/server` on `:8080`, Playwright Chromium

## Commands

| Gate | Command | Result |
|------|---------|--------|
| Go format | `test -z "$(gofmt -l ./cmd ./internal)"` | pass |
| Go vet | `GOTOOLCHAIN=local go vet ./...` | pass |
| Go tests | `GOTOOLCHAIN=local go test -race -count=1 ./...` (clean env, no `PUBLIC_BASE_URL`) | **pass** — 183 tests / 30 packages |
| Web lint | `./node_modules/.bin/biome ci .` in `web/` | **pass** — 62 files |
| Web typecheck | `cd web && pnpm typecheck` | pass |
| Web unit | `cd web && pnpm test` | **pass** — 11 files / 36 tests |
| Web build | `cd web && pnpm build` | pass |
| OpenAPI | `npx @redocly/cli@1.34.5 lint api/openapi.yaml` | **pass** — valid, 5 existing warnings |
| E2E public | seed + server + `cd web && pnpm e2e` | **pass** — 4/4 (smoke, offline, phase4 auth/me, phase4 vapid); re-run after option-button click harden also 4/4 |

## E2E detail

```
✓ phase4-api › auth/me reports configuration without requiring a session
✓ phase4-api › push vapid endpoint responds according to VAPID configuration
✓ offline resilience › shows cached club data and an offline banner…
✓ public smoke › search → club → match → share
4 passed (11.3s)
```

## Residuals

- Auth / Push browser e2e still residual (TEST-011 / Push part of TEST-008) until TASK-035 or dedicated test credentials.
- Postgres store integration tests still require a live DB locally; CI e2e brings its own Postgres service.
- Beta denominators (TEST-015) remain Phase 5.
- Local note: exporting `PUBLIC_BASE_URL` while running `go test` fails `TestLoadDefaults` (env pollution); CI Go job does not set that var.
