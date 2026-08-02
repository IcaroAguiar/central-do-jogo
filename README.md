# Central do Jogo

PWA open source de pré-jogo para o futebol brasileiro: onde assistir, escalações oficiais e links com proveniência e confiança explícitas.

> Status: Fase 2 concluída (domínio, persistência e fila de jobs). Fase 3 em curso (busca, agenda do clube, detalhe de partida, SSR e seed de dados de exemplo). Allowlist: Série A condicional (ADR 0001).

## Requisitos locais

- Go `1.26.x` (toolchain pinado em `go1.26.5`)
- Node.js `24.18.1` (Active LTS) e pnpm `11.18.0`
- Docker / Docker Compose (para Postgres e imagem completa)

Detalhes em [`docs/adr/0000-toolchain-versions.md`](docs/adr/0000-toolchain-versions.md).

## Desenvolvimento rápido

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

### Seed de dados de exemplo

Após aplicar as migrações, popule o banco com clubes da Série A, uma
competição e um conjunto variado de partidas para desenvolvimento local
(idempotente; não substitui os adapters de ingestão, que permanecem no-op
nesta fase):

```bash
go run ./cmd/seed
```

Isso habilita as rotas públicas: `GET /api/v1/search?q=...`,
`GET /api/v1/clubs/{slug}`, `GET /api/v1/clubs/{slug}/matches`,
`GET /api/v1/matches/{slug}`, além das páginas renderizadas no servidor `/`,
`/clubes/{slug}` e `/jogos/{slug}`. Contrato completo em
[`api/openapi.yaml`](api/openapi.yaml).

## Documentação

- Produto: [`docs/product/`](docs/product/)
- ADRs: [`docs/adr/`](docs/adr/)
- Deploy: [`deploy/README.md`](deploy/README.md)
- English entry: [`README.en.md`](README.en.md)

## Licença

Apache License 2.0. Escudos de clubes não são distribuídos neste repositório.
