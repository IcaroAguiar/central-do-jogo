-- +goose Up
CREATE TABLE clubs (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    short_name TEXT NOT NULL DEFAULT '',
    aliases TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE season_clubs (
    season INT NOT NULL CHECK (season >= 1900),
    club_id TEXT NOT NULL REFERENCES clubs (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (season, club_id)
);

CREATE TABLE competitions (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    season INT NOT NULL CHECK (season >= 1900),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (slug, season)
);

CREATE TABLE matches (
    id TEXT PRIMARY KEY,
    competition_id TEXT NOT NULL REFERENCES competitions (id),
    home_club_id TEXT NOT NULL REFERENCES clubs (id),
    away_club_id TEXT NOT NULL REFERENCES clubs (id),
    slug TEXT NOT NULL UNIQUE,
    round TEXT NOT NULL DEFAULT '',
    venue TEXT NOT NULL DEFAULT '',
    kickoff_at TIMESTAMPTZ,
    kickoff_state TEXT NOT NULL CHECK (kickoff_state IN ('published', 'indefinite', 'changed')),
    broadcast_state TEXT NOT NULL CHECK (broadcast_state IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    lineup_state TEXT NOT NULL CHECK (lineup_state IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    news_state TEXT NOT NULL CHECK (news_state IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (home_club_id <> away_club_id),
    CHECK (
        (kickoff_state = 'indefinite' AND kickoff_at IS NULL)
        OR (kickoff_state <> 'indefinite')
    )
);

CREATE INDEX matches_competition_id_idx ON matches (competition_id);
CREATE INDEX matches_kickoff_at_idx ON matches (kickoff_at);

CREATE TABLE sources (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    home_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE evidence (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources (id),
    match_id TEXT REFERENCES matches (id),
    data_type TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL,
    parser_version TEXT NOT NULL,
    run_id TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    raw_ref TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX evidence_match_id_idx ON evidence (match_id);
CREATE INDEX evidence_source_id_idx ON evidence (source_id);

CREATE TABLE broadcasts (
    id TEXT PRIMARY KEY,
    match_id TEXT NOT NULL REFERENCES matches (id),
    evidence_id TEXT NOT NULL REFERENCES evidence (id),
    channel TEXT NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    access TEXT NOT NULL CHECK (access IN ('free', 'subscription', 'unknown')),
    region TEXT NOT NULL DEFAULT '',
    official_url TEXT NOT NULL DEFAULT '',
    confidence TEXT NOT NULL CHECK (confidence IN ('high', 'medium', 'low')),
    verified_at TIMESTAMPTZ NOT NULL,
    availability TEXT NOT NULL CHECK (availability IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX broadcasts_match_id_idx ON broadcasts (match_id);

CREATE TABLE lineups (
    id TEXT PRIMARY KEY,
    match_id TEXT NOT NULL REFERENCES matches (id),
    club_id TEXT NOT NULL REFERENCES clubs (id),
    evidence_id TEXT NOT NULL REFERENCES evidence (id),
    side TEXT NOT NULL CHECK (side IN ('home', 'away')),
    formation TEXT NOT NULL DEFAULT '',
    coach TEXT NOT NULL DEFAULT '',
    players JSONB NOT NULL DEFAULT '[]'::jsonb,
    official BOOLEAN NOT NULL DEFAULT false,
    availability TEXT NOT NULL CHECK (availability IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (match_id, side, evidence_id)
);

CREATE INDEX lineups_match_id_idx ON lineups (match_id);

CREATE TABLE news_links (
    id TEXT PRIMARY KEY,
    match_id TEXT NOT NULL REFERENCES matches (id),
    evidence_id TEXT NOT NULL REFERENCES evidence (id),
    source_id TEXT NOT NULL REFERENCES sources (id),
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    availability TEXT NOT NULL CHECK (availability IN ('available', 'awaiting_publication', 'not_found', 'divergent', 'no_coverage')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (match_id, url)
);

CREATE INDEX news_links_match_id_idx ON news_links (match_id);

CREATE TABLE audit_events (
    id BIGSERIAL PRIMARY KEY,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    before_json JSONB,
    after_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_events_entity_idx ON audit_events (entity_type, entity_id);

-- +goose Down
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS news_links;
DROP TABLE IF EXISTS lineups;
DROP TABLE IF EXISTS broadcasts;
DROP TABLE IF EXISTS evidence;
DROP TABLE IF EXISTS sources;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS competitions;
DROP TABLE IF EXISTS season_clubs;
DROP TABLE IF EXISTS clubs;
