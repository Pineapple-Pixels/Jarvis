-- 002_cron_jobs

CREATE TABLE IF NOT EXISTS cron_jobs (
    id                  TEXT PRIMARY KEY,
    hour                INTEGER NOT NULL,
    minute              INTEGER NOT NULL,
    weekday             INTEGER,
    prompt              TEXT    NOT NULL DEFAULT '',
    delivery_mode       TEXT    NOT NULL DEFAULT 'log',
    delivery_to         TEXT    NOT NULL DEFAULT '',
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at         TIMESTAMPTZ,
    last_run_status     TEXT,
    consecutive_errors  INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
