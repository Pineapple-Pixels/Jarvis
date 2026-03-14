-- 002_cron_jobs
-- Persistent storage for cron job state so jobs survive restarts.

CREATE TABLE IF NOT EXISTS cron_jobs (
    id                  TEXT PRIMARY KEY,
    hour                INTEGER NOT NULL,
    minute              INTEGER NOT NULL,
    weekday             INTEGER,               -- 0=Sunday..6=Saturday, NULL=every day
    prompt              TEXT    NOT NULL DEFAULT '',
    delivery_mode       TEXT    NOT NULL DEFAULT 'log',   -- 'log' | 'whatsapp' | 'webhook'
    delivery_to         TEXT    NOT NULL DEFAULT '',
    enabled             INTEGER NOT NULL DEFAULT 1,
    last_run_at         DATETIME,
    last_run_status     TEXT,                  -- 'ok' | 'error'
    consecutive_errors  INTEGER NOT NULL DEFAULT 0,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);
