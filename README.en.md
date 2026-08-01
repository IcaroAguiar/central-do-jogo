# Central do Jogo

Open-source pre-match PWA for Brazilian football: where to watch, official lineups, and related links with explicit provenance and confidence.

> Status: technical foundation (scaffold) is in place. Product implementation stays blocked until source-feasibility research (Phase 0) is accepted.

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

# Worker (stub)
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

## Docs

- Product (pt-BR): [`docs/product/`](docs/product/)
- ADRs: [`docs/adr/`](docs/adr/)
- Deploy: [`deploy/README.md`](deploy/README.md)
- Portuguese entry: [`README.md`](README.md)

## License

Apache License 2.0. Club crests are not distributed in this repository.
