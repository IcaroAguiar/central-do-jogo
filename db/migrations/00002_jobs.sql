-- +goose Up
CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    job_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'dead')) DEFAULT 'pending',
    lease_owner TEXT NOT NULL DEFAULT '',
    lease_expires_at TIMESTAMPTZ,
    run_after TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    idempotency_key TEXT NOT NULL UNIQUE,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX jobs_status_run_after_idx ON jobs (status, run_after)
    WHERE status IN ('pending', 'failed');

CREATE TABLE job_attempts (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    attempt_no INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    error TEXT NOT NULL DEFAULT '',
    outcome TEXT NOT NULL CHECK (outcome IN ('running', 'success', 'error')) DEFAULT 'running',
    UNIQUE (job_id, attempt_no)
);

CREATE INDEX job_attempts_job_id_idx ON job_attempts (job_id);

CREATE TABLE source_health (
    source_id TEXT PRIMARY KEY REFERENCES sources (id),
    last_success_at TIMESTAMPTZ,
    last_error_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    next_run_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    consecutive_failures INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS source_health;
DROP TABLE IF EXISTS job_attempts;
DROP TABLE IF EXISTS jobs;
