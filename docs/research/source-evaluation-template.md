# Source evaluation template

Use one copy of this template per candidate source. Fill every field before proposing an adapter. Prefer structured public data over HTML.

## Identity

| Field | Value |
|-------|-------|
| Source ID (slug) | |
| Display name | |
| Owner / publisher | |
| Primary URL | |
| Data type(s) | schedule / broadcast / lineup / news / other |
| Competitions / clubs covered | |
| Evaluator | |
| Evaluation date (America/Sao_Paulo) | |

## Access and terms (SEC-004)

| Field | Value |
|-------|-------|
| Access method | public HTTP / feed / documented API / HTML page / other |
| Authentication required? | no / yes (describe, without secrets) |
| robots.txt / crawl policy summary | |
| Terms of use summary (attribution, redistribution, commercial use) | |
| Explicit prohibition of automated access? | no / yes / unclear |
| Attribution required in product UI? | |
| Removal / takedown contact if known | |

## Technical shape

| Field | Value |
|-------|-------|
| Format | JSON / XML / RSS / Atom / HTML / image / mixed |
| Stable URL pattern? | |
| Sample endpoint or page path (public only) | |
| Observed fields available | |
| Missing fields vs MVP need | |
| Freshness / update cadence observed | |
| Timezone of published timestamps | |
| Rate limit / throttling observed or documented | |
| Caching guidance | |
| Stability risk (layout/API churn) | low / medium / high |

## Coverage and quality

| Field | Value |
|-------|-------|
| Coverage estimate for supported Serie A clubs | |
| Latency vs official publication (lineups/broadcasts) | |
| Can support REQ-023 / REQ-024 path? | yes / conditional / no |
| Provenance fields preservable (source, observed value, instant) | |
| Known failure modes | |

## Evidence

| Field | Value |
|-------|-------|
| Redacted fixture path under `docs/research/evidence/` | |
| Capture method | |
| Capture timestamp (UTC) | |
| What was redacted and why | |

## Decision

| Field | Value |
|-------|-------|
| Decision | `viable` / `conditional` / `rejected` |
| Conditions (if conditional) | |
| Adapter priority | |
| Rationale | |
| Follow-up work | |

## Checklist before matrix row

- [ ] Terms/robots reviewed without inventing permission
- [ ] No secrets, cookies, or private URLs recorded
- [ ] Redacted evidence committed or linked under `docs/research/evidence/`
- [ ] Decision is explicit and dated
- [ ] Row ready for `docs/research/source-matrix.md`
