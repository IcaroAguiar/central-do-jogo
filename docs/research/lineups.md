# Lineup source research

Evaluation date: 2026-08-01 (America/Sao_Paulo)  
Focus: REQ-024 (≈95% correct within 5 minutes of **official observable** publication).

## Candidates

### `cbf_match_center`

| Field | Value |
|-------|-------|
| Example | https://www.cbf.com.br/futebol-brasileiro/jogos/campeonato-brasileiro/serie-a/2025/317445/fluminense-x-gremio/829305?view=escalacao |
| HTTP | 200 on 2026-08-01 probe |
| Format | HTML with headings `Escalação` and `Reservas`; shirt numbers + names in text |
| robots | `/robots.txt` not machine-readable (HTML 404) |
| Auth | none |
| Formation / coach | not observed on sampled page |
| Evidence | `evidence/lineups/cbf-fluminense-gremio-2025.redacted.json` |
| Decision | **conditional** (primary official for national matches) |

Rationale: Best free official structured-enough HTML for Brazilian domestic lineups. Publication lag vs club social graphics is a known product risk.

### `ge_globo_editorial`

| Field | Value |
|-------|-------|
| Examples | Inter x Flamengo, Coritiba x Cruzeiro, Grêmio x Flamengo “onde assistir / escalações” pages |
| Format | HTML; often **probable** XIs (`Escalação provável`) or editorial named XIs |
| Evidence | three files under `evidence/lineups/*ge-editorial*` |
| Decision | **conditional** secondary |

Rationale: Useful early signal, but not a substitute for official CBF publication when defining REQ-024 “official observable”.

### `club_official_social`

| Field | Value |
|-------|-------|
| Format | image cards on X/Instagram |
| Decision | **rejected** for free automated MVP |

Rationale: ToS/API cost, login walls, and OCR fragility across 20 club designs.

### `conmebol_gol_portal`

| Field | Value |
|-------|-------|
| URL | https://gol.conmebol.com/ (HTTP 200 shell) |
| Observation | no lineup payload extracted from the landing page in this pass |
| Decision | **conditional / unverified** for continental match centers — needs fixture-level follow-up before adapter work |

## Three real cases captured

1. CBF Fluminense x Grêmio (2025 page still public) — official HTML starters/bench sample.
2. ge Inter x Flamengo (2026-07-29) — editorial probable XI + broadcast mentions.
3. ge Coritiba x Cruzeiro (2026-07-30) — editorial named XI + SporTV/Premiere mention.
4. (extra) ge Grêmio x Flamengo (2026-05-10) — additional editorial sample.

## REQ-024 feasibility

Define **official observable publication** for MVP as: first public CBF match-center HTML that contains `Escalação` with named starters.

Under that definition, a free adapter with adaptive polling in the pre-match window is **plausible**.

If “official” is redefined as club social image time, free automation is **not plausible** without paid social APIs/OCR.

## Adapter implications

1. Primary: `cbf_match_center` HTML adapter + fixtures map.
2. Secondary: ge editorial for early “awaiting official” UI states — never upgrade confidence to official without CBF (or maintainer override).
3. Preserve evidence blobs for every published lineup version.
