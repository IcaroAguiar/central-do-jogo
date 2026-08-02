# Deploy notes

## Local Compose

Compose is for **local development only**. Do not reuse these defaults in production.

```bash
cp .env.example .env
docker compose -f deploy/compose.yaml up --build
```

- API and static PWA: http://127.0.0.1:8080
- Health probe: `GET /healthz`
- PostgreSQL: `postgres:17.10-alpine` on host port `5433` by default (avoids clashing with a local Postgres on `5432`)
- Default DB password in Compose/`.env.example` is a non-secret local placeholder (`central_dev_only`) with `sslmode=disable`. Replace before any shared or remote environment.

The API image is distroless and has no shell/curl. Use an external probe (reverse proxy, Compose `curl` sidecar later, or `curl` from the host) against `/healthz`.

## Image layout

Multi-stage `deploy/Dockerfile`:

1. Node 24 + pnpm 11 builds `web/dist`
2. Go 1.26.5 builds `server` and `worker`
3. Distroless static runtime serves the PWA from `STATIC_DIR` (default `/app/web/dist`)

No Node runtime is present in production.

## Migrations

SQL migrations live in `db/migrations/` and are embedded via package `db`.

The API and worker apply migrations on startup (`DATABASE_URL` is required).

```bash
export DATABASE_URL=postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable
go run ./cmd/server
```

## Seed data (development only)

`cmd/seed` applies migrations and then idempotently upserts sample Serie A
clubs, one competition, and a varied set of matches (covering kickoff
states, availability gap states, and low/medium-confidence broadcasts,
lineups, and news) so the public read journeys (`/api/v1/search`,
`/api/v1/clubs/{slug}`, `/api/v1/clubs/{slug}/matches`,
`/api/v1/matches/{slug}`, and the SSR pages) have something to render
locally:

```bash
export DATABASE_URL=postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable
go run ./cmd/seed
```

Seeding is a local/demo convenience and is safe to re-run. It is not a
substitute for the source ingest pipeline (`internal/sources/`, `cmd/worker`),
which remains a no-op placeholder until ingest handlers are wired (outside
GOAL-004 acceptance).

## Search rate limiting (SEC-001)

`GET /api/v1/search` uses an in-process token bucket keyed by client IP.

When the API sits behind a reverse proxy, set `TRUSTED_PROXY_CIDRS` to the
proxy CIDRs/IPs so `X-Forwarded-For` is honored. Leaving it empty (default)
ignores forwarded headers and keys on `RemoteAddr` — correct for direct
exposure, but collapses every user into one bucket if an unconfigured proxy
terminates TLS in front of the process.

