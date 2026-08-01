# ADR 0001: Source feasibility gate

- Status: Accepted (conditional allowlist)
- Date: 2026-08-01
- Deciders: Icaro Aguiar
- Issue: https://github.com/IcaroAguiar/central-do-jogo/issues/2

## Context

GOAL-001 requires proving that free, admissible sources can support the pre-match MVP promise before product domain work expands. Research artifacts:

- `docs/research/schedules.md`
- `docs/research/broadcasts.md`
- `docs/research/lineups.md`
- `docs/research/news.md`
- `docs/research/source-matrix.md`
- redacted probes under `docs/research/evidence/`

Constraints: CON-001 (no mandatory paid services), SEC-004 (source manifests), PAT-003 (fixtures/tests, no live CI), REQ-023 / REQ-024 beta gates.

## Decision

**Accept GOAL-001 with a reduced, conditional allowlist.**

Product domain work (Phase 2+) may begin **only** for:

1. Serie A clubs in the official 2026 snapshot
2. Schedule ingest centered on CBF public HTML/PDF evidence (openfootball secondary/non-authoritative)
3. Lineups centered on CBF match-center HTML, with editorial sources marked non-official
4. News link cards from allowlisted RSS (starting with Gazeta Esportiva)
5. Broadcasts as **candidate** claims from public editorial pages, with low confidence until maintainer confirmation

Explicitly **out of the first implementation slice** until their matrix rows are upgraded:

- Copa do Brasil, Libertadores, Sudamericana as fully supported competitions
- Any paid API as a required dependency
- Club social image OCR as an ingest path

### REQ-023 / REQ-024

- **REQ-024:** Plausible if “official observable” means CBF match-center publication time.
- **REQ-023:** Plausible only as a human-assisted pipeline. Do not promise push-quality broadcast accuracy before beta measurement and maintainer tooling exist.

This ADR does **not** waive the beta numeric gates; it only unlocks building toward them.

## Consequences

- Update `AGENTS.md`: Phase 0 research complete; domain implementation may start within the allowlist above.
- Every future adapter must ship manifest + redacted fixtures + deterministic tests.
- If broadcast accuracy cannot be staffed, ship schedules/lineups/news first and keep broadcasts in `awaiting_publication` / low-confidence states without notifications.

## Alternatives considered

- **Block GOAL-001 entirely** because broadcasts lack a structured free API — rejected; schedules/lineups/news have enough path to learn, and broadcasts can degrade gracefully.
- **Require paid sports API** — rejected under CON-001 for the default OSS deployment.
- **Full four-competition launch** — rejected until continental/national cup adapters are evidenced at the same depth as Serie A.
