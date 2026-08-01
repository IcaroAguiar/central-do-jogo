# ADR 0000: Toolchain versions

- Status: Accepted
- Date: 2026-07-31
- Deciders: Icaro Aguiar

## Context

The project needs pinned, reproducible toolchains for an open-source modular monolith (Go API/worker, React + Vite PWA, PostgreSQL, Docker Compose). Versions were researched against current stable/LTS lines on 2026-07-31 before scaffolding the empty foundation.

An intentional sequencing exception allows this foundation scaffold before Phase 0 source feasibility (`GOAL-001`). Product domain code, adapters, and feature work remain blocked until the source matrix and `docs/adr/0001-source-feasibility-gate.md` are accepted.

## Decision

| Layer | Pin | Rationale |
|-------|-----|-----------|
| Go language | `go 1.26.0` | Application module (not a shared library); current stable major line. |
| Go toolchain | `toolchain go1.26.5` | Latest 1.26 patch for reproducible builds; older local installs auto-download via the Go toolchain switcher. |
| golangci-lint | `v2.12.x` (CI action) | Go 1.26 support since v2.9.0; use official release binaries. |
| HTTP router | `net/http.ServeMux` | Stdlib method/path routing since Go 1.22; zero extra dependency. Revisit chi only if route grouping becomes painful. |
| Node.js | `^24.18.1` (Active LTS Krypton) | Prefer Active LTS over Current 26 for contributors and CI stability. |
| pnpm | `11.18.0` via `packageManager` | Current stable major; Corepack/CI reproducibility. |
| Vite | `8.2.0` | Current stable; Rolldown-based production builds. |
| `@vitejs/plugin-react` | `6.x` | Matched to Vite 8. |
| React | `19.2.8` | Current stable. |
| TypeScript | `6.0.3` | TypeScript 7.0 ships a native `tsc` but lacks a stable programmatic API for many tools; stay on 6.x until the ecosystem is ready (~7.1). |
| Vitest | `4.x` | Aligned with Vite 8 ecosystem. |
| Biome | `2.5.6` | Project standard for frontend lint/format; replaces Vite-template ESLint. |
| PWA | `vite-plugin-pwa` `1.3.0` + minimal web manifest | Enough for scaffold; full offline product behavior comes later. |
| PostgreSQL (Compose) | `postgres:17.10` | Matches local host major (17.x); community support through 2029-11. Avoid PostgreSQL 18 image volume/`PGDATA` layout change during scaffold. |
| OpenAPI | minimal internal `/healthz` contract | No third-party compatibility SLA (CON-005). |
| License | Apache-2.0 | CON-007. |

Module path: `github.com/IcaroAguiar/central-do-jogo`.

## Consequences

- Contributors need Node 24 LTS and pnpm 11 (or Corepack to fetch it).
- TypeScript stays on 6.x until ecosystem tooling supports TypeScript 7.1+.
- Frontend lint/format is Biome-only under `web/`; do not reintroduce ESLint/Prettier without a new ADR.
- PostgreSQL 18 upgrade remains a deliberate follow-up with volume migration notes.
- Empty foundation may exist before source research; product implementation must not expand past health/shell until `ADR 0001` accepts the source gate.

## Follow-ups

- Revisit TypeScript 7 after 7.1 lands programmatic API support.
- Revisit PostgreSQL 18 when Compose volume migration is planned.
- Revisit chi (or equivalent) only if stdlib ServeMux grouping becomes a maintenance cost.
- Defer a pnpm workspace / monorepo layout until multi-package Node needs appear; keep the single `web/` package for now.
