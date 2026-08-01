# Central do Jogo

## Project

Central do Jogo is an open-source, low-cost pre-match PWA for Brazilian football. Its core promise is to show where to watch, official lineups, and related source links with explicit provenance and confidence.

## Current phase

- Empty technical foundation (scaffold) is allowed and present: Go server/worker stubs, React/Vite shell, OpenAPI health contract, Compose, and CI.
- Product implementation remains blocked until the source matrix is documented and accepted (`GOAL-001` / `docs/adr/0001-source-feasibility-gate.md`).
- Next delivery work is Phase 0 research under full open-source governance (issue + branch + PR + review).
- Start durable project documentation under `docs/`; keep decisions separate from unverified assumptions.
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

Go modular monolith, React + Vite PWA, PostgreSQL, REST + OpenAPI, and Docker Compose. Toolchain pins live in `docs/adr/0000-toolchain-versions.md`. Product features still wait on source feasibility.
