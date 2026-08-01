# Broadcast source research

Evaluation date: 2026-08-01 (America/Sao_Paulo)  
Focus: REQ-023 path (≈97% accuracy / ≈90% coverage in beta) for “where to watch”.

## Rights landscape (Serie A 2026)

Public editorial/Wikipedia summaries describe a split across Globo ecosystem (TV Globo, SporTV, Premiere, Ge TV), Record, CazéTV, and Prime Video, depending on club block (Libra / FFU). Rights are match-specific and change near kickoff.

This pass does **not** treat Wikipedia as an operational source; it is background only.

## Candidates

### `ge_globo_editorial_onde_assistir`

| Field | Value |
|-------|-------|
| Owner | Grupo Globo (ge.globo) |
| Example URLs | see `evidence/broadcasts/*.ge-editorial.redacted.json` |
| Access | public HTTP |
| robots | agenda/content paths generally allowed; `/servico` and `/dynamo` disallowed |
| Format | HTML editorial (“onde assistir … escalações”) |
| Fields observed | channel/platform names in prose (`Premiere`, `SporTV`, sometimes `Prime Video`) |
| Missing vs REQ-007 | stable free-vs-pay flag, region, official deep link often absent |
| Evidence | three redacted captures under `evidence/broadcasts/` |
| Decision | **conditional** |

Rationale: Best free public signal found for per-match Brazilian “where to watch” text. Not a structured rights API. Suitable as candidate adapter **only** with strong provenance labels, conservative confidence, and maintainer confirmation for high-risk matches.

### `gzh_editorial` (sample)

| Field | Value |
|-------|-------|
| URL probed | GaúchaZH round-1 “onde assistir” article (HTTP 200) |
| Format | HTML listing kickoff + channels (Premiere/SporTV/etc.) |
| Decision | **conditional** secondary editorial |

### Official competition / club pages

| Field | Value |
|-------|-------|
| Observation | No free, comprehensive, structured official rights feed found in this pass |
| Decision | **rejected** as sole automated MVP source |

### Paid sports data APIs

| Field | Value |
|-------|-------|
| Decision | **rejected** as mandatory dependency (CON-001) |

## REQ-023 feasibility

| Question | Assessment |
|----------|------------|
| Can channel names be collected legally from public HTML? | Plausible for editorial pages with robots-aware crawling and attribution |
| Can 90% coverage be reached for Serie A? | Only with multi-source editorial coverage + gaps marked `awaiting_publication` / `not_found` |
| Can 97% accuracy be reached without humans? | **Unlikely.** Rights copy is noisy and changes late |
| Plausible MVP path? | **Yes, conditional:** automated editorial candidates + maintainer panel (REQ-013) before notifying users; beta metrics may fail if ops capacity is insufficient |

## Adapter implications

1. Store every broadcast claim with source URL, observed text, and verification timestamp.
2. Default confidence low for pay TV / exclusives until confirmed.
3. Never notify on unconfirmed oscillations (REQ-012).
4. Prefer not launching push-for-broadcast until beta accuracy is measured.
