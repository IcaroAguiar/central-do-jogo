# Central do Jogo

## Project

Central do Jogo is an open-source, low-cost pre-match PWA for Brazilian football. Its core promise is to show where to watch, official lineups, and related source links with explicit provenance and confidence.

## Current phase

- Phase 0 source feasibility is documented: see `docs/research/source-matrix.md` and `docs/adr/0001-source-feasibility-gate.md`.
- GOAL-001 is **accepted with a conditional Serie A allowlist**. Domain/adapters may start only inside that allowlist.
- Phase 2 / GOAL-003 is **done**: domain types, PostgreSQL migrations, structured logging, jobs queue.
- Phase 3 / GOAL-004 is **done**: public read journeys (search, club detail/agenda, match detail), SSR, PWA offline shell, and `cmd/seed` demo data (ingest adapters remain a no-op placeholder).
- Phase 4 / GOAL-005 is **done on `main`** (PR #16): privacy/admin/reports, ADR 0002 package layout, evidence pack under `docs/validation/phase-4-evidence-pack.md`. Checklist residuals remain for TASK-035.
- Next: Phase 5 / GOAL-006 (beta protocol, full Playwright, a11y/perf, backup/ops, release gate).
- QA evidence contract lives in `docs/validation/`; public Playwright smoke is required CI on PR/`main` (TEST-008/TEST-010).
- Broadcasts remain human-assisted for accuracy; Copa do Brasil / Libertadores / Sudamericana are deferred pending deeper source rows.
- Keep contributor-useful product, research, and architecture documents public. Put personal drafts or non-redistributable evidence only in the ignored private-document paths described in `docs/README.md`.

## Working rules

- Preserve unrelated user changes and never invent execution evidence.
- Prefer public structured data; use isolated HTML adapters only when reviewed and necessary.
- Every source adapter must have a manifest, redacted fixtures, and deterministic tests. CI must not depend on live sources.
- Never commit secrets, tokens, cookies, credentials, private data, or unlicensed third-party assets.
- Treat `.gitignore` as accident prevention, not secret storage; sensitive credentials must never be written to the repository tree.
- Product and user documentation use pt-BR. Code, contracts, and contributor-facing technical documentation use English.
- Keep the MVP focused on pre-match data. AI, live scores, match events, statistics, and post-match features are out of scope.
- Respect package boundaries in `docs/adr/0002-package-layout.md`: features use ports (no cross-feature imports, no `store` response types in ports); persistence lives in `internal/store/`; web product UI under `web/src/features/` (`privacy` ↔ `settings`).

## Planned stack

Go modular monolith, React + Vite PWA, PostgreSQL, REST + OpenAPI, and Docker Compose. Toolchain pins live in `docs/adr/0000-toolchain-versions.md`. Package layout lives in `docs/adr/0002-package-layout.md`. Implementation must respect the Phase 0 allowlist in ADR 0001.

## Cursor Cloud specific instructions

The Cloud Agent environment already installs toolchains (Go 1.26.5, Node 24.18.1 via nvm, pnpm 11.18.0, PostgreSQL 17, Playwright Chromium), refreshes deps, builds `web/dist`, starts Postgres, seeds the dev DB, and launches the API. Canonical dev/build/test commands live in `README.md`, `CONTRIBUTING.md`, and `deploy/README.md`; the notes below are only the non-obvious gotchas.

- **Node PATH gotcha.** A bundled Node (v22) at `/exec-daemon/node` is force-prepended to `PATH` on every command, so a bare `node`/`pnpm` is the wrong version. Prefix any Node/pnpm/Playwright command with `export PATH="$HOME/.nvm/versions/node/v24.18.1/bin:$PATH"` to get the pinned Node 24.18.1 + pnpm 11.18.0.
- **PostgreSQL runs as a local cluster, not Docker.** It listens on port **5433** (matching `.env.example`). Start/reconcile with `sudo pg_ctlcluster 17 main start`; check with `pg_isready -h 127.0.0.1 -p 5433`. Docker/Compose is not available in this environment.
- **Integration tests wipe the database in `DATABASE_URL`.** The `internal/store/...` and `db/...` tests self-migrate and `DELETE` every row of whatever `DATABASE_URL` points at. Run them against the dedicated `central_test` database (`DATABASE_URL=postgres://central:central_dev_only@127.0.0.1:5433/central_test?sslmode=disable go test -race ./...`) so the seeded `central_do_jogo` dev DB survives. Re-run `go run ./cmd/seed` if you ever point tests at the dev DB.
- **Do not export app config env vars globally.** `internal/platform/config` `TestLoadDefaults` asserts unset defaults (e.g. `PUBLIC_BASE_URL`), so scope `DATABASE_URL`/`PUBLIC_BASE_URL`/etc. to the server/worker/seed commands only. Config reads env vars directly and does **not** auto-load `.env`, so exported vars are required for `go run` processes.
- **The Go server on `:8080` is the real app surface** (SSR pages + built PWA + `/api/v1/*`). `pnpm dev` (Vite, `:5173`) is only for frontend HMR and proxies `/api` and `/healthz` to `:8080`; it deliberately does not serve SSR or the service worker, so e2e/offline flows must target `:8080`.
