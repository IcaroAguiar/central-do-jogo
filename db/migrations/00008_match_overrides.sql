-- +goose Up
CREATE TABLE match_overrides (
    id TEXT PRIMARY KEY,
    match_id TEXT NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    data_type TEXT NOT NULL CHECK (data_type IN ('broadcast', 'lineup', 'news', 'kickoff')),
    field TEXT NOT NULL DEFAULT 'state',
    value TEXT NOT NULL,
    justification TEXT NOT NULL,
    actor TEXT NOT NULL,
    version INT NOT NULL CHECK (version >= 1),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (match_id, data_type, field, version)
);

CREATE INDEX match_overrides_match_id_idx ON match_overrides (match_id);

-- +goose Down
DROP TABLE IF EXISTS match_overrides;
