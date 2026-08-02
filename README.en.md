# Central do Jogo

Open-source pre-match PWA for Brazilian football: where to watch, official lineups, and related links with explicit provenance and confidence.

> Status: Phase 2 done (domain, persistence, jobs queue). Phase 3 in progress (search, club agenda, match detail, SSR, and sample data seeding). Allowlist: conditional Serie A (ADR 0001).

## Local requirements

- Go `1.26.x` (toolchain pinned to `go1.26.5`)
- Node.js `24.18.1` (Active LTS) and pnpm `11.18.0`
- Docker / Docker Compose (Postgres and full image)

See [`docs/adr/0000-toolchain-versions.md`](docs/adr/0000-toolchain-versions.md).

## Quick start

```bash
# API
cp .env.example .env
go run ./cmd/server

# Worker
go run ./cmd/worker

# Frontend
cd web
pnpm install
pnpm dev
```

Compose (API + worker + PostgreSQL 17.10):

```bash
docker compose -f deploy/compose.yaml up --build
```

Probe: `GET http://127.0.0.1:8080/healthz`

### Seed sample data

After running migrations, populate the database with Serie A clubs, one
competition, and a varied set of matches for local development (idempotent;
this does not replace the ingest adapters, which remain a no-op in this
phase):

```bash
go run ./cmd/seed
```

This enables the public read routes: `GET /api/v1/search?q=...`,
`GET /api/v1/clubs/{slug}`, `GET /api/v1/clubs/{slug}/matches`,
`GET /api/v1/matches/{slug}`, plus the server-rendered pages at `/`,
`/clubes/{slug}`, and `/jogos/{slug}`. Full contract in
[`api/openapi.yaml`](api/openapi.yaml).

## Docs

- Product (pt-BR): [`docs/product/`](docs/product/)
- ADRs: [`docs/adr/`](docs/adr/)
- Deploy: [`deploy/README.md`](deploy/README.md)
- Portuguese entry: [`README.md`](README.md)

## License

Apache License 2.0. Club crests are not distributed in this repository.
