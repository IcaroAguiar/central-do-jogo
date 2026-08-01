# News source research

Evaluation date: 2026-08-01 (America/Sao_Paulo)  
Scope: up to five related links per match, 72h window, attribution, no full-text republication (REQ-009).

## Candidates

### `gazetaesportiva_rss`

| Field | Value |
|-------|-------|
| Feed | https://www.gazetaesportiva.com/feed/ |
| HTTP | 200, XML RSS |
| robots | allows `/feed/` |
| Auth | none |
| Evidence | `evidence/news/gazetaesportiva-feed.meta.json` |
| Decision | **viable** for allowlist trial |

Rationale: Structured RSS, public, robots-friendly. Items are general sports — match linking needs club/competition keyword filters and dedupe.

### `ge_globo_rss_plantao_futebol`

| Field | Value |
|-------|-------|
| URL tried | https://ge.globo.com/rss/plantao/futebol/ |
| HTTP | **400** Bad Request on 2026-08-01 probe |
| Decision | **rejected** (endpoint not usable as probed) |

### Club site RSS / HTML

| Field | Value |
|-------|-------|
| Flamengo robots | `User-agent: *` / `Disallow:` empty (permissive) |
| Decision | **conditional** per club — evaluate individually before allowlisting |

### Aggregators requiring API keys / ToS-hostile scrapes

| Field | Value |
|-------|-------|
| Decision | **rejected** |

## Allowlist proposal (initial)

1. `gazetaesportiva.com` (RSS)
2. Add club official news feeds only after per-club SEC-004 sheets
3. ge.globo HTML articles only if a stable, robots-allowed feed/API alternative appears — not in MVP v0 ingest

## Product rules

- Store title, outlet, published time, canonical URL only.
- Dedupe by normalized title + URL.
- Prefer items within 72h before kickoff; otherwise show empty/`not_found` rather than stale filler.
- Never copy full article body into the product database.
