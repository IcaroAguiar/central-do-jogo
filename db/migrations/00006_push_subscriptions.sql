-- +goose Up
CREATE TABLE push_subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled_at TIMESTAMPTZ,
    UNIQUE (endpoint)
);

CREATE INDEX push_subscriptions_user_id_idx ON push_subscriptions (user_id);
CREATE INDEX push_subscriptions_disabled_at_idx ON push_subscriptions (disabled_at)
    WHERE disabled_at IS NOT NULL;

CREATE TABLE push_outbox (
    id TEXT PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    alert_type TEXT NOT NULL,
    match_id TEXT,
    version TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'failed', 'dead')),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at TIMESTAMPTZ
);

CREATE INDEX push_outbox_status_idx ON push_outbox (status, created_at);

-- +goose Down
DROP TABLE IF EXISTS push_outbox;
DROP TABLE IF EXISTS push_subscriptions;
