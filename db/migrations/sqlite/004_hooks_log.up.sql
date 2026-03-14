-- 004_hooks_log
-- Audit log of hook events for debugging and observability.

CREATE TABLE IF NOT EXISTS hooks_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    payload    TEXT,                            -- JSON
    status     TEXT NOT NULL DEFAULT 'ok',      -- 'ok' | 'error'
    error_msg  TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hooks_log_event
    ON hooks_log(event_type, created_at DESC);
