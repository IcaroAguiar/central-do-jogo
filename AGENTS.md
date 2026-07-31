# Central do Jogo

## Project

Central do Jogo is an open-source, low-cost pre-match PWA for Brazilian football. Its core promise is to show where to watch, official lineups, and related source links with explicit provenance and confidence.

## Current phase

- The project is in source-feasibility research.
- Do not implement product code before the source matrix is documented and accepted.
- Start durable project documentation under `docs/`; keep decisions separate from unverified assumptions.

## Working rules

- Preserve unrelated user changes and never invent execution evidence.
- Prefer public structured data; use isolated HTML adapters only when reviewed and necessary.
- Every source adapter must have a manifest, redacted fixtures, and deterministic tests. CI must not depend on live sources.
- Never commit secrets, tokens, cookies, credentials, private data, or unlicensed third-party assets.
- Product and user documentation use pt-BR. Code, contracts, and contributor-facing technical documentation use English.
- Keep the MVP focused on pre-match data. AI, live scores, match events, statistics, and post-match features are out of scope.

## Planned stack

Go modular monolith, React + Vite PWA, PostgreSQL, REST + OpenAPI, and Docker Compose. The stack remains a planning constraint until source feasibility is proven.
