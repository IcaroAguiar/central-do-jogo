# ADR 0002: Package layout and module boundaries

- Status: Accepted
- Date: 2026-08-09
- Deciders: Icaro Aguiar

## Context

The modular monolith already has vertical features under `internal/features/` and `web/src/features/`, but persistence lived under `internal/platform/store`, feature ports returned store record types, `internal/api` collided with the top-level OpenAPI tree `/api`, and HTTP wiring grew a field-per-handler bag. Phase 4 (privacy, admin, reports) would widen that debt without explicit boundaries.

## Decision

### Go layout

| Package | Role |
|---------|------|
| `internal/domain/` | Domain types and read models used by feature ports |
| `internal/features/<name>/` | Use case + HTTP handler + unit tests for one feature |
| `internal/store/` | PostgreSQL persistence implementing ports |
| `internal/platform/` | Cross-cutting only: config, HTTP kernel, logging, render, database, ratelimit, brasilia |
| `internal/httpapi/` | Shared contract DTOs / SSR page constants (not OpenAPI YAML) |
| `internal/sources/`, `jobs/`, `reconciliation/` | Adapters, job queue, reconciliation as today |
| `cmd/` | Thin composition root and process entrypoints |

OpenAPI YAML stays at `api/openapi.yaml`. Generated TS types stay at `web/src/api/generated/`.

SSR HTML templates remain in `web/server-templates/` (Go `embed` consumed by `platform/render`); do not relocate without a follow-up ADR.

### Import rules

- Features must not import other features' handlers or services. Depend on local ports; wire implementations in `cmd/` (or a small `internal/app` helper).
- Features must not import `internal/store` for response/port types. Port return types live in `domain` (or feature-local DTOs). `store` maps rows to those types.
- `platform` must not import features.
- Auth/session/maintainer access crosses features only via ports (`SessionResolver`, `MaintainerGate`), not by importing `features/auth`.

### HTTP composition

Each feature exposes a `Register(mux *http.ServeMux, …)` (or equivalent) so the HTTP kernel mounts health/static/SSR shell and delegates API routes to feature registrars. Avoid growing a monolithic handler-slot struct as the primary extension point.

### Web layout

| Path | Role |
|------|------|
| `web/src/features/<name>/` | Feature UI, hooks, route pages, and Vitest files |
| `web/src/pages/` | Generic shells only (`HomePage`, `NotFoundPage`, `LoadErrorPage`) |
| `web/src/lib/` | Domain-agnostic helpers |
| `web/src/test/` | Shared Vitest harness only |

Phase 4 naming: Go package `privacy` pairs with web feature `settings` (account UI). Do not add a parallel `web/src/features/privacy/`. `admin` and `reports` use the same name on both sides.

### Test layers

| Layer | Location | Scope |
|-------|----------|-------|
| Unit | Co-located (`*_test.go`, `*.test.ts(x)`) | Pure logic and handlers with fakes |
| Persistence | `internal/store/*_test.go` (+ `jobs`) | SQL, constraints, leases |
| Contract | OpenAPI CI + `sources/*/adapter_test.go` fixtures | No live network |
| E2E public | `e2e/public/tests/` | Visitor journeys |
| E2E admin | `e2e/admin/` | Maintainer journeys (full suite TASK-035) |

Do not create mirrored trees such as `internal/features/foo/tests/` or `web/tests/` that duplicate package layout.

### Config

Group `platform/config` by concern (`Auth`, `Push`, `Privacy`, `Admin`, `Reports`, …) without renaming existing environment variables.

## Consequences

- Moving `platform/store` → `store` and `internal/api` → `httpapi` is a mechanical import update for existing code.
- New Phase 4 features must follow these paths and ports from day one.
- Contributors review import direction in PRs; no compile-time architecture linter is required for MVP.
- FILE-* entries in `docs/product/initial-plan.md` track the same layout.
