# Central do Jogo

PWA open source de pré-jogo para o futebol brasileiro: onde assistir, escalações oficiais e links com proveniência e confiança explícitas.

> Status: fundação técnica (scaffold) pronta. Implementação de produto permanece bloqueada até a pesquisa de viabilidade de fontes (Fase 0).

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

## Documentação

- Produto: [`docs/product/`](docs/product/)
- ADRs: [`docs/adr/`](docs/adr/)
- Deploy: [`deploy/README.md`](deploy/README.md)
- English entry: [`README.en.md`](README.en.md)

## Licença

Apache License 2.0. Escudos de clubes não são distribuídos neste repositório.
