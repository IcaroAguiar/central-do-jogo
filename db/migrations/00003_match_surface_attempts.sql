-- +goose Up
ALTER TABLE matches
    ADD COLUMN broadcast_last_attempt_at TIMESTAMPTZ,
    ADD COLUMN lineup_last_attempt_at TIMESTAMPTZ,
    ADD COLUMN news_last_attempt_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE matches
    DROP COLUMN IF EXISTS news_last_attempt_at,
    DROP COLUMN IF EXISTS lineup_last_attempt_at,
    DROP COLUMN IF EXISTS broadcast_last_attempt_at;
