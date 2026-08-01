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
