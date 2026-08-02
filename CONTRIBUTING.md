# Contributing

Thanks for considering a contribution to Central do Jogo.

## Before you start

- Product domain work must stay inside the Phase 0 allowlist (`docs/adr/0001-source-feasibility-gate.md`, `AGENTS.md`).
- Prefer public structured data. Isolated HTML adapters need review, manifests, redacted fixtures, and deterministic tests.
- Do not commit secrets, tokens, cookies, credentials, private data, or unlicensed third-party assets.

## Language

- Product and user-facing docs: Brazilian Portuguese (pt-BR).
- Code, contracts, ADRs, and contributor technical docs: English.
- Keep the bilingual README pair (`README.md` / `README.en.md`) concise.

## Development

1. Install Go 1.26.x, Node 24.18.1 (Active LTS), and pnpm 11.18.0.
2. Copy `.env.example` to `.env` and adjust only local non-secret values.
3. Run checks:

```bash
go test -race ./...
go vet ./...
cd web && pnpm install && pnpm lint && pnpm typecheck && pnpm test && pnpm build
```

Frontend lint/format uses Biome (`web/biome.json`). Apply fixes locally with `pnpm format`.

4. Optional full stack: `docker compose -f deploy/compose.yaml up --build`
5. Optional public-journey e2e smoke (Playwright, needs the full stack running and seeded): see `e2e/public/README.md`.

## Pull requests

- Use semantic branches (`feat/`, `fix/`, `chore/`, `docs/`, `refactor/`).
- Keep diffs small and reviewable.
- Explain behavior covered by tests; list residual risk when checks are skipped.
- Wait for CI (Go, web, OpenAPI) to pass.
