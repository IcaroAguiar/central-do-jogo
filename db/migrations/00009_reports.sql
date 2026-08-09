-- +goose Up
CREATE TABLE reports (
    id TEXT PRIMARY KEY,
    context_type TEXT NOT NULL CHECK (context_type IN ('match', 'club', 'other')),
    context_slug TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'reviewed', 'dismissed')),
    ip_hash TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);

CREATE INDEX reports_status_created_at_idx ON reports (status, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS reports;
