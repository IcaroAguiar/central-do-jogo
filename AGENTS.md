# Central do Jogo

## Project

Central do Jogo is an open-source, low-cost pre-match PWA for Brazilian football. Its core promise is to show where to watch, official lineups, and related source links with explicit provenance and confidence.

## Current phase

- The project is in source-feasibility research.
- Do not implement product code before the source matrix is documented and accepted.
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

Go modular monolith, React + Vite PWA, PostgreSQL, REST + OpenAPI, and Docker Compose. The stack remains a planning constraint until source feasibility is proven.

## Cursor Cloud specific instructions

- This repository is documentation-only right now (Phase 0, source-feasibility research). There is no application, dependency manifest, test suite, linter, or build system yet, and that is intentional — see `docs/product/initial-plan.md`. Do not scaffold product foundation (`go.mod`, `web/`, `cmd/`, `internal/`, Docker Compose, etc.) as part of environment setup; that is Phase 1 (GOAL-002) and is gated behind acceptance of `docs/research/source-matrix.md` (GOAL-001).
- "Running the project" today means working with the docs under `docs/`. There is no server, no dev command, and no port to open. A quick sanity check is validating Markdown (UTF-8, headings, and the YAML front-matter in `docs/product/initial-plan.md`).
- Toolchain preinstalled on the VM for the planned stack: Go 1.22, Node 22, pnpm 10, npm, Python 3.12, and GNU Make. Docker and `psql` are NOT installed; install them only once Phase 1 introduces code that actually needs PostgreSQL and Docker Compose.
- The startup update script is intentionally a guarded no-op today: it installs Go modules only if `go.mod` exists and frontend deps (via pnpm) only if `web/package.json` exists. Once Phase 1 lands those files, the same script starts installing dependencies automatically with no changes needed.
