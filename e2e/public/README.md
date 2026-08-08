# Public journey e2e smoke (Playwright)

Covers the public read journeys end to end against a real, built, running
stack: search → club agenda → match detail → share (TEST-008 smoke), offline
resilience (TEST-010), and Phase 4 API readiness probes (auth/push config
surfaces without requiring live Google/VAPID secrets).

These tests intentionally do **not** start their own server: they need the
real Go SSR pages, the real `/api/v1/*` responses, and a real installed
service worker, none of which the plain Vite dev server provides (the PWA
plugin's `devOptions.enabled` is `false` on purpose — see
`web/vite.config.ts`).

## Running locally

1. Start Postgres and the API (either Docker Compose or a local Postgres):

   ```bash
   cp .env.example .env
   docker compose -f deploy/compose.yaml up --build
   ```

2. Seed sample data so the agenda/broadcasts/lineups/news sections have
   something to render:

   ```bash
   export DATABASE_URL=postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable
   go run ./cmd/seed
   ```

3. Build the web assets so the Go server has `web/dist/app.js` / `app.css`
   and the compiled service worker to serve (Compose already does this in
   its image build; for a bare `go run ./cmd/server` instead, build once):

   ```bash
   cd web && pnpm install && pnpm build
   ```

4. Install Playwright's browser binaries once, then run the suite:

   ```bash
   cd web && pnpm exec playwright install --with-deps chromium
   pnpm e2e
   ```

   By default this targets `http://127.0.0.1:8080` (the Go server). Point
   it elsewhere with `E2E_BASE_URL`, e.g. against a deployed preview:

   ```bash
   E2E_BASE_URL=https://preview.example.com pnpm e2e
   ```

## CI

The `e2e-public` job runs on every `pull_request` and every push to `main`
(and on `workflow_dispatch`). It boots Postgres, seeds data, builds the web
app, starts `cmd/server`, and uploads the Playwright HTML report as an
artifact. See `docs/validation/` for the evidence contract.
