-- +goose Up
CREATE TABLE analytics_events (
    id TEXT PRIMARY KEY,
    anonymous_id TEXT NOT NULL,
    user_id TEXT REFERENCES users (id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX analytics_events_created_at_idx ON analytics_events (created_at);
CREATE INDEX analytics_events_user_id_idx ON analytics_events (user_id)
    WHERE user_id IS NOT NULL;
CREATE INDEX analytics_events_anonymous_id_idx ON analytics_events (anonymous_id);

-- +goose Down
DROP TABLE IF EXISTS analytics_events;
