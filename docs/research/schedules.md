# Schedule source research

Evaluation date: 2026-08-01 (America/Sao_Paulo)  
Capture window (UTC): 2026-08-01T02:01:41Z  
Goal: compare public/free sources for Serie A, Copa do Brasil, Libertadores, and Sudamericana schedules.

## Candidates

### `cbf_official_site`

| Field | Value |
|-------|-------|
| Owner | Confederação Brasileira de Futebol |
| Primary URL | https://www.cbf.com.br/ |
| Data types | schedule (national competitions) |
| Access | public HTTP; no auth observed |
| robots.txt | `GET /robots.txt` returned HTML 404 page (no machine-readable allow/deny) |
| Terms | site terms/privacy apply; no explicit automated-access license found in this pass |
| Format | HTML news + downloadable PDF “tabela básica” |
| Freshness | seasonal PDF + ongoing site updates; kickoff changes require re-check |
| Evidence | `evidence/schedules/cbf-tabela-basica-2026.meta.json` |
| Decision | **conditional** |

Rationale: Official for Brasileirão / Copa do Brasil context. Structured machine API not found. PDF + HTML are admissible with attribution, rate limits, and change detection. Parser fragility is the main risk.

### `openfootball_brazil`

| Field | Value |
|-------|-------|
| Owner | openfootball community (GitHub) |
| Primary URL | https://github.com/openfootball/brazil |
| Sample | https://raw.githubusercontent.com/openfootball/brazil/master/brazil/2026_br1.txt |
| Format | Football.TXT (convertible to JSON) |
| Auth | none |
| Terms | project LICENSE on repo; community-maintained, not an official federation feed |
| Evidence | `evidence/schedules/openfootball-brazil-2026-br1.excerpt.txt` |
| Decision | **conditional** (secondary / bootstrap only) |

Rationale: Excellent structured fixture text for Serie A 2026, including kickoff local times. Not authoritative vs CBF; use for scaffolding/tests and cross-check, never as sole provenance for user-facing truth.

### `football_data_org`

| Field | Value |
|-------|-------|
| Owner | football-data.org |
| Probe | `GET https://api.football-data.org/v4/competitions` → 200; lists `BSA`, `CDB`, `CLI`, `CS` |
| Restricted | `GET /v4/competitions/BSA` → **403** without paid token |
| Decision | **rejected** for free MVP path |

Rationale: CON-001 forbids mandatory paid services. Free tier does not unlock Brazilian match payloads needed for MVP.

### `dadosfutebol` / `api-futebol`

| Field | Value |
|-------|-------|
| Access | documented JSON APIs with **API key** |
| Decision | **rejected** as mandatory free path; **conditional** only as optional operator-paid enrichment |

Rationale: Useful later if an operator opts in; cannot be required for OSS low-cost promise.

### `ge_globo_agenda`

| Field | Value |
|-------|-------|
| URL | https://ge.globo.com/agenda/ |
| robots | https://ge.globo.com/robots.txt — general crawl allowed with Disallows for `/servico`, `/dynamo`, etc. |
| Format | large HTML |
| Decision | **conditional** secondary calendar UX source |

Rationale: Public editorial agenda exists. Broadcast/lineup truth still needs per-match pages and provenance rules.

### `conmebol_site`

| Field | Value |
|-------|-------|
| URL | https://www.conmebol.com/pt-br/competicoes/ |
| Libertadores hub | https://www.conmebol.com/pt-br/competicoes/conmebol-libertadores-2026/ (HTTP 200) |
| robots | `/robots.txt` returned nginx 404 HTML |
| Format | HTML |
| Decision | **conditional** for continental fixtures of Serie A clubs |

## Coverage judgment

| Competition | Viable free path? | Notes |
|-------------|-------------------|-------|
| Serie A | yes, conditional | CBF PDF/HTML primary; openfootball secondary |
| Copa do Brasil | conditional | CBF site/PDF expected; less structured feed observed in this pass |
| Libertadores | conditional | CONMEBOL HTML; no free JSON confirmed |
| Sudamericana | conditional | same as Libertadores |

## Adapter implications

1. Prefer CBF PDF/HTML ingest with immutable evidence blobs.
2. Keep openfootball fixtures for deterministic tests (committed redacted excerpts only).
3. Do not hard-depend on football-data.org paid tiers.
4. Continental competitions need dedicated HTML adapters and may ship after Serie A if capacity is limited.
