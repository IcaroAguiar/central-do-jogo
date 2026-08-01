# Broadcast source research

Evaluation date: 2026-08-01 (America/Sao_Paulo)  
Focus: REQ-023 path (≈97% accuracy / ≈90% coverage in beta) for “where to watch”.

## Rights landscape (Serie A 2026)

Public editorial/Wikipedia summaries describe a split across Globo ecosystem (TV Globo, SporTV, Premiere, Ge TV), Record, CazéTV, and Prime Video, depending on club block (Libra / FFU). Rights are match-specific and change near kickoff.

This pass does **not** treat Wikipedia as an operational source; it is background only.

## Critical failure mode: incomplete channel lists

Editorial “onde assistir” pages are **not exhaustive**. Observed and product-risk patterns:

1. **Competitor omission:** Globo-family pages often emphasize Premiere / SporTV / Globo and may **omit or under-mention** CazéTV, Prime Video, Record, or other holders even when those platforms carry the match.
2. **Partial prose:** a page may mention one live carrier (“Premiere transmite ao vivo”) while the same round’s rights map includes additional free or exclusive outlets.
3. **False completeness:** scraping a single editorial and rendering it as “todas as transmissões” would violate REQ-007 and destroy REQ-023 accuracy via **false negatives** (missing channels), not only false positives.

**Product rule:** automated editorial claims are **candidate / incomplete by default**. The UI must not imply an exhaustive list until a maintainer confirms, or until multiple independent sources agree on the full set. Missing a free outlet (e.g. CazéTV) is as bad as inventing one.

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
| Missing vs REQ-007 | stable free-vs-pay flag, region, official deep link often absent; **lists frequently incomplete vs known rights map** |
| Evidence | three redacted captures under `evidence/broadcasts/` |
| Decision | **conditional** (candidate signal only — never sole truth) |

Rationale: Useful free per-match text signal, but structurally biased toward Globo platforms and unsuitable as the only broadcast authority. Must be merged with other editorials + maintainer confirmation; treat omissions of CazéTV / Prime Video / Record as expected failure mode.

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
| Can 90% coverage be reached for Serie A? | Only with multi-source editorial coverage + explicit incomplete state + maintainer fill of omitted outlets (CazéTV, Prime, etc.) |
| Can 97% accuracy be reached without humans? | **No.** Single-editorial scrapes systematically miss channels; late rights changes add noise |
| Plausible MVP path? | **Conditional:** multi-source candidates + maintainer panel (REQ-013) as the path to publishable truth; automated list alone must stay labeled incomplete |

## Adapter implications

1. Store every broadcast claim with source URL, observed text, and verification timestamp.
2. Default confidence low; mark lists as `incomplete` until maintainer confirmation or multi-source concordance.
3. When reconciling sources, **union** channels across editorials — do not trust the shortest Globo-centric list.
4. Never notify on unconfirmed oscillations (REQ-012); never notify from a single unconfirmed editorial scrape.
5. Prefer not launching push-for-broadcast until beta accuracy is measured with omission-aware sampling (false negatives for CazéTV/Prime/Record count as failures).
