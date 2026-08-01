# Source matrix (Phase 0)

Date: 2026-08-01  
Issue: https://github.com/IcaroAguiar/central-do-jogo/issues/2  
Status legend: `viable` | `conditional` | `rejected`

## MVP allowlist actually supportable

**In scope for first product slice (after GOAL-001 accept):**

- Clubs: official Serie A 2026 snapshot clubs
- Competitions prioritized: **Serie A first**
- Copa do Brasil / Libertadores / Sudamericana: **deferred** until dedicated adapters pass the same matrix bar (still researched as conditional)

**Data types:**

| Data | Primary source | Backup | Status |
|------|----------------|--------|--------|
| Schedule | `cbf_official_site` (PDF/HTML) | `openfootball_brazil` (non-authoritative) | conditional |
| Broadcast | maintainer-confirmed set (REQ-013) | multi-editorial candidates (ge + others); never single ge list alone | conditional |
| Lineup | `cbf_match_center` | ge editorial as early non-official signal | conditional |
| News links | `gazetaesportiva_rss` | club feeds after SEC-004 | viable / conditional |

## Matrix rows

| Data | Competition | Source ID | Status | Notes |
|------|-------------|-----------|--------|-------|
| schedule | Serie A | cbf_official_site | conditional | PDF/HTML official; no free JSON API |
| schedule | Serie A | openfootball_brazil | conditional | structured; community; secondary only |
| schedule | Serie A | football_data_org | rejected | BSA payload 403 without paid plan |
| schedule | Serie A | dadosfutebol / api-futebol | rejected | API key required; not free-default |
| schedule | Copa do Brasil | cbf_official_site | conditional | expected official path; less probe depth |
| schedule | Libertadores | conmebol_site | conditional | HTML hubs public |
| schedule | Sudamericana | conmebol_site | conditional | HTML hubs public |
| broadcast | Serie A | ge_globo_editorial | conditional | candidate only; often omits CazéTV/Prime/Record; not exhaustive |
| broadcast | Serie A | multi_editorial_union + maintainer | conditional | required path for publishable “full” broadcast lists |
| broadcast | Serie A | official rights feed | rejected | not found free/structured |
| broadcast | Copa do Brasil | ge_globo_editorial | conditional | same class of editorial pages (unverified volume) |
| broadcast | Libertadores / Sudamericana | editorial HTML | conditional | fragmented rights; high ops risk |
| lineup | Serie A | cbf_match_center | conditional | verified HTML Escalação/Reservas |
| lineup | Serie A | ge_globo_editorial | conditional | often probable XI; not official |
| lineup | Serie A | club_official_social | rejected | images/API walls |
| lineup | Libertadores / Sudamericana | conmebol match center | conditional | landing verified; fixture depth unverified |
| news | all MVP clubs | gazetaesportiva_rss | viable | RSS allowed by robots |
| news | all MVP clubs | ge_rss_plantao | rejected | HTTP 400 on probed URL |

## REQ gates (preview)

| Gate | Plausible free path? | Condition |
|------|----------------------|-----------|
| REQ-023 broadcasts | conditional | multi-source + maintainer; score false negatives (missing CazéTV/Prime/etc.) |
| REQ-024 lineups | conditional | define official instant as CBF match-center HTML publish time |

## References

- `schedules.md`, `broadcasts.md`, `lineups.md`, `news.md`
- Redacted evidence under `evidence/`
- Template: `source-evaluation-template.md`
