-- +goose Up
CREATE TABLE user_preferences (
    user_id TEXT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    primary_club_slug TEXT,
    favorite_club_slugs TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_preferences;
