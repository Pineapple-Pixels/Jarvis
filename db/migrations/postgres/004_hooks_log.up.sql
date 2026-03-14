-- 004_hooks_log

CREATE TABLE IF NOT EXISTS hooks_log (
    id         SERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload    JSONB,
    status     TEXT NOT NULL DEFAULT 'ok',
    error_msg  TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hooks_log_event
    ON hooks_log(event_type, created_at DESC);

-- Auto-cleanup: partition-friendly index for deleting old logs
CREATE INDEX IF NOT EXISTS idx_hooks_log_created
    ON hooks_log(created_at);
