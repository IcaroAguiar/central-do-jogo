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
3. **False completeness:** scraping a single editorial **or** a single club post and rendering it as “todas as transmissões” would violate REQ-007 and destroy REQ-023 accuracy via **false negatives** (missing channels), not only false positives.
4. **Club coverage gaps:** many official club calendars/jogo hubs do not expose channel text in HTML; useful signals often live only in matchday posts or PDF guides — uneven across the Serie A snapshot.

**Product rule:** automated editorial and club claims are **candidate / incomplete by default**. The UI must not imply an exhaustive list until a maintainer confirms, or until multiple independent sources agree on the full set. Missing a free outlet (e.g. CazéTV) is as bad as inventing one. Club matchday posts are especially useful for catching non-Globo holders that Globo editorials skip.

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

### `club_official_matchday_html`

| Field | Value |
|-------|-------|
| Owner | Individual Serie A clubs (heterogeneous CMS) |
| Example URLs | Atlético Mineiro matchday posts under `/…-hoje-tem-jogo…` — see `evidence/broadcasts/*.club-matchday.redacted.json` |
| Access | public HTTP on many clubs; some sites return bot walls (e.g. Cloudflare/captcha on probed Corinthians/Vasco surfaces) |
| Format | HTML news/matchday posts; sometimes PDF “guia da partida” (Botafogo, Bahia) without reliable free structured text in this pass |
| Fields observed | Explicit `Transmissão:` label on Atlético posts with channel names (`SporTV`, `Premiere`/`Premiete`, `Prime Vídeo`); club streaming (e.g. Galotv) often listed separately as pré-jogo/narração |
| Missing vs REQ-007 | Not universal across clubs; no stable free-vs-pay/region/deep-link; typos and club-centric wording; PDF guides not automatable without fragile PDF/OCR |
| Evidence | three redacted Atlético captures under `evidence/broadcasts/` |
| Decision | **conditional** (high-value candidate class for multi-source union — never sole truth) |

Rationale: Club pages were underweighted in the first pass. Where clubs publish matchday service posts, they can surface non-Globo holders (e.g. Prime Video) that Globo-family editorials omit. Coverage is **uneven** (calendars/jogo hubs for Flamengo, São Paulo, Internacional often lack channel text; Botafogo/Bahia lean on PDFs). Treat as per-club adapters after SEC-004, robots checks, and fixtures — not one generic scraper.

Related probes (same pass, not promoted to primary evidence):

- Atlético: `/partida/…` and guia hubs — weak/no channel field; matchday “hoje tem jogo” posts are the useful surface.
- Botafogo `/guia-da-partida`, Bahia `/guia-das-partidas/` — PDF guides linked; not used as automated MVP path here.
- Flamengo `/jogos/…`, São Paulo calendário, Internacional jogo/serviço pages — HTTP 200 but no reliable `Transmissão` prose in the probed HTML.

### Official competition rights feed / sole club authority

| Field | Value |
|-------|-------|
| Observation | No free, comprehensive, structured official rights feed found; club pages are partial and club-specific |
| Decision | **rejected** as sole automated MVP source (club matchday HTML remains a **conditional candidate** above) |

### Paid sports data APIs

| Field | Value |
|-------|-------|
| Decision | **rejected** as mandatory dependency (CON-001) |

## REQ-023 feasibility

| Question | Assessment |
|----------|------------|
| Can channel names be collected legally from public HTML? | Plausible for editorial **and** selected club matchday pages with robots-aware crawling and attribution |
| Can 90% coverage be reached for Serie A? | Only with multi-source coverage (editorials **+** club matchday/HTML where available) + explicit incomplete state + maintainer fill of omitted outlets (CazéTV, Prime, etc.) |
| Can 97% accuracy be reached without humans? | **No.** Single-source scrapes systematically miss channels; club coverage is uneven; late rights changes add noise |
| Plausible MVP path? | **Conditional:** multi-source candidates (ge + other editorials + club official matchday HTML) + maintainer panel (REQ-013); automated list alone must stay labeled incomplete |

## Adapter implications

1. Store every broadcast claim with source URL, observed text, and verification timestamp.
2. Default confidence low; mark lists as `incomplete` until maintainer confirmation or multi-source concordance.
3. When reconciling sources, **union** channels across editorials **and** club matchday posts — do not trust the shortest Globo-centric list.
4. Prefer club `Transmissão` (or equivalent) fields when present as an independent signal for non-Globo holders; do not conflate club OTT/narration products with broadcast rights.
5. Never notify on unconfirmed oscillations (REQ-012); never notify from a single unconfirmed scrape (editorial or club).
6. Prefer not launching push-for-broadcast until beta accuracy is measured with omission-aware sampling (false negatives for CazéTV/Prime/Record count as failures).
7. Onboard club adapters one club at a time (SEC-004 + fixtures); skip PDF-only guides until a reviewed PDF strategy exists.
