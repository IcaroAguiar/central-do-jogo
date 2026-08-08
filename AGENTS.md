# Central do Jogo

## Project

Central do Jogo is an open-source, low-cost pre-match PWA for Brazilian football. Its core promise is to show where to watch, official lineups, and related source links with explicit provenance and confidence.

## Current phase

- Phase 0 source feasibility is documented: see `docs/research/source-matrix.md` and `docs/adr/0001-source-feasibility-gate.md`.
- GOAL-001 is **accepted with a conditional Serie A allowlist**. Domain/adapters may start only inside that allowlist.
- Phase 2 / GOAL-003 is **done**: domain types, PostgreSQL migrations, structured logging, jobs queue.
- Phase 3 / GOAL-004 is **done**: public read journeys (search, club detail/agenda, match detail), SSR, PWA offline shell, and `cmd/seed` demo data (ingest adapters remain a no-op placeholder).
- Phase 4 / GOAL-005 is in progress: OAuth (TASK-027), preferences (TASK-028), and Web Push (TASK-029) done; privacy, admin, and reports follow.
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

## Planned stack

Go modular monolith, React + Vite PWA, PostgreSQL, REST + OpenAPI, and Docker Compose. Toolchain pins live in `docs/adr/0000-toolchain-versions.md`. Implementation must respect the Phase 0 allowlist in ADR 0001.
